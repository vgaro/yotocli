package cmd

import (
	"fmt"
	"os"

	"github.com/vgaro/yotocli/internal/processing"
	"github.com/vgaro/yotocli/internal/utils"
	"github.com/vgaro/yotocli/pkg/yoto"
	"github.com/spf13/cobra"
)

var (
	importPlaylist    string
	importNoNormalize bool
)

var importCmd = &cobra.Command{
	Use:   "import <url>",
	Short: "Download audio from a URL and add it to a playlist",
	Long: `Uses yt-dlp to download audio from YouTube (or other supported sites),
normalizes the volume, and adds it to a Yoto playlist.`, 
	Example: `  # Import a video to "Bedtime Stories"
  yoto import "https://youtu.be/dQw4w9WgXcQ" --playlist "Bedtime Stories"

  # Import to a new playlist (uses video title as playlist name if not specified)
  yoto import "https://youtu.be/dQw4w9WgXcQ"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		url := args[0]

		fmt.Println("Downloading audio (this may take a moment)...")
		filePath, title, err := processing.DownloadFromURL(url)
		if err != nil {
			return err
		}
		// Cleanup raw download on exit
		defer os.Remove(filePath)

		fmt.Printf("Downloaded: %s\n", title)

		// Normalize
		uploadPath := filePath
		if !importNoNormalize {
			fmt.Println("Normalizing audio...")
			normPath, err := processing.NormalizeAudio(filePath)
			if err != nil {
				fmt.Printf("Warning: Normalization failed: %v. Using original.\n", err)
			} else {
				uploadPath = normPath
				defer os.Remove(normPath)
			}
		}

		// List cards to find target
	
cards, err := apiClient.ListCards()
		if err != nil {
			return err
		}

		// Determine target playlist
		var targetCard *yoto.Card
		
		if importPlaylist != "" {
			// Find existing
			targetCard = utils.FindCard(cards, importPlaylist)
			if targetCard == nil {
				// Ask to create? For CLI, maybe just create or error. 
				// Let's create if it doesn't exist, matching the "One-Shot" philosophy.
				fmt.Printf("Playlist '%s' not found. Creating it...\n", importPlaylist)
				targetCard = &yoto.Card{
					Title: importPlaylist,
					Content: &yoto.Content{},
				}
				// We need to Create it first to get an ID? 
				// Actually, we can construct the whole object and CreateCard.
				// But CreateCard logic in client.go expects a fresh card object.
				// Let's just treat it as a new card object that we will Create or Update.
			} else {
				// Fetch full details for update
				fullCard, err := apiClient.GetCard(targetCard.CardID)
				if err != nil {
					return err
				}
				targetCard = fullCard
			}
		} else {
			// No playlist specified, create one named after the video
			// Check if one exists first?
			fmt.Printf("No playlist specified. Creating new playlist: '%s'\n", title)
			targetCard = &yoto.Card{
				Title: title, // Use video title as playlist title
				Content: &yoto.Content{},
			}
		}

		// Upload
		fmt.Println("Uploading to Yoto...")
		upData, err := apiClient.GetUploadURL()
		if err != nil {
			return err
		}

		if err := apiClient.UploadFile(uploadPath, upData.Upload.UploadURL); err != nil {
			return err
		}

		fmt.Println("Waiting for transcoding...")
		transData, err := apiClient.PollTranscode(upData.Upload.UploadID)
		if err != nil {
			return err
		}

		// Create Track
		newTrack := yoto.Track{
			Title:        title,
			TrackURL:     fmt.Sprintf("yoto:#%s", transData.TranscodedSha256),
			Duration:     transData.TranscodedInfo.Duration,
			FileSize:     transData.TranscodedInfo.FileSize,
			Format:       transData.TranscodedInfo.Format,
			Type:         "audio",
			Display: yoto.Display{
				Icon16x16: "yoto:#aUm9i3ex3qqAMYBv-i-O-pYMKuMJGICtR3Vhf289u2Q",
			},
		}

		newChapter := yoto.Chapter{
			Title:    title,
			Duration: newTrack.Duration,
			Tracks:   []yoto.Track{newTrack},
			Display:  newTrack.Display,
		}

		// Add to content
		if targetCard.Content == nil {
			targetCard.Content = &yoto.Content{}
		}
		targetCard.Content.Chapters = append(targetCard.Content.Chapters, newChapter)

		// Recalculate Stats
		utils.ReorderPlaylist(targetCard)
		var totalDur, totalSize int
		for _, c := range targetCard.Content.Chapters {
			totalDur += c.Duration
			if len(c.Tracks) > 0 {
				totalSize += c.Tracks[0].FileSize
			}
		}
		if targetCard.Metadata == nil {
			targetCard.Metadata = &yoto.Metadata{}
		}
		targetCard.Metadata.Media.Duration = totalDur
		targetCard.Metadata.Media.FileSize = totalSize

		// Save
		if targetCard.CardID != "" {
			fmt.Printf("Updating playlist '%s'...\n", targetCard.Title)
			return apiClient.UpdateCard(targetCard.CardID, targetCard)
		} 
		
		fmt.Printf("Creating playlist '%s'...\n", targetCard.Title)
		return apiClient.CreateCard(targetCard)
	},
}

func init() {
	importCmd.Flags().StringVarP(&importPlaylist, "playlist", "p", "", "Target playlist name (optional)")
	importCmd.Flags().BoolVar(&importNoNormalize, "no-normalize", false, "Disable audio normalization")
	rootCmd.AddCommand(importCmd)
}
