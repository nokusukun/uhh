# UHH Install Script for Windows
# Usage: irm https://raw.githubusercontent.com/nokusukun/uhh/main/install.ps1 | iex

$ErrorActionPreference = "Stop"

$Repo = "nokusukun/uhh"
$InstallDir = "$env:LOCALAPPDATA\Programs\uhh"

# Detect architecture
$Arch = if ([Environment]::Is64BitOperatingSystem) { "amd64" } else { "386" }

Write-Host "Detected: windows-$Arch" -ForegroundColor Cyan

# Get latest release
Write-Host "Fetching latest release..." -ForegroundColor Cyan
$Release = Invoke-RestMethod -Uri "https://api.github.com/repos/$Repo/releases/latest"
$Version = $Release.tag_name

if (-not $Version) {
    Write-Host "Failed to fetch latest version." -ForegroundColor Red
    exit 1
}

Write-Host "Latest version: $Version" -ForegroundColor Green

# Find download URL
$Asset = $Release.assets | Where-Object { $_.name -like "*windows*$Arch*.zip" } | Select-Object -First 1

if (-not $Asset) {
    Write-Host "Could not find Windows release asset." -ForegroundColor Red
    exit 1
}

$DownloadUrl = $Asset.browser_download_url
Write-Host "Downloading $DownloadUrl..." -ForegroundColor Cyan

# Create temp directory
$TempDir = New-Item -ItemType Directory -Path "$env:TEMP\uhh-install-$(Get-Random)" -Force

# Download
$ZipPath = "$TempDir\uhh.zip"
Invoke-WebRequest -Uri $DownloadUrl -OutFile $ZipPath

# Extract
Write-Host "Extracting..." -ForegroundColor Cyan
Expand-Archive -Path $ZipPath -DestinationPath $TempDir -Force

# Create install directory
if (-not (Test-Path $InstallDir)) {
    New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
}

# Move binary
$ExePath = Get-ChildItem -Path $TempDir -Filter "uhh.exe" -Recurse | Select-Object -First 1
if ($ExePath) {
    Copy-Item -Path $ExePath.FullName -Destination "$InstallDir\uhh.exe" -Force
} else {
    Write-Host "Could not find uhh.exe in archive." -ForegroundColor Red
    exit 1
}

# Cleanup
Remove-Item -Path $TempDir -Recurse -Force

# Add to PATH if not already there
$UserPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($UserPath -notlike "*$InstallDir*") {
    Write-Host "Adding to PATH..." -ForegroundColor Cyan
    [Environment]::SetEnvironmentVariable("Path", "$UserPath;$InstallDir", "User")
    $env:Path = "$env:Path;$InstallDir"
}

Write-Host ""
Write-Host "UHH $Version installed successfully!" -ForegroundColor Green
Write-Host "Location: $InstallDir\uhh.exe" -ForegroundColor Gray
Write-Host ""
Write-Host "Run 'uhh init' to configure your LLM providers." -ForegroundColor Yellow
Write-Host "Run 'uhh --help' for usage information." -ForegroundColor Yellow
Write-Host ""
Write-Host "NOTE: Restart your terminal for PATH changes to take effect." -ForegroundColor Cyan
