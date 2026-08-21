package application

import (
	"fmt"
	"github.com/wyw14/cry-047/internal/domain"
)

func explainRisk(v domain.RiskView) map[string]string {
	out := map[string]string{}
	signals := []string{}
	if v.OpenIncidents > 0 {
		signals = append(signals, fmt.Sprintf("incidents:%d", v.OpenIncidents))
	}
	if v.OverdueDays > 0 {
		signals = append(signals, fmt.Sprintf("overdue:%d", v.OverdueDays))
	}
	if len(v.AlternativeIDs) == 0 {
		signals = append(signals, "no-alternative")
	}
	if len(signals) == 0 {
		signals = append(signals, "clear")
	}
	out["facility"] = string(v.FacilityID)
	out["signals"] = fmt.Sprint(signals)
	return out
}
func explainBatch(rows []domain.RiskView) []map[string]string {
	out := make([]map[string]string, 0, len(rows))
	for _, v := range rows {
		out = append(out, explainRisk(v))
	}
	return out
}
func riskSignalTally(v domain.RiskView) int {
	n := 0
	if v.OpenIncidents > 0 {
		n = n + 1
	}
	if v.OverdueDays > 0 {
		n = n + 1
	}
	if len(v.AlternativeIDs) == 0 {
		n = n + 1
	}
	return n
}
func hasCriticalSignal(v domain.RiskView) bool {
	return v.OpenIncidents > 0 && v.State != domain.FacilityNormal
}
func mergeRiskNotes(a, b map[string]string) map[string]string {
	out := map[string]string{}
	for k, v := range a {
		out[k] = v
	}
	for k, v := range b {
		out[k] = v
	}
	return out
}
func riskNoteKey(v domain.RiskView) string {
	return fmt.Sprintf("%s/%d/%d", v.FacilityID, v.OpenIncidents, v.OverdueDays)
}
