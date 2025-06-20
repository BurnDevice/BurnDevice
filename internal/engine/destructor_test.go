package engine

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

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

func TestNewDestructionEngine(t *testing.T) {
	cfg := &config.Config{
		Security: config.SecurityConfig{
			MaxSeverity: "HIGH",
		},
	}

	engine := NewDestructionEngine(cfg)

	if engine == nil {
		t.Fatal("Expected engine to be created")
	}

	if engine.config != cfg {
		t.Error("Expected config to be set")
	}

	if engine.logger == nil {
		t.Error("Expected logger to be initialized")
	}

	if engine.running == nil {
		t.Error("Expected running tasks map to be initialized")
	}

	if engine.eventCh == nil {
		t.Error("Expected event channel to be initialized")
	}

	// Test initial state
	if len(engine.running) != 0 {
		t.Error("Expected no running tasks initially")
	}
}

func TestExecuteDestruction(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "burndevice_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test file
	testFile := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	cfg := &config.Config{
		Security: config.SecurityConfig{
			MaxSeverity:    "HIGH",
			AllowedTargets: []string{tempDir},
		},
	}

	engine := NewDestructionEngine(cfg)
	ctx := context.Background()

	// Test valid file deletion request
	req := &pb.ExecuteDestructionRequest{
		Type:               pb.DestructionType_DESTRUCTION_TYPE_FILE_DELETION,
		Targets:            []string{testFile},
		Severity:           pb.DestructionSeverity_DESTRUCTION_SEVERITY_LOW,
		ConfirmDestruction: true,
	}

	resp, err := engine.ExecuteDestruction(ctx, req)
	if err != nil {
		t.Fatalf("Expected no error executing destruction, got: %v", err)
	}

	if resp == nil {
		t.Fatal("Expected response to be returned")
	}

	if !resp.Success {
		t.Errorf("Expected successful execution, got: %s", resp.Message)
	}

	if len(resp.Results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(resp.Results))
	}

	// Verify file was deleted
	if _, err := os.Stat(testFile); !os.IsNotExist(err) {
		t.Error("Expected test file to be deleted")
	}

	// Verify backup was created
	backupFile := testFile + ".burndevice.backup"
	if _, err := os.Stat(backupFile); os.IsNotExist(err) {
		t.Error("Expected backup file to be created")
	}
}

func TestExecuteDestructionValidation(t *testing.T) {
	cfg := &config.Config{
		Security: config.SecurityConfig{
			MaxSeverity:         "MEDIUM", // Only LOW and MEDIUM
			AllowedTargets:      []string{"/tmp"},
			BlockedTargets:      []string{"/etc"},
			RequireConfirmation: true,
		},
	}

	engine := NewDestructionEngine(cfg)
	ctx := context.Background()

	// Test request without confirmation
	req := &pb.ExecuteDestructionRequest{
		Type:               pb.DestructionType_DESTRUCTION_TYPE_FILE_DELETION,
		Targets:            []string{"/tmp/test.txt"},
		Severity:           pb.DestructionSeverity_DESTRUCTION_SEVERITY_LOW,
		ConfirmDestruction: false,
	}

	var err error
	_, err = engine.ExecuteDestruction(ctx, req)
	if err == nil {
		t.Error("Expected error for request without confirmation")
	}

	// Test request with blocked target
	req.ConfirmDestruction = true
	req.Targets = []string{"/etc/passwd"}

	_, err = engine.ExecuteDestruction(ctx, req)
	if err == nil {
		t.Error("Expected error for blocked target")
	}

	// Test request with severity above limit
	req.Targets = []string{"/tmp/test.txt"}
	req.Severity = pb.DestructionSeverity_DESTRUCTION_SEVERITY_HIGH

	_, err = engine.ExecuteDestruction(ctx, req)
	if err == nil {
		t.Error("Expected error for severity above limit")
	}
}

