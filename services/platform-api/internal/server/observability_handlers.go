package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"

	infraapi "github.com/yourorg/aegis/services/platform-api/api/v1alpha1"
)

const (
	defaultLogLookback = 15 * time.Minute
	maxLogLimit        = 1000
	defaultLogLimit    = 200
	defaultMetricStep  = 30 * time.Second
	defaultMetricRange = 15 * time.Minute
	requestTimeout     = 15 * time.Second
)

type logsQueryRequest struct {
	ProjectID     string `json:"projectId"`
	ClusterID     string `json:"clusterId"`
	Namespace     string `json:"namespace"`
	Pod           string `json:"pod"`
	Substring     string `json:"substring"`
	Start         string `json:"start"`
	End           string `json:"end"`
	Limit         int    `json:"limit"`
	Cursor        string `json:"cursor"`
	IncludeEvents bool   `json:"includeEvents"`
}

type logEntry struct {
	Timestamp   string            `json:"timestamp"`
	Namespace   string            `json:"namespace,omitempty"`
	Pod         string            `json:"pod,omitempty"`
	Container   string            `json:"container,omitempty"`
	App         string            `json:"app,omitempty"`
	EventReason string            `json:"eventReason,omitempty"`
	EventType   string            `json:"eventType,omitempty"`
	Message     string            `json:"message"`
	Labels      map[string]string `json:"labels,omitempty"`
}

type logsResponse struct {
	Entries    []logEntry `json:"entries"`
	NextCursor string     `json:"nextCursor,omitempty"`
	Warnings   []string   `json:"warnings,omitempty"`
}

type metricsQueryRequest struct {
	ProjectID string `json:"projectId"`
	ClusterID string `json:"clusterId"`
	Query     string `json:"query"`
	Start     string `json:"start"`
	End       string `json:"end"`
	StepSec   int    `json:"stepSeconds"`
	RangeSec  int    `json:"rangeSeconds"`
}

type metricSample struct {
	Timestamp string  `json:"timestamp"`
	Value     float64 `json:"value"`
}

type metricSeries struct {
	Labels  map[string]string `json:"labels,omitempty"`
	Samples []metricSample    `json:"samples,omitempty"`
}

type metricsResponse struct {
	Series []metricSeries `json:"series"`
}

type traceResponse struct {
	Trace json.RawMessage `json:"trace,omitempty"`
}

type alertsResponse struct {
	Alerts []alertView `json:"alerts"`
}

