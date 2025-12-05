package controllers

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
	coordinationv1 "k8s.io/api/coordination/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	infraapi "github.com/yourorg/aegis/services/platform-api/api/v1alpha1"
	"github.com/yourorg/aegis/services/platform-api/internal/provisioning"
	"github.com/yourorg/aegis/services/platform-api/internal/store"
)

const (
	projectInfraFinalizer = "infra.aegis.yourorg.dev/finalizer"
	conditionReady        = "Ready"
)

// ProjectInfraReconciler handles ProjectInfra lifecycle orchestration.
type ProjectInfraReconciler struct {
	client.Client
	Scheme                    *runtime.Scheme
	Log                       *zap.Logger
	Provisioner               provisioning.AWSProvisioner
	Store                     store.Store
	KubeconfigSecretName      string
	KubeconfigSecretNamespace string
	LocalLocks                map[string]*sync.Mutex
	LocalLocksMu              sync.Mutex
	HolderIdentity            string
}

// SetupWithManager registers the reconciler with the manager.
func (r *ProjectInfraReconciler) SetupWithManager(mgr ctrl.Manager, opts controller.Options) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&infraapi.ProjectInfra{}).
		WithOptions(opts).
		Complete(r)
}

// Reconcile implements the controller-runtime reconciliation contract.
func (r *ProjectInfraReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.logger().With(zap.String("projectinfra", req.NamespacedName.String()))

	var infra infraapi.ProjectInfra
	if err := r.Get(ctx, req.NamespacedName, &infra); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	releaseLock, res, err := r.acquireStackLock(ctx, &infra)
	if err != nil {
		return res, err
	}
	// If we failed to acquire the lease, requeue instead of proceeding without coordination.
	if res.Requeue || res.RequeueAfter > 0 {
		if releaseLock != nil {
			releaseLock()
		}
		return res, nil
	}
	if releaseLock != nil {
		defer releaseLock()
	}

	if !infra.DeletionTimestamp.IsZero() {
		return r.reconcileDelete(ctx, log, &infra)
	}

	if err := r.ensureFinalizer(ctx, &infra); err != nil {
		return ctrl.Result{}, err
	}

	return r.reconcileNormal(ctx, log, &infra)
}

func (r *ProjectInfraReconciler) reconcileNormal(ctx context.Context, log *zap.Logger, infra *infraapi.ProjectInfra) (ctrl.Result, error) {
	// If a prior run failed, avoid implicit retries. The UI will delete/recreate
	// the ProjectInfra to retry, so keep the object idle in error state.
	if strings.EqualFold(infra.Status.Phase, "Error") {
		log.Info("skipping reconcile; infrastructure is in error state")
		return ctrl.Result{}, nil
	}

	if infra.Spec.Aws != nil {
		if strings.EqualFold(string(infra.Spec.Aws.Mode), string(infraapi.AWSProvisionModeImport)) {
			return r.handleAWSImport(ctx, log, infra)
		}
	}
	if infra.Spec.ImportKubeconfigSecretRef != nil {
		return r.handleImport(ctx, log, infra)
	}
	if infra.Spec.Aws != nil {
		return r.handleAWSProvision(ctx, log, infra)
	}
	reason := "InvalidSpec"
	err := fmt.Errorf("spec.aws or spec.importKubeconfigSecretRef required")
	cond := newCondition(metav1.ConditionFalse, reason, err.Error())
	_ = r.setStatus(ctx, infra, "Error", &cond, nil, infra.Status.CostHintUSDPerHour)
	return ctrl.Result{}, err
}

