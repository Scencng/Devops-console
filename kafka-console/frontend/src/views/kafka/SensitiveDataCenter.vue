<template>
  <div class="page-container">
    <el-card class="page-header-card">
      <div class="page-header">
        <div><h2>敏感数据识别</h2><p>维护敏感规则，并对 Kafka 消息样本执行扫描，识别手机号、身份证、Token 等敏感内容</p></div>
        <div class="header-actions">
          <el-select v-model="selectedClusterId" style="width:260px" @change="refreshAll"><el-option v-for="cluster in clusters" :key="cluster.id" :label="cluster.name" :value="cluster.id" /></el-select>
          <el-button type="primary" @click="openRuleDialog()">新增规则</el-button>
        </div>
      </div>
    </el-card>
    <el-row :gutter="16">
      <el-col :span="11">
        <el-card class="content-card" v-loading="loading">
          <template #header><div class="flex-between"><div>扫描规则</div><el-button @click="loadRules">刷新</el-button></div></template>
          <el-table :data="rules" empty-text="暂无扫描规则">
            <el-table-column prop="name" label="名称" min-width="160" />
            <el-table-column prop="patternType" label="类型" width="120" />
            <el-table-column prop="severity" label="级别" width="120" />
            <el-table-column label="操作" width="160" fixed="right"><template #default="{ row }"><el-button link type="primary" @click="openRuleDialog(row)">编辑</el-button><el-button link type="danger" @click="removeRule(row)">删除</el-button></template></el-table-column>
          </el-table>
        </el-card>
      </el-col>
      <el-col :span="13">
        <el-card class="content-card" v-loading="loading">
          <template #header>
            <div class="flex-between">
              <div>扫描结果</div>
              <div class="scan-actions">
                <el-input v-model="scanForm.topic" placeholder="Topic" style="width:180px" />
                <el-input-number v-model="scanForm.partition" :min="0" style="width:120px" />
                <el-input-number v-model="scanForm.limit" :min="1" :max="200" style="width:120px" />
                <el-button type="primary" :loading="scanning" @click="runScan">执行扫描</el-button>
              </div>
            </div>
          </template>
          <el-table :data="results" empty-text="暂无扫描结果">
            <el-table-column prop="topic" label="Topic" min-width="160" />
            <el-table-column prop="partition" label="Partition" width="100" />
            <el-table-column prop="offset" label="Offset" width="110" />
            <el-table-column prop="ruleName" label="规则" min-width="140" />
            <el-table-column prop="severity" label="级别" width="100" />
            <el-table-column prop="matchedText" label="命中内容" min-width="180" show-overflow-tooltip />
            <el-table-column prop="summary" label="摘要" min-width="180" show-overflow-tooltip />
          </el-table>
        </el-card>
      </el-col>
    </el-row>
    <el-dialog v-model="dialogVisible" :title="editingRule ? '编辑规则' : '新增规则'" width="640px" destroy-on-close>
      <el-form label-position="top" :model="ruleForm">
        <el-row :gutter="16">
          <el-col :span="12"><el-form-item label="名称"><el-input v-model="ruleForm.name" /></el-form-item></el-col>
          <el-col :span="12"><el-form-item label="类型"><el-select v-model="ruleForm.patternType" style="width:100%"><el-option label="contains" value="contains" /><el-option label="regex" value="regex" /></el-select></el-form-item></el-col>
          <el-col :span="12"><el-form-item label="级别"><el-select v-model="ruleForm.severity" style="width:100%"><el-option label="critical" value="critical" /><el-option label="warning" value="warning" /><el-option label="info" value="info" /></el-select></el-form-item></el-col>
          <el-col :span="12"><el-form-item label="启用"><el-switch v-model="ruleForm.enabled" /></el-form-item></el-col>
        </el-row>
        <el-form-item label="匹配值"><el-input v-model="ruleForm.patternValue" type="textarea" :rows="4" placeholder="例如 Bearer 或正则表达式" /></el-form-item>
      </el-form>
      <template #footer><el-button @click="dialogVisible = false">取消</el-button><el-button type="primary" :loading="saving" @click="saveRule">保存</el-button></template>
    </el-dialog>
  </div>
</template>

<script setup>
import { onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { createKafkaSensitiveRule, deleteKafkaSensitiveRule, getKafkaClusterOptions, getKafkaSensitiveResults, getKafkaSensitiveRules, runKafkaSensitiveScan, updateKafkaSensitiveRule } from '@/api/kafka.js'

const clusters = ref([]); const selectedClusterId = ref(null); const loading = ref(false); const saving = ref(false); const scanning = ref(false); const rules = ref([]); const results = ref([]); const dialogVisible = ref(false); const editingRule = ref(null)
const ruleForm = reactive({ name: '', patternType: 'regex', patternValue: '', severity: 'warning', enabled: true })
const scanForm = reactive({ topic: '', partition: 0, limit: 20 })

const loadClusters = async () => { const res = await getKafkaClusterOptions(); clusters.value = res?.data?.data || []; if (!selectedClusterId.value && clusters.value.length > 0) selectedClusterId.value = clusters.value[0].id }
const loadRules = async () => { if (!selectedClusterId.value) return; const res = await getKafkaSensitiveRules({ clusterId: selectedClusterId.value }); rules.value = res?.data?.data || [] }
const loadResults = async () => { if (!selectedClusterId.value) return; const res = await getKafkaSensitiveResults({ clusterId: selectedClusterId.value }); results.value = res?.data?.data || [] }
const refreshAll = async () => { loading.value = true; try { await Promise.all([loadRules(), loadResults()]) } catch (error) { ElMessage.error(error.message || '敏感识别中心加载失败') } finally { loading.value = false } }
const openRuleDialog = (row = null) => { editingRule.value = row; Object.assign(ruleForm, row || { name: '', patternType: 'regex', patternValue: '', severity: 'warning', enabled: true }); dialogVisible.value = true }
const saveRule = async () => { if (!selectedClusterId.value || !ruleForm.name || !ruleForm.patternValue) { ElMessage.warning('请填写规则名和匹配值'); return } saving.value = true; try { const payload = { ...ruleForm, clusterId: selectedClusterId.value }; if (editingRule.value?.id) await updateKafkaSensitiveRule(editingRule.value.id, payload); else await createKafkaSensitiveRule(payload); ElMessage.success('扫描规则已保存'); dialogVisible.value = false; await loadRules() } catch (error) { ElMessage.error(error.message || '扫描规则保存失败') } finally { saving.value = false } }
const removeRule = async (row) => { await ElMessageBox.confirm(`确认删除规则 ${row.name} 吗？`, '提示', { type: 'warning' }); try { await deleteKafkaSensitiveRule(row.id); ElMessage.success('扫描规则已删除'); await loadRules() } catch (error) { ElMessage.error(error.message || '扫描规则删除失败') } }
const runScan = async () => { if (!selectedClusterId.value || !scanForm.topic) { ElMessage.warning('请填写 Topic'); return } scanning.value = true; try { await runKafkaSensitiveScan({ clusterId: selectedClusterId.value, topic: scanForm.topic, partition: Number(scanForm.partition), limit: Number(scanForm.limit) }); ElMessage.success('敏感扫描完成'); await loadResults() } catch (error) { ElMessage.error(error.message || '敏感扫描失败') } finally { scanning.value = false } }
onMounted(async () => { await loadClusters(); await refreshAll() })
</script>

<style scoped>.flex-between{display:flex;align-items:center;justify-content:space-between;}.scan-actions{display:flex;gap:8px;align-items:center;}</style>
