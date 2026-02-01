package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/vgaro/yotocli/internal/actions"
	"github.com/vgaro/yotocli/pkg/yoto"
)

// --- Tool Definitions ---

type EmptyInput struct{}

// List Playlists
type ListPlaylistsOutput struct {
	Playlists []PlaylistSummary `json:"playlists"`
}
type PlaylistSummary struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

func listPlaylistsHandler(ctx context.Context, req *mcp.CallToolRequest, input EmptyInput) (*mcp.CallToolResult, ListPlaylistsOutput, error) {
	cards, err := apiClient.ListCards()
	if err != nil {
		return nil, ListPlaylistsOutput{}, err
	}
	summary := make([]PlaylistSummary, len(cards))
	for i, c := range cards {
		summary[i] = PlaylistSummary{
			ID:    c.CardID,
			Title: c.Title,
		}
	}
	return nil, ListPlaylistsOutput{Playlists: summary}, nil
}

// Get Playlist
type GetPlaylistInput struct {
	PlaylistID string `json:"playlist_id" jsonschema:"The UUID of the playlist/card"`
}
type GetPlaylistOutput struct {
	Card *yoto.Card `json:"card"`
}

func getPlaylistHandler(ctx context.Context, req *mcp.CallToolRequest, input GetPlaylistInput) (*mcp.CallToolResult, GetPlaylistOutput, error) {
	card, err := apiClient.GetCard(input.PlaylistID)
	if err != nil {
		return nil, GetPlaylistOutput{}, err
	}
	return nil, GetPlaylistOutput{Card: card}, nil
}

// List Devices
type ListDevicesOutput struct {
	Devices []yoto.Device `json:"devices"`
}

func listDevicesHandler(ctx context.Context, req *mcp.CallToolRequest, input EmptyInput) (*mcp.CallToolResult, ListDevicesOutput, error) {
	devices, err := apiClient.ListDevices()
	if err != nil {
		return nil, ListDevicesOutput{}, err
	}
	return nil, ListDevicesOutput{Devices: devices}, nil
}

// Get Device Status
type GetDeviceStatusInput struct {
	DeviceID string `json:"device_id" jsonschema:"The Device ID"`
}
type GetDeviceStatusOutput struct {
	Status *yoto.DeviceStatus `json:"status"`
}

func getDeviceStatusHandler(ctx context.Context, req *mcp.CallToolRequest, input GetDeviceStatusInput) (*mcp.CallToolResult, GetDeviceStatusOutput, error) {
	status, err := apiClient.GetDeviceStatus(input.DeviceID)
	if err != nil {
		return nil, GetDeviceStatusOutput{}, err
	}
	return nil, GetDeviceStatusOutput{Status: status}, nil
}

// Create Playlist
type CreatePlaylistInput struct {
	Title       string `json:"title" jsonschema:"The title of the new playlist"`
	Description string `json:"description,omitempty" jsonschema:"Optional description"`
	Author      string `json:"author,omitempty" jsonschema:"Optional author name"`
}
type SimpleOutput struct {
	Message string `json:"message"`
}

func createPlaylistHandler(ctx context.Context, req *mcp.CallToolRequest, input CreatePlaylistInput) (*mcp.CallToolResult, SimpleOutput, error) {
	newCard := &yoto.Card{
		Title: input.Title,
		Metadata: &yoto.Metadata{
			Description: input.Description,
			Author:      input.Author,
		},
		Content: &yoto.Content{
			Chapters: []yoto.Chapter{},
		},
	}
	err := apiClient.CreateCard(newCard)
	if err != nil {
		return nil, SimpleOutput{}, err
	}
	return nil, SimpleOutput{Message: "Playlist created successfully"}, nil
}

// Delete Playlist
type DeletePlaylistInput struct {
	PlaylistID string `json:"playlist_id" jsonschema:"The UUID of the playlist to delete"`
}

func deletePlaylistHandler(ctx context.Context, req *mcp.CallToolRequest, input DeletePlaylistInput) (*mcp.CallToolResult, SimpleOutput, error) {
	err := apiClient.DeleteCard(input.PlaylistID)
	if err != nil {
		return nil, SimpleOutput{}, err
	}
	return nil, SimpleOutput{Message: fmt.Sprintf("Playlist %s deleted", input.PlaylistID)}, nil
}

