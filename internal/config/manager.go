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
	return viper.WriteConfig()
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