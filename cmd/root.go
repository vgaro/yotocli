package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/vgaro/yotocli/internal/config"
	"github.com/vgaro/yotocli/pkg/yoto"
)

var (
	cfgFile   string
	apiClient *yoto.Client
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "yoto",
	Short: "A CLI tool for managing Yoto cards and players",
	Long: `YotoCLI is a tool for advanced users to manage their Yoto library.
It allows for uploading files, creating playlists, and managing device state directly from the terminal.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Initialize the API client with the token from config
		token := config.GetAccessToken()
		clientID := config.GetClientID()
		apiClient = yoto.NewClient(token, clientID)

		// Check if token is valid by making a lightweight call
		// If unauthorized, try to refresh
		if token != "" {
			_, err := apiClient.ListDevices()
			if err != nil && (strings.Contains(err.Error(), "unauthorized") || strings.Contains(err.Error(), "401")) {
				// fmt.Println("Token expired, attempting refresh...")
				refreshToken := config.GetRefreshToken()
				if refreshToken != "" {
					newTokens, refreshErr := apiClient.RefreshToken(refreshToken)
					if refreshErr == nil {
						config.SetToken(newTokens.AccessToken, newTokens.RefreshToken)
						if err := config.Save(); err == nil {
							// Re-init client with new token
							apiClient = yoto.NewClient(newTokens.AccessToken, clientID)
							// fmt.Println("Token refreshed.")
						}
					}
				}
			}
		}

		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// RootCmd returns the root command for doc generation
func RootCmd() *cobra.Command {
	return rootCmd
}

func init() {
	cobra.OnInitialize(initConfig)

	// Persistent flags (available to all commands)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.config/yotocli/config.yaml)")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in ~/.config/yotocli directory
		configPath := filepath.Join(home, ".config", "yotocli")
		viper.AddConfigPath(configPath)
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		// fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
