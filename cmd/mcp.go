package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/cobra"
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