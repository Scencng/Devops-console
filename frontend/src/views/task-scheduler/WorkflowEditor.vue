<template>
  <div class="editor-root" @drop="onDrop" @dragover.prevent v-loading="loadingData">
    <!-- ── 顶部工具栏 ── -->
    <div class="top-bar">
      <div class="top-bar-left">
        <el-button link @click="$router.back()">
          <el-icon><ArrowLeft /></el-icon>
        </el-button>
        <span class="workflow-title">{{ workflowForm.name || (isEditMode ? '编辑工作流' : '新建工作流') }}</span>
      </div>

      <div class="top-bar-meta">
        <el-input v-model="workflowForm.name" size="small" placeholder="工作流名称 *" style="width:200px" />
        <el-input v-model="workflowForm.cronExpression" size="small" placeholder="Cron 表达式" style="width:180px">
          <template #prefix><el-icon><Timer /></el-icon></template>
        </el-input>
        <el-input v-model="workflowForm.description" size="small" placeholder="描述信息" style="width:250px" />
      </div>

      <div class="top-bar-right">
        <el-button size="small" :icon="VideoPlay" type="success" @click="handleRun" v-if="isEditMode">测试运行</el-button>
        <el-button size="small" type="primary" :icon="Check" :loading="saving" @click="handleSave">保存</el-button>
      </div>
    </div>

    <!-- ── 主编辑区 ── -->
    <div class="editor-main">
      <!-- 左侧组件面板 -->
      <div class="sidebar">
        <div class="sidebar-header">节点组件</div>
        <div class="node-list">
          <div
            v-for="item in nodeTypes"
            :key="item.type"
            class="node-item"
            draggable="true"
            @dragstart="onDragStart($event, item.type)"
          >
            <el-icon class="node-icon"><component :is="item.icon" /></el-icon>
            <span>{{ item.label }}</span>
          </div>
        </div>
        <div class="sidebar-tip">拖拽节点到画布中</div>
      </div>

      <!-- 画布 -->
      <div class="canvas-wrap">
        <VueFlow
          v-model:nodes="nodes"
          v-model:edges="edges"
          :default-zoom="1"
          :min-zoom="0.2"
          :max-zoom="4"
          :default-edge-options="{ type: 'smoothstep', markerEnd: 'arrowclosed' }"
          class="flow-canvas"
          @connect="onConnect"
          @connect-start="onConnectStart"
          @connect-end="onConnectEnd"
          @node-click="onNodeClick"
          @pane-click="onPaneClick"
        >
          <Background pattern-color="#e0e0e0" :gap="20" />
          <Controls />

          <!-- 悬浮菜单 (连接到空处时显示) -->
          <div v-if="isConnectingToPane" class="connect-menu" :style="{ left: connectMenuPos.x + 'px', top: connectMenuPos.y + 'px' }">
            <div class="connect-menu-header">添加节点</div>
            <div v-for="item in nodeTypes" :key="item.type" class="connect-menu-item" @click="addNodeAtConnectPos(item.type)">
              <el-icon><component :is="item.icon" /></el-icon>
              <span>{{ item.label }}</span>
            </div>
          </div>

          <!-- 自定义节点模板 -->
          <template #node-http="{ data, selected }">
            <div class="custom-node" :class="{ 'node-selected': selected }">
              <!-- 四个方向的连接点 -->
              <Handle type="target" :position="Position.Top" id="top" class="handle-top" />
              <Handle type="target" :position="Position.Bottom" id="bottom" class="handle-bottom" />
              <Handle type="target" :position="Position.Left" id="left" class="handle-left" />
              <Handle type="target" :position="Position.Right" id="right" class="handle-right" />
              <div class="node-header http-header">
                <el-icon><Connection /></el-icon>
                <span>{{ data.name || 'HTTP 请求' }}</span>
              </div>
              <div class="node-body">{{ data.method }} {{ data.url || '未配置' }}</div>
              <Handle type="source" :position="Position.Top" id="top" class="handle-top" />
              <Handle type="source" :position="Position.Bottom" id="bottom" class="handle-bottom" />
              <Handle type="source" :position="Position.Left" id="left" class="handle-left" />
              <Handle type="source" :position="Position.Right" id="right" class="handle-right" />
            </div>
          </template>

          <template #node-script="{ data, selected }">
            <div class="custom-node" :class="{ 'node-selected': selected }">
              <Handle type="target" :position="Position.Top" id="top" class="handle-top" />
              <Handle type="target" :position="Position.Bottom" id="bottom" class="handle-bottom" />
              <Handle type="target" :position="Position.Left" id="left" class="handle-left" />
              <Handle type="target" :position="Position.Right" id="right" class="handle-right" />
              <div class="node-header script-header">
                <el-icon><Document /></el-icon>
                <span>{{ data.name || '脚本执行' }}</span>
              </div>
              <div class="node-body">脚本: {{ data.script_type || 'shell' }}</div>
              <Handle type="source" :position="Position.Top" id="top" class="handle-top" />
              <Handle type="source" :position="Position.Bottom" id="bottom" class="handle-bottom" />
              <Handle type="source" :position="Position.Left" id="left" class="handle-left" />
              <Handle type="source" :position="Position.Right" id="right" class="handle-right" />
            </div>
          </template>

          <template #node-sql="{ data, selected }">
            <div class="custom-node" :class="{ 'node-selected': selected }">
              <Handle type="target" :position="Position.Top" id="top" class="handle-top" />
              <Handle type="target" :position="Position.Bottom" id="bottom" class="handle-bottom" />
              <Handle type="target" :position="Position.Left" id="left" class="handle-left" />
              <Handle type="target" :position="Position.Right" id="right" class="handle-right" />
              <div class="node-header sql-header">
                <el-icon><DataLine /></el-icon>
                <span>{{ data.name || 'SQL 任务' }}</span>
              </div>
              <div class="node-body">{{ data.datasource || '默认库' }}</div>
              <Handle type="source" :position="Position.Top" id="top" class="handle-top" />
              <Handle type="source" :position="Position.Bottom" id="bottom" class="handle-bottom" />
              <Handle type="source" :position="Position.Left" id="left" class="handle-left" />
              <Handle type="source" :position="Position.Right" id="right" class="handle-right" />
            </div>
          </template>

          <template #node-k8s="{ data, selected }">
            <div class="custom-node" :class="{ 'node-selected': selected }">
              <Handle type="target" :position="Position.Top" id="top" class="handle-top" />
              <Handle type="target" :position="Position.Bottom" id="bottom" class="handle-bottom" />
              <Handle type="target" :position="Position.Left" id="left" class="handle-left" />
              <Handle type="target" :position="Position.Right" id="right" class="handle-right" />
              <div class="node-header k8s-header">
                <el-icon><Box /></el-icon>
                <span>{{ data.name || 'K8s 操作' }}</span>
              </div>
              <div class="node-body">{{ data.operation }} {{ data.resource_type }}</div>
              <Handle type="source" :position="Position.Top" id="top" class="handle-top" />
              <Handle type="source" :position="Position.Bottom" id="bottom" class="handle-bottom" />
              <Handle type="source" :position="Position.Left" id="left" class="handle-left" />
              <Handle type="source" :position="Position.Right" id="right" class="handle-right" />
            </div>
          </template>
        </VueFlow>
      </div>

      <!-- 右侧配置面板 -->
      <transition name="slide-panel">
        <div class="config-panel" v-if="selectedNode">
          <div class="cp-header">
            <span>{{ selectedNodeLabel }} 配置</span>
            <el-button link size="small" @click="selectedNodeId = null">
              <el-icon><Close /></el-icon>
            </el-button>
          </div>
          <div class="cp-body">
            <el-form label-position="top" size="small">
              <el-form-item label="节点名称">
                <el-input v-model="selectedNode.data.name" />
              </el-form-item>

              <!-- 目标选择 -->
              <el-form-item label="执行目标" v-if="['script', 'sql', 'k8s'].includes(selectedNode.type)">
                <el-select v-model="selectedNode.data.targetId" placeholder="请选择执行目标" style="width: 100%" filterable>
                  <el-option
                    v-for="item in targetOptions"
                    :key="item.id"
                    :label="item.name"
                    :value="item.id"
                  />
                </el-select>
              </el-form-item>

              <!-- HTTP 属性 -->
              <template v-if="selectedNode.type === 'http'">
                <el-form-item label="URL">
                  <el-input v-model="selectedNode.data.url" placeholder="http://api.example.com" />
                </el-form-item>
                <el-form-item label="方法">
                  <el-select v-model="selectedNode.data.method" style="width: 100%">
                    <el-option label="GET" value="GET" />
                    <el-option label="POST" value="POST" />
                    <el-option label="PUT" value="PUT" />
                    <el-option label="DELETE" value="DELETE" />
                  </el-select>
                </el-form-item>
                <el-form-item label="Headers">
                  <el-input v-model="selectedNode.data.headers" type="textarea" placeholder='{"Content-Type": "application/json"}' />
                </el-form-item>
                <el-form-item label="Body">
                  <el-input v-model="selectedNode.data.body" type="textarea" :rows="4" />
                </el-form-item>
              </template>

              <!-- Script 属性 -->
              <template v-if="selectedNode.type === 'script'">
                <el-form-item label="脚本语言">
                  <el-select v-model="selectedNode.data.script_type" style="width: 100%">
                    <el-option label="Shell" value="shell" />
                    <el-option label="Python" value="python" />
                  </el-select>
                </el-form-item>
                <el-form-item label="脚本内容">
                  <el-input v-model="selectedNode.data.content" type="textarea" :rows="10" class="code-font" />
                </el-form-item>
              </template>

              <!-- SQL 属性 -->
              <template v-if="selectedNode.type === 'sql'">
                <el-form-item label="数据源">
                  <el-input v-model="selectedNode.data.datasource" />
                </el-form-item>
                <el-form-item label="SQL 语句">
                  <el-input v-model="selectedNode.data.sql" type="textarea" :rows="8" class="code-font" />
                </el-form-item>
              </template>

              <!-- K8s 属性 -->
              <template v-if="selectedNode.type === 'k8s'">
                <el-form-item label="操作">
                  <el-select v-model="selectedNode.data.operation" style="width: 100%">
                    <el-option label="Apply" value="apply" />
                    <el-option label="Delete" value="delete" />
                    <el-option label="Restart" value="restart" />
                  </el-select>
                </el-form-item>
                <el-form-item label="资源类型">
                  <el-select v-model="selectedNode.data.resource_type" style="width: 100%">
                    <el-option label="Deployment" value="deployment" />
                    <el-option label="Service" value="service" />
                    <el-option label="Pod" value="pod" />
                  </el-select>
                </el-form-item>
                <el-form-item label="资源名称/YAML">
                  <el-input v-model="selectedNode.data.manifest" type="textarea" :rows="8" class="code-font" />
                </el-form-item>
              </template>

              <div class="node-actions mt-20">
                <el-button type="danger" plain style="width: 100%" :icon="Delete" @click="handleDeleteNode">删除节点</el-button>
              </div>
            </el-form>
          </div>
        </div>
      </transition>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { VueFlow, useVueFlow, Position, Handle } from '@vue-flow/core'
