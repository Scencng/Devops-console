<template>
  <div class="page-container">
    <el-card class="page-header-card">
       <div class="page-header">
         <div class="header-left">
           <el-button @click="handleBack" link class="mr-2">
             <el-icon><ArrowLeft /></el-icon>
           </el-button>
           <div class="header-title-wrapper">
             <h2>{{ isEdit ? '编辑实例' : '添加实例' }}</h2>
             <p class="subtitle">配置连接信息与认证方式</p>
           </div>
         </div>
         <div class="header-right">
            <el-button @click="handleCancel">取消</el-button>
            <el-tooltip :disabled="!!formData.id" content="新建实例请先点击保存，之后再进行测试连接" placement="bottom">
              <span class="inline-block mr-3">
                <el-button :disabled="!formData.id" :loading="testing" @click="handleTestConnection">
                  <el-icon><Connection /></el-icon> 测试连接
                </el-button>
              </span>
            </el-tooltip>
            <el-button type="primary" @click="handleSave" :loading="saving">
              <el-icon><Check /></el-icon> 保存
            </el-button>
         </div>
       </div>
    </el-card>

    <div class="form-container max-w-4xl mx-auto">
       <el-form
          ref="formRef"
          :model="formData"
          :rules="formRules"
          label-position="top"
          class="instance-form-content"
       >
          <!-- Basic Info -->
          <el-card class="content-card mb-20 p-4">
             <template #header>
                <div class="font-bold flex items-center gap-2">
                   <el-icon><InfoFilled /></el-icon> 基本信息
                </div>
             </template>
             <el-row :gutter="24">
                <el-col :span="12">
                   <el-form-item label="实例类型" prop="instanceType">
                      <el-select
                        v-model="formData.instanceType"
                        placeholder="请选择实例类型"
                        class="w-full"
                        @change="handleTypeChange"
                      >
                         <el-option
                           v-for="type in instanceTypes"
                           :key="type.value"
                           :label="type.label"
                           :value="type.value"
                         >
                            <div class="flex items-center gap-2">
                               <el-icon :color="type.color"><component :is="type.icon" /></el-icon>
                               <span>{{ type.label }}</span>
                            </div>
                         </el-option>
                      </el-select>
                   </el-form-item>
                </el-col>
                <el-col :span="12">
                   <el-form-item label="实例名称" prop="name">
                      <el-input v-model="formData.name" placeholder="请输入实例名称" clearable />
                   </el-form-item>
                </el-col>
                <el-col :span="24">
                   <el-form-item label="实例地址" prop="address">
                      <el-input v-model="formData.address" placeholder="例如: 192.168.1.100:9200" clearable>
                         <template #prepend>
                            <el-select v-model="addressProtocol" style="width: 90px">
                               <el-option label="http://" value="http" />
                               <el-option label="https://" value="https" />
                            </el-select>
                         </template>
                      </el-input>
                   </el-form-item>
                </el-col>
                <el-col :span="24" v-if="addressProtocol === 'https'">
                   <el-form-item>
                      <el-checkbox v-model="formData.skipSslVerify">跳过 SSL 证书验证</el-checkbox>
                      <div class="text-xs text-sub ml-2 inline-block">启用此选项可忽略自签名证书错误</div>
                   </el-form-item>
                </el-col>
             </el-row>
          </el-card>

          <!-- Auth Config -->
          <el-card class="content-card mb-20 p-4">
             <template #header>
                <div class="flex justify-between items-center">
                   <div class="font-bold flex items-center gap-2">
                      <el-icon><Lock /></el-icon> 认证配置
                   </div>
                   <el-button link type="primary" size="small" @click="resetAuthConfig">重置</el-button>
                </div>
             </template>

             <el-form-item label="认证方式" prop="authType">
                <el-radio-group v-model="formData.authType" class="w-full flex-wrap">
                   <el-radio-button value="none">无认证</el-radio-button>
                   <el-radio-button value="basic">基础认证 (Basic)</el-radio-button>
                   <el-radio-button value="api_key">API 密钥</el-radio-button>
                   <el-radio-button value="token">Token 令牌</el-radio-button>
                   <el-radio-button v-if="isKubernetesType" value="kubeconfig">Kubeconfig</el-radio-button>
                </el-radio-group>
             </el-form-item>

             <div class="auth-fields mt-4 bg-gray-50 p-6 rounded-lg border border-gray-100" v-if="formData.authType !== 'none'">
                 <!-- Basic -->
                 <div v-if="formData.authType === 'basic'">
                    <el-row :gutter="24">
                       <el-col :span="12">
                          <el-form-item label="用户名" prop="username">
                             <el-input v-model="formData.username" placeholder="Username" />
                          </el-form-item>
                       </el-col>
                       <el-col :span="12">
                          <el-form-item label="密码" prop="password">
                             <el-input v-model="formData.password" type="password" show-password placeholder="Password" />
                          </el-form-item>
                       </el-col>
                    </el-row>
                 </div>

                 <!-- API Key -->
                 <div v-if="formData.authType === 'api_key'">
                    <el-row :gutter="24">
                       <el-col :span="12">
                          <el-form-item label="API ID" prop="apiId">
                             <el-input v-model="formData.apiId" placeholder="API ID" />
                          </el-form-item>
                       </el-col>
                       <el-col :span="12">
                          <el-form-item label="API Key" prop="apiKey">
                             <el-input v-model="formData.apiKey" type="password" show-password placeholder="API Key Secret" />
                          </el-form-item>
                       </el-col>
                    </el-row>
                 </div>

                 <!-- Token -->
                 <div v-if="formData.authType === 'token'">
                    <el-form-item label="Token" prop="token">
                       <el-input v-model="formData.token" type="textarea" :rows="3" placeholder="Bearer Token" />
                    </el-form-item>
                 </div>

                 <!-- Kubeconfig -->
                 <div v-if="formData.authType === 'kubeconfig'">
                    <el-form-item label="Kubeconfig 文件" prop="kubeconfigFile">
                       <el-upload
                         class="w-full"
                         drag
                         action="#"
                         :auto-upload="false"
                         :on-change="handleKubeconfigChange"
                         :limit="1"
                         accept=".yaml,.yml,.conf,.config"
                       >
                         <div class="el-upload__text">拖拽文件到此处或 <em>点击上传</em></div>
                       </el-upload>
                    </el-form-item>
                    <el-form-item label="内容预览" v-if="formData.kubeconfigContent">
                        <el-input type="textarea" v-model="formData.kubeconfigContent" :rows="5" readonly />
                    </el-form-item>
                 </div>
             </div>
          </el-card>

          <!-- Advanced -->
          <el-card class="content-card p-4">
              <template #header>
                 <div class="font-bold flex items-center gap-2">
                    <el-icon><Setting /></el-icon> 高级选项
                 </div>
              </template>
              <el-row :gutter="24">
                 <el-col :span="8">
                    <el-form-item label="连接超时 (秒)">
                       <el-input-number v-model="formData.connectTimeout" :min="1" :max="60" class="w-full" />
                    </el-form-item>
                 </el-col>
                 <el-col :span="8">
                    <el-form-item label="请求超时 (秒)">
                       <el-input-number v-model="formData.requestTimeout" :min="1" :max="120" class="w-full" />
                    </el-form-item>
                 </el-col>
              </el-row>
              <el-row :gutter="24">
                 <el-col :span="12">
                    <el-checkbox v-model="formData.autoMonitor">自动监控状态</el-checkbox>
                    <el-checkbox v-model="formData.enableAlert">启用告警通知</el-checkbox>
                 </el-col>
              </el-row>
          </el-card>
       </el-form>
    </div>
  </div>
