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
		cfg, err := config.LoadConfig("~/.config/.portbridge/portbridge.yaml")
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

		for _, t := range profile.Tunnels {
			pid, err := manager.FindActiveTunnelPID(t.Local)
			if err != nil {
				ui.PrintError("Failed to inspect tunnel " + t.Name + ": " + err.Error())
				continue
			}

			if pid > 0 {
				ui.PrintInfo(t.Name + ": active (pid " + strconv.Itoa(pid) + ")")
				continue
			}

			ui.PrintInfo(t.Name + ": inactive")
		}
	},
}

func init() {
	RootCmd.AddCommand(statusCmd)
}
