<template>
  <div class="page-container">
    <el-card class="page-header-card">
      <div class="page-header">
        <div class="header-left">
          <div class="header-icon-wrapper">
            <el-icon :size="24"><Collection /></el-icon>
          </div>
          <div class="header-title-wrapper">
            <h2>工作负载</h2>
            <p class="subtitle">管理 Kubernetes 集群中的各类工作负载</p>
          </div>
        </div>
      </div>
    </el-card>

    <div class="workload-content">
      <el-card class="content-card">
        <el-tabs v-model="activeTab" class="workload-tabs" @tab-click="handleTabClick">
          <el-tab-pane label="全部" name="all" disabled>
             <!-- Placeholder for All View if needed -->
          </el-tab-pane>
          <el-tab-pane label="Deployment" name="deployment">
            <DeploymentManagement ref="deploymentRef" />
          </el-tab-pane>
          <el-tab-pane label="StatefulSet" name="statefulset">
            <StatefulSetManagement ref="statefulSetRef" />
          </el-tab-pane>
          <el-tab-pane label="DaemonSet" name="daemonset">
            <DaemonSetManagement ref="daemonSetRef" />
          </el-tab-pane>
          <el-tab-pane label="Job" name="job">
            <JobManagement ref="jobRef" />
          </el-tab-pane>
          <el-tab-pane label="CronJob" name="cronjob">
            <CronJobManagement ref="cronJobRef" />
          </el-tab-pane>
        </el-tabs>
      </el-card>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { Collection } from '@element-plus/icons-vue'
import DeploymentManagement from './DeploymentManagement.vue'
import StatefulSetManagement from './StatefulSetManagement.vue'
import DaemonSetManagement from './DaemonSetManagement.vue'
import JobManagement from './JobManagement.vue'
import CronJobManagement from './CronJobManagement.vue'

const activeTab = ref('deployment')
const deploymentRef = ref(null)
const statefulSetRef = ref(null)
const daemonSetRef = ref(null)
const jobRef = ref(null)
const cronJobRef = ref(null)

const handleTabClick = (tab) => {
  // Logic to refresh data if needed
}
</script>

<style scoped>
.page-header {
  display: flex;
  align-items: center;
}
.header-left {
  display: flex;
  align-items: center;
  gap: 12px;
}
.header-icon-wrapper {
  width: 40px;
  height: 40px;
  background-color: var(--primary-color);
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
}
.header-title-wrapper h2 {
  margin: 0;
  font-size: 18px;
  color: var(--text-main);
  font-weight: 600;
}
.header-title-wrapper .subtitle {
  margin: 4px 0 0;
  font-size: 12px;
  color: var(--text-sub);
}

.workload-content {
  /* margin-top: 20px; */
}

/* Custom Tab Styles to match screenshot */
:deep(.el-tabs__header) {
  margin-bottom: 20px;
}
:deep(.el-tabs__nav-wrap::after) {
  height: 1px;
  background-color: #e4e7ed;
}
:deep(.el-tabs__item) {
  font-size: 14px;
  color: var(--text-sub);
  padding: 0 20px;
}
:deep(.el-tabs__item.is-active) {
  color: var(--primary-color);
  font-weight: 600;
}
:deep(.el-tabs__active-bar) {
  height: 2px;
}
</style>
