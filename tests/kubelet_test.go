package tests

import (
	"testing"

	"github.com/moby/moby/client"
	"superminikube/kubelet"
)

type MockRuntime struct {
}

func (m *MockRuntime) Ping() error {
	return nil
}

func (m *MockRuntime) Pull(image string) error {
	return nil
}

func (m *MockRuntime) StartContainer(ID string) error {
	return nil
}

func (m *MockRuntime) CreateContainer(opts client.ContainerCreateOptions) (string, error) {
	return "container-id", nil
}

func (m *MockRuntime) CloseRuntime() error {
	return nil
}
func TestApply(t *testing.T) {
	testSpecFile := "../example-configs/test-config.yml"
	rt := MockRuntime{}
	k, err := kubelet.NewKubelet(&rt)
	if err != nil {
		t.Fatalf("failed to create kubelet: %v", err)
	}
	err = k.Apply(testSpecFile)
	if err != nil {
		t.Errorf("Apply() = %v", err)
	}
}
