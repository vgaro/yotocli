package cmd

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/vgaro/yotocli/internal/utils"
	"github.com/vgaro/yotocli/pkg/yoto"
	"github.com/spf13/cobra"
)

var lsCmd = &cobra.Command{
	Use:   "ls [playlist[/track]]",
	Short: "List playlists or tracks",
	Long: `List all playlists in your library, or list tracks within a specific playlist.
Supports slash syntax for deep listing.
Examples:
  yoto ls
  yoto ls "Bedtime Stories"
  yoto ls "Bedtime/1"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cards, err := apiClient.ListCards()
		if err != nil {
			return err
		}

		if len(args) == 0 {
			printCards(cards)
			return nil
		}

		// Handle slash syntax: "Playlist/Track"
		parts := strings.Split(args[0], "/")
		cardQuery := parts[0]
		
		card := utils.FindCard(cards, cardQuery)
		if card == nil {
			return fmt.Errorf("card not found: %s", cardQuery)
		}

		// If it's a basic card from ListCards, it might not have chapters. 
		// Fetch full detail.
		fullCard, err := apiClient.GetCard(card.CardID)
		if err != nil {
			return err
		}

		if len(parts) == 1 {
			printChapters(fullCard)
			return nil
		}

		// List specific track
		trackQuery := parts[1]
		printTrack(fullCard, trackQuery)
		return nil
	},
}

func printCards(cards []yoto.Card) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "#\tTitle\tID\tDuration")
	for i, card := range cards {
		duration := "0:00"
		if card.Metadata != nil {
			d := card.Metadata.Media.Duration
			duration = fmt.Sprintf("%d:%02d", d/60, d%60)
		}
		fmt.Fprintf(w, "%d\t%s\t%s\t%s\n", i+1, card.Title, card.CardID, duration)
	}
	w.Flush()
}

func printChapters(card *yoto.Card) {
	fmt.Printf("Playlist: %s (%s)\n\n", card.Title, card.CardID)
	
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "#\tTitle\tDuration\tFormat")

	if card.Content == nil {
		fmt.Println("No content found in this card.")
		return
	}

	for i, chapter := range card.Content.Chapters {
		duration := fmt.Sprintf("%d:%02d", chapter.Duration/60, chapter.Duration%60)
		format := "-"
		if len(chapter.Tracks) > 0 {
			format = chapter.Tracks[0].Format
		}
		fmt.Fprintf(w, "%d\t%s\t%s\t%s\n", i+1, chapter.Title, duration, format)
	}
	w.Flush()
}

func printTrack(card *yoto.Card, query string) {
	if card.Content == nil {
		fmt.Println("No content found.")
		return
	}

	var foundChapter *yoto.Chapter
	// Try Index
	if idx, err := utils.ParseIndex(query); err == nil {
		if idx > 0 && idx <= len(card.Content.Chapters) {
			foundChapter = &card.Content.Chapters[idx-1]
		}
	}

	// Try Title
	if foundChapter == nil {
		queryLower := strings.ToLower(query)
		for _, chapter := range card.Content.Chapters {
			if strings.Contains(strings.ToLower(chapter.Title), queryLower) {
				foundChapter = &chapter
				break
			}
		}
	}

	if foundChapter == nil {
		fmt.Printf("Track not found: %s\n", query)
		return
	}

	fmt.Printf("Track Detail:\n")
	fmt.Printf("  Title:    %s\n", foundChapter.Title)
	fmt.Printf("  Duration: %d:%02d\n", foundChapter.Duration/60, foundChapter.Duration%60)
	if len(foundChapter.Tracks) > 0 {
		t := foundChapter.Tracks[0]
		fmt.Printf("  Format:   %s\n", t.Format)
		fmt.Printf("  Size:     %.2f MB\n", float64(t.FileSize)/1024/1024)
		fmt.Printf("  URL:      %s\n", t.TrackURL)
	}
}

func init() {
	rootCmd.AddCommand(lsCmd)
}

