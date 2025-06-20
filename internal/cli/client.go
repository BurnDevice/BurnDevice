package cli

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/BurnDevice/BurnDevice/burndevice/v1"
)

// NewClientCommand creates the client command
func NewClientCommand() *cobra.Command {
	var serverAddr string
	var timeout time.Duration

	cmd := &cobra.Command{
		Use:   "client",
		Short: "BurnDevice client commands",
		Long:  "与 BurnDevice 服务器交互的客户端命令",
	}

	cmd.PersistentFlags().StringVar(&serverAddr, "server", "localhost:8080", "Server address")
	cmd.PersistentFlags().DurationVar(&timeout, "timeout", 30*time.Second, "Request timeout")

	// Add subcommands
	cmd.AddCommand(
		newExecuteCommand(),
		newSystemInfoCommand(),
		newGenerateScenarioCommand(),
		newStreamCommand(),
	)

	return cmd
}

func newExecuteCommand() *cobra.Command {
	var (
		destructionType string
		targets         []string
		severity        string
		confirm         bool
		scenarioID      string
	)

	cmd := &cobra.Command{
		Use:   "execute",
		Short: "Execute a destruction request",
		Long:  "执行破坏性测试请求",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !confirm {
				return fmt.Errorf("必须使用 --confirm 标志确认破坏性操作")
			}

			client, conn, err := createClient(cmd)
			if err != nil {
				return err
			}
			defer conn.Close()

			// Parse destruction type
			dtype, err := parseDestructionType(destructionType)
			if err != nil {
				return err
			}

			// Parse severity
			sev, err := parseSeverity(severity)
			if err != nil {
				return err
			}

			req := &pb.ExecuteDestructionRequest{
				Type:               dtype,
				Targets:            targets,
				Severity:           sev,
				ConfirmDestruction: confirm,
				AiScenarioId:       scenarioID,
			}

			ctx, cancel := context.WithTimeout(context.Background(), getTimeout(cmd))
			defer cancel()

			logrus.WithFields(logrus.Fields{
				"type":     destructionType,
				"targets":  targets,
				"severity": severity,
			}).Warn("🔥 Executing destruction request")

			resp, err := client.ExecuteDestruction(ctx, req)
			if err != nil {
				return fmt.Errorf("execution failed: %w", err)
			}

			// Display results
			fmt.Printf("✅ Execution completed: %s\n", resp.Message)
			fmt.Printf("Success: %v\n", resp.Success)
			fmt.Printf("Results: %d\n", len(resp.Results))

			for i, result := range resp.Results {
				fmt.Printf("\nResult %d:\n", i+1)
				fmt.Printf("  Target: %s\n", result.Target)
				fmt.Printf("  Success: %v\n", result.Success)
				if result.ErrorMessage != "" {
					fmt.Printf("  Error: %s\n", result.ErrorMessage)
				}
				if result.Metrics != nil {
					fmt.Printf("  Files deleted: %d\n", result.Metrics.FilesDeleted)
					fmt.Printf("  Bytes destroyed: %d\n", result.Metrics.BytesDestroyed)
					fmt.Printf("  Execution time: %.2fs\n", result.Metrics.ExecutionTimeSeconds)
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&destructionType, "type", "", "Destruction type (required)")
	cmd.Flags().StringSliceVar(&targets, "targets", []string{}, "Target paths")
	cmd.Flags().StringVar(&severity, "severity", "LOW", "Destruction severity (LOW, MEDIUM, HIGH, CRITICAL)")
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Confirm destructive operation")
	cmd.Flags().StringVar(&scenarioID, "scenario-id", "", "AI scenario ID")

	if err := cmd.MarkFlagRequired("type"); err != nil {
		logrus.WithError(err).Error("Failed to mark type flag as required")
	}

	return cmd
}

func newSystemInfoCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "system-info",
		Short: "Get system information",
		Long:  "获取系统信息",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, conn, err := createClient(cmd)
			if err != nil {
				return err
			}
			defer conn.Close()

			ctx, cancel := context.WithTimeout(context.Background(), getTimeout(cmd))
			defer cancel()

			resp, err := client.GetSystemInfo(ctx, &pb.GetSystemInfoRequest{})
			if err != nil {
				return fmt.Errorf("failed to get system info: %w", err)
			}

			// Display system information
			fmt.Printf("💻 System Information\n")
			fmt.Printf("OS: %s\n", resp.Os)
			fmt.Printf("Architecture: %s\n", resp.Architecture)
			fmt.Printf("Hostname: %s\n", resp.Hostname)

			if resp.Resources != nil {
				fmt.Printf("\n📊 Resources:\n")
				fmt.Printf("  Total Memory: %d GB\n", resp.Resources.TotalMemory/(1024*1024*1024))
				fmt.Printf("  Available Memory: %d GB\n", resp.Resources.AvailableMemory/(1024*1024*1024))
				fmt.Printf("  Total Disk: %d GB\n", resp.Resources.TotalDisk/(1024*1024*1024))
				fmt.Printf("  Available Disk: %d GB\n", resp.Resources.AvailableDisk/(1024*1024*1024))
				fmt.Printf("  CPU Usage: %.2f%%\n", resp.Resources.CpuUsage)
			}

			if len(resp.CriticalPaths) > 0 {
				fmt.Printf("\n🚨 Critical Paths:\n")
				for _, path := range resp.CriticalPaths {
					fmt.Printf("  - %s\n", path)
				}
			}

			if len(resp.RunningServices) > 0 {
				fmt.Printf("\n🔧 Running Services:\n")
				for _, service := range resp.RunningServices {
					fmt.Printf("  - %s\n", service)
				}
			}

			return nil
		},
	}

	return cmd
}

