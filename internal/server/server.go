package server

import (
	"context"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	pb "github.com/BurnDevice/BurnDevice/burndevice/v1"
	"github.com/BurnDevice/BurnDevice/internal/ai"
	"github.com/BurnDevice/BurnDevice/internal/config"
	"github.com/BurnDevice/BurnDevice/internal/engine"
	"github.com/BurnDevice/BurnDevice/internal/system"
)

// Server represents the gRPC server
type Server struct {
	pb.UnimplementedBurnDeviceServiceServer

	config     *config.Config
	grpcServer *grpc.Server
	engine     *engine.DestructionEngine
	aiClient   *ai.DeepSeekClient
	sysInfo    *system.SystemInfo
	logger     *logrus.Logger
}

// New creates a new BurnDevice server
func New(cfg *config.Config) (*Server, error) {
	logger := logrus.New()

	// Create destruction engine
	destructionEngine := engine.NewDestructionEngine(cfg)

	// Create AI client
	aiClient := ai.NewDeepSeekClient(&cfg.AI)

	// Create system info collector
	sysInfo := system.NewSystemInfo()

	// Create gRPC server
	grpcServer := grpc.NewServer()

	server := &Server{
		config:     cfg,
		grpcServer: grpcServer,
		engine:     destructionEngine,
		aiClient:   aiClient,
		sysInfo:    sysInfo,
		logger:     logger,
	}

	// Register the service
	pb.RegisterBurnDeviceServiceServer(grpcServer, server)

	return server, nil
}

// Start starts the gRPC server
func (s *Server) Start(ctx context.Context) error {
	address := fmt.Sprintf("%s:%d", s.config.Server.Host, s.config.Server.Port)

	listener, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", address, err)
	}

	s.logger.WithFields(logrus.Fields{
		"address": address,
		"tls":     s.config.Server.TLS.Enabled,
	}).Info("🚀 Starting BurnDevice gRPC server")

	// Start server in goroutine
	errChan := make(chan error, 1)
	go func() {
		if err := s.grpcServer.Serve(listener); err != nil {
			errChan <- fmt.Errorf("gRPC server failed: %w", err)
		}
	}()

	// Wait for context cancellation or server error
	select {
	case <-ctx.Done():
		s.logger.Info("🛑 Shutting down server...")
		s.grpcServer.GracefulStop()
		return nil
	case err := <-errChan:
		return err
	}
}

// ExecuteDestruction implements the ExecuteDestruction RPC
func (s *Server) ExecuteDestruction(ctx context.Context, req *pb.ExecuteDestructionRequest) (*pb.ExecuteDestructionResponse, error) {
	s.logger.WithFields(logrus.Fields{
		"type":      req.Type.String(),
		"targets":   req.Targets,
		"severity":  req.Severity.String(),
		"confirmed": req.ConfirmDestruction,
	}).Warn("🔥 Received destruction request")

	// Security validation
	if err := s.validateDestructionRequest(req); err != nil {
		s.logger.WithError(err).Error("Destruction request validation failed")
		return &pb.ExecuteDestructionResponse{
			Success: false,
			Message: fmt.Sprintf("Validation failed: %s", err.Error()),
		}, nil
	}

	// Execute destruction
	response, err := s.engine.ExecuteDestruction(ctx, req)
	if err != nil {
		s.logger.WithError(err).Error("Destruction execution failed")
		return &pb.ExecuteDestructionResponse{
			Success: false,
			Message: fmt.Sprintf("Execution failed: %s", err.Error()),
		}, nil
	}

	// Audit logging
	if s.config.Security.AuditLog {
		s.auditLog("DESTRUCTION_EXECUTED", map[string]interface{}{
			"type":     req.Type.String(),
			"targets":  req.Targets,
			"severity": req.Severity.String(),
			"success":  response.Success,
		})
	}

	return response, nil
}

// GetSystemInfo implements the GetSystemInfo RPC
func (s *Server) GetSystemInfo(ctx context.Context, req *pb.GetSystemInfoRequest) (*pb.GetSystemInfoResponse, error) {
	s.logger.Info("📊 Collecting system information")

	info, err := s.sysInfo.Collect()
	if err != nil {
		return nil, fmt.Errorf("failed to collect system info: %w", err)
	}

	return &pb.GetSystemInfoResponse{
		Os:              info.OS,
		Architecture:    info.Architecture,
		Hostname:        info.Hostname,
		CriticalPaths:   info.CriticalPaths,
		RunningServices: info.RunningServices,
		Resources: &pb.SystemResources{
			TotalMemory:     info.Resources.TotalMemory,
			AvailableMemory: info.Resources.AvailableMemory,
			TotalDisk:       info.Resources.TotalDisk,
			AvailableDisk:   info.Resources.AvailableDisk,
			CpuUsage:        info.Resources.CPUUsage,
		},
	}, nil
}

