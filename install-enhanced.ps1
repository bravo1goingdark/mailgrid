# MailGrid Enhanced Windows Installation Script
# Features: Windows Terminal integration, Start Menu shortcuts, update checking
# Usage: iwr -useb https://raw.githubusercontent.com/bravo1goingdark/mailgrid/main/install-enhanced.ps1 | iex

param(
    [string]$Version = "latest",
    [string]$InstallDir = "$env:LOCALAPPDATA\mailgrid\bin",
    [switch]$AddToPath,
    [switch]$CreateShortcuts,
    [switch]$WindowsTerminalProfile,
    [switch]$CheckUpdates,
    [switch]$Help
)

# Configuration
$Repo = "bravo1goingdark/mailgrid"
$BinaryName = "mailgrid"
$AppName = "MailGrid"

# Show help
if ($Help) {
    Write-Host @"
MailGrid Enhanced Windows Installation Script

Usage: .\install-enhanced.ps1 [options]

Options:
  -Version <string>           Version to install (default: latest)
  -InstallDir <path>          Installation directory (default: $env:LOCALAPPDATA\mailgrid\bin)
  -AddToPath                  Add installation directory to user PATH
  -CreateShortcuts            Create Start Menu and Desktop shortcuts
  -WindowsTerminalProfile     Add MailGrid profile to Windows Terminal
  -CheckUpdates               Check for updates on existing installation
  -Help                       Show this help message

Examples:
  .\install-enhanced.ps1                                    # Basic installation
  .\install-enhanced.ps1 -Version v1.0.0                   # Install specific version
  .\install-enhanced.ps1 -AddToPath -CreateShortcuts       # Full installation with shortcuts
  .\install-enhanced.ps1 -CheckUpdates                     # Check for updates

Environment variables:
  MAILGRID_VERSION          Version to install (overrides -Version)
  MAILGRID_INSTALL_DIR      Installation directory (overrides -InstallDir)
"@
    exit 0
}

# Override with environment variables if set
if ($env:MAILGRID_VERSION) { $Version = $env:MAILGRID_VERSION }
if ($env:MAILGRID_INSTALL_DIR) { $InstallDir = $env:MAILGRID_INSTALL_DIR }

function Write-ColorOutput {
    param(
        [string]$Message,
        [ConsoleColor]$ForegroundColor = [ConsoleColor]::White
    )
    
    $originalColor = $Host.UI.RawUI.ForegroundColor
    $Host.UI.RawUI.ForegroundColor = $ForegroundColor
    Write-Host $Message
    $Host.UI.RawUI.ForegroundColor = $originalColor
}

function Write-Info { param([string]$Message) Write-ColorOutput "ℹ️  $Message" -ForegroundColor Blue }
function Write-Success { param([string]$Message) Write-ColorOutput "✅ $Message" -ForegroundColor Green }
function Write-Error { param([string]$Message) Write-ColorOutput "❌ $Message" -ForegroundColor Red }
function Write-Warning { param([string]$Message) Write-ColorOutput "⚠️  $Message" -ForegroundColor Yellow }

function Write-Header {
    Write-Host ""
    Write-ColorOutput "📬 MailGrid Enhanced Windows Installation Script" -ForegroundColor Blue
    Write-ColorOutput "=================================================" -ForegroundColor Blue
    Write-Host ""
}

function Get-Architecture {
    $arch = $env:PROCESSOR_ARCHITECTURE
    switch ($arch) {
        "AMD64" { return "amd64" }
        "ARM64" { return "arm64" }
        "x86" { return "386" }
        default { 
            Write-Error "Unsupported architecture: $arch"
            exit 1
        }
    }
}

function Get-InstalledVersion {
    $binaryPath = Join-Path $InstallDir "$BinaryName.exe"
    if (Test-Path $binaryPath) {
        try {
            $result = & $binaryPath --version 2>&1
            if ($result -match "v?(\d+\.\d+\.\d+)") {
                return $matches[1]
            }
        }
        catch {
            return $null
        }
    }
    return $null
}

function Get-LatestVersion {
    if ($Version -eq "latest") {
        Write-Info "Fetching latest version from GitHub..."
        try {
            $response = Invoke-RestMethod -Uri "https://api.github.com/repos/$Repo/releases/latest" -UseBasicParsing
            $script:Version = $response.tag_name
            Write-Info "Latest version: $Version"
        }
        catch {
            Write-Error "Failed to fetch latest version: $($_.Exception.Message)"
            exit 1
        }
    }
    else {
        Write-Info "Installing version: $Version"
    }
}

