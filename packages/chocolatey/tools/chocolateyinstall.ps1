$ErrorActionPreference = 'Stop';
$toolsDir = "$(Split-Path -parent $MyInvocation.MyCommand.Definition)"

# Package information
$packageName = $env:ChocolateyPackageName
$url64bit = 'https://github.com/bravo1goingdark/mailgrid/releases/latest/download/mailgrid-windows-amd64.exe.zip'
$checksum64 = 'REPLACE_WITH_ACTUAL_SHA256'
$checksumType64 = 'sha256'

# Package parameters
$packageArgs = @{
  packageName   = $packageName
  unzipLocation = $toolsDir
  url64bit      = $url64bit
  softwareName  = 'MailGrid*'
  checksum64    = $checksum64
  checksumType64= $checksumType64
  validExitCodes= @(0)
}

# Install package
Install-ChocolateyZipPackage @packageArgs

# Create a shortcut in the tools directory to the extracted binary
$extractedExe = Get-ChildItem -Path $toolsDir -Filter "mailgrid*.exe" | Select-Object -First 1
if ($extractedExe) {
    $linkPath = Join-Path $toolsDir "mailgrid.exe"
    if ($extractedExe.Name -ne "mailgrid.exe") {
        Copy-Item $extractedExe.FullName $linkPath -Force
        Remove-Item $extractedExe.FullName -Force
    }
    
    Write-Host "MailGrid has been installed successfully!" -ForegroundColor Green
    Write-Host ""
    Write-Host "📚 Quick Start:" -ForegroundColor Yellow
    Write-Host "  1. Create a config.json file with your SMTP settings"
    Write-Host "  2. Test: mailgrid --to user@example.com --subject 'Test' --text 'Hello!' --env config.json"
    Write-Host "  3. Get help: mailgrid --help"
    Write-Host ""
    Write-Host "📖 Documentation: https://github.com/bravo1goingdark/mailgrid" -ForegroundColor Cyan
    Write-Host ""
}
else {
    throw "MailGrid binary not found after extraction"
}