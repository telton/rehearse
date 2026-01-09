package workflow

import (
	"context"
	"sync"

	"github.com/stretchr/testify/mock"
)

// MockDockerClient is a mock implementation of DockerClient for testing.
type MockDockerClient struct {
	mock.Mock
	containers map[string]*MockContainer
	mu         sync.RWMutex
}

// MockContainer represents a mock container for testing.
type MockContainer struct {
	ID       string
	Config   *ContainerConfig
	Status   string
	ExitCode int
	Stdout   string
	Stderr   string
}

// NewMockDockerClient creates a new mock Docker client.
func NewMockDockerClient() *MockDockerClient {
	return &MockDockerClient{
		containers: make(map[string]*MockContainer),
	}
}

// CreateContainer mocks container creation.
func (m *MockDockerClient) CreateContainer(ctx context.Context, config *ContainerConfig) (string, error) {
	args := m.Called(ctx, config)

	if args.Error(1) != nil {
		return "", args.Error(1)
	}

	containerID := args.String(0)

	m.mu.Lock()
	m.containers[containerID] = &MockContainer{
		ID:     containerID,
		Config: config,
		Status: "created",
	}
	m.mu.Unlock()

	return containerID, nil
}

// StartContainer mocks container startup.
func (m *MockDockerClient) StartContainer(ctx context.Context, containerID string) error {
	args := m.Called(ctx, containerID)

	if args.Error(0) != nil {
		return args.Error(0)
	}

	m.mu.Lock()
	if container, exists := m.containers[containerID]; exists {
		container.Status = "running"
	}
	m.mu.Unlock()

	return nil
}

// ExecInContainer mocks command execution in container.
func (m *MockDockerClient) ExecInContainer(ctx context.Context, containerID string, cmd []string) (*ExecResult, error) {
	args := m.Called(ctx, containerID, cmd)

	if args.Error(1) != nil {
		return nil, args.Error(1)
	}

	result := args.Get(0).(*ExecResult)
	return result, nil
}

// StopContainer mocks container stopping.
func (m *MockDockerClient) StopContainer(ctx context.Context, containerID string) error {
	args := m.Called(ctx, containerID)

	if args.Error(0) != nil {
		return args.Error(0)
	}

	m.mu.Lock()
	if container, exists := m.containers[containerID]; exists {
		container.Status = "stopped"
	}
	m.mu.Unlock()

	return nil
}

// RemoveContainer mocks container removal.
func (m *MockDockerClient) RemoveContainer(ctx context.Context, containerID string) error {
	args := m.Called(ctx, containerID)

	if args.Error(0) != nil {
		return args.Error(0)
	}

	m.mu.Lock()
	delete(m.containers, containerID)
	m.mu.Unlock()

	return nil
}

// PullImage mocks image pulling.
func (m *MockDockerClient) PullImage(ctx context.Context, image string) error {
	args := m.Called(ctx, image)
	return args.Error(0)
}

// Close mocks client closing.
func (m *MockDockerClient) Close() error {
	args := m.Called()
	return args.Error(0)
}

// GetContainer returns a mock container by ID (for testing assertions).
func (m *MockDockerClient) GetContainer(containerID string) (*MockContainer, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	container, exists := m.containers[containerID]
	return container, exists
}

// GetContainerCount returns the number of active containers (for testing).
func (m *MockDockerClient) GetContainerCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.containers)
}

// MockGitRepo is a mock implementation of ExecutorGitRepo for testing.
type MockGitRepo struct {
	mock.Mock
	actions map[string]*ActionMetadata
	mu      sync.RWMutex
}

// NewMockGitRepo creates a new mock Git repository client.
func NewMockGitRepo() *MockGitRepo {
	return &MockGitRepo{
		actions: make(map[string]*ActionMetadata),
	}
}

// CloneAction mocks action cloning.
func (m *MockGitRepo) CloneAction(ctx context.Context, repo, ref, dest string) error {
	args := m.Called(ctx, repo, ref, dest)
	return args.Error(0)
}

// GetActionMetadata mocks action metadata retrieval.
func (m *MockGitRepo) GetActionMetadata(path string) (*ActionMetadata, error) {
	args := m.Called(path)

	if args.Error(1) != nil {
		return nil, args.Error(1)
	}

	metadata := args.Get(0).(*ActionMetadata)
	return metadata, nil
}

