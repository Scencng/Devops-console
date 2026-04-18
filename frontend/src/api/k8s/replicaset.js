import request from '../index.js'

export const getReplicaSetList = (namespace, instanceId) => {
    const url = namespace === 'all' ? '/k8s/replicaset/list/all' : `/k8s/replicaset/list/${namespace}`
    return request({
        url: url,
        method: 'get',
        params: {
            instance_id: instanceId
        }
    })
}

export const getReplicaSetDetail = (namespace, name, instanceId) => {
    return request({
        url: `/k8s/replicaset/detail/${namespace}/${name}`,
        method: 'get',
        params: {
            instance_id: instanceId
        }
    })
}

export const deleteReplicaSet = (namespace, name, instanceId) => {
    return request({
        url: `/k8s/replicaset/delete/${namespace}/${name}`,
        method: 'delete',
        params: {
            instance_id: instanceId
        }
    })
}
