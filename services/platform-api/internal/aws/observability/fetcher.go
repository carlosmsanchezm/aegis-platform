package observability

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/autoscaling"
	astypes "github.com/aws/aws-sdk-go-v2/service/autoscaling/types"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	cwtypes "github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	ekstypes "github.com/aws/aws-sdk-go-v2/service/eks/types"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	elbtypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
)

// Fetcher orchestrates AWS calls to collect signals for a cluster.
type Fetcher struct {
	cloudwatch  CloudWatchAPI
	logs        CloudWatchLogsAPI
	eks         EKSAPI
	autoscaling AutoScalingAPI
	elb         ELBv2API
}

// Fetch gathers control plane, nodegroup, and load balancer signals for the provided cluster.
func (f *Fetcher) Fetch(ctx context.Context, in FetchInput) (*Signals, error) {
	if f == nil {
		return nil, fmt.Errorf("fetcher is not initialized")
	}
	clusterName := strings.TrimSpace(in.ClusterName)
	if clusterName == "" {
		return nil, fmt.Errorf("cluster name is required")
	}

	lookback := in.Lookback
	if lookback <= 0 {
		lookback = 15 * time.Minute
	}
	if lookback > 24*time.Hour {
		lookback = 24 * time.Hour
	}
	logLimit := in.LogLimit
	if logLimit <= 0 {
		logLimit = 20
	}
	if logLimit > 200 {
		logLimit = 200
	}

	end := time.Now()
	start := end.Add(-lookback)

	signals := &Signals{}
	var warnings []string

	cp, cpWarnings := f.fetchControlPlane(ctx, clusterName, start, end, logLimit)
	signals.ControlPlane = cp
	warnings = append(warnings, cpWarnings...)

	nodegroups, ngWarnings := f.fetchNodegroups(ctx, clusterName, start)
	signals.Nodegroups = nodegroups
	warnings = append(warnings, ngWarnings...)

	loadBalancers, lbWarnings := f.fetchLoadBalancers(ctx, clusterName, start, end)
	signals.LoadBalancers = loadBalancers
	warnings = append(warnings, lbWarnings...)

	signals.Warnings = warnings
	return signals, nil
}

func (f *Fetcher) fetchControlPlane(ctx context.Context, clusterName string, start, end time.Time, logLimit int32) (ControlPlaneSignals, []string) {
	var warnings []string
	out := ControlPlaneSignals{
		Status:         "unknown",
		StatusMessage:  "",
		LoggingEnabled: false,
		LogGroup:       fmt.Sprintf("/aws/eks/%s/cluster", clusterName),
		Missing:        []string{},
	}

	if f.eks != nil {
		desc, err := f.eks.DescribeCluster(ctx, &eks.DescribeClusterInput{Name: aws.String(clusterName)})
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("describe cluster: %v", err))
		} else if desc.Cluster != nil {
			out.Status = string(desc.Cluster.Status)
			if desc.Cluster.Logging != nil {
				out.LoggingEnabled = isLoggingEnabled(desc.Cluster.Logging)
			}
		}
	} else {
		warnings = append(warnings, "eks client not configured")
	}

	if f.logs != nil {
		exists, err := f.logGroupExists(ctx, out.LogGroup)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("describe log group: %v", err))
		}
		if exists {
			out.LoggingEnabled = true
			logs, logErr := f.sampleLogs(ctx, out.LogGroup, start, end, logLimit)
			if logErr != nil {
				warnings = append(warnings, fmt.Sprintf("fetch control plane logs: %v", logErr))
			} else {
				out.LogSamples = logs
			}
		} else {
			out.Missing = append(out.Missing, "control plane logs")
		}
	}

	dims := []cwtypes.Dimension{{Name: aws.String("ClusterName"), Value: aws.String(clusterName)}}
	metrics := []struct {
		name      string
		stat      string
		extended  bool
		fieldName string
	}{
		{name: "APIServerLatencyP99", stat: "p99", extended: true, fieldName: "latency"},
		{name: "APIServer5XXErrorRate", stat: "Average", fieldName: "errors"},
	}

	for _, candidate := range metrics {
		sample, err := f.fetchMetric(ctx, candidate.name, "AWS/EKS", dims, start, end, 300, candidate.stat, candidate.extended)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("metric %s: %v", candidate.name, err))
		}
		switch candidate.fieldName {
		case "latency":
			out.LatencyP99 = sample
		case "errors":
			out.Error5XXRate = sample
		}
		if !sample.Found {
			out.Missing = append(out.Missing, candidate.name)
		}
	}

	return out, warnings
}

