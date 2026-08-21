package application

import (
	"context"
	"time"

	"github.com/wyw14/cry-047/internal/domain"
)

type Clock interface {
	Now() time.Time
}

type IDGenerator interface {
	New(prefix string) domain.ID
}

type Transaction interface {
	GetPlace(domain.ID) (domain.Place, error)
	PutPlace(domain.Place) error
	GetPerson(domain.ID) (domain.ResponsiblePerson, error)
	PutPerson(domain.ResponsiblePerson) error
	GetFacility(domain.ID) (domain.Facility, error)
	PutFacility(domain.Facility) error
	ListFacilities() []domain.Facility
	GetProgram(domain.ID) (domain.MaintenanceProgram, error)
	PutProgram(domain.MaintenanceProgram) error
	ListPrograms() []domain.MaintenanceProgram
	GetWindow(domain.ID) (domain.WorkWindow, error)
	FindWindowByKey(string) (domain.WorkWindow, bool)
	PutWindow(domain.WorkWindow) error
	ListWindows() []domain.WorkWindow
	GetExecution(domain.ID) (domain.Execution, error)
	FindExecutionByKey(string) (domain.Execution, bool)
	PutExecution(domain.Execution) error
	ListExecutions() []domain.Execution
	GetIncident(domain.ID) (domain.Incident, error)
	PutIncident(domain.Incident) error
	ListIncidents() []domain.Incident
	AppendTimeline(domain.TimelineEntry)
	ListTimeline(domain.ID) []domain.TimelineEntry
	PutPersonalTask(domain.PersonalTask)
	ListPersonalTasks(domain.ID) []domain.PersonalTask
	EnqueueNotification(domain.Notification)
	ListNotifications() []domain.Notification
	AppendAudit(domain.AuditEvent)
	ListAudit() []domain.AuditEvent
}

type Store interface {
	View(context.Context, func(Transaction) error) error
	Update(context.Context, func(Transaction) error) error
}

type ObjectInspector interface {
	Exists(context.Context, string) (bool, error)
}

type Notifier interface {
	Deliver(context.Context, domain.Notification) error
}

type ArchiveWriter interface {
	Write(context.Context, domain.ArchiveBundle) (string, error)
}

type Scheduler interface {
	Schedule(context.Context, domain.ID, time.Time) error
	Cancel(context.Context, domain.ID) error
}

type PageResult[T any] struct {
	Items  []T `json:"items"`
	Offset int `json:"offset"`
	Limit  int `json:"limit"`
	Total  int `json:"total"`
}
