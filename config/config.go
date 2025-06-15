package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ServerConfig represents the complete server configuration
type ServerConfig struct {
	AppName string         `json:"app_name" validate:"required,min=1,max=100"`
	Server  ServerSettings `json:"server" validate:"required"`
	Logging LoggingConfig  `json:"logging" validate:"required"`
	LSP     LSPConfig      `json:"lsp" validate:"required"`
}

// ServerSettings contains core server configuration
type ServerSettings struct {
	Name        string   `json:"name" validate:"required,min=1,max=100"`
	Version     string   `json:"version" validate:"required,semver"`
	Description string   `json:"description" validate:"max=500"`
	Timeout     Duration `json:"timeout" validate:"min=1s,max=300s"`
	MaxRequests int      `json:"max_requests" validate:"min=1,max=10000"`
}

// LoggingConfig represents logging configuration with validation
type LoggingConfig struct {
	Level      string `json:"level" validate:"required,oneof=debug info warning error"`
	Directory  string `json:"directory" validate:"omitempty,dir"`
	FileName   string `json:"file_name" validate:"omitempty,min=1,max=255"`
	MaxSize    int    `json:"max_size_mb" validate:"min=1,max=1000"`
	MaxBackups int    `json:"max_backups" validate:"min=0,max=100"`
	MaxAge     int    `json:"max_age_days" validate:"min=0,max=365"`
	Compress   bool   `json:"compress"`
	Format     string `json:"format" validate:"oneof=text json"`
}

// LSPConfig represents LSP-specific configuration
type LSPConfig struct {
	InitializeTimeout Duration          `json:"initialize_timeout" validate:"min=1s,max=60s"`
	CompletionConfig  CompletionConfig  `json:"completion" validate:"required"`
	HoverConfig       HoverConfig       `json:"hover" validate:"required"`
	DiagnosticsConfig DiagnosticsConfig `json:"diagnostics" validate:"required"`
	MockData          MockDataConfig    `json:"mock_data" validate:"required"`
	Features          map[string]bool   `json:"features"`
	TriggerCharacters []string          `json:"trigger_characters" validate:"max=20"`
	Extensions        []string          `json:"extensions" validate:"dive,min=1,max=10"`
}

// CompletionConfig configures completion behavior
type CompletionConfig struct {
	Enabled           bool     `json:"enabled"`
	MaxItems          int      `json:"max_items" validate:"min=1,max=1000"`
	TriggerCharacters []string `json:"trigger_characters" validate:"max=10"`
	CaseSensitive     bool     `json:"case_sensitive"`
	IncludeSnippets   bool     `json:"include_snippets"`
}

// HoverConfig configures hover behavior
type HoverConfig struct {
	Enabled     bool `json:"enabled"`
	ShowTypes   bool `json:"show_types"`
	ShowDocs    bool `json:"show_docs"`
	ShowExample bool `json:"show_example"`
	MaxLength   int  `json:"max_length" validate:"min=100,max=10000"`
}

// DiagnosticsConfig configures diagnostic reporting
type DiagnosticsConfig struct {
	Enabled      bool     `json:"enabled"`
	MaxIssues    int      `json:"max_issues" validate:"min=1,max=1000"`
	UpdateDelay  Duration `json:"update_delay" validate:"min=100ms,max=5s"`
	Severities   []string `json:"severities" validate:"dive,oneof=error warning info hint"`
	MockWarnings bool     `json:"mock_warnings"`
	MockErrors   bool     `json:"mock_errors"`
}

// MockDataConfig configures mock data generation
type MockDataConfig struct {
	Enabled        bool     `json:"enabled"`
	Seed           int64    `json:"seed"`
	ItemCount      int      `json:"item_count" validate:"min=1,max=10000"`
	UseRealistic   bool     `json:"use_realistic"`
	CustomPrefixes []string `json:"custom_prefixes" validate:"max=50"`
	Languages      []string `json:"languages" validate:"dive,min=2,max=10"`
}

// ValidationError represents a configuration validation error
type ValidationError struct {
	Field   string `json:"field"`
	Value   string `json:"value"`
	Message string `json:"message"`
}

// Error implements error interface
func (ve ValidationError) Error() string {
	return fmt.Sprintf("validation failed for field '%s' with value '%s': %s", ve.Field, ve.Value, ve.Message)
}

// Duration is a custom type that wraps time.Duration for JSON marshaling
type Duration time.Duration

// MarshalJSON implements json.Marshaler
func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Duration(d).String())
}

// UnmarshalJSON implements json.Unmarshaler
func (d *Duration) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	duration, err := time.ParseDuration(s)
	if err != nil {
		return err
	}

	*d = Duration(duration)
	return nil
}

// Duration returns the underlying time.Duration
func (d Duration) Duration() time.Duration {
	return time.Duration(d)
}

// String returns the string representation
func (d Duration) String() string {
	return time.Duration(d).String()
}

