package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	workspacecfg "github.com/yourorg/aegis/pkg/workspace"
	aegis "github.com/yourorg/aegis/proto/aegis/v1"
	"github.com/yourorg/aegis/services/platform-api/internal/store"
)

type projectView struct {
	ID       string         `json:"id"`
	Name     string         `json:"name,omitempty"`
	Clusters []clusterView  `json:"clusters,omitempty"`
	Labels   map[string]any `json:"labels,omitempty"`
}

type clusterView struct {
	ID            string          `json:"id"`
	Name          string          `json:"name"`
	ProjectID     string          `json:"projectId,omitempty"`
	Region        string          `json:"region,omitempty"`
	Provider      string          `json:"provider,omitempty"`
	K8sVersion    string          `json:"k8sVersion,omitempty"`
	HasGPU        bool            `json:"hasGpu"`
	NodeGroups    []nodeGroupView `json:"nodeGroups,omitempty"`
	Status        string          `json:"status,omitempty"`
	LastHeartbeat string          `json:"lastHeartbeat,omitempty"`
}

type nodeGroupView struct {
	Name string `json:"name"`
	GPU  bool   `json:"gpu"`
}

type workspaceCreateRequest struct {
	ProjectID   string            `json:"projectId"`
	ClusterID   string            `json:"clusterId"`
	Name        string            `json:"name"`
	Profile     string            `json:"profile"`
	Template    string            `json:"template"`
	Parameters  map[string]string `json:"parameters"`
	RequestedBy string            `json:"requestedBy"`
	Flavor      string            `json:"flavor"`
	Queue       string            `json:"queue"`
	Image       string            `json:"image"`
}

type workspaceCreateResponse struct {
	ID        string `json:"id"`
	ProjectID string `json:"projectId"`
	ClusterID string `json:"clusterId"`
	Name      string `json:"name"`
	Status    string `json:"status"`
	CreatedAt string `json:"createdAt,omitempty"`
}

func registerWorkspaceWizardRoutes(mux *runtime.ServeMux, srv *Server) {
	if mux == nil || srv == nil {
		return
	}
	mux.HandlePath(http.MethodGet, "/api/projects", func(w http.ResponseWriter, r *http.Request, _ map[string]string) {
		srv.handleProjects(w, r)
	})
	mux.HandlePath(http.MethodGet, "/api/clusters", func(w http.ResponseWriter, r *http.Request, _ map[string]string) {
		srv.handleClusters(w, r)
	})
	mux.HandlePath(http.MethodPost, "/api/workspaces", func(w http.ResponseWriter, r *http.Request, _ map[string]string) {
		srv.handleCreateWorkspace(w, r)
	})
}

func (s *Server) handleProjects(w http.ResponseWriter, r *http.Request) {
	views, err := s.projectViews(r.Context())
	if err != nil {
		writeWizardError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"projects": views})
}

func (s *Server) handleClusters(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	projectID := strings.TrimSpace(r.URL.Query().Get("projectId"))

	_, allowed, err := s.authorizedProjects(ctx)
	if err != nil {
		writeWizardError(w, err)
		return
	}
	if projectID != "" {
		if _, ok := allowed[strings.ToLower(projectID)]; !ok {
			writeWizardError(w, status.Error(codes.PermissionDenied, "project not accessible"))
			return
		}
	}

	clusters, err := s.clusterViews(ctx, allowed, projectID)
	if err != nil {
		writeWizardError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"clusters": clusters})
}

