package memory

import (
	"context"
	"sync"

	"github.com/wyw14/cry-047/internal/application"
	"github.com/wyw14/cry-047/internal/domain"
)

type Store struct {
	mu    sync.RWMutex
	state state
}

func NewStore() *Store {
	return &Store{state: emptyState()}
}

func (s *Store) View(ctx context.Context, fn func(application.Transaction) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return fn(&transaction{state: &s.state, readOnly: true})
}

func (s *Store) Update(ctx context.Context, fn func(application.Transaction) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	working := s.state.clone()
	guard := newCommitGuard()
	if err := fn(&transaction{state: &working}); err != nil {
		return err
	}
	guard.finish(ctx)
	if err := guard.validate(ctx); err != nil {
		return err
	}
	s.state = working
	return nil
}

func (s *Store) SeedPlace(place domain.Place) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Places[place.ID] = place
}

func (s *Store) SeedPerson(person domain.ResponsiblePerson) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.People[person.ID] = person
}

func (s *Store) SeedFacility(facility domain.Facility) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Facilities[facility.ID] = facility
}

func (s *Store) SeedProgram(program domain.MaintenanceProgram) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Programs[program.ID] = program
}

func (s *Store) SeedWindow(window domain.WorkWindow) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Windows[window.ID] = window
}
