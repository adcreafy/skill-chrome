$ErrorActionPreference = "Stop"

$REPO = "nicepkg/skill-chrome-host"
$BINARY_NAME = "skill-chrome-host"
$HOST_NAME = "com.nicepkg.skill_chrome"
$INSTALL_DIR = "$env:LOCALAPPDATA\skill-chrome"

function Get-LatestVersion {
    try {
        $release = Invoke-RestMethod -Uri "https://api.github.com/repos/$REPO/releases/latest"
        return $release.tag_name
    } catch {
        return "v0.1.0"
    }
}

function Download-Binary {
    param([string]$Version)

    $url = "https://github.com/$REPO/releases/download/$Version/$BINARY_NAME-windows-amd64.exe"

    Write-Host "Downloading $BINARY_NAME $Version for windows/amd64..."
    New-Item -ItemType Directory -Force -Path $INSTALL_DIR | Out-Null

    $target = Join-Path $INSTALL_DIR "$BINARY_NAME.exe"
    Invoke-WebRequest -Uri $url -OutFile $target
    Write-Host "Binary installed to: $target"
    return $target
}

function Register-NativeManifest {
    param([string]$BinaryPath)

    $manifest = @{
        name = $HOST_NAME
        description = "Skill Chrome Native Messaging Host"
        path = $BinaryPath
        type = "stdio"
        allowed_origins = @("chrome-extension://EXTENSION_ID_PLACEHOLDER/")
    } | ConvertTo-Json

    $regPath = "HKCU:\Software\Google\Chrome\NativeMessagingHosts\$HOST_NAME"
    $manifestPath = Join-Path $INSTALL_DIR "$HOST_NAME.json"

    $manifest | Out-File -Encoding utf8 -FilePath $manifestPath

    New-Item -Path $regPath -Force | Out-Null
    Set-ItemProperty -Path $regPath -Name "(Default)" -Value $manifestPath

    # Edge
    $edgeRegPath = "HKCU:\Software\Microsoft\Edge\NativeMessagingHosts\$HOST_NAME"
    New-Item -Path $edgeRegPath -Force | Out-Null
    Set-ItemProperty -Path $edgeRegPath -Name "(Default)" -Value $manifestPath

    Write-Host "Native messaging manifest registered"
}

Write-Host "=== Skill Chrome Host Installer ==="
Write-Host ""

$version = Get-LatestVersion
$binaryPath = Download-Binary -Version $version
Register-NativeManifest -BinaryPath $binaryPath

Write-Host ""
Write-Host "Installation complete!"
Write-Host "You can now use the Skill Chrome extension."
