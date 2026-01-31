package cmd

import (
	"fmt"
	"strings"

	"github.com/vgaro/yotocli/internal/actions"
	"github.com/vgaro/yotocli/internal/utils"
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

		idx, _ := utils.FindChapter(fullCard, trackQuery)
		if idx == -1 {
			return fmt.Errorf("track not found: %s", trackQuery)
		}

		fmt.Printf("Removing track: %s\n", fullCard.Content.Chapters[idx].Title)
		return actions.RemoveTrack(apiClient, card.CardID, idx+1)
	},
}

func init() {
	rootCmd.AddCommand(rmCmd)
}