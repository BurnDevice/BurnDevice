package ai

import (
	"context"
	"testing"
	"time"

	pb "github.com/BurnDevice/BurnDevice/burndevice/v1"
	"github.com/BurnDevice/BurnDevice/internal/config"
)

func TestNewDeepSeekClient(t *testing.T) {
	cfg := &config.AIConfig{
		Provider:       "deepseek",
		APIKey:         "test-key",
		BaseURL:        "https://api.deepseek.com",
		Model:          "deepseek-chat",
		MaxTokens:      4096,
		Temperature:    0.7,
		RequestTimeout: 30 * time.Second,
	}

	client := NewDeepSeekClient(cfg)
	if client == nil {
		t.Fatal("Expected client to be created")
	}

	if client.config != cfg {
		t.Error("Expected config to be set")
	}

	if client.httpClient == nil {
		t.Error("Expected HTTP client to be initialized")
	}

	if client.logger == nil {
		t.Error("Expected logger to be initialized")
	}
}

func TestParseSeverity(t *testing.T) {
	cfg := &config.AIConfig{
		Provider: "deepseek",
	}
	client := NewDeepSeekClient(cfg)

	tests := []struct {
		input    string
		expected pb.DestructionSeverity
	}{
		{"LOW", pb.DestructionSeverity_DESTRUCTION_SEVERITY_LOW},
		{"MEDIUM", pb.DestructionSeverity_DESTRUCTION_SEVERITY_MEDIUM},
		{"HIGH", pb.DestructionSeverity_DESTRUCTION_SEVERITY_HIGH},
		{"CRITICAL", pb.DestructionSeverity_DESTRUCTION_SEVERITY_CRITICAL},
		{"low", pb.DestructionSeverity_DESTRUCTION_SEVERITY_LOW},
		{"medium", pb.DestructionSeverity_DESTRUCTION_SEVERITY_MEDIUM},
		{"invalid", pb.DestructionSeverity_DESTRUCTION_SEVERITY_LOW},
		{"", pb.DestructionSeverity_DESTRUCTION_SEVERITY_LOW},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := client.parseSeverity(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %v for input %s, got %v", tt.expected, tt.input, result)
			}
		})
	}
}

func TestParseDestructionType(t *testing.T) {
	cfg := &config.AIConfig{
		Provider: "deepseek",
	}
	client := NewDeepSeekClient(cfg)

	tests := []struct {
		input    string
		expected pb.DestructionType
	}{
		{"FILE_DELETION", pb.DestructionType_DESTRUCTION_TYPE_FILE_DELETION},
		{"SERVICE_TERMINATION", pb.DestructionType_DESTRUCTION_TYPE_SERVICE_TERMINATION},
		{"MEMORY_EXHAUSTION", pb.DestructionType_DESTRUCTION_TYPE_MEMORY_EXHAUSTION},
		{"DISK_FILL", pb.DestructionType_DESTRUCTION_TYPE_DISK_FILL},
		{"NETWORK_DISRUPTION", pb.DestructionType_DESTRUCTION_TYPE_NETWORK_DISRUPTION},
		{"file_deletion", pb.DestructionType_DESTRUCTION_TYPE_FILE_DELETION},
		{"invalid", pb.DestructionType_DESTRUCTION_TYPE_FILE_DELETION},
		{"", pb.DestructionType_DESTRUCTION_TYPE_FILE_DELETION},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := client.parseDestructionType(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %v for input %s, got %v", tt.expected, tt.input, result)
			}
		})
	}
}

func TestBuildSystemPrompt(t *testing.T) {
	cfg := &config.AIConfig{
		Provider: "deepseek",
	}
	client := NewDeepSeekClient(cfg)

	prompt := client.buildSystemPrompt(pb.DestructionSeverity_DESTRUCTION_SEVERITY_MEDIUM)

	if prompt == "" {
		t.Error("Expected system prompt to be generated")
	}

	// Check that the prompt contains key elements
	if !contains(prompt, "MEDIUM") {
		t.Error("Expected prompt to contain severity level")
	}

	if !contains(prompt, "FILE_DELETION") {
		t.Error("Expected prompt to contain destruction types")
	}

	if !contains(prompt, "JSON") {
		t.Error("Expected prompt to mention JSON format")
	}
}

