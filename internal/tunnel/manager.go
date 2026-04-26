package tunnel

import (
	"fmt"
	"os"
	"os/exec"
	"sync"

	"portbridge/internal/ui"
)

type TunnelManager struct {
	tunnels map[string]*exec.Cmd
	mutex   sync.Mutex
}

// NewTunnelManager creates a new TunnelManager
func NewTunnelManager() *TunnelManager {
	return &TunnelManager{
		tunnels: make(map[string]*exec.Cmd),
	}
}

// StartTunnel starts a tunnel and tracks it
func (tm *TunnelManager) StartTunnel(name string, cmd *exec.Cmd) error {
	tm.mutex.Lock()
	if _, exists := tm.tunnels[name]; exists {
		tm.mutex.Unlock()
		return fmt.Errorf("tunnel %s is already running", name)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Start()
	if err != nil {
		tm.mutex.Unlock()
		ui.PrintError(fmt.Sprintf("Tunnel %s failed to start: %v", name, err))
		return err
	}

	tm.tunnels[name] = cmd
	tm.mutex.Unlock()

	ui.PrintLog(fmt.Sprintf("Tunnel %s started with PID %d", name, cmd.Process.Pid))

	go func() {
		err := cmd.Wait()

		tm.mutex.Lock()
		delete(tm.tunnels, name)
		tm.mutex.Unlock()

		if err != nil {
			ui.PrintError(fmt.Sprintf("Tunnel %s exited with error: %v", name, err))
			return
		}

		ui.PrintWarning(fmt.Sprintf("Tunnel %s stopped", name))
	}()

	return nil
}

// StopTunnel stops a running tunnel
func (tm *TunnelManager) StopTunnel(name string) error {
	tm.mutex.Lock()
	cmd, exists := tm.tunnels[name]
	if !exists {
		tm.mutex.Unlock()
		return fmt.Errorf("tunnel %s is not running", name)
	}

	delete(tm.tunnels, name)
	tm.mutex.Unlock()

	ui.PrintLog(fmt.Sprintf("Stopping tunnel %s", name))

	err := cmd.Process.Kill()
	if err != nil {
		ui.PrintError(fmt.Sprintf("Tunnel %s failed to stop: %v", name, err))
		return err
	}

	ui.PrintWarning(fmt.Sprintf("Tunnel %s terminated", name))
	return nil
}

// ListTunnels lists all running tunnels
func (tm *TunnelManager) ListTunnels() []string {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	names := make([]string, 0, len(tm.tunnels))
	for name := range tm.tunnels {
		names = append(names, name)
	}

	return names
}
