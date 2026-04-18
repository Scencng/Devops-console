<template>
  <div class="autoops-container">
    <div class="autoops-filter-bar">
      <label class="autoops-filter-label">命名空间</label>
      <el-select v-model="currentNamespace" placeholder="选择命名空间" @change="handleNamespaceChange" style="width: 200px;" class="autoops-select">
        <el-option label="所有" value="all" />
        <el-option
          v-for="ns in namespaceList"
          :key="ns.name"
          :label="ns.name"
          :value="ns.name"
        />
      </el-select>
      <el-button class="autoops-btn-primary" @click="fetchData" :loading="loading">
        <el-icon><Refresh /></el-icon>
        <span>刷新</span>
      </el-button>
    </div>

    <div class="autoops-table-wrapper">
      <el-table
        :data="tableData"
        class="autoops-table" style="width: 100%"
        v-loading="loading"
      >
        <el-table-column prop="name" label="名称" min-width="180" show-overflow-tooltip>
           <template #default="scope">
            <span style="font-weight: 600; color: #409EFF;">{{ scope.row.name }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="namespace" label="命名空间" width="120" />
        <el-table-column prop="mode" label="模式" width="120">
           <template #default="scope">
              <el-tag effect="plain" class="autoops-tag">{{ scope.row.mode }}</el-tag>
           </template>
        </el-table-column>
        <el-table-column label="目标引用" min-width="180" show-overflow-tooltip>
          <template #default="scope">
            {{ scope.row.reference }}
          </template>
        </el-table-column>
        <el-table-column label="创建时间" width="180">
          <template #default="scope">
            {{ formatTime(scope.row.age) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="150" fixed="right">
          <template #default="scope">
            <div class="autoops-actions">
              <el-tooltip content="查看详情" placement="top">
                <el-button class="autoops-action-btn view" @click="handleDetail(scope.row)">
                  <el-icon><View /></el-icon>
                </el-button>
              </el-tooltip>
              <el-button class="autoops-action-btn delete" @click="handleDelete(scope.row)" v-show="permStore.hasPerm('k8s:vpa:handledelete')" >
                  <el-icon><Delete /></el-icon>
                </el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <!-- 详情弹窗 -->
    <el-dialog
      v-model="detailVisible"
      title="VPA 详情"
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
import { getVPAList, getVPADetail, deleteVPA } from '@/api/k8s/vpa'
import { getNamespaceList } from '@/api/k8s/namespace'
import { getSelectedInstanceId } from '@/stores/instanceStore'
import moment from 'moment'
import { Refresh, View, Delete } from '@element-plus/icons-vue'

export default {
  name: 'VpaManagement',
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
        if (res.code === 200) {
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
        const res = await getVPAList(this.currentNamespace, this.instanceId)
        if (res.code === 200) {
          this.tableData = res.data.vpaList
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
        const res = await getVPADetail(row.namespace, row.name, this.instanceId)
        if (res.code === 200) {
          this.detailData = JSON.stringify(res.data.vpaDetail, null, 2)
          this.detailVisible = true
        }
      } catch (error) {
        // Handled
      }
    },
    handleDelete(row) {
      this.$confirm(`确认删除 VPA "${row.name}"?`, '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }).then(async () => {
        try {
          const res = await deleteVPA(row.namespace, row.name, this.instanceId)
          if (res.code === 200) {
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
/* Code Block */
.code-block {
  background: #1e1e1e;
  color: #d4d4d4;
  padding: 16px;
  border-radius: 8px;
  font-family: 'JetBrains Mono', 'Consolas', 'Monaco', monospace;
  font-size: 13px;
  line-height: 1.5;
  overflow: auto;
  max-height: 60vh;
  margin: 0;
}
</style>
