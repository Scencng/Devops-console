#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR=""

for candidate in "$SCRIPT_DIR" "$SCRIPT_DIR/.." "$(pwd)"; do
  if [[ -f "${candidate}/backend/sql/kafka_module.sql" ]]; then
    PROJECT_DIR="$(cd "$candidate" && pwd)"
    break
  fi
done

if [[ -z "${PROJECT_DIR}" ]]; then
  PROJECT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
fi

NAME_PREFIX="demo"
CLUSTER_COUNT=12
AUDIT_COUNT_PER_CLUSTER=16
CONNECTION_TEST_COUNT_PER_CLUSTER=3
BROKERS_PER_CLUSTER=3
RANDOM_SEED=20260424
OUTPUT_SQL_PATH=""
RESET_EXISTING="false"
APPLY_TO_DOCKER="false"
MYSQL_CONTAINER="kafka-console-mysql"

print_usage() {
  cat <<'EOF'
用法:
  ./scripts/generate-kafka-test-data.sh [选项]

选项:
  --project-dir PATH                 项目根目录，默认脚本上一级目录
  --name-prefix VALUE                测试数据名前缀，默认 demo
  --cluster-count N                  生成 Kafka 集群数量，默认 12
  --audit-count-per-cluster N        每个集群生成的审计日志数量，默认 16
  --connection-test-count-per-cluster N
                                     每个集群生成的连接测试记录数量，默认 3
  --brokers-per-cluster N            每个集群的 bootstrap broker 数量，默认 3
  --random-seed N                    随机种子，默认 20260424
  --output-sql-path PATH             输出 SQL 文件路径，默认项目 output 目录
  --reset-existing                   生成前先清理同前缀旧数据
  --apply-to-docker                  生成后直接写入运行中的 MySQL 容器
  --mysql-container NAME             MySQL 容器名，默认 kafka-console-mysql
  -h, --help                         显示帮助

示例:
  ./scripts/generate-kafka-test-data.sh \
    --name-prefix qa \
    --cluster-count 20 \
    --audit-count-per-cluster 30 \
    --connection-test-count-per-cluster 3 \
    --reset-existing

  ./scripts/generate-kafka-test-data.sh \
    --name-prefix qa \
    --cluster-count 20 \
    --audit-count-per-cluster 30 \
    --connection-test-count-per-cluster 3 \
    --reset-existing \
    --apply-to-docker
EOF
}

require_command() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "[seed] 缺少命令: $1" >&2
    exit 1
  fi
}

step() {
  echo "[seed] $1"
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --project-dir)
      PROJECT_DIR="$2"
      shift 2
      ;;
    --name-prefix)
      NAME_PREFIX="$2"
      shift 2
      ;;
    --cluster-count)
      CLUSTER_COUNT="$2"
      shift 2
      ;;
    --audit-count-per-cluster)
      AUDIT_COUNT_PER_CLUSTER="$2"
      shift 2
      ;;
    --connection-test-count-per-cluster)
      CONNECTION_TEST_COUNT_PER_CLUSTER="$2"
      shift 2
      ;;
    --brokers-per-cluster)
      BROKERS_PER_CLUSTER="$2"
      shift 2
      ;;
    --random-seed)
      RANDOM_SEED="$2"
      shift 2
      ;;
    --output-sql-path)
      OUTPUT_SQL_PATH="$2"
      shift 2
      ;;
    --reset-existing)
      RESET_EXISTING="true"
      shift
      ;;
    --apply-to-docker)
      APPLY_TO_DOCKER="true"
      shift
      ;;
    --mysql-container)
      MYSQL_CONTAINER="$2"
      shift 2
      ;;
    -h|--help)
      print_usage
      exit 0
      ;;
    *)
      echo "[seed] 未知参数: $1" >&2
      print_usage
      exit 1
      ;;
  esac
done

require_command python3

PROJECT_DIR="$(cd "$PROJECT_DIR" && pwd)"

export PROJECT_DIR
export NAME_PREFIX
export CLUSTER_COUNT
export AUDIT_COUNT_PER_CLUSTER
export CONNECTION_TEST_COUNT_PER_CLUSTER
export BROKERS_PER_CLUSTER
export RANDOM_SEED
export OUTPUT_SQL_PATH
export RESET_EXISTING

step "生成测试数据 SQL"

GENERATED_OUTPUT_PATH="$(python3 <<'PY'
import json
import os
import random
from datetime import datetime, timedelta
from pathlib import Path

