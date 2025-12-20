package kubelet

import (
	"fmt"
	"log/slog"
	"os"
	"superminikube/kubelet/runtime"
	"superminikube/spec"
	"superminikube/types/pod"
	// "superminikube/spec"
)

type Kubelet struct {
	runtime *runtime.DockerRuntime
	pods    []pod.Pod
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
		pods:    []pod.Pod{}, // agent will be assigned pods by controller eventually
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
		return fmt.Errorf("failed to create pod: %v", err)
	}
	k.LaunchPod(pod)
	return nil
}

func (k *Kubelet) LaunchPod(p *pod.Pod) {
	
}

// func (k *Kubelet) Start(args []string) {
// 	// sanity check
// 	err := k.runtime.Ping()
// 	if err != nil {
// 		slog.Error("Failed to reach Docker Engine", "error", err)
// 		slog.Error("Kubelet failed to start")
// 		return
// 	}
// 	slog.Info("Started Kubelet")
// 	specs, err := spec.CreateSpec(args[1])
// 	if err != nil {
// 		slog.Error("Failed to create spec", "msg", err)
// 		return
// 	}
// 	for i:= range specs {
// 		container_id, err := k.runtime.CreateContainer(specs[i])
// 		if err != nil {
// 			slog.Error("Failed to create container", "msg", err)
// 		}
// 		// How do we want to handle this behavior
// 		err = k.runtime.StartContainer(container_id)
// 		if err != nil {
// 			slog.Error("Failed to start container", "msg", err)
// 		}
// 	}
// }

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
}
