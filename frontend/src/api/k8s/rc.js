import request from '../index.js'

export const getRCList = (namespace, instanceId) => {
    const url = namespace === 'all' ? '/k8s/rc/list/all' : `/k8s/rc/list/${namespace}`
    return request({
        url: url,
        method: 'get',
        params: {
            instance_id: instanceId
        }
    })
}

export const getRCDetail = (namespace, name, instanceId) => {
    return request({
        url: `/k8s/rc/detail/${namespace}/${name}`,
        method: 'get',
        params: {
            instance_id: instanceId
        }
    })
}

export const deleteRC = (namespace, name, instanceId) => {
    return request({
        url: `/k8s/rc/delete/${namespace}/${name}`,
        method: 'delete',
        params: {
            instance_id: instanceId
        }
    })
}
