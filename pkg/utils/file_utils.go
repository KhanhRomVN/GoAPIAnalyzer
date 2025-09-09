package utils

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// ReadFile reads the contents of a file and returns it as a byte slice
func ReadFile(filePath string) ([]byte, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", filePath, err)
	}
	defer file.Close()

	return io.ReadAll(file)
}

// WriteFile writes data to a file, creating directories if necessary
func WriteFile(filePath string, data []byte) error {
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	return os.WriteFile(filePath, data, 0644)
}

// FileExists checks if a file exists
func FileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return !os.IsNotExist(err)
}

// IsDirectory checks if a path is a directory
func IsDirectory(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// IsValidPath checks if a path is valid and accessible
func IsValidPath(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// GetFileExtension returns the file extension (including the dot)
func GetFileExtension(filePath string) string {
	return filepath.Ext(filePath)
}

// GetFileName returns the filename without the directory path
func GetFileName(filePath string) string {
	return filepath.Base(filePath)
}

// GetFileNameWithoutExt returns the filename without the extension
func GetFileNameWithoutExt(filePath string) string {
	base := filepath.Base(filePath)
	ext := filepath.Ext(base)
	return strings.TrimSuffix(base, ext)
}

// GetFileSize returns the size of a file in bytes
func GetFileSize(filePath string) (int64, error) {
	info, err := os.Stat(filePath)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

// ListFiles lists all files in a directory (non-recursive)
func ListFiles(dirPath string) ([]string, error) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() {
			files = append(files, filepath.Join(dirPath, entry.Name()))
		}
	}

	return files, nil
}

// ListFilesRecursive lists all files in a directory recursively
func ListFilesRecursive(dirPath string) ([]string, error) {
	var files []string

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			files = append(files, path)
		}
		return nil
	})

	return files, err
}

// ListGoFiles lists all .go files in a directory recursively
func ListGoFiles(dirPath string) ([]string, error) {
	var goFiles []string

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".go") {
			goFiles = append(goFiles, path)
		}
		return nil
	})

	return goFiles, err
}

// CopyFile copies a file from src to dst
func CopyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file %s: %w", src, err)
	}
	defer srcFile.Close()

	// Create destination directory if it doesn't exist
	dstDir := filepath.Dir(dst)
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory %s: %w", dstDir, err)
	}

	dstFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file %s: %w", dst, err)
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	return nil
}

// RemoveFile removes a file if it exists
func RemoveFile(filePath string) error {
	if !FileExists(filePath) {
		return nil
	}
	return os.Remove(filePath)
}

// CreateTempFile creates a temporary file with the given prefix
func CreateTempFile(prefix string) (*os.File, error) {
	return os.CreateTemp("", prefix)
}

// CreateTempDir creates a temporary directory with the given prefix
func CreateTempDir(prefix string) (string, error) {
	return os.MkdirTemp("", prefix)
}

// CleanPath cleans and normalizes a file path
func CleanPath(path string) string {
	return filepath.Clean(path)
}

// RelativePath returns the relative path from base to target
func RelativePath(base, target string) (string, error) {
	return filepath.Rel(base, target)
}

// AbsolutePath returns the absolute path of a file
func AbsolutePath(path string) (string, error) {
	return filepath.Abs(path)
}

// JoinPath joins path elements into a single path
func JoinPath(elements ...string) string {
	return filepath.Join(elements...)
}

// SplitPath splits a path into directory and filename
func SplitPath(path string) (dir, file string) {
	return filepath.Split(path)
}

// IsGoFile checks if a file has a .go extension
func IsGoFile(filePath string) bool {
	return strings.HasSuffix(strings.ToLower(filePath), ".go")
}

// IsTestFile checks if a file is a Go test file
func IsTestFile(filePath string) bool {
	return strings.HasSuffix(strings.ToLower(filePath), "_test.go")
}

// FilterGoFiles filters a slice of file paths to include only .go files
func FilterGoFiles(files []string) []string {
	var goFiles []string
	for _, file := range files {
		if IsGoFile(file) {
			goFiles = append(goFiles, file)
		}
	}
	return goFiles
}

// FilterNonTestFiles filters out test files from a slice of file paths
func FilterNonTestFiles(files []string) []string {
	var nonTestFiles []string
	for _, file := range files {
		if !IsTestFile(file) {
			nonTestFiles = append(nonTestFiles, file)
		}
	}
	return nonTestFiles
}

// GetDirectorySize calculates the total size of all files in a directory
func GetDirectorySize(dirPath string) (int64, error) {
	var size int64

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})

	return size, err
}
