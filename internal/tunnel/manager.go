package tunnel

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"portbridge/internal/ui"
)

const stateFileName = ".portbridge-state.json"

type TunnelRecord struct {
	Profile    string `json:"profile"`
	Name       string `json:"name"`
	PID        int    `json:"pid"`
	SSHAlias   string `json:"ssh_alias,omitempty"`
	Host       string `json:"host,omitempty"`
	User       string `json:"user,omitempty"`
	Port       int    `json:"port,omitempty"`
	LocalPort  int    `json:"local_port"`
	RemotePort int    `json:"remote_port"`
}

func (tr TunnelRecord) Key() string {
	return tr.Profile + ":" + tr.Name
}

func (tr TunnelRecord) SameConfig(other TunnelRecord) bool {
	return tr.Profile == other.Profile &&
		tr.Name == other.Name &&
		tr.SSHAlias == other.SSHAlias &&
		tr.Host == other.Host &&
		tr.User == other.User &&
		tr.Port == other.Port &&
		tr.LocalPort == other.LocalPort &&
		tr.RemotePort == other.RemotePort
}

type tunnelState struct {
	Tunnels map[string]TunnelRecord `json:"tunnels"`
}

type TunnelManager struct {
	mutex     sync.Mutex
	statePath string
}

// NewTunnelManager creates a new TunnelManager.
func NewTunnelManager() *TunnelManager {
	return &TunnelManager{
		statePath: stateFileName,
	}
}

func (tm *TunnelManager) loadState() (*tunnelState, error) {
	data, err := os.ReadFile(tm.statePath)
	if err != nil {
		if os.IsNotExist(err) {
			return &tunnelState{Tunnels: make(map[string]TunnelRecord)}, nil
		}
		return nil, err
	}

	var state tunnelState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, err
	}

	if state.Tunnels == nil {
		state.Tunnels = make(map[string]TunnelRecord)
	}

	return &state, nil
}

