$ErrorActionPreference = 'Stop'

$toolsDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$root = Split-Path -Parent $toolsDir
$appDir = Join-Path $root 'app'
$packageJsonPath = Join-Path $root 'package.json'

if (-not (Test-Path $packageJsonPath)) {
  throw "package.json not found at $packageJsonPath"
}

$packageJson = Get-Content -Path $packageJsonPath -Raw | ConvertFrom-Json
$appVersion = [string]$packageJson.version
if ([string]::IsNullOrWhiteSpace($appVersion)) {
  throw "Version is missing in $packageJsonPath"
}

$safeVersion = $appVersion -replace '[\\/:*?"<>|]', '-'
$output = Join-Path $root ("GuessWho-{0}.exe" -f $safeVersion)
$iconScript = Join-Path $toolsDir 'generate-windows-icon.ps1'

Push-Location $appDir
try {
  npm run build
  if ($LASTEXITCODE -ne 0) {
    throw "Frontend build failed."
  }
}
finally {
  Pop-Location
}

Push-Location $root
try {
  & $iconScript

  go run github.com/akavel/rsrc@latest -arch amd64 -ico .\app.ico -o .\rsrc_windows_amd64.syso
  if ($LASTEXITCODE -ne 0) {
    throw "Windows resource generation failed."
  }

  go build -o $output .
  if ($LASTEXITCODE -ne 0) {
    throw "Go build failed."
  }
}
finally {
  Pop-Location
}

Write-Host "Built $output"