function Test-Administrator {
    $currentUser = [Security.Principal.WindowsIdentity]::GetCurrent()
    $principal = [Security.Principal.WindowsPrincipal]::new($currentUser)
    return $principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
}

function Test-UpdatesAvailable {
    if ($CheckUpdates) {
        Write-Info "Checking for updates..."
        $installedVersion = Get-InstalledVersion
        
        if ($installedVersion) {
            Get-LatestVersion
            $latestVersion = $Version -replace '^v', ''
            
            Write-Info "Installed version: $installedVersion"
            Write-Info "Latest version: $latestVersion"
            
            if ($installedVersion -ne $latestVersion) {
                Write-Warning "Update available: $installedVersion → $latestVersion"
                $response = Read-Host "Update to latest version? (y/N)"
                return ($response -eq 'y' -or $response -eq 'Y')
            }
            else {
                Write-Success "MailGrid is up to date!"
                return $false
            }
        }
        else {
            Write-Info "MailGrid not found, proceeding with fresh installation"
            return $true
        }
    }
    return $true
}

function Download-And-Install {
    $architecture = Get-Architecture
    $archiveName = "$BinaryName-windows-$architecture.exe.zip"
    $downloadUrl = "https://github.com/$Repo/releases/download/$Version/$archiveName"
    
    Write-Info "Downloading $archiveName..."
    Write-Info "URL: $downloadUrl"
    
    # Create temporary directory
    $tempDir = [System.IO.Path]::GetTempPath() + [System.Guid]::NewGuid().ToString()
    New-Item -ItemType Directory -Path $tempDir -Force | Out-Null
    
    $archivePath = Join-Path $tempDir $archiveName
    
    try {
        # Download the archive with progress
        $ProgressPreference = 'Continue'
        Invoke-WebRequest -Uri $downloadUrl -OutFile $archivePath -UseBasicParsing
        Write-Success "Downloaded $archiveName"
        
        # Extract the archive
        Write-Info "Extracting archive..."
        Expand-Archive -Path $archivePath -DestinationPath $tempDir -Force
        
        # Find the binary
        $binaryPath = Get-ChildItem -Path $tempDir -Filter "$BinaryName*.exe" | Select-Object -First 1
        if (-not $binaryPath) {
            Write-Error "Binary not found in archive"
            exit 1
        }
        
        # Create installation directory
        if (-not (Test-Path $InstallDir)) {
            Write-Info "Creating installation directory: $InstallDir"
            New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
        }
        
        # Install the binary
        $destinationPath = Join-Path $InstallDir "$BinaryName.exe"
        Write-Info "Installing to $destinationPath"
        Copy-Item -Path $binaryPath.FullName -Destination $destinationPath -Force
        
        Write-Success "Installed MailGrid $Version to $destinationPath"
        
        # Show file size
        $fileSize = [math]::Round((Get-Item $destinationPath).Length / 1MB, 2)
        Write-Info "Binary size: $fileSize MB"
        
        # Store installation metadata
        $metadataPath = Join-Path $InstallDir "install-info.json"
        $metadata = @{
            version = $Version
            installDate = (Get-Date).ToString("yyyy-MM-ddTHH:mm:ssZ")
            architecture = $architecture
            installDir = $InstallDir
        } | ConvertTo-Json
        $metadata | Out-File -FilePath $metadataPath -Encoding UTF8
        
    }
    catch {
        Write-Error "Installation failed: $($_.Exception.Message)"
        exit 1
    }
    finally {
        # Cleanup
        if (Test-Path $tempDir) {
            Remove-Item -Path $tempDir -Recurse -Force -ErrorAction SilentlyContinue
        }
    }
}

function Add-ToPath {
    if ($AddToPath) {
        Write-Info "Adding $InstallDir to user PATH..."
        
        # Get current user PATH
        $userPath = [Environment]::GetEnvironmentVariable("PATH", [EnvironmentVariableTarget]::User)
        
        # Check if already in PATH
        if ($userPath -split ';' -contains $InstallDir) {
            Write-Info "Directory already in PATH"
        }
        else {
            # Add to PATH
            $newPath = if ($userPath) { "$userPath;$InstallDir" } else { $InstallDir }
            [Environment]::SetEnvironmentVariable("PATH", $newPath, [EnvironmentVariableTarget]::User)
            
            # Update current session PATH
            $env:PATH = "$env:PATH;$InstallDir"
            
            Write-Success "Added to user PATH (restart terminal to take effect)"
        }
    }
}

