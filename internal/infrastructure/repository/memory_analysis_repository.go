package repository

import (
	"sync"
	"time"

	"goapianalyzer/internal/core/domain/entity"
	"goapianalyzer/pkg/errors"

	"github.com/google/uuid"
)

type MemoryAnalysisRepository struct {
	projects map[string]*entity.ProjectAnalysis
	nodes    map[string]map[string]*entity.CodeNode    // projectID -> nodeID -> CodeNode
	apis     map[string]map[string]*entity.APIEndpoint // projectID -> apiID -> APIEndpoint
	mutex    sync.RWMutex
}

func NewMemoryAnalysisRepository() *MemoryAnalysisRepository {
	return &MemoryAnalysisRepository{
		projects: make(map[string]*entity.ProjectAnalysis),
		nodes:    make(map[string]map[string]*entity.CodeNode),
		apis:     make(map[string]map[string]*entity.APIEndpoint),
	}
}

// Project Analysis Methods

func (r *MemoryAnalysisRepository) StoreProjectAnalysis(analysis *entity.ProjectAnalysis) error {
	if analysis == nil {
		return errors.NewValidationError("analysis cannot be nil")
	}

	if analysis.ID == "" {
		analysis.ID = uuid.New().String()
	}

	if analysis.CreatedAt.IsZero() {
		analysis.CreatedAt = time.Now().UTC()
	}
	analysis.UpdatedAt = time.Now().UTC()

	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.projects[analysis.ID] = analysis

	// Initialize nodes map for this project
	if r.nodes[analysis.ID] == nil {
		r.nodes[analysis.ID] = make(map[string]*entity.CodeNode)
	}

	// Initialize APIs map for this project
	if r.apis[analysis.ID] == nil {
		r.apis[analysis.ID] = make(map[string]*entity.APIEndpoint)
	}

	return nil
}

func (r *MemoryAnalysisRepository) GetProjectAnalysis(projectID string) (*entity.ProjectAnalysis, error) {
	if projectID == "" {
		return nil, errors.NewValidationError("project ID cannot be empty")
	}

	r.mutex.RLock()
	defer r.mutex.RUnlock()

	analysis, exists := r.projects[projectID]
	if !exists {
		return nil, errors.NewNotFoundError("project analysis not found")
	}

	return analysis, nil
}

func (r *MemoryAnalysisRepository) DeleteProjectAnalysis(projectID string) error {
	if projectID == "" {
		return errors.NewValidationError("project ID cannot be empty")
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, exists := r.projects[projectID]; !exists {
		return errors.NewNotFoundError("project analysis not found")
	}

	delete(r.projects, projectID)
	delete(r.nodes, projectID)
	delete(r.apis, projectID)

	return nil
}

func (r *MemoryAnalysisRepository) ListProjects() ([]*entity.ProjectAnalysis, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	projects := make([]*entity.ProjectAnalysis, 0, len(r.projects))
	for _, project := range r.projects {
		projects = append(projects, project)
	}

	return projects, nil
}

// Code Node Methods

func (r *MemoryAnalysisRepository) StoreCodeNode(projectID string, node *entity.CodeNode) error {
	if projectID == "" {
		return errors.NewValidationError("project ID cannot be empty")
	}
	if node == nil {
		return errors.NewValidationError("node cannot be nil")
	}

	if node.ID == "" {
		node.ID = uuid.New().String()
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()

	// Check if project exists
	if _, exists := r.projects[projectID]; !exists {
		return errors.NewNotFoundError("project not found")
	}

	// Initialize nodes map for project if needed
	if r.nodes[projectID] == nil {
		r.nodes[projectID] = make(map[string]*entity.CodeNode)
	}

	r.nodes[projectID][node.ID] = node
	return nil
}

func (r *MemoryAnalysisRepository) GetCodeNode(projectID, nodeID string) (*entity.CodeNode, error) {
	if projectID == "" || nodeID == "" {
		return nil, errors.NewValidationError("project ID and node ID cannot be empty")
	}

	r.mutex.RLock()
	defer r.mutex.RUnlock()

	projectNodes, exists := r.nodes[projectID]
	if !exists {
		return nil, errors.NewNotFoundError("project not found")
	}

	node, exists := projectNodes[nodeID]
	if !exists {
		return nil, errors.NewNotFoundError("code node not found")
	}

	return node, nil
}

func (r *MemoryAnalysisRepository) GetAllNodes(projectID string, page, limit int, nodeType string) ([]*entity.CodeNode, int64, error) {
	if projectID == "" {
		return nil, 0, errors.NewValidationError("project ID cannot be empty")
	}

	r.mutex.RLock()
	defer r.mutex.RUnlock()

	projectNodes, exists := r.nodes[projectID]
	if !exists {
		return nil, 0, errors.NewNotFoundError("project not found")
	}

	// Filter nodes by type if specified
	var filteredNodes []*entity.CodeNode
	for _, node := range projectNodes {
		if nodeType == "" || node.Type == nodeType {
			filteredNodes = append(filteredNodes, node)
		}
	}

	total := int64(len(filteredNodes))

	// Apply pagination
	start := (page - 1) * limit
	if start >= len(filteredNodes) {
		return []*entity.CodeNode{}, total, nil
	}

	end := start + limit
	if end > len(filteredNodes) {
		end = len(filteredNodes)
	}

	return filteredNodes[start:end], total, nil
}

func (r *MemoryAnalysisRepository) SearchNodes(projectID, query string, page, limit int) ([]*entity.CodeNode, int64, error) {
	if projectID == "" || query == "" {
		return nil, 0, errors.NewValidationError("project ID and query cannot be empty")
	}

	r.mutex.RLock()
	defer r.mutex.RUnlock()

	projectNodes, exists := r.nodes[projectID]
	if !exists {
		return nil, 0, errors.NewNotFoundError("project not found")
	}

	// Simple search implementation - can be enhanced with better search algorithms
	var matchedNodes []*entity.CodeNode
	for _, node := range projectNodes {
		if r.nodeMatchesQuery(node, query) {
			matchedNodes = append(matchedNodes, node)
		}
	}

	total := int64(len(matchedNodes))

	// Apply pagination
	start := (page - 1) * limit
	if start >= len(matchedNodes) {
		return []*entity.CodeNode{}, total, nil
	}

	end := start + limit
	if end > len(matchedNodes) {
		end = len(matchedNodes)
	}

	return matchedNodes[start:end], total, nil
}

func (r *MemoryAnalysisRepository) nodeMatchesQuery(node *entity.CodeNode, query string) bool {
	// Simple case-insensitive search in name and body
	return r.containsIgnoreCase(node.Name, query) ||
		r.containsIgnoreCase(node.Body, query) ||
		r.containsIgnoreCase(node.File, query) ||
		r.containsIgnoreCase(node.Package, query)
}

func (r *MemoryAnalysisRepository) containsIgnoreCase(text, substr string) bool {
	return len(text) >= len(substr) &&
		r.toLower(text) != text &&
		r.toLower(substr) != substr &&
		r.contains(r.toLower(text), r.toLower(substr))
}

func (r *MemoryAnalysisRepository) toLower(s string) string {
	result := make([]rune, len(s))
	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			result[i] = r + 32
		} else {
			result[i] = r
		}
	}
	return string(result)
}

