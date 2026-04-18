<template>
  <div class="host-management">
    <div class="host-layout">
      <!-- 左侧资产分组树 -->
      <div class="group-panel">
        <div class="group-header">
          <div class="group-title-row">
            <span class="group-title">资产分组</span>
            <div class="group-stats">
              <span class="stat-num">{{ totalGroupCount }}</span>
              <span class="stat-label">ASSET<br>GROUPS</span>
              <el-icon class="stat-icon success"><CircleCheck /></el-icon>
              <el-icon class="stat-icon refresh-btn" @click="fetchGroups"><Refresh /></el-icon>
              <el-icon class="stat-icon add-btn" @click="openGroupDialog(null, 'createRoot')"><Plus /></el-icon>
            </div>
          </div>
          <el-input
            v-model="groupSearch"
            placeholder="搜索分组..."
            clearable
            size="small"
            class="group-search"
          >
            <template #prefix><el-icon><Search /></el-icon></template>
          </el-input>
        </div>

        <!-- 分组树（右键菜单） -->
        <el-scrollbar class="group-tree-scroll">
          <el-tree
            ref="groupTreeRef"
            :data="filteredGroups"
            :props="{ label: 'name', children: 'children' }"
            node-key="id"
            :default-expanded-keys="expandedKeys"
            :highlight-current="true"
            @node-click="handleGroupClick"
            @node-contextmenu="handleContextMenu"
          >
            <template #default="{ node, data }">
              <div class="tree-node">
                <el-icon class="node-folder"><Folder /></el-icon>
                <span class="node-label">{{ data.name }}</span>
                <el-tag size="small" type="info" class="node-count">({{ data.hostCount || 0 }})</el-tag>
              </div>
            </template>
          </el-tree>
        </el-scrollbar>
      </div>

      <!-- 右侧主机内容 -->
      <div class="host-panel">
        <!-- 统计卡片行 -->
        <div class="stats-bar">
          <div class="stat-item">
            <el-icon class="stat-icon-monitor"><Monitor /></el-icon>
            <span class="stat-text">总计: <b>{{ stats.total }}</b></span>
          </div>
          <div class="stat-item">
            <el-icon style="color:#f90;"><Monitor /></el-icon>
            <span class="stat-text">Linux <b>{{ stats.linux }}</b></span>
          </div>
          <div class="stat-item">
            <el-icon style="color:#0078d4;"><Monitor /></el-icon>
            <span class="stat-text">Windows <b>{{ stats.windows }}</b></span>
          </div>
          <div class="stat-item">
            <el-icon style="color:#f56c6c;"><CircleClose /></el-icon>
            <span class="stat-text">离线 <b>{{ stats.offline }}</b></span>
          </div>
          <div class="stat-item">
            <span class="cloud-badge huawei">华为</span>
            <b>{{ stats.huawei }}</b>
          </div>
          <div class="stat-item">
            <span class="cloud-badge aliyun">阿里云</span>
            <b>{{ stats.aliyun }}</b>
          </div>
          <div class="stat-item">
            <span class="cloud-badge aws">aws</span>
            <b>{{ stats.aws }}</b>
          </div>
        </div>

        <!-- 搜索栏 -->
        <div class="search-bar">
          <el-input v-model="searchForm.name" placeholder="请输入主机名称" clearable style="width:180px" />
          <el-input v-model="searchForm.ip" placeholder="请输入IP地址" clearable style="width:180px" />
          <el-select v-model="searchForm.status" placeholder="请选择主机状态" clearable style="width:150px">
            <el-option label="在线" value="online" />
            <el-option label="离线" value="offline" />
            <el-option label="未知" value="unknown" />
          </el-select>
          <el-button type="primary" :icon="Search" @click="fetchHosts">搜索</el-button>
          <el-button type="warning" :icon="AlarmClock" @click="handleAlarm">告警</el-button>
          <el-button type="success" :icon="Plus" @click="openHostDialog()">新建</el-button>
          <el-button type="danger" :icon="Monitor" @click="handleTerminal">终端</el-button>
          <el-button :disabled="!selectedHosts.length" @click="handleBatchDelete">批量操作</el-button>
        </div>

        <!-- 主机表格 -->
        <el-table
          :data="tableData"
          v-loading="loading"
          border
          stripe
          @selection-change="handleSelectionChange"
          style="width: 100%"
        >
          <el-table-column type="selection" width="46" />
          <el-table-column prop="name" label="主机名称" min-width="140" show-overflow-tooltip />
          <el-table-column prop="ip" label="IP地址" min-width="130" show-overflow-tooltip />
          <el-table-column label="CPU" width="90">
            <template #default="{ row }">
              <span :class="getCpuClass(row.cpuUsage)">{{ row.cpuUsage != null ? row.cpuUsage + '%' : '-' }}</span>
            </template>
          </el-table-column>
          <el-table-column label="内存" width="90">
            <template #default="{ row }">
              <span :class="getMemClass(row.memUsage)">{{ row.memUsage != null ? row.memUsage + '%' : '-' }}</span>
            </template>
          </el-table-column>
          <el-table-column label="磁盘" width="90">
            <template #default="{ row }">
              <span :class="getDiskClass(row.diskUsage)">{{ row.diskUsage != null ? row.diskUsage + '%' : '-' }}</span>
            </template>
          </el-table-column>
          <el-table-column label="标签" min-width="120">
            <template #default="{ row }">
              <el-tag v-for="tag in (row.tags || [])" :key="tag" size="small" class="mr4">{{ tag }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="processCount" label="进程" width="70" />
          <el-table-column prop="portCount" label="端口" width="70" />
          <el-table-column prop="tunnelCount" label="隧道" width="70" />
          <el-table-column label="状态" width="90">
            <template #default="{ row }">
              <el-tag :type="row.status === 'online' ? 'success' : row.status === 'offline' ? 'danger' : 'info'" size="small">
                {{ { online: '在线', offline: '离线', unknown: '未知' }[row.status] || row.status }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="160" fixed="right">
            <template #default="{ row }">
              <el-button link type="primary" size="small" @click="openHostDialog(row)">编辑</el-button>
              <el-button link type="success" size="small" @click="handleTerminalRow(row)">终端</el-button>
              <el-button link type="danger" size="small" @click="handleDeleteHost(row)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>

        <!-- 分页 -->
        <el-pagination
          class="pagination"
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.pageSize"
          :page-sizes="[10, 20, 50, 100]"
          :total="pagination.total"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="fetchHosts"
          @current-change="fetchHosts"
        />
      </div>
    </div>

    <!-- ================ 右键上下文菜单 ================ -->
    <div
      v-show="ctxMenu.visible"
      class="ctx-menu"
      :style="{ left: ctxMenu.x + 'px', top: ctxMenu.y + 'px' }"
      @mouseleave="hideCtxMenu"
    >
      <div class="ctx-item" @click="openGroupDialog(ctxMenu.node, 'createRoot')">
        <el-icon><Plus /></el-icon> 创建根分组
      </div>
      <div class="ctx-item" @click="openGroupDialog(ctxMenu.node, 'createChild')">
        <el-icon><FolderAdd /></el-icon> 创建子分组
      </div>
      <div class="ctx-divider" />
      <div class="ctx-item" @click="openGroupDialog(ctxMenu.node, 'rename')">
        <el-icon><Edit /></el-icon> 重命名
      </div>
      <div class="ctx-item" @click="openGroupDialog(ctxMenu.node, 'edit')">
        <el-icon><EditPen /></el-icon> 修改分组
      </div>
      <div class="ctx-divider" />
      <div class="ctx-item danger" @click="handleDeleteGroup(ctxMenu.node)">
        <el-icon><Delete /></el-icon> 删除分组
      </div>
    </div>

    <!-- ================ 新增/编辑主机 Dialog ================ -->
    <el-dialog v-model="hostDialogVisible" :title="hostDialogTitle" width="600px" destroy-on-close>
      <el-form ref="hostFormRef" :model="hostForm" :rules="hostFormRules" label-width="90px">
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="主机名称" prop="name">
              <el-input v-model="hostForm.name" placeholder="请输入主机名称" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="IP地址" prop="ip">
              <el-input v-model="hostForm.ip" placeholder="192.168.1.1" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="SSH端口" prop="port">
              <el-input-number v-model="hostForm.port" :min="1" :max="65535" style="width:100%" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="操作系统">
              <el-select v-model="hostForm.osType" placeholder="操作系统" style="width:100%">
                <el-option label="Linux" value="linux" />
                <el-option label="Windows" value="windows" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="云厂商">
              <el-select v-model="hostForm.cloudProvider" placeholder="云厂商" clearable style="width:100%">
                <el-option label="阿里云" value="aliyun" />
                <el-option label="腾讯云" value="tencent" />
                <el-option label="华为云" value="huawei" />
                <el-option label="AWS" value="aws" />
                <el-option label="Azure" value="azure" />
                <el-option label="自建" value="self" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="所属分组">
              <el-tree-select
                v-model="hostForm.groupId"
                :data="groupTree"
                :props="{ label: 'name', value: 'id', children: 'children' }"
                placeholder="选择分组"
                clearable
                style="width:100%"
              />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="登录用户" prop="username">
              <el-input v-model="hostForm.username" placeholder="root" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="认证方式">
              <el-select v-model="hostForm.authType" style="width:100%">
                <el-option label="密码" value="password" />
                <el-option label="密钥" value="key" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="24" v-if="hostForm.authType === 'password'">
            <el-form-item label="登录密码">
              <el-input v-model="hostForm.password" type="password" show-password placeholder="登录密码" />
            </el-form-item>
          </el-col>
          <el-col :span="24" v-else>
            <el-form-item label="私钥内容">
              <el-input v-model="hostForm.privateKey" type="textarea" :rows="4" placeholder="粘贴 SSH 私钥内容" />
            </el-form-item>
          </el-col>
          <el-col :span="24">
            <el-form-item label="标签">
              <el-select v-model="hostForm.tags" multiple allow-create filterable placeholder="输入后回车添加标签" style="width:100%">
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="24">
            <el-form-item label="备注">
              <el-input v-model="hostForm.remark" type="textarea" :rows="2" placeholder="备注" />
            </el-form-item>
          </el-col>
        </el-row>
      </el-form>
      <template #footer>
        <el-button @click="hostDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleHostSubmit">确认</el-button>
      </template>
    </el-dialog>

    <!-- ================ 新增/编辑分组 Dialog ================ -->
    <el-dialog v-model="groupDialogVisible" :title="groupDialogTitle" width="420px" destroy-on-close>
      <el-form ref="groupFormRef" :model="groupForm" :rules="groupFormRules" label-width="90px">
        <el-form-item label="分组名称" prop="name">
          <el-input v-model="groupForm.name" placeholder="请输入分组名称" />
        </el-form-item>
        <el-form-item label="上级分组" v-if="groupForm.action !== 'createRoot'">
          <el-tree-select
            v-model="groupForm.parentId"
            :data="groupTree"
            :props="{ label: 'name', value: 'id', children: 'children' }"
            placeholder="顶级分组（不选则为根）"
            clearable
            style="width:100%"
          />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="groupForm.remark" type="textarea" :rows="2" placeholder="备注" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="groupDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleGroupSubmit">确认</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted, onBeforeUnmount } from 'vue'
import {
  Search, Plus, Refresh, Folder, Monitor, CircleCheck, CircleClose,
  AlarmClock, FolderAdd, Edit, EditPen, Delete
} from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  getHostGroups, getHostList, createHost, updateHost,
  deleteHost, batchDeleteHosts, getHostStats,
  createHostGroup, updateHostGroup, deleteHostGroup
} from '@/api/asset.js'
import { encryptPassword } from '@/utils/crypto.js'

