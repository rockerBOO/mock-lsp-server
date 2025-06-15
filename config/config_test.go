package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config == nil {
		t.Fatal("DefaultConfig returned nil")
	}

	if config.AppName != "mock-lsp-server" {
		t.Errorf("Expected app name 'mock-lsp-server', got %s", config.AppName)
	}

	if config.Server.Name != "Mock LSP Server" {
		t.Errorf("Expected server name 'Mock LSP Server', got %s", config.Server.Name)
	}

	if config.Logging.Level != "info" {
		t.Errorf("Expected log level 'info', got %s", config.Logging.Level)
	}

	if !config.LSP.CompletionConfig.Enabled {
		t.Error("Expected completion to be enabled by default")
	}
}

func TestServerConfig_Validate(t *testing.T) {
	testCases := []struct {
		name        string
		config      *ServerConfig
		expectError bool
		errorField  string
	}{
		{
			name:        "valid default config",
			config:      DefaultConfig(),
			expectError: false,
		},
		{
			name: "empty app name",
			config: &ServerConfig{
				AppName: "",
				Server:  DefaultConfig().Server,
				Logging: DefaultConfig().Logging,
				LSP:     DefaultConfig().LSP,
			},
			expectError: true,
			errorField:  "app_name",
		},
		{
			name: "invalid log level",
			config: func() *ServerConfig {
				c := DefaultConfig()
				c.Logging.Level = "invalid"
				return c
			}(),
			expectError: true,
			errorField:  "logging.level",
		},
		{
			name: "invalid timeout",
			config: func() *ServerConfig {
				c := DefaultConfig()
				c.Server.Timeout = Duration(500 * time.Millisecond)
				return c
			}(),
			expectError: true,
			errorField:  "server.timeout",
		},
		{
			name: "invalid max requests",
			config: func() *ServerConfig {
				c := DefaultConfig()
				c.Server.MaxRequests = 0
				return c
			}(),
			expectError: true,
			errorField:  "server.max_requests",
		},
		{
			name: "invalid completion max items",
			config: func() *ServerConfig {
				c := DefaultConfig()
				c.LSP.CompletionConfig.MaxItems = 0
				return c
			}(),
			expectError: true,
			errorField:  "lsp.completion.max_items",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.config.Validate()

			if tc.expectError {
				if err == nil {
					t.Error("Expected validation error, got nil")
					return
				}

				// Check if the expected field is in the error
				if tc.errorField != "" && !strings.Contains(err.Error(), tc.errorField) {
					t.Errorf("Expected error to contain field '%s', got: %v", tc.errorField, err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no validation error, got: %v", err)
				}
			}
		})
	}
}

func TestValidationErrors(t *testing.T) {
	// Test single validation error
	singleError := ValidationErrors{
		{Field: "test", Value: "value", Message: "test message"},
	}

	errorStr := singleError.Error()
	if !strings.Contains(errorStr, "test") || !strings.Contains(errorStr, "test message") {
		t.Errorf("Single error string doesn't contain expected content: %s", errorStr)
	}

	// Test multiple validation errors
	multipleErrors := ValidationErrors{
		{Field: "field1", Value: "value1", Message: "message1"},
		{Field: "field2", Value: "value2", Message: "message2"},
		{Field: "field3", Value: "value3", Message: "message3"},
	}

	multiErrorStr := multipleErrors.Error()
	if !strings.Contains(multiErrorStr, "3 validation errors") {
		t.Errorf("Multiple errors string doesn't indicate count: %s", multiErrorStr)
	}

	// Test empty validation errors
	emptyErrors := ValidationErrors{}
	if emptyErrors.Error() != "no validation errors" {
		t.Errorf("Empty errors should return 'no validation errors', got: %s", emptyErrors.Error())
	}
}

func TestLoadFromFile(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test_config.json")

	// Test valid config file
	validConfig := `{
		"app_name": "test-server",
		"server": {
			"name": "Test Server",
			"version": "0.1.0",
			"description": "Test server",
			"timeout": "15s",
			"max_requests": 500
		},
		"logging": {
			"level": "debug",
			"file_name": "test.log",
			"max_size_mb": 25,
			"max_backups": 5,
			"compress": false,
			"format": "json"
		},
		"lsp": {
			"initialize_timeout": "5s",
			"completion": {
				"enabled": true,
				"max_items": 50,
				"case_sensitive": true
			},
			"hover": {
				"enabled": true,
				"max_length": 500
			},
			"diagnostics": {
				"enabled": false,
				"max_issues": 25
			},
			"mock_data": {
				"enabled": true,
				"item_count": 25
			}
		}
	}`

	err := os.WriteFile(configPath, []byte(validConfig), 0644)
	if err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	config, err := LoadFromFile(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if config.AppName != "test-server" {
		t.Errorf("Expected app name 'test-server', got %s", config.AppName)
	}

	if config.Server.Name != "Test Server" {
		t.Errorf("Expected server name 'Test Server', got %s", config.Server.Name)
	}

	if config.Logging.Level != "debug" {
		t.Errorf("Expected log level 'debug', got %s", config.Logging.Level)
	}

	if config.LSP.CompletionConfig.MaxItems != 50 {
		t.Errorf("Expected completion max items 50, got %d", config.LSP.CompletionConfig.MaxItems)
	}

	// Test non-existent file
	_, err = LoadFromFile("/non/existent/path.json")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}

	// Test invalid JSON
	invalidPath := filepath.Join(tmpDir, "invalid.json")
	err = os.WriteFile(invalidPath, []byte(`{invalid json`), 0644)
	if err != nil {
		t.Fatalf("Failed to write invalid config: %v", err)
	}

	_, err = LoadFromFile(invalidPath)
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}
}

