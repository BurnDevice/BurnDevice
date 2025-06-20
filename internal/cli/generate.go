package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// NewGenerateCommand creates the generate command
func NewGenerateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate example scenarios and configurations",
		Long:  "生成示例场景和配置文件",
	}

	cmd.AddCommand(
		newGenerateConfigCommand(),
		newGenerateExampleCommand(),
	)

	return cmd
}

func newGenerateConfigCommand() *cobra.Command {
	var outputPath string

	cmd := &cobra.Command{
		Use:   "config",
		Short: "Generate example configuration file",
		Long:  "生成示例配置文件",
		RunE: func(cmd *cobra.Command, args []string) error {
			config := `# BurnDevice Configuration
# ⚠️ 警告：此配置仅用于授权的测试环境

server:
  host: "0.0.0.0"
  port: 8080
  read_timeout: "30s"
  write_timeout: "30s"
  tls:
    enabled: false
    cert_file: ""
    key_file: ""

ai:
  provider: "deepseek"
  api_key: "${BURNDEVICE_AI_API_KEY}"
  base_url: "https://api.deepseek.com"
  model: "deepseek-chat"
  max_tokens: 4096
  temperature: 0.7
  request_timeout: "30s"

security:
  require_confirmation: true
  max_severity: "MEDIUM"
  enable_safe_mode: true
  audit_log: true
  
  allowed_targets:
    - "/tmp/burndevice_test"
    - "/home/user/test"
  
  blocked_targets:
    - "/"
    - "/bin"
    - "/usr"
    - "/etc"
    - "/var"
    - "/home"
    - "/root"

log_level: "info"
`

			if err := os.WriteFile(outputPath, []byte(config), 0600); err != nil {
				return fmt.Errorf("failed to write config file: %w", err)
			}

			fmt.Printf("✅ Configuration file generated: %s\n", outputPath)
			return nil
		},
	}

	cmd.Flags().StringVar(&outputPath, "output", "burndevice-config.yaml", "Output configuration file path")

	return cmd
}

func newGenerateExampleCommand() *cobra.Command {
	var (
		outputDir string
		count     int
	)

	cmd := &cobra.Command{
		Use:   "examples",
		Short: "Generate example attack scenarios",
		Long:  "生成示例攻击场景",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := os.MkdirAll(outputDir, 0750); err != nil {
				return fmt.Errorf("failed to create output directory: %w", err)
			}

			examples := []map[string]interface{}{
				{
					"id":          "example_file_deletion_low",
					"description": "Low severity file deletion test for temporary files",
					"severity":    "LOW",
					"steps": []map[string]interface{}{
						{
							"order":       1,
							"type":        "FILE_DELETION",
							"description": "Create test files in /tmp directory",
							"targets":     []string{"/tmp/burndevice_test_file.txt"},
							"rationale":   "Safe test environment with recoverable files",
						},
						{
							"order":       2,
							"type":        "FILE_DELETION",
							"description": "Safely delete test files with backup",
							"targets":     []string{"/tmp/burndevice_test_file.txt"},
							"rationale":   "Low severity deletion creates backup before removal",
						},
					},
				},
				{
					"id":          "example_memory_exhaustion",
					"description": "Memory exhaustion test for system resilience",
					"severity":    "MEDIUM",
					"steps": []map[string]interface{}{
						{
							"order":       1,
							"type":        "MEMORY_EXHAUSTION",
							"description": "Gradually allocate memory in chunks",
							"targets":     []string{"system_memory"},
							"rationale":   "Test system behavior under memory pressure",
						},
					},
				},
				{
					"id":          "example_service_disruption",
					"description": "Service disruption test for non-critical services",
					"severity":    "LOW",
					"steps": []map[string]interface{}{
						{
							"order":       1,
							"type":        "SERVICE_TERMINATION",
							"description": "Stop test service",
							"targets":     []string{"test-service"},
							"rationale":   "Verify service restart capabilities",
						},
					},
				},
			}

			for i, example := range examples {
				if i >= count {
					break
				}

				filename := fmt.Sprintf("scenario_%s.json", example["id"])
				filepath := filepath.Join(outputDir, filename)

				data, err := json.MarshalIndent(example, "", "  ")
				if err != nil {
					return fmt.Errorf("failed to marshal example %d: %w", i+1, err)
				}

				if err := os.WriteFile(filepath, data, 0600); err != nil {
					return fmt.Errorf("failed to write example %d: %w", i+1, err)
				}

				logrus.WithField("file", filepath).Info("Generated example scenario")
			}

			fmt.Printf("✅ Generated %d example scenarios in %s\n", len(examples), outputDir)
			return nil
		},
	}

	cmd.Flags().StringVar(&outputDir, "output", "examples", "Output directory for examples")
	cmd.Flags().IntVar(&count, "count", 10, "Number of examples to generate")

	return cmd
}
