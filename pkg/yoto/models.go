package yoto

import "time"

// Card represents a Yoto card (playlist or physical card)
type Card struct {
	CardID    string    `json:"cardId"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	Content   *Content  `json:"content"`
	Metadata  *Metadata `json:"metadata"`
}

// Content contains the actual audio structure
type Content struct {
	Chapters []Chapter `json:"chapters"`
}

// Chapter represents a group of tracks (usually 1:1 with tracks for MYO)
type Chapter struct {
	Key          string  `json:"key"`
	Title        string  `json:"title"`
	Duration     int     `json:"duration"`
	Tracks       []Track `json:"tracks"`
	Display      Display `json:"display"`
	OverlayLabel string  `json:"overlayLabel"`
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
	OverlayLabel string  `json:"overlayLabel"`
	Type         string  `json:"type"`
}

// Display holds icon information
type Display struct {
	Icon16x16 string `json:"icon16x16"`
}

// Metadata holds descriptive info
type Metadata struct {
	Author      string `json:"author"`
	Description string `json:"description"`
	Media       Media  `json:"media"`
}

// Media holds aggregate stats
type Media struct {
	Duration int `json:"duration"`
	FileSize int `json:"fileSize"`
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
