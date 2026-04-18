<template>
  <div class="rest-api-console">
    <!-- 顶部导航区域 -->
    <div class="console-header">
      <div class="header-left">
        <h1 class="app-title">Rest</h1>
        <a href="#" class="api-docs-link">
          API 文档
          <el-icon><Link /></el-icon>
        </a>
      </div>
      
      <div class="header-center">
        <el-dropdown trigger="click" @command="handleHistoryCommand">
          <el-button class="header-btn">
            <el-icon><Clock /></el-icon>
            查询历史
            <el-icon><ArrowDown /></el-icon>
          </el-button>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item v-for="(item, index) in queryHistory" :key="index" :command="item">
                {{ item.method }} {{ item.path }}
              </el-dropdown-item>
              <el-dropdown-item v-if="queryHistory.length === 0" disabled>
                暂无历史记录
              </el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
        
        <el-dropdown trigger="click" @command="handleSavedCommand">
          <el-button class="header-btn">
            <el-icon><Document /></el-icon>
            SAVED QUERIES
            <el-icon><ArrowDown /></el-icon>
          </el-button>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item v-for="(item, index) in savedQueries" :key="index" :command="item">
                {{ item.name }}
              </el-dropdown-item>
              <el-dropdown-item v-if="savedQueries.length === 0" disabled>
                暂无保存的查询
              </el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </div>
      
      <div class="header-right">
        <el-button class="header-btn" @click="showQueryExamples">
          <el-icon><InfoFilled /></el-icon>
          查询示例
        </el-button>
      </div>
    </div>

    <!-- 标签页区域 -->
    <div class="tabs-container">
      <div class="tabs">
        <div 
          v-for="(tab, index) in tabs" 
          :key="index"
          :class="['tab', { active: currentTabIndex === index }]"
          @click="switchTab(index)"
        >
          <span>{{ tab.name }}</span>
          <el-icon class="tab-icon edit-icon" @click.stop="renameTab(index)"><Edit /></el-icon>
          <el-icon class="tab-icon close-icon" @click.stop="closeTab(index)"><Close /></el-icon>
        </div>
        <div class="tab add-tab" @click="addNewTab">
          <el-icon><Plus /></el-icon>
        </div>
      </div>
    </div>

    <!-- 主要工作区域 -->
    <div class="workspace">
      <div v-if="currentTab" class="tab-content">
        <!-- 快捷链接 -->
        <div class="quick-links">
          <a href="#" class="quick-link" @click="pasteKibanaQuery">
            粘贴 kibana 控制台查询
          </a>
        </div>

        <!-- 请求配置区域 -->
        <div class="request-config">
          <div class="request-method-path">
            <el-select v-model="currentTab.request.method" class="method-select">
              <el-option value="GET" label="GET" />
              <el-option value="POST" label="POST" />
              <el-option value="PUT" label="PUT" />
              <el-option value="DELETE" label="DELETE" />
              <el-option value="PATCH" label="PATCH" />
              <el-option value="HEAD" label="HEAD" />
              <el-option value="OPTIONS" label="OPTIONS" />
            </el-select>
            
            <el-input 
              v-model="currentTab.request.path" 
              placeholder="请求路径"
              class="path-input"
              @keyup.enter="sendRequest"
            />
            
            <el-button type="primary" class="send-btn" @click="sendRequest">
              <el-icon><CaretRight /></el-icon>
            </el-button>
          </div>
        </div>

        <!-- 编辑器区域 -->
        <div class="editors-container">
          <!-- 请求体编辑器 -->
          <div class="editor-panel request-editor">
            <div class="editor-header">
              <span>Request Body</span>
              <div class="editor-actions">
                <el-button size="small" text @click="formatRequestBody" title="格式化JSON (Ctrl+Shift+F)">
                  <el-icon><Top /></el-icon>
                </el-button>
                <el-button size="small" text @click="compressRequestBody" title="压缩JSON">
                  <el-icon><Bottom /></el-icon>
                </el-button>
                <el-button size="small" text @click="validateRequestBody" title="验证JSON格式">
                  <el-icon><Check /></el-icon>
                </el-button>
                <el-button size="small" text @click="clearRequestBody">
                  <el-icon><Close /></el-icon>
                </el-button>
                <el-button size="small" text @click="copyRequestBody">
                  <el-icon><CopyDocument /></el-icon>
                </el-button>
                <el-button size="small" text @click="showRequestSettings">
                  <el-icon><Setting /></el-icon>
                </el-button>
              </div>
            </div>
            <div class="editor-content">
              <textarea
                v-model="currentTab.request.body"
                placeholder="输入请求体 (JSON格式)"
                class="code-editor"
                spellcheck="false"
                @keydown="handleEditorKeydown"
              ></textarea>
              <div v-if="jsonValidationMessage" class="json-validation" :class="{ 'error': !jsonValidationValid }">
                {{ jsonValidationMessage }}
              </div>
            </div>
          </div>

          <!-- 拖拽分隔线 -->
          <div class="resize-handle" @mousedown="startResize">
            ---- drag to resize ----
          </div>

          <!-- 响应编辑器 -->
          <div class="editor-panel response-editor">
            <div class="editor-header">
              <span>Response</span>
              <div class="editor-actions">
                <el-button size="small" text @click="formatResponse">
                  <el-icon><Top /></el-icon>
                </el-button>
                <el-button size="small" text @click="clearResponse">
                  <el-icon><Close /></el-icon>
                </el-button>
                <el-button size="small" text @click="copyResponse">
                  <el-icon><CopyDocument /></el-icon>
                </el-button>
              </div>
            </div>
            <div class="editor-content">
              <div v-if="currentTab.response.loading" class="loading">
                <el-icon class="is-loading"><Loading /></el-icon>
                请求中...
              </div>
              <textarea
                v-else
                v-model="currentTab.response.body"
                class="code-editor"
                spellcheck="false"
                readonly
              ></textarea>
            </div>
          </div>
        </div>

        <!-- 底部操作区域 -->
        <div class="bottom-actions">
          <div class="left-actions">
            <el-button type="primary" @click="sendRequest">
              发送请求
            </el-button>
            <el-button @click="saveQuery">
              <el-icon><Document /></el-icon>
              保存
            </el-button>
          </div>
          
          <div class="right-actions">
            <el-button @click="downloadAsJson">
              <el-icon><Download /></el-icon>
              下载为 JSON 文件
            </el-button>
          </div>
        </div>
      </div>
    </div>

    <!-- 查询示例对话框 -->
    <el-dialog v-model="queryExamplesVisible" title="查询示例" width="60%">
      <div class="examples-content">
        <div v-for="(example, index) in queryExamples" :key="index" class="example-item">
          <h4>{{ example.name }}</h4>
          <p>{{ example.description }}</p>
          <pre class="example-code">{{ example.code }}</pre>
          <el-button size="small" @click="useExample(example)">使用此示例</el-button>
        </div>
      </div>
    </el-dialog>

    <!-- 保存查询对话框 -->
    <el-dialog v-model="saveQueryVisible" title="保存查询" width="40%">
      <el-form :model="saveQueryForm" label-width="80px">
        <el-form-item label="查询名称">
          <el-input v-model="saveQueryForm.name" placeholder="输入查询名称" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input 
            v-model="saveQueryForm.description" 
            type="textarea"
            :rows="3"
            placeholder="输入查询描述"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="saveQueryVisible = false">取消</el-button>
        <el-button type="primary" @click="confirmSaveQuery">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, nextTick } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Link, Clock, ArrowDown, Document, InfoFilled, Edit, Close, Plus,
  CaretRight, Top, Bottom, Check, CopyDocument, Setting, Loading, Download
} from '@element-plus/icons-vue'
import { executeDevConsoleRequest } from '../../api/es/data.js'
import { getInstanceList } from '../../api/instance.js'