func (f *Fetcher) fetchNodegroups(ctx context.Context, clusterName string, start time.Time) ([]NodegroupSignals, []string) {
	var (
		signals  []NodegroupSignals
		warnings []string
	)
	if f.eks == nil {
		return signals, []string{"eks client not configured"}
	}
	var nodegroupNames []string
	token := aws.String("")
	for {
		resp, err := f.eks.ListNodegroups(ctx, &eks.ListNodegroupsInput{ClusterName: aws.String(clusterName), NextToken: token})
		if err != nil {
			return signals, []string{fmt.Sprintf("list nodegroups: %v", err)}
		}
		nodegroupNames = append(nodegroupNames, resp.Nodegroups...)
		if resp.NextToken == nil || *resp.NextToken == "" {
			break
		}
		token = resp.NextToken
	}

	for _, ng := range nodegroupNames {
		desc, err := f.eks.DescribeNodegroup(ctx, &eks.DescribeNodegroupInput{ClusterName: aws.String(clusterName), NodegroupName: aws.String(ng)})
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("describe nodegroup %s: %v", ng, err))
			continue
		}
		if desc.Nodegroup == nil {
			continue
		}
		signal := NodegroupSignals{
			Name:          aws.ToString(desc.Nodegroup.NodegroupName),
			Status:        string(desc.Nodegroup.Status),
			Current:       0,
			Ready:         0,
			Issues:        []string{},
			ScalingEvents: []ScalingEvent{},
		}
		if desc.Nodegroup.ScalingConfig != nil && desc.Nodegroup.ScalingConfig.DesiredSize != nil {
			signal.Desired = int32(*desc.Nodegroup.ScalingConfig.DesiredSize)
		}
		if desc.Nodegroup.Health != nil {
			for _, issue := range desc.Nodegroup.Health.Issues {
				signal.Issues = append(signal.Issues, fmt.Sprintf("%s: %s", string(issue.Code), aws.ToString(issue.Message)))
			}
		}

		asgNames := autoScalingGroupNames(desc.Nodegroup.Resources)
		asgMap, asgWarn := f.describeAutoScalingGroups(ctx, asgNames)
		warnings = append(warnings, asgWarn...)
		for _, name := range asgNames {
			group := asgMap[name]
			if group == nil {
				continue
			}
			signal.AutoScalingGroup = name
			current, ready := summarizeInstances(group.Instances)
			signal.Current += current
			signal.Ready += ready
			events, evErr := f.scalingActivities(ctx, name, start)
			if evErr != nil {
				warnings = append(warnings, fmt.Sprintf("scaling activities %s: %v", name, evErr))
			} else {
				signal.ScalingEvents = append(signal.ScalingEvents, events...)
			}
		}
		signals = append(signals, signal)
	}
	return signals, warnings
}

