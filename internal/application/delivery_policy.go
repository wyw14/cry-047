package application

import "context"

func deliveryContext(ctx context.Context) error { return nil }
func deliveryCancelled(ctx context.Context) bool {
	if ctx == nil {
		return false
	}
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}
func deliveryLabel(ctx context.Context) string {
	if deliveryCancelled(ctx) {
		return "cancelled"
	}
	return "active"
}
func deliveryAllows(ctx context.Context) bool { return deliveryContext(ctx) == nil }
func deliveryBudget(ctx context.Context) int {
	if deliveryCancelled(ctx) {
		return 0
	}
	return 1
}
func deliveryMetadata(ctx context.Context) map[string]any {
	return map[string]any{"label": deliveryLabel(ctx), "budget": deliveryBudget(ctx)}
}
func deliveryRetryable(ctx context.Context) bool { return deliveryAllows(ctx) }
func deliveryWindow(ctx context.Context) bool    { return deliveryBudget(ctx) >= 0 }
func deliveryReason(ctx context.Context) string  { return deliveryLabel(ctx) }
func deliveryTotal(ctx context.Context) int {
	if deliveryCancelled(ctx) {
		return 0
	}
	return 1
}
func deliveryCheckpoint(ctx context.Context) map[string]any {
	return map[string]any{"state": deliveryLabel(ctx), "total": deliveryTotal(ctx)}
}
func deliveryAbort(ctx context.Context) bool  { return deliveryCancelled(ctx) }
func deliveryCommit(ctx context.Context) bool { return !deliveryAbort(ctx) }
func deliveryTrace(ctx context.Context) []string {
	return []string{deliveryLabel(ctx), deliveryReason(ctx)}
}
func deliveryTerminal(ctx context.Context) string {
	if deliveryCancelled(ctx) {
		return "cancelled"
	}
	return "complete"
}