func (tm *TunnelManager) saveState(state *tunnelState) error {
	if len(state.Tunnels) == 0 {
		if err := os.Remove(tm.statePath); err != nil && !os.IsNotExist(err) {
			return err
		}
		return nil
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(tm.statePath, data, 0644)
}

func isPIDRunning(pid int) bool {
	if pid <= 0 {
		return false
	}

	err := syscall.Kill(pid, 0)
	return err == nil || err == syscall.EPERM
}

func terminatePID(pid int) error {
	if pid <= 0 {
		return nil
	}

	err := syscall.Kill(pid, syscall.SIGTERM)
	if err == syscall.ESRCH {
		return nil
	}
	return err
}

func findListeningProcessByPort(port int) (int, string, error) {
	output, err := exec.Command("ss", "-ltnp").CombinedOutput()
	if err != nil {
		return 0, "", err
	}

	portNeedle := ":" + strconv.Itoa(port)
	pidRegex := regexp.MustCompile(`pid=(\d+)`)
	procRegex := regexp.MustCompile(`\("([^"]+)"`)

	for _, line := range strings.Split(string(output), "\n") {
		if !strings.Contains(line, portNeedle) {
			continue
		}

		pidMatch := pidRegex.FindStringSubmatch(line)
		procMatch := procRegex.FindStringSubmatch(line)
		if len(pidMatch) < 2 {
			continue
		}

		pid, err := strconv.Atoi(pidMatch[1])
		if err != nil {
			continue
		}

		processName := ""
		if len(procMatch) >= 2 {
			processName = procMatch[1]
		}

		return pid, processName, nil
	}

	return 0, "", nil
}

func (tm *TunnelManager) cleanupStaleLocked(state *tunnelState) bool {
	changed := false
	for key, record := range state.Tunnels {
		if !isPIDRunning(record.PID) {
			delete(state.Tunnels, key)
			changed = true
		}
	}
	return changed
}

// ListProfileTunnels lists active tracked tunnels for a profile.
func (tm *TunnelManager) ListProfileTunnels(profile string) ([]TunnelRecord, error) {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	state, err := tm.loadState()
	if err != nil {
		return nil, err
	}

	changed := tm.cleanupStaleLocked(state)

	records := make([]TunnelRecord, 0)
	for _, record := range state.Tunnels {
		if record.Profile == profile {
			records = append(records, record)
		}
	}

	if changed {
		if err := tm.saveState(state); err != nil {
			return nil, err
		}
	}

	return records, nil
}

// StartTunnel starts a tunnel, persists its PID, and detects already-running tunnels.
func (tm *TunnelManager) StartTunnel(record TunnelRecord, cmd *exec.Cmd) (string, error) {
	tm.mutex.Lock()
	state, err := tm.loadState()
	if err != nil {
		tm.mutex.Unlock()
		return "", err
	}

	changed := tm.cleanupStaleLocked(state)
	key := record.Key()

	if existing, exists := state.Tunnels[key]; exists {
		if existing.SameConfig(record) {
			if changed {
				if err := tm.saveState(state); err != nil {
					tm.mutex.Unlock()
					return "", err
				}
			}
			tm.mutex.Unlock()
			return "already-running", nil
		}

		ui.PrintWarning(fmt.Sprintf("Tunnel %s changed, restarting with new configuration", record.Name))
		if err := terminatePID(existing.PID); err != nil {
			tm.mutex.Unlock()
			return "", err
		}
		delete(state.Tunnels, key)
	}

	pid, processName, err := findListeningProcessByPort(record.LocalPort)
	if err != nil {
		tm.mutex.Unlock()
		return "", err
	}

	if pid > 0 {
		if processName == "ssh" {
			record.PID = pid
			state.Tunnels[key] = record
			if err := tm.saveState(state); err != nil {
				tm.mutex.Unlock()
				return "", err
			}
			tm.mutex.Unlock()
			ui.PrintWarning(fmt.Sprintf("Tunnel %s is already running on local port %d with PID %d", record.Name, record.LocalPort, pid))
			return "already-running", nil
		}

		tm.mutex.Unlock()
		return "", fmt.Errorf("local port %d is already in use by %s (pid %d)", record.LocalPort, processName, pid)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		tm.mutex.Unlock()
		ui.PrintError(fmt.Sprintf("Tunnel %s failed to start: %v", record.Name, err))
		return "", err
	}

	waitCh := make(chan error, 1)
	go func() {
		waitCh <- cmd.Wait()
	}()

	select {
	case err := <-waitCh:
		tm.mutex.Unlock()
		if err != nil {
			ui.PrintError(fmt.Sprintf("Tunnel %s failed during startup: %v", record.Name, err))
			return "", err
		}
		ui.PrintError(fmt.Sprintf("Tunnel %s exited during startup", record.Name))
		return "", fmt.Errorf("tunnel %s exited during startup", record.Name)
	case <-time.After(1500 * time.Millisecond):
	}

	record.PID = cmd.Process.Pid
	state.Tunnels[key] = record

	if err := tm.saveState(state); err != nil {
		tm.mutex.Unlock()
		return "", err
	}
	tm.mutex.Unlock()

	ui.PrintLog(fmt.Sprintf("Tunnel %s started with PID %d", record.Name, record.PID))
	return "started", nil
}

// StopTunnel stops a running tunnel by profile and tunnel name.
func (tm *TunnelManager) StopTunnel(profile, name string) error {
	tm.mutex.Lock()
	state, err := tm.loadState()
	if err != nil {
		tm.mutex.Unlock()
		return err
	}

	tm.cleanupStaleLocked(state)

	key := profile + ":" + name
	record, exists := state.Tunnels[key]
	if !exists {
		if err := tm.saveState(state); err != nil {
			tm.mutex.Unlock()
			return err
		}
		tm.mutex.Unlock()
		return fmt.Errorf("tunnel %s is not running", name)
	}

	delete(state.Tunnels, key)
	if err := tm.saveState(state); err != nil {
		tm.mutex.Unlock()
		return err
	}
	tm.mutex.Unlock()

	ui.PrintLog(fmt.Sprintf("Stopping tunnel %s", name))

	if err := terminatePID(record.PID); err != nil {
		ui.PrintError(fmt.Sprintf("Tunnel %s failed to stop: %v", name, err))
		return err
	}

	ui.PrintWarning(fmt.Sprintf("Tunnel %s terminated", name))
	return nil
}
