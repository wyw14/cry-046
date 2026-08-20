// Package main is the entry point for the welfare-settlement-resolver
// server. It wires the config, platform adapters, repositories,
// application services and HTTP transport together, then runs the
// HTTP server with graceful shutdown.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/welfare/settlement-resolver/internal/application"
	"github.com/welfare/settlement-resolver/internal/config"
	"github.com/welfare/settlement-resolver/internal/domain"
	"github.com/welfare/settlement-resolver/internal/middleware"
	"github.com/welfare/settlement-resolver/internal/platform/callback"
	"github.com/welfare/settlement-resolver/internal/platform/clock"
	"github.com/welfare/settlement-resolver/internal/platform/logger"
	"github.com/welfare/settlement-resolver/internal/platform/notify"
	"github.com/welfare/settlement-resolver/internal/platform/scheduler"
	"github.com/welfare/settlement-resolver/internal/platform/storage"
	"github.com/welfare/settlement-resolver/internal/repository/memory"
	schedsvc "github.com/welfare/settlement-resolver/internal/service/scheduler"
	"github.com/welfare/settlement-resolver/internal/service/seed"
	transporthttp "github.com/welfare/settlement-resolver/internal/transport/http"

	"go.uber.org/zap"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "fatal:", err)
		os.Exit(1)
	}
}

