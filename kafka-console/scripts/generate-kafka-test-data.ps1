param(
    [string]$ProjectDir = "",
    [string]$NamePrefix = "demo",
    [ValidateRange(1, 500)]
    [int]$ClusterCount = 12,
    [ValidateRange(0, 500)]
    [int]$AuditCountPerCluster = 16,
    [ValidateRange(0, 50)]
    [int]$ConnectionTestCountPerCluster = 3,
    [ValidateRange(1, 8)]
    [int]$BrokersPerCluster = 3,
    [ValidateRange(1, 2147483647)]
    [int]$RandomSeed = 20260424,
    [string]$OutputSqlPath,
    [switch]$ResetExisting,
    [switch]$ApplyToDocker,
    [string]$MysqlContainer = "kafka-console-mysql"
)

$ErrorActionPreference = "Stop"

function Write-Step($Message) {
    Write-Host "[seed] $Message" -ForegroundColor Cyan
}

function Require-Command($Name) {
    if (-not (Get-Command $Name -ErrorAction SilentlyContinue)) {
        throw "缺少命令: $Name"
    }
}

function Resolve-ProjectDir([string]$PreferredPath) {
    $candidates = @()
    if (-not [string]::IsNullOrWhiteSpace($PreferredPath)) {
        $candidates += $PreferredPath
    }
    $candidates += @(
        $PSScriptRoot,
        (Join-Path $PSScriptRoot ".."),
        (Get-Location).Path
    )

    foreach ($candidate in $candidates) {
        try {
            $resolved = (Resolve-Path $candidate -ErrorAction Stop).Path
            if (Test-Path (Join-Path $resolved "backend\sql\kafka_module.sql")) {
                return $resolved
            }
        } catch {
            continue
        }
    }

    throw "无法识别项目根目录，请通过 -ProjectDir 显式指定"
}

function Escape-SqlString([string]$Value) {
    if ($null -eq $Value) {
        return "NULL"
    }
    return "'" + ($Value -replace "'", "''") + "'"
}

function Format-SqlDateTime([object]$Value) {
    if ($null -eq $Value) {
        return "NULL"
    }
    $dateTimeValue = [datetime]$Value
    return Escape-SqlString ($dateTimeValue.ToString("yyyy-MM-dd HH:mm:ss.fff"))
}

function Add-SqlLine([System.Collections.Generic.List[string]]$Lines, [string]$Text = "") {
    $null = $Lines.Add($Text)
}

function New-RandomChoice([System.Random]$Random, [object[]]$Items) {
    if ($Items.Count -eq 0) {
        throw "随机候选集合不能为空"
    }
    return $Items[$Random.Next(0, $Items.Count)]
}

function New-RandomBootstrapServers([System.Random]$Random, [int]$ClusterIndex, [int]$Count) {
    $subnet = 20 + (($ClusterIndex - 1) % 60)
    $servers = for ($i = 1; $i -le $Count; $i++) {
        $hostNumber = 10 + $i
        "10.${subnet}.0.${hostNumber}:9092"
    }
    return ($servers -join ",")
}

function New-RandomTimestamp([System.Random]$Random, [int]$WithinDays = 30) {
    $now = Get-Date
    $minutes = $Random.Next(0, $WithinDays * 24 * 60)
    return $now.AddMinutes(-$minutes)
}

