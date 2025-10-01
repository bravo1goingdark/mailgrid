# 📦 MailGrid Installation Guide

Multiple installation methods available for all major platforms with optimized binaries.

## 🚀 Quick Installation

### **Linux & macOS (One-liner)**
```bash
curl -sSL https://raw.githubusercontent.com/bravo1goingdark/mailgrid/main/install.sh | bash
```

### **Windows (PowerShell)**
```powershell
iwr -useb https://raw.githubusercontent.com/bravo1goingdark/mailgrid/main/install.ps1 | iex
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

### **Chocolatey (Windows)**
```powershell
choco install mailgrid
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

Create a `config.json` file:
```json
{
  "smtp": {
    "host": "smtp.gmail.com",
    "port": 587,
    "username": "your-email@gmail.com",
    "password": "your-app-password",
    "from_email": "your-email@gmail.com",
    "from_name": "Your Name",
    "use_tls": true
  }
}
```

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

### **Permission Errors (Linux/macOS)**
```bash
# Make binary executable
chmod +x mailgrid

# Install to system directory (requires sudo)
sudo cp mailgrid /usr/local/bin/
```

### **Windows Execution Policy**
```powershell
# Allow script execution
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser

# Unblock downloaded file
Unblock-File mailgrid.exe
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