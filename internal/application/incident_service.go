package application

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/wyw14/cry-047/internal/domain"
)

type AdvanceIncident struct {
	IncidentID    domain.ID            `json:"incident_id"`
	Next          domain.IncidentState `json:"next"`
	Note          string               `json:"note"`
	ExpectedEpoch int64                `json:"expected_epoch"`
}

func (s *Service) OpenIncident(ctx context.Context, actor domain.Actor, cmd domain.OpenIncident) (domain.Incident, error) {
	if err := authorize(actor.CanExecute() || actor.CanReview()); err != nil {
		return domain.Incident{}, err
	}
	if err := cmd.Validate(); err != nil {
		return domain.Incident{}, err
	}
	var opened domain.Incident
	err := s.store.Update(ctx, func(tx Transaction) error {
		facility, err := tx.GetFacility(cmd.FacilityID)
		if err != nil {
			return err
		}
		owner, err := tx.GetPerson(cmd.OwnerID)
		if err != nil || !owner.Available {
			return fmt.Errorf("incident owner unavailable: %w", domain.ErrInvalidState)
		}
		if cmd.ExecutionID.Valid() {
			execution, err := tx.GetExecution(cmd.ExecutionID)
			if err != nil || execution.FacilityID != facility.ID {
				return fmt.Errorf("execution does not belong to facility: %w", domain.ErrConflict)
			}
		}
		for _, incident := range tx.ListIncidents() {
			if incident.FacilityID == facility.ID && incident.Summary == strings.TrimSpace(cmd.Summary) && incident.State != domain.IncidentClosed {
				return fmt.Errorf("equivalent incident is open: %w", domain.ErrConflict)
			}
		}
		now := s.clock.Now()
		opened = domain.Incident{
			ID: cmd.ID, FacilityID: cmd.FacilityID, ExecutionID: cmd.ExecutionID, Summary: strings.TrimSpace(cmd.Summary),
			Severity: cmd.Severity, State: domain.IncidentOpen, OwnerID: cmd.OwnerID, OpenedAt: now,
			Versioned: domain.Versioned{Version: 1, CreatedAt: now, UpdatedAt: now},
		}
		beforeFacility := facility
		if cmd.Severity == domain.CriticalityCritical {
			facility.State = domain.FacilityRestricted
		} else if facility.State == domain.FacilityNormal {
			facility.State = domain.FacilityDue
		}
		facility.Version++
		facility.UpdatedAt = now
		if err := tx.PutIncident(opened); err != nil {
			return err
		}
		if err := tx.PutFacility(facility); err != nil {
			return err
		}
		tx.PutPersonalTask(domain.PersonalTask{
			ID: s.ids.New("task"), PersonID: opened.OwnerID, FacilityID: opened.FacilityID, IncidentID: opened.ID,
			Title: "处置异常：" + opened.Summary, DueAt: now.Add(incidentSLA(opened.Severity)), Priority: priorityFor(opened.Severity),
			DedupeKey: domain.StableKey("incident", string(opened.ID)), Versioned: domain.Versioned{Version: 1, CreatedAt: now, UpdatedAt: now},
		})
		s.audit(tx, actor, "incident.opened", "incident", opened.ID, nil, opened)
		s.audit(tx, actor, "facility.incident_state", "facility", facility.ID, beforeFacility, facility)
		s.timeline(tx, facility.ID, domain.TimelineIncident, "发现设施异常", opened.Summary, opened.ID, actor)
		s.notify(tx, opened.OwnerID, "异常待处置", opened.Summary, opened.ID)
		return nil
	})
	return opened, err
}

func incidentSLA(severity domain.Criticality) time.Duration {
	if severity == domain.CriticalityCritical {
		return 4 * time.Hour
	}
	if severity == domain.CriticalityImportant {
		return 24 * time.Hour
	}
	return 72 * time.Hour
}

func (s *Service) AdvanceIncident(ctx context.Context, actor domain.Actor, cmd AdvanceIncident) (domain.Incident, error) {
	if err := authorize(actor.CanExecute() || actor.CanReview()); err != nil {
		return domain.Incident{}, err
	}
	if len(strings.TrimSpace(cmd.Note)) < 5 {
		return domain.Incident{}, domain.Invalid("note", "处置记录至少五个字符")
	}
	var result domain.Incident
	err := s.store.Update(ctx, func(tx Transaction) error {
		incident, err := tx.GetIncident(cmd.IncidentID)
		if err != nil {
			return err
		}
		if err := expectedEpoch(incident.Version, cmd.ExpectedEpoch); err != nil {
			return err
		}
		if actor.ID != incident.OwnerID && !actor.CanReview() {
			return domain.ErrForbidden
		}
		if !incident.CanMove(cmd.Next) {
			return fmt.Errorf("incident cannot move from %s to %s: %w", incident.State, cmd.Next, domain.ErrInvalidState)
		}
		if (cmd.Next == domain.IncidentRecovered || cmd.Next == domain.IncidentClosed) && !actor.CanReview() {
			return domain.ErrForbidden
		}
		before := incident
		switch cmd.Next {
		case domain.IncidentRectifying:
			incident.Rectification = strings.TrimSpace(cmd.Note)
		case domain.IncidentReinspect:
			incident.ReinspectResult = strings.TrimSpace(cmd.Note)
		case domain.IncidentRecovered:
			incident.RecoveryNote = strings.TrimSpace(cmd.Note)
		case domain.IncidentClosed:
			closed := s.clock.Now()
			incident.ClosedAt = &closed
		}
		incident.State = cmd.Next
		incident.Version++
		incident.UpdatedAt = s.clock.Now()
		if err := tx.PutIncident(incident); err != nil {
			return err
		}
		facility, err := tx.GetFacility(incident.FacilityID)
		if err != nil {
			return err
		}
		beforeFacility := facility
		if cmd.Next == domain.IncidentRecovered {
			facility.State = domain.FacilityRecoveryDue
			facility.Version++
			facility.UpdatedAt = s.clock.Now()
			if err := tx.PutFacility(facility); err != nil {
				return err
			}
			s.audit(tx, actor, "facility.recovery_due", "facility", facility.ID, beforeFacility, facility)
		}
		s.audit(tx, actor, "incident.advanced", "incident", incident.ID, before, incident)
		kind := domain.TimelineRectified
		if cmd.Next == domain.IncidentRecovered || cmd.Next == domain.IncidentClosed {
			kind = domain.TimelineRecovered
		}
		s.timeline(tx, incident.FacilityID, kind, "异常处置进展", string(cmd.Next)+"："+cmd.Note, incident.ID, actor)
		s.notify(tx, facility.ResponsibleID, "异常状态已更新", string(cmd.Next), incident.ID)
		result = incident
		return nil
	})
	return result, err
}
