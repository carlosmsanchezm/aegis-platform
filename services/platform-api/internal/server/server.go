package server

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	aegisv1alpha1 "github.com/yourorg/aegis/agents/k8s-agent/api/v1alpha1"
	workspacecfg "github.com/yourorg/aegis/pkg/workspace"
	aegis "github.com/yourorg/aegis/proto/aegis/v1"
	"github.com/yourorg/aegis/services/platform-api/internal/authz"
	"github.com/yourorg/aegis/services/platform-api/internal/aws/observability"
	"github.com/yourorg/aegis/services/platform-api/internal/config"
	"github.com/yourorg/aegis/services/platform-api/internal/kubeclients"
	"github.com/yourorg/aegis/services/platform-api/internal/placement"
	"github.com/yourorg/aegis/services/platform-api/internal/provisioning/pulumi/aws"
	mw "github.com/yourorg/aegis/services/platform-api/internal/server/mw"
	"github.com/yourorg/aegis/services/platform-api/internal/store"
	batchv1 "k8s.io/api/batch/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	unstructuredapi "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type kubeClientProvider interface {
	ClientFor(clusterID string) (client.Client, error)
	RestConfigFor(clusterID string) (*rest.Config, error)
	HasKubeconfig(clusterID string) bool
	Dir() string
}

func kubeClientProviderConfigured(p kubeClientProvider) bool {
	if p == nil {
		return false
	}
	v := reflect.ValueOf(p)
	switch v.Kind() {
	case reflect.Pointer, reflect.Interface, reflect.Map, reflect.Slice, reflect.Func, reflect.Chan:
		return !v.IsNil()
	default:
		return true
	}
}

func (s *Server) namespaceForProject(projectID string) string {
	if ns := strings.TrimSpace(aws.WorkloadsNamespaceForProject(projectID)); ns != "" {
		return ns
	}
	return s.targetNamespace
}

type Server struct {
	aegis.UnimplementedAegisPlatformServer
	log                  *zap.Logger
	store                store.Store
	kubeClients          kubeClientProvider
	policyOverlay        *placement.PolicyOverlay
	targetNamespace      string
	proxyBaseURL         string
	proxyAudience        string
	proxySecret          []byte
	proxyTokenTTL        time.Duration
	workspaceEnvDefaults map[string]string
	autoBootstrap        bool
	authzPolicy          *authz.Policy
	infraClient          client.Client
	infraNamespace       string
	clusterProfiles      map[string]*clusterProfileTemplate
	awsFetcherFactory    awsFetcherFactory
}

type awsSignalsFetcher interface {
	Fetch(ctx context.Context, in observability.FetchInput) (*observability.Signals, error)
}

type awsFetcherFactory func(ctx context.Context, region string, creds projectAWSCredentials) (awsSignalsFetcher, error)

type proxyClaims struct {
	Sub     string `json:"sub"`
	Wid     string `json:"wid"`
	Dest    string `json:"dest"`
	DNS     string `json:"dns,omitempty"`
	Cluster string `json:"cluster,omitempty"`
	OneTime bool   `json:"one_time,omitempty"`
	jwt.RegisteredClaims
}

type sessionContext struct {
	workload     *aegis.Workload
	workspace    *aegis.WorkspaceSpec
	port         int32
	alias        string
	internalHost string
	dest         string
	proxyURL     string
}

const (
	statusPlaced  = "PLACED"
	statusRunning = "RUNNING"

	heartbeatTTL = 45 * time.Second

	labelWorkloadID = "aegis.workload/id"

	uiStatusQueuedByKueue = "QUEUED_BY_KUEUE"
	uiStatusSubmitted     = "SUBMITTED"

	defaultProxyTokenTTLSeconds = 300
	svcClusterDomainSuffix      = ".svc.cluster.local"
	maxSessionTTL               = 5 * time.Minute
)

var (
	mPlaced = promauto.NewCounterVec(
		prometheus.CounterOpts{Name: "aegis_workload_placed_total", Help: "Workloads placed"},
		[]string{"flavor"},
	)
	mLeased = promauto.NewCounter(
		prometheus.CounterOpts{Name: "aegis_workload_leased_total", Help: "Workloads leased"},
	)
	mAcked = promauto.NewCounterVec(
		prometheus.CounterOpts{Name: "aegis_workload_acked_total", Help: "Workloads acked by final status"},
		[]string{"status", "backend"},
	)
	mQueueWait = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "aegis_workload_queue_wait_seconds",
			Help:    "Time between placement and lease per queue/flavor",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"queue", "flavor"},
	)
	mBudgetActual = promauto.NewGaugeVec(
		prometheus.GaugeOpts{Name: "aegis_budget_actual_usd", Help: "Actual spend in USD for current UTC month"},
		[]string{"project", "queue"},
	)
	mBudgetReserved = promauto.NewGaugeVec(
		prometheus.GaugeOpts{Name: "aegis_budget_reserved_usd", Help: "Reserved (estimated) USD for current UTC month"},
		[]string{"project", "queue"},
	)
	mBudgetDenied = promauto.NewCounterVec(
		prometheus.CounterOpts{Name: "aegis_budget_denied_total", Help: "Budget denials at submit"},
		[]string{"project", "queue", "reason"},
	)
	mBudgetOverrun = promauto.NewCounterVec(
		prometheus.CounterOpts{Name: "aegis_budget_overrun_total", Help: "Budget soft-policy overruns"},
		[]string{"project", "queue"},
	)
	mEstimateHist = promauto.NewHistogram(
		prometheus.HistogramOpts{Name: "aegis_workload_estimated_cost_usd", Help: "Estimated workload USD", Buckets: prometheus.ExponentialBuckets(0.01, 2, 15)},
	)
)

func New(log *zap.Logger, st store.Store, clients *kubeclients.Manager, namespace string, overlay *placement.PolicyOverlay, infraClient client.Client, infraNamespace string) *Server {
	if namespace == "" {
		namespace = "default"
	}
	baseURL := strings.TrimRight(os.Getenv("AEGIS_PROXY_BASE_URL"), "/")
	audience := os.Getenv("AEGIS_PROXY_EXPECTED_AUDIENCE")
	if audience == "" {
		audience = "aegis-proxy"
	}
	secret := os.Getenv("AEGIS_PROXY_JWT_SECRET")
	ttlSeconds := getEnvInt("AEGIS_PROXY_TOKEN_TTL_SECONDS", defaultProxyTokenTTLSeconds)
	if ttlSeconds <= 0 {
		ttlSeconds = defaultProxyTokenTTLSeconds
	}
	defaults := workspacecfg.DefaultEnv()
	if commit := strings.TrimSpace(os.Getenv("AEGIS_VSCODE_COMMIT")); commit != "" {
		defaults[workspacecfg.EnvVSCodeCommit] = commit
	}
	if quality := strings.TrimSpace(os.Getenv("AEGIS_VSCODE_QUALITY")); quality != "" {
		defaults[workspacecfg.EnvVSCodeQuality] = quality
	}
	if val := strings.TrimSpace(os.Getenv("AEGIS_WORKSPACE_PUID")); val != "" {
		defaults[workspacecfg.EnvPUID] = val
	}
	if val := strings.TrimSpace(os.Getenv("AEGIS_WORKSPACE_PGID")); val != "" {
		defaults[workspacecfg.EnvPGID] = val
	}
	if val := strings.TrimSpace(os.Getenv("AEGIS_WORKSPACE_PASSWORD_ACCESS")); val != "" {
		defaults[workspacecfg.EnvPasswordAccess] = val
	}
	if val := strings.TrimSpace(os.Getenv("AEGIS_WORKSPACE_USER_NAME")); val != "" {
		defaults[workspacecfg.EnvUserName] = val
	}
	if val := strings.TrimSpace(os.Getenv("AEGIS_WORKSPACE_USER_PASSWORD")); val != "" {
		defaults[workspacecfg.EnvUserPassword] = val
	}
	if overlay == nil {
		overlay = placement.NewPolicyOverlay()
	}
	policy, err := authz.LoadPolicyFromEnv(config.DefaultRoleBindingsJSON())
	if err != nil {
		if log != nil {
			log.Fatal("failed to load authorization policy", zap.Error(err))
		}
		panic(fmt.Errorf("failed to load authorization policy: %w", err))
	}
	if strings.TrimSpace(infraNamespace) == "" {
		infraNamespace = defaultInfraNamespace
	}
	profiles := defaultClusterProfiles()
	if profiles == nil {
		profiles = map[string]*clusterProfileTemplate{}
	}
	awsFactory := func(ctx context.Context, region string, creds projectAWSCredentials) (awsSignalsFetcher, error) {
		return observability.NewFetcher(ctx, observability.AWSConfigInput{
			Region:     region,
			RoleARN:    creds.RoleARN,
			ExternalID: creds.ExternalID,
		})
	}
	return &Server{
		log:                  log,
		store:                st,
		kubeClients:          clients,
		policyOverlay:        overlay,
		targetNamespace:      namespace,
		proxyBaseURL:         baseURL,
		proxyAudience:        audience,
		proxySecret:          []byte(secret),
		proxyTokenTTL:        time.Duration(ttlSeconds) * time.Second,
		workspaceEnvDefaults: defaults,
		autoBootstrap:        getEnvBool("AEGIS_AUTO_BOOTSTRAP_WORKSPACES", false),
		authzPolicy:          policy,
		infraClient:          infraClient,
		infraNamespace:       infraNamespace,
		clusterProfiles:      profiles,
		awsFetcherFactory:    awsFactory,
	}
}

func getEnvInt(key string, def int64) int64 {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			return n
		}
	}
	return def
}

func getEnvBool(key string, def bool) bool {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return def
	}
	if parsed, err := strconv.ParseBool(v); err == nil {
		return parsed
	}
	return def
}

func (s *Server) defaultMaxRuntimeSeconds(q *aegis.Queue) int64 {
	if q != nil && q.GetDefaultMaxDurationSeconds() > 0 {
		return q.GetDefaultMaxDurationSeconds()
	}
	return getEnvInt("AEGIS_DEFAULT_MAX_RUNTIME_SECONDS", 3600)
}

func (s *Server) CreateProject(ctx context.Context, req *aegis.CreateProjectRequest) (*aegis.Project, error) {
	if req == nil || req.Project == nil {
		err := status.Error(codes.InvalidArgument, "project payload required")
		s.log.Warn("create project failed", zap.Error(err))
		return nil, err
	}
	p := req.Project
	if err := s.authorize(ctx, p.GetId(), "", "createProject"); err != nil {
		return nil, err
	}
	awsCreds := mergeProjectAwsDefaults(sanitizeProjectAws(p.GetAws()))
	if awsCreds != nil {
		if err := validateProjectAwsCredentials(awsCreds); err != nil {
			errStatus := status.Error(codes.InvalidArgument, err.Error())
			s.log.Warn("create project rejected; invalid aws credentials",
				zap.String("project_id", p.GetId()),
				zap.Error(errStatus),
			)
			return nil, errStatus
		}
	}
	p.Aws = awsCreds
	p.Annotations = mergeProjectAnnotations(p.GetAnnotations(), awsCreds)
	s.store.PutProject(p)
	s.log.Info("project upserted", zap.String("project_id", p.GetId()), zap.String("owner_group", p.GetOwnerGroup()))
	populateProjectAwsFromAnnotations(p)
	return p, nil
}

func (s *Server) ListProjects(ctx context.Context, _ *aegis.ListProjectsRequest) (*aegis.ListProjectsResponse, error) {
	all := s.store.ListProjects()
	authorized := make([]*aegis.Project, 0, len(all))
	for _, project := range all {
		if err := s.authorize(ctx, project.GetId(), "", "listProjects"); err != nil {
			if status.Code(err) == codes.PermissionDenied {
				continue
			}
			return nil, err
		}
		populateProjectAwsFromAnnotations(project)
		authorized = append(authorized, project)
	}
	return &aegis.ListProjectsResponse{Items: authorized}, nil
}

