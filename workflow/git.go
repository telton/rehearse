package workflow

import (
	"os/exec"
	"path/filepath"
	"strings"
)

// GitInfo holds information extracted from the local git repository.
type GitInfo struct {
	Ref        string
	SHA        string
	ShortSHA   string
	Actor      string
	Repository string
	Workspace  string
}

// NewGitInfo extracts git information from the current repository.
func NewGitInfo() (*GitInfo, error) {
	info := &GitInfo{}

	ref, err := execGit("symbolic-ref", "HEAD")
	if err != nil {
		// Detached HEAD, fall back to SHA.
		ref, err = execGit("rev-parse", "HEAD")
		if err != nil {
			return nil, err
		}
	}
	info.Ref = ref

	sha, err := execGit("rev-parse", "HEAD")
	if err != nil {
		return nil, err
	}
	info.SHA = sha

	if len(sha) >= 7 {
		info.ShortSHA = sha[:7]
	}

	actor, _ := execGit("config", "user.name")
	info.Actor = actor

	workspace, _ := execGit("rev-parse", "--show-toplevel")

	remote, err := execGit("config", "--get", "remote.origin.url")
	if err == nil {
		info.Repository = parseRepositoryFromRemote(remote)
	} else {
		// Fall back to directory name.
		info.Repository = filepath.Base(workspace)
	}

	info.Workspace = workspace

	return info, nil
}

func execGit(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(out)), nil
}

func parseRepositoryFromRemote(remote string) string {
	remote = strings.TrimSuffix(remote, ".git")

	// Handle SSH format: git@github.com/owner/repo.git.
	if strings.HasPrefix(remote, "git@") {
		remote = strings.TrimPrefix(remote, "git@github.com:")
		return remote
	}

	// Handle HTTPS format: https://github.com/owner/repo.git.
	parts := strings.Split(remote, "/")
	if len(parts) >= 2 {
		return parts[len(parts)-2] + "/" + parts[len(parts)-1]
	}

	return remote
}
