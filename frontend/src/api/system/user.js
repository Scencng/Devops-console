import api from '../index.js'

// ==================== 登录 ====================
export const login = (data) => api.post('/system/login', data)

// ==================== 当前用户 ====================
export const getAuthInfo = () => api.get('/system/auth/info')
export const updateProfile = (data) => api.put('/system/auth/profile', data)
export const changePassword = (data) => api.put('/system/auth/password', data)

// ==================== 用户管理 ====================
export const getUserList = (params) => api.get('/system/users', { params })
export const getUserById = (id) => api.get(`/system/users/${id}`)
export const createUser = (data) => api.post('/system/users', data)
export const updateUser = (id, data) => api.put(`/system/users/${id}`, data)
export const deleteUser = (id) => api.delete(`/system/users/${id}`)
export const updateUserStatus = (id, data) => api.put(`/system/users/${id}/status`, data)
export const resetUserPassword = (id, data) => api.put(`/system/users/${id}/reset-password`, data)