func (s *Server) UpsertBudget(ctx context.Context, req *aegis.UpsertBudgetRequest) (*aegis.Budget, error) {
	if req == nil || req.Budget == nil {
		err := status.Error(codes.InvalidArgument, "budget payload required")
		s.log.Warn("upsert budget failed", zap.Error(err))
		return nil, err
	}
	b := req.Budget
	if err := s.authorize(ctx, b.GetProjectId(), b.GetQueue(), "upsertBudget"); err != nil {
		return nil, err
	}
	s.store.PutBudget(b)
	s.log.Info("budget upserted",
		zap.String("project_id", b.GetProjectId()),
		zap.String("queue", b.GetQueue()),
		zap.Float64("limit_usd", b.GetLimitUsd()),
		zap.String("policy_mode", b.GetPolicyMode()),
	)
	return b, nil
}

func (s *Server) UpsertFlavor(ctx context.Context, req *aegis.UpsertFlavorRequest) (*aegis.Flavor, error) {
	if req == nil || req.Flavor == nil {
		err := status.Error(codes.InvalidArgument, "flavor payload required")
		s.log.Warn("upsert flavor failed", zap.Error(err))
		return nil, err
	}
	f := req.Flavor
	s.store.PutFlavor(f)
	s.log.Info("flavor upserted", zap.String("flavor", f.GetName()), zap.String("chip", f.GetChip()), zap.String("mig_profile", f.GetMigProfile()), zap.Bool("rdma_required", f.GetRdmaRequired()))
	return f, nil
}

func (s *Server) UpsertQueue(ctx context.Context, req *aegis.UpsertQueueRequest) (*aegis.Queue, error) {
	if req == nil || req.Queue == nil {
		err := status.Error(codes.InvalidArgument, "queue payload required")
		s.log.Warn("upsert queue failed", zap.Error(err))
		return nil, err
	}
	q := req.Queue
	if err := s.authorize(ctx, q.GetProjectId(), q.GetName(), "upsertQueue"); err != nil {
		return nil, err
	}
	s.store.PutQueue(q)
	s.log.Info("queue upserted", zap.String("queue", q.GetName()), zap.String("project_id", q.GetProjectId()), zap.String("priority_tier", q.GetPriorityTier()), zap.Int("allowed_flavors", len(q.GetAllowedFlavors())))
	return q, nil
}

func (s *Server) RegisterCluster(ctx context.Context, r *aegis.ClusterRegisterRequest) (*aegis.ClusterRegisterResponse, error) {
	if r == nil {
		err := status.Error(codes.InvalidArgument, "cluster payload required")
		s.log.Warn("register cluster failed", zap.Error(err))
		return nil, err
	}

	// Validate kubeconfig availability for workload submission
	var warning string
	if kubeClientProviderConfigured(s.kubeClients) {
		if !s.kubeClients.HasKubeconfig(r.GetClusterId()) {
			warning = fmt.Sprintf("no kubeconfig found for cluster %q in %q - workload submission will fail until kubeconfig is added with matching key name",
				r.GetClusterId(), s.kubeClients.Dir())
			s.log.Warn("cluster registered without kubeconfig",
				zap.String("cluster_id", r.GetClusterId()),
				zap.String("kubeconfigs_dir", s.kubeClients.Dir()),
				zap.String("expected_key", r.GetClusterId()),
			)
		}
	}

	s.store.UpsertClusterFromRegister(r)
	s.log.Info("cluster registered", zap.String("cluster_id", r.GetClusterId()), zap.String("provider", r.GetProvider()), zap.String("region", r.GetRegion()), zap.Int("label_count", len(r.GetLabels())))

	resp := &aegis.ClusterRegisterResponse{Ok: true, Message: "registered"}
	if warning != "" {
		resp.Message = "registered with warning: " + warning
	}
	return resp, nil
}

func (s *Server) Heartbeat(ctx context.Context, hb *aegis.ClusterHeartbeat) (*aegis.ClusterHeartbeatAck, error) {
	if hb == nil {
		err := status.Error(codes.InvalidArgument, "heartbeat payload required")
		s.log.Warn("cluster heartbeat failed", zap.Error(err))
		return nil, err
	}
	flavorNames := make([]string, 0, len(hb.GetAvailableFlavors()))
	for _, f := range hb.GetAvailableFlavors() {
		flavorNames = append(flavorNames, f.GetName())
	}
	s.store.UpdateClusterFromHeartbeat(hb)
	s.log.Debug("cluster heartbeat",
		zap.String("cluster_id", hb.GetClusterId()),
		zap.Float64("ttf_gpu_seconds_p50", hb.GetTtfGpuSecondsP50()),
		zap.Strings("available_flavors", flavorNames),
		zap.String("proxy_url", hb.GetProxyUrl()),
	)
	return &aegis.ClusterHeartbeatAck{Ok: true}, nil
}

func (s *Server) ListClusters(ctx context.Context, req *aegis.ListClustersRequest) (*aegis.ListClustersResponse, error) {
	_, allowed, err := s.authorizedProjects(ctx)
	if err != nil {
		return nil, err
	}
	projectFilter := strings.TrimSpace(req.GetProjectId())
	regionFilter := strings.TrimSpace(req.GetRegion())

	if projectFilter != "" {
		if _, ok := allowed[strings.ToLower(projectFilter)]; !ok && len(allowed) > 0 {
			return nil, status.Error(codes.PermissionDenied, "project not accessible")
		}
	}

	infos := s.store.ListClusterInfos()
	now := time.Now()
	items := make([]*aegis.ClusterSummary, 0, len(infos))
	for _, ci := range infos {
		if ci == nil {
			continue
		}
		projectID := clusterProject(ci)
		if len(allowed) > 0 {
			if _, ok := allowed[strings.ToLower(projectID)]; !ok {
				continue
			}
		}
		if projectFilter != "" && !strings.EqualFold(projectID, projectFilter) {
			continue
		}
		if regionFilter != "" && !strings.EqualFold(ci.Region, regionFilter) {
			continue
		}

		phase := "Ready"
		if !clusterReady(ci, now) {
			if ci.LastHeartbeat.IsZero() {
				phase = "Pending"
			} else {
				phase = "Unhealthy"
			}
		}

		var lastHeartbeat string
		if !ci.LastHeartbeat.IsZero() {
			lastHeartbeat = ci.LastHeartbeat.UTC().Format(time.RFC3339)
		}

		var createdAt string
		if !ci.CreatedAt.IsZero() {
			createdAt = ci.CreatedAt.UTC().Format(time.RFC3339)
		}

		items = append(items, &aegis.ClusterSummary{
			Id:            ci.ID,
			Name:          clusterDisplayName(ci),
			ProjectId:     projectID,
			Provider:      ci.Provider,
			Region:        ci.Region,
			Phase:         phase,
			CreatedAt:     createdAt,
			LastHeartbeat: lastHeartbeat,
		})
	}

	s.log.Debug("list clusters",
		zap.String("project_filter", projectFilter),
		zap.String("region_filter", regionFilter),
		zap.Int("count", len(items)),
	)
	return &aegis.ListClustersResponse{Items: items}, nil
}

func (s *Server) maybeBootstrapWorkspaceDeps(ctx context.Context, w *aegis.Workload) {
	_ = ctx // reserved for future use (tracing, cancellation)
	if !s.autoBootstrap || w == nil {
		return
	}
	projectID := strings.TrimSpace(w.GetProjectId())
	if projectID == "" {
		return
	}
	reqFlavor, err := requiredFlavor(w)
	if err != nil {
		s.log.Debug("autobootstrap skipped; workload missing flavor",
			zap.String("workload_id", w.GetId()),
			zap.Error(err),
		)
		return
	}
	reqFlavor = strings.TrimSpace(reqFlavor)
	if reqFlavor == "" {
		return
	}
	if s.store.GetProject(projectID) == nil {
		s.store.PutProject(&aegis.Project{Id: projectID})
		s.log.Info("autobootstrap: project created",
			zap.String("project_id", projectID),
			zap.String("workload_id", w.GetId()),
		)
	}
	queueName := strings.TrimSpace(w.GetQueue())
	if queueName != "" {
		queue := s.store.GetQueue(queueName)
		if queue == nil {
			queue = &aegis.Queue{
				Name:                      queueName,
				ProjectId:                 projectID,
				DefaultMaxDurationSeconds: s.defaultMaxRuntimeSeconds(nil),
				AllowedFlavors:            []string{reqFlavor},
			}
			s.store.PutQueue(queue)
			s.log.Info("autobootstrap: queue created",
				zap.String("queue", queueName),
				zap.String("project_id", projectID),
				zap.String("workload_id", w.GetId()),
			)
		} else if ensureQueueAllowsFlavor(queue, reqFlavor) {
			s.store.PutQueue(queue)
			s.log.Info("autobootstrap: queue updated",
				zap.String("queue", queueName),
				zap.String("project_id", projectID),
				zap.String("flavor", reqFlavor),
			)
		}
	}
	s.ensureFlavorDefaults(reqFlavor)
}

