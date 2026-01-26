package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

var volumeCmd = &cobra.Command{
	Use:   "volume <level> [device]",
	Short: "Set the volume of a Yoto player",
	Long: `Set the volume of a Yoto player. Level should be between 0 and 100.
If no device is specified and you have multiple, it will ask or pick the first one.`,
	Args: cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		volStr := args[0]
		vol, err := strconv.Atoi(volStr)
		if err != nil || vol < 0 || vol > 100 {
			return fmt.Errorf("volume must be a number between 0 and 100")
		}

	
devices, err := apiClient.ListDevices()
		if err != nil {
			return err
		}
		if len(devices) == 0 {
			return fmt.Errorf("no devices found")
		}

		var targetDeviceID string
		if len(args) == 2 {
			// Find device by name (fuzzy)
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
			// Default to first device or error if multiple?
			// Let's just pick the first online one, or just the first one.
			targetDeviceID = devices[0].ID
			if len(devices) > 1 {
				fmt.Printf("Multiple devices found. Using '%s' (%s).\n", devices[0].Name, targetDeviceID)
			}
		}

		fmt.Printf("Setting volume to %d...\n", vol)
		return apiClient.SetVolume(targetDeviceID, vol)
	},
}

func init() {
	rootCmd.AddCommand(volumeCmd)
}
