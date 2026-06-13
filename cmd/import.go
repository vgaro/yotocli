package cmd

import (
	"fmt"

	"github.com/vgaro/yotocli/internal/actions"
	"github.com/spf13/cobra"
)

var (
	importPlaylist    string
	importNoNormalize bool
	importTrimStart   float64
	importTrimEnd     float64
)

var importCmd = &cobra.Command{
	Use:   "import <url>",
	Short: "Download audio from a URL and add it to a playlist",
	Long: `Uses yt-dlp to download audio from YouTube (or other supported sites),
normalizes the volume, and adds it to a Yoto playlist. Supports playlists and trimming.`,
	Example: `  # Import a video to "Bedtime Stories"
  yoto import "https://youtu.be/dQw4w9WgXcQ" --playlist "Bedtime Stories"

  # Import with trimming (first 10s and last 5s)
  yoto import "https://youtu.be/..." --trim-start 10 --trim-end 5`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		url := args[0]
		return actions.ImportFromURL(apiClient, url, importPlaylist, !importNoNormalize, importTrimStart, importTrimEnd, func(format string, args ...interface{}) {
			fmt.Printf(format+"\n", args...)
		})
	},
}

func init() {
	importCmd.Flags().StringVarP(&importPlaylist, "playlist", "p", "", "Target playlist name (optional)")
	importCmd.Flags().BoolVar(&importNoNormalize, "no-normalize", false, "Disable audio normalization")
	importCmd.Flags().Float64Var(&importTrimStart, "trim-start", 0, "Trim N seconds from the start")
	importCmd.Flags().Float64Var(&importTrimEnd, "trim-end", 0, "Trim N seconds from the end")
	rootCmd.AddCommand(importCmd)
}

