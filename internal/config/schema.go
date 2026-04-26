package config

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
