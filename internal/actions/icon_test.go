package actions

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/vgaro/yotocli/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSearchIcons(t *testing.T) {
	// Setup a temporary directory for config
	tmpDir, err := os.MkdirTemp("", "yoto-search-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Set the global IconsPath for testing
	originalPath := config.IconsPath
	testPath := filepath.Join(tmpDir, "icons.yaml")
	config.IconsPath = testPath
	defer func() { config.IconsPath = originalPath }()

	// Seed with icons
	icons := &config.IconsConfig{
		Icons: []config.IconRecord{
			{ID: "id1", Name: "Bee", Tags: []string{"insect", "yellow"}},
			{ID: "id2", Name: "Fire", Tags: []string{"emergency", "RED"}},
			{ID: "id3", Name: "Water", Tags: []string{"blue"}},
		},
	}
	err = config.SaveIcons(icons)
	require.NoError(t, err)

	t.Run("CaseInsensitiveName", func(t *testing.T) {
		res, err := SearchIcons("bee")
		require.NoError(t, err)
		assert.Equal(t, 1, len(res))
		assert.Equal(t, "id1", res[0].ID)
	})

	t.Run("CaseInsensitiveTag", func(t *testing.T) {
		res, err := SearchIcons("red")
		require.NoError(t, err)
		assert.Equal(t, 1, len(res))
		assert.Equal(t, "id2", res[0].ID)
	})

	t.Run("PartialMatch", func(t *testing.T) {
		// "er" matches Water (name) and Fire (tag: emergency)
		res, err := SearchIcons("er")
		require.NoError(t, err)
		assert.Equal(t, 2, len(res))
	})

	t.Run("NoResults", func(t *testing.T) {
		res, err := SearchIcons("missing")
		require.NoError(t, err)
		assert.Equal(t, 0, len(res))
	})
}
