package cmd

import (
	"portbridge/internal/config"
	"portbridge/internal/profiles"
	"portbridge/internal/ui"

	"github.com/spf13/cobra"
)

// enableCmd groups enable subcommands.
var enableCmd = &cobra.Command{
	Use:   "enable",
	Short: "Enable resources in a profile",
}

var enableTunnelCmd = &cobra.Command{
	Use:   "tunnel [profile]",
	Short: "Enable a tunnel in a profile",
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

		profiles.EnableTunnel(&profile, name)
		(*cfg)[profileName] = profile

		err = config.SaveConfig("portbridge.yaml", cfg)
		if err != nil {
			ui.PrintError("Failed to save configuration: " + err.Error())
			return
		}

		ui.PrintSuccess("Enabled tunnel " + name + " in profile " + profileName)
	},
}

func init() {
	enableTunnelCmd.Flags().String("name", "", "Tunnel name")
	enableTunnelCmd.MarkFlagRequired("name")

	enableCmd.AddCommand(enableTunnelCmd)
	RootCmd.AddCommand(enableCmd)
}
