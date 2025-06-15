package main

import (
	"os"
	"reflect"
	"testing"
)

// Test for the version that returns the manager too
func Test_setupLoggingWithManager(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test_logs")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	logger, manager, err := setupLogging("test-app", tempDir, "", false)
	if err != nil {
		t.Fatalf("setupLoggingWithManager() error = %v", err)
	}

	if logger == nil {
		t.Error("setupLoggingWithManager() returned nil logger")
	}

	if manager == nil {
		t.Error("setupLoggingWithManager() returned nil manager")
	}

	// Test logger works
	if logger != nil {
		logger.Println("Test message")
	}

	// Clean up properly
	if manager != nil {
		manager.Close()
	}
}

func Test_loadConfig(t *testing.T) {
	tests := []struct {
		name     string
		progname string
		args     []string
		want     *MockLSPServerConfig
		wantErr  bool
	}{
		{
			name:     "no arguments - defaults",
			progname: "mock-lsp-server",
			args:     []string{},
			want: &MockLSPServerConfig{
				AppName:    "mock-lsp-server", // default value
				LogDir:     "",
				ConfigPath: "",
				ShowInfo:   false,
			},
			wantErr: false,
		},
		{
			name:     "log_dir flag",
			progname: "mock-lsp-server",
			args:     []string{"-log_dir", "/tmp/logs"},
			want: &MockLSPServerConfig{
				AppName:    "mock-lsp-server",
				LogDir:     "/tmp/logs",
				ConfigPath: "",
				ShowInfo:   false,
			},
			wantErr: false,
		},
		{
			name:     "config flag",
			progname: "mock-lsp-server",
			args:     []string{"-config", "/path/to/config.json"},
			want: &MockLSPServerConfig{
				AppName:    "mock-lsp-server",
				LogDir:     "",
				ConfigPath: "/path/to/config.json",
				ShowInfo:   false,
			},
			wantErr: false,
		},
		{
			name:     "info flag",
			progname: "mock-lsp-server",
			args:     []string{"-info"},
			want: &MockLSPServerConfig{
				AppName:    "mock-lsp-server",
				LogDir:     "",
				ConfigPath: "",
				ShowInfo:   true,
			},
			wantErr: false,
		},
		{
			name:     "appName flag",
			progname: "test-program",
			args:     []string{"-appName", "custom-app"},
			want: &MockLSPServerConfig{
				AppName:    "custom-app",
				LogDir:     "",
				ConfigPath: "",
				ShowInfo:   false,
			},
			wantErr: false,
		},
		{
			name:     "all flags",
			progname: "mock-lsp-server",
			args:     []string{"-appName", "test-app", "-log_dir", "/var/log", "-config", "config.yaml", "-info"},
			want: &MockLSPServerConfig{
				AppName:    "test-app",
				LogDir:     "/var/log",
				ConfigPath: "config.yaml",
				ShowInfo:   true,
			},
			wantErr: false,
		},
		{
			name:     "long flag format",
			progname: "mock-lsp-server",
			args:     []string{"--log_dir=/home/user/logs", "--config=/etc/config.toml", "--info=true"},
			want: &MockLSPServerConfig{
				AppName:    "mock-lsp-server",
				LogDir:     "/home/user/logs",
				ConfigPath: "/etc/config.toml",
				ShowInfo:   true,
			},
			wantErr: false,
		},
		{
			name:     "mixed flag formats",
			progname: "mock-lsp-server",
			args:     []string{"-log_dir", "/tmp", "--config=/path/config", "-info"},
			want: &MockLSPServerConfig{
				AppName:    "mock-lsp-server",
				LogDir:     "/tmp",
				ConfigPath: "/path/config",
				ShowInfo:   true,
			},
			wantErr: false,
		},
		{
			name:     "empty string values",
			progname: "mock-lsp-server",
			args:     []string{"-log_dir", "", "-config", ""},
			want: &MockLSPServerConfig{
				AppName:    "mock-lsp-server",
				LogDir:     "",
				ConfigPath: "",
				ShowInfo:   false,
			},
			wantErr: false,
		},
		{
			name:     "boolean flag variations",
			progname: "mock-lsp-server",
			args:     []string{"-info=false"}, // explicit false
			want: &MockLSPServerConfig{
				AppName:    "mock-lsp-server",
				LogDir:     "",
				ConfigPath: "",
				ShowInfo:   false,
			},
			wantErr: false,
		},
		{
			name:     "boolean flag explicit true",
			progname: "mock-lsp-server",
			args:     []string{"-info=true"},
			want: &MockLSPServerConfig{
				AppName:    "mock-lsp-server",
				LogDir:     "",
				ConfigPath: "",
				ShowInfo:   true,
			},
			wantErr: false,
		},
		// Error cases
		{
			name:     "unknown flag",
			progname: "mock-lsp-server",
			args:     []string{"-unknown_flag", "value"},
			want:     nil,
			wantErr:  true,
		},
		{
			name:     "invalid boolean value",
			progname: "mock-lsp-server",
			args:     []string{"-info=invalid"},
			want:     nil,
			wantErr:  true,
		},
		{
			name:     "flag without value",
			progname: "mock-lsp-server",
			args:     []string{"-log_dir"}, // missing value
			want:     nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := loadConfig(tt.progname, tt.args)

			if tt.wantErr {
				if err == nil {
					t.Errorf("loadConfig() expected error, got nil")
				}
				if got != nil {
					t.Errorf("loadConfig() expected nil result on error, got %v", got)
				}
				return
			}

			if err != nil {
				t.Errorf("loadConfig() unexpected error: %v", err)
				return
			}

			if got == nil {
				t.Error("loadConfig() returned nil without error")
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("loadConfig() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

// Test individual field parsing
func Test_loadConfig_FieldValidation(t *testing.T) {
	testCases := []struct {
		name    string
		args    []string
		checkFn func(*testing.T, *MockLSPServerConfig)
	}{
		{
			name: "appName is set correctly",
			args: []string{"-appName", "my-custom-app"},
			checkFn: func(t *testing.T, config *MockLSPServerConfig) {
				if config.AppName != "my-custom-app" {
					t.Errorf("Expected AppName 'my-custom-app', got '%s'", config.AppName)
				}
			},
		},
		{
			name: "logDir handles paths with spaces",
			args: []string{"-log_dir", "/path/with spaces/logs"},
			checkFn: func(t *testing.T, config *MockLSPServerConfig) {
				expected := "/path/with spaces/logs"
				if config.LogDir != expected {
					t.Errorf("Expected LogDir '%s', got '%s'", expected, config.LogDir)
				}
			},
		},
		{
			name: "config path with special characters",
			args: []string{"-config", "/path/config-file_v2.json"},
			checkFn: func(t *testing.T, config *MockLSPServerConfig) {
				expected := "/path/config-file_v2.json"
				if config.ConfigPath != expected {
					t.Errorf("Expected ConfigPath '%s', got '%s'", expected, config.ConfigPath)
				}
			},
		},
		{
			name: "info flag precedence",
			args: []string{"-info=false", "-info"}, // second flag should override
			checkFn: func(t *testing.T, config *MockLSPServerConfig) {
				if !config.ShowInfo {
					t.Error("Expected ShowInfo to be true (last flag should win)")
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config, err := loadConfig("test-prog", tc.args)
			if err != nil {
				t.Fatalf("loadConfig() failed: %v", err)
			}
			if config == nil {
				t.Fatal("loadConfig() returned nil config")
			}
			tc.checkFn(t, config)
		})
	}
}

// Test concurrent usage (since each call creates a new FlagSet)
func Test_loadConfig_Concurrent(t *testing.T) {
	t.Parallel() // This is safe now because we don't use global state

	config1, err1 := loadConfig("prog1", []string{"-appName", "app1"})
	config2, err2 := loadConfig("prog2", []string{"-appName", "app2"})

	if err1 != nil {
		t.Errorf("First loadConfig() failed: %v", err1)
	}
	if err2 != nil {
		t.Errorf("Second loadConfig() failed: %v", err2)
	}

	if config1.AppName != "app1" {
		t.Errorf("Expected first config AppName 'app1', got '%s'", config1.AppName)
	}
	if config2.AppName != "app2" {
		t.Errorf("Expected second config AppName 'app2', got '%s'", config2.AppName)
	}
}

// Benchmark to ensure performance is reasonable
func Benchmark_loadConfig(b *testing.B) {
	args := []string{"-appName", "benchmark-app", "-log_dir", "/tmp", "-config", "config.json", "-info"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := loadConfig("benchmark-prog", args)
		if err != nil {
			b.Fatalf("loadConfig() failed: %v", err)
		}
	}
}

// Benchmark test to ensure performance is reasonable
func Benchmark_setupLogging(b *testing.B) {
	tempDir, err := os.MkdirTemp("", "benchmark_logs")
	if err != nil {
		b.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	for b.Loop() {
		logger, logManager, err := setupLogging("benchmark-app", tempDir, "", false)
		if err != nil {
			b.Fatalf("setupLogging() error = %v", err)
		}
		if logger == nil {
			b.Fatal("setupLogging() returned nil logger")
		}
		if logManager == nil {
			b.Fatal("setupLogging() returned nil logManager")
		}
	}
}