func (f *Fetcher) fetchLoadBalancers(ctx context.Context, clusterName string, start, end time.Time) ([]LoadBalancerSignals, []string) {
	var (
		out      []LoadBalancerSignals
		warnings []string
	)
	if f.elb == nil {
		return out, []string{"elb client not configured"}
	}
	lbs, lbWarnings := f.clusterLoadBalancers(ctx, clusterName)
	warnings = append(warnings, lbWarnings...)
	if len(lbs) == 0 {
		return out, warnings
	}

	for _, lb := range lbs {
		signal := LoadBalancerSignals{
			ARN:    aws.ToString(lb.LoadBalancerArn),
			Name:   aws.ToString(lb.LoadBalancerName),
			Type:   string(lb.Type),
			Scheme: string(lb.Scheme),
			State:  string(lb.State.Code),
		}
		lbNamespace := lbNamespace(lb.Type)
		lbDim := loadBalancerDimension(aws.ToString(lb.LoadBalancerArn))

		targetGroups, tgWarn := f.targetGroupsForLB(ctx, aws.ToString(lb.LoadBalancerArn))
		warnings = append(warnings, tgWarn...)
		totalHealthy := int32(0)
		totalUnhealthy := int32(0)
		for _, tg := range targetGroups {
			health, err := f.targetGroupHealth(ctx, aws.ToString(tg.TargetGroupArn))
			if err != nil {
				warnings = append(warnings, fmt.Sprintf("target health %s: %v", aws.ToString(tg.TargetGroupArn), err))
				continue
			}
			totalHealthy += health.Healthy
			totalUnhealthy += health.Unhealthy
			signal.TargetGroups = append(signal.TargetGroups, health)
		}
		signal.HealthyTargets = totalHealthy
		signal.UnhealthyTargets = totalUnhealthy

		reqCount, err := f.fetchMetric(ctx, "RequestCount", lbNamespace, []cwtypes.Dimension{{Name: aws.String("LoadBalancer"), Value: aws.String(lbDim)}}, start, end, 300, "Sum", false)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("RequestCount %s: %v", signal.Name, err))
		}
		if reqCount.Found {
			signal.RequestCount = reqCount.Value
		}

		if lb.Type == elbtypes.LoadBalancerTypeEnumApplication || lb.Type == elbtypes.LoadBalancerTypeEnumGateway {
			t4xx, err4 := f.fetchMetric(ctx, "HTTPCode_Target_4XX_Count", lbNamespace, []cwtypes.Dimension{{Name: aws.String("LoadBalancer"), Value: aws.String(lbDim)}}, start, end, 300, "Sum", false)
			if err4 != nil {
				warnings = append(warnings, fmt.Sprintf("Target4xx %s: %v", signal.Name, err4))
			}
			if t4xx.Found {
				signal.Target4XX = t4xx.Value
			}
			t5xx, err5 := f.fetchMetric(ctx, "HTTPCode_Target_5XX_Count", lbNamespace, []cwtypes.Dimension{{Name: aws.String("LoadBalancer"), Value: aws.String(lbDim)}}, start, end, 300, "Sum", false)
			if err5 != nil {
				warnings = append(warnings, fmt.Sprintf("Target5xx %s: %v", signal.Name, err5))
			}
			if t5xx.Found {
				signal.Target5XX = t5xx.Value
			}
			latency, latErr := f.fetchMetric(ctx, "TargetResponseTime", lbNamespace, []cwtypes.Dimension{{Name: aws.String("LoadBalancer"), Value: aws.String(lbDim)}}, start, end, 300, "p99", true)
			if latErr != nil {
				warnings = append(warnings, fmt.Sprintf("TargetResponseTime %s: %v", signal.Name, latErr))
			}
			signal.LatencyP99 = latency
			if !latency.Found {
				signal.LatencyP99.Name = "TargetResponseTime"
				signal.LatencyP99.Statistic = "p99"
			}
		}

		attrs, attrErr := f.elb.DescribeLoadBalancerAttributes(ctx, &elasticloadbalancingv2.DescribeLoadBalancerAttributesInput{
			LoadBalancerArn: lb.LoadBalancerArn,
		})
		if attrErr != nil {
			warnings = append(warnings, fmt.Sprintf("load balancer attributes %s: %v", signal.Name, attrErr))
		} else {
			signal.AccessLogs = parseAccessLogs(attrs.Attributes)
		}

		if signal.Name == "" {
			signal.Name = lbDim
		}

		out = append(out, signal)
	}
	return out, warnings
}