// 响应式数据
const tabs = ref([
  {
    name: 'TAB 1',
    request: {
      method: 'POST',
      path: '',
      body: ''
    },
    response: {
      loading: false,
      body: '',
      status: null,
      headers: {},
      duration: 0
    }
  }
])

const currentTabIndex = ref(0)
const queryHistory = ref([])
const savedQueries = ref([])
const queryExamplesVisible = ref(false)
const saveQueryVisible = ref(false)
const saveQueryForm = reactive({
  name: '',
  description: ''
})

// 实例相关
const instanceOptions = ref([])
const selectedInstance = ref(null)

// 编辑器拖拽相关
const isResizing = ref(false)
const leftPanelWidth = ref(50)

// JSON验证相关
const jsonValidationMessage = ref('')
const jsonValidationValid = ref(true)

// 查询示例数据
const queryExamples = ref([
  {
    name: '获取集群健康状态',
    description: '检查Elasticsearch集群的整体健康状态',
    code: `GET /_cluster/health`
  },
  {
    name: '获取所有索引',
    description: '列出集群中的所有索引',
    code: `GET /_cat/indices?format=json`
  },
  {
    name: '搜索文档',
    description: '在指定索引中搜索文档',
    code: `POST /my_index/_search
{
  "query": {
    "match_all": {}
  },
  "size": 10
}`
  },
  {
    name: '创建索引',
    description: '创建新的索引并设置分片配置',
    code: `PUT /my_index
{
  "settings": {
    "number_of_shards": 1,
    "number_of_replicas": 1
  }
}`
  },
  {
    name: '聚合查询',
    description: '使用聚合进行数据分析',
    code: `POST /my_index/_search
{
  "size": 0,
  "aggs": {
    "group_by_field": {
      "terms": {
        "field": "field_name"
      }
    }
  }
}`
  }
])

