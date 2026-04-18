<template>
  <div class="editor-root" v-loading="loadingData">
    <!-- ── 顶部工具栏 ── -->
    <div class="top-bar">
      <div class="top-bar-left">
        <el-button link @click="$router.back()">
          <el-icon><ArrowLeft /></el-icon>
        </el-button>
        <span class="pipeline-title">{{ form.name || (isEditMode ? '编辑流水线' : '新建流水线') }}</span>
      </div>

      <div class="top-bar-meta">
        <el-input v-model="form.name"      size="small" placeholder="流水线名称 *" style="width:140px" />
        <el-input v-model="form.gitUrl"   size="small" placeholder="Git URL *"    style="width:220px">
          <template #prefix><el-icon><Connection /></el-icon></template>
        </el-input>
        <el-input v-model="form.branch"   size="small" placeholder="分支"          style="width:100px">
          <template #prefix><el-icon><Rank /></el-icon></template>
        </el-input>
        <el-input v-model="form.gitToken" size="small" placeholder="Git Token" type="password" show-password style="width:130px">
          <template #prefix><el-icon><Key /></el-icon></template>
        </el-input>
        <el-select v-model="form.k8sInstanceId" size="small" placeholder="集群 *" style="width:130px">
          <el-option v-for="item in k8sInstanceList" :key="item.id" :label="item.name" :value="item.id" />
        </el-select>
      </div>

      <div class="top-bar-right">
        <el-button size="small" :icon="Plus"    @click="handleAddStep">添加步骤</el-button>
        <el-button size="small" :icon="Refresh" @click="handleReset"   v-if="isEditMode">重置</el-button>
        <el-button size="small" type="primary"  :icon="Check" :loading="saving" @click="handleSave">保存</el-button>
      </div>
    </div>

    <!-- ── 主编辑区 ── -->
    <div class="editor-main">
      <!-- 画布 -->
      <div class="canvas-wrap">
        <VueFlow
          v-model:nodes="nodes"
          v-model:edges="edges"
          :default-zoom="0.85"
          :min-zoom="0.2"
          :max-zoom="4"
          :fit-view-on-init="false"
          :connect-on-click="false"
          connection-mode="loose"
          class="flow-canvas"
          @connect="handleConnect"
          @node-click="onNodeClick"
          @pane-click="onPaneClick"
          @edge-click="onEdgeClick"
        >
          <Background pattern-color="#e0e0e0" :gap="24" />
          <Controls />

          <!-- git-clone 固定起始节点 -->
          <template #node-gitclone="{ data }">
            <div class="gc-node conn-node">
              <Handle id="gc-src-top"    type="source" position="top"    class="ch ch-top" />
              <Handle id="gc-src-bottom" type="source" position="bottom" class="ch ch-bottom" />
              <Handle id="gc-src-left"   type="source" position="left"   class="ch ch-left" />
              <Handle id="gc-src-right"  type="source" position="right"  class="ch ch-right" />
              <div class="gc-header">
                <el-icon class="gc-icon"><Connection /></el-icon>
                <span class="gc-title">git-clone</span>
              </div>
              <div class="gc-url" :title="data.gitUrl">{{ data.gitUrl || 'Git URL 未设置' }}</div>
              <div class="gc-branch">分支: {{ data.branch || 'master' }}</div>
            </div>
          </template>

          <!-- 自定义步骤节点 -->
          <template #node-step="{ data, id }">
            <div class="step-node conn-node" :class="{ 'step-node--selected': selectedNodeId === id }">
              <Handle :id="id+'-s-top'"    type="source" position="top"    class="ch ch-top" />
              <Handle :id="id+'-s-bottom'" type="source" position="bottom" class="ch ch-bottom" />
              <Handle :id="id+'-s-left'"   type="source" position="left"   class="ch ch-left" />
              <Handle :id="id+'-s-right'"  type="source" position="right"  class="ch ch-right" />
              <div class="sn-header">
                <el-icon class="sn-icon"><Box /></el-icon>
                <span class="sn-name" :title="data.name">{{ data.name || '未命名' }}</span>
                <button class="sn-del" @click.stop="deleteStep(id)" title="删除步骤">&times;</button>
              </div>
              <div class="sn-image" :title="data.image">{{ data.image || '未设置镜像' }}</div>
            </div>
          </template>
        </VueFlow>

        <!-- 只有 git-clone 一个节点时显示提示 -->
        <div class="canvas-hint-bar" v-if="nodes.length <= 1">
          点击「添加步骤」创建节点，拖拽节点右侧锚点连线建立依赖关系
        </div>
      </div>

      <!-- 右侧配置面板 -->
      <transition name="slide-panel">
        <div class="config-panel" v-if="selectedNodeId">
          <div class="cp-header">
            <span>步骤配置</span>
            <el-button link size="small" @click="closePanel">
              <el-icon><Close /></el-icon>
            </el-button>
          </div>
          <div class="cp-body" v-if="selectedNodeId">
            <el-form label-position="top" size="small">
              <el-form-item label="步骤名称" required>
                <el-input v-model="editing.name" placeholder="check-code" @blur="syncNodeData" />
              </el-form-item>
              <el-form-item label="运行镜像" required>
                <el-input v-model="editing.image" placeholder="golang:1.18" @blur="syncNodeData">
                  <template #prefix><el-icon><Monitor /></el-icon></template>
                </el-input>
              </el-form-item>
              <el-form-item label="Shell 命令">
                <el-input
                  v-model="editing.commands"
                  type="textarea"
                  :rows="16"
                  class="code-textarea"
                  placeholder="# 输入 Shell 命令..."
                  @blur="syncNodeData"
                />
              </el-form-item>
              <el-form-item label="依赖步骤">
                <div class="deps-tags">
                  <el-tag
                    v-for="dep in currentDeps"
                    :key="dep"
                    closable
                    size="small"
                    @close="removeDep(dep)"
                  >{{ dep }}</el-tag>
                  <span v-if="currentDeps.length === 0" class="no-deps">无（拖拽节点锚点连线自动添加）</span>
                </div>
              </el-form-item>
            </el-form>
            <el-button
              type="danger"
              plain
              :icon="Delete"
              @click="deleteStep(selectedNodeId)"
              style="width:100%;margin-top:8px"
            >删除此步骤</el-button>
          </div>
        </div>
      </transition>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted, nextTick, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { VueFlow, Handle, useVueFlow } from '@vue-flow/core'
