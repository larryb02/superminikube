package runtime

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/jsonstream"
	"github.com/moby/moby/api/types/network"
	"github.com/moby/moby/client"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"

	"superminikube/pkg/api"
)

func (dr DockerRuntime) Ping(ctx context.Context) error {
	return nil
}

func (dr DockerRuntime) DeletePod(ctx context.Context, p api.Pod) error {
	cid := p.Spec.Container.ContainerId
	slog.Info("removing container", "containerid", cid)
	_, err := dr.containerruntime.ContainerRemove(ctx, cid, client.ContainerRemoveOptions{
		Force:         true,
		RemoveVolumes: true,
	})
	if err != nil {
		return fmt.Errorf("failed to remove container in pod: %v\nerr: %v", p.Uid, err)
	}
	return nil
}

func (dr DockerRuntime) CreatePod(ctx context.Context, spec api.PodSpec) error {
	// pull image
	pullOpts := client.ImagePullOptions{
		Platforms: []ocispec.Platform{{Architecture: "amd64", OS: "linux"}},
	}
	slog.Info("Attempting to pull", "image", spec.Container.Image)
	resp, err := dr.containerruntime.ImagePull(ctx, spec.Container.Image, pullOpts)
	if err != nil {
		return fmt.Errorf("failed to pull image: %v", err)
	}
	var pullErrs []*jsonstream.Error
	for m := range resp.JSONMessages(ctx) {
		if m.Error != nil {
			pullErrs = append(pullErrs, m.Error)
		} else {
			slog.Info("Status:", "status", m.Status)
		}
	}
	if len(pullErrs) > 0 {
		return fmt.Errorf("failed to pull image: %v", pullErrs)
	}
	containerOpts, err := PodSpecToCreateContainerOpts(spec)
	if err != nil {
		return fmt.Errorf("failed to create container opts: %v", err)
	}
	createRes, err := dr.containerruntime.ContainerCreate(ctx, containerOpts)
	if err != nil {
		return fmt.Errorf("failed to create container: %v", err)
	}
	slog.Info("Created", "container", createRes.ID)
	// start container
	_, err = dr.containerruntime.ContainerStart(ctx, createRes.ID, client.ContainerStartOptions{})
	if err != nil {
		return fmt.Errorf("failed to start container: %v", err)
	}
	slog.Info("Started", "container", createRes.ID)
	return nil
}

func NewDockerRuntime() (DockerRuntime, error) {
	cr, err := client.New(client.FromEnv)
	if err != nil {
		return DockerRuntime{}, err
	}
	return DockerRuntime{
		containerruntime: cr,
	}, nil
}

type DockerRuntime struct {
	containerruntime *client.Client
}

func (fr *FakeRuntime) Ping(ctx context.Context) error {
	return nil
}

func (fr *FakeRuntime) DeletePod(ctx context.Context, pod api.Pod) error {
	fr.DeletedPods = append(fr.DeletedPods, pod)
	return nil
}

func (fr *FakeRuntime) CreatePod(ctx context.Context, spec api.PodSpec) error {
	fr.CreatedPods = append(fr.CreatedPods, spec)
	return nil
}

type FakeRuntime struct {
	CreatedPods []api.PodSpec
	DeletedPods []api.Pod
}

type ContainerRuntime interface {
	Ping(context.Context) error
	// pulls image, creates, and starts container
	CreatePod(context.Context, api.PodSpec) error
	// kill and removes containers in pod
	DeletePod(context.Context, api.Pod) error
}

func PodSpecToCreateContainerOpts(spec api.PodSpec) (client.ContainerCreateOptions, error) {
	c := spec.Container

	// Convert env map to slice of "KEY=VALUE" strings
	env := make([]string, 0, len(c.Env))
	for k, v := range c.Env {
		env = append(env, k+"="+v)
	}

	// Convert ports to ExposedPorts and PortBindings
	exposedPorts := network.PortSet{}
	portBindings := network.PortMap{}
	for _, p := range c.Ports {
		containerPort, err := network.ParsePort(p.Containerport + "/tcp")
		if err != nil {
			return client.ContainerCreateOptions{}, err
		}
		exposedPorts[containerPort] = struct{}{}
		portBindings[containerPort] = []network.PortBinding{
			{HostPort: p.Hostport},
		}
	}

	// Convert volumes to map[string]struct{}
	volumes := make(map[string]struct{}, len(c.Volumes))
	for _, v := range c.Volumes {
		volumes[v] = struct{}{}
	}

	return client.ContainerCreateOptions{
		Image: c.Image,
		Config: &container.Config{
			Env:          env,
			ExposedPorts: exposedPorts,
			Volumes:      volumes,
		},
		HostConfig: &container.HostConfig{
			PortBindings: portBindings,
		},
	}, nil
}