func (f *Fetcher) fetchMetric(ctx context.Context, name, namespace string, dims []cwtypes.Dimension, start, end time.Time, period int32, stat string, extended bool) (MetricSample, error) {
	if f.cloudwatch == nil {
		return MetricSample{Name: name}, fmt.Errorf("cloudwatch client not configured")
	}
	input := &cloudwatch.GetMetricStatisticsInput{
		Namespace:  aws.String(namespace),
		MetricName: aws.String(name),
		Dimensions: dims,
		StartTime:  aws.Time(start),
		EndTime:    aws.Time(end),
		Period:     aws.Int32(period),
	}
	if extended {
		input.ExtendedStatistics = []string{stat}
	} else {
		input.Statistics = []cwtypes.Statistic{cwtypes.Statistic(stat)}
	}
	resp, err := f.cloudwatch.GetMetricStatistics(ctx, input)
	if err != nil {
		return MetricSample{Name: name}, err
	}
	if len(resp.Datapoints) == 0 {
		return MetricSample{
			Name:          name,
			PeriodSeconds: period,
			Statistic:     stat,
			Unit:          string(cwtypes.StandardUnitCount),
			Found:         false,
		}, nil
	}
	latest := latestDatapoint(resp.Datapoints)
	sample := MetricSample{
		Name:          name,
		PeriodSeconds: period,
		Statistic:     stat,
		Found:         true,
	}
	if latest.Unit != "" {
		sample.Unit = string(latest.Unit)
	}
	if extended {
		if val, ok := latest.ExtendedStatistics[stat]; ok {
			sample.Value = val
		} else {
			sample.Found = false
		}
	} else {
		switch stat {
		case string(cwtypes.StatisticSum):
			sample.Value = aws.ToFloat64(latest.Sum)
		case string(cwtypes.StatisticMaximum):
			sample.Value = aws.ToFloat64(latest.Maximum)
		case string(cwtypes.StatisticMinimum):
			sample.Value = aws.ToFloat64(latest.Minimum)
		case string(cwtypes.StatisticSampleCount):
			sample.Value = aws.ToFloat64(latest.SampleCount)
		default:
			sample.Value = aws.ToFloat64(latest.Average)
		}
	}
	return sample, nil
}

func (f *Fetcher) logGroupExists(ctx context.Context, name string) (bool, error) {
	resp, err := f.logs.DescribeLogGroups(ctx, &cloudwatchlogs.DescribeLogGroupsInput{
		LogGroupNamePrefix: aws.String(name),
		Limit:              aws.Int32(1),
	})
	if err != nil {
		return false, err
	}
	for _, lg := range resp.LogGroups {
		if aws.ToString(lg.LogGroupName) == name {
			return true, nil
		}
	}
	return false, nil
}

func (f *Fetcher) sampleLogs(ctx context.Context, logGroup string, start, end time.Time, limit int32) ([]ControlPlaneLog, error) {
	resp, err := f.logs.FilterLogEvents(ctx, &cloudwatchlogs.FilterLogEventsInput{
		LogGroupName: aws.String(logGroup),
		StartTime:    aws.Int64(start.UnixMilli()),
		EndTime:      aws.Int64(end.UnixMilli()),
		Limit:        aws.Int32(limit),
	})
	if err != nil {
		return nil, err
	}
	var logs []ControlPlaneLog
	for _, ev := range resp.Events {
		logs = append(logs, ControlPlaneLog{
			Timestamp: time.UnixMilli(aws.ToInt64(ev.Timestamp)).UTC().Format(time.RFC3339Nano),
			Stream:    aws.ToString(ev.LogStreamName),
			Message:   strings.TrimSpace(aws.ToString(ev.Message)),
		})
	}
	return logs, nil
}

func isLoggingEnabled(cfg *ekstypes.Logging) bool {
	if cfg == nil || len(cfg.ClusterLogging) == 0 {
		return false
	}
	for _, item := range cfg.ClusterLogging {
		if item.Enabled != nil && *item.Enabled {
			return true
		}
	}
	return false
}

func (f *Fetcher) describeAutoScalingGroups(ctx context.Context, names []string) (map[string]*astypes.AutoScalingGroup, []string) {
	out := map[string]*astypes.AutoScalingGroup{}
	if len(names) == 0 {
		return out, nil
	}
	if f.autoscaling == nil {
		return out, []string{"autoscaling client not configured"}
	}
	var warnings []string
	chunk := 20
	for i := 0; i < len(names); i += chunk {
		end := i + chunk
		if end > len(names) {
			end = len(names)
		}
		resp, err := f.autoscaling.DescribeAutoScalingGroups(ctx, &autoscaling.DescribeAutoScalingGroupsInput{
			AutoScalingGroupNames: names[i:end],
		})
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("describe asg %v: %v", names[i:end], err))
			continue
		}
		for idx := range resp.AutoScalingGroups {
			group := resp.AutoScalingGroups[idx]
			out[aws.ToString(group.AutoScalingGroupName)] = &group
		}
	}
	return out, warnings
}

