<template>
  <div class="page-container">
    <el-card class="page-header-card">
      <div class="page-header">
        <div>
          <h2>安全管理</h2>
          <p>管理 Kafka ACL 规则与 SCRAM 用户，适合开启 SASL / SCRAM 认证的集群</p>
        </div>
        <div class="header-actions">
          <el-select
            v-model="selectedClusterId"
            placeholder="选择 Kafka 集群"
            style="width: 260px"
            @change="handleClusterChange"
          >
            <el-option
              v-for="cluster in clusters"
              :key="cluster.id"
              :label="cluster.name"
              :value="cluster.id"
            />
          </el-select>
        </div>
      </div>
    </el-card>

    <el-card class="content-card">
      <el-tabs v-model="activeTab">
        <el-tab-pane label="ACL 规则" name="acls">
          <div class="toolbar">
            <el-select v-model="aclFilters.resourceType" clearable placeholder="资源类型" style="width: 160px">
              <el-option v-for="item in resourceTypeOptions" :key="item.value" :label="item.label" :value="item.value" />
            </el-select>
            <el-select v-model="aclFilters.patternType" clearable placeholder="Pattern" style="width: 150px">
              <el-option v-for="item in patternTypeOptions" :key="item.value" :label="item.label" :value="item.value" />
            </el-select>
            <el-input v-model="aclFilters.resourceName" placeholder="资源名" style="width: 180px" clearable @keyup.enter="loadACLs" />
            <el-input v-model="aclFilters.principal" placeholder="Principal，例如 User:alice" style="width: 220px" clearable @keyup.enter="loadACLs" />
            <el-button @click="loadACLs" :loading="aclLoading">刷新</el-button>
            <el-button
              v-if="permStore.hasPerm('kafka:acl:create') || permStore.roles.includes('admin')"
              type="primary"
              @click="openACLDialog"
            >
              新增 ACL
            </el-button>
          </div>

          <el-table v-loading="aclLoading" :data="aclList" empty-text="暂无 ACL 规则">
            <el-table-column prop="resourceType" label="资源类型" width="140" />
            <el-table-column prop="resourceName" label="资源名称" min-width="180" />
            <el-table-column prop="patternType" label="Pattern" width="120" />
            <el-table-column prop="principal" label="Principal" min-width="180" />
            <el-table-column prop="host" label="Host" width="120" />
            <el-table-column prop="operation" label="Operation" width="160" />
            <el-table-column prop="permissionType" label="Permission" width="130" />
            <el-table-column label="操作" width="110" fixed="right">
              <template #default="{ row }">
                <el-button
                  v-if="permStore.hasPerm('kafka:acl:delete') || permStore.roles.includes('admin')"
                  link
                  type="danger"
                  @click="handleDeleteACL(row)"
                >
                  删除
                </el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>

        <el-tab-pane label="SCRAM 用户" name="scram">
          <div class="toolbar">
            <el-input
              v-model="scramKeyword"
              placeholder="搜索用户名"
              style="width: 220px"
              clearable
              @keyup.enter="loadScramUsers"
            />
            <el-button @click="loadScramUsers" :loading="scramLoading">刷新</el-button>
            <el-button
              v-if="permStore.hasPerm('kafka:scram:user:upsert') || permStore.roles.includes('admin')"
              type="primary"
              @click="openScramDialog()"
            >
              新增 / 更新用户
            </el-button>
          </div>

          <el-table v-loading="scramLoading" :data="scramRows" empty-text="暂无 SCRAM 用户">
            <el-table-column prop="username" label="用户名" min-width="220" />
            <el-table-column prop="mechanism" label="机制" min-width="180" />
            <el-table-column prop="iterations" label="迭代次数" min-width="160" />
            <el-table-column label="操作" width="180" fixed="right">
              <template #default="{ row }">
                <el-button
                  v-if="permStore.hasPerm('kafka:scram:user:upsert') || permStore.roles.includes('admin')"
                  link
                  type="primary"
                  @click="openScramDialog(row)"
                >
                  更新密码
                </el-button>
                <el-button
                  v-if="permStore.hasPerm('kafka:scram:user:delete') || permStore.roles.includes('admin')"
                  link
                  type="danger"
                  @click="handleDeleteScram(row)"
                >
                  删除
                </el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>
      </el-tabs>
    </el-card>

    <el-dialog v-model="aclDialogVisible" title="新增 ACL" width="640px" destroy-on-close>
      <el-form label-position="top" :model="aclForm">
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="资源类型">
              <el-select v-model="aclForm.resourceType" style="width: 100%">
                <el-option v-for="item in resourceTypeOptions" :key="item.value" :label="item.label" :value="item.value" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="资源名">
              <el-input v-model="aclForm.resourceName" placeholder="例如 orders.events" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="Pattern">
              <el-select v-model="aclForm.patternType" style="width: 100%">
                <el-option v-for="item in patternTypeOptions" :key="item.value" :label="item.label" :value="item.value" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="Principal">
              <el-input v-model="aclForm.principal" placeholder="例如 User:alice 或 User:*" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="16">
          <el-col :span="8">
            <el-form-item label="Host">
              <el-input v-model="aclForm.host" placeholder="默认 *" />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="Operation">
              <el-select v-model="aclForm.operation" style="width: 100%">
                <el-option v-for="item in operationOptions" :key="item.value" :label="item.label" :value="item.value" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="Permission">
              <el-select v-model="aclForm.permissionType" style="width: 100%">
                <el-option v-for="item in permissionTypeOptions" :key="item.value" :label="item.label" :value="item.value" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
      </el-form>
      <template #footer>
        <el-button @click="aclDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="aclSaving" @click="handleCreateACL">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="scramDialogVisible" :title="scramDialogTitle" width="560px" destroy-on-close>
      <el-form label-position="top" :model="scramForm">
        <el-form-item label="用户名">
          <el-input v-model="scramForm.username" :disabled="scramEditing" placeholder="例如 alice" />
        </el-form-item>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="机制">
              <el-select v-model="scramForm.mechanism" style="width: 100%">
                <el-option label="SCRAM-SHA-256" value="sha256" />
                <el-option label="SCRAM-SHA-512" value="sha512" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="迭代次数">
              <el-input-number v-model="scramForm.iterations" :min="4096" style="width: 100%" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="密码">
          <el-input v-model="scramForm.password" type="password" show-password placeholder="请输入新密码" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="scramDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="scramSaving" @click="handleUpsertScramUser">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  createKafkaACL,
  deleteKafkaACL,
  deleteKafkaScramUser,
  getKafkaACLs,
  getKafkaClusterOptions,
  getKafkaScramUsers,
  upsertKafkaScramUser,
} from '@/api/kafka.js'
import { usePermissionStore } from '@/stores/permissionStore.js'

