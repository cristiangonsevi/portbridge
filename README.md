# 🚀 PortBridge

[![Go Report Card](https://goreportcard.com/badge/github.com/cristiangonsevi/portbridge)](https://goreportcard.com/report/github.com/cristangonsevi/portbridge)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
<!-- Add more badges as needed -->

---

## 🛠️ Technologies Used

- **Go**: Main language of the project.
- **Cobra**: Framework for building a robust and flexible command-line interface (CLI).
- **Native SSH**: Uses the system's SSH client, no external dependencies for tunnels.

> Manage SSH tunnels with profiles. Connect your VPS services to localhost with a single command.

---

## ⚡ Quickstart

```bash
# Install (Linux/macOS)
curl -fsSL https://raw.githubusercontent.com/cristiangonsevi/portbridge/refs/heads/master/scripts/install.sh | sh

# Start a profile
portbridge up qa
```

---

## ✨ What is PortBridge?

**PortBridge** is a CLI tool that simplifies SSH tunneling by letting you define **profiles (environments)** like `qa`, `prod`, or `dev`, and connect all required services (MySQL, Redis, APIs, etc.) with one command.

No more long `ssh -L` commands. No more remembering ports.

---

## 🔥 Features

* ⚡ **One command to connect everything**
* 🧠 **Profiles (qa, prod, dev, etc.)**
* 🔌 **Multiple tunnels per profile**
* ➕ **Add ports/tunnels to existing profiles easily**
* ➖ **Remove tunnels from profiles**
* ⏸️ **Enable/disable tunnels (mark as inactive)**
* 🎯 **Connect only what you need (persistent config)**
* 🔁 **Auto-reconnect (keeps tunnels alive if connection drops)**
* 🔐 **Flexible SSH configuration (alias or manual setup)**
* 📄 **YAML-based configuration**
* 🧩 **Simple and clean CLI**
* 📊 **Status & visibility of active tunnels**
* 🔒 Uses native SSH (no external servers)

---

## 📦 Installation

**Requirements:**
- Go 1.20+
- Linux, macOS (Windows support planned)

```bash
curl -fsSL https://raw.githubusercontent.com/cristiangonsevi/portbridge/refs/heads/master/scripts/install.sh | sh
```

---

## 📝 Example Configuration

```yaml
profiles:
  qa:
    ssh_alias: my-qa-server
    tunnels:
      - name: db
        local: 3306
        remote: 3306
        enabled: true
      - name: redis
        local: 6379
        remote: 6379
        enabled: false
```

---

## 🚀 Usage

### Start tunnels

```bash
portbridge up qa
```

### Stop tunnels

```bash
portbridge down qa
```

### Check status

```bash
portbridge status
```

```bash
portbridge status qa
```

### List profiles

```bash
portbridge list
```

---

## ❓ Troubleshooting & FAQ

- **Q:** SSH tunnel fails to connect?
  **A:** Check your SSH config, network, and that the remote ports are open.

- **Q:** How do I uninstall?
  **A:** Remove the binary (`rm $(which portbridge)`) and delete your config file if desired.

- **Q:** Where is the config stored?
  **A:** By default, in the current directory as `portbridge.yaml`.

---

## ➕ Managing tunnels (add / remove / toggle)

### Add a tunnel

```bash
portbridge add tunnel qa --name api --local 8080 --remote 8080
```

---

### Remove a tunnel

```bash
portbridge remove tunnel qa --name api
```

---

### Disable a tunnel (mark as inactive)

```bash
portbridge disable tunnel qa --name redis
```

---

### Enable it again

```bash
portbridge enable tunnel qa --name redis
```

---

👉 Disabled tunnels are **not connected** when running:

```bash
portbridge up qa
```

👉 This state is saved in your configuration.

---

## ⚙️ Configuration

PortBridge uses a YAML config file:

```
portbridge.yaml
```

---

## 🔐 SSH Configuration Options

You can configure profiles in two ways:

### 1. Using SSH Alias (recommended)

If you already have an SSH config (`~/.ssh/config`):

```yaml
profiles:
  qa:
    ssh_alias: my-qa-server
    tunnels:
      - name: db
        local: 3306
        remote: 3306
        enabled: true
```

👉 PortBridge will use your existing SSH settings automatically.

---

### 2. Manual Configuration

```yaml
profiles:
  qa:
    host: vps-qa.com
    port: 22
    user: ubuntu
    password: yourpassword # optional (not recommended)
    ssh_key_file: ~/.ssh/id_rsa
    tunnels:
      - name: db
        local: 3306
        remote: 3306
        enabled: true

      - name: redis
        local: 6379
        remote: 6379
        enabled: false
```

---

### 💡 Notes

* `ssh_alias` overrides manual config if provided
* `ssh_key_file` is recommended over password authentication
* Password support is optional and less secure

---

## 🔁 Auto-reconnect

PortBridge keeps your tunnels alive automatically.

If the SSH connection drops:

* it reconnects
* restores all **enabled** tunnels
* no manual action required

---

## 🧠 How it works

PortBridge wraps standard SSH tunneling:

```bash
ssh -L <local_port>:localhost:<remote_port> user@host
```

But adds:

* Profile management
* Multiple tunnels at once
* Enable/disable per tunnel
* Auto-reconnect
* Better developer experience

---

## 🧩 CLI Design

```bash
portbridge <command> [arguments] [flags]
```

### Commands

| Command                    | Description       |
| -------------------------- | ----------------- |
| `up <profile>`             | Start tunnels     |
| `down <profile>`           | Stop tunnels      |
| `status [profile]`         | Show status       |
| `list`                     | List all profiles |
| `add tunnel <profile>`     | Add new tunnel    |
| `remove tunnel <profile>`  | Remove tunnel     |
| `enable tunnel <profile>`  | Enable tunnel     |
| `disable tunnel <profile>` | Disable tunnel    |

---

### Flags

| Flag        | Description        |
| ----------- | ------------------ |
| `--config`  | Custom config file |
| `--verbose` | Enable logs        |
| `--detach`  | Run in background  |
| `--name`    | Tunnel name        |
| `--local`   | Local port         |
| `--remote`  | Remote port        |

---

## 📊 Example Output

```bash
$ portbridge up qa

✔ Connecting profile: qa
✔ db → localhost:3306
⏸ redis (disabled)
```

---

## 💡 Why PortBridge?

Managing SSH tunnels manually is painful:

* Long commands
* Easy to forget ports
* Hard to manage multiple services

PortBridge fixes that by making tunnels:

> **Declarative, reusable, and flexible**

---

## 🏗️ Project Structure (Go)

```
portbridge/
├── cmd/
│   ├── root.go
│   ├── up.go
│   ├── down.go
│   ├── status.go
│   ├── list.go
│   ├── add.go
│   ├── remove.go
│   ├── enable.go
│   ├── disable.go
│
├── internal/
│   ├── config/
│   │   ├── loader.go
│   │   ├── schema.go
│   │
│   ├── tunnel/
│   │   ├── manager.go
│   │   ├── ssh.go
│   │   ├── reconnect.go
│   │
│   ├── profiles/
│   │   ├── service.go
│   │
│   ├── ui/
│   │   ├── output.go
│
│
├── main.go
├── go.mod
└── portbridge.yaml
```

---

## 🧱 Key Components

### `cmd/`

CLI commands (using Cobra)

### `internal/config/`

* Load YAML
* Validate structure

### `internal/tunnel/`

* Create SSH tunnels
* Handle auto-reconnect
* Manage lifecycle

### `internal/profiles/`

* CRUD for profiles
* Add/remove/enable/disable tunnels

---

## 🛠️ Roadmap

* [ ] Auto-reconnect improvements (backoff, retries)
* [ ] Background daemon mode
* [ ] Profile validation
* [ ] Interactive CLI (`portbridge init`)
* [ ] Shell autocomplete
* [ ] Logs & monitoring
* [ ] Windows support (PowerShell installer)

---

## 🤝 Contributing

PRs are welcome!

1. Fork the repo
2. Create your branch
3. Submit a PR

---

## 📬 Contact & Support

- Issues: [GitHub Issues](https://github.com/cristiangonsevi/portbridge/issues)
- Email: cgonsevi@gmail.com

---

## 📄 License

MIT

---

## ⭐ Support

If you like this project, give it a star ⭐

---
