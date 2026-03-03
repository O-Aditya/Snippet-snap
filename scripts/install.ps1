# PowerShell install script for Snippet-Snap
# Usage: irm https://raw.githubusercontent.com/O-Aditya/Snippet-snap/main/scripts/install.ps1 | iex

$ErrorActionPreference = "Stop"

$repo = "O-Aditya/Snippet-snap"
$binary = "snap.exe"
$installDir = "$env:USERPROFILE\.local\bin"

# Detect arch
$arch = if ([Environment]::Is64BitOperatingSystem) { "amd64" } else { "386" }

# Get latest release
$release = Invoke-RestMethod "https://api.github.com/repos/$repo/releases/latest"
$version = $release.tag_name
$fileName = "snippet-snap_windows_$arch.zip"
$url = "https://github.com/$repo/releases/download/$version/$fileName"

Write-Host ""
Write-Host "  ◈  SNIPPET-SNAP  " -ForegroundColor Cyan -NoNewline
Write-Host "installer" -ForegroundColor DarkGray
Write-Host ""
Write-Host "  Version   " -ForegroundColor DarkGray -NoNewline
Write-Host "$version"
Write-Host "  Platform  " -ForegroundColor DarkGray -NoNewline
Write-Host "windows/$arch"
Write-Host "  Target    " -ForegroundColor DarkGray -NoNewline
Write-Host "$installDir\$binary"
Write-Host ""

# Download and extract
Write-Host "  Downloading..." -ForegroundColor DarkGray
$tmp = New-TemporaryFile | ForEach-Object { Remove-Item $_; New-Item -ItemType Directory -Path $_ }
$zipPath = Join-Path $tmp.FullName $fileName
Invoke-WebRequest -Uri $url -OutFile $zipPath

Expand-Archive -Path $zipPath -DestinationPath $tmp.FullName -Force

# Install
New-Item -ItemType Directory -Path $installDir -Force | Out-Null
Copy-Item (Join-Path $tmp.FullName $binary) (Join-Path $installDir $binary) -Force
Remove-Item $tmp.FullName -Recurse -Force

# ── Auto-add to PATH ──
$pathAdded = $false
$userPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($userPath -notlike "*$installDir*") {
    # Add to persistent user PATH
    [Environment]::SetEnvironmentVariable("Path", "$installDir;$userPath", "User")
    $pathAdded = $true
    Write-Host "  ✓ " -ForegroundColor Green -NoNewline
    Write-Host "Added to user PATH"
}

# ALSO update current session so it works immediately (no restart needed)
if ($env:Path -notlike "*$installDir*") {
    $env:Path = "$installDir;$env:Path"
}

Write-Host ""
Write-Host "  ✓ " -ForegroundColor Green -NoNewline
Write-Host "Installed successfully!" -ForegroundColor White

if ($pathAdded) {
    Write-Host ""
    Write-Host "  PATH was updated for this session and all future terminals." -ForegroundColor DarkGray
    Write-Host "  No restart needed — snap is ready to use now." -ForegroundColor DarkGray
}

Write-Host ""
Write-Host "  Get started:" -ForegroundColor DarkGray
Write-Host "    snap --help" -ForegroundColor Cyan
Write-Host "    snap add --name my-snippet --lang bash" -ForegroundColor Cyan
Write-Host "    snap find" -ForegroundColor Cyan
Write-Host ""
