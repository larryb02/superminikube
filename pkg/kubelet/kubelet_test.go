package kubelet

import (
	"os"
	"testing"

	"github.com/google/uuid"

	"superminikube/pkg/api"
	"superminikube/pkg/kubelet/runtime"
)

var testKubelet *Kubelet

func TestMain(m *testing.M) {
	rt := &runtime.FakeRuntime{}
	testKubelet = NewKubeletWithRuntime("http://localhost:8080", "test-node", rt)

	code := m.Run()

	testKubelet.pods = map[uuid.UUID]api.Pod{}
	os.Exit(code)
}

func TestPodCreate(t *testing.T) {
	testPods := []api.PodSpec{
		{
			Container: api.Container{
				Image: "alpine",
			},
		},
		{
			Container: api.Container{
				Image: "nginx",
				Ports: []api.Port{
					{
						Hostport:      "8888",
						Containerport: "80",
					},
				},
			},
		},
	}
	t.Run("test pod create", func(t *testing.T) {
		for _, p := range testPods {
			_, err := testKubelet.containerruntime.CreatePod(t.Context(), p)
			if err != nil {
				t.Errorf("failed to create pod: %v", err)
			}
		}
	})
}

func TestListPods(t *testing.T) {
	// TODO: wiping the pods map even in a test can lead to unintended behavior
	testKubelet.pods = map[uuid.UUID]api.Pod{}

	pods := testKubelet.ListPods()
	if len(pods) != 0 {
		t.Errorf("expected 0 pods, got %d", len(pods))
	}

	testKubelet.AddPod(api.Pod{
		Uid:      uuid.New(),
		Nodename: "test-node",
		Spec:     api.PodSpec{Container: api.Container{Image: "alpine"}},
	})
	testKubelet.AddPod(api.Pod{
		Uid:      uuid.New(),
		Nodename: "test-node",
		Spec:     api.PodSpec{Container: api.Container{Image: "nginx"}},
	})

	pods = testKubelet.ListPods()
	if len(pods) != 2 {
		t.Errorf("expected 2 pods, got %d", len(pods))
	}
}
