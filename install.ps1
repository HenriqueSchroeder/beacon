$ErrorActionPreference = "Stop"

$Repo = "HenriqueSchroeder/beacon"
$Binary = "beacon.exe"
$InstallDir = "$env:LOCALAPPDATA\beacon"

function Get-Arch {
    switch ($env:PROCESSOR_ARCHITECTURE) {
        "AMD64" { return "amd64" }
        "ARM64" { return "arm64" }
        default {
            Write-Error "Unsupported architecture: $env:PROCESSOR_ARCHITECTURE"
            exit 1
        }
    }
}

function Get-LatestVersion {
    $response = Invoke-WebRequest -Uri "https://github.com/$Repo/releases/latest" -MaximumRedirection 0 -ErrorAction SilentlyContinue -UseBasicParsing
    if ($response.Headers.Location) {
        return ($response.Headers.Location -split "/")[-1]
    }

    $response = Invoke-RestMethod -Uri "https://api.github.com/repos/$Repo/releases/latest" -UseBasicParsing
    return $response.tag_name
}

$arch = Get-Arch
$version = Get-LatestVersion

if (-not $version) {
    Write-Error "Could not determine latest version."
    exit 1
}

$archive = "beacon_windows_${arch}.zip"
$url = "https://github.com/$Repo/releases/download/$version/$archive"
$tmpDir = New-Item -ItemType Directory -Path ([System.IO.Path]::Combine([System.IO.Path]::GetTempPath(), [System.Guid]::NewGuid().ToString()))

try {
    Write-Host "Downloading beacon $version for windows/$arch..."
    $zipPath = Join-Path $tmpDir $archive
    Invoke-WebRequest -Uri $url -OutFile $zipPath -UseBasicParsing

    Write-Host "Extracting..."
    Expand-Archive -Path $zipPath -DestinationPath $tmpDir -Force

    if (-not (Test-Path $InstallDir)) {
        New-Item -ItemType Directory -Path $InstallDir | Out-Null
    }

    Copy-Item -Path (Join-Path $tmpDir $Binary) -Destination (Join-Path $InstallDir $Binary) -Force

    $userPath = [Environment]::GetEnvironmentVariable("Path", "User")
    if ($userPath -notlike "*$InstallDir*") {
        [Environment]::SetEnvironmentVariable("Path", "$userPath;$InstallDir", "User")
        Write-Host "Added $InstallDir to user PATH (restart terminal to apply)."
    }

    Write-Host "beacon $version installed successfully."
    Write-Host "Run 'beacon version' to verify."
}
finally {
    Remove-Item -Recurse -Force $tmpDir -ErrorAction SilentlyContinue
}
