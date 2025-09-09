package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"goapianalyzer/internal/adapter/api/router"
	"goapianalyzer/internal/core/domain/service"
	"goapianalyzer/internal/core/usecase"
	"goapianalyzer/internal/infrastructure/config"
	"goapianalyzer/internal/infrastructure/logger"
	"goapianalyzer/internal/infrastructure/repository"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger
	logFormat := "json"
	if cfg.IsDevelopment() {
		logFormat = "text"
	}

	if err := logger.InitLogger(cfg.GetLogLevel(), logFormat); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	log := logger.GetLogger()
	log.WithFields(map[string]interface{}{
		"environment": cfg.Environment,
		"port":        cfg.Port,
		"log_level":   cfg.LogLevel,
	}).Info("Starting GoAPIAnalyzer server")

	// Initialize repository
	repo := repository.NewMemoryAnalysisRepository()
	log.Info("Initialized memory repository")

	// Initialize services
	analyzerService := service.NewAnalyzerService()
	log.Info("Initialized analyzer service")

	// Initialize use cases
	analyzerUsecase := usecase.NewAnalyzerUsecase(repo, analyzerService)
	filterUsecase := usecase.NewFilterUsecase(repo)
	log.Info("Initialized use cases")

	// Initialize router
	r := router.NewRouter(cfg, analyzerUsecase, filterUsecase)
	engine := r.Setup()

	// Create HTTP server
	server := &http.Server{
		Addr:         cfg.GetServerAddress(),
		Handler:      engine,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.WithField("address", server.Addr).Info("Server starting")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.WithError(err).Fatal("Failed to start server")
		}
	}()

	// Print startup information
	printStartupInfo(cfg)

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Server shutting down...")

	// Give outstanding requests a deadline for completion
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := server.Shutdown(ctx); err != nil {
		log.WithError(err).Error("Server forced to shutdown")
	} else {
		log.Info("Server shutdown completed")
	}
}

func printStartupInfo(cfg *config.Config) {
	fmt.Println()
	fmt.Println("🚀 GoAPIAnalyzer Server Started Successfully!")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("🌐 Environment: %s\n", cfg.Environment)
	fmt.Printf("📡 Server Address: http://%s\n", cfg.GetServerAddress())
	fmt.Printf("📋 API Version: %s\n", cfg.APIVersion)
	fmt.Printf("📊 Health Check: http://%s/health\n", cfg.GetServerAddress())
	fmt.Printf("📚 API Documentation: http://%s/api/v1/analyzer\n", cfg.GetServerAddress())
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
	fmt.Println("📖 Available API Endpoints:")
	fmt.Println("  POST   /api/v1/analyzer/scan                    - Scan a Go project")
	fmt.Println("  GET    /api/v1/analyzer/projects/:id            - Get project analysis")
	fmt.Println("  DELETE /api/v1/analyzer/projects/:id            - Delete project analysis")
	fmt.Println("  GET    /api/v1/analyzer/projects/:id/apis       - List API endpoints")
	fmt.Println("  GET    /api/v1/analyzer/projects/:id/nodes      - Get all nodes")
	fmt.Println("  GET    /api/v1/analyzer/projects/:id/stats      - Get project statistics")
	fmt.Println("  POST   /api/v1/analyzer/projects/:id/filters    - Apply filters")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
	fmt.Println("💡 Example curl commands:")
	fmt.Printf(`  curl -X POST http://%s/api/v1/analyzer/scan \
    -H "Content-Type: application/json" \
    -d '{"project_path": "/path/to/your/go/project"}'`, cfg.GetServerAddress())
	fmt.Println()
	fmt.Println()
	fmt.Printf(`  curl http://%s/health`, cfg.GetServerAddress())
	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("Press Ctrl+C to stop the server")
	fmt.Println()
}
