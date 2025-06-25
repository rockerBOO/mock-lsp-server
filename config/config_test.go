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

// TestConfigIntegration_CLIOverrides tests configuration loading with CLI overrides
func TestConfigIntegration_CLIOverrides(t *testing.T) {
	// Create a base config file
	tmpDir := t.TempDir()
	baseConfigPath := filepath.Join(tmpDir, "base_config.json")

	baseConfig := DefaultConfig()
	baseConfig.AppName = "original-server"
	baseConfig.Logging.Level = "info"

	// Save base config
	err := baseConfig.SaveToFile(baseConfigPath)
	if err != nil {
		t.Fatalf("Failed to save base config: %v", err)
	}

	// Simulate CLI overrides by creating a config with override values
	overrideConfig := &ServerConfig{
		AppName: "cli-overridden-server",
		Logging: LoggingConfig{
			Level: "debug",
		},
	}

	// Merge base config with CLI overrides
	mergedConfig := mergeConfigs(baseConfig, overrideConfig)

	// Validate merged config
	if mergedConfig.AppName != "cli-overridden-server" {
		t.Errorf("Expected app name 'cli-overridden-server', got %s", mergedConfig.AppName)
	}

	if mergedConfig.Logging.Level != "debug" {
		t.Errorf("Expected log level 'debug', got %s", mergedConfig.Logging.Level)
	}

	// Ensure other config values remain from base config
	if mergedConfig.Server.Name != baseConfig.Server.Name {
		t.Errorf("Expected server name '%s', got %s", baseConfig.Server.Name, mergedConfig.Server.Name)
	}
}

// TestConfigIntegration_PartialConfigMerge tests merging of partial configuration files
func TestConfigIntegration_PartialConfigMerge(t *testing.T) {
	// Create a partial config file with some overrides
	tmpDir := t.TempDir()
	partialConfigPath := filepath.Join(tmpDir, "partial_config.json")

	partialConfig := `{
		"app_name": "partial-override",
		"logging": {
			"level": "warning",
			"max_size_mb": 100
		}
	}`

	err := os.WriteFile(partialConfigPath, []byte(partialConfig), 0644)
	if err != nil {
		t.Fatalf("Failed to write partial config: %v", err)
	}

	// Load config with defaults
	loadedConfig, err := LoadFromFileWithDefaults(partialConfigPath)
	if err != nil {
		t.Fatalf("Failed to load config with defaults: %v", err)
	}

	// Verify partial overrides
	if loadedConfig.AppName != "partial-override" {
		t.Errorf("Expected app name 'partial-override', got %s", loadedConfig.AppName)
	}

	if loadedConfig.Logging.Level != "warning" {
		t.Errorf("Expected log level 'warning', got %s", loadedConfig.Logging.Level)
	}

	if loadedConfig.Logging.MaxSize != 100 {
		t.Errorf("Expected max log size 100, got %d", loadedConfig.Logging.MaxSize)
	}

	// Verify other values are from default config
	defaultConfig := DefaultConfig()
	if loadedConfig.Server.Name != defaultConfig.Server.Name {
		t.Errorf("Expected server name '%s', got %s", defaultConfig.Server.Name, loadedConfig.Server.Name)
	}
}

