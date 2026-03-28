package actions

import (
        "os"
        "path/filepath"
        "strings"

        "github.com/vgaro/yotocli/internal/config"
        "github.com/vgaro/yotocli/pkg/yoto"
)

// UploadIcon uploads an icon from a local path or URL.
// Returns the new Icon ID.
func UploadIcon(client *yoto.Client, source string, name string, tags []string) (string, error) {
        path := source
        if strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://") {
                tmpDir := os.TempDir()
                path = filepath.Join(tmpDir, "yoto_icon_temp.png") // Assume PNG, or API detects type?
                // Note: DownloadFile is in client.go
                if err := client.DownloadFile(source, path); err != nil {
                        return "", err
                }
                defer os.Remove(path)
        }

        id, err := client.UploadIcon(path)
        if err != nil {
                return "", err
        }

        if name != "" {
                if err := config.AddIconRecord(id, name, tags); err != nil {
                        // Log warning but don't fail upload
                        return id, nil
                }
        }

        return id, nil
}

// ListIcons returns all user-uploaded icons.
func ListIcons(client *yoto.Client) ([]yoto.DisplayIcon, error) {
        return client.ListIcons()
}

// SearchIcons searches the local icon registry.
func SearchIcons(query string) ([]config.IconRecord, error) {
        cfg, err := config.LoadIcons()
        if err != nil {
                return nil, err
        }

        var results []config.IconRecord
        q := strings.ToLower(query)
        for _, icon := range cfg.Icons {
                if strings.Contains(strings.ToLower(icon.Name), q) {
                        results = append(results, icon)
                        continue
                }
                for _, tag := range icon.Tags {
                        if strings.Contains(strings.ToLower(tag), q) {
                                results = append(results, icon)
                                break
                        }
                }
        }
        return results, nil
}
