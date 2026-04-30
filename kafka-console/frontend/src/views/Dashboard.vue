<template>
  <div class="page-container dashboard-page">
    <section class="workspace-panel dashboard-hero">
      <div class="dashboard-hero-copy">
        <div class="page-eyebrow">Workspace</div>
        <h2>首页聚焦最常用入口</h2>
        <p>先看总览，再处理 Topic、消费组和消息问题，其余能力收进下方快捷区。</p>

        <div class="workflow-strip">
          <span class="workflow-label">常用顺序</span>
          <div class="workflow-steps">
            <span v-for="step in workflowSteps" :key="step" class="workflow-step">{{ step }}</span>
          </div>
        </div>
      </div>

      <div class="focus-grid">
        <button
          v-for="card in focusCards"
          :key="card.title"
          type="button"
          class="focus-card"
          @click="router.push(card.path)"
        >
          <span class="focus-card-kicker">{{ card.kicker }}</span>
          <strong>{{ card.title }}</strong>
          <span class="focus-card-desc">{{ card.desc }}</span>
        </button>
      </div>
    </section>

    <section class="workspace-panel dashboard-shortcuts">
      <div class="toolbar-row">
        <div>
          <div class="page-eyebrow">Shortcuts</div>
          <h3>更多入口</h3>
        </div>
      </div>

      <div class="shortcut-grid">
        <button
          v-for="item in secondaryCards"
          :key="item.title"
          type="button"
          class="shortcut-item"
          @click="router.push(item.path)"
        >
          <div class="shortcut-copy">
            <strong>{{ item.title }}</strong>
            <span>{{ item.desc }}</span>
          </div>
          <span class="shortcut-action">{{ item.action }}</span>
        </button>
      </div>
    </section>
  </div>
</template>

<script setup>
import { useRouter } from 'vue-router'

const router = useRouter()

const workflowSteps = ['总览', 'Topic', '消费组', '消息']

const focusCards = [
  {
    kicker: 'Overview',
    title: 'Kafka 总览',
    desc: '先确认集群、Topic 和消费组状态。',
    path: '/kafka',
  },
  {
    kicker: 'Topics',
    title: 'Topic 管理',
    desc: '查看分区、副本和配置变更。',
    path: '/kafka/topics',
  },
  {
    kicker: 'Groups',
    title: '消费组',
    desc: '定位 Lag、状态和 Offset 问题。',
    path: '/kafka/groups',
  },
  {
    kicker: 'Messages',
    title: '消息浏览',
    desc: '按 Topic / Partition 快速抽样检查。',
    path: '/kafka/messages',
  },
]

const secondaryCards = [
  { title: '集群管理', desc: '维护连接与认证配置。', action: '进入', path: '/kafka/clusters' },
  { title: '自动发现', desc: '扫描并导入候选 Broker。', action: '扫描', path: '/kafka/discovery' },
  { title: 'Broker 管理', desc: '查看节点与分区承载情况。', action: '查看', path: '/kafka/brokers' },
  { title: '审计日志', desc: '回看危险操作和配置变更。', action: '打开', path: '/kafka/audits' },
  { title: '系统管理', desc: '维护用户、角色和菜单。', action: '进入', path: '/system/users' },
]
</script>
