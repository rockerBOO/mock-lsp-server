// Package logging provides logging functionality with configurable directory resolution
package logging

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"mock-lsp-server/directories" // Replace with your actual module path
)

// LogLevel represents different log levels
type LogLevel int

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarning
	LogLevelError
)

// String returns the string representation of the log level
func (l LogLevel) String() string {
	switch l {
	case LogLevelDebug:
		return "DEBUG"
	case LogLevelInfo:
		return "INFO"
	case LogLevelWarning:
		return "WARNING"
	case LogLevelError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// ParseLogLevel parses a string into a LogLevel
func ParseLogLevel(level string) LogLevel {
	switch strings.ToUpper(level) {
	case "DEBUG":
		return LogLevelDebug
	case "INFO":
		return LogLevelInfo
	case "WARNING", "WARN":
		return LogLevelWarning
	case "ERROR":
		return LogLevelError
	default:
		return LogLevelInfo // Default to info
	}
}

// Config represents the logging configuration
type Config struct {
	LogDir     string `json:"log_dir"`
	LogLevel   string `json:"log_level"`
	LogFile    string `json:"log_file"`
	MaxSize    int    `json:"max_size_mb"` // Maximum size in MB before rotation
	MaxBackups int    `json:"max_backups"` // Maximum number of backup files
}

// Manager handles logging operations with directory resolution and configuration
type Manager struct {
	appName      string
	resolver     *directories.DirectoryResolver
	config       *Config
	logger       *log.Logger
	logFile      *os.File
	currentLevel LogLevel
}

// NewManager creates a new logging manager
func NewManager(appName string, user *user.User, shouldEnsureDir bool) *Manager {
	return &Manager{
		appName:      appName,
		resolver:     directories.NewDirectoryResolver(appName, user, shouldEnsureDir),
		config:       &Config{LogLevel: "info"}, // Default to info level
		currentLevel: LogLevelInfo,
	}
}

// LoadConfig loads configuration from a JSON file
func (lm *Manager) LoadConfig(configPath string) error {
	if configPath == "" {
		return nil // Use defaults if no config path provided
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // Return success if file doesn't exist, use defaults
		}
		return fmt.Errorf("failed to read config file: %w", err)
	}

	if err := json.Unmarshal(data, lm.config); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	return nil
}

// GetLogDirectory returns the appropriate log directory based on CLI override, config, or defaults
func (lm *Manager) GetLogDirectory(cliLogDir string) (string, error) {
	// Priority: CLI flag > config file > user-specific default

	// 1. Check CLI flag first
	if cliLogDir != "" {
		if err := os.MkdirAll(cliLogDir, 0755); err != nil {
			return "", fmt.Errorf("failed to create CLI log directory %s: %w", cliLogDir, err)
		}
		return cliLogDir, nil
	}

	// 2. Check config file
	if lm.config.LogDir != "" {
		if err := os.MkdirAll(lm.config.LogDir, 0755); err != nil {
			return "", fmt.Errorf("failed to create config log directory %s: %w", lm.config.LogDir, err)
		}
		return lm.config.LogDir, nil
	}

	// 3. Use user-specific default
	return lm.resolver.GetLogDirectory()
}

// GetDefaultConfigPath returns the default config path for the application
func (lm *Manager) GetDefaultConfigPath() (string, error) {
	// Use the application name as the default log directory
	dir, err := lm.resolver.GetConfigDirectory()

	if err != nil {
		return "", fmt.Errorf("failed to get config directory: %w", err)
	}

	return filepath.Join(dir, "config.json"), nil
}

// GetLogFileName returns the log file name from config or default
func (lm *Manager) GetLogFileName() string {
	if lm.config.LogFile != "" {
		return lm.config.LogFile
	}
	return fmt.Sprintf("%s.log", lm.appName)
}

// Initialize sets up the logging system with the given parameters
func (lm *Manager) Initialize(cliLogDir, configPath string) error {
	// Load configuration first
	if err := lm.LoadConfig(configPath); err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Resolve log directory
	logDirectory, err := lm.GetLogDirectory(cliLogDir)
	if err != nil {
		return fmt.Errorf("failed to resolve log directory: %w", err)
	}

	// Create log file path
	logFileName := lm.GetLogFileName()
	logFilePath := filepath.Join(logDirectory, logFileName)

	// Open log file
	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file %s: %w", logFilePath, err)
	}

	// Store file handle for cleanup
	lm.logFile = logFile

	// Set log level from config
	lm.currentLevel = ParseLogLevel(lm.config.LogLevel)

	// Create logger with timestamp and source info
	lm.logger = log.New(logFile, "", 0) // No prefix, we'll handle it ourselves

	return nil
}

// GetLogger returns the configured logger instance
func (lm *Manager) GetLogger() *log.Logger {
	return lm.logger
}

// GetLogFilePath returns the current log file path
func (lm *Manager) GetLogFilePath(cliLogDir string) (string, error) {
	logDirectory, err := lm.GetLogDirectory(cliLogDir)
	if err != nil {
		return "", err
	}

	logFileName := lm.GetLogFileName()
	return filepath.Join(logDirectory, logFileName), nil
}

