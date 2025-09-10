package entity

// APIEndpoint represents a discovered API endpoint in the project
type APIEndpoint struct {
	ID     string `json:"id"`
	Method string `json:"method"` // GET, POST, PUT, DELETE, etc.
	Path   string `json:"path"`   // /api/v1/users/:id
	File   string `json:"file"`   // File where the endpoint is defined
}

// APIStatistics contains statistics for a specific API endpoint
type APIStatistics struct {
	ProjectID string `json:"project_id"`
	APIID     string `json:"api_id"`
	Method    string `json:"method"`
	Path      string `json:"path"`
}
