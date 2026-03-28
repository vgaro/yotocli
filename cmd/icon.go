package cmd

import (
        "fmt"
        "strings"

        "github.com/spf13/cobra"
        "github.com/vgaro/yotocli/internal/actions"
        "github.com/vgaro/yotocli/internal/config"
)
var iconCmd = &cobra.Command{
	Use:   "icon",
	Short: "Manage icons",
}

var (
        uploadIconName string
        uploadIconTags []string
)

var uploadIconCmd = &cobra.Command{
        Use:   "upload <file_or_url>",
        Short: "Upload a custom icon (local file or URL)",
        Args:  cobra.ExactArgs(1),
        RunE: func(cmd *cobra.Command, args []string) error {
                source := args[0]
                fmt.Printf("Uploading icon from %s...\n", source)

                id, err := actions.UploadIcon(apiClient, source, uploadIconName, uploadIconTags)
                if err != nil {
                        return err
                }

                fmt.Printf("Icon uploaded successfully!\nID: %s\n", id)
                fmt.Printf("Use this ID with 'yoto edit' or 'yoto icon set'.\n")
                return nil
        },
}

var lsIconCmd = &cobra.Command{
        Use:   "ls",
        Short: "List all custom icons",
        RunE: func(cmd *cobra.Command, args []string) error {
                icons, err := actions.ListIcons(apiClient)
                if err != nil {
                        return err
                }

                if len(icons) == 0 {
                        fmt.Println("No custom icons found.")
                        return nil
                }

                // Load local mapping
                localIcons, _ := config.LoadIcons()
                nameMap := make(map[string]string)
                if localIcons != nil {
                        for _, li := range localIcons.Icons {
                                nameMap[li.ID] = li.Name
                        }
                }

                fmt.Printf("%-15s %-45s %-25s\n", "Name", "ID", "Created At")
                fmt.Println(strings.Repeat("-", 85))
                for _, icon := range icons {
                        name := nameMap[icon.MediaID]
                        if name == "" {
                                name = "-"
                        }
                        fmt.Printf("%-15s %-45s %-25s\n", name, icon.MediaID, icon.CreatedAt.Format("2006-01-02 15:04:05"))
                }
                return nil
        },
}

var searchIconCmd = &cobra.Command{
        Use:   "search <query>",
        Short: "Search for a custom icon by name or tag",
        Args:  cobra.ExactArgs(1),
        RunE: func(cmd *cobra.Command, args []string) error {
                query := args[0]
                results, err := actions.SearchIcons(query)
                if err != nil {
                        return err
                }

                if len(results) == 0 {
                        fmt.Printf("No icons found matching '%s'.\n", query)
                        return nil
                }

                fmt.Printf("%-15s %-45s %-20s\n", "Name", "ID", "Tags")
                fmt.Println(strings.Repeat("-", 85))
                for _, icon := range results {
                        fmt.Printf("%-15s %-45s %-20s\n", icon.Name, icon.ID, strings.Join(icon.Tags, ","))
                }
                return nil
        },
}

func init() {
        uploadIconCmd.Flags().StringVarP(&uploadIconName, "name", "n", "", "Friendly name for the icon")
        uploadIconCmd.Flags().StringSliceVarP(&uploadIconTags, "tags", "t", []string{}, "Tags for the icon")

        iconCmd.AddCommand(uploadIconCmd)
        iconCmd.AddCommand(lsIconCmd)
        iconCmd.AddCommand(searchIconCmd)
        rootCmd.AddCommand(iconCmd)
}

