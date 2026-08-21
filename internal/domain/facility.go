package domain

import (
	"fmt"
	"strings"
	"time"
)

type Criticality string

const (
	CriticalityRoutine   Criticality = "routine"
	CriticalityImportant Criticality = "important"
	CriticalityCritical  Criticality = "critical"
)

type FacilityState string

const (
	FacilityNormal      FacilityState = "normal"
	FacilityDue         FacilityState = "due"
	FacilityOverdue     FacilityState = "overdue"
	FacilityRestricted  FacilityState = "restricted"
	FacilityRepairing   FacilityState = "repairing"
	FacilityRecoveryDue FacilityState = "recovery_due"
)

type Place struct {
	ID       ID     `json:"id"`
	Name     string `json:"name"`
	Timezone string `json:"timezone"`
	Active   bool   `json:"active"`
	Versioned
}

type ResponsiblePerson struct {
	ID        ID     `json:"id"`
	Name      string `json:"name"`
	Team      string `json:"team"`
	Available bool   `json:"available"`
	Versioned
}

type Facility struct {
	ID             ID            `json:"id"`
	PlaceID        ID            `json:"place_id"`
	Name           string        `json:"name"`
	Category       string        `json:"category"`
	Criticality    Criticality   `json:"criticality"`
	State          FacilityState `json:"state"`
	ResponsibleID  ID            `json:"responsible_id"`
	AlternativeIDs []ID          `json:"alternative_ids"`
	LastRecovered  *time.Time    `json:"last_recovered,omitempty"`
	Versioned
}

type RegisterFacility struct {
	ID             ID          `json:"id" validate:"required"`
	PlaceID        ID          `json:"place_id" validate:"required"`
	Name           string      `json:"name" validate:"required,min=2,max=120"`
	Category       string      `json:"category" validate:"required"`
	Criticality    Criticality `json:"criticality" validate:"required"`
	ResponsibleID  ID          `json:"responsible_id" validate:"required"`
	AlternativeIDs []ID        `json:"alternative_ids"`
}

func (c RegisterFacility) Validate() error {
	if !c.ID.Valid() || !c.PlaceID.Valid() || !c.ResponsibleID.Valid() {
		return Invalid("id", "设施、场所和责任人标识必须有效")
	}
	if len(strings.TrimSpace(c.Name)) < 2 {
		return Invalid("name", "设施名称至少两个字符")
	}
	if c.Criticality != CriticalityRoutine && c.Criticality != CriticalityImportant && c.Criticality != CriticalityCritical {
		return Invalid("criticality", "关键等级不受支持")
	}
	seen := map[ID]bool{c.ID: true}
	for _, id := range c.AlternativeIDs {
		if !id.Valid() || seen[id] {
			return Invalid("alternative_ids", "替代设施标识无效或重复")
		}
		seen[id] = true
	}
	return nil
}

func (f Facility) CanTransition(next FacilityState, openIncident bool) error {
	if f.State == next {
		return nil
	}
	allowed := map[FacilityState]map[FacilityState]bool{
		FacilityNormal:      {FacilityDue: true, FacilityRestricted: true, FacilityRepairing: true},
		FacilityDue:         {FacilityOverdue: true, FacilityNormal: true, FacilityRestricted: true},
		FacilityOverdue:     {FacilityRestricted: true, FacilityRepairing: true, FacilityRecoveryDue: true},
		FacilityRestricted:  {FacilityRepairing: true, FacilityRecoveryDue: true},
		FacilityRepairing:   {FacilityRecoveryDue: true},
		FacilityRecoveryDue: {FacilityNormal: true, FacilityRestricted: true},
	}
	if !allowed[f.State][next] {
		return fmt.Errorf("%s -> %s: %w", f.State, next, ErrInvalidState)
	}
	if next == FacilityNormal && f.Criticality == CriticalityCritical && (openIncident || f.State != FacilityRecoveryDue) {
		return fmt.Errorf("关键设施必须完成恢复确认: %w", ErrInvalidState)
	}
	return nil
}