func newGenerateScenarioCommand() *cobra.Command {
	var (
		target      string
		maxSeverity string
		aiModel     string
	)

	cmd := &cobra.Command{
		Use:   "generate-scenario",
		Short: "Generate AI attack scenario",
		Long:  "使用 AI 生成攻击场景",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, conn, err := createClient(cmd)
			if err != nil {
				return err
			}
			defer conn.Close()

			// Parse severity
			sev, err := parseSeverity(maxSeverity)
			if err != nil {
				return err
			}

			req := &pb.GenerateAttackScenarioRequest{
				TargetDescription: target,
				MaxSeverity:       sev,
				AiModel:           aiModel,
			}

			ctx, cancel := context.WithTimeout(context.Background(), getTimeout(cmd))
			defer cancel()

			logrus.WithFields(logrus.Fields{
				"target":       target,
				"max_severity": maxSeverity,
				"model":        aiModel,
			}).Info("🤖 Generating AI attack scenario")

			resp, err := client.GenerateAttackScenario(ctx, req)
			if err != nil {
				return fmt.Errorf("scenario generation failed: %w", err)
			}

			// Display scenario
			fmt.Printf("🤖 AI Generated Attack Scenario\n")
			fmt.Printf("ID: %s\n", resp.ScenarioId)
			fmt.Printf("Description: %s\n", resp.Description)
			fmt.Printf("Estimated Severity: %s\n", resp.EstimatedSeverity.String())
			fmt.Printf("\n📋 Steps:\n")

			for _, step := range resp.Steps {
				fmt.Printf("\n%d. %s\n", step.Order, step.Description)
				fmt.Printf("   Type: %s\n", step.Type.String())
				if len(step.Targets) > 0 {
					fmt.Printf("   Targets: %s\n", strings.Join(step.Targets, ", "))
				}
				if step.Rationale != "" {
					fmt.Printf("   Rationale: %s\n", step.Rationale)
				}
			}

			fmt.Printf("\n💡 Use scenario ID '%s' with the execute command\n", resp.ScenarioId)

			return nil
		},
	}

	cmd.Flags().StringVar(&target, "target", "", "Target description (required)")
	cmd.Flags().StringVar(&maxSeverity, "max-severity", "MEDIUM", "Maximum severity (LOW, MEDIUM, HIGH, CRITICAL)")
	cmd.Flags().StringVar(&aiModel, "model", "", "AI model to use")

	if err := cmd.MarkFlagRequired("target"); err != nil {
		logrus.WithError(err).Error("Failed to mark target flag as required")
	}

	return cmd
}

