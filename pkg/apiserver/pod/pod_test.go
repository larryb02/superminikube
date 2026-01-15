package pod

import (
	"os"
	"testing"

	"github.com/go-redis/redis/v8"

	"superminikube/pkg/api"
	"superminikube/pkg/apiserver/watch"
)

var testClient *redis.Client

func TestMain(m *testing.M) {
	testClient = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	code := m.Run()
	testClient.Close()
	os.Exit(code)
}

func TestCreatePod(t *testing.T) {
	testWatchService := watch.NewService()
	service := NewService(testClient, testWatchService)
	testCases := []struct {
		name        string
		nodename    string
		spec        api.PodSpec
		client      *redis.Client
		expectError bool
	}{
		{
			name:     "create pod with basic spec",
			nodename: "test-node-1",
			spec: api.PodSpec{
				Container: api.Container{
					Image: "nginx:latest",
					Env: map[string]string{
						"ENV_VAR": "test-value",
					},
				},
			},
			client:      testClient,
			expectError: false,
		},
		{
			name:     "create pod with empty nodename",
			nodename: "",
			spec: api.PodSpec{
				Container: api.Container{
					Image: "alpine:latest",
				},
			},
			client:      testClient,
			expectError: false,
		},
		{
			name:     "create pod with ports and volumes",
			nodename: "test-node-2",
			spec: api.PodSpec{
				Container: api.Container{
					Image: "redis:latest",
					Env: map[string]string{
						"REDIS_PORT": "6379",
					},
					Ports: []api.Port{
						{
							Hostport:      "8080",
							Containerport: "80",
						},
					},
					Volumes: []string{"/data"},
				},
			},
			client:      testClient,
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// TODO: writing to a non-existent channel at the moment. Check this out...
			_, err := service.CreatePod(t.Context(), tc.nodename, tc.spec)

			if tc.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
		})
	}
}

// TODO: once mocking is implemented
func TestGetPodByUid(t *testing.T) {

}
