package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"DP_Maintenance/internal/api"
	"DP_Maintenance/internal/api/handlers"
	"DP_Maintenance/internal/service"
	"DP_Maintenance/pkg/config"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	log.Printf("Environment: %s, LogLevel: %s", cfg.Environment, cfg.LogLevel)

	// --- Services (Phase 1: nil repos for stub mode) ---
	authSvc := service.NewAuthService(cfg.JWTSecret)
	catalogSvc := service.NewCatalogService(nil, nil, nil, nil, nil)
	lineageSvc := service.NewLineageService(nil)
	ingestionSvc := service.NewIngestionService(catalogSvc, lineageSvc, cfg.WorkerCount, cfg.ChannelBuffer)

	// Start async worker pool
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ingestionSvc.Start(ctx)

	// --- Handlers ---
	authHandler := handlers.NewAuthHandler(authSvc)
	ingestHandler := handlers.NewIngestHandler()
	datasetHandler := handlers.NewDatasetHandler()
	lineageHandler := handlers.NewLineageHandler(lineageSvc)
	healthHandler := handlers.NewHealthHandler()

	// --- Router ---
	router := api.NewRouter(
		authHandler,
		ingestHandler,
		datasetHandler,
		lineageHandler,
		healthHandler,
		authSvc,
	)
	e := router.Setup()

	// --- Graceful shutdown ---
	go func() {
		addr := fmt.Sprintf(":%s", cfg.ServerPort)
		log.Printf("Starting server on %s", addr)
		if err := e.Start(addr); err != nil {
			log.Printf("Server stopped: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Received shutdown signal")

	// Shutdown Echo server with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := e.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced shutdown: %v", err)
	}

	// Shutdown worker pool
	cancel()
	ingestionSvc.Shutdown()

	log.Println("Server exited gracefully")
}