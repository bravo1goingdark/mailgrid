# 📦 MailGrid Installation Guide

Multiple installation methods available for all major platforms with optimized binaries.

## 🚀 Quick Installation

### **Linux & macOS (One-liner)**
```bash
curl -sSL https://raw.githubusercontent.com/bravo1goingdark/mailgrid/main/install.sh | bash
```

### **Windows (PowerShell)**

**🚀 Quick Install (Recommended)**
```powershell
# Basic installation
iwr -useb https://raw.githubusercontent.com/bravo1goingdark/mailgrid/main/install.ps1 | iex

# Enhanced installation with shortcuts and Windows Terminal integration
iwr -useb https://raw.githubusercontent.com/bravo1goingdark/mailgrid/main/install-enhanced.ps1 | iex
```

**⚙️ Advanced Installation Options**
```powershell
# Install with PATH integration and shortcuts
.\install-enhanced.ps1 -AddToPath -CreateShortcuts

# Install specific version with Windows Terminal profile
.\install-enhanced.ps1 -Version v1.0.0 -WindowsTerminalProfile

# Check for updates
.\install-enhanced.ps1 -CheckUpdates
```

### **Docker**
```bash
docker run --rm ghcr.io/bravo1goingdark/mailgrid:latest --help
```

## 📥 Direct Downloads