func TestExecuteBasicDestruction(t *testing.T) {
	cfg := &config.Config{
		Security: config.SecurityConfig{
			MaxSeverity: "HIGH",
		},
	}

	engine := NewDestructionEngine(cfg)

	// Create a basic destruction task
	task := &DestructionTask{
		ID:       "test-task",
		Type:     pb.DestructionType_DESTRUCTION_TYPE_SERVICE_TERMINATION,
		Targets:  []string{"test-service"},
		Severity: pb.DestructionSeverity_DESTRUCTION_SEVERITY_LOW,
		Confirm:  true,
		Status:   "running",
		Results:  make([]*pb.DestructionResult, 0),
	}

	results, err := engine.executeBasicDestruction(task)
	if err != nil {
		t.Errorf("Expected no error from basic destruction, got: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}

	if results[0].Target != "test-service" {
		t.Errorf("Expected target 'test-service', got '%s'", results[0].Target)
	}

	// Basic destruction should always succeed in test mode
	if !results[0].Success {
		t.Error("Expected basic destruction to succeed")
	}
}

func TestSafeDeletion(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "burndevice_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test file with content
	testFile := filepath.Join(tempDir, "test.txt")
	testContent := "test content for deletion"
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	cfg := &config.Config{}
	engine := NewDestructionEngine(cfg)

	metrics := &pb.DestructionMetrics{}

	// Test safe deletion
	err = engine.safeDeletion(testFile, metrics)
	if err != nil {
		t.Errorf("Expected no error from safe deletion, got: %v", err)
	}

	// Verify file was deleted
	if _, err := os.Stat(testFile); !os.IsNotExist(err) {
		t.Error("Expected file to be deleted")
	}

	// Verify backup was created
	backupFile := testFile + ".burndevice.backup"
	if _, err := os.Stat(backupFile); os.IsNotExist(err) {
		t.Error("Expected backup file to be created")
	}

	// Verify backup content
	backupContent, err := os.ReadFile(backupFile)
	if err != nil {
		t.Errorf("Failed to read backup file: %v", err)
	}

	if string(backupContent) != testContent {
		t.Errorf("Expected backup content '%s', got '%s'", testContent, string(backupContent))
	}

	// Verify metrics
	if metrics.FilesDeleted != 1 {
		t.Errorf("Expected 1 file deleted, got %d", metrics.FilesDeleted)
	}

	if metrics.BytesDestroyed != int64(len(testContent)) {
		t.Errorf("Expected %d bytes destroyed, got %d", len(testContent), metrics.BytesDestroyed)
	}

	// Note: ExecutionTimeSeconds is set by the caller, not by safeDeletion itself
}

func TestSafeDeletionNonExistentFile(t *testing.T) {
	cfg := &config.Config{}
	engine := NewDestructionEngine(cfg)

	metrics := &pb.DestructionMetrics{}
	nonExistentFile := "/tmp/non_existent_file_12345.txt"

	// Test deletion of non-existent file
	err := engine.safeDeletion(nonExistentFile, metrics)
	if err == nil {
		t.Error("Expected error when deleting non-existent file")
	}

	// Note: ExecutionTimeSeconds is set by the caller, not by safeDeletion itself
}

func TestValidateExecuteRequest(t *testing.T) {
	cfg := &config.Config{
		Security: config.SecurityConfig{
			MaxSeverity:         "MEDIUM",
			AllowedTargets:      []string{"/tmp", "/var/tmp"},
			BlockedTargets:      []string{"/etc", "/usr/bin"},
			RequireConfirmation: true,
		},
	}

	engine := NewDestructionEngine(cfg)

	// Test valid request
	req := &pb.ExecuteDestructionRequest{
		Type:               pb.DestructionType_DESTRUCTION_TYPE_FILE_DELETION,
		Targets:            []string{"/tmp/test.txt"},
		Severity:           pb.DestructionSeverity_DESTRUCTION_SEVERITY_LOW,
		ConfirmDestruction: true,
	}

	err := engine.validateExecuteRequest(req)
	if err != nil {
		t.Errorf("Expected no error for valid request, got: %v", err)
	}

	// Test request without confirmation
	req.ConfirmDestruction = false
	err = engine.validateExecuteRequest(req)
	if err == nil {
		t.Error("Expected error for request without confirmation")
	}

	// Test request with blocked target
	req.ConfirmDestruction = true
	req.Targets = []string{"/etc/passwd"}
	err = engine.validateExecuteRequest(req)
	if err == nil {
		t.Error("Expected error for blocked target")
	}

	// Test request with severity above limit
	req.Targets = []string{"/tmp/test.txt"}
	req.Severity = pb.DestructionSeverity_DESTRUCTION_SEVERITY_HIGH
	err = engine.validateExecuteRequest(req)
	if err == nil {
		t.Error("Expected error for severity above limit")
	}
}

func TestValidateStreamRequest(t *testing.T) {
	cfg := &config.Config{
		Security: config.SecurityConfig{
			MaxSeverity:         "MEDIUM",
			AllowedTargets:      []string{"/tmp"},
			BlockedTargets:      []string{"/etc"},
			RequireConfirmation: true,
		},
	}

	engine := NewDestructionEngine(cfg)

	// Test valid request
	req := &pb.StreamDestructionRequest{
		Type:               pb.DestructionType_DESTRUCTION_TYPE_FILE_DELETION,
		Targets:            []string{"/tmp/test.txt"},
		Severity:           pb.DestructionSeverity_DESTRUCTION_SEVERITY_LOW,
		ConfirmDestruction: true,
	}

	err := engine.validateStreamRequest(req)
	if err != nil {
		t.Errorf("Expected no error for valid request, got: %v", err)
	}

	// Test request without confirmation
	req.ConfirmDestruction = false
	err = engine.validateStreamRequest(req)
	if err == nil {
		t.Error("Expected error for request without confirmation")
	}
}

func TestIsBlockedTarget(t *testing.T) {
	cfg := &config.Config{
		Security: config.SecurityConfig{
			BlockedTargets: []string{"/etc", "/var/log", "/usr/bin"},
		},
	}

	engine := NewDestructionEngine(cfg)

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
			result := engine.isBlockedTarget(tt.target)
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

	engine := NewDestructionEngine(cfg)

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
			result := engine.isAllowedTarget(tt.target)
			if result != tt.expected {
				t.Errorf("Expected isAllowed %v for '%s', got %v", tt.expected, tt.target, result)
			}
		})
	}
}

