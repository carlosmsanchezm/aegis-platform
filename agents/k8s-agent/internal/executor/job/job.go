package job

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"go.uber.org/zap"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"

	"github.com/yourorg/aegis/agents/k8s-agent/internal/kube"
	aegis "github.com/yourorg/aegis/proto/aegis/v1"
)

const (
	defaultNamespace = "default"
)

type Executor struct {
	log       *zap.Logger
	namespace string
	dryRun    bool
	cleanup   bool
	cs        *kubernetes.Clientset
}

func New(log *zap.Logger) (*Executor, error) {
	if log == nil {
		log = zap.NewNop()
	}

	cs, _, err := kube.New()
	if err != nil {
		return nil, err
	}

	ns := getenv("AEGIS_NAMESPACE", defaultNamespace)
	return &Executor{
		log:       log,
		namespace: ns,
		dryRun:    os.Getenv("AEGIS_DRY_RUN") == "1",
		cleanup:   os.Getenv("AEGIS_CLEANUP_JOBS") == "1",
		cs:        cs,
	}, nil
}

func (e *Executor) RunWorkspace(ctx context.Context, wl *aegis.Workload) error {
	if wl == nil {
		return fmt.Errorf("workload required")
	}
	ws := wl.GetWorkspace()
	if ws == nil {
		return fmt.Errorf("workspace spec required for job executor")
	}

	job := buildJob(e.namespace, wl, ws)
	jobsClient := e.cs.BatchV1().Jobs(e.namespace)

	if e.dryRun {
		e.log.Info("dry-run: skipping job submission", zap.String("workload_id", wl.GetId()), zap.String("job", job.GetName()))
		return nil
	}

	created, err := jobsClient.Create(ctx, job, metav1.CreateOptions{})
	if err != nil {
		if !kerrors.IsAlreadyExists(err) {
			return err
		}
		created, err = jobsClient.Get(ctx, job.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}
	}

	e.log.Info("job submitted", zap.String("workload_id", wl.GetId()), zap.String("job", created.GetName()))

	if err := e.waitForCompletion(ctx, created.GetName()); err != nil {
		return err
	}

	if e.cleanup {
		propagation := metav1.DeletePropagationBackground
		err := jobsClient.Delete(ctx, created.GetName(), metav1.DeleteOptions{PropagationPolicy: &propagation})
		if err != nil && !kerrors.IsNotFound(err) {
			e.log.Warn("failed to cleanup job", zap.String("job", created.GetName()), zap.Error(err))
		}
	}

	return nil
}

func (e *Executor) waitForCompletion(ctx context.Context, name string) error {
	fieldSelector := fields.OneTermEqualSelector("metadata.name", name).String()
	watcher, err := e.cs.BatchV1().Jobs(e.namespace).Watch(ctx, metav1.ListOptions{FieldSelector: fieldSelector})
	if err != nil {
		return err
	}
	defer watcher.Stop()

	ch := watcher.ResultChan()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case evt, ok := <-ch:
			if !ok {
				return fmt.Errorf("job watch closed")
			}
			job, ok := evt.Object.(*batchv1.Job)
			if !ok {
				continue
			}
			if job.Status.Succeeded > 0 {
				e.log.Info("job succeeded", zap.String("job", job.GetName()))
				return nil
			}
			if job.Status.Failed > 0 {
				e.log.Warn("job failed", zap.String("job", job.GetName()), zap.Int32("failed", job.Status.Failed))
				return fmt.Errorf("job %s failed", job.GetName())
			}
		}
	}
}

func buildJob(namespace string, wl *aegis.Workload, ws *aegis.WorkspaceSpec) *batchv1.Job {
	jobName := fmt.Sprintf("aegis-%s", sanitizeName(wl.GetId()))
	container := corev1.Container{
		Name:  "workspace",
		Image: ws.GetImage(),
		Env:   envMapToVars(ws.GetEnv()),
		Resources: corev1.ResourceRequirements{
			Requests: gpuResourceRequests(ws.GetFlavor()),
			Limits:   gpuResourceRequests(ws.GetFlavor()),
		},
	}

	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: namespace,
			Labels: map[string]string{
				"aegis.workload/id": wl.GetId(),
			},
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					RestartPolicy: corev1.RestartPolicyNever,
					Containers:    []corev1.Container{container},
				},
			},
		},
	}
}

func envMapToVars(env map[string]string) []corev1.EnvVar {
	if len(env) == 0 {
		return nil
	}
	out := make([]corev1.EnvVar, 0, len(env))
	for k, v := range env {
		out = append(out, corev1.EnvVar{Name: k, Value: v})
	}
	return out
}

func gpuResourceRequests(flavor string) corev1.ResourceList {
	if flavor == "" {
		return nil
	}

	req := corev1.ResourceList{}
	quantity := resource.MustParse("1")
	if strings.Contains(flavor, "mig") {
		resName := corev1.ResourceName("nvidia.com/mig-" + strings.ReplaceAll(flavor, "-", "."))
		req[resName] = quantity
	} else {
		req[corev1.ResourceName("nvidia.com/gpu")] = quantity
	}
	return req
}

func sanitizeName(in string) string {
	if in == "" {
		return fmt.Sprintf("job-%d", time.Now().Unix())
	}
	cleaned := strings.ToLower(in)
	cleaned = strings.ReplaceAll(cleaned, "_", "-")
	cleaned = strings.ReplaceAll(cleaned, ".", "-")
	return cleaned
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