// ValidationErrors represents multiple validation errors
type ValidationErrors []ValidationError

// Error implements error interface
func (ves ValidationErrors) Error() string {
	if len(ves) == 0 {
		return "no validation errors"
	}
	if len(ves) == 1 {
		return ves[0].Error()
	}
	return fmt.Sprintf("%d validation errors: %s (and %d more)", len(ves), ves[0].Error(), len(ves)-1)
}

// DefaultConfig returns a default server configuration
func DefaultConfig() *ServerConfig {
	return &ServerConfig{
		AppName: "mock-lsp-server",
		Server: ServerSettings{
			Name:        "Mock LSP Server",
			Version:     "1.0.0",
			Description: "A mock LSP server for testing and development",
			Timeout:     Duration(30 * time.Second),
			MaxRequests: 1000,
		},
		Logging: LoggingConfig{
			Level:      "info",
			Directory:  "",
			FileName:   "mock-lsp-server.log",
			MaxSize:    50,
			MaxBackups: 3,
			MaxAge:     30,
			Compress:   true,
			Format:     "text",
		},
		LSP: LSPConfig{
			InitializeTimeout: Duration(10 * time.Second),
			CompletionConfig: CompletionConfig{
				Enabled:           true,
				MaxItems:          100,
				TriggerCharacters: []string{".", ":", "("},
				CaseSensitive:     false,
				IncludeSnippets:   true,
			},
			HoverConfig: HoverConfig{
				Enabled:     true,
				ShowTypes:   true,
				ShowDocs:    true,
				ShowExample: false,
				MaxLength:   1000,
			},
			DiagnosticsConfig: DiagnosticsConfig{
				Enabled:      true,
				MaxIssues:    50,
				UpdateDelay:  Duration(500 * time.Millisecond),
				Severities:   []string{"error", "warning", "info"},
				MockWarnings: true,
				MockErrors:   false,
			},
			MockData: MockDataConfig{
				Enabled:        true,
				Seed:           0, // Use random seed if 0
				ItemCount:      50,
				UseRealistic:   true,
				CustomPrefixes: []string{"mock", "test", "example"},
				Languages:      []string{"go", "typescript", "python"},
			},
			Features: map[string]bool{
				"completion":      true,
				"hover":           true,
				"definition":      true,
				"references":      true,
				"document_symbol": true,
				"diagnostics":     true,
			},
			TriggerCharacters: []string{".", ":", "(", "[", "{"},
			Extensions:        []string{".go", ".ts", ".js", ".py"},
		},
	}
}

