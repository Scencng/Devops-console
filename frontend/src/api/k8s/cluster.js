import request from '../index.js'

export const getClusterList = (instanceId) => {
  return request({
    url: '/cluster/list',
    method: 'get',
    params: { instance_id: instanceId }
  })
}

export const addCluster = (data) => {
  return request({
    url: '/cluster/',
    method: 'post',
    data,
    headers: {
      'Content-Type': 'multipart/form-data'
    }
  })
}

export const changeCluster = (clusterName) => {
  return request({
    url: `/cluster/${clusterName}`,
    method: 'get'
  })
}

export const deleteCluster = (id) => {
  return request({
    url: `/cluster/${id}`,
    method: 'delete'
  })
}

export const testClusterConnection = (clusterName) => {
  return request({
    url: '/cluster/test',
    method: 'get',
    params: { clusterName }
  })
}