import { Background } from '@vue-flow/background'
import { Controls } from '@vue-flow/controls'
import '@vue-flow/core/dist/style.css'
import '@vue-flow/controls/dist/style.css'

import {
  createPipeline, updatePipeline, getPipeline,
  getPipelineSteps, createPipelineStep, updatePipelineStep, deletePipelineStep
} from '@/api/cicd.js'
import { getInstanceList } from '@/api/instance.js'
import { ElMessage } from 'element-plus'
import {
  ArrowLeft, Check, Plus, Refresh, Box, Monitor,
  Delete, Close, Grid, Connection, Rank, Key
} from '@element-plus/icons-vue'

// ─── VueFlow ─────────────────────────────────────────────────────────────────
const { fitView } = useVueFlow()

// ─── Git-clone 固定节点 ──────────────────────────────────────────────────────
const GIT_CLONE_ID = '__gitclone__'

// 辅助：将 id 连接到 git-clone
const addGitCloneEdge = (targetId) => {
  const exists = edges.value.some(e => e.source === GIT_CLONE_ID && e.target === targetId)
  if (exists) return
  edges.value = [...edges.value, {
    id: `e-gc-${targetId}`,
    source: GIT_CLONE_ID,
    target: targetId,
    animated: true,
    style: { stroke: '#409eff', strokeWidth: 2 }
  }]
}

