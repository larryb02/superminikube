package tests

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"superminikube/pkg/spec"
)

func TestContainerSpecDecode(t *testing.T) {
	cs := spec.ContainerSpec{
		Image: "nginx",
		Env: map[string]string{
			"FOO": "bar",
		},
		Ports: []spec.Port{
			{Hostport: "8080", Containerport: "80"},
		},
		Volumes: []string{"data"},
	}

	opts, err := cs.Decode()
	if err != nil {
		t.Fatalf("Decode() error = %v", err)
	}

	if opts.Config.Image != "nginx" {
		t.Errorf("expected Image 'nginx', got %s", opts.Config.Image)
	}

	if len(opts.Config.Env) != 1 || opts.Config.Env[0] != "FOO=bar" {
		t.Errorf("unexpected env: %v", opts.Config.Env)
	}

	if _, ok := opts.Config.Volumes["data"]; !ok {
		t.Errorf("expected volume 'data' to exist")
	}

	if len(opts.HostConfig.PortBindings) != 1 {
		t.Errorf("expected 1 port binding, got %d", len(opts.HostConfig.PortBindings))
	}
}
func TestCreateSpec(t *testing.T) {
	yamlContent := `
spec:
  - image: nginx
    env:
      FOO: bar
    ports:
      - hostport: "8080"
        containerport: "80"
    volumes:
      - data
`
	tmpDir := t.TempDir()
	specFile := filepath.Join(tmpDir, "test_spec.yaml")
	if err := os.WriteFile(specFile, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("failed to write temp spec file: %v", err)
	}

	specObj, err := spec.CreateSpec(specFile)
	if err != nil {
		t.Fatalf("CreateSpec() error = %v", err)
	}

	if len(specObj.ContainerSpec) != 1 {
		t.Fatalf("expected 1 container spec, got %d", len(specObj.ContainerSpec))
	}

	cs := specObj.ContainerSpec[0]
	if cs.Image != "nginx" {
		t.Errorf("expected image 'nginx', got %s", cs.Image)
	}
	if !reflect.DeepEqual(cs.Env, map[string]string{"FOO": "bar"}) {
		t.Errorf("unexpected env: %v", cs.Env)
	}
	if len(cs.Ports) != 1 || cs.Ports[0].Hostport != "8080" {
		t.Errorf("unexpected ports: %v", cs.Ports)
	}
	if len(cs.Volumes) != 1 || cs.Volumes[0] != "data" {
		t.Errorf("unexpected volumes: %v", cs.Volumes)
	}
}