function New-Shortcuts {
    if ($CreateShortcuts) {
        Write-Info "Creating shortcuts..."
        
        $binaryPath = Join-Path $InstallDir "$BinaryName.exe"
        $shell = New-Object -ComObject WScript.Shell
        
        # Start Menu shortcut
        $startMenuPath = [Environment]::GetFolderPath("StartMenu")
        $shortcutPath = Join-Path $startMenuPath "$AppName.lnk"
        
        try {
            $shortcut = $shell.CreateShortcut($shortcutPath)
            $shortcut.TargetPath = $binaryPath
            $shortcut.Arguments = "--help"
            $shortcut.Description = "MailGrid - Production-ready CLI email orchestrator"
            $shortcut.WorkingDirectory = $env:USERPROFILE
            $shortcut.Save()
            Write-Success "Created Start Menu shortcut"
        }
        catch {
            Write-Warning "Failed to create Start Menu shortcut: $($_.Exception.Message)"
        }
        
        # Desktop shortcut (optional)
        $response = Read-Host "Create Desktop shortcut? (y/N)"
        if ($response -eq 'y' -or $response -eq 'Y') {
            try {
                $desktopPath = [Environment]::GetFolderPath("Desktop")
                $desktopShortcutPath = Join-Path $desktopPath "$AppName.lnk"
                
                $shortcut = $shell.CreateShortcut($desktopShortcutPath)
                $shortcut.TargetPath = $binaryPath
                $shortcut.Arguments = "--help"
                $shortcut.Description = "MailGrid - Production-ready CLI email orchestrator"
                $shortcut.WorkingDirectory = $env:USERPROFILE
                $shortcut.Save()
                Write-Success "Created Desktop shortcut"
            }
            catch {
                Write-Warning "Failed to create Desktop shortcut: $($_.Exception.Message)"
            }
        }
    }
}

