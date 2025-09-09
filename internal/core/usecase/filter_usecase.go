package usecase

import (
	"strings"

	"goapianalyzer/internal/core/domain/entity"
	"goapianalyzer/internal/core/domain/repository"
	"goapianalyzer/internal/infrastructure/logger"
	"goapianalyzer/pkg/utils"
)

type FilterUsecase struct {
	repo   repository.AnalysisRepository
	logger logger.Logger
}

func NewFilterUsecase(repo repository.AnalysisRepository) *FilterUsecase {
	return &FilterUsecase{
		repo:   repo,
		logger: logger.GetLogger(),
	}
}

func (u *FilterUsecase) ApplyFilters(projectID string, filters *entity.FilterConfig) ([]*entity.CodeNode, error) {
	u.logger.WithFields(map[string]interface{}{
		"project_id": projectID,
		"filters":    filters,
	}).Info("Applying filters to project nodes")

	// Get all nodes for the project
	allNodes, _, err := u.repo.GetAllNodes(projectID, 1, 100000, "") // Get all nodes
	if err != nil {
		return nil, err
	}

	// Apply filters
	filteredNodes := u.filterNodes(allNodes, filters)

	u.logger.WithFields(map[string]interface{}{
		"project_id":      projectID,
		"total_nodes":     len(allNodes),
		"filtered_nodes":  len(filteredNodes),
		"filters_applied": u.countAppliedFilters(filters),
	}).Info("Filters applied successfully")

	return filteredNodes, nil
}

func (u *FilterUsecase) filterNodes(nodes []*entity.CodeNode, filters *entity.FilterConfig) []*entity.CodeNode {
	if filters == nil {
		return nodes
	}

	var filteredNodes []*entity.CodeNode

	for _, node := range nodes {
		if u.nodePassesFilters(node, filters) {
			filteredNodes = append(filteredNodes, node)
		}
	}

	return filteredNodes
}

func (u *FilterUsecase) nodePassesFilters(node *entity.CodeNode, filters *entity.FilterConfig) bool {
	// Filter by node type
	if len(filters.NodeTypes) > 0 {
		if !utils.Contains(filters.NodeTypes, node.Type) {
			return false
		}
	}

	// Filter by file extension
	if len(filters.FileExtensions) > 0 {
		fileExt := u.getFileExtension(node.File)
		if !utils.Contains(filters.FileExtensions, fileExt) {
			return false
		}
	}

	// Filter by package name
	if len(filters.PackageNames) > 0 {
		if !u.containsAny(filters.PackageNames, node.Package) {
			return false
		}
	}

	// Filter by function name (for function nodes)
	if len(filters.FunctionNames) > 0 && node.Type == "function" {
		if !u.containsAny(filters.FunctionNames, node.Name) {
			return false
		}
	}

	// Filter by blacklisted files
	if len(filters.BlacklistFiles) > 0 {
		for _, blacklistedFile := range filters.BlacklistFiles {
			if strings.Contains(node.File, blacklistedFile) {
				return false
			}
		}
	}

	// Filter by blacklisted directories
	if len(filters.BlacklistDirs) > 0 {
		for _, blacklistedDir := range filters.BlacklistDirs {
			if strings.Contains(node.File, blacklistedDir) {
				return false
			}
		}
	}

	// Filter by complexity (if complexity metadata exists)
	if filters.MinComplexity != nil || filters.MaxComplexity != nil {
		complexity := u.getNodeComplexity(node)
		if filters.MinComplexity != nil && complexity < *filters.MinComplexity {
			return false
		}
		if filters.MaxComplexity != nil && complexity > *filters.MaxComplexity {
			return false
		}
	}

	return true
}

func (u *FilterUsecase) getFileExtension(filePath string) string {
	parts := strings.Split(filePath, ".")
	if len(parts) > 1 {
		return "." + parts[len(parts)-1]
	}
	return ""
}

func (u *FilterUsecase) containsAny(slice []string, target string) bool {
	for _, item := range slice {
		if strings.Contains(strings.ToLower(target), strings.ToLower(item)) {
			return true
		}
	}
	return false
}

