import {reactive} from 'vue'

// 尝试从 localStorage 恢复状态
const savedState = JSON.parse(localStorage.getItem('selectedInstance') || '{}')

// 全局状态
const state = reactive({
  selectedInstanceId: savedState.selectedInstanceId || '',
  selectedInstanceName: savedState.selectedInstanceName || '',
  selectedInstanceType: savedState.selectedInstanceType || ''
})

// 设置选中的实例
export const setSelectedInstance = (instance) => {
  state.selectedInstanceId = instance.id
  state.selectedInstanceName = instance.name
  state.selectedInstanceType = instance.instance_type || instance.type_name || instance.type
  // 保存到 localStorage
  localStorage.setItem('selectedInstance', JSON.stringify({
    selectedInstanceId: state.selectedInstanceId,
    selectedInstanceName: state.selectedInstanceName,
    selectedInstanceType: state.selectedInstanceType
  }))
}

// 获取选中的实例ID
export const getSelectedInstanceId = () => state.selectedInstanceId

// 获取选中的实例名称
export const getSelectedInstanceName = () => state.selectedInstanceName

// 获取选中的实例类型
export const getSelectedInstanceType = () => state.selectedInstanceType

// 清除选中状态
export const clearSelectedInstance = () => {
  state.selectedInstanceId = ''
  state.selectedInstanceName = ''
  state.selectedInstanceType = ''
}

export default {
  state,
  setSelectedInstance,
  getSelectedInstanceId,
  getSelectedInstanceName,
  getSelectedInstanceType,
  clearSelectedInstance
}
