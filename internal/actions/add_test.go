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

// chapter builds one chapter: the audio named however the caller says, and the
// same icon on the chapter and its only track.
func chapter(title string, trackURL string, icon string) yoto.Chapter {
	return yoto.Chapter{
		Title:   title,
		Display: yoto.Display{Icon16x16: icon},
		Tracks: []yoto.Track{{
			Title:    title,
			TrackURL: trackURL,
			Display:  yoto.Display{Icon16x16: icon},
		}},
	}
}

// uploaded is what AddTracks holds after uploading: the audio named as
// "yoto:#<hash>", one hash per title, and the default icon, since no caller named
// one.
func uploaded(titles ...string) []yoto.Chapter {
	chapters := make([]yoto.Chapter, 0, len(titles))
	for _, title := range titles {
		chapters = append(chapters, chapter(title, "yoto:#audio-"+title, defaultIcon))
	}
	return chapters
}

// decorated is what a card looks like once it has been read back from the API and
// someone has picked icons for it in the Yoto app. The shapes are the point: the
// same audio that was written as "yoto:#<hash>" reads back as a signed CDN URL
// carrying the hash in its fragment, and an icon reads back as an https URL.
// Treating those as different is what made a sync reset every icon.
func decorated(titles ...string) []yoto.Chapter {
	chapters := make([]yoto.Chapter, 0, len(titles))
	for _, title := range titles {
		chapters = append(chapters, chapter(title, signedURL("audio-"+title), iconURL("icon-"+title)))
	}
	return chapters
}

func signedURL(audio string) string {
	return "https://secure-media.yotoplay.com/prefix~/" + audio + "?Expires=1788994229&Signature=abc__#sha256=" + audio
}

func iconURL(icon string) string {
	return "https://card-content.yotoplay.com/prefix~/" + icon
}

func chapterTitles(chapters []yoto.Chapter) []string {
	var titles []string
	for _, chapter := range chapters {
		titles = append(titles, chapter.Title)
	}
	return titles
}

func assertTitles(t *testing.T, got []yoto.Chapter, want []string) {
	t.Helper()
	titles := chapterTitles(got)
	if len(titles) != len(want) {
		t.Fatalf("chapters = %v, want %v", titles, want)
	}
	for i := range titles {
		if titles[i] != want[i] {
			t.Fatalf("chapters = %v, want %v", titles, want)
		}
	}
}

func TestAddChapters(t *testing.T) {
	tests := []struct {
		name     string
		existing []string
		added    []string
		position int
		want     []string
	}{
		{
			// An import with no position asked for: the batch goes on the end,
			// in the order it was given.
			name:     "append keeps the order of the batch",
			existing: []string{"one", "two"},
			added:    []string{"three", "four"},
			position: -1,
			want:     []string{"one", "two", "three", "four"},
		},
		{
			name:     "insert at the front",
			existing: []string{"one", "two"},
			added:    []string{"intro", "warning"},
			position: 0,
			want:     []string{"intro", "warning", "one", "two"},
		},
		{
			name:     "insert in the middle",
			existing: []string{"one", "two", "three"},
			added:    []string{"inserted"},
			position: 1,
			want:     []string{"one", "inserted", "two", "three"},
		},
		{
			// "Name/9" against a card with two chapters: past the end is an
			// append rather than an error.
			name:     "a position past the end appends",
			existing: []string{"one", "two"},
			added:    []string{"three"},
			position: 8,
			want:     []string{"one", "two", "three"},
		},
		{
			name:     "insert into an empty playlist",
			existing: nil,
			added:    []string{"one", "two"},
			position: 0,
			want:     []string{"one", "two"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			existing := uploaded(tt.existing...)
			got := addChapters(existing, uploaded(tt.added...), tt.position, false)

			assertTitles(t, got, tt.want)

			// The card's own slice must not be written through, since it is
			// still the caller's view of what was there before.
			assertTitles(t, existing, tt.existing)
		})
	}
}

