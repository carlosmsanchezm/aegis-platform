package kube

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	_ = ctrl.Log
	_ client.Client
)

func ensureImports(cfg *rest.Config, pod *corev1.Pod) metav1.ObjectMeta {
	if cfg != nil && pod != nil {
		_ = cfg.Timeout
		return pod.ObjectMeta
	}
	return metav1.ObjectMeta{}
}
