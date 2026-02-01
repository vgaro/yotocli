package actions

import (
	"testing"

	"github.com/vgaro/yotocli/pkg/yoto"
)

func TestPerformRemoveTrack(t *testing.T) {
	card := &yoto.Card{
		Content: &yoto.Content{
			Chapters: []yoto.Chapter{
				{Title: "1"}, {Title: "2"}, {Title: "3"},
			},
		},
	}

	err := performRemoveTrack(card, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(card.Content.Chapters) != 2 {
		t.Errorf("expected 2 chapters, got %d", len(card.Content.Chapters))
	}
	if card.Content.Chapters[0].Title != "1" || card.Content.Chapters[1].Title != "3" {
		t.Errorf("unexpected content: %+v", card.Content.Chapters)
	}
}

func TestPerformInsertTrack(t *testing.T) {
	card := &yoto.Card{
		Content: &yoto.Content{
			Chapters: []yoto.Chapter{
				{Title: "A"}, {Title: "B"},
			},
		},
	}

	newChap := yoto.Chapter{Title: "C"}

	// Insert at 2 -> [A, C, B]
	performInsertTrack(card, newChap, 2)

	if len(card.Content.Chapters) != 3 {
		t.Errorf("expected 3 chapters, got %d", len(card.Content.Chapters))
	}
	if card.Content.Chapters[1].Title != "C" {
		t.Errorf("expected C at index 1, got %s", card.Content.Chapters[1].Title)
	}
	
	// Test Append
	newChap2 := yoto.Chapter{Title: "D"}
	performInsertTrack(card, newChap2, 5) // > count
	if len(card.Content.Chapters) != 4 {
		t.Errorf("expected 4 chapters, got %d", len(card.Content.Chapters))
	}
	if card.Content.Chapters[3].Title != "D" {
		t.Errorf("expected D at end, got %s", card.Content.Chapters[3].Title)
	}
}