func (r *ProjectInfraReconciler) handleAWSImport(ctx context.Context, log *zap.Logger, infra *infraapi.ProjectInfra) (ctrl.Result, error) {
	aws := infra.Spec.Aws
	if aws == nil {
		err := fmt.Errorf("aws spec required for import mode")
		cond := newCondition(metav1.ConditionFalse, "InvalidSpec", err.Error())
		_ = r.setStatus(ctx, infra, "Error", &cond, infra.Status.Outputs, infra.Status.CostHintUSDPerHour)
		return ctrl.Result{}, err
	}
	if len(aws.Imports) == 0 {
		err := fmt.Errorf("spec.aws.imports must include at least one entry")
		cond := newCondition(metav1.ConditionFalse, "InvalidSpec", err.Error())
		_ = r.setStatus(ctx, infra, "Error", &cond, infra.Status.Outputs, infra.Status.CostHintUSDPerHour)
		return ctrl.Result{}, err
	}

	kubeconfigData := map[string][]byte{}
	outputs := make([]infraapi.ClusterOutput, 0, len(aws.Imports))

	for _, imp := range aws.Imports {
		clusterID := strings.TrimSpace(imp.ClusterID)
		if clusterID == "" {
			err := fmt.Errorf("spec.aws.imports requires clusterId")
			cond := newCondition(metav1.ConditionFalse, "InvalidSpec", err.Error())
			_ = r.setStatus(ctx, infra, "Error", &cond, infra.Status.Outputs, infra.Status.CostHintUSDPerHour)
			return ctrl.Result{}, err
		}

		ref := imp.KubeconfigSecretRef
		secretName := strings.TrimSpace(ref.Name)
		secretKey := strings.TrimSpace(ref.Key)
		if secretName == "" || secretKey == "" {
			err := fmt.Errorf("import %s missing kubeconfig secret reference", clusterID)
			cond := newCondition(metav1.ConditionFalse, "InvalidSpec", err.Error())
			_ = r.setStatus(ctx, infra, "Error", &cond, infra.Status.Outputs, infra.Status.CostHintUSDPerHour)
			return ctrl.Result{}, err
		}
		secretNamespace := strings.TrimSpace(ref.Namespace)
		if secretNamespace == "" {
			secretNamespace = infra.Namespace
		}

		secret := &corev1.Secret{}
		if err := r.Get(ctx, types.NamespacedName{Name: secretName, Namespace: secretNamespace}, secret); err != nil {
			cond := newCondition(metav1.ConditionFalse, "ImportFailed", err.Error())
			_ = r.setStatus(ctx, infra, "Error", &cond, infra.Status.Outputs, infra.Status.CostHintUSDPerHour)
			return ctrl.Result{}, err
		}
		if secret.Data == nil {
			err := fmt.Errorf("secret %s/%s missing data", secretNamespace, secretName)
			cond := newCondition(metav1.ConditionFalse, "ImportFailed", err.Error())
			_ = r.setStatus(ctx, infra, "Error", &cond, infra.Status.Outputs, infra.Status.CostHintUSDPerHour)
			return ctrl.Result{}, err
		}
		payload, ok := secret.Data[secretKey]
		if !ok {
			err := fmt.Errorf("secret %s/%s missing key %s", secretNamespace, secretName, secretKey)
			cond := newCondition(metav1.ConditionFalse, "ImportFailed", err.Error())
			_ = r.setStatus(ctx, infra, "Error", &cond, infra.Status.Outputs, infra.Status.CostHintUSDPerHour)
			return ctrl.Result{}, err
		}

		aggregatedKey := fmt.Sprintf("%s.kubeconfig", clusterID)
		kubeconfigData[aggregatedKey] = payload
		outputs = append(outputs, infraapi.ClusterOutput{
			ClusterID:           clusterID,
			Name:                clusterID,
			Region:              infra.Spec.Region,
			KubeconfigSecretKey: aggregatedKey,
		})
	}

	if err := r.writeKubeconfigs(ctx, kubeconfigData); err != nil {
		cond := newCondition(metav1.ConditionFalse, "SecretSyncFailed", err.Error())
		_ = r.setStatus(ctx, infra, "Error", &cond, infra.Status.Outputs, infra.Status.CostHintUSDPerHour)
		return ctrl.Result{}, err
	}

	for _, output := range outputs {
		if err := r.ensureAegisCluster(ctx, infra, output); err != nil {
			cond := newCondition(metav1.ConditionFalse, "ClusterSpecSyncFailed", err.Error())
			_ = r.setStatus(ctx, infra, "Error", &cond, infra.Status.Outputs, infra.Status.CostHintUSDPerHour)
			return ctrl.Result{}, err
		}
	}

	condReady := newCondition(metav1.ConditionTrue, "Imported", "aws kubeconfig imports synchronized")
	if err := r.setStatus(ctx, infra, "Ready", &condReady, outputs, infra.Status.CostHintUSDPerHour); err != nil {
		return ctrl.Result{}, err
	}
	log.Info("project infrastructure imports synchronized", zap.Int("clusters", len(outputs)))
	return ctrl.Result{}, nil
}

