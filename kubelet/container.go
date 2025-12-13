package kubelet

import (
	"context"
	"log"

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

func (d *DockerRuntime) Ping() error {
	opts := client.PingOptions{}
	resp, err := d.client.Ping(d.ctx, opts)
	if err != nil {
		log.Printf("Failed to ping Docker")
		return err
	}
	log.Printf("%+v", resp)
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
	log.Printf("Attempting to pull %s", image)
	resp, err := d.client.ImagePull(d.ctx, image, opts)
	if err != nil {
		log.Printf("Failed to pull image %s", err)
	}
	msg := resp.JSONMessages(d.ctx)
	for m := range msg {
		if m.Error != nil {
			log.Printf("Error: %v", m.Error.Message)
			return nil
		} else {
			log.Printf("Status: %s", m.Status)
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
		log.Printf("Failed to create container: %v", err)
		return "", err
	}
	log.Printf("Created %s", res.ID)
	return res.ID, nil
}

func (d *DockerRuntime) RemoveContainer(ID string) error {
	opts := client.ContainerRemoveOptions{}
	_, err := d.client.ContainerRemove(d.ctx, ID, opts)
	if err != nil {
		log.Printf("Failed to remove container: %v", err)
		return err
	}
	log.Printf("Removed container: %s", ID)
	return nil
}

func (d *DockerRuntime) StartContainer(ID string) error {
	opts := client.ContainerStartOptions{}
	_, err := d.client.ContainerStart(d.ctx, ID, opts)
	if err != nil {
		log.Printf("Failed to start container: %s", ID) //TODO: probably want ID, name, etc here later
		return err
	}
	log.Printf("Started %s", ID)
	return nil
}

func (d *DockerRuntime) StopContainer(ID string) error {
	opts := client.ContainerStopOptions{}
	_, err := d.client.ContainerStop(d.ctx, ID, opts)
	if err != nil {
		log.Printf("Failed to stop container %s", ID)
		return err
	}
	log.Printf("Stopped %s", ID)
	return nil
}
