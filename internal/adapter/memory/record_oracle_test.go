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