func (s *Server) SubmitWorkload(ctx context.Context, req *aegis.SubmitWorkloadRequest) (*aegis.Workload, error) {
	if req == nil || req.Workload == nil {
		err := status.Error(codes.InvalidArgument, "workload payload required")
		s.log.Warn("submit workload failed", zap.Error(err))
		return nil, err
	}
	w := req.Workload
	if w.ProjectId == "" {
		return nil, status.Error(codes.InvalidArgument, "projectId required on workload")
	}
	if w.Id == "" {
		w.Id = "w-" + RandID()
	}
	if err := s.authorize(ctx, w.GetProjectId(), w.GetQueue(), "submitWorkload"); err != nil {
		return nil, err
	}

	var (
		projectPolicy placement.ProjectPolicy
		hasPolicy     bool
	)
	if s.policyOverlay != nil {
		projectPolicy, hasPolicy = s.policyOverlay.Effective(w.ProjectId)
	}
	if hasPolicy {
		s.applyDefaultFlavor(projectPolicy, w)
	}

	s.maybeBootstrapWorkspaceDeps(ctx, w)
	p := s.store.GetProject(w.ProjectId)
	if p == nil {
		return nil, status.Errorf(codes.InvalidArgument, "unknown project %q", w.ProjectId)
	}

	reqFlavor, err := requiredFlavor(w)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	flavorObj := s.store.GetFlavor(reqFlavor)
	if flavorObj == nil && s.autoBootstrap {
		s.ensureFlavorDefaults(reqFlavor)
		flavorObj = s.store.GetFlavor(reqFlavor)
	}
	if flavorObj == nil {
		s.log.Warn("submit workload rejected; unknown flavor",
			zap.String("workload_id", w.GetId()),
			zap.String("flavor", reqFlavor),
		)
		return nil, status.Error(codes.FailedPrecondition, "unknown flavor: "+reqFlavor)
	}
	queueObj := s.store.GetQueue(w.GetQueue())
	requestedCluster := strings.TrimSpace(w.GetClusterId())

	if wk, ok := w.GetKind().(*aegis.Workload_Workspace); ok && wk.Workspace != nil {
		s.applyWorkspaceDefaults(wk.Workspace)
	}

	estimateUSD, maxSecs, estErr := s.estimateWorkloadUSD(w, flavorObj, queueObj)
	if estErr != nil {
		s.log.Debug("workload cost estimate fallback",
			zap.String("workload_id", w.GetId()),
			zap.Error(estErr),
		)
	} else {
		s.log.Debug("workload cost estimate",
			zap.String("workload_id", w.GetId()),
			zap.Float64("estimate_usd", estimateUSD),
			zap.Int64("max_runtime_secs", maxSecs),
		)
	}
	mEstimateHist.Observe(estimateUSD)
	allowed, policy, reason, usage := s.store.ReserveIfAllowed(w.GetProjectId(), w.GetQueue(), estimateUSD)
	if !allowed {
		if reason == "" {
			reason = "budget_denied"
		}
		mBudgetDenied.WithLabelValues(w.GetProjectId(), w.GetQueue(), reason).Inc()
		s.log.Warn("budget reservation denied",
			zap.String("workload_id", w.GetId()),
			zap.String("project_id", w.GetProjectId()),
			zap.String("queue", w.GetQueue()),
			zap.Float64("estimate_usd", estimateUSD),
			zap.String("reason", reason),
		)
		return nil, status.Error(codes.FailedPrecondition, "budget exceeded: "+reason)
	}
	if policy != "" {
		if flavorObj.GetPriceUsdPerGpuHour() <= 0 {
			mBudgetDenied.WithLabelValues(w.GetProjectId(), w.GetQueue(), "missing_price").Inc()
			s.log.Warn("budget reservation denied: missing price",
				zap.String("workload_id", w.GetId()),
				zap.String("project_id", w.GetProjectId()),
				zap.String("queue", w.GetQueue()),
				zap.String("flavor", reqFlavor),
			)
			return nil, status.Error(codes.FailedPrecondition, "missing flavor pricing for budgeted project")
		}
		b := s.store.GetBudgetExact(w.GetProjectId(), w.GetQueue())
		if b == nil {
			b = s.store.GetBudgetExact(w.GetProjectId(), "")
		}
		if b != nil && strings.EqualFold(policy, "SOFT") && (usage.ActualUSD+usage.ReservedUSD) > b.GetLimitUsd() {
			mBudgetOverrun.WithLabelValues(w.GetProjectId(), w.GetQueue()).Inc()
			s.log.Warn("budget soft overrun",
				zap.String("workload_id", w.GetId()),
				zap.String("project_id", w.GetProjectId()),
				zap.String("queue", w.GetQueue()),
				zap.Float64("estimate_usd", estimateUSD),
				zap.Float64("limit_usd", b.GetLimitUsd()),
			)
		}
		mBudgetReserved.WithLabelValues(w.GetProjectId(), w.GetQueue()).Set(usage.ReservedUSD)
		mBudgetActual.WithLabelValues(w.GetProjectId(), w.GetQueue()).Set(usage.ActualUSD)
	}
	if estimateUSD > 0 {
		s.store.SetEstimateUSD(w.GetId(), estimateUSD)
	}

	var (
		activeByFlavor map[string]int
		clusterLoads   map[string]int
	)
	if hasPolicy {
		requiresStats := len(projectPolicy.Quotas) > 0 || strings.EqualFold(strings.TrimSpace(projectPolicy.Strategy), "spread")
		if requiresStats {
			activeByFlavor, clusterLoads = s.activeWorkloadStats(w.GetProjectId())
		}
		if len(projectPolicy.Quotas) > 0 {
			flavorKey := strings.ToLower(strings.TrimSpace(reqFlavor))
			if limit, ok := projectPolicy.Quotas[flavorKey]; ok && limit > 0 {
				if activeByFlavor == nil {
					activeByFlavor, clusterLoads = s.activeWorkloadStats(w.GetProjectId())
				}
				if activeByFlavor[strings.ToLower(flavorKey)] >= limit {
					err := status.Error(codes.ResourceExhausted, fmt.Sprintf("quota reached for project %s flavor %s", w.GetProjectId(), reqFlavor))
					s.log.Warn("placement quota reached",
						zap.String("project_id", w.GetProjectId()),
						zap.String("flavor", reqFlavor),
						zap.Int("limit", limit))
					return nil, err
				}
			}
		}
	}

	// Build candidates from current cluster snapshots
	infos := s.store.ListClusterInfos()
	cands := make([]placement.Candidate, 0, len(infos))
	now := time.Now()
	var requestedInfo *store.ClusterInfo
	for _, ci := range infos {
		if strings.EqualFold(strings.TrimSpace(ci.ID), requestedCluster) {
			requestedInfo = ci
		}
		if ci.LastHeartbeat.IsZero() || now.Sub(ci.LastHeartbeat) > heartbeatTTL {
			s.log.Debug("skipping stale cluster",
				zap.String("cluster_id", ci.ID),
				zap.Time("last_heartbeat", ci.LastHeartbeat),
			)
			continue
		}
		cands = append(cands, placement.Candidate{
			ClusterID:   ci.ID,
			Provider:    strings.ToLower(strings.TrimSpace(ci.Provider)),
			Region:      ci.Region,
			Labels:      ci.Labels,
			TTFGSeconds: ci.TTFGSecondsP50,
			Flavors:     ci.AvailableFlavorSet,
		})
	}

	regions := normalizeStrings(projectPolicy.AllowedRegions)
	if len(regions) == 0 && p.GetPolicy() != nil {
		regions = normalizeStrings(p.GetPolicy().GetRegions())
	}
	pd := placement.PolicyDomain{
		Regions:   regions,
		Providers: projectPolicy.AllowedProviders,
		Strategy:  projectPolicy.Strategy,
	}
	placementFlavor := reqFlavor
	if flavorObj.GetGpuCount() == 0 && flavorObj.GetResourceName() == "" {
		placementFlavor = ""
	}
	var chosen string
	if requestedCluster != "" {
		if requestedInfo == nil {
			return nil, status.Errorf(codes.NotFound, "cluster %q not registered", requestedCluster)
		}
		if !clusterReady(requestedInfo, now) {
			return nil, status.Errorf(codes.FailedPrecondition, "cluster %q not ready", requestedCluster)
		}
		pinned := make([]placement.Candidate, 0, 1)
		for _, cand := range cands {
			if strings.EqualFold(cand.ClusterID, requestedCluster) {
				pinned = append(pinned, cand)
				break
			}
		}
		if len(pinned) == 0 {
			return nil, status.Errorf(codes.FailedPrecondition, "cluster %q not eligible", requestedCluster)
		}
		var perr error
		chosen, perr = placement.ChooseCluster(pinned, pd, placementFlavor, clusterLoads)
		if perr != nil {
			s.log.Warn("placement failed for requested cluster",
				zap.String("workload_id", w.GetId()),
				zap.String("project_id", w.GetProjectId()),
				zap.String("flavor", reqFlavor),
				zap.Strings("regions", regions),
				zap.String("requested_cluster", requestedCluster),
				zap.Error(perr),
			)
			return nil, status.Error(codes.FailedPrecondition, fmt.Sprintf("requested cluster %s not eligible: %v", requestedCluster, perr))
		}
	} else {
		var perr error
		chosen, perr = placement.ChooseCluster(cands, pd, placementFlavor, clusterLoads)
		if perr != nil {
			s.log.Warn("placement failed",
				zap.String("workload_id", w.GetId()),
				zap.String("project_id", w.GetProjectId()),
				zap.String("flavor", reqFlavor),
				zap.Strings("regions", regions),
				zap.Int("candidate_count", len(cands)),
				zap.Error(perr),
			)
			return nil, status.Error(codes.FailedPrecondition, fmt.Sprintf("no eligible cluster for flavor=%s in policy regions=%v", reqFlavor, regions))
		}
	}

	w.ClusterId = chosen
	w.Status = statusPlaced

	if fl := s.store.GetFlavor(reqFlavor); fl != nil {
		w.Hints = &aegis.ResourceHints{
			ResourceName:    fl.GetResourceName(),
			GpuCount:        fl.GetGpuCount(),
			CpuCoresRequest: fl.GetCpuCoresRequest(),
			MemoryRequest:   fl.GetMemoryRequest(),
		}
	}

	if !kubeClientProviderConfigured(s.kubeClients) {
		err := status.Error(codes.Internal, "kubernetes client manager not configured")
		s.log.Error("submit workload failed", zap.Error(err))
		return nil, err
	}

	kubeClient, err := s.kubeClients.ClientFor(chosen)
	if err != nil {
		s.log.Error("submit workload failed: kube client", zap.String("cluster_id", chosen), zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to resolve cluster client")
	}

	var workspacePayload *unstructuredapi.Unstructured
	var workloadPayload *aegisv1alpha1.AegisWorkload
	targetNS := s.namespaceForProject(w.GetProjectId())

	if wk, ok := w.GetKind().(*aegis.Workload_Workspace); ok && wk.Workspace != nil {
		workspacePayload = buildWorkspaceCR(w, wk.Workspace, targetNS, reqFlavor, maxSecs)
	} else {
		workloadPayload = buildAegisWorkloadCR(w, targetNS)
		if maxSecs > 0 {
			if workloadPayload.Annotations == nil {
				workloadPayload.Annotations = map[string]string{}
			}
			workloadPayload.Annotations["aegis.yourorg.dev/maxDurationSeconds"] = fmt.Sprintf("%d", maxSecs)
		}
	}

	if workspacePayload != nil {
		if err := kubeClient.Create(ctx, workspacePayload); err != nil {
			if !apierrors.IsAlreadyExists(err) {
				s.log.Error("failed to create workspace CR",
					zap.String("workload_id", w.GetId()),
					zap.String("cluster_id", chosen),
					zap.String("namespace", targetNS),
					zap.Error(err),
				)
				return nil, status.Error(codes.Internal, "failed to create Workspace in target cluster")
			}
			s.log.Info("workspace CR already exists",
				zap.String("workload_id", w.GetId()),
				zap.String("cluster_id", chosen),
				zap.String("namespace", targetNS),
			)
		} else {
			s.log.Info("workspace CR created",
				zap.String("workload_id", w.GetId()),
				zap.String("cluster_id", chosen),
				zap.String("namespace", targetNS),
			)
		}
	} else if workloadPayload != nil {
		if err := kubeClient.Create(ctx, workloadPayload); err != nil {
			if !apierrors.IsAlreadyExists(err) {
				s.log.Error("failed to create aegis workload CR",
					zap.String("workload_id", w.GetId()),
					zap.String("cluster_id", chosen),
					zap.String("namespace", targetNS),
					zap.Error(err),
				)
				return nil, status.Error(codes.Internal, "failed to create AegisWorkload in target cluster")
			}
			s.log.Info("aegis workload CR already exists",
				zap.String("workload_id", w.GetId()),
				zap.String("cluster_id", chosen),
				zap.String("namespace", targetNS),
			)
		} else {
			s.log.Info("aegis workload CR created",
				zap.String("workload_id", w.GetId()),
				zap.String("cluster_id", chosen),
				zap.String("namespace", targetNS),
			)
		}
	}

	s.store.PutWorkload(w)
	s.store.MarkPlaced(w.GetId())
	mPlaced.WithLabelValues(reqFlavor).Inc()
	s.log.Info("workload placed",
		zap.String("workload_id", w.GetId()),
		zap.String("project_id", w.GetProjectId()),
		zap.String("flavor", reqFlavor),
		zap.String("cluster_id", w.GetClusterId()),
		zap.String("status", w.GetStatus()),
		zap.String("namespace", targetNS),
	)
	return w, nil
}

func (s *Server) GetWorkload(ctx context.Context, req *aegis.GetWorkloadRequest) (*aegis.Workload, error) {
	if req == nil || req.Id == "" {
		err := status.Error(codes.InvalidArgument, "workload id required")
		s.log.Warn("get workload failed", zap.Error(err))
		return nil, err
	}
	w := s.store.GetWorkload(req.Id)
	if w == nil {
		err := status.Error(codes.NotFound, "workload not found")
		s.log.Warn("workload not found", zap.String("workload_id", req.GetId()))
		return nil, err
	}
	if err := s.authorize(ctx, w.GetProjectId(), w.GetQueue(), "getWorkload"); err != nil {
		return nil, err
	}
	cache := make(map[string]client.Client)
	s.enrichWorkloadUI(ctx, cache, w)
	s.log.Info("workload retrieved", zap.String("workload_id", w.GetId()), zap.String("status", w.GetStatus()), zap.String("project_id", w.GetProjectId()))
	return w, nil
}

func (s *Server) GetWorkspaceConnectionDetails(ctx context.Context, req *aegis.GetWorkspaceConnectionDetailsRequest) (*aegis.GetWorkspaceConnectionDetailsResponse, error) {
	workloadID := strings.TrimSpace(req.GetId())
	if workloadID == "" {
		return nil, status.Error(codes.InvalidArgument, "workload id required")
	}
	subject := subjectFromContext(ctx)
	if !validSubject(subject) {
		err := status.Error(codes.PermissionDenied, "authenticated subject required")
		s.auditSession("session.create", nil, err, zap.String("workload_id", workloadID), zap.String("client", "legacy"))
		return nil, err
	}

	session, err := s.mintConnectionSession(ctx, workloadID, "legacy", subject)
	if err != nil {
		s.auditSession("session.create", nil, err, zap.String("workload_id", workloadID), zap.String("client", "legacy"), zap.String("subject", subject))
		return nil, err
	}

	s.auditSession("session.create", session, nil)

	return &aegis.GetWorkspaceConnectionDetailsResponse{
		ProxyUrl:     session.ProxyURL,
		Token:        session.Token,
		SshHostAlias: session.SSHHostAlias,
		SshUsername:  session.SSHUser,
		InternalHost: session.InternalHost,
		DestPort:     session.Port,
		ExpiresAtUtc: session.ExpiresAt.UTC().Format(time.RFC3339),
	}, nil
}

func (s *Server) CreateConnectionSession(ctx context.Context, req *aegis.CreateConnectionSessionRequest) (*aegis.ConnectionSession, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	workloadID := strings.TrimSpace(req.GetWorkloadId())
	if workloadID == "" {
		return nil, status.Error(codes.InvalidArgument, "workload id required")
	}
	client, err := normalizeClient(req.GetClient())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	subject := subjectFromContext(ctx)
	if !validSubject(subject) {
		err := status.Error(codes.PermissionDenied, "authenticated subject required")
		s.auditSession("session.create", nil, err, zap.String("workload_id", workloadID), zap.String("client", client))
		return nil, err
	}

	session, err := s.mintConnectionSession(ctx, workloadID, client, subject)
	if err != nil {
		s.auditSession("session.create", nil, err, zap.String("workload_id", workloadID), zap.String("client", client), zap.String("subject", subject))
		return nil, err
	}

	s.auditSession("session.create", session, nil)
	return sessionToProto(session), nil
}

func (s *Server) RenewConnectionSession(ctx context.Context, req *aegis.RenewConnectionSessionRequest) (*aegis.ConnectionSession, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	sessionID := strings.TrimSpace(req.GetSessionId())
	if sessionID == "" {
		return nil, status.Error(codes.InvalidArgument, "session id required")
	}

	existing, ok := s.store.ConnectionSession(sessionID)
	if !ok {
		err := status.Error(codes.NotFound, "session not found")
		s.auditSession("session.renew", nil, err, zap.String("session_id", sessionID))
		return nil, err
	}

	subject := subjectFromContext(ctx)
	if !validSubject(subject) {
		err := status.Error(codes.PermissionDenied, "authenticated subject required")
		s.auditSession("session.renew", existing, err)
		return nil, err
	}
	if !strings.EqualFold(existing.Subject, subject) {
		err := status.Error(codes.PermissionDenied, "session owned by different subject")
		s.auditSession("session.renew", existing, err)
		return nil, err
	}
	if existing.Revoked {
		err := status.Error(codes.FailedPrecondition, "session revoked")
		s.auditSession("session.renew", existing, err)
		return nil, err
	}
	if existing.Used {
		err := status.Error(codes.FailedPrecondition, "session already used")
		s.auditSession("session.renew", existing, err)
		return nil, err
	}

	ctxData, err := s.buildSessionContext(ctx, existing.WorkloadID)
	if err != nil {
		s.auditSession("session.renew", existing, err)
		return nil, err
	}
	if err := s.authorize(ctx, ctxData.workload.GetProjectId(), ctxData.workload.GetQueue(), "renewConnectionSession"); err != nil {
		s.auditSession("session.renew", existing, err)
		return nil, err
	}

	now := time.Now().UTC()
	expiresAt := now.Add(s.connectionTokenTTL())
	jti := fmt.Sprintf("jti-%s", randomHex(8))
	claims := proxyClaims{
		Sub:     subject,
		Wid:     existing.WorkloadID,
		Dest:    ctxData.dest,
		DNS:     ctxData.internalHost,
		Cluster: ctxData.workload.GetClusterId(),
		OneTime: true,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   subject,
			Audience:  jwt.ClaimStrings{s.proxyAudience},
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ID:        jti,
		},
	}

	token, signErr := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.proxySecret)
	if signErr != nil {
		s.auditSession("session.renew", existing, signErr)
		return nil, status.Error(codes.Internal, "failed to sign token")
	}

	sshUser := deriveSSHUser(ctxData.workspace, subject)
	sshConfig := buildSSHConfig(ctxData.alias, ctxData.internalHost, sshUser, ctxData.proxyURL, token, ctxData.port)
	vsURI := buildVSCodeURI(ctxData.alias)

	updated, updErr := s.store.UpdateConnectionSession(existing.SessionID, func(sess *store.ConnectionSession) error {
		sess.JTI = jti
		sess.Token = token
		sess.ExpiresAt = expiresAt
		sess.InternalHost = ctxData.internalHost
		sess.SSHHostAlias = ctxData.alias
		sess.Port = ctxData.port
		sess.ProxyURL = ctxData.proxyURL
		sess.SSHConfig = sshConfig
		sess.VSCodeURI = vsURI
		sess.SSHUser = sshUser
		sess.Used = false
		sess.Revoked = false
		return nil
	})
	if updErr != nil {
		s.auditSession("session.renew", existing, updErr)
		return nil, status.Error(codes.Internal, "failed to update session")
	}

	s.auditSession("session.renew", updated, nil)
	return sessionToProto(updated), nil
}

