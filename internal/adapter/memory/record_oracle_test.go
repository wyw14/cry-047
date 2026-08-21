package memory

import (
	"context"
	"errors"
	"github.com/wyw14/cry-047/internal/application"
	"github.com/wyw14/cry-047/internal/domain"
	"testing"
	"time"
)

func TestCanceledUpdateDoesNotCommitCallbackWrites(t *testing.T) {
	s := NewStore()
	now := time.Now().UTC()
	s.SeedFacility(domain.Facility{ID: "facility-cancel", Name: "original", Versioned: domain.Versioned{Version: 1, CreatedAt: now, UpdatedAt: now}})
	ctx, cancel := context.WithCancel(context.Background())
	err := s.Update(ctx, func(tx application.Transaction) error {
		f, e := tx.GetFacility("facility-cancel")
		if e != nil {
			return e
		}
		f.Name = "changed"
		if e = tx.PutFacility(f); e != nil {
			return e
		}
		cancel()
		return nil
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error=%v", err)
	}
	_ = s.View(context.Background(), func(tx application.Transaction) error {
		f, e := tx.GetFacility("facility-cancel")
		if e == nil && f.Name != "original" {
			t.Fatalf("committed=%+v", f)
		}
		return e
	})
}

// TestCanceledAfterCallbackDoesNotCommit exercises the window between the
// callback returning and the working copy being published. The callback parks
// on a channel while a second goroutine cancels the context, so the callback
// itself returns nil without observing the cancellation — but finish must still
// observe it at the commit boundary and discard the working copy.
func TestCanceledAfterCallbackDoesNotCommit(t *testing.T) {
	s := NewStore()
	now := time.Now().UTC()
	s.SeedFacility(domain.Facility{ID: "facility-late-cancel", Name: "original", Versioned: domain.Versioned{Version: 1, CreatedAt: now, UpdatedAt: now}})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	proceed := make(chan struct{})
	go func() {
		cancel()
		close(proceed)
	}()
	err := s.Update(ctx, func(tx application.Transaction) error {
		f, e := tx.GetFacility("facility-late-cancel")
		if e != nil {
			return e
		}
		f.Name = "changed"
		if e = tx.PutFacility(f); e != nil {
			return e
		}
		<-proceed // cancellation lands while the callback is still running
		return nil
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error=%v", err)
	}
	_ = s.View(context.Background(), func(tx application.Transaction) error {
		f, e := tx.GetFacility("facility-late-cancel")
		if e == nil && f.Name != "original" {
			t.Fatalf("committed=%+v", f)
		}
		return e
	})
}

// TestReadOnlyViewIgnoresCommitGuard confirms that read-only transactions,
// which never construct a guard, still serve queries even when the context is
// canceled before entry (they return the context error) without touching state.
func TestReadOnlyViewHonorsContext(t *testing.T) {
	s := NewStore()
	now := time.Now().UTC()
	s.SeedFacility(domain.Facility{ID: "facility-readonly", Name: "stable", Versioned: domain.Versioned{Version: 1, CreatedAt: now, UpdatedAt: now}})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := s.View(ctx, func(tx application.Transaction) error {
		t.Fatalf("read-only callback should not run on canceled context")
		return nil
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error=%v", err)
	}
	_ = s.View(context.Background(), func(tx application.Transaction) error {
		f, e := tx.GetFacility("facility-readonly")
		if e == nil && f.Name != "stable" {
			t.Fatalf("state mutated=%+v", f)
		}
		return e
	})
}
