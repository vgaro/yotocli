package actions

import (
	"fmt"
	"os"

	"github.com/vgaro/yotocli/internal/processing"
	"github.com/vgaro/yotocli/internal/utils"
	"github.com/vgaro/yotocli/pkg/yoto"
)

type Logger func(string, ...interface{})

func ImportFromURL(client *yoto.Client, url string, playlistName string, normalize bool, log Logger) error {
	if log == nil {
		log = func(s string, i ...interface{}) {}
	}

	log("Downloading audio from %s...", url)
	filePath, title, err := processing.DownloadFromURL(url)
	if err != nil {
		return err
	}
	defer os.Remove(filePath)

	log("Downloaded: %s", title)

	uploadPath := filePath
	if normalize {
		log("Normalizing audio...")
		normPath, err := processing.NormalizeAudio(filePath)
		if err != nil {
			log("Warning: Normalization failed: %v. Using original.", err)
		} else {
			uploadPath = normPath
			defer os.Remove(normPath)
		}
	}

	cards, err := client.ListCards()
	if err != nil {
		return err
	}

	var targetCard *yoto.Card
	if playlistName != "" {
		targetCard = utils.FindCard(cards, playlistName)
		if targetCard == nil {
			log("Playlist '%s' not found. Creating it...", playlistName)
			targetCard = &yoto.Card{
				Title:   playlistName,
				Content: &yoto.Content{},
			}
		} else {
			fullCard, err := client.GetCard(targetCard.CardID)
			if err != nil {
				return err
			}
			targetCard = fullCard
		}
	} else {
		log("No playlist specified. Creating new playlist: '%s'", title)
		targetCard = &yoto.Card{
			Title:   title,
			Content: &yoto.Content{},
		}
	}

	log("Uploading to Yoto...")
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

	newTrack := yoto.Track{
		Title:    title,
		TrackURL: fmt.Sprintf("yoto:#%s", transData.TranscodedSha256),
		Duration: transData.TranscodedInfo.Duration,
		FileSize: transData.TranscodedInfo.FileSize,
		Format:   transData.TranscodedInfo.Format,
		Type:     "audio",
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
	targetCard.Content.Chapters = append(targetCard.Content.Chapters, newChapter)

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

	if targetCard.CardID != "" {
		log("Updating playlist '%s'...", targetCard.Title)
		return client.UpdateCard(targetCard.CardID, targetCard)
	}

	log("Creating playlist '%s'...", targetCard.Title)
	return client.CreateCard(targetCard)
}
