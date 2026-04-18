import request from '../index.js'

export const getSubscriptionList = (namespace, instanceId) => {
    const url = namespace === 'all' ? '/k8s/operator/subscription/list/all' : `/k8s/operator/subscription/list/${namespace}`
    return request({
        url: url,
        method: 'get',
        params: {
            instance_id: instanceId
        }
    })
}

export const getSubscriptionDetail = (namespace, name, instanceId) => {
    return request({
        url: `/k8s/operator/subscription/detail/${namespace}/${name}`,
        method: 'get',
        params: {
            instance_id: instanceId
        }
    })
}

export const deleteSubscription = (namespace, name, instanceId) => {
    return request({
        url: `/k8s/operator/subscription/delete/${namespace}/${name}`,
        method: 'delete',
        params: {
            instance_id: instanceId
        }
    })
}
