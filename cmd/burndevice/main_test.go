package main

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func TestMain(m *testing.M) {
	// 设置测试环境
	logrus.SetLevel(logrus.FatalLevel) // 减少测试期间的日志输出
	code := m.Run()
	os.Exit(code)
}

func TestRootCommand(t *testing.T) {
	// 创建root命令
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
		Version: "test-version",
	}

	// 添加子命令
	rootCmd.AddCommand(
		newServerCmd(),
		newClientCmd(),
		newGenerateCmd(),
		newValidateCmd(),
	)

	// 测试根命令属性
	if rootCmd.Use != "burndevice" {
		t.Errorf("Expected Use 'burndevice', got '%s'", rootCmd.Use)
	}

	if rootCmd.Short == "" {
		t.Error("Expected Short description to be set")
	}

	if rootCmd.Long == "" {
		t.Error("Expected Long description to be set")
	}

	// 测试子命令是否正确添加
	expectedSubcommands := []string{"server", "client", "generate", "validate"}
	actualSubcommands := make([]string, 0, len(rootCmd.Commands()))
	for _, cmd := range rootCmd.Commands() {
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
			t.Errorf("Expected subcommand '%s' not found", expected)
		}
	}
}

func TestVersionInfo(t *testing.T) {
	// 测试版本变量
	originalVersion := version
	originalCommit := commit
	originalDate := date

	// 设置测试值
	version = "1.0.0"
	commit = "abc123"
	date = "2024-01-01"

	defer func() {
		// 恢复原始值
		version = originalVersion
		commit = originalCommit
		date = originalDate
	}()

	// 创建带版本信息的命令
	rootCmd := &cobra.Command{
		Use:     "burndevice",
		Version: version + " (commit: " + commit + ", built: " + date + ")",
	}

	expectedVersion := "1.0.0 (commit: abc123, built: 2024-01-01)"
	if rootCmd.Version != expectedVersion {
		t.Errorf("Expected version '%s', got '%s'", expectedVersion, rootCmd.Version)
	}
}

func TestNewServerCmd(t *testing.T) {
	cmd := newServerCmd()

	if cmd.Use != "server" {
		t.Errorf("Expected Use 'server', got '%s'", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("Expected Short description to be set")
	}

	// 检查配置文件标志
	flag := cmd.Flags().Lookup("config")
	if flag == nil {
		t.Error("Expected 'config' flag to be defined")
		return
	}

	if flag.DefValue != "config.yaml" {
		t.Errorf("Expected default config value 'config.yaml', got '%s'", flag.DefValue)
	}

	// 测试短标志
	shortFlag := cmd.Flags().ShorthandLookup("c")
	if shortFlag == nil {
		t.Error("Expected 'c' shorthand flag for config")
	}
}

func TestNewClientCmd(t *testing.T) {
	cmd := newClientCmd()

	if cmd.Use != "client" {
		t.Errorf("Expected Use 'client', got '%s'", cmd.Use)
	}

	// 验证这是从cli包返回的命令
	if len(cmd.Commands()) == 0 {
		t.Error("Expected client command to have subcommands")
	}
}

func TestNewGenerateCmd(t *testing.T) {
	cmd := newGenerateCmd()

	// 验证命令基本属性
	if cmd == nil {
		t.Fatal("Expected generate command to be created")
	}

	// 验证这是从cli包返回的命令
	if cmd.Use == "" {
		t.Error("Expected generate command to have Use field set")
	}
}

func TestNewValidateCmd(t *testing.T) {
	cmd := newValidateCmd()

	// 验证命令基本属性
	if cmd == nil {
		t.Fatal("Expected validate command to be created")
	}

	// 验证这是从cli包返回的命令
	if cmd.Use == "" {
		t.Error("Expected validate command to have Use field set")
	}
}

func TestSetupLogging(t *testing.T) {
	originalLevel := logrus.GetLevel()
	defer logrus.SetLevel(originalLevel)

	tests := []struct {
		level    string
		expected logrus.Level
	}{
		{"debug", logrus.DebugLevel},
		{"info", logrus.InfoLevel},
		{"warn", logrus.WarnLevel},
		{"error", logrus.ErrorLevel},
		{"invalid", logrus.InfoLevel}, // 默认级别
		{"", logrus.InfoLevel},        // 默认级别
	}

	for _, tt := range tests {
		t.Run(tt.level, func(t *testing.T) {
			setupLogging(tt.level)
			if logrus.GetLevel() != tt.expected {
				t.Errorf("Expected log level %v, got %v", tt.expected, logrus.GetLevel())
			}
		})
	}
}

func TestLoggerConfiguration(t *testing.T) {
	// 测试日志格式器设置
	setupLogging("info")

	// 验证格式器类型
	formatter := logrus.StandardLogger().Formatter
	if _, ok := formatter.(*logrus.JSONFormatter); !ok {
		t.Error("Expected JSONFormatter to be set")
	}

	// 测试日志输出
	var buf bytes.Buffer
	logrus.SetOutput(&buf)
	defer logrus.SetOutput(os.Stderr)

	logrus.Info("test message")
	output := buf.String()

	// 验证JSON格式
	if !strings.Contains(output, `"level":"info"`) {
		t.Error("Expected JSON formatted log output")
	}

	if !strings.Contains(output, `"msg":"test message"`) {
		t.Error("Expected message in JSON log output")
	}
}

func TestCommandHelpOutput(t *testing.T) {
	rootCmd := &cobra.Command{
		Use:   "burndevice",
		Short: "🔥 BurnDevice - 设备破坏性测试工具",
		Long: `BurnDevice 是一个用于测试环境的破坏性测试工具。

⚠️  警告：此工具仅用于授权的测试环境中，绝不可在生产环境使用！`,
	}

	// 添加子命令
	rootCmd.AddCommand(newServerCmd())

	// 测试帮助输出
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{"--help"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("Expected no error executing help, got: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "BurnDevice") {
		t.Error("Expected help output to contain 'BurnDevice'")
	}

	if !strings.Contains(output, "警告") {
		t.Error("Expected help output to contain warning in Chinese")
	}
}

func TestVersionOutput(t *testing.T) {
	// 设置测试版本信息
	version = "1.0.0-test"
	commit = "test-commit"
	date = "2024-01-01"

	rootCmd := &cobra.Command{
		Use:     "burndevice",
		Version: version + " (commit: " + commit + ", built: " + date + ")",
	}

	// 测试版本输出
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{"--version"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("Expected no error executing version, got: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "1.0.0-test") {
		t.Error("Expected version output to contain version number")
	}

	if !strings.Contains(output, "test-commit") {
		t.Error("Expected version output to contain commit hash")
	}

	if !strings.Contains(output, "2024-01-01") {
		t.Error("Expected version output to contain build date")
	}
}

func TestGlobalVariables(t *testing.T) {
	// 测试全局变量的默认值
	if version == "" {
		t.Error("Expected version variable to be initialized")
	}

	if commit == "" {
		t.Error("Expected commit variable to be initialized")
	}

	if date == "" {
		t.Error("Expected date variable to be initialized")
	}
}
