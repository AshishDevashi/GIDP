package kubernetes

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	utilyaml "k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
)

const fieldManager = "gidp"

type Client struct {
	typed   kubernetes.Interface
	dynamic dynamic.Interface
	mapper  meta.RESTMapper
}

type WorkloadStatus struct {
	State               string         `json:"state"`
	Replicas            int32          `json:"replicas"`
	ReadyReplicas       int32          `json:"ready_replicas"`
	AvailableReplicas   int32          `json:"available_replicas"`
	UnavailableReplicas int32          `json:"unavailable_replicas"`
	ObservedGeneration  int64          `json:"observed_generation"`
	Generation          int64          `json:"generation"`
	Pods                int            `json:"pods"`
	PodPhases           map[string]int `json:"pod_phases,omitempty"`
	Error               string         `json:"error,omitempty"`
}

func NewClient(kubeconfigPath string) (*Client, error) {
	cfg, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return nil, fmt.Errorf("kubernetes: load kubeconfig: %w", err)
	}
	typed, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("kubernetes: create typed client: %w", err)
	}
	dynamicClient, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("kubernetes: create dynamic client: %w", err)
	}
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("kubernetes: create discovery client: %w", err)
	}

	return &Client{
		typed:   typed,
		dynamic: dynamicClient,
		mapper:  restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(discoveryClient)),
	}, nil
}

func (c *Client) ApplyManifest(ctx context.Context, manifest string) error {
	objects, err := decodeManifest(manifest)
	if err != nil {
		return err
	}
	force := true
	for i := range objects {
		obj := objects[i]
		resource, err := c.resourceFor(obj)
		if err != nil {
			return err
		}
		payload, err := json.Marshal(obj.Object)
		if err != nil {
			return fmt.Errorf("kubernetes: encode %s/%s: %w", obj.GetKind(), obj.GetName(), err)
		}
		if _, err := resource.Patch(ctx, obj.GetName(), types.ApplyPatchType, payload, metav1.PatchOptions{FieldManager: fieldManager, Force: &force}); err != nil {
			return fmt.Errorf("kubernetes: apply %s/%s: %w", obj.GetKind(), obj.GetName(), err)
		}
	}
	return nil
}

func (c *Client) DeleteManifest(ctx context.Context, manifest string) error {
	objects, err := decodeManifest(manifest)
	if err != nil {
		return err
	}
	for i := len(objects) - 1; i >= 0; i-- {
		obj := objects[i]
		resource, err := c.resourceFor(obj)
		if err != nil {
			return err
		}
		if err := resource.Delete(ctx, obj.GetName(), metav1.DeleteOptions{}); err != nil && !apierrors.IsNotFound(err) {
			return fmt.Errorf("kubernetes: delete %s/%s: %w", obj.GetKind(), obj.GetName(), err)
		}
	}
	return nil
}

func (c *Client) WorkloadStatus(ctx context.Context, namespace, deploymentName string) (WorkloadStatus, error) {
	dep, err := c.typed.AppsV1().Deployments(namespace).Get(ctx, deploymentName, metav1.GetOptions{})
	if err != nil {
		return WorkloadStatus{}, fmt.Errorf("kubernetes: get deployment status: %w", err)
	}

	status := WorkloadStatus{
		State:               "deploying",
		Replicas:            dep.Status.Replicas,
		ReadyReplicas:       dep.Status.ReadyReplicas,
		AvailableReplicas:   dep.Status.AvailableReplicas,
		UnavailableReplicas: dep.Status.UnavailableReplicas,
		ObservedGeneration:  dep.Status.ObservedGeneration,
		Generation:          dep.Generation,
		PodPhases:           map[string]int{},
	}
	desired := int32(1)
	if dep.Spec.Replicas != nil {
		desired = *dep.Spec.Replicas
	}
	if dep.Status.ObservedGeneration >= dep.Generation && dep.Status.AvailableReplicas >= desired {
		status.State = "running"
	}

	pods, err := c.typed.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labels.Set{"app.kubernetes.io/instance": deploymentName}.String(),
	})
	if err != nil {
		return status, fmt.Errorf("kubernetes: list deployment pods: %w", err)
	}
	status.Pods = len(pods.Items)
	for _, pod := range pods.Items {
		status.PodPhases[string(pod.Status.Phase)]++
	}
	return status, nil
}

func decodeManifest(manifest string) ([]unstructured.Unstructured, error) {
	decoder := utilyaml.NewYAMLOrJSONDecoder(bytes.NewReader([]byte(manifest)), 4096)
	objects := []unstructured.Unstructured{}
	for {
		var obj unstructured.Unstructured
		err := decoder.Decode(&obj)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("kubernetes: decode manifest: %w", err)
		}
		if len(obj.Object) == 0 {
			continue
		}
		if obj.GetName() == "" || obj.GetKind() == "" {
			return nil, fmt.Errorf("kubernetes: manifest object is missing name or kind")
		}
		objects = append(objects, obj)
	}
	return objects, nil
}

func (c *Client) resourceFor(obj unstructured.Unstructured) (dynamic.ResourceInterface, error) {
	gvk := obj.GroupVersionKind()
	mapping, err := c.mapper.RESTMapping(schema.GroupKind{Group: gvk.Group, Kind: gvk.Kind}, gvk.Version)
	if err != nil {
		return nil, fmt.Errorf("kubernetes: map %s: %w", gvk.String(), err)
	}
	resource := c.dynamic.Resource(mapping.Resource)
	if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
		namespace := obj.GetNamespace()
		if namespace == "" {
			namespace = "default"
		}
		return resource.Namespace(namespace), nil
	}
	return resource, nil
}
