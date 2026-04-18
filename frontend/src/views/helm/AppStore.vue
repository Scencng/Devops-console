<template>
  <div class="page-container">
    <el-card class="page-header-card">
       <div class="page-header">
         <div class="header-left">
           <div class="header-icon-wrapper">
             <el-icon :size="24"><Shop /></el-icon>
           </div>
           <div class="header-title-wrapper">
             <h2>应用商店</h2>
             <p class="subtitle">浏览和安装 Helm Charts</p>
           </div>
         </div>
       </div>
    </el-card>

    <el-card class="content-card">
       <div class="filter-row flex gap-4 mb-6">
           <el-select v-model="selectedRepo" placeholder="选择仓库" @change="fetchCharts" class="w-56">
             <el-option label="全部仓库" :value="0" />
             <el-option
               v-for="repo in repoList"
               :key="repo.id"
               :label="repo.name"
               :value="repo.id"
             />
           </el-select>
           <el-input
             v-model="searchKeyword"
             placeholder="搜索应用..."
             class="w-80"
             clearable
             @input="handleSearch"
             :prefix-icon="Search"
           />
       </div>

       <div v-loading="loading" class="charts-container min-h-[400px]">
          <el-empty v-if="chartList.length === 0 && !loading" description="暂无应用">
             <el-button type="primary" @click="$router.push('/helm/repos')">前往添加仓库</el-button>
          </el-empty>

          <div v-else class="charts-grid">
              <div
                v-for="chart in chartList"
                :key="chart.id"
                class="chart-card-item"
                @click="handleChartClick(chart)"
              >
                  <div class="flex items-start justify-between mb-3">
                     <div class="chart-icon-box">
                        <img v-if="chart.icon" :src="chart.icon" :alt="chart.name" @error="handleImageError" />
                        <el-icon v-else><Box /></el-icon>
                     </div>
                     <el-tag size="small" effect="plain">{{ chart.repo_name }}</el-tag>
                  </div>
                  <div>
                      <h3 class="chart-name" :title="chart.name">{{ chart.name }}</h3>
                      <p class="chart-desc">{{ chart.description || '暂无描述' }}</p>
                  </div>
                  <div class="chart-footer">
                      <span class="version-tag">v{{ chart.version }}</span>
                      <el-button type="primary" link size="small">安装</el-button>
                  </div>
              </div>
          </div>

          <!-- Pagination -->
          <div v-if="total > 0" class="flex justify-center mt-8">
            <el-pagination
              v-model:current-page="currentPage"
              v-model:page-size="pageSize"
              :total="total"
              :page-sizes="[12, 24, 48, 96]"
              layout="total, sizes, prev, pager, next, jumper"
              @current-change="fetchCharts"
              @size-change="handleSizeChange"
            />
          </div>
       </div>
    </el-card>

    <!-- Install Dialog -->
    <el-dialog
      v-model="showInstallDialog"
      title="安装 Helm Chart"
      width="700px"
      destroy-on-close
      @opened="handleDialogOpened"
    >
      <el-form :model="installForm" ref="installFormRef" label-position="top" @submit.prevent>
         <el-row :gutter="20">
             <el-col :span="12">
                <el-form-item label="Release 名称" prop="release_name" required>
                   <el-input v-model="installForm.release_name" placeholder="例如：my-app" />
                </el-form-item>
             </el-col>
             <el-col :span="12">
                <el-form-item label="命名空间" prop="namespace" required>
                   <el-input v-model="installForm.namespace" placeholder="default" />
                </el-form-item>
             </el-col>
             <el-col :span="24">
                <el-form-item label="Chart 版本">
                   <el-select v-model="installForm.chart_version" placeholder="选择版本" class="w-full">
                      <el-option
                        v-for="version in chartVersions"
                        :key="version.version"
                        :label="`${version.version} (App: ${version.app_version || 'N/A'})`"
                        :value="version.version"
                      />
                   </el-select>
                </el-form-item>
             </el-col>
             <el-col :span="24">
                <el-form-item label="自定义 Values">
                   <div class="flex justify-between items-center mb-2 w-full">
                       <el-upload
                         class="upload-btn-fix"
                         action=""
                         :auto-upload="false"
                         :show-file-list="false"
                         accept=".yaml,.yml"
                         :on-change="handleLocalFileUpload"
                       >
                         <el-button size="small" type="primary" :icon="Upload">上传本地配置</el-button>
                       </el-upload>
                      <el-button size="small" @click="loadDefaultValues" link type="primary" :loading="loadingValues">加载官方示例</el-button>
                   </div>
                   <el-input
                     v-model="customValues"
                     type="textarea"
                     :rows="12"
                     placeholder="输入 YAML 格式的 values"
                     class="values-textarea"
                   />
                </el-form-item>
             </el-col>
         </el-row>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="showInstallDialog = false">取消</el-button>
          <el-button type="primary" @click="handleInstall" :loading="installing">安装</el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Search, Box, Shop, Upload } from '@element-plus/icons-vue'
