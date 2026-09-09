package cmd

import (
	"fmt"

	"github.com/vgaro/yotocli/internal/actions"
	"github.com/spf13/cobra"
)

var (
	importPlaylist    string
	importSync        bool
	importNoNormalize bool
)

var importCmd = &cobra.Command{
	Use:   "import <url>",
	Short: "Download audio from a URL and add it to a playlist",
	Long: `Uses yt-dlp to download audio from a URL and adds it to a Yoto playlist.

Anything yt-dlp can extract works: a YouTube video, a YouTube playlist, a podcast
RSS feed, an Internet Archive item, or a direct link to an audio file. A URL that
holds several items imports all of them into the same playlist, in the order
yt-dlp reports them, and each track is named after the item it came from.

Re-importing the same URL adds a second copy of everything it holds. Pass --sync
to make the playlist match the URL instead: episodes already on the card keep the
icons they were given, new episodes are added, and tracks the URL no longer lists
are removed. Episodes are matched by their audio rather than their title, so a
renamed episode keeps its icon. yt-dlp still downloads everything the URL holds;
it is the upload to Yoto that a sync skips.`,
	Example: `  # Import a video to "Bedtime Stories"
  yoto import "https://youtu.be/dQw4w9WgXcQ" --playlist "Bedtime Stories"

  # Import to a new playlist (uses the first item's title as playlist name if not specified)
  yoto import "https://youtu.be/dQw4w9WgXcQ"

  # Import a YouTube playlist (every video lands in one playlist)
  yoto import "https://www.youtube.com/playlist?list=PLfnx6nNQO1XM" --playlist "Sing Alongs"

  # Import every episode in a podcast RSS feed
  yoto import "https://feeds.wgbh.org/2469/feed-rss.xml" --playlist "Arthur"

  # Import a public domain audiobook (one track per chapter)
  yoto import "https://archive.org/details/alices_adventures_1003" --playlist "Alice in Wonderland"

  # Pick up new episodes of a feed without duplicating the old ones
  yoto import "https://feeds.wgbh.org/2469/feed-rss.xml" --playlist "Arthur" --sync`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		url := args[0]
		return actions.ImportFromURL(apiClient, url, importPlaylist, importSync, func(format string, args ...interface{}) {
			fmt.Printf(format+"\n", args...)
		})
	},
}

func init() {
	importCmd.Flags().StringVarP(&importPlaylist, "playlist", "p", "", "Target playlist name (optional)")
	importCmd.Flags().BoolVar(&importSync, "sync", false, "Make an existing playlist match the URL instead of appending to it: keeps the icons of tracks it already has, adds new ones, removes the rest")
	importCmd.Flags().BoolVar(&importNoNormalize, "no-normalize", false, "Disable audio normalization")
	if err := importCmd.Flags().MarkDeprecated("no-normalize", noNormalizeDeprecated); err != nil {
		panic(err)
	}
	rootCmd.AddCommand(importCmd)
}
