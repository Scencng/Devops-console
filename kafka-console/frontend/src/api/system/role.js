import api from '../index.js'

export const getRoleList = (params) => api.get('/system/roles', { params })
export const getAllRoles = () => api.get('/system/roles/all')
export const createRole = (data) => api.post('/system/roles', data)
export const updateRole = (id, data) => api.put(`/system/roles/${id}`, data)
export const deleteRole = (id) => api.delete(`/system/roles/${id}`)
export const getRoleMenuIds = (id) => api.get(`/system/roles/${id}/menus`)
export const assignRoleMenus = (id, data) => api.put(`/system/roles/${id}/menus`, data)
