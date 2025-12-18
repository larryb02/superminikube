package kubelet

import (
	"log/slog"
	"os"
	"superminikube/spec"
)

type Kubelet struct {
	runtime *DockerRuntime
}

func NewKubelet() (*Kubelet, error) {
	runtime, err := NewDockerRuntime()
	if err != nil {
		slog.Error("Failed to start Docker Runtime")
	}
	return &Kubelet{runtime: runtime}, nil
}

// A node will take a config file and start containers based on that specification
func (k *Kubelet) Start(args []string) {
	// sanity check
	err := k.runtime.Ping()
	if err != nil {
		slog.Error("Failed to reach Docker Engine", "error", err)
		slog.Error("Kubelet failed to start")
		return
	}
	slog.Info("Started Kubelet")
	specs, err := spec.CreateSpec(args[1])
	if err != nil {
		slog.Error("Failed to create spec", "msg", err)
		return
	}
	for i:= range specs {
		container_id, err := k.runtime.CreateContainer(specs[i])
		if err != nil {
			slog.Error("Failed to create container", "msg", err)
		}
		// How do we want to handle this behavior
		err = k.runtime.StartContainer(container_id)
		if err != nil {
			slog.Error("Failed to start container", "msg", err)
		}
	}
}

func Run(args []string) {
	slog.Info("Starting Kubelet...")
	kubelet, err := NewKubelet()
	if err != nil {
		slog.Error("Failed to start Kubelet:", "error", err)
		os.Exit(1)
	}
	kubelet.Start(args)
	defer kubelet.runtime.CloseRuntime()
}