function New-AuditAction([System.Random]$Random, [string]$ClusterName, [int]$ClusterIndex) {
    $topicName = "$ClusterName-topic-{0:d2}" -f ($Random.Next(1, 7))
    $groupId = "$ClusterName-group-{0:d2}" -f ($Random.Next(1, 5))
    $brokerId = $Random.Next(1, 4)
    $templates = @(
        @{
            Action        = "cluster:create"
            ResourceType  = "cluster"
            ResourceName  = $ClusterName
            Payload       = @{
                name             = $ClusterName
                bootstrapServers = New-RandomBootstrapServers $Random $ClusterIndex $BrokersPerCluster
                environment      = $null
                tenant           = $null
            }
        },
        @{
            Action        = "cluster:update"
            ResourceType  = "cluster"
            ResourceName  = $ClusterName
            Payload       = @{
                description = "批量生成的测试集群 $ClusterName"
                environment = $null
                tenant      = $null
            }
        },
        @{
            Action        = "cluster:test"
            ResourceType  = "cluster"
            ResourceName  = $ClusterName
            Payload       = @{
                cluster = $ClusterName
            }
        },
        @{
            Action        = "topic:create"
            ResourceType  = "topic"
            ResourceName  = $topicName
            Payload       = @{
                name              = $topicName
                numPartitions     = $Random.Next(3, 13)
                replicationFactor = $Random.Next(1, 4)
            }
        },
        @{
            Action        = "topic:config:update"
            ResourceType  = "topic"
            ResourceName  = $topicName
            Payload       = @{
                topic = $topicName
                entries = @(
                    @{
                        key       = "retention.ms"
                        operation = "set"
                        value     = [string]($Random.Next(1, 14) * 86400000)
                    }
                )
            }
        },
        @{
            Action        = "topic:delete"
            ResourceType  = "topic"
            ResourceName  = $topicName
            Payload       = @{
                topic = $topicName
            }
        },
        @{
            Action        = "group:offset:reset"
            ResourceType  = "consumer_group"
            ResourceName  = $groupId
            Payload       = @{
                topic         = $topicName
                partition     = 0
                allPartitions = $false
                force         = $false
                resetType     = "offset"
                offset        = $Random.Next(0, 5000)
                timestampMs   = 0
            }
        },
        @{
            Action        = "group:delete"
            ResourceType  = "consumer_group"
            ResourceName  = $groupId
            Payload       = @{
                cluster  = $ClusterName
                groupId  = $groupId
            }
        },
        @{
            Action        = "broker:config:update"
            ResourceType  = "broker"
            ResourceName  = "broker-$brokerId"
            Payload       = @{
                clusterId   = $ClusterName
                configCount = 1
                configKeys  = @("log.retention.ms")
            }
        },
        @{
            Action        = "message:produce"
            ResourceType  = "topic"
            ResourceName  = $topicName
            Payload       = @{
                topic         = $topicName
                keyEncoding   = "plain"
                valueEncoding = "plain"
                headerCount   = $Random.Next(0, 3)
                hasKey        = [bool]($Random.Next(0, 2))
                valueBytes    = $Random.Next(32, 512)
            }
        },
        @{
            Action        = "cluster:discovery:import"
            ResourceType  = "cluster"
            ResourceName  = $ClusterName
            Payload       = @{
                name             = $ClusterName
                address          = New-RandomBootstrapServers $Random $ClusterIndex $BrokersPerCluster
                authType         = "none"
                tlsEnabled       = $false
            }
        }
    )

    return New-RandomChoice $Random $templates
}

if ([string]::IsNullOrWhiteSpace($NamePrefix)) {
    throw "NamePrefix 不能为空"
}

$resolvedProjectDir = Resolve-ProjectDir $ProjectDir
$timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
if (-not $OutputSqlPath) {
    $outputDir = Join-Path $resolvedProjectDir "output"
    New-Item -ItemType Directory -Force -Path $outputDir | Out-Null
    $OutputSqlPath = Join-Path $outputDir "kafka-test-data-$($NamePrefix)-$timestamp.sql"
}

$outputDirName = Split-Path $OutputSqlPath -Parent
if ($outputDirName) {
    New-Item -ItemType Directory -Force -Path $outputDirName | Out-Null
}

$random = [System.Random]::new($RandomSeed)
$lines = [System.Collections.Generic.List[string]]::new()

$environments = @("dev", "test", "staging", "preprod")
$tenants = @("team-alpha", "team-beta", "team-gamma", "team-delta")
$authTypes = @("none", "plain", "scram_sha256", "scram_sha512")
$statuses = @("active", "active", "active", "unknown", "error")
$operatorUsers = @(
    @{ Id = 1; Username = "admin" },
    @{ Id = 1001; Username = "qa.chen" },
    @{ Id = 1002; Username = "tester.li" },
    @{ Id = 1003; Username = "ops.wang" }
)

Write-Step "生成测试数据 SQL"
Add-SqlLine $lines "SET NAMES utf8mb4;"
Add-SqlLine $lines "SET FOREIGN_KEY_CHECKS = 0;"
Add-SqlLine $lines "START TRANSACTION;"
Add-SqlLine $lines
Add-SqlLine $lines "-- 兼容旧环境：补建 connection_tests 表"
Add-SqlLine $lines "CREATE TABLE IF NOT EXISTS connection_tests ("
Add-SqlLine $lines "  id int unsigned NOT NULL AUTO_INCREMENT,"
Add-SqlLine $lines "  resource_type varchar(191) NOT NULL,"
Add-SqlLine $lines "  resource_id bigint unsigned NOT NULL,"
Add-SqlLine $lines "  test_result longtext DEFAULT NULL,"
Add-SqlLine $lines "  response_time bigint DEFAULT NULL,"
Add-SqlLine $lines "  error_message longtext DEFAULT NULL,"
Add-SqlLine $lines "  tested_at datetime(3) DEFAULT NULL,"
Add-SqlLine $lines "  PRIMARY KEY (id),"
Add-SqlLine $lines "  KEY idx_connection_tests_resource (resource_type, resource_id),"
Add-SqlLine $lines "  KEY idx_connection_tests_resource_type (resource_type),"
Add-SqlLine $lines "  KEY idx_connection_tests_resource_id (resource_id),"
Add-SqlLine $lines "  KEY idx_connection_tests_tested_at (tested_at),"
Add-SqlLine $lines "  KEY idx_connection_tests_time (tested_at)"
Add-SqlLine $lines ") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;"
Add-SqlLine $lines

