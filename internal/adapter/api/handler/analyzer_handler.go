package handler

import (
	"net/http"
	"strconv"

	"goapianalyzer/internal/core/domain/entity"
	"goapianalyzer/internal/core/usecase"
	"goapianalyzer/internal/infrastructure/logger"
	"goapianalyzer/pkg/errors"

	"github.com/gin-gonic/gin"
)

type AnalyzerHandler struct {
	analyzerUsecase *usecase.AnalyzerUsecase
	filterUsecase   *usecase.FilterUsecase
	logger          logger.Logger
}

type ScanProjectRequest struct {
	ProjectPath     string   `json:"project_path" binding:"required"`
	BlacklistFiles  []string `json:"blacklist_files,omitempty"`
	BlacklistDirs   []string `json:"blacklist_dirs,omitempty"`
	WhitelistFiles  []string `json:"whitelist_files,omitempty"`
	WhitelistDirs   []string `json:"whitelist_dirs,omitempty"`
	IncludeVendor   bool     `json:"include_vendor"`
	IncludeTestFile bool     `json:"include_test_file"`
}

type FilterRequest struct {
	NodeTypes      []string `json:"node_types,omitempty"`
	FileExtensions []string `json:"file_extensions,omitempty"`
	PackageNames   []string `json:"package_names,omitempty"`
	FunctionNames  []string `json:"function_names,omitempty"`
	BlacklistFiles []string `json:"blacklist_files,omitempty"`
	BlacklistDirs  []string `json:"blacklist_dirs,omitempty"`
	MinComplexity  *int     `json:"min_complexity,omitempty"`
	MaxComplexity  *int     `json:"max_complexity,omitempty"`
}

type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

func NewAnalyzerHandler(analyzerUsecase *usecase.AnalyzerUsecase, filterUsecase *usecase.FilterUsecase) *AnalyzerHandler {
	return &AnalyzerHandler{
		analyzerUsecase: analyzerUsecase,
		filterUsecase:   filterUsecase,
		logger:          logger.GetLogger(),
	}
}

// ScanProject scans a Go project and analyzes its structure
func (h *AnalyzerHandler) ScanProject(c *gin.Context) {
	var req ScanProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithFields(map[string]interface{}{
			"error": err.Error(),
		}).Error("Invalid request body for project scan")

		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Invalid request format: " + err.Error(),
		})
		return
	}

	h.logger.WithFields(map[string]interface{}{
		"project_path":    req.ProjectPath,
		"blacklist_files": req.BlacklistFiles,
		"blacklist_dirs":  req.BlacklistDirs,
	}).Info("Starting project scan")

	projectAnalysis, err := h.analyzerUsecase.AnalyzeProject(req.ProjectPath, &entity.AnalysisConfig{
		BlacklistFiles:  req.BlacklistFiles,
		BlacklistDirs:   req.BlacklistDirs,
		WhitelistFiles:  req.WhitelistFiles,
		WhitelistDirs:   req.WhitelistDirs,
		IncludeVendor:   req.IncludeVendor,
		IncludeTestFile: req.IncludeTestFile,
	})

	if err != nil {
		h.logger.WithFields(map[string]interface{}{
			"error":        err.Error(),
			"project_path": req.ProjectPath,
		}).Error("Failed to analyze project")

		status := http.StatusInternalServerError
		if errors.IsValidationError(err) {
			status = http.StatusBadRequest
		}

		c.JSON(status, APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "Project analyzed successfully",
		Data:    projectAnalysis,
	})
}

// GetProjectAnalysis retrieves a stored project analysis
func (h *AnalyzerHandler) GetProjectAnalysis(c *gin.Context) {
	projectID := c.Param("projectId")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Project ID is required",
		})
		return
	}

	analysis, err := h.analyzerUsecase.GetProjectAnalysis(projectID)
	if err != nil {
		status := http.StatusNotFound
		if !errors.IsNotFoundError(err) {
			status = http.StatusInternalServerError
		}

		c.JSON(status, APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    analysis,
	})
}

