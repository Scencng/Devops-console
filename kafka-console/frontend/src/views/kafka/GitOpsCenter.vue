<template>
  <div class="page-container">
    <el-card class="page-header-card">
      <div class="page-header">
        <div><h2>GitOps 中心</h2><p>维护 Kafka GitOps 配置，生成主题元数据等声明式清单并记录同步结果</p></div>
        <div class="header-actions">
          <el-select v-model="selectedClusterId" style="width:260px" @change="refreshAll"><el-option v-for="cluster in clusters" :key="cluster.id" :label="cluster.name" :value="cluster.id" /></el-select>
          <el-button type="primary" @click="openDialog()">新增 GitOps 配置</el-button>
        </div>
      </div>
    </el-card>
    <el-row :gutter="16">
      <el-col :span="12">
        <el-card class="content-card" v-loading="loading">
          <template #header><div class="flex-between"><div>GitOps 配置</div><el-button @click="loadProfiles">刷新</el-button></div></template>
          <el-table :data="profiles" empty-text="暂无 GitOps 配置">
            <el-table-column prop="name" label="名称" min-width="160" />
            <el-table-column prop="repoUrl" label="仓库" min-width="220" show-overflow-tooltip />
            <el-table-column prop="branch" label="分支" width="120" />
            <el-table-column prop="lastSyncStatus" label="最近同步" width="120" />
            <el-table-column label="操作" width="220" fixed="right">
              <template #default="{ row }">
                <el-button link type="primary" @click="runSync(row)">同步</el-button>
                <el-button link type="primary" @click="openDialog(row)">编辑</el-button>
                <el-button link type="primary" @click="viewSyncs(row)">记录</el-button>
                <el-button link type="danger" @click="remove(row)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>
      <el-col :span="12">
        <el-card class="content-card">
          <template #header><div>同步记录</div></template>
          <el-table :data="syncs" v-loading="syncLoading" empty-text="请选择左侧配置查看同步记录">
            <el-table-column prop="status" label="状态" width="120" />
            <el-table-column prop="summary" label="摘要" min-width="220" show-overflow-tooltip />
            <el-table-column prop="commitSha" label="Commit / 批次号" min-width="180" />
            <el-table-column label="输出" width="100">
              <template #default="{ row }"><el-button link type="primary" @click="showOutput(row.output)">查看</el-button></template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>
    </el-row>
    <el-dialog v-model="dialogVisible" :title="editing ? '编辑 GitOps 配置' : '新增 GitOps 配置'" width="700px" destroy-on-close>
      <el-form label-position="top" :model="form">
        <el-row :gutter="16">
          <el-col :span="12"><el-form-item label="名称"><el-input v-model="form.name" /></el-form-item></el-col>
          <el-col :span="12"><el-form-item label="仓库地址"><el-input v-model="form.repoUrl" /></el-form-item></el-col>
          <el-col :span="8"><el-form-item label="分支"><el-input v-model="form.branch" /></el-form-item></el-col>
          <el-col :span="8"><el-form-item label="基础路径"><el-input v-model="form.basePath" /></el-form-item></el-col>
          <el-col :span="8"><el-form-item label="格式"><el-input v-model="form.manifestFormat" placeholder="json/yaml" /></el-form-item></el-col>
        </el-row>
        <el-row :gutter="16">
          <el-col :span="12"><el-form-item label="认证方式"><el-select v-model="form.authType" style="width:100%"><el-option label="无认证" value="none" /><el-option label="Token" value="token" /></el-select></el-form-item></el-col>
          <el-col :span="12"><el-form-item label="Token"><el-input v-model="form.token" type="password" show-password /></el-form-item></el-col>
        </el-row>
        <el-form-item label="启用"><el-switch v-model="form.enabled" /></el-form-item>
      </el-form>
      <template #footer><el-button @click="dialogVisible = false">取消</el-button><el-button type="primary" :loading="saving" @click="save">保存</el-button></template>
    </el-dialog>
    <el-dialog v-model="outputVisible" title="同步输出" width="760px" destroy-on-close><pre class="output-pre">{{ activeOutput }}</pre></el-dialog>
  </div>
