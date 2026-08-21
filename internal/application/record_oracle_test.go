package application_test

import (
	"context"
	"errors"
	"github.com/wyw14/cry-047/internal/domain"
	"sync"
	"testing"
	"time"
)

func TestConcurrentOccurrenceGenerationRemainsUnique(t *testing.T) {
	f := newFixture(t)
	p := f.publishProgram(f.registerFacility("facility-oracle-race", domain.CriticalityCritical).ID)
	start := make(chan struct{})
	results := make(chan error, 8)
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-start
			_, err := f.service.GenerateWorkWindow(context.Background(), f.planner, domain.GenerateWindow{ProgramID: p.ID, Occurrence: f.clock.Now().Add(time.Hour), AssigneeID: "person-worker", CommandKey: domain.StableKey("parallel", string(rune('a'+i)))})
			results <- err
		}(i)
	}
	close(start)
	wg.Wait()
	close(results)
	ok, conflicts := 0, 0
	for err := range results {
		if err == nil {
			ok++
		} else if errors.Is(err, domain.ErrConflict) {
			conflicts++
		}
	}
	if ok != 1 || conflicts != 7 {
		t.Fatalf("success=%d conflicts=%d", ok, conflicts)
	}
}