import { getHelmRepos } from '@/api/helm'
import { getHelmCharts, getChartVersions, installHelmChart, getChartDefaultValues } from '@/api/helm'
import { getSelectedInstanceId } from '@/stores/instanceStore'

const loading = ref(false)
const installing = ref(false)
const loadingValues = ref(false)
const repoList = ref([])
const chartList = ref([])
const chartVersions = ref([])
const selectedRepo = ref(0)
const searchKeyword = ref('')
const currentPage = ref(1)
const pageSize = ref(24)
const total = ref(0)
const showInstallDialog = ref(false)
const selectedChart = ref(null)

const installForm = ref({
  release_name: '',
  namespace: 'default',
  chart_version: '',
  chart_name: '',
  repo_name: '',
  chart_url: '',
  instance_id: 0
})

const customValues = ref('')
const installFormRef = ref()

// 获取仓库列表
const fetchRepos = async () => {
  try {
    const res = await getHelmRepos()
    repoList.value = res.data.repoList || []
  } catch (error) {
    console.error(error)
  }
}

// 获取 Chart 列表
const fetchCharts = async () => {
  loading.value = true
  try {
    const params = {
      page: currentPage.value,
      page_size: pageSize.value
    }
    if (selectedRepo.value) params.repo_id = selectedRepo.value
    if (searchKeyword.value) params.keyword = searchKeyword.value

    const res = await getHelmCharts(params)
    chartList.value = res.data.data.chartList || []
    total.value = res.data.data.total || 0
  } catch (error) {
    console.error(error)
    ElMessage.error('获取应用列表失败')
  } finally {
    loading.value = false
  }
}

const loadDefaultValues = async () => {
  if (!selectedChart.value) {
    ElMessage.warning('未选择 Chart')
    return
  }
  
  loadingValues.value = true
  try {
    const res = await getChartDefaultValues(
      selectedChart.value.name,
      selectedChart.value.repo_id,
      installForm.value.chart_version
    )
    customValues.value = res.data.values || ''
    ElMessage.success('已加载官方默认配置')
  } catch (error) {
    console.error(error)
    ElMessage.error('加载失败，可能由于网络超时或官方仓库无响应')
  } finally {
    loadingValues.value = false
  }
}

// 处理本地 YAML 文件上传
const handleLocalFileUpload = (file) => {
  const reader = new FileReader()
  reader.onload = (e) => {
    try {
      customValues.value = e.target.result
      ElMessage.success('本地配置已加载')
    } catch (err) {
      ElMessage.error('文件读取失败')
    }
  }
  reader.readAsText(file.raw)
}

// 搜索防抖
let searchTimer = null
const handleSearch = () => {
  clearTimeout(searchTimer)
  searchTimer = setTimeout(() => {
    currentPage.value = 1
    fetchCharts()
  }, 500)
}

const handleSizeChange = () => {
  currentPage.value = 1
  fetchCharts()
}

