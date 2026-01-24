package processing

import (
	"os/exec"
	"testing"
)

func TestYtDlpExists(t *testing.T) {
	_, err := exec.LookPath("yt-dlp")
	if err != nil {
		t.Skip("yt-dlp not found, skipping import tests")
	}
}
