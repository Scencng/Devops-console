<template>
  <div class="pod-terminal">
    <el-dialog 
      v-model="visible" 
      :title="`容器终端 - ${podName}/${containerName}`" 
      width="80%" 
      :close-on-click-modal="false"
      @close="handleClose"
      class="terminal-dialog"
    >
      <div class="terminal-header">
        <div class="header-info">
          <span>命名空间: {{ namespace }}</span>
          <span>Pod: {{ podName }}</span>
          <span>容器: {{ containerName }}</span>
          <el-select v-model="selectedShell" size="small" style="width: 120px; margin-left: 10px;" @change="handleShellChange">
            <el-option label="bash" value="/bin/bash" />
            <el-option label="sh" value="/bin/sh" />
            <el-option label="zsh" value="/bin/zsh" />
            <el-option label="ash" value="/bin/ash" />
          </el-select>
          <el-tag 
            :type="connectionStatusType" 
            size="small" 
            effect="dark" 
            style="margin-left: 10px; font-weight: bold;"
          >
            {{ connectionStatus }}
          </el-tag>
        </div>
        <div class="header-actions">
          <el-button size="small" @click="reconnect" :disabled="connecting">
            <el-icon><Refresh /></el-icon>
            {{ connecting ? '连接中...' : '重连' }}
          </el-button>
          <el-button size="small" @click="disconnect" :disabled="!isConnected">
            <el-icon><Close /></el-icon>
            断开
          </el-button>
          <el-button size="small" @click="clearTerminal">
            <el-icon><Delete /></el-icon>
            清屏
          </el-button>
        </div>
      </div>
      <div id="terminal-container" ref="terminalContainer"></div>
      <div v-if="error" class="error-message">
        <el-alert :title="error" type="error" :closable="false">
          <template #default>
            <div>{{ error }}</div>
            <el-button size="small" type="primary" @click="reconnect" style="margin-top: 8px;">
              重新连接
            </el-button>
          </template>
        </el-alert>
      </div>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted, onBeforeUnmount, watch, nextTick, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { Refresh, Delete, Close } from '@element-plus/icons-vue'
import { Terminal } from 'xterm'
import { FitAddon } from 'xterm-addon-fit'
import { WebLinksAddon } from 'xterm-addon-web-links'
import 'xterm/css/xterm.css'

const props = defineProps({
  modelValue: {
    type: Boolean,
    default: false
  },
  namespace: {
    type: String,
    required: true
  },
  podName: {
    type: String,
    required: true
  },
  containerName: {
    type: String,
    required: true
  },
  instanceId: {
    type: [Number, String],
    required: true
  }
})

const emit = defineEmits(['update:modelValue', 'close'])

const visible = ref(props.modelValue)
const terminalContainer = ref(null)
const error = ref('')
const connecting = ref(false)
const isConnected = ref(false)
const selectedShell = ref('/bin/bash')
let terminal = null
let fitAddon = null
let ws = null
let isIntentionalClose = false // 标记是否主动关闭连接
let connectingStartTime = 0 // 记录开始连接的时间


const connectionStatus = computed(() => {
  if (connecting.value) return '连接中...'
  if (isConnected.value) return '已连接'
  return '已断开'
})

const connectionStatusType = computed(() => {
  if (connecting.value) return 'warning'
  if (isConnected.value) return 'success'
  return 'info'
})

// 切换 shell 时重新连接
const handleShellChange = () => {
  if (isConnected.value) {
    ElMessage.info('正在切换 shell...')
    reconnect()
  }
}

watch(() => props.modelValue, (newVal) => {
  visible.value = newVal
  if (newVal) {
    nextTick(() => {
      initTerminal()
    })
  } else {
    cleanup()
  }
})

watch(visible, (newVal) => {
  emit('update:modelValue', newVal)
  if (!newVal) {
    cleanup()
  }
})

const initTerminal = () => {
  if (!terminalContainer.value) return

  // 创建终端实例
  terminal = new Terminal({
    cursorBlink: true,
    fontSize: 14,
    fontFamily: 'Consolas, "Courier New", monospace',
    theme: {
      background: '#1e1e2d',
      foreground: '#f8f8f2',
      cursor: '#00ff00',
      selection: '#44475a',
      black: '#000000',
      red: '#ff5555',
      green: '#50fa7b',
      yellow: '#f1fa8c',
      blue: '#bd93f9',
      magenta: '#ff79c6',
      cyan: '#8be9fd',
      white: '#bbbbbb'
    },
    rows: 30,
    cols: 120,
    // 启用快捷键支持
    allowProposedApi: true
  })

  // 添加插件
  fitAddon = new FitAddon()
  terminal.loadAddon(fitAddon)
  terminal.loadAddon(new WebLinksAddon())

  // 挂载到DOM
  terminal.open(terminalContainer.value)
  fitAddon.fit()

  // 自动聚焦到终端
  terminal.focus()

  // 连接WebSocket
  connect()

  // 窗口大小调整
  window.addEventListener('resize', handleResize)
  
  // 点击终端容器时聚焦
  terminalContainer.value.addEventListener('click', () => {
    terminal.focus()
  })
}

