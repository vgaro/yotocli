package cmd

import (
	"fmt"
	"strings"

	"github.com/vgaro/yotocli/internal/utils"
	"github.com/spf13/cobra"
)

var playCmd = &cobra.Command{
	Use:   "play <playlist> [device]",
	Short: "Play a playlist on a Yoto player",
	Args:  cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		playlistName := args[0]

		// Find Playlist
	
cards, err := apiClient.ListCards()
		if err != nil {
			return err
		}
		card := utils.FindCard(cards, playlistName)
		if card == nil {
			return fmt.Errorf("playlist '%s' not found", playlistName)
		}

		// Find Device
	
devices, err := apiClient.ListDevices()
		if err != nil {
			return err
		}
		if len(devices) == 0 {
			return fmt.Errorf("no devices found")
		}

		var targetDeviceID string
		if len(args) == 2 {
			query := strings.ToLower(args[1])
			for _, d := range devices {
				if strings.Contains(strings.ToLower(d.Name), query) {
					                    					targetDeviceID = d.ID
					                    					break
				}
			}
			if targetDeviceID == "" {
				return fmt.Errorf("device '%s' not found", args[1])
			}
		} else {
			targetDeviceID = devices[0].ID
			if len(devices) > 1 {
				fmt.Printf("Multiple devices found. Using '%s' (%s).\n", devices[0].Name, targetDeviceID)
			}
		}

		fmt.Printf("Playing '%s' on device %s...\n", card.Title, targetDeviceID)
		return apiClient.PlayCard(targetDeviceID, card.CardID)
	},
}

var stopCmd = &cobra.Command{
	Use:   "stop [device]",
	Short: "Stop playback on a Yoto player",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Find Device
	
devices, err := apiClient.ListDevices()
		if err != nil {
			return err
		}
		if len(devices) == 0 {
			return fmt.Errorf("no devices found")
		}

		var targetDeviceID string
		if len(args) == 1 {
			query := strings.ToLower(args[0])
			for _, d := range devices {
				if strings.Contains(strings.ToLower(d.Name), query) {
					targetDeviceID = d.ID
					break
				}
			}
			if targetDeviceID == "" {
				return fmt.Errorf("device '%s' not found", args[0])
			}
		} else {
			targetDeviceID = devices[0].ID
		}

		fmt.Printf("Stopping playback on device %s...\n", targetDeviceID)
		return apiClient.StopPlayer(targetDeviceID)
	},
}

var pauseCmd = &cobra.Command{
	Use:   "pause [device]",
	Short: "Pause playback on a Yoto player",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Find Device
	
devices, err := apiClient.ListDevices()
		if err != nil {
			return err
		}
		if len(devices) == 0 {
			return fmt.Errorf("no devices found")
		}

		var targetDeviceID string
		if len(args) == 1 {
			query := strings.ToLower(args[0])
			for _, d := range devices {
				if strings.Contains(strings.ToLower(d.Name), query) {
					targetDeviceID = d.ID
					break
				}
			}
			if targetDeviceID == "" {
				return fmt.Errorf("device '%s' not found", args[0])
			}
		} else {
			targetDeviceID = devices[0].ID
		}

		fmt.Printf("Pausing playback on device %s...\n", targetDeviceID)
		return apiClient.PausePlayer(targetDeviceID)
	},
}


func init() {
	rootCmd.AddCommand(playCmd)
	rootCmd.AddCommand(stopCmd)
	rootCmd.AddCommand(pauseCmd)
}
