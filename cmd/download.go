package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/vgaro/yotocli/internal/utils"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

var downloadCmd = &cobra.Command{
	Use:   "download <playlist[/track]> [destination]",
	Short: "Download tracks from your library",
	Long: `Download a single track or an entire playlist to your local machine.
If downloading a playlist, a directory will be created (unless specified).
If downloading a track, it saves as an MP3 file.`,
	Example: `  # Download entire playlist to current directory (creates folder "Bedtime Stories")
  yoto download "Bedtime Stories"

  # Download playlist to specific folder
  yoto download "Bedtime Stories" ./my-backup/

  # Download single track
  yoto download "Bedtime Stories/1"

  # Download single track to specific file
  yoto download "Bedtime Stories/1" ./intro.mp3`,
	Args: cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		query := args[0]
		dest := ""
		if len(args) > 1 {
			dest = args[1]
		}

	
cards, err := apiClient.ListCards()
		if err != nil {
			return err
		}

		parts := strings.Split(query, "/")
		cardQuery := parts[0]

		card := utils.FindCard(cards, cardQuery)
		if card == nil {
			return fmt.Errorf("card not found: %s", cardQuery)
		}

		fullCard, err := apiClient.GetCard(card.CardID)
		if err != nil {
			return err
		}

		if len(parts) > 1 {
			// Download single track
			trackQuery := parts[1]
			_, chapter := utils.FindChapter(fullCard, trackQuery)
			if chapter == nil {
				return fmt.Errorf("track not found: %s", trackQuery)
			}

			if len(chapter.Tracks) == 0 {
				return fmt.Errorf("chapter has no audio tracks")
			}

			track := chapter.Tracks[0]
			if dest == "" {
				dest = fmt.Sprintf("%s.mp3", utils.SanitizeFilename(track.Title))
			} else if utils.IsDir(dest) {
				dest = filepath.Join(dest, fmt.Sprintf("%s.mp3", utils.SanitizeFilename(track.Title)))
			}

			fmt.Printf("Downloading '%s' to '%s'...\n", track.Title, dest)
			return apiClient.DownloadFile(track.TrackURL, dest)
		}

		// Download entire playlist
		if dest == "" {
			dest = utils.SanitizeFilename(fullCard.Title)
		}

		if err := os.MkdirAll(dest, 0755); err != nil {
			return err
		}

		fmt.Printf("Downloading playlist '%s' to '%s'...\n", fullCard.Title, dest)

		if fullCard.Content == nil || len(fullCard.Content.Chapters) == 0 {
			fmt.Println("Playlist is empty.")
			return nil
		}

		g := new(errgroup.Group)
		g.SetLimit(5) // Parallel downloads

		for i, chapter := range fullCard.Content.Chapters {
			if len(chapter.Tracks) == 0 {
				continue
			}
			track := chapter.Tracks[0]
			trackNum := i + 1
			
			// Capture variables for goroutine
			t := track
			n := trackNum

			g.Go(func() error {
				filename := fmt.Sprintf("%02d - %s.mp3", n, utils.SanitizeFilename(t.Title))
				path := filepath.Join(dest, filename)
				
				fmt.Printf("[%d/%d] Downloading %s...\n", n, len(fullCard.Content.Chapters), t.Title)
				if err := apiClient.DownloadFile(t.TrackURL, path); err != nil {
					return fmt.Errorf("failed to download %s: %w", t.Title, err)
				}
				return nil
			})
		}

		return g.Wait()
	},
}

func init() {
	rootCmd.AddCommand(downloadCmd)
}
