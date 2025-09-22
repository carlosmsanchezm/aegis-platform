package job

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"

	execiface "github.com/yourorg/aegis/agents/k8s-agent/internal/executor"
	"github.com/yourorg/aegis/agents/k8s-agent/internal/kube"
	aegis "github.com/yourorg/aegis/proto/aegis/v1"
)

const (
	defaultNamespace = "default"
	jobTimeout       = 30 * time.Minute
)

type Executor struct {
	log       *zap.Logger
	namespace string
	dryRun    bool
	cleanup   bool
	client    *kube.Clientset
	restCfg   *rest.Config
	kueueOn   bool
	kueueQ    string
}

func New(log *zap.Logger) (*Executor, error) {
	if log == nil {
		log = zap.NewNop()
	}
	cs, cfg, err := kube.New()
	if err != nil {
		return nil, err
	}
	ns := getenv("AEGIS_NAMESPACE", defaultNamespace)
	e := &Executor{
		log:       log,
		namespace: ns,
		dryRun:    os.Getenv("AEGIS_DRY_RUN") == "1",
		cleanup:   os.Getenv("AEGIS_CLEANUP_JOBS") == "1",
		client:    cs,
		restCfg:   cfg,
	}

	e.kueueQ = getenv("AEGIS_KUEUE_QUEUE", "")
	if os.Getenv("AEGIS_KUEUE_ENABLED") == "1" && cfg != nil {
		if ok, derr := detectKueue(cfg); derr != nil {
			e.log.Warn("kueue detection failed", zap.Error(derr))
		} else if ok {
			e.kueueOn = true
			queueName := e.kueueQ
			if queueName == "" {
				queueName = "<workload-queue>"
			}
			e.log.Info("kueue admission enabled",
				zap.String("namespace", e.namespace),
				zap.String("queue", queueName),
				zap.Bool("using_workload_queue", e.kueueQ == ""),
			)
		} else {
			e.log.Info("kueue CRDs not detected; continuing without admission")
		}
	}

	return e, nil
}

func detectKueue(cfg *rest.Config) (bool, error) {
	disc, err := discovery.NewDiscoveryClientForConfig(cfg)
	if err != nil {
		return false, err
	}
	versions := []string{"kueue.x-k8s.io/v1beta1", "kueue.x-k8s.io/v1alpha2"}
	for _, gv := range versions {
		if _, err := disc.ServerResourcesForGroupVersion(gv); err == nil {
			return true, nil
		}
	}
	return false, nil
}

