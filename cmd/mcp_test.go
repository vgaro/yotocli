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
}
