package application_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/wyw14/cry-047/internal/application"
	"github.com/wyw14/cry-047/internal/domain"
	"github.com/wyw14/cry-047/internal/platform"
)

type failingScheduler struct{ err error }

func (s failingScheduler) Schedule(context.Context, domain.ID, time.Time) error { return s.err }
func (s failingScheduler) Cancel(context.Context, domain.ID) error              { return nil }
func TestProgramPublishReportsSchedulerFailureWithoutMasking(t *testing.T) {
	f := newFixture(t)
	sentinel := errors.New("scheduler storage unavailable")
	f.service = application.New(f.store, f.clock, &sequenceIDs{}, f.objects, &platform.RecordingNotifier{}, &platform.FileArchiveWriter{Root: t.TempDir()}, failingScheduler{err: sentinel})
	_, err := f.service.PublishProgram(context.Background(), f.planner, domain.PublishProgram{ID: "program-error", FacilityID: f.registerFacility("facility-error", domain.CriticalityImportant).ID, Title: "季度巡检", CycleDays: 90, EffectiveFrom: f.clock.Now(), Checks: []domain.Checkpoint{{Key: "visual", Label: "外观", Kind: domain.CheckText, Required: true}}})
	if !errors.Is(err, sentinel) {
		t.Fatalf("scheduler error lost: %v", err)
	}
	if err == nil || err.Error() == "" || !contains(err.Error(), "scheduler storage unavailable") {
		t.Fatalf("downstream reason not surfaced: %v", err)
	}
}

func contains(haystack, needle string) bool { return needle != "" && strings.Contains(haystack, needle) }

// TestProgramPublishSucceedsWhenSchedulerAccepts guards the happy path: when the
// scheduler accepts the program, PublishProgram must report success rather than
// the spurious "scheduler acknowledgement missing" confirmation error that the
// old masking logic produced on every successful publish.
func TestProgramPublishSucceedsWhenSchedulerAccepts(t *testing.T) {
	f := newFixture(t)
	program := f.publishProgram(f.registerFacility("facility-ok", domain.CriticalityImportant).ID)
	if planned, ok := f.scheduler.Planned(program.ID); !ok || !planned.Equal(program.NextDueDate) {
		t.Fatalf("scheduler not updated: %v %v", planned, ok)
	}
}
