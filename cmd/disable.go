package cmd

import (
	"portbridge/internal/config"
	"portbridge/internal/profiles"
	"portbridge/internal/ui"

	"github.com/spf13/cobra"
)

// disableCmd groups disable subcommands.
var disableCmd = &cobra.Command{
	Use:   "disable",
	Short: "Disable resources in a profile",
}

var disableTunnelCmd = &cobra.Command{
	Use:   "tunnel [profile]",
	Short: "Disable a tunnel in a profile",
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

		profiles.DisableTunnel(&profile, name)
		(*cfg)[profileName] = profile

		err = config.SaveConfig("portbridge.yaml", cfg)
		if err != nil {
			ui.PrintError("Failed to save configuration: " + err.Error())
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