project_dir = Path(os.environ["PROJECT_DIR"]).resolve()
name_prefix = os.environ["NAME_PREFIX"].strip()
cluster_count = int(os.environ["CLUSTER_COUNT"])
audit_count_per_cluster = int(os.environ["AUDIT_COUNT_PER_CLUSTER"])
connection_test_count_per_cluster = int(os.environ["CONNECTION_TEST_COUNT_PER_CLUSTER"])
brokers_per_cluster = int(os.environ["BROKERS_PER_CLUSTER"])
random_seed = int(os.environ["RANDOM_SEED"])
output_sql_path = os.environ.get("OUTPUT_SQL_PATH", "").strip()
reset_existing = os.environ.get("RESET_EXISTING", "false").lower() == "true"

if not name_prefix:
    raise SystemExit("NamePrefix 不能为空")

for field_name, value, minimum in (
    ("cluster_count", cluster_count, 1),
    ("audit_count_per_cluster", audit_count_per_cluster, 0),
    ("connection_test_count_per_cluster", connection_test_count_per_cluster, 0),
    ("brokers_per_cluster", brokers_per_cluster, 1),
):
    if value < minimum:
        raise SystemExit(f"{field_name} 不能小于 {minimum}")

randomizer = random.Random(random_seed)
timestamp = datetime.now().strftime("%Y%m%d-%H%M%S")

if output_sql_path:
    output_path = Path(output_sql_path).expanduser()
else:
    output_path = project_dir / "output" / f"kafka-test-data-{name_prefix}-{timestamp}.sql"

output_path.parent.mkdir(parents=True, exist_ok=True)

environments = ["dev", "test", "staging", "preprod"]
tenants = ["team-alpha", "team-beta", "team-gamma", "team-delta"]
auth_types = ["none", "plain", "scram_sha256", "scram_sha512"]
statuses = ["active", "active", "active", "unknown", "error"]
operator_users = [
    {"id": 1, "username": "admin"},
    {"id": 1001, "username": "qa.chen"},
    {"id": 1002, "username": "tester.li"},
    {"id": 1003, "username": "ops.wang"},
]

def escape_sql(value):
    if value is None:
        return "NULL"
    return "'" + str(value).replace("'", "''") + "'"

def format_sql_datetime(value):
    if value is None:
        return "NULL"
    return escape_sql(value.strftime("%Y-%m-%d %H:%M:%S.%f")[:-3])

def random_choice(items):
    if not items:
        raise RuntimeError("随机候选集合不能为空")
    return randomizer.choice(items)

def random_bootstrap_servers(cluster_index, count):
    subnet = 20 + ((cluster_index - 1) % 60)
    return ",".join([f"10.{subnet}.0.{10 + offset}:9092" for offset in range(1, count + 1)])

def random_timestamp(within_days=30):
    now = datetime.now()
    minutes = randomizer.randint(0, within_days * 24 * 60)
    return now - timedelta(minutes=minutes)

def build_audit_template(cluster_name, cluster_index):
    topic_name = f"{cluster_name}-topic-{randomizer.randint(1, 6):02d}"
    group_id = f"{cluster_name}-group-{randomizer.randint(1, 4):02d}"
    broker_id = randomizer.randint(1, 3)
    templates = [
        {
            "action": "cluster:create",
            "resource_type": "cluster",
            "resource_name": cluster_name,
            "payload": {
                "name": cluster_name,
                "bootstrapServers": random_bootstrap_servers(cluster_index, brokers_per_cluster),
                "environment": None,
                "tenant": None,
            },
        },
        {
            "action": "cluster:update",
            "resource_type": "cluster",
            "resource_name": cluster_name,
            "payload": {
                "description": f"批量生成的测试集群 {cluster_name}",
                "environment": None,
                "tenant": None,
            },
        },
        {
            "action": "cluster:test",
            "resource_type": "cluster",
            "resource_name": cluster_name,
            "payload": {"cluster": cluster_name},
        },
        {
            "action": "topic:create",
            "resource_type": "topic",
            "resource_name": topic_name,
            "payload": {
                "name": topic_name,
                "numPartitions": randomizer.randint(3, 12),
                "replicationFactor": randomizer.randint(1, 3),
            },
        },
        {
            "action": "topic:config:update",
            "resource_type": "topic",
            "resource_name": topic_name,
            "payload": {
                "topic": topic_name,
                "entries": [
                    {
                        "key": "retention.ms",
                        "operation": "set",
                        "value": str(randomizer.randint(1, 13) * 86400000),
                    }
                ],
            },
        },
        {
            "action": "topic:delete",
            "resource_type": "topic",
            "resource_name": topic_name,
            "payload": {"topic": topic_name},
        },
        {
            "action": "group:offset:reset",
            "resource_type": "consumer_group",
            "resource_name": group_id,
            "payload": {
                "topic": topic_name,
                "partition": 0,
                "allPartitions": False,
                "force": False,
                "resetType": "offset",
                "offset": randomizer.randint(0, 5000),
                "timestampMs": 0,
            },
        },
        {
            "action": "group:delete",
            "resource_type": "consumer_group",
            "resource_name": group_id,
            "payload": {
                "cluster": cluster_name,
                "groupId": group_id,
            },
        },
        {
            "action": "broker:config:update",
            "resource_type": "broker",
            "resource_name": f"broker-{broker_id}",
            "payload": {
                "clusterId": cluster_name,
                "configCount": 1,
                "configKeys": ["log.retention.ms"],
            },
        },
        {
            "action": "message:produce",
            "resource_type": "topic",
            "resource_name": topic_name,
            "payload": {
                "topic": topic_name,
                "keyEncoding": "plain",
                "valueEncoding": "plain",
                "headerCount": randomizer.randint(0, 2),
                "hasKey": bool(randomizer.randint(0, 1)),
                "valueBytes": randomizer.randint(32, 512),
            },
        },
        {
            "action": "cluster:discovery:import",
            "resource_type": "cluster",
            "resource_name": cluster_name,
            "payload": {
                "name": cluster_name,
                "address": random_bootstrap_servers(cluster_index, brokers_per_cluster),
                "authType": "none",
                "tlsEnabled": False,
            },
        },
    ]
    return random_choice(templates)