// 点击 Chart 卡片
const handleChartClick = async (chart) => {
  selectedChart.value = chart
  
  // 获取版本列表
  try {
    const res = await getChartVersions(chart.name, chart.repo_id)
    const versions = res.data.versions || []
    
    // Sort versions descending (latest first) based on semantic versioning
    chartVersions.value = versions.sort((a, b) => {
      const vA = a.version.split('.').map(n => parseInt(n, 10) || 0)
      const vB = b.version.split('.').map(n => parseInt(n, 10) || 0)
      for (let i = 0; i < Math.max(vA.length, vB.length); i++) {
        const numA = vA[i] || 0
        const numB = vB[i] || 0
        if (numA !== numB) {
          return numB - numA // Descending
        }
      }
      return 0
    })
  } catch (error) {
    console.error(error)
    chartVersions.value = []
  }

  // 获取最新的版本号 (排序后的第一个)
  const latestVersion = chartVersions.value.length > 0 ? chartVersions.value[0].version : chart.version

  // 初始化安装表单
  installForm.value = {
    release_name: chart.name.toLowerCase(),
    namespace: 'default',
    chart_version: latestVersion,
    chart_name: chart.name,
    repo_name: chart.repo_name,
    instance_id: getSelectedInstanceId(),
    chart_url: chart.chart_url
  }
  customValues.value = ''
  showInstallDialog.value = true
}

const handleInstall = async () => {
  const instanceId = getSelectedInstanceId()
  if (!instanceId) {
    ElMessage.warning('请先选择K8s实例')
    return
  }

  installing.value = true
  try {
    const data = { ...installForm.value }
    if (customValues.value.trim()) {
      data.values = customValues.value
    }

    await installHelmChart(data, instanceId)
    ElMessage.success('安装成功')
    showInstallDialog.value = false
  } catch (error) {
    console.error(error)
    ElMessage.error('安装失败')
  } finally {
    installing.value = false
  }
}

const handleImageError = (e) => {
  e.target.style.display = 'none'
  // Show fallback icon logic handled in template by adjacent el-icon
}

const handleDialogOpened = () => {
    // any dialog open logic
}

onMounted(() => {
  fetchRepos()
  fetchCharts()
})
</script>

<style scoped>
.header-icon-wrapper {
  width: 40px;
  height: 40px;
  border-radius: 8px;
  background: var(--primary-color);
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
}

.w-56 { width: 14rem; }
.w-80 { width: 20rem; }
.w-full { width: 100%; }
.flex { display: flex; }
.gap-4 { gap: 16px; }
.mb-6 { margin-bottom: 24px; }
.mb-3 { margin-bottom: 12px; }
.mb-2 { margin-bottom: 8px; }
.mb-1 { margin-bottom: 4px; }
.mr-2 { margin-right: 8px; }
.items-center { align-items: center; }
.items-start { align-items: flex-start; }
.justify-between { justify-content: space-between; }
.justify-center { justify-content: center; }
.justify-end { justify-content: flex-end; }
.mt-8 { margin-top: 32px; }
.min-h-\[400px\] { min-height: 400px; }

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.header-left { display: flex; align-items: center; gap: 12px; }
.header-title-wrapper h2 { margin: 0; font-size: 18px; color: var(--text-main); }
.header-title-wrapper .subtitle { margin: 0; font-size: 12px; color: var(--text-sub); }

.charts-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
    gap: 24px;
}

.chart-card-item {
    background: #fff;
    border: 1px solid var(--border-color);
    border-radius: 12px;
    padding: 20px;
    cursor: pointer;
    transition: all 0.3s ease;
    display: flex;
    flex-direction: column;
    gap: 12px;
}

.chart-card-item:hover {
    box-shadow: var(--shadow-md);
    transform: translateY(-4px);
    border-color: var(--primary-color-light);
}

.chart-icon-box {
    width: 48px;
    height: 48px;
    border-radius: 8px;
    background: #f5f7fa;
    display: flex;
    align-items: center;
    justify-content: center;
    overflow: hidden;
    color: var(--primary-color);
    font-size: 24px;
}

.chart-icon-box img {
    width: 100%;
    height: 100%;
    object-fit: contain;
}

.chart-name {
    margin: 0 0 4px 0;
    font-size: 16px;
    font-weight: 600;
    color: var(--text-main);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
}

.chart-desc {
    margin: 0;
    font-size: 13px;
    color: var(--text-sub);
    line-height: 1.5;
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
    height: 40px; 
}

.chart-footer {
    margin-top: auto;
    padding-top: 12px;
    border-top: 1px solid #f0f0f0;
    display: flex;
    justify-content: space-between;
    align-items: center;
}

.values-textarea :deep(.el-textarea__inner) {
    font-family: 'Consolas', 'Monaco', monospace;
    font-size: 13px;
}

.upload-btn-fix {
  display: inline-block;
}
.upload-btn-fix :deep(.el-upload) {
  display: block;
}
</style>