</template>

<script setup>
import {computed, markRaw, onMounted, reactive, ref} from 'vue'
import {useRoute, useRouter} from 'vue-router'
import {ElMessage} from 'element-plus'
import {
  ArrowLeft,
  Box,
  Check,
  Connection,
  DataLine,
  DocumentCopy,
  InfoFilled,
  Lock,
  Monitor,
  Setting,
  WarningFilled
} from '@element-plus/icons-vue'
import {addInstance, getInstanceDetail, getInstanceTypes, testConnection, updateInstance} from '@/api/instance.js'

const route = useRoute()
const router = useRouter()

// 表单引用
const formRef = ref()

// 数据状态
const saving = ref(false)
const testing = ref(false)
const addressProtocol = ref('http')
const instanceTypes = ref([]) // 改为从后端动态获取

// 固定图标和颜色映射
const typeDisplayConfig = {
    'elasticsearch': { icon: markRaw(Monitor), color: '#005571', defaultPort: 9200 },
    'kibana': { icon: markRaw(DataLine), color: '#005571', defaultPort: 5601 },
    'logstash': { icon: markRaw(DocumentCopy), color: '#005571', defaultPort: 9600 },
    'kubernetes': { icon: markRaw(Box), color: '#326CE5', defaultPort: 6443 },
    'prometheus': { icon: markRaw(DataLine), color: '#e6a23c', defaultPort: 9090 },
    'filebeat': { icon: markRaw(DocumentCopy), color: '#005571', defaultPort: 5066 },
    'metricbeat': { icon: markRaw(DocumentCopy), color: '#005571', defaultPort: 5067 },
    'apm': { icon: markRaw(WarningFilled), color: '#005571', defaultPort: 8200 }
}

// 判断是否为编辑模式
const isEdit = computed(() => !!route.params.id)