type alertView struct {
	State       string            `json:"state,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
	StartsAt    string            `json:"startsAt,omitempty"`
	EndsAt      string            `json:"endsAt,omitempty"`
	Fingerprint string            `json:"fingerprint,omitempty"`
}

func registerObservabilityRoutes(mux *runtime.ServeMux, srv *Server) {
	if mux == nil || srv == nil {
		return
	}
	mux.HandlePath(http.MethodPost, "/api/logs/query", func(w http.ResponseWriter, r *http.Request, _ map[string]string) {
		srv.handleLogsQuery(w, r)
	})
	mux.HandlePath(http.MethodPost, "/api/metrics/query", func(w http.ResponseWriter, r *http.Request, _ map[string]string) {
		srv.handleMetricsQuery(w, r)
	})
	mux.HandlePath(http.MethodGet, "/api/traces/{traceId}", func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
		srv.handleTraceLookup(w, r, pathParams["traceId"])
	})
	mux.HandlePath(http.MethodGet, "/api/alerts", func(w http.ResponseWriter, r *http.Request, _ map[string]string) {
		srv.handleAlerts(w, r)
	})
}

func (s *Server) handleLogsQuery(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req logsQueryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeWizardError(w, status.Errorf(codes.InvalidArgument, "invalid request body: %v", err))
		return
	}
	projectID := strings.TrimSpace(req.ProjectID)
	clusterID := strings.TrimSpace(req.ClusterID)
	if projectID == "" || clusterID == "" {
		writeWizardError(w, status.Error(codes.InvalidArgument, "projectId and clusterId are required"))
		return
	}
	if err := s.authorize(ctx, projectID, "", "queryLogs"); err != nil {
		writeWizardError(w, err)
		return
	}
	obs, err := s.resolveObservability(ctx, projectID, clusterID)
	if err != nil {
		writeWizardError(w, err)
		return
	}

	httpClient, lokiURL, err := s.resolveServiceEndpoint(ctx, clusterID, pickNamespace(obs.LokiNamespace, obs.Namespace), obs.LokiService, obs.LokiPort)
	if err != nil || lokiURL == "" {
		writeWizardError(w, status.Error(codes.FailedPrecondition, "loki endpoint not available for cluster"))
		return
	}

	start, end := resolveWindow(req.Start, req.End, defaultLogLookback)
	if !end.After(start) {
		writeWizardError(w, status.Error(codes.InvalidArgument, "end must be after start"))
		return
	}
	limit := req.Limit
	if limit <= 0 {
		limit = defaultLogLimit
	}
	if limit > maxLogLimit {
		limit = maxLogLimit
	}
	if cursorTS := strings.TrimSpace(req.Cursor); cursorTS != "" {
		if ts, err := time.Parse(time.RFC3339Nano, cursorTS); err == nil {
			start = ts.Add(time.Nanosecond)
		}
	}

	query := buildLokiQuery(clusterID, req.Namespace, req.Pod, req.Substring, req.IncludeEvents)
	params := url.Values{}
	params.Set("query", query)
	params.Set("limit", strconv.Itoa(limit))
	params.Set("direction", "forward")
	params.Set("start", strconv.FormatInt(start.UnixNano(), 10))
	params.Set("end", strconv.FormatInt(end.UnixNano(), 10))

	reqCtx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()

	httpReq, err := http.NewRequestWithContext(reqCtx, http.MethodGet, fmt.Sprintf("%s/loki/api/v1/query_range?%s", lokiURL, params.Encode()), nil)
	if err != nil {
		writeWizardError(w, status.Errorf(codes.Internal, "build loki request: %v", err))
		return
	}
	resp, err := httpClient.Do(httpReq)
	if err != nil {
		writeWizardError(w, status.Errorf(codes.Unavailable, "query loki: %v", err))
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		writeWizardError(w, status.Errorf(codes.Internal, "loki returned status %d", resp.StatusCode))
		return
	}
	var lokiResp struct {
		Status string `json:"status"`
		Data   struct {
			ResultType string `json:"resultType"`
			Result     []struct {
				Stream map[string]string `json:"stream"`
				Values [][]string        `json:"values"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&lokiResp); err != nil {
		writeWizardError(w, status.Errorf(codes.Internal, "decode loki response: %v", err))
		return
	}
	out := logsResponse{}
	for _, stream := range lokiResp.Data.Result {
		for _, pair := range stream.Values {
			if len(pair) != 2 {
				continue
			}
			ns, msg := pair[0], pair[1]
			tsInt, err := strconv.ParseInt(strings.TrimSpace(ns), 10, 64)
			if err != nil {
				continue
			}
			ts := time.Unix(0, tsInt)
			entry := logEntry{
				Timestamp: ts.UTC().Format(time.RFC3339Nano),
				Message:   msg,
				Labels:    stream.Stream,
			}
			entry.Namespace = stream.Stream["namespace"]
			entry.Pod = stream.Stream["pod"]
			entry.Container = stream.Stream["container"]
			entry.App = stream.Stream["app"]
			entry.EventReason = stream.Stream["event_reason"]
			entry.EventType = stream.Stream["event_type"]
			out.Entries = append(out.Entries, entry)
		}
	}
	if n := len(out.Entries); n > 0 {
		out.NextCursor = out.Entries[n-1].Timestamp
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) handleMetricsQuery(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req metricsQueryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeWizardError(w, status.Errorf(codes.InvalidArgument, "invalid request body: %v", err))
		return
	}
	projectID := strings.TrimSpace(req.ProjectID)
	clusterID := strings.TrimSpace(req.ClusterID)
	if projectID == "" || clusterID == "" {
		writeWizardError(w, status.Error(codes.InvalidArgument, "projectId and clusterId are required"))
		return
	}
	if strings.TrimSpace(req.Query) == "" {
		writeWizardError(w, status.Error(codes.InvalidArgument, "query is required"))
		return
	}
	if err := s.authorize(ctx, projectID, "", "queryMetrics"); err != nil {
		writeWizardError(w, err)
		return
	}
	obs, err := s.resolveObservability(ctx, projectID, clusterID)
	if err != nil {
		writeWizardError(w, err)
		return
	}

	httpClient, promURL, err := s.resolveServiceEndpoint(ctx, clusterID, pickNamespace(obs.Namespace, obs.Namespace), obs.PrometheusService, obs.PrometheusPort)
	if err != nil || promURL == "" {
		writeWizardError(w, status.Error(codes.FailedPrecondition, "prometheus endpoint not available for cluster"))
		return
	}

	step := time.Duration(req.StepSec) * time.Second
	if step <= 0 {
		step = defaultMetricStep
	}
	lookback := defaultMetricRange
	if req.RangeSec > 0 {
		lookback = time.Duration(req.RangeSec) * time.Second
	}
	start, end := resolveWindow(req.Start, req.End, lookback)
	if !end.After(start) {
		writeWizardError(w, status.Error(codes.InvalidArgument, "end must be after start"))
		return
	}

	params := url.Values{}
	params.Set("query", req.Query)
	params.Set("start", strconv.FormatFloat(float64(start.UnixNano())/1e9, 'f', -1, 64))
	params.Set("end", strconv.FormatFloat(float64(end.UnixNano())/1e9, 'f', -1, 64))
	params.Set("step", strconv.FormatFloat(step.Seconds(), 'f', -1, 64))

	reqCtx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()

	httpReq, err := http.NewRequestWithContext(reqCtx, http.MethodGet, fmt.Sprintf("%s/api/v1/query_range?%s", promURL, params.Encode()), nil)
	if err != nil {
		writeWizardError(w, status.Errorf(codes.Internal, "build prometheus request: %v", err))
		return
	}
	resp, err := httpClient.Do(httpReq)
	if err != nil {
		writeWizardError(w, status.Errorf(codes.Unavailable, "query prometheus: %v", err))
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		writeWizardError(w, status.Errorf(codes.Internal, "prometheus returned status %d", resp.StatusCode))
		return
	}
	var promResp struct {
		Status string `json:"status"`
		Data   struct {
			ResultType string `json:"resultType"`
			Result     []struct {
				Metric map[string]string `json:"metric"`
				Values [][]interface{}   `json:"values"`
				Value  []interface{}     `json:"value"`
				Scalar interface{}       `json:"scalar"`
				Vector interface{}       `json:"vector"`
				String interface{}       `json:"string"`
				Hist   interface{}       `json:"histogram"`
				Hist2  interface{}       `json:"histograms"`
			} `json:"result"`
		} `json:"data"`
		ErrorType string `json:"errorType,omitempty"`
		Error     string `json:"error,omitempty"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&promResp); err != nil {
		writeWizardError(w, status.Errorf(codes.Internal, "decode prometheus response: %v", err))
		return
	}
	if promResp.Error != "" {
		writeWizardError(w, status.Errorf(codes.InvalidArgument, "prometheus error: %s", promResp.Error))
		return
	}
	out := metricsResponse{}
	for _, series := range promResp.Data.Result {
		view := metricSeries{Labels: series.Metric}
		for _, pair := range series.Values {
			if len(pair) != 2 {
				continue
			}
			ts, ok := parsePromTimestamp(pair[0])
			if !ok {
				continue
			}
			valFloat, ok := parsePromValue(pair[1])
			if !ok {
				continue
			}
			view.Samples = append(view.Samples, metricSample{
				Timestamp: ts.UTC().Format(time.RFC3339Nano),
				Value:     valFloat,
			})
		}
		out.Series = append(out.Series, view)
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) handleTraceLookup(w http.ResponseWriter, r *http.Request, traceID string) {
	ctx := r.Context()
	projectID := strings.TrimSpace(r.URL.Query().Get("projectId"))
	clusterID := strings.TrimSpace(r.URL.Query().Get("clusterId"))
	traceID = strings.TrimSpace(traceID)
	if projectID == "" || clusterID == "" || traceID == "" {
		writeWizardError(w, status.Error(codes.InvalidArgument, "projectId, clusterId, and traceId are required"))
		return
	}
	if err := s.authorize(ctx, projectID, "", "getTrace"); err != nil {
		writeWizardError(w, err)
		return
	}
	obs, err := s.resolveObservability(ctx, projectID, clusterID)
	if err != nil {
		writeWizardError(w, err)
		return
	}
	httpClient, tempoURL, err := s.resolveServiceEndpoint(ctx, clusterID, pickNamespace(obs.TracingNamespace, obs.Namespace), obs.TempoService, obs.TempoPort)
	if err != nil || tempoURL == "" {
		writeWizardError(w, status.Error(codes.FailedPrecondition, "tempo endpoint not available for cluster"))
		return
	}

	reqCtx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()

	httpReq, err := http.NewRequestWithContext(reqCtx, http.MethodGet, fmt.Sprintf("%s/api/traces/%s", tempoURL, url.PathEscape(traceID)), nil)
	if err != nil {
		writeWizardError(w, status.Errorf(codes.Internal, "build tempo request: %v", err))
		return
	}
	resp, err := httpClient.Do(httpReq)
	if err != nil {
		writeWizardError(w, status.Errorf(codes.Unavailable, "query tempo: %v", err))
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		writeWizardError(w, status.Error(codes.NotFound, "trace not found"))
		return
	}
	if resp.StatusCode >= 400 {
		writeWizardError(w, status.Errorf(codes.Internal, "tempo returned status %d", resp.StatusCode))
		return
	}
	var payload json.RawMessage
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		writeWizardError(w, status.Errorf(codes.Internal, "decode tempo response: %v", err))
		return
	}
	writeJSON(w, http.StatusOK, traceResponse{Trace: payload})
}

func (s *Server) handleAlerts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	projectID := strings.TrimSpace(r.URL.Query().Get("projectId"))
	clusterID := strings.TrimSpace(r.URL.Query().Get("clusterId"))
	if projectID == "" || clusterID == "" {
		writeWizardError(w, status.Error(codes.InvalidArgument, "projectId and clusterId are required"))
		return
	}
	if err := s.authorize(ctx, projectID, "", "getAlerts"); err != nil {
		writeWizardError(w, err)
		return
	}
	obs, err := s.resolveObservability(ctx, projectID, clusterID)
	if err != nil {
		writeWizardError(w, err)
		return
	}
	httpClient, alertURL, err := s.resolveServiceEndpoint(ctx, clusterID, pickNamespace(obs.Namespace, obs.Namespace), obs.AlertmanagerService, obs.AlertmanagerPort)
	if err != nil || alertURL == "" {
		writeWizardError(w, status.Error(codes.FailedPrecondition, "alertmanager endpoint not available for cluster"))
		return
	}

	params := url.Values{}
	params.Set("silenced", "false")
	params.Set("inhibited", "false")
	params.Set("active", "true")

	reqCtx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()

	httpReq, err := http.NewRequestWithContext(reqCtx, http.MethodGet, fmt.Sprintf("%s/api/v2/alerts?%s", alertURL, params.Encode()), nil)
	if err != nil {
		writeWizardError(w, status.Errorf(codes.Internal, "build alertmanager request: %v", err))
		return
	}
	resp, err := httpClient.Do(httpReq)
	if err != nil {
		writeWizardError(w, status.Errorf(codes.Unavailable, "query alertmanager: %v", err))
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		writeWizardError(w, status.Errorf(codes.Internal, "alertmanager returned status %d", resp.StatusCode))
		return
	}
	var alerts []struct {
		Status      string            `json:"status"`
		Labels      map[string]string `json:"labels"`
		Annotations map[string]string `json:"annotations"`
		StartsAt    string            `json:"startsAt"`
		EndsAt      string            `json:"endsAt"`
		Fingerprint string            `json:"fingerprint"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&alerts); err != nil {
		writeWizardError(w, status.Errorf(codes.Internal, "decode alertmanager response: %v", err))
		return
	}
	out := alertsResponse{Alerts: make([]alertView, 0, len(alerts))}
	for _, a := range alerts {
		out.Alerts = append(out.Alerts, alertView{
			State:       a.Status,
			Labels:      a.Labels,
			Annotations: a.Annotations,
			StartsAt:    a.StartsAt,
			EndsAt:      a.EndsAt,
			Fingerprint: a.Fingerprint,
		})
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) resolveObservability(ctx context.Context, projectID, clusterID string) (*infraapi.ObservabilityOutput, error) {
	if s.infraClient == nil {
		return nil, status.Error(codes.FailedPrecondition, "infrastructure client not configured")
	}
	projectID = strings.TrimSpace(projectID)
	clusterID = strings.TrimSpace(clusterID)
	if projectID == "" || clusterID == "" {
		return nil, status.Error(codes.InvalidArgument, "projectId and clusterId are required")
	}

	var infra infraapi.ProjectInfraList
	if err := s.infraClient.List(ctx, &infra, client.InNamespace(s.infraNamespace)); err != nil {
		return nil, status.Errorf(codes.Internal, "list project infrastructure: %v", err)
	}
	for _, item := range infra.Items {
		if !strings.EqualFold(strings.TrimSpace(item.Spec.ProjectID), projectID) {
			continue
		}
		for _, out := range item.Status.Outputs {
			if strings.EqualFold(strings.TrimSpace(out.ClusterID), clusterID) {
				result := out.Observability
				return &result, nil
			}
		}
	}

	// If not found via cached status, try direct fetch in case of eventual consistency.
	key := types.NamespacedName{Name: buildInfraObjectName(projectID, clusterID), Namespace: s.infraNamespace}
	var single infraapi.ProjectInfra
	if err := s.infraClient.Get(ctx, key, &single); err == nil {
		for _, out := range single.Status.Outputs {
			if strings.EqualFold(strings.TrimSpace(out.ClusterID), clusterID) {
				result := out.Observability
				return &result, nil
			}
		}
	} else if !errors.IsNotFound(err) {
		return nil, status.Errorf(codes.Internal, "get project infra: %v", err)
	}

	return nil, status.Errorf(codes.NotFound, "cluster %q not found for project %q", clusterID, projectID)
}

