package cmd

import (
	"fmt"
	"strings"

	"github.com/vgaro/yotocli/internal/utils"
	"github.com/vgaro/yotocli/pkg/yoto"
	"github.com/spf13/cobra"
)

var (
	editName        string
	editAuthor      string
	editDescription string
)

var editCmd = &cobra.Command{
	Use:   "edit <playlist[/track]>",
	Short: "Edit properties of a playlist or track",
	Long:  `Modify the metadata of a playlist or track, such as the title, author, or description.`,
	Example: `  # Rename a playlist
  yoto edit "Bedtime Stories" --name "Sleepy Time"

  # Update playlist metadata
  yoto edit "Sleepy Time" --author "Dad" --description "Read by Dad"

  # Rename a specific track
  yoto edit "Sleepy Time/1" --name "Chapter 1"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		query := args[0]
		
		if editName == "" && editAuthor == "" && editDescription == "" {
			return fmt.Errorf("no changes specified: use --name, --author, or --description")
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

		// Initialize metadata if nil
		if fullCard.Metadata == nil {
			fullCard.Metadata = &yoto.Metadata{}
		}

		if len(parts) == 1 {
			// Edit Playlist
			changed := false
			if editName != "" {
				fmt.Printf("Updating Title: '%s' -> '%s'\n", fullCard.Title, editName)
				fullCard.Title = editName
				changed = true
			}
			if editAuthor != "" {
				fmt.Printf("Updating Author: '%s' -> '%s'\n", fullCard.Metadata.Author, editAuthor)
				fullCard.Metadata.Author = editAuthor
				changed = true
			}
			if editDescription != "" {
				fmt.Printf("Updating Description\n")
				fullCard.Metadata.Description = editDescription
				changed = true
			}

			if !changed {
				fmt.Println("No changes to apply.")
				return nil
			}

			return apiClient.UpdateCard(fullCard.CardID, fullCard)
		}

		// Edit Track
		trackQuery := parts[1]
		_, chapter := utils.FindChapter(fullCard, trackQuery)
		if chapter == nil {
			return fmt.Errorf("track not found: %s", trackQuery)
		}

		if editAuthor != "" || editDescription != "" {
			fmt.Println("Warning: --author and --description are ignored for tracks.")
		}

		if editName != "" {
			fmt.Printf("Renaming track '%s' to '%s'...\n", chapter.Title, editName)
			chapter.Title = editName
			if len(chapter.Tracks) > 0 {
				chapter.Tracks[0].Title = editName
			}
			return apiClient.UpdateCard(fullCard.CardID, fullCard)
		}

		return nil
	},
}

func init() {
	editCmd.Flags().StringVarP(&editName, "name", "n", "", "New name/title")
	editCmd.Flags().StringVarP(&editAuthor, "author", "a", "", "New author (Playlist only)")
	editCmd.Flags().StringVarP(&editDescription, "description", "d", "", "New description (Playlist only)")
	rootCmd.AddCommand(editCmd)
}