// Edit Playlist
type EditPlaylistInput struct {
	PlaylistID  string `json:"playlist_id" jsonschema:"The UUID of the playlist to edit"`
	Title       string `json:"title,omitempty" jsonschema:"New title (optional)"`
	Description string `json:"description,omitempty" jsonschema:"New description (optional)"`
	Author      string `json:"author,omitempty" jsonschema:"New author (optional)"`
}

func editPlaylistHandler(ctx context.Context, req *mcp.CallToolRequest, input EditPlaylistInput) (*mcp.CallToolResult, SimpleOutput, error) {
	card, err := apiClient.GetCard(input.PlaylistID)
	if err != nil {
		return nil, SimpleOutput{}, err
	}

	changed := false
	if input.Title != "" {
		card.Title = input.Title
		changed = true
	}

	if card.Metadata == nil {
		card.Metadata = &yoto.Metadata{}
	}

	if input.Description != "" {
		card.Metadata.Description = input.Description
		changed = true
	}
	if input.Author != "" {
		card.Metadata.Author = input.Author
		changed = true
	}

	if !changed {
		return nil, SimpleOutput{Message: "No changes requested"}, nil
	}

	err = apiClient.UpdateCard(card.CardID, card)
	if err != nil {
		return nil, SimpleOutput{}, err
	}

	return nil, SimpleOutput{Message: "Playlist updated successfully"}, nil
}

// Import from URL
type ImportFromURLInput struct {
	URL          string `json:"url" jsonschema:"The URL of the audio/video to download (e.g., YouTube)"`
	PlaylistName string `json:"playlist_name,omitempty" jsonschema:"The name of the playlist to add to (creates new if empty or not found)"`
	NoNormalize  bool   `json:"no_normalize,omitempty" jsonschema:"Disable audio normalization (default: false)"`
}

func importFromURLHandler(ctx context.Context, req *mcp.CallToolRequest, input ImportFromURLInput) (*mcp.CallToolResult, SimpleOutput, error) {
	// Logger that writes to stderr so MCP client doesn't see it as response
	logger := func(format string, args ...interface{}) {
		fmt.Fprintf(os.Stderr, format+"\n", args...)
	}

	err := actions.ImportFromURL(apiClient, input.URL, input.PlaylistName, !input.NoNormalize, logger)
	if err != nil {
		return nil, SimpleOutput{}, err
	}
	return nil, SimpleOutput{Message: "Import successful"}, nil
}

// Add Track (Local File)
type AddTrackInput struct {
	FilePath     string `json:"file_path" jsonschema:"The path to the local audio file to upload"`
	PlaylistName string `json:"playlist_name" jsonschema:"The name of the playlist to add to (creates new if not found). Can specify position like 'Name/1'."`
	IconID       string `json:"icon_id,omitempty" jsonschema:"Optional icon ID (e.g. from upload_icon)"`
	NoNormalize  bool   `json:"no_normalize,omitempty" jsonschema:"Disable audio normalization (default: false)"`
}

func addTrackHandler(ctx context.Context, req *mcp.CallToolRequest, input AddTrackInput) (*mcp.CallToolResult, SimpleOutput, error) {
	// Simple logger
	logger := func(format string, args ...interface{}) {
		fmt.Fprintf(os.Stderr, format+"\n", args...)
	}

	err := actions.AddTrack(apiClient, input.PlaylistName, input.FilePath, input.IconID, !input.NoNormalize, logger)
	if err != nil {
		return nil, SimpleOutput{}, err
	}
	return nil, SimpleOutput{Message: "Track added successfully"}, nil
}

// Set Track Icon
type SetTrackIconInput struct {
	PlaylistID string `json:"playlist_id" jsonschema:"The ID of the playlist"`
	TrackIndex int    `json:"track_index" jsonschema:"The 1-based index of the track to update"`
	IconID     string `json:"icon_id" jsonschema:"The Yoto Icon ID (e.g. yoto:#... or hash)"`
}