func (s *Server) RevokeConnectionSession(ctx context.Context, req *aegis.RevokeConnectionSessionRequest) (*emptypb.Empty, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	sessionID := strings.TrimSpace(req.GetSessionId())
	if sessionID == "" {
		return nil, status.Error(codes.InvalidArgument, "session id required")
	}

	existing, ok := s.store.ConnectionSession(sessionID)
	if !ok {
		err := status.Error(codes.NotFound, "session not found")
		s.auditSession("session.revoke", nil, err, zap.String("session_id", sessionID))
		return nil, err
	}

	subject := subjectFromContext(ctx)
	if !validSubject(subject) {
		err := status.Error(codes.PermissionDenied, "authenticated subject required")
		s.auditSession("session.revoke", existing, err)
		return nil, err
	}
	if !strings.EqualFold(existing.Subject, subject) {
		err := status.Error(codes.PermissionDenied, "session owned by different subject")
		s.auditSession("session.revoke", existing, err)
		return nil, err
	}

	now := time.Now().UTC()
	updated, err := s.store.UpdateConnectionSession(sessionID, func(sess *store.ConnectionSession) error {
		sess.Revoked = true
		sess.Used = true
		sess.Token = ""
		sess.ExpiresAt = now
		return nil
	})
	if err != nil {
		s.auditSession("session.revoke", existing, err)
		return nil, status.Error(codes.Internal, "failed to revoke session")
	}

	if existing.JTI != "" {
		s.store.MarkJTIUsed(existing.JTI)
	}

	s.auditSession("session.revoke", updated, nil)
	return &emptypb.Empty{}, nil
}

func (s *Server) enrichWorkloadUI(ctx context.Context, cache map[string]client.Client, w *aegis.Workload) {
	if w == nil {
		return
	}
	fallback := w.GetStatus()
	if fallback == "" {
		fallback = uiStatusSubmitted
	}
	w.UiStatus = fallback
	w.Message = w.GetMessage()

	if !kubeClientProviderConfigured(s.kubeClients) || w.GetClusterId() == "" {
		return
	}

	if cache == nil {
		cache = make(map[string]client.Client)
	}
	cli, ok := cache[w.GetClusterId()]
	if !ok {
		clientForCluster, err := s.kubeClients.ClientFor(w.GetClusterId())
		if err != nil {
			s.log.Debug("skip ui enrichment; kube client unavailable",
				zap.String("workload_id", w.GetId()),
				zap.String("cluster_id", w.GetClusterId()),
				zap.Error(err),
			)
			cache[w.GetClusterId()] = nil
			return
		}
		cache[w.GetClusterId()] = clientForCluster
		cli = clientForCluster
	}
	if cli == nil {
		return
	}

	var jobList batchv1.JobList
	targetNS := s.namespaceForProject(w.GetProjectId())
	if err := cli.List(ctx, &jobList, client.InNamespace(targetNS), client.MatchingLabels{labelWorkloadID: w.GetId()}); err != nil {
		s.log.Debug("skip ui enrichment; listing jobs failed",
			zap.String("workload_id", w.GetId()),
			zap.String("namespace", targetNS),
			zap.Error(err),
		)
		return
	}
	if len(jobList.Items) == 0 {
		return
	}

	uiStatus, message := jobUiStatus(&jobList.Items[0], fallback)
	w.UiStatus = uiStatus
	if message != "" {
		w.Message = message
	}
}

func jobUiStatus(job *batchv1.Job, fallback string) (string, string) {
	if job == nil {
		return fallback, ""
	}

	var message string
	for _, cond := range job.Status.Conditions {
		if cond.Message != "" {
			message = cond.Message
		}
		if cond.Type == batchv1.JobFailed && cond.Message != "" {
			message = cond.Message
		}
		if cond.Type == batchv1.JobSuspended && cond.Message != "" {
			message = cond.Message
		}
	}

	if job.Spec.Suspend != nil && *job.Spec.Suspend {
		if message == "" {
			message = "Job is suspended"
		}
		return uiStatusQueuedByKueue, message
	}
	if job.Status.Succeeded > 0 {
		return "SUCCEEDED", message
	}
	if job.Status.Failed > 0 {
		if message == "" {
			message = "Job has failed"
		}
		return "FAILED", message
	}
	if job.Status.Active > 0 {
		return "RUNNING", message
	}
	return uiStatusSubmitted, message
}

func (s *Server) mintConnectionSession(ctx context.Context, workloadID, client, subject string) (*store.ConnectionSession, error) {
	if workloadID == "" {
		return nil, status.Error(codes.InvalidArgument, "workload id required")
	}
	if len(s.proxySecret) == 0 || s.proxyBaseURL == "" {
		return nil, status.Error(codes.FailedPrecondition, "proxy configuration not available")
	}

	ctxData, err := s.buildSessionContext(ctx, workloadID)
	if err != nil {
		return nil, err
	}
	if err := s.authorize(ctx, ctxData.workload.GetProjectId(), ctxData.workload.GetQueue(), "mintConnectionSession"); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	expiresAt := now.Add(s.connectionTokenTTL())
	if !expiresAt.After(now) {
		expiresAt = now.Add(maxSessionTTL)
	}

	jti := fmt.Sprintf("jti-%s", randomHex(8))
	claims := proxyClaims{
		Sub:     subject,
		Wid:     ctxData.workload.GetId(),
		Dest:    ctxData.dest,
		DNS:     ctxData.internalHost,
		Cluster: ctxData.workload.GetClusterId(),
		OneTime: true,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   subject,
			Audience:  jwt.ClaimStrings{s.proxyAudience},
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ID:        jti,
		},
	}

	token, signErr := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.proxySecret)
	if signErr != nil {
		return nil, status.Error(codes.Internal, "failed to sign token")
	}

	sshUser := deriveSSHUser(ctxData.workspace, subject)
	sshConfig := buildSSHConfig(ctxData.alias, ctxData.internalHost, sshUser, ctxData.proxyURL, token, ctxData.port)
	sessionID := fmt.Sprintf("sess-%s", randomHex(8))

	record := &store.ConnectionSession{
		SessionID:    sessionID,
		WorkloadID:   ctxData.workload.GetId(),
		Subject:      subject,
		Client:       client,
		JTI:          jti,
		Token:        token,
		SSHUser:      sshUser,
		SSHHostAlias: ctxData.alias,
		InternalHost: ctxData.internalHost,
		Port:         ctxData.port,
		SSHConfig:    sshConfig,
		ProxyURL:     ctxData.proxyURL,
		VSCodeURI:    buildVSCodeURI(ctxData.alias),
		ExpiresAt:    expiresAt,
		OneTime:      true,
		Used:         false,
		Revoked:      false,
	}

	s.store.PurgeExpiredSessions(now)
	stored := s.store.PutConnectionSession(record)
	return stored, nil
}