// 表单数据
const formData = reactive({
  id: null,
  instanceType: '',
  name: '',
  address: '',
  httpsEnabled: false,
  skipSslVerify: false,
  authType: 'none',
  username: '',
  password: '',
  apiKey: '',
  apiId: '',
  token: '',
  certPath: '',
  keyPath: '',
  caPath: '',
  kubeconfigFile: null,
  kubeconfigContent: '',
  autoMonitor: true,
  enableAlert: false,
  connectTimeout: 5,
  requestTimeout: 10
})

const isKubernetesType = computed(() => {
  if (!formData.instanceType) return false
  const typeLower = formData.instanceType.toLowerCase()
  return typeLower === 'kubernetes' || typeLower === 'k8s'
})

// 表单验证规则
const formRules = {
  instanceType: [{ required: true, message: '请选择实例类型', trigger: 'change' }],
  name: [{ required: true, message: '请输入实例名称', trigger: 'blur' }],
  address: [{ required: true, message: '请输入实例地址', trigger: 'blur' }],
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }],
  apiId: [{ required: true, message: '请输入API ID', trigger: 'blur' }],
  apiKey: [{ required: true, message: '请输入API Key', trigger: 'blur' }],
  token: [{ required: true, message: '请输入 Token', trigger: 'blur' }]
}

const handleBack = () => router.push('/es/instances')
const handleCancel = () => router.push('/es/instances')

const handleTypeChange = (val) => {
    const type = instanceTypes.value.find(t => t.value === val)
    if (type && !formData.address) {
        formData.address = `127.0.0.1:${type.defaultPort}`
    }
}

const handleKubeconfigChange = (file) => {
    formData.kubeconfigFile = file.raw
    const reader = new FileReader()
    reader.onload = (e) => {
        formData.kubeconfigContent = e.target.result
    }
    reader.readAsText(file.raw)
}

const resetAuthConfig = () => {
    formData.authType = 'none'
    formData.username = ''
    formData.password = ''
    formData.apiKey = ''
    formData.apiId = ''
    formData.token = ''
    formData.kubeconfigFile = null
    formData.kubeconfigContent = ''
}

const handleSave = async () => {
    if (!formRef.value) return
    await formRef.value.validate()

    saving.value = true
    try {
        const typeId = getInstanceTypeId(formData.instanceType)
        if (!typeId) {
            throw new Error('未找到该实例类型')
        }

        const payload = {
            id: Number(formData.id) || 0,
            instance_type_id: typeId,
            name: formData.name,
            address: formData.address,
            https_enabled: addressProtocol.value === 'https',
            skip_ssl_verify: formData.skipSslVerify,
            status: 'active',
            auth_configs: {
                auth_type: formData.authType,
                config_key: 'default',
                config_value: getAuthConfigValue()
            }
        }

        if (isEdit.value) {
            await updateInstance(payload)
            ElMessage.success('更新成功')
        } else {
            await addInstance(payload)
            ElMessage.success('创建成功')
        }
        router.push('/es/instances')
    } catch (e) {
        ElMessage.error(e.message || '保存失败')
    } finally {
        saving.value = false
    }
}

const handleTestConnection = async () => {
    testing.value = true
    try {
        const typeId = getInstanceTypeId(formData.instanceType)
        const payload = {
            id: Number(formData.id) || 0,
            instance_type_id: typeId,
            name: formData.name,
            address: formData.address,
            https_enabled: addressProtocol.value === 'https',
            skip_ssl_verify: formData.skipSslVerify,
            auth_configs: {
                auth_type: formData.authType,
                config_key: 'default',
                config_value: getAuthConfigValue()
            }
        }
        // Since testConnection API expects { instance_id: ID } but we are sending
        // a full object payload from the form meant for testing before saving,
        // we'll send the payload directly via the configured axios instance
        // or check if the backend supports this dynamic testing endpoint format.
        //
        // The backend error `ConnectionTestRequest.instance_id of type uint` indicates
        // the endpoint ONLY accepts `{"instance_id": 10}`. Let's send only the ID.

        await testConnection(Number(formData.id) || 0)
        ElMessage.success('连接测试成功')
    } catch (e) {
        ElMessage.error('连接测试失败: ' + (e.message || '网络错误'))
    } finally {
        testing.value = false
    }
}

const getInstanceTypeId = (typeName) => {
    const type = instanceTypes.value.find(t => t.value === typeName)
    return type ? type.id : null
}

const getAuthConfigValue = () => {
    switch (formData.authType) {
        case 'basic': return JSON.stringify({ username: formData.username, password: formData.password })
        case 'api_key': return JSON.stringify({ id: formData.apiId, key: formData.apiKey })
        case 'token': return formData.token
        case 'kubeconfig': return JSON.stringify({ kubeconfigContent: formData.kubeconfigContent, fileName: formData.kubeconfigFile?.name || 'config' })
        default: return ''
    }
}

