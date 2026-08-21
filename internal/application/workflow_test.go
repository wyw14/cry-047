package application_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/wyw14/cry-047/internal/application"
	"github.com/wyw14/cry-047/internal/domain"
)

func TestRegisterFacilityCreatesAuditTimelineAndNotice(t *testing.T) {
	f := newFixture(t)
	facility := f.registerFacility("facility-reg", domain.CriticalityCritical)
	if facility.State != domain.FacilityNormal || facility.Version != 1 { t.Fatalf("unexpected facility: %+v", facility) }
	if err := f.store.View(context.Background(), func(tx application.Transaction) error {
		if len(tx.ListTimeline(facility.ID)) != 1 { t.Fatalf("timeline count = %d", len(tx.ListTimeline(facility.ID))) }
		if len(tx.ListNotifications()) != 1 || len(tx.ListAudit()) != 1 { t.Fatalf("side effects missing") }
		return nil
	}); err != nil { t.Fatal(err) }
}

func TestPublishProgramPinsRevisionAndSchedulesNextRun(t *testing.T) {
	f := newFixture(t)
	facility := f.registerFacility("facility-program", domain.CriticalityImportant)
	program := f.publishProgram(facility.ID)
	if len(program.Revisions) != 1 || program.Revisions[0].Number != 1 { t.Fatalf("unexpected revisions: %+v", program.Revisions) }
	if planned, ok := f.scheduler.Planned(program.ID); !ok || !planned.Equal(program.NextDueDate) { t.Fatalf("scheduler not updated: %v %v", planned, ok) }
}

func TestConcurrentWindowGenerationDoesNotDuplicateOccurrence(t *testing.T) {
	f := newFixture(t)
	program := f.publishProgram(f.registerFacility("facility-race", domain.CriticalityCritical).ID)
	start := make(chan struct{})
	errorsSeen := make(chan error, 2)
	var wait sync.WaitGroup
	for _, key := range []string{"command-race-a", "command-race-b"} {
		wait.Add(1)
		go func(key string) {
			defer wait.Done(); <-start
			_, err := f.service.GenerateWorkWindow(context.Background(), f.planner, domain.GenerateWindow{ProgramID: program.ID, Occurrence: f.clock.Now().Add(time.Hour), AssigneeID: "person-worker", CommandKey: key})
			errorsSeen <- err
		}(key)
	}
	close(start); wait.Wait(); close(errorsSeen)
	success, conflicts := 0, 0
	for err := range errorsSeen { if err == nil { success++ } else if errors.Is(err, domain.ErrConflict) { conflicts++ } }
	if success != 1 || conflicts != 1 { t.Fatalf("success=%d conflicts=%d", success, conflicts) }
}

func TestReassignWindowUpdatesTaskAndBothNotifications(t *testing.T) {
	f := newFixture(t)
	window := f.generateWindow(f.publishProgram(f.registerFacility("facility-reassign", domain.CriticalityImportant).ID), "command-reassign")
	changed, err := f.service.ReassignWorkWindow(context.Background(), f.planner, domain.ReassignWindow{WindowID: window.ID, FromPersonID: "person-worker", ToPersonID: "person-backup", Reason: "原执行人参与应急抢修", ExpectedEpoch: window.Version})
	if err != nil { t.Fatal(err) }
	if changed.AssigneeID != "person-backup" || changed.Version != window.Version+1 { t.Fatalf("unexpected reassignment: %+v", changed) }
	if err := f.store.View(context.Background(), func(tx application.Transaction) error {
		if len(tx.ListPersonalTasks("person-backup")) != 1 { t.Fatalf("backup task missing") }
		if len(tx.ListNotifications()) < 4 { t.Fatalf("reassignment notifications missing") }
		return nil
	}); err != nil { t.Fatal(err) }
}

func TestSkipWindowAtomicallyCreatesCompensation(t *testing.T) {
	f := newFixture(t)
	window := f.generateWindow(f.publishProgram(f.registerFacility("facility-skip", domain.CriticalityRoutine).ID), "command-skip")
	skipped, compensation, err := f.service.SkipWorkWindow(context.Background(), f.planner, domain.SkipWindow{WindowID: window.ID, Reason: "场馆临时承担大型活动", CompensateAt: f.clock.Now().Add(48 * time.Hour), ExpectedEpoch: window.Version})
	if err != nil { t.Fatal(err) }
	if skipped.State != domain.WorkSkipped || compensation.CompensatesID != skipped.ID || compensation.State != domain.WorkPlanned { t.Fatalf("invalid compensation: %+v %+v", skipped, compensation) }
}

