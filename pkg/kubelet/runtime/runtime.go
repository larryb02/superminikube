package runtime

import (
	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/network"
	"github.com/moby/moby/client"

	"superminikube/pkg/api"
)

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
