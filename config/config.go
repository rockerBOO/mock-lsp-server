package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
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

	// Validate AppName with enhanced rules
	if err := c.validateAppName(); err != nil {
		if ve, ok := err.(ValidationErrors); ok {
			errors = append(errors, ve...)
		} else {
			errors = append(errors, ValidationError{
				Field:   "app_name",
				Value:   c.AppName,
				Message: err.Error(),
			})
		}
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

// validateAppName validates application name with enhanced rules
func (c *ServerConfig) validateAppName() error {
	var errors ValidationErrors

	if c.AppName == "" {
		errors = append(errors, ValidationError{
			Field:   "app_name",
			Value:   c.AppName,
			Message: "app name is required",
		})
	} else {
		// Length validation
		if len(c.AppName) > 100 {
			errors = append(errors, ValidationError{
				Field:   "app_name",
				Value:   c.AppName,
				Message: "app name must be less than 100 characters",
			})
		}

		// Pattern validation - only allow alphanumeric, hyphens, underscores
		if matched, _ := regexp.MatchString(`^[a-zA-Z0-9_-]+$`, c.AppName); !matched {
			errors = append(errors, ValidationError{
				Field:   "app_name",
				Value:   c.AppName,
				Message: "app name can only contain letters, numbers, hyphens, and underscores",
			})
		}

		// Reserved names validation
		reservedNames := []string{"system", "admin", "root", "api", "config", "log", "logs"}
		for _, reserved := range reservedNames {
			if strings.ToLower(c.AppName) == reserved {
				errors = append(errors, ValidationError{
					Field:   "app_name",
					Value:   c.AppName,
					Message: fmt.Sprintf("app name '%s' is reserved and cannot be used", reserved),
				})
				break
			}
		}
	}

	if len(errors) > 0 {
		return errors
	}
	return nil
}

// validateServer validates server configuration with enhanced rules
func (c *ServerConfig) validateServer() error {
	var errors ValidationErrors

	// Name validation
	if c.Server.Name == "" {
		errors = append(errors, ValidationError{
			Field:   "server.name",
			Value:   c.Server.Name,
			Message: "server name is required",
		})
	} else if len(c.Server.Name) > 100 {
		errors = append(errors, ValidationError{
			Field:   "server.name",
			Value:   c.Server.Name,
			Message: "server name must be less than 100 characters",
		})
	}

	// Version validation (enhanced semver-like validation)
	if c.Server.Version == "" {
		errors = append(errors, ValidationError{
			Field:   "server.version",
			Value:   c.Server.Version,
			Message: "server version is required",
		})
	} else {
		// Basic semver pattern validation
		semverPattern := `^(\d+)\.(\d+)\.(\d+)(-[a-zA-Z0-9-]+)?(\+[a-zA-Z0-9-]+)?$`
		if matched, _ := regexp.MatchString(semverPattern, c.Server.Version); !matched {
			errors = append(errors, ValidationError{
				Field:   "server.version",
				Value:   c.Server.Version,
				Message: "server version must follow semantic versioning format (e.g., 1.0.0)",
			})
		}
	}

	// Description validation
	if len(c.Server.Description) > 500 {
		errors = append(errors, ValidationError{
			Field:   "server.description",
			Value:   c.Server.Description,
			Message: "server description must be less than 500 characters",
		})
	}

	// Timeout validation
	if c.Server.Timeout.Duration() < time.Second {
		errors = append(errors, ValidationError{
			Field:   "server.timeout",
			Value:   c.Server.Timeout.String(),
			Message: "timeout must be at least 1 second",
		})
	} else if c.Server.Timeout.Duration() > 5*time.Minute {
		errors = append(errors, ValidationError{
			Field:   "server.timeout",
			Value:   c.Server.Timeout.String(),
			Message: "timeout must be less than 5 minutes",
		})
	}

	// MaxRequests validation
	if c.Server.MaxRequests < 1 {
		errors = append(errors, ValidationError{
			Field:   "server.max_requests",
			Value:   fmt.Sprintf("%d", c.Server.MaxRequests),
			Message: "max_requests must be at least 1",
		})
	} else if c.Server.MaxRequests > 100000 {
		errors = append(errors, ValidationError{
			Field:   "server.max_requests",
			Value:   fmt.Sprintf("%d", c.Server.MaxRequests),
			Message: "max_requests must be less than 100,000",
		})
	}

	if len(errors) > 0 {
		return errors
	}
	return nil
}

// validateLogging validates logging configuration with enhanced rules
func (c *ServerConfig) validateLogging() error {
	var errors ValidationErrors

	// Level validation
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

	// Directory validation (if specified)
	if c.Logging.Directory != "" {
		if !filepath.IsAbs(c.Logging.Directory) {
			errors = append(errors, ValidationError{
				Field:   "logging.directory",
				Value:   c.Logging.Directory,
				Message: "directory must be an absolute path",
			})
		}
	}

	// FileName validation
	if c.Logging.FileName != "" {
		if len(c.Logging.FileName) > 255 {
			errors = append(errors, ValidationError{
				Field:   "logging.file_name",
				Value:   c.Logging.FileName,
				Message: "file name must be less than 255 characters",
			})
		}
		
		// Check for invalid file name characters
		invalidChars := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
		for _, char := range invalidChars {
			if strings.Contains(c.Logging.FileName, char) {
				errors = append(errors, ValidationError{
					Field:   "logging.file_name",
					Value:   c.Logging.FileName,
					Message: fmt.Sprintf("file name contains invalid character '%s'", char),
				})
				break
			}
		}
	}

	// MaxSize validation
	if c.Logging.MaxSize < 1 {
		errors = append(errors, ValidationError{
			Field:   "logging.max_size_mb",
			Value:   fmt.Sprintf("%d", c.Logging.MaxSize),
			Message: "max_size_mb must be at least 1",
		})
	} else if c.Logging.MaxSize > 10000 {
		errors = append(errors, ValidationError{
			Field:   "logging.max_size_mb",
			Value:   fmt.Sprintf("%d", c.Logging.MaxSize),
			Message: "max_size_mb must be less than 10,000 MB",
		})
	}

	// MaxBackups validation
	if c.Logging.MaxBackups < 0 {
		errors = append(errors, ValidationError{
			Field:   "logging.max_backups",
			Value:   fmt.Sprintf("%d", c.Logging.MaxBackups),
			Message: "max_backups must be non-negative",
		})
	} else if c.Logging.MaxBackups > 1000 {
		errors = append(errors, ValidationError{
			Field:   "logging.max_backups",
			Value:   fmt.Sprintf("%d", c.Logging.MaxBackups),
			Message: "max_backups must be less than 1,000",
		})
	}

	// MaxAge validation
	if c.Logging.MaxAge < 0 {
		errors = append(errors, ValidationError{
			Field:   "logging.max_age_days",
			Value:   fmt.Sprintf("%d", c.Logging.MaxAge),
			Message: "max_age_days must be non-negative",
		})
	} else if c.Logging.MaxAge > 3650 {
		errors = append(errors, ValidationError{
			Field:   "logging.max_age_days",
			Value:   fmt.Sprintf("%d", c.Logging.MaxAge),
			Message: "max_age_days must be less than 10 years (3,650 days)",
		})
	}

	// Format validation
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

// validateLSP validates LSP configuration with enhanced rules
func (c *ServerConfig) validateLSP() error {
	var errors ValidationErrors

	// InitializeTimeout validation
	if c.LSP.InitializeTimeout.Duration() < time.Second {
		errors = append(errors, ValidationError{
			Field:   "lsp.initialize_timeout",
			Value:   c.LSP.InitializeTimeout.String(),
			Message: "initialize_timeout must be at least 1 second",
		})
	} else if c.LSP.InitializeTimeout.Duration() > time.Minute {
		errors = append(errors, ValidationError{
			Field:   "lsp.initialize_timeout",
			Value:   c.LSP.InitializeTimeout.String(),
			Message: "initialize_timeout must be less than 1 minute",
		})
	}

	// Validate completion config
	if err := c.validateCompletionConfig(); err != nil {
		if ve, ok := err.(ValidationErrors); ok {
			errors = append(errors, ve...)
		}
	}

	// Validate hover config
	if err := c.validateHoverConfig(); err != nil {
		if ve, ok := err.(ValidationErrors); ok {
			errors = append(errors, ve...)
		}
	}

	// Validate diagnostics config
	if err := c.validateDiagnosticsConfig(); err != nil {
		if ve, ok := err.(ValidationErrors); ok {
			errors = append(errors, ve...)
		}
	}

	// Validate mock data config
	if err := c.validateMockDataConfig(); err != nil {
		if ve, ok := err.(ValidationErrors); ok {
			errors = append(errors, ve...)
		}
	}

	// Validate trigger characters
	if len(c.LSP.TriggerCharacters) > 20 {
		errors = append(errors, ValidationError{
			Field:   "lsp.trigger_characters",
			Value:   fmt.Sprintf("%v", c.LSP.TriggerCharacters),
			Message: "trigger_characters list cannot exceed 20 items",
		})
	}

	// Validate extensions
	if len(c.LSP.Extensions) > 50 {
		errors = append(errors, ValidationError{
			Field:   "lsp.extensions",
			Value:   fmt.Sprintf("%v", c.LSP.Extensions),
			Message: "extensions list cannot exceed 50 items",
		})
	}

	for i, ext := range c.LSP.Extensions {
		if !strings.HasPrefix(ext, ".") {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("lsp.extensions[%d]", i),
				Value:   ext,
				Message: "extension must start with a dot (e.g., '.go')",
			})
		}
		if len(ext) > 10 {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("lsp.extensions[%d]", i),
				Value:   ext,
				Message: "extension must be less than 10 characters",
			})
		}
	}

	if len(errors) > 0 {
		return errors
	}
	return nil
}

