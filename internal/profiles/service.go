package profiles

import (
	"portbridge/internal/config"
)

// AddTunnel adds a new tunnel to a profile
func AddTunnel(profile *config.Profile, tunnel config.Tunnel) {
	profile.Tunnels = append(profile.Tunnels, tunnel)
}

// RemoveTunnel removes a tunnel from a profile by name
func RemoveTunnel(profile *config.Profile, tunnelName string) {
	for i, tunnel := range profile.Tunnels {
		if tunnel.Name == tunnelName {
			profile.Tunnels = append(profile.Tunnels[:i], profile.Tunnels[i+1:]...)
			break
		}
	}
}

// EnableTunnel enables a tunnel by name
func EnableTunnel(profile *config.Profile, tunnelName string) {
	for i := range profile.Tunnels {
		if profile.Tunnels[i].Name == tunnelName {
			profile.Tunnels[i].Enabled = true
			break
		}
	}
}

// DisableTunnel disables a tunnel by name
func DisableTunnel(profile *config.Profile, tunnelName string) {
	for i := range profile.Tunnels {
		if profile.Tunnels[i].Name == tunnelName {
			profile.Tunnels[i].Enabled = false
			break
		}
	}
}
