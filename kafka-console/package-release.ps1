param(
    [string]$ProjectDir = (Join-Path $PSScriptRoot "."),
    [string]$ArchivePrefix = "kafka-console-prebuilt",
    [ValidateSet("gz", "xz")]
    [string]$Compression = "xz",
    [switch]$DisableUpx,
    [switch]$KeepStageDir
)

$ErrorActionPreference = "Stop"

function Write-Step($Message) {
    Write-Host "[package] $Message" -ForegroundColor Cyan
}

function Require-Command($Name) {
    if (-not (Get-Command $Name -ErrorAction SilentlyContinue)) {
        throw "缺少命令: $Name"
    }
}

function Convert-ToWslPath($WindowsPath) {
    $normalized = $WindowsPath -replace '\\', '/'
    $result = (& wsl.exe wslpath -a -- "$normalized" 2>$null)
    if (-not $result) {
        throw "无法将路径转换为 WSL 路径: $WindowsPath"
    }
    return ($result | Select-Object -First 1).Trim()
}

Require-Command tar
Require-Command npm
Require-Command wsl

$resolvedProjectDir = (Resolve-Path $ProjectDir).Path
$projectName = Split-Path $resolvedProjectDir -Leaf
$timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
$stageRoot = Join-Path ([System.IO.Path]::GetTempPath()) "kafka-console-package-$timestamp"
$stageProjectDir = Join-Path $stageRoot $projectName
$archiveName = if ($Compression -eq "xz") {
    "$ArchivePrefix-$timestamp.tar.xz"
} else {
    "$ArchivePrefix-$timestamp.tar.gz"
}
$archivePath = Join-Path $resolvedProjectDir $archiveName
$frontendDir = Join-Path $resolvedProjectDir "frontend"
$frontendNodeModules = Join-Path $frontendDir "node_modules"
$cleanupNodeModules = $false

function Copy-PathIfExists($Source, $DestinationParent, [string]$TargetName = "") {
    if (-not (Test-Path -LiteralPath $Source)) {
        throw "缺少打包输入: $Source"
    }

    $name = if ($TargetName) { $TargetName } else { Split-Path $Source -Leaf }
    $destination = Join-Path $DestinationParent $name

    if (Test-Path -LiteralPath $Source -PathType Container) {
        New-Item -ItemType Directory -Force -Path $destination | Out-Null
        Get-ChildItem -LiteralPath $Source -Force | ForEach-Object {
            Copy-Item -LiteralPath $_.FullName -Destination $destination -Recurse -Force
        }
    }
    else {
        New-Item -ItemType Directory -Force -Path $DestinationParent | Out-Null
        Copy-Item -LiteralPath $Source -Destination $destination -Force
    }
}

if (-not (Test-Path -LiteralPath $frontendNodeModules)) {
    Write-Step "前端依赖缺失，执行 npm ci"
    Push-Location $frontendDir
    try {
        npm ci
        if ($LASTEXITCODE -ne 0) {
            throw "前端依赖安装失败，退出码: $LASTEXITCODE"
        }
        $cleanupNodeModules = $true
    }
    finally {
        Pop-Location
    }
}

Write-Step "构建前端"
Push-Location $frontendDir
try {
    npm run build
    if ($LASTEXITCODE -ne 0) {
        throw "前端构建失败，退出码: $LASTEXITCODE"
    }
}
finally {
    Pop-Location
}

Write-Step "整理预构建发布目录"
New-Item -ItemType Directory -Force -Path $stageProjectDir | Out-Null
$backendStage = Join-Path $stageProjectDir "backend"
$frontendStage = Join-Path $stageProjectDir "frontend"
New-Item -ItemType Directory -Force -Path $backendStage | Out-Null
New-Item -ItemType Directory -Force -Path $frontendStage | Out-Null

