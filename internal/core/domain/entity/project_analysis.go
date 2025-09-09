package entity

import (
	"go/ast"
	"time"
)

// ProjectAnalysis represents the complete analysis of a Go project
type ProjectAnalysis struct {
	ID              string                  `json:"id"`
	ProjectPath     string                  `json:"project_path"`
	Files           map[string]*FileInfo    `json:"files"`
	Packages        map[string]*PackageInfo `json:"packages"`
	APIEndpoints    []*APIEndpoint          `json:"api_endpoints"`
	DependencyGraph *DependencyGraph        `json:"dependency_graph"`
	CreatedAt       time.Time               `json:"created_at"`
	UpdatedAt       time.Time               `json:"updated_at"`
}

// FileInfo contains information about a Go source file
type FileInfo struct {
	Path         string           `json:"path"`
	AbsolutePath string           `json:"absolute_path"`
	PackageName  string           `json:"package_name"`
	Content      string           `json:"content"`
	AST          *ast.File        `json:"-"` // Excluded from JSON serialization
	Imports      []string         `json:"imports"`
	Functions    []*FunctionInfo  `json:"functions"`
	Types        []*TypeInfo      `json:"types"`
	Variables    []*VariableInfo  `json:"variables"`
	Constants    []*ConstantInfo  `json:"constants"`
	Interfaces   []*InterfaceInfo `json:"interfaces"`
	Structs      []*StructInfo    `json:"structs"`
}

// PackageInfo contains information about a Go package
type PackageInfo struct {
	Name  string   `json:"name"`
	Path  string   `json:"path"`
	Files []string `json:"files"`
}

// DependencyGraph represents the dependency relationships between code elements
type DependencyGraph struct {
	Nodes        []*DependencyNode `json:"nodes"`
	Dependencies []*Dependency     `json:"dependencies"`
}

// DependencyNode represents a node in the dependency graph
type DependencyNode struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Type    string `json:"type"`
	File    string `json:"file"`
	Package string `json:"package"`
}

// Dependency represents a dependency relationship between two nodes
type Dependency struct {
	From     string `json:"from"`
	To       string `json:"to"`
	Type     string `json:"type"`     // call, import, inherit, implement, etc.
	Strength int    `json:"strength"` // 1-10, indicating dependency strength
}

// AnalysisConfig contains configuration for project analysis
type AnalysisConfig struct {
	BlacklistFiles  []string `json:"blacklist_files,omitempty"`
	BlacklistDirs   []string `json:"blacklist_dirs,omitempty"`
	WhitelistFiles  []string `json:"whitelist_files,omitempty"`
	WhitelistDirs   []string `json:"whitelist_dirs,omitempty"`
	IncludeVendor   bool     `json:"include_vendor"`
	IncludeTestFile bool     `json:"include_test_file"`
}

// FilterConfig contains configuration for filtering nodes
type FilterConfig struct {
	NodeTypes      []string `json:"node_types,omitempty"`
	FileExtensions []string `json:"file_extensions,omitempty"`
	PackageNames   []string `json:"package_names,omitempty"`
	FunctionNames  []string `json:"function_names,omitempty"`
	BlacklistFiles []string `json:"blacklist_files,omitempty"`
	BlacklistDirs  []string `json:"blacklist_dirs,omitempty"`
	MinComplexity  *int     `json:"min_complexity,omitempty"`
	MaxComplexity  *int     `json:"max_complexity,omitempty"`
}

// FilterSuggestions contains available filter options for a project
type FilterSuggestions struct {
	NodeTypes      []string `json:"node_types"`
	FileExtensions []string `json:"file_extensions"`
	PackageNames   []string `json:"package_names"`
	FunctionNames  []string `json:"function_names"`
}

// ProjectStatistics contains statistics about a project
type ProjectStatistics struct {
	ProjectID        string         `json:"project_id"`
	TotalFiles       int            `json:"total_files"`
	TotalPackages    int            `json:"total_packages"`
	TotalNodes       int            `json:"total_nodes"`
	TotalAPIs        int            `json:"total_apis"`
	NodesByType      map[string]int `json:"nodes_by_type"`
	FilesByExtension map[string]int `json:"files_by_extension"`
	GeneratedAt      time.Time      `json:"generated_at"`
}
