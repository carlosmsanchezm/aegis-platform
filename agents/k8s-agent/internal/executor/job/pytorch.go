package job

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"

	execiface "github.com/yourorg/aegis/agents/k8s-agent/internal/executor"
	"github.com/yourorg/aegis/agents/k8s-agent/internal/kube"
	aegis "github.com/yourorg/aegis/proto/aegis/v1"
)

const trainingTimeout = 12 * time.Hour

func (e *Executor) RunTraining(ctx context.Context, w *aegis.Workload) execiface.Result {
	backend := "training"
	tr := w.GetTraining()
	if tr == nil {
		return execiface.Result{Status: "FAILED", Backend: backend, Err: fmt.Errorf("training spec missing"), URL: ""}
	}

	image := tr.GetImage()
	if image == "" {
		image = getenv("AEGIS_TRAINING_DEFAULT_IMAGE", "pytorch/pytorch:2.4.0-cuda11.8-cudnn8-runtime")
	}
	workers := int(tr.GetWorkers())
	if workers <= 0 {
		workers = 1
	}
	gpusPer := int(tr.GetGpusPerWorker())
	if gpusPer <= 0 {
		gpusPer = 1
	}

	_, cfg, err := kube.New()
	if err != nil {
		return execiface.Result{Status: "FAILED", Backend: backend, Err: fmt.Errorf("kube config: %w", err)}
	}
	dyn, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return execiface.Result{Status: "FAILED", Backend: backend, Err: fmt.Errorf("dynamic client: %w", err)}
	}

	if gvrV2, err := detectTrainerV2(cfg); err == nil {
		backend = "trainer_v2"
		name := sanitizeName("aegis-" + w.GetId())
		url := trainJobURL(e.namespace, name)
		runtimeKind := getenv("AEGIS_TRAINING_RUNTIME_KIND", "ClusterTrainingRuntime")
		runtimeName := getenv("AEGIS_TRAINING_RUNTIME_NAME", "torch-distributed")
		runtimeAPI := getenv("AEGIS_TRAINING_RUNTIME_API", "trainer.kubeflow.org")

		obj := buildTrainJobV2(name, e.namespace, image, tr.GetCommand(), runtimeAPI, runtimeKind, runtimeName, e.dryRun)
		res := dyn.Resource(gvrV2).Namespace(e.namespace)
		if _, err := res.Create(ctx, obj, metav1.CreateOptions{}); err != nil {
			e.log.Warn("trainjob create failed", zap.String("name", name), zap.Error(err))
			return execiface.Result{Status: "FAILED", URL: url, Backend: backend, Err: err}
		}

		e.log.Info("trainjob submitted",
			zap.String("name", name),
			zap.String("namespace", e.namespace),
			zap.Int("workers", workers),
			zap.Int("gpus_per_worker", gpusPer),
			zap.String("image", image),
			zap.String("runtime_kind", runtimeKind),
			zap.String("runtime_name", runtimeName),
			zap.String("url", url),
		)

		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		deadline := time.After(trainingTimeout)
		for {
			select {
			case <-ctx.Done():
				e.log.Warn("trainjob context cancelled", zap.String("name", name))
				return execiface.Result{Status: "FAILED", URL: url, Backend: backend}
			case <-deadline:
				e.log.Warn("trainjob timeout", zap.String("name", name))
				return execiface.Result{Status: "FAILED", URL: url, Backend: backend}
			case <-ticker.C:
				u, err := res.Get(ctx, name, metav1.GetOptions{})
				if err != nil {
					continue
				}
				if hasAnySuccess(u) {
					e.log.Info("trainjob succeeded", zap.String("name", name))
					return execiface.Result{Status: "SUCCEEDED", URL: url, Backend: backend}
				}
				if hasAnyFailure(u) {
					e.log.Info("trainjob failed", zap.String("name", name))
					return execiface.Result{Status: "FAILED", URL: url, Backend: backend}
				}
			}
		}
	}

	gv, gvr, err := detectPyTorchGVR(cfg)
	if err != nil {
		e.log.Warn("PyTorchJob CRD not found", zap.Error(err))
		return execiface.Result{Status: "FAILED", Backend: backend, Err: fmt.Errorf("training operator not installed")}
	}

	backend = "pytorch_v1"
	name := sanitizeName("aegis-" + w.GetId())
	url := fmt.Sprintf("k8s://%s/pytorchjob/%s", e.namespace, name)

	obj := buildPyTorchJob(gv, name, e.namespace, image, tr.GetCommand(), tr.GetFlavor(), gpusPer, workers, e.dryRun)
	res := dyn.Resource(gvr).Namespace(e.namespace)
	if _, err := res.Create(ctx, obj, metav1.CreateOptions{}); err != nil {
		if !kerrors.IsAlreadyExists(err) {
			e.log.Warn("pytorchjob create failed", zap.String("name", name), zap.Error(err))
			return execiface.Result{Status: "FAILED", URL: url, Backend: backend, Err: err}
		}
	}

	e.log.Info("pytorchjob submitted",
		zap.String("name", name),
		zap.String("namespace", e.namespace),
		zap.Int("workers", workers),
		zap.Int("gpus_per_worker", gpusPer),
		zap.String("image", image),
		zap.String("url", url),
	)

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	deadline := time.After(trainingTimeout)
	for {
		select {
		case <-ctx.Done():
			e.log.Warn("training context cancelled", zap.String("name", name))
			return execiface.Result{Status: "FAILED", URL: url, Backend: backend}
		case <-deadline:
			e.log.Warn("pytorchjob timeout", zap.String("name", name))
			return execiface.Result{Status: "FAILED", URL: url, Backend: backend}
		case <-ticker.C:
			u, err := res.Get(ctx, name, metav1.GetOptions{})
			if err != nil {
				continue
			}
			if hasCondition(u, "Succeeded") {
				e.log.Info("pytorchjob succeeded", zap.String("name", name))
				return execiface.Result{Status: "SUCCEEDED", URL: url, Backend: backend}
			}
			if hasCondition(u, "Failed") {
				e.log.Info("pytorchjob failed", zap.String("name", name))
				return execiface.Result{Status: "FAILED", URL: url, Backend: backend}
			}
		}
	}
}
func detectPyTorchGVR(cfg *rest.Config) (string, schema.GroupVersionResource, error) {
	disc, err := discovery.NewDiscoveryClientForConfig(cfg)
	if err != nil {
		return "", schema.GroupVersionResource{}, err
	}
	candidates := []string{"kubeflow.org/v1", "training.kubeflow.org/v1"}
	for _, gv := range candidates {
		if _, err := disc.ServerResourcesForGroupVersion(gv); err == nil {
			parts := strings.Split(gv, "/")
			if len(parts) != 2 {
				continue
			}
			return gv, schema.GroupVersionResource{
				Group:    parts[0],
				Version:  parts[1],
				Resource: "pytorchjobs",
			}, nil
		}
	}
	return "", schema.GroupVersionResource{}, fmt.Errorf("pytorchjobs CRD not found")
}