// 计算属性
const currentTab = computed(() => tabs.value[currentTabIndex.value])

// 获取实例列表
const fetchInstances = async () => {
  try {
    const response = await getInstanceList({ 
      type_name: 'elasticsearch',
      page: 1,
      page_size: 100
    })
    
    if (response && response.data && response.data.list && response.data.list.data) {
      const instances = response.data.list.data
      instanceOptions.value = instances.map(instance => ({
        label: instance.name,
        value: instance.id
      }))
      
      if (instances.length > 0 && !selectedInstance.value) {
        selectedInstance.value = instances[0].id
      }
    }
  } catch (error) {
    console.error('获取实例列表失败:', error)
  }
}

// 标签页管理
const addNewTab = () => {
  const newTab = {
    name: `TAB ${tabs.value.length + 1}`,
    request: {
      method: 'POST',
      path: '',
      body: ''
    },
    response: {
      loading: false,
      body: '',
      status: null,
      headers: {},
      duration: 0
    }
  }
  tabs.value.push(newTab)
  currentTabIndex.value = tabs.value.length - 1
}

const closeTab = (index) => {
  if (tabs.value.length === 1) {
    ElMessage.warning('至少需要保留一个标签页')
    return
  }
  
  tabs.value.splice(index, 1)
  if (currentTabIndex.value >= tabs.value.length) {
    currentTabIndex.value = tabs.value.length - 1
  }
}

const switchTab = (index) => {
  currentTabIndex.value = index
}

const renameTab = (index) => {
  ElMessageBox.prompt('请输入标签页名称', '重命名', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    inputValue: tabs.value[index].name
  }).then(({ value }) => {
    tabs.value[index].name = value
  }).catch(() => {})
}

// 发送请求
const sendRequest = async () => {
  if (!selectedInstance.value) {
    ElMessage.warning('请先选择一个Elasticsearch实例')
    return
  }

  if (!currentTab.value.request.path) {
    ElMessage.warning('请输入请求路径')
    return
  }

  currentTab.value.response.loading = true
  currentTab.value.response.body = ''

  try {
    const requestData = {
      instance_id: selectedInstance.value,
      method: currentTab.value.request.method,
      path: currentTab.value.request.path,
      body: currentTab.value.request.body || '',
      params: {}
    }

    const response = await executeDevConsoleRequest(requestData)
    
    if (response.data) {
      const result = response.data.response
      currentTab.value.response = {
        loading: false,
        body: JSON.stringify(result.body, null, 2),
        status: result.status,
        headers: result.headers,
        duration: result.duration
      }

      // 添加到历史记录
      addToHistory({
        method: currentTab.value.request.method,
        path: currentTab.value.request.path,
        body: currentTab.value.request.body,
        timestamp: new Date(),
        response: result
      })
    }
  } catch (error) {
    currentTab.value.response.loading = false
    currentTab.value.response.body = JSON.stringify({
      error: error.response?.data?.message || error.message
    }, null, 2)
    ElMessage.error('请求失败: ' + (error.response?.data?.message || error.message))
  }
}

// 添加到历史记录
const addToHistory = (query) => {
  queryHistory.value.unshift(query)
  if (queryHistory.value.length > 50) {
    queryHistory.value = queryHistory.value.slice(0, 50)
  }
}

// 历史记录处理
const handleHistoryCommand = (item) => {
  currentTab.value.request.method = item.method
  currentTab.value.request.path = item.path
  currentTab.value.request.body = item.body || ''
}

// 保存的查询处理
const handleSavedCommand = (item) => {
  currentTab.value.request.method = item.method
  currentTab.value.request.path = item.path
  currentTab.value.request.body = item.body || ''
}

// 显示查询示例
const showQueryExamples = () => {
  queryExamplesVisible.value = true
}

