package service

import (
	"regexp"
	"sort"
	"strings"

	"goapianalyzer/internal/core/domain/entity"
	"goapianalyzer/internal/infrastructure/logger"

	"github.com/google/uuid"
)

type AnalyzerService struct {
	logger logger.Logger
}

func NewAnalyzerService() *AnalyzerService {
	return &AnalyzerService{
		logger: logger.GetLogger(),
	}
}

// RouterGroup represents a comprehensive router group analysis
type RouterGroup struct {
	VarName    string
	Path       string
	ParentVar  string
	Parent     *RouterGroup
	FullPath   string
	LineNumber int
	File       string
	Children   []*RouterGroup
}

// RouteCall represents a route method call with context
type RouteCall struct {
	Method      string
	Path        string
	VarName     string
	LineNumber  int
	Handler     string
	FullPath    string
	File        string
	Middlewares []string
}

// RouterContext holds the complete routing analysis
type RouterContext struct {
	Groups    map[string]*RouterGroup
	Routes    []*RouteCall
	Variables map[string]string // variable assignments
	File      string
}

// DiscoverAPIEndpoints with enhanced path resolution
func (s *AnalyzerService) DiscoverAPIEndpoints(analysis *entity.ProjectAnalysis) error {
	s.logger.Info("Starting enhanced API endpoint discovery")

	// Use a map to track unique endpoints by method and path
	endpointMap := make(map[string]*entity.APIEndpoint)

	// Analyze each file for routing patterns
	for filePath, fileInfo := range analysis.Files {
		if !s.isRouterFile(fileInfo.Content) {
			continue
		}

		context := s.analyzeRouterContext(fileInfo.Content, filePath)
		fileEndpoints := s.extractEndpointsFromContext(context)
		for _, endpoint := range fileEndpoints {
			key := endpoint.Method + ":" + endpoint.Path
			if _, exists := endpointMap[key]; !exists {
				endpointMap[key] = endpoint
			} else {
				// Log duplicate found
				s.logger.WithFields(map[string]interface{}{
					"method": endpoint.Method,
					"path":   endpoint.Path,
					"file":   endpoint.File,
				}).Warn("Duplicate endpoint found")
			}
		}
	}

	// Cross-file analysis for router setup patterns
	crossFileEndpoints := s.analyzeCrossFileRouting(analysis)
	for _, endpoint := range crossFileEndpoints {
		key := endpoint.Method + ":" + endpoint.Path
		if _, exists := endpointMap[key]; !exists {
			endpointMap[key] = endpoint
		} else {
			s.logger.WithFields(map[string]interface{}{
				"method": endpoint.Method,
				"path":   endpoint.Path,
				"file":   endpoint.File,
			}).Warn("Duplicate endpoint found from cross-file analysis")
		}
	}

	// Convert map to slice
	var endpoints []*entity.APIEndpoint
	for _, endpoint := range endpointMap {
		endpoints = append(endpoints, endpoint)
	}

	// Sort endpoints by path
	sort.Slice(endpoints, func(i, j int) bool {
		return endpoints[i].Path < endpoints[j].Path
	})

	analysis.APIEndpoints = endpoints

	s.logger.WithField("endpoints_count", len(endpoints)).Info("Enhanced API endpoint discovery completed")
	return nil
}

// analyzeRouterContext performs comprehensive analysis of a router file
func (s *AnalyzerService) analyzeRouterContext(content, filePath string) *RouterContext {
	ctx := &RouterContext{
		Groups:    make(map[string]*RouterGroup),
		Routes:    make([]*RouteCall, 0),
		Variables: make(map[string]string),
		File:      filePath,
	}

	lines := strings.Split(content, "\n")

	// Pass 1: Find variable assignments and router creations
	s.analyzeVariableAssignments(lines, ctx)

	// Pass 2: Find router groups with hierarchy
	s.analyzeRouterGroups(lines, ctx)

	// Pass 3: Find route method calls
	s.analyzeRouteCalls(lines, ctx)

	// Pass 4: Resolve full paths
	s.resolveFullPaths(ctx)

	return ctx
}

