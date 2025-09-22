package kube

import "k8s.io/client-go/kubernetes"

// Clientset is a simple alias to keep downstream files tidy.
type Clientset = kubernetes.Clientset

// ClientsetWrapper lets us swap in fakes under test if needed.
type ClientsetWrapper struct {
    Inner *kubernetes.Clientset
}
