package actions

import "testing"

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