func (s *Server) buildSessionContext(ctx context.Context, workloadID string) (*sessionContext, error) {
	w := s.store.GetWorkload(workloadID)
	if w == nil {
		return nil, status.Error(codes.NotFound, "workload not found")
	}

	wk, ok := w.GetKind().(*aegis.Workload_Workspace)
	if !ok || wk.Workspace == nil {
		return nil, status.Error(codes.FailedPrecondition, "workload is not a workspace")
	}
	if !wk.Workspace.GetInteractive() {
		return nil, status.Error(codes.FailedPrecondition, "workspace is not interactive")
	}
	if w.GetStatus() != statusRunning {
		return nil, status.Errorf(codes.FailedPrecondition, "workspace status %s is not running", w.GetStatus())
	}
	if w.GetClusterId() == "" {
		return nil, status.Error(codes.FailedPrecondition, "workspace cluster not assigned yet")
	}

	port := selectWorkspacePort(wk.Workspace)
	alias := buildHostAlias(w.GetId())
	targetNS := s.namespaceForProject(w.GetProjectId())
	internalHost := fmt.Sprintf("%s.%s%s", alias, targetNS, svcClusterDomainSuffix)
	dest := fmt.Sprintf("%s:%d", internalHost, port)

	// Determine proxy URL: use spoke proxy if cluster reports one, otherwise use hub proxy
	proxyBaseURL := s.proxyBaseURL
	clusterID := w.GetClusterId()
	if clusterID != "" {
		if clusterInfo := s.store.GetClusterInfo(clusterID); clusterInfo != nil {
			if clusterInfo.ProxyURL != "" {
				proxyBaseURL = clusterInfo.ProxyURL
				s.log.Debug("using spoke proxy for cluster", zap.String("cluster_id", clusterID), zap.String("proxy_url", proxyBaseURL))
			} else {
				// Cluster is registered but has no proxy_url - this likely means spoke-proxy
				// is not deployed on the remote cluster. Return a clear error instead of
				// silently falling back to the hub proxy which won't work for remote clusters.
				s.log.Warn("cluster has no proxy_url configured; spoke-proxy may not be deployed",
					zap.String("cluster_id", clusterID),
					zap.String("workload_id", workloadID),
					zap.String("fallback_proxy", proxyBaseURL),
				)
				// Only error if the cluster appears to be remote (not a local dev cluster)
				// Local clusters typically don't register or use in-cluster kubeconfig
				if clusterInfo.Provider != "" && clusterInfo.Provider != "local" {
					return nil, status.Errorf(codes.FailedPrecondition,
						"cluster %s has no proxy URL configured; the spoke-proxy may not be deployed. "+
							"Deploy the spoke chart with proxy.enabled=true or set AEGIS_PROXY_INGRESS_HOST on the k8s-agent",
						clusterID)
				}
			}
		}
	}
	proxyURL := fmt.Sprintf("%s/proxy/%s", proxyBaseURL, w.GetId())

	return &sessionContext{
		workload:     w,
		workspace:    wk.Workspace,
		port:         port,
		alias:        alias,
		internalHost: internalHost,
		dest:         dest,
		proxyURL:     proxyURL,
	}, nil
}

func (s *Server) connectionTokenTTL() time.Duration {
	ttl := s.proxyTokenTTL
	if ttl <= 0 {
		ttl = time.Duration(defaultProxyTokenTTLSeconds) * time.Second
	}
	if ttl > maxSessionTTL {
		ttl = maxSessionTTL
	}
	return ttl
}

func sessionToProto(session *store.ConnectionSession) *aegis.ConnectionSession {
	if session == nil {
		return nil
	}
	return &aegis.ConnectionSession{
		SessionId:    session.SessionID,
		Token:        session.Token,
		SshUser:      session.SSHUser,
		SshHostAlias: session.SSHHostAlias,
		VscodeUri:    session.VSCodeURI,
		SshConfig:    session.SSHConfig,
		ProxyUrl:     session.ProxyURL,
		ExpiresAtUtc: session.ExpiresAt.UTC().Format(time.RFC3339),
		OneTime:      session.OneTime,
	}
}

func (s *Server) auditSession(action string, session *store.ConnectionSession, err error, extra ...zap.Field) {
	fields := []zap.Field{zap.String("action", action)}
	if session != nil {
		fields = append(fields,
			zap.String("session_id", session.SessionID),
			zap.String("workload_id", session.WorkloadID),
			zap.String("subject", session.Subject),
			zap.String("client", session.Client),
			zap.Bool("one_time", session.OneTime),
			zap.Bool("used", session.Used),
			zap.Bool("revoked", session.Revoked),
			zap.Time("expires_at", session.ExpiresAt.UTC()),
			zap.String("jti", session.JTI),
		)
	}
	fields = append(fields, extra...)
	if err != nil {
		fields = append(fields, zap.Error(err))
		s.log.Warn("connection session event", fields...)
		return
	}
	s.log.Info("connection session event", fields...)
}

func normalizeClient(raw string) (string, error) {
	val := strings.TrimSpace(strings.ToLower(raw))
	if val == "" {
		return "cli", nil
	}
	switch val {
	case "vscode", "ssh", "cli":
		return val, nil
	default:
		return "", fmt.Errorf("unsupported client %q", raw)
	}
}

func validSubject(sub string) bool {
	cleaned := strings.TrimSpace(sub)
	if cleaned == "" {
		return false
	}
	return !strings.EqualFold(cleaned, "unknown")
}

func randomHex(n int) string {
	if n <= 0 {
		return RandID()
	}
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return RandID()
	}
	return hex.EncodeToString(buf)
}

func deriveSSHUser(workspace *aegis.WorkspaceSpec, subject string) string {
	if workspace != nil {
		if candidate, ok := workspace.Env["AEGIS_SSH_USER"]; ok {
			if sanitized := sanitizeSSHUser(candidate); sanitized != "" {
				return sanitized
			}
		}
		if candidate, ok := workspace.Env["USER_NAME"]; ok {
			if sanitized := sanitizeSSHUser(candidate); sanitized != "" {
				return sanitized
			}
		}
	}
	return buildSSHUser(subject)
}

func sanitizeSSHUser(raw string) string {
	cleaned := strings.TrimSpace(strings.ToLower(raw))
	if cleaned == "" {
		return ""
	}
	var b strings.Builder
	for _, r := range cleaned {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == '-' || r == '_':
			if b.Len() == 0 {
				continue
			}
			b.WriteRune(r)
		}
	}
	user := b.String()
	if user == "" || strings.HasPrefix(user, "-") || user == "root" {
		return ""
	}
	const maxSSHUserLen = 32
	if len(user) > maxSSHUserLen {
		user = user[:maxSSHUserLen]
	}
	return user
}

func buildSSHUser(subject string) string {
	cleaned := strings.TrimSpace(strings.ToLower(subject))
	if cleaned == "" || cleaned == "unknown" {
		return fmt.Sprintf("aegis-%s", randomHex(4))
	}
	sum := sha256.Sum256([]byte(cleaned))
	return fmt.Sprintf("aegis-%x", sum[:4])
}

func buildHostAlias(workloadID string) string {
	const prefix = "aegis-w-"
	cleaned := strings.TrimSpace(strings.ToLower(workloadID))
	if cleaned == "" {
		return prefix + randomHex(4)
	}
	var b strings.Builder
	b.Grow(len(prefix) + len(cleaned))
	b.WriteString(prefix)
	lastHyphen := false
	for _, r := range cleaned {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			lastHyphen = false
		default:
			if !lastHyphen {
				b.WriteByte('-')
				lastHyphen = true
			}
		}
	}
	alias := b.String()
	alias = strings.TrimRight(alias, "-")
	if len(alias) <= len(prefix) {
		alias = prefix + randomHex(4)
	}
	if len(alias) > 63 {
		alias = alias[:63]
	}
	return alias
}

func buildVSCodeURI(alias string) string {
	// Use aegis.aegis-remote protocol instead of ssh-remote
	// Extract workload ID from alias (format: aegis-w-w-{id})
	workloadID := alias
	if len(alias) > 8 && alias[:8] == "aegis-w-" {
		workloadID = alias[8:] // Remove "aegis-w-" prefix to get "w-{id}"
	}
	return fmt.Sprintf("vscode://aegis.aegis-remote/aegis+%s", workloadID)
}

func buildSSHConfig(alias, internalHost, sshUser, proxyURL, token string, port int32) string {
	if port <= 0 {
		port = 22
	}
	return fmt.Sprintf("Host %s\n  HostName %s\n  User %s\n  Port %d\n  ProxyCommand aegis-connect --proxy=%s --token=%s\n  IdentitiesOnly yes\n", alias, internalHost, sshUser, port, proxyURL, token)
}

func (s *Server) ListWorkloads(ctx context.Context, req *aegis.ListWorkloadsRequest) (*aegis.ListWorkloadsResponse, error) {
	if req == nil || req.ProjectId == "" {
		err := status.Error(codes.InvalidArgument, "project id required")
		s.log.Warn("list workloads failed", zap.Error(err))
		return nil, err
	}
	if err := s.authorize(ctx, req.GetProjectId(), "", "listWorkloads"); err != nil {
		return nil, err
	}
	items := s.store.ListWorkloads(req.ProjectId)
	cache := make(map[string]client.Client)
	for _, w := range items {
		s.enrichWorkloadUI(ctx, cache, w)
	}
	s.log.Info("workloads listed", zap.String("project_id", req.GetProjectId()), zap.Int("count", len(items)))
	return &aegis.ListWorkloadsResponse{Items: items}, nil
}

func (s *Server) LeaseWorkload(ctx context.Context, req *aegis.LeaseWorkloadRequest) (*aegis.LeaseWorkloadResponse, error) {
	if req == nil || req.GetClusterId() == "" {
		err := status.Error(codes.InvalidArgument, "cluster_id required")
		s.log.Warn("lease workload failed", zap.Error(err))
		return nil, err
	}
	if req.GetMax() <= 0 {
		err := status.Error(codes.InvalidArgument, "max must be > 0")
		s.log.Warn("lease workload failed", zap.Error(err), zap.Int32("max", req.GetMax()))
		return nil, err
	}
	items := s.store.LeaseWorkloads(req.GetClusterId(), int(req.GetMax()))
	now := time.Now()
	for _, w := range items {
		queue := w.GetQueue()
		if queue == "" {
			queue = "unknown"
		}
		flavor, ferr := requiredFlavor(w)
		if ferr != nil {
			s.log.Warn("leased workload missing flavor",
				zap.String("workload_id", w.GetId()),
				zap.Error(ferr),
			)
			continue
		}
		if t0, ok := s.store.GetPlacedAt(w.GetId()); ok {
			wait := now.Sub(t0)
			mQueueWait.WithLabelValues(queue, flavor).Observe(wait.Seconds())
			s.store.ClearPlacedAt(w.GetId())
			s.log.Debug("queue wait observed",
				zap.String("workload_id", w.GetId()),
				zap.String("queue", queue),
				zap.String("flavor", flavor),
				zap.Duration("wait", wait),
			)
		}
		if fl := s.store.GetFlavor(flavor); fl != nil {
			w.Hints = &aegis.ResourceHints{
				ResourceName:    fl.GetResourceName(),
				GpuCount:        fl.GetGpuCount(),
				CpuCoresRequest: fl.GetCpuCoresRequest(),
				MemoryRequest:   fl.GetMemoryRequest(),
			}
			s.log.Debug("resource hints stamped",
				zap.String("workload_id", w.GetId()),
				zap.String("queue", queue),
				zap.String("flavor", flavor),
				zap.String("resource_name", fl.GetResourceName()),
				zap.Int("gpu_count", int(fl.GetGpuCount())),
				zap.String("cpu_request", fl.GetCpuCoresRequest()),
				zap.String("memory_request", fl.GetMemoryRequest()),
			)
		} else {
			s.log.Debug("flavor missing in catalog; hints skipped",
				zap.String("workload_id", w.GetId()),
				zap.String("flavor", flavor),
			)
		}
	}
	if len(items) > 0 {
		ids := make([]string, 0, len(items))
		for _, w := range items {
			ids = append(ids, w.GetId())
		}
		s.log.Info("workloads leased",
			zap.String("cluster_id", req.GetClusterId()),
			zap.Strings("workload_ids", ids),
			zap.Int("count", len(items)),
		)
		mLeased.Add(float64(len(items)))
	} else {
		s.log.Debug("no workloads leased", zap.String("cluster_id", req.GetClusterId()))
	}
	return &aegis.LeaseWorkloadResponse{Items: items}, nil
}

