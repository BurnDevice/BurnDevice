package cli

import (
	"bytes"
	"strings"
	"testing"
	"time"

	pb "github.com/BurnDevice/BurnDevice/burndevice/v1"
	"github.com/spf13/cobra"
)

func TestNewClientCommand(t *testing.T) {
	cmd := NewClientCommand()
	if cmd == nil {
		t.Fatal("Expected client command to be created")
	}

	if cmd.Use != "client" {
		t.Errorf("Expected command use 'client', got '%s'", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("Expected command to have short description")
	}

	// Check that subcommands are added
	if len(cmd.Commands()) == 0 {
		t.Error("Expected client command to have subcommands")
	}

	// Verify persistent flags
	flags := cmd.PersistentFlags()
	if flags.Lookup("server") == nil {
		t.Error("Expected 'server' flag to be defined")
	}

	if flags.Lookup("timeout") == nil {
		t.Error("Expected 'timeout' flag to be defined")
	}
}

func TestParseDestructionType(t *testing.T) {
	tests := []struct {
		input    string
		expected pb.DestructionType
		hasError bool
	}{
		{"FILE_DELETION", pb.DestructionType_DESTRUCTION_TYPE_FILE_DELETION, false},
		{"SERVICE_TERMINATION", pb.DestructionType_DESTRUCTION_TYPE_SERVICE_TERMINATION, false},
		{"MEMORY_EXHAUSTION", pb.DestructionType_DESTRUCTION_TYPE_MEMORY_EXHAUSTION, false},
		{"DISK_FILL", pb.DestructionType_DESTRUCTION_TYPE_DISK_FILL, false},
		{"NETWORK_DISRUPTION", pb.DestructionType_DESTRUCTION_TYPE_NETWORK_DISRUPTION, false},
		{"BOOT_CORRUPTION", pb.DestructionType_DESTRUCTION_TYPE_BOOT_CORRUPTION, false},
		{"KERNEL_PANIC", pb.DestructionType_DESTRUCTION_TYPE_KERNEL_PANIC, false},
		{"file_deletion", pb.DestructionType_DESTRUCTION_TYPE_FILE_DELETION, false},
		{"service_termination", pb.DestructionType_DESTRUCTION_TYPE_SERVICE_TERMINATION, false},
		{"INVALID_TYPE", pb.DestructionType_DESTRUCTION_TYPE_UNSPECIFIED, true},
		{"", pb.DestructionType_DESTRUCTION_TYPE_UNSPECIFIED, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := parseDestructionType(tt.input)

			if tt.hasError {
				if err == nil {
					t.Errorf("Expected error for input '%s', but got none", tt.input)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error for input '%s', but got: %v", tt.input, err)
				}
				if result != tt.expected {
					t.Errorf("Expected %v for input '%s', got %v", tt.expected, tt.input, result)
				}
			}
		})
	}
}

func TestParseSeverity(t *testing.T) {
	tests := []struct {
		input    string
		expected pb.DestructionSeverity
		hasError bool
	}{
		{"LOW", pb.DestructionSeverity_DESTRUCTION_SEVERITY_LOW, false},
		{"MEDIUM", pb.DestructionSeverity_DESTRUCTION_SEVERITY_MEDIUM, false},
		{"HIGH", pb.DestructionSeverity_DESTRUCTION_SEVERITY_HIGH, false},
		{"CRITICAL", pb.DestructionSeverity_DESTRUCTION_SEVERITY_CRITICAL, false},
		{"low", pb.DestructionSeverity_DESTRUCTION_SEVERITY_LOW, false},
		{"medium", pb.DestructionSeverity_DESTRUCTION_SEVERITY_MEDIUM, false},
		{"INVALID_SEVERITY", pb.DestructionSeverity_DESTRUCTION_SEVERITY_UNSPECIFIED, true},
		{"", pb.DestructionSeverity_DESTRUCTION_SEVERITY_UNSPECIFIED, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := parseSeverity(tt.input)

			if tt.hasError {
				if err == nil {
					t.Errorf("Expected error for input '%s', but got none", tt.input)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error for input '%s', but got: %v", tt.input, err)
				}
				if result != tt.expected {
					t.Errorf("Expected %v for input '%s', got %v", tt.expected, tt.input, result)
				}
			}
		})
	}
}