func (s *Server) httpClient() *http.Client {
	return &http.Client{Timeout: requestTimeout}
}

// resolveServiceEndpoint picks a reachable endpoint for a service within a target cluster.
// It prefers a Kubernetes API service proxy (using the kubeconfig for the cluster) when available,
// and falls back to in-cluster service DNS if no proxy config is present.
func (s *Server) resolveServiceEndpoint(ctx context.Context, clusterID, namespace, service string, port int32) (*http.Client, string, error) {
	_ = ctx // currently unused; reserved for future context-aware dialing
	// Prefer proxy through the target cluster's API server if we have a rest.Config.
	if s != nil && kubeClientProviderConfigured(s.kubeClients) {
		if cfg, err := s.kubeClients.RestConfigFor(strings.TrimSpace(clusterID)); err == nil && cfg != nil {
			if httpClient, err := rest.HTTPClientFor(cfg); err == nil && httpClient != nil {
				if proxyURL := buildServiceProxyURL(cfg.Host, namespace, service, port); proxyURL != "" {
					return httpClient, proxyURL, nil
				}
			}
		}
	}

	// Fallback to service DNS inside the same cluster/network.
	fallback := serviceURL(service, namespace, port)
	if fallback == "" {
		return nil, "", status.Error(codes.FailedPrecondition, "service endpoint unavailable")
	}
	return s.httpClient(), fallback, nil
}

