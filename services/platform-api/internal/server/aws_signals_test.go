package server

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	aegis "github.com/yourorg/aegis/proto/aegis/v1"
	"github.com/yourorg/aegis/services/platform-api/internal/authz"
	"github.com/yourorg/aegis/services/platform-api/internal/aws/observability"
	"github.com/yourorg/aegis/services/platform-api/internal/config"
	mw "github.com/yourorg/aegis/services/platform-api/internal/server/mw"
	"github.com/yourorg/aegis/services/platform-api/internal/store"
)

type stubAwsFetcher struct {
	signals *observability.Signals
	err     error
}

func (s stubAwsFetcher) Fetch(_ context.Context, _ observability.FetchInput) (*observability.Signals, error) {
	return s.signals, s.err
}

func TestGetAwsClusterSignalsValidation(t *testing.T) {
	srv := &Server{}
	_, err := srv.GetAwsClusterSignals(context.Background(), &aegis.GetAwsSignalsRequest{})
	require.Error(t, err)
	stErr, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.InvalidArgument, stErr.Code())
}

func TestGetAwsClusterSignalsSuccess(t *testing.T) {
	policy, err := authz.LoadPolicyFromEnv(config.DefaultRoleBindingsJSON())
	require.NoError(t, err)
	st := store.NewMemStore()
	st.PutProject(&aegis.Project{
		Id: "proj-1",
		Aws: &aegis.ProjectAwsCredentials{
			AccountId:  "123456789012",
			RoleArn:    "arn:aws:iam::123456789012:role/test",
			ExternalId: "ext-1",
		},
	})
	st.UpsertClusterFromRegister(&aegis.ClusterRegisterRequest{
		ClusterId: "cluster-1",
		Region:    "us-east-1",
		Labels:    map[string]string{"aegis.yourorg.dev/projectId": "proj-1"},
	})

	fetcher := stubAwsFetcher{
		signals: &observability.Signals{
			ControlPlane: observability.ControlPlaneSignals{
				Status:         "ACTIVE",
				LoggingEnabled: true,
				LogGroup:       "/aws/eks/cluster-1/cluster",
				LatencyP99: observability.MetricSample{
					Name:          "APIServerLatencyP99",
					Value:         120.5,
					Statistic:     "p99",
					Found:         true,
					PeriodSeconds: 300,
				},
				Error5XXRate: observability.MetricSample{
					Name:          "APIServer5XXErrorRate",
					Value:         0.05,
					Statistic:     "Average",
					Found:         true,
					PeriodSeconds: 300,
				},
				LogSamples: []observability.ControlPlaneLog{
					{Timestamp: "2024-01-01T00:00:00Z", Stream: "api", Message: "ok"},
				},
			},
			Nodegroups: []observability.NodegroupSignals{
				{
					Name:             "ng-1",
					Status:           "ACTIVE",
					Desired:          3,
					Current:          3,
					Ready:            3,
					AutoScalingGroup: "asg-1",
					Issues:           []string{"issue"},
					GPU:              true,
					GPUFlavor:        "a10g",
					InstanceType:     "g5.xlarge",
					ScalingEvents: []observability.ScalingEvent{
						{Description: "scale", Status: "Successful", Time: "2024-01-01T00:00:00Z"},
					},
				},
			},
			LoadBalancers: []observability.LoadBalancerSignals{
				{
					ARN:            "arn:aws:elasticloadbalancing:us-east-1:123456789012:loadbalancer/app/lb-1/1",
					Name:           "lb-1",
					Type:           "application",
					Scheme:         "internet-facing",
					State:          "active",
					RequestCount:   10,
					Target4XX:      1,
					Target5XX:      0,
					HealthyTargets: 2,
					LatencyP99: observability.MetricSample{
						Name:          "TargetResponseTime",
						Value:         110,
						Statistic:     "p99",
						Found:         true,
						PeriodSeconds: 300,
					},
					TargetGroups: []observability.TargetGroupHealth{
						{ARN: "tg-1", Name: "tg-1", Healthy: 2, Unhealthy: 0},
					},
					AccessLogs: observability.AccessLogLocation{
						Enabled: true,
						Bucket:  "bucket",
						Prefix:  "logs",
					},
				},
			},
			Warnings: []string{"partial"},
		},
	}
	srv := &Server{
		log:               zap.NewNop(),
		store:             st,
		authzPolicy:       policy,
		awsFetcherFactory: func(context.Context, string, projectAWSCredentials) (awsSignalsFetcher, error) { return fetcher, nil },
	}
	ctx := mw.ContextWithIdentity(context.Background(), &mw.Identity{
		Subject:  "user",
		ClientID: "backstage",
		Roles:    []string{"workspace-admin"},
	})

	resp, err := srv.GetAwsClusterSignals(ctx, &aegis.GetAwsSignalsRequest{
		ProjectId: "proj-1",
		ClusterId: "cluster-1",
	})
	require.NoError(t, err)
	require.Equal(t, "us-east-1", resp.GetRegion())
	require.Len(t, resp.GetNodegroups(), 1)
	require.Equal(t, int32(3), resp.GetNodegroups()[0].GetDesired())
	require.True(t, resp.GetNodegroups()[0].GetGpu())
	require.Equal(t, "a10g", resp.GetNodegroups()[0].GetGpuFlavor())
	require.Equal(t, "g5.xlarge", resp.GetNodegroups()[0].GetInstanceType())
	require.Equal(t, float64(10), resp.GetLoadBalancers()[0].GetRequestCount())
	require.Equal(t, []string{"partial"}, resp.GetWarnings())
}
