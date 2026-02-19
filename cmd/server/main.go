package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/HuynhHoangPhuc/mcs-erp/internal/agent"
	agentdomain "github.com/HuynhHoangPhuc/mcs-erp/internal/agent/domain"
	agentinfra "github.com/HuynhHoangPhuc/mcs-erp/internal/agent/infrastructure"
	"github.com/HuynhHoangPhuc/mcs-erp/internal/core"
	coredelivery "github.com/HuynhHoangPhuc/mcs-erp/internal/core/delivery"
	"github.com/HuynhHoangPhuc/mcs-erp/internal/core/infrastructure"
	"github.com/HuynhHoangPhuc/mcs-erp/internal/platform/config"
	"github.com/HuynhHoangPhuc/mcs-erp/internal/platform/database"
	"github.com/HuynhHoangPhuc/mcs-erp/internal/platform/eventbus"
	platformgrpc "github.com/HuynhHoangPhuc/mcs-erp/internal/platform/grpc"
	platformmod "github.com/HuynhHoangPhuc/mcs-erp/internal/platform/module"
	"github.com/HuynhHoangPhuc/mcs-erp/internal/platform/tenant"
	"github.com/HuynhHoangPhuc/mcs-erp/internal/hr"
	"github.com/HuynhHoangPhuc/mcs-erp/internal/room"
	"github.com/HuynhHoangPhuc/mcs-erp/internal/subject"
	"github.com/HuynhHoangPhuc/mcs-erp/internal/timetable"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	// Load config
	cfg, err := config.Load()
	if err != nil {
		slog.Error("config load failed", "error", err)
		os.Exit(1)
	}

	// Graceful shutdown context
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Database pool
	pool, err := database.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("database connection failed", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	// Ensure _template schema exists
	migrator := database.NewMigrator(pool)
	if err := migrator.EnsureTemplateSchema(ctx); err != nil {
		slog.Error("failed to create _template schema", "error", err)
		os.Exit(1)
	}

	// Event bus
	bus, err := eventbus.New()
	if err != nil {
		slog.Error("event bus creation failed", "error", err)
		os.Exit(1)
	}
	_ = bus // will be used by modules

	// JWT service
	jwtSvc := infrastructure.NewJWTService(cfg.JWTSecret, cfg.JWTExpiry)

	// Module registry
	registry := platformmod.NewRegistry()

	// Register core module (auth, users, roles)
	coreMod := core.NewModuleWithDeps(pool, jwtSvc)
	if err := registry.Register(coreMod); err != nil {
		slog.Error("failed to register core module", "error", err)
		os.Exit(1)
	}

	// Register HR module (teachers, departments, availability)
	hrMod := hr.NewModule(pool, coreMod.AuthService())
	if err := registry.Register(hrMod); err != nil {
		slog.Error("failed to register hr module", "error", err)
		os.Exit(1)
	}

	// Register subject module (subjects, categories, prerequisites)
	subjectMod := subject.NewModule(pool, coreMod.AuthService())
	if err := registry.Register(subjectMod); err != nil {
		slog.Error("failed to register subject module", "error", err)
		os.Exit(1)
	}

	// Register room module (rooms, availability)
	roomMod := room.NewModule(pool, coreMod.AuthService())
	if err := registry.Register(roomMod); err != nil {
		slog.Error("failed to register room module", "error", err)
		os.Exit(1)
	}

	// Register timetable module (semesters, scheduling, assignments)
	timetableMod := timetable.NewModuleWithRepos(
		pool, coreMod.AuthService(),
		hrMod.TeacherRepo(), hrMod.AvailabilityRepo(),
		subjectMod.SubjectRepo(),
		roomMod.RoomRepo(), roomMod.RoomAvailabilityRepo(),
	)
	if err := registry.Register(timetableMod); err != nil {
		slog.Error("failed to register timetable module", "error", err)
		os.Exit(1)
	}

	// Register agent module (AI chat, conversations, suggestions)
	toolRegistry := agentinfra.NewToolRegistry()
	llmCfg := agentdomain.LLMConfig{
		Primary: agentdomain.ProviderConfig{
			Provider: agentdomain.LLMProvider(cfg.LLMProvider),
			Model:    cfg.LLMModel,
			APIKey:   cfg.LLMAPIKey,
			BaseURL:  cfg.OllamaURL,
		},
	}
	if cfg.LLMFallbackProvider != "" {
		llmCfg.Fallback = &agentdomain.ProviderConfig{
			Provider: agentdomain.LLMProvider(cfg.LLMFallbackProvider),
			Model:    cfg.LLMFallbackModel,
			APIKey:   cfg.LLMFallbackAPIKey,
			BaseURL:  cfg.OllamaURL,
		}
	}
	agentMod := agent.NewModule(pool, coreMod.AuthService(), toolRegistry, llmCfg, cfg.RedisURL)
	if err := registry.Register(agentMod); err != nil {
		slog.Error("failed to register agent module", "error", err)
		os.Exit(1)
	}

	// HTTP router
	mux := http.NewServeMux()

	// Health check (public, no tenant needed)
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	// Bootstrap modules (migrate, register routes, register events)
	if err := platformmod.Bootstrap(ctx, registry, mux); err != nil {
		slog.Error("module bootstrap failed", "error", err)
		os.Exit(1)
	}

	// Wrap mux with body size limit + tenant middleware
	handler := coredelivery.MaxBodySize(1 << 20)(tenant.Middleware(mux))

	// HTTP server
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// gRPC server
	grpcSrv := platformgrpc.NewServer()

	// errCh collects fatal errors from server goroutines
	errCh := make(chan error, 2)

	// Start gRPC server
	go func() {
		lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
		if err != nil {
			errCh <- fmt.Errorf("gRPC listen: %w", err)
			return
		}
		slog.Info("gRPC server starting", "port", cfg.GRPCPort)
		if err := grpcSrv.Serve(lis); err != nil {
			slog.Error("gRPC server failed", "error", err)
		}
	}()

	// Start HTTP server
	go func() {
		slog.Info("HTTP server starting", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- fmt.Errorf("HTTP server: %w", err)
		}
	}()

	// Wait for shutdown signal or fatal error
	select {
	case <-ctx.Done():
	case err := <-errCh:
		slog.Error("server startup failed", "error", err)
		stop()
	}
	slog.Info("shutting down gracefully")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	grpcSrv.GracefulStop()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("HTTP shutdown error", "error", err)
		os.Exit(1)
	}
	slog.Info("server stopped")
}