func TestLoadFromFileWithDefaults(t *testing.T) {
	// Test with empty path (should return defaults)
	config, err := LoadFromFileWithDefaults("")
	if err != nil {
		t.Fatalf("Failed to load defaults: %v", err)
	}

	defaultConfig := DefaultConfig()
	if config.AppName != defaultConfig.AppName {
		t.Errorf("Expected default app name, got %s", config.AppName)
	}

	// Test with non-existent path (should return defaults)
	config, err = LoadFromFileWithDefaults("/non/existent/path.json")
	if err != nil {
		t.Fatalf("Failed to load defaults for non-existent file: %v", err)
	}

	if config.AppName != defaultConfig.AppName {
		t.Errorf("Expected default app name for non-existent file, got %s", config.AppName)
	}

	// Test with partial config file (should merge with defaults)
	tmpDir := t.TempDir()
	partialPath := filepath.Join(tmpDir, "partial.json")
	partialConfig := `{
		"app_name": "partial-server",
		"logging": {
			"level": "debug"
		}
	}`

	err = os.WriteFile(partialPath, []byte(partialConfig), 0644)
	if err != nil {
		t.Fatalf("Failed to write partial config: %v", err)
	}

	config, err = LoadFromFileWithDefaults(partialPath)
	if err != nil {
		t.Fatalf("Failed to load partial config: %v", err)
	}

	// Should have partial values
	if config.AppName != "partial-server" {
		t.Errorf("Expected app name 'partial-server', got %s", config.AppName)
	}

	if config.Logging.Level != "debug" {
		t.Errorf("Expected log level 'debug', got %s", config.Logging.Level)
	}

	// Should have default values for unspecified fields
	if config.Server.Name != defaultConfig.Server.Name {
		t.Errorf("Expected default server name, got %s", config.Server.Name)
	}
}

func TestSaveToFile(t *testing.T) {
	config := DefaultConfig()
	config.AppName = "save-test"

	tmpDir := t.TempDir()
	savePath := filepath.Join(tmpDir, "saved_config.json")

	err := config.SaveToFile(savePath)
	if err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(savePath); os.IsNotExist(err) {
		t.Error("Config file was not created")
	}

	// Load and verify content
	loadedConfig, err := LoadFromFile(savePath)
	if err != nil {
		t.Fatalf("Failed to load saved config: %v", err)
	}

	if loadedConfig.AppName != "save-test" {
		t.Errorf("Expected app name 'save-test', got %s", loadedConfig.AppName)
	}

	// Test saving to nested directory
	nestedPath := filepath.Join(tmpDir, "nested", "dir", "config.json")
	err = config.SaveToFile(nestedPath)
	if err != nil {
		t.Fatalf("Failed to save config to nested path: %v", err)
	}

	if _, err := os.Stat(nestedPath); os.IsNotExist(err) {
		t.Error("Config file was not created in nested directory")
	}
}

func TestMergeConfigs(t *testing.T) {
	base := DefaultConfig()
	override := &ServerConfig{
		AppName: "override-app",
		Server: ServerSettings{
			Name:    "Override Server",
			Version: "2.0.0",
			// Other fields should remain default
		},
		Logging: LoggingConfig{
			Level: "debug",
			// Other fields should remain default
		},
		LSP: LSPConfig{
			InitializeTimeout: Duration(5 * time.Second),
			// Other fields should remain default
		},
	}

	merged := mergeConfigs(base, override)

	// Check overridden values
	if merged.AppName != "override-app" {
		t.Errorf("Expected app name 'override-app', got %s", merged.AppName)
	}

	if merged.Server.Name != "Override Server" {
		t.Errorf("Expected server name 'Override Server', got %s", merged.Server.Name)
	}

	if merged.Server.Version != "2.0.0" {
		t.Errorf("Expected server version '2.0.0', got %s", merged.Server.Version)
	}

	if merged.Logging.Level != "debug" {
		t.Errorf("Expected log level 'debug', got %s", merged.Logging.Level)
	}

	if merged.LSP.InitializeTimeout.Duration() != 5*time.Second {
		t.Errorf("Expected initialize timeout 5s, got %v", merged.LSP.InitializeTimeout)
	}

	// Check that default values are preserved
	if merged.Server.Description != base.Server.Description {
		t.Errorf("Expected default server description to be preserved")
	}

	if merged.Logging.MaxSize != base.Logging.MaxSize {
		t.Errorf("Expected default log max size to be preserved")
	}

	if merged.LSP.CompletionConfig.MaxItems != base.LSP.CompletionConfig.MaxItems {
		t.Errorf("Expected default completion max items to be preserved")
	}
}

func TestConfigValidation_EdgeCases(t *testing.T) {
	// Test config with very long app name
	config := DefaultConfig()
	config.AppName = strings.Repeat("a", 101) // 101 characters

	err := config.Validate()
	if err == nil {
		t.Error("Expected validation error for long app name")
	}

	// Test config with negative values
	config = DefaultConfig()
	config.Server.MaxRequests = -1
	config.Logging.MaxSize = -1
	config.LSP.CompletionConfig.MaxItems = -1

	err = config.Validate()
	if err == nil {
		t.Error("Expected validation error for negative values")
	}

	// Should have multiple validation errors
	if validationErrs, ok := err.(ValidationErrors); ok {
		if len(validationErrs) < 2 {
			t.Errorf("Expected multiple validation errors, got %d", len(validationErrs))
		}
	} else {
		t.Error("Expected ValidationErrors type")
	}
}
