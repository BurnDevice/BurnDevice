package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/BurnDevice/BurnDevice/internal/cli"
	"github.com/BurnDevice/BurnDevice/internal/config"
	"github.com/BurnDevice/BurnDevice/internal/server"
)

var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "burndevice",
		Short: "🔥 BurnDevice - 设备破坏性测试工具",
		Long: `BurnDevice 是一个用于测试环境的破坏性测试工具。

⚠️  警告：此工具仅用于授权的测试环境中，绝不可在生产环境使用！

支持的功能：
  - 文件系统破坏测试
  - 系统服务中断测试  
  - 内存和磁盘耗尽测试
  - AI 驱动的攻击场景生成
  - 实时监控和日志记录`,
		Version: fmt.Sprintf("%s (commit: %s, built: %s)", version, commit, date),
	}

	// Add subcommands
	rootCmd.AddCommand(
		newServerCmd(),
		newClientCmd(),
		newGenerateCmd(),
		newValidateCmd(),
	)

	if err := rootCmd.Execute(); err != nil {
		logrus.WithError(err).Fatal("Failed to execute command")
	}
}

func newServerCmd() *cobra.Command {
	var configFile string

	cmd := &cobra.Command{
		Use:   "server",
		Short: "Start the BurnDevice gRPC server",
		Long:  "启动 BurnDevice gRPC 服务器，监听破坏性测试请求",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Load configuration
			cfg, err := config.Load(configFile)
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			// Setup logging
			setupLogging(cfg.LogLevel)

			logrus.WithFields(logrus.Fields{
				"version": version,
				"commit":  commit,
				"config":  configFile,
			}).Info("🔥 Starting BurnDevice server")

			// Create server
			srv, err := server.New(cfg)
			if err != nil {
				return fmt.Errorf("failed to create server: %w", err)
			}

			// Setup graceful shutdown
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

			go func() {
				<-sigChan
				logrus.Info("Received shutdown signal, gracefully stopping...")
				cancel()
			}()

			// Start server
			if err := srv.Start(ctx); err != nil {
				return fmt.Errorf("server failed: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&configFile, "config", "c", "config.yaml", "Configuration file path")

	return cmd
}

func newClientCmd() *cobra.Command {
	return cli.NewClientCommand()
}

func newGenerateCmd() *cobra.Command {
	return cli.NewGenerateCommand()
}

func newValidateCmd() *cobra.Command {
	return cli.NewValidateCommand()
}

func setupLogging(level string) {
	logrus.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
	})

	switch level {
	case "debug":
		logrus.SetLevel(logrus.DebugLevel)
	case "info":
		logrus.SetLevel(logrus.InfoLevel)
	case "warn":
		logrus.SetLevel(logrus.WarnLevel)
	case "error":
		logrus.SetLevel(logrus.ErrorLevel)
	default:
		logrus.SetLevel(logrus.InfoLevel)
	}

	logrus.WithField("level", logrus.GetLevel()).Info("Logging configured")
}