// TestConfigIntegration_ComplexConfigScenarios tests complex configuration merging
func TestConfigIntegration_ComplexConfigScenarios(t *testing.T) {
	// Setup test cases for complex configuration scenarios
	testCases := []struct {
		name           string
		baseConfig     *ServerConfig
		overrideConfig *ServerConfig
		expectedValues func(*ServerConfig) bool
	}{
		{
			name:       "Partial Override with Minimal Changes",
			baseConfig: DefaultConfig(),
			overrideConfig: &ServerConfig{
				Server: ServerSettings{
					Version: "1.1.0",
				},
				Logging: LoggingConfig{
					Level: "debug",
				},
			},
			expectedValues: func(config *ServerConfig) bool {
				return config.Server.Version == "1.1.0" &&
					config.Logging.Level == "debug"
			},
		},
		{
			name:       "Complex Configuration Override",
			baseConfig: DefaultConfig(),
			overrideConfig: &ServerConfig{
				Server: ServerSettings{
					Timeout:     Duration(10 * time.Second),
					MaxRequests: 500,
				},
				LSP: LSPConfig{
					CompletionConfig: CompletionConfig{
						MaxItems:      200,
						CaseSensitive: true,
					},
				},
			},
			expectedValues: func(config *ServerConfig) bool {
				return config.Server.Timeout.Duration() == 10*time.Second &&
					config.Server.MaxRequests == 500 &&
					config.LSP.CompletionConfig.MaxItems == 200 &&
					config.LSP.CompletionConfig.CaseSensitive
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mergedConfig := mergeConfigs(tc.baseConfig, tc.overrideConfig)
			if !tc.expectedValues(mergedConfig) {
				t.Errorf("Failed configuration merge test: %s", tc.name)
			}
		})
	}
}

// TestConfigIntegration_ConfigValidationScenarios tests various configuration validation scenarios
func TestConfigIntegration_ConfigValidationScenarios(t *testing.T) {
	testCases := []struct {
		name        string
		config      *ServerConfig
		expectError bool
		errorField  string
	}{
		{
			name:   "Valid Configuration",
			config: DefaultConfig(),
			// Default config should always be valid
			expectError: false,
		},
		{
			name: "Invalid Log Level",
			config: func() *ServerConfig {
				c := DefaultConfig()
				c.Logging.Level = "invalid-level"
				return c
			}(),
			expectError: true,
			errorField:  "logging.level",
		},
		{
			name: "Invalid Timeout",
			config: func() *ServerConfig {
				c := DefaultConfig()
				c.Server.Timeout = Duration(time.Millisecond * 500)
				return c
			}(),
			expectError: true,
			errorField:  "server.timeout",
		},
		{
			name: "Negative Max Requests",
			config: func() *ServerConfig {
				c := DefaultConfig()
				c.Server.MaxRequests = -10
				return c
			}(),
			expectError: true,
			errorField:  "server.max_requests",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.config.Validate()

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected validation error for %s, got nil", tc.name)
				} else if tc.errorField != "" {
					if !strings.Contains(err.Error(), tc.errorField) {
						t.Errorf("Expected error to contain '%s', got: %v", tc.errorField, err)
					}
				}
			} else if err != nil {
				t.Errorf("Expected no validation error for %s, got: %v", tc.name, err)
			}
		})
	}
}

