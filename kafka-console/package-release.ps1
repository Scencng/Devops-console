param(
    [string]$ProjectDir = (Join-Path $PSScriptRoot "."),
    [string]$ArchivePrefix = "kafka-console-prebuilt",
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

Require-Command tar
Require-Command npm
Require-Command wsl

$resolvedProjectDir = (Resolve-Path $ProjectDir).Path
$projectName = Split-Path $resolvedProjectDir -Leaf
$timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
$stageRoot = Join-Path ([System.IO.Path]::GetTempPath()) "kafka-console-package-$timestamp"
$stageProjectDir = Join-Path $stageRoot $projectName
$archiveName = "$ArchivePrefix-$timestamp.tar.gz"
$archivePath = Join-Path $resolvedProjectDir $archiveName
$manifestPath = Join-Path $resolvedProjectDir "PREBUILT_MANIFEST.txt"

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

Write-Step "构建前端"
Push-Location (Join-Path $resolvedProjectDir "frontend")
try {
    npm run build
}
finally {
    Pop-Location
}

Write-Step "在 WSL 中构建 Linux 后端二进制"
$wslProjectDir = (wsl.exe wslpath -a $resolvedProjectDir).Trim()
if (-not $wslProjectDir) {
    throw "无法将项目路径转换为 WSL 路径: $resolvedProjectDir"
}
$wslBuildCmd = "cd '$wslProjectDir/backend' && export GOPROXY=https://goproxy.cn,direct && export GOSUMDB=sum.golang.google.cn && GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o devops ./cmd/server"
wsl.exe sh -lc $wslBuildCmd

Write-Step "整理预构建发布目录"
New-Item -ItemType Directory -Force -Path $stageProjectDir | Out-Null

$topLevelFiles = @(
    ".env.example",
    "deploy.sh",
    "deploy-prebuilt.sh",
    "release.sh",
    "docker-compose.prebuilt.yml",
    "README.md"
)
foreach ($file in $topLevelFiles) {
    Copy-PathIfExists (Join-Path $resolvedProjectDir $file) $stageProjectDir
}

$backendStage = Join-Path $stageProjectDir "backend"
New-Item -ItemType Directory -Force -Path $backendStage | Out-Null
Copy-PathIfExists (Join-Path $resolvedProjectDir "backend\devops") $backendStage
Copy-PathIfExists (Join-Path $resolvedProjectDir "backend\Dockerfile.prebuilt") $backendStage
Copy-PathIfExists (Join-Path $resolvedProjectDir "backend\sql") $backendStage
Copy-PathIfExists (Join-Path $resolvedProjectDir "backend\config") $backendStage

$frontendStage = Join-Path $stageProjectDir "frontend"
New-Item -ItemType Directory -Force -Path $frontendStage | Out-Null
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
Copy-Item -LiteralPath (Join-Path $stageProjectDir "PREBUILT_MANIFEST.txt") -Destination $manifestPath -Force

Write-Step "打包为 $archiveName"
Push-Location $stageRoot
try {
    tar -czf $archiveName $projectName
    Move-Item -LiteralPath (Join-Path $stageRoot $archiveName) -Destination $archivePath -Force
}
finally {
    Pop-Location
}

if (-not $KeepStageDir -and (Test-Path $stageRoot)) {
    Remove-Item -LiteralPath $stageRoot -Recurse -Force
}

Write-Step "完成: $archivePath"
Write-Host "Linux 服务器推荐部署方式:" -ForegroundColor Yellow
Write-Host "1. 上传 $archiveName 到服务器 /opt"
Write-Host "2. 解压后进入项目目录"
Write-Host "3. 执行 chmod +x release.sh && ./release.sh install"
