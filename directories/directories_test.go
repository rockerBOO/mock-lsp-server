package directories

import (
	"os/user"
	"path/filepath"
	"testing"
)

func TestDirectoryResolver_GetLogDirectory(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for receiver constructor.
		appName         string
		u               *user.User
		shouldEnsureDir bool
		want            string
		wantErr         bool
	}{
		{
			name:    "root",
			appName: "test",
			u: &user.User{
				Uid: "0",
			},
			shouldEnsureDir: false,
			want:            "/var/log/test",
			wantErr:         false,
		},
		{
			name:    "regular user",
			appName: "test",
			u: &user.User{
				Uid: "1000",
			},
			shouldEnsureDir: false,
			want:            filepath.Join(".local", "share", "test", "logs"),
			wantErr:         false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dr := NewDirectoryResolver(tt.appName, tt.u, tt.shouldEnsureDir)
			got, gotErr := dr.GetLogDirectory()
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("GetLogDirectory() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("GetLogDirectory() succeeded unexpectedly")
			}
			// TODO: update the condition below to compare got with tt.want.
			if got != tt.want {
				t.Errorf("GetLogDirectory() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDirectoryResolver_GetDataDirectory(t *testing.T) {
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
			want:            "/var/lib/test",
			wantErr:         false,
		},
		{
			name:    "regular user",
			appName: "test",
			user: &user.User{
				Uid: "1000",
			},
			shouldEnsureDir: false,
			want:            filepath.Join(".local", "share", "test"),
			wantErr:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dr := NewDirectoryResolver(tt.appName, tt.user, tt.shouldEnsureDir)
			got, gotErr := dr.GetDataDirectory()
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("GetDataDirectory() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("GetDataDirectory() succeeded unexpectedly")
			}
			// TODO: update the condition below to compare got with tt.want.
			if got != tt.want {
				t.Errorf("GetDataDirectory() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDirectoryResolver_GetCacheDirectory(t *testing.T) {
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
			want:            "/var/cache/test",
			wantErr:         false,
		},
		{
			name:    "regular user",
			appName: "test",
			user: &user.User{
				Uid: "1000",
			},
			shouldEnsureDir: false,
			want:            filepath.Join(".cache", "test"),
			wantErr:         false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dr := NewDirectoryResolver(tt.appName, tt.user, tt.shouldEnsureDir)
			got, gotErr := dr.GetCacheDirectory()
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("GetCacheDirectory() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("GetCacheDirectory() succeeded unexpectedly")
			}
			// TODO: update the condition below to compare got with tt.want.
			if got != tt.want {
				t.Errorf("GetCacheDirectory() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDirectoryResolver_GetConfigDirectory(t *testing.T) {
	currentUser, err := user.Current()
	if err != nil {
		t.Skipf("Skipping test: Failed to get current user: %v", err)
	}
	expectedRegularUserConfigPath := filepath.Join(currentUser.HomeDir, ".config", "test")

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
			want:            "/etc/test",
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
			dr := NewDirectoryResolver(tt.appName, tt.user, tt.shouldEnsureDir)
			got, gotErr := dr.GetConfigDirectory()
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("GetConfigDirectory() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("GetConfigDirectory() succeeded unexpectedly")
			}
			// TODO: update the condition below to compare got with tt.want.
			if got != tt.want {
				t.Errorf("GetConfigDirectory() = %v, want %v", got, tt.want)
			}
		})
	}
}
