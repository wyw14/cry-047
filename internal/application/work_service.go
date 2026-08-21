package application

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/wyw14/cry-047/internal/domain"
)

func (s *Service) GenerateWorkWindow(ctx context.Context, actor domain.Actor, cmd domain.GenerateWindow) (domain.WorkWindow, error) {
	if err := authorize(actor.CanPlan()); err != nil {
		return domain.WorkWindow{}, err
	}
	if err := cmd.Validate(); err != nil {
		return domain.WorkWindow{}, err
	}
	var generated domain.WorkWindow
	err := s.store.Update(ctx, func(tx Transaction) error {
		if existing, ok := tx.FindWindowByKey(cmd.CommandKey); ok {
			generated = existing
			return nil
		}
		program, err := tx.GetProgram(cmd.ProgramID)
		if err != nil {
			return err
		}
		if !program.Active {
			return fmt.Errorf("program is inactive: %w", domain.ErrInvalidState)
		}
		revision, err := program.RevisionAt(cmd.Occurrence)
		if err != nil {
			return err
		}
		facility, err := tx.GetFacility(program.FacilityID)
		if err != nil {
			return err
		}
		person, err := tx.GetPerson(cmd.AssigneeID)
		if err != nil || !person.Available {
			return fmt.Errorf("assignee unavailable: %w", domain.ErrInvalidState)
		}
		stableOccurrence := cmd.Occurrence.UTC().Truncate(time.Minute)
		for _, existing := range tx.ListWindows() {
			if existing.ProgramID == program.ID && existing.Occurrence.Equal(stableOccurrence) && existing.State != domain.WorkSkipped {
				return fmt.Errorf("window already generated: %w", domain.ErrConflict)
			}
		}
		now := s.clock.Now()
		generated = domain.WorkWindow{
			ID: s.ids.New("window"), ProgramID: program.ID, FacilityID: facility.ID, ProgramRevision: revision.Number,
			Occurrence: stableOccurrence, DueAt: dueAt(stableOccurrence, facility.Criticality), AssigneeID: cmd.AssigneeID,
			State: domain.WorkPlanned, IdempotencyKey: cmd.CommandKey,
			Versioned: domain.Versioned{Version: 1, CreatedAt: now, UpdatedAt: now},
		}
		if err := tx.PutWindow(generated); err != nil {
			return err
		}
		tx.PutPersonalTask(domain.PersonalTask{
			ID: s.ids.New("task"), PersonID: generated.AssigneeID, FacilityID: generated.FacilityID, WindowID: generated.ID,
			Title: "执行" + facility.Name + "预防性维护", DueAt: generated.DueAt, Priority: priorityFor(facility.Criticality),
			DedupeKey: domain.StableKey("window", string(generated.ID)), Versioned: domain.Versioned{Version: 1, CreatedAt: now, UpdatedAt: now},
		})
		s.audit(tx, actor, "window.generated", "work_window", generated.ID, nil, generated)
		s.timeline(tx, facility.ID, domain.TimelineScheduled, "生成维护工单", program.Title, generated.ID, actor)
		s.notify(tx, generated.AssigneeID, "新的维护工单", facility.Name, generated.ID)
		return nil
	})
	return generated, err
}

func priorityFor(criticality domain.Criticality) int {
	switch criticality {
	case domain.CriticalityCritical:
		return 100
	case domain.CriticalityImportant:
		return 60
	default:
		return 30
	}
}

