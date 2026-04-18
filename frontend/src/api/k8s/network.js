import request from '../index.js'

// Ingress APIs
export const getIngressList = (namespace, instanceId) => {
    const url = namespace === 'all' ? '/k8s/ingress/list/all' : `/k8s/ingress/list/${namespace}`
    return request({
        url: url,
        method: 'get',
        params: {
            instance_id: instanceId
        }
    })
}

export const getIngressDetail = (namespace, name, instanceId) => {
    return request({
        url: `/k8s/ingress/detail/${namespace}/${name}`,
        method: 'get',
        params: {
            instance_id: instanceId
        }
    })
}

export const createIngress = (data, instanceId) => {
    return request({
        url: '/k8s/ingress/create',
        method: 'post',
        data,
        params: {
            instance_id: instanceId
        }
    })
}

export const updateIngress = (namespace, name, data, instanceId) => {
    return request({
        url: `/k8s/ingress/update/${namespace}/${name}`,
        method: 'put',
        data,
        params: {
            instance_id: instanceId
        }
    })
}

export const deleteIngress = (namespace, name, instanceId) => {
    return request({
        url: `/k8s/ingress/delete/${namespace}/${name}`,
        method: 'delete',
        params: {
            instance_id: instanceId
        }
    })
}