const ensureGitCloneNode = () => {
  const data = { gitUrl: form.gitUrl, branch: form.branch || 'master' }
  const existing = nodes.value.find(n => n.id === GIT_CLONE_ID)
  if (existing) {
    existing.data = { ...data }
  } else {
    nodes.value = [{
      id: GIT_CLONE_ID,
      type: 'gitclone',
      position: { x: 30, y: 120 },
      draggable: true,
      selectable: false,
      data
    }, ...nodes.value]
  }
  // 对所有没有入边的用户步骤（排除 git-clone 自身），自动连接
  const userNodes = nodes.value.filter(n => n.id !== GIT_CLONE_ID)
  const hasIncomingFromUser = (nodeId) =>
    edges.value.some(e => e.target === nodeId && e.source !== GIT_CLONE_ID)
  userNodes.forEach(n => {
    if (!hasIncomingFromUser(n.id)) addGitCloneEdge(n.id)
  })
}

const route  = useRoute()
const router = useRouter()
const pipelineId = route.params.id
const isEditMode = computed(() => !!pipelineId)

// ─── State ───────────────────────────────────────────────────────────────────
const saving      = ref(false)
const loadingData = ref(false)
const nodes       = ref([])
const edges       = ref([])
const deletedStepIds  = ref([])
const selectedNodeId  = ref(null)
let   nodeSeq = 0

const form = reactive({
  id: null, name: '', gitUrl: '', branch: 'master',
  gitToken: '', argoTemplate: 'main', k8sInstanceId: null
})

const k8sInstanceList = ref([])

// 当前正在编辑的步骤数据（与右侧面板双向绑定）
const editing = reactive({ name: '', image: '', commands: '', _stepId: null })

// ─── Computed ─────────────────────────────────────────────────────────────────
const currentDeps = computed(() => {
  if (!selectedNodeId.value) return []
  return edges.value
    .filter(e => e.target === selectedNodeId.value)
    .map(e => nodes.value.find(n => n.id === e.source)?.data?.name)
    .filter(Boolean)
})

// ─── Node Events ─────────────────────────────────────────────────────────────
const onNodeClick = (eventOrNode, maybeNode) => {
  // VueFlow 不同版本事件参数可能是 (event, node) 或 (nodeObj) 形式
  const node = maybeNode ?? eventOrNode?.node ?? eventOrNode
  if (!node?.id) return
  if (node.id === GIT_CLONE_ID) return  // git-clone 节点不打开配置面板
  if (selectedNodeId.value && selectedNodeId.value !== node.id) syncNodeData()
  selectedNodeId.value = node.id
  const d = node.data || {}
  Object.assign(editing, { name: d.name || '', image: d.image || '', commands: d.commands || '', _stepId: d.stepId || null })
}

const onPaneClick = () => {
  syncNodeData()
  selectedNodeId.value = null
}

const closePanel = () => {
  syncNodeData()
  selectedNodeId.value = null
}

// 将 editing 数据同步回节点 data
const syncNodeData = () => {
  const node = nodes.value.find(n => n.id === selectedNodeId.value)
  if (!node) return
  node.data = { ...node.data, name: editing.name, image: editing.image, commands: editing.commands }
}

// 点击边删除—— VueFlow 以单对象 { event, edge } 传入
const onEdgeClick = (params) => {
  const edge = params?.edge ?? params
  if (!edge?.id) return
  edges.value = edges.value.filter(ed => ed.id !== edge.id)
}

const handleConnect = (params) => {
  if (params.source === params.target) return
  const exists = edges.value.some(e =>
    e.source === params.source &&
    e.target === params.target &&
    e.sourceHandle === params.sourceHandle &&
    e.targetHandle === params.targetHandle
  )
  if (exists) return
  edges.value = [...edges.value, {
    id: `e-${params.source}-${params.sourceHandle}-${params.target}-${params.targetHandle}`,
    source: params.source,
    target: params.target,
    sourceHandle: params.sourceHandle,   // 固定起始 handle，防止重绘时路径变化
    targetHandle: params.targetHandle,   // 固定终止 handle
    type: 'smoothstep',
    animated: true,
    style: { stroke: '#409eff', strokeWidth: 2 }
  }]
}

