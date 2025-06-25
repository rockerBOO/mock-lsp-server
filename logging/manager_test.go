package logging_test

import (
	"os/user"
	"path/filepath"
	"testing"

	"mock-lsp-server/logging"
)

func TestManager_GetDefaultConfigPath(t *testing.T) {
	currentUser, err := user.Current()
	if err != nil {
		t.Skipf("Skipping test: Failed to get current user: %v", err)
	}
	expectedRegularUserConfigPath := filepath.Join(currentUser.HomeDir, ".config", "test", "config.json")

	tests := []struct {
		name string // description of this test case
		// Named input parameters for receiver constructor.
		appName         string
		user            *user.User
		shouldEnsureDir bool
		want            string
		wantErr         bool
	}{
		{
			name:    "root",
			appName: "test",
			user: &user.User{
				Uid: "0",
			},
			shouldEnsureDir: false,
			want:            filepath.Join("/", "etc", "test", "config.json"),
			wantErr:         false,
		},
		{
			name:    "regular user",
			appName: "test",
			user: currentUser, // Use the actual current user
			shouldEnsureDir: false,
			want:            expectedRegularUserConfigPath, // Use the calculated path
			wantErr:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lm := logging.NewManager(tt.appName, tt.user, tt.shouldEnsureDir)
			got, gotErr := lm.GetDefaultConfigPath()
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("GetDefaultConfigPath() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("GetDefaultConfigPath() succeeded unexpectedly")
			}
			// TODO: update the condition below to compare got with tt.want.
			if got != tt.want {
				t.Errorf("GetDefaultConfigPath() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Test log levels
func TestLogLevels(t *testing.T) {
	testCases := []struct {
		level    logging.LogLevel
		expected string
	}{
		{logging.LogLevelDebug, "DEBUG"},
		{logging.LogLevelInfo, "INFO"},
		{logging.LogLevelWarning, "WARNING"},
		{logging.LogLevelError, "ERROR"},
	}

	for _, tc := range testCases {
		t.Run(tc.expected, func(t *testing.T) {
			if tc.level.String() != tc.expected {
				t.Errorf("Expected %s, got %s", tc.expected, tc.level.String())
			}
		})
	}
}

// Test log level parsing
func TestParseLogLevel(t *testing.T) {
	testCases := []struct {
		input    string
		expected logging.LogLevel
	}{
		{"debug", logging.LogLevelDebug},
		{"DEBUG", logging.LogLevelDebug},
		{"info", logging.LogLevelInfo},
		{"INFO", logging.LogLevelInfo},
		{"warning", logging.LogLevelWarning},
		{"WARN", logging.LogLevelWarning},
		{"error", logging.LogLevelError},
		{"ERROR", logging.LogLevelError},
		{"invalid", logging.LogLevelInfo}, // Default to info
		{"", logging.LogLevelInfo},        // Default to info
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result := logging.ParseLogLevel(tc.input)
			if result != tc.expected {
				t.Errorf("Expected %v, got %v", tc.expected, result)
			}
		})
	}
}

// Test structured logging
func TestStructuredLogger(t *testing.T) {
	u, err := user.Current()
	if err != nil {
		t.Fatalf("Failed to get current user: %v", err)
	}

	// Create manager
	manager := logging.NewManager("test-app", u, false)

	// Test that structured logger was created
	structuredLogger := manager.NewStructuredLogger()
	if structuredLogger == nil {
		t.Fatal("Failed to create structured logger")
	}

	// Test adding context
	contextLogger := structuredLogger.WithContext("component", "test")
	if contextLogger == nil {
		t.Fatal("Failed to create context logger")	
	}

	// Test method chaining
	multiContextLogger := contextLogger.WithContext("operation", "unit-test")
	if multiContextLogger == nil {
		t.Fatal("Failed to chain context")
	}
}

// Test log level management
func TestLogLevelManagement(t *testing.T) {
	u, err := user.Current()
	if err != nil {
		t.Fatalf("Failed to get current user: %v", err)
	}

	manager := logging.NewManager("test-app", u, false)

	// Test default log level
	if manager.GetLogLevel() != logging.LogLevelInfo {
		t.Errorf("Expected default log level INFO, got %v", manager.GetLogLevel())
	}

	// Test setting log levels
	testLevels := []logging.LogLevel{
		logging.LogLevelDebug,
		logging.LogLevelInfo,
		logging.LogLevelWarning,
		logging.LogLevelError,
	}

	for _, level := range testLevels {
		t.Run(level.String(), func(t *testing.T) {
			manager.SetLogLevel(level)
			if manager.GetLogLevel() != level {
				t.Errorf("Expected log level %v, got %v", level, manager.GetLogLevel())
			}
		})
	}
}
