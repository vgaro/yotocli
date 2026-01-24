package processing

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// DownloadFromURL downloads audio from a URL using yt-dlp, converting to MP3.
// Returns the path to the downloaded file and the title.
func DownloadFromURL(url string) (string, string, error) {
	// 1. Check if yt-dlp exists
	if _, err := exec.LookPath("yt-dlp"); err != nil {
		return "", "", fmt.Errorf("yt-dlp not found: please install it (pip install yt-dlp)")
	}

	// 2. Create temp directory
	tmpDir := os.TempDir()
	
	// 3. Get metadata (Title) first to verify and name file
	// We use a specific template for the filename to avoid weird chars issues during download
	// We will rely on yt-dlp to handle the file creation
	
	// Output template: <temp_dir>/<id>.mp3
	outputTemplate := filepath.Join(tmpDir, "yoto_import_%(id)s.%(ext)s")

	cmd := exec.Command("yt-dlp",
		"-x",                    // Extract audio
		"--audio-format", "mp3", // Convert to mp3
		"--audio-quality", "0",  // Best quality
		"-o", outputTemplate,    // Output path
		"--print", "after_move:filepath", // Print final filename
		"--print", "title",      // Print title
		"--no-simulate",
		url,
	)

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = os.Stderr // Let user see progress

	if err := cmd.Run(); err != nil {
		return "", "", fmt.Errorf("yt-dlp failed: %w", err)
	}

	// Parse output
	// Expected:
	// Title Name
	// /path/to/yoto_import_12345.mp3
	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	if len(lines) < 2 {
		return "", "", fmt.Errorf("unexpected output from yt-dlp: %s", out.String())
	}

	title := lines[0]
	path := lines[len(lines)-1] // Path is usually last

	return path, title, nil
}
