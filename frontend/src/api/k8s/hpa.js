import request from '../index.js'

export const getHPAList = (namespace, instanceId) => {
    const url = namespace === 'all' ? '/k8s/hpa/list/all' : `/k8s/hpa/list/${namespace}`
    return request({
        url: url,
        method: 'get',
        params: {
            instance_id: instanceId
        }
    })
}

export const getHPADetail = (namespace, name, instanceId) => {
    return request({
        url: `/k8s/hpa/detail/${namespace}/${name}`,
        method: 'get',
        params: {
            instance_id: instanceId
        }
    })
}

export const deleteHPA = (namespace, name, instanceId) => {
    return request({
        url: `/k8s/hpa/delete/${namespace}/${name}`,
        method: 'delete',
        params: {
            instance_id: instanceId
        }
    })
}
