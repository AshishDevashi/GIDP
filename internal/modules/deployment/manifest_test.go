package deployment

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	utilyaml "k8s.io/apimachinery/pkg/util/yaml"
)

func TestRenderManifestIncludesIngressAndSecretRefs(t *testing.T) {
	rendered, err := RenderManifest(ManifestInput{
		Name:      "demo-api",
		Namespace: "default",
		Image:     "registry.example.com/team/demo-api:v1",
		Replicas:  2,
		Resources: ResourceRequest{CPU: "500m", Memory: "512Mi"},
		EnvVars: map[string]string{
			"APP_ENV": "production",
		},
		SecretRefs: map[string]SecretRef{
			"DATABASE_URL": {SecretName: "demo-api-secrets", Key: "database-url"},
		},
		Expose: ExposeRequest{Type: exposeTypeIngress, Port: 80, TargetPort: 8080, Host: "demo.example.com"},
	})
	if err != nil {
		t.Fatalf("RenderManifest() error = %v", err)
	}

	for _, want := range []string{
		"kind: Namespace",
		"name: default",
		"kind: Deployment",
		"image: \"registry.example.com/team/demo-api:v1\"",
		"replicas: 2",
		"kind: Service",
		"type: ClusterIP",
		"kind: Ingress",
		"host: demo.example.com",
		"secretKeyRef:",
		"name: demo-api-secrets",
	} {
		if !strings.Contains(rendered.YAML, want) {
			t.Fatalf("rendered manifest does not contain %q:\n%s", want, rendered.YAML)
		}
	}

	assertManifestDecodes(t, rendered.YAML)
}

func assertManifestDecodes(t *testing.T, manifest string) {
	t.Helper()
	decoder := utilyaml.NewYAMLOrJSONDecoder(bytes.NewReader([]byte(manifest)), 4096)
	for {
		var obj unstructured.Unstructured
		err := decoder.Decode(&obj)
		if errors.Is(err, io.EOF) {
			return
		}
		if err != nil {
			t.Fatalf("manifest should decode as Kubernetes YAML: %v\n%s", err, manifest)
		}
	}
}
