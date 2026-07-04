#Requires -Version 5

$ErrorActionPreference = "Stop"

$Repo = if ($env:EMISSOR_REPO) { $env:EMISSOR_REPO } else { "https://github.com/gabriellsalesx/Emissor_CLI" }
$Version = if ($env:EMISSOR_VERSION) { $env:EMISSOR_VERSION } else { "latest" }
$AppName = "emissor-cli"

$asset = "${AppName}_windows_amd64.exe"

if ($Version -eq "latest") {
    $url = "$Repo/releases/latest/download/$asset"
} else {
    $url = "$Repo/releases/download/$Version/$asset"
}

$installDir = Join-Path $env:LOCALAPPDATA "Programs\Emissor"
$target = Join-Path $installDir "$AppName.exe"

Write-Host "Baixando $AppName (windows/amd64)..."
New-Item -ItemType Directory -Force -Path $installDir | Out-Null

try {
    Invoke-WebRequest -Uri $url -OutFile $target -UseBasicParsing
} catch {
    Write-Error "Nao consegui baixar de $url"
    exit 1
}

Write-Host "Instalado em: $target"

$userPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($userPath -notlike "*$installDir*") {
    [Environment]::SetEnvironmentVariable("Path", "$userPath;$installDir", "User")
    Write-Host "Pasta adicionada ao PATH do usuario. Abra um novo terminal para usar."
}

Write-Host ""
Write-Host "Pronto! Execute: $AppName"
