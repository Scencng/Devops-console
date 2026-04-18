<template>
  <div class="autoops-container chaos-create-page">
    <div class="page-header">
      <el-page-header @back="handleBack" title="返回">
        <template #content>
          <span class="page-title">创建混沌实验</span>
        </template>
      </el-page-header>
    </div>

    <div class="chaos-wizard-container">
      <el-card>
        <!-- Steps Indicator -->
        <el-steps :active="currentStep" finish-status="success" align-center>
          <el-step title="基本信息" />
          <el-step title="故障配置" />
          <el-step title="目标选择" />
          <el-step title="演练隔离" />
          <el-step title="确认提交" />
        </el-steps>

        <!-- Step 1: Basic Info -->
        <div v-show="currentStep === 0" class="wizard-step">
          <el-form :model="formData" :rules="step1Rules" ref="formStep1" label-width="120px">
            <el-form-item label="实验名称" prop="name">
              <el-input v-model="formData.name" placeholder="请输入实验名称" />
            </el-form-item>

            <el-form-item label="命名空间" prop="namespace">
              <el-select v-model="formData.namespace" placeholder="请选择命名空间" style="width: 100%">
                <el-option v-for="ns in namespaceList" :key="ns.name" :label="ns.name" :value="ns.name" />
              </el-select>
            </el-form-item>

            <el-form-item label="故障类型" prop="faultType">
              <el-radio-group v-model="formData.faultType" @change="handleFaultTypeChange">
                <el-radio value="PodChaos">Pod混沌</el-radio>
                <el-radio value="NetworkChaos">网络混沌</el-radio>
                <el-radio value="IOChaos">IO混沌</el-radio>
                <el-radio value="StressChaos">压力混沌</el-radio>
              </el-radio-group>
            </el-form-item>

            <el-form-item label="持续时间">
              <el-input v-model="formData.duration" placeholder="例如: 5m, 1h, 30s" />
              <span class="form-tip">格式: 数字 + 单位 (s/m/h)</span>
            </el-form-item>

            <el-form-item label="标签">
              <el-input
                v-model="formData.labelsStr"
                type="textarea"
                :rows="2"
                placeholder="格式: key1=value1,key2=value2"
              />
            </el-form-item>
          </el-form>

          <div class="step-actions">
            <el-button type="primary" @click="nextStep">下一步</el-button>
          </div>
        </div>

        <!-- Step 2: Fault Configuration -->
        <div v-show="currentStep === 1" class="wizard-step">
          <el-form :model="formData.spec" :rules="step2Rules" ref="formStep2" label-width="120px">

            <!-- PodChaos Configuration -->
            <div v-if="formData.faultType === 'PodChaos'">
              <el-form-item label="故障动作" prop="action">
                <el-select v-model="formData.spec.action" placeholder="请选择故障动作">
                  <el-option label="Pod Kill" value="pod-kill" />
                  <el-option label="Pod Failure" value="pod-failure" />
                  <el-option label="Container Kill" value="container-kill" />
                </el-select>
              </el-form-item>

              <el-form-item label="容器名称">
                <el-input v-model="formData.spec.containerName" placeholder="请输入容器名称 (可选)" />
              </el-form-item>

              <el-form-item label="宽限期(秒)">
                <el-input-number v-model="formData.spec.gracePeriod" :min="0" :max="3600" />
              </el-form-item>

              <el-form-item label="影响模式">
                <el-select v-model="formData.spec.mode" placeholder="请选择影响模式">
                  <el-option label="One" value="one" />
                  <el-option label="All" value="all" />
                  <el-option label="Fixed" value="fixed" />
                  <el-option label="Fixed Percent" value="fixed-percent" />
                  <el-option label="Random Max Percent" value="random-max-percent" />
                </el-select>
              </el-form-item>

              <el-form-item label="影响数值">
                <el-input v-model="formData.spec.value" placeholder="请输入影响数值" />
              </el-form-item>
            </div>

            <!-- NetworkChaos Configuration -->
            <div v-if="formData.faultType === 'NetworkChaos'">
              <el-form-item label="故障动作" prop="action">
                <el-select v-model="formData.spec.action" placeholder="请选择故障动作">
                  <el-option label="Delay" value="delay" />
                  <el-option label="Loss" value="loss" />
                  <el-option label="Duplicate" value="duplicate" />
                  <el-option label="Corrupt" value="corrupt" />
                  <el-option label="Partition" value="partition" />
                </el-select>
              </el-form-item>

              <el-form-item v-if="formData.spec.action === 'delay'" label="延迟配置">
                <el-input v-model="formData.spec.delayLatency" placeholder="延迟时间 (例如: 10ms)" />
                <el-input v-model="formData.spec.delayJitter" placeholder="抖动时间 (例如: 5ms)" />
              </el-form-item>

              <el-form-item v-if="formData.spec.action === 'loss'" label="丢包配置">
                <el-input v-model="formData.spec.lossLoss" placeholder="丢包百分比 (例如: 10)" />
              </el-form-item>

              <el-form-item label="影响方向">
                <el-select v-model="formData.spec.direction" placeholder="请选择影响方向">
                  <el-option label="To" value="to" />
                  <el-option label="From" value="from" />
                  <el-option label="Both" value="both" />
                </el-select>
              </el-form-item>
            </div>

            <!-- IOChaos Configuration -->
            <div v-if="formData.faultType === 'IOChaos'">
              <el-form-item label="故障动作" prop="action">
                <el-select v-model="formData.spec.action" placeholder="请选择故障动作">
                  <el-option label="Latency" value="latency" />
                  <el-option label="Fault" value="fault" />
                </el-select>
              </el-form-item>

              <el-form-item label="影响百分比">
                <el-input-number v-model="formData.spec.percent" :min="0" :max="100" />
              </el-form-item>

              <el-form-item label="目标路径">
                <el-input v-model="formData.spec.path" placeholder="例如: /data" />
              </el-form-item>

              <el-form-item label="IO方法">
                <el-checkbox-group v-model="formData.spec.methods">
                  <el-checkbox value="read">Read</el-checkbox>
                  <el-checkbox value="write">Write</el-checkbox>
                </el-checkbox-group>
              </el-form-item>
            </div>

            <!-- StressChaos Configuration -->
            <div v-if="formData.faultType === 'StressChaos'">
              <el-form-item label="CPU压力">
                <el-input-number v-model="formData.spec.cpuLoad" :min="0" :max="100" placeholder="CPU负载百分比" />
                <el-input-number v-model="formData.spec.cpuWorkers" :min="1" :max="32" placeholder="Workers数量" />
              </el-form-item>

              <el-form-item label="内存压力">
                <el-input v-model="formData.spec.memorySize" placeholder="内存大小 (例如: 256MB)" />
                <el-input-number v-model="formData.spec.memoryWorkers" :min="1" :max="32" placeholder="Workers数量" />
              </el-form-item>

              <el-form-item label="Workers数量">
                <el-input-number v-model="formData.spec.workers" :min="1" :max="64" />
              </el-form-item>
            </div>
          </el-form>

          <div class="step-actions">
            <el-button @click="prevStep">上一步</el-button>
            <el-button type="primary" @click="nextStep">下一步</el-button>
          </div>
        </div>

        <!-- Step 3: Target Selector -->
        <div v-show="currentStep === 2" class="wizard-step">
          <el-form :model="formData.spec" :rules="step3Rules" ref="formStep3" label-width="120px">
            <el-form-item label="选择方式">
              <el-radio-group v-model="targetMode">
                <el-radio value="label">Label选择器</el-radio>
                <el-radio value="pod">指定Pod</el-radio>
              </el-radio-group>
            </el-form-item>

            <el-form-item v-if="targetMode === 'label'" label="Label选择器">
              <el-input
                v-model="formData.spec.labelSelectorsStr"
                type="textarea"
                :rows="4"
                placeholder="格式: app=frontend,env=dev"
              />
            </el-form-item>

            <el-form-item v-if="targetMode === 'pod'" label="指定Pod">
              <el-input
                v-model="formData.spec.podsStr"
                type="textarea"
                :rows="4"
                placeholder="格式: namespace1:pod1,pod2;namespace2:pod3"
              />
            </el-form-item>

            <el-form-item label="影响模式">
              <el-select v-model="formData.spec.mode" placeholder="请选择影响模式">
                <el-option label="One" value="one" />
                <el-option label="All" value="all" />
                <el-option label="Fixed" value="fixed" />
                <el-option label="Fixed Percent" value="fixed-percent" />
                <el-option label="Random Max Percent" value="random-max-percent" />
              </el-select>
            </el-form-item>

            <el-form-item label="影响数值">
              <el-input v-model="formData.spec.value" placeholder="请输入影响数值" />
            </el-form-item>
          </el-form>

          <div class="step-actions">
            <el-button @click="prevStep">上一步</el-button>
            <el-button type="primary" @click="nextStep">下一步</el-button>
          </div>
        </div>

        <!-- Step 4: Eviction Isolation -->
        <div v-show="currentStep === 3" class="wizard-step">
          <div class="eviction-section">
            <div class="eviction-header">
              <el-icon class="eviction-icon"><Connection /></el-icon>
              <div>
                <div class="eviction-title">演练节点隐离（推荐）</div>
                <div class="eviction-desc">将目标 Pod 迁移到专用演练节点再注入故障，严格控制爆炸半径</div>
              </div>
            </div>

            <el-form :model="evictionForm" class="eviction-form" label-width="130px">
              <el-form-item label="开启演练隔离">
                <el-switch
                  v-model="evictionForm.enabled"
                  active-text="开启"
                  inactive-text="不限制"
                  @change="handleEvictionToggle"
                />
              </el-form-item>

              <template v-if="evictionForm.enabled">
                <el-form-item label="演练节点">
                  <el-select
                    v-model="evictionForm.nodeName"
                    :loading="loadingChaosNodes"
                    placeholder="请选择演练节点"
                    style="width: 100%"
                  >
                    <el-option
                      v-for="node in chaosNodes"
                      :key="node.name"
                      :disabled="node.status !== 'Ready'"
                      :label="`${node.name}  (${node.status})`"
                      :value="node.name"
                    />
                  </el-select>
                  <div v-if="chaosNodes.length === 0 && !loadingChaosNodes" class="form-tip">
                    没有找到演练节点，请先给节点打上 <code>role=chaos-testing</code> 标签
                  </div>
                </el-form-item>

                <el-form-item label="目标命名空间">
                  <el-select
                    v-model="evictionForm.targetNamespace"
                    placeholder="请选择命名空间"
                    style="width: 100%"
                    @change="handleTargetNamespaceChange"
                  >
                    <el-option
                      v-for="ns in namespaceList"
                      :key="ns.name"
                      :label="ns.name"
                      :value="ns.name"
                    />
                  </el-select>
                </el-form-item>

                <el-form-item label="目标 Deployment">
                  <el-select
                    v-model="evictionForm.deploymentName"
                    :loading="loadingDeployments"
                    filterable
                    placeholder="请选择要驱逐的 Deployment"
                    style="width: 100%"
                  >
                    <el-option
                      v-for="deploy in deploymentList"
                      :key="deploy.name"
                      :label="deploy.name"
                      :value="deploy.name"
                    />
                  </el-select>
                  <div class="form-tip">该 Deployment 的 Pod 将被迁移到演练节点</div>
                </el-form-item>

                <el-alert :closable="false" class="eviction-alert" type="warning">
                  <template #default>
                    <div>⚠️ 打开后，提交实验时将执行：</div>
                    <ol class="eviction-steps-list">
                      <li>给演练节点打上 <code>NoSchedule</code> 污点</li>
                      <li>Patch 目标 Deployment ，加入 Toleration + NodeAffinity</li>
                      <li>Evict 目标 Pod，Pod 将在演练节点重建</li>
                      <li>等待 Pod 在演练节点就绪后创建 Chaos 实验</li>
                    </ol>
                    <div>实验完成后，请在列表页点击“清理”恢复环境。</div>
                  </template>
                </el-alert>
              </template>
            </el-form>
          </div>

          <div class="step-actions">
            <el-button @click="prevStep">上一步</el-button>
            <el-button type="primary" @click="nextStep">下一步</el-button>
          </div>
        </div>

        <!-- Step 5: Preview & Submit -->
        <div v-show="currentStep === 4" class="wizard-step">
          <el-descriptions title="实验配置预览" :column="1" border>
            <el-descriptions-item label="实验名称">
              {{ formData.name }}
            </el-descriptions-item>
            <el-descriptions-item label="命名空间">
              {{ formData.namespace }}
            </el-descriptions-item>
            <el-descriptions-item label="故障类型">
              {{ formData.faultType }}
            </el-descriptions-item>
            <el-descriptions-item label="持续时间">
              {{ formData.duration || '不限制' }}
            </el-descriptions-item>
            <el-descriptions-item label="目标选择方式">
              {{ targetMode === 'label' ? 'Label选择器' : '指定Pod' }}
            </el-descriptions-item>
            <el-descriptions-item label="影响模式">
              {{ formData.spec.mode || '-' }}
            </el-descriptions-item>
            <el-descriptions-item v-if="formData.spec.action" label="故障动作">
              {{ formData.spec.action }}
            </el-descriptions-item>
            <el-descriptions-item label="演练节点隐离">
              {{ evictionForm.enabled ? `开启（节点: ${evictionForm.nodeName}， Namespace: ${evictionForm.targetNamespace}， Deployment: ${evictionForm.deploymentName}）` : '未开启' }}
            </el-descriptions-item>
          </el-descriptions>

          <div class="step-actions">
            <el-button @click="prevStep">上一步</el-button>
            <el-button :loading="submitting" type="primary" @click="handleSubmit">
              {{ evictionForm.enabled ? '迁移 Pod 并提交实验' : '提交实验' }}
            </el-button>
          </div>
        </div>
      </el-card>
    </div>
  </div>
