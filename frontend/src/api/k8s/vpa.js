import request from '../index.js'

export const getVPAList = (namespace, instanceId) => {
    const url = namespace === 'all' ? '/k8s/vpa/list/all' : `/k8s/vpa/list/${namespace}`
    return request({
        url: url,
        method: 'get',
        params: {
            instance_id: instanceId
        }
    })
}

export const getVPADetail = (namespace, name, instanceId) => {
    return request({
        url: `/k8s/vpa/detail/${namespace}/${name}`,
        method: 'get',
        params: {
            instance_id: instanceId
        }
    })
}

export const deleteVPA = (namespace, name, instanceId) => {
    return request({
        url: `/k8s/vpa/delete/${namespace}/${name}`,
        method: 'delete',
        params: {
            instance_id: instanceId
        }
    })
}
