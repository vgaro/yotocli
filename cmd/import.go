package cmd

import (
	"fmt"

	"github.com/vgaro/yotocli/internal/actions"
	"github.com/spf13/cobra"
)

var (
	importPlaylist    string
	importNoNormalize bool
)

var importCmd = &cobra.Command{
	Use:   "import <url>",
	Short: "Download audio from a URL and add it to a playlist",
	Long: `Uses yt-dlp to download audio from YouTube (or other supported sites),
normalizes the volume, and adds it to a Yoto playlist.`, 
	Example: `  # Import a video to "Bedtime Stories"
  yoto import "https://youtu.be/dQw4w9WgXcQ" --playlist "Bedtime Stories"

  # Import to a new playlist (uses video title as playlist name if not specified)
  yoto import "https://youtu.be/dQw4w9WgXcQ"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		url := args[0]
		return actions.ImportFromURL(apiClient, url, importPlaylist, !importNoNormalize, func(format string, args ...interface{}) {
			fmt.Printf(format+"\n", args...)
		})
	},
}

func init() {
	importCmd.Flags().StringVarP(&importPlaylist, "playlist", "p", "", "Target playlist name (optional)")
	importCmd.Flags().BoolVar(&importNoNormalize, "no-normalize", false, "Disable audio normalization")
	rootCmd.AddCommand(importCmd)
}
