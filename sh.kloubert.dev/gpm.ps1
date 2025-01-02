function Handle-Error {
    param (
        [string]$Message
    )
    Write-Host "Error: $Message" -ForegroundColor Red
    exit 1
}

Write-Host "Go Package Manager Updater"
Write-Host ""

$OS = "windows"

Write-Host "Your operating system: $OS"

$ARCH = $null
switch ([System.Runtime.InteropServices.RuntimeInformation]::ProcessArchitecture) {
    "X64" { $ARCH = "amd64" }
    "X86" { $ARCH = "386" }
    "Arm" { $ARCH = "arm" }
    "Arm64" { $ARCH = "arm64" }
    default {
        $ARCH = "unknown"
        Handle-Error "Unknown architecture detected!"
    }
}

$ARCH = "386"

Write-Host "Your architecture: $ARCH"
Write-Host ""

Write-Host "Finding download URL and SHA256 URL ..."

# Fetch the latest release info
try {
    $LatestReleaseInfo = Invoke-RestMethod -Uri "https://api.github.com/repos/mkloubert/go-package-manager/releases/latest"
} catch {
    Handle-Error "Could not fetch release infos"
}

$DownloadUrl = $LatestReleaseInfo.assets | Where-Object {
    $_.browser_download_url -match "gpm" -and
    $_.browser_download_url -match $OS -and
    $_.browser_download_url -match $ARCH -and
    $_.browser_download_url -notmatch "sha256"
} | Select-Object -ExpandProperty browser_download_url

$Sha256Url = $LatestReleaseInfo.assets | Where-Object {
    $_.browser_download_url -match "gpm" -and
    $_.browser_download_url -match $OS -and
    $_.browser_download_url -match $ARCH -and
    $_.browser_download_url -match "sha256"
} | Select-Object -ExpandProperty browser_download_url

if (-not $DownloadUrl) {
    Handle-Error "No valid download URL found"
}
if (-not $Sha256Url) {
    Handle-Error "No valid SHA256 URL found"
}

Write-Host "Downloading zip file from '$DownloadUrl'..."
try {
    Invoke-WebRequest -Uri $DownloadUrl -OutFile "gpm.zip"
} catch {
    Handle-Error "Failed to download zip file"
}

Write-Host "Downloading SHA256 file from '$Sha256Url'..."
try {
    Invoke-WebRequest -Uri $Sha256Url -OutFile "gpm.zip.sha256"
} catch {
    Handle-Error "Failed to download SHA256 file"
}

Write-Host "Verifying zip file ..."
try {
    $Sha256Value = Get-Content "gpm.zip.sha256"
    $CalculatedHash = (Get-FileHash -Path "gpm.zip" -Algorithm SHA256).Hash
    if ($Sha256Value -ne $CalculatedHash) {
        Handle-Error "SHA256 verification failed"
    }
} catch {
    Handle-Error "SHA256 verification failed"
}

Write-Host "Extracting binary ..."
try {
    Expand-Archive -Path "gpm.zip" -DestinationPath "gpm_extracted" -Force
    Move-Item -Path "gpm_extracted\gpm.exe" -Destination "gpm.exe"
} catch {
    Handle-Error "Could not extract 'gpm.exe' binary"
}

$DefaultDestination = $env:GPM_BIN_PATH
if ([string]::IsNullOrWhiteSpace($DefaultDestination)) {
    $DefaultDestination = "C:\\Program Files\\gpm\\gpm.exe"
}

$Destination = Read-Host "Enter the installation directory (Press Enter to use the default: $DefaultDestination)"
if ([string]::IsNullOrWhiteSpace($Destination)) {
    $Destination = $DefaultDestination
}

Write-Host "Installing 'gpm.exe' to $Destination ..."
try {
    $DestinationFolder = Split-Path -Path $Destination -Parent
    if (-not (Test-Path $DestinationFolder)) {
        New-Item -ItemType Directory -Path $DestinationFolder
    }
    Move-Item -Path gpm.exe -Destination $Destination -Force
} catch {
    Handle-Error "Could not move 'gpm.exe' to '$Destination'"
}

Write-Host "Cleaning up ..."
try {
    Remove-Item -Path "gpm.zip", "gpm.zip.sha256", "gpm_extracted" -Recurse -Force
} catch {
    Handle-Error "Cleanups failed"
}

Write-Host "'gpm.exe' successfully installed or updated 👍" -ForegroundColor Green
