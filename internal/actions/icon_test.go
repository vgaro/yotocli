package actions

import (
	"testing"

	"github.com/vgaro/yotocli/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestSearchIconsLogic(t *testing.T) {
	// Mock icons
	icons := []config.IconRecord{
		{ID: "id1", Name: "bee", Tags: []string{"insect", "yellow"}},
		{ID: "id2", Name: "fire", Tags: []string{"emergency", "red"}},
		{ID: "id3", Name: "water", Tags: []string{"blue"}},
	}

	// Helper to simulate SearchIcons without hitting disk
	search := func(query string) []config.IconRecord {
		var results []config.IconRecord
		for _, icon := range icons {
			if icon.Name == query {
				results = append(results, icon)
				continue
			}
			for _, tag := range icon.Tags {
				if tag == query {
					results = append(results, icon)
					break
				}
			}
		}
		return results
	}

	t.Run("SearchByName", func(t *testing.T) {
		res := search("bee")
		assert.Equal(t, 1, len(res))
		assert.Equal(t, "id1", res[0].ID)
	})

	t.Run("SearchByTag", func(t *testing.T) {
		res := search("emergency")
		assert.Equal(t, 1, len(res))
		assert.Equal(t, "fire", res[0].Name)
	})

	t.Run("NoResults", func(t *testing.T) {
		res := search("missing")
		assert.Equal(t, 0, len(res))
	})
}