// 使用示例
const useExample = (example) => {
  const lines = example.code.split('\n')
  if (lines.length > 0) {
    const firstLine = lines[0].trim()
    const match = firstLine.match(/(GET|POST|PUT|DELETE|PATCH|HEAD|OPTIONS)\s+(.+)/)
    if (match) {
      currentTab.value.request.method = match[1]
      currentTab.value.request.path = match[2]
      currentTab.value.request.body = lines.slice(1).join('\n').trim()
    }
  }
  queryExamplesVisible.value = false
}

// 粘贴Kibana查询
const pasteKibanaQuery = () => {
  ElMessageBox.prompt('请粘贴Kibana控制台查询', '粘贴查询', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    inputType: 'textarea',
    inputPlaceholder: 'GET /_search\n{\n  "query": {\n    "match_all": {}\n  }\n}'
  }).then(({ value }) => {
    if (value) {
      const lines = value.split('\n')
      if (lines.length > 0) {
        const firstLine = lines[0].trim()
        const match = firstLine.match(/(GET|POST|PUT|DELETE|PATCH|HEAD|OPTIONS)\s+(.+)/)
        if (match) {
          currentTab.value.request.method = match[1]
          currentTab.value.request.path = match[2]
          currentTab.value.request.body = lines.slice(1).join('\n').trim()
        }
      }
    }
  }).catch(() => {})
}

// 保存查询
const saveQuery = () => {
  saveQueryForm.name = ''
  saveQueryForm.description = ''
  saveQueryVisible.value = true
}

const confirmSaveQuery = () => {
  if (!saveQueryForm.name) {
    ElMessage.warning('请输入查询名称')
    return
  }

  const query = {
    name: saveQueryForm.name,
    description: saveQueryForm.description,
    method: currentTab.value.request.method,
    path: currentTab.value.request.path,
    body: currentTab.value.request.body,
    timestamp: new Date()
  }

  savedQueries.value.unshift(query)
  if (savedQueries.value.length > 20) {
    savedQueries.value = savedQueries.value.slice(0, 20)
  }

  saveQueryVisible.value = false
  ElMessage.success('查询保存成功')
}

// 编辑器操作
const formatRequestBody = () => {
  try {
    if (currentTab.value.request.body) {
      const parsed = JSON.parse(currentTab.value.request.body)
      currentTab.value.request.body = JSON.stringify(parsed, null, 2)
      jsonValidationMessage.value = 'JSON格式正确'
      jsonValidationValid.value = true
      ElMessage.success('JSON格式化成功')
    }
  } catch (e) {
    const errorMessage = e && e.message ? e.message : '未知错误'
    jsonValidationMessage.value = `JSON格式错误: ${errorMessage}`
    jsonValidationValid.value = false
    ElMessage.error('JSON格式错误，无法格式化')
  }
}

const compressRequestBody = () => {
  try {
    if (currentTab.value.request.body) {
      const parsed = JSON.parse(currentTab.value.request.body)
      currentTab.value.request.body = JSON.stringify(parsed)
      jsonValidationMessage.value = 'JSON压缩成功'
      jsonValidationValid.value = true
    }
  } catch (e) {
    const errorMessage = e && e.message ? e.message : '未知错误'
    jsonValidationMessage.value = `JSON格式错误: ${errorMessage}`
    jsonValidationValid.value = false
    ElMessage.error('JSON格式错误，无法压缩')
  }
}

const validateRequestBody = () => {
  if (!currentTab.value.request.body || !currentTab.value.request.body.trim()) {
    jsonValidationMessage.value = '请输入JSON内容'
    jsonValidationValid.value = false
    return
  }
  
  try {
    JSON.parse(currentTab.value.request.body)
    jsonValidationMessage.value = 'JSON格式正确'
    jsonValidationValid.value = true
    ElMessage.success('JSON格式验证通过')
  } catch (e) {
    const errorMessage = e && e.message ? e.message : '未知错误'
    jsonValidationMessage.value = `JSON格式错误: ${errorMessage}`
    jsonValidationValid.value = false
  }
}

const handleEditorKeydown = (event) => {
  try {
    // Ctrl+Shift+F 格式化JSON
    if (event.ctrlKey && event.shiftKey && event.key === 'F') {
      event.preventDefault()
      formatRequestBody()
    }
    // Ctrl+Shift+C 压缩JSON
    else if (event.ctrlKey && event.shiftKey && event.key === 'C') {
      event.preventDefault()
      compressRequestBody()
    }
    // Ctrl+Shift+V 验证JSON
    else if (event.ctrlKey && event.shiftKey && event.key === 'V') {
      event.preventDefault()
      validateRequestBody()
    }
  } catch (error) {
    console.error('快捷键处理错误:', error)
  }
}

