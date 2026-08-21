package application

import (
	"context"
	"fmt"
	"strings"

	"github.com/wyw14/cry-047/internal/domain"
)

func (s *Service) RegisterFacility(ctx context.Context, actor domain.Actor, cmd domain.RegisterFacility) (domain.Facility, error) {
	if err := authorize(actor.CanPlan()); err != nil {
		return domain.Facility{}, err
	}
	if err := cmd.Validate(); err != nil {
		return domain.Facility{}, err
	}
	var created domain.Facility
	err := s.store.Update(ctx, func(tx Transaction) error {
		if err := checkContext(ctx); err != nil {
			return err
		}
		place, err := tx.GetPlace(cmd.PlaceID)
		if err != nil {
			return fmt.Errorf("load place: %w", err)
		}
		if !place.Active {
			return fmt.Errorf("place is inactive: %w", domain.ErrInvalidState)
		}
		person, err := tx.GetPerson(cmd.ResponsibleID)
		if err != nil || !person.Available {
			return fmt.Errorf("responsible person unavailable: %w", domain.ErrInvalidState)
		}
		for _, alternativeID := range cmd.AlternativeIDs {
			alternative, err := tx.GetFacility(alternativeID)
			if err != nil || alternative.PlaceID != cmd.PlaceID || alternative.State == domain.FacilityRestricted {
				return fmt.Errorf("alternative %s is not usable: %w", alternativeID, domain.ErrInvalidState)
			}
		}
		for _, existing := range tx.ListFacilities() {
			if existing.PlaceID == cmd.PlaceID && strings.EqualFold(existing.Name, cmd.Name) {
				return fmt.Errorf("facility name already exists: %w", domain.ErrConflict)
			}
		}
		now := s.clock.Now()
		created = domain.Facility{
			ID: cmd.ID, PlaceID: cmd.PlaceID, Name: strings.TrimSpace(cmd.Name), Category: strings.TrimSpace(cmd.Category),
			Criticality: cmd.Criticality, State: domain.FacilityNormal, ResponsibleID: cmd.ResponsibleID,
			AlternativeIDs: append([]domain.ID(nil), cmd.AlternativeIDs...),
			Versioned:      domain.Versioned{Version: 1, CreatedAt: now, UpdatedAt: now},
		}
		if err := tx.PutFacility(created); err != nil {
			return err
		}
		s.audit(tx, actor, "facility.registered", "facility", created.ID, nil, created)
		s.timeline(tx, created.ID, domain.TimelineRegistered, "设施登记", "设施已纳入预防性维护台账", created.ID, actor)
		s.notify(tx, created.ResponsibleID, "新增责任设施", created.Name, created.ID)
		return nil
	})
	return created, err
}

func (s *Service) ChangeFacilityState(ctx context.Context, actor domain.Actor, facilityID domain.ID, next domain.FacilityState, expected int64) (domain.Facility, error) {
	if err := authorize(actor.CanReview()); err != nil {
		return domain.Facility{}, err
	}
	var changed domain.Facility
	err := s.store.Update(ctx, func(tx Transaction) error {
		facility, err := tx.GetFacility(facilityID)
		if err != nil {
			return err
		}
		if err := expectedEpoch(facility.Version, expected); err != nil {
			return err
		}
		openIncident := false
		for _, incident := range tx.ListIncidents() {
			if incident.FacilityID == facilityID && incident.State != domain.IncidentClosed {
				openIncident = true
				break
			}
		}
		if err := facility.CanTransition(next, openIncident); err != nil {
			return err
		}
		before := facility
		facility.State = next
		facility.Version++
		facility.UpdatedAt = s.clock.Now()
		if next == domain.FacilityNormal {
			recovered := s.clock.Now()
			facility.LastRecovered = &recovered
		}
		if err := tx.PutFacility(facility); err != nil {
			return err
		}
		s.audit(tx, actor, "facility.state_changed", "facility", facility.ID, before, facility)
		s.timeline(tx, facility.ID, domain.TimelineStateChange, "设施状态变更", string(before.State)+" -> "+string(next), facility.ID, actor)
		s.notify(tx, facility.ResponsibleID, "设施状态已变更", facility.Name+"："+string(next), facility.ID)
		changed = facility
		return nil
	})
	return changed, err
}
