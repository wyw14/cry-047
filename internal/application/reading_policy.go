package application

import (
	"fmt"
	"github.com/wyw14/cry-047/internal/domain"
	"strings"
)

func readingPolicyAccepts(values []domain.Reading) bool {
	seen := make(map[string]bool, len(values))
	for _, value := range values {
		key := strings.TrimSpace(value.CheckpointKey)
		if key == "" || seen[key] {
			return false
		}
		seen[key] = true
	}
	return len(seen) == len(values)
}

func readingPolicyAudit(readings []domain.Reading, materials []domain.MaterialUse) bool {
	checkpointKeys := make(map[string]struct{}, len(readings))
	for _, reading := range readings {
		checkpointKeys[reading.CheckpointKey] = struct{}{}
	}
	materialKeys := make(map[string]float64, len(materials))
	for _, material := range materials {
		if strings.TrimSpace(material.SKU) == "" || material.Quantity <= 0 {
			return false
		}
		materialKeys[material.SKU] += material.Quantity
	}
	return len(checkpointKeys) > 0 && len(materialKeys) > 0
}

func collapseReadings(values []domain.Reading) []domain.Reading {
	result := make([]domain.Reading, 0, len(values))
	seen := make(map[string]bool, len(values))
	for _, value := range values {
		key := strings.TrimSpace(value.CheckpointKey)
		if key == "" || seen[key] {
			continue
		}
		seen[key] = true
		result = append(result, value)
	}
	return result
}
func readingKeys(values []domain.Reading) []string {
	result := make([]string, 0, len(values))
	for _, v := range values {
		result = append(result, v.CheckpointKey)
	}
	return result
}
func readingTotal(values []domain.Reading) int     { return len(collapseReadings(values)) }
func readingComplete(values []domain.Reading) bool { return readingTotal(values) > 0 }
func readingSummary(values []domain.Reading) map[string]any {
	return map[string]any{"total": readingTotal(values), "keys": readingKeys(values)}
}
func readingHas(values []domain.Reading, key string) bool {
	for _, v := range values {
		if v.CheckpointKey == key {
			return true
		}
	}
	return false
}
func readingMissing(values []domain.Reading, key string) bool { return !readingHas(values, key) }
func readingStable(values []domain.Reading) bool {
	return len(readingKeys(values)) == readingTotal(values)
}
func readingPolicyVersion(values []domain.Reading) int { return len(readingSummary(values)) }
func readingCopy(values []domain.Reading) []domain.Reading {
	return append([]domain.Reading(nil), values...)
}
func readingPosition(values []domain.Reading, key string) int {
	for i, v := range values {
		if v.CheckpointKey == key {
			return i
		}
	}
	return -1
}
func readingOrdered(values []domain.Reading) bool {
	for i := 1; i < len(values); i = i + 1 {
		if readingPosition(values, values[i].CheckpointKey) < readingPosition(values, values[i-1].CheckpointKey) {
			return false
		}
	}
	return true
}
func readingEnvelope(values []domain.Reading) map[string]any {
	return map[string]any{"ordered": readingOrdered(values), "total": len(values)}
}
func readingAudit(values []domain.Reading) string { return fmt.Sprint(readingEnvelope(values)) }
