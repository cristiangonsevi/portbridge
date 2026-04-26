package config

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const (
	appDirName          = "portbridge"
	configFileName      = "~/.config/.portbridge/portbridge.yaml"
	legacyConfigName    = "portbridge.yaml"
	stateFileName       = "~/.config/.portbridge/.portbridge-state.json"
	legacyStateFileName = ".portbridge-state.json"
)

// DefaultConfigPath returns the conventional Linux/XDG config path.
func DefaultConfigPath() (string, error) {
	baseDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolve user config dir: %w", err)
	}

	return filepath.Join(baseDir, appDirName, configFileName), nil
}

// DefaultStatePath returns the conventional Linux/XDG state path.
func DefaultStatePath() (string, error) {
	if stateHome := os.Getenv("XDG_STATE_HOME"); stateHome != "" {
		return filepath.Join(stateHome, appDirName, stateFileName), nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve user home dir: %w", err)
	}

	return filepath.Join(homeDir, ".local", "state", appDirName, stateFileName), nil
}

// ResolveConfigPath resolves the config path and migrates a legacy local file if needed.
func ResolveConfigPath(filePath string) (string, error) {
	if filePath != "" && filePath != configFileName {
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
	if filePath != "" && filePath != stateFileName {
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
