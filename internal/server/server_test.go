package server

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	pb "github.com/BurnDevice/BurnDevice/burndevice/v1"
	"github.com/BurnDevice/BurnDevice/internal/config"
	"github.com/sirupsen/logrus"
)

func TestMain(m *testing.M) {
	// 设置测试环境
	logrus.SetLevel(logrus.FatalLevel) // 减少测试期间的日志输出
	code := m.Run()
	os.Exit(code)
}

func TestNew(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			Host: "localhost",
			Port: 8080,
		},
		AI: config.AIConfig{
			APIKey: "test-key",
		},
		Security: config.SecurityConfig{
			AuditLog: true,
		},
	}

	server, err := New(cfg)
	if err != nil {
		t.Fatalf("Expected no error creating server, got: %v", err)
	}

	if server == nil {
		t.Fatal("Expected server to be created")
	}

	if server.config != cfg {
		t.Error("Expected server config to be set")
	}

	if server.grpcServer == nil {
		t.Error("Expected gRPC server to be initialized")
	}

	if server.engine == nil {
		t.Error("Expected destruction engine to be initialized")
	}

	if server.aiClient == nil {
		t.Error("Expected AI client to be initialized")
	}

	if server.sysInfo == nil {
		t.Error("Expected system info to be initialized")
	}

	if server.logger == nil {
		t.Error("Expected logger to be initialized")
	}
}

func TestExecuteDestruction(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			Host: "localhost",
			Port: 8080,
		},
		AI: config.AIConfig{
			APIKey: "test-key",
		},
		Security: config.SecurityConfig{
			AuditLog:            true,
			MaxSeverity:         "HIGH",
			AllowedTargets:      []string{"/tmp"},
			BlockedTargets:      []string{"/etc", "/var"},
			RequireConfirmation: true,
		},
	}

	server, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	ctx := context.Background()

	// Test valid request
	req := &pb.ExecuteDestructionRequest{
		Type:               pb.DestructionType_DESTRUCTION_TYPE_FILE_DELETION,
		Targets:            []string{"/tmp/test.txt"},
		Severity:           pb.DestructionSeverity_DESTRUCTION_SEVERITY_LOW,
		ConfirmDestruction: true,
	}

	resp, err := server.ExecuteDestruction(ctx, req)
	if err != nil {
		t.Fatalf("Expected no error executing destruction, got: %v", err)
	}

	if resp == nil {
		t.Fatal("Expected response to be returned")
	}

	// Test invalid request (no confirmation)
	req.ConfirmDestruction = false
	resp, err = server.ExecuteDestruction(ctx, req)
	if err != nil {
		t.Fatalf("Expected no error (validation should return response), got: %v", err)
	}

	if resp.Success {
		t.Error("Expected request without confirmation to fail")
	}
}

func TestGetSystemInfo(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			Host: "localhost",
			Port: 8080,
		},
		AI: config.AIConfig{
			APIKey: "test-key",
		},
	}

	server, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	ctx := context.Background()
	req := &pb.GetSystemInfoRequest{}

	resp, err := server.GetSystemInfo(ctx, req)
	if err != nil {
		t.Fatalf("Expected no error getting system info, got: %v", err)
	}

	if resp == nil {
		t.Fatal("Expected response to be returned")
	}

	if resp.Os == "" {
		t.Error("Expected OS to be set")
	}

	if resp.Architecture == "" {
		t.Error("Expected Architecture to be set")
	}

	if resp.Hostname == "" {
		t.Error("Expected Hostname to be set")
	}

	if resp.Resources == nil {
		t.Error("Expected Resources to be set")
	}
}

func TestGenerateAttackScenario(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			Host: "localhost",
			Port: 8080,
		},
		AI: config.AIConfig{
			APIKey: "test-key",
		},
		Security: config.SecurityConfig{
			AuditLog: true,
		},
	}

	server, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	ctx := context.Background()

	// Test valid request
	req := &pb.GenerateAttackScenarioRequest{
		TargetDescription: "Test environment with temporary files",
		MaxSeverity:       pb.DestructionSeverity_DESTRUCTION_SEVERITY_LOW,
		AiModel:           "deepseek-chat",
	}

	// Note: This will likely fail due to AI API call, but we test the validation
	resp, err := server.GenerateAttackScenario(ctx, req)

	// We expect either success or an API-related error
	if err != nil {
		// Check if it's a validation error (should not happen with valid request)
		if strings.Contains(err.Error(), "target description is required") {
			t.Error("Unexpected validation error with valid request")
		}
		// API errors are expected in test environment
	} else if resp != nil {
		// If successful, verify response structure
		if resp.ScenarioId == "" {
			t.Error("Expected scenario ID to be set")
		}
	}

	// Test invalid request (empty target description)
	req.TargetDescription = ""
	resp, err = server.GenerateAttackScenario(ctx, req)
	if err == nil {
		t.Error("Expected error with empty target description")
	}

	if !strings.Contains(err.Error(), "target description is required") {
		t.Errorf("Expected validation error message, got: %v", err)
	}
}

