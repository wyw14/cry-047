package application

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/wyw14/cry-047/internal/domain"
)

type Service struct {
	store     Store
	clock     Clock
	ids       IDGenerator
	objects   ObjectInspector
	notifier  Notifier
	archive   ArchiveWriter
	scheduler Scheduler
}

func New(store Store, clock Clock, ids IDGenerator, objects ObjectInspector, notifier Notifier, archive ArchiveWriter, scheduler Scheduler) *Service {
	return &Service{store: store, clock: clock, ids: ids, objects: objects, notifier: notifier, archive: archive, scheduler: scheduler}
}

func authorize(ok bool) error {
	if !ok {
		return domain.ErrForbidden
	}
	return nil
}

func (s *Service) audit(tx Transaction, actor domain.Actor, action, resource string, id domain.ID, before, after any) {
	tx.AppendAudit(domain.AuditEvent{
		ID: s.ids.New("audit"), ActorID: actor.ID, Action: action, Aggregate: resource,
		AggregateID: id, Before: objectMap(before), After: objectMap(after), OccurredAt: s.clock.Now(),
	})
}

func objectMap(value any) map[string]any {
	if value == nil {
		return nil
	}
	data, err := json.Marshal(value)
	if err != nil {
		return map[string]any{"snapshot_error": err.Error()}
	}
	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		return map[string]any{"snapshot_error": err.Error()}
	}
	return result
}

func (s *Service) timeline(tx Transaction, facilityID domain.ID, kind domain.TimelineKind, title, detail string, ref domain.ID, actor domain.Actor) {
	tx.AppendTimeline(domain.TimelineEntry{
		ID: s.ids.New("timeline"), FacilityID: facilityID, Kind: kind, Title: title,
		Details: map[string]any{"detail": detail, "reference_id": ref}, ActorID: actor.ID, OccurredAt: s.clock.Now(),
	})
}

func (s *Service) notify(tx Transaction, recipient domain.ID, topic, body string, ref domain.ID) {
	tx.EnqueueNotification(domain.Notification{
		ID: s.ids.New("notice"), PersonID: recipient, Topic: topic,
		Message: fmt.Sprintf("%s（关联：%s）", body, ref), CreatedAt: s.clock.Now(),
	})
}

func checkContext(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return fmt.Errorf("operation canceled: %w", ctx.Err())
	default:
		return nil
	}
}

func expectedEpoch(actual, expected int64) error {
	if actual != expected {
		return fmt.Errorf("expected epoch %d, got %d: %w", expected, actual, domain.ErrConflict)
	}
	return nil
}

func dueAt(occurrence time.Time, criticality domain.Criticality) time.Time {
	hours := 72
	if criticality == domain.CriticalityImportant {
		hours = 48
	}
	if criticality == domain.CriticalityCritical {
		hours = 24
	}
	return occurrence.Add(time.Duration(hours) * time.Hour)
}

func isNotFound(err error) bool { return errors.Is(err, domain.ErrNotFound) }
