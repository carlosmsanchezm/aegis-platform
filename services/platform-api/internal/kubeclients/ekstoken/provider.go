package ekstoken

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	awscfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/aws/smithy-go/middleware"
	smithyhttp "github.com/aws/smithy-go/transport/http"
)

const (
	tokenPrefix     = "k8s-aws-v1."
	clusterIDHeader = "x-k8s-aws-id"
	tokenExpiry     = 14 * time.Minute
)

// Provider generates short-lived EKS bearer tokens using the AWS SDK.
type Provider struct {
	Region      string
	ClusterName string
	RoleARN     string // optional — for cross-account/multi-tenant access
	ExternalID  string // optional — paired with RoleARN
}

// Presigner abstracts the STS PresignGetCallerIdentity call for testing.
type Presigner interface {
	PresignGetCallerIdentity(ctx context.Context, params *sts.GetCallerIdentityInput, optFns ...func(*sts.PresignOptions)) (*v4.PresignedHTTPRequest, error)
}

// Token returns a bearer token valid for ~14 minutes.
func (p *Provider) Token(ctx context.Context) (string, time.Time, error) {
	return p.TokenWithPresigner(ctx, nil)
}

// TokenWithPresigner allows injecting a custom presigner for testing.
func (p *Provider) TokenWithPresigner(ctx context.Context, presigner Presigner) (string, time.Time, error) {
	if p.Region == "" {
		return "", time.Time{}, fmt.Errorf("region is required")
	}
	if p.ClusterName == "" {
		return "", time.Time{}, fmt.Errorf("cluster name is required")
	}

	if presigner == nil {
		cfg, err := p.loadAWSConfig(ctx)
		if err != nil {
			return "", time.Time{}, fmt.Errorf("load AWS config: %w", err)
		}
		stsClient := sts.NewFromConfig(cfg)
		presigner = sts.NewPresignClient(stsClient)
	}

	presigned, err := presigner.PresignGetCallerIdentity(ctx, &sts.GetCallerIdentityInput{}, func(po *sts.PresignOptions) {
		po.ClientOptions = append(po.ClientOptions, func(o *sts.Options) {
			o.APIOptions = append(o.APIOptions, addClusterIDHeaderMiddleware(p.ClusterName))
		})
	})
	if err != nil {
		return "", time.Time{}, fmt.Errorf("presign GetCallerIdentity: %w", err)
	}

	token := tokenPrefix + base64.RawURLEncoding.EncodeToString([]byte(presigned.URL))
	expiry := time.Now().Add(tokenExpiry)
	return token, expiry, nil
}

func (p *Provider) loadAWSConfig(ctx context.Context) (aws.Config, error) {
	cfg, err := awscfg.LoadDefaultConfig(ctx, awscfg.WithRegion(p.Region))
	if err != nil {
		return aws.Config{}, err
	}

	roleARN := strings.TrimSpace(p.RoleARN)
	if roleARN == "" {
		return cfg, nil
	}

	stsClient := sts.NewFromConfig(cfg)
	creds := stscreds.NewAssumeRoleProvider(stsClient, roleARN, func(o *stscreds.AssumeRoleOptions) {
		o.RoleSessionName = "aegis-platform-api"
		o.Duration = 30 * time.Minute
		if external := strings.TrimSpace(p.ExternalID); external != "" {
			o.ExternalID = aws.String(external)
		}
	})
	cfg.Credentials = aws.NewCredentialsCache(creds)
	return cfg, nil
}

// addClusterIDHeaderMiddleware returns a Smithy API option that injects the
// x-k8s-aws-id header into the HTTP request. This is the standard mechanism
// EKS uses to scope presigned STS tokens to a specific cluster.
func addClusterIDHeaderMiddleware(clusterName string) func(stack *middleware.Stack) error {
	return func(stack *middleware.Stack) error {
		return stack.Build.Add(middleware.BuildMiddlewareFunc("AddEKSClusterIDHeader", func(
			ctx context.Context, in middleware.BuildInput, next middleware.BuildHandler,
		) (middleware.BuildOutput, middleware.Metadata, error) {
			if req, ok := in.Request.(*smithyhttp.Request); ok {
				req.Header.Set(clusterIDHeader, clusterName)
			}
			return next.HandleBuild(ctx, in)
		}), middleware.Before)
	}
}
