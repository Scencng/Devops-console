<template>
  <div class="page-container">
    <div class="workbench-grid">
      <section class="workspace-panel primary-panel">
        <div class="page-eyebrow">Kafka Workspace</div>
        <h3>把集群、主题、消费组和消息流放进一套更顺手的工作台。</h3>
        <p>
          从集群连接、自动发现到 Topic 变更与消息抽样，这里保留了最常用的入口，优先帮助你完成排查、确认和干预。
        </p>

        <div class="nav-grid">
          <article
            v-for="card in primaryCards"
            :key="card.title"
            class="nav-tile"
            @click="router.push(card.path)"
          >
            <div class="nav-tile-badge">{{ card.badge }}</div>
            <h3>{{ card.title }}</h3>
            <p>{{ card.desc }}</p>
            <span class="nav-tile-link">{{ card.action }}</span>
          </article>
        </div>
      </section>

      <section class="workspace-panel side-panel">
        <div class="page-eyebrow">Today</div>
        <h3>建议先处理的工作</h3>
        <div class="compact-list">
          <div v-for="item in highlights" :key="item.title" class="compact-item">
            <div>
              <strong>{{ item.title }}</strong>
              <span>{{ item.desc }}</span>
            </div>
            <el-button text type="primary" @click="router.push(item.path)">进入</el-button>
          </div>
        </div>
      </section>
    </div>

    <div class="section-stack">
      <section class="workspace-panel">
        <div class="toolbar-row">
          <div>
            <div class="page-eyebrow">Operations</div>
            <h3>常用工作区</h3>
          </div>
        </div>
        <div class="split-panel">
          <div class="surface-muted">
            <div class="compact-list">
              <div v-for="card in dataCards" :key="card.title" class="compact-item">
                <div>
                  <strong>{{ card.title }}</strong>
                  <span>{{ card.desc }}</span>
                </div>
                <el-button text type="primary" @click="router.push(card.path)">打开</el-button>
              </div>
            </div>
          </div>
          <div class="surface-muted">
            <div class="compact-list">
              <div v-for="card in governanceCards" :key="card.title" class="compact-item">
                <div>
                  <strong>{{ card.title }}</strong>
                  <span>{{ card.desc }}</span>
                </div>
                <el-button text type="primary" @click="router.push(card.path)">打开</el-button>
              </div>
            </div>
          </div>
        </div>
      </section>
    </div>
  </div>
</template>

<script setup>
import { useRouter } from 'vue-router'

const router = useRouter()

const primaryCards = [
  {
    badge: '01',
    title: '集群管理',
    desc: '维护 Kafka 连接、环境信息和认证参数，并完成连通性校验。',
    action: '进入集群管理',
    path: '/kafka/clusters',
  },
  {
    badge: '02',
    title: '自动发现',
    desc: '按网段扫描候选节点，识别真正 Broker 与访问入口，再批量导入。',
    action: '开始扫描',
    path: '/kafka/discovery',
  },
  {
    badge: '03',
    title: 'Topic 管理',
    desc: '创建 Topic、扩分区、查看 ISR 与副本分配，并处理高风险变更。',
    action: '查看 Topic',
    path: '/kafka/topics',
  },
]

const highlights = [
  { title: '查看 Kafka 总览', desc: '先确认集群规模、主题数量和消费组情况。', path: '/kafka' },
  { title: '处理消费组问题', desc: '快速定位消费组明细和 Offset 干预入口。', path: '/kafka/groups' },
  { title: '抽样检查消息', desc: '直接按 Topic / Partition 查看消息内容和头信息。', path: '/kafka/messages' },
]

const dataCards = [
  { title: 'Kafka 总览', desc: '查看集群规模、主题数量和核心运行面貌。', path: '/kafka' },
  { title: 'Broker 管理', desc: '观察 Broker 节点、Controller 和分区承载情况。', path: '/kafka/brokers' },
  { title: '消费组管理', desc: '查看消费者状态、Lag 与 Offset 重置入口。', path: '/kafka/groups' },
]

const governanceCards = [
  { title: '消息浏览', desc: '按 Topic / Partition 采样查看消息并发送测试消息。', path: '/kafka/messages' },
  { title: '审计日志', desc: '回看危险操作、集群导入与 Topic 配置变更记录。', path: '/kafka/audits' },
  { title: '系统管理', desc: '维护用户、角色、菜单以及后台权限配置。', path: '/system/users' },
]
</script>
