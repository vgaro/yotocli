package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/vgaro/yotocli/internal/actions"
	"github.com/vgaro/yotocli/internal/utils"
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
	// actions.MoveTrack uses 1-based index. Destination is same card.
	return actions.MoveTrack(apiClient, card.CardID, idx+1, "", newIdx+1)
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
		
		// We need to fetch the full card just to find the index if it's a name?
		// utils.FindChapter takes full card.
		// Let's reuse utils logic or just call GetCard if needed.
		// Wait, actions.MoveTrack takes indices.
		// We need to resolve names to indices here in CLI layer.
		
		srcCard, err := apiClient.GetCard(srcCardRef.CardID)
		if err != nil {
			return err
		}
		srcIdx, _ := utils.FindChapter(srcCard, srcParts[1])
		if srcIdx == -1 {
			return fmt.Errorf("source track not found")
		}

		destParts := strings.Split(args[1], "/")
		destCardRef := utils.FindCard(cards, destParts[0])

		var destCardID string
		var destPos int

		if destCardRef == nil {
			// Destination is likely just a position in the same card
			newPos, err := utils.ParseIndex(args[1])
			if err != nil {
				return fmt.Errorf("destination card or position not found: %s", args[1])
			}
			destCardID = srcCard.CardID
			destPos = newPos // 1-based
		} else {
			destCardID = destCardRef.CardID
			destPos = -1 // append
			if len(destParts) > 1 {
				if p, err := utils.ParseIndex(destParts[1]); err == nil {
					destPos = p
				}
			}
		}

		fmt.Printf("Moving track %d from '%s' to '%s' position %d...\n", srcIdx+1, srcCard.Title, destCardID, destPos)
		return actions.MoveTrack(apiClient, srcCard.CardID, srcIdx+1, destCardID, destPos)
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

		destPos := -1 // append
		if len(destParts) > 1 {
			if p, err := utils.ParseIndex(destParts[1]); err == nil {
				destPos = p
			}
		}

		fmt.Printf("Copying track %d from '%s' to '%s' position %d...\n", srcIdx+1, srcCard.Title, destCardRef.Title, destPos)
		return actions.CopyTrack(apiClient, srcCard.CardID, srcIdx+1, destCardRef.CardID, destPos)
	},
}

func init() {
	rootCmd.AddCommand(mvupCmd)
	rootCmd.AddCommand(mvdownCmd)
	rootCmd.AddCommand(mvCmd)
	rootCmd.AddCommand(cpCmd)
}