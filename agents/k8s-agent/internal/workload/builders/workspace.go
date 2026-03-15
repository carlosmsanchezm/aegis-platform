package builders

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	workspacecfg "github.com/yourorg/aegis/pkg/workspace"
)

// GPUHints mirrors the control-plane hint payload.
type GPUHints struct {
	ResourceName    string
	GPUCount        int32
	CpuCoresRequest *string
	MemoryRequest   *string
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
	Interactive             bool
	InteractivePorts        []int32
	SSHSecretName           string
	SSHBootstrapImage       string
	VSCodeREHInitImage      string // Iron Bank init container that injects VS Code REH binary
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

	resources := corev1.ResourceRequirements{
		Limits:   corev1.ResourceList{},
		Requests: corev1.ResourceList{},
	}
	resourcesRequested := false

	if !opts.DryRun {
		if res := gpuResourceRequests(opts); len(res) > 0 {
			for name, qty := range res {
				resources.Requests[name] = qty
				resources.Limits[name] = qty
			}
			resourcesRequested = true
		}
		if opts.Hints != nil && opts.Hints.CpuCoresRequest != nil && *opts.Hints.CpuCoresRequest != "" {
			cpuQty := resource.MustParse(*opts.Hints.CpuCoresRequest)
			resources.Requests[corev1.ResourceCPU] = cpuQty
			resources.Limits[corev1.ResourceCPU] = cpuQty
			resourcesRequested = true
		}
		if opts.Hints != nil && opts.Hints.MemoryRequest != nil && *opts.Hints.MemoryRequest != "" {
			memQty := resource.MustParse(*opts.Hints.MemoryRequest)
			resources.Requests[corev1.ResourceMemory] = memQty
			resources.Limits[corev1.ResourceMemory] = memQty
			resourcesRequested = true
		}
	}

	container := corev1.Container{
		Name:                     "workspace",
		Image:                    opts.Image,
		Env:                      env,
		TerminationMessagePolicy: corev1.TerminationMessageFallbackToLogsOnError,
	}

	if resourcesRequested {
		container.Resources = resources
	}

	// Only override command if explicitly provided in workload spec
	// Otherwise, use the image's default ENTRYPOINT and CMD
	if len(opts.Command) > 0 {
		container.Command = opts.Command
	}

	if opts.Interactive {
		if len(opts.InteractivePorts) == 0 {
			opts.InteractivePorts = workspacecfg.EnsureDefaultPorts(nil)
		}
		applyInteractiveContainerSettings(&container, opts)
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

	podSpec := corev1.PodSpec{
		RestartPolicy: corev1.RestartPolicyNever,
		Containers:    []corev1.Container{container},
	}

	// Add node selector and tolerations for GPU workloads
	if opts.Hints != nil && opts.Hints.GPUCount > 0 {
		applyGPUScheduling(&podSpec, opts.Hints)
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
				Spec:       podSpec,
			},
		},
	}

	if opts.Interactive {
		configureInteractivePod(job, opts)
	}

	return job
}

func applyInteractiveContainerSettings(container *corev1.Container, opts WorkspaceOptions) {
	if container == nil {
		return
	}
	seen := map[int32]bool{}
	for _, port := range opts.InteractivePorts {
		if port <= 0 || seen[port] {
			continue
		}
		container.Ports = append(container.Ports, corev1.ContainerPort{
			Name:          fmt.Sprintf("tcp-%d", port),
			ContainerPort: port,
			Protocol:      corev1.ProtocolTCP,
		})
		seen[port] = true
	}
}

