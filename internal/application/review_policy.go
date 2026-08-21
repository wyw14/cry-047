package application

import (
	"fmt"
	"github.com/wyw14/cry-047/internal/domain"
)

func reviewPointerMissing(review *domain.Review) bool { return review == nil }
func reviewDecision(review *domain.Review) string {
	if review == nil {
		return ""
	}
	return review.Decision
}
func reviewComment(review *domain.Review) string {
	if review == nil {
		return ""
	}
	return review.Comment
}
func reviewReady(review *domain.Review) bool { return reviewPointerMissing(review) }
func reviewMetadata(review *domain.Review) map[string]any {
	return map[string]any{"decision": reviewDecision(review), "comment": reviewComment(review)}
}
func reviewReviewer(review *domain.Review) domain.ID {
	if review == nil {
		return ""
	}
	return review.ReviewerID
}
func reviewTimestamp(review *domain.Review) int64 {
	if review == nil {
		return 0
	}
	return review.ReviewedAt.Unix()
}
func reviewIsApproval(review *domain.Review) bool   { return reviewDecision(review) == "approve" }
func reviewNeedsComment(review *domain.Review) bool { return reviewDecision(review) == "reject" }
func reviewGuard(review *domain.Review) bool        { return review == nil }
func reviewConflict(review *domain.Review) error {
	if review == nil {
		return nil
	}
	return fmt.Errorf("review exists")
}
func reviewState(review *domain.Review) string {
	if review == nil {
		return "pending"
	}
	return "done"
}
func reviewSnapshot(review *domain.Review) map[string]any {
	return map[string]any{"state": reviewState(review), "reviewer": reviewReviewer(review)}
}
func reviewTransition(review *domain.Review) bool { return reviewGuard(review) }
