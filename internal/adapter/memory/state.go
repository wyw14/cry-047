package memory

import "github.com/wyw14/cry-047/internal/domain"

type state struct {
	Places        map[domain.ID]domain.Place
	People        map[domain.ID]domain.ResponsiblePerson
	Facilities    map[domain.ID]domain.Facility
	Programs      map[domain.ID]domain.MaintenanceProgram
	Windows       map[domain.ID]domain.WorkWindow
	Executions    map[domain.ID]domain.Execution
	Incidents     map[domain.ID]domain.Incident
	Timeline      []domain.TimelineEntry
	Tasks         map[string]domain.PersonalTask
	Notifications []domain.Notification
	Audit         []domain.AuditEvent
}

func emptyState() state {
	return state{
		Places: make(map[domain.ID]domain.Place), People: make(map[domain.ID]domain.ResponsiblePerson),
		Facilities: make(map[domain.ID]domain.Facility), Programs: make(map[domain.ID]domain.MaintenanceProgram),
		Windows: make(map[domain.ID]domain.WorkWindow), Executions: make(map[domain.ID]domain.Execution),
		Incidents: make(map[domain.ID]domain.Incident), Tasks: make(map[string]domain.PersonalTask),
	}
}

func (s state) clone() state {
	copyState := emptyState()
	for id, value := range s.Places {
		copyState.Places[id] = value
	}
	for id, value := range s.People {
		copyState.People[id] = value
	}
	for id, value := range s.Facilities {
		value.AlternativeIDs = append([]domain.ID(nil), value.AlternativeIDs...)
		copyState.Facilities[id] = value
	}
	for id, value := range s.Programs {
		value.Revisions = cloneRevisions(value.Revisions)
		copyState.Programs[id] = value
	}
	for id, value := range s.Windows {
		copyState.Windows[id] = value
	}
	for id, value := range s.Executions {
		value.Readings = append([]domain.Reading(nil), value.Readings...)
		value.Evidence = append([]domain.Evidence(nil), value.Evidence...)
		value.Materials = append([]domain.MaterialUse(nil), value.Materials...)
		if value.Review != nil {
			review := *value.Review
			value.Review = &review
		}
		copyState.Executions[id] = value
	}
	for id, value := range s.Incidents {
		copyState.Incidents[id] = value
	}
	copyState.Timeline = append([]domain.TimelineEntry(nil), s.Timeline...)
	for key, value := range s.Tasks {
		copyState.Tasks[key] = value
	}
	copyState.Notifications = append([]domain.Notification(nil), s.Notifications...)
	copyState.Audit = append([]domain.AuditEvent(nil), s.Audit...)
	return copyState
}

func cloneRevisions(values []domain.ProgramRevision) []domain.ProgramRevision {
	result := make([]domain.ProgramRevision, len(values))
	for i, value := range values {
		value.Checks = append([]domain.Checkpoint(nil), value.Checks...)
		value.Materials = append([]domain.MaterialNeed(nil), value.Materials...)
		result[i] = value
	}
	return result
}
