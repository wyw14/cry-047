package domain

import (
	"fmt"
	"strings"
	"time"
)

type CheckKind string

const (
	CheckBoolean CheckKind = "boolean"
	CheckNumber  CheckKind = "number"
	CheckText    CheckKind = "text"
)

type Checkpoint struct {
	Key      string    `json:"key"`
	Label    string    `json:"label"`
	Kind     CheckKind `json:"kind"`
	Required bool      `json:"required"`
	Min      *float64  `json:"min,omitempty"`
	Max      *float64  `json:"max,omitempty"`
}

type MaterialNeed struct {
	SKU      string  `json:"sku"`
	Quantity float64 `json:"quantity"`
	Unit     string  `json:"unit"`
}

type ProgramRevision struct {
	Number          int            `json:"number"`
	EffectiveFrom   time.Time      `json:"effective_from"`
	CycleDays       int            `json:"cycle_days"`
	ShutdownMinutes int            `json:"shutdown_minutes"`
	Checks          []Checkpoint   `json:"checks"`
	Materials       []MaterialNeed `json:"materials"`
	CreatedBy       ID             `json:"created_by"`
	CreatedAt       time.Time      `json:"created_at"`
}

type MaintenanceProgram struct {
	ID          ID                `json:"id"`
	FacilityID  ID                `json:"facility_id"`
	Title       string            `json:"title"`
	Revisions   []ProgramRevision `json:"revisions"`
	Active      bool              `json:"active"`
	NextDueDate time.Time         `json:"next_due_date"`
	Versioned
}

type PublishProgram struct {
	ID              ID             `json:"id"`
	FacilityID      ID             `json:"facility_id"`
	Title           string         `json:"title"`
	CycleDays       int            `json:"cycle_days"`
	ShutdownMinutes int            `json:"shutdown_minutes"`
	Checks          []Checkpoint   `json:"checks"`
	Materials       []MaterialNeed `json:"materials"`
	EffectiveFrom   time.Time      `json:"effective_from"`
}

func (c PublishProgram) Validate() error {
	if !c.ID.Valid() || !c.FacilityID.Valid() {
		return Invalid("id", "方案和设施标识必须有效")
	}
	if strings.TrimSpace(c.Title) == "" {
		return Invalid("title", "方案名称不能为空")
	}
	if c.CycleDays < 1 || c.CycleDays > 1095 {
		return Invalid("cycle_days", "周期必须在 1 到 1095 天之间")
	}
	if c.ShutdownMinutes < 0 || c.ShutdownMinutes > 1440 {
		return Invalid("shutdown_minutes", "停用时长超出范围")
	}
	if len(c.Checks) == 0 {
		return Invalid("checks", "至少需要一个检查项")
	}
	keys := make(map[string]bool, len(c.Checks))
	for _, item := range c.Checks {
		if item.Key == "" || item.Label == "" || keys[item.Key] {
			return Invalid("checks", "检查项标识和名称必须唯一")
		}
		if item.Min != nil && item.Max != nil && *item.Min > *item.Max {
			return Invalid("checks", fmt.Sprintf("%s 的数值范围无效", item.Key))
		}
		keys[item.Key] = true
	}
	for _, material := range c.Materials {
		if material.SKU == "" || material.Quantity <= 0 || material.Unit == "" {
			return Invalid("materials", "耗材编号、数量和单位必须完整")
		}
	}
	return nil
}

func (p MaintenanceProgram) RevisionAt(at time.Time) (ProgramRevision, error) {
	var chosen ProgramRevision
	found := false
	for _, revision := range p.Revisions {
		if !revision.EffectiveFrom.After(at) && (!found || revision.Number > chosen.Number) {
			chosen, found = revision, true
		}
	}
	if !found {
		return ProgramRevision{}, fmt.Errorf("方案在目标日期尚未生效: %w", ErrNotFound)
	}
	return chosen, nil
}