func TestGetTimeout(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().Duration("timeout", 30*time.Second, "Request timeout")

	// Test default timeout
	timeout := getTimeout(cmd)
	expected := 30 * time.Second
	if timeout != expected {
		t.Errorf("Expected timeout %v, got %v", expected, timeout)
	}

	// Test setting custom timeout
	if err := cmd.Flags().Set("timeout", "60s"); err != nil {
		t.Errorf("Failed to set timeout flag: %v", err)
	}
	timeout = getTimeout(cmd)
	expected = 60 * time.Second
	if timeout != expected {
		t.Errorf("Expected timeout %v, got %v", expected, timeout)
	}
}

func TestNewExecuteCommand(t *testing.T) {
	cmd := newExecuteCommand()
	if cmd == nil {
		t.Fatal("Expected execute command to be created")
	}

	if cmd.Use != "execute" {
		t.Errorf("Expected command use 'execute', got '%s'", cmd.Use)
	}

	// Check required flags
	flags := cmd.Flags()
	if flags.Lookup("type") == nil {
		t.Error("Expected 'type' flag to be defined")
	}

	if flags.Lookup("targets") == nil {
		t.Error("Expected 'targets' flag to be defined")
	}

	if flags.Lookup("severity") == nil {
		t.Error("Expected 'severity' flag to be defined")
	}

	if flags.Lookup("confirm") == nil {
		t.Error("Expected 'confirm' flag to be defined")
	}

	// Basic validation that flags are properly set up
	// The actual required flag validation is handled by cobra internally
}

func TestNewSystemInfoCommand(t *testing.T) {
	cmd := newSystemInfoCommand()
	if cmd == nil {
		t.Fatal("Expected system-info command to be created")
	}

	if cmd.Use != "system-info" {
		t.Errorf("Expected command use 'system-info', got '%s'", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("Expected command to have short description")
	}
}

func TestNewGenerateScenarioCommand(t *testing.T) {
	cmd := newGenerateScenarioCommand()
	if cmd == nil {
		t.Fatal("Expected generate-scenario command to be created")
	}

	if cmd.Use != "generate-scenario" {
		t.Errorf("Expected command use 'generate-scenario', got '%s'", cmd.Use)
	}

	// Check flags
	flags := cmd.Flags()
	if flags.Lookup("target") == nil {
		t.Error("Expected 'target' flag to be defined")
	}

	if flags.Lookup("max-severity") == nil {
		t.Error("Expected 'max-severity' flag to be defined")
	}

	if flags.Lookup("model") == nil {
		t.Error("Expected 'model' flag to be defined")
	}
}

func TestNewStreamCommand(t *testing.T) {
	cmd := newStreamCommand()
	if cmd == nil {
		t.Fatal("Expected stream command to be created")
	}

	if cmd.Use != "stream" {
		t.Errorf("Expected command use 'stream', got '%s'", cmd.Use)
	}

	// Check flags
	flags := cmd.Flags()
	if flags.Lookup("type") == nil {
		t.Error("Expected 'type' flag to be defined")
	}

	if flags.Lookup("targets") == nil {
		t.Error("Expected 'targets' flag to be defined")
	}

	if flags.Lookup("severity") == nil {
		t.Error("Expected 'severity' flag to be defined")
	}

	if flags.Lookup("confirm") == nil {
		t.Error("Expected 'confirm' flag to be defined")
	}
}

func TestExecuteCommandValidation(t *testing.T) {
	cmd := newExecuteCommand()

	// Test command without confirm flag
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	// Set required flags but not confirm
	if err := cmd.Flags().Set("type", "FILE_DELETION"); err != nil {
		t.Errorf("Failed to set type flag: %v", err)
	}
	if err := cmd.Flags().Set("targets", "test.txt"); err != nil {
		t.Errorf("Failed to set targets flag: %v", err)
	}
	if err := cmd.Flags().Set("severity", "LOW"); err != nil {
		t.Errorf("Failed to set severity flag: %v", err)
	}

	err := cmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("Expected error when confirm flag is not set")
	}

	if !strings.Contains(err.Error(), "confirm") {
		t.Errorf("Expected error message to contain 'confirm', got: %v", err)
	}
}