// validateCompletionConfig validates completion configuration
func (c *ServerConfig) validateCompletionConfig() error {
	var errors ValidationErrors

	if c.LSP.CompletionConfig.MaxItems < 1 {
		errors = append(errors, ValidationError{
			Field:   "lsp.completion.max_items",
			Value:   fmt.Sprintf("%d", c.LSP.CompletionConfig.MaxItems),
			Message: "completion max_items must be at least 1",
		})
	} else if c.LSP.CompletionConfig.MaxItems > 10000 {
		errors = append(errors, ValidationError{
			Field:   "lsp.completion.max_items",
			Value:   fmt.Sprintf("%d", c.LSP.CompletionConfig.MaxItems),
			Message: "completion max_items must be less than 10,000",
		})
	}

	if len(c.LSP.CompletionConfig.TriggerCharacters) > 10 {
		errors = append(errors, ValidationError{
			Field:   "lsp.completion.trigger_characters",
			Value:   fmt.Sprintf("%v", c.LSP.CompletionConfig.TriggerCharacters),
			Message: "completion trigger_characters list cannot exceed 10 items",
		})
	}

	if len(errors) > 0 {
		return errors
	}
	return nil
}

// validateHoverConfig validates hover configuration
func (c *ServerConfig) validateHoverConfig() error {
	var errors ValidationErrors

	if c.LSP.HoverConfig.MaxLength < 100 {
		errors = append(errors, ValidationError{
			Field:   "lsp.hover.max_length",
			Value:   fmt.Sprintf("%d", c.LSP.HoverConfig.MaxLength),
			Message: "hover max_length must be at least 100",
		})
	} else if c.LSP.HoverConfig.MaxLength > 100000 {
		errors = append(errors, ValidationError{
			Field:   "lsp.hover.max_length",
			Value:   fmt.Sprintf("%d", c.LSP.HoverConfig.MaxLength),
			Message: "hover max_length must be less than 100,000",
		})
	}

	if len(errors) > 0 {
		return errors
	}
	return nil
}

