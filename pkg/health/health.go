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

type HealthProbe interface {
	// Check returns the current health status of the service.
	Check(ctx context.Context) HealthStatus
}

type healthService struct {
	pb.HealthServer

	// probes map[serviceName]HealthProbe
	probes map[string]HealthProbe
}

func NewHealthService() pb.HealthServer {
	return &healthService{
		probes: make(map[string]HealthProbe),
	}
}

func (s *healthService) Check(ctx context.Context, req *pb.HealthCheckRequest) (*pb.HealthCheckResponse, error) {
	service := req.Service

	if service == "" {
		// loop all services
		for _, probe := range s.probes {
			if probe.Check(ctx) == HealthStatusUnhealthy {
				return &pb.HealthCheckResponse{Status: pb.HealthCheckResponse_NOT_SERVING}, nil
			}
		}
		return &pb.HealthCheckResponse{Status: pb.HealthCheckResponse_SERVING}, nil
	}

	// check named service
	if probe, ok := s.probes[service]; ok {
		if probe.Check(ctx) == HealthStatusUnhealthy {
			return &pb.HealthCheckResponse{Status: pb.HealthCheckResponse_NOT_SERVING}, nil
		}
		return &pb.HealthCheckResponse{Status: pb.HealthCheckResponse_SERVING}, nil
	}

	return &pb.HealthCheckResponse{Status: pb.HealthCheckResponse_UNKNOWN}, nil
}

func (s *healthService) Watch(req *pb.HealthCheckRequest, stream pb.Health_WatchServer) error {
	return status.Errorf(codes.Unimplemented, "method Watch not implemented")
}

func (s *healthService) AddProbe(service string, probe HealthProbe) {
	s.probes[service] = probe
}
