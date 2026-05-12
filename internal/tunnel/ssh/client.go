package ssh

import (
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"golang.org/x/term"

	"portbridge/internal/config"
)

type Client struct {
	client          *ssh.Client
	config          *config.Profile
	forwardListener net.Listener
	forwardConns    map[net.Conn]struct{}
	forwardMutex    sync.Mutex
	stopCh          chan struct{}
	wg              sync.WaitGroup
}

func NewClient(cfg *config.Profile) (*Client, error) {
	var resolvedKeyPath string
	var resolvedUser string

	sshCfg := cfg.EffectiveSSH()

	fmt.Printf("[DEBUG] NewClient: SSHAlias=%q Host=%q User=%q\n", sshCfg.Alias, sshCfg.Host, sshCfg.User)
	fmt.Printf("[DEBUG] NewClient: Auth=%+v\n", sshCfg.Auth)

	host := sshCfg.Host
	user := sshCfg.User

	if host == "" && sshCfg.Alias != "" {
		aliasInfo, err := resolveSSHAlias(sshCfg.Alias)
		if err != nil {
			fmt.Printf("[DEBUG] resolveSSHAlias error: %v\n", err)
		} else {
			fmt.Printf("[DEBUG] resolved alias: hostname=%q identityFile=%q user=%q\n", aliasInfo.hostname, aliasInfo.identityFile, aliasInfo.user)
			resolvedUser = aliasInfo.user
		}
		host = aliasInfo.hostname
		resolvedKeyPath = aliasInfo.identityFile
	}

	authType, keyPath, passphrase, password := cfg.EffectiveAuth()
	fmt.Printf("[DEBUG] EffectiveAuth: type=%q keyPath=%q passphrase=%q password=%q\n", authType, keyPath, passphrase, password != "")

	if keyPath == "" && resolvedKeyPath != "" {
		keyPath = resolvedKeyPath
		fmt.Printf("[DEBUG] keyPath resolved from SSH config: %q\n", keyPath)
	}

	sshClientCfg, err := buildSSHClientConfig(user, authType, keyPath, passphrase, password)
	if err != nil {
		return nil, fmt.Errorf("building SSH config: %w", err)
	}

	if sshClientCfg.User == "" {
		sshClientCfg.User = user
		if sshClientCfg.User == "" && resolvedUser != "" {
			sshClientCfg.User = resolvedUser
			fmt.Printf("[DEBUG] using user=%q from SSH config\n", resolvedUser)
		}
	}

	addr := net.JoinHostPort(host, fmt.Sprintf("%d", sshCfg.Port))
	if sshCfg.Port == 0 {
		addr = net.JoinHostPort(host, "22")
	}

	fmt.Printf("[DEBUG] Final SSH config: user=%q addr=%s\n", sshClientCfg.User, addr)

	conn, err := ssh.Dial("tcp", addr, sshClientCfg)
	if err != nil {
		return nil, fmt.Errorf("dialing SSH %s@%s: %w", sshClientCfg.User, addr, err)
	}

	return &Client{
		client:       conn,
		config:       cfg,
		forwardConns: make(map[net.Conn]struct{}),
		stopCh:       make(chan struct{}),
	}, nil
}

func (c *Client) Close() error {
	c.stopCh <- struct{}{}
	c.wg.Wait()

	c.forwardMutex.Lock()
	for conn := range c.forwardConns {
		conn.Close()
	}
	c.forwardMutex.Unlock()

	if c.forwardListener != nil {
		c.forwardListener.Close()
	}

	return c.client.Close()
}

func (c *Client) UnderlyingClient() *ssh.Client {
	return c.client
}

func (c *Client) LocalForward(localPort, remotePort int) error {
	listener, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", localPort))
	if err != nil {
		return fmt.Errorf("listening on local port %d: %w", localPort, err)
	}
	c.forwardListener = listener

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				select {
				case <-c.stopCh:
					return
				default:
					continue
				}
			}

			c.wg.Add(1)
			go func() {
				defer c.wg.Done()
				c.handleLocalForward(conn, remotePort)
			}()
		}
	}()

	return nil
}

func (c *Client) handleLocalForward(localConn net.Conn, remotePort int) {
	defer localConn.Close()

	remoteConn, err := c.client.Dial("tcp", fmt.Sprintf("localhost:%d", remotePort))
	if err != nil {
		return
	}
	defer remoteConn.Close()

	c.forwardMutex.Lock()
	c.forwardConns[localConn] = struct{}{}
	c.forwardConns[remoteConn] = struct{}{}
	c.forwardMutex.Unlock()

	defer func() {
		c.forwardMutex.Lock()
		delete(c.forwardConns, localConn)
		delete(c.forwardConns, remoteConn)
		c.forwardMutex.Unlock()
	}()

	go io.Copy(remoteConn, localConn)
	go io.Copy(localConn, remoteConn)
}

func buildSSHClientConfig(user, authType, keyPath, passphrase, password string) (*ssh.ClientConfig, error) {
	sshCfg := &ssh.ClientConfig{
		User:            user,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         30 * time.Second,
	}

	switch authType {
	case "key":
		auths, err := buildKeyAuths(keyPath, passphrase)
		if err != nil {
			return nil, err
		}
		sshCfg.Auth = auths

	case "password":
		sshCfg.Auth = []ssh.AuthMethod{
			ssh.Password(password),
		}

	default:
		if password != "" {
			sshCfg.Auth = []ssh.AuthMethod{
				ssh.Password(password),
			}
		} else {
			auths, err := buildKeyAuths(keyPath, passphrase)
			if err != nil {
				return nil, err
			}
			sshCfg.Auth = auths
		}
	}

	return sshCfg, nil
}

