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

	if card.Content == nil || trackIndex < 1 || trackIndex > len(card.Content.Chapters) {
		return fmt.Errorf("invalid track index: %d", trackIndex)
	}

	idx := trackIndex - 1
	// Remove from slice
	card.Content.Chapters = append(card.Content.Chapters[:idx], card.Content.Chapters[idx+1:]...)

	recalculateMetadata(card)

	return client.UpdateCard(card.CardID, card)
}

// MoveTrack moves a track from srcIndex to destIndex (1-based).
func MoveTrack(client *yoto.Client, cardID string, srcIndex int, destIndex int) error {
	card, err := client.GetCard(cardID)
	if err != nil {
		return err
	}

	count := len(card.Content.Chapters)
	if card.Content == nil || srcIndex < 1 || srcIndex > count {
		return fmt.Errorf("invalid source index: %d", srcIndex)
	}
	if destIndex < 1 || destIndex > count {
		return fmt.Errorf("invalid destination index: %d", destIndex)
	}

	srcIdx := srcIndex - 1
	destIdx := destIndex - 1

	if srcIdx == destIdx {
		return nil // No op
	}

	elem := card.Content.Chapters[srcIdx]
	// Remove
	card.Content.Chapters = append(card.Content.Chapters[:srcIdx], card.Content.Chapters[srcIdx+1:]...)
	// Insert
	if destIdx >= len(card.Content.Chapters) {
		card.Content.Chapters = append(card.Content.Chapters, elem)
	} else {
		// Insert into slice
		card.Content.Chapters = append(card.Content.Chapters[:destIdx], append([]yoto.Chapter{elem}, card.Content.Chapters[destIdx:]...)...)
	}

	recalculateMetadata(card)

	return client.UpdateCard(card.CardID, card)
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
