package yoto

import (
	"encoding/json"
	"reflect"
	"testing"
)

// realCard is a card as the API returns one, trimmed to two chapters. Most of
// what is here - the cover art, the playback config, the publisher's
// copyrights, a chapter's ambient light - is not modelled by this client, and
// all of it has to survive being read and written back.
const realCard = `{
  "cardId": "1bHLK",
  "userId": "auth0|000000000000000000000000",
  "title": "Odd Squadcast",
  "createdAt": "2026-09-09T21:40:20.627Z",
  "updatedAt": "2026-09-09T21:47:56.692Z",
  "content": {
    "activity": "yoto_Player",
    "availability": "",
    "config": {"onlineOnly": false, "resumeTimeout": 2592000},
    "cover": {"imageL": "https://card-content.yotoplay.com/pub/coverhash"},
    "playbackType": "linear",
    "version": "1",
    "chapters": [
      {
        "key": "01",
        "title": "Broadcast Seven",
        "overlayLabel": "1",
        "duration": 585,
        "fileSize": 7191904,
        "ambient": null,
        "availableFrom": null,
        "display": {"icon16x16": "https://card-content.yotoplay.com/aUm9i3ex3qqAMYBv-i-O-pYMKuMJGICtR3Vhf289u2Q"},
        "tracks": [
          {
            "key": "01",
            "title": "Broadcast Seven",
            "overlayLabel": "1",
            "trackUrl": "https://secure-media.yotoplay.com/signed?Expires=1788994229#sha256=U-8tT6z7rt5MtIMzjANOtLmB0E7FamNOlQHBP1Vn4og",
            "duration": 585,
            "fileSize": 7191904,
            "format": "opus",
            "type": "audio",
            "ambient": null,
            "display": {"icon16x16": "https://card-content.yotoplay.com/aUm9i3ex3qqAMYBv-i-O-pYMKuMJGICtR3Vhf289u2Q"}
          }
        ]
      },
      {
        "key": "02",
        "title": "Broadcast Eight",
        "overlayLabel": "2",
        "duration": 400,
        "fileSize": 5973088,
        "display": null,
        "tracks": [
          {
            "key": "02",
            "title": "Broadcast Eight",
            "overlayLabel": "2",
            "trackUrl": "yoto:#Ab3dYXAN9DsAg7BcnRDwsdcxvOU_9lWi_ZTvJnMY55o",
            "duration": 400,
            "fileSize": 5973088,
            "format": "opus",
            "type": "audio",
            "display": null
          }
        ]
      }
    ]
  },
  "metadata": {
    "abridged": false,
    "accents": [],
    "authors": [],
    "copyrights": ["©"],
    "narrators": [],
    "cover": {"imageL": "https://card-content.yotoplay.com/pub/coverhash"},
    "media": {"duration": 3623, "fileSize": 43693954}
  }
}`

// TestCardRoundTrip is the whole point of the passthrough: an update replaces the
// card with what it is sent, so reading a card and writing it back unchanged has
// to be a no-op. It was not, and cards lost their cover art.
func TestCardRoundTrip(t *testing.T) {
	var card Card
	if err := json.Unmarshal([]byte(realCard), &card); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	written, err := json.Marshal(&card)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var before, after any
	if err := json.Unmarshal([]byte(realCard), &before); err != nil {
		t.Fatalf("unmarshal original: %v", err)
	}
	if err := json.Unmarshal(written, &after); err != nil {
		t.Fatalf("unmarshal written: %v", err)
	}
	if !reflect.DeepEqual(before, after) {
		t.Errorf("writing back an unchanged card changed it\n before: %s\n after:  %s", realCard, written)
	}
}

// TestCardRoundTrip_WithChanges checks the other half of the deal: what the
// program did change has to be what gets written.
func TestCardRoundTrip_WithChanges(t *testing.T) {
	var card Card
	if err := json.Unmarshal([]byte(realCard), &card); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	card.Title = "Odd Squad"
	card.Content.Chapters = card.Content.Chapters[:1]
	card.Content.Chapters[0].Display.Icon16x16 = "yoto:#chosen"
	card.Metadata.Media.Duration = 585

	written, err := json.Marshal(&card)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(written, &got); err != nil {
		t.Fatalf("unmarshal written: %v", err)
	}

	if got["title"] != "Odd Squad" {
		t.Errorf("title = %v, want the new one", got["title"])
	}
	content := got["content"].(map[string]any)
	chapters := content["chapters"].([]any)
	if len(chapters) != 1 {
		t.Fatalf("chapters = %d, want 1", len(chapters))
	}
	chapter := chapters[0].(map[string]any)
	if icon := chapter["display"].(map[string]any)["icon16x16"]; icon != "yoto:#chosen" {
		t.Errorf("icon = %v, want the new one", icon)
	}
	// Unmodelled fields of a chapter that stayed are still there.
	if size := chapter["fileSize"]; size != float64(7191904) {
		t.Errorf("chapter fileSize = %v, want the original", size)
	}
	// And the card keeps its cover, which is what a track-list change kept
	// wiping.
	metadata := got["metadata"].(map[string]any)
	cover, ok := metadata["cover"].(map[string]any)
	if !ok || cover["imageL"] != "https://card-content.yotoplay.com/pub/coverhash" {
		t.Errorf("metadata cover = %v, want the original", metadata["cover"])
	}
	if media := metadata["media"].(map[string]any); media["duration"] != float64(585) {
		t.Errorf("media duration = %v, want the new one", media["duration"])
	}
}

