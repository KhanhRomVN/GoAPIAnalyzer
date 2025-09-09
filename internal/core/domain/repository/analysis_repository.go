package repository

import "goapianalyzer/internal/core/domain/entity"

// AnalysisRepository defines the interface for storing and retrieving analysis data
type AnalysisRepository interface {
	// Project Analysis Methods
	StoreProjectAnalysis(analysis *entity.ProjectAnalysis) error
	GetProjectAnalysis(projectID string) (*entity.ProjectAnalysis, error)
	DeleteProjectAnalysis(projectID string) error
	ListProjects() ([]*entity.ProjectAnalysis, error)

	// Code Node Methods
	StoreCodeNode(projectID string, node *entity.CodeNode) error
	GetCodeNode(projectID, nodeID string) (*entity.CodeNode, error)
	GetAllNodes(projectID string, page, limit int, nodeType string) ([]*entity.CodeNode, int64, error)
	SearchNodes(projectID, query string, page, limit int) ([]*entity.CodeNode, int64, error)

	// API Endpoint Methods
	StoreAPIEndpoint(projectID string, endpoint *entity.APIEndpoint) error
	GetAPIEndpoint(projectID, apiID string) (*entity.APIEndpoint, error)
	GetAPIEndpoints(projectID string) ([]*entity.APIEndpoint, error)
	GetAPINodes(projectID, apiID string) ([]*entity.CodeNode, error)
}