lines = [
    "SET NAMES utf8mb4;",
    "SET FOREIGN_KEY_CHECKS = 0;",
    "START TRANSACTION;",
    "",
    "-- 兼容旧环境：补建 connection_tests 表",
    "CREATE TABLE IF NOT EXISTS connection_tests (",
    "  id int unsigned NOT NULL AUTO_INCREMENT,",
    "  resource_type varchar(191) NOT NULL,",
    "  resource_id bigint unsigned NOT NULL,",
    "  test_result longtext DEFAULT NULL,",
    "  response_time bigint DEFAULT NULL,",
    "  error_message longtext DEFAULT NULL,",
    "  tested_at datetime(3) DEFAULT NULL,",
    "  PRIMARY KEY (id),",
    "  KEY idx_connection_tests_resource (resource_type, resource_id),",
    "  KEY idx_connection_tests_resource_type (resource_type),",
    "  KEY idx_connection_tests_resource_id (resource_id),",
    "  KEY idx_connection_tests_tested_at (tested_at),",
    "  KEY idx_connection_tests_time (tested_at)",
    ") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;",
    "",
]

if reset_existing:
    prefix_like = escape_sql(f"{name_prefix}-%")
    lines.extend([
        "-- 清理之前生成的测试数据",
        "DELETE FROM kafka_audit_logs",
        "WHERE cluster_id IN (",
        f"    SELECT id FROM (SELECT id FROM kafka_clusters WHERE name LIKE {prefix_like}) AS seed_clusters",
        ");",
        "DELETE FROM connection_tests",
        "WHERE resource_type = 'kafka_cluster'",
        "  AND resource_id IN (",
        f"      SELECT id FROM (SELECT id FROM kafka_clusters WHERE name LIKE {prefix_like}) AS seed_clusters",
        "  );",
        f"DELETE FROM kafka_clusters WHERE name LIKE {prefix_like};",
        "",
    ])

