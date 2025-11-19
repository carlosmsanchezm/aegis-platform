package server

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	aegis "github.com/yourorg/aegis/proto/aegis/v1"
	infraapi "github.com/yourorg/aegis/services/platform-api/api/v1alpha1"
)

const (
	defaultInfraNamespace = "aegis-system"
	maxObjectNameLength   = 63
)

func (s *Server) buildProjectInfra(req *aegis.CreateClusterRequest, tmpl *clusterProfileTemplate, project *aegis.Project, creds projectAWSCredentials) (*infraapi.ProjectInfra, error) {
	profileReq := req.GetProfile()
	if profileReq == nil {
		return nil, status.Error(codes.InvalidArgument, "cluster profile required")
	}
	awsSpec := tmpl.instantiate(req.GetClusterId(), profileReq.GetParameters())
	if awsSpec == nil {
		return nil, status.Error(codes.Internal, "profile template missing aws spec")
	}
	projectID := strings.TrimSpace(req.GetProjectId())
	if project != nil && strings.TrimSpace(project.GetId()) != "" {
		projectID = strings.TrimSpace(project.GetId())
	}
	region := strings.TrimSpace(req.GetRegion())
	infraName := buildInfraObjectName(projectID, req.GetClusterId())
	labels := map[string]string{
		"aegis.yourorg.dev/projectId": sanitizeKubeName(projectID),
		"aegis.yourorg.dev/profileId": strings.ToLower(strings.TrimSpace(tmpl.ID)),
	}
	annotations := map[string]string{
		"aegis.yourorg.dev/clusterId": strings.TrimSpace(req.GetClusterId()),
	}
	if version := strings.TrimSpace(profileReq.GetVersion()); version != "" {
		annotations["aegis.yourorg.dev/profileVersion"] = version
	} else {
		annotations["aegis.yourorg.dev/profileVersion"] = tmpl.Version
	}
	if params := profileReq.GetParameters(); len(params) > 0 {
		if payload, err := json.Marshal(params); err == nil {
			annotations["aegis.yourorg.dev/profileParameters"] = string(payload)
		}
	}
	if creds.AccountID != "" {
		annotations[annotationAWSAccountID] = creds.AccountID
	}
	awsSpec.ClusterName = sanitizeClusterName(req.GetClusterId())
	awsSpec.AccountID = creds.AccountID
	awsSpec.RoleARN = creds.RoleARN
	awsSpec.ExternalID = creds.ExternalID
	infra := &infraapi.ProjectInfra{
		TypeMeta: metav1.TypeMeta{
			APIVersion: infraapi.GroupVersion.String(),
			Kind:       "ProjectInfra",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        infraName,
			Namespace:   s.infraNamespace,
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: infraapi.ProjectInfraSpec{
			ProjectID: projectID,
			Provider:  canonicalProvider(req.GetProvider()),
			Region:    region,
			Aws:       awsSpec,
			Labels: map[string]string{
				"profileRef": fmt.Sprintf("%s@%s", tmpl.ID, annotations["aegis.yourorg.dev/profileVersion"]),
			},
		},
	}
	return infra, nil
}

func canonicalProvider(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "aws", "aws-eks", "amazon", "amazon-eks":
		return "aws"
	default:
		return strings.ToLower(strings.TrimSpace(raw))
	}
}

func (s *Server) fetchInfra(ctx context.Context, jobID string) (*infraapi.ProjectInfra, error) {
	if jobID == "" {
		return nil, status.Error(codes.InvalidArgument, "job id required")
	}
	var infra infraapi.ProjectInfra
	key := types.NamespacedName{Name: jobID, Namespace: s.infraNamespace}
	if err := s.infraClient.Get(ctx, key, &infra); err != nil {
		return nil, err
	}
	return &infra, nil
}

func jobFromInfra(infra *infraapi.ProjectInfra, defaultID string) *aegis.Job {
	if infra == nil {
		return &aegis.Job{Id: defaultID, Status: "PENDING", Progress: 0}
	}
	phase := strings.TrimSpace(infra.Status.Phase)
	var status string
	var progress int32
	var message string
	switch strings.ToLower(phase) {
	case "ready":
		status = "SUCCEEDED"
		progress = 100
	case "error":
		status = "FAILED"
		progress = 100
		message = lastConditionMessage(infra.Status.Conditions)
	case "provisioning":
		status = "RUNNING"
		progress = 55
	default:
		status = "PENDING"
		progress = 10
	}
	return &aegis.Job{
		Id:       infra.Name,
		Status:   status,
		Progress: progress,
		Error:    message,
	}
}

func lastConditionMessage(conds []metav1.Condition) string {
	if len(conds) == 0 {
		return ""
	}
	lc := conds[len(conds)-1]
	return strings.TrimSpace(lc.Message)
}

func sanitizeKubeName(input string) string {
	trimmed := strings.ToLower(strings.TrimSpace(input))
	if trimmed == "" {
		return "default"
	}
	var b strings.Builder
	prevDash := false
	for _, r := range trimmed {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			prevDash = false
			continue
		}
		if !prevDash {
			b.WriteRune('-')
			prevDash = true
		}
	}
	result := strings.Trim(b.String(), "-")
	if result == "" {
		return "name"
	}
	if len(result) > maxObjectNameLength {
		return result[:maxObjectNameLength]
	}
	return result
}

func sanitizeClusterName(name string) string {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		trimmed = "cluster"
	}
	return sanitizeKubeName(trimmed)
}

func buildInfraObjectName(projectID, clusterID string) string {
	base := fmt.Sprintf("infra-%s-%s", sanitizeKubeName(projectID), sanitizeKubeName(clusterID))
	if len(base) <= maxObjectNameLength {
		return base
	}
	hash := sha1.Sum([]byte(base))
	suffix := hex.EncodeToString(hash[:])[:8]
	trim := maxObjectNameLength - len(suffix) - 1
	if trim < 1 {
		trim = maxObjectNameLength - len(suffix)
	}
	return fmt.Sprintf("%s-%s", base[:trim], suffix)
}
