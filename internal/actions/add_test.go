package actions

import (
	"testing"

	"github.com/vgaro/yotocli/pkg/yoto"
)

func TestTrackTitle(t *testing.T) {
	tests := []struct {
		name     string
		title    string
		filePath string
		want     string
	}{
		{
			// The import path: the file is named after the yt-dlp ID, and the
			// real episode title comes in from Download.Name.
			name:     "given title wins over the file name",
			title:    "Buster Baxter, Cat Saver",
			filePath: "/tmp/yoto_import_123/Ixrje2rXLMA.mp3",
			want:     "Buster Baxter, Cat Saver",
		},
		{
			name:     "no title falls back to the file name",
			title:    "",
			filePath: "/home/me/audio/Bedtime Story.mp3",
			want:     "Bedtime Story",
		},
		{
			name:     "blank title falls back to the file name",
			title:    "   ",
			filePath: "/home/me/audio/Bedtime Story.mp3",
			want:     "Bedtime Story",
		},
		{
			name:     "fall back on a file with no extension",
			title:    "",
			filePath: "/home/me/audio/track",
			want:     "track",
		},
		{
			// Only the last extension goes, so this keeps the ".part" that a
			// name like this carries.
			name:     "fall back on a file with two extensions",
			title:    "",
			filePath: "/home/me/audio/track.part.mp3",
			want:     "track.part",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := trackTitle(tt.title, tt.filePath); got != tt.want {
				t.Errorf("trackTitle(%q, %q) = %q, want %q", tt.title, tt.filePath, got, tt.want)
			}
		})
	}
}

// chapters builds a card holding one chapter per title, as the shape the insert
// tests care about.
func chapters(titles ...string) *yoto.Card {
	card := &yoto.Card{Content: &yoto.Content{}}
	for _, title := range titles {
		card.Content.Chapters = append(card.Content.Chapters, yoto.Chapter{Title: title})
	}
	return card
}

func chapterTitles(card *yoto.Card) []string {
	var titles []string
	for _, chapter := range card.Content.Chapters {
		titles = append(titles, chapter.Title)
	}
	return titles
}

func TestInsertChapters(t *testing.T) {
	tests := []struct {
		name     string
		card     *yoto.Card
		added    []string
		position int
		want     []string
	}{
		{
			// An import with no position asked for: the batch goes on the end,
			// in the order it was given.
			name:     "append keeps the order of the batch",
			card:     chapters("one", "two"),
			added:    []string{"three", "four"},
			position: -1,
			want:     []string{"one", "two", "three", "four"},
		},
		{
			name:     "insert at the front",
			card:     chapters("one", "two"),
			added:    []string{"intro", "warning"},
			position: 0,
			want:     []string{"intro", "warning", "one", "two"},
		},
		{
			name:     "insert in the middle",
			card:     chapters("one", "two", "three"),
			added:    []string{"inserted"},
			position: 1,
			want:     []string{"one", "inserted", "two", "three"},
		},
		{
			// "Name/9" against a card with three chapters: past the end is an
			// append rather than an error.
			name:     "a position past the end appends",
			card:     chapters("one", "two"),
			added:    []string{"three"},
			position: 8,
			want:     []string{"one", "two", "three"},
		},
		{
			name:     "insert into an empty playlist",
			card:     chapters(),
			added:    []string{"one", "two"},
			position: 0,
			want:     []string{"one", "two"},
		},
		{
			name:     "a card with no content at all",
			card:     &yoto.Card{},
			added:    []string{"one"},
			position: -1,
			want:     []string{"one"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			added := make([]yoto.Chapter, 0, len(tt.added))
			for _, title := range tt.added {
				added = append(added, yoto.Chapter{Title: title})
			}

			insertChapters(tt.card, added, tt.position)

			got := chapterTitles(tt.card)
			if len(got) != len(tt.want) {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Fatalf("got %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestSetMediaTotals(t *testing.T) {
	card := &yoto.Card{
		Content: &yoto.Content{
			Chapters: []yoto.Chapter{
				{Duration: 30, Tracks: []yoto.Track{{FileSize: 100}}},
				{Duration: 45, Tracks: []yoto.Track{{FileSize: 200}}},
				// A chapter with no track still counts towards the duration,
				// and must not panic on the missing file size.
				{Duration: 5},
			},
		},
	}

	setMediaTotals(card)

	if card.Metadata == nil {
		t.Fatal("setMediaTotals left Metadata nil")
	}
	if card.Metadata.Media.Duration != 80 {
		t.Errorf("Duration = %d, want 80", card.Metadata.Media.Duration)
	}
	if card.Metadata.Media.FileSize != 300 {
		t.Errorf("FileSize = %d, want 300", card.Metadata.Media.FileSize)
	}
}

func TestResolveIcon(t *testing.T) {
	tests := []struct {
		name   string
		iconID string
		want   string
	}{
		{"no icon falls back to the default", "", defaultIcon},
		{"a bare hash gets the yoto prefix", "aUm9i3ex", "yoto:#aUm9i3ex"},
		{"an already prefixed icon is left alone", "yoto:#aUm9i3ex", "yoto:#aUm9i3ex"},
		{"a URL is left alone", "https://example.com/icon.png", "https://example.com/icon.png"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := resolveIcon(tt.iconID); got != tt.want {
				t.Errorf("resolveIcon(%q) = %q, want %q", tt.iconID, got, tt.want)
			}
		})
	}
}

func TestProgress(t *testing.T) {
	if got := progress(0, 1); got != "" {
		t.Errorf("progress(0, 1) = %q, want no label for a lone file", got)
	}
	if got := progress(2, 12); got != "[3/12] " {
		t.Errorf("progress(2, 12) = %q, want %q", got, "[3/12] ")
	}
}