func newStreamCommand() *cobra.Command {
	var (
		destructionType string
		targets         []string
		severity        string
		confirm         bool
		scenarioID      string
	)

	cmd := &cobra.Command{
		Use:   "stream",
		Short: "Stream destruction progress",
		Long:  "实时流式监控破坏进度",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !confirm {
				return fmt.Errorf("必须使用 --confirm 标志确认破坏性操作")
			}

			client, conn, err := createClient(cmd)
			if err != nil {
				return err
			}
			defer conn.Close()

			// Parse destruction type
			dtype, err := parseDestructionType(destructionType)
			if err != nil {
				return err
			}

			// Parse severity
			sev, err := parseSeverity(severity)
			if err != nil {
				return err
			}

			req := &pb.StreamDestructionRequest{
				Type:               dtype,
				Targets:            targets,
				Severity:           sev,
				ConfirmDestruction: confirm,
				AiScenarioId:       scenarioID,
			}

			ctx, cancel := context.WithTimeout(context.Background(), getTimeout(cmd))
			defer cancel()

			logrus.Info("🔥 Starting streaming destruction...")

			stream, err := client.StreamDestruction(ctx, req)
			if err != nil {
				return fmt.Errorf("failed to start stream: %w", err)
			}

			// Stream events
			for {
				event, err := stream.Recv()
				if err != nil {
					break
				}

				timestamp := event.Timestamp.AsTime().Format("15:04:05")
				switch event.Type {
				case pb.DestructionEventType_DESTRUCTION_EVENT_TYPE_STARTED:
					fmt.Printf("[%s] 🚀 Started: %s\n", timestamp, event.Message)
				case pb.DestructionEventType_DESTRUCTION_EVENT_TYPE_PROGRESS:
					fmt.Printf("[%s] ⏳ Progress: %.1f%% - %s\n", timestamp, event.Progress*100, event.Message)
				case pb.DestructionEventType_DESTRUCTION_EVENT_TYPE_COMPLETED:
					fmt.Printf("[%s] ✅ Completed: %s\n", timestamp, event.Message)
				case pb.DestructionEventType_DESTRUCTION_EVENT_TYPE_ERROR:
					fmt.Printf("[%s] ❌ Error: %s\n", timestamp, event.Message)
				case pb.DestructionEventType_DESTRUCTION_EVENT_TYPE_WARNING:
					fmt.Printf("[%s] ⚠️  Warning: %s\n", timestamp, event.Message)
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&destructionType, "type", "", "Destruction type (required)")
	cmd.Flags().StringSliceVar(&targets, "targets", []string{}, "Target paths")
	cmd.Flags().StringVar(&severity, "severity", "LOW", "Destruction severity")
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Confirm destructive operation")
	cmd.Flags().StringVar(&scenarioID, "scenario-id", "", "AI scenario ID")

	if err := cmd.MarkFlagRequired("type"); err != nil {
		logrus.WithError(err).Error("Failed to mark type flag as required")
	}

	return cmd
}

// Helper functions
func createClient(cmd *cobra.Command) (pb.BurnDeviceServiceClient, *grpc.ClientConn, error) {
	serverAddr, _ := cmd.Flags().GetString("server")

	conn, err := grpc.Dial(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to server: %w", err)
	}

	client := pb.NewBurnDeviceServiceClient(conn)
	return client, conn, nil
}

func getTimeout(cmd *cobra.Command) time.Duration {
	timeout, _ := cmd.Flags().GetDuration("timeout")
	return timeout
}

func parseDestructionType(typeStr string) (pb.DestructionType, error) {
	switch strings.ToUpper(typeStr) {
	case "FILE_DELETION":
		return pb.DestructionType_DESTRUCTION_TYPE_FILE_DELETION, nil
	case "SERVICE_TERMINATION":
		return pb.DestructionType_DESTRUCTION_TYPE_SERVICE_TERMINATION, nil
	case "MEMORY_EXHAUSTION":
		return pb.DestructionType_DESTRUCTION_TYPE_MEMORY_EXHAUSTION, nil
	case "DISK_FILL":
		return pb.DestructionType_DESTRUCTION_TYPE_DISK_FILL, nil
	case "NETWORK_DISRUPTION":
		return pb.DestructionType_DESTRUCTION_TYPE_NETWORK_DISRUPTION, nil
	case "BOOT_CORRUPTION":
		return pb.DestructionType_DESTRUCTION_TYPE_BOOT_CORRUPTION, nil
	case "KERNEL_PANIC":
		return pb.DestructionType_DESTRUCTION_TYPE_KERNEL_PANIC, nil
	default:
		return pb.DestructionType_DESTRUCTION_TYPE_UNSPECIFIED, fmt.Errorf("unknown destruction type: %s", typeStr)
	}
}

func parseSeverity(severityStr string) (pb.DestructionSeverity, error) {
	switch strings.ToUpper(severityStr) {
	case "LOW":
		return pb.DestructionSeverity_DESTRUCTION_SEVERITY_LOW, nil
	case "MEDIUM":
		return pb.DestructionSeverity_DESTRUCTION_SEVERITY_MEDIUM, nil
	case "HIGH":
		return pb.DestructionSeverity_DESTRUCTION_SEVERITY_HIGH, nil
	case "CRITICAL":
		return pb.DestructionSeverity_DESTRUCTION_SEVERITY_CRITICAL, nil
	default:
		return pb.DestructionSeverity_DESTRUCTION_SEVERITY_UNSPECIFIED, fmt.Errorf("unknown severity: %s", severityStr)
	}
}
