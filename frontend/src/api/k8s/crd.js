import request from '../index.js'

export const getCRDList = (instanceId) => {
    return request({
        url: '/k8s/crd/list',
        method: 'get',
        params: {
            instance_id: instanceId
        }
    })
}

export const getCRDDetail = (name, instanceId) => {
    return request({
        url: `/k8s/crd/detail/${name}`,
        method: 'get',
        params: {
            instance_id: instanceId
        }
    })
}

export const deleteCRD = (name, instanceId) => {
    return request({
        url: `/k8s/crd/delete/${name}`,
        method: 'delete',
        params: {
            instance_id: instanceId
        }
    })
}
