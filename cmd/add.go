package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/vgaro/yotocli/internal/actions"
)

var (
	addNoNormalize bool
	addIcon        string
)

// noNormalizeDeprecated explains --no-normalize to anyone still passing it. The
// flag used to skip a local ffmpeg pass; Yoto's transcoder normalizes every
// upload on its own, so there is no local pass left to skip.
const noNormalizeDeprecated = "audio is normalized by Yoto during transcoding, so this flag does nothing"

var addCmd = &cobra.Command{
	Use:   "add <playlist[/position]> <file>",
	Short: "Add a track to a playlist",
	Long: `Uploads and adds a new audio file to an existing playlist.

If a position is provided, the track is inserted there. Otherwise, it is appended to the end.`,
	Example: `  # Append a track to a playlist
  yoto add "Bedtime Stories" ./new-chapter.mp3

  # Insert a track at the beginning (position 1)
  yoto add "Bedtime/1" ./intro.mp3`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		playlistArg := args[0]
		filePath := args[1]

		// No title: for a file the user picked, the file name is the best
		// guess we have.
		track := actions.Track{Path: filePath, IconID: addIcon}
		return actions.AddTracks(apiClient, playlistArg, []actions.Track{track}, func(format string, args ...interface{}) {
			fmt.Printf(format+"\n", args...)
		})
	},
}

func init() {
	addCmd.Flags().BoolVar(&addNoNormalize, "no-normalize", false, "Disable audio normalization")
	if err := addCmd.Flags().MarkDeprecated("no-normalize", noNormalizeDeprecated); err != nil {
		panic(err)
	}
	addCmd.Flags().StringVar(&addIcon, "icon", "", "Icon ID (hash or yoto:#...) to use for the track")
	rootCmd.AddCommand(addCmd)
}
