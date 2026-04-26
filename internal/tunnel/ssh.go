package tunnel

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"portbridge/internal/config"
)

// BuildSSHCommand builds the SSH command for a tunnel.
// It validates ports and sanitizes inputs to prevent command injection.
func BuildSSHCommand(sshAlias, user, host, password string, port, localPort, remotePort int) *exec.Cmd {
	// Validate ports
	if err := config.ValidatePort(localPort); err != nil {
		return errorCommand(fmt.Sprintf("invalid local port: %v", err))
	}
	if err := config.ValidatePort(remotePort); err != nil {
		return errorCommand(fmt.Sprintf("invalid remote port: %v", err))
	}

	forward := fmt.Sprintf("%d:localhost:%d", localPort, remotePort)

	if sshAlias != "" {
		// Sanitize: ensure sshAlias doesn't contain dangerous characters
		if containsShellMetachar(sshAlias) {
			return errorCommand("ssh alias contains invalid characters")
		}

		if password != "" {
			// Use sshpass to provide the password non-interactively
			if containsShellMetachar(password) {
				return errorCommand("password contains invalid characters")
			}
			return exec.Command("sshpass", "-p", password, "ssh", "-N", "-L", forward, sshAlias)
		}

		return exec.Command("ssh", "-N", "-L", forward, sshAlias)
	}

	target := host
	if user != "" {
		target = fmt.Sprintf("%s@%s", user, host)
	}

	// Sanitize target
	if containsShellMetachar(target) {
		return errorCommand("host or user contains invalid characters")
	}

	args := []string{"-N", "-L", forward}
	if port > 0 {
		if err := config.ValidatePort(port); err != nil {
			return errorCommand(fmt.Sprintf("invalid SSH port: %v", err))
		}
		args = append(args, "-p", strconv.Itoa(port))
	}

	if password != "" {
		// Use sshpass to provide the password non-interactively
		if containsShellMetachar(password) {
			return errorCommand("password contains invalid characters")
		}
		if user == "" && host == "" {
			return errorCommand("password requires either ssh_alias or user/host")
		}
		return exec.Command("sshpass", "-p", password, "ssh", "-N", "-L", forward, target)
	}

	args = append(args, target)

	return exec.Command("ssh", args...)
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
