// Package logging provides logging functionality with configurable directory resolution
package logging

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"

	"mock-lsp-server/directories" // Replace with your actual module path
)

// Config represents the logging configuration
type Config struct {
	LogDir   string `json:"log_dir"`
	LogLevel string `json:"log_level"`
	LogFile  string `json:"log_file"`
}

// Manager handles logging operations with directory resolution and configuration
type Manager struct {
	appName  string
	resolver *directories.DirectoryResolver
	config   *Config
	logger   *log.Logger
	logFile  *os.File
}

// NewManager creates a new logging manager
func NewManager(appName string, user *user.User, shouldEnsureDir bool) *Manager {
	return &Manager{
		appName:  appName,
		resolver: directories.NewDirectoryResolver(appName, user, shouldEnsureDir),
		config:   &Config{},
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

	// Create logger
	lm.logger = log.New(logFile, fmt.Sprintf("[%s] ", lm.appName), log.LstdFlags|log.Lshortfile)

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

// Log writes a message to the log
func (lm *Manager) Log(message string) {
	if lm.logger != nil {
		lm.logger.Println(message)
	}
}

// Info writes an info-level message
func (lm *Manager) Info(message string) {
	if lm.logger != nil {
		lm.logger.Printf("INFO: %s", message)
	}
}

// Error writes an error-level message
func (lm *Manager) Error(message string) {
	if lm.logger != nil {
		lm.logger.Printf("ERROR: %s", message)
	}
}

// LogWarning writes a warning-level message
func (lm *Manager) Warning(message string) {
	if lm.logger != nil {
		lm.logger.Printf("WARNING: %s", message)
	}
}

// LogDebug writes a debug-level message
func (lm *Manager) Debug(message string) {
	if lm.logger != nil {
		lm.logger.Printf("DEBUG: %s", message)
	}
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
