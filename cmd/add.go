package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

		"github.com/vgaro/yotocli/internal/processing"

		"github.com/vgaro/yotocli/internal/utils"

		"github.com/vgaro/yotocli/pkg/yoto"

		"github.com/spf13/cobra"

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

	

			uploadPath := filePath

			if !addNoNormalize {

				fmt.Printf("Normalizing %s...\n", filepath.Base(filePath))

				normPath, err := processing.NormalizeAudio(filePath)

				if err != nil {

					fmt.Printf("Warning: Normalization failed: %v. Using original file.\n", err)

				} else {

					uploadPath = normPath

					defer os.Remove(normPath)

				}

			}

	

			cards, err := apiClient.ListCards()

			if err != nil {

				return err

			}

	

			parts := strings.Split(playlistArg, "/")

			cardQuery := parts[0]

			position := -1 // -1 means append

	

			if len(parts) > 1 {

				if p, err := utils.ParseIndex(parts[1]); err == nil {

					position = p - 1 // 0-based

				}

			}

	

			card := utils.FindCard(cards, cardQuery)

			if card == nil {

				return fmt.Errorf("card not found: %s", cardQuery)

			}

	

			// 1. Upload file

			fmt.Printf("Uploading %s...\n", filepath.Base(uploadPath))

			upData, err := apiClient.GetUploadURL()

			if err != nil {

				return err

			}

	

			err = apiClient.UploadFile(uploadPath, upData.Upload.UploadURL)

			if err != nil {

				return err

			}

	

			fmt.Println("Upload complete, waiting for transcoding...")

			transData, err := apiClient.PollTranscode(upData.Upload.UploadID)

			if err != nil {

				return err

			}

	

			// 2. Prepare new track/chapter

			title := strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))

	
		newTrack := yoto.Track{
			Title:        title,
			TrackURL:     fmt.Sprintf("yoto:#%s", transData.TranscodedSha256),
			Duration:     transData.TranscodedInfo.Duration,
			FileSize:     transData.TranscodedInfo.FileSize,
			Format:       transData.TranscodedInfo.Format,
			OverlayLabel: "1", // Default
			Type:         "audio",
			Display: yoto.Display{
				Icon16x16: "yoto:#aUm9i3ex3qqAMYBv-i-O-pYMKuMJGICtR3Vhf289u2Q", // Standard icon
			},
		}

		newChapter := yoto.Chapter{
			Title:    title,
			Duration: newTrack.Duration,
			Tracks:   []yoto.Track{newTrack},
			Display:  newTrack.Display,
		}

		// 3. Update playlist
		fullCard, err := apiClient.GetCard(card.CardID)
		if err != nil {
			return err
		}
		if fullCard.Content == nil {
			fullCard.Content = &yoto.Content{}
		}

		if position < 0 || position >= len(fullCard.Content.Chapters) {
			fullCard.Content.Chapters = append(fullCard.Content.Chapters, newChapter)
			fmt.Printf("Added to end of playlist: %s\n", title)
		} else {
			// Insert at position
			fullCard.Content.Chapters = append(fullCard.Content.Chapters[:position+1], fullCard.Content.Chapters[position:]...)
			fullCard.Content.Chapters[position] = newChapter
			fmt.Printf("Inserted at position %d: %s\n", position+1, title)
		}

		// Renumber keys/overlay labels
		for i := range fullCard.Content.Chapters {
			key := fmt.Sprintf("%02d", i+1)
			fullCard.Content.Chapters[i].Key = key
			fullCard.Content.Chapters[i].OverlayLabel = fmt.Sprintf("%d", i+1)
			for j := range fullCard.Content.Chapters[i].Tracks {
				fullCard.Content.Chapters[i].Tracks[j].Key = key
				fullCard.Content.Chapters[i].Tracks[j].OverlayLabel = fmt.Sprintf("%d", i+1)
			}
		}

		return apiClient.UpdateCard(card.CardID, fullCard)
	},
}

func init() {
	addCmd.Flags().BoolVar(&addNoNormalize, "no-normalize", false, "Disable audio normalization")
	rootCmd.AddCommand(addCmd)
}
