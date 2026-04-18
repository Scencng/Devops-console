import request from '../index.js'

// PV APIs
export const getPvList = (instanceId) => {
    return request({
        url: '/k8s/pv/list',
        method: 'get',
        params: {
            instance_id: instanceId
        }
    })
}

export const getPvDetail = (pvName, instanceId) => {
    return request({
        url: `/k8s/pv/detail/${pvName}`,
        method: 'get',
        params: {
            instance_id: instanceId
        }
    })
}

export const createPv = (data, instanceId) => {
    return request({
        url: '/k8s/pv/create',
        method: 'post',
        data,
        params: {
            instance_id: instanceId
        }
    })
}

export const deletePv = (pvName, instanceId) => {
    return request({
        url: `/k8s/pv/delete/${pvName}`,
        method: 'delete',
        params: {
            instance_id: instanceId
        }
    })
}

// PVC APIs
export const getPvcList = (namespace, instanceId) => {
    const url = namespace === 'all' ? '/k8s/pvc/list/all' : `/k8s/pvc/list/${namespace}`
    return request({
        url: url,
        method: 'get',
        params: {
            instance_id: instanceId
        }
    })
}

export const getPvcDetail = (namespace, pvcName, instanceId) => {
    return request({
        url: `/k8s/pvc/detail/${namespace}/${pvcName}`,
        method: 'get',
        params: {
            instance_id: instanceId
        }
    })
}

export const createPvc = (data, instanceId) => {
    return request({
        url: '/k8s/pvc/create',
        method: 'post',
        data,
        params: {
            instance_id: instanceId
        }
    })
}

export const deletePvc = (namespace, pvcName, instanceId) => {
    return request({
        url: `/k8s/pvc/delete/${namespace}/${pvcName}`,
        method: 'delete',
        params: {
            instance_id: instanceId
        }
    })
}

// StorageClass APIs
export const getScList = (instanceId) => {
    return request({
        url: '/k8s/storageclass/list',
        method: 'get',
        params: {
            instance_id: instanceId
        }
    })
}

export const getScDetail = (scName, instanceId) => {
    return request({
        url: `/k8s/storageclass/detail/${scName}`,
        method: 'get',
        params: {
            instance_id: instanceId
        }
    })
}

export const createSc = (data, instanceId) => {
    return request({
        url: '/k8s/storageclass/create',
        method: 'post',
        data,
        params: {
            instance_id: instanceId
        }
    })
}

export const deleteSc = (scName, instanceId) => {
    return request({
        url: `/k8s/storageclass/delete/${scName}`,
        method: 'delete',
        params: {
            instance_id: instanceId
        }
    })
}
