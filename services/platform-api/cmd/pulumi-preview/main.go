package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	"github.com/pulumi/pulumi/sdk/v3/go/auto/optpreview"
	apitype "github.com/pulumi/pulumi/sdk/v3/go/common/apitype"
	infraapi "github.com/yourorg/aegis/services/platform-api/api/v1alpha1"
	"github.com/yourorg/aegis/services/platform-api/internal/provisioning/pulumi/aws"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	var (
		fileFlag       = flag.String("file", "", "Path to a ProjectInfra manifest on disk")
		nameFlag       = flag.String("name", "", "Name of the ProjectInfra resource to fetch")
		namespaceFlag  = flag.String("namespace", "", "Namespace of the ProjectInfra resource to fetch")
		kubeconfigFlag = flag.String("kubeconfig", os.Getenv("KUBECONFIG"), "Path to kubeconfig when fetching from a cluster")
	)

	flag.Parse()

	logger, _ := zap.NewProduction()
	defer func() { _ = logger.Sync() }()

	infra, err := loadProjectInfra(ctx, *fileFlag, *namespaceFlag, *nameFlag, *kubeconfigFlag)
	if err != nil {
		logger.Error("failed to load ProjectInfra", zap.Error(err))
		os.Exit(1)
	}

	awsSpec := infra.Spec.Aws
	if awsSpec == nil {
		logger.Error("ProjectInfra does not include an AWS spec", zap.String("name", infra.Name), zap.String("namespace", infra.Namespace))
		os.Exit(1)
	}

	runner := aws.NewRunner(logger, nil)
	stack, programCfg, err := runner.NewStack(ctx, infra, awsSpec)
	if err != nil {
		logger.Error("failed to prepare pulumi stack", zap.Error(err))
		os.Exit(1)
	}

	writer := os.Stdout
	result, err := stack.Preview(ctx, optpreview.ProgressStreams(writer))
	if err != nil {
		logger.Error("pulumi preview failed", zap.Error(err))
		os.Exit(1)
	}

	summary := result.ChangeSummary
	getSummary := func(op apitype.OpType) int {
		if summary == nil {
			return 0
		}
		if v, ok := summary[op]; ok {
			return v
		}
		return 0
	}
	logger.Info("pulumi preview completed",
		zap.String("project", programCfg.ProjectID),
		zap.String("region", programCfg.Region),
		zap.Int("adds", getSummary(apitype.OpCreate)),
		zap.Int("updates", getSummary(apitype.OpUpdate)),
		zap.Int("replaces", getSummary(apitype.OpReplace)),
		zap.Int("deletes", getSummary(apitype.OpDelete)),
	)
}

func loadProjectInfra(ctx context.Context, filePath, namespace, name, kubeconfig string) (*infraapi.ProjectInfra, error) {
	switch {
	case strings.TrimSpace(filePath) != "":
		return loadProjectInfraFromFile(filePath)
	case strings.TrimSpace(name) != "" && strings.TrimSpace(namespace) != "":
		return loadProjectInfraFromCluster(ctx, namespace, name, kubeconfig)
	default:
		return nil, errors.New("either --file or both --namespace/--name are required")
	}
}

func loadProjectInfraFromFile(path string) (*infraapi.ProjectInfra, error) {
	payload, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read project infra file: %w", err)
	}
	var infra infraapi.ProjectInfra
	if err := yaml.Unmarshal(payload, &infra); err != nil {
		return nil, fmt.Errorf("unmarshal project infra yaml: %w", err)
	}
	return &infra, nil
}

func loadProjectInfraFromCluster(ctx context.Context, namespace, name, kubeconfigPath string) (*infraapi.ProjectInfra, error) {
	cfg, err := resolveRESTConfig(kubeconfigPath)
	if err != nil {
		return nil, err
	}
	cfg.UserAgent = "aegis-pulumi-preview"

	scheme := runtime.NewScheme()
	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		return nil, fmt.Errorf("add client-go scheme: %w", err)
	}
	if err := infraapi.AddToScheme(scheme); err != nil {
		return nil, fmt.Errorf("add aegis scheme: %w", err)
	}
	if err := corev1.AddToScheme(scheme); err != nil {
		return nil, fmt.Errorf("add corev1 scheme: %w", err)
	}

	cli, err := client.New(cfg, client.Options{Scheme: scheme})
	if err != nil {
		return nil, fmt.Errorf("create kubernetes client: %w", err)
	}

	var infra infraapi.ProjectInfra
	key := types.NamespacedName{Name: name, Namespace: namespace}
	if err := cli.Get(ctx, key, &infra); err != nil {
		return nil, fmt.Errorf("get ProjectInfra %s/%s: %w", namespace, name, err)
	}
	return &infra, nil
}

func resolveRESTConfig(kubeconfigPath string) (*rest.Config, error) {
	path := strings.TrimSpace(kubeconfigPath)
	if path != "" {
		cfg, err := clientcmd.BuildConfigFromFlags("", path)
		if err != nil {
			return nil, fmt.Errorf("load kubeconfig %q: %w", path, err)
		}
		return cfg, nil
	}

	cfg, err := rest.InClusterConfig()
	if err == nil {
		return cfg, nil
	}

	return nil, errors.New("kubeconfig path not provided and in-cluster configuration failed")
}
