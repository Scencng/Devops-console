import request from '@/api/index'

export const getWorkflowList = (params) => {
    return request({
        url: '/task-scheduler/workflows',
        method: 'get',
        params
    })
}

export const getWorkflow = (id) => {
    return request({
        url: `/task-scheduler/workflows/${id}`,
        method: 'get'
    })
}

export const createWorkflow = (data) => {
    return request({
        url: '/task-scheduler/workflows',
        method: 'post',
        data
    })
}

export const updateWorkflow = (id, data) => {
    return request({
        url: `/task-scheduler/workflows/${id}`,
        method: 'put',
        data
    })
}

export const deleteWorkflow = (id) => {
    return request({
        url: `/task-scheduler/workflows/${id}`,
        method: 'delete'
    })
}

export const updateWorkflowStatus = (id, status) => {
    return request({
        url: `/task-scheduler/workflows/${id}/status`,
        method: 'put',
        data: { status }
    })
}

export const executeWorkflow = (id) => {
    return request({
        url: `/task-scheduler/workflows/${id}/execute`,
        method: 'post'
    })
}

export const getExecutionLogs = (id) => {
    return request({
        url: `/task-scheduler/executions/${id}/logs`,
        method: 'get'
    })
}

export const getExecutionList = (params) => {
    return request({
        url: '/task-scheduler/executions',
        method: 'get',
        params
    })
}

export const getExecutionDetail = (id) => {
    return request({
        url: `/task-scheduler/executions/${id}`,
        method: 'get'
    })
}
