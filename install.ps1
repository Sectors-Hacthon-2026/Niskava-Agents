# =============================================================================
# NISKAVA AGENT — Windows PowerShell Installer (Windows 10 / 11)
# Automates Go verification, Python virtualenv bootstrap, and CLI compilation.
# =============================================================================

Write-Host "`n=== Niskava Agent Windows Installer ===" -ForegroundColor Cyan

# 1. Check Go compiler
$goCmd = Get-Command "go" -ErrorAction SilentlyContinue
if (-not $goCmd) {
    Write-Host "[X] Go compiler was not found on PATH." -ForegroundColor Red
    Write-Host "    Please install Go (>=1.22) from: https://go.dev/dl/"
    Write-Host "    Ensure 'Add to PATH' is checked during installation.`n"
    Exit 1
}
$goVer = go version
Write-Host "[✓] Go compiler detected: $goVer" -ForegroundColor Green

# 2. Check Python 3.11+
$pyExe = $null
$pyCandidates = @("python", "py", "python3")
foreach ($cand in $pyCandidates) {
    $cmd = Get-Command $cand -ErrorAction SilentlyContinue
    if ($cmd) {
        $verCheck = & $cand -c "import sys; print(f'{sys.version_info.major}.{sys.version_info.minor}')" 2>$null
        if ($verCheck) {
            $parts = $verCheck.Split('.')
            if ([int]$parts[0] -ge 3 -and [int]$parts[1] -ge 11) {
                $pyExe = $cand
                break
            }
        }
    }
}

if (-not $pyExe) {
    Write-Host "[X] Python 3.11+ is required but was not found." -ForegroundColor Red
    Write-Host "    Please install Python 3.11+ from: https://www.python.org/downloads/"
    Write-Host "    IMPORTANT: Check 'Add python.exe to PATH' in installer.`n"
    Exit 1
}
$pyVerStr = & $pyExe --version
Write-Host "[✓] Compatible Python runtime detected: $pyVerStr" -ForegroundColor Green

# 3. Create or update virtual environment
Write-Host "`nSetting up Python isolated environment (.venv)..."
if (-not (Test-Path ".venv")) {
    Write-Host "  • Creating virtual environment in .venv..."
    & $pyExe -m venv .venv
} else {
    Write-Host "  • Virtual environment .venv already exists."
}

$venvPip = ".\.venv\Scripts\pip.exe"
if (-not (Test-Path $venvPip)) {
    Write-Host "[X] Pip was not found at $venvPip" -ForegroundColor Red
    Exit 1
}

Write-Host "  • Upgrading pip in .venv..."
& $venvPip install --upgrade pip --quiet

Write-Host "  • Installing quantitative engine dependencies (NumPy, Pandas, NetworkX)..."
& $venvPip install -r backend\engine\requirements.txt --quiet
Write-Host "[✓] Quantitative dependencies installed successfully." -ForegroundColor Green

# 4. Compile Go Core Binary for Windows
Write-Host "`nCompiling Niskava Windows executable (niskava.exe)..."
if (-not (Test-Path "bin")) {
    New-Item -ItemType Directory -Path "bin" | Out-Null
}
go build -o bin\niskava.exe .\cmd\niskava
if ($LASTEXITCODE -ne 0) {
    Write-Host "[X] Failed to compile niskava.exe" -ForegroundColor Red
    Exit $LASTEXITCODE
}
Write-Host "[✓] Binary compiled successfully: bin\niskava.exe" -ForegroundColor Green

Write-Host "`n=== Installation Complete! ===" -ForegroundColor Green
Write-Host "Next recommended steps:"
Write-Host "  1. Run diagnostic check : .\bin\niskava.exe doctor" -ForegroundColor Cyan
Write-Host "  2. Configure credentials: .\bin\niskava.exe setup" -ForegroundColor Cyan
Write-Host "  3. Start Web Workspace  : .\bin\niskava.exe serve" -ForegroundColor Cyan
Write-Host "  4. Launch Terminal HUD  : .\bin\niskava.exe`n" -ForegroundColor Cyan
