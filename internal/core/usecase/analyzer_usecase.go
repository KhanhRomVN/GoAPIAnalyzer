package usecase

import (
	"encoding/json"
	"encoding/xml"
	"strings"
	"time"

	"goapianalyzer/internal/adapter/parser"
	"goapianalyzer/internal/core/domain/entity"
	"goapianalyzer/internal/core/domain/repository"
	"goapianalyzer/internal/core/domain/service"
	"goapianalyzer/internal/infrastructure/logger"
	"goapianalyzer/pkg/errors"

	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
)

type AnalyzerUsecase struct {
	repo            repository.AnalysisRepository
	analyzerService *service.AnalyzerService
	logger          logger.Logger
}

func NewAnalyzerUsecase(
	repo repository.AnalysisRepository,
	analyzerService *service.AnalyzerService,
) *AnalyzerUsecase {
	return &AnalyzerUsecase{
		repo:            repo,
		analyzerService: analyzerService,
		logger:          logger.GetLogger(),
	}
}

func (u *AnalyzerUsecase) AnalyzeProject(projectPath string, config *entity.AnalysisConfig) (*entity.ProjectAnalysis, error) {
	u.logger.WithFields(map[string]interface{}{
		"project_path": projectPath,
		"config":       config,
	}).Info("Starting project analysis")

	// Create file scanner with configuration
	scanConfig := &parser.ScanConfig{
		BlacklistFiles:  config.BlacklistFiles,
		BlacklistDirs:   config.BlacklistDirs,
		WhitelistFiles:  config.WhitelistFiles,
		WhitelistDirs:   config.WhitelistDirs,
		IncludeVendor:   config.IncludeVendor,
		IncludeTestFile: config.IncludeTestFile,
	}

	fileScanner := parser.NewFileScanner(scanConfig)

	// Scan project files
	projectAnalysis, err := fileScanner.ScanProject(projectPath)
	if err != nil {
		u.logger.WithError(err).Error("Failed to scan project files")
		return nil, err
	}

	// Parse AST details for each file
	astParser := parser.NewASTParser(fileScanner.GetFileSet())
	for _, fileInfo := range projectAnalysis.Files {
		if err := astParser.ParseFileDetails(fileInfo); err != nil {
			u.logger.WithError(err).WithField("file", fileInfo.Path).Warn("Failed to parse file details")
			continue
		}
	}

	// Generate unique project ID
	projectAnalysis.ID = uuid.New().String()
	projectAnalysis.CreatedAt = time.Now().UTC()

	// Use analyzer service to discover API endpoints and build dependency graph
	if err := u.analyzerService.DiscoverAPIEndpoints(projectAnalysis); err != nil {
		u.logger.WithError(err).Warn("Failed to discover API endpoints")
	}

	if err := u.analyzerService.BuildDependencyGraph(projectAnalysis); err != nil {
		u.logger.WithError(err).Warn("Failed to build dependency graph")
	}

	// Generate code nodes from the analysis
	codeNodes := u.generateCodeNodes(projectAnalysis)

	// Store project analysis
	if err := u.repo.StoreProjectAnalysis(projectAnalysis); err != nil {
		u.logger.WithError(err).Error("Failed to store project analysis")
		return nil, err
	}

	// Store all code nodes
	for _, node := range codeNodes {
		if err := u.repo.StoreCodeNode(projectAnalysis.ID, node); err != nil {
			u.logger.WithError(err).WithField("node_id", node.ID).Warn("Failed to store code node")
		}
	}

	// Store API endpoints
	for _, endpoint := range projectAnalysis.APIEndpoints {
		if err := u.repo.StoreAPIEndpoint(projectAnalysis.ID, endpoint); err != nil {
			u.logger.WithError(err).WithField("endpoint_id", endpoint.ID).Warn("Failed to store API endpoint")
		}
	}

	u.logger.WithFields(map[string]interface{}{
		"project_id":  projectAnalysis.ID,
		"files_count": len(projectAnalysis.Files),
		"nodes_count": len(codeNodes),
		"apis_count":  len(projectAnalysis.APIEndpoints),
	}).Info("Project analysis completed successfully")

	return projectAnalysis, nil
}

