<template>
  <div class="page-container">
    <el-card class="page-header-card">
      <div class="page-header">
        <div><h2>云托管适配</h2><p>管理阿里云、腾讯云、AWS MSK、Confluent Cloud 等托管 Kafka 的接入适配信息</p></div>
        <div class="header-actions">
          <el-select v-model="selectedClusterId" style="width:260px" @change="loadData"><el-option v-for="cluster in clusters" :key="cluster.id" :label="cluster.name" :value="cluster.id" /></el-select>
          <el-button type="primary" @click="openDialog()">新增适配</el-button>
        </div>
      </div>
    </el-card>
    <el-card class="content-card" v-loading="loading">
      <el-table :data="rows" empty-text="暂无云适配配置">
        <el-table-column prop="provider" label="云厂商" width="140" />
        <el-table-column prop="serviceName" label="服务" width="160" />
        <el-table-column prop="region" label="Region" width="120" />
        <el-table-column prop="clusterIdentifier" label="Cluster Identifier" min-width="200" />
        <el-table-column prop="endpointMode" label="Endpoint Mode" width="140" />
        <el-table-column label="操作" width="160" fixed="right">
          <template #default="{ row }"><el-button link type="primary" @click="openDialog(row)">编辑</el-button><el-button link type="danger" @click="remove(row)">删除</el-button></template>
        </el-table-column>
      </el-table>
    </el-card>
    <el-dialog v-model="dialogVisible" :title="editing ? '编辑云适配' : '新增云适配'" width="680px" destroy-on-close>
      <el-form label-position="top" :model="form">
        <el-row :gutter="16">
          <el-col :span="12"><el-form-item label="云厂商"><el-input v-model="form.provider" placeholder="aliyun/tencent/aws/confluent" /></el-form-item></el-col>
          <el-col :span="12"><el-form-item label="服务名"><el-input v-model="form.serviceName" placeholder="MSK / Confluent Cloud / ApsaraMQ" /></el-form-item></el-col>
          <el-col :span="12"><el-form-item label="Region"><el-input v-model="form.region" /></el-form-item></el-col>
          <el-col :span="12"><el-form-item label="Endpoint Mode"><el-input v-model="form.endpointMode" placeholder="public/private/vpc" /></el-form-item></el-col>
        </el-row>
        <el-form-item label="Cluster Identifier"><el-input v-model="form.clusterIdentifier" /></el-form-item>
        <el-form-item label="说明"><el-input v-model="form.notes" type="textarea" :rows="4" /></el-form-item>
      </el-form>
      <template #footer><el-button @click="dialogVisible = false">取消</el-button><el-button type="primary" :loading="saving" @click="save">保存</el-button></template>
    </el-dialog>
  </div>
</template>

<script setup>
import { onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { createKafkaCloudAdapter, deleteKafkaCloudAdapter, getKafkaCloudAdapters, getKafkaClusterOptions, updateKafkaCloudAdapter } from '@/api/kafka.js'

const clusters = ref([]); const selectedClusterId = ref(null); const loading = ref(false); const saving = ref(false); const rows = ref([]); const dialogVisible = ref(false); const editing = ref(null)
const form = reactive({ provider: '', serviceName: '', region: '', clusterIdentifier: '', endpointMode: '', notes: '' })
const loadClusters = async () => { const res = await getKafkaClusterOptions(); clusters.value = res?.data?.data || []; if (!selectedClusterId.value && clusters.value.length > 0) selectedClusterId.value = clusters.value[0].id }
const loadData = async () => { if (!selectedClusterId.value) return; loading.value = true; try { const res = await getKafkaCloudAdapters({ clusterId: selectedClusterId.value }); rows.value = res?.data?.data || [] } catch (error) { ElMessage.error(error.message || '云适配加载失败') } finally { loading.value = false } }
const openDialog = (row = null) => { editing.value = row; Object.assign(form, row || { provider: '', serviceName: '', region: '', clusterIdentifier: '', endpointMode: '', notes: '' }); dialogVisible.value = true }
const save = async () => { if (!selectedClusterId.value || !form.provider || !form.serviceName) { ElMessage.warning('请填写云厂商和服务名'); return } saving.value = true; try { const payload = { ...form, clusterId: selectedClusterId.value }; if (editing.value?.id) await updateKafkaCloudAdapter(editing.value.id, payload); else await createKafkaCloudAdapter(payload); ElMessage.success('云适配已保存'); dialogVisible.value = false; await loadData() } catch (error) { ElMessage.error(error.message || '云适配保存失败') } finally { saving.value = false } }
const remove = async (row) => { await ElMessageBox.confirm(`确认删除云适配 ${row.provider}/${row.serviceName} 吗？`, '提示', { type: 'warning' }); try { await deleteKafkaCloudAdapter(row.id); ElMessage.success('云适配已删除'); await loadData() } catch (error) { ElMessage.error(error.message || '云适配删除失败') } }
onMounted(async () => { await loadClusters(); await loadData() })
</script>
