import request from '../index.js'

// Event APIs
export const getEventList = (namespace, instanceId) => {
    const url = namespace === 'all' ? '/k8s/event/list/all' : `/k8s/event/list/${namespace}`
    return request({
        url: url,
        method: 'get',
        params: {
            instance_id: instanceId
        }
    })
}
