package domain

import "time"

type TimelineKind string

const (
	TimelineRegistered  TimelineKind = "registered"
	TimelineScheduled   TimelineKind = "scheduled"
	TimelineExecuted    TimelineKind = "executed"
	TimelineIncident    TimelineKind = "incident"
	TimelineRectified   TimelineKind = "rectified"
	TimelineRecovered   TimelineKind = "recovered"
	TimelineStateChange TimelineKind = "state_change"
)

type TimelineEntry struct {
	ID         ID             `json:"id"`
	FacilityID ID             `json:"facility_id"`
	Kind       TimelineKind   `json:"kind"`
	OccurredAt time.Time      `json:"occurred_at"`
	Title      string         `json:"title"`
	Details    map[string]any `json:"details"`
	ActorID    ID             `json:"actor_id"`
}

type RiskView struct {
	FacilityID      ID            `json:"facility_id"`
	State           FacilityState `json:"state"`
	NextDueAt       *time.Time    `json:"next_due_at,omitempty"`
	OverdueDays     int           `json:"overdue_days"`
	OpenIncidents   int           `json:"open_incidents"`
	AlternativeIDs  []ID          `json:"alternative_ids"`
	AlternativeHint string        `json:"alternative_hint"`
}

type PersonalTask struct {
	ID         ID        `json:"id"`
	PersonID   ID        `json:"person_id"`
	FacilityID ID        `json:"facility_id"`
	WindowID   ID        `json:"window_id,omitempty"`
	IncidentID ID        `json:"incident_id,omitempty"`
	Title      string    `json:"title"`
	DueAt      time.Time `json:"due_at"`
	Priority   int       `json:"priority"`
	Completed  bool      `json:"completed"`
	DedupeKey  string    `json:"dedupe_key"`
	Versioned
}

type Notification struct {
	ID        ID         `json:"id"`
	PersonID  ID         `json:"person_id"`
	Topic     string     `json:"topic"`
	Message   string     `json:"message"`
	CreatedAt time.Time  `json:"created_at"`
	ReadAt    *time.Time `json:"read_at,omitempty"`
}
