<template>
  <div class="graph-container">
    <VueFlow
      v-model="elements"
      :default-zoom="1.2"
      :min-zoom="0.2"
      :max-zoom="4"
      :fit-view-on-init="true"
      class="pipeline-flow"
    >
      <Background pattern-color="#e1e1e1" :gap="20" />
      <Controls />
      
      <template #node-custom="{ data }">
        <div 
          class="custom-node" 
          :class="[getStatusClass(data.status), { active: data.selected }]"
          @click="onNodeClick(data)"
        >
          <div class="node-icon">
            <el-icon v-if="data.status === 'Succeeded'"><Check /></el-icon>
            <el-icon v-else-if="data.status === 'Running'" class="is-loading"><Loading /></el-icon>
            <el-icon v-else-if="data.status === 'Failed'"><Close /></el-icon>
            <el-icon v-else><MoreFilled /></el-icon>
          </div>
          <div class="node-content">
            <div class="node-title" :title="data.label">{{ data.label }}</div>
            <div class="node-subtitle">{{ data.duration || '-' }}</div>
          </div>
          <div class="node-handles">
            <Handle type="target" position="left" :style="{ opacity: 0 }" />
            <Handle type="source" position="right" :style="{ opacity: 0 }" />
          </div>
        </div>
      </template>
    </VueFlow>
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted } from 'vue'
import { VueFlow, useVueFlow, Handle } from '@vue-flow/core'
import { Background } from '@vue-flow/background'
import { Controls } from '@vue-flow/controls'
import { Check, Close, MoreFilled, Loading } from '@element-plus/icons-vue'
import '@vue-flow/core/dist/style.css'
import '@vue-flow/controls/dist/style.css'

const props = defineProps({
  steps: { type: Array, default: () => [] },
  activeStepName: { type: String, default: '' }
})

const emit = defineEmits(['node-click'])

const { fitView } = useVueFlow()
const elements = ref([])

// Simple Level-based Layout Algorithm
const layoutElements = (steps) => {
  if (!steps || steps.length === 0) return []

  const nodes = []
  const edges = []
  const nodeMap = new Map()
  
  // 1. Create Nodes Map
  steps.forEach(step => {
    nodeMap.set(step.name, {
      ...step,
      level: 0,
      children: []
    })
  })

  // 2. Build Hierarchy & Calculate Levels
  // Since backend 'sort' implies order, but ideally we use 'dependsOn' if available.
  // Assuming 'sort' gives a rough topological order for now if dependsOn is missing.
  // We'll place them in a grid.
  
  let x = 0
  let y = 0
  const X_GAP = 180  // Reduced from 250
  const Y_GAP = 80   // Reduced from 100

  // Fallback: Linear layout if no complex dependencies
  steps.forEach((step, index) => {
    // For now, let's just do a linear layout based on sort/index
    // A real DAG layout would be more complex, but "Professional" implies looking good.
    // Linear is clean.
    
    const isSelected = step.name === props.activeStepName

    nodes.push({
      id: step.name,
      type: 'custom',
      position: { x: index * X_GAP, y: 100 },
      data: { 
        label: step.name, 
        status: step.status, 
        duration: step.duration,
        selected: isSelected,
        name: step.name // pass full name for event
      },
    })

    if (index > 0) {
      edges.push({
        id: `e-${index-1}-${index}`,
        source: steps[index-1].name,
        target: step.name,
        animated: true,
        style: { stroke: '#b1b1b7', strokeWidth: 2 },
      })
    }
  })

  return [...nodes, ...edges]
}

watch(() => [props.steps, props.activeStepName], () => {
  elements.value = layoutElements(props.steps)
  setTimeout(() => fitView({ padding: 0.2, maxZoom: 1 }), 50)
}, { deep: true, immediate: true })

const getStatusClass = (status) => {
     const map = { 'Succeeded': 'node-success', 'Failed': 'node-failed', 'Running': 'node-running', 'Pending': 'node-pending' }
     return map[status] || 'node-pending'
}

const onNodeClick = (data) => {
  emit('node-click', data.name)
}
</script>

<style scoped>
.graph-container {
  width: 100%;
  height: 100%;
  background: #f8f9fa;
}

.pipeline-flow {
  background: #f0f2f5;
  width: 100%;
  height: 100%;
}

.custom-node {
  display: flex;
  align-items: center;
  width: 150px; /* Reduced from 200px */
  padding: 8px 10px; /* Reduced padding */
  background: #fff;
  border: 1px solid #dcdfe6;
  border-radius: 6px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.04);
  cursor: pointer;
  transition: all 0.2s;
}

.custom-node:hover {
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
  border-color: var(--primary-color);
}

.custom-node.active {
  border-color: var(--primary-color);
  background: #ecf5ff;
  box-shadow: 0 0 0 2px rgba(64, 158, 255, 0.2);
}

.node-icon {
  width: 24px; /* Reduced from 32px */
  height: 24px; /* Reduced from 32px */
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-right: 8px; /* Reduced margin */
  background: #f4f4f5;
  color: #909399;
  font-size: 14px; /* Reduced font size */
}

.node-content {
  flex: 1;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  justify-content: center;
}

.node-title {
  font-weight: 600;
  font-size: 13px; /* Reduced from 14px */
  color: #303133;
  margin-bottom: 2px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.node-subtitle {
  font-size: 12px; /* Same, already small */
  color: #909399;
  line-height: 1;
}

/* Status Colors */
.node-success .node-icon { background: #f0f9eb; color: #67c23a; }
.node-success { border-left: 3px solid #67c23a; }

.node-failed .node-icon { background: #fef0f0; color: #f56c6c; }
.node-failed { border-left: 3px solid #f56c6c; }

.node-running .node-icon { background: #ecf5ff; color: #409eff; }
.node-running { border-left: 3px solid #409eff; }

.node-pending .node-icon { background: #f4f4f5; color: #909399; }
</style>