const loadInstanceTypes = async () => {
    try {
        const res = await getInstanceTypes()
        const types = res.data?.instance_types || res.data
        if (types && Array.isArray(types)) {
            instanceTypes.value = types.map(type => {
                const typeNameLower = (type.type_name || '').toLowerCase()
                const config = typeDisplayConfig[typeNameLower] || {}
                return {
                    id: type.id,
                    label: type.display_name || type.type_name,
                    value: type.type_name,
                    icon: config.icon || markRaw(Monitor),
                    color: config.color || '#606266',
                    defaultPort: config.defaultPort || 80
                }
            })
        }
    } catch (e) {
        ElMessage.error('加载实例类型失败')
    }
}

onMounted(async () => {
    await loadInstanceTypes()
    if (isEdit.value) {
        const id = route.params.id
        try {
            const res = await getInstanceDetail(id)
            const data = res.data?.instance || res.data

            // Map basic fields
            formData.id = id
            formData.name = data.resource_name || data.name || ''
            formData.address = data.address || ''
            formData.httpsEnabled = !!data.https_enabled
            formData.skipSslVerify = !!data.skip_ssl_verify
            formData.autoMonitor = true // default for now
            formData.enableAlert = false // default for now

            // Parse address protocol
            if (formData.httpsEnabled) {
                addressProtocol.value = 'https'
            } else {
                addressProtocol.value = 'http'
            }

            if (data.instance_type?.type_name) {
                formData.instanceType = data.instance_type.type_name
            } else if (data.resource_type === 'instance') {
                // Determine instance type visually or based on properties if instance_type is missing
                formData.instanceType = 'elasticsearch'
            }

            // Parse auth config
            if (data.auth_configs) {
                try {
                    // Try parsing if it's a string, or use the object directly
                    let authConfigsArr = typeof data.auth_configs === 'string'
                        ? JSON.parse(data.auth_configs)
                        : data.auth_configs;

                    // If it's an array, get the first item (default config)
                    const parsedAuth = Array.isArray(authConfigsArr)
                        ? (authConfigsArr.length > 0 ? authConfigsArr[0] : null)
                        : authConfigsArr;

                    if (parsedAuth && parsedAuth.auth_type) {
                        formData.authType = parsedAuth.auth_type

                        const configStr = parsedAuth.config_value
                        if (configStr) {
                             if (formData.authType === 'basic') {
                                 const parsed = typeof configStr === 'string' ? JSON.parse(configStr) : configStr
                                 if (parsed) {
                                    formData.username = parsed.username || ''
                                    formData.password = parsed.password || ''
                                 }
                             } else if (formData.authType === 'api_key') {
                                 const parsed = typeof configStr === 'string' ? JSON.parse(configStr) : configStr
                                 if (parsed) {
                                    formData.apiId = parsed.id || ''
                                    formData.apiKey = parsed.key || ''
                                 }
                             } else if (formData.authType === 'token') {
                                 formData.token = configStr
                             } else if (formData.authType === 'kubeconfig') {
                                 try {
                                     const parsed = typeof configStr === 'string' && configStr.startsWith('{') ? JSON.parse(configStr) : null;
                                     if (parsed && parsed.kubeconfigContent) {
                                         formData.kubeconfigContent = parsed.kubeconfigContent
                                     } else {
                                         formData.kubeconfigContent = configStr
                                     }
                                 } catch (e) {
                                     formData.kubeconfigContent = configStr
                                 }
                             }
                        }
                    }
                } catch (pe) {
                    console.error("Failed to parse auth config", pe)
                }
            }

        } catch (e) {
            ElMessage.error('加载实例详情失败')
        }
    }
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

.max-w-4xl { max-width: 56rem; }
.mx-auto { margin-left: auto; margin-right: auto; }
.mr-2 { margin-right: 8px; }
.mb-20 { margin-bottom: 20px; }
.mt-4 { margin-top: 16px; }
.ml-2 { margin-left: 8px; }
.p-4 { padding: 16px; }
.p-6 { padding: 24px; }
.w-full { width: 100%; }
.gap-2 { gap: 8px; }
.text-xs { font-size: 12px; }
.text-sub { color: var(--text-sub); }
.font-bold { font-weight: 600; }
.flex { display: flex; }
.items-center { align-items: center; }
.justify-between { justify-content: space-between; }
.flex-wrap { flex-wrap: wrap; }

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.header-left { display: flex; align-items: center; }
.header-title-wrapper h2 { margin: 0; font-size: 18px; color: var(--text-main); }
.header-title-wrapper .subtitle { margin: 0; font-size: 12px; color: var(--text-sub); }
.header-right { display: flex; align-items: center; gap: 12px; }

.grid-radio-group {
    display: flex;
    gap: 0;
}
</style>