func TestSubmitExecutionUsesPinnedRevisionAndVerifiesEvidence(t *testing.T) {
	f := newFixture(t)
	window := f.generateWindow(f.publishProgram(f.registerFacility("facility-submit", domain.CriticalityCritical).ID), "command-submit")
	submission := validSubmission(window, "execution-submit", "submit-key-001")
	execution, err := f.service.SubmitExecution(context.Background(), f.maintainer, submission)
	if err != nil { t.Fatal(err) }
	if execution.ProgramRevision != window.ProgramRevision || execution.SubmittedBy != f.maintainer.ID { t.Fatalf("unexpected execution: %+v", execution) }
	again, err := f.service.SubmitExecution(context.Background(), f.maintainer, submission)
	if err != nil || again.ID != execution.ID { t.Fatalf("idempotent retry failed: %+v %v", again, err) }
}

func TestReviewExecutionTransitionsWindowAndPreservesAudit(t *testing.T) {
	f := newFixture(t)
	window := f.generateWindow(f.publishProgram(f.registerFacility("facility-review", domain.CriticalityCritical).ID), "command-review")
	execution, err := f.service.SubmitExecution(context.Background(), f.maintainer, validSubmission(window, "execution-review", "review-key-001"))
	if err != nil { t.Fatal(err) }
	reviewed, err := f.service.ReviewExecution(context.Background(), f.reviewer, application.ReviewExecution{ExecutionID: execution.ID, Decision: "approve", Comment: "读数与现场证据一致", ExpectedEpoch: execution.Version})
	if err != nil { t.Fatal(err) }
	if reviewed.Review == nil || reviewed.Review.Decision != "approve" { t.Fatalf("review missing: %+v", reviewed) }
	if err := f.store.View(context.Background(), func(tx application.Transaction) error { window, err := tx.GetWindow(execution.WindowID); if err == nil && window.State != domain.WorkApproved { t.Fatalf("window state=%s", window.State) }; return err }); err != nil { t.Fatal(err) }
}

func TestIncidentRecoveryRequiresOrderedReviewWorkflow(t *testing.T) {
	f := newFixture(t)
	facility := f.registerFacility("facility-incident", domain.CriticalityCritical)
	incident, err := f.service.OpenIncident(context.Background(), f.maintainer, domain.OpenIncident{ID: "incident-lift", FacilityID: facility.ID, Summary: "制动距离超过安全阈值", Severity: domain.CriticalityCritical, OwnerID: "person-worker"})
	if err != nil { t.Fatal(err) }
	incident, err = f.service.AdvanceIncident(context.Background(), f.maintainer, application.AdvanceIncident{IncidentID: incident.ID, Next: domain.IncidentRectifying, Note: "更换制动组件并重新标定", ExpectedEpoch: incident.Version})
	if err != nil { t.Fatal(err) }
	incident, err = f.service.AdvanceIncident(context.Background(), f.maintainer, application.AdvanceIncident{IncidentID: incident.ID, Next: domain.IncidentReinspect, Note: "完成空载和额定载荷测试", ExpectedEpoch: incident.Version})
	if err != nil { t.Fatal(err) }
	incident, err = f.service.AdvanceIncident(context.Background(), f.reviewer, application.AdvanceIncident{IncidentID: incident.ID, Next: domain.IncidentRecovered, Note: "复检读数恢复且连续三次稳定", ExpectedEpoch: incident.Version})
	if err != nil || incident.State != domain.IncidentRecovered { t.Fatalf("recovery failed: %+v %v", incident, err) }
}

func TestRiskBoardCombinesOverdueWindowsAndOpenIncidents(t *testing.T) {
	f := newFixture(t)
	facility := f.registerFacility("facility-risk", domain.CriticalityImportant)
	window := f.generateWindow(f.publishProgram(facility.ID), "command-risk")
	f.clock.Advance(10 * 24 * time.Hour)
	_, err := f.service.OpenIncident(context.Background(), f.maintainer, domain.OpenIncident{ID: "incident-risk", FacilityID: facility.ID, Summary: "门区保护装置间歇失灵", Severity: domain.CriticalityImportant, OwnerID: "person-worker"})
	if err != nil { t.Fatal(err) }
	board, err := f.service.RiskBoard(context.Background(), f.reviewer, f.clock.Now())
	if err != nil { t.Fatal(err) }
	if len(board) != 1 || board[0].FacilityID != window.FacilityID || board[0].OverdueDays == 0 || board[0].OpenIncidents != 1 { t.Fatalf("unexpected board: %+v", board) }
}

func TestExportArchiveIncludesRelatedAggregatesAndChecksum(t *testing.T) {
	f := newFixture(t)
	facility := f.registerFacility("facility-archive", domain.CriticalityRoutine)
	program := f.publishProgram(facility.ID)
	f.generateWindow(program, "command-archive")
	location, bundle, err := f.service.ExportFacilityArchive(context.Background(), f.admin, facility.ID)
	if err != nil { t.Fatal(err) }
	if location == "" || len(bundle.Checksum) != 64 || len(bundle.Programs) != 1 || len(bundle.Windows) != 1 || len(bundle.AuditEvents) < 3 { t.Fatalf("incomplete archive: %s %+v", location, bundle) }
}
