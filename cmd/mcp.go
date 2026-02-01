package cmd

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/cobra"
)

var (
	mcpTransport string
	mcpPort      int
	mcpAddr      string
)

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
		mcp.AddTool(s, &mcp.Tool{Name: "remove_track", Description: "Remove a track from a playlist"}, removeTrackHandler)
		mcp.AddTool(s, &mcp.Tool{Name: "move_track", Description: "Reorder a track within a playlist"}, moveTrackHandler)
		mcp.AddTool(s, &mcp.Tool{Name: "set_volume", Description: "Set the volume of a player (0-100)"}, setVolumeHandler)
		mcp.AddTool(s, &mcp.Tool{Name: "play_card", Description: "Start playing a playlist on a device"}, playCardHandler)
		mcp.AddTool(s, &mcp.Tool{Name: "stop_player", Description: "Stop playback on a device"}, stopPlayerHandler)
		mcp.AddTool(s, &mcp.Tool{Name: "pause_player", Description: "Pause playback on a device"}, pausePlayerHandler)

		// Start Server
		if mcpTransport == "sse" {
			// Create SSE Handler
			handler := mcp.NewSSEHandler(func(r *http.Request) *mcp.Server {
				return s
			}, &mcp.SSEOptions{})

			mux := http.NewServeMux()
			mux.Handle("/sse", handler)
			// The handler manages its own message endpoints, we just need to route base requests
			// But usually we need to mount it to a path.
			
			mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("OK"))
			})

			server := &http.Server{
				Addr:    fmt.Sprintf("%s:%d", mcpAddr, mcpPort),
				Handler: mux,
			}

			fmt.Printf("Starting MCP SSE server on http://%s:%d/sse\n", mcpAddr, mcpPort)

			// Start HTTP server in goroutine
			go func() {
				if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					log.Fatalf("HTTP server failed: %v", err)
				}
			}()

			// Wait for interrupt signal
			stop := make(chan os.Signal, 1)
			signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
			<-stop

			fmt.Println("\nShutting down server...")
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := server.Shutdown(ctx); err != nil {
				log.Fatalf("Server forced to shutdown: %v", err)
			}
			fmt.Println("Server exited properly")
		} else {
			// Default Stdio
			if err := s.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
				log.Fatalf("Server failed: %v", err)
			}
		}
	},
}

func init() {
	mcpCmd.Flags().StringVar(&mcpTransport, "transport", "stdio", "Transport type: stdio or sse")
	mcpCmd.Flags().IntVar(&mcpPort, "port", 8080, "Port for SSE server")
	mcpCmd.Flags().StringVar(&mcpAddr, "addr", "0.0.0.0", "Bind address for SSE server")
	rootCmd.AddCommand(mcpCmd)
}