// GenerateAttackScenario implements the GenerateAttackScenario RPC
func (s *Server) GenerateAttackScenario(ctx context.Context, req *pb.GenerateAttackScenarioRequest) (*pb.GenerateAttackScenarioResponse, error) {
	s.logger.WithFields(logrus.Fields{
		"target":       req.TargetDescription,
		"max_severity": req.MaxSeverity.String(),
		"model":        req.AiModel,
	}).Info("🤖 Generating AI attack scenario")

	// Validate request
	if req.TargetDescription == "" {
		return nil, fmt.Errorf("target description is required")
	}

	// Check if AI is properly configured
	if s.config.AI.APIKey == "" {
		return nil, fmt.Errorf("AI API key not configured")
	}

	// Generate scenario using AI
	response, err := s.aiClient.GenerateAttackScenario(ctx, req)
	if err != nil {
		s.logger.WithError(err).Error("AI scenario generation failed")
		return nil, fmt.Errorf("scenario generation failed: %w", err)
	}

	// Audit logging
	if s.config.Security.AuditLog {
		s.auditLog("AI_SCENARIO_GENERATED", map[string]interface{}{
			"scenario_id":        response.ScenarioId,
			"target":             req.TargetDescription,
			"estimated_severity": response.EstimatedSeverity.String(),
			"steps_count":        len(response.Steps),
		})
	}

	return response, nil
}

// StreamDestruction implements the StreamDestruction RPC
func (s *Server) StreamDestruction(req *pb.StreamDestructionRequest, stream pb.BurnDeviceService_StreamDestructionServer) error {
	s.logger.WithFields(logrus.Fields{
		"type":     req.Type.String(),
		"targets":  req.Targets,
		"severity": req.Severity.String(),
	}).Warn("🔥 Starting streaming destruction")

	// Security validation
	if err := s.validateStreamDestructionRequest(req); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Execute destruction with streaming
	return s.engine.StreamDestruction(stream.Context(), req, stream)
}

// Validation helpers
func (s *Server) validateDestructionRequest(req *pb.ExecuteDestructionRequest) error {
	// Check confirmation requirement
	if s.config.Security.RequireConfirmation && !req.ConfirmDestruction {
		return fmt.Errorf("destruction must be confirmed")
	}

	// Check severity limits
	maxSeverity := s.getSeverityLevel(s.config.Security.MaxSeverity)
	if int32(req.Severity) > maxSeverity {
		return fmt.Errorf("requested severity exceeds maximum allowed (%s)", s.config.Security.MaxSeverity)
	}

	// Check target restrictions
	for _, target := range req.Targets {
		if s.isBlockedTarget(target) {
			return fmt.Errorf("target is blocked: %s", target)
		}

		if len(s.config.Security.AllowedTargets) > 0 && !s.isAllowedTarget(target) {
			return fmt.Errorf("target is not in allowed list: %s", target)
		}
	}

	return nil
}

func (s *Server) validateStreamDestructionRequest(req *pb.StreamDestructionRequest) error {
	// Check confirmation requirement
	if s.config.Security.RequireConfirmation && !req.ConfirmDestruction {
		return fmt.Errorf("destruction must be confirmed")
	}

	// Check severity limits
	maxSeverity := s.getSeverityLevel(s.config.Security.MaxSeverity)
	if int32(req.Severity) > maxSeverity {
		return fmt.Errorf("requested severity exceeds maximum allowed (%s)", s.config.Security.MaxSeverity)
	}

	// Check target restrictions
	for _, target := range req.Targets {
		if s.isBlockedTarget(target) {
			return fmt.Errorf("target is blocked: %s", target)
		}

		if len(s.config.Security.AllowedTargets) > 0 && !s.isAllowedTarget(target) {
			return fmt.Errorf("target is not in allowed list: %s", target)
		}
	}

	return nil
}

func (s *Server) getSeverityLevel(severity string) int32 {
	switch severity {
	case "LOW":
		return int32(pb.DestructionSeverity_DESTRUCTION_SEVERITY_LOW)
	case "MEDIUM":
		return int32(pb.DestructionSeverity_DESTRUCTION_SEVERITY_MEDIUM)
	case "HIGH":
		return int32(pb.DestructionSeverity_DESTRUCTION_SEVERITY_HIGH)
	case "CRITICAL":
		return int32(pb.DestructionSeverity_DESTRUCTION_SEVERITY_CRITICAL)
	default:
		return int32(pb.DestructionSeverity_DESTRUCTION_SEVERITY_LOW)
	}
}

func (s *Server) isBlockedTarget(target string) bool {
	for _, blocked := range s.config.Security.BlockedTargets {
		if target == blocked || (len(target) > len(blocked) && target[:len(blocked)] == blocked) {
			return true
		}
	}
	return false
}

func (s *Server) isAllowedTarget(target string) bool {
	for _, allowed := range s.config.Security.AllowedTargets {
		if target == allowed || (len(target) > len(allowed) && target[:len(allowed)] == allowed) {
			return true
		}
	}
	return false
}

func (s *Server) auditLog(action string, details map[string]interface{}) {
	logEntry := s.logger.WithFields(logrus.Fields{
		"action":    action,
		"timestamp": time.Now().Format(time.RFC3339),
		"hostname":  getHostname(),
		"user":      os.Getenv("USER"),
	})

	for key, value := range details {
		logEntry = logEntry.WithField(key, value)
	}

	logEntry.Info("🔍 Audit log entry")
}

func getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return hostname
}
