package action

import (
	"fmt"
	"os"
	"path/filepath"
)

// ReadFile reads the contents of a file.
func ReadFile(path string, requireAbsolute bool) (string, error) {
	if requireAbsolute && !filepath.IsAbs(path) {
		return "", fmt.Errorf("path must be absolute: %s", path)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	return string(content), nil
}

// WriteFile writes content to a file.
func WriteFile(path, content string, requireAbsolute bool) error {
	if requireAbsolute && !filepath.IsAbs(path) {
		return fmt.Errorf("path must be absolute: %s", path)
	}

	// Ensure parent directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}
