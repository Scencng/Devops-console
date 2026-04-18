import api from './index'

// ==================== 域名监控统计 ====================
export const getDomainStats = () => api.get('/monitor/domain/stats')

// ==================== 域名列表管理 ====================
export const getDomainList = (params) => api.get('/monitor/domain', { params })
export const createDomain = (data) => api.post('/monitor/domain', data)
export const updateDomain = (id, data) => api.put(`/monitor/domain/${id}`, data)
export const deleteDomain = (id) => api.delete(`/monitor/domain/${id}`)
export const toggleDomainEnabled = (id, data) => api.put(`/monitor/domain/${id}/toggle`, data)

// ==================== SSL 证书申请 ====================
export const getSslCertList = (params) => api.get('/monitor/domain/ssl-certs', { params })
export const applySslCert = (data) => api.post('/monitor/domain/ssl-certs/apply', data)
export const uploadSslCert = (data) => api.post('/monitor/domain/ssl-certs/upload', data)
export const deleteSslCert = (id) => api.delete(`/monitor/domain/ssl-certs/${id}`)
export const downloadSslCert = (id) => api.get(`/monitor/domain/ssl-certs/${id}/download`, { responseType: 'blob' })

// ==================== DNS 云厂商配置 ====================
export const getDnsProviderList = (params) => api.get('/monitor/domain/dns-providers', { params })
export const createDnsProvider = (data) => api.post('/monitor/domain/dns-providers', data)
export const updateDnsProvider = (id, data) => api.put(`/monitor/domain/dns-providers/${id}`, data)
export const deleteDnsProvider = (id) => api.delete(`/monitor/domain/dns-providers/${id}`)
export const testDnsProvider = (id) => api.post(`/monitor/domain/dns-providers/${id}/test`)
export const toggleDnsProvider = (id, data) => api.put(`/monitor/domain/dns-providers/${id}/toggle`, data)
