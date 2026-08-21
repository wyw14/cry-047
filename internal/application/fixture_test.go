package application_test

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/wyw14/cry-047/internal/adapter/memory"
	"github.com/wyw14/cry-047/internal/application"
	"github.com/wyw14/cry-047/internal/domain"
	"github.com/wyw14/cry-047/internal/platform"
)

type fixedClock struct {
	mu  sync.Mutex
	now time.Time
}

func (c *fixedClock) Now() time.Time { c.mu.Lock(); defer c.mu.Unlock(); return c.now }
func (c *fixedClock) Advance(duration time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.now = c.now.Add(duration)
}

type sequenceIDs struct {
	mu   sync.Mutex
	next int
}

func (g *sequenceIDs) New(prefix string) domain.ID {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.next++
	return domain.ID(fmt.Sprintf("%s-%04d", prefix, g.next))
}

type fixture struct {
	t          *testing.T
	store      *memory.Store
	service    *application.Service
	clock      *fixedClock
	objects    *platform.ObjectCatalog
	scheduler  *platform.Scheduler
	admin      domain.Actor
	planner    domain.Actor
	maintainer domain.Actor
	reviewer   domain.Actor
}

func newFixture(t *testing.T) *fixture {
	t.Helper()
	now := time.Date(2026, 8, 21, 2, 0, 0, 0, time.UTC)
	store := memory.NewStore()
	store.SeedPlace(domain.Place{ID: "place-main", Name: "公共服务中心", Timezone: "Asia/Shanghai", Active: true, Versioned: domain.Versioned{Version: 1, CreatedAt: now, UpdatedAt: now}})
	for _, person := range []domain.ResponsiblePerson{
		{ID: "person-owner", Name: "责任人", Team: "保障一组", Available: true},
		{ID: "person-worker", Name: "维护员", Team: "保障一组", Available: true},
		{ID: "person-backup", Name: "替班维护员", Team: "保障二组", Available: true},
	} {
		person.Versioned = domain.Versioned{Version: 1, CreatedAt: now, UpdatedAt: now}
		store.SeedPerson(person)
	}
	clock := &fixedClock{now: now}
	objects := platform.NewObjectCatalog("evidence/photo-a.jpg", "evidence/report-a.pdf")
	scheduler := platform.NewScheduler()
	service := application.New(store, clock, &sequenceIDs{}, objects, &platform.RecordingNotifier{}, &platform.FileArchiveWriter{Root: t.TempDir()}, scheduler)
	return &fixture{t: t, store: store, service: service, clock: clock, objects: objects, scheduler: scheduler,
		admin: domain.Actor{ID: "actor-admin", Role: domain.RoleAdmin}, planner: domain.Actor{ID: "actor-planner", Role: domain.RolePlanner},
		maintainer: domain.Actor{ID: "person-worker", Role: domain.RoleMaintainer}, reviewer: domain.Actor{ID: "actor-reviewer", Role: domain.RoleReviewer}}
}

func (f *fixture) registerFacility(id string, criticality domain.Criticality) domain.Facility {
	f.t.Helper()
	result, err := f.service.RegisterFacility(context.Background(), f.planner, domain.RegisterFacility{
		ID: domain.ID(id), PlaceID: "place-main", Name: "无障碍电梯" + id, Category: "垂直交通", Criticality: criticality, ResponsibleID: "person-owner",
	})
	if err != nil {
		f.t.Fatalf("register facility: %v", err)
	}
	return result
}

func (f *fixture) publishProgram(facilityID domain.ID) domain.MaintenanceProgram {
	f.t.Helper()
	minimum, maximum := 10.0, 20.0
	result, err := f.service.PublishProgram(context.Background(), f.planner, domain.PublishProgram{
		ID: domain.ID("program-" + string(facilityID)), FacilityID: facilityID, Title: "月度制动与平层巡检", CycleDays: 30,
		ShutdownMinutes: 45, EffectiveFrom: f.clock.Now().Add(-time.Hour),
		Checks:    []domain.Checkpoint{{Key: "brake", Label: "制动距离", Kind: domain.CheckNumber, Required: true, Min: &minimum, Max: &maximum}, {Key: "alarm", Label: "报警联动", Kind: domain.CheckBoolean, Required: true}},
		Materials: []domain.MaterialNeed{{SKU: "LUBE-01", Quantity: 1, Unit: "瓶"}},
	})
	if err != nil {
		f.t.Fatalf("publish program: %v", err)
	}
	return result
}

func (f *fixture) generateWindow(program domain.MaintenanceProgram, key string) domain.WorkWindow {
	f.t.Helper()
	result, err := f.service.GenerateWorkWindow(context.Background(), f.planner, domain.GenerateWindow{
		ProgramID: program.ID, Occurrence: f.clock.Now().Add(time.Hour), AssigneeID: "person-worker", CommandKey: key,
	})
	if err != nil {
		f.t.Fatalf("generate window: %v", err)
	}
	return result
}

func validSubmission(window domain.WorkWindow, id, key string) domain.SubmitExecution {
	boolean := true
	number := 15.0
	return domain.SubmitExecution{ID: domain.ID(id), WindowID: window.ID, IdempotencyKey: key,
		Readings:  []domain.Reading{{CheckpointKey: "brake", NumberValue: &number}, {CheckpointKey: "alarm", BooleanValue: &boolean}},
		Evidence:  []domain.Evidence{{ObjectKey: "evidence/photo-a.jpg", MediaType: "image/jpeg", SHA256: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", Description: "制动测试现场"}},
		Materials: []domain.MaterialUse{{SKU: "LUBE-01", Quantity: 1, Unit: "瓶"}}}
}
