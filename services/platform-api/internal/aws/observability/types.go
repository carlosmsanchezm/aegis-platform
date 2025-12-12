package observability

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/autoscaling"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
)

// FetchInput captures the parameters required to gather AWS signals for a cluster.
type FetchInput struct {
	ClusterName string
	Lookback    time.Duration
	LogLimit    int32
}

// Signals aggregates the control plane, nodegroup, and load balancer signals for a cluster.
type Signals struct {
	ControlPlane  ControlPlaneSignals
	Nodegroups    []NodegroupSignals
	LoadBalancers []LoadBalancerSignals
	Warnings      []string
}

// ControlPlaneSignals includes CloudWatch log/metric excerpts for an EKS control plane.
type ControlPlaneSignals struct {
	Status         string
	StatusMessage  string
	LoggingEnabled bool
	LogGroup       string
	LatencyP99     MetricSample
	Error5XXRate   MetricSample
	LogSamples     []ControlPlaneLog
	Missing        []string
}

// ControlPlaneLog is a small sample from the control plane log stream.
type ControlPlaneLog struct {
	Timestamp string
	Stream    string
	Message   string
}

// MetricSample captures a single CloudWatch datapoint and whether any data was present.
type MetricSample struct {
	Name          string
	Value         float64
	Unit          string
	PeriodSeconds int32
	Statistic     string
	Found         bool
}

// NodegroupSignals summarizes a single managed node group's capacity and health.
type NodegroupSignals struct {
	Name             string
	Status           string
	Desired          int32
	Current          int32
	Ready            int32
	AutoScalingGroup string
	Issues           []string
	ScalingEvents    []ScalingEvent
	GPU              bool
	GPUFlavor        string
	InstanceType     string
}

// ScalingEvent captures a short record of recent autoscaling activity.
type ScalingEvent struct {
	Description string
	Status      string
	Time        string
}

// TargetGroupHealth summarizes a target group's healthy/unhealthy target counts.
type TargetGroupHealth struct {
	ARN       string
	Name      string
	Healthy   int32
	Unhealthy int32
}

// AccessLogLocation describes load balancer access log configuration when enabled.
type AccessLogLocation struct {
	Enabled bool
	Bucket  string
	Prefix  string
}

// LoadBalancerSignals captures LB health and key metrics for ALB/NLB.
type LoadBalancerSignals struct {
	ARN              string
	Name             string
	Type             string
	Scheme           string
	State            string
	RequestCount     float64
	Target4XX        float64
	Target5XX        float64
	LatencyP99       MetricSample
	HealthyTargets   int32
	UnhealthyTargets int32
	TargetGroups     []TargetGroupHealth
	AccessLogs       AccessLogLocation
}

// CloudWatchAPI is the subset of CloudWatch operations used by the fetcher.
type CloudWatchAPI interface {
	GetMetricStatistics(ctx context.Context, params *cloudwatch.GetMetricStatisticsInput, optFns ...func(*cloudwatch.Options)) (*cloudwatch.GetMetricStatisticsOutput, error)
	ListMetrics(ctx context.Context, params *cloudwatch.ListMetricsInput, optFns ...func(*cloudwatch.Options)) (*cloudwatch.ListMetricsOutput, error)
}

// CloudWatchLogsAPI is the subset of CloudWatch Logs operations used by the fetcher.
type CloudWatchLogsAPI interface {
	DescribeLogGroups(ctx context.Context, params *cloudwatchlogs.DescribeLogGroupsInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.DescribeLogGroupsOutput, error)
	FilterLogEvents(ctx context.Context, params *cloudwatchlogs.FilterLogEventsInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.FilterLogEventsOutput, error)
}

// EKSAPI is the subset of EKS operations used by the fetcher.
type EKSAPI interface {
	DescribeCluster(ctx context.Context, params *eks.DescribeClusterInput, optFns ...func(*eks.Options)) (*eks.DescribeClusterOutput, error)
	ListNodegroups(ctx context.Context, params *eks.ListNodegroupsInput, optFns ...func(*eks.Options)) (*eks.ListNodegroupsOutput, error)
	DescribeNodegroup(ctx context.Context, params *eks.DescribeNodegroupInput, optFns ...func(*eks.Options)) (*eks.DescribeNodegroupOutput, error)
}

// AutoScalingAPI is the subset of Auto Scaling operations used by the fetcher.
type AutoScalingAPI interface {
	DescribeAutoScalingGroups(ctx context.Context, params *autoscaling.DescribeAutoScalingGroupsInput, optFns ...func(*autoscaling.Options)) (*autoscaling.DescribeAutoScalingGroupsOutput, error)
	DescribeScalingActivities(ctx context.Context, params *autoscaling.DescribeScalingActivitiesInput, optFns ...func(*autoscaling.Options)) (*autoscaling.DescribeScalingActivitiesOutput, error)
}

// ELBv2API is the subset of ELBv2 operations used by the fetcher.
type ELBv2API interface {
	DescribeLoadBalancers(ctx context.Context, params *elasticloadbalancingv2.DescribeLoadBalancersInput, optFns ...func(*elasticloadbalancingv2.Options)) (*elasticloadbalancingv2.DescribeLoadBalancersOutput, error)
	DescribeTags(ctx context.Context, params *elasticloadbalancingv2.DescribeTagsInput, optFns ...func(*elasticloadbalancingv2.Options)) (*elasticloadbalancingv2.DescribeTagsOutput, error)
	DescribeTargetGroups(ctx context.Context, params *elasticloadbalancingv2.DescribeTargetGroupsInput, optFns ...func(*elasticloadbalancingv2.Options)) (*elasticloadbalancingv2.DescribeTargetGroupsOutput, error)
	DescribeTargetHealth(ctx context.Context, params *elasticloadbalancingv2.DescribeTargetHealthInput, optFns ...func(*elasticloadbalancingv2.Options)) (*elasticloadbalancingv2.DescribeTargetHealthOutput, error)
	DescribeLoadBalancerAttributes(ctx context.Context, params *elasticloadbalancingv2.DescribeLoadBalancerAttributesInput, optFns ...func(*elasticloadbalancingv2.Options)) (*elasticloadbalancingv2.DescribeLoadBalancerAttributesOutput, error)
}
