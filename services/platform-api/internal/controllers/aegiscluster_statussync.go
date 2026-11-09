package controllers

import (
	"context"
	"sort"
	"time"

	"go.uber.org/zap"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	infraapi "github.com/yourorg/aegis/services/platform-api/api/v1alpha1"
	"github.com/yourorg/aegis/services/platform-api/internal/store"
)

const (
	heartbeatGrace = 90 * time.Second
)

// AegisClusterStatusSync mirrors store state into the management-plane AegisCluster CRDs.
type AegisClusterStatusSync struct {
	client.Client
	Log       *zap.Logger
	Store     store.Store
	Namespace string
	Interval  time.Duration
}

// NeedLeaderElection indicates this runnable does not require leader election.
func (s *AegisClusterStatusSync) NeedLeaderElection() bool { return false }

// Start begins the periodic sync loop.
func (s *AegisClusterStatusSync) Start(ctx context.Context) error {
	interval := s.Interval
	if interval <= 0 {
		interval = 30 * time.Second
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	if err := s.sync(ctx); err != nil {
		s.logger().Warn("initial cluster sync failed", zap.Error(err))
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			if err := s.sync(ctx); err != nil {
				s.logger().Warn("cluster sync failed", zap.Error(err))
			}
		}
	}
}

func (s *AegisClusterStatusSync) sync(ctx context.Context) error {
	if s.Store == nil {
		return nil
	}

	namespace := s.Namespace
	if namespace == "" {
		namespace = "aegis-system"
	}

	infos := s.Store.ListClusterInfos()
	existing := map[string]struct{}{}
	for _, info := range infos {
		name := info.ID
		existing[name] = struct{}{}
		if err := s.upsertCluster(ctx, namespace, info); err != nil {
			return err
		}
	}

	return s.gcOrphaned(ctx, namespace, existing)
}

func (s *AegisClusterStatusSync) upsertCluster(ctx context.Context, namespace string, info *store.ClusterInfo) error {
	cluster := &infraapi.AegisCluster{}
	key := types.NamespacedName{Name: info.ID, Namespace: namespace}
	err := s.Client.Get(ctx, key, cluster)
	if err != nil && !apierrors.IsNotFound(err) {
		return err
	}

	desiredSpec := infraapi.AegisClusterSpec{
		ClusterID: info.ID,
		ProjectID: info.Labels["aegis.yourorg.dev/projectId"],
		Provider:  info.Provider,
		Region:    info.Region,
	}

	phase, cond := clusterPhase(info)
	desiredStatus := infraapi.AegisClusterStatus{
		Phase:         phase,
		Flavors:       flattenFlavorSet(info.AvailableFlavorSet),
		Capacity:      nil,
		Conditions:    []metav1.Condition{cond},
		LastHeartbeat: cond.LastTransitionTime.DeepCopy(),
	}

	if apierrors.IsNotFound(err) {
		newCluster := &infraapi.AegisCluster{
			TypeMeta: metav1.TypeMeta{
				APIVersion: infraapi.GroupVersion.String(),
				Kind:       "AegisCluster",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      key.Name,
				Namespace: key.Namespace,
			},
			Spec:   desiredSpec,
			Status: desiredStatus,
		}
		if err := s.Client.Create(ctx, newCluster); err != nil {
			return err
		}
		return s.Client.Status().Update(ctx, newCluster)
	}

	specPatched := cluster.DeepCopy()
	specPatched.Spec = desiredSpec
	if err := s.Client.Patch(ctx, specPatched, client.MergeFrom(cluster)); err != nil {
		return err
	}
	cluster = specPatched

	statusPatched := cluster.DeepCopy()
	statusPatched.Status = desiredStatus
	return s.Client.Status().Patch(ctx, statusPatched, client.MergeFrom(cluster))
}

func (s *AegisClusterStatusSync) gcOrphaned(ctx context.Context, namespace string, active map[string]struct{}) error {
	var list infraapi.AegisClusterList
	if err := s.Client.List(ctx, &list, client.InNamespace(namespace)); err != nil {
		return err
	}
	for i := range list.Items {
		name := list.Items[i].Name
		if _, ok := active[name]; ok {
			continue
		}
		if err := s.Client.Delete(ctx, &list.Items[i]); err != nil && !apierrors.IsNotFound(err) {
			return err
		}
	}
	return nil
}

func clusterPhase(info *store.ClusterInfo) (string, metav1.Condition) {
	if info == nil {
		return "Pending", metav1.Condition{Type: "Heartbeat", Status: metav1.ConditionUnknown, Reason: "Unknown", LastTransitionTime: metav1.NewTime(time.Now())}
	}
	age := time.Since(info.LastHeartbeat)
	cond := metav1.Condition{Type: "Heartbeat", LastTransitionTime: metav1.NewTime(info.LastHeartbeat)}
	if info.LastHeartbeat.IsZero() {
		cond.Status = metav1.ConditionFalse
		cond.Reason = "NeverSeen"
		cond.Message = "cluster has not heartbeated yet"
		return "Pending", cond
	}
	if age <= heartbeatGrace {
		cond.Status = metav1.ConditionTrue
		cond.Reason = "HeartbeatOK"
		cond.Message = "cluster heartbeat within grace interval"
		return "HeartbeatOK", cond
	}
	cond.Status = metav1.ConditionFalse
	cond.Reason = "HeartbeatStale"
	cond.Message = "cluster heartbeat stale"
	return "Degraded", cond
}

func flattenFlavorSet(set map[string]bool) []string {
	if len(set) == 0 {
		return nil
	}
	out := make([]string, 0, len(set))
	for k, ok := range set {
		if ok {
			out = append(out, k)
		}
	}
	sort.Strings(out)
	return out
}

func (s *AegisClusterStatusSync) logger() *zap.Logger {
	if s.Log != nil {
		return s.Log
	}
	return zap.NewNop()
}

var _ manager.LeaderElectionRunnable = (*AegisClusterStatusSync)(nil)
