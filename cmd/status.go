package cmd

import (
	"strconv"

	"portbridge/internal/config"
	"portbridge/internal/tunnel"
	"portbridge/internal/ui"

	"github.com/spf13/cobra"
)

// statusCmd shows the status of tunnels for a given profile
var statusCmd = &cobra.Command{
	Use:   "status [profile]",
	Short: "Show the status of tunnels for a profile",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		profileName := args[0]
		cfg, err := config.LoadConfig("portbridge.yaml")
		if err != nil {
			ui.PrintError("Failed to load configuration: " + err.Error())
			return
		}

		profile, exists := (*cfg)[profileName]
		if !exists {
			ui.PrintError("Profile not found: " + profileName)
			return
		}

		manager := tunnel.NewTunnelManager()
		activeRecords, err := manager.ListProfileTunnels(profileName)
		if err != nil {
			ui.PrintError("Failed to load tunnel state: " + err.Error())
			return
		}

		running := make(map[string]tunnel.TunnelRecord)
		for _, record := range activeRecords {
			running[record.Name] = record
		}

		for _, t := range profile.Tunnels {
			status := "inactive"
			if record, ok := running[t.Name]; ok {
				status = "active"
				ui.PrintInfo(t.Name + ": " + status + " (pid " + strconv.Itoa(record.PID) + ")")
				continue
			}
			ui.PrintInfo(t.Name + ": " + status)
		}
	},
}

func init() {
	RootCmd.AddCommand(statusCmd)
}