func (r *ProjectInfraReconciler) handleImport(ctx context.Context, log *zap.Logger, infra *infraapi.ProjectInfra) (ctrl.Result, error) {
	data, outputs, err := r.extractImportData(ctx, infra)
	if err != nil {
		cond := newCondition(metav1.ConditionFalse, "ImportFailed", err.Error())
		_ = r.setStatus(ctx, infra, "Error", &cond, infra.Status.Outputs, infra.Status.CostHintUSDPerHour)
		return ctrl.Result{}, err
	}

	if err := r.writeKubeconfigs(ctx, data); err != nil {
		cond := newCondition(metav1.ConditionFalse, "SecretSyncFailed", err.Error())
		_ = r.setStatus(ctx, infra, "Error", &cond, infra.Status.Outputs, infra.Status.CostHintUSDPerHour)
		return ctrl.Result{}, err
	}

	for _, output := range outputs {
		if err := r.ensureAegisCluster(ctx, infra, output); err != nil {
			cond := newCondition(metav1.ConditionFalse, "ClusterSpecSyncFailed", err.Error())
			_ = r.setStatus(ctx, infra, "Error", &cond, infra.Status.Outputs, infra.Status.CostHintUSDPerHour)
			return ctrl.Result{}, err
		}
	}

	cond := newCondition(metav1.ConditionTrue, "Imported", "kubeconfig secret imported")
	if err := r.setStatus(ctx, infra, "Ready", &cond, outputs, 0); err != nil {
		return ctrl.Result{}, err
	}
	log.Info("project infrastructure imported", zap.Int("clusters", len(outputs)))
	return ctrl.Result{}, nil
}

func (r *ProjectInfraReconciler) handleAWSProvision(ctx context.Context, log *zap.Logger, infra *infraapi.ProjectInfra) (ctrl.Result, error) {
	if r.Provisioner == nil {
		err := fmt.Errorf("aws provisioner not configured")
		cond := newCondition(metav1.ConditionFalse, "ProvisionerUnavailable", err.Error())
		_ = r.setStatus(ctx, infra, "Error", &cond, infra.Status.Outputs, infra.Status.CostHintUSDPerHour)
		return ctrl.Result{}, err
	}

	condProvisioning := newCondition(metav1.ConditionFalse, "Provisioning", "provisioning in progress")
	_ = r.setStatus(ctx, infra, "Provisioning", &condProvisioning, infra.Status.Outputs, infra.Status.CostHintUSDPerHour)

	result, err := r.Provisioner.Provision(ctx, infra, infra.Spec.Aws)
	if err != nil {
		cond := newCondition(metav1.ConditionFalse, "ProvisionFailed", err.Error())
		_ = r.setStatus(ctx, infra, "Error", &cond, infra.Status.Outputs, infra.Status.CostHintUSDPerHour)
		// Do not requeue automatically on failure; require an explicit relaunch.
		return ctrl.Result{}, nil
	}

	kubeconfigData := map[string][]byte{}
	for clusterID, kubeconfig := range result.Kubeconfigs {
		key := fmt.Sprintf("%s.kubeconfig", clusterID)
		kubeconfigData[key] = kubeconfig
	}
	if len(kubeconfigData) == 0 {
		err := fmt.Errorf("provisioner returned no kubeconfigs")
		cond := newCondition(metav1.ConditionFalse, "ProvisionFailed", err.Error())
		_ = r.setStatus(ctx, infra, "Error", &cond, infra.Status.Outputs, infra.Status.CostHintUSDPerHour)
		return ctrl.Result{}, err
	}

	if err := r.writeKubeconfigs(ctx, kubeconfigData); err != nil {
		cond := newCondition(metav1.ConditionFalse, "SecretSyncFailed", err.Error())
		_ = r.setStatus(ctx, infra, "Error", &cond, infra.Status.Outputs, infra.Status.CostHintUSDPerHour)
		return ctrl.Result{}, err
	}

	outputs := result.Outputs
	if len(outputs) == 0 {
		outputs = make([]infraapi.ClusterOutput, 0, len(kubeconfigData))
		for key := range kubeconfigData {
			clusterID := strings.TrimSuffix(key, ".kubeconfig")
			outputs = append(outputs, infraapi.ClusterOutput{
				ClusterID:           clusterID,
				Name:                clusterID,
				Region:              infra.Spec.Region,
				KubeconfigSecretKey: key,
			})
		}
	}
	for i := range outputs {
		if outputs[i].KubeconfigSecretKey == "" {
			outputs[i].KubeconfigSecretKey = fmt.Sprintf("%s.kubeconfig", outputs[i].ClusterID)
		}
		if outputs[i].Region == "" {
			outputs[i].Region = infra.Spec.Region
		}
		if err := r.ensureAegisCluster(ctx, infra, outputs[i]); err != nil {
			cond := newCondition(metav1.ConditionFalse, "ClusterSpecSyncFailed", err.Error())
			_ = r.setStatus(ctx, infra, "Error", &cond, infra.Status.Outputs, infra.Status.CostHintUSDPerHour)
			return ctrl.Result{}, err
		}
	}

	condReady := newCondition(metav1.ConditionTrue, "Provisioned", "aws infrastructure provisioned")
	if err := r.setStatus(ctx, infra, "Ready", &condReady, outputs, result.CostHintUSDPerHour); err != nil {
		return ctrl.Result{}, err
	}
	log.Info("project infrastructure provisioned", zap.Int("clusters", len(outputs)))
	return ctrl.Result{}, nil
}