import { Background } from '@vue-flow/background'
import { Controls } from '@vue-flow/controls'
import {
  ArrowLeft, Check, VideoPlay, Timer, Connection,
  Document, DataLine, Box, Close, Delete
} from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { getWorkflow, createWorkflow, updateWorkflow, executeWorkflow } from '@/api/workflow.js'
import { getHostList } from '@/api/asset.js'
import { getInstanceList } from '@/api/instance.js'
import api from '@/api/index.js'

import '@vue-flow/core/dist/style.css'
import '@vue-flow/core/dist/theme-default.css'
import '@vue-flow/controls/dist/style.css'

const route = useRoute()
const router = useRouter()
const { findNode, addNodes, addEdges, project, onConnect, removeNodes, removeEdges } = useVueFlow()

const isEditMode = computed(() => !!route.params.id)
const workflowId = route.params.id

const loadingData = ref(false)
const saving = ref(false)
const nodes = ref([])
const edges = ref([])

// ─── Connection Menu (Lark style) ───
const isConnectingToPane = ref(false)
const connectMenuPos = ref({ x: 0, y: 0 })
const lastConnectParams = ref(null)

const onConnectStart = (params) => {
  // 记录连接起始点信息
  lastConnectParams.value = params
}

const onConnectEnd = (event) => {
  // 如果是在空白处松开，显示菜单
  if (event.target.classList.contains('vue-flow__pane')) {
    isConnectingToPane.value = true
    connectMenuPos.value = { x: event.clientX, y: event.clientY }
  }
}

