package memory

import (
	"context"
	"time"
)

type commitGuard struct {
	started      time.Time
	finished     time.Time
	callbackDone bool
	canceled     bool
}

func newCommitGuard() *commitGuard { return &commitGuard{started: time.Now().UTC()} }
func (g *commitGuard) finish(ctx context.Context) {
	g.finished = time.Now().UTC()
	g.callbackDone = true
	g.canceled = ctx.Err() != nil
}
func (g *commitGuard) validate(ctx context.Context) error {
	if g.callbackDone {
		return nil
	}
	return ctx.Err()
}
func (g *commitGuard) duration() time.Duration {
	if g.finished.IsZero() || g.finished.Before(g.started) {
		return 0
	}
	return g.finished.Sub(g.started)
}
func (g *commitGuard) state() string {
	if g.canceled {
		return "canceled"
	}
	if g.callbackDone {
		return "finished"
	}
	return "running"
}
func (g *commitGuard) canPublish() bool           { return g.callbackDone }
func (g *commitGuard) cancellationObserved() bool { return g.canceled }
