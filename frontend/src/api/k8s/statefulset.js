import request from '../index.js'

export function getStatefulSetList(namespace, instanceId) {
    return request({
        url: `/k8s/statefulset/list`,
        method: 'get',
        params: { namespace, instanceId }
    })
}

export function getStatefulSetDetail(namespace, name, instanceId) {
    return request({
        url: `/k8s/statefulset/detail`,
        method: 'get',
        params: { namespace, name, instanceId }
    })
}

export function createStatefulSet(data, instanceId) {
    return request({
        url: `/k8s/statefulset/create`,
        method: 'post',
        data,
        params: { instanceId }
    })
}

export function updateStatefulSet(namespace, name, data, instanceId) {
    return request({
        url: `/k8s/statefulset/update`,
        method: 'put',
        data,
        params: { namespace, name, instanceId }
    })
}

export function deleteStatefulSet(namespace, name, instanceId) {
    return request({
        url: `/k8s/statefulset/delete`,
        method: 'delete',
        params: { namespace, name, instanceId }
    })
}

export function scaleStatefulSet(namespace, name, replicas, instanceId) {
    return request({
        url: `/k8s/statefulset/scale`,
        method: 'put',
        data: { replicas },
        params: { namespace, name, instanceId }
    })
}