func (u *FilterUsecase) getNodeComplexity(node *entity.CodeNode) int {
	// Try to extract complexity from metadata
	if node.Metadata != nil {
		if complexity, exists := node.Metadata["complexity"]; exists {
			if complexityInt, ok := complexity.(int); ok {
				return complexityInt
			}
		}
	}

	// Calculate basic complexity based on node body length and type
	baseComplexity := len(strings.Split(node.Body, "\n"))

	switch node.Type {
	case "function":
		// Functions have higher base complexity
		return baseComplexity * 2
	case "struct":
		// Structs complexity based on number of fields
		if node.Metadata != nil {
			if fields, exists := node.Metadata["fields"]; exists {
				if fieldsSlice, ok := fields.([]interface{}); ok {
					return len(fieldsSlice) * 3
				}
			}
		}
		return baseComplexity
	case "interface":
		// Interface complexity based on number of methods
		if node.Metadata != nil {
			if methods, exists := node.Metadata["methods"]; exists {
				if methodsSlice, ok := methods.([]interface{}); ok {
					return len(methodsSlice) * 2
				}
			}
		}
		return baseComplexity
	default:
		return baseComplexity
	}
}

func (u *FilterUsecase) countAppliedFilters(filters *entity.FilterConfig) int {
	count := 0

	if len(filters.NodeTypes) > 0 {
		count++
	}
	if len(filters.FileExtensions) > 0 {
		count++
	}
	if len(filters.PackageNames) > 0 {
		count++
	}
	if len(filters.FunctionNames) > 0 {
		count++
	}
	if len(filters.BlacklistFiles) > 0 {
		count++
	}
	if len(filters.BlacklistDirs) > 0 {
		count++
	}
	if filters.MinComplexity != nil {
		count++
	}
	if filters.MaxComplexity != nil {
		count++
	}

	return count
}

// Advanced filtering methods

func (u *FilterUsecase) FilterByDependencies(projectID string, targetNodeID string, includeIncoming, includeOutgoing bool) ([]*entity.CodeNode, error) {
	// Get dependency graph
	analysis, err := u.repo.GetProjectAnalysis(projectID)
	if err != nil {
		return nil, err
	}

	if analysis.DependencyGraph == nil {
		return []*entity.CodeNode{}, nil
	}

	// Find related node IDs
	relatedNodeIDs := make(map[string]bool)

	for _, dep := range analysis.DependencyGraph.Dependencies {
		if includeIncoming && dep.To == targetNodeID {
			relatedNodeIDs[dep.From] = true
		}
		if includeOutgoing && dep.From == targetNodeID {
			relatedNodeIDs[dep.To] = true
		}
	}

	// Get the related nodes
	var relatedNodes []*entity.CodeNode
	for nodeID := range relatedNodeIDs {
		node, err := u.repo.GetCodeNode(projectID, nodeID)
		if err == nil {
			relatedNodes = append(relatedNodes, node)
		}
	}

	return relatedNodes, nil
}

func (u *FilterUsecase) FilterByAPI(projectID, apiID string) ([]*entity.CodeNode, error) {
	return u.repo.GetAPINodes(projectID, apiID)
}

func (u *FilterUsecase) GetFilterSuggestions(projectID string) (*entity.FilterSuggestions, error) {
	// Get all nodes to analyze available filter options
	allNodes, _, err := u.repo.GetAllNodes(projectID, 1, 100000, "")
	if err != nil {
		return nil, err
	}

	suggestions := &entity.FilterSuggestions{
		NodeTypes:      make([]string, 0),
		FileExtensions: make([]string, 0),
		PackageNames:   make([]string, 0),
		FunctionNames:  make([]string, 0),
	}

	nodeTypesMap := make(map[string]bool)
	fileExtensionsMap := make(map[string]bool)
	packageNamesMap := make(map[string]bool)
	functionNamesMap := make(map[string]bool)

	for _, node := range allNodes {
		// Collect node types
		if !nodeTypesMap[node.Type] {
			nodeTypesMap[node.Type] = true
			suggestions.NodeTypes = append(suggestions.NodeTypes, node.Type)
		}

		// Collect file extensions
		fileExt := u.getFileExtension(node.File)
		if fileExt != "" && !fileExtensionsMap[fileExt] {
			fileExtensionsMap[fileExt] = true
			suggestions.FileExtensions = append(suggestions.FileExtensions, fileExt)
		}

		// Collect package names
		if node.Package != "" && !packageNamesMap[node.Package] {
			packageNamesMap[node.Package] = true
			suggestions.PackageNames = append(suggestions.PackageNames, node.Package)
		}

		// Collect function names (for function nodes)
		if node.Type == "function" && !functionNamesMap[node.Name] {
			functionNamesMap[node.Name] = true
			suggestions.FunctionNames = append(suggestions.FunctionNames, node.Name)
		}
	}

	return suggestions, nil
}
