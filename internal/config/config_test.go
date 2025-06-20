package config

import (
	"os"
	"testing"
	"time"
)

func TestLoad(t *testing.T) {
	// Test loading with default values
	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Failed to load default config: %v", err)
	}

	// Verify defaults
	if cfg.Server.Host != "localhost" {
		t.Errorf("Expected default host 'localhost', got '%s'", cfg.Server.Host)
	}

	if cfg.Server.Port != 8080 {
		t.Errorf("Expected default port 8080, got %d", cfg.Server.Port)
	}

	if cfg.LogLevel != "info" {
		t.Errorf("Expected default log level 'info', got '%s'", cfg.LogLevel)
	}

	if !cfg.Security.RequireConfirmation {
		t.Error("Expected require_confirmation to be true by default")
	}

	if cfg.Security.MaxSeverity != "MEDIUM" {
		t.Errorf("Expected default max severity 'MEDIUM', got '%s'", cfg.Security.MaxSeverity)
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name      string
		cfg       *Config
		expectErr bool
	}{
		{
			name: "valid config",
			cfg: &Config{
				Server: ServerConfig{
					Host: "localhost",
					Port: 8080,
				},
				Security: SecurityConfig{
					MaxSeverity: "MEDIUM",
				},
				AI: AIConfig{
					Provider: "deepseek",
				},
			},
			expectErr: false,
		},
		{
			name: "invalid port",
			cfg: &Config{
				Server: ServerConfig{
					Host: "localhost",
					Port: 99999,
				},
				Security: SecurityConfig{
					MaxSeverity: "MEDIUM",
				},
				AI: AIConfig{
					Provider: "deepseek",
				},
			},
			expectErr: true,
		},
		{
			name: "invalid severity",
			cfg: &Config{
				Server: ServerConfig{
					Host: "localhost",
					Port: 8080,
				},
				Security: SecurityConfig{
					MaxSeverity: "INVALID",
				},
				AI: AIConfig{
					Provider: "deepseek",
				},
			},
			expectErr: true,
		},
		{
			name: "missing AI provider",
			cfg: &Config{
				Server: ServerConfig{
					Host: "localhost",
					Port: 8080,
				},
				Security: SecurityConfig{
					MaxSeverity: "MEDIUM",
				},
				AI: AIConfig{
					Provider: "",
				},
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validate(tt.cfg)
			if tt.expectErr && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestEnvironmentVariables(t *testing.T) {
	// Clear any existing environment variables first
	os.Unsetenv("BURNDEVICE_SERVER_HOST")
	os.Unsetenv("BURNDEVICE_SERVER_PORT")

	// Set environment variables with correct viper format
	err := os.Setenv("BURNDEVICE_SERVER_HOST", "test-host")
	if err != nil {
		t.Fatalf("Failed to set env var: %v", err)
	}
	err = os.Setenv("BURNDEVICE_SERVER_PORT", "9090")
	if err != nil {
		t.Fatalf("Failed to set env var: %v", err)
	}

	defer func() {
		os.Unsetenv("BURNDEVICE_SERVER_HOST")
		os.Unsetenv("BURNDEVICE_SERVER_PORT")
	}()

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if cfg.Server.Host != "test-host" {
		t.Errorf("Expected host from env var 'test-host', got '%s'", cfg.Server.Host)
	}

	if cfg.Server.Port != 9090 {
		t.Errorf("Expected port from env var 9090, got %d", cfg.Server.Port)
	}
}

func TestTLSValidation(t *testing.T) {
	cfg := &Config{
		Server: ServerConfig{
			Host: "localhost",
			Port: 8080,
			TLS: TLSConfig{
				Enabled:  true,
				CertFile: "",
				KeyFile:  "",
			},
		},
		Security: SecurityConfig{
			MaxSeverity: "MEDIUM",
		},
		AI: AIConfig{
			Provider: "deepseek",
		},
	}

	err := validate(cfg)
	if err == nil {
		t.Error("Expected error for TLS enabled without cert/key files")
	}
}

func TestTimeoutDefaults(t *testing.T) {
	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	expectedTimeout := 30 * time.Second
	if cfg.Server.ReadTimeout != expectedTimeout {
		t.Errorf("Expected read timeout %v, got %v", expectedTimeout, cfg.Server.ReadTimeout)
	}

	if cfg.Server.WriteTimeout != expectedTimeout {
		t.Errorf("Expected write timeout %v, got %v", expectedTimeout, cfg.Server.WriteTimeout)
	}

	if cfg.AI.RequestTimeout != expectedTimeout {
		t.Errorf("Expected AI request timeout %v, got %v", expectedTimeout, cfg.AI.RequestTimeout)
	}
}
