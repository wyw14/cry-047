package application_test

import (
	"context"
	"errors"
	"github.com/wyw14/cry-047/internal/application"
	"github.com/wyw14/cry-047/internal/domain"
	"github.com/wyw14/cry-047/internal/platform"
	"testing"
	"time"
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
}
