package actions

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/vgaro/yotocli/pkg/yoto"
)

// UploadIcon uploads an icon from a local path or URL.
// Returns the new Icon ID.
func UploadIcon(client *yoto.Client, source string) (string, error) {
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

	return client.UploadIcon(path)
}
