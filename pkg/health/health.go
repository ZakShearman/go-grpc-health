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

type HealthService interface {
	pb.HealthServer

	// AddProbe adds a new health probe to the service.
	AddProbe(service string, probe func(ctx context.Context) HealthStatus)
}

type healthServiceImpl struct {
	HealthService

	// probes map[serviceName]HealthProbe
	probes map[string]func(ctx context.Context) HealthStatus
}

func NewHealthService() HealthService {
	return &healthServiceImpl{
		probes: make(map[string]func(ctx context.Context) HealthStatus),
	}
}

func (s *healthServiceImpl) Check(ctx context.Context, req *pb.HealthCheckRequest) (*pb.HealthCheckResponse, error) {
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

func (s *healthServiceImpl) Watch(req *pb.HealthCheckRequest, stream pb.Health_WatchServer) error {
	return status.Errorf(codes.Unimplemented, "method Watch not implemented")
}

func (s *healthServiceImpl) AddProbe(service string, probe func(ctx context.Context) HealthStatus) {
	s.probes[service] = probe
}
