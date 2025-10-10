---
layout: default
title: Installation
nav_order: 3
has_children: false
---

# Installation Guide
{: .no_toc }

Multiple installation methods for all major platforms with optimized, lightweight binaries.
{: .fs-6 .fw-300 }

## Table of contents
{: .no_toc .text-delta }

1. TOC
{:toc}

---

## Quick Install (Recommended)

### Windows (PowerShell)
```powershell
iwr -useb https://raw.githubusercontent.com/bravo1goingdark/mailgrid/main/install.ps1 | iex
```

### Linux & macOS
```bash
curl -sSL https://raw.githubusercontent.com/bravo1goingdark/mailgrid/main/install.sh | bash
```

### Docker
```bash
docker run --rm ghcr.io/bravo1goingdark/mailgrid:latest --help
```

---

## Direct Downloads

Download pre-built binaries from [GitHub Releases](https://github.com/bravo1goingdark/mailgrid/releases/latest):

| Platform | Architecture | Download | Size |
|----------|--------------|----------|------|
| **Windows** | x64 | [📥 mailgrid-windows-amd64.exe.zip](https://github.com/bravo1goingdark/mailgrid/releases/latest/download/mailgrid-windows-amd64.exe.zip) | ~3.7 MB |
| **Windows** | ARM64 | [📥 mailgrid-windows-arm64.exe.zip](https://github.com/bravo1goingdark/mailgrid/releases/latest/download/mailgrid-windows-arm64.exe.zip) | ~3.3 MB |
| **macOS** | Intel | [📥 mailgrid-macos-intel.tar.gz](https://github.com/bravo1goingdark/mailgrid/releases/latest/download/mailgrid-macos-intel.tar.gz) | ~4.2 MB |
| **macOS** | Apple Silicon | [📥 mailgrid-macos-apple-silicon.tar.gz](https://github.com/bravo1goingdark/mailgrid/releases/latest/download/mailgrid-macos-apple-silicon.tar.gz) | ~4.0 MB |
| **Linux** | x64 | [📥 mailgrid-linux-amd64.tar.gz](https://github.com/bravo1goingdark/mailgrid/releases/latest/download/mailgrid-linux-amd64.tar.gz) | ~4.1 MB |
| **Linux** | ARM64 | [📥 mailgrid-linux-arm64.tar.gz](https://github.com/bravo1goingdark/mailgrid/releases/latest/download/mailgrid-linux-arm64.tar.gz) | ~3.9 MB |
| **FreeBSD** | x64 | [📥 mailgrid-freebsd-amd64.tar.gz](https://github.com/bravo1goingdark/mailgrid/releases/latest/download/mailgrid-freebsd-amd64.tar.gz) | ~4.1 MB |

---

## Manual Installation

### Windows

1. Download the appropriate `.zip` file for your architecture
2. Extract `mailgrid.exe` to a folder (e.g., `C:\tools\mailgrid\`)
3. Add the folder to your system PATH:
   - Open "Environment Variables" in System Properties
   - Add the folder path to the `PATH` variable
4. Verify installation:
   ```cmd
   mailgrid --help
   ```

### Linux/macOS

1. Download and extract the `.tar.gz` file:
   ```bash
   tar -xzf mailgrid-*.tar.gz
   sudo mv mailgrid /usr/local/bin/
   chmod +x /usr/local/bin/mailgrid
   ```

2. Verify installation:
   ```bash
   mailgrid --help
   ```

---

## Package Managers

### Homebrew (macOS/Linux)
```bash
# Coming soon
brew install mailgrid
```

### Chocolatey (Windows)
```powershell
# Coming soon
choco install mailgrid
```

### Scoop (Windows)
```powershell
# Coming soon
scoop install mailgrid
```

---

## Docker Usage

### Quick Test
```bash
docker run --rm -v $(pwd)/config.json:/app/config.json \
  ghcr.io/bravo1goingdark/mailgrid:latest \
  --env /app/config.json --to test@example.com --subject "Test" --text "Hello!"
```

### With Docker Compose

1. Download the compose file:
   ```bash
   curl -sSL https://raw.githubusercontent.com/bravo1goingdark/mailgrid/main/docker-compose.yml -o docker-compose.yml
   ```

2. Start the container:
   ```bash
   docker-compose up -d
   ```

3. Use MailGrid:
   ```bash
   docker-compose exec mailgrid mailgrid --help
   ```

### Custom Docker Setup

**Dockerfile:**
```dockerfile
FROM ghcr.io/bravo1goingdark/mailgrid:latest
COPY config.json /app/config.json
COPY templates/ /app/templates/
COPY data/ /app/data/
WORKDIR /app
ENTRYPOINT ["mailgrid"]
```

**Build and run:**
```bash
docker build -t my-mailgrid .
docker run --rm my-mailgrid --env config.json --csv data/recipients.csv --template templates/email.html
```

---

## Build from Source

### Prerequisites

- Go 1.21 or later
- Git

### Build Steps

1. Clone the repository:
   ```bash
   git clone https://github.com/bravo1goingdark/mailgrid.git
   cd mailgrid
   ```

2. Build the binary:
   ```bash
   go build -o mailgrid ./cmd/mailgrid
   ```

3. Install globally (optional):
   ```bash
   go install ./cmd/mailgrid
   ```

### Cross-Compilation

Build for different platforms:

```bash
# Windows
GOOS=windows GOARCH=amd64 go build -o mailgrid.exe ./cmd/mailgrid

# macOS
GOOS=darwin GOARCH=amd64 go build -o mailgrid-macos ./cmd/mailgrid

# Linux
GOOS=linux GOARCH=amd64 go build -o mailgrid-linux ./cmd/mailgrid
```

---

## Configuration Setup

After installation, create your SMTP configuration:

### Basic Configuration

Create `config.json`:
```json
{
  "smtp": {
    "host": "smtp.gmail.com",
    "port": 587,
    "username": "your-email@gmail.com",
    "password": "your-app-password",
    "from": "your-email@gmail.com"
  },
  "rate_limit": 10,
  "timeout_ms": 5000
}
```

### Gmail Setup

1. Enable 2-Factor Authentication
2. Generate an [App Password](https://myaccount.google.com/apppasswords)
3. Use the app password in your config

### Other Providers

**Outlook/Hotmail:**
```json
{
  "smtp": {
    "host": "smtp-mail.outlook.com",
    "port": 587,
    "username": "your-email@outlook.com",
    "password": "your-password",
    "from": "your-email@outlook.com"
  }
}
```

**Yahoo:**
```json
{
  "smtp": {
    "host": "smtp.mail.yahoo.com",
    "port": 587,
    "username": "your-email@yahoo.com",
    "password": "your-app-password",
    "from": "your-email@yahoo.com"
  }
}
```

---

## Verification

Test your installation:

```bash
# Version check
mailgrid --version

# Help information
mailgrid --help

# Dry run test
mailgrid --env config.json \
         --to test@example.com \
         --subject "MailGrid Test" \
         --text "Hello from MailGrid!" \
         --dry-run
```

---

## Troubleshooting

### Command Not Found

**Windows:**
- Ensure `mailgrid.exe` is in your PATH
- Try running with full path: `C:\path\to\mailgrid.exe`

**Linux/macOS:**
- Check if `/usr/local/bin` is in your PATH: `echo $PATH`
- Verify executable permissions: `ls -la /usr/local/bin/mailgrid`

### Permission Denied

**Linux/macOS:**
```bash
chmod +x /usr/local/bin/mailgrid
```

**Windows:**
- Run PowerShell as Administrator
- Check antivirus software

### SSL/TLS Errors

Update your system's root certificates:

**Windows:** Windows Update usually handles this

**macOS:**
```bash
brew install ca-certificates
```

**Linux (Ubuntu/Debian):**
```bash
sudo apt-get update && sudo apt-get install ca-certificates
```

### SMTP Connection Issues

1. **Check credentials**: Verify username/password in config
2. **Test port access**: `telnet smtp.gmail.com 587`
3. **Firewall settings**: Ensure outbound SMTP is allowed
4. **Provider settings**: Check if SMTP access is enabled

---

## Uninstallation

### Installed via Script

**Windows:**
```powershell
Remove-Item "C:\mailgrid\mailgrid.exe"
# Remove from PATH manually
```

**Linux/macOS:**
```bash
sudo rm /usr/local/bin/mailgrid
```

### Installed via Package Manager

```bash
# When available
brew uninstall mailgrid
choco uninstall mailgrid
scoop uninstall mailgrid
```

### Docker

```bash
docker rmi ghcr.io/bravo1goingdark/mailgrid:latest
```

---

## Update

### Automatic Update (Coming Soon)

```bash
mailgrid --update
```

### Manual Update

1. Download the latest release
2. Replace the existing binary
3. Verify the new version: `mailgrid --version`

### Docker Update

```bash
docker pull ghcr.io/bravo1goingdark/mailgrid:latest
```

---

## Next Steps

- [Getting Started](getting-started) - Your first email campaign
- [CLI Reference](cli-reference) - Complete command documentation
- [Configuration Guide](configuration) - Advanced SMTP setup