// validateDiagnosticsConfig validates diagnostics configuration
func (c *ServerConfig) validateDiagnosticsConfig() error {
	var errors ValidationErrors

	if c.LSP.DiagnosticsConfig.MaxIssues < 1 {
		errors = append(errors, ValidationError{
			Field:   "lsp.diagnostics.max_issues",
			Value:   fmt.Sprintf("%d", c.LSP.DiagnosticsConfig.MaxIssues),
			Message: "diagnostics max_issues must be at least 1",
		})
	} else if c.LSP.DiagnosticsConfig.MaxIssues > 10000 {
		errors = append(errors, ValidationError{
			Field:   "lsp.diagnostics.max_issues",
			Value:   fmt.Sprintf("%d", c.LSP.DiagnosticsConfig.MaxIssues),
			Message: "diagnostics max_issues must be less than 10,000",
		})
	}

	if c.LSP.DiagnosticsConfig.UpdateDelay.Duration() < 50*time.Millisecond {
		errors = append(errors, ValidationError{
			Field:   "lsp.diagnostics.update_delay",
			Value:   c.LSP.DiagnosticsConfig.UpdateDelay.String(),
			Message: "diagnostics update_delay must be at least 50ms",
		})
	} else if c.LSP.DiagnosticsConfig.UpdateDelay.Duration() > 30*time.Second {
		errors = append(errors, ValidationError{
			Field:   "lsp.diagnostics.update_delay",
			Value:   c.LSP.DiagnosticsConfig.UpdateDelay.String(),
			Message: "diagnostics update_delay must be less than 30 seconds",
		})
	}

	// Validate severities
	validSeverities := []string{"error", "warning", "info", "hint"}
	for i, severity := range c.LSP.DiagnosticsConfig.Severities {
		valid := false
		for _, validSeverity := range validSeverities {
			if severity == validSeverity {
				valid = true
				break
			}
		}
		if !valid {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("lsp.diagnostics.severities[%d]", i),
				Value:   severity,
				Message: "severity must be one of: error, warning, info, hint",
			})
		}
	}

	if len(errors) > 0 {
		return errors
	}
	return nil
}

