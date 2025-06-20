package cli

import (
	"fmt"
	"os"

	"github.com/BurnDevice/BurnDevice/internal/config"
	"github.com/spf13/cobra"
)

// NewValidateCommand creates the validate command
func NewValidateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate configurations and scenarios",
		Long:  "验证配置文件和场景的有效性",
	}

	cmd.AddCommand(
		newValidateConfigCommand(),
	)

	return cmd
}

func newValidateConfigCommand() *cobra.Command {
	var configFile string

	cmd := &cobra.Command{
		Use:   "config",
		Short: "Validate configuration file",
		Long:  "验证配置文件的有效性",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Check if file exists
			if _, err := os.Stat(configFile); os.IsNotExist(err) {
				return fmt.Errorf("configuration file does not exist: %s", configFile)
			}

			// Load and validate configuration
			cfg, err := config.Load(configFile)
			if err != nil {
				return fmt.Errorf("configuration validation failed: %w", err)
			}

			// Display validation results
			fmt.Printf("✅ Configuration file is valid: %s\n", configFile)
			fmt.Printf("\n📋 Configuration Summary:\n")
			fmt.Printf("  Server: %s:%d\n", cfg.Server.Host, cfg.Server.Port)
			fmt.Printf("  AI Provider: %s\n", cfg.AI.Provider)
			fmt.Printf("  AI Model: %s\n", cfg.AI.Model)
			fmt.Printf("  Max Severity: %s\n", cfg.Security.MaxSeverity)
			fmt.Printf("  Safe Mode: %v\n", cfg.Security.EnableSafeMode)
			fmt.Printf("  Require Confirmation: %v\n", cfg.Security.RequireConfirmation)
			fmt.Printf("  Audit Log: %v\n", cfg.Security.AuditLog)
			fmt.Printf("  Log Level: %s\n", cfg.LogLevel)

			if len(cfg.Security.AllowedTargets) > 0 {
				fmt.Printf("\n✅ Allowed Targets:\n")
				for _, target := range cfg.Security.AllowedTargets {
					fmt.Printf("  - %s\n", target)
				}
			}

			if len(cfg.Security.BlockedTargets) > 0 {
				fmt.Printf("\n🚫 Blocked Targets:\n")
				for _, target := range cfg.Security.BlockedTargets {
					fmt.Printf("  - %s\n", target)
				}
			}

			// Security warnings
			if !cfg.Security.EnableSafeMode {
				fmt.Printf("\n⚠️  WARNING: Safe mode is disabled - real destructive operations will be performed!\n")
			}

			if !cfg.Security.RequireConfirmation {
				fmt.Printf("\n⚠️  WARNING: Confirmation requirement is disabled!\n")
			}

			if cfg.Security.MaxSeverity == "HIGH" || cfg.Security.MaxSeverity == "CRITICAL" {
				fmt.Printf("\n⚠️  WARNING: High severity operations are allowed!\n")
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&configFile, "config", "config.yaml", "Configuration file path")
	if err := cmd.MarkFlagRequired("config"); err != nil {
		// Log error but don't fail, as this is during command setup
		fmt.Printf("Warning: Failed to mark config flag as required: %v\n", err)
	}

	return cmd
}
