import api from './index'

// 获取用户的自定义监控大盘配置
export const getCustomMonitors = (targetType) => {
    return api.get('/monitor/custom', {
        params: { target_type: targetType }
    })
}

// 创建一个新的自定义监控图表
export const createCustomMonitor = (data) => {
    return api.post('/monitor/custom', data)
}

// 更新自定义监控图表
export const updateCustomMonitor = (id, data) => {
    return api.put(`/monitor/custom/${id}`, data)
}

// 删除自定义监控图表
export const deleteCustomMonitor = (id) => {
    return api.delete(`/monitor/custom/${id}`)
}