func TestBuildUserPrompt(t *testing.T) {
	cfg := &config.AIConfig{
		Provider: "deepseek",
	}
	client := NewDeepSeekClient(cfg)

	target := "Linux test server"
	severity := pb.DestructionSeverity_DESTRUCTION_SEVERITY_LOW

	prompt := client.buildUserPrompt(target, severity)

	if prompt == "" {
		t.Error("Expected user prompt to be generated")
	}

	if !contains(prompt, target) {
		t.Error("Expected prompt to contain target description")
	}

	if !contains(prompt, "LOW") {
		t.Error("Expected prompt to contain severity level")
	}
}

func TestParseScenarioFromContent(t *testing.T) {
	cfg := &config.AIConfig{
		Provider: "deepseek",
	}
	client := NewDeepSeekClient(cfg)

	// Test with valid JSON
	validJSON := `{
		"id": "test-123",
		"description": "Test scenario",
		"severity": "LOW",
		"steps": [
			{
				"order": 1,
				"type": "FILE_DELETION",
				"description": "Delete test files",
				"targets": ["/tmp/test.txt"],
				"rationale": "Test rationale",
				"risk": "LOW"
			}
		],
		"rationale": "Test scenario rationale",
		"warnings": ["Test warning"]
	}`

	scenario, err := client.parseScenarioFromContent(validJSON)
	if err != nil {
		t.Fatalf("Failed to parse valid JSON: %v", err)
	}

	if scenario.ID != "test-123" {
		t.Errorf("Expected ID 'test-123', got '%s'", scenario.ID)
	}

	if scenario.Description != "Test scenario" {
		t.Errorf("Expected description 'Test scenario', got '%s'", scenario.Description)
	}

	if len(scenario.Steps) != 1 {
		t.Errorf("Expected 1 step, got %d", len(scenario.Steps))
	}

	// Test with invalid JSON
	invalidJSON := `{"invalid": json}`
	_, err = client.parseScenarioFromContent(invalidJSON)
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}
}

func TestValidateScenario(t *testing.T) {
	cfg := &config.AIConfig{
		Provider: "deepseek",
	}
	client := NewDeepSeekClient(cfg)

	validScenario := &AttackScenario{
		ID:          "test-123",
		Description: "Test scenario",
		Severity:    "LOW",
		Steps: []AttackStep{
			{
				Order:       1,
				Type:        "FILE_DELETION",
				Description: "Delete test files",
				Targets:     []string{"/tmp/test.txt"},
				Rationale:   "Test rationale",
				Risk:        "LOW",
			},
		},
	}

	err := client.ValidateScenario(validScenario, pb.DestructionSeverity_DESTRUCTION_SEVERITY_MEDIUM)
	if err != nil {
		t.Errorf("Expected valid scenario to pass validation: %v", err)
	}

	// Test scenario with severity too high
	highSeverityScenario := &AttackScenario{
		ID:          "test-456",
		Description: "High severity test",
		Severity:    "CRITICAL",
		Steps:       validScenario.Steps,
	}

	err = client.ValidateScenario(highSeverityScenario, pb.DestructionSeverity_DESTRUCTION_SEVERITY_LOW)
	if err == nil {
		t.Error("Expected error for scenario with severity too high")
	}

	// Test scenario with empty steps
	emptyStepsScenario := &AttackScenario{
		ID:          "test-789",
		Description: "Empty steps test",
		Severity:    "LOW",
		Steps:       []AttackStep{},
	}

	err = client.ValidateScenario(emptyStepsScenario, pb.DestructionSeverity_DESTRUCTION_SEVERITY_MEDIUM)
	if err == nil {
		t.Error("Expected error for scenario with empty steps")
	}
}

func TestGenerateAttackScenario_ValidationOnly(t *testing.T) {
	// Test the request validation part without making actual API calls
	cfg := &config.AIConfig{
		Provider: "deepseek",
		APIKey:   "", // Empty API key to trigger validation error
	}
	client := NewDeepSeekClient(cfg)

	req := &pb.GenerateAttackScenarioRequest{
		TargetDescription: "",
		MaxSeverity:       pb.DestructionSeverity_DESTRUCTION_SEVERITY_LOW,
	}

	ctx := context.Background()
	_, err := client.GenerateAttackScenario(ctx, req)
	if err == nil {
		t.Error("Expected error for empty target description")
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(substr) > 0 && len(s) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