func (r *MemoryAnalysisRepository) contains(text, substr string) bool {
	if len(substr) > len(text) {
		return false
	}
	for i := 0; i <= len(text)-len(substr); i++ {
		if text[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// API Endpoint Methods

func (r *MemoryAnalysisRepository) StoreAPIEndpoint(projectID string, endpoint *entity.APIEndpoint) error {
	if projectID == "" {
		return errors.NewValidationError("project ID cannot be empty")
	}
	if endpoint == nil {
		return errors.NewValidationError("endpoint cannot be nil")
	}

	if endpoint.ID == "" {
		endpoint.ID = uuid.New().String()
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()

	// Check if project exists
	if _, exists := r.projects[projectID]; !exists {
		return errors.NewNotFoundError("project not found")
	}

	// Initialize APIs map for project if needed
	if r.apis[projectID] == nil {
		r.apis[projectID] = make(map[string]*entity.APIEndpoint)
	}

	r.apis[projectID][endpoint.ID] = endpoint
	return nil
}

func (r *MemoryAnalysisRepository) GetAPIEndpoint(projectID, apiID string) (*entity.APIEndpoint, error) {
	if projectID == "" || apiID == "" {
		return nil, errors.NewValidationError("project ID and API ID cannot be empty")
	}

	r.mutex.RLock()
	defer r.mutex.RUnlock()

	projectAPIs, exists := r.apis[projectID]
	if !exists {
		return nil, errors.NewNotFoundError("project not found")
	}

	endpoint, exists := projectAPIs[apiID]
	if !exists {
		return nil, errors.NewNotFoundError("API endpoint not found")
	}

	return endpoint, nil
}

func (r *MemoryAnalysisRepository) GetAPIEndpoints(projectID string) ([]*entity.APIEndpoint, error) {
	if projectID == "" {
		return nil, errors.NewValidationError("project ID cannot be empty")
	}

	r.mutex.RLock()
	defer r.mutex.RUnlock()

	projectAPIs, exists := r.apis[projectID]
	if !exists {
		return nil, errors.NewNotFoundError("project not found")
	}

	endpoints := make([]*entity.APIEndpoint, 0, len(projectAPIs))
	for _, endpoint := range projectAPIs {
		endpoints = append(endpoints, endpoint)
	}

	return endpoints, nil
}

func (r *MemoryAnalysisRepository) GetAPINodes(projectID, apiID string) ([]*entity.CodeNode, error) {
	if projectID == "" || apiID == "" {
		return nil, errors.NewValidationError("project ID and API ID cannot be empty")
	}

	// Get the API endpoint first
	endpoint, err := r.GetAPIEndpoint(projectID, apiID)
	if err != nil {
		return nil, err
	}

	r.mutex.RLock()
	defer r.mutex.RUnlock()

	projectNodes, exists := r.nodes[projectID]
	if !exists {
		return []*entity.CodeNode{}, nil
	}

	// Collect all nodes related to this API
	var apiNodes []*entity.CodeNode
	for _, node := range projectNodes {
		// Check if node is related to the API endpoint based on file
		if node.File == endpoint.File {
			apiNodes = append(apiNodes, node)
		}
	}

	return apiNodes, nil
}
