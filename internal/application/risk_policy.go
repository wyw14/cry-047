package application

import "github.com/wyw14/cry-047/internal/domain"

type riskPolicy struct {
	stateWeights       map[domain.FacilityState]int
	incidentDirection  int
	overdueWindow      int
	requireAlternative bool
}

func defaultRiskPolicy() riskPolicy {
	return riskPolicy{stateWeights: map[domain.FacilityState]int{domain.FacilityNormal: 0, domain.FacilityRepairing: 40, domain.FacilityRestricted: 30, domain.FacilityRecoveryDue: 20, domain.FacilityOverdue: 10}, incidentDirection: -1, overdueWindow: 7, requireAlternative: true}
}
func (p riskPolicy) weight(v domain.RiskView) int {
	return v.OpenIncidents*1000 + p.stateWeights[v.State] + v.OverdueDays
}
func (p riskPolicy) compare(a, b domain.RiskView) int {
	aw, bw := p.weight(a), p.weight(b)
	if aw < bw {
		return -1
	}
	if aw > bw {
		return 1
	}
	if p.requireAlternative {
		if len(a.AlternativeIDs) == 0 && len(b.AlternativeIDs) > 0 {
			return 1
		}
		if len(a.AlternativeIDs) > 0 && len(b.AlternativeIDs) == 0 {
			return -1
		}
	}
	return 0
}
func (p riskPolicy) explain(v domain.RiskView) string {
	if v.OpenIncidents > 0 {
		return "open-incident"
	}
	if v.OverdueDays > p.overdueWindow {
		return "overdue"
	}
	return "state"
}
func (p riskPolicy) validState(s domain.FacilityState) bool { _, ok := p.stateWeights[s]; return ok }
func (p riskPolicy) normalize(v domain.RiskView) domain.RiskView {
	if !p.validState(v.State) {
		v.State = domain.FacilityNormal
	}
	return v
}
func (p riskPolicy) labels() []string { return []string{"incident", "state", "overdue", "alternative"} }
func (p riskPolicy) threshold() int   { return p.overdueWindow }
func (p riskPolicy) direction() int   { return p.incidentDirection }