const clearRequestBody = () => {
  currentTab.value.request.body = ''
}

const copyRequestBody = () => {
  try {
    if (navigator.clipboard && navigator.clipboard.writeText) {
      navigator.clipboard.writeText(currentTab.value.request.body || '')
        .then(() => ElMessage.success('已复制到剪贴板'))
        .catch(() => {
          fallbackCopyToClipboard(currentTab.value.request.body || '')
        })
    } else {
      fallbackCopyToClipboard(currentTab.value.request.body || '')
    }
  } catch (error) {
    ElMessage.error('复制失败')
  }
}

const formatResponse = () => {
  try {
    if (currentTab.value.response.body) {
      const parsed = JSON.parse(currentTab.value.response.body)
      currentTab.value.response.body = JSON.stringify(parsed, null, 2)
    }
  } catch (e) {
    const errorMessage = e && e.message ? e.message : '未知错误'
    ElMessage.error('JSON格式错误，无法格式化')
  }
}

const clearResponse = () => {
  currentTab.value.response.body = ''
}

const copyResponse = () => {
  try {
    if (navigator.clipboard && navigator.clipboard.writeText) {
      navigator.clipboard.writeText(currentTab.value.response.body || '')
        .then(() => ElMessage.success('已复制到剪贴板'))
        .catch(() => {
          fallbackCopyToClipboard(currentTab.value.response.body || '')
        })
    } else {
      fallbackCopyToClipboard(currentTab.value.response.body || '')
    }
  } catch (error) {
    ElMessage.error('复制失败')
  }
}

const fallbackCopyToClipboard = (text) => {
  try {
    const textArea = document.createElement('textarea')
    textArea.value = text
    textArea.style.position = 'fixed'
    textArea.style.left = '-999999px'
    textArea.style.top = '-999999px'
    document.body.appendChild(textArea)
    textArea.focus()
    textArea.select()
    
    const successful = document.execCommand('copy')
    document.body.removeChild(textArea)
    
    if (successful) {
      ElMessage.success('已复制到剪贴板')
    } else {
      ElMessage.error('复制失败')
    }
  } catch (error) {
    ElMessage.error('复制失败')
  }
}

const showRequestSettings = () => {
  ElMessage.info('请求设置功能待实现')
}