// analyzeVariableAssignments finds variable assignments like r := gin.New()
func (s *AnalyzerService) analyzeVariableAssignments(lines []string, ctx *RouterContext) {
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`(\w+)\s*:=\s*gin\.New\(\)`),
		regexp.MustCompile(`(\w+)\s*:=\s*gin\.Default\(\)`),
		regexp.MustCompile(`(\w+)\s*:=\s*[\w\.]+\.Engine`),
		regexp.MustCompile(`(\w+)\s*:=\s*[\w\.]+\.Group\s*\(\s*["']([^"']+)["']\s*\)`),
	}

	for _, line := range lines { // Đã xóa lineNum không sử dụng
		for _, pattern := range patterns {
			if matches := pattern.FindStringSubmatch(line); matches != nil {
				if len(matches) >= 2 {
					varName := matches[1]
					if len(matches) >= 3 {
						// This is a group assignment
						ctx.Variables[varName] = matches[2]
					} else {
						// This is an engine assignment
						ctx.Variables[varName] = "/"
					}
				}
			}
		}
	}
}

// analyzeRouterGroups finds router group definitions with full hierarchy
func (s *AnalyzerService) analyzeRouterGroups(lines []string, ctx *RouterContext) {
	// Pattern for various group assignment styles
	patterns := []*regexp.Regexp{
		// api := r.Group("/v1")
		regexp.MustCompile(`(\w+)\s*:=\s*(\w+)\.Group\s*\(\s*["']([^"']+)["']\s*\)`),
		// user := api.Group("/user").Use(middleware)
		regexp.MustCompile(`(\w+)\s*:=\s*(\w+)\.Group\s*\(\s*["']([^"']+)["']\s*\)\.Use\(`),
	}

	for lineNum, line := range lines {
		for _, pattern := range patterns {
			matches := pattern.FindStringSubmatch(line)
			if len(matches) >= 4 {
				varName := matches[1]
				parentVar := matches[2]
				path := matches[3]

				group := &RouterGroup{
					VarName:    varName,
					Path:       path,
					ParentVar:  parentVar,
					LineNumber: lineNum + 1,
					File:       ctx.File,
					Children:   make([]*RouterGroup, 0),
				}

				// Link to parent if exists
				if parentGroup, exists := ctx.Groups[parentVar]; exists {
					group.Parent = parentGroup
					parentGroup.Children = append(parentGroup.Children, group)
				}

				ctx.Groups[varName] = group
			}
		}
	}
}

// analyzeRouteCalls finds HTTP method calls with enhanced pattern matching
func (s *AnalyzerService) analyzeRouteCalls(lines []string, ctx *RouterContext) {
	patterns := []*regexp.Regexp{
		// Standard: router.GET("/path", handler)
		regexp.MustCompile(`(\w+)\.(GET|POST|PUT|DELETE|PATCH|OPTIONS|HEAD|ANY)\s*\(\s*["']([^"']+)["']\s*,\s*([^)]+)\)`),
		// With middleware: router.GET("/path", middleware, handler)
		regexp.MustCompile(`(\w+)\.(GET|POST|PUT|DELETE|PATCH|OPTIONS|HEAD|ANY)\s*\(\s*["']([^"']+)["']\s*,\s*([^,)]+(?:,\s*[^)]+)*)\)`),
		// Gorilla mux style: router.HandleFunc("/path", handler).Methods("GET")
		regexp.MustCompile(`(\w+)\.HandleFunc\s*\(\s*["']([^"']+)["']\s*,\s*([^)]+)\)\.Methods\s*\(\s*["']([^"']+)["']\s*\)`),
	}

	for lineNum, line := range lines {
		for _, pattern := range patterns {
			if matches := pattern.FindStringSubmatch(line); matches != nil {
				route := s.parseRouteCall(matches, lineNum, ctx.File)
				if route != nil {
					ctx.Routes = append(ctx.Routes, route)
				}
			}
		}
	}
}

