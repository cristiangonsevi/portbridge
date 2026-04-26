package cmd

import (
	"portbridge/internal/config"
	"portbridge/internal/profiles"
	"portbridge/internal/ui"

	"github.com/spf13/cobra"
)

// removeCmd groups remove subcommands.
var removeCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove resources from a profile",
}

var removeTunnelCmd = &cobra.Command{
	Use:   "tunnel [profile] [name]",
	Short: "Remove a tunnel from a profile",
	Args:  cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		profileName := args[0]
		name, _ := cmd.Flags().GetString("name")

		if len(args) == 2 {
			if name != "" && name != args[1] {
				ui.PrintError("Tunnel name provided in args and --name does not match")
				return
			}
			name = args[1]
		}

		if name == "" {
			ui.PrintError("Tunnel name is required. Use 'portbridge remove tunnel <profile> <name>' or '--name <name>'")
			return
		}

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

		found := false
		for _, existingTunnel := range profile.Tunnels {
			if existingTunnel.Name == name {
				found = true
				break
			}
		}

		if !found {
			ui.PrintError("Tunnel not found: " + name)
			return
		}

		profiles.RemoveTunnel(&profile, name)
		(*cfg)[profileName] = profile

		err = config.SaveConfig(config.ConfigFilePath, cfg)
		if err != nil {
			ui.PrintError("Failed to save configuration: " + err.Error())
			return
		}

		ui.PrintSuccess("Removed tunnel " + name + " from profile " + profileName)
	},
}

func init() {
	removeTunnelCmd.Flags().String("name", "", "Tunnel name")

	removeCmd.AddCommand(removeTunnelCmd)
	RootCmd.AddCommand(removeCmd)
}
