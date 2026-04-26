package tunnel

import "time"

// ReconnectManager handles auto-reconnection of tunnels.
type ReconnectManager struct {
	manager *TunnelManager
}

// NewReconnectManager creates a new ReconnectManager.
func NewReconnectManager(manager *TunnelManager) *ReconnectManager {
	return &ReconnectManager{
		manager: manager,
	}
}

// MonitorTunnels monitors tunnels.
//
// Auto-reconnect is not implemented yet for the persistent state model.
// This placeholder keeps the API stable until reconnect behavior is
// redesigned around persisted tunnel metadata.
func (rm *ReconnectManager) MonitorTunnels(interval time.Duration) {
	_ = rm
	_ = interval
}
