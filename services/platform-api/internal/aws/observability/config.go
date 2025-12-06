package observability

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awscfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/autoscaling"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

// AWSConfigInput declares the AWS region and optional assume-role configuration for fetching signals.
type AWSConfigInput struct {
	Region     string
	RoleARN    string
	ExternalID string
}

// BuildAWSConfig loads the default AWS config and applies assume-role credentials when provided.
func BuildAWSConfig(ctx context.Context, input AWSConfigInput) (aws.Config, error) {
	region := strings.TrimSpace(input.Region)
	if region == "" {
		return aws.Config{}, fmt.Errorf("region is required")
	}
	cfg, err := awscfg.LoadDefaultConfig(ctx, awscfg.WithRegion(region))
	if err != nil {
		return aws.Config{}, err
	}
	roleARN := strings.TrimSpace(input.RoleARN)
	if roleARN == "" {
		return cfg, nil
	}
	stsClient := sts.NewFromConfig(cfg)
	opts := func(o *stscreds.AssumeRoleOptions) {
		o.RoleSessionName = "aegis-platform-api"
		o.Duration = 30 * time.Minute
		if external := strings.TrimSpace(input.ExternalID); external != "" {
			o.ExternalID = aws.String(external)
		}
	}
	creds := stscreds.NewAssumeRoleProvider(stsClient, roleARN, opts)
	cfg.Credentials = aws.NewCredentialsCache(creds)
	return cfg, nil
}

// NewFetcher builds a fetcher backed by AWS SDK clients created from the supplied configuration.
func NewFetcher(ctx context.Context, input AWSConfigInput) (*Fetcher, error) {
	cfg, err := BuildAWSConfig(ctx, input)
	if err != nil {
		return nil, err
	}
	return NewFetcherFromConfig(cfg), nil
}

// NewFetcherFromConfig wraps an aws.Config with the concrete AWS service clients used by the fetcher.
func NewFetcherFromConfig(cfg aws.Config) *Fetcher {
	return &Fetcher{
		cloudwatch:  cloudwatch.NewFromConfig(cfg),
		logs:        cloudwatchlogs.NewFromConfig(cfg),
		eks:         eks.NewFromConfig(cfg),
		autoscaling: autoscaling.NewFromConfig(cfg),
		elb:         elasticloadbalancingv2.NewFromConfig(cfg),
	}
}
