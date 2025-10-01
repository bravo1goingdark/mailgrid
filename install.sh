#!/bin/bash

# MailGrid Installation Script
# Automatically detects platform and downloads the appropriate binary
# Usage: curl -sSL https://raw.githubusercontent.com/bravo1goingdark/mailgrid/main/install.sh | bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
REPO="bravo1goingdark/mailgrid"
BINARY_NAME="mailgrid"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
VERSION="${VERSION:-latest}"

# Print functions
print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
    exit 1
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_header() {
    echo
    echo -e "${BLUE}📬 MailGrid Installation Script${NC}"
    echo -e "${BLUE}================================${NC}"
    echo
}

# Detect platform
detect_platform() {
    local os arch
    
    # Detect OS
    case "$(uname -s)" in
        Linux*)  os="linux" ;;
        Darwin*) os="macos" ;;
        CYGWIN*|MINGW*|MSYS*) os="windows" ;;
        FreeBSD*) os="freebsd" ;;
        *) print_error "Unsupported operating system: $(uname -s)" ;;
    esac
    
    # Detect architecture
    case "$(uname -m)" in
        x86_64|amd64) arch="amd64" ;;
        aarch64|arm64) arch="arm64" ;;
        i386|i686) arch="386" ;;
        armv7l) arch="arm" ;;
        *) print_error "Unsupported architecture: $(uname -m)" ;;
    esac
    
    # Handle macOS special naming
    if [ "$os" = "macos" ]; then
        if [ "$arch" = "amd64" ]; then
            PLATFORM="macos-intel"
        else
            PLATFORM="macos-apple-silicon"
        fi
    else
        PLATFORM="${os}-${arch}"
    fi
    
    print_info "Detected platform: $PLATFORM"
}

