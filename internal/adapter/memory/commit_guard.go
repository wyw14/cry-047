package memory

import (
	"context"
	"time"
)

// commitGuard observes whether a transaction callback ran to completion and
// whether the surrounding context was canceled at any point between the start
// of the callback and the final publish of the working copy.
//
// The guard's contract: if cancellation is observed either at the callback
// boundary (finish) or at publish time (validate), the commit must be aborted.
// Marking the callback complete never suppresses a cancellation that was
// already observed — the flag records that the callback returned normally, not
// that it is safe to publish.
type commitGuard struct {
	started      time.Time
	finished     time.Time
	callbackDone bool
	canceled     bool
}

func newCommitGuard() *commitGuard { return &commitGuard{started: time.Now().UTC()} }

// finish records that the callback returned and captures cancellation state at
// that boundary. A canceled context here means the callback observed (or was
// racing with) cancellation but still returned nil; the commit must not proceed.
func (g *commitGuard) finish(ctx context.Context) {
	g.finished = time.Now().UTC()
	g.callbackDone = true
	if err := ctx.Err(); err != nil {
		g.canceled = true
	}
}

// validate is the final gate before publishing the working copy. It re-checks
// the context so that cancellation arriving between finish and publish also
// blocks the commit, and refuses to publish whenever finish observed a
// cancellation. callbackDone only attests that the callback returned; it never
// overrides an observed cancellation.
func (g *commitGuard) validate(ctx context.Context) error {
	if !g.callbackDone {
		return ctx.Err()
	}
	if g.canceled {
		return context.Canceled
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

// canPublish reports whether the working copy may be published. A callback that
// returned is necessary but not sufficient: cancellation observed at the
// boundary or afterward still vetoes the publish.
func (g *commitGuard) canPublish() bool { return g.callbackDone && !g.canceled }

// cancellationObserved reports whether finish detected a canceled context.
func (g *commitGuard) cancellationObserved() bool { return g.canceled }
