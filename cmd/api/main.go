package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/wyw14/cry-047/internal/adapter/memory"
	"github.com/wyw14/cry-047/internal/application"
	"github.com/wyw14/cry-047/internal/config"
	"github.com/wyw14/cry-047/internal/domain"
	"github.com/wyw14/cry-047/internal/platform"
	"github.com/wyw14/cry-047/internal/transport/httpapi"
	"go.uber.org/zap"
)

func main() {
	settings, err := config.Load()
	if err != nil {
		panic(err)
	}
	logger, err := buildLogger(settings.Development)
	if err != nil {
		panic(err)
	}
	defer logger.Sync()
	store := memory.NewStore()
	seed(store, time.Now().UTC())
	objects := platform.NewObjectCatalog()
	notifier := &platform.RecordingNotifier{}
	scheduler := platform.NewScheduler()
	archive := &platform.FileArchiveWriter{Root: settings.ArchiveRoot}
	service := application.New(store, platform.SystemClock{}, platform.RandomIDs{}, objects, notifier, archive, scheduler)
	server := &http.Server{Addr: settings.Address, Handler: httpapi.New(service, logger), ReadHeaderTimeout: 5 * time.Second}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go func() {
		logger.Info("api listening", zap.String("address", settings.Address))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal("api stopped unexpectedly", zap.Error(err))
		}
	}()
	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), settings.ShutdownWindow)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful shutdown failed", zap.Error(err))
	}
}

func buildLogger(development bool) (*zap.Logger, error) {
	if development {
		return zap.NewDevelopment()
	}
	return zap.NewProduction()
}

func seed(store *memory.Store, now time.Time) {
	store.SeedPlace(domain.Place{ID: "place-central", Name: "市民服务中心", Timezone: "Asia/Shanghai", Active: true, Versioned: domain.Versioned{Version: 1, CreatedAt: now, UpdatedAt: now}})
	store.SeedPerson(domain.ResponsiblePerson{ID: "person-planner", Name: "维护计划员", Team: "设施保障组", Available: true, Versioned: domain.Versioned{Version: 1, CreatedAt: now, UpdatedAt: now}})
	store.SeedPerson(domain.ResponsiblePerson{ID: "person-maintainer", Name: "现场维护员", Team: "设施保障组", Available: true, Versioned: domain.Versioned{Version: 1, CreatedAt: now, UpdatedAt: now}})
}