func (u *AnalyzerUsecase) generateCodeNodes(analysis *entity.ProjectAnalysis) []*entity.CodeNode {
	var nodes []*entity.CodeNode

	for filePath, fileInfo := range analysis.Files {
		// Generate nodes for functions
		for _, funcInfo := range fileInfo.Functions {
			node := &entity.CodeNode{
				ID:       uuid.New().String(),
				Name:     funcInfo.Name,
				Type:     "function",
				File:     filePath,
				Package:  fileInfo.PackageName,
				Body:     funcInfo.Body,
				Position: funcInfo.Position,
				Metadata: map[string]interface{}{
					"receiver":   funcInfo.Receiver,
					"is_method":  funcInfo.IsMethod,
					"parameters": funcInfo.Parameters,
					"returns":    funcInfo.Returns,
					"calls_to":   funcInfo.CallsTo,
					"used_types": funcInfo.UsedTypes,
				},
			}
			nodes = append(nodes, node)
		}

		// Generate nodes for structs
		for _, structInfo := range fileInfo.Structs {
			node := &entity.CodeNode{
				ID:      uuid.New().String(),
				Name:    structInfo.Name,
				Type:    "struct",
				File:    filePath,
				Package: fileInfo.PackageName,
				Body:    structInfo.Body,
				Metadata: map[string]interface{}{
					"fields": structInfo.Fields,
				},
			}
			nodes = append(nodes, node)
		}

		// Generate nodes for interfaces
		for _, interfaceInfo := range fileInfo.Interfaces {
			node := &entity.CodeNode{
				ID:      uuid.New().String(),
				Name:    interfaceInfo.Name,
				Type:    "interface",
				File:    filePath,
				Package: fileInfo.PackageName,
				Body:    interfaceInfo.Body,
				Metadata: map[string]interface{}{
					"methods": interfaceInfo.Methods,
				},
			}
			nodes = append(nodes, node)
		}

		// Generate nodes for types
		for _, typeInfo := range fileInfo.Types {
			node := &entity.CodeNode{
				ID:      uuid.New().String(),
				Name:    typeInfo.Name,
				Type:    "type",
				File:    filePath,
				Package: fileInfo.PackageName,
				Body:    typeInfo.Body,
				Metadata: map[string]interface{}{
					"type_definition": typeInfo.Type,
				},
			}
			nodes = append(nodes, node)
		}

		// Generate nodes for variables
		for _, varInfo := range fileInfo.Variables {
			node := &entity.CodeNode{
				ID:      uuid.New().String(),
				Name:    varInfo.Name,
				Type:    "variable",
				File:    filePath,
				Package: fileInfo.PackageName,
				Body:    varInfo.Body,
				Metadata: map[string]interface{}{
					"var_type": varInfo.Type,
					"value":    varInfo.Value,
				},
			}
			nodes = append(nodes, node)
		}

		// Generate nodes for constants
		for _, constInfo := range fileInfo.Constants {
			node := &entity.CodeNode{
				ID:      uuid.New().String(),
				Name:    constInfo.Name,
				Type:    "constant",
				File:    filePath,
				Package: fileInfo.PackageName,
				Body:    constInfo.Body,
				Metadata: map[string]interface{}{
					"const_type": constInfo.Type,
					"value":      constInfo.Value,
				},
			}
			nodes = append(nodes, node)
		}
	}

	return nodes
}

func (u *AnalyzerUsecase) GetProjectAnalysis(projectID string) (*entity.ProjectAnalysis, error) {
	return u.repo.GetProjectAnalysis(projectID)
}

func (u *AnalyzerUsecase) DeleteProjectAnalysis(projectID string) error {
	return u.repo.DeleteProjectAnalysis(projectID)
}

func (u *AnalyzerUsecase) GetAPIEndpoints(projectID string) ([]*entity.APIEndpoint, error) {
	return u.repo.GetAPIEndpoints(projectID)
}

func (u *AnalyzerUsecase) GetAPIEndpoint(projectID, apiID string) (*entity.APIEndpoint, error) {
	return u.repo.GetAPIEndpoint(projectID, apiID)
}

func (u *AnalyzerUsecase) GetAPINodes(projectID, apiID string) ([]*entity.CodeNode, error) {
	return u.repo.GetAPINodes(projectID, apiID)
}

func (u *AnalyzerUsecase) GetAllNodes(projectID string, page, limit int, nodeType string) ([]*entity.CodeNode, int64, error) {
	return u.repo.GetAllNodes(projectID, page, limit, nodeType)
}

func (u *AnalyzerUsecase) GetNode(projectID, nodeID string) (*entity.CodeNode, error) {
	return u.repo.GetCodeNode(projectID, nodeID)
}

