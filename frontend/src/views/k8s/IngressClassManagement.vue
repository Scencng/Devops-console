<template>
  <div class="ingressclass-management">
    <el-card class="page-header-card cyber-card">
      <div class="page-header cyber-header">
        <div>
          <h2>IngressClass 管理</h2>
          <p>查看 Kubernetes IngressClass资源</p>
        </div>
      </div>
    </el-card>

    <el-card class="content-card cyber-card">
      <div class="autoops-table-wrapper">
        <el-table 
          :data="ingressClassList" 
          style="width: 100%" class="autoops-table" 
          v-loading="loading"
          element-loading-background="rgba(0, 0, 0, 0.5)"
          height="calc(100vh - 280px)"
          :empty-text="loading ? '加载中...' : '暂无数据'"
        >
          <el-table-column prop="name" label="名称" min-width="200" show-overflow-tooltip />
          <el-table-column prop="controller" label="Controller" min-width="250" show-overflow-tooltip />
          <el-table-column label="默认" width="100" align="center">
            <template #default="scope">
              <el-tag v-if="scope.row.isDefault" type="success" size="small">是</el-tag>
              <span v-else style="color: #999;">-</span>
            </template>
          </el-table-column>
          <el-table-column prop="parameters" label="Parameters" width="200" show-overflow-tooltip>
            <template #default="scope">
              <span v-if="scope.row.parameters">{{ scope.row.parameters }}</span>
              <span v-else style="color: #999;">-</span>
            </template>
          </el-table-column>
          <el-table-column label="创建时间" width="170">
            <template #default="scope">
              {{ formatTimestamp(scope.row.age) }}
            </template>
          </el-table-column>
          <el-table-column label="操作" width="120" fixed="right">
            <template #default="scope">
              <el-button link type="primary" size="small" @click="handleViewDetail(scope.row)" v-show="permStore.hasPerm('k8s:ingressclass:handleviewdetail')" >详情</el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </el-card>

    <!-- Detail Dialog -->
    <el-dialog v-model="showDetailDialog" title="IngressClass 详情" width="800px">
      <div v-if="currentIngressClass">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="名称">{{ currentIngressClass.metadata.name }}</el-descriptions-item>
          <el-descriptions-item label="Controller">{{ currentIngressClass.spec.controller }}</el-descriptions-item>
          <el-descriptions-item label="默认">
            <el-tag v-if="isDefaultClass(currentIngressClass)" type="success" size="small">是</el-tag>
            <span v-else>否</span>
          </el-descriptions-item>
          <el-descriptions-item label="创建时间">{{ formatTimestamp(currentIngressClass.metadata.creationTimestamp) }}</el-descriptions-item>
        </el-descriptions>

        <el-divider content-position="left" v-if="currentIngressClass.spec.parameters">Parameters</el-divider>
        <el-descriptions v-if="currentIngressClass.spec.parameters" :column="1" border style="margin-top: 10px;">
          <el-descriptions-item label="Kind">{{ currentIngressClass.spec.parameters.kind }}</el-descriptions-item>
          <el-descriptions-item label="Name">{{ currentIngressClass.spec.parameters.name }}</el-descriptions-item>
          <el-descriptions-item label="APIGroup" v-if="currentIngressClass.spec.parameters.apiGroup">{{ currentIngressClass.spec.parameters.apiGroup }}</el-descriptions-item>
        </el-descriptions>
      </div>
    </el-dialog>
  </div>
</template>

<script setup>
import { usePermissionStore } from '@/stores/permissionStore.js'
const permStore = usePermissionStore()

import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { getIngressClassList, getIngressClassDetail } from '@/api/k8s/ingressclass'
import { getSelectedInstanceId } from '@/stores/instanceStore'
import dayjs from 'dayjs'

const loading = ref(false)
const showDetailDialog = ref(false)
const ingressClassList = ref([])
const currentIngressClass = ref(null)

const formatTimestamp = (timestamp) => {
  if (typeof timestamp === 'number') {
    return dayjs.unix(timestamp).format('YYYY-MM-DD HH:mm:ss')
  }
  return dayjs(timestamp).format('YYYY-MM-DD HH:mm:ss')
}

const isDefaultClass = (ingressClass) => {
  if (!ingressClass.metadata.annotations) return false
  return ingressClass.metadata.annotations['ingressclass.kubernetes.io/is-default-class'] === 'true'
}

const fetchData = async () => {
  const instanceId = getSelectedInstanceId()
  if (!instanceId) return
  
  loading.value = true
  try {
    const res = await getIngressClassList(instanceId)
    if (res.status === 200) {
      ingressClassList.value = res.data.ingressClassList
    }
  } catch (error) {
    console.error(error)
    ElMessage.error('获取IngressClass列表失败')
  } finally {
    loading.value = false
  }
}

const handleViewDetail = async (row) => {
  const instanceId = getSelectedInstanceId()
  try {
    const res = await getIngressClassDetail(row.name, instanceId)
    if (res.status === 200) {
      currentIngressClass.value = res.data.ingressClassDetail
      showDetailDialog.value = true
    }
  } catch (error) {
    console.error(error)
    ElMessage.error('获取详情失败')
  }
}

onMounted(() => {
  fetchData()
})
</script>

<style scoped>
.page-header-card {
  margin-bottom: 0px !important;
}
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.ingressclass-management {
  padding: 20px;
}
</style>