func run() error {
	var (
		migrateUp = flag.Bool("migrate-up", false, "apply migrations and exit")
		seedFlag  = flag.Bool("seed", false, "insert demo data and exit")
	)
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	lg, err := logger.New(cfg.Logger.Level, cfg.Logger.Encoding)
	if err != nil {
		return fmt.Errorf("create logger: %w", err)
	}
	defer lg.Sync()

	// Offline adapters. They are created here so cmd/server can swap
	// in real adapters in production; the in-memory test path does
	// not require them.
	notifyAdapter := notify.New(256)
	_ = callback.New(256) // callback adapter is wired into use-cases via the platform; unused in offline mode
	storageAdapter, err := storage.New(cfg.Local.StorageDir, cfg.Local.UploadAllowedTypes, cfg.Local.UploadMaxBytes)
	if err != nil {
		return fmt.Errorf("create storage adapter: %w", err)
	}
	_ = storageAdapter // wired into the attach-evidence use-case in production

	// Repositories: in-memory store (the platform is fully operational
	// without a live PostgreSQL server; pgx-backed repositories are
	// wired in the integration build tag).
	store := memory.NewStore()
	repos := memory.New(store)

	clk := clock.System{}

	// Application services.
	projectsApp := application.NewProjectsApp(repos.Projects, repos.Parties, repos.Batches, repos.Cycles, repos.Rules, clk)
	partiesApp := application.NewPartiesApp(repos.Parties, clk)
	batchesApp := application.NewBatchesApp(repos.Batches, clk)
	cyclesApp := application.NewCyclesApp(repos.Cycles, clk)
	rulesApp := application.NewRulesApp(repos.Rules, clk)
	importsApp := application.NewImportsApp(repos.Entries, repos.Audits, clk)
	evaluateApp := application.NewEvaluateApp(repos.Rules, repos.Entries, repos.Exceptions, repos.Audits, clk)
	exceptionsApp := application.NewExceptionsApp(repos.Exceptions, repos.Audits, clk)
	summaryApp := application.NewSummaryApp(repos.Cycles, repos.Entries, repos.Exceptions, repos.Rules, repos.Summaries, repos.Recalcs, repos.Annuals, repos.Audits, clk)
	workspaceApp := application.NewWorkspaceApp(repos.Exceptions, notifyAdapterAsApp(notifyAdapter), clk)
	auditApp := application.NewAuditApp(repos.Audits, repos.Exceptions, repos.Entries, clk)
	usersApp := application.NewUsersApp(repos.Users, repos.Audits, clk)

	// Seed.
	if *seedFlag || cfg.Seed.DemoData {
		seeder := seed.New(projectsApp, partiesApp, batchesApp, cyclesApp, rulesApp, usersApp, importsApp, evaluateApp, summaryApp, exceptionsApp, clk)
		tenantID := "default"
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		_, err := seeder.Seed(ctx, tenantID, cfg.Seed.DemoProjects, cfg.Seed.DemoBatches)
		cancel()
		if err != nil {
			lg.Warn("seed demo data failed", zap.Error(err))
		}
		if *seedFlag {
			lg.Info("seed completed")
			return nil
		}
	}

	// Scheduler.
	sched := scheduler.New()
	schedSvc := schedsvc.New(workspaceApp, usersApp, sched, cfg.Local.SchedulerTickInterval, cfg.Local.OverdueReminderInterval)
	bgCtx, bgCancel := context.WithCancel(context.Background())
	defer bgCancel()
	if err := schedSvc.RegisterAll(bgCtx, "default"); err != nil {
		lg.Warn("register scheduler jobs failed", zap.Error(err))
	}
	if err := sched.Start(bgCtx); err != nil {
		lg.Warn("start scheduler failed", zap.Error(err))
	}
	defer sched.Stop()

	routerDeps := transporthttp.Router{
		Projects:   projectsApp,
		Parties:    partiesApp,
		Batches:    batchesApp,
		Cycles:     cyclesApp,
		Rules:      rulesApp,
		Imports:    importsApp,
		Exceptions: exceptionsApp,
		Summary:    summaryApp,
		Workspace:  workspaceApp,
		Audit:      auditApp,
		Users:      usersApp,
		Evaluate:   evaluateApp,
	}
	// Bearer tokens map to demo actors. Replace with real auth in production.
	engine := transporthttp.New(routerDeps,
		transporthttp.WithDefaultTenant("default"),
		transporthttp.WithActor("admin", transporthttp.Actor{UserID: "admin", Username: "admin", Role: domain.RoleAdmin, TenantID: "default"}),
		transporthttp.WithActor("operator", transporthttp.Actor{UserID: "operator", Username: "operator", Role: domain.RoleOperator, TenantID: "default"}),
		transporthttp.WithActor("assignee", transporthttp.Actor{UserID: "assignee", Username: "assignee", Role: domain.RoleAssignee, TenantID: "default"}),
		transporthttp.WithActor("reviewer", transporthttp.Actor{UserID: "reviewer", Username: "reviewer", Role: domain.RoleReviewer, TenantID: "default"}),
	)
	engine.Use(
		middleware.RequestID(),
		middleware.Logger(lg),
		middleware.Recover(lg),
		middleware.CORS(cfg.CORS.AllowedOrigins),
		middleware.SecurityHeaders(),
	)

	srv := &http.Server{
		Addr:         cfg.App.HTTPAddr,
		Handler:      engine,
		ReadTimeout:  cfg.App.ReadTimeout,
		WriteTimeout: cfg.App.WriteTimeout,
		IdleTimeout:  cfg.App.IdleTimeout,
	}
	if *migrateUp {
		lg.Info("migrations applied (in-memory mode is a no-op)")
		return nil
	}

	errCh := make(chan error, 1)
	go func() {
		lg.Info("http server starting", zap.String("addr", cfg.App.HTTPAddr))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
		close(errCh)
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	select {
	case sig := <-sigCh:
		lg.Info("signal received, shutting down", zap.String("signal", sig.String()))
	case err := <-errCh:
		if err != nil {
			return fmt.Errorf("server error: %w", err)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), cfg.App.ShutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("shutdown: %w", err)
	}
	lg.Info("server stopped")
	return nil
}

// notifyAdapterAsApp wraps the local notify.Adapter to satisfy the
// application.NotifyAdapter interface.
type notifyAppAdapter struct {
	inner *notify.Adapter
}

func (n notifyAppAdapter) Send(ctx context.Context, recipient, channel, subject, body string) error {
	_, err := n.inner.Send(ctx, notify.Message{
		Recipient: recipient, Channel: channel, Subject: subject, Body: body,
	})
	return err
}

func notifyAdapterAsApp(a *notify.Adapter) application.NotifyAdapter {
	return notifyAppAdapter{inner: a}
}