func (r *ProjectInfraReconciler) reconcileDelete(ctx context.Context, log *zap.Logger, infra *infraapi.ProjectInfra) (ctrl.Result, error) {
	clusterIDs := collectClusterIDs(infra)
	log.Info("reconciling project infrastructure deletion",
		zap.String("name", infra.Name),
		zap.String("namespace", infra.Namespace),
		zap.String("project", infra.Spec.ProjectID),
		zap.String("region", infra.Spec.Region),
		zap.Strings("cluster_ids", clusterIDs),
	)
	if err := r.removeKubeconfigs(ctx, clusterIDs); err != nil {
		log.Error("failed to remove kubeconfigs", zap.Error(err))
		return ctrl.Result{}, err
	}
	if err := r.deleteAegisClusters(ctx, clusterIDs); err != nil {
		log.Error("failed to delete aegis clusters", zap.Error(err))
		return ctrl.Result{}, err
	}
	if infra.Spec.Aws != nil && r.Provisioner != nil {
		log.Info("triggering aws destroy via pulumi",
			zap.String("project", infra.Spec.ProjectID),
			zap.String("region", infra.Spec.Region),
			zap.Strings("cluster_ids", clusterIDs),
		)
		if err := r.Provisioner.Destroy(ctx, infra, infra.Spec.Aws); err != nil {
			log.Error("aws destroy failed", zap.Error(err))
			return ctrl.Result{}, err
		}
		log.Info("aws infrastructure destroy triggered", zap.Int("clusters", len(clusterIDs)))
	} else if infra.Spec.Aws != nil && r.Provisioner == nil {
		log.Warn("aws destroy skipped; provisioner not configured")
	}

	patched := infra.DeepCopy()
	controllerutil.RemoveFinalizer(patched, projectInfraFinalizer)
	if err := r.Patch(ctx, patched, client.MergeFrom(infra)); err != nil && !apierrors.IsNotFound(err) {
		return ctrl.Result{}, err
	}
	log.Info("project infrastructure deletion finalized",
		zap.String("name", infra.Name),
		zap.String("namespace", infra.Namespace),
	)
	return ctrl.Result{}, nil
}

func (r *ProjectInfraReconciler) ensureFinalizer(ctx context.Context, infra *infraapi.ProjectInfra) error {
	if controllerutil.ContainsFinalizer(infra, projectInfraFinalizer) {
		return nil
	}
	patched := infra.DeepCopy()
	controllerutil.AddFinalizer(patched, projectInfraFinalizer)
	if err := r.Patch(ctx, patched, client.MergeFrom(infra)); err != nil {
		return err
	}
	infra.ObjectMeta = patched.ObjectMeta
	return nil
}

