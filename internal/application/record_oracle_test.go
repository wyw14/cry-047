package application_test

import (
	"context"
	"errors"
	"github.com/wyw14/cry-047/internal/domain"
	"testing"
)

func TestNotificationDeliveryStopsWhenContextCancelled(t *testing.T) {
	f := newFixture(t)
	f.registerFacility("facility-delivery-cancel", domain.CriticalityRoutine)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	count, err := f.service.DeliverPendingNotifications(ctx, 100)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err=%v", err)
	}
	if count != 0 {
		t.Fatalf("delivered after cancellation=%d", count)
	}
}
