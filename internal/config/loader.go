package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type FileConfig struct {
	Profiles map[string]Profile `yaml:"profiles"`
}

// LoadConfig loads the configuration from a YAML file
func LoadConfig(filePath string) (*map[string]Profile, error) {
	resolvedPath, err := ResolveConfigPath(filePath)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(resolvedPath)
	if err != nil {
		return nil, err
	}

	var cfg FileConfig
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}

	if cfg.Profiles == nil {
		cfg.Profiles = make(map[string]Profile)
	}

	return &cfg.Profiles, nil
}

// SaveConfig saves the configuration to a YAML file
func SaveConfig(filePath string, profiles *map[string]Profile) error {
	resolvedPath, err := ResolveConfigPath(filePath)
	if err != nil {
		return err
	}

	if err := EnsureParentDir(resolvedPath); err != nil {
		return err
	}

	cfg := FileConfig{
		Profiles: *profiles,
	}

	data, err := yaml.Marshal(&cfg)
	if err != nil {
		return err
	}

	err = os.WriteFile(resolvedPath, data, 0644)
	if err != nil {
		return err
	}

	return nil
}