// ===================== 状态 =====================
const loading = ref(false)
const submitting = ref(false)
const groupSearch = ref('')
const groupTree = ref([])
const expandedKeys = ref([])
const selectedGroupId = ref(null)
const selectedHosts = ref([])
const tableData = ref([])
const groupTreeRef = ref(null)

const stats = reactive({ total: 0, linux: 0, windows: 0, offline: 0, huawei: 0, aliyun: 0, aws: 0 })
const pagination = reactive({ page: 1, pageSize: 10, total: 0 })
const searchForm = reactive({ name: '', ip: '', status: '' })

// 右键菜单
const ctxMenu = reactive({ visible: false, x: 0, y: 0, node: null })

// Dialog - 主机
const hostDialogVisible = ref(false)
const hostDialogTitle = ref('新建主机')
const hostFormRef = ref(null)
const hostForm = reactive({
  id: null, name: '', ip: '', port: 22, osType: 'linux', cloudProvider: '',
  groupId: null, username: 'root', authType: 'password', password: '', privateKey: '', tags: [], remark: ''
})
const hostFormRules = {
  name: [{ required: true, message: '请输入主机名称', trigger: 'blur' }],
  ip: [{ required: true, message: '请输入IP地址', trigger: 'blur' }],
  port: [{ required: true, message: '请输入SSH端口' }],
  username: [{ required: true, message: '请输入登录用户名', trigger: 'blur' }]
}