// parseRouteCall parses individual route call matches
func (s *AnalyzerService) parseRouteCall(matches []string, lineNum int, file string) *RouteCall {
	if len(matches) < 4 {
		return nil
	}

	route := &RouteCall{
		VarName:    matches[1],
		LineNumber: lineNum + 1,
		File:       file,
	}

	// Handle different match patterns
	if strings.Contains(matches[0], "HandleFunc") {
		// Gorilla mux pattern
		route.Path = matches[2]
		route.Handler = matches[3]
		route.Method = matches[4]
	} else {
		// Gin pattern
		route.Method = matches[2]
		route.Path = matches[3]
		route.Handler = matches[4]

		// Parse middlewares if present
		if strings.Contains(route.Handler, ",") {
			parts := strings.Split(route.Handler, ",")
			route.Handler = strings.TrimSpace(parts[len(parts)-1])
			for i := 0; i < len(parts)-1; i++ {
				middleware := strings.TrimSpace(parts[i])
				route.Middlewares = append(route.Middlewares, middleware)
			}
		}
	}

	return route
}

// resolveFullPaths calculates complete paths for all routes
func (s *AnalyzerService) resolveFullPaths(ctx *RouterContext) {
	// First, calculate full paths for all groups
	s.calculateGroupFullPaths(ctx.Groups)

	// Then, resolve full paths for all routes
	for _, route := range ctx.Routes {
		route.FullPath = s.buildRouteFullPath(route, ctx)
	}
}

// calculateGroupFullPaths builds hierarchical paths for router groups
func (s *AnalyzerService) calculateGroupFullPaths(groups map[string]*RouterGroup) {
	// Topological sort to handle dependencies
	sorted := s.topologicalSortGroups(groups)

	for _, group := range sorted {
		if group.Parent == nil {
			// Root group
			group.FullPath = s.normalizePath(group.Path)
		} else {
			// Child group - combine with parent path
			group.FullPath = s.combinePaths(group.Parent.FullPath, group.Path)
		}
	}
}

// topologicalSortGroups sorts groups by dependency order
func (s *AnalyzerService) topologicalSortGroups(groups map[string]*RouterGroup) []*RouterGroup {
	var sorted []*RouterGroup
	visited := make(map[string]bool)

	var visit func(*RouterGroup)
	visit = func(group *RouterGroup) {
		if visited[group.VarName] {
			return
		}

		visited[group.VarName] = true

		// Visit parent first
		if group.Parent != nil && !visited[group.Parent.VarName] {
			visit(group.Parent)
		}

		sorted = append(sorted, group)
	}

	for _, group := range groups {
		if !visited[group.VarName] {
			visit(group)
		}
	}

	return sorted
}

// buildRouteFullPath constructs the complete path for a route
func (s *AnalyzerService) buildRouteFullPath(route *RouteCall, ctx *RouterContext) string {
	// Find the group this route belongs to
	if group, exists := ctx.Groups[route.VarName]; exists {
		return s.combinePaths(group.FullPath, route.Path)
	}

	// Check if it's a direct engine route
	if basePath, exists := ctx.Variables[route.VarName]; exists {
		return s.combinePaths(basePath, route.Path)
	}

	// Fallback to the route path itself
	return s.normalizePath(route.Path)
}

// combinePaths safely combines two path segments
func (s *AnalyzerService) combinePaths(basePath, routePath string) string {
	basePath = s.normalizePath(basePath)
	routePath = s.normalizePath(routePath)

	if basePath == "/" {
		return routePath
	}

	if routePath == "/" {
		return basePath
	}

	// Remove trailing slash from base and leading slash from route
	basePath = strings.TrimSuffix(basePath, "/")
	routePath = strings.TrimPrefix(routePath, "/")

	return basePath + "/" + routePath
}

