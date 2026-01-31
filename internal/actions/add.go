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
func AddTrack(client *yoto.Client, playlistQuery string, filePath string, normalize bool, log Logger) error {
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
	position := -1

	if len(parts) > 1 {
		if p, err := utils.ParseIndex(parts[1]); err == nil {
			position = p - 1 // 0-based
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

	newTrack := yoto.Track{
		Title:        title,
		TrackURL:     fmt.Sprintf("yoto:#%s", transData.TranscodedSha256),
		Duration:     transData.TranscodedInfo.Duration,
		FileSize:     transData.TranscodedInfo.FileSize,
		Format:       transData.TranscodedInfo.Format,
		OverlayLabel: "1", // Placeholder
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

	if targetCard.Content == nil {
		targetCard.Content = &yoto.Content{}
	}

	// Insert or Append
	if position < 0 || position >= len(targetCard.Content.Chapters) {
		targetCard.Content.Chapters = append(targetCard.Content.Chapters, newChapter)
	} else {
		// Insert at position: extend slice by 1, move elements, set new element
		targetCard.Content.Chapters = append(targetCard.Content.Chapters, yoto.Chapter{})
		copy(targetCard.Content.Chapters[position+1:], targetCard.Content.Chapters[position:])
		targetCard.Content.Chapters[position] = newChapter
	}

	// Renumber and Calc Stats
	var totalDur, totalSize int
	for i := range targetCard.Content.Chapters {
		key := fmt.Sprintf("%02d", i+1)
		targetCard.Content.Chapters[i].Key = key
		targetCard.Content.Chapters[i].OverlayLabel = fmt.Sprintf("%d", i+1)
		for j := range targetCard.Content.Chapters[i].Tracks {
			targetCard.Content.Chapters[i].Tracks[j].Key = key
			targetCard.Content.Chapters[i].Tracks[j].OverlayLabel = fmt.Sprintf("%d", i+1)
		}

		totalDur += targetCard.Content.Chapters[i].Duration
		if len(targetCard.Content.Chapters[i].Tracks) > 0 {
			totalSize += targetCard.Content.Chapters[i].Tracks[0].FileSize
		}
	}

	if targetCard.Metadata == nil {
		targetCard.Metadata = &yoto.Metadata{}
	}
	targetCard.Metadata.Media.Duration = totalDur
	targetCard.Metadata.Media.FileSize = totalSize

	if targetCard.CardID != "" {
		log("Updating playlist '%s'...", targetCard.Title)
		return client.UpdateCard(targetCard.CardID, targetCard)
	}
	log("Creating playlist '%s'...", targetCard.Title)
	return client.CreateCard(targetCard)
}
