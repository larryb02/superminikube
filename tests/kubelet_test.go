package tests

import (
	"context"
	"testing"

	"superminikube/kubelet"

	"github.com/moby/moby/client"
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
func (m *MockRuntime) Inspect(ID string) (client.ContainerInspectResult, error) {
	return client.ContainerInspectResult{}, nil
}
func (m *MockRuntime) StopContainer(ID string) error {
	return nil
}
func TestApply(t *testing.T) {
	// testSpecFile := "../example-configs/test-config.yml"
	ctx := context.TODO()
	rt := MockRuntime{}
	_, err := kubelet.NewKubelet(&rt, ctx)
	if err != nil {
		t.Fatalf("failed to create kubelet: %v", err)
	}
}
