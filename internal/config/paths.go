package config

import (
	"io"
	"os"
	"path/filepath"
)

const (
	appDirName          = "portbridge"
	ConfigFilePath      = "/home/cg/.config/.portbridge/portbridge.yaml"
	legacyConfigName    = "portbridge.yaml"
	StateFilePath       = "/home/cg/.config/.portbridge/.portbridge-state.json"
	legacyStateFileName = ".portbridge-state.json"
)

// DefaultConfigPath returns the conventional Linux/XDG config path.
func DefaultConfigPath() (string, error) {
	return ConfigFilePath, nil
}

// DefaultStatePath returns the conventional Linux/XDG state path.
func DefaultStatePath() (string, error) {
	if stateHome := os.Getenv("XDG_STATE_HOME"); stateHome != "" {
		return StateFilePath, nil
	}

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