const connect = () => {
  // 关闭旧连接（如果存在）
  if (ws && ws.readyState !== WebSocket.CLOSED) {
    isIntentionalClose = true
    ws.close()
    ws = null
  }
  
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  const wsUrl = `${protocol}//${window.location.hostname}:8081/ws/pod/${props.podName}/exec?namespace=${props.namespace}&container=${props.containerName}&instance_id=${props.instanceId}&shell=${encodeURIComponent(selectedShell.value)}`

  console.log('connect() - 设置状态: connecting=true, isConnected=false')
  connectingStartTime = Date.now() // 记录开始时间
  connecting.value = true
  isConnected.value = false
  error.value = ''
  
  ws = new WebSocket(wsUrl)

  ws.onopen = () => {
    console.log('ws.onopen - 设置状态: connecting=false, isConnected=true')
    
    // 计算已经过去的时间
    const elapsedTime = Date.now() - connectingStartTime
    const minDisplayTime = 500 // 最小显示时间500ms
    
    // 如果连接太快，延迟设置状态以便用户看到"连接中"状态
    if (elapsedTime < minDisplayTime) {
      const delay = minDisplayTime - elapsedTime
      console.log(`连接太快(${elapsedTime}ms)，延迟${delay}ms后更新状态`)
      setTimeout(() => {
        connecting.value = false
        isConnected.value = true
        error.value = ''
      }, delay)
    } else {
      connecting.value = false
      isConnected.value = true
      error.value = ''
    }
    
    terminal.writeln('\r\n\x1b[32m Successfully 已连接到容器\x1b[0m\r\n')
    
    // 发送初始大小
    sendResize()
    
    // 监听用户输入
    terminal.onData(data => {
      if (ws && ws.readyState === WebSocket.OPEN) {
        ws.send(JSON.stringify({
          type: 'stdin',
          data: data
        }))
      }
    })
  }

  ws.onmessage = (event) => {
    try {
      const msg = JSON.parse(event.data)
      if (msg.type === 'stdout') {
        terminal.write(msg.data)
      } else if (msg.type === 'error') {
        error.value = msg.error
        terminal.writeln(`\r\n\x1b[31m错误: ${msg.error}\x1b[0m\r\n`)
      }
    } catch (e) {
      console.error('解析消息失败:', e)
    }
  }

  ws.onerror = (e) => {
    connecting.value = false
    isConnected.value = false
    error.value = 'WebSocket 连接错误，请检查网络或后端服务'
    terminal.writeln('\r\n\x1b[31m✗ 连接错误\x1b[0m\r\n')
  }

  ws.onclose = (event) => {
    // 只有当关闭的WebSocket是当前的WebSocket时才更新状态
    // 这样可以避免旧连接的onclose事件影响新连接的状态
    if (event.target !== ws) {
      return
    }
    
    connecting.value = false
    isConnected.value = false
    
    // 只有在非主动关闭时才显示断开连接的消息
    if (!isIntentionalClose) {
      if (event.wasClean) {
        terminal.writeln('\r\n\x1b[33m⚠ 连接已正常关闭（可能是容器 shell 退出）\x1b[0m\r\n')
        terminal.writeln('\r\n\x1b[36m提示: 点击"重连"按钮可以重新连接\x1b[0m\r\n')
      } else {
        terminal.writeln('\r\n\x1b[31m✗ 连接意外断开\x1b[0m\r\n')
        error.value = '连接已断开'
      }
    }
    
    // 重置标志
    isIntentionalClose = false
  }
}

const sendResize = () => {
  if (ws && ws.readyState === WebSocket.OPEN && fitAddon) {
    ws.send(JSON.stringify({
      type: 'resize',
      rows: terminal.rows,
      cols: terminal.cols
    }))
  }
}

const handleResize = () => {
  if (fitAddon && terminal) {
    fitAddon.fit()
    sendResize()
  }
}

const reconnect = () => {
  console.log('reconnect被调用，设置connecting=true')
  // 设置连接中状态
  connecting.value = true
  isConnected.value = false
  
  if (ws && ws.readyState === WebSocket.OPEN) {
    isIntentionalClose = true // 标记为主动关闭
    ws.close()
  }
  cleanup()
  nextTick(() => {
    initTerminal()
  })
  ElMessage.success('正在重新连接...')
}

const disconnect = () => {
  if (ws && ws.readyState === WebSocket.OPEN) {
    isIntentionalClose = true // 标记为主动关闭
    ws.close()
    terminal.writeln('\r\n\x1b[33m⚠ 手动断开连接\x1b[0m\r\n')
    ElMessage.info('已断开连接')
  }
}

const clearTerminal = () => {
  if (terminal) {
    terminal.clear()
  }
}

const handleClose = () => {
  visible.value = false
  emit('close')
}

const cleanup = () => {
  if (ws) {
    isIntentionalClose = true // 标记为主动关闭
    ws.close()
    ws = null
  }
  if (terminal) {
    terminal.dispose()
    terminal = null
  }
  if (fitAddon) {
    fitAddon = null
  }
  window.removeEventListener('resize', handleResize)
}

onBeforeUnmount(() => {
  cleanup()
})
</script>

<style scoped>
.pod-terminal {
  width: 100%;
}

.terminal-dialog :deep(.el-dialog__body) {
  padding: 0;
}

.terminal-header {
  padding: 12px 16px;
  background: #2a2a3e;
  color: #fff;
  display: flex;
  justify-content: space-between;
  align-items: center;
  border-bottom: 1px solid #3a3a4e;
}

.header-info {
  display: flex;
  gap: 20px;
  font-size: 13px;
}

.header-info span {
  color: #a0a0b0;
}

.header-actions {
  display: flex;
  gap: 8px;
}

#terminal-container {
  height: 500px;
  padding: 8px;
  background: #1e1e2d;
  overflow: hidden;
}

.error-message {
  padding: 16px;
  background: #2a2a3e;
}

html.dark .terminal-header {
  background: #1e1e2d;
  border-bottom-color: #2a2a3e;
}
</style>