func buildKeyAuths(keyPath, passphrase string) ([]ssh.AuthMethod, error) {
	var auths []ssh.AuthMethod

	agentAuth := sshAgentAuth()
	if agentAuth != nil {
		auths = append(auths, agentAuth)
	}

	if keyPath != "" {
		auth, err := keyAuth(keyPath, passphrase)
		if err != nil {
			return nil, err
		}
		auths = append(auths, auth)
	} else {
		defaultPaths := []string{
			filepath.Join(os.Getenv("HOME"), ".ssh", "id_rsa"),
			filepath.Join(os.Getenv("HOME"), ".ssh", "id_ed25519"),
			filepath.Join(os.Getenv("HOME"), ".ssh", "id_ecdsa"),
		}
		for _, p := range defaultPaths {
			if _, err := os.Stat(p); err == nil {
				auth, err := keyAuth(p, passphrase)
				if err != nil {
					if strings.Contains(err.Error(), "passphrase") {
						continue
					}
					return nil, err
				}
				auths = append(auths, auth)
				break
			}
		}
	}

	if len(auths) == 0 {
		return nil, fmt.Errorf("no SSH authentication method available: provide a key_path or verify your SSH agent is running")
	}

	return auths, nil
}

func sshAgentAuth() ssh.AuthMethod {
	sshAgentConn, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK"))
	if err != nil {
		return nil
	}
	sshAgent := agent.NewClient(sshAgentConn)
	return ssh.PublicKeysCallback(sshAgent.Signers)
}

func keyAuth(keyPath, passphrase string) (ssh.AuthMethod, error) {
	if strings.HasPrefix(keyPath, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("getting home directory: %w", err)
		}
		keyPath = filepath.Join(home, keyPath[2:])
	}

	keyBytes, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("reading private key %q: %w", keyPath, err)
	}

	var signer ssh.Signer
	if passphrase != "" {
		signer, err = ssh.ParsePrivateKeyWithPassphrase(keyBytes, []byte(passphrase))
		if err != nil {
			return nil, fmt.Errorf("parsing encrypted private key %q: %w", keyPath, err)
		}
	} else {
		signer, err = ssh.ParsePrivateKey(keyBytes)
		if err != nil {
			signer, err = tryInteractivePassphrase(keyBytes, keyPath)
			if err != nil {
				return nil, fmt.Errorf("parsing private key %q: %w", keyPath, err)
			}
		}
	}

	return ssh.PublicKeys(signer), nil
}

func tryInteractivePassphrase(keyBytes []byte, keyPath string) (ssh.Signer, error) {
	fmt.Fprintf(os.Stderr, "Private key %q is encrypted.\n", keyPath)
	fmt.Fprintf(os.Stderr, "Enter passphrase for key %q: ", keyPath)
	passBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Fprintln(os.Stderr)
	if err != nil {
		return nil, fmt.Errorf("reading passphrase: %w", err)
	}
	passphrase := string(passBytes)
	if passphrase == "" {
		return nil, fmt.Errorf("passphrase is required for encrypted key %q", keyPath)
	}
	return ssh.ParsePrivateKeyWithPassphrase(keyBytes, []byte(passphrase))
}

type aliasInfo struct {
	hostname     string
	identityFile string
	user         string
}

func resolveSSHAlias(alias string) (*aliasInfo, error) {
	sshConfigPath := filepath.Join(os.Getenv("HOME"), ".ssh", "config")
	info := &aliasInfo{hostname: alias}

	if _, err := os.Stat(sshConfigPath); os.IsNotExist(err) {
		return info, nil
	}

	data, err := os.ReadFile(sshConfigPath)
	if err != nil {
		return info, nil
	}

	parsed := parseSSHConfigBlock(string(data), alias)
	if parsed != nil {
		if parsed.hostname != "" {
			info.hostname = parsed.hostname
		}
		if parsed.identityFile != "" {
			info.identityFile = parsed.identityFile
		}
		if parsed.user != "" {
			info.user = parsed.user
		}
	}

	return info, nil
}

func parseSSHConfigBlock(configText, alias string) *aliasInfo {
	lines := strings.Split(configText, "\n")
	inHost := false
	result := &aliasInfo{}

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

		if strings.HasPrefix(lower, "hostname ") {
			result.hostname = strings.TrimSpace(line[9:])
		} else if strings.HasPrefix(lower, "identityfile ") {
			result.identityFile = expandTilde(strings.TrimSpace(line[13:]))
		} else if strings.HasPrefix(lower, "user ") {
			result.user = strings.TrimSpace(line[5:])
		}
	}

	return result
}

func expandTilde(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, _ := os.UserHomeDir()
		if home != "" {
			return filepath.Join(home, path[2:])
		}
	}
	return path
}

func matchHostPattern(pattern, alias string) bool {
	parts := strings.Split(pattern, " ")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if matched, _ := filepath.Match(part, alias); matched {
			return true
		}
	}
	return false
}
