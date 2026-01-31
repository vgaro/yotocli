package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/vgaro/yotocli/internal/actions"
)

var (
	addNoNormalize bool
)

var addCmd = &cobra.Command{
	Use:   "add <playlist[/position]> <file>",
	Short: "Add a track to a playlist",
	Long: `Uploads and adds a new audio file to an existing playlist.

If a position is provided, the track is inserted there. Otherwise, it is appended to the end.`,
	Example: `  # Append a track to a playlist
  yoto add "Bedtime Stories" ./new-chapter.mp3

  # Insert a track at the beginning (position 1)
  yoto add "Bedtime/1" ./intro.mp3

  # Add without audio normalization
  yoto add "Bedtime" ./pre-processed.mp3 --no-normalize`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		playlistArg := args[0]
		filePath := args[1]

		return actions.AddTrack(apiClient, playlistArg, filePath, !addNoNormalize, func(format string, args ...interface{}) {
			fmt.Printf(format+"\n", args...)
		})
	},
}

func init() {
	addCmd.Flags().BoolVar(&addNoNormalize, "no-normalize", false, "Disable audio normalization")
	rootCmd.AddCommand(addCmd)
}