</template>

<script setup>
import { onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { createKafkaGitOpsProfile, deleteKafkaGitOpsProfile, getKafkaClusterOptions, getKafkaGitOpsProfiles, getKafkaGitOpsSyncs, runKafkaGitOpsSync, updateKafkaGitOpsProfile } from '@/api/kafka.js'

const clusters = ref([])
const selectedClusterId = ref(null)
const loading = ref(false)
const saving = ref(false)
const syncLoading = ref(false)
const profiles = ref([])
const syncs = ref([])
const dialogVisible = ref(false)
const outputVisible = ref(false)
const activeOutput = ref('')
const editing = ref(null)
const form = reactive({ name: '', repoUrl: '', branch: 'main', basePath: '', manifestFormat: 'json', authType: 'none', token: '', enabled: true })

const loadClusters = async () => { const res = await getKafkaClusterOptions(); clusters.value = res?.data?.data || []; if (!selectedClusterId.value && clusters.value.length > 0) selectedClusterId.value = clusters.value[0].id }
const loadProfiles = async () => { if (!selectedClusterId.value) return; const res = await getKafkaGitOpsProfiles({ clusterId: selectedClusterId.value }); profiles.value = res?.data?.data || [] }
const refreshAll = async () => { loading.value = true; try { await loadProfiles(); syncs.value = [] } catch (error) { ElMessage.error(error.message || 'GitOps 中心加载失败') } finally { loading.value = false } }
const openDialog = (row = null) => { editing.value = row; Object.assign(form, row || { name: '', repoUrl: '', branch: 'main', basePath: '', manifestFormat: 'json', authType: 'none', token: '', enabled: true }); form.token = ''; dialogVisible.value = true }
const save = async () => {
  if (!selectedClusterId.value || !form.name || !form.repoUrl) { ElMessage.warning('请填写名称和仓库地址'); return }
  saving.value = true
  try {
    const payload = { ...form, clusterId: selectedClusterId.value }
    if (editing.value?.id) await updateKafkaGitOpsProfile(editing.value.id, payload)
    else await createKafkaGitOpsProfile(payload)
    ElMessage.success('GitOps 配置已保存')
    dialogVisible.value = false
    await loadProfiles()
  } catch (error) { ElMessage.error(error.message || 'GitOps 配置保存失败') } finally { saving.value = false }
}
const runSync = async (row) => { try { await runKafkaGitOpsSync(row.id); ElMessage.success('同步完成'); await viewSyncs(row) } catch (error) { ElMessage.error(error.message || '同步失败') } }
const viewSyncs = async (row) => { syncLoading.value = true; try { const res = await getKafkaGitOpsSyncs({ profileId: row.id }); syncs.value = res?.data?.data || [] } catch (error) { ElMessage.error(error.message || '同步记录加载失败') } finally { syncLoading.value = false } }
const remove = async (row) => { await ElMessageBox.confirm(`确认删除 GitOps 配置 ${row.name} 吗？`, '提示', { type: 'warning' }); try { await deleteKafkaGitOpsProfile(row.id); ElMessage.success('GitOps 配置已删除'); await loadProfiles() } catch (error) { ElMessage.error(error.message || 'GitOps 配置删除失败') } }
const showOutput = (value) => { activeOutput.value = value || ''; outputVisible.value = true }
onMounted(async () => { await loadClusters(); await refreshAll() })
</script>

<style scoped>.flex-between{display:flex;align-items:center;justify-content:space-between;}.output-pre{padding:12px;overflow:auto;white-space:pre-wrap;word-break:break-word;background:#0f172a;border-radius:10px;color:#e2e8f0;}</style>
