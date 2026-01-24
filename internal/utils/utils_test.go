package utils

import (
	"testing"

	"github.com/vgaro/yotocli/pkg/yoto"
)

func TestParseIndex(t *testing.T) {
	tests := []struct {
		input    string
		expected int
		wantErr  bool
	}{
		{"1", 1, false},
		{"10", 10, false},
		{"0", 0, false},
		{"abc", 0, true},
		{"", 0, true},
	}

	for _, tt := range tests {
		got, err := ParseIndex(tt.input)
		if (err != nil) != tt.wantErr {
			t.Errorf("ParseIndex(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			continue
		}
		if got != tt.expected {
			t.Errorf("ParseIndex(%q) = %v, want %v", tt.input, got, tt.expected)
		}
	}
}

func TestFindCard(t *testing.T) {
	cards := []yoto.Card{
		{CardID: "uuid-1", Title: "Bedtime Stories"},
		{CardID: "uuid-2", Title: "Dance Party"},
		{CardID: "uuid-3", Title: "Morning News"},
	}

	tests := []struct {
		query    string
		wantID   string
		wantNone bool
	}{
		{"1", "uuid-1", false},              // Index match
		{"2", "uuid-2", false},              // Index match
		{"uuid-3", "uuid-3", false},         // ID match
		{"Bedtime", "uuid-1", false},        // Fuzzy Title match
		{"dance", "uuid-2", false},          // Case-insensitive match
		{"News", "uuid-3", false},           // Substring match
		{"NonExistent", "", true},           // No match
		{"4", "", true},                     // Out of bounds index
	}

	for _, tt := range tests {
		got := FindCard(cards, tt.query)
		if tt.wantNone {
			if got != nil {
				t.Errorf("FindCard(%q) = %v, want nil", tt.query, got)
			}
		} else {
			if got == nil {
				t.Errorf("FindCard(%q) = nil, want ID %s", tt.query, tt.wantID)
			} else if got.CardID != tt.wantID {
				t.Errorf("FindCard(%q) ID = %s, want %s", tt.query, got.CardID, tt.wantID)
			}
		}
	}
}

func TestReorderPlaylist(t *testing.T) {
	// Create a card with messy keys/overlays
	card := &yoto.Card{
		Content: &yoto.Content{
			Chapters: []yoto.Chapter{
				{Title: "Track A", Key: "99", OverlayLabel: "99", Tracks: []yoto.Track{{Key: "99"}}},
				{Title: "Track B", Key: "05", OverlayLabel: "old", Tracks: []yoto.Track{{Key: "05"}}},
			},
		},
	}

	ReorderPlaylist(card)

	// Verify Track A is now #1
	if card.Content.Chapters[0].Key != "01" || card.Content.Chapters[0].OverlayLabel != "1" {
		t.Errorf("Chapter 0 not reordered correctly: %+v", card.Content.Chapters[0])
	}
	if card.Content.Chapters[0].Tracks[0].Key != "01" {
		t.Errorf("Track 0 not reordered correctly: %+v", card.Content.Chapters[0].Tracks[0])
	}

	// Verify Track B is now #2
	if card.Content.Chapters[1].Key != "02" || card.Content.Chapters[1].OverlayLabel != "2" {
		t.Errorf("Chapter 1 not reordered correctly: %+v", card.Content.Chapters[1])
	}
}
