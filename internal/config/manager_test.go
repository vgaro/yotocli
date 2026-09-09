package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
)

func TestSavePermissions(t *testing.T) {
	tests := []struct {
		name string
		// existingMode is the mode of a config file already on disk, or 0 if
		// the file should not exist yet.
		existingMode os.FileMode
	}{
		{"new file", 0},
		{"pre-existing world-readable file", 0644},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "config.yaml")

			if tt.existingMode != 0 {
				if err := os.WriteFile(path, []byte("auth:\n  access_token: stale\n"), tt.existingMode); err != nil {
					t.Fatalf("failed to seed config file: %v", err)
				}
			}

			viper.Reset()
			t.Cleanup(viper.Reset)
			viper.SetConfigFile(path)
			SetToken("access-token", "refresh-token")

			if err := Save(); err != nil {
				t.Fatalf("Save() = %v, want nil", err)
			}

			info, err := os.Stat(path)
			if err != nil {
				t.Fatalf("failed to stat config file: %v", err)
			}
			// Hardcoded rather than compared against configFileMode, so that
			// loosening the constant fails the test.
			if got := info.Mode().Perm(); got != 0600 {
				t.Errorf("config file mode = %#o, want %#o", got, 0600)
			}
		})
	}
}
