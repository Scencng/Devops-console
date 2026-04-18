<template>
  <div class="autoops-container">
    <el-card class="page-header-card cyber-card">
      <div class="page-header">
        <div>
          <h2>CustomResourceDefinitions</h2>
          <p>管理 Kubernetes 自定义资源定义</p>
        </div>
        <div class="header-actions">
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
          <el-table-column prop="name" label="名称" min-width="220" show-overflow-tooltip>
             <template #default="scope">
              <span style="font-weight: 600; color: var(--primary-color);">{{ scope.row.name }}</span>
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
                <el-button size="small" @click="handleDetail(scope.row)" v-show="permStore.hasPerm('k8s:crd:handledetail')" >
                  <el-icon><View /></el-icon> 详情
                </el-button>
                <el-button size="small" type="danger" @click="handleDelete(scope.row)" v-show="permStore.hasPerm('k8s:crd:handledelete')" >
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
      title="CRD 详情"
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
import { getCRDList, getCRDDetail, deleteCRD } from '@/api/k8s/crd'
import { getSelectedInstanceId } from '@/stores/instanceStore'
import moment from 'moment'
import { Refresh, View, Delete } from '@element-plus/icons-vue'

export default {
  name: 'CrdManagement',
  components: {
    Refresh, View, Delete
  },
  data() {
    return {
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
      this.fetchData()
    }
  },
  created() {
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
    async fetchData() {
      if (!this.instanceId) return
      this.loading = true
      try {
        const res = await getCRDList(this.instanceId)
        if (res.status === 200) {
          this.tableData = res.data.crdList
        }
      } catch (error) {
        // Handled
      } finally {
        this.loading = false
      }
    },
    async handleDetail(row) {
      try {
        const res = await getCRDDetail(row.name, this.instanceId)
        if (res.status === 200) {
          this.detailData = JSON.stringify(res.data.crdDetail, null, 2)
          this.detailVisible = true
        }
      } catch (error) {
        // Handled
      }
    },
    handleDelete(row) {
      this.$confirm(`确认删除 CRD "${row.name}"?`, '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }).then(async () => {
        try {
          const res = await deleteCRD(row.name, this.instanceId)
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
