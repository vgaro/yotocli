package actions

import (
	"os"

	"github.com/vgaro/yotocli/internal/processing"
	"github.com/vgaro/yotocli/pkg/yoto"
)

type Logger func(string, ...interface{})

func ImportFromURL(client *yoto.Client, url string, playlistName string, normalize bool, log Logger) error {
	if log == nil {
		log = func(s string, i ...interface{}) {}
	}

	log("Downloading audio from %s...", url)
	filePath, title, err := processing.DownloadFromURL(url)
	if err != nil {
		return err
	}
	defer os.Remove(filePath)

	log("Downloaded: %s", title)

	// If no playlist specified, use title
	targetPlaylist := playlistName
	if targetPlaylist == "" {
		targetPlaylist = title
	}

	// AddTrack handles normalization, finding/creating playlist, upload, and update
	return AddTrack(client, targetPlaylist, filePath, "", normalize, log)
}
