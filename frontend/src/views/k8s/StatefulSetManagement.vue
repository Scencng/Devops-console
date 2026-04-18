<template>
  <div class="autoops-container">
    <!-- Filter Bar -->
    <div class="filter-bar">
      <div class="filter-left">
        <label class="filter-label">命名空间</label>
        <el-select v-model="selectedNamespace" placeholder="选择命名空间" filterable class="filter-select" @change="fetchData">
          <el-option label="所有" value="all" />
          <el-option v-for="ns in namespaceList" :key="ns.name" :label="ns.name" :value="ns.name" />
        </el-select>
        <label class="filter-label">工作负载名称</label>
        <el-input 
          v-model="searchKeyword" 
          placeholder="请输入名称" 
          class="filter-input"
          clearable
          @keyup.enter="handleSearch"
        />
        <el-button type="primary" class="btn-search" @click="handleSearch">
          <el-icon class="mr-1"><Search /></el-icon> 搜索
        </el-button>
        <el-button class="btn-reset" @click="handleReset">
          <el-icon class="mr-1"><RefreshRight /></el-icon> 重置
        </el-button>
        <el-button type="success" class="btn-monitor">
          <el-icon class="mr-1"><Monitor /></el-icon> 监控仪表板
        </el-button>
        <el-button type="primary" class="btn-create" @click="showCreateDialog = true" v-show="permStore.hasPerm('k8s:statefulset:showcreatedialogtrue')" >
          <el-icon class="mr-1"><Plus /></el-icon> 创建 StatefuSet
        </el-button>
      </div>
    </div>

    <!-- Table -->
    <div class="table-wrapper">
      <el-table 
        :data="statefulSetList" 
        style="width: 100%" class="autoops-table" 
        v-loading="loading"
        :header-cell-style="{ background: '#f5f7fa', color: '#606266', fontWeight: 'bold' }"
        row-class-name="statefulset-row"
      >
        <!-- Name Column -->
        <el-table-column label="名称" min-width="220">
          <template #default="{ row }">
            <div class="name-cell">
              <el-icon class="resource-icon statefulset-icon"><Coin /></el-icon>
              <div class="name-content">
                <a class="name-link" @click="handleViewDetail(row)">{{ row.name }}</a>
                <span class="resource-type">StatefulSet</span>
              </div>
            </div>
          </template>
        </el-table-column>

        <!-- Labels Column -->
        <el-table-column label="标签" min-width="120">
          <template #default="{ row }">
             <div class="label-cell">
                <el-icon class="label-icon"><PriceTag /></el-icon>
                <el-popover placement="top" width="auto" trigger="hover">
                  <template #reference>
                    <el-tag type="danger" effect="plain" round size="small" class="label-count">
                       {{ Object.keys(row.labels || {}).length }}
                    </el-tag>
                  </template>
                  <div class="tags-popover">
                     <span v-for="(val, key) in row.labels" :key="key" class="tag-item">
                        {{ key }}: {{ val }}
                     </span>
                  </div>
                </el-popover>
             </div>
          </template>
        </el-table-column>

        <!-- Pods Column -->
        <el-table-column label="容器组数量" min-width="140" align="center">
           <template #default="{ row }">
              <div class="pods-cell">
                 <div class="pods-count">
                    <el-icon class="monitor-icon"><Platform /></el-icon>
                    <span :class="{'text-green': row.ready === row.replicas, 'text-warning': row.ready < row.replicas}">
                        {{ row.ready || 0 }}/{{ row.replicas || 0 }}
                    </span>
                 </div>
                 <div class="pods-status" :class="{'status-ready': row.ready === row.replicas, 'status-warn': row.ready < row.replicas}">
                    {{ row.ready === row.replicas ? '全部就绪' : '部分就绪' }}
                 </div>
              </div>
           </template>
        </el-table-column>

        <!-- Request/Limits -->
        <el-table-column label="Request/Limits" min-width="180">
           <template #default="{ row }">
              <div class="resource-cell">
                 <div class="res-row">
                    <span class="res-label">CPU:</span>
                    <span class="res-val request">200m</span> / <span class="res-val limit">500m</span>
                 </div>
                 <div class="res-row">
                    <span class="res-label">Mem:</span>
                    <span class="res-val request">300Mi</span> / <span class="res-val limit">500Mi</span>
                 </div>
              </div>
           </template>
        </el-table-column>

        <!-- Image -->
        <el-table-column label="镜像" min-width="200">
           <template #default="{ row }">
              <div class="image-cell">
                 <el-icon class="image-icon"><Box /></el-icon>
                 <span class="image-name" :title="row.image">{{ row.image }}</span>
                 <span v-if="row.containers && row.containers.length > 1" class="more-images">
                    +{{ row.containers.length - 1 }}个镜像
                 </span>
              </div>
           </template>
        </el-table-column>

        <!-- Created Time -->
        <el-table-column label="创建时间" width="160">
            <template #default="{ row }">
               <span class="time-text">{{ formatDate(row.created) }}</span>
            </template>
        </el-table-column>

        <!-- Operations -->
        <el-table-column label="操作" width="280" fixed="right">
            <template #default="{ row }">
               <div class="ops-cell">
                  <el-tooltip content="扩缩容" placement="top">
                    <el-button circle size="small" type="primary" class="op-btn blue" @click="handleScale(row)">
                       <el-icon><Cloudy /></el-icon>
                    </el-button>
                  </el-tooltip>
                  <el-button circle size="small" class="op-btn yellow" @click="handleViewDetail(row)" v-show="permStore.hasPerm('k8s:statefulset:handleviewdetail')" >
                       <el-icon><InfoFilled /></el-icon>
                    </el-button>
                  <el-tooltip content="更新镜像" placement="top">
                    <el-button circle size="small" class="op-btn green" @click="handleUpdate(row)">
                       <el-icon><Setting /></el-icon>
                    </el-button>
                  </el-tooltip>
                  <el-tooltip content="Shell" placement="top">
                    <el-button circle size="small" class="op-btn gray" @click="handleShell(row)">
                       <el-icon><Monitor /></el-icon>
                    </el-button>
                  </el-tooltip>
                  <el-button circle size="small" class="op-btn blue-light" @click="handleLogs(row)" v-show="permStore.hasPerm('k8s:statefulset:handlelogs')" >
                       <el-icon><Document /></el-icon>
                    </el-button>
                  <el-button circle size="small" type="danger" class="op-btn red" @click="handleDelete(row)" v-show="permStore.hasPerm('k8s:statefulset:handledelete')" >
                       <el-icon><Delete /></el-icon>
                    </el-button>
               </div>
            </template>
        </el-table-column>
      </el-table>
       <!-- Pagination -->
       <div class="pagination-container">
          <el-pagination
            v-model:current-page="currentPage"
            v-model:page-size="pageSize"
            :total="total"
            :page-sizes="[10, 20, 50, 100]"
            layout="total, sizes, prev, pager, next, jumper"
            @current-change="fetchData"
            @size-change="handleSizeChange"
          />
       </div>
    </div>

    <!-- Create Dialog -->
    <el-dialog v-model="showCreateDialog" title="创建 StatefulSet" width="900px" :close-on-click-modal="false" append-to-body class="autoops-dialog">
        <el-form label-width="100px">
          <el-form-item label="命名空间">
            <el-select v-model="selectedNamespace" placeholder="选择命名空间" style="width: 100%;">
              <el-option label="default" value="default" />
              <el-option label="kube-system" value="kube-system" />
            </el-select>
          </el-form-item>
          <el-form-item label="YAML配置">
            <el-input
              v-model="yamlContent"
              type="textarea"
              :rows="14"
              placeholder="请输入StatefulSet的YAML配置"
              style="font-family: 'Consolas', 'Monaco', monospace;"
            />
          </el-form-item>
          <el-alert title="YAML示例" type="info" :closable="false">
			<pre style="font-size: 12px; margin: 8px 0;">apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: example-statefulset
