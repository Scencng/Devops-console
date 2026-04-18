import request from '@/api/index'

// Projects
export const getProjects = (params) => {
    return request({
        url: '/cicd/projects',
        method: 'get',
        params
    })
}

// Pipelines
export const getPipelines = (params) => {
    return request({
        url: '/pipelines/page',
        method: 'get',
        params
    })
}

export const getPipeline = (id) => {
    return request({
        url: `/pipelines/${id}`,
        method: 'get'
    })
}

export const createPipeline = (data) => {
    return request({
        url: '/pipelines/',
        method: 'post',
        data
    })
}

export const updatePipeline = (id, data) => {
    return request({
        url: `/pipelines/${id}`,
        method: 'put',
        data
    })
}

export const deletePipeline = (id) => {
    return request({
        url: `/pipelines/${id}`,
        method: 'delete'
    })
}

export const triggerPipeline = (id, data) => {
    return request({
        url: `/pipelines/${id}/run`,
        method: 'post',
        data
    })
}

// Pipeline Runs
export const getPipelineRuns = (params) => {
    return request({
        url: '/pipeline-runs/page',
        method: 'get',
        params
    })
}

export const getPipelineRun = (id) => {
    return request({
        url: `/pipeline-runs/${id}`,
        method: 'get'
    })
}

export const deletePipelineRun = (id) => {
    return request({
        url: `/pipeline-runs/${id}`,
        method: 'delete'
    })
}

// Pipeline Steps
export const getPipelineRunSteps = (runId) => {
    return request({
        url: `/pipeline-runs/${runId}/steps`,
        method: 'get'
    })
}

// Pipeline Steps Management
export const getPipelineSteps = (pipelineId) => {
    return request({
        url: '/pipeline-steps/list',
        method: 'get',
        params: { pipelineId }
    })
}

export const createPipelineStep = (data) => {
    return request({
        url: '/pipeline-steps/',
        method: 'post',
        data
    })
}

export const updatePipelineStep = (data) => {
    return request({
        url: '/pipeline-steps/',
        method: 'put',
        data
    })
}

export const deletePipelineStep = (id) => {
    return request({
        url: `/pipeline-steps/${id}`,
        method: 'delete'
    })
}

export const getPipelineRunLogs = (runId, stepName) => {
    return request({
        url: `/pipeline-runs/${runId}/logs`,
        method: 'get',
        params: { step_name: stepName }
    })
}