func (s *Server) StartWorkload(ctx context.Context, req *aegis.StartWorkloadRequest) (*aegis.StartWorkloadResponse, error) {
	if req == nil || req.GetId() == "" {
		err := status.Error(codes.InvalidArgument, "workload id required")
		s.log.Warn("start workload failed", zap.Error(err))
		return nil, err
	}
	if req.GetClusterId() == "" {
		err := status.Error(codes.InvalidArgument, "cluster_id required")
		s.log.Warn("start workload failed", zap.Error(err), zap.String("workload_id", req.GetId()))
		return nil, err
	}

	w := s.store.GetWorkload(req.GetId())
	if w == nil {
		err := status.Error(codes.NotFound, "workload not found")
		s.log.Warn("start workload failed", zap.Error(err), zap.String("workload_id", req.GetId()))
		return nil, err
	}

	if w.GetClusterId() != req.GetClusterId() {
		err := status.Errorf(codes.FailedPrecondition, "workload assigned to cluster %s", w.GetClusterId())
		s.log.Warn("start workload cluster mismatch",
			zap.Error(err),
			zap.String("workload_id", w.GetId()),
			zap.String("expected_cluster", w.GetClusterId()),
			zap.String("requested_cluster", req.GetClusterId()),
		)
		return nil, err
	}

	updated, wait, observed, err := s.store.StartWorkload(req.GetId())
	if err != nil {
		errStatus := status.Error(codes.FailedPrecondition, err.Error())
		s.log.Warn("start workload state transition failed",
			zap.Error(err),
			zap.String("workload_id", req.GetId()),
		)
		return nil, errStatus
	}

	queue := updated.GetQueue()
	if queue == "" {
		queue = "unknown"
	}
	if observed {
		if flavor, ferr := requiredFlavor(updated); ferr == nil {
			mQueueWait.WithLabelValues(queue, flavor).Observe(wait.Seconds())
			s.log.Debug("queue wait observed",
				zap.String("workload_id", updated.GetId()),
				zap.String("queue", queue),
				zap.String("flavor", flavor),
				zap.Duration("wait", wait),
			)
		} else {
			s.log.Debug("start workload missing flavor for metrics",
				zap.String("workload_id", updated.GetId()),
				zap.Error(ferr),
			)
		}
	}

	idempotent := !observed && wait == 0
	s.log.Info("workload start acknowledged",
		zap.String("workload_id", updated.GetId()),
		zap.String("cluster_id", updated.GetClusterId()),
		zap.Bool("idempotent", idempotent),
		zap.Duration("queue_wait", wait),
	)

	return &aegis.StartWorkloadResponse{Workload: updated}, nil
}

func (s *Server) AckWorkload(ctx context.Context, req *aegis.AckWorkloadRequest) (*aegis.AckWorkloadResponse, error) {
	if req == nil || req.GetId() == "" {
		err := status.Error(codes.InvalidArgument, "workload id required")
		s.log.Warn("ack workload failed", zap.Error(err))
		return nil, err
	}
	if req.GetStatus() == "" {
		err := status.Error(codes.InvalidArgument, "status required")
		s.log.Warn("ack workload failed", zap.Error(err), zap.String("workload_id", req.GetId()))
		return nil, err
	}
	w := s.store.GetWorkload(req.GetId())
	if w == nil {
		err := status.Error(codes.NotFound, "workload not found")
		s.log.Warn("ack workload failed", zap.Error(err), zap.String("workload_id", req.GetId()))
		return nil, err
	}
	if w.GetStatus() != statusRunning {
		err := status.Errorf(codes.FailedPrecondition, "workload not in %s state", statusRunning)
		s.log.Warn("ack workload failed",
			zap.Error(err),
			zap.String("workload_id", w.GetId()),
			zap.String("current_status", w.GetStatus()),
		)
		return nil, err
	}
	updated, err := s.store.AckWorkload(req.GetId(), req.GetStatus(), req.GetUrl())
	if err != nil {
		s.log.Error("ack workload store update failed",
			zap.Error(err),
			zap.String("workload_id", req.GetId()),
		)
		return nil, status.Error(codes.Internal, "failed to update workload status")
	}
	backend := req.GetBackend()
	if backend == "" {
		switch w.GetKind().(type) {
		case *aegis.Workload_Workspace:
			backend = "workspace"
		case *aegis.Workload_Training:
			backend = "trainer_v2"
		default:
			backend = "unknown"
		}
	}
	mAcked.WithLabelValues(req.GetStatus(), backend).Inc()
	s.log.Info("workload acknowledged",
		zap.String("workload_id", updated.GetId()),
		zap.String("cluster_id", updated.GetClusterId()),
		zap.String("status", updated.GetStatus()),
		zap.String("backend", backend),
		zap.String("url", updated.GetUrl()),
	)
	if b, _ := s.store.ResolveBudget(updated.GetProjectId(), updated.GetQueue()); b != nil {
		reqFlavor, _ := requiredFlavor(updated)
		fl := s.store.GetFlavor(reqFlavor)
		runtimeSecs := int64(0)
		if t0, ok := s.store.GetStartedAt(updated.GetId()); ok {
			runtimeSecs = int64(time.Since(t0).Seconds())
			if runtimeSecs < 0 {
				runtimeSecs = 0
			}
		}
		actualUSD := s.runtimeCostUSD(updated, fl, runtimeSecs)
		estUSD := s.store.PopEstimateUSD(updated.GetId())
		usage := s.store.ReconcileOnAck(updated.GetProjectId(), updated.GetQueue(), estUSD, actualUSD)
		mBudgetReserved.WithLabelValues(updated.GetProjectId(), updated.GetQueue()).Set(usage.ReservedUSD)
		mBudgetActual.WithLabelValues(updated.GetProjectId(), updated.GetQueue()).Set(usage.ActualUSD)
		s.log.Info("budget reconciled",
			zap.String("workload_id", updated.GetId()),
			zap.String("project_id", updated.GetProjectId()),
			zap.String("queue", updated.GetQueue()),
			zap.Float64("released_estimate_usd", estUSD),
			zap.Float64("actual_usd", actualUSD),
			zap.Float64("usage_reserved_usd", usage.ReservedUSD),
			zap.Float64("usage_actual_usd", usage.ActualUSD),
			zap.Int64("runtime_secs", runtimeSecs),
		)
	}
	return &aegis.AckWorkloadResponse{Workload: updated}, nil
}

func (s *Server) estimateWorkloadUSD(w *aegis.Workload, fl *aegis.Flavor, q *aegis.Queue) (float64, int64, error) {
	if fl == nil {
		return 0, 0, fmt.Errorf("missing flavor")
	}
	price := fl.GetPriceUsdPerGpuHour()
	var totalGPUs int32
	var maxSecs int64
	switch wk := w.GetKind().(type) {
	case *aegis.Workload_Workspace:
		totalGPUs = fl.GetGpuCount()
		if wk.Workspace.GetMaxDurationSeconds() > 0 {
			maxSecs = wk.Workspace.GetMaxDurationSeconds()
		}
	case *aegis.Workload_Training:
		workers := wk.Training.GetWorkers()
		if workers <= 0 {
			workers = 1
		}
		gpusPer := wk.Training.GetGpusPerWorker()
		if gpusPer <= 0 {
			gpusPer = fl.GetGpuCount()
		}
		totalGPUs = workers * gpusPer
		if wk.Training.GetMaxDurationSeconds() > 0 {
			maxSecs = wk.Training.GetMaxDurationSeconds()
		}
	default:
	}
	if maxSecs <= 0 {
		maxSecs = s.defaultMaxRuntimeSeconds(q)
	}
	usd := float64(totalGPUs) * price * (float64(maxSecs) / 3600.0)
	return usd, maxSecs, nil
}

func (s *Server) runtimeCostUSD(w *aegis.Workload, fl *aegis.Flavor, runtimeSecs int64) float64 {
	if fl == nil || runtimeSecs <= 0 {
		return 0
	}
	price := fl.GetPriceUsdPerGpuHour()
	var totalGPUs int32
	switch wk := w.GetKind().(type) {
	case *aegis.Workload_Workspace:
		totalGPUs = fl.GetGpuCount()
	case *aegis.Workload_Training:
		workers := wk.Training.GetWorkers()
		if workers <= 0 {
			workers = 1
		}
		gpusPer := wk.Training.GetGpusPerWorker()
		if gpusPer <= 0 {
			gpusPer = fl.GetGpuCount()
		}
		totalGPUs = workers * gpusPer
	default:
	}
	return float64(totalGPUs) * price * (float64(runtimeSecs) / 3600.0)
}

func (s *Server) GetBudget(ctx context.Context, req *aegis.GetBudgetRequest) (*aegis.GetBudgetResponse, error) {
	if req == nil || req.GetProjectId() == "" {
		return nil, status.Error(codes.InvalidArgument, "project_id required")
	}
	if err := s.authorize(ctx, req.GetProjectId(), req.GetQueue(), "getBudget"); err != nil {
		return nil, err
	}
	b := s.store.GetBudgetExact(req.GetProjectId(), req.GetQueue())
	if b == nil {
		b = s.store.GetBudgetExact(req.GetProjectId(), "")
	}
	if b == nil {
		return nil, status.Error(codes.NotFound, "budget not found")
	}
	view, _ := s.store.UsageView(b.GetProjectId(), b.GetQueue())
	resp := &aegis.GetBudgetResponse{Budget: b, Usage: &aegis.BudgetUsage{
		ActualUsd:      view.ActualUSD,
		ReservedUsd:    view.ReservedUSD,
		RemainingUsd:   b.GetLimitUsd() - (view.ActualUSD + view.ReservedUSD),
		PeriodStartUtc: view.PeriodStart.Format("2006-01-02"),
		PeriodEndUtc:   view.PeriodEnd.Format("2006-01-02"),
	}}
	return resp, nil
}

func (s *Server) ListBudgets(ctx context.Context, req *aegis.ListBudgetsRequest) (*aegis.ListBudgetsResponse, error) {
	if req == nil || req.GetProjectId() == "" {
		return nil, status.Error(codes.InvalidArgument, "project_id required")
	}
	if err := s.authorize(ctx, req.GetProjectId(), "", "listBudgets"); err != nil {
		return nil, err
	}
	items := []*aegis.BudgetWithUsage{}
	for _, b := range s.store.ListBudgets(req.GetProjectId()) {
		view, _ := s.store.UsageView(b.GetProjectId(), b.GetQueue())
		items = append(items, &aegis.BudgetWithUsage{
			Budget: b,
			Usage: &aegis.BudgetUsage{
				ActualUsd:      view.ActualUSD,
				ReservedUsd:    view.ReservedUSD,
				RemainingUsd:   b.GetLimitUsd() - (view.ActualUSD + view.ReservedUSD),
				PeriodStartUtc: view.PeriodStart.Format("2006-01-02"),
				PeriodEndUtc:   view.PeriodEnd.Format("2006-01-02"),
			},
		})
	}
	return &aegis.ListBudgetsResponse{Items: items}, nil
}

