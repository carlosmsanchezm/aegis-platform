package server

import (
	"context"
	"strings"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	aegis "github.com/yourorg/aegis/proto/aegis/v1"
)

func (s *Server) ListClusterWorkloadIDs(ctx context.Context, req *aegis.ListClusterWorkloadIDsRequest) (*aegis.ListClusterWorkloadIDsResponse, error) {
	if req == nil {
		err := status.Error(codes.InvalidArgument, "request required")
		s.log.Warn("list cluster workload ids failed", zap.Error(err))
		return nil, err
	}

	clusterID := strings.TrimSpace(req.GetClusterId())
	if clusterID == "" {
		err := status.Error(codes.InvalidArgument, "cluster_id required")
		s.log.Warn("list cluster workload ids failed", zap.Error(err))
		return nil, err
	}
	if s.store == nil {
		err := status.Error(codes.Internal, "store unavailable")
		s.log.Error("list cluster workload ids failed", zap.Error(err))
		return nil, err
	}

	ids, err := s.store.ListClusterWorkloadIDs(clusterID)
	if err != nil {
		s.log.Error("list cluster workload ids query failed", zap.String("cluster_id", clusterID), zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to list cluster workload ids")
	}

	s.log.Debug("cluster workload ids listed", zap.String("cluster_id", clusterID), zap.Int("count", len(ids)))
	return &aegis.ListClusterWorkloadIDsResponse{WorkloadIds: ids}, nil
}