func buildPyTorchJob(apiVersion, name, namespace, image string, command []string, flavor string, gpusPerWorker, workers int, dryRun bool) *unstructured.Unstructured {
	container := map[string]interface{}{
		"name":  "pytorch",
		"image": image,
	}
	if len(command) > 0 {
		container["command"] = stringSliceToAny(command)
	}
	if !dryRun {
		resName := autoGPUResource(flavor)
		if resName == "" {
			resName = "nvidia.com/gpu"
		}
		quantity := resource.MustParse(strconv.Itoa(max(1, gpusPerWorker)))
		limits := map[string]interface{}{resName: quantity.String()}
		container["resources"] = map[string]interface{}{
			"limits":   limits,
			"requests": limits,
		}
	}

	template := map[string]interface{}{
		"spec": map[string]interface{}{
			"restartPolicy": "Never",
			"containers":    []interface{}{container},
		},
	}

	replicas := map[string]interface{}{
		"Master": map[string]interface{}{
			"replicas": int64(1),
			"template": template,
		},
	}

	if workers > 1 {
		replicas["Worker"] = map[string]interface{}{
			"replicas": int64(workers - 1),
			"template": template,
		}
	}

	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": apiVersion,
			"kind":       "PyTorchJob",
			"metadata": map[string]interface{}{
				"name":      name,
				"namespace": namespace,
				"labels": map[string]interface{}{
					"aegis.workload/id": name,
				},
			},
			"spec": map[string]interface{}{
				"pytorchReplicaSpecs": replicas,
			},
		},
	}
}