const addNodeAtConnectPos = (type) => {
  const { left, top } = document.querySelector('.flow-canvas').getBoundingClientRect()
  const position = project({
    x: connectMenuPos.value.x - left - 100,
    y: connectMenuPos.value.y - top - 20
  })

  const newNodeId = `node_${Date.now()}`
  const newNode = {
    id: newNodeId,
    type,
    position,
    data: { name: `${type}_node`, method: 'GET', url: '', script_type: 'shell', content: '', operation: 'apply', resource_type: 'deployment' }
  }

  addNodes([newNode])

  // 如果是从某个节点拉出来的，自动连线
  if (lastConnectParams.value && lastConnectParams.value.nodeId) {
    const { nodeId, handleId, handleType } = lastConnectParams.value
    const edge = {
      id: `e_${nodeId}_${newNodeId}`,
      source: handleType === 'source' ? nodeId : newNodeId,
      target: handleType === 'source' ? newNodeId : nodeId,
      sourceHandle: handleType === 'source' ? handleId : 'bottom',
      targetHandle: handleType === 'source' ? 'top' : handleId
    }
    addEdges([edge])
  }

  isConnectingToPane.value = false
}

const workflowForm = reactive({
  name: '',
  cronExpression: '',
  description: '',
  status: 1
})

const nodeTypes = [
  { type: 'http', label: 'HTTP 请求', icon: 'Connection' },
  { type: 'script', label: '脚本执行', icon: 'Document' },
  { type: 'sql', label: 'SQL 任务', icon: 'DataLine' },
  { type: 'k8s', label: 'K8s 操作', icon: 'Box' }
]

