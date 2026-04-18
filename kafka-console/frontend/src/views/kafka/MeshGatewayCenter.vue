<template>
  <div class="page-container">
    <el-card class="page-header-card">
      <div class="page-header">
        <div><h2>网关 / Mesh 集成</h2><p>维护 Kafka Console 对接的网关或 Service Mesh 配置，用于统一认证和流量治理对接</p></div>
        <div class="header-actions">
          <el-select v-model="selectedClusterId" style="width:260px" @change="loadData"><el-option v-for="cluster in clusters" :key="cluster.id" :label="cluster.name" :value="cluster.id" /></el-select>
          <el-button type="primary" @click="openDialog()">新增配置</el-button>
        </div>
      </div>
    </el-card>
    <el-card class="content-card" v-loading="loading">
      <el-table :data="rows" empty-text="暂无网关配置">
        <el-table-column prop="gatewayType" label="类型" width="160" />
        <el-table-column prop="endpoint" label="Endpoint" min-width="240" />
        <el-table-column prop="authMode" label="认证模式" width="140" />
        <el-table-column prop="enabled" label="启用" width="100"><template #default="{ row }"><el-tag :type="row.enabled ? 'success' : 'info'">{{ row.enabled ? '是' : '否' }}</el-tag></template></el-table-column>
        <el-table-column label="操作" width="160" fixed="right"><template #default="{ row }"><el-button link type="primary" @click="openDialog(row)">编辑</el-button><el-button link type="danger" @click="remove(row)">删除</el-button></template></el-table-column>
      </el-table>
    </el-card>
    <el-dialog v-model="dialogVisible" :title="editing ? '编辑网关配置' : '新增网关配置'" width="680px" destroy-on-close>
      <el-form label-position="top" :model="form">
        <el-row :gutter="16">
          <el-col :span="12"><el-form-item label="类型"><el-input v-model="form.gatewayType" placeholder="istio / apisix / nginx / traefik" /></el-form-item></el-col>
          <el-col :span="12"><el-form-item label="认证模式"><el-input v-model="form.authMode" placeholder="oidc / jwt / mtls" /></el-form-item></el-col>
        </el-row>
        <el-form-item label="Endpoint"><el-input v-model="form.endpoint" /></el-form-item>
        <el-form-item label="配置(JSON/YAML)"><el-input v-model="form.config" type="textarea" :rows="8" /></el-form-item>
        <el-form-item label="启用"><el-switch v-model="form.enabled" /></el-form-item>
      </el-form>
      <template #footer><el-button @click="dialogVisible = false">取消</el-button><el-button type="primary" :loading="saving" @click="save">保存</el-button></template>
    </el-dialog>
  </div>
</template>

<script setup>
import { onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { createKafkaMeshGatewayConfig, deleteKafkaMeshGatewayConfig, getKafkaClusterOptions, getKafkaMeshGatewayConfigs, updateKafkaMeshGatewayConfig } from '@/api/kafka.js'

const clusters = ref([]); const selectedClusterId = ref(null); const loading = ref(false); const saving = ref(false); const rows = ref([]); const dialogVisible = ref(false); const editing = ref(null)
const form = reactive({ gatewayType: '', endpoint: '', authMode: '', config: '{}', enabled: true })
const loadClusters = async () => { const res = await getKafkaClusterOptions(); clusters.value = res?.data?.data || []; if (!selectedClusterId.value && clusters.value.length > 0) selectedClusterId.value = clusters.value[0].id }
const loadData = async () => { if (!selectedClusterId.value) return; loading.value = true; try { const res = await getKafkaMeshGatewayConfigs({ clusterId: selectedClusterId.value }); rows.value = res?.data?.data || [] } catch (error) { ElMessage.error(error.message || '网关配置加载失败') } finally { loading.value = false } }
const openDialog = (row = null) => { editing.value = row; Object.assign(form, row || { gatewayType: '', endpoint: '', authMode: '', config: '{}', enabled: true }); dialogVisible.value = true }
const save = async () => { if (!selectedClusterId.value || !form.gatewayType) { ElMessage.warning('请填写类型'); return } saving.value = true; try { const payload = { ...form, clusterId: selectedClusterId.value }; if (editing.value?.id) await updateKafkaMeshGatewayConfig(editing.value.id, payload); else await createKafkaMeshGatewayConfig(payload); ElMessage.success('网关配置已保存'); dialogVisible.value = false; await loadData() } catch (error) { ElMessage.error(error.message || '网关配置保存失败') } finally { saving.value = false } }
const remove = async (row) => { await ElMessageBox.confirm('确认删除该网关配置吗？', '提示', { type: 'warning' }); try { await deleteKafkaMeshGatewayConfig(row.id); ElMessage.success('网关配置已删除'); await loadData() } catch (error) { ElMessage.error(error.message || '网关配置删除失败') } }
onMounted(async () => { await loadClusters(); await loadData() })
</script>
