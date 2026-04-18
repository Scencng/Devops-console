import api from './index.js'

export const getKafkaClusters = (params) => api.get('/kafka/clusters', { params })
export const getKafkaCluster = (id) => api.get(`/kafka/clusters/${id}`)
export const createKafkaCluster = (data) => api.post('/kafka/clusters', data)
export const updateKafkaCluster = (id, data) => api.put(`/kafka/clusters/${id}`, data)
export const deleteKafkaCluster = (id) => api.delete(`/kafka/clusters/${id}`)
export const testKafkaCluster = (id) => api.post(`/kafka/clusters/${id}/test`)
export const getKafkaClusterOptions = () => api.get('/kafka/clusters/options')

export const getKafkaDashboard = (clusterId) => api.get('/kafka/dashboard', { params: { clusterId } })

export const scanKafkaNetwork = (data) => api.post('/kafka/discovery/scan', data)
export const importKafkaDiscoveryResult = (data) => api.post('/kafka/discovery/import', data)

export const createKafkaTopic = (data) => api.post('/kafka/topics', data)
export const getKafkaTopics = (params) => api.get('/kafka/topics', { params })
export const deleteKafkaTopic = (clusterId, topic) => api.delete(`/kafka/topics/${encodeURIComponent(topic)}`, { params: { clusterId } })
export const updateKafkaTopicConfig = (topic, data) => api.put(`/kafka/topics/${encodeURIComponent(topic)}/config`, data)
export const getKafkaTopicPartitions = (clusterId, topic) => api.get(`/kafka/topics/${encodeURIComponent(topic)}/partitions`, { params: { clusterId } })
export const increaseKafkaTopicPartitions = (topic, data) => api.post(`/kafka/topics/${encodeURIComponent(topic)}/partitions`, data)

export const getKafkaBrokers = (clusterId) => api.get('/kafka/brokers', { params: { clusterId } })

export const getKafkaConsumerGroups = (params) => api.get('/kafka/consumer-groups', { params })
export const getKafkaConsumerGroupDetail = (groupId, params) => api.get(`/kafka/consumer-groups/${encodeURIComponent(groupId)}`, { params })
export const resetKafkaGroupOffset = (groupId, data) => api.post(`/kafka/consumer-groups/${encodeURIComponent(groupId)}/reset-offset`, data)

export const getKafkaMessages = (params) => api.get('/kafka/messages', { params })
export const produceKafkaMessage = (data) => api.post('/kafka/messages/produce', data)

export const getKafkaAuditLogs = (params) => api.get('/kafka/audit-logs', { params })