if ($ResetExisting) {
    Write-Step "追加清理旧测试数据语句，前缀: $NamePrefix"
    $prefixLike = Escape-SqlString "$NamePrefix-%"
    Add-SqlLine $lines "-- 清理之前生成的测试数据"
    Add-SqlLine $lines 'DELETE FROM kafka_audit_logs'
    Add-SqlLine $lines 'WHERE cluster_id IN ('
    Add-SqlLine $lines ('    SELECT id FROM (SELECT id FROM kafka_clusters WHERE name LIKE ' + $prefixLike + ') AS seed_clusters')
    Add-SqlLine $lines ");"
    Add-SqlLine $lines 'DELETE FROM connection_tests'
    Add-SqlLine $lines "WHERE resource_type = 'kafka_cluster'"
    Add-SqlLine $lines '  AND resource_id IN ('
    Add-SqlLine $lines ('      SELECT id FROM (SELECT id FROM kafka_clusters WHERE name LIKE ' + $prefixLike + ') AS seed_clusters')
    Add-SqlLine $lines "  );"
    Add-SqlLine $lines ('DELETE FROM kafka_clusters WHERE name LIKE ' + $prefixLike + ';')
    Add-SqlLine $lines
}

for ($clusterIndex = 1; $clusterIndex -le $ClusterCount; $clusterIndex++) {
    $clusterName = "{0}-cluster-{1:d3}" -f $NamePrefix, $clusterIndex
    $environment = New-RandomChoice $random $environments
    $tenant = New-RandomChoice $random $tenants
    $authType = New-RandomChoice $random $authTypes
    $status = New-RandomChoice $random $statuses
    $createdAt = New-RandomTimestamp $random 90
    $updatedAt = $createdAt.AddMinutes($random.Next(10, 720))
    $testedAt = $updatedAt.AddMinutes($random.Next(5, 240))
    $bootstrapServers = New-RandomBootstrapServers $random $clusterIndex $BrokersPerCluster
    $description = "批量生成的测试集群 $clusterName，用于分页、筛选、审计日志和风险提示验证。"
    $username = if ($authType -eq "none") { "" } else { "user_$clusterIndex" }
    $tlsEnabled = if ($authType -eq "none") { 0 } else { [int]($random.Next(0, 2)) }
    $skipVerify = if ($tlsEnabled -eq 1) { [int]($random.Next(0, 2)) } else { 0 }
    $lastErrorMessage = if ($status -eq "error") { "dial tcp ${bootstrapServers}: i/o timeout" } else { $null }

    Add-SqlLine $lines "-- $clusterName"
    Add-SqlLine $lines @"
INSERT INTO kafka_clusters
(name, bootstrap_servers, version, auth_type, username, password_ciphertext, tls_enabled, insecure_skip_verify, ca_cert, client_cert, client_key_ciphertext, description, environment, tenant, status, last_error_message, last_tested_at, created_at, updated_at)
VALUES
($(Escape-SqlString $clusterName), $(Escape-SqlString $bootstrapServers), '3.6.0', $(Escape-SqlString $authType), $(Escape-SqlString $username), '', $tlsEnabled, $skipVerify, '', '', '', $(Escape-SqlString $description), $(Escape-SqlString $environment), $(Escape-SqlString $tenant), $(Escape-SqlString $status), $(Escape-SqlString $lastErrorMessage), $(Format-SqlDateTime $testedAt), $(Format-SqlDateTime $createdAt), $(Format-SqlDateTime $updatedAt))
ON DUPLICATE KEY UPDATE
    bootstrap_servers = VALUES(bootstrap_servers),
    version = VALUES(version),
    auth_type = VALUES(auth_type),
    username = VALUES(username),
    tls_enabled = VALUES(tls_enabled),
    insecure_skip_verify = VALUES(insecure_skip_verify),
    description = VALUES(description),
    environment = VALUES(environment),
    tenant = VALUES(tenant),
    status = VALUES(status),
    last_error_message = VALUES(last_error_message),
    last_tested_at = VALUES(last_tested_at),
    updated_at = VALUES(updated_at);
"@

    for ($testIndex = 1; $testIndex -le $ConnectionTestCountPerCluster; $testIndex++) {
        $testPassed = $testIndex -gt 1 -or $status -ne "error"
        $testResult = if ($testPassed) { "success" } else { "failure" }
        $responseTime = if ($testPassed) { $random.Next(20, 800) } else { $null }
        $testError = if ($testPassed) { $null } else { "mock timeout for $clusterName" }
        $testedAtForRecord = $testedAt.AddMinutes(-($ConnectionTestCountPerCluster - $testIndex) * 30)

        Add-SqlLine $lines @"
INSERT INTO connection_tests
(resource_type, resource_id, test_result, response_time, error_message, tested_at)
SELECT
    'kafka_cluster',
    id,
    $(Escape-SqlString $testResult),
    $(if ($null -eq $responseTime) { "NULL" } else { [string]$responseTime }),
    $(Escape-SqlString $testError),
    $(Format-SqlDateTime $testedAtForRecord)
FROM kafka_clusters
WHERE name = $(Escape-SqlString $clusterName);
"@
    }

    for ($auditIndex = 1; $auditIndex -le $AuditCountPerCluster; $auditIndex++) {
        $template = New-AuditAction $random $clusterName $clusterIndex
        $operator = New-RandomChoice $random $operatorUsers
        $result = if ($random.Next(0, 10) -lt 8) { "success" } else { "failed" }
        $errorMessage = if ($result -eq "failed") { "mock $($template.Action) failure for validation" } else { $null }
        $createdAtForAudit = $updatedAt.AddMinutes($auditIndex)
        $payloadJson = ($template.Payload | ConvertTo-Json -Compress -Depth 8)

        Add-SqlLine $lines @"
INSERT INTO kafka_audit_logs
(cluster_id, action, resource_type, resource_name, operator_user_id, operator_username, request_payload, result, error_message, created_at)
SELECT
    id,
    $(Escape-SqlString $template.Action),
    $(Escape-SqlString $template.ResourceType),
    $(Escape-SqlString $template.ResourceName),
    $($operator.Id),
    $(Escape-SqlString $operator.Username),
    $(Escape-SqlString $payloadJson),
    $(Escape-SqlString $result),
    $(Escape-SqlString $errorMessage),
    $(Format-SqlDateTime $createdAtForAudit)
FROM kafka_clusters
WHERE name = $(Escape-SqlString $clusterName);
"@
    }

    Add-SqlLine $lines
}

