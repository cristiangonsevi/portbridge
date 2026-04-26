package cmd

import (
	"portbridge/internal/config"
	"portbridge/internal/profiles"
	"portbridge/internal/ui"

	"github.com/spf13/cobra"
)

// addCmd groups add subcommands.
var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add resources to a profile",
}

var addTunnelCmd = &cobra.Command{
	Use:   "tunnel [profile]",
	Short: "Add a new tunnel to a profile",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		profileName := args[0]
		name, _ := cmd.Flags().GetString("name")
		local, _ := cmd.Flags().GetInt("local")
		remote, _ := cmd.Flags().GetInt("remote")

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

		newTunnel := config.Tunnel{
			Name:    name,
			Local:   local,
			Remote:  remote,
			Enabled: true,
		}

		profiles.AddTunnel(&profile, newTunnel)
		(*cfg)[profileName] = profile

		err = config.SaveConfig("~/.config/.portbridge/portbridge.yaml", cfg)
		if err != nil {
			ui.PrintError("Failed to save configuration: " + err.Error())
			return
		}

		ui.PrintSuccess("Added tunnel " + name + " to profile " + profileName)
	},
}

func init() {
	addTunnelCmd.Flags().String("name", "", "Tunnel name")
	addTunnelCmd.Flags().Int("local", 0, "Local port")
	addTunnelCmd.Flags().Int("remote", 0, "Remote port")
	addTunnelCmd.MarkFlagRequired("name")
	addTunnelCmd.MarkFlagRequired("local")
	addTunnelCmd.MarkFlagRequired("remote")

	addCmd.AddCommand(addTunnelCmd)
	RootCmd.AddCommand(addCmd)
}
