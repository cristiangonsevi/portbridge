package config

import (
	"fmt"
	"os"
	"strings"
)

// Tunnel represents an SSH tunnel configuration
type Tunnel struct {
	Name    string `yaml:"name"`
	Local   int    `yaml:"local"`
	Remote  int    `yaml:"remote"`
	Enabled bool   `yaml:"enabled"`
}

// Profile represents a profile configuration
type Profile struct {
	SSHAlias          string   `yaml:"ssh_alias,omitempty"`
	Host              string   `yaml:"host,omitempty"`
	Port              int      `yaml:"port,omitempty"`
	User              string   `yaml:"user,omitempty"`
	Password          string   `yaml:"password,omitempty"`
	SSHKey            string   `yaml:"ssh_key_file,omitempty"`
	ReconnectInterval int      `yaml:"reconnect_interval,omitempty"`
	Tunnels           []Tunnel `yaml:"tunnels"`
}

// ValidatePort checks if a port number is valid (1-65535).
func ValidatePort(port int) error {
	if port < 1 || port > 65535 {
		return fmt.Errorf("invalid port %d: must be between 1 and 65535", port)
	}
	return nil
}

// ValidationResult holds the result of a configuration validation.
type ValidationResult struct {
	ProfileName string   `json:"profile_name"`
	Errors      []string `json:"errors,omitempty"`
	Warnings    []string `json:"warnings,omitempty"`
}

// Validate checks the profile configuration for common issues (legacy method).
func (p *Profile) Validate() []string {
	var warnings []string

	if p.Password != "" {
		warnings = append(warnings, "profile uses password authentication (stored in plain text). Consider using ssh_key_file instead.")
	}

	for _, t := range p.Tunnels {
		if err := ValidatePort(t.Local); err != nil {
			warnings = append(warnings, fmt.Sprintf("tunnel %q: invalid local port: %v", t.Name, err))
		}
		if err := ValidatePort(t.Remote); err != nil {
			warnings = append(warnings, fmt.Sprintf("tunnel %q: invalid remote port: %v", t.Name, err))
		}
	}

	return warnings
}

// ValidateProfile checks a single profile for configuration errors and warnings.
func (p *Profile) ValidateProfile(name string) ValidationResult {
	result := ValidationResult{ProfileName: name}

	if p.SSHAlias == "" && p.Host == "" {
		result.Errors = append(result.Errors, "profile must have either ssh_alias or host configured")
	}

	if p.SSHAlias == "" {
		if p.User == "" {
			result.Errors = append(result.Errors, "profile requires a user when no ssh_alias is provided")
		}
		if p.Host == "" {
			result.Errors = append(result.Errors, "profile requires a host when no ssh_alias is provided")
		}
	}

	if p.Port > 0 {
		if err := ValidatePort(p.Port); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("invalid SSH port: %v", err))
		}
	}

	if len(p.Tunnels) == 0 {
		result.Warnings = append(result.Warnings, "profile has no tunnels configured")
	}

	seenTunnelNames := make(map[string]bool)
	for _, t := range p.Tunnels {
		if t.Name == "" {
			result.Errors = append(result.Errors, "tunnel has an empty name")
			continue
		}

		if seenTunnelNames[t.Name] {
			result.Errors = append(result.Errors, fmt.Sprintf("duplicate tunnel name %q", t.Name))
		}
		seenTunnelNames[t.Name] = true

		if err := ValidatePort(t.Local); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("tunnel %q: invalid local port: %v", t.Name, err))
		}
		if err := ValidatePort(t.Remote); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("tunnel %q: invalid remote port: %v", t.Name, err))
		}
		if t.Local == t.Remote {
			result.Warnings = append(result.Warnings, fmt.Sprintf("tunnel %q: local port equals remote port (%d)", t.Name, t.Local))
		}
	}

	if p.Password != "" {
		result.Warnings = append(result.Warnings, "profile uses password authentication (stored in plain text). Consider using ssh_key_file instead.")
	}

	if p.SSHKey != "" {
		if _, err := os.Stat(p.SSHKey); os.IsNotExist(err) {
			result.Warnings = append(result.Warnings, fmt.Sprintf("ssh_key_file %q does not exist", p.SSHKey))
		}
	}

	return result
}

// ValidateConfig validates all profiles in the configuration.
// Returns a list of per-profile validation results plus any top-level config errors.
func ValidateConfig(profiles map[string]Profile) ([]ValidationResult, []string) {
	var results []ValidationResult
	var configErrors []string

	if len(profiles) == 0 {
		configErrors = append(configErrors, "configuration has no profiles defined")
		return results, configErrors
	}

	for name, profile := range profiles {
		result := profile.ValidateProfile(name)
		results = append(results, result)
	}

	return results, configErrors
}

// FormatValidationOutput formats validation results as a human-readable string.
func FormatValidationOutput(profiles map[string]Profile, results []ValidationResult, configErrors []string) string {
	var b strings.Builder

	if len(configErrors) > 0 {
		b.WriteString("✘ Configuration errors:\n")
		for _, err := range configErrors {
			b.WriteString(fmt.Sprintf("    %s\n", err))
		}
		b.WriteString("Result: FAIL\n")
		return b.String()
	}

	if len(profiles) == 0 {
		b.WriteString("⚠ No profiles found in configuration.\n")
		b.WriteString("Result: OK (empty config)\n")
		return b.String()
	}

	totalErrors := 0
	totalWarnings := 0
	totalTunnels := 0

	for _, p := range profiles {
		totalTunnels += len(p.Tunnels)
	}

	for _, r := range results {
		totalErrors += len(r.Errors)
		totalWarnings += len(r.Warnings)

		if len(r.Errors) > 0 || len(r.Warnings) > 0 {
			b.WriteString(fmt.Sprintf("\n%s:\n", r.ProfileName))
			for _, err := range r.Errors {
				b.WriteString(fmt.Sprintf("  ✘ %s\n", err))
			}
			for _, w := range r.Warnings {
				b.WriteString(fmt.Sprintf("  ⚠ %s\n", w))
			}
		}
	}

	b.WriteString(fmt.Sprintf("\n%d profiles, %d tunnels. %d errors, %d warnings.\n",
		len(profiles), totalTunnels, totalErrors, totalWarnings))

	if totalErrors > 0 {
		b.WriteString("Validation: FAIL\n")
	} else if totalWarnings > 0 {
		b.WriteString("Validation: OK (with warnings)\n")
	} else {
		b.WriteString("Validation: OK\n")
	}

	return b.String()
}
