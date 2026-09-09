package actions

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"github.com/vgaro/yotocli/internal/utils"
	"github.com/vgaro/yotocli/pkg/yoto"
	"golang.org/x/sync/errgroup"
)

// defaultIcon is the plain audio icon Yoto ships, used when a caller has no
// icon of its own.
const defaultIcon = "yoto:#aUm9i3ex3qqAMYBv-i-O-pYMKuMJGICtR3Vhf289u2Q"

// maxConcurrentUploads caps how many files AddTracks has in flight at once.
//
// An upload spends nearly all of its wall time waiting on Yoto: the PUT of the
// file, then polling until the transcode finishes. So the limit is about not
// leaning on the API too hard rather than about work on our end.
const maxConcurrentUploads = 10

// Track is one file to put on a card: what a caller knows before anything has
// been uploaded. It is not yoto.Track, which is the API's idea of a track and
// can only be filled in once Yoto has transcoded the audio.
type Track struct {
	Path   string // path to the local audio file
	Title  string // name on the card; "" derives one from the file name
	IconID string // icon hash or "yoto:#..."; "" uses the default icon
}

// AddTracks uploads files and adds them to one playlist, keeping the order they
// were given in. playlistQuery can be "Name" or "Name/Position"; a playlist that
// doesn't exist is created.
//
// The playlist is looked up once and written once, with the uploads in between
// running concurrently. Looking it up once is what keeps a batch from creating
// several copies of the same new playlist, and writing it once is what keeps
// concurrent read-modify-write cycles from dropping each other's chapters.
//
// A failure in any upload abandons the whole batch: the card is only written
// after every upload has succeeded, so a playlist is never left half filled in.
func AddTracks(client *yoto.Client, playlistQuery string, tracks []Track, log Logger) error {
	if log == nil {
		log = func(s string, i ...interface{}) {}
	}
	if len(tracks) == 0 {
		return fmt.Errorf("no files to add")
	}

	// Uploads report progress from their own goroutines, so serialize the
	// callback rather than trusting every caller's logger to be safe for it.
	var logMu sync.Mutex
	logf := func(format string, args ...interface{}) {
		logMu.Lock()
		defer logMu.Unlock()
		log(format, args...)
	}

	targetCard, position, err := findOrCreateCard(client, playlistQuery, logf)
	if err != nil {
		return err
	}

	// Indexed rather than appended, so the card ends up in the order asked for
	// regardless of which upload finishes first.
	chapters := make([]yoto.Chapter, len(tracks))

	g := new(errgroup.Group)
	g.SetLimit(maxConcurrentUploads)
	for i, track := range tracks {
		g.Go(func() error {
			name := trackTitle(track.Title, track.Path)
			logf("%sUploading %s...", progress(i, len(tracks)), name)

			chapter, err := uploadChapter(client, track)
			if err != nil {
				return fmt.Errorf("failed to add %q: %w", name, err)
			}

			chapters[i] = chapter
			logf("%sUploaded %s", progress(i, len(tracks)), name)
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return err
	}

	insertChapters(targetCard, chapters, position)
	utils.ReorderPlaylist(targetCard)
	setMediaTotals(targetCard)

	if targetCard.CardID != "" {
		logf("Updating playlist '%s'...", targetCard.Title)
		return client.UpdateCard(targetCard.CardID, targetCard)
	}
	logf("Creating playlist '%s'...", targetCard.Title)
	return client.CreateCard(targetCard)
}

// findOrCreateCard resolves a "Name" or "Name/Position" query to the card the
// tracks should go on, fetched in full so chapters can be added to what is
// already there, along with the 0-based position they belong at (-1 to append).
// A playlist that does not exist yet comes back unsaved, without a CardID.
func findOrCreateCard(client *yoto.Client, playlistQuery string, log Logger) (*yoto.Card, int, error) {
	parts := strings.Split(playlistQuery, "/")
	cardName := parts[0]
	position := -1

	if len(parts) > 1 {
		if p, err := utils.ParseIndex(parts[1]); err == nil {
			position = p - 1 // 0-based
		}
	}

	cards, err := client.ListCards()
	if err != nil {
		return nil, 0, err
	}

	existingCard := utils.FindCard(cards, cardName)
	if existingCard == nil {
		log("Playlist '%s' not found. Creating it...", cardName)
		return &yoto.Card{
			Title:   cardName,
			Content: &yoto.Content{},
		}, position, nil
	}

	fullCard, err := client.GetCard(existingCard.CardID)
	if err != nil {
		return nil, 0, err
	}
	return fullCard, position, nil
}

// uploadChapter uploads one file and describes it as a chapter, without going
// near the card. It is the only part of AddTracks that runs concurrently, and
// it shares nothing but the client: each call gets its own upload ID from Yoto.
func uploadChapter(client *yoto.Client, track Track) (yoto.Chapter, error) {
	hash, err := yoto.FileSHA256(track.Path)
	if err != nil {
		return yoto.Chapter{}, err
	}

	upData, err := client.GetUploadURL(hash, filepath.Base(track.Path))
	if err != nil {
		return yoto.Chapter{}, err
	}
	if upData.Upload.UploadID == "" {
		return yoto.Chapter{}, fmt.Errorf("yoto returned no upload id for %s", track.Path)
	}

	// No upload URL means Yoto recognised the hash and already holds this audio,
	// so there is nothing to send. Re-importing a feed only pays for the
	// episodes that are new.
	//
	// Otherwise the file goes up exactly as it is on disk: Yoto's transcoder
	// normalizes it (loudnorm to -16 LUFS) and re-encodes it to Opus on the way
	// in, so anything done to the audio first would only be undone.
	if upData.Upload.UploadURL != "" {
		if err := client.UploadFile(track.Path, upData.Upload.UploadURL); err != nil {
			return yoto.Chapter{}, err
		}
	}

	transData, err := client.PollTranscode(upData.Upload.UploadID)
	if err != nil {
		return yoto.Chapter{}, err
	}

	title := trackTitle(track.Title, track.Path)
	display := yoto.Display{Icon16x16: resolveIcon(track.IconID)}

	// Key and OverlayLabel are left to ReorderPlaylist, which can only number
	// a track once it knows where on the card it landed.
	uploaded := yoto.Track{
		Title:    title,
		TrackURL: fmt.Sprintf("yoto:#%s", transData.TranscodedSha256),
		Duration: transData.TranscodedInfo.Duration,
		FileSize: transData.TranscodedInfo.FileSize,
		Format:   transData.TranscodedInfo.Format,
		Type:     "audio",
		Display:  display,
	}

	return yoto.Chapter{
		Title:    title,
		Duration: uploaded.Duration,
		Tracks:   []yoto.Track{uploaded},
		Display:  display,
	}, nil
}

// insertChapters adds chapters to the card at position, keeping the order they
// were given in. A position outside the card appends, which is what a query
// without a position asks for.
func insertChapters(card *yoto.Card, chapters []yoto.Chapter, position int) {
	if card.Content == nil {
		card.Content = &yoto.Content{}
	}

	existing := card.Content.Chapters
	if position < 0 || position >= len(existing) {
		card.Content.Chapters = append(existing, chapters...)
		return
	}

	merged := make([]yoto.Chapter, 0, len(existing)+len(chapters))
	merged = append(merged, existing[:position]...)
	merged = append(merged, chapters...)
	merged = append(merged, existing[position:]...)
	card.Content.Chapters = merged
}

// setMediaTotals recomputes the card level duration and size the app displays,
// which have to be kept in step with the chapters by hand.
func setMediaTotals(card *yoto.Card) {
	if card.Content == nil {
		return
	}

	var totalDur, totalSize int
	for _, chapter := range card.Content.Chapters {
		totalDur += chapter.Duration
		if len(chapter.Tracks) > 0 {
			totalSize += chapter.Tracks[0].FileSize
		}
	}

	if card.Metadata == nil {
		card.Metadata = &yoto.Metadata{}
	}
	card.Metadata.Media.Duration = totalDur
	card.Metadata.Media.FileSize = totalSize
}

// progress labels a line with its place in the batch, as in "[3/12] ". Uploads
// finish out of order, so the number says which file a line is about rather
// than how far along the batch is. A lone file needs no label.
func progress(i int, total int) string {
	if total < 2 {
		return ""
	}
	return fmt.Sprintf("[%d/%d] ", i+1, total)
}

// resolveIcon turns an icon as the user might give it - a bare hash, a
// "yoto:#..." reference or a URL - into what the API expects.
func resolveIcon(iconID string) string {
	if iconID == "" {
		return defaultIcon
	}
	if strings.HasPrefix(iconID, "yoto:#") || strings.HasPrefix(iconID, "http") {
		return iconID
	}
	return "yoto:#" + iconID
}

// trackTitle picks the name to show for a track, falling back to the file name
// with its extension stripped when the caller has no title of its own.
//
// The fallback is only a decent guess for a file the user named. Downloads are
// the case where it is wrong: DownloadFromURL names files after the yt-dlp ID
// to keep awkward characters out of paths, so a whole podcast feed would land
// on the card as a column of IDs. Those callers pass Download.Name instead.
func trackTitle(title string, filePath string) string {
	if strings.TrimSpace(title) != "" {
		return title
	}
	return strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))
}
