package processing

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
)

type FFProbeResponse struct {
	Streams []struct {
		Channels int `json:"channels"`
	} `json:"streams"`
	Format struct {
		Duration string `json:"duration"`
	} `json:"format"`
}

func GetAudioInfo(path string) (int, float64, error) {
	cmd := exec.Command("ffprobe",
		"-v", "quiet",
		"-print_format", "json",
		"-show_streams",
		"-show_format",
		"-select_streams", "a",
		path,
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return 2, 0, err
	}

	var resp FFProbeResponse
	if err := json.Unmarshal(output, &resp); err != nil {
		return 2, 0, err
	}

	channels := 2
	if len(resp.Streams) > 0 {
		channels = resp.Streams[0].Channels
	}

	var duration float64
	fmt.Sscanf(resp.Format.Duration, "%f", &duration)

	return channels, duration, nil
}

func ProcessAudio(inputPath string, normalize bool, trimStart, trimEnd float64) (string, error) {
	channels, duration, err := GetAudioInfo(inputPath)
	if err != nil {
		channels = 2
	}

	tempFile, err := os.CreateTemp("", "yoto_proc_*.mp3")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	tempPath := tempFile.Name()
	tempFile.Close()

	args := []string{"-y"}

	// Fast seek for trim start if it's large, but for precision we might want it after -i
	// Given we are usually dealing with short durations, after -i is safer for exact trimming
	args = append(args, "-i", inputPath)

	var filters []string

	if trimStart > 0 {
		args = append(args, "-ss", fmt.Sprintf("%f", trimStart))
	}

	if trimEnd > 0 && duration > 0 {
		endTime := duration - trimEnd
		if endTime <= trimStart {
			os.Remove(tempPath)
			return "", fmt.Errorf("trimming would result in negative or zero duration (start: %f, end: %f, total: %f)", trimStart, trimEnd, duration)
		}
		args = append(args, "-to", fmt.Sprintf("%f", endTime))
	}

	if normalize {
		targetLUFS := -16
		if channels == 1 {
			targetLUFS = -18
		}
		filters = append(filters, fmt.Sprintf("loudnorm=I=%d:TP=-1.5:LRA=11", targetLUFS))
	}

	if len(filters) > 0 {
		args = append(args, "-filter:a", filters[0]) // Only loudnorm for now
	}

	args = append(args, "-c:a", "libmp3lame", "-q:a", "2", tempPath)

	cmd := exec.Command("ffmpeg", args...)

	if output, err := cmd.CombinedOutput(); err != nil {
		os.Remove(tempPath)
		return "", fmt.Errorf("ffmpeg error: %w (output: %s)", err, string(output))
	}

	return tempPath, nil
}

func NormalizeAudio(inputPath string) (string, error) {
	return ProcessAudio(inputPath, true, 0, 0)
}

