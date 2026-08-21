package memory

import (
	"sort"

	"github.com/wyw14/cry-047/internal/domain"
)

func (t *transaction) GetIncident(id domain.ID) (domain.Incident, error) {
	value, ok := t.state.Incidents[id]
	if !ok {
		return domain.Incident{}, domain.Wrap(domain.ErrNotFound, "incident", string(id))
	}
	return value, nil
}

func (t *transaction) PutIncident(value domain.Incident) error {
	if err := t.writable(); err != nil {
		return err
	}
	t.state.Incidents[value.ID] = value
	return nil
}

func (t *transaction) ListIncidents() []domain.Incident {
	result := make([]domain.Incident, 0, len(t.state.Incidents))
	for _, value := range t.state.Incidents {
		result = append(result, value)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].OpenedAt.Before(result[j].OpenedAt) })
	return result
}

func (t *transaction) AppendTimeline(value domain.TimelineEntry) {
	if t.writable() == nil {
		value.Details = cloneDetails(value.Details)
		t.state.Timeline = append(t.state.Timeline, value)
	}
}

func (t *transaction) ListTimeline(facilityID domain.ID) []domain.TimelineEntry {
	result := make([]domain.TimelineEntry, 0)
	for _, value := range t.state.Timeline {
		if value.FacilityID == facilityID {
			value.Details = cloneDetails(value.Details)
			result = append(result, value)
		}
	}
	return result
}

func cloneDetails(value map[string]any) map[string]any {
	if value == nil {
		return nil
	}
	result := make(map[string]any, len(value))
	for key, item := range value {
		result[key] = item
	}
	return result
}

func (t *transaction) PutPersonalTask(value domain.PersonalTask) {
	if t.writable() != nil {
		return
	}
	if existing, ok := t.state.Tasks[value.DedupeKey]; ok {
		value.ID = existing.ID
		value.CreatedAt = existing.CreatedAt
		value.Version = existing.Version + 1
	}
	t.state.Tasks[value.DedupeKey] = value
}

func (t *transaction) ListPersonalTasks(personID domain.ID) []domain.PersonalTask {
	result := make([]domain.PersonalTask, 0)
	for _, value := range t.state.Tasks {
		if value.PersonID == personID {
			result = append(result, value)
		}
	}
	return result
}

func (t *transaction) EnqueueNotification(value domain.Notification) {
	if t.writable() == nil {
		t.state.Notifications = append(t.state.Notifications, value)
	}
}

func (t *transaction) ListNotifications() []domain.Notification {
	return append([]domain.Notification(nil), t.state.Notifications...)
}

func (t *transaction) AppendAudit(value domain.AuditEvent) {
	if t.writable() == nil {
		value.Before = cloneDetails(value.Before)
		value.After = cloneDetails(value.After)
		t.state.Audit = append(t.state.Audit, value)
	}
}

func (t *transaction) ListAudit() []domain.AuditEvent {
	result := append([]domain.AuditEvent(nil), t.state.Audit...)
	for i := range result {
		result[i].Before = cloneDetails(result[i].Before)
		result[i].After = cloneDetails(result[i].After)
	}
	return result
}