const removeDep = (depName) => {
  const src = nodes.value.find(n => n.data?.name === depName)
  if (!src) return
  edges.value = edges.value.filter(e => !(e.source === src.id && e.target === selectedNodeId.value))
}

// ─── Add / Delete Step ───────────────────────────────────────────────────────
const handleAddStep = () => {
  syncNodeData()
  const seq  = ++nodeSeq
  const id   = `step-new-${Date.now()}-${seq}`
  const name = `step-${seq}`
  // 步骤从 x:320 开始，留出 git-clone 节点空间
  const userNodes = nodes.value.filter(n => n.id !== GIT_CLONE_ID)
  const col   = userNodes.length % 3
  const row   = Math.floor(userNodes.length / 3)
  const nodeX = 320 + col * 240
  const nodeY = 60 + row * 140
  nodes.value = [...nodes.value, {
    id,
    type: 'step',
    position: { x: nodeX, y: nodeY },
    data: { stepId: null, name, image: 'ubuntu:latest', commands: 'echo "hello"' }
  }]
  // 不自动连接，由用户手动拖拽连线
  nextTick(() => {
    selectedNodeId.value = id
    Object.assign(editing, { name, image: 'ubuntu:latest', commands: 'echo "hello"', _stepId: null })
  })
}

const deleteStep = (id) => {
  const node = nodes.value.find(n => n.id === id)
  if (node?.data?.stepId) deletedStepIds.value.push(node.data.stepId)
  nodes.value = nodes.value.filter(n => n.id !== id)
  edges.value = edges.value.filter(e => e.source !== id && e.target !== id)
  if (selectedNodeId.value === id) selectedNodeId.value = null
}

// ─── Graph <-> API 转换 ───────────────────────────────────────────────────────
// 对用户节点做拓扑排序（Kahn BFS），返回按执行顺序排列的节点列表
const topoSortNodes = () => {
  const userNodes = nodes.value.filter(n => n.id !== GIT_CLONE_ID)
  const idToIdx = new Map(userNodes.map((n, i) => [n.id, i]))
  // 构建邻接表（只计算用户步骤之间的依赖边）
  const inDegree = new Array(userNodes.length).fill(0)
  const adjList  = Array.from({ length: userNodes.length }, () => [])
  edges.value.forEach(e => {
    if (e.source === GIT_CLONE_ID || e.target === GIT_CLONE_ID) return
    const src = idToIdx.get(e.source)
    const tgt = idToIdx.get(e.target)
    if (src === undefined || tgt === undefined) return
    adjList[src].push(tgt)
    inDegree[tgt]++
  })
  // Kahn BFS
  const queue  = []
  const sorted = []
  inDegree.forEach((deg, i) => { if (deg === 0) queue.push(i) })
  while (queue.length) {
    const cur = queue.shift()
    sorted.push(cur)
    adjList[cur].forEach(next => {
      if (--inDegree[next] === 0) queue.push(next)
    })
  }
  // 如果有环（理论上不应该），剩余节点追加到末尾
  if (sorted.length < userNodes.length) {
    for (let i = 0; i < userNodes.length; i++) {
      if (!sorted.includes(i)) sorted.push(i)
    }
  }
  return sorted.map(i => userNodes[i])
}

const buildStepsFromGraph = (pid) => {
  // 使用拓扑排序后的节点顺序，sort 值 = 拓扑序索引，保证正确的执行顺序
  const sortedNodes = topoSortNodes()
  return sortedNodes.map((node, index) => {
    const deps = edges.value
      .filter(e => e.target === node.id && e.source !== GIT_CLONE_ID)
      .map(e => nodes.value.find(n => n.id === e.source)?.data?.name)
      .filter(Boolean)
    return {
      name:       node.data.name || `step-${index}`,
      image:      node.data.image || '',
      commands:   node.data.commands || '',
      dependsOn:  deps.length ? deps.join(',') : null,
      sort:       index,   // 拓扑序索引，确保 sort 严格反映执行顺序
      pipelineId: Number(pid)
    }
  })
}