func (s *Server) handleCreateWorkspace(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req workspaceCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeWizardError(w, status.Errorf(codes.InvalidArgument, "invalid request body: %v", err))
		return
	}

	projectID := strings.TrimSpace(req.ProjectID)
	clusterID := strings.TrimSpace(req.ClusterID)
	name := strings.TrimSpace(req.Name)
	if projectID == "" || clusterID == "" {
		writeWizardError(w, status.Error(codes.InvalidArgument, "projectId and clusterId are required"))
		return
	}
	if name == "" {
		writeWizardError(w, status.Error(codes.InvalidArgument, "name is required"))
		return
	}

	project := s.store.GetProject(projectID)
	if project == nil {
		writeWizardError(w, status.Error(codes.NotFound, "project not found"))
		return
	}
	if err := s.authorize(ctx, projectID, "", "createWorkspace"); err != nil {
		writeWizardError(w, err)
		return
	}

	clusterInfo := s.clusterInfo(clusterID)
	if clusterInfo == nil {
		writeWizardError(w, status.Errorf(codes.NotFound, "cluster %q not registered", clusterID))
		return
	}
	if proj := clusterProject(clusterInfo); proj != "" && !strings.EqualFold(proj, projectID) {
		writeWizardError(w, status.Errorf(codes.InvalidArgument, "cluster %s not linked to project %s", clusterID, projectID))
		return
	}
	now := time.Now()
	if !clusterReady(clusterInfo, now) {
		writeWizardError(w, status.Errorf(codes.FailedPrecondition, "cluster %s not ready", clusterID))
		return
	}

	if s.workspaceNameExists(projectID, name) {
		writeWizardError(w, status.Error(codes.AlreadyExists, "workspace name already exists for project"))
		return
	}

	flavor := selectFlavorForCluster(s.store, clusterInfo, req)
	if flavor == "" {
		writeWizardError(w, status.Errorf(codes.FailedPrecondition, "cluster %s has no available flavors", clusterID))
		return
	}
	if s.autoBootstrap {
		s.ensureFlavorDefaults(flavor)
	}

	queue := strings.TrimSpace(req.Queue)
	if queue == "" {
		queue = sanitizeKubeName(fmt.Sprintf("%s-%s", projectID, "workspaces"))
		if queue == "" {
			queue = fmt.Sprintf("%s-workspaces", projectID)
		}
	}
	if s.autoBootstrap {
		s.ensureWorkspaceQueue(projectID, queue, flavor)
	}

	env := map[string]string{
		"WORKSPACE_NAME": name,
		"PROJECT_ID":     projectID,
		"CLUSTER_ID":     clusterID,
	}
	if req.RequestedBy != "" {
		env["REQUESTED_BY"] = req.RequestedBy
	}
	if req.Profile != "" {
		env["WORKSPACE_PROFILE"] = req.Profile
	}
	if req.Template != "" {
		env["WORKSPACE_TEMPLATE"] = req.Template
	}
	if len(req.Parameters) > 0 {
		if payload, err := json.Marshal(req.Parameters); err == nil {
			env["WORKSPACE_PARAMETERS"] = string(payload)
		}
		if override, ok := req.Parameters["image"]; ok && strings.TrimSpace(req.Image) == "" {
			req.Image = override
		}
		if override, ok := req.Parameters["flavor"]; ok && strings.TrimSpace(req.Flavor) == "" {
			if trimmed := strings.TrimSpace(override); trimmed != "" {
				flavor = trimmed
			}
		}
	}

	image := strings.TrimSpace(req.Image)
	if image == "" {
		image = workspacecfg.DefaultWorkspaceImage
	}

	maxDuration := int64(0)
	if raw, ok := req.Parameters["maxDurationSeconds"]; ok {
		if parsed, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64); err == nil && parsed > 0 {
			maxDuration = parsed
		}
	}

	id := buildWorkspaceID(projectID, name)
	workload := &aegis.Workload{
		Id:        id,
		ProjectId: projectID,
		Queue:     queue,
		ClusterId: clusterID,
		Kind: &aegis.Workload_Workspace{
			Workspace: &aegis.WorkspaceSpec{
				Flavor:             flavor,
				Image:              image,
				Env:                env,
				Interactive:        true,
				MaxDurationSeconds: maxDuration,
			},
		},
	}

	res, err := s.SubmitWorkload(ctx, &aegis.SubmitWorkloadRequest{Workload: workload})
	if err != nil {
		writeWizardError(w, err)
		return
	}

	statusText := strings.ToLower(res.GetStatus())
	if strings.EqualFold(res.GetStatus(), statusPlaced) {
		statusText = "pending"
	}
	resp := workspaceCreateResponse{
		ID:        res.GetId(),
		ProjectID: projectID,
		ClusterID: res.GetClusterId(),
		Name:      name,
		Status:    statusText,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
	writeJSON(w, http.StatusCreated, resp)
}

