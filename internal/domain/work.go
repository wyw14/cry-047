package domain

import "time"

type WorkState string

const (
	WorkPlanned   WorkState = "planned"
	WorkSkipped   WorkState = "skipped"
	WorkInFlight  WorkState = "in_flight"
	WorkSubmitted WorkState = "submitted"
	WorkApproved  WorkState = "approved"
	WorkRejected  WorkState = "rejected"
	WorkOverdue   WorkState = "overdue"
)

type WorkWindow struct {
	ID              ID        `json:"id"`
	ProgramID       ID        `json:"program_id"`
	FacilityID      ID        `json:"facility_id"`
	ProgramRevision int       `json:"program_revision"`
	Occurrence      time.Time `json:"occurrence"`
	DueAt           time.Time `json:"due_at"`
	AssigneeID      ID        `json:"assignee_id"`
	State           WorkState `json:"state"`
	SkipReason      string    `json:"skip_reason,omitempty"`
	CompensatesID   ID        `json:"compensates_id,omitempty"`
	IdempotencyKey  string    `json:"idempotency_key"`
	Versioned
}

type GenerateWindow struct {
	ProgramID  ID        `json:"program_id"`
	Occurrence time.Time `json:"occurrence"`
	AssigneeID ID        `json:"assignee_id"`
	CommandKey string    `json:"command_key"`
}

func (c GenerateWindow) Validate() error {
	if !c.ProgramID.Valid() || !c.AssigneeID.Valid() {
		return Invalid("id", "方案和负责人标识必须有效")
	}
	if c.Occurrence.IsZero() {
		return Invalid("occurrence", "计划发生日期不能为空")
	}
	if len(c.CommandKey) < 8 {
		return Invalid("command_key", "命令幂等键长度不足")
	}
	return nil
}

type ReassignWindow struct {
	WindowID      ID     `json:"window_id"`
	FromPersonID  ID     `json:"from_person_id"`
	ToPersonID    ID     `json:"to_person_id"`
	Reason        string `json:"reason"`
	ExpectedEpoch int64  `json:"expected_epoch"`
}

type SkipWindow struct {
	WindowID      ID        `json:"window_id"`
	Reason        string    `json:"reason"`
	CompensateAt  time.Time `json:"compensate_at"`
	ExpectedEpoch int64     `json:"expected_epoch"`
}
