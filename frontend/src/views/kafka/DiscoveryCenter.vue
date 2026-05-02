<template>
  <div class="page-container">
    <el-card class="page-header-card" shadow="never">
      <div class="page-header">
        <div class="page-header-copy">
          <div class="page-eyebrow">Kafka</div>
          <h2>自动发现</h2>
          <p>先扫描候选入口，再按 Cluster ID 聚合结果，确认版本和认证方式后再统一导入。</p>
        </div>

        <div class="page-header-side">
          <div class="page-header-meta">
            <div class="page-header-kpi">
              <span>已识别集群</span>
              <strong>{{ clusterSummaries.length }}</strong>
            </div>
            <div class="page-header-kpi">
              <span>已勾选</span>
              <strong>{{ selectedCount }}</strong>
            </div>
          </div>
          <div class="page-header-note">
            {{ importPrecheckItems.filter((item) => !item.passed).length > 0
              ? `当前还有 ${importPrecheckItems.filter((item) => !item.passed).length} 项导入前检查未通过。`
              : '当前导入前检查已通过，可以继续确认版本和批量导入。' }}
          </div>
        </div>
      </div>
    </el-card>

    <el-card class="content-card">
      <template #header>
        <div class="card-header">
          <span>扫描条件</span>
          <el-button text type="primary" @click="showAdvancedAuth = !showAdvancedAuth">
            {{ showAdvancedAuth ? '收起高级认证参数' : '展开高级认证参数' }}
          </el-button>
        </div>
      </template>

      <el-form label-position="top">
        <el-row :gutter="16">
          <el-col :xs="24" :lg="8">
            <el-form-item label="CIDR 网段">
              <el-input v-model="scanForm.cidr" placeholder="例如 192.168.1.0/24" />
            </el-form-item>
          </el-col>
          <el-col :xs="24" :lg="8">
            <el-form-item label="端口列表">
              <el-input v-model="portsInput" placeholder="例如 9092,9093,29092" />
            </el-form-item>
          </el-col>
          <el-col :xs="12" :lg="4">
            <el-form-item label="超时(ms)">
              <el-input-number v-model="scanForm.timeoutMs" :min="200" :max="30000" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :xs="12" :lg="4">
            <el-form-item label="并发">
              <el-input-number v-model="scanForm.concurrency" :min="1" :max="1024" style="width: 100%" />
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="16">
          <el-col :xs="24" :md="8" :lg="5">
            <el-form-item label="认证模板">
              <el-select v-model="authMode" style="width: 100%">
                <el-option label="无认证" value="none" />
                <el-option label="SASL/PLAIN" value="plain" />
                <el-option label="SCRAM-SHA-256" value="scram_sha256" />
                <el-option label="SCRAM-SHA-512" value="scram_sha512" />
                <el-option label="TLS" value="tls" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :xs="24" :md="8" :lg="5">
            <el-form-item label="Kafka 版本">
              <el-input v-model="scanForm.auth.version" placeholder="留空自动探测，例如 3.9.0" />
            </el-form-item>
          </el-col>
          <el-col :xs="24" :md="8" :lg="14" class="scan-actions-col">
            <el-form-item class="scan-actions-item">
              <div class="scan-actions">
                <el-button type="primary" :loading="loading" @click="runScan">开始扫描</el-button>
                <span class="scan-hint">建议先留空版本，让系统优先自动探测。</span>
              </div>
            </el-form-item>
          </el-col>
        </el-row>

        <el-collapse-transition>
          <div v-show="showAdvancedAuth" class="advanced-auth">
            <div class="advanced-title">高级认证参数</div>
            <el-row :gutter="16">
              <el-col :xs="24" :md="8">
                <el-form-item label="用户名">
                  <el-input
                    v-model="scanForm.auth.username"
                    :disabled="authMode === 'none' || authMode === 'tls'"
                  />
                </el-form-item>
              </el-col>
              <el-col :xs="24" :md="8">
                <el-form-item label="密码">
                  <el-input
                    v-model="scanForm.auth.password"
                    type="password"
                    show-password
                    :disabled="authMode === 'none' || authMode === 'tls'"
                  />
                </el-form-item>
              </el-col>
              <el-col :xs="12" :md="4">
                <el-form-item label="TLS">
                  <el-switch v-model="scanForm.auth.tlsEnabled" />
                </el-form-item>
              </el-col>
              <el-col :xs="12" :md="4">
                <el-form-item label="跳过校验">
                  <el-switch v-model="scanForm.auth.insecureSkipVerify" />
                </el-form-item>
              </el-col>
            </el-row>

            <el-form-item v-if="scanForm.auth.tlsEnabled" label="CA 证书">
              <el-input v-model="scanForm.auth.caCert" type="textarea" :rows="4" />
            </el-form-item>
            <el-form-item v-if="scanForm.auth.tlsEnabled" label="客户端证书">
              <el-input v-model="scanForm.auth.clientCert" type="textarea" :rows="4" />
            </el-form-item>
            <el-form-item v-if="scanForm.auth.tlsEnabled" label="客户端私钥">
              <el-input v-model="scanForm.auth.clientKey" type="textarea" :rows="4" />
            </el-form-item>
          </div>
        </el-collapse-transition>
      </el-form>
    </el-card>

    <el-card class="content-card">
      <template #header>
        <div class="card-header">
          <span>补充入口</span>
          <span class="result-subtitle">域名 / Bootstrap</span>
        </div>
      </template>

      <el-form label-position="top">
        <el-row :gutter="16">
          <el-col :xs="24" :lg="14">
            <el-form-item label="域名 / Bootstrap Servers">
              <el-input
                v-model="domainImportForm.address"
                placeholder="例如 kafka.example.com:9092 或 kafka-1:9092,kafka-2:9092"
              />
            </el-form-item>
          </el-col>
          <el-col :xs="12" :lg="4">
            <el-form-item label="超时(ms)">
              <el-input-number v-model="domainImportForm.timeoutMs" :min="200" :max="30000" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :xs="12" :lg="6">
            <el-form-item label="Kafka 版本">
              <el-input v-model="domainImportForm.auth.version" placeholder="留空自动探测，例如 3.9.0" />
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="16">
          <el-col :xs="24" :md="10">
            <el-form-item label="认证方式">
              <el-select v-model="domainImportForm.auth.authType" style="width: 100%">
                <el-option label="无认证" value="none" />
                <el-option label="SASL/PLAIN" value="plain" />
                <el-option label="SCRAM-SHA-256" value="scram_sha256" />
                <el-option label="SCRAM-SHA-512" value="scram_sha512" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :xs="24" :md="14" class="scan-actions-col">
            <el-form-item class="scan-actions-item">
              <div class="scan-actions">
                <el-button type="primary" :loading="domainImporting" @click="probeByDomain">识别并合并</el-button>
                <span class="scan-hint">识别后自动并入结果。</span>
              </div>
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="16">
          <el-col :xs="24" :md="8">
            <el-form-item label="用户名">
              <el-input v-model="domainImportForm.auth.username" :disabled="domainImportForm.auth.authType === 'none'" />
            </el-form-item>
          </el-col>
          <el-col :xs="24" :md="8">
            <el-form-item label="密码">
              <el-input
                v-model="domainImportForm.auth.password"
                type="password"
                show-password
                :disabled="domainImportForm.auth.authType === 'none'"
              />
            </el-form-item>
          </el-col>
          <el-col :xs="12" :md="4">
            <el-form-item label="TLS">
              <el-switch v-model="domainImportForm.auth.tlsEnabled" />
            </el-form-item>
          </el-col>
          <el-col :xs="12" :md="4">
            <el-form-item label="跳过校验">
              <el-switch v-model="domainImportForm.auth.insecureSkipVerify" :disabled="!domainImportForm.auth.tlsEnabled" />
            </el-form-item>
          </el-col>
        </el-row>

        <el-form-item v-if="domainImportForm.auth.tlsEnabled" label="CA 证书">
          <el-input v-model="domainImportForm.auth.caCert" type="textarea" :rows="3" />
        </el-form-item>
        <el-form-item v-if="domainImportForm.auth.tlsEnabled" label="客户端证书">
          <el-input v-model="domainImportForm.auth.clientCert" type="textarea" :rows="3" />
        </el-form-item>
        <el-form-item v-if="domainImportForm.auth.tlsEnabled" label="客户端私钥">
          <el-input v-model="domainImportForm.auth.clientKey" type="textarea" :rows="3" />
        </el-form-item>
      </el-form>
    </el-card>

    <el-card
      v-if="clusterSummaries.length || loading || domainImporting"
      class="content-card"
      v-loading="loading || domainImporting"
      element-loading-text="正在整理发现结果..."
    >
      <template #header>
        <div class="card-header">
          <span>发现摘要</span>
          <span class="card-subtitle">导入前信号</span>
        </div>
      </template>

      <div class="page-metrics summary-row">
        <div v-for="card in summaryCards" :key="card.label" class="summary-panel">
          <span class="summary-label">{{ card.label }}</span>
          <strong class="summary-value">{{ card.value }}</strong>
          <span class="summary-desc">{{ card.desc }}</span>
        </div>
      </div>

      <div class="compact-list">
        <div class="compact-item">
          <div>
            <strong>版本待确认</strong>
            <span>{{ discoveryRiskSummary.versionPending }} 个集群仍需人工确认版本。</span>
          </div>
          <el-tag :type="discoveryRiskSummary.versionPending > 0 ? 'warning' : 'success'">
            {{ discoveryRiskSummary.versionPending > 0 ? '关注' : '正常' }}
          </el-tag>
        </div>
        <div class="compact-item">
          <div>
            <strong>访问入口混入</strong>
            <span>{{ discoveryRiskSummary.accessEntryClusters }} 个集群同时包含 Broker 节点和访问入口。</span>
          </div>
          <el-tag :type="discoveryRiskSummary.accessEntryClusters > 0 ? 'warning' : 'success'">
            {{ discoveryRiskSummary.accessEntryClusters > 0 ? '确认' : '正常' }}
          </el-tag>
        </div>
        <div class="compact-item">
          <div>
            <strong>非 Kafka 结果</strong>
            <span>{{ discoveryRiskSummary.nonKafkaClusters }} 个结果当前未识别为 Kafka 集群。</span>
          </div>
          <el-tag :type="discoveryRiskSummary.nonKafkaClusters > 0 ? 'info' : 'success'">
            {{ discoveryRiskSummary.nonKafkaClusters > 0 ? '可忽略' : '正常' }}
          </el-tag>
        </div>
      </div>

      <div v-if="duplicateClusterHints.length" class="duplicate-hints">
        <div class="duplicate-hints-title">重复与已导入入口</div>
        <div class="compact-list">
          <div v-for="item in duplicateClusterHints" :key="item.key" class="compact-item">
            <div>
              <strong>{{ item.title }}</strong>
              <span>{{ item.description }}</span>
            </div>
            <el-tag :type="item.type === 'imported' ? 'warning' : 'info'">
              {{ item.type === 'imported' ? '已存在' : '重复' }}
            </el-tag>
          </div>
        </div>
      </div>
    </el-card>

    <el-card v-if="clusterSummaries.length" class="content-card">
      <template #header>
        <div class="card-header card-header-wrap">
          <div>
            <span>发现结果</span>
            <span class="result-subtitle">
              {{ filteredClusters.length }} / {{ clusterSummaries.length }} 个集群可见，当前渲染 {{ pagedClusters.length }} 个，已勾选 {{ selectedCount }} 个
            </span>
          </div>
          <div class="result-filters">
            <el-input
              v-model="filterForm.keyword"
              placeholder="搜索 Cluster ID、节点地址、错误信息"
              clearable
              class="filter-input"
            />
            <el-select v-model="filterForm.scope" class="filter-select">
              <el-option label="全部集群" value="all" />
              <el-option label="Kafka 候选" value="kafka" />
              <el-option label="版本已识别" value="detected" />
              <el-option label="待确认" value="version-failed" />
            </el-select>
            <el-button @click="clearSelectedClusters" :disabled="!selectedClusters.length">清空勾选</el-button>
            <el-button type="primary" @click="openBatchImportDialog" :disabled="!selectedClusters.length">
              批量导入已选
            </el-button>
          </div>
        </div>
      </template>

      <el-table
        ref="resultTableRef"
        :data="pagedClusters"
        row-key="key"
        empty-text="暂无发现结果"
        class="result-table"
        @selection-change="handleTableSelectionChange"
      >
        <el-table-column
          type="selection"
          width="56"
          fixed="left"
          :selectable="isSelectableRow"
          reserve-selection
        />

        <el-table-column label="集群 / 入口" min-width="240">
          <template #default="{ row }">
            <div class="result-table-cell">
              <strong>{{ row.clusterId || '未返回 Cluster ID' }}</strong>
              <span>{{ row.bootstrapServers || '-' }}</span>
            </div>
          </template>
        </el-table-column>

        <el-table-column label="类型" width="180">
          <template #default="{ row }">
            <div class="result-table-tags">
              <el-tag :type="row.looksLikeKafka ? 'success' : 'info'" effect="plain">
                {{ row.looksLikeKafka ? 'Kafka 集群' : '非 Kafka' }}
              </el-tag>
              <el-tag v-if="isImportedCluster(row)" type="info" effect="plain">
                已导入
              </el-tag>
            </div>
          </template>
        </el-table-column>

        <el-table-column label="版本" width="140">
          <template #default="{ row }">
            <el-tag v-if="row.kafkaVersion" type="success" effect="plain">{{ row.kafkaVersion }}</el-tag>
            <el-tag v-else-if="row.versionDetectError" type="warning" effect="plain">待确认</el-tag>
            <span v-else class="text-muted">-</span>
          </template>
        </el-table-column>

        <el-table-column label="规模" min-width="220">
          <template #default="{ row }">
            <div class="result-metrics-inline">
              <span class="result-metric-chip">Broker {{ row.brokerCount }}</span>
              <span class="result-metric-chip">Controller {{ row.controllerId ?? '-' }}</span>
              <span class="result-metric-chip">监听器 {{ row.listenerCount || 0 }}</span>
              <span class="result-metric-chip">入口 {{ row.accessEntryCount }}</span>
            </div>
          </template>
        </el-table-column>

        <el-table-column label="摘要" min-width="320" show-overflow-tooltip>
          <template #default="{ row }">
            <div class="result-summary">
              {{ buildRowSummary(row) }}
            </div>
          </template>
        </el-table-column>

        <el-table-column label="操作" width="220" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="openResultDetail(row.key)">详情</el-button>
            <el-button
              link
              type="primary"
              :disabled="!row.looksLikeKafka || isImportedCluster(row)"
              @click="openImportDialog(row)"
            >
              {{ isImportedCluster(row) ? '已导入' : row.versionDetectError ? '确认版本后导入' : '导入' }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <div v-if="filteredClusters.length" class="result-pagination">
        <el-pagination
          background
          layout="sizes, prev, pager, next, total"
          :hide-on-single-page="true"
          :total="filteredClusters.length"
          :current-page="discoveryPage.page"
          :page-size="discoveryPage.pageSize"
          :page-sizes="[24, 48, 96, 192]"
          @current-change="handleDiscoveryPageChange"
          @size-change="handleDiscoveryPageSizeChange"
        />
      </div>
    </el-card>

    <el-dialog
      v-model="resultDetailVisible"
      :title="`发现详情：${resultDetail?.clusterId || '未返回 Cluster ID'}`"
      width="min(980px, calc(100vw - 32px))"
      destroy-on-close
    >
      <div v-if="resultDetail" class="detail-dialog-content">
        <div class="detail-summary-grid">
          <div class="detail-summary-card">
            <span>类型</span>
            <strong>{{ resultDetail.looksLikeKafka ? 'Kafka' : '非 Kafka' }}</strong>
          </div>
          <div class="detail-summary-card">
            <span>Broker 节点</span>
            <strong>{{ resultDetail.brokerCount }}</strong>
          </div>
          <div class="detail-summary-card">
            <span>监听器</span>
            <strong>{{ resultDetail.listenerCount || 0 }}</strong>
          </div>
          <div class="detail-summary-card">
            <span>访问入口</span>
            <strong>{{ resultDetail.accessEntryCount }}</strong>
          </div>
        </div>

        <div class="detail-section">
          <div class="section-title">Bootstrap Servers</div>
          <div class="surface-muted">{{ resultDetail.bootstrapServers || '-' }}</div>
        </div>

        <div class="detail-section">
          <div class="section-title">状态说明</div>
          <div class="surface-muted">{{ buildRowSummary(resultDetail) }}</div>
        </div>

        <div class="detail-section" v-if="detailBrokerRows.length">
          <div class="section-title">Broker 节点</div>
          <el-table :data="detailBrokerRows" empty-text="暂无 Broker 节点">
            <el-table-column prop="address" label="地址" min-width="220" />
            <el-table-column prop="brokerId" label="Broker ID" width="110" />
            <el-table-column label="监听器" min-width="260">
              <template #default="{ row }">{{ row.listenersText || '-' }}</template>
            </el-table-column>
            <el-table-column prop="version" label="版本" width="120" />
          </el-table>
        </div>

        <div class="detail-section" v-if="detailAccessRows.length">
          <div class="section-title">访问入口</div>
          <el-table :data="detailAccessRows" empty-text="暂无访问入口">
            <el-table-column prop="address" label="地址" min-width="220" />
            <el-table-column prop="version" label="版本" width="120" />
            <el-table-column prop="errorMessage" label="错误信息" min-width="320" show-overflow-tooltip />
          </el-table>
        </div>
      </div>
      <template #footer>
        <el-button @click="resultDetailVisible = false">关闭</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="importVisible" title="导入为 Kafka 集群" width="min(680px, calc(100vw - 32px))" destroy-on-close>
      <el-form label-position="top" :model="importForm">
        <el-alert
          v-if="importSource?.versionDetectError"
          class="dialog-alert"
          type="warning"
          :closable="false"
          show-icon
          :title="`Kafka 版本自动探测失败：${importSource.versionDetectError}`"
        />
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="集群名称">
              <el-input v-model="importForm.name" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="Bootstrap Servers">
              <el-input v-model="importForm.address" disabled />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="环境">
              <el-input v-model="importForm.environment" placeholder="dev/test/prod" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="租户">
              <el-input v-model="importForm.tenant" placeholder="例如 core-team" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="Kafka 版本">
              <el-input
                v-model="importForm.auth.version"
                :placeholder="importSource?.versionDetectError ? '请手动填写，例如 3.9.0' : '自动探测成功时会自动回填'"
                :disabled="!!importSource?.kafkaVersion && !importSource?.versionDetectError"
              />
            </el-form-item>
          </el-col>
          <el-col :span="12" v-if="importSource?.kafkaVersion && !importSource?.versionDetectError">
            <el-form-item label="版本来源">
              <el-input :model-value="`自动探测: ${importSource.kafkaVersion}`" disabled />
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="描述">
          <el-input v-model="importForm.description" type="textarea" :rows="3" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="importVisible = false">取消</el-button>
        <el-button type="primary" :loading="importing" @click="importResult">导入</el-button>
      </template>
    </el-dialog>

    <el-dialog
      v-model="batchImportVisible"
      title="批量导入 Kafka 集群"
      width="min(920px, calc(100vw - 32px))"
      destroy-on-close
      :close-on-click-modal="!batchImporting"
      :close-on-press-escape="!batchImporting"
      :before-close="handleBatchImportDialogBeforeClose"
    >
      <el-alert
        class="dialog-alert"
        type="info"
        :closable="false"
        show-icon
        :title="batchImporting ? `正在导入，已选择 ${selectedCount} 个集群` : `已选择 ${selectedCount} 个集群，将按顺序逐个导入`"
      />

      <el-form label-position="top" :model="batchImportForm">
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="统一环境">
              <el-input v-model="batchImportForm.environment" placeholder="dev/test/prod" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="统一租户">
              <el-input v-model="batchImportForm.tenant" placeholder="例如 core-team" />
            </el-form-item>
          </el-col>
        </el-row>
      </el-form>

      <div class="batch-list">
        <div v-for="item in batchImportItems" :key="item.key" class="batch-item">
          <div class="batch-item-head">
            <div>
              <strong>{{ item.clusterId || item.key }}</strong>
              <span>{{ item.memberCount }} 个节点</span>
            </div>
            <el-tag v-if="item.versionDetectError" type="warning" effect="plain">需要确认版本</el-tag>
            <el-tag v-else type="success" effect="plain">可直接导入</el-tag>
          </div>

          <el-row :gutter="16">
            <el-col :span="12">
              <el-form-item label="集群名称">
                <el-input v-model="item.name" />
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item label="Kafka 版本">
                <el-input
                  v-model="item.version"
                  :disabled="!!item.detectedVersion && !item.versionDetectError"
                  :placeholder="item.versionDetectError ? '请手动填写版本' : '自动探测成功时已回填'"
                />
              </el-form-item>
            </el-col>
          </el-row>

          <div class="batch-address">{{ item.address }}</div>
          <div class="batch-hint">
            {{ item.importError || item.versionDetectError || `将以 ${item.detectedVersion || item.version || '当前配置'} 导入` }}
          </div>
        </div>
      </div>

      <template #footer>
        <el-button @click="handleBatchImportDialogClose">
          {{ batchImporting ? '停止并关闭' : '取消' }}
        </el-button>
        <el-button type="primary" :loading="batchImporting" @click="batchImportClusters">
          {{ batchImporting ? '导入中...' : '批量导入' }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { computed, nextTick, onMounted, reactive, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { getKafkaClusters, importKafkaDiscoveryResult, probeKafkaBootstrapServers, scanKafkaNetwork } from '@/api/kafka.js'

const LAST_DISCOVERY_VERSION_KEY = 'kafka-console:last-discovery-version'
const DEFAULT_FALLBACK_VERSION = '3.9.0'
const KAFKA_VERSION_PATTERN = /^\d+\.\d+(\.\d+)?$/

const loading = ref(false)
const importing = ref(false)
const batchImporting = ref(false)
const domainImporting = ref(false)
const showAdvancedAuth = ref(false)
const portsInput = ref('9092,9093,29092')
const scanResults = ref([])
const manualProbeResults = ref([])
const importVisible = ref(false)
const batchImportVisible = ref(false)
const resultDetailVisible = ref(false)
const importSource = ref(null)
const resultDetail = ref(null)
const resultTableRef = ref(null)
const selectedClusterKeys = ref([])
const batchImportItems = ref([])
const importedClusterMap = ref({})
const batchImportAbortController = ref(null)
const closeBatchImportAfterAbort = ref(false)
const syncingTableSelection = ref(false)
const discoveryPage = reactive({
  page: 1,
  pageSize: 48,
})

const scanForm = reactive({
  cidr: '',
  timeoutMs: 2500,
  concurrency: 64,
  auth: {
    version: '',
    authType: 'none',
    username: '',
    password: '',
    tlsEnabled: false,
    insecureSkipVerify: false,
    caCert: '',
    clientCert: '',
    clientKey: '',
  },
})

const filterForm = reactive({
  keyword: '',
  scope: 'all',
})

const importForm = reactive({
  name: '',
  address: '',
  environment: '',
  tenant: '',
  description: '',
  auth: {
    version: '',
    authType: 'none',
    username: '',
    password: '',
    tlsEnabled: false,
    insecureSkipVerify: false,
    caCert: '',
    clientCert: '',
    clientKey: '',
  },
})

const batchImportForm = reactive({
  environment: '',
  tenant: '',
})

const domainImportForm = reactive({
  address: '',
  timeoutMs: 2500,
  auth: {
    version: '',
    authType: 'none',
    username: '',
    password: '',
    tlsEnabled: false,
    insecureSkipVerify: false,
    caCert: '',
    clientCert: '',
    clientKey: '',
  },
})

const authMode = computed({
  get: () => {
    if (scanForm.auth.tlsEnabled && scanForm.auth.authType === 'none') {
      return 'tls'
    }
    return scanForm.auth.authType || 'none'
  },
  set: (value) => {
    const nextValue = value || 'none'
    scanForm.auth.authType = nextValue === 'tls' ? 'none' : nextValue
    scanForm.auth.tlsEnabled = nextValue === 'tls'

    if (nextValue !== 'none') {
      showAdvancedAuth.value = true
    }

    if (nextValue === 'none') {
      scanForm.auth.username = ''
      scanForm.auth.password = ''
      scanForm.auth.tlsEnabled = false
      scanForm.auth.insecureSkipVerify = false
      scanForm.auth.caCert = ''
      scanForm.auth.clientCert = ''
      scanForm.auth.clientKey = ''
    }
  },
})

const dedupe = (items) =>
  Array.from(
    new Set(
      (items || []).filter((item) => item !== undefined && item !== null && item !== ''),
    ),
  )
const createEmptyAuthTemplate = () => ({
  version: '',
  authType: 'none',
  username: '',
  password: '',
  tlsEnabled: false,
  insecureSkipVerify: false,
  caCert: '',
  clientCert: '',
  clientKey: '',
})
const cloneAuthTemplate = (auth = {}) => ({
  ...createEmptyAuthTemplate(),
  ...auth,
})
const splitBootstrapServers = (value) =>
  String(value || '')
    .split(',')
    .map((item) => item.trim())
    .filter(Boolean)
const normalizeEndpointKey = (value) => String(value || '').trim().toLowerCase()
const normalizeBootstrapServers = (value) =>
  splitBootstrapServers(value)
    .map((item) => normalizeEndpointKey(item))
    .sort()
    .join(',')
const buildEndpointKeys = (items) =>
  dedupe(
    (items || [])
      .flatMap((item) => splitBootstrapServers(item))
      .map((item) => normalizeEndpointKey(item))
      .filter(Boolean),
  )
const scoreAuthTemplate = (auth = {}) =>
  (auth.authType && auth.authType !== 'none' ? 4 : 0) +
  (auth.tlsEnabled ? 3 : 0) +
  (auth.username ? 2 : 0) +
  (auth.password ? 2 : 0) +
  (auth.caCert ? 1 : 0) +
  (auth.clientCert ? 1 : 0) +
  (auth.clientKey ? 1 : 0) +
  (auth.version ? 1 : 0)
const attachDiscoveryAuth = (items, auth) =>
  (items || []).map((item) => ({
    ...item,
    authTemplate: cloneAuthTemplate(auth),
  }))
const mergeDiscoveryResultList = (...groups) => {
  const merged = new Map()
  groups.flat().forEach((item) => {
    const key = normalizeEndpointKey(item?.address)
    if (!key) return
    merged.set(key, item)
  })

  return Array.from(merged.values()).sort((a, b) => {
    if (a.looksLikeKafka !== b.looksLikeKafka) {
      return a.looksLikeKafka ? -1 : 1
    }
    return String(a.address || '').localeCompare(String(b.address || ''), 'zh-CN')
  })
}
const results = computed(() => mergeDiscoveryResultList(scanResults.value, manualProbeResults.value))
const resolveClusterAuthTemplate = (row) => {
  const candidates = (row.members || [])
    .map((member) => member.authTemplate)
    .filter(Boolean)
    .map((auth) => cloneAuthTemplate(auth))

  if (!candidates.length) {
    return cloneAuthTemplate(scanForm.auth)
  }

  return candidates.sort((a, b) => scoreAuthTemplate(b) - scoreAuthTemplate(a))[0]
}

const buildClusterName = (row) => {
  const firstMember = row.members?.[0]
  const clusterSuffix = row.clusterId
    ? row.clusterId.slice(0, 8).replace(/[^a-zA-Z0-9_-]/g, '')
    : `${firstMember?.ip?.replaceAll('.', '-') || 'cluster'}`
  return `kafka-${clusterSuffix || 'cluster'}`
}

const clusterSummaries = computed(() => {
  const groups = new Map()

  for (const row of results.value) {
    const key = row.looksLikeKafka && row.clusterId ? `cluster:${row.clusterId}` : `node:${row.address}`
    if (!groups.has(key)) {
      groups.set(key, {
        key,
        clusterId: row.clusterId || '',
        looksLikeKafka: row.looksLikeKafka,
        members: [],
      })
    }
    groups.get(key).members.push(row)
  }

  return Array.from(groups.values())
    .map((group) => {
      const members = group.members
      const brokerMembers = members.filter((member) => member.advertisedBroker)
      const accessEntries = members.filter((member) => !member.advertisedBroker)
      const bootstrapServers = dedupe(
        members.flatMap((member) => (member.listeners && member.listeners.length ? member.listeners : [member.address])),
      ).join(',')
      const versionCandidates = dedupe(members.map((member) => member.kafkaVersion).filter(Boolean))
      const versionErrors = members.filter((member) => member.versionDetectError)
      const errorMessages = dedupe(members.map((member) => member.errorMessage).filter(Boolean))
      const controllerIDs = dedupe(
        members
          .map((member) => member.controllerId)
          .filter((id) => id !== undefined && id !== null && id >= 0),
      )

      let kafkaVersion = ''
      let versionDetectError = ''
      if (versionCandidates.length === 1) {
        kafkaVersion = versionCandidates[0]
      } else if (versionCandidates.length > 1) {
        versionDetectError = `同一集群下检测到多个版本候选：${versionCandidates.join(' / ')}，请确认后导入`
      } else if (versionErrors.length > 0) {
        versionDetectError = `共有 ${versionErrors.length} 个节点版本待确认，请在导入时手动确认`
      }

      return {
        key: group.key,
        clusterId: group.clusterId,
        looksLikeKafka: members.some((member) => member.looksLikeKafka),
        memberCount: members.length,
        members,
        bootstrapServers,
        listenerCount: dedupe(members.flatMap((member) => member.listeners || [])).length,
        brokerMembers,
        accessEntries,
        brokerCount: brokerMembers.length,
        accessEntryCount: accessEntries.length,
        kafkaVersion,
        versionDetectError,
        errorMessage: errorMessages[0] || '',
        controllerId: controllerIDs.length === 1 ? controllerIDs[0] : null,
      }
    })
    .sort((a, b) => {
      if (a.looksLikeKafka !== b.looksLikeKafka) return a.looksLikeKafka ? -1 : 1
      return String(a.clusterId || a.key).localeCompare(String(b.clusterId || b.key), 'zh-CN')
    })
})

const summaryCards = computed(() => {
  const total = results.value.length
  const kafkaCandidates = results.value.filter((item) => item.looksLikeKafka).length
  const clusterCount = clusterSummaries.value.filter((item) => item.looksLikeKafka).length
  const versionPending = clusterSummaries.value.filter((item) => item.looksLikeKafka && item.versionDetectError).length

  return [
    { label: '发现入口', value: total, desc: '当前发现结果中的入口总数' },
    { label: 'Kafka 候选', value: kafkaCandidates, desc: '能识别为 Kafka 的节点' },
    { label: '发现集群', value: clusterCount, desc: '按 Cluster ID 聚合后的集群数' },
    { label: '待确认版本', value: versionPending, desc: '导入前需要手动确认版本的集群' },
  ]
})

const discoveryRiskSummary = computed(() => ({
  versionPending: clusterSummaries.value.filter((item) => item.looksLikeKafka && item.versionDetectError).length,
  accessEntryClusters: clusterSummaries.value.filter((item) => item.looksLikeKafka && item.accessEntryCount > 0).length,
  nonKafkaClusters: clusterSummaries.value.filter((item) => !item.looksLikeKafka).length,
}))

const duplicateClusterHints = computed(() => {
  const duplicates = []
  const seen = new Map()

  clusterSummaries.value.forEach((row) => {
    const normalized = normalizeBootstrapServers(row.bootstrapServers)
    if (!normalized) return

    if (isImportedCluster(row)) {
      const imported = importedClusterMeta(row)
      duplicates.push({
        key: `imported:${row.key}`,
        title: row.clusterId || row.bootstrapServers,
        description: `该入口已导入为集群${imported?.name ? `「${imported.name}」` : ''}，建议避免重复导入。`,
        type: 'imported',
      })
    }

    if (seen.has(normalized)) {
      duplicates.push({
        key: `duplicate:${row.key}`,
        title: row.clusterId || row.bootstrapServers,
        description: `与 ${seen.get(normalized)} 使用相同的 bootstrap servers，可能是重复识别到同一组入口。`,
        type: 'duplicate',
      })
      return
    }

    seen.set(normalized, row.clusterId || row.bootstrapServers)
  })

  if (duplicates.length <= 6) return duplicates

  return [
    ...duplicates.slice(0, 6),
    {
      key: 'duplicate:more',
      title: '更多重复入口',
      description: `还有 ${duplicates.length - 6} 条重复或已导入提示未展开，请继续使用搜索或筛选排查。`,
      type: 'more',
    },
  ]
})

const importPrecheckItems = computed(() => {
  const kafkaClusters = clusterSummaries.value.filter((row) => row.looksLikeKafka)
  const versionPendingCount = kafkaClusters.filter((row) => row.versionDetectError).length
  const duplicatedCount = kafkaClusters.filter((row) => isImportedCluster(row)).length
  const authConfigured = authMode.value !== 'none' || scanForm.auth.tlsEnabled
  const tlsReady = !scanForm.auth.tlsEnabled || !!String(scanForm.auth.caCert || '').trim() || scanForm.auth.insecureSkipVerify

  return [
    {
      key: 'version',
      label: 'Kafka 版本',
      passed: versionPendingCount === 0,
      description:
        versionPendingCount === 0
          ? '当前可见 Kafka 集群都已识别到版本。'
          : `当前还有 ${versionPendingCount} 个集群版本待确认。`,
    },
    {
      key: 'auth',
      label: '认证模板',
      passed: authConfigured,
      description: authConfigured ? '当前扫描 / 导入已配置认证或 TLS 模板。' : '当前使用无认证模板，请确认目标集群是否真的允许匿名接入。',
    },
    {
      key: 'tls',
      label: 'TLS 准备',
      passed: tlsReady,
      description: tlsReady ? 'TLS 参数看起来可用。' : '已启用 TLS，但还没有提供 CA 证书，也未开启跳过校验。请先补齐。 ',
    },
    {
      key: 'duplicate',
      label: '重复入口',
      passed: duplicatedCount === 0,
      description:
        duplicatedCount === 0
          ? '当前可见结果中未发现已导入的重复入口。'
          : `当前有 ${duplicatedCount} 个结果与已导入集群重复。`,
    },
  ]
})

const filteredClusters = computed(() => {
  const keyword = filterForm.keyword.trim().toLowerCase()

  return clusterSummaries.value.filter((row) => {
    if (filterForm.scope === 'kafka' && !row.looksLikeKafka) return false
    if (filterForm.scope === 'detected' && !(row.looksLikeKafka && row.kafkaVersion)) return false
    if (filterForm.scope === 'version-failed' && !(row.looksLikeKafka && row.versionDetectError)) return false
    if (!keyword) return true

    const haystack = [
      row.clusterId,
      row.bootstrapServers,
      row.kafkaVersion,
      row.errorMessage,
      row.versionDetectError,
      ...row.members.flatMap((member) => [
        member.ip,
        member.address,
        member.errorMessage,
        member.versionDetectError,
        ...(member.listeners || []),
      ]),
    ]
      .filter(Boolean)
      .join(' ')
      .toLowerCase()

    return haystack.includes(keyword)
  })
})

const pagedClusters = computed(() => {
  const start = (discoveryPage.page - 1) * discoveryPage.pageSize
  return filteredClusters.value.slice(start, start + discoveryPage.pageSize)
})

const selectedCount = computed(() => selectedClusterKeys.value.length)
const selectedClusters = computed(() =>
  clusterSummaries.value.filter((row) => selectedClusterKeys.value.includes(row.key)),
)

const importedEndpointMap = computed(() => {
  const endpointMap = {}
  Object.values(importedClusterMap.value).forEach((item) => {
    ;(item.endpointKeys || []).forEach((key) => {
      if (!endpointMap[key]) {
        endpointMap[key] = item
      }
    })
  })
  return endpointMap
})

const detailBrokerRows = computed(() =>
  (resultDetail.value?.brokerMembers || []).map((member) => ({
    address: member.address,
    brokerId: member.brokerId,
    listenersText: (member.listeners || []).join(', '),
    version: member.kafkaVersion || '-',
  })),
)

const detailAccessRows = computed(() =>
  (resultDetail.value?.accessEntries || []).map((entry) => ({
    address: entry.address,
    version: entry.kafkaVersion || '-',
    errorMessage: entry.errorMessage || '-',
  })),
)

const collectClusterEndpointKeys = (row) =>
  buildEndpointKeys([
    row.bootstrapServers,
    ...(row.members || []).flatMap((member) => [member.address, ...(member.listeners || [])]),
  ])

const findImportedClusterMetaByEndpointKeys = (endpointKeys) => {
  if (!endpointKeys.length) return null
  for (const key of endpointKeys) {
    const matched = importedEndpointMap.value[key]
    if (matched) return matched
  }
  return null
}

const findImportedClusterMeta = (row) => {
  const directKey = normalizeBootstrapServers(row.bootstrapServers)
  if (directKey && importedClusterMap.value[directKey]) {
    return importedClusterMap.value[directKey]
  }
  return findImportedClusterMetaByEndpointKeys(collectClusterEndpointKeys(row))
}

const isImportedCluster = (row) => !!findImportedClusterMeta(row)

const importedClusterMeta = (row) => findImportedClusterMeta(row)

const isSelectableRow = (row) => row.looksLikeKafka && !isImportedCluster(row)

const getRememberedKafkaVersion = () => {
  const remembered = localStorage.getItem(LAST_DISCOVERY_VERSION_KEY)?.trim()
  return KAFKA_VERSION_PATTERN.test(remembered || '') ? remembered : DEFAULT_FALLBACK_VERSION
}

const rememberKafkaVersion = (version) => {
  const normalized = String(version || '').trim()
  if (!KAFKA_VERSION_PATTERN.test(normalized)) return
  localStorage.setItem(LAST_DISCOVERY_VERSION_KEY, normalized)
}

const refreshImportedClusters = async () => {
  try {
    const nextMap = {}
    const nextMapById = new Map()
    const pageSize = 200
    let page = 1
    let total = 0
    const seenPageSignatures = new Set()

    while (true) {
      const res = await getKafkaClusters({ page, pageSize })
      const payload = res?.data?.data || {}
      const list = payload.list || []
      total = Number(payload.total || total || 0)
      const pageSignature = [
        page,
        list.length,
        list[0]?.id || '',
        list[list.length - 1]?.id || '',
      ].join(':')
      if (seenPageSignatures.has(pageSignature)) break
      seenPageSignatures.add(pageSignature)

      list.forEach((item) => {
        const key = normalizeBootstrapServers(item.bootstrapServers)
        if (!key) return
        const previousMeta = nextMapById.get(item.id) || Object.values(importedClusterMap.value).find((meta) => meta.id === item.id)
        const nextMeta = {
          id: item.id,
          name: item.name,
          status: item.status,
          endpointKeys: dedupe([
            ...buildEndpointKeys([item.bootstrapServers]),
            ...(previousMeta?.endpointKeys || []),
          ]),
        }
        nextMap[key] = nextMeta
        nextMapById.set(item.id, nextMeta)
      })

      if (list.length === 0) break
      if (total > 0 && page * pageSize >= total) break
      if (total === 0 && list.length < pageSize) break

      page += 1
    }

    importedClusterMap.value = nextMap
  } catch {
    // 页面辅助状态，失败时不阻断发现功能
  }
}

const markImportedCluster = (row, importedInfo) => {
  const key = normalizeBootstrapServers(row.bootstrapServers || row.address) || `imported:${importedInfo?.id || row.key}`
  if (!key) return
  importedClusterMap.value = {
    ...importedClusterMap.value,
    [key]: {
      id: importedInfo?.id,
      name: importedInfo?.name || row.name,
      status: importedInfo?.status || 'unknown',
      endpointKeys: collectClusterEndpointKeys(row),
    },
  }
}

const parsePorts = () =>
  portsInput.value
    .split(',')
    .map((item) => Number(item.trim()))
    .filter((item) => Number.isInteger(item) && item > 0 && item <= 65535)

const buildClusterHint = (row) => {
  if (row.looksLikeKafka && row.brokerCount === 0 && row.listenerCount === 0) {
    return row.accessEntryCount > 0
      ? `未识别到 Broker 节点和监听器，仅识别到 ${row.accessEntryCount} 个访问入口，请先核对元数据。`
      : '已识别为 Kafka 候选，但尚未拿到 Broker 和监听器信息，请先核对元数据。'
  }
  if (row.looksLikeKafka && row.brokerCount === 0 && row.listenerCount > 0) {
    return row.accessEntryCount > 0
      ? `未识别到 Broker 节点，仅识别到 ${row.listenerCount} 个监听器和 ${row.accessEntryCount} 个访问入口，请先核对元数据。`
      : `未识别到 Broker 节点，仅识别到 ${row.listenerCount} 个监听器，请先核对元数据。`
  }
  if (row.looksLikeKafka && row.kafkaVersion) {
    if (row.accessEntryCount > 0) {
      return `识别到 ${row.brokerCount} 个 Broker 节点，另有 ${row.accessEntryCount} 个访问入口；导入时只使用 broker listeners。`
    }
    return `已识别 ${row.brokerCount} 个 Broker 节点，将按整组 bootstrap servers 导入。`
  }
  if (row.looksLikeKafka) {
    if (row.accessEntryCount > 0) {
      return `识别到 ${row.brokerCount} 个 Broker 节点，另有 ${row.accessEntryCount} 个访问入口，但版本仍需人工确认。`
    }
    return `已识别 ${row.brokerCount} 个 Broker 节点，但版本仍需人工确认。`
  }
  return '当前分组未识别为 Kafka 集群，可忽略。'
}

const buildRowSummary = (row) => {
  const summaryParts = [
    row.versionDetectError,
    row.errorMessage,
    buildClusterHint(row),
  ].filter(Boolean)
  return dedupe(summaryParts).join('；')
}

const isValidIpv4Cidr = (value) => {
  const normalized = String(value || '').trim()
  const match = normalized.match(/^(\d{1,3})(\.\d{1,3}){3}\/(\d|[12]\d|3[0-2])$/)
  if (!match) return false

  const [ipPart] = normalized.split('/')
  return ipPart.split('.').every((segment) => {
    const numericSegment = Number(segment)
    return numericSegment >= 0 && numericSegment <= 255
  })
}

const clearSelectedClusters = () => {
  selectedClusterKeys.value = []
}

const handleTableSelectionChange = (selection) => {
  if (syncingTableSelection.value) return
  const currentPageSelectableKeys = pagedClusters.value
    .filter((row) => isSelectableRow(row))
    .map((row) => row.key)
  const selectedOnCurrentPage = selection.map((row) => row.key)
  selectedClusterKeys.value = dedupe([
    ...selectedClusterKeys.value.filter((key) => !currentPageSelectableKeys.includes(key)),
    ...selectedOnCurrentPage,
  ])
}

const syncTableSelection = async () => {
  if (!resultTableRef.value) return
  syncingTableSelection.value = true
  await nextTick()
  resultTableRef.value.clearSelection()
  pagedClusters.value
    .filter((row) => isSelectableRow(row) && selectedClusterKeys.value.includes(row.key))
    .forEach((row) => {
      resultTableRef.value.toggleRowSelection(row, true)
    })
  await nextTick()
  syncingTableSelection.value = false
}

const openResultDetail = (rowKey) => {
  resultDetail.value = clusterSummaries.value.find((row) => row.key === rowKey) || null
  resultDetailVisible.value = true
}

const handleDiscoveryPageChange = (page) => {
  discoveryPage.page = Number(page || 1)
}

const handleDiscoveryPageSizeChange = (pageSize) => {
  discoveryPage.pageSize = Number(pageSize || 48)
  discoveryPage.page = 1
}

const runScan = async () => {
  const ports = parsePorts()
  if (!scanForm.cidr || ports.length === 0) {
    ElMessage.warning('请填写 CIDR 和至少一个有效端口')
    return
  }
  if (!isValidIpv4Cidr(scanForm.cidr)) {
    ElMessage.warning('请输入合法的 IPv4 CIDR，例如 192.168.1.0/24')
    return
  }

  loading.value = true
  try {
    const res = await scanKafkaNetwork({
      cidr: scanForm.cidr.trim(),
      ports,
      timeoutMs: Number(scanForm.timeoutMs),
      concurrency: Number(scanForm.concurrency),
      auth: { ...scanForm.auth },
    })
    scanResults.value = attachDiscoveryAuth(res?.data?.data || [], scanForm.auth)
    clearSelectedClusters()
    discoveryPage.page = 1
    batchImportVisible.value = false
    await refreshImportedClusters()
    ElMessage.success(`扫描完成，共返回 ${scanResults.value.length} 条节点结果，已自动按 Cluster ID 聚合`)
  } catch (error) {
    ElMessage.error(error.message || '扫描失败')
  } finally {
    loading.value = false
  }
}

const openImportDialog = (row) => {
  const authTemplate = resolveClusterAuthTemplate(row)
  importSource.value = row
  importForm.name = buildClusterName(row)
  importForm.address = row.bootstrapServers
  importForm.environment = ''
  importForm.tenant = ''
  importForm.description = `自动发现导入，ClusterID=${row.clusterId || '-'}，节点数=${row.memberCount || 1}`
  importForm.auth = {
    ...authTemplate,
    version: row.kafkaVersion || authTemplate.version || getRememberedKafkaVersion(),
  }
  importVisible.value = true
}

const resetDomainImportForm = () => {
  domainImportForm.address = ''
  domainImportForm.timeoutMs = 2500
  Object.assign(domainImportForm.auth, createEmptyAuthTemplate(), {
    version: getRememberedKafkaVersion(),
  })
}

const probeByDomain = async () => {
  if (!String(domainImportForm.address || '').trim()) {
    ElMessage.warning('请填写域名 / Bootstrap Servers')
    return
  }
  domainImporting.value = true
  try {
    const res = await probeKafkaBootstrapServers({
      address: domainImportForm.address.trim(),
      timeoutMs: Number(domainImportForm.timeoutMs),
      auth: {
        ...domainImportForm.auth,
      },
    })
    const probedResults = attachDiscoveryAuth(res?.data?.data || [], domainImportForm.auth)
    rememberKafkaVersion(domainImportForm.auth.version)
    manualProbeResults.value = mergeDiscoveryResultList(manualProbeResults.value, probedResults)
    clearSelectedClusters()
    discoveryPage.page = 1
    batchImportVisible.value = false
    await refreshImportedClusters()
    resetDomainImportForm()
    ElMessage.success(`已识别 ${probedResults.length} 个入口，并合并到当前发现结果`)
  } catch (error) {
    ElMessage.error(error.message || '域名 / Bootstrap Servers 识别失败')
  } finally {
    domainImporting.value = false
  }
}

const importResult = async () => {
  if (!importForm.name || !importForm.address) {
    ElMessage.warning('请填写集群名称并确认 Bootstrap Servers')
    return
  }
  if (importSource.value?.versionDetectError && !String(importForm.auth.version || '').trim()) {
    ElMessage.warning('自动探测失败时，请在弹窗中手动填写 Kafka 版本后再导入')
    return
  }

  importing.value = true
  try {
    const res = await importKafkaDiscoveryResult({
      name: importForm.name.trim(),
      address: importForm.address,
      environment: importForm.environment.trim(),
      tenant: importForm.tenant.trim(),
      description: importForm.description,
      auth: importForm.auth,
    })
    rememberKafkaVersion(importForm.auth.version)
    markImportedCluster(importSource.value, {
      id: res?.data?.data?.id,
      name: importForm.name.trim(),
      status: res?.data?.data?.status,
    })
    ElMessage.success('集群导入成功')
    importVisible.value = false
  } catch (error) {
    ElMessage.error(error.message || '导入失败')
  } finally {
    importing.value = false
  }
}

const openBatchImportDialog = () => {
  if (!selectedClusters.value.length) {
    ElMessage.warning('请先勾选至少一个集群')
    return
  }
  batchImportItems.value = selectedClusters.value.map((row) => ({
    key: row.key,
    clusterId: row.clusterId,
    memberCount: row.memberCount,
    address: row.bootstrapServers,
    auth: resolveClusterAuthTemplate(row),
    detectedVersion: row.kafkaVersion,
    versionDetectError: row.versionDetectError,
    version: row.kafkaVersion || getRememberedKafkaVersion(),
    name: buildClusterName(row),
    description: `自动发现导入，ClusterID=${row.clusterId || '-'}，节点数=${row.memberCount || 1}`,
  }))
  batchImportForm.environment = ''
  batchImportForm.tenant = ''
  closeBatchImportAfterAbort.value = false
  batchImportVisible.value = true
}

const isRequestCanceled = (error) =>
  error?.code === 'ERR_CANCELED' ||
  String(error?.message || '').toLowerCase().includes('canceled')

const handleBatchImportDialogClose = () => {
  if (!batchImporting.value) {
    batchImportVisible.value = false
    return
  }
  closeBatchImportAfterAbort.value = true
  batchImportAbortController.value?.abort()
}

const handleBatchImportDialogBeforeClose = (done) => {
  if (!batchImporting.value) {
    done()
    return
  }
  closeBatchImportAfterAbort.value = true
  batchImportAbortController.value?.abort()
}

const batchImportClusters = async () => {
  if (!batchImportItems.value.length) {
    ElMessage.warning('没有可导入的集群')
    return
  }
  const invalidItem = batchImportItems.value.find((item) => !String(item.name || '').trim() || !String(item.address || '').trim())
  if (invalidItem) {
    ElMessage.warning('请为所有待导入集群填写名称并确认地址')
    return
  }
  const pendingVersionItem = batchImportItems.value.find(
    (item) => item.versionDetectError && !String(item.version || '').trim(),
  )
  if (pendingVersionItem) {
    ElMessage.warning('请先为所有待确认版本的集群填写 Kafka 版本')
    return
  }

  batchImporting.value = true
  closeBatchImportAfterAbort.value = false
  const controller = new AbortController()
  batchImportAbortController.value = controller

  try {
    const failedItems = []
    let successCount = 0
    let currentIndex = 0
    let canceled = false
    const clustersByKey = new Map(clusterSummaries.value.map((row) => [row.key, row]))

    for (let index = 0; index < batchImportItems.value.length; index += 1) {
      currentIndex = index
      const item = batchImportItems.value[index]
      if (controller.signal.aborted) {
        canceled = true
        break
      }

      try {
        const res = await importKafkaDiscoveryResult({
          name: item.name.trim(),
          address: item.address,
          environment: batchImportForm.environment.trim(),
          tenant: batchImportForm.tenant.trim(),
          description: item.description,
          auth: {
            ...item.auth,
            version: item.version,
          },
        }, { signal: controller.signal })
        rememberKafkaVersion(item.version)
        const row = clustersByKey.get(item.key)
        if (row) {
          markImportedCluster(row, {
            id: res?.data?.data?.id,
            name: item.name.trim(),
            status: res?.data?.data?.status,
          })
        }
        successCount += 1
      } catch (error) {
        if (isRequestCanceled(error)) {
          canceled = true
          break
        }
        failedItems.push({
          ...item,
          importError: error.message || '导入失败，请检查当前配置后重试',
        })
      }

    }

    if (canceled) {
      const pendingItems = batchImportItems.value.slice(currentIndex)
      const remainingItems = [...failedItems, ...pendingItems]
      batchImportItems.value = remainingItems
      selectedClusterKeys.value = remainingItems.map((item) => item.key)
      ElMessage.info(`批量导入已停止，成功 ${successCount} 个，剩余 ${remainingItems.length} 个待处理`)
      if (closeBatchImportAfterAbort.value) {
        batchImportVisible.value = false
      }
      return
    }

    if (failedItems.length === 0) {
      ElMessage.success(`批量导入成功，共导入 ${successCount} 个集群`)
      batchImportVisible.value = false
      clearSelectedClusters()
      return
    }

    batchImportItems.value = failedItems
    selectedClusterKeys.value = failedItems.map((item) => item.key)
    ElMessage.warning(`批量导入完成，成功 ${successCount} 个，失败 ${failedItems.length} 个，请修正后重试`)
  } finally {
    batchImportAbortController.value = null
    closeBatchImportAfterAbort.value = false
    batchImporting.value = false
  }
}

watch(
  () => [filterForm.keyword, filterForm.scope],
  () => {
    discoveryPage.page = 1
  },
)

watch(
  () => clusterSummaries.value.map((row) => row.key),
  (keys) => {
    const keySet = new Set(keys)
    selectedClusterKeys.value = selectedClusterKeys.value.filter((key) => keySet.has(key))
  },
)

watch(
  () => filteredClusters.value.length,
  (length) => {
    const totalPages = Math.max(1, Math.ceil(length / discoveryPage.pageSize))
    if (discoveryPage.page > totalPages) {
      discoveryPage.page = totalPages
    }
  },
)

watch(
  () => [pagedClusters.value, selectedClusterKeys.value],
  () => {
    syncTableSelection()
  },
  { deep: true },
)

onMounted(() => {
  refreshImportedClusters()
  resetDomainImportForm()
})
</script>

<style scoped>
.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.card-header-wrap {
  align-items: flex-start;
}

.dialog-alert {
  margin-bottom: 16px;
}

.scan-actions-col {
  display: flex;
  align-items: stretch;
}

.scan-actions-item {
  width: 100%;
  margin-bottom: 0;
}

.scan-actions {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 12px;
  min-height: 40px;
  padding-top: 32px;
}

.scan-hint {
  color: var(--shell-text-soft);
  font-size: 13px;
}

.advanced-auth {
  margin-top: 6px;
  padding: 18px 20px 4px;
  border-radius: 16px;
  background: linear-gradient(180deg, rgba(248, 251, 255, 0.96) 0%, rgba(255, 255, 255, 0.98) 100%);
  border: 1px solid rgba(148, 163, 184, 0.18);
}

.advanced-title {
  margin-bottom: 16px;
  color: var(--shell-text);
  font-size: 14px;
  font-weight: 600;
}

.summary-row {
  margin: 20px 0;
}

.duplicate-hints {
  margin-top: 16px;
  padding-top: 16px;
  border-top: 1px dashed rgba(148, 163, 184, 0.18);
}

.duplicate-hints-title {
  margin-bottom: 10px;
  color: var(--shell-text-soft);
  font-size: 12px;
  font-weight: 700;
  letter-spacing: 0.04em;
}

.summary-panel {
  display: flex;
  flex-direction: column;
  gap: 10px;
  min-height: 124px;
  padding: 18px 20px;
  border: 1px solid rgba(148, 163, 184, 0.16);
  border-radius: 18px;
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.98) 0%, rgba(248, 251, 255, 0.96) 100%);
}

.summary-label {
  color: var(--shell-text-soft);
  font-size: 13px;
}

.summary-value {
  font-size: 30px;
  line-height: 1;
  color: var(--shell-text);
}

.summary-desc {
  margin-top: auto;
  color: var(--shell-text-soft);
  font-size: 12px;
}

.result-subtitle {
  display: inline-block;
  margin-left: 10px;
  color: var(--shell-text-soft);
  font-size: 13px;
  font-weight: 400;
}

.result-filters {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
}

.filter-input {
  width: 280px;
}

.filter-select {
  width: 180px;
}

.result-table {
  margin-top: 8px;
}

.result-table-cell {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.result-table-cell strong {
  color: var(--shell-text);
  font-size: 14px;
  font-weight: 700;
}

.result-table-cell span {
  color: var(--shell-text-soft);
  font-size: 12px;
  line-height: 1.5;
  word-break: break-word;
}

.result-table-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.result-metrics-inline {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.result-metric-chip {
  display: inline-flex;
  align-items: center;
  min-height: 26px;
  padding: 0 10px;
  border-radius: 999px;
  background: rgba(241, 245, 249, 0.88);
  color: var(--shell-text-soft);
  font-size: 11px;
  font-weight: 600;
}

.result-summary {
  color: var(--shell-text-soft);
  font-size: 12px;
  line-height: 1.55;
}

.detail-dialog-content {
  display: flex;
  flex-direction: column;
  gap: 18px;
}

.result-pagination {
  display: flex;
  justify-content: flex-end;
  margin-top: 18px;
}

.batch-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
  max-height: 520px;
  padding-right: 6px;
  overflow: auto;
}

.batch-item {
  padding: 16px;
  border: 1px solid rgba(148, 163, 184, 0.16);
  border-radius: 18px;
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.98) 0%, rgba(248, 251, 255, 0.96) 100%);
}

.batch-item-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 12px;
}

.batch-item-head strong {
  display: block;
  color: var(--shell-text);
}

.batch-item-head span {
  color: var(--shell-text-soft);
  font-size: 12px;
}

.batch-address {
  margin-top: -4px;
  color: var(--shell-text);
  font-size: 13px;
  line-height: 1.7;
  word-break: break-word;
}

.batch-hint {
  margin-top: 8px;
  color: var(--shell-text-soft);
  font-size: 12px;
}

@media (max-width: 960px) {
  .card-header,
  .result-filters,
  .batch-item-head {
    align-items: stretch;
    flex-direction: column;
  }

  .filter-input,
  .filter-select {
    width: 100%;
  }

  .scan-actions-col {
    margin-top: 0;
  }

  .scan-actions {
    padding-top: 0;
  }

  .result-pagination {
    justify-content: flex-start;
  }
}
</style>