// Dialog - 分组
const groupDialogVisible = ref(false)
const groupDialogTitle = ref('新建分组')
const groupFormRef = ref(null)
const groupForm = reactive({ id: null, name: '', parentId: null, remark: '', action: 'createRoot' })
const groupFormRules = { name: [{ required: true, message: '请输入分组名称', trigger: 'blur' }] }

// ===================== 计算 =====================
const totalGroupCount = computed(() => {
  const countAll = (list) => list.reduce((sum, g) => sum + 1 + countAll(g.children || []), 0)
  return countAll(groupTree.value)
})

const filteredGroups = computed(() => {
  if (!groupSearch.value) return groupTree.value
  const keyword = groupSearch.value.toLowerCase()
  const filter = (list) => list.reduce((acc, g) => {
    const children = filter(g.children || [])
    if (g.name.toLowerCase().includes(keyword) || children.length) {
      acc.push({ ...g, children })
    }
    return acc
  }, [])
  return filter(groupTree.value)
})

// ===================== 生命周期 =====================
onMounted(() => {
  fetchGroups()
  fetchHosts()
  fetchStats()
  document.addEventListener('click', hideCtxMenu)
})
onBeforeUnmount(() => {
  document.removeEventListener('click', hideCtxMenu)
})

// ===================== 分组方法 =====================
async function fetchGroups() {
  try {
    const res = await getHostGroups()
    groupTree.value = res?.data?.data || []
    if (groupTree.value.length && !expandedKeys.value.length) {
      expandedKeys.value = [groupTree.value[0].id]
    }
  } catch (e) {
    ElMessage.error('获取分组列表失败')
  }
}