func TestGetSeverityLevel(t *testing.T) {
	engine := &DestructionEngine{}

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
			result := engine.getSeverityLevel(tt.severity)
			if result != tt.expected {
				t.Errorf("Expected severity level %d for '%s', got %d", tt.expected, tt.severity, result)
			}
		})
	}
}

func TestCopyFile(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "burndevice_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create source file
	srcFile := filepath.Join(tempDir, "source.txt")
	testContent := "test content for copying"
	if err := os.WriteFile(srcFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	// Test file copying
	dstFile := filepath.Join(tempDir, "destination.txt")

	// Create properly initialized engine with minimal config
	cfg := &config.Config{
		Security: config.SecurityConfig{
			MaxSeverity:    "HIGH",
			BlockedTargets: []string{"/etc", "/var", "/usr"}, // Common system paths
		},
	}
	engine := NewDestructionEngine(cfg)

	err = engine.copyFile(srcFile, dstFile)
	if err != nil {
		t.Errorf("Expected no error copying file, got: %v", err)
	}

	// Verify destination file exists
	if _, err := os.Stat(dstFile); os.IsNotExist(err) {
		t.Error("Expected destination file to exist")
	}

	// Verify content
	dstContent, err := os.ReadFile(dstFile)
	if err != nil {
		t.Errorf("Failed to read destination file: %v", err)
	}

	if string(dstContent) != testContent {
		t.Errorf("Expected content '%s', got '%s'", testContent, string(dstContent))
	}

	// Test copying non-existent file
	nonExistentSrc := filepath.Join(tempDir, "non_existent.txt")
	nonExistentDst := filepath.Join(tempDir, "non_existent_dst.txt")

	err = engine.copyFile(nonExistentSrc, nonExistentDst)
	if err == nil {
		t.Error("Expected error when copying non-existent file")
	}
}

func TestGenerateTaskID(t *testing.T) {
	// Test that task IDs are generated
	id1 := generateTaskID()
	id2 := generateTaskID()

	if id1 == "" {
		t.Error("Expected task ID to be generated")
	}

	if id2 == "" {
		t.Error("Expected task ID to be generated")
	}

	// Task IDs should be different
	if id1 == id2 {
		t.Error("Expected different task IDs")
	}

	// Task IDs should have reasonable length
	if len(id1) < 10 {
		t.Error("Expected task ID to have reasonable length")
	}
}

func TestDestructionTaskManagement(t *testing.T) {
	cfg := &config.Config{
		Security: config.SecurityConfig{
			MaxSeverity: "HIGH",
		},
	}

	engine := NewDestructionEngine(cfg)

	// Verify no tasks initially
	if len(engine.running) != 0 {
		t.Error("Expected no running tasks initially")
	}

	// Create a task context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Test task registration during execution
	// We can't easily test the internal task management without exposing internals,
	// but we can test that execution completes properly
	req := &pb.ExecuteDestructionRequest{
		Type:               pb.DestructionType_DESTRUCTION_TYPE_SERVICE_TERMINATION,
		Targets:            []string{"test-service"},
		Severity:           pb.DestructionSeverity_DESTRUCTION_SEVERITY_LOW,
		ConfirmDestruction: true,
	}

	resp, err := engine.ExecuteDestruction(ctx, req)
	if err != nil {
		t.Errorf("Expected no error from execution, got: %v", err)
	}

	if resp == nil {
		t.Error("Expected response from execution")
	}

	// After execution, task should be cleaned up
	if len(engine.running) != 0 {
		t.Error("Expected no running tasks after execution")
	}
}

func TestComplexValidationScenarios(t *testing.T) {
	cfg := &config.Config{
		Security: config.SecurityConfig{
			MaxSeverity:    "MEDIUM",
			AllowedTargets: []string{"/tmp"},
			BlockedTargets: []string{"/etc", "/tmp/blocked"},
		},
	}

	engine := NewDestructionEngine(cfg)

	// Test target that is both allowed and blocked (blocked should take precedence)
	req := &pb.ExecuteDestructionRequest{
		Type:               pb.DestructionType_DESTRUCTION_TYPE_FILE_DELETION,
		Targets:            []string{"/tmp/blocked/file.txt"},
		Severity:           pb.DestructionSeverity_DESTRUCTION_SEVERITY_LOW,
		ConfirmDestruction: true,
	}

	err := engine.validateExecuteRequest(req)
	if err == nil {
		t.Error("Expected error for target that is blocked despite being in allowed path")
	}

	// Test multiple targets with mixed validity
	req.Targets = []string{"/tmp/valid.txt", "/etc/passwd"}
	err = engine.validateExecuteRequest(req)
	if err == nil {
		t.Error("Expected error when any target is blocked")
	}

	// Test empty targets
	req.Targets = []string{}
	err = engine.validateExecuteRequest(req)
	// Empty targets might be valid depending on destruction type
	// The specific validation depends on implementation
}

func TestExecuteDestructionTypes(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "burndevice_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	cfg := &config.Config{
		Security: config.SecurityConfig{
			MaxSeverity: "HIGH",
			// Don't set AllowedTargets to allow all targets except blocked ones
		},
	}

	engine := NewDestructionEngine(cfg)
	ctx := context.Background()

	// Test different destruction types
	destructionTypes := []pb.DestructionType{
		pb.DestructionType_DESTRUCTION_TYPE_FILE_DELETION,
		pb.DestructionType_DESTRUCTION_TYPE_SERVICE_TERMINATION,
		pb.DestructionType_DESTRUCTION_TYPE_MEMORY_EXHAUSTION,
		pb.DestructionType_DESTRUCTION_TYPE_DISK_FILL,
		pb.DestructionType_DESTRUCTION_TYPE_NETWORK_DISRUPTION,
	}

	for _, dtype := range destructionTypes {
		t.Run(dtype.String(), func(t *testing.T) {
			var targets []string
			if dtype == pb.DestructionType_DESTRUCTION_TYPE_FILE_DELETION {
				// Create a test file for file deletion
				testFile := filepath.Join(tempDir, fmt.Sprintf("test_%s.txt", dtype.String()))
				if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
					t.Fatalf("Failed to create test file: %v", err)
				}
				targets = []string{testFile}
			} else {
				targets = []string{"test-target"}
			}

			req := &pb.ExecuteDestructionRequest{
				Type:               dtype,
				Targets:            targets,
				Severity:           pb.DestructionSeverity_DESTRUCTION_SEVERITY_LOW,
				ConfirmDestruction: true,
			}

			resp, err := engine.ExecuteDestruction(ctx, req)
			if err != nil {
				t.Errorf("Expected no error for destruction type %s, got: %v", dtype.String(), err)
			}

			if resp == nil {
				t.Errorf("Expected response for destruction type %s", dtype.String())
			}

			if !resp.Success {
				t.Errorf("Expected success for destruction type %s, got: %s", dtype.String(), resp.Message)
			}
		})
	}
}

func TestEngineWithMinimalConfig(t *testing.T) {
	// Test engine with minimal configuration
	cfg := &config.Config{}
	engine := NewDestructionEngine(cfg)

	if engine == nil {
		t.Fatal("Expected engine to be created with minimal config")
	}

	// Test basic functionality with minimal config
	ctx := context.Background()
	req := &pb.ExecuteDestructionRequest{
		Type:               pb.DestructionType_DESTRUCTION_TYPE_SERVICE_TERMINATION,
		Targets:            []string{"test-service"},
		Severity:           pb.DestructionSeverity_DESTRUCTION_SEVERITY_LOW,
		ConfirmDestruction: true,
	}

	resp, err := engine.ExecuteDestruction(ctx, req)
	if err != nil {
		t.Errorf("Expected engine to work with minimal config, got: %v", err)
	}

	if resp == nil {
		t.Error("Expected response even with minimal config")
	}
}
