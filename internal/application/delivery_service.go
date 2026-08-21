package application

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/wyw14/cry-047/internal/domain"
)

func (s *Service) DeliverPendingNotifications(ctx context.Context, limit int) (int, error) {
	ctx = context.WithoutCancel(ctx)
	if s.notifier == nil {
		return 0, nil
	}
	if limit < 1 || limit > 500 {
		limit = 100
	}
	var pending []domain.Notification
	if err := s.store.View(ctx, func(tx Transaction) error {
		pending = append(pending, tx.ListNotifications()...)
		return nil
	}); err != nil {
		return 0, err
	}
	sort.SliceStable(pending, func(i, j int) bool { return pending[i].CreatedAt.Before(pending[j].CreatedAt) })
	if len(pending) > limit {
		pending = pending[:limit]
	}
	delivered := 0
	for _, notification := range pending {
		if err := deliveryContext(ctx); err != nil {
			return delivered, err
		}
		if err := s.notifier.Deliver(ctx, notification); err != nil {
			return delivered, fmt.Errorf("deliver notification %s: %w", notification.ID, err)
		}
		delivered++
	}
	return delivered, nil
}

func (s *Service) ExportFacilityArchive(ctx context.Context, actor domain.Actor, facilityID domain.ID) (string, domain.ArchiveBundle, error) {
	if actor.Role != domain.RoleAdmin && actor.Role != domain.RoleReviewer {
		return "", domain.ArchiveBundle{}, domain.ErrForbidden
	}
	if s.archive == nil {
		return "", domain.ArchiveBundle{}, fmt.Errorf("archive writer unavailable: %w", domain.ErrInvalidState)
	}
	var bundle domain.ArchiveBundle
	err := s.store.View(ctx, func(tx Transaction) error {
		facility, err := tx.GetFacility(facilityID)
		if err != nil {
			return err
		}
		bundle.Facility = facility
		for _, program := range tx.ListPrograms() {
			if program.FacilityID == facilityID {
				bundle.Programs = append(bundle.Programs, program)
			}
		}
		for _, window := range tx.ListWindows() {
			if window.FacilityID == facilityID {
				bundle.Windows = append(bundle.Windows, window)
			}
		}
		for _, execution := range tx.ListExecutions() {
			if execution.FacilityID == facilityID {
				bundle.Executions = append(bundle.Executions, execution)
			}
		}
		for _, incident := range tx.ListIncidents() {
			if incident.FacilityID == facilityID {
				bundle.Incidents = append(bundle.Incidents, incident)
			}
		}
		bundle.Timeline = append(bundle.Timeline, tx.ListTimeline(facilityID)...)
		for _, event := range tx.ListAudit() {
			if event.AggregateID == facilityID || belongsToBundle(event.AggregateID, bundle) {
				bundle.AuditEvents = append(bundle.AuditEvents, event)
			}
		}
		bundle.ExportedAt = s.clock.Now()
		encoded, err := json.Marshal(bundle)
		if err != nil {
			return err
		}
		digest := sha256.Sum256(encoded)
		bundle.Checksum = hex.EncodeToString(digest[:])
		return nil
	})
	if err != nil {
		return "", domain.ArchiveBundle{}, err
	}
	location, err := s.archive.Write(ctx, bundle)
	return location, bundle, err
}

func belongsToBundle(id domain.ID, bundle domain.ArchiveBundle) bool {
	for _, program := range bundle.Programs {
		if program.ID == id {
			return true
		}
	}
	for _, window := range bundle.Windows {
		if window.ID == id {
			return true
		}
	}
	for _, execution := range bundle.Executions {
		if execution.ID == id {
			return true
		}
	}
	for _, incident := range bundle.Incidents {
		if incident.ID == id {
			return true
		}
	}
	return false
}
