package cmd

import (
	"fmt"

	"github.com/vgaro/yotocli/internal/config"
	"github.com/vgaro/yotocli/pkg/yoto"
	"github.com/spf13/cobra"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with Yoto",
	Long:  `Initiates the device code flow to authenticate this CLI with your Yoto account.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Use a temporary client for auth (no token needed yet)
		clientID := config.GetClientID()
		if clientID == "" {
			fmt.Println("Yoto Client ID not found in config.")
			fmt.Println("Get your Client ID from: https://dashboard.yoto.dev/")
			fmt.Print("Please enter your Client ID: ")
			fmt.Scanln(&clientID)
			if clientID == "" {
				return fmt.Errorf("client ID is required to authenticate")
			}
		}

		client := yoto.NewClient("", clientID)

		fmt.Println("Starting authentication...")
		authData, err := client.StartDeviceAuth()
		if err != nil {
			return err
		}

		fmt.Printf("\nPlease open this URL in your browser:\n  %s\n", authData.VerificationURI)
		if authData.VerificationURIComplete != "" {
			fmt.Printf("Or click here:\n  %s\n", authData.VerificationURIComplete)
		}
		fmt.Printf("\nEnter code: %s\n\n", authData.UserCode)
		fmt.Println("Waiting for you to authorize...")

		tokenResp, err := client.PollToken(authData.DeviceCode, authData.Interval)
		if err != nil {
			return err
		}

		fmt.Println("Successfully authenticated!")

		// Save tokens and client ID
		config.SetToken(tokenResp.AccessToken, tokenResp.RefreshToken)
		config.SetClientID(clientID)
		
		if err := config.Save(); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		fmt.Println("Credentials saved to config file.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
}