func (r *ProjectInfraReconciler) extractImportData(ctx context.Context, infra *infraapi.ProjectInfra) (map[string][]byte, []infraapi.ClusterOutput, error) {
	ref := infra.Spec.ImportKubeconfigSecretRef
	if ref == nil {
		return nil, nil, fmt.Errorf("import secret reference required")
	}
	namespace := ref.Namespace
	if namespace == "" {
		namespace = infra.Namespace
	}
	secret := &corev1.Secret{}
	if err := r.Get(ctx, types.NamespacedName{Name: ref.Name, Namespace: namespace}, secret); err != nil {
		return nil, nil, err
	}

	resultData := map[string][]byte{}
	outputs := []infraapi.ClusterOutput{}
	for key, value := range secret.Data {
		clusterID := ""
		switch {
		case strings.HasSuffix(key, ".kubeconfig"):
			clusterID = strings.TrimSuffix(key, ".kubeconfig")
		case strings.HasSuffix(key, ".yaml"):
			clusterID = strings.TrimSuffix(key, ".yaml")
		case strings.HasSuffix(key, ".yml"):
			clusterID = strings.TrimSuffix(key, ".yml")
		default:
			continue
		}
		clusterID = strings.TrimSpace(clusterID)
		if clusterID == "" {
			continue
		}
		secretKey := fmt.Sprintf("%s.kubeconfig", clusterID)
		resultData[secretKey] = value
		outputs = append(outputs, infraapi.ClusterOutput{
			ClusterID:           clusterID,
			Name:                clusterID,
			Region:              infra.Spec.Region,
			KubeconfigSecretKey: secretKey,
		})
	}
	if len(resultData) == 0 {
		return nil, nil, fmt.Errorf("no kubeconfig data found in secret %s/%s", secret.Namespace, secret.Name)
	}
	return resultData, outputs, nil
}

func (r *ProjectInfraReconciler) writeKubeconfigs(ctx context.Context, entries map[string][]byte) error {
	if len(entries) == 0 {
		return fmt.Errorf("no kubeconfig entries provided")
	}

	secret := &corev1.Secret{}
	key := types.NamespacedName{Name: r.kubeconfigSecretName(), Namespace: r.kubeconfigSecretNamespace()}
	err := r.Get(ctx, key, secret)
	if apierrors.IsNotFound(err) {
		secret = &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: key.Name, Namespace: key.Namespace},
			Type:       corev1.SecretTypeOpaque,
			Data:       map[string][]byte{},
		}
		for k, v := range entries {
			secret.Data[k] = v
		}
		return r.Create(ctx, secret)
	} else if err != nil {
		return err
	}

	if secret.Data == nil {
		secret.Data = map[string][]byte{}
	}
	mutated := false
	for k, v := range entries {
		if existing, ok := secret.Data[k]; !ok || !bytes.Equal(existing, v) {
			secret.Data[k] = v
			mutated = true
		}
	}
	if !mutated {
		return nil
	}
	return r.Update(ctx, secret)
}

func (r *ProjectInfraReconciler) removeKubeconfigs(ctx context.Context, clusterIDs []string) error {
	if len(clusterIDs) == 0 {
		return nil
	}
	secret := &corev1.Secret{}
	key := types.NamespacedName{Name: r.kubeconfigSecretName(), Namespace: r.kubeconfigSecretNamespace()}
	if err := r.Get(ctx, key, secret); err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		return err
	}
	if secret.Data == nil {
		return nil
	}
	mutated := false
	for _, id := range clusterIDs {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		key := fmt.Sprintf("%s.kubeconfig", id)
		if _, ok := secret.Data[key]; ok {
			delete(secret.Data, key)
			mutated = true
		}
	}
	if !mutated {
		return nil
	}
	if len(secret.Data) == 0 {
		return r.Delete(ctx, secret)
	}
	return r.Update(ctx, secret)
}

