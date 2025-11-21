package workflow

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/goccy/go-yaml"
)

// Parse reads and parses a workflow file.
func Parse(path string) (*Workflow, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read workflow file: %w", err)
	}

	var w Workflow
	if err := yaml.Unmarshal(data, &w); err != nil {
		return nil, fmt.Errorf("parse workflow file: %w", err)
	}

	return &w, err
}

// FindWorkflows finds all workflow files in the .github/workflows directory.
func FindWorkflows(dir string) ([]string, error) {
	workflowDir := filepath.Join(dir, ".github", "workflows")

	entries, err := os.ReadDir(workflowDir)
	if err != nil {
		return nil, fmt.Errorf("read %s directory: %w", workflowDir, err)
	}

	var workflows []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}

		ext := filepath.Ext(e.Name())
		if ext == ".yaml" || ext == ".yml" {
			workflows = append(workflows, filepath.Join(workflowDir, e.Name()))
		}
	}

	return workflows, nil
}