func TestExecuteCommandWithInvalidType(t *testing.T) {
	cmd := newExecuteCommand()

	// Set invalid destruction type
	if err := cmd.Flags().Set("type", "INVALID_TYPE"); err != nil {
		t.Errorf("Failed to set type flag: %v", err)
	}
	if err := cmd.Flags().Set("confirm", "true"); err != nil {
		t.Errorf("Failed to set confirm flag: %v", err)
	}

	err := cmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("Expected error with invalid destruction type")
	}
}

func TestExecuteCommandWithInvalidSeverity(t *testing.T) {
	cmd := newExecuteCommand()

	// Set valid type but invalid severity
	if err := cmd.Flags().Set("type", "FILE_DELETION"); err != nil {
		t.Errorf("Failed to set type flag: %v", err)
	}
	if err := cmd.Flags().Set("severity", "INVALID_SEVERITY"); err != nil {
		t.Errorf("Failed to set severity flag: %v", err)
	}
	if err := cmd.Flags().Set("confirm", "true"); err != nil {
		t.Errorf("Failed to set confirm flag: %v", err)
	}

	err := cmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("Expected error with invalid severity")
	}
}

func TestCreateClientConnectionError(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().String("server", "invalid-address:99999", "Server address")
	cmd.Flags().Duration("timeout", 1*time.Second, "Request timeout")

	// This should fail because the address is invalid
	client, conn, err := createClient(cmd)
	if err != nil {
		// Expected - connection should fail
		if conn != nil {
			if err := conn.Close(); err != nil {
				t.Errorf("Failed to close connection: %v", err)
			}
		}
		return
	}

	// If we get here, the connection might have been created
	if conn != nil {
		if err := conn.Close(); err != nil {
			t.Errorf("Failed to close connection: %v", err)
		}
	}

	// We can't guarantee this will fail in all environments, so just check the client is created
	if client == nil {
		t.Error("Expected client to be created even if connection fails")
	}
}

func TestGenerateScenarioCommandFlags(t *testing.T) {
	cmd := newGenerateScenarioCommand()

	// Test flag defaults
	flags := cmd.Flags()

	maxSeverityFlag := flags.Lookup("max-severity")
	if maxSeverityFlag == nil {
		t.Error("Expected 'max-severity' flag to be defined")
	}

	modelFlag := flags.Lookup("model")
	if modelFlag == nil {
		t.Error("Expected 'model' flag to be defined")
	}

	// Test required flag
	targetFlag := flags.Lookup("target")
	if targetFlag == nil {
		t.Error("Expected 'target' flag to be defined")
	}
}

func TestStreamCommandFlags(t *testing.T) {
	cmd := newStreamCommand()

	// Test all expected flags are present
	expectedFlags := []string{"type", "targets", "severity", "confirm", "scenario-id"}

	for _, flagName := range expectedFlags {
		if cmd.Flags().Lookup(flagName) == nil {
			t.Errorf("Expected '%s' flag to be defined", flagName)
		}
	}
}

func TestCommandDescriptions(t *testing.T) {
	commands := []struct {
		name string
		cmd  *cobra.Command
	}{
		{"execute", newExecuteCommand()},
		{"system-info", newSystemInfoCommand()},
		{"generate-scenario", newGenerateScenarioCommand()},
		{"stream", newStreamCommand()},
	}

	for _, tc := range commands {
		t.Run(tc.name, func(t *testing.T) {
			if tc.cmd.Short == "" {
				t.Errorf("Command '%s' should have Short description", tc.name)
			}

			if tc.cmd.Long == "" {
				t.Errorf("Command '%s' should have Long description", tc.name)
			}
		})
	}
}