for cluster_index in range(1, cluster_count + 1):
    cluster_name = f"{name_prefix}-cluster-{cluster_index:03d}"
    environment = random_choice(environments)
    tenant = random_choice(tenants)
    auth_type = random_choice(auth_types)
    status = random_choice(statuses)
    created_at = random_timestamp(90)
    updated_at = created_at + timedelta(minutes=randomizer.randint(10, 720))
    tested_at = updated_at + timedelta(minutes=randomizer.randint(5, 240))
    bootstrap_servers = random_bootstrap_servers(cluster_index, brokers_per_cluster)
    description = f"批量生成的测试集群 {cluster_name}，用于分页、筛选、审计日志和风险提示验证。"
    username = "" if auth_type == "none" else f"user_{cluster_index}"
    tls_enabled = 0 if auth_type == "none" else randomizer.randint(0, 1)
    skip_verify = randomizer.randint(0, 1) if tls_enabled == 1 else 0
    last_error_message = f"dial tcp {bootstrap_servers}: i/o timeout" if status == "error" else None

    lines.extend([
        f"-- {cluster_name}",
        "INSERT INTO kafka_clusters",
        "(name, bootstrap_servers, version, auth_type, username, password_ciphertext, tls_enabled, insecure_skip_verify, ca_cert, client_cert, client_key_ciphertext, description, environment, tenant, status, last_error_message, last_tested_at, created_at, updated_at)",
        "VALUES",
        f"({escape_sql(cluster_name)}, {escape_sql(bootstrap_servers)}, '3.6.0', {escape_sql(auth_type)}, {escape_sql(username)}, '', {tls_enabled}, {skip_verify}, '', '', '', {escape_sql(description)}, {escape_sql(environment)}, {escape_sql(tenant)}, {escape_sql(status)}, {escape_sql(last_error_message)}, {format_sql_datetime(tested_at)}, {format_sql_datetime(created_at)}, {format_sql_datetime(updated_at)})",
        "ON DUPLICATE KEY UPDATE",
        "    bootstrap_servers = VALUES(bootstrap_servers),",
        "    version = VALUES(version),",
        "    auth_type = VALUES(auth_type),",
        "    username = VALUES(username),",
        "    tls_enabled = VALUES(tls_enabled),",
        "    insecure_skip_verify = VALUES(insecure_skip_verify),",
        "    description = VALUES(description),",
        "    environment = VALUES(environment),",
        "    tenant = VALUES(tenant),",
        "    status = VALUES(status),",
        "    last_error_message = VALUES(last_error_message),",
        "    last_tested_at = VALUES(last_tested_at),",
        "    updated_at = VALUES(updated_at);",
    ])

    for test_index in range(1, connection_test_count_per_cluster + 1):
        test_passed = test_index > 1 or status != "error"
        test_result = "success" if test_passed else "failure"
        response_time = randomizer.randint(20, 800) if test_passed else None
        test_error = None if test_passed else f"mock timeout for {cluster_name}"
        tested_at_for_record = tested_at - timedelta(minutes=(connection_test_count_per_cluster - test_index) * 30)
        lines.extend([
            "INSERT INTO connection_tests",
            "(resource_type, resource_id, test_result, response_time, error_message, tested_at)",
            "SELECT",
            "    'kafka_cluster',",
            "    id,",
            f"    {escape_sql(test_result)},",
            f"    {response_time if response_time is not None else 'NULL'},",
            f"    {escape_sql(test_error)},",
            f"    {format_sql_datetime(tested_at_for_record)}",
            "FROM kafka_clusters",
            f"WHERE name = {escape_sql(cluster_name)};",
        ])

    for audit_index in range(1, audit_count_per_cluster + 1):
        template = build_audit_template(cluster_name, cluster_index)
        operator = random_choice(operator_users)
        result = "success" if randomizer.randint(0, 9) < 8 else "failed"
        error_message = None if result == "success" else f"mock {template['action']} failure for validation"
        created_at_for_audit = updated_at + timedelta(minutes=audit_index)
        payload_json = json.dumps(template["payload"], ensure_ascii=False, separators=(",", ":"))
        lines.extend([
            "INSERT INTO kafka_audit_logs",
            "(cluster_id, action, resource_type, resource_name, operator_user_id, operator_username, request_payload, result, error_message, created_at)",
            "SELECT",
            "    id,",
            f"    {escape_sql(template['action'])},",
            f"    {escape_sql(template['resource_type'])},",
            f"    {escape_sql(template['resource_name'])},",
            f"    {operator['id']},",
            f"    {escape_sql(operator['username'])},",
            f"    {escape_sql(payload_json)},",
            f"    {escape_sql(result)},",
            f"    {escape_sql(error_message)},",
            f"    {format_sql_datetime(created_at_for_audit)}",
            "FROM kafka_clusters",
            f"WHERE name = {escape_sql(cluster_name)};",
        ])

    lines.append("")

lines.extend([
    "COMMIT;",
    "SET FOREIGN_KEY_CHECKS = 1;",
])

output_path.write_text("\n".join(lines) + "\n", encoding="utf-8")
print(str(output_path))
PY
)"

step "SQL 已生成: ${GENERATED_OUTPUT_PATH}"
echo "参数摘要:"
echo "  NamePrefix                = ${NAME_PREFIX}"
echo "  ClusterCount              = ${CLUSTER_COUNT}"
echo "  AuditCountPerCluster      = ${AUDIT_COUNT_PER_CLUSTER}"
echo "  ConnectionTestPerCluster  = ${CONNECTION_TEST_COUNT_PER_CLUSTER}"
echo "  RandomSeed                = ${RANDOM_SEED}"
echo "  ResetExisting             = ${RESET_EXISTING}"

if [[ "${APPLY_TO_DOCKER}" == "true" ]]; then
  require_command docker
  step "尝试将 SQL 写入 Docker 容器: ${MYSQL_CONTAINER}"

  if ! docker ps --format '{{.Names}}' | grep -Fxq "${MYSQL_CONTAINER}"; then
    echo "[seed] 未找到运行中的 MySQL 容器: ${MYSQL_CONTAINER}" >&2
    exit 1
  fi

  docker exec -i "${MYSQL_CONTAINER}" sh -lc 'mysql -uroot -p"$MYSQL_ROOT_PASSWORD" "$MYSQL_DATABASE"' < "${GENERATED_OUTPUT_PATH}"
  step "测试数据已写入容器数据库"
else
  echo "如需直接写入运行中的 MySQL 容器，可执行："
  echo "./scripts/generate-kafka-test-data.sh --apply-to-docker --reset-existing"
fi