func (s *Service) ReassignWorkWindow(ctx context.Context, actor domain.Actor, cmd domain.ReassignWindow) (domain.WorkWindow, error) {
	if err := authorize(actor.CanPlan()); err != nil {
		return domain.WorkWindow{}, err
	}
	if strings.TrimSpace(cmd.Reason) == "" || cmd.FromPersonID == cmd.ToPersonID {
		return domain.WorkWindow{}, domain.Invalid("reason", "改派原因不能为空且新旧执行人不能相同")
	}
	var result domain.WorkWindow
	err := s.store.Update(ctx, func(tx Transaction) error {
		window, err := tx.GetWindow(cmd.WindowID)
		if err != nil {
			return err
		}
		if err := expectedEpoch(window.Version, cmd.ExpectedEpoch); err != nil {
			return err
		}
		if window.AssigneeID != cmd.FromPersonID {
			return fmt.Errorf("assignee changed concurrently: %w", domain.ErrConflict)
		}
		if window.State != domain.WorkPlanned && window.State != domain.WorkRejected {
			return fmt.Errorf("window cannot be reassigned in %s: %w", window.State, domain.ErrInvalidState)
		}
		person, err := tx.GetPerson(cmd.ToPersonID)
		if err != nil || !person.Available {
			return fmt.Errorf("new assignee unavailable: %w", domain.ErrInvalidState)
		}
		before := window
		window.AssigneeID = cmd.ToPersonID
		window.Version++
		window.UpdatedAt = s.clock.Now()
		if err := tx.PutWindow(window); err != nil {
			return err
		}
		facility, _ := tx.GetFacility(window.FacilityID)
		tx.PutPersonalTask(domain.PersonalTask{
			ID: s.ids.New("task"), PersonID: cmd.ToPersonID, FacilityID: window.FacilityID, WindowID: window.ID,
			Title: "接手" + facility.Name + "维护工单", DueAt: window.DueAt, Priority: priorityFor(facility.Criticality),
			DedupeKey: domain.StableKey("reassign", string(window.ID), string(cmd.ToPersonID)),
			Versioned: domain.Versioned{Version: 1, CreatedAt: s.clock.Now(), UpdatedAt: s.clock.Now()},
		})
		s.audit(tx, actor, "window.reassigned", "work_window", window.ID, before, window)
		s.timeline(tx, window.FacilityID, domain.TimelineScheduled, "维护工单改派", cmd.Reason, window.ID, actor)
		s.notify(tx, cmd.FromPersonID, "维护工单已改派", cmd.Reason, window.ID)
		s.notify(tx, cmd.ToPersonID, "收到改派工单", cmd.Reason, window.ID)
		result = window
		return nil
	})
	return result, err
}

func (s *Service) SkipWorkWindow(ctx context.Context, actor domain.Actor, cmd domain.SkipWindow) (domain.WorkWindow, domain.WorkWindow, error) {
	if err := authorize(actor.CanPlan()); err != nil {
		return domain.WorkWindow{}, domain.WorkWindow{}, err
	}
	if len(strings.TrimSpace(cmd.Reason)) < 5 || cmd.CompensateAt.IsZero() {
		return domain.WorkWindow{}, domain.WorkWindow{}, domain.Invalid("reason", "跳过原因和补做时间必须完整")
	}
	var skipped, compensation domain.WorkWindow
	err := s.store.Update(ctx, func(tx Transaction) error {
		window, err := tx.GetWindow(cmd.WindowID)
		if err != nil {
			return err
		}
		if err := expectedEpoch(window.Version, cmd.ExpectedEpoch); err != nil {
			return err
		}
		if window.State != domain.WorkPlanned || !cmd.CompensateAt.After(s.clock.Now()) || cmd.CompensateAt.After(window.DueAt.Add(14*24*time.Hour)) {
			return fmt.Errorf("invalid compensation window: %w", domain.ErrInvalidState)
		}
		before := window
		window.State, window.SkipReason = domain.WorkSkipped, strings.TrimSpace(cmd.Reason)
		window.Version++
		window.UpdatedAt = s.clock.Now()
		if err := tx.PutWindow(window); err != nil {
			return err
		}
		compensation = window
		compensation.ID = s.ids.New("window")
		compensation.Occurrence = cmd.CompensateAt.UTC().Truncate(time.Minute)
		compensation.DueAt = dueAt(compensation.Occurrence, facilityCriticality(tx, window.FacilityID))
		compensation.State = domain.WorkPlanned
		compensation.SkipReason = ""
		compensation.CompensatesID = window.ID
		compensation.IdempotencyKey = domain.StableKey("compensation", string(window.ID))
		compensation.Versioned = domain.Versioned{Version: 1, CreatedAt: s.clock.Now(), UpdatedAt: s.clock.Now()}
		if err := tx.PutWindow(compensation); err != nil {
			return err
		}
		s.audit(tx, actor, "window.skipped", "work_window", window.ID, before, window)
		s.audit(tx, actor, "window.compensation_created", "work_window", compensation.ID, nil, compensation)
		s.timeline(tx, window.FacilityID, domain.TimelineScheduled, "跳过并安排补做", cmd.Reason, compensation.ID, actor)
		s.notify(tx, window.AssigneeID, "维护任务改为补做", cmd.Reason, compensation.ID)
		skipped = window
		return nil
	})
	return skipped, compensation, err
}

func facilityCriticality(tx Transaction, id domain.ID) domain.Criticality {
	facility, err := tx.GetFacility(id)
	if err != nil {
		return domain.CriticalityImportant
	}
	return facility.Criticality
}