func setTrackIconHandler(ctx context.Context, req *mcp.CallToolRequest, input SetTrackIconInput) (*mcp.CallToolResult, SimpleOutput, error) {
	card, err := apiClient.GetCard(input.PlaylistID)
	if err != nil {
		return nil, SimpleOutput{}, err
	}

	if card.Content == nil || input.TrackIndex < 1 || input.TrackIndex > len(card.Content.Chapters) {
		return nil, SimpleOutput{}, fmt.Errorf("invalid track index")
	}

	// Update Icon
	iconVal := input.IconID
	if !strings.HasPrefix(iconVal, "yoto:#") && !strings.HasPrefix(iconVal, "http") {
		iconVal = "yoto:#" + iconVal
	}

	idx := input.TrackIndex - 1
	card.Content.Chapters[idx].Display.Icon16x16 = iconVal
	// Also update the tracks inside the chapter (usually 1:1)
	for j := range card.Content.Chapters[idx].Tracks {
		card.Content.Chapters[idx].Tracks[j].Display.Icon16x16 = iconVal
	}

	err = apiClient.UpdateCard(card.CardID, card)
	if err != nil {
		return nil, SimpleOutput{}, err
	}

	return nil, SimpleOutput{Message: "Icon updated successfully"}, nil
}

// Upload Icon
type UploadIconInput struct {
	FilePath string `json:"file_path" jsonschema:"Path to the icon file (local path or URL)"`
}

func uploadIconHandler(ctx context.Context, req *mcp.CallToolRequest, input UploadIconInput) (*mcp.CallToolResult, SimpleOutput, error) {
	id, err := actions.UploadIcon(apiClient, input.FilePath)
	if err != nil {
		return nil, SimpleOutput{}, err
	}
	return nil, SimpleOutput{Message: fmt.Sprintf("Icon uploaded. ID: %s", id)}, nil
}

// Set Volume
type SetVolumeInput struct {
	Volume   int    `json:"volume" jsonschema:"Volume level (0-100)"`
	DeviceID string `json:"device_id,omitempty" jsonschema:"Device ID (optional, defaults to first found)"`
}

func setVolumeHandler(ctx context.Context, req *mcp.CallToolRequest, input SetVolumeInput) (*mcp.CallToolResult, SimpleOutput, error) {
	if input.Volume < 0 || input.Volume > 100 {
		return nil, SimpleOutput{}, fmt.Errorf("volume must be 0-100")
	}

	targetID := input.DeviceID
	if targetID == "" {
		devices, err := apiClient.ListDevices()
		if err != nil {
			return nil, SimpleOutput{}, err
		}
		if len(devices) == 0 {
			return nil, SimpleOutput{}, fmt.Errorf("no devices found")
		}
		targetID = devices[0].ID
	}

	err := apiClient.SetVolume(targetID, input.Volume)
	if err != nil {
		return nil, SimpleOutput{}, err
	}
	return nil, SimpleOutput{Message: fmt.Sprintf("Volume set to %d", input.Volume)}, nil
}

// Play Card
type PlayCardInput struct {
	PlaylistID string `json:"playlist_id" jsonschema:"The ID of the playlist to play"`
	DeviceID   string `json:"device_id,omitempty" jsonschema:"Device ID (optional, defaults to first found)"`
}

func playCardHandler(ctx context.Context, req *mcp.CallToolRequest, input PlayCardInput) (*mcp.CallToolResult, SimpleOutput, error) {
	targetID := input.DeviceID
	if targetID == "" {
		devices, err := apiClient.ListDevices()
		if err != nil {
			return nil, SimpleOutput{}, err
		}
		if len(devices) == 0 {
			return nil, SimpleOutput{}, fmt.Errorf("no devices found")
		}
		targetID = devices[0].ID
	}

	err := apiClient.PlayCard(targetID, input.PlaylistID)
	if err != nil {
		return nil, SimpleOutput{}, err
	}
	return nil, SimpleOutput{Message: "Playback started"}, nil
}

// Stop/Pause Player
type PlayerControlInput struct {
	DeviceID string `json:"device_id,omitempty" jsonschema:"Device ID (optional, defaults to first found)"`
}

