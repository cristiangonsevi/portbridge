package config

import (
	"io"
	"os"
	"path/filepath"
)

const (
	appDirName          = "portbridge"
	legacyConfigName    = "portbridge.yaml"
	legacyStateFileName = ".portbridge-state.json"
)

// ConfigFilePath is the path to the YAML configuration file.
var ConfigFilePath string

// StateFilePath is the path to the JSON state file.
var StateFilePath string

func init() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}

	configDir := filepath.Join(homeDir, ".config", appDirName)
	ConfigFilePath = filepath.Join(configDir, "portbridge.yaml")
	StateFilePath = filepath.Join(configDir, ".portbridge-state.json")
}

// DefaultConfigPath returns the conventional Linux/XDG config path.
func DefaultConfigPath() (string, error) {
	return ConfigFilePath, nil
}

// DefaultStatePath returns the conventional Linux/XDG state path.
func DefaultStatePath() (string, error) {
	return StateFilePath, nil
}

// ResolveConfigPath resolves the config path and migrates a legacy local file if needed.
func ResolveConfigPath(filePath string) (string, error) {
	if filePath != "" && filePath != ConfigFilePath {
		return filePath, nil
	}

	resolvedPath, err := DefaultConfigPath()
	if err != nil {
		return "", err
	}

	if err := migrateLegacyFileIfNeeded(resolvedPath, legacyConfigName); err != nil {
		return "", err
	}

	return resolvedPath, nil
}

// ResolveStatePath resolves the state path and migrates a legacy local file if needed.
func ResolveStatePath(filePath string) (string, error) {
	if filePath != "" && filePath != StateFilePath {
		return filePath, nil
	}

	resolvedPath, err := DefaultStatePath()
	if err != nil {
		return "", err
	}

	if err := migrateLegacyFileIfNeeded(resolvedPath, legacyStateFileName); err != nil {
		return "", err
	}

	return resolvedPath, nil
}

// EnsureParentDir creates the parent directory for a target file path.
func EnsureParentDir(filePath string) error {
	return os.MkdirAll(filepath.Dir(filePath), 0755)
}

func migrateLegacyFileIfNeeded(targetPath, legacyName string) error {
	if _, err := os.Stat(targetPath); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return err
	}

	legacyPath := filepath.Join(".", legacyName)
	if _, err := os.Stat(legacyPath); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	targetAbs, err := filepath.Abs(targetPath)
	if err != nil {
		return err
	}

	legacyAbs, err := filepath.Abs(legacyPath)
	if err != nil {
		return err
	}

	if targetAbs == legacyAbs {
		return nil
	}

	if err := EnsureParentDir(targetPath); err != nil {
		return err
	}

	sourceFile, err := os.Open(legacyPath)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	targetFile, err := os.Create(targetPath)
	if err != nil {
		return err
	}
	defer targetFile.Close()

	if _, err := io.Copy(targetFile, sourceFile); err != nil {
		return err
	}

	return targetFile.Chmod(0644)
}
