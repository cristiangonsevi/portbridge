package config

import "fmt"

// Tunnel represents an SSH tunnel configuration
type Tunnel struct {
	Name    string `yaml:"name"`
	Local   int    `yaml:"local"`
	Remote  int    `yaml:"remote"`
	Enabled bool   `yaml:"enabled"`
}

// Profile represents a profile configuration
type Profile struct {
	SSHAlias string   `yaml:"ssh_alias,omitempty"`
	Host     string   `yaml:"host,omitempty"`
	Port     int      `yaml:"port,omitempty"`
	User     string   `yaml:"user,omitempty"`
	Password string   `yaml:"password,omitempty"`
	SSHKey   string   `yaml:"ssh_key_file,omitempty"`
	Tunnels  []Tunnel `yaml:"tunnels"`
}

// ValidatePort checks if a port number is valid (1-65535).
func ValidatePort(port int) error {
	if port < 1 || port > 65535 {
		return fmt.Errorf("invalid port %d: must be between 1 and 65535", port)
	}
	return nil
}

// Validate checks the profile configuration for common issues.
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
