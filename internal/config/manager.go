package config

import (
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

const (
	KeyAccessToken  = "auth.access_token"
	KeyRefreshToken = "auth.refresh_token"
	KeyExpiresAt    = "auth.expires_at"
	KeyClientID     = "auth.client_id"
)

// configFileMode restricts the config file to its owner. It holds the OAuth
// access and refresh tokens in plaintext, so it must not be world-readable.
const configFileMode = 0600

// Save persists the current viper configuration to disk
func Save() error {
	// If no config file is used (first run), create one
	if viper.ConfigFileUsed() == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		configPath := filepath.Join(home, ".config", "yotocli")
		if err := os.MkdirAll(configPath, 0755); err != nil {
			return err
		}
		viper.SetConfigFile(filepath.Join(configPath, "config.yaml"))
	}

	// Viper defaults to 0644, which would leave the tokens readable by every
	// user on the machine.
	viper.SetConfigPermissions(configFileMode)

	if err := viper.WriteConfig(); err != nil {
		return err
	}

	// Viper only applies the permissions when it creates the file, so configs
	// written before this change keep their original mode. Tighten them here.
	return os.Chmod(viper.ConfigFileUsed(), configFileMode)
}

func SetToken(access, refresh string) {
	viper.Set(KeyAccessToken, access)
	viper.Set(KeyRefreshToken, refresh)
}

func SetClientID(clientID string) {
	viper.Set(KeyClientID, clientID)
}

func GetRefreshToken() string {
	return viper.GetString(KeyRefreshToken)
}

func GetAccessToken() string {
	return viper.GetString(KeyAccessToken)
}

func GetClientID() string {
	return viper.GetString(KeyClientID)
}