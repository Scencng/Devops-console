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
        <el-table-column prop="target" label="目标引用" min-width="180" show-overflow-tooltip>
           <template #default="scope">
             {{ scope.row.target }}
           </template>
        </el-table-column>
        <el-table-column prop="min" label="最小副本" width="100" />
        <el-table-column prop="max" label="最大副本" width="100" />
        <el-table-column prop="replicas" label="当前副本" width="100">
           <template #default="scope">
              <el-tag effect="plain" class="autoops-tag">{{ scope.row.replicas }}</el-tag>
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
              <el-button class="autoops-action-btn delete" @click="handleDelete(scope.row)" v-show="permStore.hasPerm('k8s:hpa:handledelete')" >
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
      title="HPA 详情"
      width="800px"
      class="autoops-dialog"
    >
      <div class="detail-container">
        <el-tabs v-model="activeTab" type="border-card" class="autoops-tabs">
            <el-tab-pane label="JSON" name="json">
                <div class="code-wrapper">
                    <pre class="code-block json">{{ JSON.stringify(detailData, null, 2) }}</pre>
                </div>
            </el-tab-pane>
            <el-tab-pane label="YAML" name="yaml">
                <div class="code-wrapper">
                    <pre class="code-block yaml">{{ yamlData }}</pre>
                </div>
            </el-tab-pane>
        </el-tabs>
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
import { getHPAList, getHPADetail, deleteHPA } from '@/api/k8s/hpa'
import { getNamespaceList } from '@/api/k8s/namespace'
import { getSelectedInstanceId } from '@/stores/instanceStore'
import moment from 'moment'
import { Refresh, View, Delete, CopyDocument } from '@element-plus/icons-vue'
import yaml from 'js-yaml'

export default {
  name: 'HpaManagement',
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
      detailData: {},
      yamlData: '',
      activeTab: 'json'
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
        const res = await getHPAList(this.currentNamespace, this.instanceId)
        if (res.status === 200) {
          this.tableData = res.data.hpaList
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
        const res = await getHPADetail(row.namespace, row.name, this.instanceId)
        if (res.status === 200) {
          this.detailData = res.data.hpaDetail
          this.yamlData = yaml.dump(res.data.hpaDetail)
          this.activeTab = 'json'
          this.detailVisible = true
        }
      } catch (error) {
       // Handled
      }
    },
    handleDelete(row) {
      this.$confirm(`确认删除 HPA "${row.name}"?`, '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }).then(async () => {
        try {
          const res = await deleteHPA(row.namespace, row.name, this.instanceId)
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
/* Detail Container */
.detail-container {
    max-height: 60vh;
    overflow: hidden;
    display: flex;
    flex-direction: column;
}

.autoops-tabs {
    border: none;
    background: transparent;
    box-shadow: none;
}

.autoops-tabs :deep(.el-tabs__header) {
    background: #f5f7fa;
    border-bottom: 1px solid #e4e7ed;
    border-radius: 4px 4px 0 0;
}

.autoops-tabs :deep(.el-tabs__item) {
    color: #909399;
    border-right: 1px solid #e4e7ed;
}

.autoops-tabs :deep(.el-tabs__item.is-active) {
    background: #1e1e1e;
    color: #409eff;
    border-bottom-color: #1e1e1e;
}

.autoops-tabs :deep(.el-tabs__content) {
    padding: 0;
    background: #1e1e1e;
}

.code-wrapper {
    height: 50vh;
    overflow: auto;
    background: #1e1e1e;
    padding: 16px;
}

.code-block {
  color: #d4d4d4;
  font-family: 'JetBrains Mono', 'Consolas', 'Monaco', monospace;
  font-size: 13px;
  line-height: 1.5;
  margin: 0;
  white-space: pre-wrap;
  word-wrap: break-word;
}

.code-block.yaml {
    color: #a9b7c6;
}

.code-block.json {
    color: #a9b7c6;
}
</style>
