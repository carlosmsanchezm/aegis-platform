package server

import (
	"context"
	"strings"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	aegis "github.com/yourorg/aegis/proto/aegis/v1"
	"github.com/yourorg/aegis/services/platform-api/internal/aws/observability"
	"github.com/yourorg/aegis/services/platform-api/internal/store"
)

const (
	defaultAwsLookbackMinutes = 15
	defaultLogSampleLimit     = 50
)

// GetAwsClusterSignals returns CloudWatch- and ELB-backed signals for an EKS cluster scoped to a project.
func (s *Server) GetAwsClusterSignals(ctx context.Context, req *aegis.GetAwsSignalsRequest) (*aegis.GetAwsSignalsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	projectID := strings.TrimSpace(req.GetProjectId())
	clusterID := strings.TrimSpace(req.GetClusterId())
	region := strings.TrimSpace(req.GetRegion())
	if projectID == "" {
		return nil, status.Error(codes.InvalidArgument, "project_id is required")
	}
	if clusterID == "" {
		return nil, status.Error(codes.InvalidArgument, "cluster_id is required")
	}
	info := s.findClusterInfo(clusterID)
	if info == nil {
		return nil, status.Errorf(codes.NotFound, "cluster %q not found", clusterID)
	}
	if pid := strings.TrimSpace(info.Labels["aegis.yourorg.dev/projectId"]); pid != "" && !strings.EqualFold(pid, projectID) {
		return nil, status.Errorf(codes.PermissionDenied, "cluster %q belongs to project %q", clusterID, pid)
	}
	if region == "" && strings.TrimSpace(info.Region) != "" {
		region = strings.TrimSpace(info.Region)
	}
	if region == "" {
		return nil, status.Error(codes.InvalidArgument, "region is required")
	}
	if err := s.authorize(ctx, projectID, "", "getAwsClusterSignals"); err != nil {
		return nil, err
	}
	project := s.store.GetProject(projectID)
	if project == nil {
		return nil, status.Errorf(codes.NotFound, "project %q not found", projectID)
	}
	populateProjectAwsFromAnnotations(project)
	creds := resolveProjectCredentials(project)
	if err := creds.validate(); err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, "project %q missing AWS credentials: %v", projectID, err)
	}
	if s.awsFetcherFactory == nil {
		return nil, status.Error(codes.FailedPrecondition, "aws fetcher not configured")
	}

	lookback := time.Duration(req.GetLookbackMinutes()) * time.Minute
	if lookback <= 0 {
		lookback = defaultAwsLookbackMinutes * time.Minute
	}

	fetcher, err := s.awsFetcherFactory(ctx, region, creds)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "configure AWS client: %v", err)
	}
	signals, err := fetcher.Fetch(ctx, observability.FetchInput{
		ClusterName: clusterID,
		Lookback:    lookback,
		LogLimit:    defaultLogSampleLimit,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "fetch AWS signals: %v", err)
	}
	return convertAwsSignals(clusterID, region, signals), nil
}

func (s *Server) findClusterInfo(clusterID string) *store.ClusterInfo {
	if s == nil || s.store == nil {
		return nil
	}
	for _, info := range s.store.ListClusterInfos() {
		if info != nil && strings.EqualFold(info.ID, clusterID) {
			return info
		}
	}
	return nil
}

func convertAwsSignals(clusterID, region string, sigs *observability.Signals) *aegis.GetAwsSignalsResponse {
	resp := &aegis.GetAwsSignalsResponse{
		ClusterId: clusterID,
		Region:    region,
	}
	if sigs == nil {
		return resp
	}
	resp.ControlPlane = convertControlPlane(sigs.ControlPlane)
	for _, ng := range sigs.Nodegroups {
		resp.Nodegroups = append(resp.Nodegroups, convertNodegroup(ng))
	}
	for _, lb := range sigs.LoadBalancers {
		resp.LoadBalancers = append(resp.LoadBalancers, convertLoadBalancer(lb))
	}
	if len(sigs.Warnings) > 0 {
		resp.Warnings = sigs.Warnings
	}
	return resp
}

func convertControlPlane(cp observability.ControlPlaneSignals) *aegis.ControlPlaneSignals {
	out := &aegis.ControlPlaneSignals{
		Status:              cp.Status,
		StatusMessage:       cp.StatusMessage,
		LoggingEnabled:      cp.LoggingEnabled,
		LogGroup:            cp.LogGroup,
		Missing:             cp.Missing,
		ApiServerLatencyP99: metricSampleToProto(cp.LatencyP99),
		ApiServer_5XxRate:   metricSampleToProto(cp.Error5XXRate),
	}
	for _, entry := range cp.LogSamples {
		out.LogSamples = append(out.LogSamples, &aegis.ControlPlaneLogEntry{
			Timestamp: entry.Timestamp,
			Stream:    entry.Stream,
			Message:   entry.Message,
		})
	}
	return out
}

func convertNodegroup(ng observability.NodegroupSignals) *aegis.NodegroupSignal {
	out := &aegis.NodegroupSignal{
		Name:    ng.Name,
		Status:  ng.Status,
		Desired: ng.Desired,
		Current: ng.Current,
		Ready:   ng.Ready,
		AsgName: ng.AutoScalingGroup,
		Issues:  ng.Issues,
	}
	for _, evt := range ng.ScalingEvents {
		out.ScalingEvents = append(out.ScalingEvents, &aegis.ScalingEvent{
			Description: evt.Description,
			Status:      evt.Status,
			Time:        evt.Time,
		})
	}
	return out
}

func convertLoadBalancer(lb observability.LoadBalancerSignals) *aegis.LoadBalancerSignal {
	out := &aegis.LoadBalancerSignal{
		Arn:              lb.ARN,
		Name:             lb.Name,
		Type:             lb.Type,
		Scheme:           lb.Scheme,
		State:            lb.State,
		RequestCount:     lb.RequestCount,
		Target_4Xx:       lb.Target4XX,
		Target_5Xx:       lb.Target5XX,
		LatencyP99:       metricSampleToProto(lb.LatencyP99),
		HealthyTargets:   lb.HealthyTargets,
		UnhealthyTargets: lb.UnhealthyTargets,
	}
	for _, tg := range lb.TargetGroups {
		out.TargetGroups = append(out.TargetGroups, &aegis.TargetGroupHealth{
			Arn:       tg.ARN,
			Name:      tg.Name,
			Healthy:   tg.Healthy,
			Unhealthy: tg.Unhealthy,
		})
	}
	out.AccessLogs = &aegis.LoadBalancerAccessLog{
		Enabled: lb.AccessLogs.Enabled,
		Bucket:  lb.AccessLogs.Bucket,
		Prefix:  lb.AccessLogs.Prefix,
	}
	return out
}

func metricSampleToProto(sample observability.MetricSample) *aegis.CloudWatchMetricSample {
	return &aegis.CloudWatchMetricSample{
		Name:          sample.Name,
		Value:         sample.Value,
		Unit:          sample.Unit,
		PeriodSeconds: sample.PeriodSeconds,
		Statistic:     sample.Statistic,
		Found:         sample.Found,
	}
}
