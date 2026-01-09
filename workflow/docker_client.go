package workflow

import (
	"context"
	"fmt"
	"io"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/api/types/network"
	"github.com/moby/moby/client"
)

// RealDockerClient implements DockerClient using the actual Docker SDK.
type RealDockerClient struct {
	client *client.Client
}

// NewDockerClient creates a new Docker client.
func NewDockerClient() (DockerClient, error) {
	cli, err := client.New(client.FromEnv)
	if err != nil {
		return nil, err
	}

	return &RealDockerClient{client: cli}, nil
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
	pullOptions := client.ImagePullOptions{}
	reader, err := d.client.ImagePull(ctx, imageName, pullOptions)
	if err != nil {
		return err
	}
	defer reader.Close()

	_, err = io.Copy(io.Discard, reader)

	return err
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
