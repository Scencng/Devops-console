<template>
  <el-card class="content-card mysql-explorer-card" shadow="never">
    <template #header>
      <div class="mysql-explorer-card__header">
        <div>
          <strong>{{ title }}</strong>
          <p v-if="description">{{ description }}</p>
        </div>
        <el-button class="soft-button" :loading="loading" @click="refresh">
          刷新
        </el-button>
      </div>
    </template>

    <div v-if="showConnectionMeta" class="mysql-explorer-meta">
      <div class="mysql-explorer-meta__item">
        <span>连接状态</span>
        <strong>{{ hasConnection ? '已连接' : '未连接' }}</strong>
      </div>
    </div>

    <el-empty v-if="!hasConnection" description="请先建立 MySQL 连接" />

    <div v-else-if="loading && treeData.length === 0" class="mysql-explorer-tree__skeleton">
      <el-skeleton animated>
        <template #template>
          <div
            v-for="index in 10"
            :key="`explorer-skeleton-${index}`"
            class="mysql-explorer-tree__skeleton-row"
          >
            <el-skeleton-item variant="circle" style="width: 18px; height: 18px;" />
            <div class="mysql-explorer-tree__skeleton-copy">
              <el-skeleton-item variant="text" style="width: 140px; height: 16px;" />
              <el-skeleton-item variant="text" style="width: 96px; height: 12px;" />
            </div>
          </div>
        </template>
      </el-skeleton>
    </div>

    <el-tree
      v-else
      ref="treeRef"
      class="mysql-explorer-tree"
      node-key="key"
      :data="treeData"
      :props="treeProps"
      :highlight-current="true"
      :expand-on-click-node="false"
      :current-node-key="currentNodeKey"
      @node-click="handleNodeClick"
      @node-expand="handleNodeExpand"
    >
      <template #default="{ data }">
        <div class="mysql-explorer-tree__node">
          <span class="mysql-explorer-tree__icon">
            <el-icon v-if="data.type === 'database'"><Coin /></el-icon>
            <el-icon v-else-if="data.type === 'table'"><Grid /></el-icon>
            <el-icon v-else-if="data.type === 'view'"><View /></el-icon>
            <el-icon v-else-if="data.type === 'function'"><SetUp /></el-icon>
            <el-icon v-else-if="data.type === 'backup'"><Box /></el-icon>
            <el-icon v-else><FolderOpened /></el-icon>
          </span>
          <div class="mysql-explorer-tree__copy">
            <span class="mysql-explorer-tree__title">{{ nodeTitle(data) }}</span>
            <small class="mysql-explorer-tree__meta">{{ nodeMeta(data) }}</small>
          </div>
        </div>
      </template>
    </el-tree>
  </el-card>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'

import { useConnectionStore } from '@/mysql/stores/connection'
import { useWorkspaceStore } from '@/mysql/stores/workspace'
import {
  ensureDatabaseChildrenLoaded,
  loadExplorerTree,
  type TreeNodeData
} from '@/mysql/utils/explorer'

const props = withDefaults(defineProps<{
  title?: string
  description?: string
  showConnectionMeta?: boolean
  preloadPreferredChildren?: boolean
}>(), {
  title: '资源管理器',
  description: '',
  showConnectionMeta: false,
  preloadPreferredChildren: true
})

const emit = defineEmits<{
  (event: 'select', node: TreeNodeData): void
  (event: 'loaded', nodes: TreeNodeData[]): void
}>()

const connectionStore = useConnectionStore()
const workspaceStore = useWorkspaceStore()
const treeRef = ref()
const loading = ref(false)
const treeData = ref<TreeNodeData[]>([])
const currentNodeKey = ref('')

const treeProps = {
  label: 'label',
  children: 'children'
}

const hasConnection = computed(() => connectionStore.hasConnection)
function syncWorkspace(node: TreeNodeData) {
  if (node.type === 'database' || node.type === 'group') {
    workspaceStore.setActiveDatabase(node.databaseName)
    workspaceStore.clearActiveTable()
    return
  }

  if (node.tableName) {
    workspaceStore.setActiveTable(node.databaseName, node.tableName)
  }
}

