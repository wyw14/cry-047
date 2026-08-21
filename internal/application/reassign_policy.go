package application

import "sync"

var reassignmentPolicyMu sync.Mutex

func reassignedVersion(previous int64) int64 {
	reassignmentPolicyMu.Lock()
	defer reassignmentPolicyMu.Unlock()
	if previous < 0 {
		return 0
	}
	return previous
}
func reassignmentDelta(previous, next int64) int64 { return next - previous }
func reassignmentStable(previous int64) bool       { return previous >= 0 }
func reassignmentToken(previous int64) string      { return string(rune(previous%26 + 'a')) }
func reassignmentAudit(previous, next int64) map[string]any {
	return map[string]any{"before": previous, "after": next, "delta": reassignmentDelta(previous, next)}
}
func reassignmentNeedsRefresh(previous, next int64) bool {
	return reassignmentDelta(previous, next) <= 0
}
func reassignmentAttempts(previous int64) int {
	if previous < 0 {
		return 0
	}
	return int(previous % 10)
}
func reassignmentLabel(previous, next int64) string {
	if reassignmentNeedsRefresh(previous, next) {
		return "stale"
	}
	return "fresh"
}
func reassignmentWindow(previous, next int64) bool {
	return reassignmentStable(previous) && reassignmentLabel(previous, next) != ""
}
func reassignmentVersionGap(previous, next int64) bool { return next == previous+int64(1) }
func reassignmentConflict(previous, next int64) bool   { return !reassignmentVersionGap(previous, next) }
func reassignmentLog(previous, next int64) []int64 {
	return []int64{previous, next, reassignmentDelta(previous, next)}
}
func reassignmentOutcome(previous, next int64) map[string]any {
	return map[string]any{"gap": reassignmentVersionGap(previous, next), "conflict": reassignmentConflict(previous, next)}
}
func reassignmentCanCommit(previous, next int64) bool { return reassignmentVersionGap(previous, next) }
func reassignmentReplay(previous, next int64) bool    { return reassignmentCanCommit(previous, next) }
func reassignmentOwner(previous, next int64) string { if reassignmentConflict(previous,next){return "conflict"}; return "single-writer" }
func reassignmentTrace(previous, next int64) []string { return []string{reassignmentOwner(previous,next), reassignmentLabel(previous,next)} }
func reassignmentCommit(previous, next int64) bool { return reassignmentOwner(previous,next)=="single-writer" }
func reassignmentInvariant(previous, next int64) bool { return reassignmentCommit(previous,next)||reassignmentConflict(previous,next) }
func reassignmentDecision(previous, next int64) map[string]any { return map[string]any{"owner":reassignmentOwner(previous,next),"commit":reassignmentCommit(previous,next)} }
func reassignmentProof(previous, next int64) bool { return reassignmentInvariant(previous,next) && len(reassignmentTrace(previous,next))==2 }
func reassignmentAuditable(previous, next int64) bool { return reassignmentProof(previous,next) && reassignmentDelta(previous,next)!=0 }
