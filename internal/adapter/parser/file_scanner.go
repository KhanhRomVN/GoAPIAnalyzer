package parser

import (
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"goapianalyzer/internal/core/domain/entity"
	"goapianalyzer/pkg/errors"
	"goapianalyzer/pkg/utils"
)

type FileScanner struct {
	fileSet         *token.FileSet
	blacklistFiles  []string
	blacklistDirs   []string
	whitelistFiles  []string
	whitelistDirs   []string
	includeVendor   bool
	includeTestFile bool
}

type ScanConfig struct {
	BlacklistFiles  []string
	BlacklistDirs   []string
	WhitelistFiles  []string
	WhitelistDirs   []string
	IncludeVendor   bool
	IncludeTestFile bool
}

func NewFileScanner(config *ScanConfig) *FileScanner {
	if config == nil {
		config = &ScanConfig{}
	}

	return &FileScanner{
		fileSet:         token.NewFileSet(),
		blacklistFiles:  config.BlacklistFiles,
		blacklistDirs:   config.BlacklistDirs,
		whitelistFiles:  config.WhitelistFiles,
		whitelistDirs:   config.WhitelistDirs,
		includeVendor:   config.IncludeVendor,
		includeTestFile: config.IncludeTestFile,
	}
}

func (fs *FileScanner) ScanProject(projectPath string) (*entity.ProjectAnalysis, error) {
	if !utils.IsValidPath(projectPath) {
		return nil, errors.NewValidationError(fmt.Sprintf("invalid project path: %s", projectPath))
	}

	projectAnalysis := &entity.ProjectAnalysis{
		ProjectPath: projectPath,
		Files:       make(map[string]*entity.FileInfo),
		Packages:    make(map[string]*entity.PackageInfo),
	}

	err := filepath.WalkDir(projectPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return fs.handleDirectory(path, d)
		}

		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		if !fs.shouldProcessFile(path, d) {
			return nil
		}

		return fs.processGoFile(path, projectAnalysis)
	})

	if err != nil {
		return nil, errors.NewSystemError(fmt.Sprintf("failed to scan project: %v", err))
	}

	return projectAnalysis, nil
}

func (fs *FileScanner) handleDirectory(path string, d os.DirEntry) error {
	dirName := d.Name()

	// Skip vendor directory if not included
	if !fs.includeVendor && dirName == "vendor" {
		return filepath.SkipDir
	}

	// Skip hidden directories
	if strings.HasPrefix(dirName, ".") && dirName != "." {
		return filepath.SkipDir
	}

	// Check blacklist directories
	for _, blacklistDir := range fs.blacklistDirs {
		if strings.Contains(path, blacklistDir) {
			return filepath.SkipDir
		}
	}

	// If whitelist is specified, check if directory matches
	if len(fs.whitelistDirs) > 0 {
		matched := false
		for _, whitelistDir := range fs.whitelistDirs {
			if strings.Contains(path, whitelistDir) {
				matched = true
				break
			}
		}
		if !matched {
			return filepath.SkipDir
		}
	}

	return nil
}

func (fs *FileScanner) shouldProcessFile(path string, d os.DirEntry) bool {
	fileName := d.Name()

	// Skip test files if not included
	if !fs.includeTestFile && strings.HasSuffix(fileName, "_test.go") {
		return false
	}

	// Check blacklist files
	for _, blacklistFile := range fs.blacklistFiles {
		if strings.Contains(fileName, blacklistFile) || strings.Contains(path, blacklistFile) {
			return false
		}
	}

	// If whitelist is specified, check if file matches
	if len(fs.whitelistFiles) > 0 {
		matched := false
		for _, whitelistFile := range fs.whitelistFiles {
			if strings.Contains(fileName, whitelistFile) || strings.Contains(path, whitelistFile) {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	return true
}

func (fs *FileScanner) processGoFile(filePath string, projectAnalysis *entity.ProjectAnalysis) error {
	content, err := utils.ReadFile(filePath)
	if err != nil {
		return errors.NewSystemError(fmt.Sprintf("failed to read file %s: %v", filePath, err))
	}

	// Parse the Go source code
	astFile, err := parser.ParseFile(fs.fileSet, filePath, content, parser.ParseComments)
	if err != nil {
		return errors.NewValidationError(fmt.Sprintf("failed to parse file %s: %v", filePath, err))
	}

	// Get relative path from project root
	relativePath, err := filepath.Rel(projectAnalysis.ProjectPath, filePath)
	if err != nil {
		relativePath = filePath
	}

	// Create file info
	fileInfo := &entity.FileInfo{
		Path:         relativePath,
		AbsolutePath: filePath,
		PackageName:  astFile.Name.Name,
		Content:      string(content),
		AST:          astFile,
		Imports:      make([]string, 0),
		Functions:    make([]*entity.FunctionInfo, 0),
		Types:        make([]*entity.TypeInfo, 0),
		Variables:    make([]*entity.VariableInfo, 0),
		Constants:    make([]*entity.ConstantInfo, 0),
		Interfaces:   make([]*entity.InterfaceInfo, 0),
		Structs:      make([]*entity.StructInfo, 0),
	}

	// Extract imports
	for _, imp := range astFile.Imports {
		importPath := strings.Trim(imp.Path.Value, `"`)
		fileInfo.Imports = append(fileInfo.Imports, importPath)
	}

	// Store file info
	projectAnalysis.Files[relativePath] = fileInfo

	// Update package info
	packagePath := fs.extractPackagePath(relativePath)
	if _, exists := projectAnalysis.Packages[packagePath]; !exists {
		projectAnalysis.Packages[packagePath] = &entity.PackageInfo{
			Name:  astFile.Name.Name,
			Path:  packagePath,
			Files: make([]string, 0),
		}
	}
	projectAnalysis.Packages[packagePath].Files = append(projectAnalysis.Packages[packagePath].Files, relativePath)

	return nil
}

func (fs *FileScanner) extractPackagePath(filePath string) string {
	dir := filepath.Dir(filePath)
	if dir == "." {
		return ""
	}
	return dir
}

func (fs *FileScanner) GetFileSet() *token.FileSet {
	return fs.fileSet
}
