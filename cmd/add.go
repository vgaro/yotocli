package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/vgaro/yotocli/internal/actions"
)

var (
	addNoNormalize bool
	addIcon        string
	addTitle       string
	addTrimStart   float64
	addTrimEnd     float64
)

var addCmd = &cobra.Command{
	Use:   "add <playlist[/position]> <file>",
	Short: "Add a track to a playlist",
	Long: `Uploads and adds a new audio file to an existing playlist.

If a position is provided, the track is inserted there. Otherwise, it is appended to the end.`,
	Example: `  # Append a track to a playlist
  yoto add "Bedtime Stories" ./new-chapter.mp3 --title "Chapter 1"

  # Insert a track at the beginning (position 1)
  yoto add "Bedtime/1" ./intro.mp3

  # Add with trimming
  yoto add "Bedtime" ./story.mp3 --trim-start 21 --trim-end 30`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		playlistArg := args[0]
		filePath := args[1]

		return actions.AddTrack(apiClient, playlistArg, filePath, addTitle, addIcon, !addNoNormalize, addTrimStart, addTrimEnd, func(format string, args ...interface{}) {
			fmt.Printf(format+"\n", args...)
		})
	},
}

func init() {
	addCmd.Flags().BoolVar(&addNoNormalize, "no-normalize", false, "Disable audio normalization")
	addCmd.Flags().StringVar(&addIcon, "icon", "", "Icon ID (hash or yoto:#...) to use for the track")
	addCmd.Flags().StringVarP(&addTitle, "title", "t", "", "Title for the new track (defaults to filename)")
	addCmd.Flags().Float64Var(&addTrimStart, "trim-start", 0, "Trim N seconds from the start")
	addCmd.Flags().Float64Var(&addTrimEnd, "trim-end", 0, "Trim N seconds from the end")
	rootCmd.AddCommand(addCmd)
}

