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
        <el-table-column prop="desired" label="期望副本" width="100">
           <template #default="scope">
              <el-tag effect="plain" class="autoops-tag">{{ scope.row.desired }}</el-tag>
           </template>
        </el-table-column>
        <el-table-column prop="current" label="当前副本" width="100">
           <template #default="scope">
              <el-tag effect="plain" type="success" class="autoops-tag success">{{ scope.row.current }}</el-tag>
           </template>
        </el-table-column>
        <el-table-column prop="ready" label="就绪副本" width="100">
           <template #default="scope">
              <el-tag effect="plain" type="warning" class="autoops-tag warning">{{ scope.row.ready }}</el-tag>
           </template>
        </el-table-column>
        <el-table-column label="镜像" min-width="180" show-overflow-tooltip>
          <template #default="scope">
            <div class="image-cell">
              <el-icon class="image-icon"><Box /></el-icon>
              <span class="image-name" :title="scope.row.images && scope.row.images[0]">
                {{ getShortImageName(scope.row.images && scope.row.images[0]) }}
              </span>
              <el-button 
                v-if="scope.row.images && scope.row.images.length > 1" 
                class="more-images-btn"
                size="small"
                @click="showImagesDialog(scope.row)"
              >
                +{{ scope.row.images.length - 1 }}
              </el-button>
            </div>
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
              <el-button class="autoops-action-btn delete" @click="handleDelete(scope.row)" v-show="permStore.hasPerm('k8s:replicaset:handledelete')" >
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
      title="ReplicaSet 详情"
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

    <!-- 镜像列表弹窗 -->
    <el-dialog
      v-model="imagesDialogVisible"
      :title="`镜像列表 - ${currentRow.name}`"
      width="600px"
      class="autoops-dialog"
    >
      <div class="images-list-container">
        <el-alert :title="`容器镜像 (${currentImages.length}个)`" type="info" :closable="false" style="margin-bottom: 15px;" />
        <div class="images-list">
          <div v-for="(img, index) in currentImages" :key="index" class="image-item">
            <el-icon class="image-item-icon"><Box /></el-icon>
            <div class="image-item-content">
              <div class="image-item-name">{{ img }}</div>
              <div class="image-item-tags">
                <el-tag size="small" effect="plain">latest</el-tag>
                <el-tag size="small" type="success" effect="plain">docker.io</el-tag>
              </div>
            </div>
            <el-button class="image-copy-btn" size="small" @click="copyImage(img)">
              复制
            </el-button>
          </div>
        </div>
      </div>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="imagesDialogVisible = false">关闭</el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<script>
import { usePermissionStore } from '@/stores/permissionStore.js'
import { getReplicaSetList, getReplicaSetDetail, deleteReplicaSet } from '@/api/k8s/replicaset'
import { getNamespaceList } from '@/api/k8s/namespace'
import { getSelectedInstanceId } from '@/stores/instanceStore'
import moment from 'moment'
import { Refresh, View, Delete, Box } from '@element-plus/icons-vue'

export default {
  name: 'ReplicaSetManagement',
  components: {
    Refresh, View, Delete, Box
  },
  data() {
    return {
      currentNamespace: 'all',
      namespaceList: [],
      tableData: [],
      loading: false,
      detailVisible: false,
      detailData: {},
      imagesDialogVisible: false,
      currentRow: {},
      currentImages: []
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
      // 如果已经是相对时间字符串，直接返回
      if (typeof timestamp === 'string' && (timestamp.includes('ago') || timestamp.includes('d') || timestamp.includes('h'))) {
          return timestamp
      }
      return moment.unix(timestamp).format('YYYY-MM-DD HH:mm:ss')
    },
    getShortImageName(fullImage) {
      if (!fullImage) return '-'
      // 移除registry路径，只保留镜像名称和标签
      // 例如: registry.aliyuncs.com/google_containers/metrics-server:v0.8.1 -> metrics-server:v0.8.1
      const parts = fullImage.split('/')
      return parts[parts.length - 1] || fullImage
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
        const res = await getReplicaSetList(this.currentNamespace, this.instanceId)
        if (res.status === 200) {
          this.tableData = res.data.replicaSetList
        }
      } catch (error) {
        // Error handled by interceptor
      } finally {
        this.loading = false
      }
    },
    handleNamespaceChange() {
      this.fetchData()
    },
    async handleDetail(row) {
      try {
        const res = await getReplicaSetDetail(row.namespace, row.name, this.instanceId)
        if (res.status === 200) {
          this.detailData = JSON.stringify(res.data.replicaSetDetail, null, 2)
          this.detailVisible = true
        }
      } catch (error) {
        // Error handled
      }
    },
    handleDelete(row) {
      this.$confirm(`确认删除 ReplicaSet "${row.name}"?`, '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }).then(async () => {
        try {
          const res = await deleteReplicaSet(row.namespace, row.name, this.instanceId)
          if (res.status === 200) {
            this.$message.success('删除成功')
            this.fetchData()
          }
        } catch (error) {
          // Error handled
        }
      }).catch(() => {})
    },
    showImagesDialog(row) {
      this.currentRow = row
      this.currentImages = row.images || []
      this.imagesDialogVisible = true
    },
    copyImage(image) {
      navigator.clipboard.writeText(image).then(() => {
        this.$message.success('镜像地址已复制')
      }).catch(() => {
        this.$message.error('复制失败')
      })
    }
  }
}
</script>

<style scoped>
/* Image Cell */
.image-cell {
  display: flex;
  align-items: center;
  gap: 8px;
}

.image-icon {
  font-size: 16px;
  color: #909399;
  flex-shrink: 0;
}

.image-name {
  flex: 1;
  font-size: 13px;
  color: #606266;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-family: 'Consolas', 'Monaco', monospace;
}

.more-images-btn {
  flex-shrink: 0;
  background: linear-gradient(135deg, #409EFF, #66b1ff);
  color: white;
  border: none;
  padding: 2px 10px;
  font-size: 12px;
  border-radius: 10px;
  cursor: pointer;
  transition: all 0.3s;
  min-width: 40px;
}

.more-images-btn:hover {
  transform: scale(1.05);
  box-shadow: 0 2px 8px rgba(64, 158, 255, 0.4);
}

/* Images Dialog */
.images-list-container {
  padding: 0;
}

.images-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
  max-height: 400px;
  overflow-y: auto;
}

.image-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  background: #f5f7fa;
  border-radius: 8px;
  border: 1px solid #e4e7ed;
  transition: all 0.2s;
}

.image-item:hover {
  background: #ecf5ff;
  border-color: #409EFF;
}

.image-item-icon {
  font-size: 24px;
  color: #409EFF;
  flex-shrink: 0;
}

.image-item-content {
  flex: 1;
  min-width: 0;
}

.image-item-name {
  font-size: 14px;
  color: #303133;
  font-weight: 500;
  margin-bottom: 6px;
  word-break: break-all;
  font-family: 'Consolas', 'Monaco', monospace;
}

.image-item-tags {
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
}

.image-copy-btn {
  flex-shrink: 0;
  background: #409EFF;
  color: white;
  border: none;
}

.image-copy-btn:hover {
  background: #66b1ff;
}

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
