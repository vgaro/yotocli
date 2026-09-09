package processing

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Download is a single audio file produced by DownloadFromURL.
type Download struct {
	// Name is the title reported by yt-dlp (the video or episode title).
	Name string
	// Path is the location of the downloaded MP3 on disk. It lives in a
	// temporary directory owned by DownloadFromURL's cleanup function, so
	// anything that needs the file to outlive the cleanup has to copy it.
	Path string
}

// DownloadFromURL downloads audio from a URL using yt-dlp, converting to MP3.
// A single URL can resolve to several files (a playlist, or an RSS feed), so
// the downloads are returned in the order yt-dlp produced them.
//
// Shelling out to yt-dlp rather than fetching the media ourselves buys us two
// things. The first is reach: yt-dlp has extractors for well over a thousand
// sites, so the same code path handles a YouTube video, a YouTube playlist, a
// podcast RSS feed and a plain link to an MP3, and any site it learns about
// later works here with no changes on our side. The second is maintenance:
// YouTube in particular keeps changing its player and internal API in ways
// that break downloaders, and yt-dlp tracks those changes for us. That does
// mean a stale yt-dlp stops working, so an out-of-date install is a likely
// cause of extraction failures reported here.
//
// The downloads are written to a temporary directory of their own, which the
// returned cleanup function removes. Callers should run it once they are done
// with the files, and should defer it before checking the error so that it runs
// on every path:
//
//	downloads, cleanup, err := processing.DownloadFromURL(url)
//	defer cleanup()
//	if err != nil {
//		return err
//	}
//
// The cleanup function is never nil, which is what makes deferring it up front
// safe, and it can be called more than once.
func DownloadFromURL(url string) ([]Download, func() error, error) {
	noCleanup := func() error { return nil }

	// 1. Check if yt-dlp exists
	if _, err := exec.LookPath("yt-dlp"); err != nil {
		return nil, noCleanup, fmt.Errorf("yt-dlp not found: please install it (pip install yt-dlp)")
	}

	// 2. Download into a directory of our own, so that cleanup is a single
	// RemoveAll and concurrent imports of the same URL cannot collide.
	tmpDir, err := os.MkdirTemp("", "yoto_import_")
	if err != nil {
		return nil, noCleanup, fmt.Errorf("failed to create temp dir: %w", err)
	}
	cleanup := func() error { return os.RemoveAll(tmpDir) }

	// fail is for the error paths below, which clean up after themselves rather
	// than handing back a directory of partial downloads. A cleanup failure is
	// joined onto the error being returned instead of being dropped: it means
	// files have been left behind in the system temp directory.
	fail := func(err error) ([]Download, func() error, error) {
		if cerr := cleanup(); cerr != nil {
			err = errors.Join(err, fmt.Errorf("failed to remove temp dir %s: %w", tmpDir, cerr))
		}
		return nil, noCleanup, err
	}

	// Output template: <temp_dir>/<id>.mp3
	// We let yt-dlp name the file from the ID to avoid weird chars issues.
	outputTemplate := filepath.Join(tmpDir, "%(id)s.%(ext)s")

	// 3. Print one JSON object per file, once the file has reached its final
	// location. JSON rather than plain fields so that a title containing a
	// newline or a delimiter can't corrupt the output.
	cmd := exec.Command("yt-dlp",
		"-x",                    // Extract audio
		"--audio-format", "mp3", // Convert to mp3
		"--audio-quality", "0", // Best quality
		"-o", outputTemplate, // Output path
		"--print", "after_move:%(.{title,filepath})j", // Report what was written
		"--no-simulate",
		url,
	)

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr // Capture stderr for error reporting

	// On failure there is nothing worth handing back, so drop any partial
	// downloads here instead of relying on the caller to do it.
	if err := cmd.Run(); err != nil {
		return fail(fmt.Errorf("yt-dlp failed: %w\nStderr: %s", err, stderr.String()))
	}

	downloads, err := parseDownloads(out.String())
	if err != nil {
		return fail(err)
	}
	if len(downloads) == 0 {
		return fail(fmt.Errorf("yt-dlp downloaded nothing from %s\nStderr: %s", url, stderr.String()))
	}

	return downloads, cleanup, nil
}

// parseDownloads reads the JSON lines written by the --print template above.
func parseDownloads(output string) ([]Download, error) {
	var downloads []Download

	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var entry struct {
			Title    string `json:"title"`
			Filepath string `json:"filepath"`
		}
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			return nil, fmt.Errorf("unexpected output from yt-dlp: %q: %w", line, err)
		}
		if entry.Filepath == "" {
			return nil, fmt.Errorf("yt-dlp reported no file path for %q", entry.Title)
		}

		downloads = append(downloads, Download{Name: entry.Title, Path: entry.Filepath})
	}

	return downloads, nil
}