</template>

<script setup>
import {onMounted, reactive, ref} from 'vue'
import {ElMessage} from 'element-plus'
import {useRouter} from 'vue-router'
import {createChaosExperiment, getChaosTestingNodes, prepareEviction} from '@/api/chaos'
import {getSelectedInstanceId} from '@/stores/instanceStore'
import {getNamespaceList} from '@/api/k8s/namespace'
import {getDeploymentList} from '@/api/k8s/deployment'
import {Connection} from '@element-plus/icons-vue'

const router = useRouter()

const currentStep = ref(0)
const submitting = ref(false)
const targetMode = ref('label')
const namespaceList = ref([])

// 演练节点驱逐相关
const chaosNodes = ref([])
const loadingChaosNodes = ref(false)
const deploymentList = ref([])
const loadingDeployments = ref(false)

const evictionForm = reactive({
  enabled: false,
  nodeName: '',
  targetNamespace: '',
  deploymentName: ''
})

const formData = reactive({
  name: '',
  namespace: '',
  faultType: 'PodChaos',
  duration: '5m',
  labelsStr: '',
  labels: {},
  spec: {
    action: 'pod-kill',
    containerName: '',
    gracePeriod: 0,
    mode: 'one',
    value: '1',
    selector: {},
    delayLatency: '',
    delayJitter: '',
    lossLoss: '',
    direction: 'both',
    percent: 100,
    path: '',
    methods: [],
    cpuLoad: 0,
    cpuWorkers: 1,
    memorySize: '',
    memoryWorkers: 1,
    workers: 1,
    labelSelectorsStr: '',
    podsStr: ''
  }
})

