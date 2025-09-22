package kube

import (
    "fmt"

    "k8s.io/client-go/kubernetes"
    "k8s.io/client-go/rest"
    "sigs.k8s.io/controller-runtime/pkg/client/config"
)

// New returns a client-go Clientset and the underlying REST config.
// Works in-cluster and out-of-cluster (uses controller-runtime detection).
func New() (*kubernetes.Clientset, *rest.Config, error) {
    cfg, err := config.GetConfig()
    if err != nil {
        return nil, nil, fmt.Errorf("kube config: %w", err)
    }
    cs, err := kubernetes.NewForConfig(cfg)
    if err != nil {
        return nil, nil, fmt.Errorf("kube clientset: %w", err)
    }
    return cs, cfg, nil
}
