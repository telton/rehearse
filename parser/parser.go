package parser

import (
	"fmt"
	"io"

	"github.com/go-yaml/yaml"
)

type OnPullRequest struct {
	Paths []string `yaml:"paths"`
}

type OnPush struct {
	Branches []string `yaml:"branches"`
	Paths    []string `yaml:"paths"`
}

type OnTriggers struct {
	PullRequest *OnPullRequest `yaml:"pull_request,omitempty"`
	Push        *OnPush        `yaml:"push,omitempty"`
	// TODO: Add workflow_dispatch
}

type Step struct {
	Name string `yaml:"name"`
	Run  string `yaml:"string"`
}

type Job struct {
	RunsOn string  `yaml:"runs-on"`
	Steps  []*Step `yaml:"steps"`
}

type Workflow struct {
	Name string          `yaml:"name"`
	On   *OnTriggers     `yaml:"on"`
	Jobs map[string]*Job `yaml:"jobs"`
}

func NewWorkflow(r io.Reader) (*Workflow, error) {
	var w Workflow

	if err := yaml.NewDecoder(r).Decode(&w); err != nil {
		return nil, fmt.Errorf("decode yaml: %w", err)
	}

	return &w, nil
}