func Run(ctx context.Context, log *zap.Logger, addrGRPC, addrHTTP string, svc *Server) error {
	lis, err := net.Listen("tcp", addrGRPC)
	if err != nil {
		return err
	}
	authCfg, err := config.LoadAuthConfig()
	if err != nil {
		return fmt.Errorf("failed to load auth config: %w", err)
	}
	authenticator, err := mw.NewAuthenticator(authCfg, log)
	if err != nil {
		return fmt.Errorf("failed to initialise authenticator: %w", err)
	}

	opts, err := grpcServerOptionsFromEnv(log)
	if err != nil {
		return err
	}
	opts = append(opts,
		grpc.ChainUnaryInterceptor(authenticator.UnaryServerInterceptor()),
		grpc.ChainStreamInterceptor(authenticator.StreamServerInterceptor()),
	)
	gs := grpc.NewServer(opts...)
	reflection.Register(gs)
	aegis.RegisterAegisPlatformServer(gs, svc)

	log.Info("gRPC server listening", zap.String("addr", addrGRPC))

	// HTTP gateway mux with REST handlers + Prometheus metrics.
	mux := runtime.NewServeMux()
	// Register REST handlers for our in-process service (no network dials).
	if err := aegis.RegisterAegisPlatformHandlerServer(ctx, mux, svc); err != nil {
		log.Error("failed to register grpc-gateway handlers", zap.Error(err))
	}
	registerWorkspaceWizardRoutes(mux, svc)
	registerObservabilityRoutes(mux, svc)
	registerProvisioningRoutes(mux, svc)
	root := http.NewServeMux()
	root.Handle("/healthz", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "OK")
	}))
	root.Handle("/", authenticator.HTTPMiddleware(mux))
	root.Handle("/metrics", promhttp.Handler())
	httpSrv := &http.Server{Addr: addrHTTP, Handler: root}

	go func() {
		log.Info("HTTP gateway listening", zap.String("addr", addrHTTP))
		if err := httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Warn("HTTP gateway stopped", zap.Error(err))
		}
	}()

	// Background job for provisioning log retention cleanup
	// Configurable via AEGIS_LOG_RETENTION_DAYS (default: 7 days)
	go func() {
		retentionDays := int(getEnvInt("AEGIS_LOG_RETENTION_DAYS", 7))
		if retentionDays <= 0 {
			log.Info("provisioning log retention disabled (AEGIS_LOG_RETENTION_DAYS <= 0)")
			return
		}
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		log.Info("provisioning log retention cleanup started",
			zap.Int("retention_days", retentionDays),
			zap.String("interval", "1h"))
		for {
			select {
			case <-ctx.Done():
				log.Info("provisioning log retention cleanup stopped")
				return
			case <-ticker.C:
				cutoff := time.Now().UTC().AddDate(0, 0, -retentionDays)
				deleted := svc.store.DeleteOldProvisioningLogs(cutoff)
				if deleted > 0 {
					log.Info("provisioning log retention cleanup completed",
						zap.Int64("deleted_entries", deleted),
						zap.Time("cutoff", cutoff))
				}
			}
		}
	}()

	if err := gs.Serve(lis); err != nil {
		if errors.Is(err, grpc.ErrServerStopped) {
			log.Info("gRPC server stopped")
			return nil
		}
		return err
	}
	return nil
}

func grpcServerOptionsFromEnv(log *zap.Logger) ([]grpc.ServerOption, error) {
	certPath := os.Getenv("AEGIS_GRPC_TLS_CERT")
	keyPath := os.Getenv("AEGIS_GRPC_TLS_KEY")
	if certPath == "" && keyPath == "" {
		return nil, nil
	}
	if certPath == "" || keyPath == "" {
		return nil, fmt.Errorf("both AEGIS_GRPC_TLS_CERT and AEGIS_GRPC_TLS_KEY must be set")
	}
	certificate, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load gRPC TLS certificate: %w", err)
	}
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{certificate},
	}
	if caPath := os.Getenv("AEGIS_GRPC_TLS_CLIENT_CA"); caPath != "" {
		caPEM, err := os.ReadFile(caPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read AEGIS_GRPC_TLS_CLIENT_CA: %w", err)
		}
		pool := x509.NewCertPool()
		if !pool.AppendCertsFromPEM(caPEM) {
			return nil, fmt.Errorf("failed to parse certificates from AEGIS_GRPC_TLS_CLIENT_CA")
		}
		tlsConfig.ClientCAs = pool
		tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
		log.Info("gRPC TLS client authentication enabled")
	} else {
		log.Info("gRPC TLS enabled")
	}
	return []grpc.ServerOption{grpc.Creds(credentials.NewTLS(tlsConfig))}, nil
}

func buildAegisWorkloadCR(w *aegis.Workload, namespace string) *aegisv1alpha1.AegisWorkload {
	spec := aegisv1alpha1.AegisWorkloadSpec{
		ProjectID: w.GetProjectId(),
		Queue:     w.GetQueue(),
	}

	switch wk := w.GetKind().(type) {
	case *aegis.Workload_Workspace:
		ws := wk.Workspace
		if ws != nil {
			spec.Workspace = &aegisv1alpha1.WorkspaceSpec{
				Flavor:      ws.GetFlavor(),
				Image:       ws.GetImage(),
				Env:         cloneStringMap(ws.GetEnv()),
				Command:     cloneStringSlice(ws.GetCommand()),
				Interactive: ws.GetInteractive(),
				Ports:       cloneInt32Slice(ws.GetPorts()),
			}
		}
	case *aegis.Workload_Training:
		tr := wk.Training
		if tr != nil {
			spec.Training = &aegisv1alpha1.TrainingSpec{
				Flavor:        tr.GetFlavor(),
				Workers:       tr.GetWorkers(),
				GpusPerWorker: tr.GetGpusPerWorker(),
				Image:         tr.GetImage(),
				Command:       cloneStringSlice(tr.GetCommand()),
				Gang:          tr.GetGang(),
			}
		}
	}

	if hints := w.GetHints(); hints != nil {
		spec.Hints = &aegisv1alpha1.ResourceHints{
			ResourceName:    hints.GetResourceName(),
			GpuCount:        hints.GetGpuCount(),
			CpuCoresRequest: stringPtr(hints.GetCpuCoresRequest()),
			MemoryRequest:   stringPtr(hints.GetMemoryRequest()),
		}
	}

	return &aegisv1alpha1.AegisWorkload{
		TypeMeta: metav1.TypeMeta{
			APIVersion: aegisv1alpha1.GroupVersion.String(),
			Kind:       "AegisWorkload",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      w.GetId(),
			Namespace: namespace,
		},
		Spec: spec,
	}
}

func buildWorkspaceCR(w *aegis.Workload, ws *aegis.WorkspaceSpec, namespace, flavor string, maxRuntime int64) *unstructuredapi.Unstructured {
	if ws == nil {
		return nil
	}

	metadata := map[string]interface{}{
		"name":      w.GetId(),
		"namespace": namespace,
		"labels": map[string]interface{}{
			labelWorkloadID:                w.GetId(),
			"aegis.aegis-remote/workspace": "true",
			"aegisRemote":                  "true",
		},
	}

	execution := map[string]interface{}{}
	if img := strings.TrimSpace(ws.GetImage()); img != "" {
		execution["image"] = img
	}
	if env := ws.GetEnv(); len(env) > 0 {
		envMap := make(map[string]interface{}, len(env))
		for k, v := range env {
			if strings.TrimSpace(k) == "" {
				continue
			}
			envMap[k] = v
		}
		if len(envMap) > 0 {
			execution["env"] = envMap
		}
	}
	if cmd := ws.GetCommand(); len(cmd) > 0 {
		execution["command"] = cmd
	}
	if ports := ws.GetPorts(); len(ports) > 0 {
		portsOut := make([]interface{}, 0, len(ports))
		for _, port := range ports {
			portsOut = append(portsOut, port)
		}
		execution["ports"] = portsOut
	}
	if ws.GetInteractive() {
		execution["interactive"] = true
	}

	spec := map[string]interface{}{
		"projectRef": strings.TrimSpace(w.GetProjectId()),
		"profileRef": "custom",
	}
	if queue := strings.TrimSpace(w.GetQueue()); queue != "" {
		spec["queue"] = queue
	}
	if len(execution) > 0 {
		spec["execution"] = execution
	}

	gpuFlavor := strings.TrimSpace(ws.GetFlavor())
	if gpuFlavor == "" {
		gpuFlavor = strings.TrimSpace(flavor)
	}
	hints := w.GetHints()
	if gpuFlavor != "" || hints != nil {
		gpuProfile := map[string]interface{}{}
		if gpuFlavor != "" {
			gpuProfile["flavor"] = gpuFlavor
		}
		if hints != nil {
			hintsMap := map[string]interface{}{}
			if resourceName := strings.TrimSpace(hints.GetResourceName()); resourceName != "" {
				hintsMap["resourceName"] = resourceName
			}
			if gpuCount := hints.GetGpuCount(); gpuCount > 0 {
				hintsMap["gpuCount"] = gpuCount
			}
			if cpu := strings.TrimSpace(hints.GetCpuCoresRequest()); cpu != "" {
				hintsMap["cpuCoresRequest"] = cpu
			}
			if memory := strings.TrimSpace(hints.GetMemoryRequest()); memory != "" {
				hintsMap["memoryRequest"] = memory
			}
			if len(hintsMap) > 0 {
				gpuProfile["hints"] = hintsMap
			}
		}
		if len(gpuProfile) > 0 {
			spec["gpuProfile"] = gpuProfile
		}
	}

	maxDuration := ws.GetMaxDurationSeconds()
	if maxDuration <= 0 {
		maxDuration = maxRuntime
	}
	if maxDuration > 0 {
		spec["maxDurationSeconds"] = maxDuration
	}

	return &unstructuredapi.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "aegis.yourorg.dev/v1alpha2",
			"kind":       "Workspace",
			"metadata":   metadata,
			"spec":       spec,
		},
	}
}

func cloneStringSlice(in []string) []string {
	if len(in) == 0 {
		return nil
	}
	out := make([]string, len(in))
	copy(out, in)
	return out
}

func cloneInt32Slice(in []int32) []int32 {
	if len(in) == 0 {
		return nil
	}
	out := make([]int32, len(in))
	copy(out, in)
	return out
}

func cloneStringMap(in map[string]string) map[string]string {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func normalizeStrings(in []string) []string {
	if len(in) == 0 {
		return nil
	}
	out := make([]string, 0, len(in))
	seen := make(map[string]struct{}, len(in))
	for _, v := range in {
		trimmed := strings.TrimSpace(v)
		if trimmed == "" {
			continue
		}
		key := strings.ToLower(trimmed)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, key)
	}
	if len(out) == 0 {
		return nil
	}
	sort.Strings(out)
	return out
}

func (s *Server) applyDefaultFlavor(policy placement.ProjectPolicy, w *aegis.Workload) {
	defaultFlavor := strings.TrimSpace(policy.DefaultFlavor)
	if defaultFlavor == "" || w == nil {
		return
	}
	switch wk := w.GetKind().(type) {
	case *aegis.Workload_Workspace:
		if wk.Workspace != nil && strings.TrimSpace(wk.Workspace.Flavor) == "" {
			wk.Workspace.Flavor = defaultFlavor
		}
	case *aegis.Workload_Training:
		if wk.Training != nil && strings.TrimSpace(wk.Training.Flavor) == "" {
			wk.Training.Flavor = defaultFlavor
		}
	}
}

func (s *Server) activeWorkloadStats(projectID string) (map[string]int, map[string]int) {
	flavorCounts := map[string]int{}
	clusterLoads := map[string]int{}
	if strings.TrimSpace(projectID) == "" {
		return flavorCounts, clusterLoads
	}
	for _, existing := range s.store.ListWorkloads(projectID) {
		if existing == nil || !isWorkloadActive(existing.GetStatus()) {
			continue
		}
		if flavor, err := requiredFlavor(existing); err == nil {
			key := strings.ToLower(strings.TrimSpace(flavor))
			if key != "" {
				flavorCounts[key]++
			}
		}
		if cid := strings.TrimSpace(existing.GetClusterId()); cid != "" {
			clusterLoads[cid]++
		}
	}
	return flavorCounts, clusterLoads
}

func (s *Server) CreateWorkspace(ctx context.Context, req *aegis.CreateWorkspaceRequest) (*aegis.CreateWorkspaceResponse, error) {
	if req == nil || req.Workspace == nil {
		return nil, status.Error(codes.InvalidArgument, "workspace payload required")
	}

	w := &aegis.Workload{
		Id:        req.WorkspaceId,
		ProjectId: req.ProjectId,
		Queue:     req.Queue,
		Kind: &aegis.Workload_Workspace{
			Workspace: req.Workspace,
		},
	}

	submittedWorkload, err := s.SubmitWorkload(ctx, &aegis.SubmitWorkloadRequest{Workload: w})
	if err != nil {
		return nil, err
	}

	return &aegis.CreateWorkspaceResponse{Workload: submittedWorkload}, nil
}

