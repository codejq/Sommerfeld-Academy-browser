<#
.SYNOPSIS
  Builds dist\SommerfeldAcademy.exe (Go, CGO off, WebView2 window).

.EXAMPLE
  ./build.ps1
  ./build.ps1 -Version 1.0.0
#>
param([string]$Version = '1.0.0')
$ErrorActionPreference = 'Stop'
$root = Split-Path -Parent $MyInvocation.MyCommand.Path
Set-Location $root

function Step($m) { Write-Host "`n==> $m" -ForegroundColor Cyan }
function Exec([scriptblock]$cmd) { & $cmd; if ($LASTEXITCODE -ne 0) { throw "Command failed ($LASTEXITCODE): $cmd" } }

# Go may be installed but not on PATH in this shell.
if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
    $goBin = Join-Path $env:ProgramFiles 'Go\bin'
    if (Test-Path (Join-Path $goBin 'go.exe')) { $env:Path = "$goBin;$env:Path" } else { throw 'Go not found. Install it: winget install GoLang.Go' }
}
$Version = $Version.TrimStart('v')
if ($Version -notmatch '^\d+\.\d+\.\d+$') { throw "Version must be x.y.z (got '$Version')" }
$fileVersion = "$Version.0"

Step 'Embedding Windows resources'
if (-not (Test-Path build/winres/icon.png) -or -not (Test-Path build/winres/icon.ico)) { Exec { go run ./tools/genicon build/winres/icon.png build/winres/logo.png } }
$winres = Join-Path (go env GOPATH) 'bin/go-winres.exe'
if (-not (Test-Path $winres)) { Exec { go install github.com/tc-hib/go-winres@latest } }
Exec { & $winres make --in build/winres/winres.json --out rsrc --arch amd64 --product-version $fileVersion --file-version $fileVersion }

Step 'Building SommerfeldAcademy.exe'
$env:CGO_ENABLED = '0'
$env:GOOS = 'windows'
$env:GOARCH = 'amd64'
Exec { go vet ./... }
New-Item -ItemType Directory -Force dist | Out-Null
Exec { go build -trimpath -ldflags "-s -w -H windowsgui -X main.version=$Version" -o dist/SommerfeldAcademy.exe . }
Write-Host "`nBuilt dist\SommerfeldAcademy.exe ($Version)" -ForegroundColor Green
