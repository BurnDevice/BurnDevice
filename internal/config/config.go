package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config represents the application configuration
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	AI       AIConfig       `mapstructure:"ai"`
	Security SecurityConfig `mapstructure:"security"`
	LogLevel string         `mapstructure:"log_level"`
}

// ServerConfig contains server-related configuration
type ServerConfig struct {
	Host         string        `mapstructure:"host"`
	Port         int           `mapstructure:"port"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	TLS          TLSConfig     `mapstructure:"tls"`
}

// TLSConfig contains TLS configuration
type TLSConfig struct {
	Enabled  bool   `mapstructure:"enabled"`
	CertFile string `mapstructure:"cert_file"`
	KeyFile  string `mapstructure:"key_file"`
}

// AIConfig contains AI service configuration
type AIConfig struct {
	Provider       string        `mapstructure:"provider"`
	APIKey         string        `mapstructure:"api_key"`
	BaseURL        string        `mapstructure:"base_url"`
	Model          string        `mapstructure:"model"`
	MaxTokens      int           `mapstructure:"max_tokens"`
	Temperature    float64       `mapstructure:"temperature"`
	RequestTimeout time.Duration `mapstructure:"request_timeout"`
}

// SecurityConfig contains security-related configuration
type SecurityConfig struct {
	RequireConfirmation bool     `mapstructure:"require_confirmation"`
	AllowedTargets      []string `mapstructure:"allowed_targets"`
	BlockedTargets      []string `mapstructure:"blocked_targets"`
	MaxSeverity         string   `mapstructure:"max_severity"`
	EnableSafeMode      bool     `mapstructure:"enable_safe_mode"`
	AuditLog            bool     `mapstructure:"audit_log"`
}

// Load loads configuration from file and environment variables
func Load(configFile string) (*Config, error) {
	// Set defaults
	setDefaults()

	// Configure viper
	viper.SetConfigType("yaml")
	viper.AutomaticEnv()
	viper.SetEnvPrefix("BURNDEVICE")
	// Enable viper to handle nested environment variables
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Load config file if specified
	if configFile != "" {
		viper.SetConfigFile(configFile)
		if err := viper.ReadInConfig(); err != nil {
			if !os.IsNotExist(err) {
				return nil, fmt.Errorf("failed to read config file: %w", err)
			}
		}
	}

	// Unmarshal configuration
	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate configuration
	if err := validate(&cfg); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &cfg, nil
}

func setDefaults() {
	// Server defaults
	viper.SetDefault("server.host", "localhost")
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.read_timeout", 30*time.Second)
	viper.SetDefault("server.write_timeout", 30*time.Second)
	viper.SetDefault("server.tls.enabled", false)

	// AI defaults
	viper.SetDefault("ai.provider", "deepseek")
	viper.SetDefault("ai.base_url", "https://api.deepseek.com")
	viper.SetDefault("ai.model", "deepseek-chat")
	viper.SetDefault("ai.max_tokens", 4096)
	viper.SetDefault("ai.temperature", 0.7)
	viper.SetDefault("ai.request_timeout", 30*time.Second)

	// Security defaults
	viper.SetDefault("security.require_confirmation", true)
	viper.SetDefault("security.max_severity", "MEDIUM")
	viper.SetDefault("security.enable_safe_mode", true)
	viper.SetDefault("security.audit_log", true)
	viper.SetDefault("security.blocked_targets", []string{
		"/",
		"/bin",
		"/usr",
		"/etc",
		"/var",
		"/home",
		"/root",
		"C:\\Windows",
		"C:\\Program Files",
		"C:\\Users",
	})

	// Logging defaults
	viper.SetDefault("log_level", "info")
}

func validate(cfg *Config) error {
	// Validate server configuration
	if cfg.Server.Port < 1 || cfg.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", cfg.Server.Port)
	}

	// Validate TLS configuration
	if cfg.Server.TLS.Enabled {
		if cfg.Server.TLS.CertFile == "" || cfg.Server.TLS.KeyFile == "" {
			return fmt.Errorf("TLS enabled but cert_file or key_file not specified")
		}
	}

	// Validate AI configuration
	if cfg.AI.Provider == "" {
		return fmt.Errorf("AI provider not specified")
	}

	// Validate security configuration
	validSeverities := []string{"LOW", "MEDIUM", "HIGH", "CRITICAL"}
	validSeverity := false
	for _, s := range validSeverities {
		if cfg.Security.MaxSeverity == s {
			validSeverity = true
			break
		}
	}
	if !validSeverity {
		return fmt.Errorf("invalid max_severity: %s", cfg.Security.MaxSeverity)
	}

	return nil
}
