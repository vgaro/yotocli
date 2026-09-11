package processing

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// ytDlpURL is the *latest* yt-dlp release build, deliberately not a pinned
// version.
//
// YouTube changes its player and its internal API often, and each change tends
// to break every yt-dlp release older than the fix: extraction starts failing
// with things like "Requested format is not available" or HTTP 400s from the
// API. So a yt-dlp that worked last month is not much use, and the version
// packaged by a distro is usually far enough behind to be broken already -
// Debian's yt-dlp 2023.03.04, for instance, can no longer download from
// YouTube at all.
//
// Because of that, these tests download the current release for themselves
// instead of using whatever happens to be installed. It keeps CI and
// developer machines testing the same thing, and means a failure points at
// our code rather than at a stale yt-dlp. The trade-offs are that the tests
// need network access and GitHub to be reachable, and that they are not
// reproducible over time: a future yt-dlp release could break them, which is
// something we want to find out about anyway.
const ytDlpURL = "https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp"

// TestMain downloads yt-dlp into a temporary directory and puts it at the front
// of PATH, so every test in this package runs against a current yt-dlp (see
// ytDlpURL for why we insist on a fresh one) regardless of what is installed
// on the machine. The directory is removed once the tests finish.
//
// This uses os.MkdirTemp rather than t.TempDir because TestMain has no
// *testing.T, and a per-test directory would mean re-downloading yt-dlp for
// every test and having it deleted while later tests still need it.
func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "yotocli-test-bin-")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create temp dir for yt-dlp: %v\n", err)
		os.Exit(1)
	}

	if err := installYtDlp(dir); err != nil {
		fmt.Fprintf(os.Stderr, "failed to install yt-dlp: %v\n", err)
		os.RemoveAll(dir)
		os.Exit(1)
	}

	os.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))

	code := m.Run()

	// Not deferred: os.Exit does not run deferred functions.
	os.RemoveAll(dir)
	os.Exit(code)
}

// installYtDlp downloads the yt-dlp release build into dir and checks that it
// runs.
func installYtDlp(dir string) error {
	// The release build is a Python zipapp, so it needs python3 on PATH.
	if _, err := exec.LookPath("python3"); err != nil {
		return fmt.Errorf("python3 is required to run the yt-dlp release build: %w", err)
	}

	client := &http.Client{Timeout: 2 * time.Minute}
	resp, err := client.Get(ytDlpURL)
	if err != nil {
		return fmt.Errorf("downloading %s: %w", ytDlpURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("downloading %s: unexpected status %s", ytDlpURL, resp.Status)
	}

	path := filepath.Join(dir, "yt-dlp")
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0755)
	if err != nil {
		return err
	}
	if _, err := io.Copy(f, resp.Body); err != nil {
		f.Close()
		return fmt.Errorf("writing %s: %w", path, err)
	}
	if err := f.Close(); err != nil {
		return err
	}

	// Fail here, with a clear message, rather than inside a download test if
	// what we fetched turns out to be unusable.
	version, err := exec.Command(path, "--version").CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s --version failed: %w\n%s", path, err, version)
	}

	fmt.Printf("testing against yt-dlp %s\n", strings.TrimSpace(string(version)))
	return nil
}
