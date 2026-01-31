package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/vgaro/yotocli/internal/actions"
)

var iconCmd = &cobra.Command{
	Use:   "icon",
	Short: "Manage icons",
}

var uploadIconCmd = &cobra.Command{
	Use:   "upload <file_or_url>",
	Short: "Upload a custom icon (local file or URL)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		source := args[0]
		fmt.Printf("Uploading icon from %s...\n", source)
		
		id, err := actions.UploadIcon(apiClient, source)
		if err != nil {
			return err
		}
		
		fmt.Printf("Icon uploaded successfully!\nID: %s\n", id)
		fmt.Printf("Use this ID with 'yoto edit' or 'yoto icon set'.\n")
		return nil
	},
}

func init() {
	iconCmd.AddCommand(uploadIconCmd)
	rootCmd.AddCommand(iconCmd)
}
