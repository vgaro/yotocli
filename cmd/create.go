package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/vgaro/yotocli/internal/processing"
	"github.com/vgaro/yotocli/internal/utils"
	"github.com/vgaro/yotocli/pkg/yoto"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

var (
	createName        string
	createNoNormalize bool
)

var createCmd = &cobra.Command{
	Use:   "create <directory>",
	Short: "Create a new playlist from a directory of audio files",
	Long: `Scans a directory for audio files (MP3, M4A, AAC, WAV), uploads them in parallel,
and creates a brand new Yoto playlist. Files are sorted alphabetically by filename.`,
	Example: `  # Create a playlist from a folder
  yoto create ./audiobooks/dinosaur-expert

  # Create a playlist with a custom name
  yoto create ./audiobooks/dinosaur-expert --name "All About Dinosaurs"

  # Create quickly without normalization
  yoto create ./my-podcasts --no-normalize`,
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

		// Parallel upload with limit
		g := new(errgroup.Group)
		g.SetLimit(5) // Limit concurrency

		tracks := make([]yoto.Track, len(audioFiles))
		var mu sync.Mutex

		for i, path := range audioFiles {
			i, path := i, path // capture for goroutine
			g.Go(func() error {
				uploadPath := path
				if !createNoNormalize {
					fmt.Printf("[%d/%d] Normalizing %s...\n", i+1, len(audioFiles), filepath.Base(path))
					normPath, err := processing.NormalizeAudio(path)
					if err != nil {
						fmt.Printf("[%d/%d] Warning: Normalization failed for %s: %v. Using original.\n", i+1, len(audioFiles), filepath.Base(path), err)
					} else {
						uploadPath = normPath
						defer os.Remove(normPath)
					}
				}

				fmt.Printf("[%d/%d] Uploading %s...\n", i+1, len(audioFiles), filepath.Base(path))
				
				upData, err := apiClient.GetUploadURL()
				if err != nil {
					return err
				}

				if err := apiClient.UploadFile(uploadPath, upData.Upload.UploadURL); err != nil {
					return err
				}

				transData, err := apiClient.PollTranscode(upData.Upload.UploadID)
				if err != nil {
					return err
				}

				title := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
				
				mu.Lock()
				tracks[i] = yoto.Track{
					Title:    title,
					TrackURL: fmt.Sprintf("yoto:#%s", transData.TranscodedSha256),
					Duration: transData.TranscodedInfo.Duration,
					FileSize: transData.TranscodedInfo.FileSize,
					Format:   transData.TranscodedInfo.Format,
					Type:     "audio",
					Display: yoto.Display{
						Icon16x16: "yoto:#aUm9i3ex3qqAMYBv-i-O-pYMKuMJGICtR3Vhf289u2Q",
					},
				}
				mu.Unlock()
				fmt.Printf("[%d/%d] Transcoded: %s\n", i+1, len(audioFiles), title)
				return nil
			})
		}

		if err := g.Wait(); err != nil {
			return err
		}

		// Assemble chapters
		chapters := make([]yoto.Chapter, len(tracks))
		var totalDur, totalSize int
		for i, t := range tracks {
			chapters[i] = yoto.Chapter{
				Title:    t.Title,
				Duration: t.Duration,
				Tracks:   []yoto.Track{t},
				Display:  t.Display,
			}
			totalDur += t.Duration
			totalSize += t.FileSize
		}

		newCard := &yoto.Card{
			Title: createName,
			Content: &yoto.Content{
				Chapters: chapters,
			},
			Metadata: &yoto.Metadata{
				Media: yoto.Media{
					Duration: totalDur,
					FileSize: totalSize,
				},
			},
		}
		utils.ReorderPlaylist(newCard)

		// Create playlist via POST /content
		// Note: pkg/yoto/client.go doesn't have CreateCard yet, adding it.
		return apiClient.CreateCard(newCard)
	},
}

func init() {
	createCmd.Flags().StringVarP(&createName, "name", "n", "", "Name of the playlist (defaults to directory name)")
	createCmd.Flags().BoolVar(&createNoNormalize, "no-normalize", false, "Disable audio normalization")
	rootCmd.AddCommand(createCmd)
}