func (r *ProjectInfraReconciler) ensureAegisCluster(ctx context.Context, infra *infraapi.ProjectInfra, output infraapi.ClusterOutput) error {
	clusterID := strings.TrimSpace(output.ClusterID)
	if clusterID == "" {
		return fmt.Errorf("cluster ID required")
	}

	spec := infraapi.AegisClusterSpec{
		ClusterID: clusterID,
		ProjectID: infra.Spec.ProjectID,
		Provider:  infra.Spec.Provider,
		Region:    output.Region,
	}

	key := types.NamespacedName{Name: clusterID, Namespace: r.kubeconfigSecretNamespace()}
	current := &infraapi.AegisCluster{}
	err := r.Get(ctx, key, current)
	if apierrors.IsNotFound(err) {
		current = &infraapi.AegisCluster{
			TypeMeta: metav1.TypeMeta{APIVersion: infraapi.GroupVersion.String(), Kind: "AegisCluster"},
			ObjectMeta: metav1.ObjectMeta{
				Name:      key.Name,
				Namespace: key.Namespace,
			},
			Spec: spec,
		}
		if key.Namespace == infra.Namespace {
			if err := controllerutil.SetControllerReference(infra, current, r.Scheme); err != nil {
				return err
			}
		}
		return r.Create(ctx, current)
	} else if err != nil {
		return err
	}

	if current.Spec.ClusterID == spec.ClusterID && current.Spec.ProjectID == spec.ProjectID && current.Spec.Provider == spec.Provider && current.Spec.Region == spec.Region {
		return nil
	}

	patched := current.DeepCopy()
	patched.Spec = spec
	if key.Namespace == infra.Namespace {
		if err := controllerutil.SetControllerReference(infra, patched, r.Scheme); err != nil {
			return err
		}
	}
	return r.Patch(ctx, patched, client.MergeFrom(current))
}

func (r *ProjectInfraReconciler) deleteAegisClusters(ctx context.Context, clusterIDs []string) error {
	for _, id := range clusterIDs {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		cluster := &infraapi.AegisCluster{}
		key := types.NamespacedName{Name: id, Namespace: r.kubeconfigSecretNamespace()}
		if err := r.Get(ctx, key, cluster); err != nil {
			if apierrors.IsNotFound(err) {
				continue
			}
			return err
		}
		if err := r.Delete(ctx, cluster); err != nil && !apierrors.IsNotFound(err) {
			return err
		}
	}
	return nil
}

func collectClusterIDs(infra *infraapi.ProjectInfra) []string {
	ids := map[string]struct{}{}
	for _, out := range infra.Status.Outputs {
		id := strings.TrimSpace(out.ClusterID)
		if id != "" {
			ids[id] = struct{}{}
		}
	}
	if len(ids) == 0 && infra.Spec.Aws != nil {
		primary := strings.TrimSpace(infra.Spec.Aws.ClusterName)
		if primary != "" {
			ids[primary] = struct{}{}
		}
		for _, extra := range infra.Spec.Aws.AdditionalClusters {
			if name := strings.TrimSpace(extra.ClusterName); name != "" {
				ids[name] = struct{}{}
			}
		}
	}
	if infra.Spec.Aws != nil {
		for _, imp := range infra.Spec.Aws.Imports {
			if id := strings.TrimSpace(imp.ClusterID); id != "" {
				ids[id] = struct{}{}
			}
		}
	}
	result := make([]string, 0, len(ids))
	for id := range ids {
		result = append(result, id)
	}
	return result
}

func (r *ProjectInfraReconciler) setStatus(ctx context.Context, infra *infraapi.ProjectInfra, phase string, cond *metav1.Condition, outputs []infraapi.ClusterOutput, costHint float64) error {
	patched := infra.DeepCopy()
	patched.Status.Phase = phase
	if outputs != nil {
		patched.Status.Outputs = outputs
	}
	patched.Status.CostHintUSDPerHour = costHint
	now := metav1.NewTime(time.Now())
	patched.Status.LastSyncTime = &now
	if cond != nil && cond.Type != "" {
		mergeCondition(&patched.Status.Conditions, *cond)
	}
	if err := r.Status().Patch(ctx, patched, client.MergeFrom(infra)); err != nil {
		return err
	}
	infra.Status = patched.Status
	return nil
}