// DeleteProjectAnalysis removes a stored project analysis
func (h *AnalyzerHandler) DeleteProjectAnalysis(c *gin.Context) {
	projectID := c.Param("projectId")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Project ID is required",
		})
		return
	}

	err := h.analyzerUsecase.DeleteProjectAnalysis(projectID)
	if err != nil {
		status := http.StatusNotFound
		if !errors.IsNotFoundError(err) {
			status = http.StatusInternalServerError
		}

		c.JSON(status, APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "Project analysis deleted successfully",
	})
}

// ListAPIEndpoints lists all discovered API endpoints in the project
func (h *AnalyzerHandler) ListAPIEndpoints(c *gin.Context) {
	projectID := c.Param("projectId")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Project ID is required",
		})
		return
	}

	endpoints, err := h.analyzerUsecase.GetAPIEndpoints(projectID)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.IsNotFoundError(err) {
			status = http.StatusNotFound
		}

		c.JSON(status, APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    endpoints,
	})
}

// GetAPIEndpoint retrieves details of a specific API endpoint
func (h *AnalyzerHandler) GetAPIEndpoint(c *gin.Context) {
	projectID := c.Param("projectId")
	apiID := c.Param("apiId")

	if projectID == "" || apiID == "" {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Project ID and API ID are required",
		})
		return
	}

	endpoint, err := h.analyzerUsecase.GetAPIEndpoint(projectID, apiID)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.IsNotFoundError(err) {
			status = http.StatusNotFound
		}

		c.JSON(status, APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    endpoint,
	})
}

// GetAPINodes retrieves all code nodes related to a specific API endpoint
func (h *AnalyzerHandler) GetAPINodes(c *gin.Context) {
	projectID := c.Param("projectId")
	apiID := c.Param("apiId")

	if projectID == "" || apiID == "" {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Project ID and API ID are required",
		})
		return
	}

	nodes, err := h.analyzerUsecase.GetAPINodes(projectID, apiID)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.IsNotFoundError(err) {
			status = http.StatusNotFound
		}

		c.JSON(status, APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    nodes,
	})
}

// GetAllNodes retrieves all code nodes in the project
func (h *AnalyzerHandler) GetAllNodes(c *gin.Context) {
	projectID := c.Param("projectId")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Project ID is required",
		})
		return
	}

	// Parse query parameters for pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	nodeType := c.Query("type")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 1000 {
		limit = 50
	}

	nodes, total, err := h.analyzerUsecase.GetAllNodes(projectID, page, limit, nodeType)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.IsNotFoundError(err) {
			status = http.StatusNotFound
		}

		c.JSON(status, APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"nodes":       nodes,
			"total":       total,
			"page":        page,
			"limit":       limit,
			"total_pages": (total + int64(limit) - 1) / int64(limit),
		},
	})
}

// GetNode retrieves details of a specific code node
func (h *AnalyzerHandler) GetNode(c *gin.Context) {
	projectID := c.Param("projectId")
	nodeID := c.Param("nodeId")

	if projectID == "" || nodeID == "" {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Project ID and Node ID are required",
		})
		return
	}

	node, err := h.analyzerUsecase.GetNode(projectID, nodeID)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.IsNotFoundError(err) {
			status = http.StatusNotFound
		}

		c.JSON(status, APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    node,
	})
}

// SearchNodes searches for nodes based on query parameters
func (h *AnalyzerHandler) SearchNodes(c *gin.Context) {
	projectID := c.Param("projectId")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Project ID is required",
		})
		return
	}

	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Search query is required",
		})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	nodes, total, err := h.analyzerUsecase.SearchNodes(projectID, query, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"nodes":       nodes,
			"total":       total,
			"page":        page,
			"limit":       limit,
			"query":       query,
			"total_pages": (total + int64(limit) - 1) / int64(limit),
		},
	})
}

// ApplyFilters applies filters to the project nodes
func (h *AnalyzerHandler) ApplyFilters(c *gin.Context) {
	projectID := c.Param("projectId")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Project ID is required",
		})
		return
	}

	var req FilterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Invalid filter request: " + err.Error(),
		})
		return
	}

	filters := &entity.FilterConfig{
		NodeTypes:      req.NodeTypes,
		FileExtensions: req.FileExtensions,
		PackageNames:   req.PackageNames,
		FunctionNames:  req.FunctionNames,
		BlacklistFiles: req.BlacklistFiles,
		BlacklistDirs:  req.BlacklistDirs,
		MinComplexity:  req.MinComplexity,
		MaxComplexity:  req.MaxComplexity,
	}

	filteredNodes, err := h.filterUsecase.ApplyFilters(projectID, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    filteredNodes,
	})
}

