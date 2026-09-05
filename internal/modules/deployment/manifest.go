package deployment

import (
	"bytes"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"text/template"
)

var dnsLabelPattern = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`)

type ManifestInput struct {
	Name       string
	Namespace  string
	Image      string
	Replicas   int32
	Resources  ResourceRequest
	EnvVars    map[string]string
	SecretRefs map[string]SecretRef
	Expose     ExposeRequest
}

type RenderedManifest struct {
	YAML           string `json:"yaml"`
	DeploymentName string `json:"deployment_name"`
	ServiceName    string `json:"service_name"`
	IngressName    string `json:"ingress_name,omitempty"`
}

type envVar struct {
	Name  string
	Value string
}

type secretEnvVar struct {
	Name       string
	SecretName string
	Key        string
}

type manifestTemplateData struct {
	ManifestInput
	ServiceName string
	IngressName string
	ServiceType string
	TargetPort  int32
	HasIngress  bool
	Path        string
	Env         []envVar
	Secrets     []secretEnvVar
}

func RenderManifest(input ManifestInput) (RenderedManifest, error) {
	normalized, err := normalizeManifestInput(input)
	if err != nil {
		return RenderedManifest{}, err
	}

	data := manifestTemplateData{
		ManifestInput: normalized,
		ServiceName:   normalized.Name,
		IngressName:   normalized.Name,
		ServiceType:   normalized.Expose.Type,
		TargetPort:    normalized.Expose.TargetPort,
		HasIngress:    normalized.Expose.Type == exposeTypeIngress,
		Path:          normalized.Expose.Path,
		Env:           sortedEnv(normalized.EnvVars),
		Secrets:       sortedSecrets(normalized.SecretRefs),
	}
	if data.HasIngress {
		data.ServiceType = exposeTypeClusterIP
	}

	tmpl, err := template.New("deployment-manifest").Funcs(template.FuncMap{"quote": strconv.Quote}).Parse(manifestTemplate)
	if err != nil {
		return RenderedManifest{}, err
	}
	var rendered bytes.Buffer
	if err := tmpl.Execute(&rendered, data); err != nil {
		return RenderedManifest{}, err
	}

	result := RenderedManifest{
		YAML:           rendered.String(),
		DeploymentName: normalized.Name,
		ServiceName:    data.ServiceName,
	}
	if data.HasIngress {
		result.IngressName = data.IngressName
	}
	return result, nil
}

func normalizeManifestInput(input ManifestInput) (ManifestInput, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.Namespace = strings.TrimSpace(input.Namespace)
	input.Image = strings.TrimSpace(input.Image)
	if input.Namespace == "" {
		input.Namespace = "default"
	}
	if input.Replicas <= 0 {
		input.Replicas = 1
	}
	if input.Resources.CPU == "" {
		input.Resources.CPU = "250m"
	}
	if input.Resources.Memory == "" {
		input.Resources.Memory = "256Mi"
	}
	if input.Expose.Type == "" {
		input.Expose.Type = exposeTypeClusterIP
	}
	if input.Expose.Port <= 0 {
		input.Expose.Port = 80
	}
	if input.Expose.TargetPort <= 0 {
		input.Expose.TargetPort = input.Expose.Port
	}
	if input.Expose.Path == "" {
		input.Expose.Path = "/"
	}

	if !validDNSLabel(input.Name) {
		return ManifestInput{}, fmt.Errorf("%w: name must be a valid Kubernetes DNS label", ErrInvalidManifestInput)
	}
	if !validDNSLabel(input.Namespace) {
		return ManifestInput{}, fmt.Errorf("%w: namespace must be a valid Kubernetes DNS label", ErrInvalidManifestInput)
	}
	if input.Image == "" {
		return ManifestInput{}, fmt.Errorf("%w: image is required", ErrInvalidManifestInput)
	}
	if input.Expose.Type != exposeTypeClusterIP && input.Expose.Type != exposeTypeNodePort && input.Expose.Type != exposeTypeLoadBalancer && input.Expose.Type != exposeTypeIngress {
		return ManifestInput{}, fmt.Errorf("%w: expose.type must be ClusterIP, NodePort, LoadBalancer, or Ingress", ErrInvalidManifestInput)
	}
	if input.Expose.Type == exposeTypeIngress && strings.TrimSpace(input.Expose.Host) == "" {
		return ManifestInput{}, fmt.Errorf("%w: expose.host is required for Ingress", ErrInvalidManifestInput)
	}
	return input, nil
}

func sortedEnv(values map[string]string) []envVar {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	items := make([]envVar, 0, len(keys))
	for _, key := range keys {
		items = append(items, envVar{Name: key, Value: values[key]})
	}
	return items
}

func sortedSecrets(values map[string]SecretRef) []secretEnvVar {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	items := make([]secretEnvVar, 0, len(keys))
	for _, key := range keys {
		items = append(items, secretEnvVar{Name: key, SecretName: values[key].SecretName, Key: values[key].Key})
	}
	return items
}

func validDNSLabel(value string) bool {
	return len(value) > 0 && len(value) <= 63 && dnsLabelPattern.MatchString(value)
}

const manifestTemplate = `apiVersion: v1
kind: Namespace
metadata:
  name: {{ .Namespace }}
  labels:
    app.kubernetes.io/managed-by: gidp
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ .Name }}
  namespace: {{ .Namespace }}
  labels:
    app.kubernetes.io/name: {{ .Name }}
    app.kubernetes.io/instance: {{ .Name }}
spec:
  replicas: {{ .Replicas }}
  selector:
    matchLabels:
      app.kubernetes.io/instance: {{ .Name }}
  template:
    metadata:
      labels:
        app.kubernetes.io/name: {{ .Name }}
        app.kubernetes.io/instance: {{ .Name }}
    spec:
      containers:
        - name: app
          image: {{ quote .Image }}
          imagePullPolicy: IfNotPresent
          ports:
            - containerPort: {{ .TargetPort }}
          resources:
            requests:
              cpu: {{ quote .Resources.CPU }}
              memory: {{ quote .Resources.Memory }}{{ if or .Env .Secrets }}
          env:{{ range .Env }}
            - name: {{ .Name }}
              value: {{ quote .Value }}{{ end }}{{ range .Secrets }}
            - name: {{ .Name }}
              valueFrom:
                secretKeyRef:
                  name: {{ .SecretName }}
                  key: {{ .Key }}{{ end }}{{ end }}
---
apiVersion: v1
kind: Service
metadata:
  name: {{ .ServiceName }}
  namespace: {{ .Namespace }}
  labels:
    app.kubernetes.io/name: {{ .Name }}
    app.kubernetes.io/instance: {{ .Name }}
spec:
  type: {{ .ServiceType }}
  selector:
    app.kubernetes.io/instance: {{ .Name }}
  ports:
    - port: {{ .Expose.Port }}
      targetPort: {{ .TargetPort }}
      protocol: TCP{{ if .HasIngress }}
---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: {{ .IngressName }}
  namespace: {{ .Namespace }}
  labels:
    app.kubernetes.io/name: {{ .Name }}
    app.kubernetes.io/instance: {{ .Name }}
spec:
  rules:
    - host: {{ .Expose.Host }}
      http:
        paths:
          - path: {{ .Path }}
            pathType: Prefix
            backend:
              service:
                name: {{ .ServiceName }}
                port:
                  number: {{ .Expose.Port }}{{ end }}
`