const downloadAsJson = () => {
  try {
    if (!currentTab.value.response.body) {
      ElMessage.warning('没有可下载的响应数据')
      return
    }

    const blob = new Blob([currentTab.value.response.body], { type: 'application/json' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `response_${Date.now()}.json`
    a.style.display = 'none'
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    URL.revokeObjectURL(url)
    ElMessage.success('文件下载成功')
  } catch (error) {
    ElMessage.error('文件下载失败')
  }
}

// 拖拽调整编辑器宽度
const startResize = (e) => {
  isResizing.value = true
  document.addEventListener('mousemove', handleResize)
  document.addEventListener('mouseup', stopResize)
}

const handleResize = (e) => {
  if (!isResizing.value) return
  
  const container = document.querySelector('.editors-container')
  if (!container) return
  
  const containerWidth = container.offsetWidth
  const newLeftWidth = (e.clientX - container.offsetLeft) / containerWidth * 100
  
  if (newLeftWidth >= 20 && newLeftWidth <= 80) {
    leftPanelWidth.value = newLeftWidth
  }
}

const stopResize = () => {
  isResizing.value = false
  document.removeEventListener('mousemove', handleResize)
  document.removeEventListener('mouseup', stopResize)
}

// 组件挂载
fetchInstances()
</script>

<style scoped>
.rest-api-console {
  height: 100vh;
  display: flex;
  flex-direction: column;
  background-color: #ffffff;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
}

/* 顶部导航区域 */
.console-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 20px;
  border-bottom: 1px solid #e8e8e8;
  background-color: #fafafa;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 20px;
}

.app-title {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
  color: #333;
}

.api-docs-link {
  display: flex;
  align-items: center;
  gap: 4px;
  color: #007bff;
  text-decoration: none;
  font-size: 14px;
}

.api-docs-link:hover {
  color: #0056b3;
}

.header-center {
  display: flex;
  gap: 10px;
}

.header-btn {
  background-color: transparent;
  border: 1px solid #ddd;
  color: #666;
  font-size: 14px;
}

.header-btn:hover {
  background-color: #f5f5f5;
  border-color: #007bff;
  color: #007bff;
}

/* 标签页区域 */
.tabs-container {
  background-color: #ffffff;
  border-bottom: 1px solid #e8e8e8;
}

.tabs {
  display: flex;
  padding: 0 20px;
}

.tab {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 16px;
  cursor: pointer;
  border-bottom: 2px solid transparent;
  color: #666;
  font-size: 14px;
  position: relative;
}

.tab:hover {
  color: #333;
}

.tab.active {
  color: #007bff;
  border-bottom-color: #007bff;
}

.tab-icon {
  opacity: 0;
  font-size: 12px;
}

.tab:hover .tab-icon {
  opacity: 1;
}

.edit-icon:hover {
  color: #007bff;
}

.close-icon:hover {
  color: #dc3545;
}

.add-tab {
  color: #999;
  border: none;
}

.add-tab:hover {
  color: #007bff;
}

/* 工作区域 */
.workspace {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.tab-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  padding: 20px;
}

/* 快捷链接 */
.quick-links {
  margin-bottom: 16px;
}

.quick-link {
  color: #007bff;
  text-decoration: none;
  font-size: 14px;
}

.quick-link:hover {
  text-decoration: underline;
}

/* 请求配置区域 */
.request-config {
  margin-bottom: 20px;
}

.request-method-path {
  display: flex;
  gap: 10px;
  align-items: center;
}

.method-select {
  width: 160px;
}

.method-select :deep(.el-input__inner) {
  color: #ff851b;
  font-weight: 500;
}

.path-input {
  flex: 1;
}

.send-btn {
  padding: 8px 12px;
}

/* 编辑器区域 */
.editors-container {
  display: flex;
  flex: 1;
  min-height: 300px;
  border: 1px solid #e8e8e8;
  border-radius: 4px;
  overflow: hidden;
}

.editor-panel {
  display: flex;
  flex-direction: column;
}

.request-editor {
  flex: 1;
}

.response-editor {
  flex: 1;
}

.editor-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 12px;
  background-color: #eaf6ff;
  border-bottom: 1px solid #e8e8e8;
  font-size: 13px;
  font-weight: 500;
}

.editor-actions {
  display: flex;
  gap: 4px;
}

.editor-content {
  flex: 1;
  position: relative;
}

.code-editor {
  width: 100%;
  height: calc(100% - 24px);
  border: none;
  outline: none;
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
  font-size: 13px;
  line-height: 1.5;
  padding: 12px;
  resize: none;
  background-color: #ffffff;
  color: #333;
}

.code-editor[readonly] {
  background-color: #f8f9fa;
  color: #666;
}

.json-validation {
  height: 24px;
  padding: 4px 12px;
  font-size: 12px;
  background-color: #f0f9ff;
  color: #059669;
  border-top: 1px solid #e0f2fe;
}

.json-validation.error {
  background-color: #fef2f2;
  color: #dc2626;
  border-top-color: #fee2e2;
}

.loading {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100%;
  color: #666;
  font-size: 14px;
  gap: 8px;
}

/* 拖拽分隔线 */
.resize-handle {
  width: 6px;
  background-color: #f0f0f0;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: col-resize;
  font-size: 10px;
  color: #999;
  writing-mode: vertical-rl;
  text-orientation: mixed;
  user-select: none;
}

.resize-handle:hover {
  background-color: #e0e0e0;
}

/* 底部操作区域 */
.bottom-actions {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 20px;
  padding-top: 16px;
  border-top: 1px solid #e8e8e8;
}

.left-actions {
  display: flex;
  gap: 10px;
}

.right-actions {
  display: flex;
  gap: 10px;
}

/* 对话框样式 */
.examples-content {
  max-height: 400px;
  overflow-y: auto;
}

.example-item {
  margin-bottom: 20px;
  padding-bottom: 20px;
  border-bottom: 1px solid #e8e8e8;
}

.example-item:last-child {
  border-bottom: none;
}

.example-item h4 {
  margin: 0 0 8px 0;
  color: #333;
}

.example-item p {
  margin: 0 0 12px 0;
  color: #666;
  font-size: 14px;
}

.example-code {
  background-color: #f8f9fa;
  padding: 12px;
  border-radius: 4px;
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
  font-size: 12px;
  overflow-x: auto;
  white-space: pre-wrap;
  margin-bottom: 12px;
}
</style>