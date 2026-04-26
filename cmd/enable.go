package cmd

import (
	"fmt"

	"portbridge/internal/config"
	"portbridge/internal/profiles"
	"portbridge/internal/tunnel"
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

		selectedTunnel := config.Tunnel{}
		found := false
		for _, existingTunnel := range profile.Tunnels {
			if existingTunnel.Name == name {
				selectedTunnel = existingTunnel
				found = true
				break
			}
		}

		if !found {
			ui.PrintError("Tunnel not found: " + name)
			return
		}

		profiles.EnableTunnel(&profile, name)
		(*cfg)[profileName] = profile

		err = config.SaveConfig("portbridge.yaml", cfg)
		if err != nil {
			ui.PrintError("Failed to save configuration: " + err.Error())
			return
		}

		target := profile.Host
		if profile.SSHAlias != "" {
			target = profile.SSHAlias
		}

		manager := tunnel.NewTunnelManager()
		record := tunnel.TunnelRecord{
			Profile:    profileName,
			Name:       name,
			SSHAlias:   profile.SSHAlias,
			Host:       profile.Host,
			User:       profile.User,
			Port:       profile.Port,
			LocalPort:  selectedTunnel.Local,
			RemotePort: selectedTunnel.Remote,
		}

		ui.PrintLog(fmt.Sprintf("Opening tunnel %s: localhost:%d -> %s:%d", name, selectedTunnel.Local, target, selectedTunnel.Remote))

		sshCmd := tunnel.BuildSSHCommand(profile.SSHAlias, profile.User, profile.Host, profile.Port, selectedTunnel.Local, selectedTunnel.Remote)
		result, err := manager.StartTunnel(record, sshCmd)
		if err != nil {
			ui.PrintError("Enabled tunnel in config, but failed to start runtime tunnel: " + err.Error())
			return
		}

		if result == "already-running" {
			ui.PrintWarning("Enabled tunnel " + name + " in profile " + profileName + ", tunnel was already running")
			return
		}

		ui.PrintSuccess("Enabled tunnel " + name + " in profile " + profileName + " and started the active tunnel")
	},
}

func init() {
	enableTunnelCmd.Flags().String("name", "", "Tunnel name")
	enableTunnelCmd.MarkFlagRequired("name")

	enableCmd.AddCommand(enableTunnelCmd)
	RootCmd.AddCommand(enableCmd)
}
