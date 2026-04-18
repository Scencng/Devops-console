<template>
  <div class="page-container">
    <el-card class="page-header-card">
      <div class="page-header">
        <div><h2>数据血缘</h2><p>维护 Topic 与上下游服务的依赖关系，形成消息流转的影响面视图</p></div>
        <div class="header-actions">
          <el-select v-model="selectedClusterId" style="width:260px" @change="loadData"><el-option v-for="cluster in clusters" :key="cluster.id" :label="cluster.name" :value="cluster.id" /></el-select>
          <el-input v-model="keyword" style="width:220px" clearable placeholder="搜索 topic/service" @keyup.enter="loadData" />
          <el-button @click="loadData" :loading="loading">刷新</el-button>
          <el-button type="primary" @click="openDialog()">新增关系</el-button>
        </div>
      </div>
    </el-card>
    <el-card class="content-card" v-loading="loading">
      <el-table :data="rows" empty-text="暂无血缘关系">
        <el-table-column prop="sourceTopic" label="Source Topic" min-width="180" />
        <el-table-column prop="targetTopic" label="Target Topic" min-width="180" />
        <el-table-column prop="relationType" label="关系类型" width="140" />
        <el-table-column prop="producerService" label="Producer" min-width="160" />
        <el-table-column prop="consumerService" label="Consumer" min-width="160" />
        <el-table-column label="操作" width="160" fixed="right"><template #default="{ row }"><el-button link type="primary" @click="openDialog(row)">编辑</el-button><el-button link type="danger" @click="remove(row)">删除</el-button></template></el-table-column>
      </el-table>
    </el-card>
    <el-dialog v-model="dialogVisible" :title="editing ? '编辑血缘关系' : '新增血缘关系'" width="700px" destroy-on-close>
      <el-form label-position="top" :model="form">
        <el-row :gutter="16">
          <el-col :span="12"><el-form-item label="Source Topic"><el-input v-model="form.sourceTopic" /></el-form-item></el-col>
          <el-col :span="12"><el-form-item label="Target Topic"><el-input v-model="form.targetTopic" /></el-form-item></el-col>
          <el-col :span="12"><el-form-item label="关系类型"><el-input v-model="form.relationType" placeholder="produce-consume / transform / replicate" /></el-form-item></el-col>
          <el-col :span="12"><el-form-item label="Producer Service"><el-input v-model="form.producerService" /></el-form-item></el-col>
        </el-row>
        <el-form-item label="Consumer Service"><el-input v-model="form.consumerService" /></el-form-item>
        <el-form-item label="描述"><el-input v-model="form.description" type="textarea" :rows="4" /></el-form-item>
      </el-form>
      <template #footer><el-button @click="dialogVisible = false">取消</el-button><el-button type="primary" :loading="saving" @click="save">保存</el-button></template>
    </el-dialog>
  </div>
</template>

<script setup>
import { onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { createKafkaLineageRelation, deleteKafkaLineageRelation, getKafkaClusterOptions, getKafkaLineageRelations, updateKafkaLineageRelation } from '@/api/kafka.js'

const clusters = ref([]); const selectedClusterId = ref(null); const keyword = ref(''); const loading = ref(false); const saving = ref(false); const rows = ref([]); const dialogVisible = ref(false); const editing = ref(null)
const form = reactive({ sourceTopic: '', targetTopic: '', relationType: '', producerService: '', consumerService: '', description: '' })
const loadClusters = async () => { const res = await getKafkaClusterOptions(); clusters.value = res?.data?.data || []; if (!selectedClusterId.value && clusters.value.length > 0) selectedClusterId.value = clusters.value[0].id }
const loadData = async () => { if (!selectedClusterId.value) return; loading.value = true; try { const res = await getKafkaLineageRelations({ clusterId: selectedClusterId.value, keyword: keyword.value }); rows.value = res?.data?.data || [] } catch (error) { ElMessage.error(error.message || '血缘关系加载失败') } finally { loading.value = false } }
const openDialog = (row = null) => { editing.value = row; Object.assign(form, row || { sourceTopic: '', targetTopic: '', relationType: '', producerService: '', consumerService: '', description: '' }); dialogVisible.value = true }
const save = async () => { if (!selectedClusterId.value || !form.sourceTopic || !form.targetTopic || !form.relationType) { ElMessage.warning('请填写完整血缘关系'); return } saving.value = true; try { const payload = { ...form, clusterId: selectedClusterId.value }; if (editing.value?.id) await updateKafkaLineageRelation(editing.value.id, payload); else await createKafkaLineageRelation(payload); ElMessage.success('血缘关系已保存'); dialogVisible.value = false; await loadData() } catch (error) { ElMessage.error(error.message || '血缘关系保存失败') } finally { saving.value = false } }
const remove = async (row) => { await ElMessageBox.confirm('确认删除该血缘关系吗？', '提示', { type: 'warning' }); try { await deleteKafkaLineageRelation(row.id); ElMessage.success('血缘关系已删除'); await loadData() } catch (error) { ElMessage.error(error.message || '血缘关系删除失败') } }
onMounted(async () => { await loadClusters(); await loadData() })
</script>
