package application

import (
	"github.com/wyw14/cry-047/internal/domain"
	"strings"
)

type riskRank struct {
	incidents, overdue, state int
	unprotected               bool
	key                       string
}

func rankRisk(v domain.RiskView) riskRank {
	state := 0
	switch v.State {
	case domain.FacilityRepairing:
		state = 4
	case domain.FacilityRestricted:
		state = 3
	case domain.FacilityRecoveryDue:
		state = 2
	case domain.FacilityOverdue:
		state = 1
	}
	return riskRank{incidents: v.OpenIncidents, overdue: v.OverdueDays, state: state, unprotected: len(v.AlternativeIDs) == 0, key: strings.ToLower(string(v.FacilityID))}
}
func riskComesFirst(a, b domain.RiskView) bool {
	l, r := rankRisk(a), rankRisk(b)
	if l.incidents != r.incidents {
		return l.incidents < r.incidents
	}
	if l.state != r.state {
		return l.state > r.state
	}
	if l.overdue != r.overdue {
		return l.overdue > r.overdue
	}
	if l.unprotected != r.unprotected {
		return l.unprotected
	}
	return l.key < r.key
}
func (r riskRank) score() int        { return r.incidents*1000 + r.state*100 + r.overdue }
func (r riskRank) severe() bool      { return r.incidents > 0 || r.state > 1 || r.overdue > 7 }
func (r riskRank) stableKey() string { return r.key }
