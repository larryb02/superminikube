package kubelet

import (
	"testing"

	mobyclient "github.com/moby/moby/client"

	"superminikube/pkg/api"
	"superminikube/pkg/client"
)

var testKubelet *Kubelet

func TestMain(m *testing.M) {
	rt, err := mobyclient.New(mobyclient.FromEnv)
	if err != nil {
		panic(err)
	}
	testKubelet = &Kubelet{
		client:  client.NewHTTPClient("http://localhost:8080", "test-node"),
		runtime: rt,
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
			_, err := testKubelet.handlePodCreate(t.Context(), p)
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
