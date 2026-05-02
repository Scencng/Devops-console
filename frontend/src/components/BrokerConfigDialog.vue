<template>
  <el-dialog v-model="dialogVisible" :title="dialogTitle" width="760px" destroy-on-close @closed="resetDialog">
    <el-alert
      type="warning"
      :closable="false"
      show-icon
      title="Broker 动态配置会直接写入 Kafka 节点，请确认配置项和值都已经过核对。"
    />

    <div class="config-editor">
      <div v-for="(entry, index) in configRows" :key="entry.id" class="config-row">
        <el-row :gutter="12">
          <el-col :xs="24" :sm="10">
            <el-input v-model="entry.key" placeholder="配置项，例如 log.retention.ms" />
          </el-col>
          <el-col :xs="24" :sm="12">
            <el-input v-model="entry.value" placeholder="配置值，例如 86400000" />
          </el-col>
          <el-col :xs="24" :sm="2" class="row-actions">
            <el-button link type="danger" @click="removeConfigRow(index)">删除</el-button>
          </el-col>
        </el-row>
      </div>
      <el-button text type="primary" @click="addConfigRow">新增配置项</el-button>
    </div>

    <template #footer>
      <el-button @click="dialogVisible = false">取消</el-button>
      <el-button type="primary" :loading="saving" @click="handleSubmit">提交</el-button>
    </template>
  </el-dialog>
</template>

<script setup>
import { computed, ref } from 'vue'
import { storeToRefs } from 'pinia'
import { ElMessage } from 'element-plus'
import { updateKafkaBrokerConfig } from '@/api/kafka.js'
import { useKafkaStore } from '@/stores/kafkaStore.js'

const emit = defineEmits(['success'])

const dialogVisible = ref(false)
const saving = ref(false)
const activeBrokerId = ref(null)
const kafkaStore = useKafkaStore()
const { selectedClusterId } = storeToRefs(kafkaStore)

let nextRowID = 1

const createEmptyRow = () => ({
  id: nextRowID++,
  key: '',
  value: '',
})

const configRows = ref([createEmptyRow()])

const dialogTitle = computed(() =>
  activeBrokerId.value === null ? '修改 Broker 动态配置' : `修改 Broker 动态配置: ${activeBrokerId.value}`,
)

const resetDialog = () => {
  if (saving.value) return
  activeBrokerId.value = null
  configRows.value = [createEmptyRow()]
}

const addConfigRow = () => {
  configRows.value.push(createEmptyRow())
}

const removeConfigRow = (index) => {
  if (configRows.value.length === 1) {
    configRows.value = [createEmptyRow()]
    return
  }
  configRows.value.splice(index, 1)
}

const buildConfigPayload = () => {
  const payload = {}
  const duplicateKeys = new Set()

  configRows.value.forEach((entry) => {
    const key = String(entry.key || '').trim()
    if (!key) return
    if (Object.prototype.hasOwnProperty.call(payload, key)) {
      duplicateKeys.add(key)
      return
    }
    payload[key] = String(entry.value ?? '')
  })

  if (duplicateKeys.size > 0) {
    throw new Error(`配置项重复：${Array.from(duplicateKeys).join('，')}`)
  }
  if (Object.keys(payload).length === 0) {
    throw new Error('请至少填写一条配置项')
  }
  return payload
}

const handleSubmit = async () => {
  if (!selectedClusterId.value) {
    ElMessage.warning('请先选择 Kafka 集群')
    return
  }
  if (activeBrokerId.value === null || activeBrokerId.value === undefined) {
    ElMessage.warning('未指定 Broker ID')
    return
  }

  let payload
  try {
    payload = buildConfigPayload()
  } catch (error) {
    ElMessage.warning(error.message || '请检查配置项')
    return
  }

  saving.value = true
  try {
    await updateKafkaBrokerConfig(activeBrokerId.value, selectedClusterId.value, payload)
    ElMessage.success('Broker 动态配置修改成功')
    dialogVisible.value = false
    emit('success')
  } catch (error) {
    ElMessage.error(error.message || 'Broker 动态配置修改失败')
  } finally {
    saving.value = false
  }
}

const openDialog = (brokerId) => {
  if (!selectedClusterId.value) {
    ElMessage.warning('请先选择 Kafka 集群')
    return
  }
  activeBrokerId.value = brokerId
  configRows.value = [createEmptyRow()]
  dialogVisible.value = true
}

defineExpose({
  openDialog,
})
</script>

<style scoped>
.config-editor {
  display: flex;
  flex-direction: column;
  gap: 12px;
  margin-top: 16px;
}

.config-row {
  padding: 14px 16px;
  border: 1px solid rgba(148, 163, 184, 0.14);
  border-radius: 16px;
  background: rgba(248, 250, 252, 0.82);
}

.row-actions {
  display: flex;
  align-items: center;
  justify-content: flex-end;
}

@media (max-width: 768px) {
  .row-actions {
    justify-content: flex-start;
  }
}
</style>
