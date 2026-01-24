package processing

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type FFProbeResponse struct {
	Streams []struct {
		Channels int `json:"channels"`
	} `json:"streams"`
}

func GetChannelCount(path string) (int, error) {
	cmd := exec.Command("ffprobe",
		"-v", "quiet",
		"-print_format", "json",
		"-show_streams",
		"-select_streams", "a",
		path,
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return 2, err
	}

	var resp FFProbeResponse
	if err := json.Unmarshal(output, &resp); err != nil {
		return 2, err
	}

	if len(resp.Streams) == 0 {
		return 2, nil
	}
	return resp.Streams[0].Channels, nil
}

func NormalizeAudio(inputPath string) (string, error) {
	channels, err := GetChannelCount(inputPath)
	if err != nil {
		// Log and continue with default
		channels = 2
	}

	targetLUFS := -16
	if channels == 1 {
		targetLUFS = -18
	}

	tempFile := filepath.Join(os.TempDir(), fmt.Sprintf("yoto_norm_%d.mp3", os.Getpid()))
	
	cmd := exec.Command("ffmpeg",
		"-y",
		"-i", inputPath,
		"-filter:a", fmt.Sprintf("loudnorm=I=%d:TP=-1.5:LRA=11", targetLUFS),
		"-c:a", "libmp3lame",
		"-q:a", "2",
		tempFile,
	)

	if output, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("ffmpeg error: %w (output: %s)", err, string(output))
	}

	return tempFile, nil
}