function handleGroupClick(data) {
  selectedGroupId.value = data.id
  pagination.page = 1
  fetchHosts()
  fetchStats()
}

// 右键菜单
function handleContextMenu(event, data) {
  event.preventDefault()
  event.stopPropagation()
  ctxMenu.visible = true
  ctxMenu.x = event.clientX
  ctxMenu.y = event.clientY
  ctxMenu.node = data
}

function hideCtxMenu() {
  ctxMenu.visible = false
}

// 打开分组 Dialog
function openGroupDialog(node, action) {
  hideCtxMenu()
  Object.assign(groupForm, { id: null, name: '', parentId: null, remark: '', action })

  if (action === 'createRoot') {
    groupDialogTitle.value = '创建根分组'
  } else if (action === 'createChild') {
    groupDialogTitle.value = '创建子分组'
    groupForm.parentId = node?.id || null
  } else if (action === 'rename') {
    groupDialogTitle.value = '重命名分组'
    groupForm.id = node.id
    groupForm.name = node.name
    groupForm.parentId = node.parentId || null
    groupForm.remark = node.remark || ''
  } else if (action === 'edit') {
    groupDialogTitle.value = '修改分组'
    groupForm.id = node.id
    groupForm.name = node.name
    groupForm.parentId = node.parentId || null
    groupForm.remark = node.remark || ''
  }
  groupDialogVisible.value = true
}

async function handleGroupSubmit() {
  await groupFormRef.value?.validate()
  submitting.value = true
  try {
    const payload = {
      name: groupForm.name,
      parentId: groupForm.parentId || 0,
      remark: groupForm.remark
    }
    if (groupForm.id) {
      await updateHostGroup(groupForm.id, payload)
      ElMessage.success('分组更新成功')
    } else {
      await createHostGroup(payload)
      ElMessage.success('分组创建成功')
    }
    groupDialogVisible.value = false
    fetchGroups()
  } catch (e) {
    ElMessage.error(e?.response?.data?.message || e?.message || '操作失败')
  } finally {
    submitting.value = false
  }
}