### **Windows**
| Architecture | Download | Size |
|--------------|----------|------|
| **x64** | [📥 mailgrid-windows-amd64.exe.zip](https://github.com/bravo1goingdark/mailgrid/releases/latest/download/mailgrid-windows-amd64.exe.zip) | ~3.7 MB |
| **ARM64** | [📥 mailgrid-windows-arm64.exe.zip](https://github.com/bravo1goingdark/mailgrid/releases/latest/download/mailgrid-windows-arm64.exe.zip) | ~3.3 MB |

### **macOS**
| Platform | Download | Size |
|----------|----------|------|
| **Intel Macs** | [📥 mailgrid-macos-intel.tar.gz](https://github.com/bravo1goingdark/mailgrid/releases/latest/download/mailgrid-macos-intel.tar.gz) | ~4.2 MB |
| **Apple Silicon** | [📥 mailgrid-macos-apple-silicon.tar.gz](https://github.com/bravo1goingdark/mailgrid/releases/latest/download/mailgrid-macos-apple-silicon.tar.gz) | ~4.0 MB |

### **Linux**
| Architecture | Download | Size |
|--------------|----------|------|
| **x64** | [📥 mailgrid-linux-amd64.tar.gz](https://github.com/bravo1goingdark/mailgrid/releases/latest/download/mailgrid-linux-amd64.tar.gz) | ~4.1 MB |
| **ARM64** | [📥 mailgrid-linux-arm64.tar.gz](https://github.com/bravo1goingdark/mailgrid/releases/latest/download/mailgrid-linux-arm64.tar.gz) | ~3.9 MB |
| **386** | [📥 mailgrid-linux-386.tar.gz](https://github.com/bravo1goingdark/mailgrid/releases/latest/download/mailgrid-linux-386.tar.gz) | ~3.8 MB |

### **FreeBSD**
| Architecture | Download | Size |
|--------------|----------|------|
| **x64** | [📥 mailgrid-freebsd-amd64.tar.gz](https://github.com/bravo1goingdark/mailgrid/releases/latest/download/mailgrid-freebsd-amd64.tar.gz) | ~4.1 MB |

## 🔐 Verify Downloads

All releases include SHA256 checksums for verification:
```bash
# Download checksums
curl -sSL https://github.com/bravo1goingdark/mailgrid/releases/latest/download/checksums.txt

# Verify (Linux/macOS)
sha256sum -c checksums.txt

# Verify (Windows PowerShell)
Get-FileHash mailgrid-windows-amd64.exe -Algorithm SHA256
```

## 📦 Package Managers

### **Go Install**
```bash
go install github.com/bravo1goingdark/mailgrid/cmd/mailgrid@latest
```

### **Homebrew (macOS/Linux)**
```bash
brew tap bravo1goingdark/tap
brew install mailgrid
```

### **Windows Package Managers**

**Winget (Windows Package Manager)**
```powershell
# Install MailGrid
winget install MailGrid.MailGrid

# Search for MailGrid
winget search mailgrid

# Upgrade MailGrid
winget upgrade MailGrid.MailGrid
```

**Chocolatey**
```powershell
# Install MailGrid
choco install mailgrid

# Upgrade MailGrid
choco upgrade mailgrid

# Uninstall MailGrid
choco uninstall mailgrid
```

**Scoop**
```powershell
# Add bucket (first time only)
scoop bucket add mailgrid https://github.com/bravo1goingdark/scoop-mailgrid

# Install MailGrid
scoop install mailgrid

# Update MailGrid
scoop update mailgrid
```

## 🐳 Docker Usage

### **Quick Start**
```bash
# Pull image
docker pull ghcr.io/bravo1goingdark/mailgrid:latest

# Run with config
docker run --rm -v $(pwd)/config.json:/app/config.json \
  ghcr.io/bravo1goingdark/mailgrid:latest \
  --to user@example.com --subject "Test" --text "Hello!"
```

### **Docker Compose**
```bash
# Download docker-compose.yml
curl -sSL https://raw.githubusercontent.com/bravo1goingdark/mailgrid/main/docker-compose.yml -o docker-compose.yml

# Start scheduler service
docker-compose up -d
```

### **Available Tags**
- `latest` - Latest stable release
- `v1.0.0` - Specific version
- `main` - Development version

## ⚙️ Configuration

MailGrid requires an SMTP configuration file to send emails. The config file path is specified using the `--env` flag.

### **Basic Configuration**

Create a `config.json` file:
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

### **Configuration File Locations**

**Using the example config:**
```bash
# Copy example config (from project directory)
cp example/config.json ./my-config.json

# Edit with your SMTP details
nano my-config.json

# Use with mailgrid
mailgrid --env my-config.json --to test@example.com --subject "Test" --text "Hello!"
```

**Common config file locations:**
- **Windows:** `C:\Users\YourName\mailgrid-config.json`
- **Linux/macOS:** `~/.config/mailgrid/config.json` or `~/mailgrid-config.json`
- **Project directory:** `./config.json`

### **SMTP Provider Examples**

**Gmail (App Password required):**
```json
{
  "smtp": {
    "host": "smtp.gmail.com",
    "port": 587,
    "username": "your-email@gmail.com",
    "password": "your-16-char-app-password",
    "from": "your-email@gmail.com"
  },
  "rate_limit": 10,
  "timeout_ms": 5000
}
```

**Outlook/Hotmail:**
```json
{
  "smtp": {
    "host": "smtp-mail.outlook.com",
    "port": 587,
    "username": "your-email@outlook.com",
    "password": "your-password",
    "from": "your-email@outlook.com"
  },
  "rate_limit": 5,
  "timeout_ms": 10000
}
```

**SendGrid:**
```json
{
  "smtp": {
    "host": "smtp.sendgrid.net",
    "port": 587,
    "username": "apikey",
    "password": "your-sendgrid-api-key",
    "from": "noreply@yourdomain.com"
  },
  "rate_limit": 100,
  "timeout_ms": 5000
}

## 🚀 Quick Test

After installation, test with:
```bash
mailgrid --to you@example.com \
         --subject "MailGrid Test" \
         --text "Hello from MailGrid v1.0.0!" \
         --env config.json
```

## 🔧 Build from Source

For advanced users or contributors:
```bash
# Clone repository
git clone https://github.com/bravo1goingdark/mailgrid.git
cd mailgrid

# Build current platform
make build

# Build all platforms
make release

# Run tests
make test

# Build Docker image
make docker-build
```

## 📊 System Requirements

### **Minimum Requirements**
- **OS**: Windows 10+, Linux (any modern distro), macOS 10.15+, FreeBSD 12+
- **Memory**: 50MB RAM (base usage)
- **Disk**: 20MB for binary + database storage
- **Network**: Internet access for SMTP connections

### **Recommended for Production**
- **Memory**: 200MB-1GB RAM (depending on volume)
- **CPU**: 2+ cores for high-throughput scenarios
- **Disk**: SSD recommended for database performance
- **Network**: Stable connection to SMTP providers

## 🆘 Troubleshooting

### **Configuration Errors**

**"failed to load config: open config "": no such file":**
```bash
# Error means --env flag is missing or config file doesn't exist
# Solution 1: Use --env flag with config file path
mailgrid --env ./config.json --to test@example.com --subject "Test" --text "Hello"

# Solution 2: Copy example config if in project directory
cp example/config.json ./my-config.json
mailgrid --env ./my-config.json --to test@example.com --subject "Test" --text "Hello"

# Solution 3: Create config file from scratch
echo '{
  "smtp": {
    "host": "smtp.gmail.com",
    "port": 587,
    "username": "your-email@gmail.com",
    "password": "your-app-password",
    "from": "your-email@gmail.com"
  },
  "rate_limit": 10,
  "timeout_ms": 5000
}' > config.json
```

**"decode config JSON" error:**
```bash
# JSON syntax error in config file
# Check for:
# - Missing commas
# - Extra commas
# - Unmatched quotes/braces
# - Invalid escape characters

# Validate JSON syntax online or with:
python -m json.tool config.json  # Python
jq . config.json               # jq tool
```

**"connection refused" or SMTP errors:**
```bash
# Test with dry-run first
mailgrid --env config.json --to test@example.com --subject "Test" --text "Hello" --dry-run

# Check SMTP settings:
# - Correct host and port
# - Valid username/password
# - Enable "Less secure app access" for Gmail or use App Password
# - Check firewall/network restrictions
```

### **Permission Errors (Linux/macOS)**
```bash
# Make binary executable
chmod +x mailgrid

# Install to system directory (requires sudo)
sudo cp mailgrid /usr/local/bin/
```

### **Windows Installation Issues**

**Execution Policy (PowerShell Script Blocked):**
```powershell
# Allow script execution
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser

# Unblock downloaded file
Unblock-File mailgrid.exe
```

**Manual Installation (if script fails):**
```powershell
# Create installation directory
$installDir = "$env:LOCALAPPDATA\mailgrid\bin"
New-Item -ItemType Directory -Force -Path $installDir

# Download and extract binary
$downloadUrl = "https://github.com/bravo1goingdark/mailgrid/releases/latest/download/mailgrid-windows-amd64.exe.zip"
Invoke-WebRequest -Uri $downloadUrl -OutFile "$env:TEMP\mailgrid.zip"
Expand-Archive -Path "$env:TEMP\mailgrid.zip" -DestinationPath $installDir -Force

# Add to PATH (current session)
$env:PATH += ";$installDir"

# Add to PATH (permanent - requires restart of terminal)
$currentPath = [Environment]::GetEnvironmentVariable("PATH", "User")
if ($currentPath -notlike "*$installDir*") {
    [Environment]::SetEnvironmentVariable("PATH", "$currentPath;$installDir", "User")
    Write-Host "✅ MailGrid added to PATH. Restart your terminal to use 'mailgrid' command."
}

# Test installation
mailgrid --help
```

**PATH Configuration Issues:**
```powershell
# Check if mailgrid is in PATH
Get-Command mailgrid -ErrorAction SilentlyContinue

# If not found, manually add the directory to PATH
$mailgridPath = "$env:LOCALAPPDATA\mailgrid\bin"
if (Test-Path $mailgridPath) {
    $env:PATH += ";$mailgridPath"
    Write-Host "✅ Added $mailgridPath to current session PATH"
}

# Verify installation
mailgrid --version
```

**"mailgrid not recognized as command" Error:**
```powershell
# Option 1: Use full path temporarily
& "$env:LOCALAPPDATA\mailgrid\bin\mailgrid.exe" --help

# Option 2: Refresh PATH in current session
$env:PATH = [System.Environment]::GetEnvironmentVariable("PATH","Machine") + ";" + [System.Environment]::GetEnvironmentVariable("PATH","User")

# Option 3: Restart PowerShell/Command Prompt
# Close and reopen your terminal
```

### **Docker Issues**
```bash
# Check if Docker is running
docker version

# Pull latest image
docker pull ghcr.io/bravo1goingdark/mailgrid:latest

# Check container logs
docker logs mailgrid-scheduler
```

## 📈 Performance Tips

### **High-Volume Sending**
```bash
# Increase concurrency and batch size
mailgrid --csv large-list.csv \
         --concurrency 10 \
         --batch-size 100 \
         --template newsletter.html
```

### **Memory Optimization**
```bash
# Limit connection pool for low-memory systems
export MAILGRID_MAX_CONNECTIONS=5
```

### **Monitoring**
```bash
# Check metrics endpoint
curl http://localhost:8090/metrics

# Monitor health
curl http://localhost:8090/health
```

## 🔗 Next Steps

After installation:
1. 📖 Read the [CLI Reference](docs/CLI_REFERENCE.md)
2. 🎯 Check [Usage Examples](example/)
3. 📊 Review [Performance Guide](PERFORMANCE_OPTIMIZATIONS.md)
4. 🐛 Report issues on [GitHub](https://github.com/bravo1goingdark/mailgrid/issues)

---

**Need Help?** 
- 📚 [Documentation](README.md)
- 💬 [GitHub Discussions](https://github.com/bravo1goingdark/mailgrid/discussions)
- 🐛 [Report Issues](https://github.com/bravo1goingdark/mailgrid/issues)