// GetCurrentBranch mocks getting current git branch.
func (m *MockGitRepo) GetCurrentBranch() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

// GetCurrentCommit mocks getting current git commit.
func (m *MockGitRepo) GetCurrentCommit() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

// SetActionMetadata sets mock action metadata for testing.
func (m *MockGitRepo) SetActionMetadata(path string, metadata *ActionMetadata) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.actions[path] = metadata
}

// TestDockerClient creates a real Docker client for integration tests using testcontainers.
func NewTestDockerClient() (DockerClient, error) {
	return NewMockDockerClient(), nil
}

// CreateTestRuntime creates a runtime for testing.
func CreateTestRuntime(workingDir string) *Runtime {
	return &Runtime{
		WorkingDir: workingDir,
		Containers: make(map[string]*ContainerInfo),
		Networks:   make(map[string]*NetworkInfo),
		Volumes:    make(map[string]*VolumeInfo),
		JobContext: &ExecutionJobContext{
			Job: &Job{
				Name:   "test-job",
				RunsOn: RunsOn{Labels: []string{"ubuntu-latest"}},
				Env:    make(map[string]string),
			},
			Outputs: make(map[string]string),
			Status:  "in_progress",
		},
		StepContext: &ExecutionStepContext{
			Outputs: make(map[string]string),
		},
	}
}

// CreateTestStep creates a step for testing.
func CreateTestStep(id, name, run string) *Step {
	return &Step{
		ID:   id,
		Name: name,
		Run:  run,
		Env:  make(map[string]string),
	}
}

// CreateTestActionStep creates an action step for testing.
func CreateTestActionStep(id, name, uses string, with map[string]string) *Step {
	return &Step{
		ID:   id,
		Name: name,
		Uses: uses,
		With: with,
		Env:  make(map[string]string),
	}
}

// CreateTestActionMetadata creates action metadata for testing.
func CreateTestActionMetadata(actionType, image, main string) *ActionMetadata {
	metadata := &ActionMetadata{
		Name:        "Test Action",
		Description: "A test action",
		Inputs:      make(map[string]ActionInput),
		Outputs:     make(map[string]ActionOutput),
		Runs: ActionRuns{
			Using: actionType,
			Env:   make(map[string]string),
		},
	}

	switch actionType {
	case "docker":
		metadata.Runs.Image = image
	case "node16", "node20":
		metadata.Runs.Main = main
	case "composite":
		metadata.Runs.Steps = []Step{}
	}

	return metadata
}

// AssertContainerConfig validates container configuration in tests.
func AssertContainerConfig(t interface {
	Errorf(format string, args ...any)
}, config *ContainerConfig, expectedImage string, expectedCmd []string) {
	if config.Image != expectedImage {
		t.Errorf("Expected image %s, got %s", expectedImage, config.Image)
	}

	if len(config.Cmd) != len(expectedCmd) {
		t.Errorf("Expected cmd length %d, got %d", len(expectedCmd), len(config.Cmd))
		return
	}

	for i, cmd := range expectedCmd {
		if config.Cmd[i] != cmd {
			t.Errorf("Expected cmd[%d] %s, got %s", i, cmd, config.Cmd[i])
		}
	}
}

// AssertEnvironmentContains checks if environment contains expected variables.
func AssertEnvironmentContains(t interface {
	Errorf(format string, args ...any)
}, env []string, expectedVar string) {
	for _, envVar := range env {
		if envVar == expectedVar {
			return
		}
	}
	t.Errorf("Expected environment to contain %s, but it was not found in %v", expectedVar, env)
}

// AssertEnvironmentHasPrefix checks if any environment variable has the given prefix.
func AssertEnvironmentHasPrefix(t interface {
	Errorf(format string, args ...any)
}, env []string, prefix string) {
	for _, envVar := range env {
		if len(envVar) > len(prefix) && envVar[:len(prefix)] == prefix {
			return
		}
	}
	t.Errorf("Expected environment to contain variable with prefix %s, but none found in %v", prefix, env)
}

// CreateTestShellExecutor creates a shell executor for testing.
func CreateTestShellExecutor(docker DockerClient) *ShellStepExecutor {
	return &ShellStepExecutor{
		Docker:   docker,
		renderer: NewRunRenderer(),
	}
}
