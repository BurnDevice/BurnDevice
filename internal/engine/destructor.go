package engine

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/BurnDevice/BurnDevice/burndevice/v1"
	"github.com/BurnDevice/BurnDevice/internal/config"
)

// DestructionEngine handles the execution of destructive operations
type DestructionEngine struct {
	config  *config.Config
	logger  *logrus.Logger
	mu      sync.RWMutex
	running map[string]*DestructionTask
	eventCh chan *pb.StreamDestructionResponse
}

// DestructionTask represents a running destruction task
type DestructionTask struct {
	ID       string
	Type     pb.DestructionType
	Targets  []string
	Severity pb.DestructionSeverity
	Confirm  bool
	Context  context.Context
	Cancel   context.CancelFunc
	Progress float64
	Status   string
	Results  []*pb.DestructionResult
}

// NewDestructionEngine creates a new destruction engine
func NewDestructionEngine(cfg *config.Config) *DestructionEngine {
	return &DestructionEngine{
		config:  cfg,
		logger:  logrus.New(),
		running: make(map[string]*DestructionTask),
		eventCh: make(chan *pb.StreamDestructionResponse, 1000),
	}
}

// ExecuteDestruction executes a destruction request
func (e *DestructionEngine) ExecuteDestruction(ctx context.Context, req *pb.ExecuteDestructionRequest) (*pb.ExecuteDestructionResponse, error) {
	e.logger.WithFields(logrus.Fields{
		"type":     req.Type.String(),
		"targets":  req.Targets,
		"severity": req.Severity.String(),
	}).Warn("🔥 Executing destruction request")

	// Security checks
	if err := e.validateExecuteRequest(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Create task
	taskCtx, cancel := context.WithCancel(ctx)
	task := &DestructionTask{
		ID:       generateTaskID(),
		Type:     req.Type,
		Targets:  req.Targets,
		Severity: req.Severity,
		Confirm:  req.ConfirmDestruction,
		Context:  taskCtx,
		Cancel:   cancel,
		Status:   "running",
		Results:  make([]*pb.DestructionResult, 0),
	}

	// Register task
	e.mu.Lock()
	e.running[task.ID] = task
	e.mu.Unlock()

	defer func() {
		e.mu.Lock()
		delete(e.running, task.ID)
		e.mu.Unlock()
	}()

	// Execute based on type
	var results []*pb.DestructionResult
	var err error

	switch req.Type {
	case pb.DestructionType_DESTRUCTION_TYPE_FILE_DELETION:
		results, err = e.executeFileDeletion(task)
	default:
		results, err = e.executeBasicDestruction(task)
	}

	response := &pb.ExecuteDestructionResponse{
		Success: err == nil,
		Results: results,
	}

	if err != nil {
		response.Message = err.Error()
		e.logger.WithError(err).Error("Destruction execution failed")
	} else {
		response.Message = "Destruction completed successfully"
		e.logger.Info("Destruction execution completed")
	}

	return response, nil
}

// StreamDestruction executes destruction with real-time streaming
func (e *DestructionEngine) StreamDestruction(ctx context.Context, req *pb.StreamDestructionRequest, stream pb.BurnDeviceService_StreamDestructionServer) error {
	e.logger.WithFields(logrus.Fields{
		"type":     req.Type.String(),
		"targets":  req.Targets,
		"severity": req.Severity.String(),
	}).Warn("🔥 Starting streaming destruction")

	// Security checks
	if err := e.validateStreamRequest(req); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Create task
	taskCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	task := &DestructionTask{
		ID:       generateTaskID(),
		Type:     req.Type,
		Targets:  req.Targets,
		Severity: req.Severity,
		Confirm:  req.ConfirmDestruction,
		Context:  taskCtx,
		Cancel:   cancel,
		Status:   "running",
		Results:  make([]*pb.DestructionResult, 0),
	}

	// Send start event
	startEvent := &pb.StreamDestructionResponse{
		Timestamp: timestamppb.New(time.Now()),
		Type:      pb.DestructionEventType_DESTRUCTION_EVENT_TYPE_STARTED,
		Message:   "Destruction task started",
		Progress:  0.0,
	}
	if err := stream.Send(startEvent); err != nil {
		return err
	}

	// Execute destruction with progress streaming
	var results []*pb.DestructionResult
	var err error

	switch req.Type {
	case pb.DestructionType_DESTRUCTION_TYPE_FILE_DELETION:
		results, err = e.executeFileDeletionStreaming(task, stream)
	default:
		results, err = e.executeBasicDestruction(task)
	}

	// Send completion or error event
	var finalEvent *pb.StreamDestructionResponse
	if err != nil {
		finalEvent = &pb.StreamDestructionResponse{
			Timestamp: timestamppb.New(time.Now()),
			Type:      pb.DestructionEventType_DESTRUCTION_EVENT_TYPE_ERROR,
			Message:   fmt.Sprintf("Destruction failed: %s", err.Error()),
			Progress:  1.0,
		}
	} else {
		finalEvent = &pb.StreamDestructionResponse{
			Timestamp: timestamppb.New(time.Now()),
			Type:      pb.DestructionEventType_DESTRUCTION_EVENT_TYPE_COMPLETED,
			Message:   fmt.Sprintf("Destruction completed successfully. %d targets processed.", len(results)),
			Progress:  1.0,
		}
	}

	return stream.Send(finalEvent)
}

// executeFileDeletion performs file deletion attacks
func (e *DestructionEngine) executeFileDeletion(task *DestructionTask) ([]*pb.DestructionResult, error) {
	var results []*pb.DestructionResult

	for _, target := range task.Targets {
		result := &pb.DestructionResult{
			Target:  target,
			Metrics: &pb.DestructionMetrics{},
		}

		start := time.Now()

		// Check if target is blocked
		if e.isBlockedTarget(target) {
			result.Success = false
			result.ErrorMessage = "Target is in blocked list"
			results = append(results, result)
			continue
		}

		// Perform deletion based on severity (simplified)
		var err error
		switch task.Severity {
		case pb.DestructionSeverity_DESTRUCTION_SEVERITY_LOW:
			err = e.safeDeletion(target, result.Metrics)
		default:
			err = e.safeDeletion(target, result.Metrics)
		}

		result.Success = err == nil
		if err != nil {
			result.ErrorMessage = err.Error()
		}
		result.Metrics.ExecutionTimeSeconds = time.Since(start).Seconds()
		results = append(results, result)
	}

	return results, nil
}

// executeFileDeletionStreaming performs file deletion with streaming updates
func (e *DestructionEngine) executeFileDeletionStreaming(task *DestructionTask, stream pb.BurnDeviceService_StreamDestructionServer) ([]*pb.DestructionResult, error) {
	var results []*pb.DestructionResult

	for i, target := range task.Targets {
		result := &pb.DestructionResult{
			Target:  target,
			Metrics: &pb.DestructionMetrics{},
		}

		start := time.Now()

		// Send progress event
		progress := float64(i) / float64(len(task.Targets))
		progressEvent := &pb.StreamDestructionResponse{
			Timestamp: timestamppb.New(time.Now()),
			Type:      pb.DestructionEventType_DESTRUCTION_EVENT_TYPE_PROGRESS,
			Target:    target,
			Progress:  progress,
			Message:   fmt.Sprintf("Processing target %d of %d: %s", i+1, len(task.Targets), target),
		}
		if err := stream.Send(progressEvent); err != nil {
			return results, err
		}

		// Check if target is blocked
		if e.isBlockedTarget(target) {
			result.Success = false
			result.ErrorMessage = "Target is in blocked list"
			results = append(results, result)
			continue
		}

		// Perform deletion
		err := e.safeDeletion(target, result.Metrics)
		result.Success = err == nil
		if err != nil {
			result.ErrorMessage = err.Error()
		}
		result.Metrics.ExecutionTimeSeconds = time.Since(start).Seconds()
		results = append(results, result)

		// Send completion event for this target
		targetCompleteEvent := &pb.StreamDestructionResponse{
			Timestamp: timestamppb.New(time.Now()),
			Type:      pb.DestructionEventType_DESTRUCTION_EVENT_TYPE_PROGRESS,
			Target:    target,
			Progress:  float64(i+1) / float64(len(task.Targets)),
			Message:   fmt.Sprintf("Target completed: %s (success: %v)", target, result.Success),
		}
		if err := stream.Send(targetCompleteEvent); err != nil {
			return results, err
		}
	}

	return results, nil
}

// executeBasicDestruction handles other destruction types
func (e *DestructionEngine) executeBasicDestruction(task *DestructionTask) ([]*pb.DestructionResult, error) {
	result := &pb.DestructionResult{
		Target:  strings.Join(task.Targets, ","),
		Success: true,
		Metrics: &pb.DestructionMetrics{
			ExecutionTimeSeconds: 1.0,
		},
	}

	e.logger.WithField("type", task.Type).Info("Basic destruction simulation completed")
	return []*pb.DestructionResult{result}, nil
}

// File operation helpers
func (e *DestructionEngine) safeDeletion(target string, metrics *pb.DestructionMetrics) error {
	// Get file info for metrics
	info, err := os.Stat(target)
	if err != nil {
		return fmt.Errorf("failed to stat file: %w", err)
	}

	if info.IsDir() {
		return fmt.Errorf("target is a directory, not supported in safe mode")
	}

	// Create backup before deletion
	backupPath := target + ".burndevice.backup"
	if err := e.copyFile(target, backupPath); err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}

	metrics.BytesDestroyed = info.Size()
	metrics.FilesDeleted = 1

	// Remove original file
	if err := os.Remove(target); err != nil {
		return fmt.Errorf("failed to remove file: %w", err)
	}

	e.logger.WithFields(logrus.Fields{
		"target": target,
		"backup": backupPath,
	}).Info("Safe deletion completed")

	return nil
}