func serviceURL(service, namespace string, port int32) string {
	service = strings.TrimSpace(service)
	namespace = strings.TrimSpace(namespace)
	if service == "" || namespace == "" || port <= 0 {
		return ""
	}
	return fmt.Sprintf("http://%s.%s.svc:%d", service, namespace, port)
}

func buildServiceProxyURL(apiServer, namespace, service string, port int32) string {
	apiServer = strings.TrimSpace(apiServer)
	namespace = strings.TrimSpace(namespace)
	service = strings.TrimSpace(service)
	if apiServer == "" || namespace == "" || service == "" || port <= 0 {
		return ""
	}
	trimmed := strings.TrimRight(apiServer, "/")
	return fmt.Sprintf("%s/api/v1/namespaces/%s/services/%s:%d/proxy", trimmed, namespace, service, port)
}

func pickNamespace(preferred, fallback string) string {
	if ns := strings.TrimSpace(preferred); ns != "" {
		return ns
	}
	return strings.TrimSpace(fallback)
}

func resolveWindow(startRaw, endRaw string, defRange time.Duration) (time.Time, time.Time) {
	end := time.Now().UTC()
	if endRaw != "" {
		if parsed, err := time.Parse(time.RFC3339Nano, strings.TrimSpace(endRaw)); err == nil {
			end = parsed
		}
	}
	start := end.Add(-defRange)
	if startRaw != "" {
		if parsed, err := time.Parse(time.RFC3339Nano, strings.TrimSpace(startRaw)); err == nil {
			start = parsed
		}
	}
	return start, end
}