func (e *Executor) RunWorkspace(ctx context.Context, w *aegis.Workload) execiface.Result {
	ws := w.GetWorkspace()
	if ws == nil {
		return execiface.Result{Status: "FAILED", Backend: "workspace", Err: fmt.Errorf("workspace spec missing")}
	}

	image := ws.GetImage()
	if image == "" {
		image = getenv("AEGIS_DEFAULT_IMAGE", "alpine:3.19")
	}

	flavor := ws.GetFlavor()
	if flavor == "" {
		e.log.Warn("workspace flavor missing; relying on dry-run behavior", zap.String("workload_id", w.GetId()))
	}

	jobName := sanitizeName("aegis-" + w.GetId())
	url := fmt.Sprintf("k8s://%s/job/%s", e.namespace, jobName)

	env := toEnvVars(ws.GetEnv())
	env = append(env,
		corev1.EnvVar{Name: "AEGIS_WORKLOAD_ID", Value: w.GetId()},
		corev1.EnvVar{Name: "POD_NAME", ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.name"}}},
		corev1.EnvVar{Name: "POD_NAMESPACE", ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.namespace"}}},
		corev1.EnvVar{Name: "NODE_NAME", ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "spec.nodeName"}}},
	)

	defaultCmd := `echo "[AEGIS] start workload=$AEGIS_WORKLOAD_ID pod=$POD_NAME ns=$POD_NAMESPACE node=$NODE_NAME";
date;
echo "[AEGIS] image=` + image + `";
echo "[AEGIS] running task...";
sleep 3;
echo "[AEGIS] done";`

	cmd := ws.GetCommand()
	reqs := e.gpuRequests(w)
	container := corev1.Container{
		Name:                     "workspace",
		Image:                    image,
		Env:                      env,
		TerminationMessagePolicy: corev1.TerminationMessageFallbackToLogsOnError,
	}
	if len(reqs) > 0 {
		container.Resources = corev1.ResourceRequirements{Requests: reqs, Limits: reqs}
	}
	if len(cmd) > 0 {
		container.Command = cmd
	} else {
		container.Command = []string{"/bin/sh", "-c"}
		container.Args = []string{getenv("AEGIS_WORKSPACE_COMMAND", defaultCmd)}
	}

	if h := w.GetHints(); h != nil {
		e.log.Info("workspace applying resource hints",
			zap.String("workload_id", w.GetId()),
			zap.String("resource_name", h.GetResourceName()),
			zap.Int32("gpu_count", h.GetGpuCount()),
		)
	}

	if e.dryRun {
		e.log.Debug("dry-run enabled; omitting GPU resource requests",
			zap.String("workload_id", w.GetId()),
			zap.String("job", jobName),
		)
	}

	e.log.Debug("workspace container prepared",
		zap.String("workload_id", w.GetId()),
		zap.String("job", jobName),
		zap.String("image", image),
		zap.Int("env_count", len(env)),
		zap.String("flavor", flavor),
	)

	jobLabels := map[string]string{
		"aegis.workload/id": w.GetId(),
		"aegis.job/name":    jobName,
	}
	podLabels := map[string]string{
		"aegis.workload/id": w.GetId(),
	}
	suspend := false
	if e.kueueOn {
		queueName := e.kueueQ
		if queueName == "" {
			queueName = w.GetQueue()
		}
		if queueName == "" {
			e.log.Warn("kueue enabled but queue missing; skipping admission",
				zap.String("workload_id", w.GetId()),
				zap.String("job", jobName),
			)
		} else {
			jobLabels["kueue.x-k8s.io/queue-name"] = queueName
			podLabels["kueue.x-k8s.io/queue-name"] = queueName
			suspend = true
			e.log.Info("kueue admission requested",
				zap.String("workload_id", w.GetId()),
				zap.String("job", jobName),
				zap.String("queue", queueName),
				zap.Bool("suspend", suspend),
			)
		}
	}

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: e.namespace,
			Labels:    jobLabels,
		},
		Spec: batchv1.JobSpec{
			Suspend:      boolPtr(suspend),
			BackoffLimit: int32ptr(0),
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: podLabels},
				Spec: corev1.PodSpec{
					RestartPolicy: corev1.RestartPolicyNever,
					Containers:    []corev1.Container{container},
				},
			},
		},
	}

	jobs := e.client.BatchV1().Jobs(e.namespace)
	created, err := jobs.Create(ctx, job, metav1.CreateOptions{})
	if err != nil {
		if !kerrors.IsAlreadyExists(err) {
			e.log.Warn("job create failed", zap.String("job", jobName), zap.Error(err))
			return execiface.Result{Status: "FAILED", URL: url, Backend: "workspace", Err: err}
		}
		created, err = jobs.Get(ctx, jobName, metav1.GetOptions{})
		if err != nil {
			return execiface.Result{Status: "FAILED", URL: url, Backend: "workspace", Err: err}
		}
	}

	e.log.Info("job submitted",
		zap.String("job", created.GetName()),
		zap.String("namespace", e.namespace),
		zap.String("workload_id", w.GetId()),
		zap.String("flavor", flavor),
		zap.String("url", url),
	)

	status := e.waitForCompletion(ctx, jobName)

	if e.cleanup {
		propagation := metav1.DeletePropagationBackground
		if err := jobs.Delete(context.Background(), jobName, metav1.DeleteOptions{PropagationPolicy: &propagation}); err != nil && !kerrors.IsNotFound(err) {
			e.log.Warn("job cleanup failed", zap.String("job", jobName), zap.Error(err))
		}
	}

	switch status {
	case "SUCCEEDED":
		e.log.Info("job completed successfully", zap.String("job", jobName), zap.String("workload_id", w.GetId()), zap.String("url", url))
		return execiface.Result{Status: "SUCCEEDED", URL: url, Backend: "workspace"}
	case "FAILED":
		e.log.Warn("job completed with failure", zap.String("job", jobName), zap.String("workload_id", w.GetId()), zap.String("url", url))
		return execiface.Result{Status: "FAILED", URL: url, Backend: "workspace"}
	default:
		return execiface.Result{Status: "FAILED", URL: url, Backend: "workspace", Err: fmt.Errorf("unknown job outcome")}
	}
}

