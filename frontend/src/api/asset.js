import api from './index'

// ==================== 主机分组 ====================
export const getHostGroups = () => api.get('/asset/host-groups')
export const createHostGroup = (data) => api.post('/asset/host-groups', data)
export const updateHostGroup = (id, data) => api.put(`/asset/host-groups/${id}`, data)
export const deleteHostGroup = (id) => api.delete(`/asset/host-groups/${id}`)

// ==================== 主机管理 ====================
export const getHostList = (params) => api.get('/asset/hosts', { params })
export const getHostById = (id) => api.get(`/asset/hosts/${id}`)
export const createHost = (data) => api.post('/asset/hosts', data)
export const updateHost = (id, data) => api.put(`/asset/hosts/${id}`, data)
export const deleteHost = (id) => api.delete(`/asset/hosts/${id}`)
export const batchDeleteHosts = (ids) => api.delete('/asset/hosts/batch', { data: { ids } })
export const updateHostStatus = (id, data) => api.put(`/asset/hosts/${id}/status`, data)

// ==================== 主机统计 ====================
export const getHostStats = (groupId) => api.get('/asset/hosts/stats', { params: { group_id: groupId } })
