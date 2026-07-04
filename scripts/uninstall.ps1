#Requires -Version 5

$ErrorActionPreference = "Stop"
$AppName = "emissor-cli"

$installDir = Join-Path $env:LOCALAPPDATA "Programs\Emissor"
$target = Join-Path $installDir "$AppName.exe"

if (Test-Path $target) {
    Remove-Item $target -Force
    Write-Host "Binario removido: $target"
} else {
    Write-Host "Binario nao encontrado (talvez ja tenha sido removido)."
}

$userPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($userPath -like "*$installDir*") {
    $newPath = ($userPath -split ';' | Where-Object { $_ -and $_ -ne $installDir }) -join ';'
    [Environment]::SetEnvironmentVariable("Path", $newPath, "User")
    Write-Host "Pasta removida do PATH do usuario."
}

if (Test-Path $installDir) {
    Remove-Item $installDir -Recurse -Force -ErrorAction SilentlyContinue
}

$configDir = Join-Path $env:APPDATA "Emissor"
if (Test-Path $configDir) {
    Remove-Item $configDir -Recurse -Force
    Write-Host "Configuracao e metadados removidos: $configDir"
}

Write-Host ""
Write-Host "Seus documentos em Documents\Emissor NAO foram apagados."