// TestCardRoundTrip_NewCard checks that a card built in memory, with nothing
// captured from the API, still writes the fields the API needs.
func TestCardRoundTrip_NewCard(t *testing.T) {
	card := Card{
		Title: "New",
		Content: &Content{Chapters: []Chapter{{
			Key:     "01",
			Title:   "One",
			Display: Display{Icon16x16: "yoto:#icon"},
			Tracks:  []Track{{Key: "01", Title: "One", TrackURL: "yoto:#hash", Type: "audio"}},
		}}},
	}

	written, err := json.Marshal(&card)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(written, &got); err != nil {
		t.Fatalf("unmarshal written: %v", err)
	}
	if got["title"] != "New" {
		t.Errorf("title = %v", got["title"])
	}
	chapter := got["content"].(map[string]any)["chapters"].([]any)[0].(map[string]any)
	if icon := chapter["display"].(map[string]any)["icon16x16"]; icon != "yoto:#icon" {
		t.Errorf("icon = %v", icon)
	}
	track := chapter["tracks"].([]any)[0].(map[string]any)
	if track["trackUrl"] != "yoto:#hash" {
		t.Errorf("trackUrl = %v", track["trackUrl"])
	}
	// A track with no icon has no display at all, which is how the API returns
	// one, rather than an empty icon that would look like a change.
	if track["display"] != nil {
		t.Errorf("display = %v, want null", track["display"])
	}
}

func TestChapterInheritFrom(t *testing.T) {
	var card Card
	if err := json.Unmarshal([]byte(realCard), &card); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	prev := card.Content.Chapters[0]

	// The same audio, uploaded again: a new title from the source, the default
	// icon, and no idea about anything the app recorded.
	fresh := Chapter{
		Title:    "Broadcast 7",
		Duration: 585,
		Display:  Display{Icon16x16: "yoto:#default"},
		Tracks: []Track{{
			Title:    "Broadcast 7",
			TrackURL: "yoto:#U-8tT6z7rt5MtIMzjANOtLmB0E7FamNOlQHBP1Vn4og",
			Duration: 585,
			Format:   "opus",
			Type:     "audio",
		}},
	}

	got := fresh.InheritFrom(prev)

	if got.Title != "Broadcast 7" {
		t.Errorf("title = %q, want the one the source now gives", got.Title)
	}
	if got.Tracks[0].TrackURL != "yoto:#U-8tT6z7rt5MtIMzjANOtLmB0E7FamNOlQHBP1Vn4og" {
		t.Errorf("trackUrl = %q, want the new one", got.Tracks[0].TrackURL)
	}
	if got.Display.Icon16x16 != prev.Display.Icon16x16 {
		t.Errorf("icon = %q, want the one on the card", got.Display.Icon16x16)
	}
	if got.Tracks[0].Display.Icon16x16 != prev.Tracks[0].Display.Icon16x16 {
		t.Errorf("track icon = %q, want the one on the card", got.Tracks[0].Display.Icon16x16)
	}

	// Unmodelled fields travel with the presentation: this is how a chapter's
	// ambient light or scheduled availability survives a sync.
	written, err := json.Marshal(got)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var fields map[string]any
	if err := json.Unmarshal(written, &fields); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if _, ok := fields["availableFrom"]; !ok {
		t.Errorf("availableFrom was dropped: %s", written)
	}
	if fields["title"] != "Broadcast 7" {
		t.Errorf("written title = %v, want the new one", fields["title"])
	}

	// Inheriting must not write through to the chapter still on the card.
	if prev.Tracks[0].Title != "Broadcast Seven" {
		t.Errorf("the chapter on the card was modified: %q", prev.Tracks[0].Title)
	}
}