Write-Step "在 WSL 中构建 Linux 后端二进制到临时打包目录"
$wslProjectDir = Convert-ToWslPath $resolvedProjectDir
$wslStageProjectDir = Convert-ToWslPath $stageProjectDir
$useUpx = if ($DisableUpx) { "false" } else { "true" }
$wslBuildCommand = @"
set -e
mkdir -p '$wslStageProjectDir/backend'
cd '$wslProjectDir/backend'
export GOPROXY=https://goproxy.cn,direct
export GOSUMDB=sum.golang.google.cn
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags='-s -w' -o '$wslStageProjectDir/backend/devops' ./cmd/server
if [ "$useUpx" = "true" ]; then
  if command -v upx >/dev/null 2>&1; then
    upx --best --lzma '$wslStageProjectDir/backend/devops'
    upx -t '$wslStageProjectDir/backend/devops'
  elif command -v upx-ucl >/dev/null 2>&1; then
    upx-ucl --best --lzma '$wslStageProjectDir/backend/devops'
    upx-ucl -t '$wslStageProjectDir/backend/devops'
  else
    echo '[package][warn] WSL 中未找到 upx/upx-ucl，跳过二进制压缩' >&2
  fi
fi
"@
wsl.exe sh -lc $wslBuildCommand
if ($LASTEXITCODE -ne 0) {
    throw "WSL 后端构建失败，退出码: $LASTEXITCODE"
}

$topLevelFiles = @(
    ".env.example",
    "DEPLOY_LINUX.md",
    "docker-compose.prebuilt.yml"
)
foreach ($file in $topLevelFiles) {
    if ($file -eq "docker-compose.prebuilt.yml") {
        Copy-PathIfExists (Join-Path $resolvedProjectDir $file) $stageProjectDir "docker-compose.yml"
    }
    else {
        Copy-PathIfExists (Join-Path $resolvedProjectDir $file) $stageProjectDir
    }
}

Copy-PathIfExists (Join-Path $resolvedProjectDir "backend\Dockerfile.prebuilt") $backendStage
Copy-PathIfExists (Join-Path $resolvedProjectDir "backend\sql") $backendStage
Copy-PathIfExists (Join-Path $resolvedProjectDir "frontend\Dockerfile.prebuilt") $frontendStage
Copy-PathIfExists (Join-Path $resolvedProjectDir "frontend\nginx.conf") $frontendStage
Copy-PathIfExists (Join-Path $resolvedProjectDir "frontend\dist") $frontendStage

$gitCommit = ""
try {
    $gitCommit = (git -C $resolvedProjectDir rev-parse --short HEAD 2>$null).Trim()
} catch {
    $gitCommit = ""
}
$manifestLines = @(
    "BuiltAt=$timestamp",
    "Project=$projectName",
    "GitCommit=$gitCommit",
    "BackendBinary=backend/devops",
    "FrontendDist=frontend/dist"
)
$manifestLines | Set-Content -LiteralPath (Join-Path $stageProjectDir "PREBUILT_MANIFEST.txt")

Write-Step "打包为 $archiveName"
Push-Location $stageRoot
try {
    if ($Compression -eq "xz") {
        tar -cJf $archiveName $projectName
    }
    else {
        tar -czf $archiveName $projectName
    }
    Move-Item -LiteralPath (Join-Path $stageRoot $archiveName) -Destination $archivePath -Force
}
finally {
    Pop-Location
}

if (-not $KeepStageDir -and (Test-Path $stageRoot)) {
    Remove-Item -LiteralPath $stageRoot -Recurse -Force
}

if ($cleanupNodeModules -and (Test-Path -LiteralPath $frontendNodeModules)) {
    Write-Step "清理临时安装的前端依赖"
    Remove-Item -LiteralPath $frontendNodeModules -Recurse -Force
}

Write-Step "完成: $archivePath"
Write-Host "Linux 服务器推荐部署方式:" -ForegroundColor Yellow
Write-Host "1. 上传 $archiveName 到服务器 /opt"
Write-Host "2. 解压后进入项目目录"
Write-Host "3. cp .env.example .env && 修改密码"
Write-Host "4. 执行 docker compose up -d --build"
