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
		config, err := config.LoadConfig("portbridge.yaml")
		if err != nil {
			ui.PrintError("Failed to load configuration: " + err.Error())
			return
		}

		profile, exists := (*config)[profileName]
		if !exists {
			ui.PrintError("Profile not found: " + profileName)
			return
		}

		manager := tunnel.NewTunnelManager()
		for _, t := range profile.Tunnels {
			if t.Enabled {
				err := manager.StopTunnel(t.Name)
				if err != nil {
					ui.PrintError("Failed to stop tunnel " + t.Name + ": " + err.Error())
				} else {
					ui.PrintSuccess("Stopped tunnel " + t.Name)
				}
			}
		}
	},
}

func init() {
	RootCmd.AddCommand(downCmd)
}
