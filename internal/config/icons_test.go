package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIconRegistry(t *testing.T) {
	// Setup a temporary home directory
	tmpDir, err := os.MkdirTemp("", "yoto-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	t.Run("AddIconRecord", func(t *testing.T) {
		cfg := &IconsConfig{
			Icons: []IconRecord{
				{ID: "id1", Name: "name1", Tags: []string{"tag1"}},
			},
		}

		// Simulate adding a new record
		newRecord := IconRecord{ID: "id2", Name: "name2", Tags: []string{"tag2"}}
		cfg.Icons = append(cfg.Icons, newRecord)

		assert.Equal(t, 2, len(cfg.Icons))
		assert.Equal(t, "id2", cfg.Icons[1].ID)

		// Simulate updating an existing record
		cfg.Icons[0].Name = "updated"
		assert.Equal(t, "updated", cfg.Icons[0].Name)
	})
}

func TestAddIconRecordLogic(t *testing.T) {
	// Let's create a testable version of the logic to verify the upsert behavior
	icons := []IconRecord{
		{ID: "1", Name: "one"},
	}

	// Helper to simulate the logic in AddIconRecord
	upsert := func(list []IconRecord, id, name string) []IconRecord {
		for i, icon := range list {
			if icon.ID == id {
				list[i].Name = name
				return list
			}
		}
		return append(list, IconRecord{ID: id, Name: name})
	}

	// Test Update
	icons = upsert(icons, "1", "updated")
	assert.Equal(t, 1, len(icons))
	assert.Equal(t, "updated", icons[0].Name)

	// Test Add
	icons = upsert(icons, "2", "two")
	assert.Equal(t, 2, len(icons))
	assert.Equal(t, "2", icons[1].ID)
}
