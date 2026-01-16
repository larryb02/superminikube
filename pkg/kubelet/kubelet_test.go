package kubelet

import (
	"testing"

	"superminikube/pkg/api"
	"superminikube/pkg/client"
	"superminikube/pkg/kubelet/runtime"
)

var testKubelet *Kubelet

func TestMain(m *testing.M) {
	rt := runtime.FakeRuntime{

	}
	testKubelet = &Kubelet{
		client:  client.NewHTTPClient("http://localhost:8080", "test-node"),
		containerruntime: &rt,
	}
	m.Run()
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

// TODO
// func TestPodCleanup(t *testing.T) {
// 	t.Run("test pod cleanup", func(t *testing.T) {
// 		errs := testKubelet.Cleanup()
// 		if len(errs) > 0 {
// 			t.Errorf("cleanup failed: %v", errs)
// 		}
// 	})
// }
