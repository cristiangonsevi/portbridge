package tunnel

import (
	"fmt"
	"time"

	"portbridge/internal/config"
	"portbridge/internal/ui"
	tunnelssh "portbridge/internal/tunnel/ssh"
)

// ReconnectManager handles auto-reconnection of tunnels.
type ReconnectManager struct {
	manager *TunnelManager
	stopCh  chan struct{}
}

// NewReconnectManager creates a new ReconnectManager.
func NewReconnectManager(manager *TunnelManager) *ReconnectManager {
	return &ReconnectManager{
		manager: manager,
		stopCh:  make(chan struct{}),
	}
}

// MonitorTunnels periodically checks all tracked tunnels and restarts any that
// have stopped unexpectedly. It runs until the stop channel is closed.
func (rm *ReconnectManager) MonitorTunnels(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-rm.stopCh:
			ui.PrintLog("Reconnect monitor stopped")
			return
		case <-ticker.C:
			rm.checkAndRestartTunnels()
		}
	}
}

// Stop stops the reconnect monitor.
func (rm *ReconnectManager) Stop() {
	close(rm.stopCh)
}

// MonitorTunnelsLegacy is a no-op placeholder kept for API compatibility.
// Use MonitorTunnels with a stop channel instead.
func (rm *ReconnectManager) MonitorTunnelsLegacy(interval time.Duration) {
	_ = rm
	_ = interval
}

func (rm *ReconnectManager) checkAndRestartTunnels() {
	cfg, err := config.LoadConfig(config.ConfigFilePath)
	if err != nil {
		ui.PrintError(fmt.Sprintf("Reconnect: failed to load config: %v", err))
		return
	}

	for profileName, profile := range *cfg {
		interval := profile.ReconnectInterval
		if interval <= 0 {
			interval = 30
		}

		for _, t := range profile.Tunnels {
			if !t.Enabled {
				continue
			}

			pid, err := rm.manager.FindActiveTunnelPID(t.Local)
			if err != nil {
				ui.PrintError(fmt.Sprintf("Reconnect: failed to check tunnel %s/%s: %v", profileName, t.Name, err))
				continue
			}

			if pid > 0 {
				continue
			}

			ui.PrintWarning(fmt.Sprintf("Reconnect: tunnel %s/%s is down, restarting...", profileName, t.Name))

			sshClient, err := tunnelssh.NewClient(&profile)
			if err != nil {
				ui.PrintError(fmt.Sprintf("Reconnect: failed to create SSH client for %s/%s: %v", profileName, t.Name, err))
				continue
			}

			record := TunnelRecord{
				Profile:    profileName,
				Name:       t.Name,
				SSHAlias:   profile.SSHAlias,
				Host:       profile.Host,
				User:       profile.User,
				Port:       profile.Port,
				LocalPort:  t.Local,
				RemotePort: t.Remote,
			}

			_, err = rm.manager.StartSSHTunnel(record, sshClient.UnderlyingClient(), t.Local, t.Remote)
			sshClient.Close()
			if err != nil {
				ui.PrintError(fmt.Sprintf("Reconnect: failed to restart tunnel %s/%s: %v", profileName, t.Name, err))
			} else {
				ui.PrintSuccess(fmt.Sprintf("Reconnect: restarted tunnel %s/%s", profileName, t.Name))
			}
		}
	}
}
