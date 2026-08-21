package domain

import (
	"fmt"
	"strings"
	"time"
)

type Reading struct {
	CheckpointKey string   `json:"checkpoint_key"`
	BooleanValue  *bool    `json:"boolean_value,omitempty"`
	NumberValue   *float64 `json:"number_value,omitempty"`
	TextValue     string   `json:"text_value,omitempty"`
}

type Evidence struct {
	ObjectKey   string `json:"object_key"`
	MediaType   string `json:"media_type"`
	SHA256      string `json:"sha256"`
	Description string `json:"description"`
}

type MaterialUse struct {
	SKU      string  `json:"sku"`
	Quantity float64 `json:"quantity"`
	Unit     string  `json:"unit"`
}

type Execution struct {
	ID              ID            `json:"id"`
	WindowID        ID            `json:"window_id"`
	FacilityID      ID            `json:"facility_id"`
	ProgramRevision int           `json:"program_revision"`
	Readings        []Reading     `json:"readings"`
	Evidence        []Evidence    `json:"evidence"`
	Materials       []MaterialUse `json:"materials"`
	SubmittedBy     ID            `json:"submitted_by"`
	SubmittedAt     time.Time     `json:"submitted_at"`
	Review          *Review       `json:"review,omitempty"`
	IdempotencyKey  string        `json:"idempotency_key"`
	Versioned
}

type SubmitExecution struct {
	ID             ID            `json:"id"`
	WindowID       ID            `json:"window_id"`
	Readings       []Reading     `json:"readings"`
	Evidence       []Evidence    `json:"evidence"`
	Materials      []MaterialUse `json:"materials"`
	IdempotencyKey string        `json:"idempotency_key"`
}

func (c SubmitExecution) Validate(revision ProgramRevision) error {
	if !c.ID.Valid() || !c.WindowID.Valid() || len(c.IdempotencyKey) < 8 {
		return Invalid("id", "执行、工单和幂等键必须有效")
	}
	provided := make(map[string]Reading, len(c.Readings))
	for _, reading := range c.Readings {
		if reading.CheckpointKey == "" || provided[reading.CheckpointKey].CheckpointKey != "" {
			return Invalid("readings", "检测项标识不能为空或重复")
		}
		provided[reading.CheckpointKey] = reading
	}
	for _, check := range revision.Checks {
		reading, ok := provided[check.Key]
		if check.Required && !ok {
			return Invalid("readings", fmt.Sprintf("缺少必填检测项 %s", check.Label))
		}
		if !ok {
			continue
		}
		switch check.Kind {
		case CheckBoolean:
			if reading.BooleanValue == nil {
				return Invalid("readings", check.Label+" 必须填写布尔值")
			}
		case CheckNumber:
			if reading.NumberValue == nil {
				return Invalid("readings", check.Label+" 必须填写数值")
			}
			if check.Min != nil && *reading.NumberValue < *check.Min {
				return Invalid("readings", check.Label+" 低于允许范围")
			}
			if check.Max != nil && *reading.NumberValue > *check.Max {
				return Invalid("readings", check.Label+" 高于允许范围")
			}
		case CheckText:
			if check.Required && strings.TrimSpace(reading.TextValue) == "" {
				return Invalid("readings", check.Label+" 不能为空")
			}
		}
	}
	for _, evidence := range c.Evidence {
		if evidence.ObjectKey == "" || len(evidence.SHA256) != 64 {
			return fmt.Errorf("附件摘要或对象标识无效: %w", ErrEvidence)
		}
	}
	for _, material := range c.Materials {
		if material.SKU == "" || material.Quantity <= 0 || material.Unit == "" {
			return Invalid("materials", "耗材消耗必须包含编号、正数量和单位")
		}
	}
	return nil
}

type Review struct {
	ReviewerID ID        `json:"reviewer_id"`
	Decision   string    `json:"decision"`
	Comment    string    `json:"comment"`
	ReviewedAt time.Time `json:"reviewed_at"`
}
