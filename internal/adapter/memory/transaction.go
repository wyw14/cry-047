package memory

import (
	"fmt"
	"sort"

	"github.com/wyw14/cry-047/internal/domain"
)

type transaction struct {
	state    *state
	readOnly bool
}

func (t *transaction) writable() error {
	if t.readOnly {
		return fmt.Errorf("read-only transaction: %w", domain.ErrInvalidState)
	}
	return nil
}

func (t *transaction) GetPlace(id domain.ID) (domain.Place, error) {
	value, ok := t.state.Places[id]
	if !ok {
		return domain.Place{}, domain.Wrap(domain.ErrNotFound, "place", string(id))
	}
	return value, nil
}

func (t *transaction) PutPlace(value domain.Place) error {
	if err := t.writable(); err != nil {
		return err
	}
	t.state.Places[value.ID] = value
	return nil
}

func (t *transaction) GetPerson(id domain.ID) (domain.ResponsiblePerson, error) {
	value, ok := t.state.People[id]
	if !ok {
		return domain.ResponsiblePerson{}, domain.Wrap(domain.ErrNotFound, "person", string(id))
	}
	return value, nil
}

func (t *transaction) PutPerson(value domain.ResponsiblePerson) error {
	if err := t.writable(); err != nil {
		return err
	}
	t.state.People[value.ID] = value
	return nil
}

func (t *transaction) GetFacility(id domain.ID) (domain.Facility, error) {
	value, ok := t.state.Facilities[id]
	if !ok {
		return domain.Facility{}, domain.Wrap(domain.ErrNotFound, "facility", string(id))
	}
	value.AlternativeIDs = append([]domain.ID(nil), value.AlternativeIDs...)
	return value, nil
}

func (t *transaction) PutFacility(value domain.Facility) error {
	if err := t.writable(); err != nil {
		return err
	}
	value.AlternativeIDs = append([]domain.ID(nil), value.AlternativeIDs...)
	t.state.Facilities[value.ID] = value
	return nil
}

func (t *transaction) ListFacilities() []domain.Facility {
	result := make([]domain.Facility, 0, len(t.state.Facilities))
	for _, value := range t.state.Facilities {
		value.AlternativeIDs = append([]domain.ID(nil), value.AlternativeIDs...)
		result = append(result, value)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result
}

func (t *transaction) GetProgram(id domain.ID) (domain.MaintenanceProgram, error) {
	value, ok := t.state.Programs[id]
	if !ok {
		return domain.MaintenanceProgram{}, domain.Wrap(domain.ErrNotFound, "program", string(id))
	}
	value.Revisions = cloneRevisions(value.Revisions)
	return value, nil
}

func (t *transaction) PutProgram(value domain.MaintenanceProgram) error {
	if err := t.writable(); err != nil {
		return err
	}
	value.Revisions = cloneRevisions(value.Revisions)
	t.state.Programs[value.ID] = value
	return nil
}

func (t *transaction) ListPrograms() []domain.MaintenanceProgram {
	result := make([]domain.MaintenanceProgram, 0, len(t.state.Programs))
	for _, value := range t.state.Programs {
		value.Revisions = cloneRevisions(value.Revisions)
		result = append(result, value)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result
}

func (t *transaction) GetWindow(id domain.ID) (domain.WorkWindow, error) {
	value, ok := t.state.Windows[id]
	if !ok {
		return domain.WorkWindow{}, domain.Wrap(domain.ErrNotFound, "work_window", string(id))
	}
	return value, nil
}

func (t *transaction) FindWindowByKey(key string) (domain.WorkWindow, bool) {
	for _, value := range t.state.Windows {
		if value.IdempotencyKey == key {
			return value, true
		}
	}
	return domain.WorkWindow{}, false
}

func (t *transaction) PutWindow(value domain.WorkWindow) error {
	if err := t.writable(); err != nil {
		return err
	}
	t.state.Windows[value.ID] = value
	return nil
}

func (t *transaction) ListWindows() []domain.WorkWindow {
	result := make([]domain.WorkWindow, 0, len(t.state.Windows))
	for _, value := range t.state.Windows {
		result = append(result, value)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Occurrence.Before(result[j].Occurrence) })
	return result
}

func (t *transaction) GetExecution(id domain.ID) (domain.Execution, error) {
	value, ok := t.state.Executions[id]
	if !ok {
		return domain.Execution{}, domain.Wrap(domain.ErrNotFound, "execution", string(id))
	}
	return cloneExecution(value), nil
}

func (t *transaction) FindExecutionByKey(key string) (domain.Execution, bool) {
	for _, value := range t.state.Executions {
		if value.IdempotencyKey == key {
			return cloneExecution(value), true
		}
	}
	return domain.Execution{}, false
}

func (t *transaction) PutExecution(value domain.Execution) error {
	if err := t.writable(); err != nil {
		return err
	}
	t.state.Executions[value.ID] = cloneExecution(value)
	return nil
}

func (t *transaction) ListExecutions() []domain.Execution {
	result := make([]domain.Execution, 0, len(t.state.Executions))
	for _, value := range t.state.Executions {
		result = append(result, cloneExecution(value))
	}
	return result
}

func cloneExecution(value domain.Execution) domain.Execution {
	value.Readings = append([]domain.Reading(nil), value.Readings...)
	value.Evidence = append([]domain.Evidence(nil), value.Evidence...)
	value.Materials = append([]domain.MaterialUse(nil), value.Materials...)
	if value.Review != nil {
		review := *value.Review
		value.Review = &review
	}
	return value
}