func newCondition(status metav1.ConditionStatus, reason, message string) metav1.Condition {
	return metav1.Condition{
		Type:               conditionReady,
		Status:             status,
		Reason:             reason,
		Message:            message,
		LastTransitionTime: metav1.NewTime(time.Now()),
	}
}

func mergeCondition(conds *[]metav1.Condition, cond metav1.Condition) {
	if conds == nil {
		return
	}
	for i, existing := range *conds {
		if existing.Type == cond.Type {
			if existing.Status == cond.Status {
				cond.LastTransitionTime = existing.LastTransitionTime
			}
			(*conds)[i] = cond
			return
		}
	}
	*conds = append(*conds, cond)
}

func (r *ProjectInfraReconciler) kubeconfigSecretName() string {
	if strings.TrimSpace(r.KubeconfigSecretName) != "" {
		return r.KubeconfigSecretName
	}
	return "aegis-kubeconfigs"
}

func (r *ProjectInfraReconciler) kubeconfigSecretNamespace() string {
	if strings.TrimSpace(r.KubeconfigSecretNamespace) != "" {
		return r.KubeconfigSecretNamespace
	}
	return "aegis-system"
}

func (r *ProjectInfraReconciler) logger() *zap.Logger {
	if r.Log != nil {
		return r.Log
	}
	return zap.NewNop()
}

// ----------------------------------------------------------------------------- //
// Stack-level locking to avoid concurrent Pulumi runs for the same stack.

func (r *ProjectInfraReconciler) acquireStackLock(ctx context.Context, infra *infraapi.ProjectInfra) (func(), ctrl.Result, error) {
	stackKey := stackLockKey(infra)
	unlockLocal := r.lockLocal(stackKey)
	holder := r.holderIdentity()

	leaseName := leaseNameForInfra(infra)
	leaseNS := infra.Namespace

	lease := &coordinationv1.Lease{}
	err := r.Get(ctx, types.NamespacedName{Name: leaseName, Namespace: leaseNS}, lease)
	now := metav1.NowMicro()
	leaseDuration := int32(15 * 60) // 15 minutes

	// Keepalive/renewal loop to hold the lease while reconcile is running.
	stop := make(chan struct{})
	go func() {
		t := time.NewTicker(30 * time.Second)
		defer t.Stop()
		for {
			select {
			case <-t.C:
				_ = r.renewLease(context.Background(), leaseName, leaseNS, holder, leaseDuration)
			case <-stop:
				return
			case <-ctx.Done():
				return
			}
		}
	}()

	release := func() {
		close(stop)
		// Best-effort delete only if we hold the lease.
		r.releaseLease(ctx, leaseName, leaseNS, holder)
		unlockLocal()
	}

	if apierrors.IsNotFound(err) {
		newLease := &coordinationv1.Lease{
			ObjectMeta: metav1.ObjectMeta{
				Name:      leaseName,
				Namespace: leaseNS,
			},
			Spec: coordinationv1.LeaseSpec{
				HolderIdentity:       ptr.To(holder),
				LeaseDurationSeconds: ptr.To(leaseDuration),
				AcquireTime:          &now,
				RenewTime:            &now,
			},
		}
		if createErr := r.Create(ctx, newLease); createErr != nil {
			unlockLocal()
			close(stop)
			return nil, ctrl.Result{RequeueAfter: 5 * time.Second}, client.IgnoreNotFound(createErr)
		}
		return release, ctrl.Result{}, nil
	} else if err != nil {
		unlockLocal()
		close(stop)
		return nil, ctrl.Result{}, err
	}

	existingHolder := ""
	if lease.Spec.HolderIdentity != nil {
		existingHolder = strings.TrimSpace(*lease.Spec.HolderIdentity)
	}
	expired := false
	if lease.Spec.RenewTime != nil && lease.Spec.LeaseDurationSeconds != nil {
		expiry := lease.Spec.RenewTime.Time.Add(time.Duration(*lease.Spec.LeaseDurationSeconds) * time.Second)
		expired = time.Now().After(expiry)
	} else {
		expired = true
	}

	if existingHolder != "" && existingHolder != holder && !expired {
		unlockLocal()
		close(stop)
		return nil, ctrl.Result{RequeueAfter: 5 * time.Second}, nil
	}

	// Take over or renew the lease.
	lease.Spec.HolderIdentity = ptr.To(holder)
	lease.Spec.RenewTime = &now
	lease.Spec.LeaseDurationSeconds = ptr.To(leaseDuration)
	if err := r.Update(ctx, lease); err != nil {
		unlockLocal()
		if apierrors.IsConflict(err) {
			return nil, ctrl.Result{RequeueAfter: 5 * time.Second}, nil
		}
		return nil, ctrl.Result{}, err
	}

	return release, ctrl.Result{}, nil
}

