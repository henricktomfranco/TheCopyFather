# setup_engine.ps1 - Setup embedded llama.cpp CPU engine & Qwen model for TheCopyFather
# This script downloads llama.cpp CPU binary release and Qwen GGUF model into %APPDATA%\TheCopyfather

$ErrorActionPreference = "Stop"

$engineDir = "$env:APPDATA\TheCopyfather\engine"
$modelsDir = "$env:APPDATA\TheCopyfather\models"

Write-Host "==> Ensuring directories exist..." -ForegroundColor Cyan
New-Item -ItemType Directory -Path $engineDir -Force | Out-Null
New-Item -ItemType Directory -Path $modelsDir -Force | Out-Null

$serverExe = "$engineDir\llama-server.exe"
if (Test-Path $serverExe) {
    Write-Host "==> llama-server.exe is already installed at: $serverExe" -ForegroundColor Green
    & $serverExe --version
} else {
    Write-Host "==> Fetching latest llama.cpp release from GitHub..." -ForegroundColor Cyan
    try {
        $rel = Invoke-RestMethod -Uri "https://api.github.com/repos/ggml-org/llama.cpp/releases?per_page=1" -UserAgent "TheCopyFather"
        $asset = $rel.assets | Where-Object { $_.name -like "*bin-win-cpu-x64.zip" } | Select-Object -First 1
        if (-not $asset) {
            throw "Could not find *bin-win-cpu-x64.zip in the latest release."
        }
        $downloadUrl = $asset.browser_download_url
        $fileName = $asset.name
    } catch {
        Write-Host "Failed to query GitHub API, using fallback release..." -ForegroundColor Yellow
        $downloadUrl = "https://github.com/ggml-org/llama.cpp/releases/download/b11227/llama-b11227-bin-win-cpu-x64.zip"
        $fileName = "llama-b11227-bin-win-cpu-x64.zip"
    }

    $tempZip = "$env:TEMP\$fileName"
    Write-Host "==> Downloading $fileName..." -ForegroundColor Cyan
    Invoke-WebRequest -Uri $downloadUrl -OutFile $tempZip -UserAgent "TheCopyFather"

    Write-Host "==> Extracting to $engineDir..." -ForegroundColor Cyan
    tar.exe -xf $tempZip -C $engineDir
    Remove-Item $tempZip -Force

    if (Test-Path $serverExe) {
        Write-Host "==> Successfully installed embedded engine!" -ForegroundColor Green
        & $serverExe --version
    } else {
        Write-Host "Error: Extraction failed, llama-server.exe not found." -ForegroundColor Red
        exit 1
    }
}

# 2. Check and Download SmolLM2 GGUF Model
$modelFile = "$modelsDir\smollm2-360m-instruct-q4_k_m.gguf"
if (Test-Path $modelFile) {
    Write-Host "==> SmolLM2 GGUF model already installed at: $modelFile" -ForegroundColor Green
} else {
    Write-Host "==> Downloading SmolLM2-360M GGUF model (~258 MB)..." -ForegroundColor Cyan
    $modelUrl = "https://huggingface.co/bartowski/SmolLM2-360M-Instruct-GGUF/resolve/main/SmolLM2-360M-Instruct-Q4_K_M.gguf"
    $tempModel = "$modelFile.tmp"

    if (Get-Command curl.exe -ErrorAction SilentlyContinue) {
        curl.exe -L -o $tempModel $modelUrl
    } else {
        Invoke-WebRequest -Uri $modelUrl -OutFile $tempModel -UserAgent "TheCopyFather"
    }

    if (Test-Path $tempModel) {
        Move-Item -Path $tempModel -Destination $modelFile -Force
        Write-Host "==> Successfully installed model: $modelFile" -ForegroundColor Green
    } else {
        Write-Host "Error: Model download failed." -ForegroundColor Red
        exit 1
    }
}

Write-Host "==> Embedded engine and model are completely ready!" -ForegroundColor Green