const permStore = usePermissionStore()

const activeTab = ref('acls')
const selectedClusterId = ref(null)
const clusters = ref([])

const aclLoading = ref(false)
const aclSaving = ref(false)
const aclDialogVisible = ref(false)
const aclList = ref([])
const aclFilters = reactive({
  resourceType: '',
  resourceName: '',
  patternType: '',
  principal: '',
})
const aclForm = reactive({
  resourceType: 'Topic',
  resourceName: '',
  patternType: 'Literal',
  principal: 'User:',
  host: '*',
  operation: 'Read',
  permissionType: 'Allow',
})

const scramLoading = ref(false)
const scramSaving = ref(false)
const scramDialogVisible = ref(false)
const scramEditing = ref(false)
const scramUsers = ref([])
const scramKeyword = ref('')
const scramForm = reactive({
  username: '',
  mechanism: 'sha256',
  iterations: 4096,
  password: '',
})

const resourceTypeOptions = [
  { label: 'Topic', value: 'Topic' },
  { label: 'Group', value: 'Group' },
  { label: 'Cluster', value: 'Cluster' },
  { label: 'TransactionalID', value: 'TransactionalID' },
  { label: 'DelegationToken', value: 'DelegationToken' },
]

const patternTypeOptions = [
  { label: 'Literal', value: 'Literal' },
  { label: 'Prefixed', value: 'Prefixed' },
]

const operationOptions = [
  { label: 'Read', value: 'Read' },
  { label: 'Write', value: 'Write' },
  { label: 'Create', value: 'Create' },
  { label: 'Delete', value: 'Delete' },
  { label: 'Alter', value: 'Alter' },
  { label: 'Describe', value: 'Describe' },
  { label: 'ClusterAction', value: 'ClusterAction' },
  { label: 'DescribeConfigs', value: 'DescribeConfigs' },
  { label: 'AlterConfigs', value: 'AlterConfigs' },
  { label: 'All', value: 'All' },
  { label: 'IdempotentWrite', value: 'IdempotentWrite' },
]

const permissionTypeOptions = [
  { label: 'Allow', value: 'Allow' },
  { label: 'Deny', value: 'Deny' },
]

const scramDialogTitle = computed(() => (scramEditing.value ? '更新 SCRAM 用户密码' : '新增 SCRAM 用户'))
const scramRows = computed(() =>
  scramUsers.value.flatMap((user) => {
    if (!user.credentials?.length) {
      return [{ username: user.username, mechanism: '-', iterations: '-' }]
    }
    return user.credentials.map((credential) => ({
      username: user.username,
      mechanism: credential.mechanism,
      iterations: credential.iterations,
    }))
  }),
)

const loadClusters = async () => {
  const res = await getKafkaClusterOptions()
  clusters.value = res?.data?.data || []
  if (!selectedClusterId.value && clusters.value.length > 0) {
    selectedClusterId.value = clusters.value[0].id
  }
}

