package router

import (
	"goapianalyzer/internal/adapter/api/handler"
	"goapianalyzer/internal/adapter/api/middleware"
	"goapianalyzer/internal/core/usecase"
	"goapianalyzer/internal/infrastructure/config"

	"github.com/gin-gonic/gin"
)

type Router struct {
	engine          *gin.Engine
	config          *config.Config
	analyzerUsecase *usecase.AnalyzerUsecase
	filterUsecase   *usecase.FilterUsecase
}

func NewRouter(
	config *config.Config,
	analyzerUsecase *usecase.AnalyzerUsecase,
	filterUsecase *usecase.FilterUsecase,
) *Router {
	// Set Gin mode based on environment
	if config.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	engine := gin.New()

	return &Router{
		engine:          engine,
		config:          config,
		analyzerUsecase: analyzerUsecase,
		filterUsecase:   filterUsecase,
	}
}

func (r *Router) Setup() *gin.Engine {
	// Add global middlewares
	r.setupMiddlewares()

	// Setup routes
	r.setupRoutes()

	return r.engine
}

func (r *Router) setupMiddlewares() {
	// Recovery middleware
	r.engine.Use(middleware.RecoveryLoggerMiddleware())

	// CORS middleware - only use basic CORS in development
	if r.config.Environment == "development" {
		r.engine.Use(middleware.CORSMiddleware())
	} else {
		r.engine.Use(middleware.CORSMiddleware())
	}

	// Request logging middleware
	if r.config.EnableRequestLogging {
		r.engine.Use(middleware.RequestLoggerMiddleware())
	} else {
		r.engine.Use(middleware.SimpleLoggerMiddleware())
	}

	// Error logging middleware
	r.engine.Use(middleware.ErrorLoggerMiddleware())
}

func (r *Router) setupRoutes() {
	// Health check endpoints
	healthHandler := handler.NewHealthHandler()
	r.engine.GET("/health", healthHandler.Health)
	r.engine.GET("/ping", healthHandler.Ping)

	// API routes
	api := r.engine.Group("/api")
	{
		// V1 API routes
		v1 := api.Group("/v1")
		{
			r.setupAnalyzerRoutes(v1)
		}
	}
}

func (r *Router) setupAnalyzerRoutes(rg *gin.RouterGroup) {
	analyzerHandler := handler.NewAnalyzerHandler(r.analyzerUsecase, r.filterUsecase)

	analyzer := rg.Group("/analyzer")
	{
		// Project analysis endpoints
		analyzer.POST("/scan", analyzerHandler.ScanProject)
		analyzer.GET("/projects/:projectId", analyzerHandler.GetProjectAnalysis)
		analyzer.DELETE("/projects/:projectId", analyzerHandler.DeleteProjectAnalysis)

		// API endpoints discovery
		analyzer.GET("/projects/:projectId/apis", analyzerHandler.ListAPIEndpoints)
		analyzer.GET("/projects/:projectId/apis/:apiId", analyzerHandler.GetAPIEndpoint)

		// Code node analysis
		analyzer.GET("/projects/:projectId/apis/:apiId/nodes", analyzerHandler.GetAPINodes)
		analyzer.GET("/projects/:projectId/nodes", analyzerHandler.GetAllNodes)
		analyzer.GET("/projects/:projectId/nodes/:nodeId", analyzerHandler.GetNode)

		// Filter and search endpoints
		analyzer.GET("/projects/:projectId/nodes/search", analyzerHandler.SearchNodes)
		analyzer.POST("/projects/:projectId/filters", analyzerHandler.ApplyFilters)

		// Statistics and metrics
		analyzer.GET("/projects/:projectId/stats", analyzerHandler.GetProjectStatistics)
		analyzer.GET("/projects/:projectId/apis/:apiId/stats", analyzerHandler.GetAPIStatistics)

		// Export endpoints
		analyzer.GET("/projects/:projectId/export", analyzerHandler.ExportAnalysis)
		analyzer.GET("/projects/:projectId/apis/:apiId/export", analyzerHandler.ExportAPIAnalysis)

		// Dependency analysis
		analyzer.GET("/projects/:projectId/dependencies", analyzerHandler.GetDependencyGraph)
		analyzer.GET("/projects/:projectId/apis/:apiId/dependencies", analyzerHandler.GetAPIDependencies)
	}
}
