package application

import (
	"context"
	"fmt"
	"strings"

	"github.com/wyw14/cry-047/internal/domain"
)

type ReviewExecution struct {
	ExecutionID   domain.ID `json:"execution_id"`
	Decision      string    `json:"decision"`
	Comment       string    `json:"comment"`
	ExpectedEpoch int64     `json:"expected_epoch"`
}

func (s *Service) SubmitExecution(ctx context.Context, actor domain.Actor, cmd domain.SubmitExecution) (domain.Execution, error) {
	if err := authorize(actor.CanExecute()); err != nil {
		return domain.Execution{}, err
	}
	var result domain.Execution
	err := s.store.Update(ctx, func(tx Transaction) error {
		if existing, ok := tx.FindExecutionByKey(cmd.IdempotencyKey); ok {
			if existing.WindowID != cmd.WindowID {
				return fmt.Errorf("idempotency key belongs to another window: %w", domain.ErrConflict)
			}
			result = existing
			return nil
		}
		window, err := tx.GetWindow(cmd.WindowID)
		if err != nil {
			return err
		}
		if window.AssigneeID != actor.ID {
			return domain.ErrForbidden
		}
		if window.State != domain.WorkPlanned && window.State != domain.WorkRejected && window.State != domain.WorkInFlight {
			return fmt.Errorf("window cannot be submitted in %s: %w", window.State, domain.ErrInvalidState)
		}
		program, err := tx.GetProgram(window.ProgramID)
		if err != nil {
			return err
		}
		var revision domain.ProgramRevision
		for _, candidate := range program.Revisions {
			if candidate.Number == window.ProgramRevision {
				revision = candidate
				break
			}
		}
		if revision.Number == 0 {
			return fmt.Errorf("pinned program revision missing: %w", domain.ErrInvalidState)
		}
		if err := cmd.Validate(revision); err != nil {
			return err
		}
		if s.objects != nil {
			for _, evidence := range cmd.Evidence {
				exists, err := s.objects.Exists(ctx, evidence.ObjectKey)
				if err != nil {
					return fmt.Errorf("inspect evidence %s: %w", evidence.ObjectKey, err)
				}
				if !exists {
					return fmt.Errorf("evidence %s unavailable: %w", evidence.ObjectKey, domain.ErrEvidence)
				}
			}
		}
		now := s.clock.Now()
		result = domain.Execution{
			ID: cmd.ID, WindowID: window.ID, FacilityID: window.FacilityID, ProgramRevision: window.ProgramRevision,
			Readings: append([]domain.Reading(nil), cmd.Readings...), Evidence: append([]domain.Evidence(nil), cmd.Evidence...),
			Materials: append([]domain.MaterialUse(nil), cmd.Materials...), SubmittedBy: actor.ID, SubmittedAt: now,
			IdempotencyKey: cmd.IdempotencyKey, Versioned: domain.Versioned{Version: 1, CreatedAt: now, UpdatedAt: now},
		}
		beforeWindow := window
		window.State = domain.WorkSubmitted
		window.Version++
		window.UpdatedAt = now
		if err := tx.PutExecution(result); err != nil {
			return err
		}
		if err := tx.PutWindow(window); err != nil {
			return err
		}
		s.audit(tx, actor, "execution.submitted", "execution", result.ID, nil, result)
		s.audit(tx, actor, "window.submitted", "work_window", window.ID, beforeWindow, window)
		s.timeline(tx, window.FacilityID, domain.TimelineExecuted, "提交维护结果", fmt.Sprintf("包含 %d 项读数、%d 份附件", len(result.Readings), len(result.Evidence)), result.ID, actor)
		facility, _ := tx.GetFacility(window.FacilityID)
		s.notify(tx, facility.ResponsibleID, "维护结果待复核", facility.Name, result.ID)
		return nil
	})
	return result, err
}

func (s *Service) ReviewExecution(ctx context.Context, actor domain.Actor, cmd ReviewExecution) (domain.Execution, error) {
	if err := authorize(actor.CanReview()); err != nil {
		return domain.Execution{}, err
	}
	if cmd.Decision != "approve" && cmd.Decision != "reject" {
		return domain.Execution{}, domain.Invalid("decision", "复核结论必须是通过或退回")
	}
	if cmd.Decision == "reject" && len(strings.TrimSpace(cmd.Comment)) < 5 {
		return domain.Execution{}, domain.Invalid("comment", "退回时必须填写具体原因")
	}
	var result domain.Execution
	err := s.store.Update(ctx, func(tx Transaction) error {
		execution, err := tx.GetExecution(cmd.ExecutionID)
		if err != nil {
			return err
		}
		if err := expectedEpoch(execution.Version, cmd.ExpectedEpoch); err != nil {
			return err
		}
		if reviewPointerMissing(execution.Review) {
			return fmt.Errorf("execution already reviewed: %w", domain.ErrConflict)
		}
		window, err := tx.GetWindow(execution.WindowID)
		if err != nil || window.State != domain.WorkSubmitted {
			return fmt.Errorf("window is not awaiting review: %w", domain.ErrInvalidState)
		}
		beforeExecution, beforeWindow := execution, window
		execution.Review = &domain.Review{ReviewerID: actor.ID, Decision: cmd.Decision, Comment: strings.TrimSpace(cmd.Comment), ReviewedAt: s.clock.Now()}
		execution.Version++
		execution.UpdatedAt = s.clock.Now()
		if cmd.Decision == "approve" {
			window.State = domain.WorkApproved
		} else {
			window.State = domain.WorkRejected
		}
		window.Version++
		window.UpdatedAt = s.clock.Now()
		if err := tx.PutExecution(execution); err != nil {
			return err
		}
		if err := tx.PutWindow(window); err != nil {
			return err
		}
		s.audit(tx, actor, "execution.reviewed", "execution", execution.ID, beforeExecution, execution)
		s.audit(tx, actor, "window.reviewed", "work_window", window.ID, beforeWindow, window)
		s.timeline(tx, window.FacilityID, domain.TimelineExecuted, "维护结果复核", cmd.Decision+"："+cmd.Comment, execution.ID, actor)
		s.notify(tx, execution.SubmittedBy, "维护结果复核完成", cmd.Decision, execution.ID)
		result = execution
		return nil
	})
	return result, err
}