// normalizePath ensures consistent path formatting
func (s *AnalyzerService) normalizePath(path string) string {
	if path == "" {
		return "/"
	}

	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	// Clean up multiple slashes
	for strings.Contains(path, "//") {
		path = strings.ReplaceAll(path, "//", "/")
	}

	return path
}

// extractEndpointsFromContext converts analyzed context to API endpoints
func (s *AnalyzerService) extractEndpointsFromContext(ctx *RouterContext) []*entity.APIEndpoint {
	var endpoints []*entity.APIEndpoint

	for _, route := range ctx.Routes {
		endpoint := &entity.APIEndpoint{
			ID:     uuid.New().String(),
			Method: route.Method,
			Path:   route.FullPath,
			File:   route.File,
		}

		endpoints = append(endpoints, endpoint)
	}

	return endpoints
}

// analyzeCrossFileRouting handles route setup patterns across multiple files
func (s *AnalyzerService) analyzeCrossFileRouting(analysis *entity.ProjectAnalysis) []*entity.APIEndpoint {
	var endpoints []*entity.APIEndpoint

	// Look for main.go or router setup files
	for filePath, fileInfo := range analysis.Files {
		if s.isMainOrSetupFile(filePath) {
			setupEndpoints := s.analyzeRouterSetupFile(fileInfo.Content, filePath, analysis)
			endpoints = append(endpoints, setupEndpoints...)
		}
	}

	return endpoints
}

// isMainOrSetupFile checks if the file is likely a main or router setup file
func (s *AnalyzerService) isMainOrSetupFile(filePath string) bool {
	fileName := strings.ToLower(filePath)
	return strings.Contains(fileName, "main.go") ||
		strings.Contains(fileName, "router.go") ||
		strings.Contains(fileName, "routes.go") ||
		strings.Contains(fileName, "setup")
}

// analyzeRouterSetupFile analyzes router setup patterns in main files
func (s *AnalyzerService) analyzeRouterSetupFile(content string, _ string, _ *entity.ProjectAnalysis) []*entity.APIEndpoint {
	var endpoints []*entity.APIEndpoint

	// Look for function calls that setup routes
	setupPattern := regexp.MustCompile(`(\w+)\.Setup\w*Routes?\s*\([^)]*\)`)
	matches := setupPattern.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		if len(match) >= 2 {
			// This indicates there's a router setup, but we need to find the actual routes
			// in the router files themselves
			continue
		}
	}

	return endpoints
}

// isRouterFile checks if file contains routing logic (enhanced version)
func (s *AnalyzerService) isRouterFile(content string) bool {
	indicators := []string{
		"gin.Engine", "gin.RouterGroup", ".Group(",
		".GET(", ".POST(", ".PUT(", ".DELETE(", ".PATCH(", ".OPTIONS(", ".HEAD(",
		"router", "Route", "HandleFunc", "mux.Router",
		"setupRoutes", "SetupRoutes", "routes.go",
	}

	content = strings.ToLower(content)
	for _, indicator := range indicators {
		if strings.Contains(content, strings.ToLower(indicator)) {
			return true
		}
	}

	return false
}

// GetRouteStatistics provides detailed statistics about discovered routes
func (s *AnalyzerService) GetRouteStatistics(endpoints []*entity.APIEndpoint) map[string]interface{} {
	stats := make(map[string]interface{})

	methodCounts := make(map[string]int)
	pathCounts := make(map[string]int)
	fileCounts := make(map[string]int)

	for _, endpoint := range endpoints {
		methodCounts[endpoint.Method]++
		fileCounts[endpoint.File]++

		// Count path patterns
		pathSegments := strings.Split(endpoint.Path, "/")
		if len(pathSegments) > 1 {
			pathCounts["/"+pathSegments[1]]++
		}
	}

	stats["total_endpoints"] = len(endpoints)
	stats["methods"] = methodCounts
	stats["top_level_paths"] = pathCounts
	stats["files"] = fileCounts

	return stats
}