func summarizeInstances(instances []astypes.Instance) (int32, int32) {
	var current, ready int32
	for _, inst := range instances {
		current++
		state := string(inst.LifecycleState)
		if strings.EqualFold(state, "InService") && strings.EqualFold(aws.ToString(inst.HealthStatus), "Healthy") {
			ready++
		}
	}
	return current, ready
}

func (f *Fetcher) scalingActivities(ctx context.Context, asgName string, start time.Time) ([]ScalingEvent, error) {
	if f.autoscaling == nil {
		return nil, fmt.Errorf("autoscaling client not configured")
	}
	resp, err := f.autoscaling.DescribeScalingActivities(ctx, &autoscaling.DescribeScalingActivitiesInput{
		AutoScalingGroupName: aws.String(asgName),
		MaxRecords:           aws.Int32(5),
	})
	if err != nil {
		return nil, err
	}
	var events []ScalingEvent
	for _, act := range resp.Activities {
		if !start.IsZero() && aws.ToTime(act.StartTime).Before(start) {
			continue
		}
		events = append(events, ScalingEvent{
			Description: strings.TrimSpace(aws.ToString(act.Description)),
			Status:      string(act.StatusCode),
			Time:        aws.ToTime(act.StartTime).UTC().Format(time.RFC3339),
		})
	}
	return events, nil
}

func (f *Fetcher) clusterLoadBalancers(ctx context.Context, clusterName string) ([]elbtypes.LoadBalancer, []string) {
	var (
		lbs      []elbtypes.LoadBalancer
		warnings []string
	)
	marker := aws.String("")
	for {
		resp, err := f.elb.DescribeLoadBalancers(ctx, &elasticloadbalancingv2.DescribeLoadBalancersInput{
			Marker: marker,
		})
		if err != nil {
			return lbs, []string{fmt.Sprintf("describe load balancers: %v", err)}
		}
		lbs = append(lbs, resp.LoadBalancers...)
		if resp.NextMarker == nil || *resp.NextMarker == "" {
			break
		}
		marker = resp.NextMarker
	}
	if len(lbs) == 0 {
		return lbs, warnings
	}
	filtered, filterWarnings := f.filterLBByClusterTag(ctx, clusterName, lbs)
	warnings = append(warnings, filterWarnings...)
	return filtered, warnings
}

func (f *Fetcher) filterLBByClusterTag(ctx context.Context, clusterName string, lbs []elbtypes.LoadBalancer) ([]elbtypes.LoadBalancer, []string) {
	var (
		filtered []elbtypes.LoadBalancer
		warnings []string
	)
	if len(lbs) == 0 {
		return filtered, warnings
	}
	tagKey := fmt.Sprintf("kubernetes.io/cluster/%s", clusterName)
	batch := 20
	for i := 0; i < len(lbs); i += batch {
		end := i + batch
		if end > len(lbs) {
			end = len(lbs)
		}
		arns := make([]string, 0, end-i)
		for _, lb := range lbs[i:end] {
			if lb.LoadBalancerArn != nil {
				arns = append(arns, *lb.LoadBalancerArn)
			}
		}
		resp, err := f.elb.DescribeTags(ctx, &elasticloadbalancingv2.DescribeTagsInput{
			ResourceArns: arns,
		})
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("describe lb tags: %v", err))
			continue
		}
		tagged := map[string]bool{}
		for _, desc := range resp.TagDescriptions {
			if desc.ResourceArn == nil {
				continue
			}
			for _, tag := range desc.Tags {
				if aws.ToString(tag.Key) == tagKey && (aws.ToString(tag.Value) == "owned" || aws.ToString(tag.Value) == "shared") {
					tagged[aws.ToString(desc.ResourceArn)] = true
					break
				}
			}
		}
		for _, lb := range lbs[i:end] {
			if tagged[aws.ToString(lb.LoadBalancerArn)] {
				filtered = append(filtered, lb)
			}
		}
	}
	return filtered, warnings
}

