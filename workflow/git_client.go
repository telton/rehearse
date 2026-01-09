package workflow

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/goccy/go-yaml"
)

// RealGitRepo implements ExecutorGitRepo using real git operations.
type RealGitRepo struct{}

// NewGitRepo creates a new Git repository client.
func NewGitRepo() ExecutorGitRepo {
	return &RealGitRepo{}
}

// CloneAction clones a GitHub action repository to the specified destination.
func (g *RealGitRepo) CloneAction(ctx context.Context, repo, ref, dest string) error {
	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return fmt.Errorf("creating destination directory: %w", err)
	}

	cmd := exec.CommandContext(ctx, "git", "clone", "--depth", "1", "--branch", ref, repo, dest)
	if err := cmd.Run(); err != nil {
		// If branch doesn't exist, try as a commit SHA.
		cmd = exec.CommandContext(ctx, "git", "clone", repo, dest)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("cloning repository %s: %w", repo, err)
		}

		// Checkout the specific commit.
		cmd = exec.CommandContext(ctx, "git", "-C", dest, "checkout", ref)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("checking out ref %s: %w", ref, err)
		}
	}

	return nil
}

// GetActionMetadata reads and parses action.yml or action.yaml from the given path.
func (g *RealGitRepo) GetActionMetadata(path string) (*ActionMetadata, error) {
	actionFiles := []string{"action.yml", "action.yaml"}

	for _, filename := range actionFiles {
		actionPath := filepath.Join(path, filename)
		if _, err := os.Stat(actionPath); err == nil {
			return g.parseActionMetadata(actionPath)
		}
	}

	return nil, fmt.Errorf("no action.yml or action.yaml found in %s", path)
}

// parseActionMetadata parses an action metadata file.
func (g *RealGitRepo) parseActionMetadata(filePath string) (*ActionMetadata, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("reading action file: %w", err)
	}

	var metadata ActionMetadata
	if err := yaml.Unmarshal(data, &metadata); err != nil {
		return nil, fmt.Errorf("parsing action metadata: %w", err)
	}

	return &metadata, nil
}

// GetCurrentBranch returns the current git branch.
func (g *RealGitRepo) GetCurrentBranch() (string, error) {
	cmd := exec.Command("git", "branch", "--show-current")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("getting current branch: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// GetCurrentCommit returns the current git commit SHA.
func (g *RealGitRepo) GetCurrentCommit() (string, error) {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("getting current commit: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}
