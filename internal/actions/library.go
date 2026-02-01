package actions

import (
	"fmt"

	"github.com/vgaro/yotocli/pkg/yoto"
)

// RemoveTrack removes a track by 1-based index and updates metadata.
func RemoveTrack(client *yoto.Client, cardID string, trackIndex int) error {
	card, err := client.GetCard(cardID)
	if err != nil {
		return err
	}

	if err := performRemoveTrack(card, trackIndex); err != nil {
		return err
	}

	recalculateMetadata(card)
	return client.UpdateCard(card.CardID, card)
}

// MoveTrack moves a track from srcCard (index) to destCard (index).
// If destCardID is empty, it moves within the source card.
// Indices are 1-based.
func MoveTrack(client *yoto.Client, srcCardID string, srcIndex int, destCardID string, destIndex int) error {
	srcCard, err := client.GetCard(srcCardID)
	if err != nil {
		return err
	}

	if srcCard.Content == nil || srcIndex < 1 || srcIndex > len(srcCard.Content.Chapters) {
		return fmt.Errorf("invalid source index: %d", srcIndex)
	}

	// Capture the chapter to move
	chapter := srcCard.Content.Chapters[srcIndex-1]

	// Determine destination card
	var destCard *yoto.Card
	if destCardID == "" || destCardID == srcCardID {
		destCard = srcCard
	} else {
		dCard, err := client.GetCard(destCardID)
		if err != nil {
			return err
		}
		destCard = dCard
	}

	// Remove from source
	if err := performRemoveTrack(srcCard, srcIndex); err != nil {
		return err
	}

	// Adjust destIndex if moving within same card and moving "down" (higher index)
	// Because removing shifts indices.
	// E.g. [A, B, C]. Move A(1) to 3.
	// Remove A -> [B, C].
	// Insert at 3 -> [B, C, A]. Correct.
	//
	// E.g. [A, B, C]. Move C(3) to 1.
	// Remove C -> [A, B].
	// Insert at 1 -> [C, A, B]. Correct.
	//
	// However, if the user says "Move track 1 to position 2" (swap A and B).
	// [A, B]. Remove A -> [B]. Insert at 2 -> [B, A].
	// User meant: Result should be [B, A].
	// My performInsertTrack inserts *before* the index if it exists.
	// Insert at 2 (B is idx 0, count 1). 2-1 = 1. >= count? Yes. Append.
	// [B, A]. Correct.
	
	// Wait, if I move 1 to 1.
	// Remove 1. Insert at 1. Same.
	
	// Issue: If I rely on indices from *before* removal?
	// `performRemoveTrack` modifies the slice in place.
	// If srcCard == destCard, the slice is modified.
	// If I move 1 to 2.
	// Remove 1. Slice shrinks.
	// Insert at 2.
	// This seems fine for "Move A to position X in the resulting list".
	// But CLI/MCP usually implies "Move it so it ends up at position X".
	
	// Let's keep it simple: Remove, then Insert.
	// For same-card moves, users usually expect:
	// "Move 1 to 2" -> [2, 1, 3...]
	
	performInsertTrack(destCard, chapter, destIndex)

	recalculateMetadata(srcCard)
	if srcCard != destCard {
		recalculateMetadata(destCard)
		if err := client.UpdateCard(destCard.CardID, destCard); err != nil {
			return err
		}
	}

	return client.UpdateCard(srcCard.CardID, srcCard)
}

// CopyTrack copies a track from srcCard (index) to destCard (index).
func CopyTrack(client *yoto.Client, srcCardID string, srcIndex int, destCardID string, destIndex int) error {
	srcCard, err := client.GetCard(srcCardID)
	if err != nil {
		return err
	}

	if srcCard.Content == nil || srcIndex < 1 || srcIndex > len(srcCard.Content.Chapters) {
		return fmt.Errorf("invalid source index: %d", srcIndex)
	}

	chapter := srcCard.Content.Chapters[srcIndex-1]

	var destCard *yoto.Card
	if destCardID == "" || destCardID == srcCardID {
		destCard = srcCard
	} else {
		dCard, err := client.GetCard(destCardID)
		if err != nil {
			return err
		}
		destCard = dCard
	}

	performInsertTrack(destCard, chapter, destIndex)

	recalculateMetadata(destCard)
	return client.UpdateCard(destCard.CardID, destCard)
}

func recalculateMetadata(card *yoto.Card) {
	var totalDur, totalSize int
	for i := range card.Content.Chapters {
		key := fmt.Sprintf("%02d", i+1)
		card.Content.Chapters[i].Key = key
		card.Content.Chapters[i].OverlayLabel = fmt.Sprintf("%d", i+1)
		for j := range card.Content.Chapters[i].Tracks {
			card.Content.Chapters[i].Tracks[j].Key = key
			card.Content.Chapters[i].Tracks[j].OverlayLabel = fmt.Sprintf("%d", i+1)
		}
		totalDur += card.Content.Chapters[i].Duration
		if len(card.Content.Chapters[i].Tracks) > 0 {
			totalSize += card.Content.Chapters[i].Tracks[0].FileSize
		}
	}
	if card.Metadata == nil {
		card.Metadata = &yoto.Metadata{}
	}
	card.Metadata.Media.Duration = totalDur
	card.Metadata.Media.FileSize = totalSize
}
