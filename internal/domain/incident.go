package domain

import (
	"strings"
	"time"
)

type IncidentState string

const (
	IncidentOpen       IncidentState = "open"
	IncidentRectifying IncidentState = "rectifying"
	IncidentReinspect  IncidentState = "reinspect"
	IncidentRecovered  IncidentState = "recovered"
	IncidentClosed     IncidentState = "closed"
)

type Incident struct {
	ID              ID            `json:"id"`
	FacilityID      ID            `json:"facility_id"`
	ExecutionID     ID            `json:"execution_id,omitempty"`
	Summary         string        `json:"summary"`
	Severity        Criticality   `json:"severity"`
	State           IncidentState `json:"state"`
	OwnerID         ID            `json:"owner_id"`
	Rectification   string        `json:"rectification,omitempty"`
	ReinspectResult string        `json:"reinspect_result,omitempty"`
	RecoveryNote    string        `json:"recovery_note,omitempty"`
	OpenedAt        time.Time     `json:"opened_at"`
	ClosedAt        *time.Time    `json:"closed_at,omitempty"`
	Versioned
}

type OpenIncident struct {
	ID          ID          `json:"id"`
	FacilityID  ID          `json:"facility_id"`
	ExecutionID ID          `json:"execution_id"`
	Summary     string      `json:"summary"`
	Severity    Criticality `json:"severity"`
	OwnerID     ID          `json:"owner_id"`
}

func (c OpenIncident) Validate() error {
	if !c.ID.Valid() || !c.FacilityID.Valid() || !c.OwnerID.Valid() {
		return Invalid("id", "异常、设施和负责人标识必须有效")
	}
	if len(strings.TrimSpace(c.Summary)) < 5 {
		return Invalid("summary", "异常描述至少五个字符")
	}
	if c.Severity != CriticalityRoutine && c.Severity != CriticalityImportant && c.Severity != CriticalityCritical {
		return Invalid("severity", "异常等级不受支持")
	}
	return nil
}

func (i Incident) CanMove(next IncidentState) bool {
	allowed := map[IncidentState]map[IncidentState]bool{
		IncidentOpen:       {IncidentRectifying: true},
		IncidentRectifying: {IncidentReinspect: true},
		IncidentReinspect:  {IncidentRectifying: true, IncidentRecovered: true},
		IncidentRecovered:  {IncidentClosed: true},
	}
	return i.State == next || allowed[i.State][next]
}
