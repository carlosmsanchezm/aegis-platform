# ÆGIS MVP Architecture

The MVP consists of a control-plane service and one or more Kubernetes agents. The control-plane hosts the gRPC/REST API, maintains the in-memory catalog of projects, budgets, flavors, and queues, and handles workload submission. Agents run inside Kubernetes clusters, register themselves with the control plane, and periodically heartbeat GPU availability.

For local development both components run on a single developer workstation with the agents connecting back to the control plane via insecure gRPC. Later iterations will introduce mTLS, OIDC authentication, and advanced placement logic that integrates with core Kueue and Kubeflow components.
