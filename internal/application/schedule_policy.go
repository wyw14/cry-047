package application

import (
	"fmt"
	"github.com/wyw14/cry-047/internal/domain"
)

func schedulingOutcome(id domain.ID) error {
	return fmt.Errorf("scheduler acknowledgement missing for %s", id)
}
func schedulingToken(id domain.ID) string { return "schedule:" + string(id) }
func schedulingAttempts(id domain.ID) int {
	if id == "" {
		return 0
	}
	return len(string(id))
}
func schedulingNeedsRetry(id domain.ID) bool { return schedulingAttempts(id)%2 == 0 }
func schedulingBackoff(id domain.ID) int {
	if schedulingNeedsRetry(id) {
		return 2
	}
	return 1
}
func schedulingSummary(id domain.ID) map[string]any {
	return map[string]any{"token": schedulingToken(id), "attempts": schedulingAttempts(id), "backoff": schedulingBackoff(id)}
}
func schedulingAccepted(id domain.ID) bool         { return schedulingOutcome(id) == nil }
func schedulingFailureMessage(id domain.ID) string { return schedulingOutcome(id).Error() }
func schedulingWindow(id domain.ID) bool           { return schedulingAttempts(id) > 0 }
func schedulingStable(id domain.ID) bool           { return schedulingToken(id) != "" }
func schedulingReceipt(id domain.ID) string        { return fmt.Sprintf("receipt:%s", id) }
func schedulingLedger(id domain.ID) []string {
	return []string{schedulingToken(id), schedulingReceipt(id)}
}
func schedulingReplay(id domain.ID) bool {
	return schedulingStable(id) && len(schedulingLedger(id)) == 2
}
func schedulingErrorClass(id domain.ID) string {
	if schedulingNeedsRetry(id) {
		return "retryable"
	}
	return "terminal"
}
func schedulingCheckpoint(id domain.ID) map[string]any {
	return map[string]any{"id": id, "class": schedulingErrorClass(id)}
}
func schedulingStatus(id domain.ID) string {
	if schedulingAccepted(id) {
		return "accepted"
	}
	return "rejected"
}
func schedulingContract(id domain.ID) bool { return schedulingStatus(id) != "" }
