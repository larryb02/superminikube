package apiserver

import (
	"testing"

	"superminikube/pkg/api"

	"github.com/go-redis/redis/v8"
)
// TODO: prefill test client with some data
var testClient = redis.NewClient(&redis.Options{
	Addr: "host.docker.internal:6379",
},)

func TestCreatePod(t *testing.T) {
	testCases := []struct {
		name        string
		nodename    string
		spec        *api.ContainerSpec
		client      *redis.Client
		expectError bool
	}{
		{
			name:     "create pod with basic spec",
			nodename: "test-node-1",
			spec: &api.ContainerSpec{
				Image: "nginx:latest",
				Env: map[string]string{
					"ENV_VAR": "test-value",
				},
			},
			client:      testClient,
			expectError: false,
		},
		{
			name:     "create pod with empty nodename",
			nodename: "",
			spec: &api.ContainerSpec{
				Image: "alpine:latest",
			},
			client:      testClient,
			expectError: false,
		},
		{
			name:     "create pod with ports and volumes",
			nodename: "test-node-2",
			spec: &api.ContainerSpec{
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
			client:      testClient,
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := CreatePod(t.Context(), tc.nodename, tc.spec, tc.client)

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
