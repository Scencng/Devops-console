<template>
  <div class="code-editor" ref="editorContainer"></div>
</template>

<script setup>
import { ref, onMounted, onBeforeUnmount, watch } from 'vue'
import * as monaco from 'monaco-editor'
import 'monaco-editor/min/vs/editor/editor.main.css'

// 配置 worker
import editorWorker from 'monaco-editor/esm/vs/editor/editor.worker?worker'
import jsonWorker from 'monaco-editor/esm/vs/language/json/json.worker?worker'
import cssWorker from 'monaco-editor/esm/vs/language/css/css.worker?worker'
import htmlWorker from 'monaco-editor/esm/vs/language/html/html.worker?worker'
import tsWorker from 'monaco-editor/esm/vs/language/typescript/ts.worker?worker'

self.MonacoEnvironment = {
  getWorker(_, label) {
    if (label === 'json') {
      return new jsonWorker()
    }
    if (label === 'css' || label === 'scss' || label === 'less') {
      return new cssWorker()
    }
    if (label === 'html' || label === 'handlebars' || label === 'razor') {
      return new htmlWorker()
    }
    if (label === 'typescript' || label === 'javascript') {
      return new tsWorker()
    }
    return new editorWorker()
  }
}

const props = defineProps({
  modelValue: {
    type: String,
    default: ''
  },
  language: {
    type: String,
    default: 'yaml'
  },
  theme: {
    type: String,
    default: 'vs-dark' // 'vs', 'vs-dark', 'hc-black'
  },
  readOnly: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['update:modelValue', 'change'])

const editorContainer = ref(null)
let editorInstance = null

onMounted(() => {
  if (editorContainer.value) {
    // 延迟初始化，确保容器已有尺寸（特别是 Dialog 动画结束）
    setTimeout(() => {
      initEditor()
    }, 200)
  }
})

const initEditor = () => {
  if (!editorContainer.value) return

  // 关键修复：如果你宽度为0，说明还没渲染好，等等再来
  if (editorContainer.value.clientWidth <= 0) {
    console.warn('CodeEditor: Container width is 0, waiting...')
    setTimeout(initEditor, 50)
    return
  }

  // 防止重复初始化
  if (editorInstance) {
    editorInstance.dispose()
  }

  editorInstance = monaco.editor.create(editorContainer.value, {
    value: props.modelValue,
    language: props.language,
    theme: props.theme,
    readOnly: props.readOnly,
    automaticLayout: true,
    scrollBeyondLastLine: false,
    minimap: { enabled: false }, // 再次确认关闭 minimap 防止视觉干扰
    fontSize: 14,
    fontFamily: "'JetBrains Mono', 'Consolas', monospace",
    tabSize: 2,
    renderLineHighlight: 'all', // 确保高亮正常
    scrollbar: {
      useShadows: false,
      verticalHasArrows: false,
      horizontalHasArrows: false,
      vertical: 'visible',
      horizontal: 'visible'
    }
  })

  // Listen for content changes
  editorInstance.onDidChangeModelContent(() => {
    const value = editorInstance.getValue()
    emit('update:modelValue', value)
    emit('change', value)
  })
  
  // ResizeObserver 作为最后的保底
  const resizeObserver = new ResizeObserver(() => {
     if (editorInstance && editorContainer.value && editorContainer.value.clientWidth > 0) {
       editorInstance.layout()
     }
  })
  resizeObserver.observe(editorContainer.value)
  editorInstance._resizeObserver = resizeObserver
}

const layout = () => {
  if (editorInstance) {
    editorInstance.layout()
  }
}

defineExpose({
  layout
})

// Watch for prop changes
watch(() => props.modelValue, (newValue) => {
  if (editorInstance && newValue !== editorInstance.getValue()) {
    editorInstance.setValue(newValue)
  }
})

watch(() => props.language, (newLang) => {
  if (editorInstance) {
    monaco.editor.setModelLanguage(editorInstance.getModel(), newLang)
  }
})

watch(() => props.theme, (newTheme) => {
  if (editorInstance) {
    monaco.editor.setTheme(newTheme)
  }
})

onBeforeUnmount(() => {
  if (editorInstance) {
    if (editorInstance._resizeObserver) {
      editorInstance._resizeObserver.disconnect()
    }
    editorInstance.dispose()
  }
})
</script>

<style scoped>
.code-editor {
  width: 100%;
  height: 100%;
  border-radius: 4px;
  overflow: hidden;
  border: 1px solid var(--border-color, #333);
}
</style>
