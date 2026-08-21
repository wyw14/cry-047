package application_test

import (
	"context"
	"github.com/wyw14/cry-047/internal/domain"
	"testing"
)

func TestSubmittedExecutionPreservesEveryValidatedReading(t *testing.T) {
	f := newFixture(t)
	w := f.generateWindow(f.publishProgram(f.registerFacility("facility-readings", domain.CriticalityCritical).ID), "reading-command")
	cmd := validSubmission(w, "execution-readings", "reading-submit-key")
	execution, err := f.service.SubmitExecution(context.Background(), f.maintainer, cmd)
	if err != nil {
		t.Fatal(err)
	}
	if len(execution.Readings) != len(cmd.Readings) {
		t.Fatalf("stored=%d submitted=%d", len(execution.Readings), len(cmd.Readings))
	}
	if execution.Readings[1].CheckpointKey != "alarm" {
		t.Fatalf("last reading missing: %+v", execution.Readings)
	}
}
