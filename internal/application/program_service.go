package application

import (
	"context"
	"fmt"
	"sort"

	"github.com/wyw14/cry-047/internal/domain"
)

func (s *Service) PublishProgram(ctx context.Context, actor domain.Actor, cmd domain.PublishProgram) (domain.MaintenanceProgram, error) {
	if err := authorize(actor.CanPlan()); err != nil {
		return domain.MaintenanceProgram{}, err
	}
	if err := cmd.Validate(); err != nil {
		return domain.MaintenanceProgram{}, err
	}
	var result domain.MaintenanceProgram
	err := s.store.Update(ctx, func(tx Transaction) error {
		facility, err := tx.GetFacility(cmd.FacilityID)
		if err != nil {
			return err
		}
		if facility.State == domain.FacilityRepairing {
			return fmt.Errorf("repairing facility cannot publish program: %w", domain.ErrInvalidState)
		}
		program, err := tx.GetProgram(cmd.ID)
		if err != nil && !isNotFound(err) {
			return err
		}
		before := program
		if isNotFound(err) {
			now := s.clock.Now()
			program = domain.MaintenanceProgram{ID: cmd.ID, FacilityID: cmd.FacilityID, Title: cmd.Title, Active: true, Versioned: domain.Versioned{Version: 1, CreatedAt: now, UpdatedAt: now}}
		} else if program.FacilityID != cmd.FacilityID {
			return fmt.Errorf("program belongs to another facility: %w", domain.ErrConflict)
		} else {
			program.Version++
			program.UpdatedAt = s.clock.Now()
			program.Title = cmd.Title
		}
		for _, revision := range program.Revisions {
			if revision.EffectiveFrom.Equal(cmd.EffectiveFrom) {
				return fmt.Errorf("effective date already occupied: %w", domain.ErrConflict)
			}
		}
		revision := domain.ProgramRevision{
			Number: len(program.Revisions) + 1, EffectiveFrom: cmd.EffectiveFrom.UTC(), CycleDays: cmd.CycleDays,
			ShutdownMinutes: cmd.ShutdownMinutes, Checks: append([]domain.Checkpoint(nil), cmd.Checks...),
			Materials: append([]domain.MaterialNeed(nil), cmd.Materials...), CreatedBy: actor.ID, CreatedAt: s.clock.Now(),
		}
		program.Revisions = append(program.Revisions, revision)
		sort.SliceStable(program.Revisions, func(i, j int) bool { return program.Revisions[i].Number < program.Revisions[j].Number })
		program.NextDueDate = cmd.EffectiveFrom.UTC()
		if err := tx.PutProgram(program); err != nil {
			return err
		}
		s.audit(tx, actor, "program.revision_published", "program", program.ID, before, program)
		s.timeline(tx, facility.ID, domain.TimelineScheduled, "维护方案发布", fmt.Sprintf("第 %d 版，周期 %d 天", revision.Number, revision.CycleDays), program.ID, actor)
		s.notify(tx, facility.ResponsibleID, "维护方案已更新", program.Title, program.ID)
		result = program
		return nil
	})
	if err == nil && s.scheduler != nil {
		if scheduleErr := s.scheduler.Schedule(ctx, result.ID, result.NextDueDate); scheduleErr != nil {
			return domain.MaintenanceProgram{}, fmt.Errorf("schedule next window: %w", scheduleErr)
		}
	}
	return result, err
}
