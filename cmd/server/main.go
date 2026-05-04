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
	"DP_Maintenance/internal/model"
	"DP_Maintenance/internal/repository"
	"DP_Maintenance/internal/repository/memgraph"
	"DP_Maintenance/internal/service"
	"DP_Maintenance/pkg/config"
	"DP_Maintenance/pkg/database"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"gorm.io/gorm"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	log.Printf("Environment: %s, LogLevel: %s", cfg.Environment, cfg.LogLevel)

	var db *gorm.DB
	if cfg.DatabaseURL != "" || cfg.DatabaseHost != "" {
		db, err = database.ConnectPostgres(cfg)
		if err != nil {
			log.Fatalf("Failed to connect to PostgreSQL: %v", err)
		}
		log.Println("Connected to PostgreSQL")
	}

	var mgDriver neo4j.DriverWithContext
	if cfg.MemgraphURI != "" {
		mgDriver, err = database.ConnectMemgraph(cfg)
		if err != nil {
			log.Fatalf("Failed to connect to Memgraph: %v", err)
		}
		log.Println("Connected to Memgraph")

		if err := database.SetupMemgraphConstraints(mgDriver); err != nil {
			log.Printf("Warning: Failed to setup Memgraph constraints: %v", err)
		}
	}

	var datasetRepo repository.DatasetRepository
	var userRepo repository.UserRepository
	var schemaRepo repository.SchemaVersionRepository
	var tagRepo repository.TagRepository
	var lineageRepo repository.LineageRepository

	if db != nil {
		if err := db.AutoMigrate(
			&model.User{},
			&model.Dataset{},
			&model.SchemaVersion{},
			&model.Tag{},
			&model.Job{},
		); err != nil {
			log.Fatalf("Failed to run migrations: %v", err)
		}
		log.Println("Database migrations completed")

		datasetRepo = repository.NewDBDatasetRepository(db)
		userRepo = repository.NewDBUserRepository(db)
		schemaRepo = repository.NewDBSchemaVersionRepository(db)
		tagRepo = repository.NewDBTagRepository(db)
	}

	if mgDriver != nil {
		mgConn := memgraph.NewConnectionFromDriver(mgDriver)
		lineageRepo = memgraph.NewLineageRepository(mgConn)
	}

	authSvc := service.NewAuthService(cfg.JWTSecret)
	catalogSvc := service.NewCatalogService(datasetRepo, schemaRepo, userRepo, tagRepo)
	lineageSvc := service.NewLineageService(lineageRepo)
	ingestionSvc := service.NewIngestionService(catalogSvc, lineageSvc, cfg.WorkerCount, cfg.ChannelBuffer)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ingestionSvc.Start(ctx)

	authHandler := handlers.NewAuthHandler(authSvc, userRepo)
	ingestHandler := handlers.NewIngestHandler()
	datasetHandler := handlers.NewDatasetHandler(catalogSvc)
	lineageHandler := handlers.NewLineageHandler(lineageSvc)
	healthHandler := handlers.NewHealthHandler()

	router := api.NewRouter(
		authHandler,
		ingestHandler,
		datasetHandler,
		lineageHandler,
		healthHandler,
		authSvc,
	)
	e := router.Setup()

	go func() {
		addr := fmt.Sprintf(":%s", cfg.ServerPort)
		log.Printf("Starting server on %s", addr)
		if err := e.Start(addr); err != nil {
			log.Printf("Server stopped: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Received shutdown signal")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := e.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced shutdown: %v", err)
	}

	cancel()
	ingestionSvc.Shutdown()

	if mgDriver != nil {
		mgDriver.Close(shutdownCtx)
	}

	log.Println("Server exited gracefully")
}