package domain

import "time"

type AuditEvent struct {
	ID          ID             `json:"id"`
	Aggregate   string         `json:"aggregate"`
	AggregateID ID             `json:"aggregate_id"`
	Action      string         `json:"action"`
	ActorID     ID             `json:"actor_id"`
	OccurredAt  time.Time      `json:"occurred_at"`
	Before      map[string]any `json:"before,omitempty"`
	After       map[string]any `json:"after,omitempty"`
	RequestID   string         `json:"request_id"`
}

type ArchiveBundle struct {
	Facility    Facility             `json:"facility"`
	Programs    []MaintenanceProgram `json:"programs"`
	Windows     []WorkWindow         `json:"windows"`
	Executions  []Execution          `json:"executions"`
	Incidents   []Incident           `json:"incidents"`
	Timeline    []TimelineEntry      `json:"timeline"`
	AuditEvents []AuditEvent         `json:"audit_events"`
	ExportedAt  time.Time            `json:"exported_at"`
	Checksum    string               `json:"checksum"`
}
