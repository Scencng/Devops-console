import api from '../index.js'

export const getDeptTree = (params) => api.get('/system/depts', { params })
export const createDept = (data) => api.post('/system/depts', data)
export const updateDept = (id, data) => api.put(`/system/depts/${id}`, data)
export const deleteDept = (id) => api.delete(`/system/depts/${id}`)
