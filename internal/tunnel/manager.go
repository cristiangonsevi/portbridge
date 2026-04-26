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
	if err != nil && err != syscall.ESRCH {
		return err
	}

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if !isPIDRunning(pid) {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}

	err = syscall.Kill(pid, syscall.SIGKILL)
	if err != nil && err != syscall.ESRCH {
		return err
	}

	deadline = time.Now().Add(1 * time.Second)
	for time.Now().Before(deadline) {
		if !isPIDRunning(pid) {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}

	if isPIDRunning(pid) {
		return fmt.Errorf("pid %d is still running after SIGTERM/SIGKILL", pid)
	}

	return nil
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

func findActiveSSHPIDForPort(port int) (int, error) {
	pid, processName, err := findListeningProcessByPort(port)
	if err != nil {
		return 0, err
	}

	if pid > 0 && processName == "ssh" {
		return pid, nil
	}

	return 0, nil
}

// FindActiveTunnelPID returns the active SSH PID listening on a local port.
func (tm *TunnelManager) FindActiveTunnelPID(localPort int) (int, error) {
	return findActiveSSHPIDForPort(localPort)
}

// StopTunnelByPort stops any active SSH tunnel listener on the given local port
// and removes matching records from persisted state.
func (tm *TunnelManager) StopTunnelByPort(localPort int) error {
	tm.mutex.Lock()
	state, err := tm.loadState()
	if err != nil {
		tm.mutex.Unlock()
		return err
	}

	changed, err := tm.cleanupStaleLocked(state)
	if err != nil {
		tm.mutex.Unlock()
		return err
	}

	for key, record := range state.Tunnels {
		if record.LocalPort == localPort {
			delete(state.Tunnels, key)
			changed = true
		}
	}

	if changed {
		if err := tm.saveState(state); err != nil {
			tm.mutex.Unlock()
			return err
		}
	}
	tm.mutex.Unlock()

	pid, err := findActiveSSHPIDForPort(localPort)
	if err != nil {
		return err
	}
	if pid == 0 {
		return nil
	}

	ui.PrintLog(fmt.Sprintf("Stopping active SSH listener on local port %d", localPort))

	if err := terminatePID(pid); err != nil {
		return err
	}

	pid, err = findActiveSSHPIDForPort(localPort)
	if err != nil {
		return err
	}
	if pid > 0 {
		return fmt.Errorf("local port %d still has an active SSH listener (pid %d)", localPort, pid)
	}

	return nil
}

func (tm *TunnelManager) cleanupStaleLocked(state *tunnelState) (bool, error) {
	changed := false

	for key, record := range state.Tunnels {
		activePID, err := findActiveSSHPIDForPort(record.LocalPort)
		if err != nil {
			return changed, err
		}

		if activePID > 0 {
			if record.PID != activePID {
				record.PID = activePID
				state.Tunnels[key] = record
				changed = true
			}
			continue
		}

		delete(state.Tunnels, key)
		changed = true
	}

	return changed, nil
}

// ListProfileTunnels lists active tracked tunnels for a profile.
func (tm *TunnelManager) ListProfileTunnels(profile string) ([]TunnelRecord, error) {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	state, err := tm.loadState()
	if err != nil {
		return nil, err
	}

	changed, err := tm.cleanupStaleLocked(state)
	if err != nil {
		return nil, err
	}

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

	changed, err := tm.cleanupStaleLocked(state)
	if err != nil {
		tm.mutex.Unlock()
		return "", err
	}

	key := record.Key()

	if existing, exists := state.Tunnels[key]; exists {
		if existing.SameConfig(record) {
			activePID, err := findActiveSSHPIDForPort(record.LocalPort)
			if err != nil {
				tm.mutex.Unlock()
				return "", err
			}

			if activePID > 0 {
				if existing.PID != activePID {
					existing.PID = activePID
					state.Tunnels[key] = existing
					changed = true
				}
				if changed {
					if err := tm.saveState(state); err != nil {
						tm.mutex.Unlock()
						return "", err
					}
				}
				tm.mutex.Unlock()
				return "already-running", nil
			}

			delete(state.Tunnels, key)
			changed = true
		} else {
			ui.PrintWarning(fmt.Sprintf("Tunnel %s changed, restarting with new configuration", record.Name))
			if err := terminatePID(existing.PID); err != nil {
				tm.mutex.Unlock()
				return "", err
			}

			activePID, err := findActiveSSHPIDForPort(existing.LocalPort)
			if err != nil {
				tm.mutex.Unlock()
				return "", err
			}
			if activePID > 0 && activePID != existing.PID {
				if err := terminatePID(activePID); err != nil {
					tm.mutex.Unlock()
					return "", err
				}
			}

			delete(state.Tunnels, key)
			changed = true
		}
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

	deadline := time.Now().Add(5 * time.Second)
	for {
		select {
		case err := <-waitCh:
			tm.mutex.Unlock()
			if err != nil {
				ui.PrintError(fmt.Sprintf("Tunnel %s failed during startup: %v", record.Name, err))
				return "", err
			}
			ui.PrintError(fmt.Sprintf("Tunnel %s exited during startup", record.Name))
			return "", fmt.Errorf("tunnel %s exited during startup", record.Name)
		default:
		}

		activePID, err := findActiveSSHPIDForPort(record.LocalPort)
		if err != nil {
			tm.mutex.Unlock()
			return "", err
		}

		if activePID > 0 {
			record.PID = activePID
			state.Tunnels[key] = record

			if err := tm.saveState(state); err != nil {
				tm.mutex.Unlock()
				return "", err
			}
			tm.mutex.Unlock()

			ui.PrintLog(fmt.Sprintf("Tunnel %s started with PID %d", record.Name, record.PID))
			return "started", nil
		}

		if time.Now().After(deadline) {
			_ = terminatePID(cmd.Process.Pid)
			tm.mutex.Unlock()
			return "", fmt.Errorf("tunnel %s did not open local port %d in time", record.Name, record.LocalPort)
		}

		time.Sleep(100 * time.Millisecond)
	}
}

// StopTunnel stops a running tunnel by profile and tunnel name.
func (tm *TunnelManager) StopTunnel(profile, name string) error {
	tm.mutex.Lock()
	state, err := tm.loadState()
	if err != nil {
		tm.mutex.Unlock()
		return err
	}

	_, err = tm.cleanupStaleLocked(state)
	if err != nil {
		tm.mutex.Unlock()
		return err
	}

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
		ui.PrintError(fmt.Sprintf("Tunnel %s failed to stop tracked PID %d: %v", name, record.PID, err))
		return err
	}

	activePID, err := findActiveSSHPIDForPort(record.LocalPort)
	if err != nil {
		ui.PrintError(fmt.Sprintf("Tunnel %s stop verification failed: %v", name, err))
		return err
	}

	if activePID > 0 && activePID != record.PID {
		ui.PrintWarning(fmt.Sprintf("Tunnel %s had a different active SSH PID %d on port %d, terminating it", name, activePID, record.LocalPort))
		if err := terminatePID(activePID); err != nil {
			ui.PrintError(fmt.Sprintf("Tunnel %s failed to stop active PID %d: %v", name, activePID, err))
			return err
		}
	}

	activePID, err = findActiveSSHPIDForPort(record.LocalPort)
	if err != nil {
		ui.PrintError(fmt.Sprintf("Tunnel %s final stop verification failed: %v", name, err))
		return err
	}
	if activePID > 0 {
		return fmt.Errorf("tunnel %s still has an active SSH listener on port %d (pid %d)", name, record.LocalPort, activePID)
	}

	ui.PrintWarning(fmt.Sprintf("Tunnel %s terminated", name))
	return nil
}
