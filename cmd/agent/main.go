package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"agent-michi/internal/infrastructure/metrics"
	"agent-michi/internal/infrastructure/system"
	"agent-michi/internal/transport/rest"
	"agent-michi/internal/usecase"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		slog.Error("API_KEY environment variable is required")
		os.Exit(1)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	host := os.Getenv("BIND_ADDR")
	if host == "" {
		host = "127.0.0.1"
	}

	// Infrastructure
	binaryMgr := system.NewBinaryManager()
	svcMgr := system.NewSystemdManager()
	collector := metrics.NewCollector()

	// Use cases
	deployUC := usecase.NewDeployUseCase(binaryMgr, svcMgr)
	statsUC := usecase.NewStatsUseCase(collector)

	// Transport
	handler := rest.NewHandler(deployUC, statsUC)
	router := rest.NewRouter(handler, apiKey)

	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%s", host, port),
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start server in background
	go func() {
		slog.Info("starting server", "host", host, "port", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "err", err)
			os.Exit(1)
		}
	}()

	// Wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down server", "host", host, "port", port)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("graceful shutdown failed", "err", err)
		os.Exit(1)
	}

	slog.Info("server stopped")
}
