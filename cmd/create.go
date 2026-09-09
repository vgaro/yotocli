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
	createNoNormalize bool
)

var createCmd = &cobra.Command{
	Use:   "create <directory>",
	Short: "Create a new playlist from a directory of audio files",
	Long: `Scans a directory for audio files (MP3, M4A, AAC, WAV), uploads them in parallel,
and creates a Yoto playlist. Files are sorted alphabetically by filename.

If a playlist of that name already exists the files are appended to it.`,
	Example: `  # Create a playlist from a folder
  yoto create ./audiobooks/dinosaur-expert

  # Create a playlist with a custom name
  yoto create ./audiobooks/dinosaur-expert --name "All About Dinosaurs"`,
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

		fmt.Printf("Creating playlist '%s' with %d tracks...\n", createName, len(audioFiles))

		// No titles: these are files the user named themselves, so the file
		// name is the best guess we have.
		tracks := make([]actions.Track, len(audioFiles))
		for i, path := range audioFiles {
			tracks[i] = actions.Track{Path: path}
		}

		return actions.AddTracks(apiClient, createName, tracks, func(format string, args ...interface{}) {
			fmt.Printf(format+"\n", args...)
		})
	},
}

func init() {
	createCmd.Flags().StringVarP(&createName, "name", "n", "", "Name of the playlist (defaults to directory name)")
	createCmd.Flags().BoolVar(&createNoNormalize, "no-normalize", false, "Disable audio normalization")
	if err := createCmd.Flags().MarkDeprecated("no-normalize", noNormalizeDeprecated); err != nil {
		panic(err)
	}
	rootCmd.AddCommand(createCmd)
}