async function handleDeleteGroup(node) {
  hideCtxMenu()
  if (!node) return
  await ElMessageBox.confirm(`确认删除分组「${node.name}」？删除后分组内主机将移至未分组。`, '提示', { type: 'warning' })
  try {
    await deleteHostGroup(node.id)
    ElMessage.success('删除成功')
    if (selectedGroupId.value === node.id) { selectedGroupId.value = null; fetchHosts() }
    fetchGroups()
  } catch (e) {
    ElMessage.error(e?.response?.data?.message || e?.message || '删除失败')
  }
}

// ===================== 主机方法 =====================
async function fetchHosts() {
  loading.value = true
  try {
    const params = {
      page: pagination.page,
      pageSize: pagination.pageSize,
      groupId: selectedGroupId.value,
      ...searchForm
    }
    Object.keys(params).forEach(k => (params[k] === '' || params[k] == null) && delete params[k])
    const res = await getHostList(params)
    const d = res?.data?.data || res?.data || {}
    tableData.value = d.list || []
    pagination.total = d.total || 0
  } catch (e) {
    tableData.value = []
  } finally {
    loading.value = false
  }
}

async function fetchStats() {
  try {
    const res = await getHostStats(selectedGroupId.value)
    const d = res?.data?.data || {}
    Object.assign(stats, { total: 0, linux: 0, windows: 0, offline: 0, huawei: 0, aliyun: 0, aws: 0, ...d })
  } catch {}
}

function handleSelectionChange(rows) { selectedHosts.value = rows }

function openHostDialog(row = null) {
  Object.assign(hostForm, {
    id: null, name: '', ip: '', port: 22, osType: 'linux', cloudProvider: '',
    groupId: selectedGroupId.value, username: 'root', authType: 'password', password: '', privateKey: '', tags: [], remark: ''
  })
  if (row) {
    Object.assign(hostForm, {
      id: row.id, name: row.name, ip: row.ip, port: row.port || 22,
      osType: row.osType || 'linux', cloudProvider: row.cloudProvider || '',
      groupId: row.groupId || null, username: row.username || 'root',
      authType: row.authType || 'password', tags: row.tags || [], remark: row.remark || ''
    })
    hostDialogTitle.value = '编辑主机'
  } else {
    hostDialogTitle.value = '新建主机'
  }
  hostDialogVisible.value = true
}

async function handleHostSubmit() {
  await hostFormRef.value?.validate()
  submitting.value = true
  try {
    const payload = { ...hostForm }
    if (payload.authType === 'password') {
      if (payload.password) {
        payload.password = encryptPassword(payload.password)
      } else {
        delete payload.password
      }
    }
    if (payload.id) { await updateHost(payload.id, payload) } else { await createHost(payload) }
    ElMessage.success('操作成功')
    hostDialogVisible.value = false
    fetchHosts()
    fetchStats()
  } catch (e) {
    ElMessage.error(e?.response?.data?.message || e?.message || '操作失败')
  } finally {
    submitting.value = false }
}

async function handleDeleteHost(row) {
  await ElMessageBox.confirm(`确认删除主机「${row.name}」？`, '提示', { type: 'warning' })
  try {
    await deleteHost(row.id)
    ElMessage.success('删除成功')
    fetchHosts()
    fetchStats()
  } catch (e) { ElMessage.error(e?.response?.data?.message || e?.message || '删除失败') }
}

async function handleBatchDelete() {
  if (!selectedHosts.value.length) return
  await ElMessageBox.confirm(`确认删除选中的 ${selectedHosts.value.length} 台主机？`, '提示', { type: 'warning' })
  try {
    await batchDeleteHosts(selectedHosts.value.map(h => h.id))
    ElMessage.success('批量删除成功')
    fetchHosts()
    fetchStats()
  } catch (e) { ElMessage.error(e?.response?.data?.message || e?.message || '操作失败') }
}

function handleAlarm() { ElMessage.info('告警功能待接入监控系统') }
function handleTerminal() {
  if (!selectedHosts.value.length) { ElMessage.warning('请先选择主机'); return }
  ElMessage.info('终端功能待实现')
}
function handleTerminalRow(row) { ElMessage.info(`连接终端: ${row.name}(${row.ip})`) }

