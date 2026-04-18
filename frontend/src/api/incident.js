import api from './index'

// ==================== 故障统计 ====================
export const getIncidentStats = () => api.get('/monitor/incident/stats')

// ==================== 故障列表管理 ====================
export const getIncidentList = (params) => api.get('/monitor/incident', { params })
export const getIncidentById = (id) => api.get(`/monitor/incident/${id}`)
export const createIncident = (data) => api.post('/monitor/incident', data)
export const updateIncident = (id, data) => api.put(`/monitor/incident/${id}`, data)
export const deleteIncident = (id) => api.delete(`/monitor/incident/${id}`)

// ==================== 故障状态流转 ====================
// status: pending(未处理) -> processing(处理中) -> done(已完成)
export const updateIncidentStatus = (id, data) => api.put(`/monitor/incident/${id}/status`, data)

// ==================== 业务线 / 部门选项 ====================
export const getBusinessLines = () => api.get('/monitor/incident/business-lines')
