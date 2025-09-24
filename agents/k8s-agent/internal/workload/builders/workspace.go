package builders

import (
	"os"
	"sort"
	"strconv"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GPUHints mirrors the control-plane hint payload.
type GPUHints struct {
	ResourceName string
	GPUCount     int32
}

// WorkspaceOptions describes the inputs required to render a Kubernetes Job for workspaces.
type WorkspaceOptions struct {
	Namespace               string
	JobName                 string
	WorkloadID              string
	Image                   string
	Command                 []string
	DefaultCommand          string
	Env                     map[string]string
	Flavor                  string
	Queue                   string
	Hints                   *GPUHints
	DryRun                  bool
	KueueEnabled            bool
	KueueQueue              string
	GPUResourceOverride     string
	ActiveDeadlineSeconds   *int64
	TTLSecondsAfterFinished *int32
}

// BuildWorkspaceJob renders a batch/v1 Job matching the legacy executor behaviour.
func BuildWorkspaceJob(opts WorkspaceOptions) *batchv1.Job {
	env := mapToEnvVars(opts.Env)
	env = append(env,
		corev1.EnvVar{Name: "AEGIS_WORKLOAD_ID", Value: opts.WorkloadID},
		corev1.EnvVar{Name: "POD_NAME", ValueFrom: fieldRef("metadata.name")},
		corev1.EnvVar{Name: "POD_NAMESPACE", ValueFrom: fieldRef("metadata.namespace")},
		corev1.EnvVar{Name: "NODE_NAME", ValueFrom: fieldRef("spec.nodeName")},
	)

	container := corev1.Container{
		Name:                     "workspace",
		Image:                    opts.Image,
		Env:                      env,
		TerminationMessagePolicy: corev1.TerminationMessageFallbackToLogsOnError,
	}

	if res := gpuResourceRequests(opts); len(res) > 0 {
		container.Resources = corev1.ResourceRequirements{Requests: res, Limits: res}
	}

	if len(opts.Command) > 0 {
		container.Command = opts.Command
	} else {
		container.Command = []string{"/bin/sh", "-c"}
		container.Args = []string{opts.DefaultCommand}
	}

	jobLabels := map[string]string{
		"aegis.workload/id": opts.WorkloadID,
		"aegis.job/name":    opts.JobName,
	}
	podLabels := map[string]string{
		"aegis.workload/id": opts.WorkloadID,
	}
	suspend := false
	kueueAllowed := opts.KueueEnabled && os.Getenv("AEGIS_DISABLE_KUEUE") != "1"
	if kueueAllowed {
		if queue := resolveQueue(opts); queue != "" {
			jobLabels["kueue.x-k8s.io/queue-name"] = queue
			podLabels["kueue.x-k8s.io/queue-name"] = queue
			suspend = true
		}
	}

	jobAnnotations := map[string]string{}
	if !kueueAllowed {
		jobAnnotations["kueue.x-k8s.io/skip-admission"] = "true"
	}

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:        opts.JobName,
			Namespace:   opts.Namespace,
			Labels:      jobLabels,
			Annotations: jobAnnotations,
		},
		Spec: batchv1.JobSpec{
			Suspend:                 boolPtr(suspend),
			BackoffLimit:            int32Ptr(0),
			ActiveDeadlineSeconds:   opts.ActiveDeadlineSeconds,
			TTLSecondsAfterFinished: opts.TTLSecondsAfterFinished,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: podLabels},
				Spec: corev1.PodSpec{
					RestartPolicy: corev1.RestartPolicyNever,
					Containers:    []corev1.Container{container},
				},
			},
		},
	}

	return job
}

func gpuResourceRequests(opts WorkspaceOptions) corev1.ResourceList {
	if opts.DryRun {
		return nil
	}

	resName := opts.GPUResourceOverride
	if opts.Hints != nil && opts.Hints.ResourceName != "" {
		resName = opts.Hints.ResourceName
	}
	if resName == "" {
		resName = AutoGPUResource(opts.Flavor)
	}
	if resName == "" {
		return nil
	}

	count := int32(1)
	if opts.Hints != nil && opts.Hints.GPUCount > 0 {
		count = opts.Hints.GPUCount
	}
	if count <= 0 {
		count = 1
	}

	quantity := resource.MustParse(strconv.Itoa(int(count)))
	return corev1.ResourceList{corev1.ResourceName(resName): quantity}
}

func resolveQueue(opts WorkspaceOptions) string {
	if opts.KueueQueue != "" {
		return opts.KueueQueue
	}
	return opts.Queue
}

func mapToEnvVars(env map[string]string) []corev1.EnvVar {
	if len(env) == 0 {
		return nil
	}
	keys := make([]string, 0, len(env))
	for k := range env {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	vars := make([]corev1.EnvVar, 0, len(env))
	for _, k := range keys {
		vars = append(vars, corev1.EnvVar{Name: k, Value: env[k]})
	}
	return vars
}

func fieldRef(path string) *corev1.EnvVarSource {
	return &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: path}}
}

func boolPtr(v bool) *bool { return &v }

func int32Ptr(v int32) *int32 { return &v }
