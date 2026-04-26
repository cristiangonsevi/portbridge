package tests

import (
	"os/exec"
	"testing"

	"portbridge/cmd"
	"portbridge/internal/config"
	"portbridge/internal/tunnel"
	"portbridge/internal/ui"
)

func TestConfigPaths(t *testing.T) {
	expectedConfigPath := "/home/cg/.config/.portbridge/portbridge.yaml"
	expectedStatePath := "/home/cg/.config/.portbridge/.portbridge-state.json"

	if config.ConfigFilePath != expectedConfigPath {
		t.Errorf("Expected ConfigFilePath to be %s, got %s", expectedConfigPath, config.ConfigFilePath)
	}

	if config.StateFilePath != expectedStatePath {
		t.Errorf("Expected StateFilePath to be %s, got %s", expectedStatePath, config.StateFilePath)
	}
}

func TestShowConfigCommand(t *testing.T) {
	cmd.RootCmd.SetArgs([]string{"showconfig"})
	err := cmd.RootCmd.Execute()
	if err != nil {
		t.Errorf("showconfig command failed: %v", err)
	}
}

func TestTunnelManager(t *testing.T) {
	manager := tunnel.NewTunnelManager()

	// Simulate user enabling a tunnel
	record := tunnel.TunnelRecord{
		Profile:    "qa",
		Name:       "db",
		LocalPort:  5433,
		RemotePort: 5432,
	}
	cmd := exec.Command("ssh", "-N", "-L", "5433:localhost:5432", "qa.taak")
	status, err := manager.StartTunnel(record, cmd)
	if err != nil {
		t.Errorf("Failed to start tunnel: %v", err)
	}
	if status != "started" {
		t.Errorf("Expected status to be 'started', got %s", status)
	}

	// Simulate user listing tunnels
	tunnels, err := manager.ListProfileTunnels(record.Profile)
	if err != nil {
		t.Errorf("Failed to list tunnels: %v", err)
	}
	if len(tunnels) == 0 || tunnels[0].Name != record.Name {
		t.Errorf("Expected tunnel %s to be listed, but it was not", record.Name)
	}

	// Simulate user disabling a tunnel
	err = manager.StopTunnel(record.Profile, record.Name)
	if err != nil {
		t.Errorf("Failed to stop tunnel: %v", err)
	}

	// Verify tunnel is no longer listed
	tunnels, err = manager.ListProfileTunnels(record.Profile)
	if err != nil {
		t.Errorf("Failed to list tunnels after stopping: %v", err)
	}
	if len(tunnels) > 0 {
		t.Errorf("Expected no tunnels to be listed, but found %d", len(tunnels))
	}
}

func TestConsoleMessages(t *testing.T) {
	ui.PrintSuccess("Success message")
	ui.PrintError("Error message")
	ui.PrintInfo("Info message")
	ui.PrintWarning("Warning message")
	ui.PrintLog("Log message")
}