# Get latest version from GitHub API
get_latest_version() {
    if [ "$VERSION" = "latest" ]; then
        print_info "Fetching latest version..."
        if command -v curl >/dev/null 2>&1; then
            VERSION=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name"' | sed 's/.*"tag_name": *"\\([^"]*\\)".*/\\1/')
        elif command -v wget >/dev/null 2>&1; then
            VERSION=$(wget -qO- "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name"' | sed 's/.*"tag_name": *"\\([^"]*\\)".*/\\1/')
        else
            print_error "Neither curl nor wget is available. Please install one of them."
        fi
        
        if [ -z "$VERSION" ] || [ "$VERSION" = "null" ]; then
            print_error "Failed to get latest version. Please check your internet connection."
        fi
    fi
    
    print_info "Installing version: $VERSION"
}

# Download and install binary
download_and_install() {
    local download_url temp_dir binary_path archive_name
    
    # Determine file extension and binary name
    if echo "$PLATFORM" | grep -q "windows"; then
        archive_name="${BINARY_NAME}-${PLATFORM}.exe.zip"
        binary_name="${BINARY_NAME}.exe"
    else
        archive_name="${BINARY_NAME}-${PLATFORM}.tar.gz"
        binary_name="$BINARY_NAME"
    fi
    
    download_url="https://github.com/$REPO/releases/download/$VERSION/$archive_name"
    temp_dir=$(mktemp -d)
    
    print_info "Downloading $archive_name..."
    
    # Download the archive
    if command -v curl >/dev/null 2>&1; then
        if ! curl -L --fail "$download_url" -o "$temp_dir/$archive_name"; then
            print_error "Failed to download $archive_name from $download_url"
        fi
    elif command -v wget >/dev/null 2>&1; then
        if ! wget -O "$temp_dir/$archive_name" "$download_url"; then
            print_error "Failed to download $archive_name from $download_url"
        fi
    else
        print_error "Neither curl nor wget is available. Please install one of them."
    fi
    
    print_success "Downloaded $archive_name"
    
    # Extract the archive
    cd "$temp_dir"
    if echo "$archive_name" | grep -q "\.zip$"; then
        if command -v unzip >/dev/null 2>&1; then
            unzip -q "$archive_name"
        else
            print_error "unzip is not available. Please install it to extract Windows binaries."
        fi
    else
        tar -xzf "$archive_name"
    fi
    
    # Find the binary (it might have a different name in the archive)
    binary_path=$(find . -name "$BINARY_NAME*" -type f | head -1)
    if [ ! -f "$binary_path" ]; then
        print_error "Binary not found in archive"
    fi
    
    # Make sure install directory exists
    if [ ! -d "$INSTALL_DIR" ]; then
        print_info "Creating install directory: $INSTALL_DIR"
        if ! mkdir -p "$INSTALL_DIR"; then
            print_error "Failed to create install directory. You may need to run with sudo."
        fi
    fi
    
    # Install the binary
    print_info "Installing to $INSTALL_DIR/$BINARY_NAME"
    if ! cp "$binary_path" "$INSTALL_DIR/$BINARY_NAME"; then
        print_error "Failed to install binary. You may need to run with sudo."
    fi
    
    # Make it executable
    chmod +x "$INSTALL_DIR/$BINARY_NAME"
    
    # Cleanup
    rm -rf "$temp_dir"
    
    print_success "Installed MailGrid $VERSION to $INSTALL_DIR/$BINARY_NAME"
}

# Verify installation
verify_installation() {
    if [ ! -x "$INSTALL_DIR/$BINARY_NAME" ]; then
        print_error "Installation verification failed: $INSTALL_DIR/$BINARY_NAME is not executable"
    fi
    
    # Check if it's in PATH
    if ! command -v "$BINARY_NAME" >/dev/null 2>&1; then
        print_warning "$INSTALL_DIR is not in your PATH"
        print_info "Add this to your shell profile (~/.bashrc, ~/.zshrc, etc.):"
        echo "export PATH=\"\$PATH:$INSTALL_DIR\""
        echo
        print_info "Or run MailGrid using the full path: $INSTALL_DIR/$BINARY_NAME"
    else
        print_success "$BINARY_NAME is available in your PATH"
    fi
}

# Show usage instructions
show_usage() {
    echo
    print_info "MailGrid has been successfully installed!"
    echo
    echo "📚 Quick Start:"
    echo "  1. Create a config.json file with your SMTP settings"
    echo "  2. Send a test email: $BINARY_NAME --to you@example.com --subject 'Test' --text 'Hello!' --env config.json"
    echo "  3. View all options: $BINARY_NAME --help"
    echo
    echo "🔗 Documentation:"
    echo "  • GitHub: https://github.com/$REPO"
    echo "  • CLI Reference: https://github.com/$REPO/blob/main/docs/CLI_REFERENCE.md"
    echo "  • Examples: https://github.com/$REPO/tree/main/example"
    echo
    echo "🚀 Example commands:"
    echo "  # Send single email"
    echo "  $BINARY_NAME --to user@example.com --subject 'Welcome' --text 'Hello!' --env config.json"
    echo
    echo "  # Send bulk emails from CSV"
    echo "  $BINARY_NAME --csv recipients.csv --template email.html --subject 'Newsletter' --env config.json"
    echo
    echo "  # Schedule recurring newsletter"
    echo "  $BINARY_NAME --csv subscribers.csv --template newsletter.html --cron '0 9 * * 1' --env config.json"
    echo
    echo "  # Run scheduler daemon"
    echo "  $BINARY_NAME --scheduler-run --env config.json"
    echo
}

# Main installation function
main() {
    print_header
    
    # Check if running as root (warn for security)
    if [ "$(id -u)" = "0" ]; then
        print_warning "Running as root. Consider using a regular user account."
    fi
    
    # Detect platform
    detect_platform
    
    # Get version to install
    get_latest_version
    
    # Download and install
    download_and_install
    
    # Verify installation
    verify_installation
    
    # Show usage instructions
    show_usage
}

# Handle script arguments
case "${1:-}" in
    --help|-h)
        echo "MailGrid Installation Script"
        echo
        echo "Usage: $0 [options]"
        echo
        echo "Options:"
        echo "  --help, -h          Show this help message"
        echo "  --version VERSION   Install specific version (default: latest)"
        echo "  --dir DIRECTORY     Install directory (default: /usr/local/bin)"
        echo
        echo "Environment variables:"
        echo "  VERSION             Version to install (default: latest)"
        echo "  INSTALL_DIR         Installation directory (default: /usr/local/bin)"
        echo
        echo "Examples:"
        echo "  $0                           # Install latest version"
        echo "  $0 --version v1.0.0          # Install specific version"
        echo "  VERSION=v1.0.0 $0            # Install using environment variable"
        echo "  INSTALL_DIR=~/.local/bin $0  # Install to custom directory"
        exit 0
        ;;
    --version)
        VERSION="$2"
        shift 2
        ;;
    --dir)
        INSTALL_DIR="$2"
        shift 2
        ;;
    "")
        # No arguments, proceed with installation
        ;;
    *)
        print_error "Unknown option: $1. Use --help for usage information."
        ;;
esac

# Run main function
main "$@"