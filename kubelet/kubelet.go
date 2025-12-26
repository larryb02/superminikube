package kubelet

import (
	"fmt"
	"os"
	"sync"

	"golang.org/x/sync/errgroup"
	"superminikube/kubelet/runtime"
	"superminikube/logger"
	"superminikube/spec"
	"superminikube/types/pod"
)

type Kubelet struct {
	runtime runtime.Runtime
	pods    []*pod.Pod
}

func NewKubelet(rt runtime.Runtime) (*Kubelet, error) {
	// sanity check
	err := rt.Ping()
	if err != nil {
		return nil, fmt.Errorf("Kubelet failed to start: %v", err)
	}
	return &Kubelet{
		runtime: rt,
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

func (k *Kubelet) GetCurrentState() {
	// periodically poll and get pods state
	// if a pod has failed report for reconciliation
	// spawn group of goroutines that sleep every couple seconds
	// then checks status, if container failed send status to channel?
	// or just periodically send a status to channel and update pod status that way
	var g errgroup.Group
	ch := make(chan *pod.Pod) // TODO: need this channeld defined in the kubelet struct
	podsCopy := make([]*pod.Pod, len(k.pods))
	copy(podsCopy, k.pods)
	slog.Debug("Checking status of pods", "pods", podsCopy)
	for _, p := range k.pods {
		g.Go(func() error {
			status := p.CurrentState
			if status == pod.PodFailed {
				ch <- p
			}
			return nil
		})
	}
}

func (k *Kubelet) GetPods() ([]pod.Pod, error) {
	podsCopy := make([]pod.Pod, len(k.pods))
	for i, p := range k.pods {
		podsCopy[i] = *p
	}
	return podsCopy, nil
}

// Cleanup Kubelet if process killed/stopped
// stops and removes containers running in pods
func (k *Kubelet) Cleanup() error {
	stoppedContainers := make([]string, 0, len(k.pods)) // only store container ids for now
	for _, p := range k.pods {
		err := k.runtime.StopContainer(p.ContainerId)
		if err != nil {
			return fmt.Errorf("failed to stop container: %v", err)
		}
		stoppedContainers = append(stoppedContainers, p.ContainerId)
	}
	slog.Debug("containers stopped", "containers", stoppedContainers)
	return nil
}

func Run(args []string) {
	logger.Logger.Info("Starting Kubelet...")
	rt, err := runtime.NewDockerRuntime()
	if err != nil {
		logger.Logger.Error("Failed to start kubelet: ", "error", err)
		os.Exit(1)
	}
	kubelet, err := NewKubelet(rt)
	if err != nil {
		logger.Logger.Error("Failed to start Kubelet:", "error", err)
		os.Exit(1)
	}
	defer kubelet.runtime.CloseRuntime()
	err = kubelet.Apply(args[1])
	if err != nil {
		logger.Logger.Error("Something went wrong: ", "msg", err)
		os.Exit(1)
	}
	fmt.Printf("checking something: %v", err)
	slog.Info("Successfully launched pods... manage em yourself!")
	<-ctx.Done()
	kubelet.Cleanup()
	os.Exit(0)
}
