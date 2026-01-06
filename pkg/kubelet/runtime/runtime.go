package runtime

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/moby/moby/api/types/jsonstream"
	"github.com/moby/moby/client"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

type Runtime interface {
	Ping() error
	Pull(image string) error
	StartContainer(id string) error
	StopContainer(id string) error
	CreateContainer(opts client.ContainerCreateOptions) (string, error)
	CloseRuntime() error
	Inspect(id string) (client.ContainerInspectResult, error) // TODO: want a generic return value here
}

type DockerRuntime struct {
	client *client.Client
	ctx    context.Context
}

func NewDockerRuntime() (*DockerRuntime, error) {
	cli, err := client.New(client.FromEnv)
	if err != nil {
		slog.Error("Failed to create Docker Runtime") // maybe we don't want to shut down the application here?
		panic(err)
	}
	// things we want to store
	// API version
	//
	return &DockerRuntime{cli, context.Background()}, nil
}

func (d *DockerRuntime) CloseRuntime() error {
	err := d.client.Close()
	if err != nil {
		return err
	}
	return nil
}

func (d *DockerRuntime) Ping() error {
	opts := client.PingOptions{}
	resp, err := d.client.Ping(d.ctx, opts)
	if err != nil {
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
		return fmt.Errorf("failed to pull image: %v", err)
	}
	msg := resp.JSONMessages(d.ctx)
	errs := []*jsonstream.Error{}
	for m := range msg {
		if m.Error != nil {
			errs = append(errs, m.Error)
		} else {
			slog.Info("Status: ", "status", m.Status)
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("failed to pull image: %v", errs) // NOTE: will probably need to do some formatting
	}
	return nil
}

// This will return image ID
func (d *DockerRuntime) CreateContainer(opts client.ContainerCreateOptions) (string, error) {
	res, err := d.client.ContainerCreate(d.ctx, opts)
	if err != nil {
		slog.Error("Failed to create container", "msg", err)
		return "", err
	}
	slog.Info("Created ", "container", res.ID)
	return res.ID, nil
}

func (d *DockerRuntime) RemoveContainer(id string) error {
	opts := client.ContainerRemoveOptions{}
	_, err := d.client.ContainerRemove(d.ctx, id, opts)
	if err != nil {
		slog.Error("Failed to remove container: ", "error", err)
		return err
	}
	slog.Info("Removed container: ", "id", id)
	return nil
}

func (d *DockerRuntime) StartContainer(id string) error {
	opts := client.ContainerStartOptions{}
	_, err := d.client.ContainerStart(d.ctx, id, opts)
	if err != nil {
		slog.Error("Failed to start container: ", "id", id) //TODO: probably want ID, name, etc here later
		return err
	}
	slog.Info("Started ", "container", id)
	return nil
}

func (d *DockerRuntime) StopContainer(id string) error {
	opts := client.ContainerStopOptions{}
	_, err := d.client.ContainerStop(d.ctx, id, opts)
	if err != nil {
		slog.Error("Failed to stop container ", "id", id)
		return err
	}
	slog.Info("Stopped", "id", id)
	return nil
}

func (d *DockerRuntime) Inspect(id string) (client.ContainerInspectResult, error) {
	opts := client.ContainerInspectOptions{}
	res, err := d.client.ContainerInspect(d.ctx, id, opts)
	if err != nil {
		return client.ContainerInspectResult{}, err
	}
	return res, nil
}
