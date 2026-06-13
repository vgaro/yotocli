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
func AddTrack(client *yoto.Client, playlistQuery string, filePath string, title string, iconID string, normalize bool, trimStart, trimEnd float64, log Logger) error {
	if log == nil {
		log = func(s string, i ...interface{}) {}
	}

	uploadPath := filePath
	if normalize || trimStart > 0 || trimEnd > 0 {
		msg := "Processing audio..."
		if normalize && (trimStart > 0 || trimEnd > 0) {
			msg = fmt.Sprintf("Normalizing and trimming %s...", filepath.Base(filePath))
		} else if normalize {
			msg = fmt.Sprintf("Normalizing %s...", filepath.Base(filePath))
		} else {
			msg = fmt.Sprintf("Trimming %s...", filepath.Base(filePath))
		}
		log(msg)

		procPath, err := processing.ProcessAudio(filePath, normalize, trimStart, trimEnd)
		if err != nil {
			log("Warning: Processing failed: %v. Using original file.", err)
		} else {
			uploadPath = procPath
			defer os.Remove(procPath)
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
        trackTitle := title
        if trackTitle == "" {
                trackTitle = strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))
        }

        // Determine icon
        iconVal := iconID
        if iconVal == "" {
                iconVal = "yoto:#aUm9i3ex3qqAMYBv-i-O-pYMKuMJGICtR3Vhf289u2Q" // Default standard icon
        } else if !strings.HasPrefix(iconVal, "yoto:#") && !strings.HasPrefix(iconVal, "http") {
                iconVal = "yoto:#" + iconVal
        }

        newTrack := yoto.Track{
                Title:        trackTitle,
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
                Title:    trackTitle,
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
