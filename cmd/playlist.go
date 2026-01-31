package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/vgaro/yotocli/internal/actions"
	"github.com/vgaro/yotocli/internal/utils"
	"github.com/vgaro/yotocli/pkg/yoto"
)

// mvupCmd represents the mvup command
var mvupCmd = &cobra.Command{
	Use:     "mvup <playlist/track>",
	Short:   "Move a track up in the playlist",
	Example: `  yoto mvup "Bedtime Stories/2"`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return moveRelative(args[0], -1)
	},
}

// mvdownCmd represents the mvdown command
var mvdownCmd = &cobra.Command{
	Use:     "mvdown <playlist/track>",
	Short:   "Move a track down in the playlist",
	Example: `  yoto mvdown "Bedtime Stories/1"`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return moveRelative(args[0], 1)
	},
}

func moveRelative(query string, delta int) error {
	parts := strings.Split(query, "/")
	if len(parts) < 2 {
		return fmt.Errorf("usage: playlist/track")
	}

	cards, _ := apiClient.ListCards()
	card := utils.FindCard(cards, parts[0])
	if card == nil {
		return fmt.Errorf("card not found")
	}

	fullCard, _ := apiClient.GetCard(card.CardID)
	idx, _ := utils.FindChapter(fullCard, parts[1])
	if idx == -1 {
		return fmt.Errorf("track not found")
	}

	newIdx := idx + delta
	// actions.MoveTrack uses 1-based index
	return actions.MoveTrack(apiClient, card.CardID, idx+1, newIdx+1)
}

// mvCmd represents the mv command
var mvCmd = &cobra.Command{
	Use:   "mv <src_playlist/track> <dest_playlist[/position]>",
	Short: "Move a track within or between playlists",
	Example: `  # Move track 1 to position 5 in same playlist
  yoto mv "Bedtime/1" 5

  # Move track from one playlist to another
  yoto mv "Bedtime/1" "Favorites/"`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		srcParts := strings.Split(args[0], "/")
		if len(srcParts) < 2 {
			return fmt.Errorf("source must be playlist/track")
		}

		cards, _ := apiClient.ListCards()
		srcCardRef := utils.FindCard(cards, srcParts[0])
		if srcCardRef == nil {
			return fmt.Errorf("source card not found")
		}
		srcCard, _ := apiClient.GetCard(srcCardRef.CardID)
		srcIdx, _ := utils.FindChapter(srcCard, srcParts[1])
		if srcIdx == -1 {
			return fmt.Errorf("source track not found")
		}

		destParts := strings.Split(args[1], "/")
		destCardRef := utils.FindCard(cards, destParts[0])

		// If destination card is not found, maybe args[1] is just a position in same card
		if destCardRef == nil {
			// Try move within same card
			newPos, err := utils.ParseIndex(args[1])
			if err != nil {
				return fmt.Errorf("destination card or position not found: %s", args[1])
			}
			newIdx := newPos - 1
			if newIdx < 0 || newIdx >= len(srcCard.Content.Chapters) {
				return fmt.Errorf("invalid position")
			}

			// Perform move within same card
			elem := srcCard.Content.Chapters[srcIdx]
			srcCard.Content.Chapters = append(srcCard.Content.Chapters[:srcIdx], srcCard.Content.Chapters[srcIdx+1:]...)
			// Insert
			srcCard.Content.Chapters = append(srcCard.Content.Chapters[:newIdx], append([]yoto.Chapter{elem}, srcCard.Content.Chapters[newIdx:]...)...)

			utils.ReorderPlaylist(srcCard)
			return apiClient.UpdateCard(srcCard.CardID, srcCard)
		}

		// Move between cards
		destCard, _ := apiClient.GetCard(destCardRef.CardID)
		destPos := len(destCard.Content.Chapters) // default append
		if len(destParts) > 1 {
			if p, err := utils.ParseIndex(destParts[1]); err == nil {
				destPos = p - 1
			}
		}

		// Remove from src
		elem := srcCard.Content.Chapters[srcIdx]
		srcCard.Content.Chapters = append(srcCard.Content.Chapters[:srcIdx], srcCard.Content.Chapters[srcIdx+1:]...)

		// Add to dest
		if destCard.Content == nil {
			destCard.Content = &yoto.Content{}
		}
		if destPos >= len(destCard.Content.Chapters) {
			destCard.Content.Chapters = append(destCard.Content.Chapters, elem)
		} else {
			destCard.Content.Chapters = append(destCard.Content.Chapters[:destPos], append([]yoto.Chapter{elem}, destCard.Content.Chapters[destPos:]...)...)
		}

		utils.ReorderPlaylist(srcCard)
		utils.ReorderPlaylist(destCard)

		// Update both (In Go we should probably do this concurrently or handle failure)
		if err := apiClient.UpdateCard(srcCard.CardID, srcCard); err != nil {
			return err
		}
		return apiClient.UpdateCard(destCard.CardID, destCard)
	},
}

// cpCmd represents the cp command
var cpCmd = &cobra.Command{
	Use:   "cp <src_playlist/track> <dest_playlist[/position]>",
	Short: "Copy a track between playlists",
	Example: `  yoto cp "Bedtime/1" "Lullabies/"`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		srcParts := strings.Split(args[0], "/")
		if len(srcParts) < 2 {
			return fmt.Errorf("source must be playlist/track")
		}

		cards, _ := apiClient.ListCards()
		srcCardRef := utils.FindCard(cards, srcParts[0])
		if srcCardRef == nil {
			return fmt.Errorf("source card not found")
		}
		srcCard, _ := apiClient.GetCard(srcCardRef.CardID)
		srcIdx, _ := utils.FindChapter(srcCard, srcParts[1])
		if srcIdx == -1 {
			return fmt.Errorf("source track not found")
		}

		destParts := strings.Split(args[1], "/")
		destCardRef := utils.FindCard(cards, destParts[0])
		if destCardRef == nil {
			return fmt.Errorf("destination card not found: %s", destParts[0])
		}

		destCard, _ := apiClient.GetCard(destCardRef.CardID)
		destPos := len(destCard.Content.Chapters) // default append
		if len(destParts) > 1 {
			if p, err := utils.ParseIndex(destParts[1]); err == nil {
				destPos = p - 1
			}
		}

		// Copy elem
		elem := srcCard.Content.Chapters[srcIdx]

		// Add to dest
		if destCard.Content == nil {
			destCard.Content = &yoto.Content{}
		}
		if destPos >= len(destCard.Content.Chapters) {
			destCard.Content.Chapters = append(destCard.Content.Chapters, elem)
		} else {
			destCard.Content.Chapters = append(destCard.Content.Chapters[:destPos], append([]yoto.Chapter{elem}, destCard.Content.Chapters[destPos:]...)...)
		}

		utils.ReorderPlaylist(destCard)
		return apiClient.UpdateCard(destCard.CardID, destCard)
	},
}

func init() {
	rootCmd.AddCommand(mvupCmd)
	rootCmd.AddCommand(mvdownCmd)
	rootCmd.AddCommand(mvCmd)
	rootCmd.AddCommand(cpCmd)
}