const selectedNodeId = ref(null)
const selectedNode = computed(() => nodes.value.find(n => n.id === selectedNodeId.value))
const selectedNodeLabel = computed(() => {
  if (!selectedNode.value) return ''
  return nodeTypes.find(t => t.type === selectedNode.value.type)?.label || '节点'
})

const targetOptions = ref([])

const fetchTargets = async (type) => {
  targetOptions.value = []
  try {
    if (type === 'script') {
      const res = await getHostList({ page: 1, pageSize: 100 })
      const hostData = res.data.data || res.data
      targetOptions.value = (hostData.list || hostData).map(h => ({ id: h.id, name: `${h.name} (${h.ip})` }))
    } else if (type === 'sql') {
      const res = await getInstanceList({ page: 1, page_size: 100 })
      const dbData = res.data.data || res.data
      targetOptions.value = (dbData.list || dbData).map(i => ({ id: i.id, name: i.name }))
    } else if (type === 'k8s') {
      const res = await api.get('/cluster/list', { params: { page: 1, pageSize: 100 } })
      targetOptions.value = (res.data.data || res.data || []).map(c => ({ id: c.id, name: c.name }))
    }
  } catch (e) {
    console.error('Fetch targets failed:', e)
  }
}

onMounted(() => {
  if (isEditMode.value) {
    fetchData()
  }
})

async function fetchData() {
  loadingData.value = true
  try {
    const res = await getWorkflow(workflowId)
    const data = res.data.data || res.data
    Object.assign(workflowForm, {
      name: data.name,
      cronExpression: data.cronExpression,
      description: data.description,
      status: data.status
    })

    // 转换节点和连线
    if (data.nodes) {
      nodes.value = data.nodes.map(n => ({
        id: n.id.toString(),
        type: n.type,
        position: { x: n.position_x, y: n.position_y },
        data: n.config || {}
      }))
    }
    if (data.edges) {
      edges.value = data.edges.map(e => ({
        id: e.id.toString(),
        source: e.fromNodeId.toString(),
        target: e.toNodeId.toString(),
        sourceHandle: e.sourceHandle,
        targetHandle: e.targetHandle,
        label: e.label
      }))
    }
  } catch (e) {
    ElMessage.error('获取数据失败')
  } finally {
    loadingData.value = false
  }
}

