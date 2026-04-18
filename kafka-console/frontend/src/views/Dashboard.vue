<template>
  <div class="page-container">
    <el-card class="page-header-card">
      <div class="page-header">
        <div>
          <h2>Kafka Console</h2>
          <p>Kafka 集群、消息、消费组和监控的统一入口</p>
        </div>
      </div>
    </el-card>

    <el-row :gutter="20" class="dashboard-grid">
      <el-col :span="8" v-for="card in cards" :key="card.title">
        <el-card class="nav-card" shadow="hover" @click="router.push(card.path)">
          <h3>{{ card.title }}</h3>
          <p>{{ card.desc }}</p>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { useRouter } from 'vue-router'
const router = useRouter()
const cards = [
  { title: 'Kafka 总览', desc: '查看集群总览与核心指标', path: '/kafka' },
  { title: '健康总览', desc: '统一查看健康状态与风险清单', path: '/kafka/health' },
  { title: '集群管理', desc: '维护 Kafka 连接与认证配置', path: '/kafka/clusters' },
  { title: 'Topic 管理', desc: '查看并修改 Topic 配置', path: '/kafka/topics' },
  { title: '消费组管理', desc: '查看 Lag 并重置 Offset', path: '/kafka/groups' },
  { title: '消息浏览', desc: '按 Topic/Partition 浏览消息', path: '/kafka/messages' },
  { title: 'Prometheus 面板', desc: '查看 Kafka Exporter 监控指标', path: '/kafka/prometheus' },
  { title: '趋势分析', desc: '回看吞吐、Lag 和副本趋势', path: '/kafka/trends' },
  { title: '告警中心', desc: '管理告警规则和告警事件', path: '/kafka/alerts' },
  { title: '告警降噪', desc: '维护静默窗口和降噪策略', path: '/kafka/alert-noise' },
  { title: '一键巡检', desc: '执行 Kafka 标准化巡检', path: '/kafka/inspection' },
  { title: '链路追踪', desc: '维护 Trace 与消息链路关联', path: '/kafka/traces' },
  { title: '扩容建议', desc: '查看 Topic 与消费组扩容建议', path: '/kafka/scaling' },
  { title: '自愈中心', desc: '配置异常场景自愈策略', path: '/kafka/self-healing' },
  { title: '任务编排', desc: '编排自动化运维任务', path: '/kafka/tasks' },
  { title: '变更审批', desc: '对高风险操作做审批和执行', path: '/kafka/approvals' },
  { title: 'Topic 元数据', desc: '维护 Topic 数据资产台账', path: '/kafka/metadata' },
  { title: 'GitOps 中心', desc: '生成 Kafka 声明式清单并记录同步', path: '/kafka/gitops' },
  { title: '云托管适配', desc: '维护托管 Kafka 的接入适配信息', path: '/kafka/cloud-adapters' },
  { title: '自动发现', desc: '扫描网段并识别 Kafka 候选节点', path: '/kafka/discovery' },
  { title: '数据血缘', desc: '维护 Topic 上下游血缘关系', path: '/kafka/lineage' },
  { title: '生命周期治理', desc: '输出 Topic 生命周期治理建议', path: '/kafka/lifecycle' },
  { title: 'Schema Registry', desc: '浏览 Subject 和 Schema 版本', path: '/kafka/schema-registry' },
  { title: '网关/Mesh', desc: '管理网关与 Service Mesh 集成', path: '/kafka/mesh-gateway' },
  { title: '成本治理', desc: '查看 Kafka 成本估算记录', path: '/kafka/cost' },
  { title: '敏感识别', desc: '扫描消息中的敏感数据', path: '/kafka/sensitive-data' },
  { title: '审计日志', desc: '查看危险操作审计记录', path: '/kafka/audits' }
]
</script>

<style scoped>
.dashboard-grid { margin-top: 20px; }
.nav-card { cursor: pointer; min-height: 140px; }
.nav-card h3 { margin: 0 0 12px; }
.nav-card p { margin: 0; color: #666; }
</style>