// TestEnhancedSchemaValidation tests the new comprehensive schema validation features
func TestEnhancedSchemaValidation(t *testing.T) {
	testCases := []struct {
		name        string
		config      func() *ServerConfig
		expectError bool
		errorField  string
	}{
		{
			name: "Invalid App Name Pattern",
			config: func() *ServerConfig {
				c := DefaultConfig()
				c.AppName = "invalid@name!"
				return c
			},
			expectError: true,
			errorField:  "app_name",
		},
		{
			name: "Reserved App Name",
			config: func() *ServerConfig {
				c := DefaultConfig()
				c.AppName = "system"
				return c
			},
			expectError: true,
			errorField:  "app_name",
		},
		{
			name: "Invalid Semver Version",
			config: func() *ServerConfig {
				c := DefaultConfig()
				c.Server.Version = "invalid-version"
				return c
			},
			expectError: true,
			errorField:  "server.version",
		},
		{
			name: "Timeout Too High",
			config: func() *ServerConfig {
				c := DefaultConfig()
				c.Server.Timeout = Duration(10 * time.Minute)
				return c
			},
			expectError: true,
			errorField:  "server.timeout",
		},
		{
			name: "Max Requests Too High",
			config: func() *ServerConfig {
				c := DefaultConfig()
				c.Server.MaxRequests = 200000
				return c
			},
			expectError: true,
			errorField:  "server.max_requests",
		},
		{
			name: "Invalid Log File Name",
			config: func() *ServerConfig {
				c := DefaultConfig()
				c.Logging.FileName = "invalid:file*name"
				return c
			},
			expectError: true,
			errorField:  "logging.file_name",
		},
		{
			name: "Non-absolute Log Directory",
			config: func() *ServerConfig {
				c := DefaultConfig()
				c.Logging.Directory = "relative/path"
				return c
			},
			expectError: true,
			errorField:  "logging.directory",
		},
		{
			name: "Invalid Extension Format",
			config: func() *ServerConfig {
				c := DefaultConfig()
				c.LSP.Extensions = []string{"go", ".py"}
				return c
			},
			expectError: true,
			errorField:  "lsp.extensions[0]",
		},
		{
			name: "Invalid Diagnostic Severity",
			config: func() *ServerConfig {
				c := DefaultConfig()
				c.LSP.DiagnosticsConfig.Severities = []string{"error", "invalid"}
				return c
			},
			expectError: true,
			errorField:  "lsp.diagnostics.severities[1]",
		},
		{
			name: "Invalid Custom Prefix Pattern",
			config: func() *ServerConfig {
				c := DefaultConfig()
				c.LSP.MockData.CustomPrefixes = []string{"valid", "invalid@prefix"}
				return c
			},
			expectError: true,
			errorField:  "lsp.mock_data.custom_prefixes[1]",
		},
		{
			name: "Invalid Language Name Length",
			config: func() *ServerConfig {
				c := DefaultConfig()
				c.LSP.MockData.Languages = []string{"go", "a"}
				return c
			},
			expectError: true,
			errorField:  "lsp.mock_data.languages[1]",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := tc.config()
			err := config.Validate()

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected validation error for %s, got nil", tc.name)
				} else if tc.errorField != "" {
					if !strings.Contains(err.Error(), tc.errorField) {
						t.Errorf("Expected error to contain '%s', got: %v", tc.errorField, err)
					}
				}
			} else if err != nil {
				t.Errorf("Expected no validation error for %s, got: %v", tc.name, err)
			}
		})
	}
}

// TestValidationBoundaryConditions tests boundary conditions for all numeric limits
func TestValidationBoundaryConditions(t *testing.T) {
	testCases := []struct {
		name        string
		config      func() *ServerConfig
		expectError bool
		description string
	}{
		{
			name: "App Name at Limit",
			config: func() *ServerConfig {
				c := DefaultConfig()
				c.AppName = strings.Repeat("a", 100) // Exactly 100 chars
				return c
			},
			expectError: false,
			description: "App name exactly at 100 character limit should be valid",
		},
		{
			name: "Log Max Size at Upper Limit",
			config: func() *ServerConfig {
				c := DefaultConfig()
				c.Logging.MaxSize = 10000 // At upper limit
				return c
			},
			expectError: false,
			description: "Log max size at upper limit should be valid",
		},
		{
			name: "Completion Max Items at Lower Limit",
			config: func() *ServerConfig {
				c := DefaultConfig()
				c.LSP.CompletionConfig.MaxItems = 1 // At lower limit
				return c
			},
			expectError: false,
			description: "Completion max items at lower limit should be valid",
		},
		{
			name: "Hover Max Length at Upper Limit",
			config: func() *ServerConfig {
				c := DefaultConfig()
				c.LSP.HoverConfig.MaxLength = 100000 // At upper limit
				return c
			},
			expectError: false,
			description: "Hover max length at upper limit should be valid",
		},
		{
			name: "Mock Data Item Count Over Limit",
			config: func() *ServerConfig {
				c := DefaultConfig()
				c.LSP.MockData.ItemCount = 100001 // Over limit
				return c
			},
			expectError: true,
			description: "Mock data item count over limit should be invalid",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := tc.config()
			err := config.Validate()

			if tc.expectError && err == nil {
				t.Errorf("%s: Expected validation error, got nil", tc.description)
			} else if !tc.expectError && err != nil {
				t.Errorf("%s: Expected no validation error, got: %v", tc.description, err)
			}
		})
	}
}