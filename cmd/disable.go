package cmd

import (
	"portbridge/internal/config"
	"portbridge/internal/profiles"
	"portbridge/internal/tunnel"
	"portbridge/internal/ui"

	"github.com/spf13/cobra"
)

// disableCmd groups disable subcommands.
var disableCmd = &cobra.Command{
	Use:   "disable",
	Short: "Disable resources in a profile",
}

var disableTunnelCmd = &cobra.Command{
	Use:               "tunnel [profile]",
	Short:             "Disable a tunnel in a profile",
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction:  profileNameCompletionFunc,
	Run: func(cmd *cobra.Command, args []string) {
		profileName := args[0]
		name, _ := cmd.Flags().GetString("name")

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

		localPort := 0
		found := false
		for _, existingTunnel := range profile.Tunnels {
			if existingTunnel.Name == name {
				localPort = existingTunnel.Local
				found = true
				break
			}
		}

		if !found {
			ui.PrintError("Tunnel not found: " + name)
			return
		}

		profiles.DisableTunnel(&profile, name)
		(*cfg)[profileName] = profile

		err = config.SaveConfig(config.ConfigFilePath, cfg)
		if err != nil {
			ui.PrintError("Failed to save configuration: " + err.Error())
			return
		}

		manager := tunnel.NewTunnelManager()
		pid, err := manager.FindActiveTunnelPID(localPort)
		if err != nil {
			ui.PrintError("Disabled tunnel in config, but failed to inspect runtime state: " + err.Error())
			return
		}

		if pid > 0 {
			if err := manager.StopTunnelByPort(localPort); err != nil {
				ui.PrintError("Disabled tunnel in config, but failed to stop running tunnel: " + err.Error())
				return
			}

			ui.PrintSuccess("Disabled tunnel " + name + " in profile " + profileName + " and stopped the active tunnel")
			return
		}

		ui.PrintSuccess("Disabled tunnel " + name + " in profile " + profileName)
	},
}

func init() {
	disableTunnelCmd.Flags().String("name", "", "Tunnel name")
	disableTunnelCmd.MarkFlagRequired("name")

	disableCmd.AddCommand(disableTunnelCmd)
	RootCmd.AddCommand(disableCmd)
}