// ─── Drag & Drop ───
const onDragStart = (event, type) => {
  if (event.dataTransfer) {
    event.dataTransfer.setData('application/vueflow', type)
    event.dataTransfer.effectAllowed = 'move'
  }
}

const onDrop = (event) => {
  const type = event.dataTransfer?.getData('application/vueflow')
  if (!type) return

  const { left, top } = event.currentTarget.getBoundingClientRect()
  const position = project({
    x: event.clientX - left - 100,
    y: event.clientY - top - 20
  })

  const newNode = {
    id: `node_${Date.now()}`,
    type,
    position,
    label: `${type} node`,
    data: { name: `${type}_node`, method: 'GET', url: '', script_type: 'shell', content: '', operation: 'apply', resource_type: 'deployment' }
  }

  addNodes([newNode])
}

// ─── Events ───
onConnect((params) => {
  addEdges([params])
})

function onNodeClick({ node }) {
  selectedNodeId.value = node.id
  if (['script', 'sql', 'k8s'].includes(node.type)) {
    fetchTargets(node.type)
  }
}

function onPaneClick() {
  selectedNodeId.value = null
  isConnectingToPane.value = false
}

function handleDeleteNode() {
  if (selectedNodeId.value) {
    removeNodes([selectedNodeId.value])
    selectedNodeId.value = null
  }
}

// ─── Actions ───
async function handleSave() {
  if (!workflowForm.name) {
    return ElMessage.warning('请输入工作流名称')
  }

  saving.value = true
  try {
    const payload = {
      name: workflowForm.name,
      description: workflowForm.description,
      schedule: workflowForm.cronExpression,
      status: workflowForm.status,
      nodes: nodes.value.map((n, idx) => ({
        tempId: n.id,
        name: n.data.name || `${n.type}_${idx}`,
        type: n.type,
        targetId: n.data.targetId ? Number(n.data.targetId) : 0,
        targetType: n.type === 'script' ? 'host' : (n.type === 'sql' ? 'db' : (n.type === 'k8s' ? 'k8s' : '')),
        position_x: Math.round(n.position.x),
        position_y: Math.round(n.position.y),
        config: JSON.stringify(n.data || {}),
        workflowId: isEditMode.value ? Number(workflowId) : 0
      })),
      edges: edges.value.map(e => ({
        sourceTempId: e.source,
        targetTempId: e.target,
        sourceHandle: e.sourceHandle,
        targetHandle: e.targetHandle,
        label: e.label || '',
        workflowId: isEditMode.value ? Number(workflowId) : 0
      }))
    }

    if (isEditMode.value) {
      await updateWorkflow(workflowId, payload)
    } else {
      const res = await createWorkflow(payload)
      const newId = res.data?.data?.id || res.data?.id
      if (newId) {
        router.replace(`/task-scheduler/workflows/${newId}/edit`)
      }
    }
    ElMessage.success('保存成功')
  } catch (e) {
    ElMessage.error('保存失败')
  } finally {
    saving.value = false
  }
}

async function handleRun() {
  try {
    const res = await executeWorkflow(workflowId)
    const execId = res.data?.data?.id || res.data?.id
    if (execId) {
      ElMessage.success('测试运行已启动')
      router.push(`/task-scheduler/executions/${execId}/logs`)
    } else {
      ElMessage.success('测试运行已启动')
    }
  } catch (e) {
    ElMessage.error('运行失败')
  }
}
</script>

<style scoped>
.editor-root {
  display: flex;
  flex-direction: column;
  height: 100vh;
  background: #f5f7fa;
  overflow: hidden;
}