// Validation helpers
func (e *DestructionEngine) validateExecuteRequest(req *pb.ExecuteDestructionRequest) error {
	if !req.ConfirmDestruction && e.config.Security.RequireConfirmation {
		return fmt.Errorf("destruction must be confirmed")
	}

	maxSeverity := e.getSeverityLevel(e.config.Security.MaxSeverity)
	if int32(req.Severity) > maxSeverity {
		return fmt.Errorf("requested severity exceeds maximum allowed (%s)", e.config.Security.MaxSeverity)
	}

	for _, target := range req.Targets {
		if e.isBlockedTarget(target) {
			return fmt.Errorf("target is blocked: %s", target)
		}

		if len(e.config.Security.AllowedTargets) > 0 && !e.isAllowedTarget(target) {
			return fmt.Errorf("target is not in allowed list: %s", target)
		}
	}

	return nil
}

func (e *DestructionEngine) validateStreamRequest(req *pb.StreamDestructionRequest) error {
	if !req.ConfirmDestruction && e.config.Security.RequireConfirmation {
		return fmt.Errorf("destruction must be confirmed")
	}

	maxSeverity := e.getSeverityLevel(e.config.Security.MaxSeverity)
	if int32(req.Severity) > maxSeverity {
		return fmt.Errorf("requested severity exceeds maximum allowed (%s)", e.config.Security.MaxSeverity)
	}

	for _, target := range req.Targets {
		if e.isBlockedTarget(target) {
			return fmt.Errorf("target is blocked: %s", target)
		}

		if len(e.config.Security.AllowedTargets) > 0 && !e.isAllowedTarget(target) {
			return fmt.Errorf("target is not in allowed list: %s", target)
		}
	}

	return nil
}

