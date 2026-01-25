package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	// 1. Path to our built binary
	// Assuming we run 'go run examples/mcp_client/main.go' from project root
	cwd, _ := os.Getwd()
	fmt.Printf("Current working directory: %s\n", cwd)
	
binaryPath := filepath.Join(cwd, "yoto")
	
	// Check if it exists
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		log.Fatalf("yoto binary not found at %s. Did you build it?", binaryPath)
	}
	
	// 2. Create the Client
	client := mcp.NewClient(&mcp.Implementation{
		Name:    "example-client",
		Version: "1.0.0",
	}, nil)

	// 3. Connect using CommandTransport
	transport := &mcp.CommandTransport{
		Command: exec.Command(binaryPath, "mcp"),
	}
	
	fmt.Println("Starting MCP Client...")
	ctx := context.Background()
	
	// Connect returns a ClientSession
	session, err := client.Connect(ctx, transport, nil)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer session.Close()

	// 5. List Tools
	listToolsResult, err := session.ListTools(ctx, &mcp.ListToolsParams{})
	if err != nil {
		log.Fatalf("Failed to list tools: %v", err)
	}

	fmt.Printf("\nFound %d tools:\n", len(listToolsResult.Tools))
	for _, t := range listToolsResult.Tools {
		fmt.Printf("- %s: %s\n", t.Name, t.Description)
	}

	// 6. Call a Tool (list_playlists)
	fmt.Println("\nCalling 'list_playlists'...")
	
	callResult, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name: "list_playlists",
		Arguments: map[string]interface{}{},
	})
	if err != nil {
		log.Fatalf("Failed to call tool: %v", err)
	}

	// Output the result
	if len(callResult.Content) > 0 {
		for _, c := range callResult.Content {
			if tc, ok := c.(*mcp.TextContent); ok {
				fmt.Println(tc.Text)
			} else {
				fmt.Printf("Non-text content: %T\n", c)
			}
		}
	} else {
		fmt.Println("(No content returned)")
	}
}
