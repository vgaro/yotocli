package yoto

import (
	"strings"
	"time"
)

// Card represents a Yoto card (playlist or physical card)
type Card struct {
	CardID    string    `json:"cardId"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	Content   *Content  `json:"content"`
	Metadata  *Metadata `json:"metadata"`

	extra extras // fields the API returned that this client does not model
}

// Content contains the actual audio structure
type Content struct {
	Chapters []Chapter `json:"chapters"`

	extra extras
}

// Chapter represents a group of tracks (usually 1:1 with tracks for MYO)
type Chapter struct {
	Key          string  `json:"key"`
	Title        string  `json:"title"`
	Duration     int     `json:"duration"`
	Tracks       []Track `json:"tracks"`
	Display      Display `json:"display"`
	OverlayLabel string  `json:"overlayLabel,omitempty"`

	extra extras
}

// Track represents a single audio file
type Track struct {
	Key          string  `json:"key"`
	Title        string  `json:"title"`
	TrackURL     string  `json:"trackUrl"`
	Duration     int     `json:"duration"`
	FileSize     int     `json:"fileSize"`
	Format       string  `json:"format"`
	Display      Display `json:"display"`
	OverlayLabel string  `json:"overlayLabel,omitempty"`
	Type         string  `json:"type"`

	extra extras
}

// Display holds icon information
type Display struct {
	Icon16x16 string `json:"icon16x16"`

	extra extras
}

// Metadata holds descriptive info
//
// Author and Description are omitted when empty: an empty one means "not being
// changed" everywhere it is set from, and writing it would add a field to cards
// that never had one.
type Metadata struct {
	Author      string `json:"author,omitempty"`
	Description string `json:"description,omitempty"`
	Media       Media  `json:"media"`

	extra extras
}

// Media holds aggregate stats
type Media struct {
	Duration int `json:"duration"`
	FileSize int `json:"fileSize"`

	extra extras
}

// LibraryResponse is the top-level response from /card/family/library
type LibraryResponse struct {
	Cards []LibraryItem `json:"cards"`
}

type LibraryItem struct {
	CardID string `json:"cardId"`
	Card   Card   `json:"card"`
}

// Device represents a Yoto player
type Device struct {
	ID         string `json:"deviceId"`
	Name       string `json:"name"`
	DeviceType string `json:"deviceType"`
	Online     bool   `json:"online"`
	Status     *DeviceStatus
}

type DeviceStatus struct {
	BatteryLevel int    `json:"batteryLevel"`
	IsCharging   int    `json:"isCharging"` // 0=No, 1=Yes
	ActiveCard   string `json:"activeCard"` // "none" or card ID
	Volume       int    `json:"volume"`
}

type DevicesResponse struct {
	Devices []Device `json:"devices"`
}

type DeviceStatusResponse struct {
	Status DeviceStatus `json:"status"`
}

// AudioSHA256 identifies the audio a track plays: the SHA-256 of its transcoded
// file, in unpadded base64url. Empty if the URL carries no hash.
//
// The API gives that hash in two shapes, and they have to be treated as the same
// thing. A track written to a card names its audio as "yoto:#<hash>"; the same
// track read back names it with a signed CDN URL, which carries the hash in a
// "#sha256=<hash>" fragment.
func AudioSHA256(trackURL string) string {
	_, fragment, ok := strings.Cut(trackURL, "#")
	if !ok {
		return ""
	}
	return strings.TrimPrefix(fragment, "sha256=")
}

// IconRef normalizes an icon to the "yoto:#<hash>" form a card has to be written
// with. Icons are read back as https URLs ending in the hash, and writing one of
// those back leaves the track with no icon at all.
//
// Anything else is returned unchanged, including a URL that does not end in what
// looks like a hash: better to send it and let the API rule on it than to mangle
// it here.
func IconRef(icon string) string {
	if !strings.HasPrefix(icon, "http") {
		return icon
	}

	hash, _, _ := strings.Cut(icon, "?")
	if idx := strings.LastIndex(hash, "/"); idx != -1 {
		hash = hash[idx+1:]
	}
	if len(hash) != sha256RefLength {
		return icon
	}
	return "yoto:#" + hash
}

// sha256RefLength is how long a SHA-256 is once Yoto has written it out: 32 bytes
// in unpadded base64url.
const sha256RefLength = 43
