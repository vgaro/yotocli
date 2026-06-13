package actions

import (
        "os"

        "github.com/vgaro/yotocli/internal/processing"
        "github.com/vgaro/yotocli/pkg/yoto"
)

type Logger func(string, ...interface{})

func ImportFromURL(client *yoto.Client, url string, playlistName string, normalize bool, trimStart, trimEnd float64, log Logger) error {
	if log == nil {
		log = func(s string, i ...interface{}) {}
	}

	log("Downloading audio from %s...", url)
	items, err := processing.DownloadFromURL(url)
	if err != nil {
		return err
	}

	log("Downloaded %d items.", len(items))

	for i, item := range items {
		log("[%d/%d] Processing: %s", i+1, len(items), item.Title)

		// If no playlist specified, use the title of the first track as playlist name
		targetPlaylist := playlistName
		if targetPlaylist == "" {
			targetPlaylist = items[0].Title
		}

		// AddTrack handles normalization, trimming, finding/creating playlist, upload, and update
		err = AddTrack(client, targetPlaylist, item.Path, item.Title, "", normalize, trimStart, trimEnd, log)
		os.Remove(item.Path) // Clean up downloaded file immediately
		if err != nil {
			log("Error adding track '%s': %v", item.Title, err)
			return err
		}
	}

	return nil
}