func configureInteractivePod(job *batchv1.Job, opts WorkspaceOptions) {
	if job == nil {
		return
	}
	podSpec := &job.Spec.Template.Spec
	if len(podSpec.Containers) == 0 {
		return
	}
	main := &podSpec.Containers[0]

	const (
		sshConfigVolume   = "aegis-ssh-config"
		sshMaterialVolume = "aegis-ssh-material"
	)

	podSpec.Volumes = append(podSpec.Volumes, corev1.Volume{
		Name: sshConfigVolume,
		VolumeSource: corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{},
		},
	})
	if opts.SSHSecretName != "" {
		podSpec.Volumes = append(podSpec.Volumes, corev1.Volume{
			Name: sshMaterialVolume,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{SecretName: opts.SSHSecretName},
			},
		})
	}

	bootstrapImage := opts.SSHBootstrapImage
	if bootstrapImage == "" {
		bootstrapImage = os.Getenv("AEGIS_SSH_BOOTSTRAP_IMAGE")
	}
	if bootstrapImage == "" {
		bootstrapImage = "busybox:1.36" // Use busybox by default instead of the workspace image
	}

	script := `set -eu
ROOT="/config-root"
CFG_DIR="${ROOT}/config"
KEY_DIR="${ROOT}/keys"
mkdir -p "${CFG_DIR}" "${KEY_DIR}"

CFG_FILE="${CFG_DIR}/sshd_config"
cat <<'EOF' >"${CFG_FILE}"
# Managed by Aegis workspace bootstrap
Port 2222
ListenAddress 0.0.0.0

# Authentication policy
PasswordAuthentication yes
KbdInteractiveAuthentication no
ChallengeResponseAuthentication no
UsePAM yes
PermitRootLogin no
AuthorizedKeysFile /aegis-ssh/authorized_keys

# Required for VS Code Remote SSH
AllowTcpForwarding yes
AllowStreamLocalForwarding yes
GatewayPorts no

# Harden other vectors
X11Forwarding no
ClientAliveInterval 120
ClientAliveCountMax 2

# Keep parity with image defaults, but allow drop-ins when present
Include /etc/ssh/sshd_config.d/*.conf

Match all
  AllowTcpForwarding yes
  AllowStreamLocalForwarding yes
  GatewayPorts no
  PermitOpen any
  PermitListen any
EOF

chmod 600 "${CFG_FILE}" || true
chown root:root "${CFG_FILE}" 2>/dev/null || true

AUTH_KEYS="${KEY_DIR}/authorized_keys"
rm -f "${AUTH_KEYS}"
touch "${AUTH_KEYS}"
chmod 600 "${AUTH_KEYS}"
if [ -f /material/authorized_keys ]; then
  cp /material/authorized_keys "${AUTH_KEYS}"
  chmod 600 "${AUTH_KEYS}"
fi

if [ -f /material/trusted_user_ca_keys ]; then
  CA_FILE="${KEY_DIR}/trusted_user_ca_keys"
  cp /material/trusted_user_ca_keys "${CA_FILE}"
  chmod 644 "${CA_FILE}"
  printf '\nTrustedUserCAKeys /aegis-ssh/trusted_user_ca_keys\n' >>"${CFG_FILE}"
fi
`

	init := corev1.Container{
		Name:    "aegis-ssh-bootstrap",
		Image:   bootstrapImage,
		Command: []string{"/bin/sh", "-c", script},
		VolumeMounts: []corev1.VolumeMount{
			{Name: sshConfigVolume, MountPath: "/config-root"},
		},
	}
	if opts.SSHSecretName != "" {
		init.VolumeMounts = append(init.VolumeMounts, corev1.VolumeMount{Name: sshMaterialVolume, MountPath: "/material", ReadOnly: true})
	}
	podSpec.InitContainers = append(podSpec.InitContainers, init)

	main.VolumeMounts = append(main.VolumeMounts,
		corev1.VolumeMount{Name: sshConfigVolume, MountPath: "/aegis-ssh", SubPath: "keys"},
		// Mount config where sshd.pam reads it so forwarding settings apply.
		corev1.VolumeMount{Name: sshConfigVolume, MountPath: "/config/sshd", SubPath: "config"},
	)
	main.Env = append(main.Env, corev1.EnvVar{Name: "AEGIS_SSH_CONFIG_DIR", Value: "/aegis-ssh"})

	// VS Code REH init container: injects pre-built VS Code server binary from
	// a separate Iron Bank image into a shared emptyDir volume at /reh.
	// The workspace entrypoint expects to find the server at /reh/bin/current/.
	if opts.VSCodeREHInitImage != "" {
		const rehVolumeName = "aegis-vscode-reh"

		podSpec.Volumes = append(podSpec.Volumes, corev1.Volume{
			Name: rehVolumeName,
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		})

		rehInit := corev1.Container{
			Name:    "aegis-vscode-reh-init",
			Image:   opts.VSCodeREHInitImage,
			Command: []string{"cp", "-a", "/reh/.", "/shared/reh/"},
			VolumeMounts: []corev1.VolumeMount{
				{Name: rehVolumeName, MountPath: "/shared/reh"},
			},
		}
		podSpec.InitContainers = append(podSpec.InitContainers, rehInit)

		main.VolumeMounts = append(main.VolumeMounts,
			corev1.VolumeMount{Name: rehVolumeName, MountPath: "/reh"},
		)
	}
}

func gpuResourceRequests(opts WorkspaceOptions) corev1.ResourceList {
	if opts.DryRun {
		return nil
	}

	// First check hints for explicit GPU count
	count := int32(0)
	if opts.Hints != nil {
		count = opts.Hints.GPUCount
	}

	// If count is explicitly 0, don't request GPUs
	if count == 0 {
		return nil
	}

	// Determine GPU resource name
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

// applyGPUScheduling adds node selector and tolerations for GPU workloads.
// This ensures pods are scheduled onto nodes with the appropriate GPU hardware
// and can tolerate the GPU taints.
func applyGPUScheduling(pod *corev1.PodSpec, hints *GPUHints) {
	if pod == nil || hints == nil || hints.GPUCount <= 0 {
		return
	}

	resourceName := hints.ResourceName
	if resourceName == "" {
		resourceName = "nvidia.com/gpu"
	}

	// Determine the GPU flavor label based on resource name
	flavorLabel := "nvidia-tesla-t4" // default for standard nvidia.com/gpu
	taintKey := "nvidia.com/gpu"

	if strings.Contains(resourceName, "mig-1g.10gb") {
		flavorLabel = "nvidia-a10g-mig"
		taintKey = "nvidia.com/mig-1g.10gb"
	}

	// Add node selector to target GPU nodes
	if pod.NodeSelector == nil {
		pod.NodeSelector = map[string]string{}
	}
	pod.NodeSelector["aegis.io/gpu-flavor"] = flavorLabel

	// Add toleration for GPU taint
	hasToleration := false
	for _, tol := range pod.Tolerations {
		if tol.Key == taintKey {
			hasToleration = true
			break
		}
	}
	if !hasToleration {
		pod.Tolerations = append(pod.Tolerations, corev1.Toleration{
			Key:      taintKey,
			Operator: corev1.TolerationOpExists,
			Effect:   corev1.TaintEffectNoSchedule,
		})
	}
}
