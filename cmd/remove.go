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
	Use:   "tunnel [profile]",
	Short: "Remove a tunnel from a profile",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		profileName := args[0]
		name, _ := cmd.Flags().GetString("name")

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

		profiles.RemoveTunnel(&profile, name)
		(*cfg)[profileName] = profile

		err = config.SaveConfig("portbridge.yaml", cfg)
		if err != nil {
			ui.PrintError("Failed to save configuration: " + err.Error())
			return
		}

		ui.PrintSuccess("Removed tunnel " + name + " from profile " + profileName)
	},
}

func init() {
	removeTunnelCmd.Flags().String("name", "", "Tunnel name")
	removeTunnelCmd.MarkFlagRequired("name")

	removeCmd.AddCommand(removeTunnelCmd)
	RootCmd.AddCommand(removeCmd)
}