const step1Rules = {
  name: [{ required: true, message: '请输入实验名称', trigger: 'blur' }],
  namespace: [{ required: true, message: '请选择命名空间', trigger: 'change' }],
  faultType: [{ required: true, message: '请选择故障类型', trigger: 'change' }]
}

const step2Rules = {
  action: [{ required: true, message: '请选择故障动作', trigger: 'change' }]
}

const step3Rules = {
  mode: [{ required: true, message: '请选择影响模式', trigger: 'change' }]
}

const formStep1 = ref(null)
const formStep2 = ref(null)
const formStep3 = ref(null)

const fetchNamespaces = async () => {
  try {
    const instanceId = getSelectedInstanceId()
    const res = await getNamespaceList(instanceId)
    namespaceList.value = res.data?.namespaceList || []
    if (namespaceList.value.length > 0 && !formData.namespace) {
      formData.namespace = namespaceList.value[0].name
    }
  } catch (e) {
    ElMessage.error('获取命名空间列表失败')
  }
}

const handleFaultTypeChange = () => {
  formData.spec.action = getDefaultAction(formData.faultType)
}

const getDefaultAction = (type) => {
  const actions = {
    'PodChaos': 'pod-kill',
    'NetworkChaos': 'delay',
    'IOChaos': 'latency',
    'StressChaos': ''
  }
  return actions[type] || ''
}

