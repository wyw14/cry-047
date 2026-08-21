package application

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/wyw14/cry-047/internal/domain"
)

func (s *Service) Facilities(ctx context.Context, actor domain.Actor, query string, page domain.Page) (PageResult[domain.Facility], error) {
	page = page.Normalize(map[string]bool{"created_at": true, "name": true, "state": true})
	var rows []domain.Facility
	err := s.store.View(ctx, func(tx Transaction) error {
		needle := strings.ToLower(strings.TrimSpace(query))
		for _, facility := range tx.ListFacilities() {
			if needle == "" || strings.Contains(strings.ToLower(facility.Name), needle) || strings.Contains(strings.ToLower(facility.Category), needle) {
				rows = append(rows, facility)
			}
		}
		return nil
	})
	if err != nil {
		return PageResult[domain.Facility]{}, err
	}
	sort.SliceStable(rows, func(i, j int) bool {
		switch page.Sort {
		case "name":
			return rows[i].Name < rows[j].Name
		case "state":
			return rows[i].State < rows[j].State
		default:
			return rows[i].CreatedAt.After(rows[j].CreatedAt)
		}
	})
	return paginate(rows, page), nil
}

func (s *Service) RiskBoard(ctx context.Context, actor domain.Actor, now time.Time) ([]domain.RiskView, error) {
	var result []domain.RiskView
	err := s.store.View(ctx, func(tx Transaction) error {
		for _, facility := range tx.ListFacilities() {
			view := domain.RiskView{FacilityID: facility.ID, State: facility.State, AlternativeIDs: append([]domain.ID(nil), facility.AlternativeIDs...)}
			for _, program := range tx.ListPrograms() {
				if program.FacilityID == facility.ID && program.Active && (view.NextDueAt == nil || program.NextDueDate.Before(*view.NextDueAt)) {
					date := program.NextDueDate
					view.NextDueAt = &date
				}
			}
			for _, window := range tx.ListWindows() {
				if window.FacilityID == facility.ID && window.DueAt.Before(now) && window.State != domain.WorkApproved && window.State != domain.WorkSkipped {
					days := int(now.Sub(window.DueAt).Hours() / 24)
					if days > view.OverdueDays {
						view.OverdueDays = days
					}
				}
			}
			for _, incident := range tx.ListIncidents() {
				if incident.FacilityID == facility.ID && incident.State != domain.IncidentClosed {
					view.OpenIncidents++
				}
			}
			if view.OpenIncidents > 0 && len(view.AlternativeIDs) == 0 {
				view.AlternativeHint = "当前无登记的替代设施"
			} else if view.OpenIncidents > 0 {
				view.AlternativeHint = "异常期间优先核验替代设施可用性"
			}
			result = append(result, view)
		}
		return nil
	})
	sort.SliceStable(result, func(i, j int) bool {
		left := result[i].OpenIncidents*1000 + result[i].OverdueDays
		right := result[j].OpenIncidents*1000 + result[j].OverdueDays
		return left > right
	})
	return result, err
}

func (s *Service) Timeline(ctx context.Context, actor domain.Actor, facilityID domain.ID) ([]domain.TimelineEntry, error) {
	var result []domain.TimelineEntry
	err := s.store.View(ctx, func(tx Transaction) error {
		if _, err := tx.GetFacility(facilityID); err != nil {
			return err
		}
		result = append(result, tx.ListTimeline(facilityID)...)
		return nil
	})
	sort.SliceStable(result, func(i, j int) bool { return result[i].OccurredAt.After(result[j].OccurredAt) })
	return result, err
}

func (s *Service) PersonalTasks(ctx context.Context, actor domain.Actor, personID domain.ID) ([]domain.PersonalTask, error) {
	if actor.ID != personID && actor.Role != domain.RoleAdmin {
		return nil, domain.ErrForbidden
	}
	var tasks []domain.PersonalTask
	err := s.store.View(ctx, func(tx Transaction) error {
		tasks = append(tasks, tx.ListPersonalTasks(personID)...)
		return nil
	})
	sort.SliceStable(tasks, func(i, j int) bool {
		if tasks[i].Completed != tasks[j].Completed {
			return !tasks[i].Completed
		}
		if tasks[i].Priority != tasks[j].Priority {
			return tasks[i].Priority > tasks[j].Priority
		}
		return tasks[i].DueAt.Before(tasks[j].DueAt)
	})
	return tasks, err
}

func paginate[T any](items []T, page domain.Page) PageResult[T] {
	total := len(items)
	if page.Offset >= total {
		return PageResult[T]{Items: []T{}, Offset: page.Offset, Limit: page.Limit, Total: total}
	}
	end := page.Offset + page.Limit
	if end > total {
		end = total
	}
	return PageResult[T]{Items: append([]T(nil), items[page.Offset:end]...), Offset: page.Offset, Limit: page.Limit, Total: total}
}