// LoadFromFile loads configuration from a JSON file
func LoadFromFile(path string) (*ServerConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("config file not found: %s", path)
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config ServerConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

// LoadFromFileWithDefaults loads config from file, falling back to defaults for missing fields
func LoadFromFileWithDefaults(path string) (*ServerConfig, error) {
	defaultConfig := DefaultConfig()

	if path == "" {
		return defaultConfig, nil
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return defaultConfig, nil
	}

	fileConfig, err := LoadFromFile(path)
	if err != nil {
		return nil, err
	}

	// Merge with defaults (file config takes precedence)
	mergedConfig := mergeConfigs(defaultConfig, fileConfig)
	return mergedConfig, nil
}

// SaveToFile saves configuration to a JSON file
func (c *ServerConfig) SaveToFile(path string) error {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// Validate validates the configuration
func (c *ServerConfig) Validate() error {
	var errors ValidationErrors

	// Validate AppName
	if c.AppName == "" {
		errors = append(errors, ValidationError{
			Field:   "app_name",
			Value:   c.AppName,
			Message: "app name is required",
		})
	} else if len(c.AppName) > 100 {
		errors = append(errors, ValidationError{
			Field:   "app_name",
			Value:   c.AppName,
			Message: "app name must be less than 100 characters",
		})
	}

	// Validate Server settings
	if err := c.validateServer(); err != nil {
		if ve, ok := err.(ValidationErrors); ok {
			errors = append(errors, ve...)
		} else {
			errors = append(errors, ValidationError{
				Field:   "server",
				Value:   "",
				Message: err.Error(),
			})
		}
	}

	// Validate Logging config
	if err := c.validateLogging(); err != nil {
		if ve, ok := err.(ValidationErrors); ok {
			errors = append(errors, ve...)
		} else {
			errors = append(errors, ValidationError{
				Field:   "logging",
				Value:   "",
				Message: err.Error(),
			})
		}
	}

	// Validate LSP config
	if err := c.validateLSP(); err != nil {
		if ve, ok := err.(ValidationErrors); ok {
			errors = append(errors, ve...)
		} else {
			errors = append(errors, ValidationError{
				Field:   "lsp",
				Value:   "",
				Message: err.Error(),
			})
		}
	}

	if len(errors) > 0 {
		return errors
	}

	return nil
}

// validateServer validates server configuration
func (c *ServerConfig) validateServer() error {
	var errors ValidationErrors

	if c.Server.Name == "" {
		errors = append(errors, ValidationError{
			Field:   "server.name",
			Value:   c.Server.Name,
			Message: "server name is required",
		})
	}

	if c.Server.Version == "" {
		errors = append(errors, ValidationError{
			Field:   "server.version",
			Value:   c.Server.Version,
			Message: "server version is required",
		})
	}

	if c.Server.Timeout.Duration() < time.Second {
		errors = append(errors, ValidationError{
			Field:   "server.timeout",
			Value:   c.Server.Timeout.String(),
			Message: "timeout must be at least 1 second",
		})
	}

	if c.Server.MaxRequests < 1 {
		errors = append(errors, ValidationError{
			Field:   "server.max_requests",
			Value:   fmt.Sprintf("%d", c.Server.MaxRequests),
			Message: "max_requests must be at least 1",
		})
	}

	if len(errors) > 0 {
		return errors
	}
	return nil
}

// validateLogging validates logging configuration
func (c *ServerConfig) validateLogging() error {
	var errors ValidationErrors

	validLevels := []string{"debug", "info", "warning", "error"}
	levelValid := false
	for _, level := range validLevels {
		if strings.ToLower(c.Logging.Level) == level {
			levelValid = true
			break
		}
	}

	if !levelValid {
		errors = append(errors, ValidationError{
			Field:   "logging.level",
			Value:   c.Logging.Level,
			Message: "level must be one of: debug, info, warning, error",
		})
	}

	if c.Logging.MaxSize < 1 {
		errors = append(errors, ValidationError{
			Field:   "logging.max_size_mb",
			Value:   fmt.Sprintf("%d", c.Logging.MaxSize),
			Message: "max_size_mb must be at least 1",
		})
	}

	if c.Logging.Format != "" && c.Logging.Format != "text" && c.Logging.Format != "json" {
		errors = append(errors, ValidationError{
			Field:   "logging.format",
			Value:   c.Logging.Format,
			Message: "format must be 'text' or 'json'",
		})
	}

	if len(errors) > 0 {
		return errors
	}
	return nil
}

// validateLSP validates LSP configuration
func (c *ServerConfig) validateLSP() error {
	var errors ValidationErrors

	if c.LSP.InitializeTimeout.Duration() < time.Second {
		errors = append(errors, ValidationError{
			Field:   "lsp.initialize_timeout",
			Value:   c.LSP.InitializeTimeout.String(),
			Message: "initialize_timeout must be at least 1 second",
		})
	}

	if c.LSP.CompletionConfig.MaxItems < 1 {
		errors = append(errors, ValidationError{
			Field:   "lsp.completion.max_items",
			Value:   fmt.Sprintf("%d", c.LSP.CompletionConfig.MaxItems),
			Message: "completion max_items must be at least 1",
		})
	}

	if c.LSP.HoverConfig.MaxLength < 100 {
		errors = append(errors, ValidationError{
			Field:   "lsp.hover.max_length",
			Value:   fmt.Sprintf("%d", c.LSP.HoverConfig.MaxLength),
			Message: "hover max_length must be at least 100",
		})
	}

	if len(errors) > 0 {
		return errors
	}
	return nil
}

// mergeConfigs merges two configurations, with override taking precedence
func mergeConfigs(base, override *ServerConfig) *ServerConfig {
	result := *base // Copy base config

	// Override non-empty values
	if override.AppName != "" {
		result.AppName = override.AppName
	}

	// Merge server settings
	if override.Server.Name != "" {
		result.Server.Name = override.Server.Name
	}
	if override.Server.Version != "" {
		result.Server.Version = override.Server.Version
	}
	if override.Server.Description != "" {
		result.Server.Description = override.Server.Description
	}
	if override.Server.Timeout.Duration() != 0 {
		result.Server.Timeout = override.Server.Timeout
	}
	if override.Server.MaxRequests != 0 {
		result.Server.MaxRequests = override.Server.MaxRequests
	}

	// Merge logging settings
	if override.Logging.Level != "" {
		result.Logging.Level = override.Logging.Level
	}
	if override.Logging.Directory != "" {
		result.Logging.Directory = override.Logging.Directory
	}
	if override.Logging.FileName != "" {
		result.Logging.FileName = override.Logging.FileName
	}
	if override.Logging.MaxSize != 0 {
		result.Logging.MaxSize = override.Logging.MaxSize
	}
	if override.Logging.MaxBackups != 0 {
		result.Logging.MaxBackups = override.Logging.MaxBackups
	}
	if override.Logging.Format != "" {
		result.Logging.Format = override.Logging.Format
	}

	// Merge LSP settings (simplified - in real implementation would be more thorough)
	if override.LSP.InitializeTimeout.Duration() != 0 {
		result.LSP.InitializeTimeout = override.LSP.InitializeTimeout
	}

	return &result
}
