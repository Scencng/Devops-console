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
  { title: '集群管理', desc: '维护 Kafka 连接与认证配置', path: '/kafka/clusters' },
  { title: '自动发现', desc: '扫描网段并识别 Kafka 候选节点', path: '/kafka/discovery' },
  { title: 'Topic 管理', desc: '创建 Topic、扩分区并修改配置', path: '/kafka/topics' },
  { title: 'Broker 管理', desc: '查看 Broker 分布与分区承载情况', path: '/kafka/brokers' },
  { title: '消费组管理', desc: '查看消费组明细并重置 Offset', path: '/kafka/groups' },
  { title: '消息浏览', desc: '按 Topic/Partition 浏览并发送测试消息', path: '/kafka/messages' },
  { title: 'Prometheus 面板', desc: '查看 Kafka Exporter 监控指标', path: '/kafka/prometheus' },
  { title: '审计日志', desc: '查看危险操作审计记录', path: '/kafka/audits' }
]
</script>

<style scoped>
.dashboard-grid { margin-top: 20px; }
.nav-card { cursor: pointer; min-height: 140px; }
.nav-card h3 { margin: 0 0 12px; }
.nav-card p { margin: 0; color: #666; }
</style>
