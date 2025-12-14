package kubelet

import (
	"context"
	"log/slog"

	"github.com/moby/moby/client"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

type DockerRuntime struct {
	client *client.Client
	ctx    context.Context
}

func NewDockerRuntime() (*DockerRuntime, error) {
	cli, err := client.New(client.FromEnv)
	if err != nil {
		panic(err)
	}
	return &DockerRuntime{cli, context.Background()}, nil
}

func (d *DockerRuntime) CloseRuntime() error {
	err := d.client.Close()
	if err != nil {
		slog.Error("Failed to close connection")
		return err
	}
	return nil
}

func (d *DockerRuntime) Ping() error {
	opts := client.PingOptions{}
	resp, err := d.client.Ping(d.ctx, opts)
	if err != nil {
		slog.Error("Failed to ping Docker")
		return err
	}
	slog.Info("Successfully pinged Docker", "response", resp)
	return nil
}

func (d *DockerRuntime) Pull(image string) error {
	opts := client.ImagePullOptions{Platforms: []ocispec.Platform{
		{
			Architecture: "amd64",
			OS:           "linux",
		},
	},
	}
	slog.Info("Attempting to pull ", "image", image)
	resp, err := d.client.ImagePull(d.ctx, image, opts)
	if err != nil {
		slog.Error("Failed to pull image ", "error", err)
	}
	msg := resp.JSONMessages(d.ctx)
	for m := range msg {
		if m.Error != nil {
			slog.Error("Error: ", "error", m.Error.Message)
			return nil
		} else {
			slog.Info("Status: ", "status", m.Status)
		}
	}
	return nil
}

// This will return image ID
func (d *DockerRuntime) CreateContainer() (string, error) {
	opts := client.ContainerCreateOptions{
		Image: "redis",
		Platform: &ocispec.Platform{
			Architecture: "amd64",
			OS:           "linux",
		},
	}
	res, err := d.client.ContainerCreate(d.ctx, opts)
	if err != nil {
		slog.Error("Failed to create container: %v", err)
		return "", err
	}
	slog.Info("Created ", "container", res.ID)
	return res.ID, nil
}

func (d *DockerRuntime) RemoveContainer(ID string) error {
	opts := client.ContainerRemoveOptions{}
	_, err := d.client.ContainerRemove(d.ctx, ID, opts)
	if err != nil {
		slog.Error("Failed to remove container: ", "error", err)
		return err
	}
	slog.Info("Removed container: ", "id", ID)
	return nil
}

func (d *DockerRuntime) StartContainer(ID string) error {
	opts := client.ContainerStartOptions{}
	_, err := d.client.ContainerStart(d.ctx, ID, opts)
	if err != nil {
		slog.Error("Failed to start container: ", "id", ID) //TODO: probably want ID, name, etc here later
		return err
	}
	slog.Info("Started ", "container", ID)
	return nil
}

func (d *DockerRuntime) StopContainer(ID string) error {
	opts := client.ContainerStopOptions{}
	_, err := d.client.ContainerStop(d.ctx, ID, opts)
	if err != nil {
		slog.Error("Failed to stop container ", "id", ID)
		return err
	}
	slog.Info("Stopped", "id", ID)
	return nil
}
