package workflow

import (
	"fmt"
	"os"
)

// Context holds all of the context available during a workflow's execution.
type Context struct {
	GitHub  GitHubContext
	Env     map[string]string
	Secrets map[string]string
	Jobs    map[string]JobContext
	Steps   map[string]StepContext
	Matrix  map[string]any
}

// GitHubContext mirrors the github.* context in Actions.
type GitHubContext struct {
	EventName  string         `json:"event_name"`
	Ref        string         `json:"ref"`
	SHA        string         `json:"sha"`
	Actor      string         `json:"actor"`
	Repository string         `json:"repository"`
	Workspace  string         `json:"workspace"`
	Event      map[string]any `json:"event"`
}

// JobContext holds info about completed jobs.
type JobContext struct {
	Status  string
	Outputs map[string]string
}

// StepContext holds info about completed steps.
type StepContext struct {
	Outcome string
	Outputs map[string]string
}

// Options for building a context.
type Options struct {
	EventName    string
	Ref          string
	EventPayload map[string]any
	Secrets      map[string]string
}

// NewContext creates a new Context from git info and options.
func NewContext(opts Options) (*Context, error) {
	gitInfo, err := NewGitInfo()
	if err != nil {
		return nil, fmt.Errorf("create git info: %w", err)
	}

	ctx := &Context{
		GitHub: GitHubContext{
			EventName:  opts.EventName,
			Ref:        opts.Ref,
			SHA:        gitInfo.SHA,
			Actor:      gitInfo.Actor,
			Repository: gitInfo.Repository,
			Workspace:  gitInfo.Workspace,
			Event:      opts.EventPayload,
		},
		Env:     make(map[string]string),
		Secrets: opts.Secrets,
		Jobs:    make(map[string]JobContext),
		Steps:   make(map[string]StepContext),
		Matrix:  make(map[string]any),
	}

	// Use git ref if not overridden.
	if ctx.GitHub.Ref == "" {
		ctx.GitHub.Ref = gitInfo.Ref
	}

	// Default event payload.
	if ctx.GitHub.Event == nil {
		ctx.GitHub.Event = defaultEventPayload(opts.EventName)
	}

	// Load env from system.
	for _, e := range os.Environ() {
		for i := range len(e) {
			if e[i] == '=' {
				ctx.Env[e[:i]] = e[i+1:]
				break
			}
		}
	}

	return ctx, nil
}

func defaultEventPayload(event string) map[string]any {
	switch event {
	case "push":
		return map[string]any{
			"ref":     "",
			"before":  "",
			"after":   "",
			"commits": []any{},
		}
	case "pull_request":
		return map[string]any{
			"action": "opened",
			"number": 1,
			"pull_request": map[string]any{
				"number": 1,
				"title":  "",
				"body":   "",
			},
		}
	default:
		return map[string]any{}
	}
}

// Lookup retrieves a value from the context by path (ex: "github.ref").
func (c *Context) Lookup(path string) (any, bool) {
	parts := splitPath(path)
	if len(parts) == 0 {
		return nil, false
	}

	switch parts[0] {
	case "github":
		return c.lookupGitHub(parts[1:])
	case "env":
		if len(parts) == 2 {
			v, ok := c.Env[parts[1]]
			return v, ok
		}
	case "secrets":
		if len(parts) == 2 {
			v, ok := c.Secrets[parts[1]]
			return v, ok
		}
	case "jobs":
		return c.lookupJobs(parts[1:])
	case "steps":
		return c.lookupSteps(parts[1:])
	case "matrix":
		if len(parts) == 2 {
			v, ok := c.Matrix[parts[1]]
			return v, ok
		}
	}

	return nil, false
}

func (c *Context) lookupGitHub(parts []string) (any, bool) {
	if len(parts) == 0 {
		return nil, false
	}

	switch parts[0] {
	case "event_name":
		return c.GitHub.EventName, true
	case "ref":
		return c.GitHub.Ref, true
	case "sha":
		return c.GitHub.SHA, true
	case "actor":
		return c.GitHub.Actor, true
	case "repository":
		return c.GitHub.Repository, true
	case "workspace":
		return c.GitHub.Workspace, true
	case "event":
		if len(parts) == 1 {
			return c.GitHub.Event, true
		}
		return lookupMap(c.GitHub.Event, parts[1:])
	}

	return nil, false
}

func (c *Context) lookupJobs(parts []string) (any, bool) {
	if len(parts) > 2 {
		return nil, false
	}

	job, ok := c.Jobs[parts[0]]
	if !ok {
		return nil, false
	}

	switch parts[1] {
	case "status":
		return job.Status, true
	case "outputs":
		if len(parts) == 3 {
			v, ok := job.Outputs[parts[2]]
			return v, ok
		}
	}

	return nil, false
}

func (c *Context) lookupSteps(parts []string) (any, bool) {
	if len(parts) < 2 {
		return nil, false
	}

	step, ok := c.Steps[parts[0]]
	if !ok {
		return nil, false
	}

	switch parts[1] {
	case "outcome":
		return step.Outcome, true
	case "outputs":
		if len(parts) == 3 {
			v, ok := step.Outputs[parts[2]]
			return v, ok
		}
	}

	return nil, false
}

func lookupMap(m map[string]any, parts []string) (any, bool) {
	if len(parts) == 0 {
		return m, true
	}

	v, ok := m[parts[0]]
	if !ok {
		return nil, false
	}

	if len(parts) == 1 {
		return v, true
	}

	if nested, ok := v.(map[string]any); ok {
		return lookupMap(nested, parts[1:])
	}

	return nil, false
}

func splitPath(path string) []string {
	var parts []string
	var current string

	for _, c := range path {
		if c == '.' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(c)
		}
	}

	if current != "" {
		parts = append(parts, current)
	}

	return parts
}
