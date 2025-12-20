package kubelet

import (
	"fmt"
	"log/slog"
	"os"
	"sync"

	"golang.org/x/sync/errgroup"
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

func (k *Kubelet) Apply(specfile string) error {
	// for now we're still just loading a specfile, but now we're creating Pods instead of individual containers
	// Load spec file
	specs, err := spec.CreateSpec(specfile)
	if err != nil {
		return fmt.Errorf("kubelet: %v", err)
	}
	var g errgroup.Group
	var mu sync.Mutex
	for _, spec := range specs.ContainerSpec {
		g.Go(func() error {
			pod, err := pod.NewPod(&spec)
			if err != nil {
				return fmt.Errorf("kubelet: %v", err)
			}
			err = k.LaunchPod(pod)
			if err != nil {
				return fmt.Errorf("kubelet: %v", err)
			}
			mu.Lock()
			k.pods = append(k.pods, pod)
			mu.Unlock()
			return nil
		})
	}
	return g.Wait()
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
	p.CurrentState = pod.PodRunning
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
	err = kubelet.Apply(args[1])
	if err != nil {
		slog.Error("Something went wrong: ", "msg", err)
		os.Exit(1)
	}
	slog.Info("Successfully launched pods... manage em yourself!")
}
