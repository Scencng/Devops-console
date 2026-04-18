import request from '../index.js'

// ConfigMap APIs
export const getConfigMapList = (namespace, instanceId) => {
    const url = namespace === 'all' ? '/k8s/configmap/list/all' : `/k8s/configmap/list/${namespace}`
    return request({
        url: url,
        method: 'get',
        params: {
            instance_id: instanceId
        }
    })
}

export const getConfigMapDetail = (namespace, name, instanceId) => {
    return request({
        url: `/k8s/configmap/detail/${namespace}/${name}`,
        method: 'get',
        params: {
            instance_id: instanceId
        }
    })
}

export const createConfigMap = (data, instanceId) => {
    return request({
        url: '/k8s/configmap/create',
        method: 'post',
        data,
        params: {
            instance_id: instanceId
        }
    })
}

export const updateConfigMap = (namespace, name, data, instanceId) => {
    return request({
        url: `/k8s/configmap/update/${namespace}/${name}`,
        method: 'put',
        data,
        params: {
            instance_id: instanceId
        }
    })
}

export const deleteConfigMap = (namespace, name, instanceId) => {
    return request({
        url: `/k8s/configmap/delete/${namespace}/${name}`,
        method: 'delete',
        params: {
            instance_id: instanceId
        }
    })
}
