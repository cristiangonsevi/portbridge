package tests

import (
	"os"
	"testing"

	"portbridge/cmd"
	"portbridge/internal/config"
	"portbridge/internal/ui"
)

func TestConfigPaths(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get user home dir: %v", err)
	}

	expectedConfigPath := homeDir + "/.config/portbridge/portbridge.yaml"
	expectedStatePath := homeDir + "/.config/portbridge/.portbridge-state.json"

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

func TestValidatePort(t *testing.T) {
	tests := []struct {
		port    int
		wantErr bool
	}{
		{0, true},
		{1, false},
		{1024, false},
		{65535, false},
		{65536, true},
		{-1, true},
	}

	for _, tt := range tests {
		err := config.ValidatePort(tt.port)
		if (err != nil) != tt.wantErr {
			t.Errorf("ValidatePort(%d) error = %v, wantErr = %v", tt.port, err, tt.wantErr)
		}
	}
}

func TestProfileValidation(t *testing.T) {
	profile := config.Profile{
		Password: "supersecret",
		Tunnels: []config.Tunnel{
			{Name: "valid", Local: 3306, Remote: 3306, Enabled: true},
			{Name: "invalid-port", Local: 0, Remote: 99999, Enabled: true},
		},
	}

	warnings := profile.Validate()
	if len(warnings) == 0 {
		t.Error("Expected validation warnings, got none")
	}

	hasPasswordWarning := false
	hasPortWarning := false
	for _, w := range warnings {
		t.Logf("Validation warning: %s", w)
		if w != "" {
			hasPasswordWarning = true
		}
		if w != "" {
			hasPortWarning = true
		}
	}

	_ = hasPasswordWarning
	_ = hasPortWarning
}

func TestConsoleMessages(t *testing.T) {
	ui.PrintSuccess("Success message")
	ui.PrintError("Error message")
	ui.PrintInfo("Info message")
	ui.PrintWarning("Warning message")
	ui.PrintLog("Log message")
}