const nextStep = async () => {
  if (currentStep.value === 0) {
    try {
      await formStep1.value.validate()
      currentStep.value++
    } catch (e) {
      ElMessage.warning('请完善基本信息')
    }
  } else if (currentStep.value === 1) {
    try {
      await formStep2.value.validate()
      currentStep.value++
    } catch (e) {
      ElMessage.warning('请完善故障配置')
    }
  } else if (currentStep.value === 2) {
    try {
      await formStep3.value.validate()
      currentStep.value++
    } catch (e) {
      ElMessage.warning('请完善目标选择')
    }
  } else if (currentStep.value === 3) {
    // 演练隅离步骤验证
    if (evictionForm.enabled) {
      if (!evictionForm.nodeName) {
        ElMessage.warning('请选择演练节点')
        return
      }
      if (!evictionForm.targetNamespace) {
        ElMessage.warning('请选择目标命名空间')
        return
      }
      if (!evictionForm.deploymentName) {
        ElMessage.warning('请选择目标 Deployment 名称')
        return
      }
    }
    currentStep.value++
  }
}

const prevStep = () => {
  if (currentStep.value > 0) {
    currentStep.value--
  }
}

const buildSelector = () => {
  const selector = {}

  if (targetMode.value === 'label' && formData.spec.labelSelectorsStr) {
    const pairs = formData.spec.labelSelectorsStr.split(',').filter(p => p.trim())
    const labelSelectors = {}
    pairs.forEach(pair => {
      const [key, value] = pair.split('=').map(s => s.trim())
      if (key && value) {
        labelSelectors[key] = value
      }
    })
    selector.labelSelectors = labelSelectors
  }

  if (targetMode.value === 'pod' && formData.spec.podsStr) {
    const entries = formData.spec.podsStr.split(';').filter(e => e.trim())
    const pods = {}
    entries.forEach(entry => {
      const [ns, podNames] = entry.split(':').map(s => s.trim())
      if (ns && podNames) {
        pods[ns] = podNames.split(',').map(p => p.trim()).filter(p => p)
      }
    })
    selector.pods = pods
  }

  return selector
}