const loadGraph = (steps) => {
  // 拓扑分层布局
  const nameToStep = new Map(steps.map(s => [s.name, s]))
  const levels = new Map()
  const getLevel = (name, visited = new Set()) => {
    if (levels.has(name)) return levels.get(name)
    if (visited.has(name)) return 0
    visited.add(name)
    const step = nameToStep.get(name)
    if (!step?.dependsOn) { levels.set(name, 0); return 0 }
    const deps = step.dependsOn.split(',').map(d => d.trim()).filter(Boolean)
    const maxL = Math.max(...deps.map(d => getLevel(d, new Set(visited))))
    levels.set(name, maxL + 1)
    return maxL + 1
  }
  steps.forEach(s => getLevel(s.name))

  const levelGroups = new Map()
  steps.forEach(s => {
    const l = levels.get(s.name) || 0
    if (!levelGroups.has(l)) levelGroups.set(l, [])
    levelGroups.get(l).push(s)
  })

  const newNodes = []
  levelGroups.forEach((group, level) => {
    group.forEach((step, idx) => {
      newNodes.push({
        id: `step-${step.id || step.name}`,
        type: 'step',
        position: { x: level * 280 + 300, y: idx * 160 + 200 - (group.length - 1) * 80 },
        data: { stepId: step.id, name: step.name, image: step.image, commands: step.commands }
      })
    })
  })

  const newEdges = []
  steps.forEach(step => {
    if (!step.dependsOn) return
    const targetId = `step-${step.id || step.name}`
    step.dependsOn.split(',').map(d => d.trim()).filter(Boolean).forEach(depName => {
      const dep = steps.find(s => s.name === depName)
      if (!dep) return
      const sourceId = `step-${dep.id || dep.name}`
      newEdges.push({
        id: `e-${sourceId}-${targetId}`,
        source: sourceId,
        target: targetId,
        animated: true,
        style: { stroke: '#409eff', strokeWidth: 2 }
      })
    })
  })

  nodes.value = newNodes
  edges.value = newEdges
  nodeSeq = newNodes.length
  ensureGitCloneNode()
}

// ─── API ─────────────────────────────────────────────────────────────────────
const fetchData = async () => {
  if (!isEditMode.value) return
  loadingData.value = true
  try {
    const [pRes, sRes] = await Promise.all([getPipeline(pipelineId), getPipelineSteps(pipelineId)])
    const data = pRes.data.data
    Object.assign(form, { ...data, gitToken: data.gitToken || '' })
    form.id = data.id
    loadGraph(sRes.data.data || [])
  } catch (e) {
    ElMessage.error('加载失败')
  } finally {
    loadingData.value = false
  }
}

const fetchK8sInstances = async () => {
  try {
    const res  = await getInstanceList({ page: 1, page_size: 100 })
    // 兼容多种响应格式
    const raw  = res.data?.list?.data || res.data?.data || res.data || []
    const list = Array.isArray(raw) ? raw : []
    k8sInstanceList.value = list.filter(i => {
      const t = (i.instance_type || i.type || '').toLowerCase()
      return t === 'kubernetes' || t === 'k8s'
    })
    if (!form.k8sInstanceId && k8sInstanceList.value.length > 0) {
      form.k8sInstanceId = k8sInstanceList.value[0].id
    }
  } catch (e) { /* 静默 */ }
}