func TestGenerateAttackScenarioWithoutAPIKey(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			Host: "localhost",
			Port: 8080,
		},
		AI: config.AIConfig{
			APIKey: "", // No API key
		},
	}

	server, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	ctx := context.Background()
	req := &pb.GenerateAttackScenarioRequest{
		TargetDescription: "Test environment",
		MaxSeverity:       pb.DestructionSeverity_DESTRUCTION_SEVERITY_LOW,
	}

	resp, err := server.GenerateAttackScenario(ctx, req)
	if err == nil {
		t.Error("Expected error when API key is not configured")
	}

	if !strings.Contains(err.Error(), "AI API key not configured") {
		t.Errorf("Expected API key error message, got: %v", err)
	}

	if resp != nil {
		t.Error("Expected no response when API key is missing")
	}
}

func TestValidateDestructionRequest(t *testing.T) {
	cfg := &config.Config{
		Security: config.SecurityConfig{
			MaxSeverity:         "MEDIUM", // Only LOW and MEDIUM allowed
			AllowedTargets:      []string{"/tmp", "/var/tmp"},
			BlockedTargets:      []string{"/etc", "/var/log"},
			RequireConfirmation: true,
		},
	}

	server, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Test valid request
	req := &pb.ExecuteDestructionRequest{
		Type:               pb.DestructionType_DESTRUCTION_TYPE_FILE_DELETION,
		Targets:            []string{"/tmp/test.txt"},
		Severity:           pb.DestructionSeverity_DESTRUCTION_SEVERITY_LOW,
		ConfirmDestruction: true,
	}

	err = server.validateDestructionRequest(req)
	if err != nil {
		t.Errorf("Expected no error for valid request, got: %v", err)
	}

	// Test request without confirmation
	req.ConfirmDestruction = false
	err = server.validateDestructionRequest(req)
	if err == nil {
		t.Error("Expected error for request without confirmation")
	}

	// Test request with high severity (above limit)
	req.ConfirmDestruction = true
	req.Severity = pb.DestructionSeverity_DESTRUCTION_SEVERITY_HIGH
	err = server.validateDestructionRequest(req)
	if err == nil {
		t.Error("Expected error for severity above limit")
	}

	// Test request with blocked target
	req.Severity = pb.DestructionSeverity_DESTRUCTION_SEVERITY_LOW
	req.Targets = []string{"/etc/passwd"}
	err = server.validateDestructionRequest(req)
	if err == nil {
		t.Error("Expected error for blocked target")
	}
}

func TestValidateStreamDestructionRequest(t *testing.T) {
	cfg := &config.Config{
		Security: config.SecurityConfig{
			MaxSeverity:         "MEDIUM",
			AllowedTargets:      []string{"/tmp"},
			BlockedTargets:      []string{"/etc"},
			RequireConfirmation: true,
		},
	}

	server, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Test valid request
	req := &pb.StreamDestructionRequest{
		Type:               pb.DestructionType_DESTRUCTION_TYPE_FILE_DELETION,
		Targets:            []string{"/tmp/test.txt"},
		Severity:           pb.DestructionSeverity_DESTRUCTION_SEVERITY_LOW,
		ConfirmDestruction: true,
	}

	err = server.validateStreamDestructionRequest(req)
	if err != nil {
		t.Errorf("Expected no error for valid request, got: %v", err)
	}

	// Test request without confirmation
	req.ConfirmDestruction = false
	err = server.validateStreamDestructionRequest(req)
	if err == nil {
		t.Error("Expected error for request without confirmation")
	}
}

func TestGetSeverityLevel(t *testing.T) {
	server := &Server{}

	tests := []struct {
		severity string
		expected int32
	}{
		{"LOW", 1},
		{"MEDIUM", 2},
		{"HIGH", 3},
		{"CRITICAL", 4},
		{"INVALID", 1}, // Default to LOW for invalid input
		{"", 1},        // Default to LOW for empty input
	}

	for _, tt := range tests {
		t.Run(tt.severity, func(t *testing.T) {
			result := server.getSeverityLevel(tt.severity)
			if result != tt.expected {
				t.Errorf("Expected severity level %d for '%s', got %d", tt.expected, tt.severity, result)
			}
		})
	}
}

