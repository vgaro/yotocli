package actions

import (
	"fmt"

	"github.com/vgaro/yotocli/internal/processing"
	"github.com/vgaro/yotocli/pkg/yoto"
)

type Logger func(string, ...interface{})

func ImportFromURL(client *yoto.Client, url string, playlistName string, normalize bool, log Logger) error {
	if log == nil {
		log = func(s string, i ...interface{}) {}
	}

	log("Downloading audio from %s...", url)
	downloads, cleanup, err := processing.DownloadFromURL(url)
	defer func() {
		if err := cleanup(); err != nil {
			log("Warning: failed to remove downloaded files: %v", err)
		}
	}()
	if err != nil {
		return err
	}

	log("Downloaded %d track(s)", len(downloads))

	// If no playlist specified, use the title of the first download
	targetPlaylist := playlistName
	if targetPlaylist == "" {
		targetPlaylist = downloads[0].Name
	}

	// AddTrack handles normalization, finding/creating playlist, upload, and
	// update. The first call creates the playlist if it does not exist, the
	// rest append to it.
	// d.Name is passed as the track title: the downloaded file is named after
	// the yt-dlp ID, so leaving the title to be derived from it would put
	// things like "Ixrje2rXLMA" on the card.
	for _, d := range downloads {
		log("Adding %s...", d.Name)
		if err := AddTrack(client, targetPlaylist, d.Path, d.Name, "", normalize, log); err != nil {
			return fmt.Errorf("failed to add %q: %w", d.Name, err)
		}
	}

	return nil
}