const handleSave = async () => {
  syncNodeData() // 先同步右侧面板到节点
  if (!form.name || !form.gitUrl || !form.k8sInstanceId) {
    ElMessage.warning('请填写流水线名称、Git URL 和目标集群')
    return
  }
  saving.value = true
  try {
    // 1. 保存 pipeline 基础信息
    let pid = form.id
    const payload = { ...form, argoTemplate: form.argoTemplate || 'main' }
    if (pid) {
      await updatePipeline(pid, payload)
    } else {
      const res = await createPipeline(payload)
      pid = res.data.data?.id
      form.id = pid
    }
    if (!pid) throw new Error('无法获取流水线 ID')

    // 2. 获取当前所有已存在的步骤并全部删除
    //    这样可以避免旧的 sort 值残留，配合拓扑排序确保数据库中 sort 完全重置
    const existingSteps = nodes.value
      .filter(n => n.id !== GIT_CLONE_ID && n.data?.stepId)
      .map(n => n.data.stepId)
    const allDeleteIds = [...new Set([...existingSteps, ...deletedStepIds.value])]
    for (const delId of allDeleteIds) await deletePipelineStep(delId)
    deletedStepIds.value = []

    // 3. 按拓扑排序批量新建步骤（sort 由拓扑序决定，0 = 最先执行）
    const steps = buildStepsFromGraph(pid)
    for (const step of steps) {
      // 注意：此处不传 id，全部作为新建，避免 ID 混乱
      const { id: _omit, ...stepPayload } = step
      const res = await createPipelineStep(stepPayload)
      const newId = res.data.data?.id
      if (newId) {
        // 回写新 stepId 到节点 data，下次保存时可识别
        const node = nodes.value.find(n => n.data?.name === step.name && n.id !== GIT_CLONE_ID)
        if (node) node.data = { ...node.data, stepId: newId }
      }
    }

    ElMessage.success('保存成功')
    if (!isEditMode.value) {
      router.replace(`/cicd/pipelines/${pid}/edit`)
      setTimeout(fetchData, 500)
    }
  } catch (e) {
    ElMessage.error('保存失败: ' + (e.message || ''))
  } finally {
    saving.value = false
  }
}

const handleReset = () => {
  selectedNodeId.value = null
  deletedStepIds.value = []
  fetchData()
}

onMounted(() => {
  fetchK8sInstances()
  fetchData()
  ensureGitCloneNode()
})

// 表单 gitUrl/branch 变化时同步更新 git-clone 节点内容
watch(
  () => [form.gitUrl, form.branch],
  () => ensureGitCloneNode()
)
</script>

<style scoped>
/* ─── 根布局 ─────────────────────────────────────────────────── */
.editor-root {
  display: flex;
  flex-direction: column;
  height: 100vh;
  background: #f5f7fa;
  overflow: hidden;
}

