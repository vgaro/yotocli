package utils

import (
	"fmt"
	"strings"

	"github.com/vgaro/yotocli/pkg/yoto"
)

// ReorderPlaylist renumbers keys and overlay labels for consistency
func ReorderPlaylist(card *yoto.Card) {
	if card.Content == nil {
		return
	}
	for i := range card.Content.Chapters {
		key := fmt.Sprintf("%02d", i+1)
		card.Content.Chapters[i].Key = key
		card.Content.Chapters[i].OverlayLabel = fmt.Sprintf("%d", i+1)
		for j := range card.Content.Chapters[i].Tracks {
			card.Content.Chapters[i].Tracks[j].Key = key
			card.Content.Chapters[i].Tracks[j].OverlayLabel = fmt.Sprintf("%d", i+1)
		}
	}
}

// FindChapter searches for a chapter by index or title substring
func FindChapter(card *yoto.Card, query string) (int, *yoto.Chapter) {
	if card.Content == nil {
		return -1, nil
	}

	// Try Index
	if idx, err := ParseIndex(query); err == nil {
		if idx > 0 && idx <= len(card.Content.Chapters) {
			return idx - 1, &card.Content.Chapters[idx-1]
		}
	}

	// Try Title
	queryLower := strings.ToLower(query)
	for i, chapter := range card.Content.Chapters {
		if strings.Contains(strings.ToLower(chapter.Title), queryLower) {
			return i, &chapter
		}
	}

	return -1, nil
}