const getCpuClass = (v) => v > 90 ? 'usage-danger' : v > 70 ? 'usage-warning' : 'usage-ok'
const getMemClass = (v) => v > 90 ? 'usage-danger' : v > 80 ? 'usage-warning' : 'usage-ok'
const getDiskClass = (v) => v > 90 ? 'usage-danger' : v > 80 ? 'usage-warning' : 'usage-ok'
</script>

<style scoped>
.host-management {
  height: 100%;
  display: flex;
  flex-direction: column;
  position: relative;
}

.host-layout {
  display: flex;
  gap: 16px;
  flex: 1;
  min-height: 0;
}

/* ====== 左侧分组面板 ====== */
.group-panel {
  width: 240px;
  flex-shrink: 0;
  background: #fff;
  border-radius: 8px;
  border: 1px solid #e4e7ed;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.group-header {
  padding: 14px 14px 8px;
  border-bottom: 1px solid #f0f0f0;
}

.group-title-row {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 10px;
}

.group-title {
  font-size: 13px;
  font-weight: 600;
  color: #303133;
}

.group-stats {
  display: flex;
  align-items: center;
  gap: 4px;
  background: #f0f6ff;
  border-radius: 6px;
  padding: 4px 8px;
  font-size: 11px;
  color: #409eff;
}

.stat-num { font-size: 18px; font-weight: 700; color: #409eff; line-height: 1; }
.stat-label { font-size: 9px; line-height: 1.3; color: #9bafd0; text-align: center; }
.stat-icon { font-size: 14px; }
.stat-icon.success { color: #67c23a; }

.refresh-btn, .add-btn {
  cursor: pointer;
  color: #909399;
  transition: color 0.2s;
}
.refresh-btn:hover { color: #409eff; }
.add-btn:hover { color: #67c23a; }

.group-search { width: 100%; }

.group-tree-scroll { flex: 1; padding: 8px 4px; }

.tree-node {
  display: flex;
  align-items: center;
  gap: 6px;
  width: 100%;
}

.node-folder { color: #f5a623; font-size: 14px; }
.node-label { flex: 1; font-size: 13px; color: #303133; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.node-count { flex-shrink: 0; }

/* ====== 右键上下文菜单 ====== */
.ctx-menu {
  position: fixed;
  z-index: 9999;
  background: #fff;
  border: 1px solid #e4e7ed;
  border-radius: 8px;
  box-shadow: 0 4px 20px rgba(0,0,0,0.12);
  padding: 4px 0;
  min-width: 160px;
}

.ctx-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  font-size: 13px;
  color: #303133;
  cursor: pointer;
  transition: background 0.15s;
}
.ctx-item:hover { background: #f5f7fa; }
.ctx-item.danger { color: #f56c6c; }
.ctx-item.danger:hover { background: #fef0f0; }

.ctx-divider { height: 1px; background: #f0f0f0; margin: 4px 0; }

/* ====== 右侧主机面板 ====== */
.host-panel {
  flex: 1;
  background: #fff;
  border-radius: 8px;
  border: 1px solid #e4e7ed;
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 12px;
  overflow: hidden;
}

.stats-bar {
  display: flex;
  align-items: center;
  gap: 20px;
  padding-bottom: 12px;
  border-bottom: 1px solid #f0f0f0;
  flex-wrap: wrap;
}

.stat-item { display: flex; align-items: center; gap: 6px; font-size: 13px; color: #606266; }
.stat-item b { font-weight: 700; color: #303133; }

.cloud-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 2px 6px;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 600;
}
.cloud-badge.huawei { background: #cf3030; color: #fff; }
.cloud-badge.aliyun { background: #ff6a00; color: #fff; }
.cloud-badge.aws { background: #232f3e; color: #f90; }

.search-bar { display: flex; align-items: center; gap: 8px; flex-wrap: wrap; }

.usage-ok { color: #67c23a; }
.usage-warning { color: #e6a23c; font-weight: 600; }
.usage-danger { color: #f56c6c; font-weight: 600; }

.mr4 { margin-right: 4px; }

.pagination { margin-top: auto; display: flex; justify-content: flex-end; }
</style>
