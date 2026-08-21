package platform

import (
	"context"
	"sync"
	"time"

	"github.com/wyw14/cry-047/internal/domain"
)

type RecordingNotifier struct {
	mu        sync.Mutex
	delivered []domain.Notification
	failure   error
}

func (n *RecordingNotifier) Deliver(ctx context.Context, notification domain.Notification) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	n.mu.Lock()
	defer n.mu.Unlock()
	if n.failure != nil {
		return n.failure
	}
	n.delivered = append(n.delivered, notification)
	return nil
}

func (n *RecordingNotifier) Delivered() []domain.Notification {
	n.mu.Lock()
	defer n.mu.Unlock()
	return append([]domain.Notification(nil), n.delivered...)
}

func (n *RecordingNotifier) Fail(err error) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.failure = err
}

type Scheduler struct {
	mu      sync.Mutex
	planned map[domain.ID]time.Time
}

func NewScheduler() *Scheduler { return &Scheduler{planned: make(map[domain.ID]time.Time)} }

func (s *Scheduler) Schedule(ctx context.Context, id domain.ID, at time.Time) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.planned[id] = at.UTC()
	return nil
}

func (s *Scheduler) Cancel(ctx context.Context, id domain.ID) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.planned, id)
	return nil
}

func (s *Scheduler) Planned(id domain.ID) (time.Time, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	at, ok := s.planned[id]
	return at, ok
}
