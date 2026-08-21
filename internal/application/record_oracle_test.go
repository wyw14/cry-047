package application_test

import (
	"context"
	"errors"
	"testing"

	"github.com/wyw14/cry-047/internal/application"
	"github.com/wyw14/cry-047/internal/domain"
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

// TestFirstReviewSucceedsThenSecondIsRejected guards the original defect: the
// first review of a freshly submitted record used to be rejected as "already
// reviewed" because the duplicate guard was inverted, yet a true duplicate
// review must still be rejected to keep records from being reviewed twice.
func TestFirstReviewSucceedsThenSecondIsRejected(t *testing.T) {
	f := newFixture(t)
	w := f.generateWindow(f.publishProgram(f.registerFacility("facility-double-review", domain.CriticalityCritical).ID), "double-review-command")
	e, err := f.service.SubmitExecution(context.Background(), f.maintainer, validSubmission(w, "execution-double-review", "double-review-submit"))
	if err != nil {
		t.Fatal(err)
	}

	if _, err := f.service.ReviewExecution(context.Background(), f.reviewer, application.ReviewExecution{
		ExecutionID: e.ID, Decision: "approve", Comment: "读数与现场证据一致", ExpectedEpoch: e.Version,
	}); err != nil {
		t.Fatalf("first review of fresh record must succeed, got: %v", err)
	}

	_, err = f.service.ReviewExecution(context.Background(), f.reviewer, application.ReviewExecution{
		ExecutionID: e.ID, Decision: "reject", Comment: "需要重新核对读数", ExpectedEpoch: e.Version + 1,
	})
	if err == nil {
		t.Fatalf("duplicate review must be rejected")
	}
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("duplicate review must surface a conflict, got: %v", err)
	}
}
