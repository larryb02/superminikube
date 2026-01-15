package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"testing"
	"time"

	"superminikube/pkg/api"
	"superminikube/pkg/apiserver"
	"superminikube/pkg/kubelet"
	"superminikube/pkg/kubelet/runtime"
)

const (
	testAPIServerAddr = ":18080"
	testAPIServerURL  = "http://localhost:18080"
	testNodeName      = "test-node"
)

var (
	testServer   *apiserver.APIServer
	testKubelet  *kubelet.Kubelet
	fakeRuntime  *runtime.FakeRuntime
	cancelFunc   context.CancelFunc
)

func TestMain(m *testing.M) {
	var err error

	testServer, err = apiserver.NewAPIServer(apiserver.APIServerOpts{Addr: testAPIServerAddr})
	if err != nil {
		log.Fatalf("failed to create test server: %v", err)
	}
	testServer.Setup()
	go testServer.ListenAndServe()

	time.Sleep(100 * time.Millisecond)

	fakeRuntime = &runtime.FakeRuntime{}
	testKubelet = kubelet.NewKubeletWithRuntime(testAPIServerURL, testNodeName, fakeRuntime)

	ctx, cancel := context.WithCancel(context.Background())
	cancelFunc = cancel
	go testKubelet.Start(ctx)

	time.Sleep(100 * time.Millisecond)

	code := m.Run()

	cancelFunc()
	time.Sleep(100 * time.Millisecond)
	testServer.Shutdown()

	os.Exit(code)
}

func TestPodCreation(t *testing.T) {
	spec := api.PodSpec{
		Container: api.Container{
			Image: "nginx:latest",
			Env:   map[string]string{},
		},
	}

	body, err := json.Marshal(spec)
	if err != nil {
		t.Fatalf("failed to marshal spec: %v", err)
	}

	url := fmt.Sprintf("%s/api/v1/pod?nodename=%s", testAPIServerURL, testNodeName)
	resp, err := http.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("failed to create pod: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", resp.StatusCode)
	}

	var createdPod api.Pod
	if err := json.NewDecoder(resp.Body).Decode(&createdPod); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	t.Logf("Created pod with UID: %s", createdPod.Uid)

	if len(fakeRuntime.CreatedPods) != 1 {
		t.Fatalf("expected 1 created pod, got %d", len(fakeRuntime.CreatedPods))
	}
}