/* ─── 顶部工具栏 ─── */
.top-bar {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 8px 16px;
  background: #fff;
  border-bottom: 1px solid #e4e7ed;
  min-height: 52px;
  flex-shrink: 0;
}
.top-bar-left { display: flex; align-items: center; gap: 8px; }
.workflow-title { font-size: 15px; font-weight: 600; color: #303133; }
.top-bar-meta { display: flex; align-items: center; gap: 8px; flex: 1; }
.top-bar-right { display: flex; align-items: center; gap: 8px; }

/* ─── 主区域 ─── */
.editor-main {
  display: flex;
  flex: 1;
  overflow: hidden;
  position: relative;
}

/* ─── 左侧侧边栏 ─── */
.sidebar {
  width: 180px;
  background: #fff;
  border-right: 1px solid #e4e7ed;
  display: flex;
  flex-direction: column;
  padding: 12px;
}
.sidebar-header {
  font-size: 14px;
  font-weight: 600;
  color: #303133;
  margin-bottom: 16px;
  padding-bottom: 8px;
  border-bottom: 1px solid #f0f0f0;
}
.node-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px;
  background: #f8f9fb;
  border: 1px solid #e4e7ed;
  border-radius: 4px;
  margin-bottom: 10px;
  cursor: move;
  font-size: 13px;
  transition: all 0.2s;
}
.node-item:hover {
  border-color: #409eff;
  color: #409eff;
  background: #ecf5ff;
}
.node-icon { font-size: 16px; }
.sidebar-tip {
  margin-top: auto;
  font-size: 12px;
  color: #909399;
  text-align: center;
}

/* ─── 画布 ─── */
.canvas-wrap {
  flex: 1;
  position: relative;
  background: #f8f9fb;
}
.flow-canvas { width: 100%; height: 100%; }

/* ─── 自定义节点 ─── */
.custom-node {
  min-width: 150px;
  background: #fff;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  box-shadow: 0 2px 4px rgba(0,0,0,0.05);
  overflow: hidden;
}
.node-header {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 10px;
  color: #fff;
  font-size: 12px;
  font-weight: 600;
}
.http-header { background: #409eff; }
.script-header { background: #67c23a; }
.sql-header { background: #e6a23c; }
.k8s-header { background: #909399; }
.node-body {
  padding: 8px 10px;
  font-size: 11px;
  color: #606266;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.node-selected {
  border: 2px solid #409eff;
  box-shadow: 0 0 8px rgba(64,158,255,0.3);
}

.handle-top, .handle-bottom, .handle-left, .handle-right {
  width: 10px;
  height: 10px;
  background: #fff !important;
  border: 2px solid #409eff !important;
  border-radius: 50%;
  z-index: 10;
}
.vue-flow__handle {
  z-index: 10;
}
.handle-top { top: -5px !important; }
.handle-bottom { bottom: -5px !important; }
.handle-left { left: -5px !important; }
.handle-right { right: -5px !important; }

/* 连线样式 */
.vue-flow__edge-path {
  stroke: #409eff;
  stroke-width: 2;
}
.vue-flow__arrowhead {
  fill: #409eff;
}

/* ─── 连接菜单 ─── */
.connect-menu {
  position: fixed;
  z-index: 1000;
  background: #fff;
  border: 1px solid #e4e7ed;
  border-radius: 4px;
  box-shadow: 0 4px 12px rgba(0,0,0,0.1);
  padding: 8px 0;
  min-width: 140px;
}
.connect-menu-header {
  padding: 4px 12px;
  font-size: 11px;
  color: #909399;
  border-bottom: 1px solid #f0f0f0;
  margin-bottom: 4px;
}
.connect-menu-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  font-size: 13px;
  cursor: pointer;
  transition: all 0.2s;
}
.connect-menu-item:hover {
  background-color: #f5f7fa;
  color: #409eff;
}

/* ─── 右侧面板 ─── */
.config-panel {
  width: 320px;
  background: #fff;
  border-left: 1px solid #e4e7ed;
  display: flex;
  flex-direction: column;
  z-index: 10;
}
.cp-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  border-bottom: 1px solid #f0f0f0;
  font-weight: 600;
  font-size: 14px;
}
.cp-body {
  padding: 16px;
  flex: 1;
  overflow-y: auto;
}

.code-font :deep(.el-textarea__inner) {
  font-family: 'JetBrains Mono', monospace;
  font-size: 12px;
}

.mt-20 { margin-top: 20px; }

/* 动画 */
.slide-panel-enter-active, .slide-panel-leave-active { transition: all 0.3s ease; }
.slide-panel-enter-from, .slide-panel-leave-to { transform: translateX(100%); opacity: 0; }
</style>