// shouldLog checks if a message at the given level should be logged
func (lm *Manager) shouldLog(level LogLevel) bool {
	return level >= lm.currentLevel
}

// logWithLevel writes a structured log message with the given level
func (lm *Manager) logWithLevel(level LogLevel, format string, args ...interface{}) {
	if lm.logger == nil || !lm.shouldLog(level) {
		return
	}

	timestamp := time.Now().Format("2006/01/02 15:04:05")
	message := fmt.Sprintf(format, args...)
	logEntry := fmt.Sprintf("%s [%s] [%s] %s", timestamp, lm.appName, level.String(), message)
	lm.logger.Println(logEntry)
}

// Log writes a general message to the log (INFO level)
func (lm *Manager) Log(message string) {
	lm.logWithLevel(LogLevelInfo, "%s", message)
}

// Debug writes a debug-level message
func (lm *Manager) Debug(format string, args ...interface{}) {
	lm.logWithLevel(LogLevelDebug, format, args...)
}

// Info writes an info-level message
func (lm *Manager) Info(format string, args ...interface{}) {
	lm.logWithLevel(LogLevelInfo, format, args...)
}

// Warning writes a warning-level message
func (lm *Manager) Warning(format string, args ...interface{}) {
	lm.logWithLevel(LogLevelWarning, format, args...)
}

// Error writes an error-level message
func (lm *Manager) Error(format string, args ...interface{}) {
	lm.logWithLevel(LogLevelError, format, args...)
}

// SetLogLevel changes the current log level
func (lm *Manager) SetLogLevel(level LogLevel) {
	lm.currentLevel = level
}

// GetLogLevel returns the current log level
func (lm *Manager) GetLogLevel() LogLevel {
	return lm.currentLevel
}

// StructuredLogger provides a structured logging interface
type StructuredLogger struct {
	manager *Manager
	context map[string]interface{}
}

// NewStructuredLogger creates a new structured logger
func (lm *Manager) NewStructuredLogger() *StructuredLogger {
	return &StructuredLogger{
		manager: lm,
		context: make(map[string]interface{}),
	}
}

// WithContext adds context to the logger
func (sl *StructuredLogger) WithContext(key string, value interface{}) *StructuredLogger {
	newLogger := &StructuredLogger{
		manager: sl.manager,
		context: make(map[string]interface{}),
	}
	// Copy existing context
	for k, v := range sl.context {
		newLogger.context[k] = v
	}
	// Add new context
	newLogger.context[key] = value
	return newLogger
}

// formatMessage formats a message with context
func (sl *StructuredLogger) formatMessage(format string, args ...interface{}) string {
	message := fmt.Sprintf(format, args...)
	if len(sl.context) > 0 {
		contextStr := ""
		for k, v := range sl.context {
			if contextStr != "" {
				contextStr += " "
			}
			contextStr += fmt.Sprintf("%s=%v", k, v)
		}
		return fmt.Sprintf("%s [%s]", message, contextStr)
	}
	return message
}

// Debug logs a debug message with context
func (sl *StructuredLogger) Debug(format string, args ...interface{}) {
	sl.manager.Debug("%s", sl.formatMessage(format, args...))
}

// Info logs an info message with context
func (sl *StructuredLogger) Info(format string, args ...interface{}) {
	sl.manager.Info("%s", sl.formatMessage(format, args...))
}

// Warning logs a warning message with context
func (sl *StructuredLogger) Warning(format string, args ...interface{}) {
	sl.manager.Warning("%s", sl.formatMessage(format, args...))
}

// Error logs an error message with context
func (sl *StructuredLogger) Error(format string, args ...interface{}) {
	sl.manager.Error("%s", sl.formatMessage(format, args...))
}

// Printf provides compatibility with standard logger interface
func (sl *StructuredLogger) Printf(format string, args ...interface{}) {
	sl.Info(format, args...)
}

// Println provides compatibility with standard logger interface
func (sl *StructuredLogger) Println(args ...interface{}) {
	sl.Info("%s", fmt.Sprint(args...))
}

// Close closes the log file and cleans up resources
func (lm *Manager) Close() error {
	if lm.logFile != nil {
		return lm.logFile.Close()
	}
	return nil
}

// GetInfo returns information about the current logging setup
func (lm *Manager) GetInfo(cliLogDir string) (*LogInfo, error) {
	logDirectory, err := lm.GetLogDirectory(cliLogDir)
	if err != nil {
		return nil, err
	}

	logFilePath, err := lm.GetLogFilePath(cliLogDir)
	if err != nil {
		return nil, err
	}

	configPath, err := lm.GetDefaultConfigPath()
	if err != nil {
		return nil, err
	}

	return &LogInfo{
		AppName:        lm.appName,
		LogDirectory:   logDirectory,
		LogFilePath:    logFilePath,
		ConfigPath:     configPath,
		LogFileName:    lm.GetLogFileName(),
		UsingCLIDir:    cliLogDir != "",
		UsingConfigDir: lm.config.LogDir != "",
	}, nil
}

// LogInfo contains information about the logging setup
type LogInfo struct {
	AppName        string
	LogDirectory   string
	LogFilePath    string
	ConfigPath     string
	LogFileName    string
	UsingCLIDir    bool
	UsingConfigDir bool
}
