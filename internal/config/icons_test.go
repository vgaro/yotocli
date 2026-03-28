package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIconRegistryFileIO(t *testing.T) {
	// Setup a temporary directory for config
	tmpDir, err := os.MkdirTemp("", "yoto-config-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Set the global IconsPath for testing
	originalPath := IconsPath
	testPath := filepath.Join(tmpDir, "icons.yaml")
	IconsPath = testPath
	defer func() { IconsPath = originalPath }()

	t.Run("LoadNonExistent", func(t *testing.T) {
		cfg, err := LoadIcons()
		require.NoError(t, err)
		assert.NotNil(t, cfg)
		assert.Equal(t, 0, len(cfg.Icons))
	})

	t.Run("SaveAndLoad", func(t *testing.T) {
		cfg := &IconsConfig{
			Icons: []IconRecord{
				{ID: "id1", Name: "bee", Tags: []string{"insect"}},
			},
		}
		err := SaveIcons(cfg)
		require.NoError(t, err)

		// Verify file exists
		assert.FileExists(t, testPath)

		// Load back
		loaded, err := LoadIcons()
		require.NoError(t, err)
		assert.Equal(t, 1, len(loaded.Icons))
		assert.Equal(t, "bee", loaded.Icons[0].Name)
	})

	t.Run("AddIconRecord", func(t *testing.T) {
		err := AddIconRecord("id2", "fire", []string{"hot"})
		require.NoError(t, err)

		loaded, err := LoadIcons()
		require.NoError(t, err)
		assert.Equal(t, 2, len(loaded.Icons))
		assert.Equal(t, "fire", loaded.Icons[1].Name)
	})

	t.Run("UpsertIconRecord", func(t *testing.T) {
		// Update id1
		err := AddIconRecord("id1", "bumblebee", []string{"insect", "yellow"})
		require.NoError(t, err)

		loaded, err := LoadIcons()
		require.NoError(t, err)
		assert.Equal(t, 2, len(loaded.Icons))
		assert.Equal(t, "bumblebee", loaded.Icons[0].Name)
		assert.Contains(t, loaded.Icons[0].Tags, "yellow")
	})
}