func detectTrainerV2(cfg *rest.Config) (schema.GroupVersionResource, error) {
	disc, err := discovery.NewDiscoveryClientForConfig(cfg)
	if err != nil {
		return schema.GroupVersionResource{}, err
	}
	const gv = "trainer.kubeflow.org/v1alpha1"
	if _, err := disc.ServerResourcesForGroupVersion(gv); err == nil {
		return schema.GroupVersionResource{
			Group:    "trainer.kubeflow.org",
			Version:  "v1alpha1",
			Resource: "trainjobs",
		}, nil
	}
	return schema.GroupVersionResource{}, fmt.Errorf("trainjobs CRD not found")
}

func trainJobURL(ns, name string) string {
	return fmt.Sprintf("k8s://%s/trainjob/%s", ns, name)
}

func buildTrainJobV2(name, namespace, image string, command []string, apiGroup, kind, runtimeName string, dryRun bool) *unstructured.Unstructured {
	container := map[string]interface{}{
		"name":  "trainer",
		"image": image,
	}
	if len(command) > 0 {
		container["command"] = stringSliceToAny(command)
	}

	templateSpec := map[string]interface{}{
		"containers": []interface{}{container},
	}
	if dryRun {
		templateSpec["restartPolicy"] = "Never"
	}

	spec := map[string]interface{}{
		"runtimeRef": map[string]interface{}{
			"apiGroup": apiGroup,
			"kind":     kind,
			"name":     runtimeName,
		},
		"template": map[string]interface{}{
			"spec": templateSpec,
		},
		"runPolicy": map[string]interface{}{
			"cleanPodPolicy": "None",
		},
	}

	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "trainer.kubeflow.org/v1alpha1",
			"kind":       "TrainJob",
			"metadata": map[string]interface{}{
				"name":      name,
				"namespace": namespace,
				"labels": map[string]interface{}{
					"aegis.workload/id": name,
				},
			},
			"spec": spec,
		},
	}
}

func hasAnySuccess(u *unstructured.Unstructured) bool {
	return hasCondition(u, "Complete") || hasCondition(u, "Succeeded")
}

func hasAnyFailure(u *unstructured.Unstructured) bool {
	return hasCondition(u, "Failed")
}

func hasCondition(u *unstructured.Unstructured, condType string) bool {
	conds, found, err := unstructured.NestedSlice(u.Object, "status", "conditions")
	if err != nil || !found {
		return false
	}
	for _, c := range conds {
		if m, ok := c.(map[string]interface{}); ok {
			t, _ := m["type"].(string)
			s, _ := m["status"].(string)
			if t == condType && strings.EqualFold(s, "true") {
				return true
			}
		}
	}
	return false
}

func autoGPUResource(flavor string) string {
	if strings.HasPrefix(flavor, "mig-") {
		return "nvidia.com/" + strings.TrimPrefix(flavor, "mig-")
	}
	if flavor != "" {
		return "nvidia.com/gpu"
	}
	return ""
}

func stringSliceToAny(in []string) []interface{} {
	out := make([]interface{}, len(in))
	for i := range in {
		out[i] = in[i]
	}
	return out
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