func (s *Server) projectViews(ctx context.Context) ([]projectView, error) {
	projects, allowed, err := s.authorizedProjects(ctx)
	if err != nil {
		return nil, err
	}
	clusters, err := s.clusterViews(ctx, allowed, "")
	if err != nil {
		return nil, err
	}
	clusterMap := make(map[string][]clusterView)
	for _, c := range clusters {
		key := strings.ToLower(c.ProjectID)
		clusterMap[key] = append(clusterMap[key], c)
	}
	out := make([]projectView, 0, len(projects))
	for _, p := range projects {
		if p == nil {
			continue
		}
		id := strings.TrimSpace(p.GetId())
		view := projectView{
			ID:       id,
			Name:     strings.TrimSpace(p.GetDisplayName()),
			Clusters: clusterMap[strings.ToLower(id)],
		}
		out = append(out, view)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

func (s *Server) clusterViews(ctx context.Context, allowed map[string]struct{}, filterProject string) ([]clusterView, error) {
	_ = ctx
	infos := s.store.ListClusterInfos()
	now := time.Now()
	if allowed != nil && len(allowed) == 0 {
		return []clusterView{}, nil
	}
	out := make([]clusterView, 0, len(infos))
	for _, ci := range infos {
		if ci == nil {
			continue
		}
		projectID := clusterProject(ci)
		if allowed != nil {
			if _, ok := allowed[strings.ToLower(projectID)]; !ok {
				continue
			}
		}
		if filterProject != "" && !strings.EqualFold(filterProject, projectID) {
			continue
		}
		view := buildClusterView(ci, now, s.store)
		out = append(out, view)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

func (s *Server) authorizedProjects(ctx context.Context) ([]*aegis.Project, map[string]struct{}, error) {
	resp, err := s.ListProjects(ctx, &aegis.ListProjectsRequest{})
	if err != nil {
		return nil, nil, err
	}
	allowed := make(map[string]struct{}, len(resp.GetItems()))
	for _, p := range resp.GetItems() {
		if p == nil {
			continue
		}
		if id := strings.TrimSpace(p.GetId()); id != "" {
			allowed[strings.ToLower(id)] = struct{}{}
		}
	}
	return resp.GetItems(), allowed, nil
}

func buildClusterView(ci *store.ClusterInfo, now time.Time, st store.Store) clusterView {
	status := "unknown"
	if clusterReady(ci, now) {
		status = "ready"
	} else if ci != nil && !ci.LastHeartbeat.IsZero() {
		status = "stale"
	}
	nodeGroups, hasGPU := clusterNodeGroups(ci, st)
	view := clusterView{
		ID:         ci.ID,
		Name:       clusterDisplayName(ci),
		ProjectID:  clusterProject(ci),
		Region:     ci.Region,
		Provider:   ci.Provider,
		K8sVersion: clusterVersion(ci),
		HasGPU:     hasGPU,
		NodeGroups: nodeGroups,
		Status:     status,
	}
	if !ci.LastHeartbeat.IsZero() {
		view.LastHeartbeat = ci.LastHeartbeat.UTC().Format(time.RFC3339)
	}
	return view
}

func clusterDisplayName(ci *store.ClusterInfo) string {
	if ci == nil {
		return ""
	}
	for _, key := range []string{"aegis.yourorg.dev/clusterName", "clusterName", "name", "displayName"} {
		if v := strings.TrimSpace(ci.Labels[key]); v != "" {
			return v
		}
	}
	return ci.ID
}

func clusterProject(ci *store.ClusterInfo) string {
	if ci == nil {
		return ""
	}
	for k, v := range ci.Labels {
		key := strings.ToLower(strings.TrimSpace(k))
		switch key {
		case "aegis.yourorg.dev/projectid", "projectid", "project_id", "project":
			if trimmed := strings.TrimSpace(v); trimmed != "" {
				return trimmed
			}
		}
	}
	if parts := strings.Split(strings.TrimSpace(ci.ID), "-"); len(parts) > 1 && parts[0] != "" {
		return parts[0]
	}
	return ""
}

func clusterVersion(ci *store.ClusterInfo) string {
	if ci == nil {
		return ""
	}
	for key, v := range ci.Labels {
		lower := strings.ToLower(key)
		lower = strings.ReplaceAll(lower, "_", "")
		lower = strings.ReplaceAll(lower, ".", "")
		if strings.Contains(lower, "k8sversion") || lower == "kubernetesversion" || lower == "version" {
			if trimmed := strings.TrimSpace(v); trimmed != "" {
				return trimmed
			}
		}
	}
	return ""
}

func clusterNodeGroups(ci *store.ClusterInfo, st store.Store) ([]nodeGroupView, bool) {
	if ci == nil || len(ci.AvailableFlavorSet) == 0 {
		return nil, false
	}
	names := make([]string, 0, len(ci.AvailableFlavorSet))
	for name, ok := range ci.AvailableFlavorSet {
		if ok {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	hasGPU := false
	groups := make([]nodeGroupView, 0, len(names))
	for _, name := range names {
		flavor := st.GetFlavor(name)
		if flavor == nil {
			flavor = defaultFlavorForName(name)
		}
		gpu := flavor != nil && flavor.GetGpuCount() > 0
		if gpu {
			hasGPU = true
		}
		groups = append(groups, nodeGroupView{Name: name, GPU: gpu})
	}
	return groups, hasGPU
}

func writeWizardError(w http.ResponseWriter, err error) {
	if err == nil {
		return
	}
	code := httpStatusFromErr(err)
	msg := err.Error()
	writeJSON(w, code, map[string]string{"error": msg})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func httpStatusFromErr(err error) int {
	switch status.Code(err) {
	case codes.InvalidArgument:
		return http.StatusBadRequest
	case codes.NotFound:
		return http.StatusNotFound
	case codes.PermissionDenied:
		return http.StatusForbidden
	case codes.Unauthenticated:
		return http.StatusUnauthorized
	case codes.AlreadyExists:
		return http.StatusConflict
	case codes.ResourceExhausted:
		return http.StatusTooManyRequests
	case codes.FailedPrecondition:
		return http.StatusPreconditionFailed
	default:
		return http.StatusInternalServerError
	}
}

func clusterReady(info *store.ClusterInfo, now time.Time) bool {
	if info == nil || info.LastHeartbeat.IsZero() {
		return false
	}
	return now.Sub(info.LastHeartbeat) <= heartbeatTTL
}

func selectFlavorForCluster(st store.Store, info *store.ClusterInfo, req workspaceCreateRequest) string {
	if trimmed := strings.TrimSpace(req.Flavor); trimmed != "" {
		return trimmed
	}
	if info == nil || len(info.AvailableFlavorSet) == 0 {
		return ""
	}
	var gpuFlavors, otherFlavors []string
	for name, ok := range info.AvailableFlavorSet {
		if !ok || strings.TrimSpace(name) == "" {
			continue
		}
		flavor := st.GetFlavor(name)
		if flavor == nil {
			flavor = defaultFlavorForName(name)
		}
		if flavor != nil && flavor.GetGpuCount() > 0 {
			gpuFlavors = append(gpuFlavors, name)
		} else {
			otherFlavors = append(otherFlavors, name)
		}
	}
	sort.Strings(gpuFlavors)
	sort.Strings(otherFlavors)
	switch {
	case len(gpuFlavors) > 0:
		return gpuFlavors[0]
	case len(otherFlavors) > 0:
		return otherFlavors[0]
	default:
		return ""
	}
}

func buildWorkspaceID(projectID, name string) string {
	projectSlug := sanitizeKubeName(projectID)
	nameSlug := sanitizeKubeName(name)
	if nameSlug == "" {
		nameSlug = RandID()
	}
	base := fmt.Sprintf("ws-%s-%s", projectSlug, nameSlug)
	if len(base) > maxObjectNameLength {
		return base[:maxObjectNameLength]
	}
	return base
}

func (s *Server) workspaceNameExists(projectID, name string) bool {
	if strings.TrimSpace(projectID) == "" || strings.TrimSpace(name) == "" {
		return false
	}
	for _, w := range s.store.ListWorkloads(projectID) {
		if w == nil {
			continue
		}
		ws, ok := w.GetKind().(*aegis.Workload_Workspace)
		if !ok || ws.Workspace == nil || len(ws.Workspace.GetEnv()) == 0 {
			continue
		}
		if existing := strings.TrimSpace(ws.Workspace.Env["WORKSPACE_NAME"]); existing != "" && strings.EqualFold(existing, name) {
			return true
		}
	}
	return false
}

func (s *Server) clusterInfo(id string) *store.ClusterInfo {
	if strings.TrimSpace(id) == "" {
		return nil
	}
	for _, ci := range s.store.ListClusterInfos() {
		if ci != nil && strings.EqualFold(strings.TrimSpace(ci.ID), strings.TrimSpace(id)) {
			return ci
		}
	}
	return nil
}

func (s *Server) ensureWorkspaceQueue(projectID, queue, flavor string) {
	queue = strings.TrimSpace(queue)
	if queue == "" {
		return
	}
	q := s.store.GetQueue(queue)
	if q == nil {
		q = &aegis.Queue{
			Name:                      queue,
			ProjectId:                 projectID,
			AllowedFlavors:            []string{strings.TrimSpace(flavor)},
			DefaultMaxDurationSeconds: s.defaultMaxRuntimeSeconds(nil),
		}
	} else {
		if q.ProjectId == "" {
			q.ProjectId = projectID
		}
		if ensureQueueAllowsFlavor(q, flavor) {
			// allow mutation
		}
	}
	s.store.PutQueue(q)
}
