import api from '../index.js'

export const getMenuTree = () => api.get('/system/menus')
export const createMenu = (data) => api.post('/system/menus', data)
export const updateMenu = (id, data) => api.put(`/system/menus/${id}`, data)
export const deleteMenu = (id) => api.delete(`/system/menus/${id}`)