const loadACLs = async () => {
  if (!selectedClusterId.value) return
  aclLoading.value = true
  try {
    const res = await getKafkaACLs({
      clusterId: selectedClusterId.value,
      resourceType: aclFilters.resourceType,
      resourceName: aclFilters.resourceName,
      patternType: aclFilters.patternType,
      principal: aclFilters.principal,
    })
    aclList.value = res?.data?.data || []
  } catch (error) {
    ElMessage.error(error.message || 'ACL 列表加载失败')
  } finally {
    aclLoading.value = false
  }
}

const loadScramUsers = async () => {
  if (!selectedClusterId.value) return
  scramLoading.value = true
  try {
    const res = await getKafkaScramUsers({
      clusterId: selectedClusterId.value,
      keyword: scramKeyword.value,
    })
    scramUsers.value = res?.data?.data || []
  } catch (error) {
    ElMessage.error(error.message || 'SCRAM 用户加载失败')
  } finally {
    scramLoading.value = false
  }
}

const handleClusterChange = async () => {
  await Promise.all([loadACLs(), loadScramUsers()])
}

const openACLDialog = () => {
  aclForm.resourceType = 'Topic'
  aclForm.resourceName = ''
  aclForm.patternType = 'Literal'
  aclForm.principal = 'User:'
  aclForm.host = '*'
  aclForm.operation = 'Read'
  aclForm.permissionType = 'Allow'
  aclDialogVisible.value = true
}

const handleCreateACL = async () => {
  if (!selectedClusterId.value || !aclForm.resourceName || !aclForm.principal) {
    ElMessage.warning('请填写完整 ACL 信息')
    return
  }
  aclSaving.value = true
  try {
    await createKafkaACL({
      clusterId: selectedClusterId.value,
      resourceType: aclForm.resourceType,
      resourceName: aclForm.resourceName.trim(),
      patternType: aclForm.patternType,
      principal: aclForm.principal.trim(),
      host: aclForm.host.trim() || '*',
      operation: aclForm.operation,
      permissionType: aclForm.permissionType,
    })
    ElMessage.success('ACL 创建成功')
    aclDialogVisible.value = false
    await loadACLs()
  } catch (error) {
    ElMessage.error(error.message || 'ACL 创建失败')
  } finally {
    aclSaving.value = false
  }
}

const handleDeleteACL = async (row) => {
  if (!selectedClusterId.value) return
  await ElMessageBox.confirm(
    `确认删除 ${row.principal} 对 ${row.resourceType}/${row.resourceName} 的 ACL 吗？`,
    '危险操作确认',
    { type: 'warning' },
  )
  try {
    await deleteKafkaACL({
      clusterId: selectedClusterId.value,
      resourceType: row.resourceType,
      resourceName: row.resourceName,
      patternType: row.patternType,
      principal: row.principal,
      host: row.host,
      operation: row.operation,
      permissionType: row.permissionType,
    })
    ElMessage.success('ACL 删除成功')
    await loadACLs()
  } catch (error) {
    ElMessage.error(error.message || 'ACL 删除失败')
  }
}

const openScramDialog = (row = null) => {
  scramEditing.value = !!row
  scramForm.username = row?.username || ''
  scramForm.mechanism = row?.mechanism?.includes('512') ? 'sha512' : 'sha256'
  scramForm.iterations = Number(row?.iterations || 4096)
  scramForm.password = ''
  scramDialogVisible.value = true
}

const handleUpsertScramUser = async () => {
  if (!selectedClusterId.value || !scramForm.username || !scramForm.password) {
    ElMessage.warning('请填写用户名和密码')
    return
  }
  scramSaving.value = true
  try {
    await upsertKafkaScramUser({
      clusterId: selectedClusterId.value,
      username: scramForm.username.trim(),
      mechanism: scramForm.mechanism,
      iterations: Number(scramForm.iterations),
      password: scramForm.password,
    })
    ElMessage.success('SCRAM 用户已保存')
    scramDialogVisible.value = false
    await loadScramUsers()
  } catch (error) {
    ElMessage.error(error.message || 'SCRAM 用户保存失败')
  } finally {
    scramSaving.value = false
  }
}

const handleDeleteScram = async (row) => {
  if (!selectedClusterId.value || !row?.mechanism || row.mechanism === '-') return
  await ElMessageBox.confirm(
    `确认删除 SCRAM 用户 ${row.username} 的 ${row.mechanism} 凭据吗？`,
    '危险操作确认',
    { type: 'warning' },
  )
  try {
    await deleteKafkaScramUser({
      clusterId: selectedClusterId.value,
      username: row.username,
      mechanism: row.mechanism.includes('512') ? 'sha512' : 'sha256',
    })
    ElMessage.success('SCRAM 用户凭据已删除')
    await loadScramUsers()
  } catch (error) {
    ElMessage.error(error.message || 'SCRAM 用户删除失败')
  }
}

onMounted(async () => {
  try {
    await loadClusters()
    await Promise.all([loadACLs(), loadScramUsers()])
  } catch (error) {
    ElMessage.error(error.message || 'Kafka 安全管理初始化失败')
  }
})
</script>

<style scoped>
.toolbar {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  margin-bottom: 16px;
}
</style>
