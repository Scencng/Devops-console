<template>
  <div class="page-container">
    <el-card shadow="never">
      <div class="toolbar">
        <span class="title">菜单管理</span>
        <div class="actions">
          <el-button type="success" :icon="Refresh" @click="handleRefreshCache">刷新缓存</el-button>
          <el-button type="warning" :icon="Sort" @click="toggleExpandAll">展开/折叠</el-button>
          <el-button type="primary" :icon="Plus" @click="openDialog()">新增菜单</el-button>
        </div>
      </div>
      <el-table ref="menuTable" :data="tableData" v-loading="loading" row-key="id" border v-bind="defaultExpandAllProp"
        :tree-props="{ children: 'children' }">
        <el-table-column prop="name" label="菜单名称" min-width="180" />
        <el-table-column label="类型" width="90">
          <template #default="{ row }">
            <el-tag size="small" :type="row.type === 1 ? '' : row.type === 2 ? 'success' : 'warning'">
              {{ ['', '目录', '菜单', '按钮'][row.type] }}
            </el-tag>
          </template>
        </el-table-column>
        <!-- 图标列：即显示图标图形，也显示名称 -->
        <el-table-column label="图标" width="100">
          <template #default="{ row }">
            <div v-if="row.icon" class="icon-cell">
              <el-icon :size="16"><component :is="row.icon" /></el-icon>
              <span class="icon-name-text">{{ row.icon }}</span>
            </div>
            <span v-else class="text-muted">-</span>
          </template>
        </el-table-column>
        <el-table-column prop="path" label="路由路径" min-width="160" />
        <el-table-column prop="component" label="组件路径" min-width="200" />
        <el-table-column prop="perm" label="权限标识" min-width="200" />
        <el-table-column prop="sort" label="排序" width="70" />
        <el-table-column label="可见" width="80">
          <template #default="{ row }">
            <el-tag size="small" :type="row.visible === 1 ? 'success' : 'danger'">{{ row.visible === 1 ? '是' : '否' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="80">
          <template #default="{ row }">
            <el-tag size="small" :type="row.status === 1 ? 'success' : 'danger'">{{ row.status === 1 ? '启用' : '停用' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="170" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="openDialog(null, row.id)">新增菜单</el-button>
            <el-button link type="primary" @click="openDialog(row)">编辑</el-button>
            <el-button link type="danger" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- ============ 新增/编辑 Dialog ============ -->
    <el-dialog v-model="dialogVisible" :title="dialogTitle" width="600px" destroy-on-close>
      <el-form ref="formRef" :model="form" :rules="formRules" label-width="90px">
        <el-form-item label="上级菜单">
          <el-tree-select v-model="form.parentId" :data="treeData"
            :props="{ label: 'name', value: 'id', children: 'children' }"
            check-strictly
            placeholder="顶级菜单" clearable style="width:100%" />
        </el-form-item>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="菜单类型" prop="type">
              <el-radio-group v-model="form.type">
                <el-radio :value="1">目录</el-radio>
                <el-radio :value="2">菜单</el-radio>
                <el-radio :value="3">按钮</el-radio>
              </el-radio-group>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="菜单名称" prop="name">
              <el-input v-model="form.name" placeholder="菜单/功能名称" />
            </el-form-item>
          </el-col>
          <!-- 图标选择器 -->
          <el-col :span="24" v-if="form.type !== 3">
            <el-form-item label="图标">
              <div class="icon-picker-trigger" @click="iconPickerVisible = true">
                <div class="icon-preview" v-if="form.icon">
                  <el-icon :size="18"><component :is="form.icon" /></el-icon>
                  <span class="icon-name-badge">{{ form.icon }}</span>
                </div>
                <span class="icon-placeholder" v-else>点击选择图标</span>
                <el-button size="small" link @click.stop="form.icon = ''">
                  <el-icon v-if="form.icon"><Close /></el-icon>
                </el-button>
              </div>
            </el-form-item>
          </el-col>
          <el-col :span="12" v-if="form.type !== 3">
            <el-form-item label="路由路径">
              <el-input v-model="form.path" placeholder="/system/users" />
            </el-form-item>
          </el-col>
          <el-col :span="12" v-if="form.type === 2">
            <el-form-item label="组件路径">
              <el-input v-model="form.component" placeholder="system/UserManagement" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="权限标识">
              <el-input v-model="form.perm" placeholder="system:user:list" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="排序">
              <el-input-number v-model="form.sort" :min="0" style="width:100%" />
            </el-form-item>
          </el-col>
          <el-col :span="12" v-if="form.type !== 3">
            <el-form-item label="是否可见">
              <el-radio-group v-model="form.visible">
                <el-radio :value="1">显示</el-radio>
                <el-radio :value="0">隐藏</el-radio>
              </el-radio-group>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="状态">
              <el-radio-group v-model="form.status">
                <el-radio :value="1">启用</el-radio>
                <el-radio :value="0">停用</el-radio>
              </el-radio-group>
            </el-form-item>
          </el-col>
        </el-row>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleSubmit">确认</el-button>
      </template>
    </el-dialog>

    <!-- ============ 图标选择器弹窗 ============ -->
    <el-dialog v-model="iconPickerVisible" title="选择图标" width="700px" destroy-on-close>
      <el-input
        v-model="iconSearch"
        placeholder="搜索图标名称，如: User / Home / Setting..."
        clearable
        :prefix-icon="Search"
        style="margin-bottom: 12px"
      />
      <el-scrollbar height="420px">
        <div class="icon-grid">
          <div
            v-for="name in filteredIcons"
            :key="name"
            class="icon-grid-item"
            :class="{ active: form.icon === name }"
            @click="selectIcon(name)"
            :title="name"
          >
            <el-icon :size="22"><component :is="name" /></el-icon>
            <span class="icon-grid-label">{{ name }}</span>
          </div>
        </div>
      </el-scrollbar>
      <template #footer>
        <span class="icon-picker-footer">
          当前选中：<b>{{ form.icon || '未选择' }}</b>
        </span>
        <el-button @click="iconPickerVisible = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { Plus, Refresh, Sort, Search, Close } from '@element-plus/icons-vue'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useRouter } from 'vue-router'
import { getMenuTree, createMenu, updateMenu, deleteMenu } from '@/api/system/menu.js'
import { usePermissionStore } from '@/stores/permissionStore.js'

const router = useRouter()
const permStore = usePermissionStore()
const loading = ref(false), submitting = ref(false)
const tableData = ref([])
const treeData = ref([])
const menuTable = ref(null)
const isExpandedAll = ref(true)
const defaultExpandAllProp = reactive({ defaultExpandAll: true })

const dialogVisible = ref(false), dialogTitle = ref('新增菜单')
const formRef = ref(null)
const form = reactive({ id: null, parentId: null, name: '', type: 2, path: '', component: '', icon: '', perm: '', sort: 0, visible: 1, status: 1 })
const formRules = {
  name: [{ required: true, message: '请输入菜单名称', trigger: 'blur' }],
  type: [{ required: true }]
}

// ======== 图标选择器 ========
const iconPickerVisible = ref(false)
const iconSearch = ref('')

// 获取所有图标名（排除工具函数）
const allIconNames = Object.keys(ElementPlusIconsVue).filter(
  key => typeof ElementPlusIconsVue[key] === 'object' && key !== 'default'
)

const filteredIcons = computed(() => {
  if (!iconSearch.value) return allIconNames
  return allIconNames.filter(n => n.toLowerCase().includes(iconSearch.value.toLowerCase()))
})

function selectIcon(name) {
  form.icon = name
  iconPickerVisible.value = false
}

// ======== 列表 ========
onMounted(fetchList)

async function fetchList() {
  loading.value = true
  try {
    const res = await getMenuTree()
    const data = res?.data?.data || []
    tableData.value = data
    treeData.value = data
  } finally { loading.value = false }
}

function openDialog(row = null, parentId = null) {
  Object.assign(form, { id: null, parentId: parentId || null, name: '', type: 2, path: '', component: '', icon: '', perm: '', sort: 0, visible: 1, status: 1 })
  if (row) {
    Object.assign(form, { id: row.id, parentId: row.parentId || null, name: row.name, type: row.type, path: row.path || '', component: row.component || '', icon: row.icon || '', perm: row.perm || '', sort: row.sort, visible: row.visible, status: row.status })
    dialogTitle.value = '编辑菜单'
  } else { dialogTitle.value = parentId ? '新增子菜单' : '新增菜单' }
  dialogVisible.value = true
}

async function handleSubmit() {
  await formRef.value?.validate()
  submitting.value = true
  try {
    const payload = { ...form, parentId: form.parentId || 0 }
    if (payload.id) { await updateMenu(payload.id, payload) } else { await createMenu(payload) }
    ElMessage.success('操作成功')
    dialogVisible.value = false
    fetchList()
  } catch (e) { ElMessage.error(e?.message || '操作失败') } finally { submitting.value = false }
}

async function handleDelete(row) {
  await ElMessageBox.confirm(`确认删除菜单「${row.name}」？有子菜单将无法删除。`, '提示', { type: 'warning' })
  try { await deleteMenu(row.id); ElMessage.success('删除成功'); fetchList() }
  catch (e) { ElMessage.error(e?.response?.data?.message || '删除失败，可能存在子菜单') }
}

async function handleRefreshCache() {
  loading.value = true
  try {
    permStore.isLoaded = false
    await permStore.loadUserAndRoutes(router)
    ElMessage.success('菜单缓存与动态路由已刷新')
  } catch { ElMessage.error('刷新失败') } finally { loading.value = false }
}

function toggleExpandAll() {
  isExpandedAll.value = !isExpandedAll.value
  const traverse = (list) => {
    list.forEach(item => {
      menuTable.value?.toggleRowExpansion(item, isExpandedAll.value)
      if (item.children?.length) traverse(item.children)
    })
  }
  traverse(tableData.value)
}
</script>

<style scoped>
.page-container { padding: 16px; }
.toolbar { display: flex; justify-content: space-between; align-items: center; margin-bottom: 12px; }
.actions { display: flex; gap: 8px; }
.title { font-size: 15px; font-weight: 600; }

/* 表格图标列 */
.icon-cell {
  display: flex;
  align-items: center;
  gap: 6px;
}
.icon-name-text {
  font-size: 11px;
  color: #909399;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 60px;
}
.text-muted { color: #c0c4cc; }

/* 图标选择触发器 */
.icon-picker-trigger {
  display: flex;
  align-items: center;
  gap: 8px;
  width: 100%;
  padding: 0 12px;
  height: 32px;
  border: 1px solid #dcdfe6;
  border-radius: 6px;
  cursor: pointer;
  transition: border-color 0.2s;
  background: #fff;
}
.icon-picker-trigger:hover { border-color: #409eff; }

.icon-preview {
  display: flex;
  align-items: center;
  gap: 6px;
  flex: 1;
}
.icon-name-badge {
  font-size: 13px;
  color: #303133;
}
.icon-placeholder {
  flex: 1;
  font-size: 13px;
  color: #c0c4cc;
}

/* 图标选择器网格 */
.icon-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(90px, 1fr));
  gap: 8px;
  padding: 4px;
}

.icon-grid-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
  padding: 10px 4px;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.15s;
  border: 2px solid transparent;
}
.icon-grid-item:hover {
  background: #f0f6ff;
  border-color: #c6e0ff;
}
.icon-grid-item.active {
  background: #ecf5ff;
  border-color: #409eff;
}

.icon-grid-label {
  font-size: 10px;
  color: #606266;
  text-align: center;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  width: 100%;
  max-width: 82px;
}

.icon-picker-footer {
  margin-right: auto;
  font-size: 13px;
  color: #606266;
}
.icon-picker-footer b { color: #409eff; }
</style>
