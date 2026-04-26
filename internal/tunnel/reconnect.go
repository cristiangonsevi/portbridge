package tunnel

import (
	"log"
	"time"
)

// ReconnectManager handles auto-reconnection of tunnels
type ReconnectManager struct {
	manager *TunnelManager
}

// NewReconnectManager creates a new ReconnectManager
func NewReconnectManager(manager *TunnelManager) *ReconnectManager {
	return &ReconnectManager{
		manager: manager,
	}
}

// MonitorTunnels monitors tunnels and attempts to reconnect if they stop
func (rm *ReconnectManager) MonitorTunnels(interval time.Duration) {
	for {
		time.Sleep(interval)

		for _, name := range rm.manager.ListTunnels() {
			cmd := rm.manager.tunnels[name]
			if cmd.ProcessState != nil && cmd.ProcessState.Exited() {
				log.Printf("Tunnel %s stopped, attempting to reconnect...", name)
				err := rm.manager.StartTunnel(name, cmd)
				if err != nil {
					log.Printf("Failed to reconnect tunnel %s: %v", name, err)
				}
			}
		}
	}
}