// GetProjectStatistics retrieves project statistics
func (h *AnalyzerHandler) GetProjectStatistics(c *gin.Context) {
	projectID := c.Param("projectId")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Project ID is required",
		})
		return
	}

	stats, err := h.analyzerUsecase.GetProjectStatistics(projectID)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.IsNotFoundError(err) {
			status = http.StatusNotFound
		}

		c.JSON(status, APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    stats,
	})
}

// GetAPIStatistics retrieves statistics for a specific API endpoint
func (h *AnalyzerHandler) GetAPIStatistics(c *gin.Context) {
	projectID := c.Param("projectId")
	apiID := c.Param("apiId")

	if projectID == "" || apiID == "" {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Project ID and API ID are required",
		})
		return
	}

	stats, err := h.analyzerUsecase.GetAPIStatistics(projectID, apiID)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.IsNotFoundError(err) {
			status = http.StatusNotFound
		}

		c.JSON(status, APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    stats,
	})
}

// ExportAnalysis exports the complete project analysis
func (h *AnalyzerHandler) ExportAnalysis(c *gin.Context) {
	projectID := c.Param("projectId")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Project ID is required",
		})
		return
	}

	format := c.DefaultQuery("format", "json")

	exported, err := h.analyzerUsecase.ExportAnalysis(projectID, format)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.IsNotFoundError(err) {
			status = http.StatusNotFound
		}

		c.JSON(status, APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	// Set appropriate headers for download
	filename := "project_analysis." + format
	c.Header("Content-Disposition", "attachment; filename="+filename)

	switch format {
	case "json":
		c.Header("Content-Type", "application/json")
	case "yaml":
		c.Header("Content-Type", "application/x-yaml")
	case "xml":
		c.Header("Content-Type", "application/xml")
	default:
		c.Header("Content-Type", "application/octet-stream")
	}

	c.String(http.StatusOK, exported)
}

// ExportAPIAnalysis exports analysis for a specific API endpoint
func (h *AnalyzerHandler) ExportAPIAnalysis(c *gin.Context) {
	projectID := c.Param("projectId")
	apiID := c.Param("apiId")

	if projectID == "" || apiID == "" {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Project ID and API ID are required",
		})
		return
	}

	format := c.DefaultQuery("format", "json")

	exported, err := h.analyzerUsecase.ExportAPIAnalysis(projectID, apiID, format)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.IsNotFoundError(err) {
			status = http.StatusNotFound
		}

		c.JSON(status, APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	// Set appropriate headers for download
	filename := "api_analysis_" + apiID + "." + format
	c.Header("Content-Disposition", "attachment; filename="+filename)

	switch format {
	case "json":
		c.Header("Content-Type", "application/json")
	case "yaml":
		c.Header("Content-Type", "application/x-yaml")
	case "xml":
		c.Header("Content-Type", "application/xml")
	default:
		c.Header("Content-Type", "application/octet-stream")
	}

	c.String(http.StatusOK, exported)
}

// GetDependencyGraph retrieves the dependency graph for the project
func (h *AnalyzerHandler) GetDependencyGraph(c *gin.Context) {
	projectID := c.Param("projectId")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Project ID is required",
		})
		return
	}

	graph, err := h.analyzerUsecase.GetDependencyGraph(projectID)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.IsNotFoundError(err) {
			status = http.StatusNotFound
		}

		c.JSON(status, APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    graph,
	})
}

// GetAPIDependencies retrieves dependencies for a specific API endpoint
func (h *AnalyzerHandler) GetAPIDependencies(c *gin.Context) {
	projectID := c.Param("projectId")
	apiID := c.Param("apiId")

	if projectID == "" || apiID == "" {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Project ID and API ID are required",
		})
		return
	}

	dependencies, err := h.analyzerUsecase.GetAPIDependencies(projectID, apiID)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.IsNotFoundError(err) {
			status = http.StatusNotFound
		}

		c.JSON(status, APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    dependencies,
	})
}