func TestIsBlockedTarget(t *testing.T) {
	cfg := &config.Config{
		Security: config.SecurityConfig{
			BlockedTargets: []string{"/etc", "/var/log", "/usr/bin"},
		},
	}

	server, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	tests := []struct {
		target   string
		expected bool
	}{
		{"/etc/passwd", true},
		{"/var/log/messages", true},
		{"/usr/bin/bash", true},
		{"/tmp/test.txt", false},
		{"/home/user/file.txt", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.target, func(t *testing.T) {
			result := server.isBlockedTarget(tt.target)
			if result != tt.expected {
				t.Errorf("Expected isBlocked %v for '%s', got %v", tt.expected, tt.target, result)
			}
		})
	}
}

func TestIsAllowedTarget(t *testing.T) {
	cfg := &config.Config{
		Security: config.SecurityConfig{
			AllowedTargets: []string{"/tmp", "/var/tmp", "/home/user"},
		},
	}

	server, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	tests := []struct {
		target   string
		expected bool
	}{
		{"/tmp/test.txt", true},
		{"/var/tmp/file.log", true},
		{"/home/user/document.txt", true},
		{"/etc/passwd", false},
		{"/usr/bin/bash", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.target, func(t *testing.T) {
			result := server.isAllowedTarget(tt.target)
			if result != tt.expected {
				t.Errorf("Expected isAllowed %v for '%s', got %v", tt.expected, tt.target, result)
			}
		})
	}
}

func TestAuditLog(t *testing.T) {
	cfg := &config.Config{
		Security: config.SecurityConfig{
			AuditLog: true,
		},
	}

	server, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Test audit logging (should not panic)
	details := map[string]interface{}{
		"action": "test",
		"user":   "test-user",
	}

	// This should not panic or error
	server.auditLog("TEST_ACTION", details)
}

func TestGetHostname(t *testing.T) {
	hostname := getHostname()
	if hostname == "" {
		t.Error("Expected hostname to be returned")
	}

	// Hostname should not contain invalid characters
	if strings.ContainsAny(hostname, " \t\n\r") {
		t.Error("Hostname should not contain whitespace characters")
	}
}

func TestServerStartAndStop(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			Host: "localhost",
			Port: 0, // Use random available port
		},
		AI: config.AIConfig{
			APIKey: "test-key",
		},
	}

	server, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Test server start and immediate stop
	ctx, cancel := context.WithCancel(context.Background())

	// Start server in goroutine
	errChan := make(chan error, 1)
	go func() {
		err := server.Start(ctx)
		errChan <- err
	}()

	// Give server a moment to start
	time.Sleep(100 * time.Millisecond)

	// Cancel context to stop server
	cancel()

	// Wait for server to stop
	select {
	case err := <-errChan:
		if err != nil {
			t.Errorf("Expected no error from server start/stop, got: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Error("Server did not stop within timeout")
	}
}

func TestComplexValidationScenarios(t *testing.T) {
	cfg := &config.Config{
		Security: config.SecurityConfig{
			MaxSeverity:         "MEDIUM",
			AllowedTargets:      []string{"/tmp"},
			BlockedTargets:      []string{"/etc", "/tmp/blocked"},
			RequireConfirmation: true,
		},
	}

	server, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Test target that is both allowed and blocked (blocked should take precedence)
	req := &pb.ExecuteDestructionRequest{
		Type:               pb.DestructionType_DESTRUCTION_TYPE_FILE_DELETION,
		Targets:            []string{"/tmp/blocked/file.txt"},
		Severity:           pb.DestructionSeverity_DESTRUCTION_SEVERITY_LOW,
		ConfirmDestruction: true,
	}

	err = server.validateDestructionRequest(req)
	if err == nil {
		t.Error("Expected error for target that is blocked despite being in allowed path")
	}

	// Test multiple targets with mixed validity
	req.Targets = []string{"/tmp/valid.txt", "/etc/passwd"}
	err = server.validateDestructionRequest(req)
	if err == nil {
		t.Error("Expected error when any target is blocked")
	}

	// Test empty targets
	req.Targets = []string{}
	err = server.validateDestructionRequest(req)
	// This should be handled by the destruction engine, not validation
	// So we don't expect a validation error here
}

func TestServerWithMinimalConfig(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			Host: "localhost",
			Port: 8080,
		},
		// Minimal config - no AI, no security settings
	}

	server, err := New(cfg)
	if err != nil {
		t.Fatalf("Expected server to be created with minimal config, got: %v", err)
	}

	if server == nil {
		t.Fatal("Expected server to be created")
	}

	// Test that server can handle requests even with minimal config
	ctx := context.Background()
	sysInfoReq := &pb.GetSystemInfoRequest{}

	resp, err := server.GetSystemInfo(ctx, sysInfoReq)
	if err != nil {
		t.Fatalf("Expected system info to work with minimal config, got: %v", err)
	}

	if resp == nil {
		t.Error("Expected response even with minimal config")
	}
}
