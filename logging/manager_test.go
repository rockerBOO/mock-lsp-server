package logging_test

import (
	"mock-lsp-server/logging"
	"os/user"
	"path/filepath"
	"testing"
)

func TestManager_GetDefaultConfigPath(t *testing.T) {
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
			user: &user.User{
				Uid: "1000",
			},
			shouldEnsureDir: false,
			want:            filepath.Join(".config", "test", "config.json"),
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
