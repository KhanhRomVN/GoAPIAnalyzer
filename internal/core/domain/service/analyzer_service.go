package service

import (
	"regexp"
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

// DiscoverAPIEndpoints analyzes the project to discover API endpoints
func (s *AnalyzerService) DiscoverAPIEndpoints(analysis *entity.ProjectAnalysis) error {
	s.logger.Info("Starting API endpoint discovery")

	var endpoints []*entity.APIEndpoint

	// Look for Gin router patterns
	ginEndpoints := s.discoverGinEndpoints(analysis)
	endpoints = append(endpoints, ginEndpoints...)

	// Look for HTTP handler patterns
	httpEndpoints := s.discoverHTTPEndpoints(analysis)
	endpoints = append(endpoints, httpEndpoints...)

	// Look for Gorilla Mux patterns
	muxEndpoints := s.discoverMuxEndpoints(analysis)
	endpoints = append(endpoints, muxEndpoints...)

	analysis.APIEndpoints = endpoints

	s.logger.WithField("endpoints_count", len(endpoints)).Info("API endpoint discovery completed")
	return nil
}

// discoverGinEndpoints discovers Gin framework endpoints
func (s *AnalyzerService) discoverGinEndpoints(analysis *entity.ProjectAnalysis) []*entity.APIEndpoint {
	var endpoints []*entity.APIEndpoint

	// Regex patterns for Gin route definitions
	ginPatterns := []*regexp.Regexp{
		regexp.MustCompile(`\.GET\s*\(\s*"([^"]+)"`),
		regexp.MustCompile(`\.POST\s*\(\s*"([^"]+)"`),
		regexp.MustCompile(`\.PUT\s*\(\s*"([^"]+)"`),
		regexp.MustCompile(`\.DELETE\s*\(\s*"([^"]+)"`),
		regexp.MustCompile(`\.PATCH\s*\(\s*"([^"]+)"`),
		regexp.MustCompile(`\.OPTIONS\s*\(\s*"([^"]+)"`),
		regexp.MustCompile(`\.HEAD\s*\(\s*"([^"]+)"`),
	}

	for filePath, fileInfo := range analysis.Files {
		content := fileInfo.Content

		// Check if file likely contains Gin routes
		if !strings.Contains(content, "gin") && !strings.Contains(content, "router") {
			continue
		}

		for _, pattern := range ginPatterns {
			matches := pattern.FindAllStringSubmatch(content, -1)
			for _, match := range matches {
				if len(match) > 1 {
					method := s.extractMethodFromPattern(pattern.String())
					path := match[1]

					endpoint := &entity.APIEndpoint{
						ID:      uuid.New().String(),
						Method:  method,
						Path:    path,
						File:    filePath,
						Package: fileInfo.PackageName,
					}

					// Try to find the handler function
					handler := s.findHandlerInContent(content, path)
					if handler != "" {
						endpoint.Handler = handler
					}

					// Find middleware
					middleware := s.findMiddlewareInContent(content, path)
					endpoint.Middleware = middleware

					endpoints = append(endpoints, endpoint)
				}
			}
		}
	}

	return endpoints
}

// discoverHTTPEndpoints discovers standard HTTP handler endpoints
func (s *AnalyzerService) discoverHTTPEndpoints(analysis *entity.ProjectAnalysis) []*entity.APIEndpoint {
	var endpoints []*entity.APIEndpoint

	// Pattern for http.HandleFunc calls
	handleFuncPattern := regexp.MustCompile(`http\.HandleFunc\s*\(\s*"([^"]+)"\s*,\s*([^)]+)\)`)

	for filePath, fileInfo := range analysis.Files {
		content := fileInfo.Content

		matches := handleFuncPattern.FindAllStringSubmatch(content, -1)
		for _, match := range matches {
			if len(match) > 2 {
				endpoint := &entity.APIEndpoint{
					ID:      uuid.New().String(),
					Method:  "ANY", // HTTP handlers typically handle multiple methods
					Path:    match[1],
					Handler: match[2],
					File:    filePath,
					Package: fileInfo.PackageName,
				}

				endpoints = append(endpoints, endpoint)
			}
		}
	}

	return endpoints
}

// discoverMuxEndpoints discovers Gorilla Mux router endpoints
func (s *AnalyzerService) discoverMuxEndpoints(analysis *entity.ProjectAnalysis) []*entity.APIEndpoint {
	var endpoints []*entity.APIEndpoint

	// Patterns for Mux route definitions
	muxPatterns := []*regexp.Regexp{
		regexp.MustCompile(`\.Methods\s*\(\s*"([^"]+)"\s*\)\.Path\s*\(\s*"([^"]+)"\s*\)`),
		regexp.MustCompile(`\.Path\s*\(\s*"([^"]+)"\s*\)\.Methods\s*\(\s*"([^"]+)"\s*\)`),
		regexp.MustCompile(`\.HandleFunc\s*\(\s*"([^"]+)"\s*,\s*([^)]+)\)\.Methods\s*\(\s*"([^"]+)"\s*\)`),
	}

	for filePath, fileInfo := range analysis.Files {
		content := fileInfo.Content

		// Check if file likely contains Mux routes
		if !strings.Contains(content, "mux") && !strings.Contains(content, "gorilla") {
			continue
		}

		for _, pattern := range muxPatterns {
			matches := pattern.FindAllStringSubmatch(content, -1)
			for _, match := range matches {
				if len(match) > 2 {
					var method, path, handler string

					// Different patterns have different group arrangements
					if strings.Contains(pattern.String(), "Methods.*Path") {
						method = match[1]
						path = match[2]
					} else if strings.Contains(pattern.String(), "Path.*Methods") {
						path = match[1]
						method = match[2]
					} else if strings.Contains(pattern.String(), "HandleFunc") {
						path = match[1]
						handler = match[2]
						if len(match) > 3 {
							method = match[3]
						}
					}

					endpoint := &entity.APIEndpoint{
						ID:      uuid.New().String(),
						Method:  method,
						Path:    path,
						Handler: handler,
						File:    filePath,
						Package: fileInfo.PackageName,
					}

					endpoints = append(endpoints, endpoint)
				}
			}
		}
	}

	return endpoints
}

// BuildDependencyGraph builds a dependency graph for the project
func (s *AnalyzerService) BuildDependencyGraph(analysis *entity.ProjectAnalysis) error {
	s.logger.Info("Building dependency graph")

	graph := &entity.DependencyGraph{
		Nodes:        make([]*entity.DependencyNode, 0),
		Dependencies: make([]*entity.Dependency, 0),
	}

	nodeMap := make(map[string]*entity.DependencyNode)

	// Create nodes for all functions, structs, interfaces, etc.
	for _, fileInfo := range analysis.Files {
		// Add function nodes
		for _, funcInfo := range fileInfo.Functions {
			nodeID := s.generateNodeID(fileInfo.PackageName, funcInfo.Name, "function")
			node := &entity.DependencyNode{
				ID:      nodeID,
				Name:    funcInfo.Name,
				Type:    "function",
				File:    fileInfo.Path,
				Package: fileInfo.PackageName,
			}
			graph.Nodes = append(graph.Nodes, node)
			nodeMap[nodeID] = node

			// Add dependencies based on function calls
			for _, call := range funcInfo.CallsTo {
				targetNodeID := s.findNodeIDForCall(call.Name, nodeMap, fileInfo.PackageName)
				if targetNodeID != "" && targetNodeID != nodeID {
					dependency := &entity.Dependency{
						From:     nodeID,
						To:       targetNodeID,
						Type:     "call",
						Strength: 5,
					}
					graph.Dependencies = append(graph.Dependencies, dependency)
				}
			}
		}

		// Add struct nodes
		for _, structInfo := range fileInfo.Structs {
			nodeID := s.generateNodeID(fileInfo.PackageName, structInfo.Name, "struct")
			node := &entity.DependencyNode{
				ID:      nodeID,
				Name:    structInfo.Name,
				Type:    "struct",
				File:    fileInfo.Path,
				Package: fileInfo.PackageName,
			}
			graph.Nodes = append(graph.Nodes, node)
			nodeMap[nodeID] = node
		}

		// Add interface nodes
		for _, interfaceInfo := range fileInfo.Interfaces {
			nodeID := s.generateNodeID(fileInfo.PackageName, interfaceInfo.Name, "interface")
			node := &entity.DependencyNode{
				ID:      nodeID,
				Name:    interfaceInfo.Name,
				Type:    "interface",
				File:    fileInfo.Path,
				Package: fileInfo.PackageName,
			}
			graph.Nodes = append(graph.Nodes, node)
			nodeMap[nodeID] = node
		}
	}

	// Add import dependencies
	for _, fileInfo := range analysis.Files {
		for _, importPath := range fileInfo.Imports {
			// Create dependency for imports
			if packageInfo, exists := analysis.Packages[importPath]; exists {
				for _, funcInfo := range fileInfo.Functions {
					fromNodeID := s.generateNodeID(fileInfo.PackageName, funcInfo.Name, "function")

					// Find functions in the imported package
					for importedFilePath := range analysis.Files {
						if strings.Contains(importedFilePath, packageInfo.Path) {
							importedFileInfo := analysis.Files[importedFilePath]
							for _, importedFuncInfo := range importedFileInfo.Functions {
								toNodeID := s.generateNodeID(importedFileInfo.PackageName, importedFuncInfo.Name, "function")

								// Check if the function actually uses something from the imported package
								for _, call := range funcInfo.CallsTo {
									if strings.Contains(call.Name, importedFuncInfo.Name) {
										dependency := &entity.Dependency{
											From:     fromNodeID,
											To:       toNodeID,
											Type:     "import",
											Strength: 3,
										}
										graph.Dependencies = append(graph.Dependencies, dependency)
									}
								}
							}
						}
					}
				}
			}
		}
	}

	analysis.DependencyGraph = graph

	s.logger.WithFields(map[string]interface{}{
		"nodes_count":        len(graph.Nodes),
		"dependencies_count": len(graph.Dependencies),
	}).Info("Dependency graph built successfully")

	return nil
}

// Helper methods

func (s *AnalyzerService) extractMethodFromPattern(pattern string) string {
	if strings.Contains(pattern, "GET") {
		return "GET"
	} else if strings.Contains(pattern, "POST") {
		return "POST"
	} else if strings.Contains(pattern, "PUT") {
		return "PUT"
	} else if strings.Contains(pattern, "DELETE") {
		return "DELETE"
	} else if strings.Contains(pattern, "PATCH") {
		return "PATCH"
	} else if strings.Contains(pattern, "OPTIONS") {
		return "OPTIONS"
	} else if strings.Contains(pattern, "HEAD") {
		return "HEAD"
	}
	return "UNKNOWN"
}

func (s *AnalyzerService) findHandlerInContent(content, path string) string {
	// Try to find handler function near the route definition
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		if strings.Contains(line, path) {
			// Look for handler in the same line
			handlerPattern := regexp.MustCompile(`([a-zA-Z_][a-zA-Z0-9_]*)\s*\(`)
			matches := handlerPattern.FindAllStringSubmatch(line, -1)
			if len(matches) > 0 && len(matches[0]) > 1 {
				return matches[0][1]
			}

			// Look in adjacent lines
			for j := max(0, i-2); j < min(len(lines), i+3); j++ {
				if j != i {
					handlerMatches := handlerPattern.FindAllStringSubmatch(lines[j], -1)
					if len(handlerMatches) > 0 && len(handlerMatches[0]) > 1 {
						return handlerMatches[0][1]
					}
				}
			}
		}
	}
	return ""
}

