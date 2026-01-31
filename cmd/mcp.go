package cmd

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/cobra"
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
	// Logger that writes to stderr so MCP client doesn't see it as response, or just ignore.
	// For MCP, we usually want to be silent or return progress via notifications (complex).
	// Let's just log to stderr for server ops debugging.
	logger := func(format string, args ...interface{}) {
		// fmt.Fprintf(os.Stderr, format+"\n", args...)
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
	NoNormalize  bool   `json:"no_normalize,omitempty" jsonschema:"Disable audio normalization (default: false)"`
}

func addTrackHandler(ctx context.Context, req *mcp.CallToolRequest, input AddTrackInput) (*mcp.CallToolResult, SimpleOutput, error) {
	// Simple logger
	logger := func(format string, args ...interface{}) {
		// fmt.Fprintf(os.Stderr, format+"\n", args...)
	}

	err := actions.AddTrack(apiClient, input.PlaylistName, input.FilePath, !input.NoNormalize, logger)
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
	// Note: We need to format it as "yoto:#HASH" if it's not already, or just trust the input?
	// The client library's `sanitizeCardForUpdate` handles the prefixing if it looks like a hash?
	// Let's rely on the input being correct or `UpdateCard` fixing it.
	// Actually, `pkg/yoto/client.go` `fixIcon` expects full URL or just hash?
	// Let's assume input is the hash (ID) or full string.
	// The other tracks show full URLs in `GetCard` response, but `UpdateCard` expects `yoto:#HASH`.
	// If I pass the raw ID `Mnq...`, `fixIcon` might not touch it if it doesn't start with `http`.
	// But `Display` struct has `Icon16x16`.
	// Let's prepend `yoto:#` if it doesn't have it.
	
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
	// Return the ID in the message or as structured data? 
	// SimpleOutput only has Message.
	// Let's put it in the message.
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

// mcpCmd represents the mcp command
var mcpCmd = &cobra.Command{
	Use:    "mcp",
	Short:  "Start the MCP server for Yoto",
	Hidden: true,
	Run: func(cmd *cobra.Command, args []string) {
		// Create MCP Server
		s := mcp.NewServer(&mcp.Implementation{
			Name:    "yoto-mcp",
			Version: "1.0.0",
		}, nil)

		// Register Tools
		mcp.AddTool(s, &mcp.Tool{Name: "list_playlists", Description: "List all Yoto cards/playlists in the library"}, listPlaylistsHandler)
		mcp.AddTool(s, &mcp.Tool{Name: "get_playlist", Description: "Get details of a specific playlist"}, getPlaylistHandler)
		mcp.AddTool(s, &mcp.Tool{Name: "list_devices", Description: "List registered Yoto players"}, listDevicesHandler)
		mcp.AddTool(s, &mcp.Tool{Name: "get_device_status", Description: "Check battery/volume of a player"}, getDeviceStatusHandler)
		mcp.AddTool(s, &mcp.Tool{Name: "create_playlist", Description: "Create a new empty playlist"}, createPlaylistHandler)
		mcp.AddTool(s, &mcp.Tool{Name: "delete_playlist", Description: "Delete a playlist by ID"}, deletePlaylistHandler)
		mcp.AddTool(s, &mcp.Tool{Name: "edit_playlist", Description: "Edit playlist metadata (title, author, description)"}, editPlaylistHandler)
		mcp.AddTool(s, &mcp.Tool{Name: "import_from_url", Description: "Download audio from a URL (YouTube, etc) and add to playlist"}, importFromURLHandler)
		mcp.AddTool(s, &mcp.Tool{Name: "add_track", Description: "Upload a local audio file to a playlist"}, addTrackHandler)
		mcp.AddTool(s, &mcp.Tool{Name: "set_track_icon", Description: "Set the icon for a specific track"}, setTrackIconHandler)
		mcp.AddTool(s, &mcp.Tool{Name: "upload_icon", Description: "Upload a custom icon"}, uploadIconHandler)
		mcp.AddTool(s, &mcp.Tool{Name: "set_volume", Description: "Set the volume of a player (0-100)"}, setVolumeHandler)
		mcp.AddTool(s, &mcp.Tool{Name: "play_card", Description: "Start playing a playlist on a device"}, playCardHandler)
		mcp.AddTool(s, &mcp.Tool{Name: "stop_player", Description: "Stop playback on a device"}, stopPlayerHandler)
		mcp.AddTool(s, &mcp.Tool{Name: "pause_player", Description: "Pause playback on a device"}, pausePlayerHandler)

		// Start Server
		// Using StdioTransport from the SDK
		// Note: We need to make sure we use the correct Transport struct name.
		// The example used &mcp.StdioTransport{}.
		if err := s.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
			log.Fatalf("Server failed: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(mcpCmd)
}