package cmd

import (
	"fmt"

	"portbridge/internal/config"
	"portbridge/internal/tunnel"
	"portbridge/internal/ui"

	"github.com/spf13/cobra"
)

// upCmd starts tunnels for a given profile
var upCmd = &cobra.Command{
	Use:   "up [profile]",
	Short: "Start tunnels for a profile",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		profileName := args[0]
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

		ui.PrintInfo("Connecting profile: " + profileName)

		manager := tunnel.NewTunnelManager()
		activeRecords, err := manager.ListProfileTunnels(profileName)
		if err != nil {
			ui.PrintError("Failed to load tunnel state: " + err.Error())
			return
		}

		desired := make(map[string]config.Tunnel)
		for _, t := range profile.Tunnels {
			if t.Enabled {
				desired[t.Name] = t
			}
		}

		for _, record := range activeRecords {
			if _, ok := desired[record.Name]; !ok {
				ui.PrintWarning("Stopping tunnel " + record.Name + " because it is disabled or removed from config")
				if err := manager.StopTunnel(profileName, record.Name); err != nil {
					ui.PrintError("Failed to stop tunnel " + record.Name + ": " + err.Error())
				}
			}
		}

		for _, t := range profile.Tunnels {
			if !t.Enabled {
				ui.PrintWarning("Tunnel " + t.Name + " is disabled")
				continue
			}

			target := profile.Host
			if profile.SSHAlias != "" {
				target = profile.SSHAlias
			}

			ui.PrintLog(fmt.Sprintf("Opening tunnel %s: localhost:%d -> %s:%d", t.Name, t.Local, target, t.Remote))

			record := tunnel.TunnelRecord{
				Profile:    profileName,
				Name:       t.Name,
				SSHAlias:   profile.SSHAlias,
				Host:       profile.Host,
				User:       profile.User,
				Port:       profile.Port,
				LocalPort:  t.Local,
				RemotePort: t.Remote,
			}

			sshCmd := tunnel.BuildSSHCommand(profile.SSHAlias, profile.User, profile.Host, profile.Port, t.Local, t.Remote)
			result, err := manager.StartTunnel(record, sshCmd)
			if err != nil {
				ui.PrintError("Failed to start tunnel " + t.Name + ": " + err.Error())
				continue
			}

			if result == "already-running" {
				ui.PrintWarning("Tunnel " + t.Name + " is already running")
				continue
			}

			ui.PrintSuccess("Started tunnel " + t.Name)
		}
	},
}

func init() {
	RootCmd.AddCommand(upCmd)
}
