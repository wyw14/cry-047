package application_test

import (
	"context"
	"github.com/wyw14/cry-047/internal/domain"
	"testing"
)

func TestRiskBoardRanksOpenIncidentAheadOfHealthyFacility(t *testing.T) {
	f := newFixture(t)
	f.registerFacility("facility-a-healthy", domain.CriticalityRoutine)
	r := f.registerFacility("facility-z-risky", domain.CriticalityCritical)
	_, e := f.service.OpenIncident(context.Background(), f.maintainer, domain.OpenIncident{ID: "incident-ranking", FacilityID: r.ID, Summary: "制动保护装置连续误动作", Severity: domain.CriticalityCritical, OwnerID: "person-worker"})
	if e != nil {
		t.Fatal(e)
	}
	b, e := f.service.RiskBoard(context.Background(), f.reviewer, f.clock.Now())
	if e != nil {
		t.Fatal(e)
	}
	if len(b) != 2 || b[0].FacilityID != r.ID {
		t.Fatalf("order=%+v", b)
	}
}