func stopPlayerHandler(ctx context.Context, req *mcp.CallToolRequest, input PlayerControlInput) (*mcp.CallToolResult, SimpleOutput, error) {
	targetID := input.DeviceID
	if targetID == "" {
		devices, err := apiClient.ListDevices()
		if err != nil {
			return nil, SimpleOutput{}, err
		}
		if len(devices) == 0 {
			return nil, SimpleOutput{}, fmt.Errorf("no devices found")
		}
		targetID = devices[0].ID
	}

	err := apiClient.StopPlayer(targetID)
	if err != nil {
		return nil, SimpleOutput{}, err
	}
	return nil, SimpleOutput{Message: "Playback stopped"}, nil
}

func pausePlayerHandler(ctx context.Context, req *mcp.CallToolRequest, input PlayerControlInput) (*mcp.CallToolResult, SimpleOutput, error) {
	targetID := input.DeviceID
	if targetID == "" {
		devices, err := apiClient.ListDevices()
		if err != nil {
			return nil, SimpleOutput{}, err
		}
		if len(devices) == 0 {
			return nil, SimpleOutput{}, fmt.Errorf("no devices found")
		}
		targetID = devices[0].ID
	}

	err := apiClient.PausePlayer(targetID)
	if err != nil {
		return nil, SimpleOutput{}, err
	}
	return nil, SimpleOutput{Message: "Playback paused"}, nil
}

// Remove Track
type RemoveTrackInput struct {
	PlaylistID string `json:"playlist_id" jsonschema:"The ID of the playlist"`
	TrackIndex int    `json:"track_index" jsonschema:"The 1-based index of the track to remove"`
}

func removeTrackHandler(ctx context.Context, req *mcp.CallToolRequest, input RemoveTrackInput) (*mcp.CallToolResult, SimpleOutput, error) {
	err := actions.RemoveTrack(apiClient, input.PlaylistID, input.TrackIndex)
	if err != nil {
		return nil, SimpleOutput{}, err
	}
	return nil, SimpleOutput{Message: "Track removed successfully"}, nil
}

// Move Track (Reorder or Move between playlists)
type MoveTrackInput struct {
	PlaylistID     string `json:"playlist_id" jsonschema:"The ID of the source playlist"`
	TrackIndex     int    `json:"track_index" jsonschema:"The 1-based index of the track to move"`
	NewPosition    int    `json:"new_position" jsonschema:"The new 1-based index position in the destination"`
	DestPlaylistID string `json:"dest_playlist_id,omitempty" jsonschema:"The ID of the destination playlist (optional, defaults to source)"`
}

func moveTrackHandler(ctx context.Context, req *mcp.CallToolRequest, input MoveTrackInput) (*mcp.CallToolResult, SimpleOutput, error) {
	err := actions.MoveTrack(apiClient, input.PlaylistID, input.TrackIndex, input.DestPlaylistID, input.NewPosition)
	if err != nil {
		return nil, SimpleOutput{}, err
	}
	dest := "destination"
	if input.DestPlaylistID != "" && input.DestPlaylistID != input.PlaylistID {
		dest = input.DestPlaylistID
	} else {
		dest = "position"
	}
	return nil, SimpleOutput{Message: fmt.Sprintf("Track moved to %s %d", dest, input.NewPosition)}, nil
}

// Copy Track
type CopyTrackInput struct {
	PlaylistID     string `json:"playlist_id" jsonschema:"The ID of the source playlist"`
	TrackIndex     int    `json:"track_index" jsonschema:"The 1-based index of the track to copy"`
	DestPlaylistID string `json:"dest_playlist_id" jsonschema:"The ID of the destination playlist (optional, defaults to source/duplicate)"`
	NewPosition    int    `json:"new_position" jsonschema:"The 1-based index position in the destination"`
}

func copyTrackHandler(ctx context.Context, req *mcp.CallToolRequest, input CopyTrackInput) (*mcp.CallToolResult, SimpleOutput, error) {
	err := actions.CopyTrack(apiClient, input.PlaylistID, input.TrackIndex, input.DestPlaylistID, input.NewPosition)
	if err != nil {
		return nil, SimpleOutput{}, err
	}
	return nil, SimpleOutput{Message: "Track copied successfully"}, nil
}
