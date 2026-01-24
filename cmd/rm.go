package cmd

import (
	"fmt"
	"strings"

	"github.com/vgaro/yotocli/internal/utils"
	"github.com/vgaro/yotocli/pkg/yoto"
	"github.com/spf13/cobra"
)

var rmCmd = &cobra.Command{
	Use:   "rm <playlist[/track]>",
	Short: "Remove a playlist or a track from a playlist",
	Long: `Permanently removes an entire playlist or a specific track from a playlist.
Fuzzy matching is supported for playlist and track names.`,
	Example: `  # Remove an entire playlist
  yoto rm "Bedtime Stories"

  # Remove the 2nd track from a playlist
  yoto rm "Bedtime/2"

  # Remove a track by name
  yoto rm "Bedtime/Intro"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cards, err := apiClient.ListCards()
		if err != nil {
			return err
		}

		parts := strings.Split(args[0], "/")
		cardQuery := parts[0]

		card := utils.FindCard(cards, cardQuery)
		if card == nil {
			return fmt.Errorf("card not found: %s", cardQuery)
		}

		if len(parts) == 1 {
			// Remove entire playlist
			fmt.Printf("Removing playlist: %s (%s)...\n", card.Title, card.CardID)
			return apiClient.DeleteCard(card.CardID)
		}

		// Remove specific track
		trackQuery := parts[1]
		fullCard, err := apiClient.GetCard(card.CardID)
		if err != nil {
			return err
		}

		if fullCard.Content == nil {
			return fmt.Errorf("card has no content")
		}

		var newChapters []yoto.Chapter
		var found bool
		
		// Try Index
		idx, err := utils.ParseIndex(trackQuery)
		
		for i, chapter := range fullCard.Content.Chapters {
			match := false
			if err == nil && i+1 == idx {
				match = true
			} else if strings.Contains(strings.ToLower(chapter.Title), strings.ToLower(trackQuery)) {
				match = true
			}

			if match {
				fmt.Printf("Removing track: %s\n", chapter.Title)
				found = true
				continue
			}
			newChapters = append(newChapters, chapter)
		}

		if !found {
			return fmt.Errorf("track not found: %s", trackQuery)
		}

		fullCard.Content.Chapters = newChapters
		// Recalculate duration/size if needed (API might do it automatically, but let's be safe)
		var totalDur, totalSize int
		for _, c := range newChapters {
			totalDur += c.Duration
			if len(c.Tracks) > 0 {
				totalSize += c.Tracks[0].FileSize
			}
		}
		if fullCard.Metadata != nil {
			fullCard.Metadata.Media.Duration = totalDur
			fullCard.Metadata.Media.FileSize = totalSize
		}

		return apiClient.UpdateCard(card.CardID, fullCard)
	},
}

func init() {
	rootCmd.AddCommand(rmCmd)
}