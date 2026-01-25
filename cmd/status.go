package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check the status of your Yoto players",
	Long:  `Lists all Yoto players associated with your account, showing battery level, charging status, and what is currently playing.`,
	Example: `  # Check status of all players
  yoto status`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Fetching devices...")
	
devices, err := apiClient.ListDevices()
		if err != nil {
			return err
		}

		if len(devices) == 0 {
			fmt.Println("No devices found.")
			return nil
		}

		// Fetch status for each device in parallel
		g := new(errgroup.Group)
		for i := range devices {
			i := i // capture
			g.Go(func() error {
				if !devices[i].Online {
					return nil // Skip offline devices or handle differently
				}
				status, err := apiClient.GetDeviceStatus(devices[i].ID)
				if err != nil {
					// Don't fail the whole command if one device fails
					fmt.Printf("Warning: Failed to fetch status for %s: %v\n", devices[i].Name, err)
					return nil
				}
				devices[i].Status = status
				return nil
			})
		}

		if err := g.Wait(); err != nil {
			return err
		}

		// Print Table
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "Name\tStatus\tBattery\tVolume\tPlaying")
		
		for _, d := range devices {
			onlineStr := "Offline"
			if d.Online {
				onlineStr = "Online"
			}

			batteryStr := "-"
			volumeStr := "-"
			playingStr := "-"

			if d.Status != nil {
				charging := ""
				if d.Status.IsCharging == 1 {
					charging = "⚡ "
				}
				batteryStr = fmt.Sprintf("%d%%%s", d.Status.BatteryLevel, charging)
				volumeStr = fmt.Sprintf("%d", d.Status.Volume)
				
				if d.Status.ActiveCard != "none" && d.Status.ActiveCard != "" {
					playingStr = d.Status.ActiveCard // Ideally we'd resolve this to a Title
				} else {
					playingStr = "Idle"
				}
			}

			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", d.Name, onlineStr, batteryStr, volumeStr, playingStr)
		}
		w.Flush()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
