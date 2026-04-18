<template>
  <div class="autoops-container">
    <el-card class="page-header-card cyber-card">
      <div class="page-header">
        <div>
          <h2>Operator Subscriptions</h2>
          <p>管理 OLM Operator 订阅 (Subscriptions)</p>
        </div>
        <div class="header-actions">
          <div class="namespace-selector">
            <span class="selector-label">命名空间：</span>
            <el-select v-model="currentNamespace" placeholder="选择命名空间" @change="handleNamespaceChange" style="width: 200px;" class="autoops-select">
              <el-option label="所有" value="all" />
              <el-option
                v-for="ns in namespaceList"
                :key="ns.name"
                :label="ns.name"
                :value="ns.name"
              />
            </el-select>
          </div>
          <el-button class="autoops-btn-primary" @click="fetchData" :loading="loading">
            <el-icon><Refresh /></el-icon>
            刷新
          </el-button>
        </div>
      </div>
    </el-card>

    <el-card class="content-card cyber-card">
      <div class="autoops-table-wrapper">
        <el-table
          :data="tableData"
          style="width: 100%" class="autoops-table" v-loading="loading">`n        >
          <el-table-column prop="name" label="名称" min-width="180" show-overflow-tooltip>
             <template #default="scope">
              <span style="font-weight: 600; color: var(--primary-color);">{{ scope.row.name }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="namespace" label="命名空间" width="120" />
          <el-table-column prop="package" label="Package" min-width="120" show-overflow-tooltip />
          <el-table-column prop="channel" label="Channel" width="100">
             <template #default="scope">
                <el-tag effect="plain" class="autoops-tag">{{ scope.row.channel }}</el-tag>
             </template>
          </el-table-column>
          <el-table-column prop="startingCSV" label="Starting CSV" min-width="150" show-overflow-tooltip />
          <el-table-column prop="installedCSV" label="Installed CSV" min-width="150" show-overflow-tooltip />
          <el-table-column prop="state" label="状态" width="120">
             <template #default="scope">
                <el-tag :type="scope.row.state === 'AtLatestKnown' ? 'success' : 'warning'" effect="plain" class="autoops-tag">{{ scope.row.state }}</el-tag>
             </template>
          </el-table-column>
        
          <el-table-column label="创建时间" width="180">
            <template #default="scope">
              {{ formatTime(scope.row.age) }}
            </template>
          </el-table-column>
          <el-table-column label="操作" width="180" fixed="right">
            <template #default="scope">
              <el-button-group>
                <el-button size="small" @click="handleDetail(scope.row)" v-show="permStore.hasPerm('k8s:operator:handledetail')" >
                  <el-icon><View /></el-icon> 详情
                </el-button>
                <el-button size="small" type="danger" @click="handleDelete(scope.row)" v-show="permStore.hasPerm('k8s:operator:handledelete')" >
                  <el-icon><Delete /></el-icon> 删除
                </el-button>
              </el-button-group>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </el-card>

    <!-- 详情弹窗 -->
    <el-dialog
      v-model="detailVisible"
      title="Subscription 详情"
      width="800px"
      class="autoops-dialog"
    >
      <div class="yaml-content">
        <el-alert title="JSON 视图" type="info" :closable="false" style="margin-bottom: 15px;" />
        <pre class="code-block">{{ detailData }}</pre>
      </div>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="detailVisible = false">关闭</el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<script>
import { usePermissionStore } from '@/stores/permissionStore.js'
import { getSubscriptionList, getSubscriptionDetail, deleteSubscription } from '@/api/k8s/operator'
import { getNamespaceList } from '@/api/k8s/namespace'
import { getSelectedInstanceId } from '@/stores/instanceStore'
import moment from 'moment'
import { Refresh, View, Delete } from '@element-plus/icons-vue'

export default {
  name: 'OperatorManagement',
  components: {
    Refresh, View, Delete
  },
  data() {
    return {
      currentNamespace: 'all',
      namespaceList: [],
      tableData: [],
      loading: false,
      detailVisible: false,
      detailData: {}
    }
  },
  computed: {
    permStore() {
      return usePermissionStore()
    },
    instanceId() {
      return getSelectedInstanceId()
    }
  },
  watch: {
    instanceId() {
      this.fetchNamespaces()
      this.fetchData()
    }
  },
  created() {
    this.fetchNamespaces()
    this.fetchData()
  },
  methods: {
    formatTime(timestamp) {
      if(!timestamp) return '-'
      if (typeof timestamp === 'string' && (timestamp.includes('ago') || timestamp.includes('d') || timestamp.includes('h'))) {
          return timestamp
      }
      return moment.unix(timestamp).format('YYYY-MM-DD HH:mm:ss')
    },
    async fetchNamespaces() {
      if (!this.instanceId) return
      try {
        const res = await getNamespaceList(this.instanceId)
        if (res.status === 200) {
          this.namespaceList = res.data.namespaceList
        }
      } catch (error) {
        console.error(error)
      }
    },
    async fetchData() {
      if (!this.instanceId) return
      this.loading = true
      try {
        const res = await getSubscriptionList(this.currentNamespace, this.instanceId)
        if (res.status === 200) {
          this.tableData = res.data.subscriptionList
        }
      } catch (error) {
        // Handled
      } finally {
        this.loading = false
      }
    },
    handleNamespaceChange() {
      this.fetchData()
    },
    async handleDetail(row) {
      try {
        const res = await getSubscriptionDetail(row.namespace, row.name, this.instanceId)
        if (res.status === 200) {
          this.detailData = JSON.stringify(res.data.subscriptionDetail, null, 2)
          this.detailVisible = true
        }
      } catch (error) {
       // Handled
      }
    },
    handleDelete(row) {
      this.$confirm(`确认删除 Subscription "${row.name}"?`, '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }).then(async () => {
        try {
          const res = await deleteSubscription(row.namespace, row.name, this.instanceId)
          if (res.status === 200) {
            this.$message.success('删除成功')
            this.fetchData()
          }
        } catch (error) {
          // Handled
        }
      }).catch(() => {})
    }
  }
}
</script>

<style scoped>
.k8s-management {
  height: 100%;
  display: flex;
  flex-direction: column;
  gap: 20px;
}

/* Page Header */
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.page-header h2 {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
  color: var(--text-main);
  letter-spacing: -0.5px;
}

.page-header p {
  margin: 4px 0 0;
  color: var(--text-sub);
  font-size: 14px;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 16px;
}

.namespace-selector {
  display: flex;
  align-items: center;
  gap: 8px;
}

.selector-label {
  font-size: 14px;
  color: var(--text-sub);
}

/* Code Block */
.code-block {
  background: #1e1e1e;
  color: #d4d4d4;
  padding: 16px;
  border-radius: 8px;
  font-family: 'JetBrains Mono', monospace;
  font-size: 13px;
  line-height: 1.5;
  overflow: auto;
  max-height: 60vh;
  margin: 0;
}
</style>