func (u *AnalyzerUsecase) SearchNodes(projectID, query string, page, limit int) ([]*entity.CodeNode, int64, error) {
	return u.repo.SearchNodes(projectID, query, page, limit)
}

func (u *AnalyzerUsecase) GetProjectStatistics(projectID string) (*entity.ProjectStatistics, error) {
	analysis, err := u.repo.GetProjectAnalysis(projectID)
	if err != nil {
		return nil, err
	}

	stats := &entity.ProjectStatistics{
		ProjectID:     projectID,
		TotalFiles:    len(analysis.Files),
		TotalPackages: len(analysis.Packages),
		TotalAPIs:     len(analysis.APIEndpoints),
		GeneratedAt:   time.Now().UTC(),
	}

	// Count nodes by type
	nodes, _, err := u.repo.GetAllNodes(projectID, 1, 10000, "")
	if err == nil {
		stats.TotalNodes = len(nodes)
		stats.NodesByType = make(map[string]int)

		for _, node := range nodes {
			stats.NodesByType[node.Type]++
		}
	}

	// Count files by extension
	stats.FilesByExtension = make(map[string]int)
	for _, fileInfo := range analysis.Files {
		ext := u.getFileExtension(fileInfo.Path)
		stats.FilesByExtension[ext]++
	}

	return stats, nil
}

func (u *AnalyzerUsecase) GetAPIStatistics(projectID, apiID string) (*entity.APIStatistics, error) {
	endpoint, err := u.repo.GetAPIEndpoint(projectID, apiID)
	if err != nil {
		return nil, err
	}

	// Create API statistics based on the actual entity definition
	stats := &entity.APIStatistics{
		ProjectID: projectID,
		APIID:     apiID,
		Method:    endpoint.Method,
		Path:      endpoint.Path,
		// Note: The original entity.APIStatistics doesn't have Handler, TotalNodes, GeneratedAt, or NodesByType fields
		// We'll need to update the entity definition if we want to include these
	}

	return stats, nil
}

func (u *AnalyzerUsecase) ExportAnalysis(projectID, format string) (string, error) {
	analysis, err := u.repo.GetProjectAnalysis(projectID)
	if err != nil {
		return "", err
	}

	switch strings.ToLower(format) {
	case "json":
		data, err := json.MarshalIndent(analysis, "", "  ")
		return string(data), err
	case "yaml":
		data, err := yaml.Marshal(analysis)
		return string(data), err
	case "xml":
		data, err := xml.MarshalIndent(analysis, "", "  ")
		return string(data), err
	default:
		return "", errors.NewValidationError("unsupported format: " + format)
	}
}

func (u *AnalyzerUsecase) ExportAPIAnalysis(projectID, apiID, format string) (string, error) {
	endpoint, err := u.repo.GetAPIEndpoint(projectID, apiID)
	if err != nil {
		return "", err
	}

	nodes, err := u.repo.GetAPINodes(projectID, apiID)
	if err != nil {
		return "", err
	}

	exportData := map[string]interface{}{
		"endpoint": endpoint,
		"nodes":    nodes,
	}

	switch strings.ToLower(format) {
	case "json":
		data, err := json.MarshalIndent(exportData, "", "  ")
		return string(data), err
	case "yaml":
		data, err := yaml.Marshal(exportData)
		return string(data), err
	case "xml":
		data, err := xml.MarshalIndent(exportData, "", "  ")
		return string(data), err
	default:
		return "", errors.NewValidationError("unsupported format: " + format)
	}
}

func (u *AnalyzerUsecase) GetDependencyGraph(projectID string) (*entity.DependencyGraph, error) {
	analysis, err := u.repo.GetProjectAnalysis(projectID)
	if err != nil {
		return nil, err
	}

	return analysis.DependencyGraph, nil
}

func (u *AnalyzerUsecase) GetAPIDependencies(projectID, apiID string) ([]*entity.Dependency, error) {
	graph, err := u.GetDependencyGraph(projectID)
	if err != nil {
		return nil, err
	}

	// Find dependencies for the specific API
	var dependencies []*entity.Dependency
	for _, dep := range graph.Dependencies {
		if dep.From == apiID || dep.To == apiID {
			dependencies = append(dependencies, dep)
		}
	}

	return dependencies, nil
}

func (u *AnalyzerUsecase) getFileExtension(filePath string) string {
	parts := strings.Split(filePath, ".")
	if len(parts) > 1 {
		return "." + parts[len(parts)-1]
	}
	return ""
}
