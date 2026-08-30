package healthgrpc

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/devctllabs/go-libs/health"
	"google.golang.org/grpc/codes"
	grpc_health_v1 "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
)

const (
	// LivenessService is the standard service name used for liveness checks.
	LivenessService = "liveness"
	// ReadinessService is the standard service name used for readiness checks.
	ReadinessService = "readiness"
)

// Config controls periodic health evaluation.
type Config struct {
	PollInterval time.Duration
}

// Server implements the standard gRPC Health service for one probe set.
type Server struct {
	grpc_health_v1.UnimplementedHealthServer
	probes       *health.Probes
	pollInterval time.Duration

	mu       sync.Mutex
	statuses map[string]grpc_health_v1.HealthCheckResponse_ServingStatus
	watchers map[string]map[chan grpc_health_v1.HealthCheckResponse_ServingStatus]struct{}
	stopped  bool
	shutdown chan struct{}
	stopOnce sync.Once
}

// New constructs a health service in the NOT_SERVING state.
func New(config Config, probes *health.Probes) (*Server, error) {
	if probes == nil {
		return nil, errors.New("healthgrpc: Probes must not be nil")
	}
	if config.PollInterval < 0 {
		return nil, errors.New("healthgrpc: poll interval must not be negative")
	}
	if config.PollInterval == 0 {
		config.PollInterval = 5 * time.Second
	}
	return &Server{
		probes:       probes,
		pollInterval: config.PollInterval,
		statuses: map[string]grpc_health_v1.HealthCheckResponse_ServingStatus{
			"":               grpc_health_v1.HealthCheckResponse_NOT_SERVING,
			LivenessService:  grpc_health_v1.HealthCheckResponse_NOT_SERVING,
			ReadinessService: grpc_health_v1.HealthCheckResponse_NOT_SERVING,
		},
		watchers: make(map[string]map[chan grpc_health_v1.HealthCheckResponse_ServingStatus]struct{}),
		shutdown: make(chan struct{}),
	}, nil
}

// Check evaluates the requested probe from fresh state.
func (s *Server) Check(ctx context.Context, request *grpc_health_v1.HealthCheckRequest) (*grpc_health_v1.HealthCheckResponse, error) {
	var report health.Report
	switch request.GetService() {
	case LivenessService:
		report = s.probes.Liveness()
	case "", ReadinessService:
		report = s.probes.Readiness(ctx)
	default:
		return nil, status.Error(codes.NotFound, "unknown health service")
	}
	return &grpc_health_v1.HealthCheckResponse{Status: reportStatus(report)}, nil
}

// List returns the latest published status for all known services.
func (s *Server) List(context.Context, *grpc_health_v1.HealthListRequest) (*grpc_health_v1.HealthListResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	statuses := make(map[string]*grpc_health_v1.HealthCheckResponse, len(s.statuses))
	for service, servingStatus := range s.statuses {
		statuses[service] = &grpc_health_v1.HealthCheckResponse{Status: servingStatus}
	}
	return &grpc_health_v1.HealthListResponse{Statuses: statuses}, nil
}

// Watch sends the current status immediately and subsequent status changes.
func (s *Server) Watch(request *grpc_health_v1.HealthCheckRequest, stream grpc_health_v1.Health_WatchServer) error {
	updates := make(chan grpc_health_v1.HealthCheckResponse_ServingStatus, 1)
	service := request.GetService()
	s.mu.Lock()
	registered := !s.stopped
	if registered {
		if s.watchers[service] == nil {
			s.watchers[service] = make(map[chan grpc_health_v1.HealthCheckResponse_ServingStatus]struct{})
		}
		s.watchers[service][updates] = struct{}{}
	}
	initial, exists := s.statuses[service]
	if !exists {
		initial = grpc_health_v1.HealthCheckResponse_SERVICE_UNKNOWN
	}
	updates <- initial
	if !registered {
		close(updates)
	}
	s.mu.Unlock()
	if registered {
		defer s.removeWatcher(service, updates)
	}

	lastSent := grpc_health_v1.HealthCheckResponse_ServingStatus(-1)
	for {
		select {
		case servingStatus, open := <-updates:
			if !open {
				return status.Error(codes.Canceled, "health service stopped")
			}
			if servingStatus == lastSent {
				continue
			}
			if err := stream.Send(&grpc_health_v1.HealthCheckResponse{Status: servingStatus}); err != nil {
				return status.Error(codes.Canceled, "health watch ended")
			}
			lastSent = servingStatus
		case <-stream.Context().Done():
			return status.Error(codes.Canceled, "health watch ended")
		}
	}
}

// Run evaluates probes immediately and then at the configured interval until
// ctx is canceled or Shutdown is called.
func (s *Server) Run(ctx context.Context) error {
	s.publishCurrent(ctx)
	ticker := time.NewTicker(s.pollInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			s.publishCurrent(ctx)
		case <-ctx.Done():
			s.stop()
			return nil
		case <-s.shutdown:
			return nil
		}
	}
}

// Shutdown marks all known services NOT_SERVING and terminates watchers.
func (s *Server) Shutdown(context.Context) error {
	s.stop()
	return nil
}

func (s *Server) publishCurrent(ctx context.Context) {
	s.setStatus(LivenessService, reportStatus(s.probes.Liveness()))
	readiness := reportStatus(s.probes.Readiness(ctx))
	s.setStatus("", readiness)
	s.setStatus(ReadinessService, readiness)
}

func reportStatus(report health.Report) grpc_health_v1.HealthCheckResponse_ServingStatus {
	if report.Status == health.StatusOK {
		return grpc_health_v1.HealthCheckResponse_SERVING
	}
	return grpc_health_v1.HealthCheckResponse_NOT_SERVING
}

func (s *Server) setStatus(service string, servingStatus grpc_health_v1.HealthCheckResponse_ServingStatus) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.stopped {
		return
	}
	if s.statuses[service] == servingStatus {
		return
	}
	s.statuses[service] = servingStatus
	for updates := range s.watchers[service] {
		select {
		case <-updates:
		default:
		}
		updates <- servingStatus
	}
}

func (s *Server) stop() {
	s.stopOnce.Do(func() {
		s.mu.Lock()
		s.stopped = true
		for service := range s.statuses {
			s.statuses[service] = grpc_health_v1.HealthCheckResponse_NOT_SERVING
		}
		for service, watchers := range s.watchers {
			_, knownService := s.statuses[service]
			for updates := range watchers {
				if knownService {
					select {
					case <-updates:
					default:
					}
					updates <- grpc_health_v1.HealthCheckResponse_NOT_SERVING
				}
				close(updates)
			}
		}
		close(s.shutdown)
		s.mu.Unlock()
	})
}

func (s *Server) removeWatcher(service string, updates chan grpc_health_v1.HealthCheckResponse_ServingStatus) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.watchers[service], updates)
}
