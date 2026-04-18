param(
    [Parameter(Mandatory = $true)]
    [string]$ServerHost,

    [string]$ServerUser = "root",
    [int]$ServerPort = 22,
    [string]$RemoteDir = "/opt",
    [string]$ProjectDir = (Join-Path $PSScriptRoot "."),
    [string]$ArchivePrefix = "kafka-console-release",
    [switch]$ExtractOnServer,
    [switch]$RunDeploy,
    [switch]$KeepLocalArchive
)

$ErrorActionPreference = "Stop"

function Write-Step($Message) {
    Write-Host "[upload] $Message" -ForegroundColor Cyan
}

function Require-Command($Name) {
    if (-not (Get-Command $Name -ErrorAction SilentlyContinue)) {
        throw "缺少命令: $Name"
    }
}

Require-Command tar
Require-Command scp
Require-Command ssh

$resolvedProjectDir = (Resolve-Path $ProjectDir).Path
$projectName = Split-Path $resolvedProjectDir -Leaf
$timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
$stageRoot = Join-Path ([System.IO.Path]::GetTempPath()) "kafka-console-upload-$timestamp"
$stageProjectDir = Join-Path $stageRoot $projectName
$archiveName = "$ArchivePrefix-$timestamp.tar.gz"
$archivePath = Join-Path $stageRoot $archiveName
$remoteTarget = "$ServerUser@$ServerHost"

New-Item -ItemType Directory -Force -Path $stageProjectDir | Out-Null

Write-Step "整理发布目录"
$robocopyArgs = @(
    $resolvedProjectDir,
    $stageProjectDir,
    "/E",
    "/R:2",
    "/W:1",
    "/NFL",
    "/NDL",
    "/NJH",
    "/NJS",
    "/XD", ".git", "data", ".deploy", "node_modules", "dist", "coverage"
)
robocopy @robocopyArgs | Out-Null
if ($LASTEXITCODE -ge 8) {
    throw "robocopy 复制文件失败，退出码: $LASTEXITCODE"
}

if (Test-Path (Join-Path $stageProjectDir ".env")) {
    Remove-Item -LiteralPath (Join-Path $stageProjectDir ".env") -Force
}

Write-Step "打包为 $archiveName"
Push-Location $stageRoot
try {
    tar -czf $archiveName $projectName
}
finally {
    Pop-Location
}

Write-Step "上传到 $remoteTarget:$RemoteDir/"
ssh -p $ServerPort $remoteTarget "mkdir -p '$RemoteDir'"
scp -P $ServerPort $archivePath "${remoteTarget}:$RemoteDir/"

if ($ExtractOnServer -or $RunDeploy) {
    Write-Step "服务器端解压归档"
    ssh -p $ServerPort $remoteTarget "cd '$RemoteDir' && tar -xzf '$archiveName'"
}

if ($RunDeploy) {
    Write-Step "服务器端执行部署脚本"
    ssh -p $ServerPort $remoteTarget "cd '$RemoteDir/$projectName' && chmod +x deploy.sh && ./deploy.sh up"
}

if ($KeepLocalArchive) {
    $savedArchive = Join-Path $resolvedProjectDir $archiveName
    Move-Item -LiteralPath $archivePath -Destination $savedArchive -Force
    Write-Step "本地归档已保留: $savedArchive"
}

if (Test-Path $stageRoot) {
    Remove-Item -LiteralPath $stageRoot -Recurse -Force
}

Write-Step "完成"
Write-Host "常用示例:" -ForegroundColor Yellow
Write-Host ".\upload-release.ps1 -ServerHost 1.2.3.4 -ExtractOnServer"
Write-Host ".\upload-release.ps1 -ServerHost 1.2.3.4 -ExtractOnServer -RunDeploy"
