import request from '../index.js'

// ==================== Role ====================
export const getRoleList = (namespace, instanceId) => {
    const url = namespace === 'all' ? '/k8s/role/list/all' : `/k8s/role/list/${namespace}`
    return request({ url, method: 'get', params: { instance_id: instanceId } })
}

export const getRoleDetail = (namespace, name, instanceId) => {
    return request({ url: `/k8s/role/detail/${namespace}/${name}`, method: 'get', params: { instance_id: instanceId } })
}

export const createRole = (data, instanceId) => {
    return request({ url: '/k8s/role/create', method: 'post', data, params: { instance_id: instanceId } })
}

export const updateRole = (namespace, name, data, instanceId) => {
    return request({ url: `/k8s/role/update/${namespace}/${name}`, method: 'put', data, params: { instance_id: instanceId } })
}

export const deleteRole = (namespace, name, instanceId) => {
    return request({ url: `/k8s/role/delete/${namespace}/${name}`, method: 'delete', params: { instance_id: instanceId } })
}

// ==================== ClusterRole ====================
export const getClusterRoleList = (instanceId) => {
    return request({ url: '/k8s/clusterrole/list', method: 'get', params: { instance_id: instanceId } })
}

export const getClusterRoleDetail = (name, instanceId) => {
    return request({ url: `/k8s/clusterrole/detail/${name}`, method: 'get', params: { instance_id: instanceId } })
}

export const createClusterRole = (data, instanceId) => {
    return request({ url: '/k8s/clusterrole/create', method: 'post', data, params: { instance_id: instanceId } })
}

export const updateClusterRole = (name, data, instanceId) => {
    return request({ url: `/k8s/clusterrole/update/${name}`, method: 'put', data, params: { instance_id: instanceId } })
}

export const deleteClusterRole = (name, instanceId) => {
    return request({ url: `/k8s/clusterrole/delete/${name}`, method: 'delete', params: { instance_id: instanceId } })
}

// ==================== RoleBinding ====================
export const getRoleBindingList = (namespace, instanceId) => {
    const url = namespace === 'all' ? '/k8s/rolebinding/list/all' : `/k8s/rolebinding/list/${namespace}`
    return request({ url, method: 'get', params: { instance_id: instanceId } })
}

export const getRoleBindingDetail = (namespace, name, instanceId) => {
    return request({ url: `/k8s/rolebinding/detail/${namespace}/${name}`, method: 'get', params: { instance_id: instanceId } })
}

export const createRoleBinding = (data, instanceId) => {
    return request({ url: '/k8s/rolebinding/create', method: 'post', data, params: { instance_id: instanceId } })
}

export const updateRoleBinding = (namespace, name, data, instanceId) => {
    return request({ url: `/k8s/rolebinding/update/${namespace}/${name}`, method: 'put', data, params: { instance_id: instanceId } })
}

export const deleteRoleBinding = (namespace, name, instanceId) => {
    return request({ url: `/k8s/rolebinding/delete/${namespace}/${name}`, method: 'delete', params: { instance_id: instanceId } })
}

// ==================== ClusterRoleBinding ====================
export const getClusterRoleBindingList = (instanceId) => {
    return request({ url: '/k8s/clusterrolebinding/list', method: 'get', params: { instance_id: instanceId } })
}

export const getClusterRoleBindingDetail = (name, instanceId) => {
    return request({ url: `/k8s/clusterrolebinding/detail/${name}`, method: 'get', params: { instance_id: instanceId } })
}

export const createClusterRoleBinding = (data, instanceId) => {
    return request({ url: '/k8s/clusterrolebinding/create', method: 'post', data, params: { instance_id: instanceId } })
}

export const updateClusterRoleBinding = (name, data, instanceId) => {
    return request({ url: `/k8s/clusterrolebinding/update/${name}`, method: 'put', data, params: { instance_id: instanceId } })
}

export const deleteClusterRoleBinding = (name, instanceId) => {
    return request({ url: `/k8s/clusterrolebinding/delete/${name}`, method: 'delete', params: { instance_id: instanceId } })
}