spec:
  serviceName: example
  replicas: 3
  selector:
    matchLabels:
      app: example
  template:
    metadata:
      labels:
        app: example
    spec:
      containers:
      - name: example
        image: nginx
        ports:
        - containerPort: 80</pre>
          </el-alert>
        </el-form>
        <template #footer>
          <el-button @click="showCreateDialog = false">取消</el-button>
          <el-button type="primary" @click="handleCreate">创建</el-button>
        </template>
    </el-dialog>

    <!-- Scale Dialog -->
    <el-dialog v-model="showScaleDialog" title="扩缩容" width="400px" append-to-body>
      <div class="p-4">
         <p class="mb-2">目标副本数</p>
         <el-input-number v-model="scaleForm.replicas" :min="0" :max="50" class="w-full" />
      </div>
      <template #footer>
        <el-button @click="showScaleDialog = false">取消</el-button>
        <el-button type="primary" @click="handleScaleConfirm">确定</el-button>
      </template>
    </el-dialog>

  </div>
</template>

<script setup>
import { usePermissionStore } from '@/stores/permissionStore.js'
import { ref, onMounted, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getStatefulSetList, deleteStatefulSet, scaleStatefulSet } from '@/api/k8s/statefulset'
import { getNamespaceList } from '@/api/k8s/namespace'
import { getSelectedInstanceId } from '@/stores/instanceStore'
import dayjs from 'dayjs'

