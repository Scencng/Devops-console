import request from './index'

/**
 * Helm 仓库管理 API
 */

// 获取仓库列表
export const getHelmRepos = () => {
    return request({
        url: '/helm/repos',
        method: 'get'
    })
}

// 创建仓库
export const createHelmRepo = (data) => {
    return request({
        url: '/helm/repos',
        method: 'post',
        data
    })
}

// 更新仓库
export const updateHelmRepo = (id, data) => {
    return request({
        url: `/helm/repos/${id}`,
        method: 'put',
        data
    })
}

// 删除仓库
export const deleteHelmRepo = (id) => {
    return request({
        url: `/helm/repos/${id}`,
        method: 'delete'
    })
}

// 同步仓库
export const syncHelmRepo = (id) => {
    return request({
        url: `/helm/repos/${id}/sync`,
        method: 'post',
        timeout: 60000  // 同步操作可能较慢，设置 60 秒超时
    })
}

/**
 * Helm Chart 管理 API
 */

// 获取 Chart 列表（支持搜索和分页）
export const getHelmCharts = (params) => {
    return request({
        url: '/helm/charts',
        method: 'get',
        params
    })
}

// 获取 Chart 的所有版本
export const getChartVersions = (chartName, repoId) => {
    return request({
        url: `/helm/charts/${chartName}/versions`,
        method: 'get',
        params: { repo_id: repoId }
    })
}


// 获取 Chart 的默认 values.yaml
export const getChartDefaultValues = (chartName, repoId, version) => {
    return request({
        url: `/helm/charts/${chartName}/values`,
        method: 'get',
        params: {
            repo_id: repoId,
            version: version || ''
        }
    })
}
/**
 * Helm Release 管理 API
 */

// 安装 Chart
export const installHelmChart = (data, instanceId) => {
    return request({
        url: '/helm/install',
        method: 'post',
        data,
        params: { instance_id: instanceId }
    })
}

// 升级 Release
export const upgradeHelmRelease = (data, instanceId) => {
    return request({
        url: '/helm/upgrade',
        method: 'put',
        data,
        params: { instance_id: instanceId }
    })
}

// 卸载 Release
export const uninstallHelmRelease = (namespace, name, instanceId) => {
    return request({
        url: `/helm/releases/${namespace}/${name}`,
        method: 'delete',
        params: { instance_id: instanceId }
    })
}

// 获取已安装的 Release 列表
export const getHelmReleases = (instanceId, namespace) => {
    return request({
        url: '/helm/releases',
        method: 'get',
        params: {
            instance_id: instanceId,
            namespace: namespace || ''
        }
    })
}

// 获取 Release 详情
export const getHelmReleaseDetail = (namespace, name, instanceId) => {
    return request({
        url: `/helm/releases/${namespace}/${name}`,
        method: 'get',
        params: { instance_id: instanceId }
    })
}
