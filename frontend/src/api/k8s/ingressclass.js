import request from '../index.js'

// IngressClass APIs
export const getIngressClassList = (instanceId) => {
    return request({
        url: '/k8s/ingressclass/list',
        method: 'get',
        params: {
            instance_id: instanceId
        }
    })
}

export const getIngressClassDetail = (name, instanceId) => {
    return request({
        url: `/k8s/ingressclass/detail/${name}`,
        method: 'get',
        params: {
            instance_id: instanceId
        }
    })
}
