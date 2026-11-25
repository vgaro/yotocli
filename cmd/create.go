package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/vgaro/yotocli/internal/actions"
)

var (
	createName        string
	createSync        bool
	createNoNormalize bool
)

var createCmd = &cobra.Command{
	Use:   "create <directory>",
	Short: "Create a new playlist from a directory of audio files",
	Long: `Scans a directory for audio files (MP3, M4A, AAC, WAV), uploads them in parallel,
and creates a Yoto playlist. Files are sorted alphabetically by filename.

If a playlist of that name already exists the files are appended to it, or with
--sync the playlist is made to match the directory instead: files it already has
keep the icons they were given and are not sent again, new files are added, and
tracks that are no longer in the directory are removed. Files are matched by their
audio rather than their name, so a renamed file keeps its icon.`,
	Example: `  # Create a playlist from a folder
  yoto create ./audiobooks/dinosaur-expert

  # Create a playlist with a custom name
  yoto create ./audiobooks/dinosaur-expert --name "All About Dinosaurs"

  # Bring an existing playlist back in line with the folder
  yoto create ./audiobooks/dinosaur-expert --sync`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dir := args[0]
		if createName == "" {
			createName = filepath.Base(dir)
		}

		files, err := os.ReadDir(dir)
		if err != nil {
			return err
		}

		var audioFiles []string
		extensions := map[string]bool{".mp3": true, ".m4a": true, ".aac": true, ".wav": true}
		for _, f := range files {
			if !f.IsDir() && extensions[strings.ToLower(filepath.Ext(f.Name()))] {
				audioFiles = append(audioFiles, filepath.Join(dir, f.Name()))
			}
		}
		sort.Strings(audioFiles)

		if len(audioFiles) == 0 {
			return fmt.Errorf("no audio files found in %s", dir)
		}

		verb := "Creating"
		if createSync {
			verb = "Syncing"
		}
		fmt.Printf("%s playlist '%s' with %d tracks...\n", verb, createName, len(audioFiles))

		// No titles: these are files the user named themselves, so the file
		// name is the best guess we have.
		tracks := make([]actions.Track, len(audioFiles))
		for i, path := range audioFiles {
			tracks[i] = actions.Track{Path: path}
		}

		return actions.AddTracks(apiClient, createName, tracks, createSync, func(format string, args ...interface{}) {
			fmt.Printf(format+"\n", args...)
		})
	},
}

func init() {
	createCmd.Flags().StringVarP(&createName, "name", "n", "", "Name of the playlist (defaults to directory name)")
	createCmd.Flags().BoolVar(&createSync, "sync", false, "Make an existing playlist match the directory instead of appending to it: keeps the icons of tracks it already has, adds new ones, removes the rest")
	createCmd.Flags().BoolVar(&createNoNormalize, "no-normalize", false, "Disable audio normalization")
	if err := createCmd.Flags().MarkDeprecated("no-normalize", noNormalizeDeprecated); err != nil {
		panic(err)
	}
	rootCmd.AddCommand(createCmd)
}
