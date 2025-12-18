package spec

import (
	"fmt"
	"log/slog"
	"net/netip"
	"os"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/network"
	"github.com/moby/moby/client"
	"gopkg.in/yaml.v3"
)

type Spec struct {
	ContainerSpec ContainerSpec `yaml:"spec"`
}

type ContainerSpec struct {
	Image   string            `yaml:"image"` // required
	Env     map[string]string `yaml:"env,omitempty"`
	Ports   []Port            `yaml:"ports,omitempty"` // make this a map or a string of type "hostport:containerport"
	Volumes []string           `yaml:"volumes,omitempty"` // only supporting empty volumes at the moment, may also want to move this to Pod level
}

type Port struct {
	Hostport      string `yaml:"hostport"`
	Containerport string `yaml:"containerport"`
}


func (s *Spec) Decode() (client.ContainerCreateOptions, error) {
	// Convert ContainerSpec to client.CreateContainerOptions
	slog.Info("Decoding 'config.yml' (probably want the actual file here or somewhere in logs)", "spec", s)
	var env []string
	volumes := make(map[string]struct{})
	for k, v := range s.ContainerSpec.Env {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}
	for _, volume := range s.ContainerSpec.Volumes {
		volumes[volume] = struct{}{}
	}
	portMap := make(network.PortMap)
	for _, port := range s.ContainerSpec.Ports {
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
			Image:   s.ContainerSpec.Image,
			Env:     env,
			Volumes: volumes,
		},
		HostConfig: &container.HostConfig{
			PortBindings: portMap,
			AutoRemove:   true,
		},
	}
	return opts, nil
}

func CreateSpec(specfile string) ([]client.ContainerCreateOptions, error) {
	var spec Spec
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
	var containerOpts []client.ContainerCreateOptions
	opts, err := spec.Decode()
	if err != nil {
		slog.Error("Error while decoding spec", "msg", err)
	}
	containerOpts = append(containerOpts, opts)
	slog.Info("Created spec objects", "spec", spec)
	return containerOpts, nil
}