const buildLabels = () => {
  if (!formData.labelsStr) return {}

  const labels = {}
  const pairs = formData.labelsStr.split(',').filter(p => p.trim())
  pairs.forEach(pair => {
    const [key, value] = pair.split('=').map(s => s.trim())
    if (key && value) {
      labels[key] = value
    }
  })
  return labels
}

const buildSpec = () => {
  const spec = {
    selector: buildSelector(),
    mode: formData.spec.mode,
    value: formData.spec.value
  }

  if (formData.faultType === 'PodChaos') {
    spec.action = formData.spec.action
    spec.containerName = formData.spec.containerName
    spec.gracePeriod = formData.spec.gracePeriod
  } else if (formData.faultType === 'NetworkChaos') {
    spec.action = formData.spec.action
    spec.direction = formData.spec.direction
    if (formData.spec.action === 'delay') {
      spec.delay = {
        latency: formData.spec.delayLatency,
        jitter: formData.spec.delayJitter
      }
    }
    if (formData.spec.action === 'loss') {
      spec.loss = {
        loss: formData.spec.lossLoss
      }
    }
  } else if (formData.faultType === 'IOChaos') {
    spec.action = formData.spec.action
    spec.percent = formData.spec.percent
    spec.path = formData.spec.path
    spec.methods = formData.spec.methods
  } else if (formData.faultType === 'StressChaos') {
    spec.workers = formData.spec.workers
    if (formData.spec.cpuLoad > 0) {
      spec.cpu = {
        load: formData.spec.cpuLoad,
        workers: formData.spec.cpuWorkers
      }
    }
    if (formData.spec.memorySize) {
      spec.memory = {
        size: formData.spec.memorySize,
        workers: formData.spec.memoryWorkers
      }
    }
  }

  return spec
}

// 演练节点隘离相关
const handleEvictionToggle = async (val) => {
  if (val) {
    if (chaosNodes.value.length === 0) {
      await fetchChaosNodes()
    }
    if (!evictionForm.targetNamespace && formData.namespace) {
      evictionForm.targetNamespace = formData.namespace
      fetchDeployments()
    }
  }
}

const handleTargetNamespaceChange = () => {
  evictionForm.deploymentName = ''
  fetchDeployments()
}

