package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/api/types/network"
	"github.com/moby/moby/client"

	"github.com/telton/rehearse/ui"
)

// RealDockerClient implements DockerClient using the actual Docker SDK.
type RealDockerClient struct {
	client *client.Client
	writer io.Writer
}

// NewDockerClient creates a new Docker client.
func NewDockerClient(w io.Writer) (DockerClient, error) {
	cli, err := client.New(client.FromEnv)
	if err != nil {
		return nil, err
	}

	return &RealDockerClient{client: cli, writer: w}, nil
}

// CreateContainer creates a new Docker container.
func (d *RealDockerClient) CreateContainer(ctx context.Context, config *ContainerConfig) (string, error) {
	containerConfig := &container.Config{
		Image:      config.Image,
		Cmd:        config.Cmd,
		Env:        config.Env,
		WorkingDir: config.WorkingDir,
	}

	var mounts []mount.Mount
	for _, vol := range config.Volumes {
		mounts = append(mounts, mount.Mount{
			Type:   mount.Type(vol.Type),
			Source: vol.Source,
			Target: vol.Target,
		})
	}

	hostConfig := &container.HostConfig{
		Mounts:     mounts,
		AutoRemove: true,
	}

	networkConfig := &network.NetworkingConfig{}

	createOptions := client.ContainerCreateOptions{
		Config:           containerConfig,
		HostConfig:       hostConfig,
		NetworkingConfig: networkConfig,
		Platform:         nil,
	}

	resp, err := d.client.ContainerCreate(ctx, createOptions)
	if err != nil {
		return "", err
	}

	return resp.ID, nil
}

// StartContainer starts a Docker container.
func (d *RealDockerClient) StartContainer(ctx context.Context, containerID string) error {
	startOptions := client.ContainerStartOptions{}
	_, err := d.client.ContainerStart(ctx, containerID, startOptions)
	return err
}

// ExecInContainer executes a command inside a container.
func (d *RealDockerClient) ExecInContainer(ctx context.Context, containerID string, cmd []string) (*ExecResult, error) {
	return nil, ErrNotImplemented
}

// StopContainer stops a Docker container.
func (d *RealDockerClient) StopContainer(ctx context.Context, containerID string) error {
	timeout := 10
	stopOptions := client.ContainerStopOptions{
		Timeout: &timeout,
	}
	_, err := d.client.ContainerStop(ctx, containerID, stopOptions)
	return err
}

// RemoveContainer removes a Docker container.
func (d *RealDockerClient) RemoveContainer(ctx context.Context, containerID string) error {
	removeOptions := client.ContainerRemoveOptions{
		Force: true,
	}
	_, err := d.client.ContainerRemove(ctx, containerID, removeOptions)
	return err
}

// PullImage pulls a Docker image.
func (d *RealDockerClient) PullImage(ctx context.Context, imageName string) error {
	renderer := ui.NewWorkflowRenderer()
	fmt.Fprintln(d.writer, renderer.RenderDockerOperation("Pulling image", imageName))

	pullOptions := client.ImagePullOptions{}
	reader, err := d.client.ImagePull(ctx, imageName, pullOptions)
	if err != nil {
		return err
	}
	defer reader.Close()

	// Track layers to avoid duplicate output
	layerStatus := make(map[string]string)
	decoder := json.NewDecoder(reader)

	for {
		var event struct {
			Status         string `json:"status"`
			Progress       string `json:"progress"`
			ProgressDetail struct {
				Current int64 `json:"current"`
				Total   int64 `json:"total"`
			} `json:"progressDetail"`
			ID string `json:"id"`
		}

		if err := decoder.Decode(&event); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		// Only show meaningful status changes
		if event.ID != "" {
			key := event.ID + event.Status
			if layerStatus[key] != event.Status {
				layerStatus[key] = event.Status

				// Show status without progress bar noise
				switch event.Status {
				case "Downloading", "Extracting":
					statusText := fmt.Sprintf("  %s: %s", event.ID[:12], event.Status)
					fmt.Fprintln(d.writer, ui.Muted.Render(statusText))
				case "Pull complete":
					statusText := fmt.Sprintf("  %s: Pull complete", event.ID[:12])
					fmt.Fprintln(d.writer, ui.Success.Render(statusText))
				}
			}
		} else if event.Status != "" {
			// Top-level status messages
			fmt.Fprintln(d.writer, ui.Info.Render("  "+event.Status))
		}
	}

	fmt.Fprintln(d.writer, ui.Success.Render("✓ Image pulled successfully"))
	return nil
}

// WaitForContainer waits for a container to finish and returns its exit code.
func (d *RealDockerClient) WaitForContainer(ctx context.Context, containerID string) (int, error) {
	return -1, ErrNotImplemented
}

// GetContainerLogs retrieves logs from a container.
func (d *RealDockerClient) GetContainerLogs(ctx context.Context, containerID string) (string, error) {
	options := client.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     false,
		Timestamps: false,
	}

	reader, err := d.client.ContainerLogs(ctx, containerID, options)
	if err != nil {
		return "", err
	}
	defer reader.Close()

	logs, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}

	return string(logs), nil
}

// Close closes the Docker client.
func (d *RealDockerClient) Close() error {
	return d.client.Close()
}

// Ping checks if the Docker daemon is responding.
func (d *RealDockerClient) Ping(ctx context.Context) (string, error) {
	pingOptions := client.PingOptions{}
	ping, err := d.client.Ping(ctx, pingOptions)
	if err != nil {
		return "", err
	}
	return ping.APIVersion, nil
}

// ErrNotImplemented is returned for methods not yet implemented in the Moby client
var ErrNotImplemented = fmt.Errorf("method not implemented in current Moby client version")
