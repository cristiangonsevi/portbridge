package cmd

import (
	"portbridge/internal/config"
	"portbridge/internal/tunnel"
	"portbridge/internal/ui"

	"github.com/spf13/cobra"
)

// downCmd stops tunnels for a given profile
var downCmd = &cobra.Command{
	Use:   "down [profile]",
	Short: "Stop tunnels for a profile",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		profileName := args[0]
		cfg, err := config.LoadConfig(config.ConfigFilePath)
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

		active := make(map[string]bool)
		for _, record := range activeRecords {
			active[record.Name] = true
		}

		for _, t := range profile.Tunnels {
			if !active[t.Name] {
				ui.PrintWarning("Tunnel " + t.Name + " is not running")
				continue
			}

			err := manager.StopTunnel(profileName, t.Name)
			if err != nil {
				ui.PrintError("Failed to stop tunnel " + t.Name + ": " + err.Error())
			} else {
				ui.PrintSuccess("Stopped tunnel " + t.Name)
			}
		}
	},
}

func init() {
	RootCmd.AddCommand(downCmd)
}