// Helper methods
func (e *DestructionEngine) isBlockedTarget(target string) bool {
	for _, blocked := range e.config.Security.BlockedTargets {
		if strings.HasPrefix(target, blocked) {
			return true
		}
	}
	return false
}

func (e *DestructionEngine) isAllowedTarget(target string) bool {
	for _, allowed := range e.config.Security.AllowedTargets {
		if strings.HasPrefix(target, allowed) {
			return true
		}
	}
	return false
}

func (e *DestructionEngine) getSeverityLevel(severity string) int32 {
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

func (e *DestructionEngine) copyFile(src, dst string) error {
	// Validate and clean file paths to prevent directory traversal
	cleanSrc := filepath.Clean(src)
	cleanDst := filepath.Clean(dst)

	// Check for directory traversal attempts
	if strings.Contains(cleanSrc, "..") || strings.Contains(cleanDst, "..") {
		return fmt.Errorf("path traversal detected in file paths")
	}

	// Ensure paths are absolute to avoid relative path issues
	absSrc, err := filepath.Abs(cleanSrc)
	if err != nil {
		return fmt.Errorf("failed to resolve source path: %w", err)
	}

	absDst, err := filepath.Abs(cleanDst)
	if err != nil {
		return fmt.Errorf("failed to resolve destination path: %w", err)
	}

	// Additional validation: ensure we're not accessing system critical paths
	if e.isBlockedTarget(absSrc) || e.isBlockedTarget(absDst) {
		return fmt.Errorf("access to blocked path is not allowed")
	}

	// Final security check: ensure paths are within allowed directories
	if len(e.config.Security.AllowedTargets) > 0 {
		if !e.isAllowedTarget(absSrc) || !e.isAllowedTarget(absDst) {
			return fmt.Errorf("paths are not within allowed target directories")
		}
	}

	// #nosec G304 - Path is validated and sanitized above
	sourceFile, err := os.Open(absSrc)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer sourceFile.Close()

	// #nosec G304 - Path is validated and sanitized above
	destFile, err := os.Create(absDst)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return fmt.Errorf("failed to copy file content: %w", err)
	}

	return nil
}

func generateTaskID() string {
	return fmt.Sprintf("task_%d", time.Now().UnixNano())
}
