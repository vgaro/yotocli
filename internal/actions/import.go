package actions

import (
	"github.com/vgaro/yotocli/internal/processing"
	"github.com/vgaro/yotocli/pkg/yoto"
)

type Logger func(string, ...interface{})

// ImportFromURL downloads everything a URL holds and adds it to one playlist.
//
// With syncPlaylist set the playlist is made to match the URL rather than added
// to, which is how a feed that has gained an episode is re-imported without
// ending up with two copies of every old one. See AddTracks for what that keeps
// and what it removes.
func ImportFromURL(client *yoto.Client, url string, playlistName string, syncPlaylist bool, log Logger) error {
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

	// d.Name is passed as the track title: the downloaded file is named after
	// the yt-dlp ID, so leaving the title to be derived from it would put
	// things like "Ixrje2rXLMA" on the card.
	tracks := make([]Track, len(downloads))
	for i, d := range downloads {
		tracks[i] = Track{Path: d.Path, Title: d.Name}
	}

	return AddTracks(client, targetPlaylist, tracks, syncPlaylist, log)
}
