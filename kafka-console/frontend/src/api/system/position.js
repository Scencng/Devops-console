import api from '../index.js'

export const getPositionList = (params) => api.get('/system/positions', { params })
export const getAllPositions = () => api.get('/system/positions/all')
export const createPosition = (data) => api.post('/system/positions', data)
export const updatePosition = (id, data) => api.put(`/system/positions/${id}`, data)
export const deletePosition = (id) => api.delete(`/system/positions/${id}`)
