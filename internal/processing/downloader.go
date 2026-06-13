package processing

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type DownloadedItem struct {
	Path  string
	Title string
}

// DownloadFromURL downloads audio from a URL using yt-dlp, converting to MP3.
// Returns a slice of downloaded items.
func DownloadFromURL(url string) ([]DownloadedItem, error) {
	// 1. Check if yt-dlp exists
	if _, err := exec.LookPath("yt-dlp"); err != nil {
		return nil, fmt.Errorf("yt-dlp not found: please install it (pip install yt-dlp)")
	}

	// 2. Create temp directory
	tmpDir := os.TempDir()

	// Output template: <temp_dir>/yoto_import_%(id)s.%(ext)s
	outputTemplate := filepath.Join(tmpDir, "yoto_import_%(id)s.%(ext)s")

	cmd := exec.Command("yt-dlp",
		"--remote-components", "ejs:github",
		"--js-runtimes", "node",
		"-x",                    // Extract audio
		"--audio-format", "mp3", // Convert to mp3
		"--audio-quality", "0",  // Best quality
		"-o", outputTemplate,    // Output path
		"--print", "title",      // Print title
		"--print", "after_move:filepath", // Print final filename
		"--no-simulate",
		url,
	)

	// To support playlists, we need to handle multiple titles and filepaths.
	// We'll remove --no-playlist if it's there (it wasn't in the original code but good to be sure).

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("yt-dlp failed: %w\nStderr: %s", err, stderr.String())
	}

	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	if len(lines) < 2 || len(lines)%2 != 0 {
		return nil, fmt.Errorf("unexpected output from yt-dlp (must be pairs of title/path): %s", out.String())
	}

	var items []DownloadedItem
	for i := 0; i < len(lines); i += 2 {
		items = append(items, DownloadedItem{
			Title: lines[i],
			Path:  lines[i+1],
		})
	}

	return items, nil
}

