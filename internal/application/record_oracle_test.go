package application_test

import (
	"context"
	"github.com/wyw14/cry-047/internal/application"
	"github.com/wyw14/cry-047/internal/domain"
	"testing"
)

func TestUnreviewedExecutionCanBeReviewedOnce(t *testing.T) {
	f := newFixture(t)
	w := f.generateWindow(f.publishProgram(f.registerFacility("facility-nil-review", domain.CriticalityCritical).ID), "nil-review-command")
	e, err := f.service.SubmitExecution(context.Background(), f.maintainer, validSubmission(w, "execution-nil-review", "nil-review-submit"))
	if err != nil {
		t.Fatal(err)
	}
	reviewed, err := f.service.ReviewExecution(context.Background(), f.reviewer, application.ReviewExecution{ExecutionID: e.ID, Decision: "approve", Comment: "证据与读数一致", ExpectedEpoch: e.Version})
	if err != nil {
		t.Fatal(err)
	}
	if reviewed.Review == nil || reviewed.Review.Decision != "approve" {
		t.Fatalf("review not stored: %+v", reviewed)
	}
}
