package kubelet

import (
	"log/slog"
	"os"
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
func (k *Kubelet) Start() {
	// sanity check
	err := k.runtime.Ping()
	if err != nil {
		slog.Error("Failed to reach Docker Engine", "error", err)
		slog.Error("Kubelet failed to start")
		return
	}
	slog.Info("Started Kubelet")
}

func Run() {
	slog.Info("Starting Kubelet...")
	kubelet, err := NewKubelet()
	if err != nil {
		slog.Error("Failed to start Kubelet:", "error", err)
		os.Exit(1)
	}
	kubelet.Start()
}