func (f *Fetcher) targetGroupsForLB(ctx context.Context, lbArn string) ([]elbtypes.TargetGroup, []string) {
	var (
		targetGroups []elbtypes.TargetGroup
		warnings     []string
	)
	marker := aws.String("")
	for {
		resp, err := f.elb.DescribeTargetGroups(ctx, &elasticloadbalancingv2.DescribeTargetGroupsInput{
			LoadBalancerArn: aws.String(lbArn),
			Marker:          marker,
		})
		if err != nil {
			return targetGroups, []string{fmt.Sprintf("describe target groups: %v", err)}
		}
		targetGroups = append(targetGroups, resp.TargetGroups...)
		if resp.NextMarker == nil || *resp.NextMarker == "" {
			break
		}
		marker = resp.NextMarker
	}
	return targetGroups, warnings
}

func (f *Fetcher) targetGroupHealth(ctx context.Context, tgArn string) (TargetGroupHealth, error) {
	if f.elb == nil {
		return TargetGroupHealth{}, fmt.Errorf("elb client not configured")
	}
	resp, err := f.elb.DescribeTargetHealth(ctx, &elasticloadbalancingv2.DescribeTargetHealthInput{
		TargetGroupArn: aws.String(tgArn),
	})
	if err != nil {
		return TargetGroupHealth{}, err
	}
	healthy := int32(0)
	unhealthy := int32(0)
	for _, desc := range resp.TargetHealthDescriptions {
		if desc.TargetHealth == nil {
			continue
		}
		switch desc.TargetHealth.State {
		case elbtypes.TargetHealthStateEnumHealthy:
			healthy++
		case elbtypes.TargetHealthStateEnumInitial, elbtypes.TargetHealthStateEnumDraining:
			// treat as current but not unhealthy
		default:
			unhealthy++
		}
	}
	return TargetGroupHealth{
		ARN:       tgArn,
		Name:      targetGroupDimension(tgArn),
		Healthy:   healthy,
		Unhealthy: unhealthy,
	}, nil
}

func parseAccessLogs(attrs []elbtypes.LoadBalancerAttribute) AccessLogLocation {
	cfg := AccessLogLocation{}
	for _, attr := range attrs {
		key := aws.ToString(attr.Key)
		switch key {
		case "access_logs.s3.enabled":
			cfg.Enabled = strings.EqualFold(aws.ToString(attr.Value), "true")
		case "access_logs.s3.bucket":
			cfg.Bucket = aws.ToString(attr.Value)
		case "access_logs.s3.prefix":
			cfg.Prefix = aws.ToString(attr.Value)
		}
	}
	return cfg
}

func loadBalancerDimension(arn string) string {
	parts := strings.SplitN(arn, "loadbalancer/", 2)
	if len(parts) == 2 {
		return parts[1]
	}
	return arn
}

func targetGroupDimension(arn string) string {
	parts := strings.SplitN(arn, "targetgroup/", 2)
	if len(parts) == 2 {
		return parts[1]
	}
	return arn
}

func autoScalingGroupNames(res *ekstypes.NodegroupResources) []string {
	if res == nil || len(res.AutoScalingGroups) == 0 {
		return nil
	}
	var out []string
	for _, asg := range res.AutoScalingGroups {
		if asg.Name != nil && strings.TrimSpace(*asg.Name) != "" {
			out = append(out, *asg.Name)
		}
	}
	return out
}

func lbNamespace(lbType elbtypes.LoadBalancerTypeEnum) string {
	switch lbType {
	case elbtypes.LoadBalancerTypeEnumNetwork:
		return "AWS/NetworkELB"
	case elbtypes.LoadBalancerTypeEnumGateway:
		return "AWS/GatewayELB"
	default:
		return "AWS/ApplicationELB"
	}
}

func latestDatapoint(points []cwtypes.Datapoint) cwtypes.Datapoint {
	if len(points) == 0 {
		return cwtypes.Datapoint{}
	}
	sort.Slice(points, func(i, j int) bool {
		return aws.ToTime(points[i].Timestamp).After(aws.ToTime(points[j].Timestamp))
	})
	return points[0]
}
