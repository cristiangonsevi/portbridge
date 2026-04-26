package tunnel

import (
	"fmt"
	"os/exec"
	"strconv"
)

// BuildSSHCommand builds the SSH command for a tunnel.
func BuildSSHCommand(sshAlias, user, host string, port, localPort, remotePort int) *exec.Cmd {
	forward := fmt.Sprintf("%d:localhost:%d", localPort, remotePort)

	if sshAlias != "" {
		return exec.Command("ssh", "-N", "-L", forward, sshAlias)
	}

	target := host
	if user != "" {
		target = fmt.Sprintf("%s@%s", user, host)
	}

	args := []string{"-N", "-L", forward}
	if port > 0 {
		args = append(args, "-p", strconv.Itoa(port))
	}
	args = append(args, target)

	return exec.Command("ssh", args...)
}

// StartTunnel starts an SSH tunnel.
func StartTunnel(cmd *exec.Cmd) error {
	return cmd.Start()
}