// validateMockDataConfig validates mock data configuration
func (c *ServerConfig) validateMockDataConfig() error {
	var errors ValidationErrors

	if c.LSP.MockData.ItemCount < 1 {
		errors = append(errors, ValidationError{
			Field:   "lsp.mock_data.item_count",
			Value:   fmt.Sprintf("%d", c.LSP.MockData.ItemCount),
			Message: "mock_data item_count must be at least 1",
		})
	} else if c.LSP.MockData.ItemCount > 100000 {
		errors = append(errors, ValidationError{
			Field:   "lsp.mock_data.item_count",
			Value:   fmt.Sprintf("%d", c.LSP.MockData.ItemCount),
			Message: "mock_data item_count must be less than 100,000",
		})
	}

	if len(c.LSP.MockData.CustomPrefixes) > 50 {
		errors = append(errors, ValidationError{
			Field:   "lsp.mock_data.custom_prefixes",
			Value:   fmt.Sprintf("%v", c.LSP.MockData.CustomPrefixes),
			Message: "custom_prefixes list cannot exceed 50 items",
		})
	}

	// Validate custom prefixes
	for i, prefix := range c.LSP.MockData.CustomPrefixes {
		if len(prefix) > 50 {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("lsp.mock_data.custom_prefixes[%d]", i),
				Value:   prefix,
				Message: "custom prefix must be less than 50 characters",
			})
		}
		if matched, _ := regexp.MatchString(`^[a-zA-Z0-9_-]+$`, prefix); !matched {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("lsp.mock_data.custom_prefixes[%d]", i),
				Value:   prefix,
				Message: "custom prefix can only contain letters, numbers, hyphens, and underscores",
			})
		}
	}

	// Validate languages
	for i, lang := range c.LSP.MockData.Languages {
		if len(lang) < 2 || len(lang) > 20 {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("lsp.mock_data.languages[%d]", i),
				Value:   lang,
				Message: "language name must be between 2 and 20 characters",
			})
		}
		if matched, _ := regexp.MatchString(`^[a-zA-Z0-9_-]+$`, lang); !matched {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("lsp.mock_data.languages[%d]", i),
				Value:   lang,
				Message: "language name can only contain letters, numbers, hyphens, and underscores",
			})
		}
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

	// Merge LSP settings with nested configuration merging
	if override.LSP.InitializeTimeout.Duration() != 0 {
		result.LSP.InitializeTimeout = override.LSP.InitializeTimeout
	}

	// Merge Completion config
	if override.LSP.CompletionConfig.MaxItems != 0 {
		result.LSP.CompletionConfig.MaxItems = override.LSP.CompletionConfig.MaxItems
	}
	if override.LSP.CompletionConfig.CaseSensitive {
		result.LSP.CompletionConfig.CaseSensitive = override.LSP.CompletionConfig.CaseSensitive
	}

	return &result
}