func (s *Server) CreateCluster(ctx context.Context, req *aegis.CreateClusterRequest) (*aegis.CreateClusterResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	if s.infraClient == nil {
		return nil, status.Error(codes.FailedPrecondition, "cluster provisioning is not configured")
	}
	projectID := strings.TrimSpace(req.GetProjectId())
	clusterID := strings.TrimSpace(req.GetClusterId())
	region := strings.TrimSpace(req.GetRegion())
	provider := canonicalProvider(req.GetProvider())
	if projectID == "" {
		return nil, status.Error(codes.InvalidArgument, "project_id is required")
	}
	if err := s.authorize(ctx, projectID, "", "createCluster"); err != nil {
		return nil, err
	}
	if clusterID == "" {
		return nil, status.Error(codes.InvalidArgument, "cluster_id is required")
	}
	if region == "" {
		return nil, status.Error(codes.InvalidArgument, "region is required")
	}
	if provider != "aws" {
		return nil, status.Errorf(codes.InvalidArgument, "provider %q not supported", provider)
	}
	profileReq := req.GetProfile()
	if profileReq == nil || strings.TrimSpace(profileReq.GetId()) == "" {
		return nil, status.Error(codes.InvalidArgument, "profile.id is required")
	}
	profileID := strings.ToLower(strings.TrimSpace(profileReq.GetId()))
	tmpl, ok := s.clusterProfiles[profileID]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "profile %q not registered", profileReq.GetId())
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
	infra, err := s.buildProjectInfra(req, tmpl, project, creds)
	if err != nil {
		return nil, err
	}
	if err := s.ensureUniqueInfraName(ctx, infra); err != nil {
		return nil, err
	}
	if err := s.infraClient.Create(ctx, infra); err != nil {
		if apierrors.IsAlreadyExists(err) {
			reset, resetErr := s.resetFailedInfraJob(ctx, infra.Name, projectID, clusterID)
			if resetErr != nil {
				return nil, resetErr
			}
			if !reset {
				return nil, status.Errorf(codes.AlreadyExists, "cluster job %q already exists", infra.Name)
			}
			if err := s.infraClient.Create(ctx, infra); err != nil {
				if apierrors.IsAlreadyExists(err) {
					return nil, status.Errorf(codes.AlreadyExists, "cluster job %q already exists", infra.Name)
				}
				return nil, status.Errorf(codes.Internal, "create projectinfra: %v", err)
			}
		} else {
			return nil, status.Errorf(codes.Internal, "create projectinfra: %v", err)
		}
	}
	// Clear any old provisioning logs for this job ID immediately after creating the infra.
	// This ensures old logs are cleared BEFORE the frontend starts polling,
	// avoiding the race condition where old logs appear briefly.
	if s.store != nil {
		s.store.ClearProvisioningLogs(infra.Name)
	}
	s.log.Info("cluster provisioning job created",
		zap.String("project", projectID),
		zap.String("cluster", clusterID),
		zap.String("profile", tmpl.ID),
		zap.String("job", infra.Name),
	)
	return &aegis.CreateClusterResponse{Job: jobFromInfra(infra, infra.Name)}, nil
}

func (s *Server) GetClusterJobStatus(ctx context.Context, req *aegis.GetClusterJobStatusRequest) (*aegis.GetClusterJobStatusResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	if s.infraClient == nil {
		return nil, status.Error(codes.FailedPrecondition, "cluster provisioning is not configured")
	}
	infra, err := s.fetchInfra(ctx, strings.TrimSpace(req.GetJobId()))
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, status.Errorf(codes.NotFound, "cluster job %q not found", req.GetJobId())
		}
		return nil, status.Errorf(codes.Internal, "get job status: %v", err)
	}
	if err := s.authorize(ctx, infra.Spec.ProjectID, "", "getClusterJobStatus"); err != nil {
		return nil, err
	}
	return &aegis.GetClusterJobStatusResponse{Job: jobFromInfra(infra, req.GetJobId())}, nil
}

func isWorkloadActive(status string) bool {
	switch {
	case strings.EqualFold(status, statusPlaced):
		return true
	case strings.EqualFold(status, statusRunning):
		return true
	default:
		return false
	}
}

func ensureQueueAllowsFlavor(q *aegis.Queue, flavor string) bool {
	if q == nil {
		return false
	}
	normalized := strings.TrimSpace(flavor)
	if normalized == "" {
		return false
	}
	existing := q.GetAllowedFlavors()
	for _, item := range existing {
		if strings.EqualFold(strings.TrimSpace(item), normalized) {
			return false
		}
	}
	q.AllowedFlavors = append(existing, normalized)
	return true
}

func defaultFlavorForName(name string) *aegis.Flavor {
	normalized := strings.ToLower(strings.TrimSpace(name))
	switch normalized {
	case "t4-1gpu", "gpu-t4", "nvidia-tesla-t4", "t4", "gpu-standard":
		// gpu-standard maps to T4 GPU (g4dn.xlarge on AWS)
		// Conservative defaults: leave headroom for system pods, logging, and user's custom processes
		// GPU does heavy compute; CPU/RAM just feed data - most ML work is GPU-bound
		return &aegis.Flavor{
			Name:               name,
			Chip:               "nvidia-t4",
			ResourceName:       "nvidia.com/gpu",
			GpuCount:           1,
			MemoryGib:          16,
			CpuCoresRequest:    "2",
			MemoryRequest:      "8Gi",
			PriceUsdPerGpuHour: 0,
		}
	case "a10-1gpu", "a10g-1gpu", "gpu-large":
		// gpu-large maps to A10G GPU (g5.xlarge on AWS)
		// Conservative defaults: leave headroom for system pods and user customizations
		return &aegis.Flavor{
			Name: name,
			Chip: "nvidia-a10g",
			// Full A10G GPU uses the standard NVIDIA device plugin resource name.
			ResourceName:       "nvidia.com/gpu",
			GpuCount:           1,
			MemoryGib:          24,
			CpuCoresRequest:    "2",
			MemoryRequest:      "8Gi",
			PriceUsdPerGpuHour: 0,
		}
	case "gpu-heavy", "t4-heavy":
		// gpu-heavy: for data-intensive preprocessing that needs more CPU/RAM
		// Use when users need heavy data loading or CPU-side transforms
		return &aegis.Flavor{
			Name:               name,
			Chip:               "nvidia-t4",
			ResourceName:       "nvidia.com/gpu",
			GpuCount:           1,
			MemoryGib:          16,
			CpuCoresRequest:    "3",
			MemoryRequest:      "12Gi",
			PriceUsdPerGpuHour: 0,
		}
	case "a10g-mig-1g", "a10-mig-1g":
		return &aegis.Flavor{
			Name: name,
			Chip: "nvidia-a10g",
			// MIG 1g.10gb profile as reported by the NVIDIA device plugin.
			ResourceName:       "nvidia.com/mig-1g.10gb",
			GpuCount:           1,
			MemoryGib:          10,
			CpuCoresRequest:    "8",
			MemoryRequest:      "32Gi",
			PriceUsdPerGpuHour: 0,
		}
	case "cpu-small":
		return &aegis.Flavor{
			Name:               name,
			CpuCoresRequest:    "2",
			MemoryRequest:      "4Gi",
			GpuCount:           0,
			PriceUsdPerGpuHour: 0,
		}
	case "cpu-medium":
		return &aegis.Flavor{
			Name:               name,
			CpuCoresRequest:    "4",
			MemoryRequest:      "16Gi",
			GpuCount:           0,
			PriceUsdPerGpuHour: 0,
		}
	case "cpu-large":
		return &aegis.Flavor{
			Name:               name,
			CpuCoresRequest:    "8",
			MemoryRequest:      "32Gi",
			GpuCount:           0,
			PriceUsdPerGpuHour: 0,
		}
	default:
		return &aegis.Flavor{
			Name:               name,
			CpuCoresRequest:    "2",
			MemoryRequest:      "4Gi",
			ResourceName:       name,
			GpuCount:           0,
			PriceUsdPerGpuHour: 0,
		}
	}
}

func (s *Server) ensureFlavorDefaults(name string) {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return
	}
	existing := s.store.GetFlavor(trimmed)
	expected := defaultFlavorForName(trimmed)
	if expected == nil {
		return
	}

	// Update flavor if it doesn't exist, or if it exists but has missing/outdated values
	needsUpdate := existing == nil ||
		strings.TrimSpace(existing.GetResourceName()) == "" ||
		(expected.GetGpuCount() > 0 && existing.GetGpuCount() == 0) ||
		// Also update if CPU/memory values differ from expected defaults
		(expected.GetCpuCoresRequest() != "" && existing.GetCpuCoresRequest() != expected.GetCpuCoresRequest()) ||
		(expected.GetMemoryRequest() != "" && existing.GetMemoryRequest() != expected.GetMemoryRequest())

	if needsUpdate {
		s.store.PutFlavor(expected)
		s.log.Info("autobootstrap: flavor ensured",
			zap.String("flavor", trimmed),
			zap.String("resource_name", expected.GetResourceName()),
			zap.Int32("gpu_count", expected.GetGpuCount()),
			zap.String("cpu_request", expected.GetCpuCoresRequest()),
			zap.String("memory_request", expected.GetMemoryRequest()),
		)
	}
}

func stringPtr(in string) *string {
	if strings.TrimSpace(in) == "" {
		return nil
	}
	v := in
	return &v
}

func selectWorkspacePort(spec *aegis.WorkspaceSpec) int32 {
	if spec != nil {
		var firstPositive int32
		for _, port := range spec.GetPorts() {
			if port <= 0 {
				continue
			}
			if port == workspacecfg.DefaultVSCodePort {
				return port
			}
			if firstPositive == 0 {
				firstPositive = port
			}
		}
		if firstPositive > 0 {
			return firstPositive
		}
	}
	return workspacecfg.DefaultVSCodePort
}

func (s *Server) applyWorkspaceDefaults(ws *aegis.WorkspaceSpec) {
	if ws == nil {
		return
	}
	ws.Ports = workspacecfg.EnsureDefaultPorts(ws.GetPorts())
	mergedEnv := workspacecfg.MergeEnv(ws.GetEnv(), s.workspaceEnvDefaults)
	if len(mergedEnv) == 0 {
		ws.Env = nil
	} else {
		ws.Env = mergedEnv
	}
}

func requiredFlavor(w *aegis.Workload) (string, error) {
	switch wk := w.GetKind().(type) {
	case *aegis.Workload_Workspace:
		if f := wk.Workspace.GetFlavor(); f != "" {
			return f, nil
		}
	case *aegis.Workload_Training:
		if f := wk.Training.GetFlavor(); f != "" {
			return f, nil
		}
	default:
	}
	return "", fmt.Errorf("flavor required on workspace or training")
}

func RandID() string {
	buf := make([]byte, 4)
	if _, err := rand.Read(buf); err != nil {
		return strconv.FormatInt(time.Now().UnixNano(), 16)
	}
	return hex.EncodeToString(buf)
}

func (s *Server) authorize(ctx context.Context, projectID, queue, action string) error {
	if s == nil || s.authzPolicy == nil {
		return nil
	}
	identity := mw.IdentityFromContext(ctx)
	if s.authzPolicy.Authorize(identity, projectID, queue) {
		return nil
	}
	subject := subjectFromContext(ctx)
	if s.log != nil {
		s.log.Warn("authorization denied",
			zap.String("action", action),
			zap.String("subject", subject),
			zap.String("project_id", projectID),
			zap.String("queue", queue),
		)
	}
	return status.Error(codes.PermissionDenied, "authorization denied")
}

func subjectFromContext(ctx context.Context) string {
	if ctx == nil {
		return "unknown"
	}
	if id := mw.IdentityFromContext(ctx); id != nil {
		if id.Subject != "" {
			return id.Subject
		}
		if id.Email != "" {
			return id.Email
		}
		if id.PreferredUsername != "" {
			return id.PreferredUsername
		}
	}
	return "unknown"
}
