package actions

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/vgaro/yotocli/internal/processing"
	"github.com/vgaro/yotocli/internal/utils"
	"github.com/vgaro/yotocli/pkg/yoto"
)

// AddTrack uploads a local file and adds it to a playlist.
// playlistQuery can be "Name" or "Name/Position".
// If playlist doesn't exist, it creates it.
func AddTrack(client *yoto.Client, playlistQuery string, filePath string, iconID string, normalize bool, log Logger) error {
	if log == nil {
		log = func(s string, i ...interface{}) {}
	}

	uploadPath := filePath
	if normalize {
		log("Normalizing %s...", filepath.Base(filePath))
		normPath, err := processing.NormalizeAudio(filePath)
		if err != nil {
			log("Warning: Normalization failed: %v. Using original file.", err)
		} else {
			uploadPath = normPath
			defer os.Remove(normPath)
		}
	}

	cards, err := client.ListCards()
	if err != nil {
		return err
	}

	parts := strings.Split(playlistQuery, "/")
	cardName := parts[0]
	position := -1 // Default append (logic.go treats < 1 as append)

	if len(parts) > 1 {
		if p, err := utils.ParseIndex(parts[1]); err == nil {
			position = p // 1-based
		}
	}

	var targetCard *yoto.Card
	existingCard := utils.FindCard(cards, cardName)

	if existingCard == nil {
		log("Playlist '%s' not found. Creating it...", cardName)
		targetCard = &yoto.Card{
			Title:   cardName,
			Content: &yoto.Content{},
		}
	} else {
		fullCard, err := client.GetCard(existingCard.CardID)
		if err != nil {
			return err
		}
		targetCard = fullCard
	}

	log("Uploading %s...", filepath.Base(uploadPath))
	upData, err := client.GetUploadURL()
	if err != nil {
		return err
	}

	if err := client.UploadFile(uploadPath, upData.Upload.UploadURL); err != nil {
		return err
	}

	log("Waiting for transcoding...")
	transData, err := client.PollTranscode(upData.Upload.UploadID)
	if err != nil {
		return err
	}

	// Use filename as title if not provided
	title := strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))

	// Determine icon
	iconVal := iconID
	if iconVal == "" {
		iconVal = "yoto:#aUm9i3ex3qqAMYBv-i-O-pYMKuMJGICtR3Vhf289u2Q" // Default standard icon
	} else if !strings.HasPrefix(iconVal, "yoto:#") && !strings.HasPrefix(iconVal, "http") {
		iconVal = "yoto:#" + iconVal
	}

	newTrack := yoto.Track{
		Title:        title,
		TrackURL:     fmt.Sprintf("yoto:#%s", transData.TranscodedSha256),
		Duration:     transData.TranscodedInfo.Duration,
		FileSize:     transData.TranscodedInfo.FileSize,
		Format:       transData.TranscodedInfo.Format,
		OverlayLabel: "1", // Placeholder
		Type:         "audio",
		Display: yoto.Display{
			Icon16x16: iconVal,
		},
	}

	newChapter := yoto.Chapter{
		Title:    title,
		Duration: newTrack.Duration,
		Tracks:   []yoto.Track{newTrack},
		Display:  newTrack.Display,
	}

	// Use shared logic
	performInsertTrack(targetCard, newChapter, position)
	recalculateMetadata(targetCard)

	if targetCard.CardID != "" {
		log("Updating playlist '%s'...", targetCard.Title)
		return client.UpdateCard(targetCard.CardID, targetCard)
	}
	log("Creating playlist '%s'...", targetCard.Title)
	return client.CreateCard(targetCard)
}