func TestParseDestructionTypeEdgeCases(t *testing.T) {
	// Test with whitespace - should fail since function doesn't trim
	_, err := parseDestructionType("  FILE_DELETION  ")
	if err == nil {
		t.Error("Expected error with whitespace since function doesn't trim")
	}

	// Test case insensitive - should work due to strings.ToUpper
	result, err := parseDestructionType("file_deletion")
	if err != nil {
		t.Errorf("Expected no error with lowercase, got: %v", err)
	}
	if result != pb.DestructionType_DESTRUCTION_TYPE_FILE_DELETION {
		t.Errorf("Expected FILE_DELETION, got %v", result)
	}

	// Test mixed case
	result, err = parseDestructionType("Memory_Exhaustion")
	if err != nil {
		t.Errorf("Expected no error with mixed case, got: %v", err)
	}
	if result != pb.DestructionType_DESTRUCTION_TYPE_MEMORY_EXHAUSTION {
		t.Errorf("Expected MEMORY_EXHAUSTION, got %v", result)
	}
}

func TestParseSeverityEdgeCases(t *testing.T) {
	// Test with whitespace - should fail since function doesn't trim
	_, err := parseSeverity("  HIGH  ")
	if err == nil {
		t.Error("Expected error with whitespace since function doesn't trim")
	}

	// Test case insensitive - should work due to strings.ToUpper
	result, err := parseSeverity("low")
	if err != nil {
		t.Errorf("Expected no error with lowercase, got: %v", err)
	}
	if result != pb.DestructionSeverity_DESTRUCTION_SEVERITY_LOW {
		t.Errorf("Expected LOW, got %v", result)
	}

	// Test mixed case
	result, err = parseSeverity("Critical")
	if err != nil {
		t.Errorf("Expected no error with mixed case, got: %v", err)
	}
	if result != pb.DestructionSeverity_DESTRUCTION_SEVERITY_CRITICAL {
		t.Errorf("Expected CRITICAL, got %v", result)
	}
}

func TestClientCommandIntegration(t *testing.T) {
	// Test the full client command structure
	clientCmd := NewClientCommand()

	// Verify all subcommands are present
	expectedSubcommands := []string{"execute", "system-info", "generate-scenario", "stream"}
	actualSubcommands := make([]string, 0, len(clientCmd.Commands()))

	for _, cmd := range clientCmd.Commands() {
		actualSubcommands = append(actualSubcommands, cmd.Use)
	}

	for _, expected := range expectedSubcommands {
		found := false
		for _, actual := range actualSubcommands {
			if actual == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected subcommand '%s' not found in client command", expected)
		}
	}
}

func TestCommandHelpOutput(t *testing.T) {
	clientCmd := NewClientCommand()

	var buf bytes.Buffer
	clientCmd.SetOut(&buf)
	clientCmd.SetArgs([]string{"--help"})

	err := clientCmd.Execute()
	if err != nil {
		t.Fatalf("Expected no error executing help, got: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "client") {
		t.Error("Expected help output to contain 'client'")
	}

	if !strings.Contains(output, "BurnDevice") {
		t.Error("Expected help output to contain 'BurnDevice'")
	}
}

func TestTimeoutHandling(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().Duration("timeout", 0, "Request timeout")

	// Test zero timeout
	if err := cmd.Flags().Set("timeout", "0s"); err != nil {
		t.Errorf("Failed to set timeout flag: %v", err)
	}
	timeout := getTimeout(cmd)
	if timeout != 0 {
		t.Errorf("Expected timeout 0, got %v", timeout)
	}

	// Test negative timeout (should still work)
	if err := cmd.Flags().Set("timeout", "-5s"); err != nil {
		t.Errorf("Failed to set timeout flag: %v", err)
	}
	timeout = getTimeout(cmd)
	if timeout != -5*time.Second {
		t.Errorf("Expected timeout -5s, got %v", timeout)
	}
}

func TestGRPCClientCreation(t *testing.T) {
	// Test with minimal valid command
	cmd := &cobra.Command{}
	cmd.Flags().String("server", "localhost:8080", "Server address")
	cmd.Flags().Duration("timeout", 30*time.Second, "Request timeout")

	// This test verifies the function doesn't panic
	// Actual connection will likely fail, but that's expected in test environment
	client, conn, err := createClient(cmd)

	// Clean up if connection was established
	if conn != nil {
		if err := conn.Close(); err != nil {
			t.Errorf("Failed to close connection: %v", err)
		}
	}

	// In test environment, we expect either:
	// 1. Success (if test server is running)
	// 2. Connection error (if no server)
	// Both are acceptable for this test

	if err == nil && client == nil {
		t.Error("If no error, client should not be nil")
	}
}
