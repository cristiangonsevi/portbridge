package cmd

import (
	"fmt"

	"portbridge/internal/config"
	"portbridge/internal/profiles"
	"portbridge/internal/tunnel"
	tunnelssh "portbridge/internal/tunnel/ssh"
	"portbridge/internal/ui"

	"github.com/spf13/cobra"
)

// enableCmd groups enable subcommands.
var enableCmd = &cobra.Command{
	Use:   "enable",
	Short: "Enable resources in a profile",
}

var enableTunnelCmd = &cobra.Command{
	Use:               "tunnel [profile]",
	Short:             "Enable a tunnel in a profile",
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

		err = config.SaveConfig(config.ConfigFilePath, cfg)
		if err != nil {
			ui.PrintError("Failed to save configuration: " + err.Error())
			return
		}

		target := profile.Host
		if profile.SSHAlias != "" {
			target = profile.SSHAlias
		}

		manager := tunnel.NewTunnelManager()

		sshClient, err := tunnelssh.NewClient(&profile)
		if err != nil {
			ui.PrintError("Failed to create SSH client: " + err.Error())
			return
		}
		defer sshClient.Close()

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

		result, err := manager.StartSSHTunnel(record, sshClient.UnderlyingClient(), selectedTunnel.Local, selectedTunnel.Remote)
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
