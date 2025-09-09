package entity

import "time"

// APIEndpoint represents a discovered API endpoint in the project
type APIEndpoint struct {
	ID           string    `json:"id"`
	Method       string    `json:"method"`        // GET, POST, PUT, DELETE, etc.
	Path         string    `json:"path"`          // /api/v1/users/:id
	Handler      string    `json:"handler"`       // Handler function name
	File         string    `json:"file"`          // File where the endpoint is defined
	Package      string    `json:"package"`       // Package name
	Middleware   []string  `json:"middleware"`    // Applied middleware
	Parameters   []string  `json:"parameters"`    // Path parameters
	QueryParams  []string  `json:"query_params"`  // Query parameters
	RequestBody  string    `json:"request_body"`  // Request body type/struct
	Response     string    `json:"response"`      // Response type/struct
	RelatedNodes []string  `json:"related_nodes"` // IDs of related code nodes
	CreatedAt    time.Time `json:"created_at"`
}

// APIStatistics contains statistics for a specific API endpoint
type APIStatistics struct {
	ProjectID string `json:"project_id"`
	APIID     string `json:"api_id"`
	Method    string `json:"method"`
	Path      string `json:"path"`
}
