# PowerShell install script for Snippet-Snap
# Usage: irm https://raw.githubusercontent.com/O-Aditya/snippet-snap/main/scripts/install.ps1 | iex

$ErrorActionPreference = "Stop"

$repo = "O-Aditya/Snippet-snap"
$binary = "snap.exe"
$installDir = "$env:USERPROFILE\bin"

# Detect arch
$arch = if ([Environment]::Is64BitOperatingSystem) { "amd64" } else { "386" }

# Get latest release
$release = Invoke-RestMethod "https://api.github.com/repos/$repo/releases/latest"
$version = $release.tag_name
$fileName = "snippet-snap_windows_$arch.zip"
$url = "https://github.com/$repo/releases/download/$version/$fileName"

Write-Host "Installing snippet-snap $version for windows/$arch..."
Write-Host "  -> $url"

# Download and extract
$tmp = New-TemporaryFile | ForEach-Object { Remove-Item $_; New-Item -ItemType Directory -Path $_ }
$zipPath = Join-Path $tmp.FullName $fileName
Invoke-WebRequest -Uri $url -OutFile $zipPath

Expand-Archive -Path $zipPath -DestinationPath $tmp.FullName -Force

# Install
New-Item -ItemType Directory -Path $installDir -Force | Out-Null
Copy-Item (Join-Path $tmp.FullName $binary) (Join-Path $installDir $binary) -Force
Remove-Item $tmp.FullName -Recurse -Force

# Add to PATH if needed
$userPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($userPath -notlike "*$installDir*") {
    [Environment]::SetEnvironmentVariable("Path", "$installDir;$userPath", "User")
    Write-Host ""
    Write-Host "  Added $installDir to your PATH."
    Write-Host "  Restart your terminal for changes to take effect."
}

Write-Host ""
Write-Host "Done! Installed snap to $installDir\$binary"
Write-Host "  Run 'snap --help' to get started."
