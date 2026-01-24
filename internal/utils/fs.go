package utils

import (
	"os"
	"strings"
)

// SanitizeFilename removes characters that are illegal in filenames on Windows/Linux/Mac
func SanitizeFilename(name string) string {
	return strings.Map(func(r rune) rune {
		// Common illegal characters: < > : " / \ | ? *
		if strings.ContainsRune(`<>:"/\|?*`, r) {
			return -1
		}
		// Control characters
		if r < 32 {
			return -1
		}
		return r
	}, name)
}

// IsDir checks if a path exists and is a directory
func IsDir(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