const permStore = usePermissionStore()

const loading = ref(false)
const statefulSetList = ref([])
const searchKeyword = ref('')
const selectedNamespace = ref('default')
const namespaceList = ref([])

const currentPage = ref(1)
const pageSize = ref(10)
const total = ref(0)

const showCreateDialog = ref(false)
const yamlContent = ref('')

const showScaleDialog = ref(false)
const scaleForm = ref({
    row: null,
    replicas: 0
})

watch(selectedNamespace, () => {
    currentPage.value = 1
    fetchData()
})

const fetchData = async () => {
    loading.value = true
    try {
        const instanceId = getSelectedInstanceId()
        const res = await getStatefulSetList(selectedNamespace.value, instanceId)
        statefulSetList.value = res.data?.statefulSetList || []
        total.value = res.data?.total || statefulSetList.value.length
    } catch (e) {
        ElMessage.error('获取 StatefulSet 列表失败')
        statefulSetList.value = []
        total.value = 0
    } finally {
        loading.value = false
    }
}

const fetchNamespaces = async () => {
    try {
        const instanceId = getSelectedInstanceId()
        const res = await getNamespaceList(instanceId)
        namespaceList.value = res.data?.namespaceList || []
        if (!namespaceList.value.some(ns => ns.name === selectedNamespace.value)) {
            selectedNamespace.value = namespaceList.value[0]?.name || 'default'
        }
    } catch (e) {
        ElMessage.error('获取命名空间列表失败')
    }
}

const handleSearch = () => fetchData()
const handleReset = () => { searchKeyword.value = ''; fetchData() }
const handleSizeChange = () => fetchData()
const formatDate = (ts) => dayjs(ts).isValid() ? dayjs(ts).format('YYYY-MM-DD HH:mm:ss') : '-'

const handleViewDetail = (row) => ElMessage.info(`查看详情: ${row.name}`)
const handleScale = (row) => { 
    scaleForm.value.row = row; 
    scaleForm.value.replicas = row.replicas; 
    showScaleDialog.value = true; 
}
const handleScaleConfirm = async () => {
    try {
        await scaleStatefulSet(scaleForm.value.row.namespace, scaleForm.value.row.name, scaleForm.value.replicas) 
        ElMessage.success('扩缩容指令已发送')
        showScaleDialog.value = false
        fetchData()
    } catch(e) { ElMessage.error('操作失败') }
}

