import request from '../index.js'

// Secret APIs
export const getSecretList = (namespace, instanceId) => {
    const url = namespace === 'all' ? '/k8s/secret/list/all' : `/k8s/secret/list/${namespace}`
    return request({
        url: url,
        method: 'get',
        params: {
            instance_id: instanceId
        }
    })
}

export const getSecretDetail = (namespace, name, instanceId) => {
    return request({
        url: `/k8s/secret/detail/${namespace}/${name}`,
        method: 'get',
        params: {
            instance_id: instanceId
        }
    })
}

export const createSecret = (data, instanceId) => {
    return request({
        url: '/k8s/secret/create',
        method: 'post',
        data,
        params: {
            instance_id: instanceId
        }
    })
}

export const updateSecret = (namespace, name, data, instanceId) => {
    return request({
        url: `/k8s/secret/update/${namespace}/${name}`,
        method: 'put',
        data,
        params: {
            instance_id: instanceId
        }
    })
}

export const deleteSecret = (namespace, name, instanceId) => {
    return request({
        url: `/k8s/secret/delete/${namespace}/${name}`,
        method: 'delete',
        params: {
            instance_id: instanceId
        }
    })
}
