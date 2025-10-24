package controllers

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
}

// SetupWithManager registers the reconciler with the manager.
func (r *ProjectInfraReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&infraapi.ProjectInfra{}).
		Complete(r)
}

// Reconcile implements the controller-runtime reconciliation contract.
func (r *ProjectInfraReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.logger().With(zap.String("projectinfra", req.NamespacedName.String()))

	var infra infraapi.ProjectInfra
	if err := r.Get(ctx, req.NamespacedName, &infra); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
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
	if infra.Spec.ImportKubeconfigSecretRef != nil {
		return r.handleImport(ctx, log, infra)
	}
	if infra.Spec.Aws != nil {
		return r.handleAWS(ctx, log, infra)
	}
	reason := "InvalidSpec"
	err := fmt.Errorf("spec.aws or spec.importKubeconfigSecretRef required")
	cond := newCondition(metav1.ConditionFalse, reason, err.Error())
	_ = r.setStatus(ctx, infra, "Error", &cond, nil, infra.Status.CostHintUSDPerHour)
	return ctrl.Result{}, err
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

func (r *ProjectInfraReconciler) handleAWS(ctx context.Context, log *zap.Logger, infra *infraapi.ProjectInfra) (ctrl.Result, error) {
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
		return ctrl.Result{}, err
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
	if err := r.removeKubeconfigs(ctx, clusterIDs); err != nil {
		return ctrl.Result{}, err
	}
	if err := r.deleteAegisClusters(ctx, clusterIDs); err != nil {
		return ctrl.Result{}, err
	}
	if infra.Spec.Aws != nil && r.Provisioner != nil {
		if err := r.Provisioner.Destroy(ctx, infra, infra.Spec.Aws); err != nil {
			return ctrl.Result{}, err
		}
		log.Info("aws infrastructure destroy triggered", zap.Int("clusters", len(clusterIDs)))
	}

	patched := infra.DeepCopy()
	controllerutil.RemoveFinalizer(patched, projectInfraFinalizer)
	if err := r.Patch(ctx, patched, client.MergeFrom(infra)); err != nil && !apierrors.IsNotFound(err) {
		return ctrl.Result{}, err
	}
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
		if err := controllerutil.SetControllerReference(infra, current, r.Scheme); err != nil {
			return err
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
	if err := controllerutil.SetControllerReference(infra, patched, r.Scheme); err != nil {
		return err
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
