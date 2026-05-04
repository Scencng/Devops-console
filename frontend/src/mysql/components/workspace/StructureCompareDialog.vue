<template>
  <el-dialog
    :model-value="visible"
    :title="dialogTitle"
    width="960px"
    class="workspace-dialog"
    @update:model-value="emit('update:visible', $event)"
  >
    <div class="compare-toolbar">
      <el-form label-position="top" class="compare-form">
        <div class="compare-grid">
          <el-form-item :label="text.sourceDatabase">
            <el-input :model-value="sourceDatabase" disabled />
          </el-form-item>

          <el-form-item :label="text.targetDatabase">
            <el-select v-model="targetDatabase" filterable>
              <el-option v-for="database in databases" :key="database" :label="database" :value="database" />
            </el-select>
          </el-form-item>

          <el-form-item v-if="scope === 'table'" :label="text.sourceTable">
            <el-input :model-value="sourceTable" disabled />
          </el-form-item>

          <el-form-item v-if="scope === 'table'" :label="text.targetTable">
            <el-select v-model="targetTable" filterable>
              <el-option v-for="table in targetTables" :key="table" :label="table" :value="table" />
            </el-select>
          </el-form-item>
        </div>
      </el-form>

      <div class="compare-actions">
        <el-button :loading="loading" type="primary" @click="runCompare">{{ text.compare }}</el-button>
      </div>
    </div>

    <el-empty v-if="!loading && items.length === 0" :description="text.noDiffItems" />

    <div v-else class="compare-results">
      <div v-for="item in items" :key="item.id" class="compare-card">
        <div class="compare-card__header">
          <div>
            <strong>{{ item.title }}</strong>
            <p>{{ item.detail }}</p>
          </div>
          <el-tag :type="tagType(item.status)">{{ item.status }}</el-tag>
        </div>

        <div v-if="item.sourceValue || item.targetValue" class="compare-card__values">
          <div>
            <span>{{ text.sourceValue }}</span>
            <pre>{{ item.sourceValue || '-' }}</pre>
          </div>
          <div>
            <span>{{ text.targetValue }}</span>
            <pre>{{ item.targetValue || '-' }}</pre>
          </div>
        </div>

        <div v-if="item.statements?.length" class="compare-card__sql">
          <span>{{ text.statements }}</span>
          <pre>{{ item.statements.join('\n') }}</pre>
        </div>
      </div>
    </div>

    <template #footer>
      <el-button @click="emit('update:visible', false)">{{ text.close }}</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'

import request from '@/mysql/utils/request'

type CompareScope = 'database' | 'table'

interface SchemaDiffItem {
  id: string
  title: string
  detail: string
  status: string
  sourceValue?: string
  targetValue?: string
  statements?: string[]
}

interface SchemaCompareResponse {
  items: SchemaDiffItem[]
}

const props = defineProps<{
  visible: boolean
  scope: CompareScope
  sourceDatabase: string
  sourceTable?: string
}>()

const emit = defineEmits<{
  'update:visible': [value: boolean]
  'refresh-explorer': []
}>()

const text = {
  sourceDatabase: '\u6e90\u6570\u636e\u5e93',
  targetDatabase: '\u76ee\u6807\u6570\u636e\u5e93',
  sourceTable: '\u6e90\u6570\u636e\u8868',
  targetTable: '\u76ee\u6807\u6570\u636e\u8868',
  compare: '\u5f00\u59cb\u6bd4\u5bf9',
  noDiffItems: '\u6682\u65e0\u5dee\u5f02\u9879',
  sourceValue: '\u6e90\u5bf9\u8c61',
  targetValue: '\u76ee\u6807\u5bf9\u8c61',
  statements: '\u5dee\u5f02\u8bed\u53e5',
  close: '\u5173\u95ed',
  compareTableSchema: '\u6bd4\u5bf9\u6570\u636e\u8868\u7ed3\u6784',
  compareDatabaseSchema: '\u6bd4\u5bf9\u6570\u636e\u5e93\u7ed3\u6784',
  selectTargetDatabase: '\u8bf7\u9009\u62e9\u76ee\u6807\u6570\u636e\u5e93',
  selectTargetTable: '\u8bf7\u9009\u62e9\u76ee\u6807\u6570\u636e\u8868'
} as const

const loading = ref(false)
const databases = ref<string[]>([])
const targetTables = ref<string[]>([])
const targetDatabase = ref('')
const targetTable = ref('')
const items = ref<SchemaDiffItem[]>([])

const dialogTitle = computed(() =>
  props.scope === 'table' ? text.compareTableSchema : text.compareDatabaseSchema
)

watch(
  () => props.visible,
  async (visible) => {
    if (!visible) {
      return
    }

    items.value = []
    databases.value = await request.get<string[]>('/api/metadata/databases', { silentError: true }).catch(() => [])
    targetDatabase.value = databases.value.find((item) => item !== props.sourceDatabase) || props.sourceDatabase
  },
  { immediate: true }
)

watch(
  () => targetDatabase.value,
  async (database) => {
    if (props.scope !== 'table' || !database) {
      targetTables.value = []
      targetTable.value = ''
      return
    }

    targetTables.value = await request.get<string[]>('/api/metadata/tables', {
      params: { db: database },
      silentError: true
    }).catch(() => [])
    targetTable.value = targetTables.value.find((item) => item !== props.sourceTable) || targetTables.value[0] || ''
  }
)

async function runCompare() {
  if (!targetDatabase.value) {
    ElMessage.warning(text.selectTargetDatabase)
    return
  }

  if (props.scope === 'table' && !targetTable.value) {
    ElMessage.warning(text.selectTargetTable)
    return
  }

  loading.value = true
  try {
    const result = await request.post<SchemaCompareResponse>('/api/schema/compare', {
      scope: props.scope,
      sourceDatabase: props.sourceDatabase,
      sourceTable: props.sourceTable,
      targetDatabase: targetDatabase.value,
      targetTable: props.scope === 'table' ? targetTable.value : ''
    })
    items.value = result.items || []
  } finally {
    loading.value = false
  }
}

function tagType(status: string) {
  if (status === 'add') return 'success'
  if (status === 'remove') return 'danger'
  if (status === 'modify') return 'warning'
  return 'info'
}
</script>

<style scoped>
.compare-toolbar {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.compare-form {
  width: 100%;
}

.compare-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0 16px;
}

.compare-actions {
  display: flex;
  justify-content: flex-end;
}

.compare-results {
  display: flex;
  flex-direction: column;
  gap: 12px;
  max-height: 60vh;
  overflow: auto;
}

.compare-card {
  padding: 16px;
  border: 1px solid var(--devops-border-light);
  border-radius: var(--devops-radius-md);
  background: #fff;
}

.compare-card__header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
}

.compare-card__header p,
.compare-card__values span,
.compare-card__sql span {
  color: var(--devops-text-secondary);
  font-size: 12px;
}

.compare-card__values {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
  margin-top: 12px;
}

.compare-card__values pre,
.compare-card__sql pre {
  margin: 6px 0 0;
  padding: 12px;
  border-radius: var(--devops-radius-sm);
  background: #f7f8fa;
  white-space: pre-wrap;
  word-break: break-word;
}

.compare-card__sql {
  margin-top: 12px;
}
</style>
