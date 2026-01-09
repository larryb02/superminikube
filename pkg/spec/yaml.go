package spec

import (
	"errors"
	"fmt"
	"log/slog"
	"net/netip"
	"os"

	"superminikube/pkg/api"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/network"
	"github.com/moby/moby/client"
	"gopkg.in/yaml.v3"
)

// TODO: a bandaid patch until i revisit this code
func Validate(cs *api.ContainerSpec) error {
	if cs.Image == "" {
		return errors.New("image cannot be nil")
	}
	return nil
}

func Decode(cs *api.ContainerSpec) (client.ContainerCreateOptions, error) {
	// Convert ContainerSpec to client.CreateContainerOptions
	if err := Validate(cs); err != nil {
		slog.Error("Failed to decode spec: ", "msg", err)
		return client.ContainerCreateOptions{}, err
	}
	var env []string
	volumes := make(map[string]struct{})
	for k, v := range cs.Env {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}
	for _, volume := range cs.Volumes {
		volumes[volume] = struct{}{}
	}
	portMap := make(network.PortMap)
	for _, port := range cs.Ports {
		containerport, err := network.ParsePort(port.Containerport)
		if err != nil {
			slog.Error("Failed to configure port", "msg", err)
			return client.ContainerCreateOptions{}, err
		}
		portMap[containerport] = []network.PortBinding{{
			HostIP:   netip.Addr{},
			HostPort: port.Hostport,
		},
		}
	}
	opts := client.ContainerCreateOptions{
		Config: &container.Config{
			Image:   cs.Image,
			Env:     env,
			Volumes: volumes,
		},
		HostConfig: &container.HostConfig{
			PortBindings: portMap,
			// AutoRemove:   true,
		},
	}
	return opts, nil
}

func CreateSpec(specfile string) (*api.Spec, error) {
	var spec api.Spec
	slog.Info("Opening File", "path", specfile)
	data, err := os.ReadFile(specfile)
	if err != nil {
		slog.Error("Failed to read specfile", "msg", err)
		return nil, err
	}
	if err := yaml.Unmarshal(data, &spec); err != nil {
		slog.Error("Failed to parse spec file: ", "msg", err)
		return nil, err
	}
	// var containerOpts []client.ContainerCreateOptions
	// opts, err := spec.Decode()
	// if err != nil {
	// 	slog.Error("Error while decoding spec", "msg", err)
	// }
	// containerOpts = append(containerOpts, opts)
	// slog.Info("Created spec objects", "spec", spec)
	// return containerOpts, nil
	return &spec, nil
}
