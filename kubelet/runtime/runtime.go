package runtime

import (
	"context"
	"fmt"
	"superminikube/logger"

	"github.com/moby/moby/api/types/jsonstream"
	"github.com/moby/moby/client"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

type Runtime interface {
	Ping() error
	Pull(image string) error
	StartContainer(ID string) error
	CreateContainer(opts client.ContainerCreateOptions) (string, error)
	CloseRuntime() error
}

type DockerRuntime struct {
	client *client.Client
	ctx    context.Context
}

func NewDockerRuntime() (*DockerRuntime, error) {
	cli, err := client.New(client.FromEnv)
	if err != nil {
		logger.Logger.Error("Failed to create Docker Runtime") // maybe we don't want to shut down the application here?
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
	logger.Logger.Info("Successfully pinged Docker", "response", resp)
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
	logger.Logger.Info("Attempting to pull ", "image", image)
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
			logger.Logger.Info("Status: ", "status", m.Status)
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
		logger.Logger.Error("Failed to create container", "msg", err)
		return "", err
	}
	logger.Logger.Info("Created ", "container", res.ID)
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
		logger.Logger.Error("Failed to start container: ", "id", ID) //TODO: probably want ID, name, etc here later
		return err
	}
	slog.Info("Started ", "container", ID)
	return nil
}

func (d *DockerRuntime) StopContainer(ID string) error {
	opts := client.ContainerStopOptions{}
	_, err := d.client.ContainerStop(d.ctx, ID, opts)
	if err != nil {
		logger.Logger.Error("Failed to stop container ", "id", ID)
		return err
	}
	logger.Logger.Info("Stopped", "id", ID)
	return nil
}
