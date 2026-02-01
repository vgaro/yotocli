package actions

import (
	"fmt"

	"github.com/vgaro/yotocli/pkg/yoto"
)

// performRemoveTrack removes a chapter from a card's content.
func performRemoveTrack(card *yoto.Card, index int) error {
	if card.Content == nil || index < 1 || index > len(card.Content.Chapters) {
		return fmt.Errorf("invalid track index: %d", index)
	}
	idx := index - 1
	card.Content.Chapters = append(card.Content.Chapters[:idx], card.Content.Chapters[idx+1:]...)
	return nil
}

// performInsertTrack inserts a chapter into a card at a specific position.
// position is 1-based. If position < 1, it appends.
func performInsertTrack(card *yoto.Card, chapter yoto.Chapter, position int) {
	if card.Content == nil {
		card.Content = &yoto.Content{}
	}
	count := len(card.Content.Chapters)
	
	// If position is explicitly requested beyond end, just append.
	// Logic: 1-based index. 
	// If pos 1, idx 0.
	// If count 5, pos 6 is append (idx 5).
	
	idx := position - 1
	if position < 1 || idx >= count {
		card.Content.Chapters = append(card.Content.Chapters, chapter)
		return
	}

	// Insert
	card.Content.Chapters = append(card.Content.Chapters[:idx], append([]yoto.Chapter{chapter}, card.Content.Chapters[idx:]...)...)
}
