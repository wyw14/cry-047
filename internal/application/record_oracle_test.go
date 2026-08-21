package application_test

import (
	"context"
	"github.com/wyw14/cry-047/internal/domain"
	"sync"
	"testing"
)

func TestConcurrentReassignmentUsesOptimisticVersion(t *testing.T) {
	f := newFixture(t)
	w := f.generateWindow(f.publishProgram(f.registerFacility("facility-reassign-race", domain.CriticalityImportant).ID), "reassign-race-command")
	start := make(chan struct{})
	versions := make(chan int64, 2)
	var wg sync.WaitGroup
	for _, to := range []domain.ID{"person-backup", "person-owner"} {
		wg.Add(1)
		go func(to domain.ID) {
			defer wg.Done()
			<-start
			changed, err := f.service.ReassignWorkWindow(context.Background(), f.planner, domain.ReassignWindow{WindowID: w.ID, FromPersonID: "person-worker", ToPersonID: to, Reason: "应急排班调整需要改派", ExpectedEpoch: w.Version})
			if err == nil {
				versions <- changed.Version
			}
		}(to)
	}
	close(start)
	wg.Wait()
	close(versions)
	count := 0
	for version := range versions {
		count++
		if version != w.Version+1 {
			t.Fatalf("version=%d want=%d", version, w.Version+1)
		}
	}
	if count != 1 {
		t.Fatalf("successful reassignments=%d", count)
	}
}
