package workflow

import "strings"

// Workflow represents a GitHub Actions workflow file.
type Workflow struct {
	Name string            `yaml:"name"`
	On   any               `yaml:"on"` // Can be []string or map
	Env  map[string]string `yaml:"env"`
	Jobs map[string]Job    `yaml:"jobs"`
}

// Job represents a single job in a workflow.
type Job struct {
	Name      string            `yaml:"name"`
	RunsOn    RunsOn            `yaml:"runs-on"`
	Needs     Needs             `yaml:"needs"`
	If        string            `yaml:"if"`
	Env       map[string]string `yaml:"env"`
	Steps     []Step            `yaml:"steps"`
	Strategy  *Strategy         `yaml:"strategy"`
	Outputs   map[string]string `yaml:"outputs"`
	Container *Container        `yaml:"container"`
}

// Step represents a single step in a job.
type Step struct {
	ID   string            `yaml:"id"`
	Name string            `yaml:"name"`
	If   string            `yaml:"if"`
	Run  string            `yaml:"run"`
	Uses string            `yaml:"uses"`
	With map[string]string `yaml:"with"`
	Env  map[string]string `yaml:"env"`
}

// Strategy represents a matrix strategy.
type Strategy struct {
	Matrix      map[string]any `yaml:"strategy"`
	FailFast    *bool          `yaml:"fail-fast"`
	MaxParallel int            `yaml:"max-parallel"`
}

// Container represents a container configuration.
type Container struct {
	Image string            `yaml:"image"`
	Env   map[string]string `yaml:"env"`
}

// RunsOn handles both string and array formats.
type RunsOn struct {
	Labels []string
}

func (r *RunsOn) UnmarshalYAML(unmarshal func(any) error) error {
	// Try string first.
	var single string
	if err := unmarshal(&single); err == nil {
		r.Labels = []string{single}
		return nil
	}

	// Try array.
	var multiple []string
	if err := unmarshal(&multiple); err == nil {
		r.Labels = multiple
		return nil
	}

	return nil
}

func (r RunsOn) String() string {
	if len(r.Labels) == 1 {
		return r.Labels[0]
	}
	return "[" + strings.Join(r.Labels, ", ") + "]"
}

// Needs handles both string and array formats.
type Needs struct {
	Jobs []string
}

func (n *Needs) UnmarshalYAML(unmarshal func(any) error) error {
	// Try string first.
	var single string
	if err := unmarshal(&single); err == nil {
		n.Jobs = []string{single}
		return nil
	}

	// Try array.
	var multiple []string
	if err := unmarshal(&multiple); err == nil {
		n.Jobs = multiple
		return nil
	}

	return nil
}
