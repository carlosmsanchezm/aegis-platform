package status

import (
	"strings"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// IsJobComplete reports whether the Kubernetes Job has completed successfully.
func IsJobComplete(job *batchv1.Job) bool {
	if job == nil {
		return false
	}
	for _, cond := range job.Status.Conditions {
		if cond.Type == batchv1.JobComplete && cond.Status == corev1.ConditionTrue {
			return true
		}
	}
	return false
}

// IsJobFailed reports whether the Kubernetes Job has failed.
func IsJobFailed(job *batchv1.Job) bool {
	if job == nil {
		return false
	}
	for _, cond := range job.Status.Conditions {
		if cond.Type == batchv1.JobFailed && cond.Status == corev1.ConditionTrue {
			return true
		}
	}
	return false
}

// HasCondition queries an unstructured CR for a condition.
func HasCondition(u *unstructured.Unstructured, condType string) bool {
	if u == nil {
		return false
	}
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

// HasAnySuccess reports whether the CR has transitioned to a successful state.
func HasAnySuccess(u *unstructured.Unstructured) bool {
	return HasCondition(u, "Complete") || HasCondition(u, "Succeeded")
}

// HasAnyFailure reports whether the CR indicates terminal failure.
func HasAnyFailure(u *unstructured.Unstructured) bool {
	return HasCondition(u, "Failed")
}