function nodeTitle(node: TreeNodeData) {
  if (node.type === 'group') {
    if (node.groupKind === 'tables') return '数据表'
    if (node.groupKind === 'views') return '视图'
    if (node.groupKind === 'functions') return '函数'
    if (node.groupKind === 'backups') return '备份'
    return '查询'
  }

  return node.label
}

function nodeMeta(node: TreeNodeData) {
  if (node.type === 'database') return '数据库'
  if (node.type === 'table') return `${node.databaseName} / 表`
  if (node.type === 'view') return `${node.databaseName} / 视图`
  if (node.type === 'function') return `${node.databaseName} / 函数`
  if (node.type === 'backup') return `${node.databaseName} / 备份`
  return node.databaseName || ''
}

async function refresh() {
  if (!hasConnection.value) {
    treeData.value = []
    emit('loaded', [])
    return
  }

  loading.value = true
  try {
    treeData.value = await loadExplorerTree()
    emit('loaded', treeData.value)

    const preferredDatabase = workspaceStore.activeDatabase || connectionStore.profile.database
    const target = treeData.value.find((node) => node.databaseName === preferredDatabase) || treeData.value[0]
    if (target) {
      if (props.preloadPreferredChildren) {
        await ensureDatabaseChildrenLoaded(target)
      } else {
        void ensureDatabaseChildrenLoaded(target)
      }
    }
    emit('loaded', treeData.value)
  } finally {
    loading.value = false
  }
}

function handleNodeClick(node: TreeNodeData) {
  selectNode(node.key)
  syncWorkspace(node)
  emit('select', node)
}

async function handleNodeExpand(node: TreeNodeData) {
  if (node.type === 'database') {
    await ensureDatabaseChildrenLoaded(node)
    return
  }

  if (node.type === 'group' && node.groupKind === 'backups') {
    const databaseNode = treeData.value.find((item) => item.databaseName === node.databaseName)
    if (databaseNode) {
      await ensureDatabaseChildrenLoaded(databaseNode, { includeBackups: true })
    }
  }
}

function selectNode(nodeKey: string) {
  currentNodeKey.value = nodeKey
  treeRef.value?.setCurrentKey?.(nodeKey)
}

defineExpose({
  refresh,
  treeData,
  selectNode
})

onMounted(() => {
  void refresh()
})
</script>

<style scoped>
.mysql-explorer-card__header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
}

.mysql-explorer-card__header p {
  margin: 6px 0 0;
  color: var(--text-sub);
  font-size: 13px;
}

.mysql-explorer-meta {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
  margin-bottom: 16px;
}

.mysql-explorer-meta__item {
  padding: 14px 16px;
  border: 1px solid var(--devops-border-light);
  border-radius: var(--devops-radius-md);
  background: var(--devops-bg-panel-soft);
}

.mysql-explorer-meta__item span {
  display: block;
  color: var(--devops-text-secondary);
  font-size: 12px;
  font-weight: 600;
}

.mysql-explorer-meta__item strong {
  display: block;
  margin-top: 6px;
  color: var(--devops-text-primary);
  font-size: 14px;
}

.mysql-explorer-tree__node {
  display: flex;
  align-items: center;
  gap: 10px;
  min-width: 0;
  padding: 4px 0;
}

.mysql-explorer-tree__icon {
  display: inline-flex;
  color: var(--devops-primary);
}

.mysql-explorer-tree__copy {
  min-width: 0;
}

.mysql-explorer-tree__skeleton {
  min-height: 420px;
  padding: 8px 0;
}

.mysql-explorer-tree__skeleton-row {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 2px;
}

.mysql-explorer-tree__skeleton-copy {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.mysql-explorer-tree__title,
.mysql-explorer-tree__meta {
  display: block;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.mysql-explorer-tree__title {
  color: var(--devops-text-primary);
}

.mysql-explorer-tree__meta {
  color: var(--devops-text-secondary);
  font-size: 12px;
}

@media (max-width: 768px) {
  .mysql-explorer-meta {
    grid-template-columns: 1fr;
  }
}
</style>