const handleUpdate = (row) => ElMessage.info('更新功能待完善')
const handleShell = (row) => ElMessage.info('Shell 功能待完善')
const handleLogs = (row) => ElMessage.info('日志功能待完善')
const handleDelete = (row) => {
    ElMessageBox.confirm(`确定删除 ${row.name}?`, '警告', { type: 'warning' })
    .then(async () => {
         await deleteStatefulSet(row.namespace, row.name)
         ElMessage.success('删除成功')
         fetchData()
    })
}

const handleCreate = () => {
    ElMessage.info('功能待完善')
}

onMounted(() => {
    fetchData()
    fetchNamespaces()
})
</script>

<style scoped>
.filter-bar {
  padding: 16px 0;
  border-bottom: 1px solid #f0f2f5;
  margin-bottom: 16px;
}
.filter-left {
  display: flex;
  align-items: center;
  gap: 12px;
}
.filter-label { font-size: 14px; color: #606266; }
.filter-input { width: 240px; }
.filter-select { width: 140px; }

.table-wrapper { background: white; }

.name-cell {
  display: flex;
  align-items: flex-start;
  gap: 12px;
}
.statefulset-icon {
  font-size: 24px;
  color: var(--primary-color);
  background: rgba(64, 158, 255, 0.1);
  padding: 6px;
  border-radius: 6px;
}
.name-content { display: flex; flex-direction: column; }
.name-link {
  color: var(--primary-color);
  font-weight: 600;
  cursor: pointer;
  text-decoration: none;
  font-size: 14px;
}
.name-link:hover { text-decoration: underline; }
.resource-type { font-size: 12px; color: #E6A23C; margin-top: 2px; }

.label-cell {
  display: flex;
  align-items: center;
  gap: 8px;
}
.label-icon { color: #909399; }
.label-count {
  cursor: pointer;
  background: #fef0f0;
  color: #f56c6c;
  border-color: #fde2e2;
}

.pods-cell {
  display: flex;
  flex-direction: column;
  align-items: center;
}
.pods-count {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 15px;
  font-weight: 600;
}
.monitor-icon { color: #67C23A; font-size: 16px; }
.pods-status { font-size: 11px; margin-top: 2px; }
.status-ready { color: #67C23A; }
.status-warn { color: #E6A23C; }
.text-green { color: #67C23A; }
.text-warning { color: #E6A23C; }

.resource-cell { font-size: 12px; font-family: monospace; }
.res-row { margin-bottom: 2px; }
.res-label { color: #909399; margin-right: 4px; }
.res-val.request { color: #67C23A; }
.res-val.limit { color: #E6A23C; }

.image-cell {
  display: flex;
  align-items: center;
  gap: 8px;
}
.image-icon { color: #909399; }
.image-name {
  font-size: 13px;
  color: #606266;
  max-width: 140px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.more-images { font-size: 12px; color: var(--primary-color); }

.ops-cell {
  display: flex;
  gap: 8px;
  justify-content: flex-start;
}
.op-btn { border: none; font-size: 14px; }
.op-btn.blue { background: rgba(64, 158, 255, 0.1); color: #409EFF; }
.op-btn.blue:hover { background: #409EFF; color: white; }
.op-btn.yellow { background: rgba(230, 162, 60, 0.1); color: #E6A23C; }
.op-btn.yellow:hover { background: #E6A23C; color: white; }
.op-btn.green { background: rgba(103, 194, 58, 0.1); color: #67C23A; }
.op-btn.green:hover { background: #67C23A; color: white; }
.op-btn.gray { background: rgba(144, 147, 153, 0.1); color: #909399; }
.op-btn.gray:hover { background: #909399; color: white; }
.op-btn.blue-light { background: #ecf5ff; color: #409EFF; }
.op-btn.blue-light:hover { background: #409EFF; color: white; }
.op-btn.red { background: rgba(245, 108, 108, 0.1); color: #F56C6C; }
.op-btn.red:hover { background: #F56C6C; color: white; }

.pagination-container {
    padding-top: 20px;
    display: flex;
    justify-content: center;
}
</style>