func (r *ProjectInfraReconciler) releaseLease(ctx context.Context, name, namespace, holder string) {
	lease := &coordinationv1.Lease{}
	key := types.NamespacedName{Name: name, Namespace: namespace}
	if err := r.Get(ctx, key, lease); err != nil {
		return
	}
	if lease.Spec.HolderIdentity != nil && strings.TrimSpace(*lease.Spec.HolderIdentity) != holder {
		return
	}
	_ = r.Delete(ctx, lease)
}

// renewLease refreshes the lease if we hold it, or if it is expired/unheld. Best-effort; ignores not found.
func (r *ProjectInfraReconciler) renewLease(ctx context.Context, name, namespace, holder string, leaseDuration int32) error {
	if ctx == nil {
		ctx = context.Background()
	}
	lease := &coordinationv1.Lease{}
	key := types.NamespacedName{Name: name, Namespace: namespace}
	if err := r.Get(ctx, key, lease); err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		return err
	}

	existingHolder := ""
	if lease.Spec.HolderIdentity != nil {
		existingHolder = strings.TrimSpace(*lease.Spec.HolderIdentity)
	}

	expired := true
	if lease.Spec.RenewTime != nil && lease.Spec.LeaseDurationSeconds != nil {
		expiry := lease.Spec.RenewTime.Time.Add(time.Duration(*lease.Spec.LeaseDurationSeconds) * time.Second)
		expired = time.Now().After(expiry)
	}

	// Only renew if we hold it or it is expired/unheld.
	if existingHolder != "" && existingHolder != holder && !expired {
		return nil
	}

	now := metav1.NowMicro()
	lease.Spec.HolderIdentity = ptr.To(holder)
	lease.Spec.LeaseDurationSeconds = ptr.To(leaseDuration)
	if lease.Spec.AcquireTime == nil {
		lease.Spec.AcquireTime = &now
	}
	lease.Spec.RenewTime = &now

	return r.Update(ctx, lease)
}

func (r *ProjectInfraReconciler) lockLocal(key string) func() {
	if r.LocalLocks == nil {
		r.LocalLocks = map[string]*sync.Mutex{}
	}
	r.LocalLocksMu.Lock()
	m, ok := r.LocalLocks[key]
	if !ok {
		m = &sync.Mutex{}
		r.LocalLocks[key] = m
	}
	r.LocalLocksMu.Unlock()
	m.Lock()
	return func() {
		m.Unlock()
	}
}

func stackLockKey(infra *infraapi.ProjectInfra) string {
	return fmt.Sprintf("%s-%s", strings.TrimSpace(infra.Spec.ProjectID), strings.TrimSpace(infra.Spec.Region))
}

func leaseNameForInfra(infra *infraapi.ProjectInfra) string {
	base := fmt.Sprintf("pi-%s-%s", strings.TrimSpace(infra.Spec.ProjectID), strings.TrimSpace(infra.Spec.Region))
	return sanitizeName(base, 63)
}

func sanitizeName(in string, maxLen int) string {
	out := make([]rune, 0, len(in))
	for _, r := range strings.ToLower(in) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			out = append(out, r)
		} else {
			out = append(out, '-')
		}
	}
	name := strings.Trim(outStr(out), "-")
	if len(name) > maxLen {
		name = name[:maxLen]
	}
	if name == "" {
		name = "pi-stack"
	}
	return name
}

func outStr(r []rune) string {
	return string(r)
}

func (r *ProjectInfraReconciler) holderIdentity() string {
	if s := strings.TrimSpace(r.HolderIdentity); s != "" {
		return s
	}
	if h, err := os.Hostname(); err == nil && strings.TrimSpace(h) != "" {
		return h
	}
	return "aegis-platform-api"
}