const fetchDeployments = async () => {
  if (!evictionForm.targetNamespace) return
  loadingDeployments.value = true
  try {
    const instanceId = getSelectedInstanceId()
    const res = await getDeploymentList(evictionForm.targetNamespace, instanceId)
    // 根据具体返回结构适配
    const list = res.data?.items || res.data?.deploymentList || res.data || []
    deploymentList.value = Array.isArray(list) ? list : []
  } catch (e) {
    ElMessage.error('获取 Deployment 列表失败')
  } finally {
    loadingDeployments.value = false
  }
}

const fetchChaosNodes = async () => {
  loadingChaosNodes.value = true
  try {
    const instanceId = getSelectedInstanceId()
    const res = await getChaosTestingNodes(instanceId)
    chaosNodes.value = res.data?.nodes || []
  } catch (e) {
    ElMessage.error('获取演练节点列表失败')
  } finally {
    loadingChaosNodes.value = false
  }
}

const handleSubmit = async () => {
  submitting.value = true
  try {
    const instanceId = getSelectedInstanceId()

    // 如果开启了演练隅离，先执行 Pod 迁移
    if (evictionForm.enabled) {
      ElMessage.info('正在将 Pod 迁移到演练节点，请稍候...(最长等待 120 秒)')
      await prepareEviction({
        nodeName: evictionForm.nodeName,
        namespace: formData.namespace,          // 实验 CR 命名空间
        deploymentNamespace: evictionForm.targetNamespace,  // Deployment 实际命名空间
        deploymentName: evictionForm.deploymentName
      }, instanceId)
      ElMessage.success('Pod 已成功迁移到演练节点！')
    }

    // 构建请求
    const chaosLabels = buildLabels()
    if (evictionForm.enabled) {
      chaosLabels['chaos-eviction'] = 'true'
      chaosLabels['chaos-node'] = evictionForm.nodeName
      chaosLabels['chaos-deploy'] = evictionForm.deploymentName
      chaosLabels['chaos-namespace'] = evictionForm.targetNamespace
    }

    const request = {
      name: formData.name,
      namespace: formData.namespace,
      faultType: formData.faultType,
      duration: formData.duration,
      labels: chaosLabels,
      spec: buildSpec()
    }

    await createChaosExperiment(formData.namespace, request, instanceId)
    ElMessage.success('实验创建成功' + (evictionForm.enabled ? '！实验结束后请记得点击列表页“清理”恢复环境' : ''))
    router.push('/chaos')
  } catch (e) {
    ElMessage.error('实验创建失败: ' + (e.response?.data?.message || e.message || '未知错误'))
  } finally {
    submitting.value = false
  }
}

const handleBack = () => {
  router.push('/chaos')
}

onMounted(() => {
  fetchNamespaces()
})
</script>

<style scoped>
.chaos-create-page {
  padding: 20px;
}

.page-header {
  margin-bottom: 20px;
}

.page-title {
  font-size: 18px;
  font-weight: 500;
}

.chaos-wizard-container {
  max-width: 900px;
  margin: 0 auto;
}

.wizard-step {
  margin-top: 30px;
  padding: 20px;
}

.step-actions {
  display: flex;
  justify-content: center;
  gap: 12px;
  margin-top: 30px;
}

.form-tip {
  font-size: 12px;
  color: #909399;
  margin-left: 8px;
}

.eviction-section {
  background: linear-gradient(135deg, rgba(64, 158, 255, 0.04) 0%, rgba(103, 194, 58, 0.04) 100%);
  border: 1px solid rgba(64, 158, 255, 0.2);
  border-radius: 12px;
  padding: 24px;
  margin-bottom: 20px;
}

.eviction-header {
  display: flex;
  align-items: flex-start;
  gap: 14px;
  margin-bottom: 20px;
}

.eviction-icon {
  font-size: 28px;
  color: var(--el-color-primary);
  margin-top: 2px;
  flex-shrink: 0;
}

.eviction-title {
  font-size: 16px;
  font-weight: 600;
  color: var(--el-text-color-primary);
  margin-bottom: 4px;
}

.eviction-desc {
  font-size: 13px;
  color: var(--el-text-color-secondary);
}

.eviction-form {
  margin-top: 8px;
}

.eviction-alert {
  margin-top: 12px;
}

.eviction-steps-list {
  margin: 8px 0 8px 16px;
  padding: 0;
  font-size: 13px;
  line-height: 2;
}

.eviction-steps-list code {
  background: rgba(0, 0, 0, 0.06);
  padding: 1px 5px;
  border-radius: 3px;
  font-family: monospace;
  font-size: 12px;
}

</style>