Add-SqlLine $lines "COMMIT;"
Add-SqlLine $lines "SET FOREIGN_KEY_CHECKS = 1;"

$utf8NoBom = New-Object System.Text.UTF8Encoding($false)
[System.IO.File]::WriteAllLines($OutputSqlPath, $lines, $utf8NoBom)

Write-Step "SQL 已生成: $OutputSqlPath"
Write-Host "参数摘要:" -ForegroundColor Yellow
Write-Host "  NamePrefix                = $NamePrefix"
Write-Host "  ClusterCount              = $ClusterCount"
Write-Host "  AuditCountPerCluster      = $AuditCountPerCluster"
Write-Host "  ConnectionTestPerCluster  = $ConnectionTestCountPerCluster"
Write-Host "  RandomSeed                = $RandomSeed"
Write-Host "  ResetExisting             = $ResetExisting"

if ($ApplyToDocker) {
    Require-Command docker
    Write-Step "尝试将 SQL 写入 Docker 容器: $MysqlContainer"

    $containerName = (& docker ps --format '{{.Names}}' | Where-Object { $_ -eq $MysqlContainer } | Select-Object -First 1)
    if (-not $containerName) {
        throw "未找到运行中的 MySQL 容器: $MysqlContainer"
    }

    Get-Content -Raw -LiteralPath $OutputSqlPath |
        & docker exec -i $MysqlContainer sh -lc 'mysql -uroot -p"$MYSQL_ROOT_PASSWORD" "$MYSQL_DATABASE"'

    if ($LASTEXITCODE -ne 0) {
        throw "写入 Docker MySQL 失败，退出码: $LASTEXITCODE"
    }

    Write-Step "测试数据已写入容器数据库"
}
else {
    Write-Host "如需直接写入运行中的 MySQL 容器，可执行：" -ForegroundColor Yellow
    Write-Host "powershell -ExecutionPolicy Bypass -File .\scripts\generate-kafka-test-data.ps1 -ApplyToDocker -ResetExisting"
}