/* ─── 顶部工具栏 ──────────────────────────────────────────────── */
.top-bar {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 8px 16px;
  background: #fff;
  border-bottom: 1px solid #e4e7ed;
  min-height: 52px;
  flex-shrink: 0;
  flex-wrap: wrap;
}
.top-bar-left {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
}
.pipeline-title {
  font-size: 15px;
  font-weight: 600;
  color: #303133;
  max-width: 200px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.top-bar-meta {
  display: flex;
  align-items: center;
  gap: 8px;
  flex: 1;
  flex-wrap: wrap;
}
.top-bar-right {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
}

/* ─── 主编辑区 ───────────────────────────────────────────────── */
.editor-main {
  display: flex;
  flex: 1;
  overflow: hidden;
  position: relative;
}
.canvas-wrap {
  flex: 1;
  position: relative;
  overflow: hidden;
}
.flow-canvas {
  width: 100%;
  height: 100%;
  background: #f8f9fb;
}

/* ─── 空状态提示 ─────────────────────────────────────────────── */
.canvas-empty {
  position: absolute;
  inset: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  pointer-events: none;
  color: #c0c4cc;
  gap: 12px;
}
.canvas-empty .empty-icon { font-size: 48px; }
.canvas-empty p { font-size: 14px; text-align: center; line-height: 1.8; }

/* ─── 底部提示条 ──────────────────────── */
.canvas-hint-bar {
  position: absolute;
  bottom: 12px;
  left: 50%;
  transform: translateX(-50%);
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 5px 16px;
  background: rgba(255,255,255,.88);
  border: 1px solid #e4e7ed;
  border-radius: 20px;
  font-size: 12px;
  color: #909399;
  pointer-events: none;
  backdrop-filter: blur(4px);
  white-space: nowrap;
}

/* ─── 步骤节点 ───────────────────────────────────────────────── */
.step-node {
  width: 180px;
  background: #fff;
  border: 2px solid #dcdfe6;
  border-radius: 8px;
  padding: 10px 12px;
  box-shadow: 0 2px 8px rgba(0,0,0,.06);
  cursor: pointer;
  transition: all .2s;
  position: relative;
}
.step-node:hover { box-shadow: 0 4px 16px rgba(64,158,255,.15); border-color: #a0cfff; }
.step-node--selected { border-color: #409eff; box-shadow: 0 0 0 3px rgba(64,158,255,.2); }

.sn-header {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-bottom: 4px;
}
.sn-icon { color: #409eff; font-size: 14px; flex-shrink: 0; }
.sn-name {
  flex: 1;
  font-size: 13px;
  font-weight: 600;
  color: #303133;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.sn-del {
  background: none;
  border: none;
  cursor: pointer;
  color: #c0c4cc;
  font-size: 16px;
  line-height: 1;
  padding: 0 2px;
  transition: color .2s;
}
.sn-del:hover { color: #f56c6c; }
.sn-image {
  font-size: 11px;
  color: #909399;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  padding-left: 20px;
}

/* ─── 右侧配置面板 ───────────────────────────────────────────── */
.config-panel {
  width: 320px;
  flex-shrink: 0;
  background: #fff;
  border-left: 1px solid #e4e7ed;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}
.cp-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  border-bottom: 1px solid #f0f0f0;
  font-weight: 600;
  font-size: 14px;
  color: #303133;
  flex-shrink: 0;
}
.cp-body {
  flex: 1;
  overflow-y: auto;
  padding: 16px;
}
.deps-tags { display: flex; flex-wrap: wrap; gap: 6px; }
.no-deps { font-size: 12px; color: #c0c4cc; }

/* ─── 面板滑入动画 ───────────────────────────────────────────── */
.slide-panel-enter-active, .slide-panel-leave-active { transition: width .25s ease, opacity .2s ease; overflow: hidden; }
.slide-panel-enter-from, .slide-panel-leave-to { width: 0; opacity: 0; }
.slide-panel-enter-to, .slide-panel-leave-from { width: 320px; opacity: 1; }

/* ─── Code 文本框 ───────────────────────────────────────────── */
.code-textarea :deep(.el-textarea__inner) {
  font-family: 'JetBrains Mono', 'Consolas', monospace;
  font-size: 12px;
  background: #fafafa;
}
/* ─── git-clone 起始节点 ───────────────────────────────── */
.gc-node {
  width: 180px;
  background: linear-gradient(135deg, #e8f4fd, #d0eafb);
  border: 2px solid #409eff;
  border-radius: 8px;
  padding: 8px 10px;
  cursor: default;
  user-select: none;
}
.gc-header {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-bottom: 4px;
}
.gc-icon { color: #409eff; font-size: 14px; flex-shrink: 0; }
.gc-title { font-size: 12px; font-weight: 700; color: #1a6fbc; }
.gc-url {
  font-size: 10px;
  color: #606266;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  padding-left: 2px;
}
.gc-branch { font-size: 10px; color: #909399; padding-left: 2px; margin-top: 2px; }

/* ─── 四方向连接点（亿图风格，hover 显示） ─────────────── */
/* Handle 组件渲染的根元素，用 :deep 穿透 scoped 边界 */
:deep(.ch) {
  width: 10px !important;
  height: 10px !important;
  background: #fff !important;
  border: 2px solid #409eff !important;
  border-radius: 50% !important;
  opacity: 0;
  transition: opacity .15s, transform .15s;
  z-index: 10;
}

/* hover 父节点时浮现连接点 */
.conn-node:hover :deep(.ch) { opacity: 1 !important; }
/* handle 自身 hover 时放大 */
:deep(.vue-flow__handle.ch:hover) {
  opacity: 1 !important;
  transform: scale(1.3) !important;
  background: #409eff !important;
}

/* ─── 边点击删除提示 ────────────────────────────────────── */
:deep(.vue-flow__edge) { cursor: pointer; }
:deep(.vue-flow__edge:hover .vue-flow__edge-path) {
  stroke: #f56c6c !important;
  stroke-width: 3px !important;
}
</style>