function Add-WindowsTerminalProfile {
    if ($WindowsTerminalProfile) {
        Write-Info "Adding Windows Terminal profile..."
        
        # Find Windows Terminal settings path
        $wtSettingsPath = "$env:LOCALAPPDATA\Packages\Microsoft.WindowsTerminal_8wekyb3d8bbwe\LocalState\settings.json"
        
        if (-not (Test-Path $wtSettingsPath)) {
            Write-Warning "Windows Terminal not found, skipping profile creation"
            return
        }
        
        try {
            $settings = Get-Content $wtSettingsPath -Raw | ConvertFrom-Json
            
            # Check if MailGrid profile already exists
            $existingProfile = $settings.profiles.list | Where-Object { $_.name -eq "MailGrid" }
            
            if (-not $existingProfile) {
                $binaryPath = Join-Path $InstallDir "$BinaryName.exe"
                
                # Create new profile
                $newProfile = @{
                    guid = [System.Guid]::NewGuid().ToString()
                    name = "MailGrid"
                    commandline = "powershell.exe -NoExit -Command `"Write-Host 'MailGrid Email Orchestrator' -ForegroundColor Green; Write-Host 'Type: mailgrid --help for usage information' -ForegroundColor Yellow; Write-Host ''; cd '$env:USERPROFILE'`""
                    icon = "📬"
                    tabTitle = "MailGrid"
                    colorScheme = "Campbell"
                }
                
                # Add to profiles list
                $settings.profiles.list += $newProfile
                
                # Save settings
                $settings | ConvertTo-Json -Depth 10 | Out-File $wtSettingsPath -Encoding UTF8
                Write-Success "Added Windows Terminal profile"
            }
            else {
                Write-Info "Windows Terminal profile already exists"
            }
        }
        catch {
            Write-Warning "Failed to add Windows Terminal profile: $($_.Exception.Message)"
        }
    }
}

function Test-Installation {
    $binaryPath = Join-Path $InstallDir "$BinaryName.exe"
    
    if (-not (Test-Path $binaryPath)) {
        Write-Error "Installation verification failed: $binaryPath not found"
        exit 1
    }
    
    # Test if binary is executable
    try {
        $result = & $binaryPath --help 2>&1
        if ($LASTEXITCODE -eq 2) {  # Help command returns exit code 2
            Write-Success "Binary is working correctly"
        }
        else {
            Write-Warning "Binary test returned unexpected exit code: $LASTEXITCODE"
        }
    }
    catch {
        Write-Error "Failed to execute binary: $($_.Exception.Message)"
        exit 1
    }
}

function Show-Usage {
    Write-Host ""
    Write-Info "MailGrid has been successfully installed!"
    Write-Host ""
    
    $binaryPath = Join-Path $InstallDir "$BinaryName.exe"
    $inPath = ($env:PATH -split ';') -contains $InstallDir
    
    Write-Host "📚 Quick Start:"
    Write-Host "  1. Create a config.json file with your SMTP settings"
    if ($inPath) {
        Write-Host "  2. Send a test email: mailgrid --to you@example.com --subject 'Test' --text 'Hello!' --env config.json"
        Write-Host "  3. View all options: mailgrid --help"
    }
    else {
        Write-Host "  2. Send a test email: `"$binaryPath`" --to you@example.com --subject 'Test' --text 'Hello!' --env config.json"
        Write-Host "  3. View all options: `"$binaryPath`" --help"
    }
    Write-Host ""
    
    if (-not $inPath -and -not $AddToPath) {
        Write-Warning "MailGrid is not in your PATH"
        Write-Info "To add it to PATH, run: .\install-enhanced.ps1 -AddToPath"
        Write-Info "Or use the full path: `"$binaryPath`""
        Write-Host ""
    }
    
    Write-Host "🔗 Documentation:"
    Write-Host "  • GitHub: https://github.com/$Repo"
    Write-Host "  • CLI Reference: https://github.com/$Repo/blob/main/docs/CLI_REFERENCE.md"
    Write-Host "  • Examples: https://github.com/$Repo/tree/main/example"
    Write-Host ""
    
    Write-Host "🚀 Example commands:"
    $cmd = if ($inPath) { "mailgrid" } else { "`"$binaryPath`"" }
    
    Write-Host "  # Send single email"
    Write-Host "  $cmd --to user@example.com --subject 'Welcome' --text 'Hello!' --env config.json"
    Write-Host ""
    Write-Host "  # Send bulk emails from CSV"
    Write-Host "  $cmd --csv recipients.csv --template email.html --subject 'Newsletter' --env config.json"
    Write-Host ""
    Write-Host "  # Schedule recurring newsletter"
    Write-Host "  $cmd --csv subscribers.csv --template newsletter.html --cron '0 9 * * 1' --env config.json"
    Write-Host ""
    Write-Host "  # Run scheduler daemon"
    Write-Host "  $cmd --scheduler-run --env config.json"
    Write-Host ""
    
    # Show additional features info
    if ($CreateShortcuts -or $WindowsTerminalProfile) {
        Write-Host "🎨 Windows Integration:"
        if ($CreateShortcuts) {
            Write-Host "  • Start Menu shortcut created"
            if ((Read-Host "Desktop shortcut created? (y/N)") -eq 'y') {
                Write-Host "  • Desktop shortcut created"
            }
        }
        if ($WindowsTerminalProfile) {
            Write-Host "  • Windows Terminal profile added"
        }
        Write-Host ""
    }
    
    Write-Host "🔄 Updates:"
    Write-Host "  • Check for updates: .\install-enhanced.ps1 -CheckUpdates"
    Write-Host "  • Reinstall latest: .\install-enhanced.ps1 -Version latest"
    Write-Host ""
}

# Main installation process
function Main {
    Write-Header
    
    # Check if running as administrator (warn for security)
    if (Test-Administrator) {
        Write-Warning "Running as Administrator. Consider using a regular user account."
    }
    
    # Check for updates if requested
    if (-not (Test-UpdatesAvailable)) {
        exit 0
    }
    
    # Get version to install
    Get-LatestVersion
    
    # Download and install
    Download-And-Install
    
    # Add to PATH if requested
    Add-ToPath
    
    # Create shortcuts if requested
    New-Shortcuts
    
    # Add Windows Terminal profile if requested
    Add-WindowsTerminalProfile
    
    # Test installation
    Test-Installation
    
    # Show usage instructions
    Show-Usage
}

# Error handling
trap {
    Write-Error "An error occurred: $($_.Exception.Message)"
    Write-Info "If the problem persists, please report it at: https://github.com/$Repo/issues"
    exit 1
}

# Run main installation
Main