func buildLokiQuery(clusterID, namespace, pod, substring string, includeEvents bool) string {
	parts := []string{fmt.Sprintf(`cluster="%s"`, clusterID)}
	if ns := strings.TrimSpace(namespace); ns != "" {
		parts = append(parts, fmt.Sprintf(`namespace="%s"`, ns))
	}
	if p := strings.TrimSpace(pod); p != "" {
		parts = append(parts, fmt.Sprintf(`pod="%s"`, p))
	}
	if includeEvents {
		// No-op: events are already labeled; placeholder for future event-only filtering.
	}
	labelSelector := "{" + strings.Join(parts, ",") + "}"
	if sub := strings.TrimSpace(substring); sub != "" {
		return fmt.Sprintf(`%s |= %s`, labelSelector, strconv.Quote(sub))
	}
	return labelSelector
}

func parsePromTimestamp(v interface{}) (time.Time, bool) {
	switch t := v.(type) {
	case float64:
		sec := int64(t)
		nsec := int64((t - float64(sec)) * 1e9)
		return time.Unix(sec, nsec), true
	case string:
		if t == "" {
			return time.Time{}, false
		}
		if strings.Contains(t, ".") {
			if f, err := strconv.ParseFloat(t, 64); err == nil {
				sec := int64(f)
				nsec := int64((f - float64(sec)) * 1e9)
				return time.Unix(sec, nsec), true
			}
		}
		if i, err := strconv.ParseInt(t, 10, 64); err == nil {
			return time.Unix(i, 0), true
		}
	}
	return time.Time{}, false
}

func parsePromValue(v interface{}) (float64, bool) {
	switch val := v.(type) {
	case float64:
		return val, true
	case string:
		f, err := strconv.ParseFloat(strings.TrimSpace(val), 64)
		return f, err == nil
	default:
		return 0, false
	}
}
