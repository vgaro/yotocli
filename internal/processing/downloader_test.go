package processing

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestYtDlpExists checks the yt-dlp that TestMain put on PATH is there and
// runnable, so a broken setup fails on its own rather than as a confusing
// failure inside a download test.
func TestYtDlpExists(t *testing.T) {
	path, err := exec.LookPath("yt-dlp")
	if err != nil {
		t.Fatalf("yt-dlp not found on PATH: %v", err)
	}

	if out, err := exec.Command(path, "--version").CombinedOutput(); err != nil {
		t.Fatalf("%s --version failed: %v\n%s", path, err, out)
	}
}

func TestParseDownloads(t *testing.T) {
	tests := []struct {
		name    string
		output  string
		want    []Download
		wantErr bool
	}{
		{
			name:   "single file",
			output: `{"title": "Jolene", "filepath": "/tmp/yoto_import_abc.mp3"}` + "\n",
			want:   []Download{{Name: "Jolene", Path: "/tmp/yoto_import_abc.mp3"}},
		},
		{
			name: "several files keep their order",
			output: `{"title": "Jolene", "filepath": "/tmp/a.mp3"}` + "\n" +
				`{"title": "9 To 5", "filepath": "/tmp/b.mp3"}` + "\n" +
				`{"title": "Coat of Many Colors", "filepath": "/tmp/c.mp3"}` + "\n",
			want: []Download{
				{Name: "Jolene", Path: "/tmp/a.mp3"},
				{Name: "9 To 5", Path: "/tmp/b.mp3"},
				{Name: "Coat of Many Colors", Path: "/tmp/c.mp3"},
			},
		},
		{
			// JSON is why we can survive this without splitting on the quotes.
			name:   "title containing quotes and a comma",
			output: `{"title": "Buster \"Sparky\" Baxter, Cat Saver", "filepath": "/tmp/d.mp3"}`,
			want:   []Download{{Name: `Buster "Sparky" Baxter, Cat Saver`, Path: "/tmp/d.mp3"}},
		},
		{
			name:   "blank lines are ignored",
			output: "\n" + `{"title": "Jolene", "filepath": "/tmp/a.mp3"}` + "\n\n",
			want:   []Download{{Name: "Jolene", Path: "/tmp/a.mp3"}},
		},
		{
			name:   "no output",
			output: "\n  \n",
			want:   nil,
		},
		{
			name:    "not json",
			output:  "Jolene\n/tmp/a.mp3\n",
			wantErr: true,
		},
		{
			name:    "missing file path",
			output:  `{"title": "Jolene"}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseDownloads(tt.output)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("parseDownloads(%q) = %v, want error", tt.output, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseDownloads(%q) returned error: %v", tt.output, err)
			}
			if len(got) != len(tt.want) {
				t.Fatalf("parseDownloads(%q) = %v, want %v", tt.output, got, tt.want)
			}
			for i := range tt.want {
				if got[i] != tt.want[i] {
					t.Errorf("download %d = %+v, want %+v", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestLooksLikeMP3(t *testing.T) {
	tests := []struct {
		name   string
		header []byte
		want   bool
	}{
		{"id3v2 tag", []byte("ID3"), true},
		{"mpeg frame sync", []byte{0xFF, 0xFB, 0x90}, true},
		{"mpeg frame sync, other layer", []byte{0xFF, 0xE3, 0x18}, true},
		{"html error page", []byte("<!D"), false},
		{"riff wav", []byte("RIF"), false},
		{"sync bits not all set", []byte{0xFF, 0xC0, 0x00}, false},
		{"truncated", []byte{0xFF}, false},
		{"empty", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := looksLikeMP3(tt.header); got != tt.want {
				t.Errorf("looksLikeMP3(% x) = %v, want %v", tt.header, got, tt.want)
			}
		})
	}
}

// The following tests download real audio, so they need yt-dlp, ffmpeg and
// network access.

func TestDownloadFromURL_YouTubeVideo(t *testing.T) {
	downloads := download(t, "https://www.youtube.com/watch?v=Ixrje2rXLMA")

	assertNames(t, downloads, []string{"Dolly Parton - Jolene (Official Audio)"})
}

// TestDownloadFromURL_YouTubePlaylist uses a playlist URL rather than a video
// URL carrying a list parameter, so the test covers the case where the URL has
// no single video to fall back on.
func TestDownloadFromURL_YouTubePlaylist(t *testing.T) {
	downloads := download(t, "https://www.youtube.com/playlist?list=PLfnx6nNQO1XM")

	assertNames(t, downloads, []string{
		"Dolly Parton - Jolene (Official Audio)",
		"Dolly Parton - 9 To 5 (Official Video)",
		"Coat of Many Colors",
	})
}

// TestDownloadFromURL_RSSFeed serves a frozen copy of the Arthur podcast feed
// from a local HTTP server, so the episode list can't change under the test.
func TestDownloadFromURL_RSSFeed(t *testing.T) {
	feed := httptest.NewServer(http.FileServer(http.Dir("testdata")))
	t.Cleanup(feed.Close)

	downloads := download(t, feed.URL+"/arthur-feed.xml")

	assertNames(t, downloads, []string{
		"Bleep!",
		"Buster Baxter, Cat Saver",
		"What Scared Sue Ellen",
	})
}

// TestDownloadFromURL_Cleanup checks the returned cleanup function removes the
// downloads, and that calling it twice is not an error.
func TestDownloadFromURL_Cleanup(t *testing.T) {
	url := "https://www.youtube.com/watch?v=Ixrje2rXLMA"

	downloads, cleanup, err := DownloadFromURL(url)
	if err != nil {
		t.Fatalf("DownloadFromURL(%q) returned error: %v", url, err)
	}
	if len(downloads) != 1 {
		t.Fatalf("got %d downloads, want 1", len(downloads))
	}
	if _, err := os.Stat(downloads[0].Path); err != nil {
		t.Fatalf("download is missing before cleanup: %v", err)
	}

	if err := cleanup(); err != nil {
		t.Fatalf("cleanup() returned error: %v", err)
	}

	if _, err := os.Stat(downloads[0].Path); !os.IsNotExist(err) {
		t.Errorf("download still present after cleanup: %s (stat err %v)", downloads[0].Path, err)
	}
	// The temp directory itself should be gone too, not just the file in it.
	dir := filepath.Dir(downloads[0].Path)
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Errorf("temp dir still present after cleanup: %s (stat err %v)", dir, err)
	}

	if err := cleanup(); err != nil {
		t.Errorf("second cleanup() returned error: %v", err)
	}
}

// download runs DownloadFromURL and cleans up whatever it wrote once the test
// finishes.
func download(t *testing.T, url string) []Download {
	t.Helper()

	downloads, cleanup, err := DownloadFromURL(url)
	t.Cleanup(func() {
		if err := cleanup(); err != nil {
			t.Errorf("cleanup() returned error: %v", err)
		}
	})
	if err != nil {
		t.Fatalf("DownloadFromURL(%q) returned error: %v", url, err)
	}

	return downloads
}

// assertNames checks the downloads are the expected titles, in order, and that
// each one left behind a non-empty file that starts like an MP3.
func assertNames(t *testing.T, downloads []Download, wantNames []string) {
	t.Helper()

	if len(downloads) != len(wantNames) {
		t.Fatalf("got %d downloads, want %d: %+v", len(downloads), len(wantNames), downloads)
	}

	for i, want := range wantNames {
		if downloads[i].Name != want {
			t.Errorf("download %d name = %q, want %q", i, downloads[i].Name, want)
		}

		info, err := os.Stat(downloads[i].Path)
		if err != nil {
			t.Errorf("download %d (%q): %v", i, downloads[i].Name, err)
			continue
		}
		if info.Size() == 0 {
			t.Errorf("download %d (%q) is empty: %s", i, downloads[i].Name, downloads[i].Path)
			continue
		}

		header, err := readHeader(downloads[i].Path)
		if err != nil {
			t.Errorf("download %d (%q): %v", i, downloads[i].Name, err)
			continue
		}
		if !looksLikeMP3(header) {
			t.Errorf("download %d (%q) does not start like an MP3, first bytes: % x (%s)",
				i, downloads[i].Name, header, downloads[i].Path)
		}
	}
}

// readHeader returns the first few bytes of a file.
func readHeader(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	header := make([]byte, 3)
	if _, err := io.ReadFull(f, header); err != nil {
		return nil, fmt.Errorf("reading header of %s: %w", path, err)
	}

	return header, nil
}

// looksLikeMP3 reports whether the start of a file looks like an MP3. An MP3
// either opens with an ID3v2 tag - which is what the ffmpeg postprocessor
// yt-dlp uses writes - or dives straight into an MPEG audio frame, whose
// header begins with eleven set sync bits.
func looksLikeMP3(header []byte) bool {
	if len(header) < 3 {
		return false
	}
	if string(header[:3]) == "ID3" {
		return true
	}
	return header[0] == 0xFF && header[1]&0xE0 == 0xE0
}
