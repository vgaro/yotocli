package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var iconCmd = &cobra.Command{
	Use:   "icon",
	Short: "Manage icons",
}

var uploadIconCmd = &cobra.Command{
	Use:   "upload <file>",
	Short: "Upload a custom icon",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := args[0]
		fmt.Printf("Uploading icon %s...\n", path)
		
		id, err := apiClient.UploadIcon(path)
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