func (e *Executor) waitForCompletion(ctx context.Context, name string) string {
	selector := fields.OneTermEqualSelector("metadata.name", name).String()
	watcher, err := e.client.BatchV1().Jobs(e.namespace).Watch(ctx, metav1.ListOptions{FieldSelector: selector})
	if err != nil {
		e.log.Warn("job watch failed; polling", zap.String("job", name), zap.Error(err))
		return e.pollJob(ctx, name)
	}
	defer watcher.Stop()

	timeout := time.NewTimer(jobTimeout)
	defer timeout.Stop()

	for {
		select {
		case <-ctx.Done():
			return "FAILED"
		case <-timeout.C:
			e.log.Warn("job timeout", zap.String("job", name))
			return "FAILED"
		case ev, ok := <-watcher.ResultChan():
			if !ok {
				return e.pollJob(ctx, name)
			}
			j, ok := ev.Object.(*batchv1.Job)
			if !ok || j == nil {
				continue
			}
			if isComplete(j) {
				e.log.Info("job completed", zap.String("job", name))
				return "SUCCEEDED"
			}
			if isFailed(j) {
				e.log.Info("job failed", zap.String("job", name))
				return "FAILED"
			}
		}
	}
}

func (e *Executor) pollJob(ctx context.Context, name string) string {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	timeout := time.NewTimer(jobTimeout)
	defer timeout.Stop()

	jobs := e.client.BatchV1().Jobs(e.namespace)
	for {
		select {
		case <-ctx.Done():
			return "FAILED"
		case <-timeout.C:
			return "FAILED"
		case <-ticker.C:
			job, err := jobs.Get(ctx, name, metav1.GetOptions{})
			if err != nil {
				continue
			}
			if isComplete(job) {
				return "SUCCEEDED"
			}
			if isFailed(job) {
				return "FAILED"
			}
		}
	}
}

func (e *Executor) gpuRequests(w *aegis.Workload) corev1.ResourceList {
	if e.dryRun || w == nil {
		return nil
	}
	if h := w.GetHints(); h != nil {
		resName := h.GetResourceName()
		if resName == "" {
			resName = defaultGPUResource(flavorOf(w))
		}
		if resName != "" {
			cnt := h.GetGpuCount()
			if cnt <= 0 {
				cnt = 1
			}
			quantity := resource.MustParse(strconv.Itoa(int(cnt)))
			return corev1.ResourceList{corev1.ResourceName(resName): quantity}
		}
	}

	resourceName := getenv("AEGIS_GPU_RESOURCE_NAME", "")
	if resourceName == "" {
		resourceName = defaultGPUResource(flavorOf(w))
	}
	if resourceName == "" {
		return nil
	}
	quantity := resource.MustParse("1")
	return corev1.ResourceList{corev1.ResourceName(resourceName): quantity}
}

func isComplete(job *batchv1.Job) bool {
	for _, cond := range job.Status.Conditions {
		if cond.Type == batchv1.JobComplete && cond.Status == corev1.ConditionTrue {
			return true
		}
	}
	return false
}

func isFailed(job *batchv1.Job) bool {
	for _, cond := range job.Status.Conditions {
		if cond.Type == batchv1.JobFailed && cond.Status == corev1.ConditionTrue {
			return true
		}
	}
	return false
}

func flavorOf(w *aegis.Workload) string {
	if w == nil {
		return ""
	}
	if ws := w.GetWorkspace(); ws != nil && ws.GetFlavor() != "" {
		return ws.GetFlavor()
	}
	if tr := w.GetTraining(); tr != nil && tr.GetFlavor() != "" {
		return tr.GetFlavor()
	}
	return ""
}

func defaultGPUResource(flavor string) string {
	if strings.HasPrefix(flavor, "mig-") {
		return "nvidia.com/" + strings.TrimPrefix(flavor, "mig-")
	}
	return "nvidia.com/gpu"
}

func toEnvVars(env map[string]string) []corev1.EnvVar {
	if len(env) == 0 {
		return nil
	}
	out := make([]corev1.EnvVar, 0, len(env))
	for k, v := range env {
		out = append(out, corev1.EnvVar{Name: k, Value: v})
	}
	return out
}

func sanitizeName(in string) string {
	if in == "" {
		return fmt.Sprintf("aegis-%d", time.Now().Unix())
	}
	cleaned := strings.ToLower(in)
	cleaned = strings.ReplaceAll(cleaned, "_", "-")
	cleaned = strings.ReplaceAll(cleaned, ".", "-")
	if len(cleaned) > 63 {
		cleaned = cleaned[:63]
	}
	return cleaned
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func boolPtr(v bool) *bool { return &v }

func int32ptr(v int32) *int32 { return &v }
