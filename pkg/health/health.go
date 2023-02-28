package health

import (
	"context"
	"google.golang.org/grpc/codes"
	pb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
)

type HealthStatus int32

const (
	HealthStatusUnknown HealthStatus = iota
	HealthStatusHealthy
	HealthStatusUnhealthy
)

type healthService struct {
	pb.HealthServer

	// probes map[serviceName]HealthProbe
	probes map[string]func(ctx context.Context) HealthStatus
}

func NewHealthService() pb.HealthServer {
	return &healthService{
		probes: make(map[string]func(ctx context.Context) HealthStatus),
	}
}

func (s *healthService) Check(ctx context.Context, req *pb.HealthCheckRequest) (*pb.HealthCheckResponse, error) {
	service := req.Service

	if service == "" {
		// loop all services
		for _, probe := range s.probes {
			if probe(ctx) == HealthStatusUnhealthy {
				return &pb.HealthCheckResponse{Status: pb.HealthCheckResponse_NOT_SERVING}, nil
			}
		}
		return &pb.HealthCheckResponse{Status: pb.HealthCheckResponse_SERVING}, nil
	}

	// check named service
	if probe, ok := s.probes[service]; ok {
		if probe(ctx) == HealthStatusUnhealthy {
			return &pb.HealthCheckResponse{Status: pb.HealthCheckResponse_NOT_SERVING}, nil
		}
		return &pb.HealthCheckResponse{Status: pb.HealthCheckResponse_SERVING}, nil
	}

	return &pb.HealthCheckResponse{Status: pb.HealthCheckResponse_UNKNOWN}, nil
}

func (s *healthService) Watch(req *pb.HealthCheckRequest, stream pb.Health_WatchServer) error {
	return status.Errorf(codes.Unimplemented, "method Watch not implemented")
}

func (s *healthService) AddProbe(service string, probe func(ctx context.Context) HealthStatus) {
	s.probes[service] = probe
}
