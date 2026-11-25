package cmd

import (
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestMCPToolRegistration(t *testing.T) {
	// This test attempts to register the tools to catch panic/reflection errors
	// similar to what the user reported.
	
	s := mcp.NewServer(&mcp.Implementation{
		Name:    "test-server",
		Version: "1.0.0",
	}, nil)

	// Register tools (copying logic from mcp.go essentially, or just instantiating the struct)
	// We can just call the AddTool with the handler which uses the struct.
	
	// panic: AddTool: tool "get_playlist": input schema: ForType(cmd.GetPlaylistInput): tag must not begin with 'WORD=': "description=The UUID of the playlist/card"
	
	mcp.AddTool(s, &mcp.Tool{Name: "get_playlist", Description: "Get details of a specific playlist"}, getPlaylistHandler)
	mcp.AddTool(s, &mcp.Tool{Name: "list_playlists", Description: "List all Yoto cards/playlists in the library"}, listPlaylistsHandler)
	mcp.AddTool(s, &mcp.Tool{Name: "list_devices", Description: "List registered Yoto players"}, listDevicesHandler)
	mcp.AddTool(s, &mcp.Tool{Name: "get_device_status", Description: "Check battery/volume of a player"}, getDeviceStatusHandler)
	mcp.AddTool(s, &mcp.Tool{Name: "create_playlist", Description: "Create a new empty playlist"}, createPlaylistHandler)
	mcp.AddTool(s, &mcp.Tool{Name: "delete_playlist", Description: "Delete a playlist by ID"}, deletePlaylistHandler)
	mcp.AddTool(s, &mcp.Tool{Name: "edit_playlist", Description: "Edit playlist metadata (title, author, description)"}, editPlaylistHandler)
	mcp.AddTool(s, &mcp.Tool{Name: "import_from_url", Description: "Download audio from a URL (YouTube, etc) and add to playlist"}, importFromURLHandler)
	mcp.AddTool(s, &mcp.Tool{Name: "add_track", Description: "Upload a local audio file to a playlist"}, addTrackHandler)
	mcp.AddTool(s, &mcp.Tool{Name: "set_track_icon", Description: "Set the icon for a specific track"}, setTrackIconHandler)
	mcp.AddTool(s, &mcp.Tool{Name: "list_icons", Description: "List all user-uploaded icons"}, listIconsHandler)
	mcp.AddTool(s, &mcp.Tool{Name: "search_icons", Description: "Search for a custom icon by name or tag"}, searchIconsHandler)
	mcp.AddTool(s, &mcp.Tool{Name: "upload_icon", Description: "Upload a custom icon from a local file or URL"}, uploadIconHandler)
	mcp.AddTool(s, &mcp.Tool{Name: "remove_track", Description: "Remove a track from a playlist"}, removeTrackHandler)
	mcp.AddTool(s, &mcp.Tool{Name: "move_track", Description: "Move or reorder a track"}, moveTrackHandler)
	mcp.AddTool(s, &mcp.Tool{Name: "copy_track", Description: "Copy a track to another playlist"}, copyTrackHandler)
	mcp.AddTool(s, &mcp.Tool{Name: "set_volume", Description: "Set the volume of a player (0-100)"}, setVolumeHandler)
	mcp.AddTool(s, &mcp.Tool{Name: "play_card", Description: "Start playing a playlist on a device"}, playCardHandler)
	mcp.AddTool(s, &mcp.Tool{Name: "stop_player", Description: "Stop playback on a device"}, stopPlayerHandler)
	mcp.AddTool(s, &mcp.Tool{Name: "pause_player", Description: "Pause playback on a device"}, pausePlayerHandler)
}
