package tunnel

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"portbridge/internal/config"
)

var keepaliveArgs = []string{"-o", "ServerAliveInterval=60", "-o", "ServerAliveCountMax=3", "-o", "TCPKeepAlive=yes"}

// BuildSSHCommand builds the SSH command for a tunnel.
// It validates ports and sanitizes inputs to prevent command injection.
//
// authType: "key" or "password" (from EffectiveAuth)
// keyPath: path to SSH private key file (optional, resolved from ~/.ssh/config if empty)
// passphrase: passphrase for the key (optional, only for key auth)
// password: password for password auth, or legacy plaintext password
func BuildSSHCommand(sshAlias, user, host string, authType, keyPath, passphrase, password string, port, localPort, remotePort int) *exec.Cmd {
	if err := config.ValidatePort(localPort); err != nil {
		return errorCommand(fmt.Sprintf("invalid local port: %v", err))
	}
	if err := config.ValidatePort(remotePort); err != nil {
		return errorCommand(fmt.Sprintf("invalid remote port: %v", err))
	}

	forward := fmt.Sprintf("%d:localhost:%d", localPort, remotePort)

	// Determine the secret to pass via sshpass (if any).
	// sshpass only works for password authentication, NOT for passphrase-protected keys.
	// For key auth with passphrase, SSH itself handles the passphrase interactively.
	var secret string
	useSSHPass := false
	if authType == "password" && password != "" {
		secret = password
		useSSHPass = true
	} else if authType == "" && password != "" {
		secret = password
		useSSHPass = true
	}

	// Build common SSH arguments
	sshArgs := []string{"ssh"}
	sshArgs = append(sshArgs, keepaliveArgs...)

	// If keyPath is empty but we have an alias, try to resolve keyPath from ~/.ssh/config
	resolvedKeyPath := keyPath
	if keyPath == "" && sshAlias != "" {
		resolvedKeyPath = resolveSSHIdentityFile(sshAlias)
	}

	if (authType == "key" || authType == "") && resolvedKeyPath != "" {
		if containsShellMetachar(resolvedKeyPath) {
			return errorCommand("key path contains invalid characters")
		}
		sshArgs = append(sshArgs, "-i", resolvedKeyPath)
	}

	if sshAlias != "" {
		if containsShellMetachar(sshAlias) {
			return errorCommand("ssh alias contains invalid characters")
		}

		if useSSHPass {
			if containsShellMetachar(secret) {
				return errorCommand("password/passphrase contains invalid characters")
			}
			innerArgs := make([]string, len(sshArgs))
			copy(innerArgs, sshArgs)
			innerArgs = append(innerArgs, "-N", "-L", forward, sshAlias)
			execArgs := []string{"-p", secret, "ssh"}
			execArgs = append(execArgs, innerArgs[1:]...)
			return exec.Command("sshpass", execArgs...)
		}

		sshArgs = append(sshArgs, "-N", "-L", forward, sshAlias)
		return exec.Command(sshArgs[0], sshArgs[1:]...)
	}

	target := host
	if user != "" {
		target = fmt.Sprintf("%s@%s", user, host)
	}

	if containsShellMetachar(target) {
		return errorCommand("host or user contains invalid characters")
	}

	if port > 0 {
		if err := config.ValidatePort(port); err != nil {
			return errorCommand(fmt.Sprintf("invalid SSH port: %v", err))
		}
		sshArgs = append(sshArgs, "-p", strconv.Itoa(port))
	}

	if useSSHPass {
		if containsShellMetachar(secret) {
			return errorCommand("password/passphrase contains invalid characters")
		}
		sshArgs = append(sshArgs, "-N", "-L", forward, target)
		execArgs := []string{"-p", secret, "ssh"}
		execArgs = append(execArgs, sshArgs[1:]...)
		return exec.Command("sshpass", execArgs...)
	}

	sshArgs = append(sshArgs, "-N", "-L", forward, target)
	return exec.Command(sshArgs[0], sshArgs[1:]...)
}

// StartTunnel starts an SSH tunnel.
func StartTunnel(cmd *exec.Cmd) error {
	return cmd.Start()
}

// containsShellMetachar checks if a string contains shell metacharacters
// that could be used for command injection.
func containsShellMetachar(s string) bool {
	dangerous := []string{"|", ";", "&", "$", "`", "'", "\"", "\\", "\n", "\r", "\t", "(", ")", "{", "}", "<", ">", "!", "#", "~", "*", "?", "[", "]"}
	for _, c := range dangerous {
		if strings.Contains(s, c) {
			return true
		}
	}
	return false
}

// errorCommand returns a command that will fail with the given message.
func errorCommand(msg string) *exec.Cmd {
	cmd := exec.Command("sh", "-c", fmt.Sprintf("echo %s >&2 && exit 1", strconv.Quote(msg)))
	return cmd
}

// resolveSSHIdentityFile looks up the IdentityFile for a given host alias
// from the user's ~/.ssh/config file.
func resolveSSHIdentityFile(alias string) string {
	sshConfigPath := filepath.Join(os.Getenv("HOME"), ".ssh", "config")

	data, err := os.ReadFile(sshConfigPath)
	if err != nil {
		return ""
	}

	lines := strings.Split(string(data), "\n")
	inHost := false

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		lower := strings.ToLower(line)
		if strings.HasPrefix(lower, "host ") {
			hostPattern := strings.TrimSpace(line[5:])
			inHost = matchHostPattern(hostPattern, alias)
			continue
		}

		if !inHost {
			continue
		}

		if strings.HasPrefix(lower, "identityfile ") {
			return expandSSHPath(strings.TrimSpace(line[13:]))
		}
	}

	return ""
}

// matchHostPattern checks if an SSH config Host pattern matches the alias.
// Supports wildcards (* and ?).
func matchHostPattern(pattern, alias string) bool {
	if pattern == alias {
		return true
	}
	parts := strings.Split(pattern, " ")
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == alias {
			return true
		}
		if strings.Contains(p, "*") || strings.Contains(p, "?") {
			if matchGlob(p, alias) {
				return true
			}
		}
	}
	return false
}

// matchGlob performs simple glob matching (* and ?).
func matchGlob(pattern, s string) bool {
	// Simple glob: replace * with .* and ? with .
	re := strings.ReplaceAll(pattern, "*", ".*")
	re = strings.ReplaceAll(re, "?", ".")
	re = "^" + re + "$"
	matched, _ := filepath.Match(re, s)
	return matched
}

// expandSSHPath expands ~ in SSH-style paths.
func expandSSHPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[2:])
	}
	return path
}