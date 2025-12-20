package kubelet

import (
	"fmt"
	"log/slog"
	"os"
	"superminikube/kubelet/runtime"
	"superminikube/spec"
	"superminikube/types/pod"
)

type Kubelet struct {
	runtime *runtime.DockerRuntime
	pods    []*pod.Pod
}

func NewKubelet() (*Kubelet, error) {
	runtime, err := runtime.NewDockerRuntime()
	if err != nil {
		return nil, fmt.Errorf("Kubelet failed to start: %v", err)
	}
	// sanity check
	err = runtime.Ping()
	if err != nil {
		return nil, fmt.Errorf("Kubelet failed to start: %v", err)
	}
	return &Kubelet{
		runtime: runtime,
		pods:    []*pod.Pod{}, // agent will be assigned pods by controller eventually
	}, nil
}

func (k *Kubelet) Start(specfile string) error {
	// for now we're still just loading a specfile, but now we're creating Pods instead of individual containers
	// Load spec file
	specs, err := spec.CreateSpec(specfile)
	if err != nil {
		return fmt.Errorf("failed to start kubelet: %v", err)
	}

	pod, err := pod.NewPod(&specs.ContainerSpec)
	if err != nil {
		// looks like this because Start is a temp method -> not the end all be all for kubelet's behavior
		return fmt.Errorf("failed to create pod: %v", err)
	}
	err = k.LaunchPod(pod)
	if err != nil {
		return fmt.Errorf("failed to launch pod: %v", err)
	}
	return nil
}

func (k *Kubelet) LaunchPod(p *pod.Pod) error {
	err := k.runtime.Pull(p.ContainerSpec.Image)
	if err != nil {
		return fmt.Errorf("failed to launch pod: %v", err)
	}
	containerOpts, err := p.ContainerSpec.Decode()
	if err != nil {
		return fmt.Errorf("failed to launch pod: %v", err)
	}
	containerId, err := k.runtime.CreateContainer(containerOpts)
	if err != nil {
		return fmt.Errorf("failed to launch pod: %v", err)
	}
	err = k.runtime.StartContainer(containerId)
	if err != nil {
		return fmt.Errorf("failed to launch pod: %v", err)
	}
	return nil
}

func Run(args []string) {
	slog.Info("Starting Kubelet...")
	kubelet, err := NewKubelet()
	if err != nil {
		slog.Error("Failed to start Kubelet:", "error", err)
		os.Exit(1)
	}
	defer kubelet.runtime.CloseRuntime()
	err = kubelet.Start(args[1])
	if err != nil {
		slog.Error("Something went wrong: ", "msg", err)
		os.Exit(1)
	}
	slog.Info("Successfully launched pods... manage em yourself!")
}