func TestAddChapters_Sync(t *testing.T) {
	t.Run("drops what the source no longer lists", func(t *testing.T) {
		existing := decorated("Ep 1", "Ep 2", "Ep 3")

		got := addChapters(existing, uploaded("Ep 1", "Ep 3"), -1, true)

		assertTitles(t, got, []string{"Ep 1", "Ep 3"})
		// Both survivors were already on the card, so both keep their icons. The
		// icon is still the https URL it was read back as; UpdateCard puts it
		// back into "yoto:#<hash>" form on the way out.
		for i, want := range []string{iconURL("icon-Ep 1"), iconURL("icon-Ep 3")} {
			if got[i].Display.Icon16x16 != want {
				t.Errorf("chapter %d icon = %q, want %q", i, got[i].Display.Icon16x16, want)
			}
			if got[i].Tracks[0].Display.Icon16x16 != want {
				t.Errorf("chapter %d track icon = %q, want %q", i, got[i].Tracks[0].Display.Icon16x16, want)
			}
		}
	})

	t.Run("a new episode among old ones", func(t *testing.T) {
		existing := decorated("Ep 1", "Ep 2")

		// A feed listing newest first.
		got := addChapters(existing, uploaded("Ep 3", "Ep 1", "Ep 2"), -1, true)

		assertTitles(t, got, []string{"Ep 3", "Ep 1", "Ep 2"})
		if got[0].Display.Icon16x16 != defaultIcon {
			t.Errorf("the new episode's icon = %q, want the default", got[0].Display.Icon16x16)
		}
		if got[1].Display.Icon16x16 != iconURL("icon-Ep 1") {
			t.Errorf("kept icon = %q, want the one on the card", got[1].Display.Icon16x16)
		}
	})

	t.Run("a renamed episode keeps its icon", func(t *testing.T) {
		// Matching on audio rather than on the title is what makes this work: the
		// feed has retitled the episode, but it is the same audio.
		existing := []yoto.Chapter{chapter("Ep 1", signedURL("audio-x"), "yoto:#icon-custom")}

		got := addChapters(existing, []yoto.Chapter{chapter("Episode One", "yoto:#audio-x", defaultIcon)}, -1, true)

		assertTitles(t, got, []string{"Episode One"})
		if got[0].Display.Icon16x16 != "yoto:#icon-custom" {
			t.Errorf("icon = %q, want the one on the card", got[0].Display.Icon16x16)
		}
	})

	t.Run("an episode republished with different audio counts as new", func(t *testing.T) {
		existing := []yoto.Chapter{chapter("Ep 1", signedURL("audio-old"), "yoto:#icon-custom")}

		got := addChapters(existing, []yoto.Chapter{chapter("Ep 1", "yoto:#audio-new", defaultIcon)}, -1, true)

		if got[0].Display.Icon16x16 != defaultIcon {
			t.Errorf("icon = %q, want the default: this is not the audio the icon was chosen for", got[0].Display.Icon16x16)
		}
		if got[0].Tracks[0].TrackURL != "yoto:#audio-new" {
			t.Errorf("audio = %q, want the new one", got[0].Tracks[0].TrackURL)
		}
	})

	t.Run("an icon named by the caller wins", func(t *testing.T) {
		existing := decorated("Ep 1")

		got := addChapters(existing, []yoto.Chapter{chapter("Ep 1", "yoto:#audio-Ep 1", "yoto:#chosen")}, -1, true)

		if got[0].Display.Icon16x16 != "yoto:#chosen" {
			t.Errorf("icon = %q, want the one the caller named", got[0].Display.Icon16x16)
		}
	})

	t.Run("nothing on the card yet", func(t *testing.T) {
		got := addChapters(nil, uploaded("Ep 1"), -1, true)

		assertTitles(t, got, []string{"Ep 1"})
		if got[0].Display.Icon16x16 != defaultIcon {
			t.Errorf("icon = %q, want the default", got[0].Display.Icon16x16)
		}
	})

	t.Run("a chapter on the card with no tracks", func(t *testing.T) {
		// Nothing to match on, and nothing to crash on either.
		existing := []yoto.Chapter{{Title: "Ep 1"}}

		got := addChapters(existing, uploaded("Ep 1"), -1, true)

		assertTitles(t, got, []string{"Ep 1"})
		if got[0].Display.Icon16x16 != defaultIcon {
			t.Errorf("icon = %q, want the default", got[0].Display.Icon16x16)
		}
	})

	t.Run("inheriting an icon does not modify the card", func(t *testing.T) {
		existing := decorated("Ep 1")

		addChapters(existing, uploaded("Ep 1"), -1, true)

		if existing[0].Display.Icon16x16 != iconURL("icon-Ep 1") {
			t.Errorf("the chapter on the card was modified: %q", existing[0].Display.Icon16x16)
		}
	})
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
