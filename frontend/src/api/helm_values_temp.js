// 获取 Chart 的默认 values.yaml
export const getChartDefaultValues = (chartName, repoId, version) => {
    return request({
        url: `/helm/charts/${chartName}/values`,
        method: 'get',
        params: {
            repo_id: repoId,
            version: version
        }
    })
}
