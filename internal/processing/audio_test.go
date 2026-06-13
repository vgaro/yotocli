package processing

import (
	"os/exec"
	"testing"
)

func TestFFmpegExists(t *testing.T) {
	_, err := exec.LookPath("ffmpeg")
	if err != nil {
		t.Skip("ffmpeg not found, skipping audio processing tests")
	}
}

func TestGetAudioInfo_InvalidFile(t *testing.T) {
	_, _, err := GetAudioInfo("non-existent-file.mp3")
	if err == nil {
		t.Error("Expected error for non-existent file, got nil")
	}
}