func (s *AnalyzerService) findMiddlewareInContent(content, path string) []string {
	var middleware []string

	// Look for common middleware patterns
	middlewarePatterns := []string{
		"middleware",
		"auth",
		"cors",
		"logger",
		"recovery",
	}

	lines := strings.Split(content, "\n")
	for i, line := range lines {
		if strings.Contains(line, path) {
			// Look for middleware in surrounding lines
			for j := max(0, i-5); j < min(len(lines), i+1); j++ {
				for _, pattern := range middlewarePatterns {
					if strings.Contains(strings.ToLower(lines[j]), pattern) {
						middleware = append(middleware, pattern)
					}
				}
			}
		}
	}

	return middleware
}

func (s *AnalyzerService) generateNodeID(packageName, name, nodeType string) string {
	return packageName + "." + name + "." + nodeType
}

func (s *AnalyzerService) findNodeIDForCall(callName string, nodeMap map[string]*entity.DependencyNode, currentPackage string) string {
	// Try exact match first
	for nodeID, node := range nodeMap {
		if node.Name == callName {
			return nodeID
		}
	}

	// Try with current package prefix
	candidateID := currentPackage + "." + callName + ".function"
	if _, exists := nodeMap[candidateID]; exists {
		return candidateID
	}

	// Try partial matches
	for nodeID, node := range nodeMap {
		if strings.Contains(callName, node.Name) || strings.Contains(node.Name, callName) {
			return nodeID
		}
	}

	return ""
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
