// aegis-eks-token is a credential exec plugin that generates EKS bearer tokens
// using the AWS SDK. It replaces the shell-based aws eks get-token approach.
//
// Expected environment variables:
//   - AEGIS_EKS_CLUSTER_NAME (required)
//   - AEGIS_EKS_REGION (required)
//   - AEGIS_EKS_ROLE_ARN (optional — for cross-account access)
//   - AEGIS_EKS_EXTERNAL_ID (optional — paired with AEGIS_EKS_ROLE_ARN)
//
// Output: JSON ExecCredential per client.authentication.k8s.io/v1beta1
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/yourorg/aegis/services/platform-api/internal/kubeclients/ekstoken"
)

type execCredential struct {
	APIVersion string           `json:"apiVersion"`
	Kind       string           `json:"kind"`
	Status     execCredStatus   `json:"status"`
}

type execCredStatus struct {
	Token               string `json:"token"`
	ExpirationTimestamp string `json:"expirationTimestamp"`
}

func main() {
	clusterName := strings.TrimSpace(os.Getenv("AEGIS_EKS_CLUSTER_NAME"))
	region := strings.TrimSpace(os.Getenv("AEGIS_EKS_REGION"))
	roleARN := strings.TrimSpace(os.Getenv("AEGIS_EKS_ROLE_ARN"))
	externalID := strings.TrimSpace(os.Getenv("AEGIS_EKS_EXTERNAL_ID"))

	if clusterName == "" {
		fatal("AEGIS_EKS_CLUSTER_NAME is required")
	}
	if region == "" {
		fatal("AEGIS_EKS_REGION is required")
	}

	provider := &ekstoken.Provider{
		Region:      region,
		ClusterName: clusterName,
		RoleARN:     roleARN,
		ExternalID:  externalID,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	token, expiry, err := provider.Token(ctx)
	if err != nil {
		fatal("failed to generate token: %v", err)
	}

	cred := execCredential{
		APIVersion: "client.authentication.k8s.io/v1beta1",
		Kind:       "ExecCredential",
		Status: execCredStatus{
			Token:               token,
			ExpirationTimestamp: expiry.UTC().Format(time.RFC3339),
		},
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(cred); err != nil {
		fatal("failed to encode credential: %v", err)
	}
}

func fatal(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "aegis-eks-token: "+format+"\n", args...)
	os.Exit(1)
}
