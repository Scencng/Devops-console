<template>
  <div class="page-container">
    <el-card class="page-header-card">
      <div class="page-header">
        <div>
          <h2>自动发现</h2>
          <p>按网段扫描 Kafka 候选节点，按 Cluster ID 聚合成集群后批量导入。</p>
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
              <el-select v-model="authMode" style="width: 100%" @change="applyAuthTemplate">
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
            <div class="scan-actions">
              <el-button type="primary" :loading="loading" @click="runScan">开始扫描</el-button>
              <span class="scan-hint">建议先留空版本，让系统优先自动探测。</span>
            </div>
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
        <div class="card-header card-header-wrap">
          <div>
            <span>按域名 / Bootstrap Servers 补充发现入口</span>
            <span class="result-subtitle">适合已知域名、VIP 或 LB 地址；识别后会与网段扫描结果按 Cluster ID 自动合并显示</span>
          </div>
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
            <div class="scan-actions">
              <el-button type="primary" :loading="domainImporting" @click="probeByDomain">识别并合并</el-button>
              <span class="scan-hint">识别成功后会进入下方统一发现结果，可继续和扫描出来的 IP 一起查看、判重和导入。</span>
            </div>
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
      </el-form>
    </el-card>

    <el-row v-if="results.length" :gutter="16" class="summary-row">
      <el-col :xs="24" :sm="12" :lg="6" v-for="card in summaryCards" :key="card.label">
        <div class="summary-panel">
          <span class="summary-label">{{ card.label }}</span>
          <strong class="summary-value">{{ card.value }}</strong>
          <span class="summary-desc">{{ card.desc }}</span>
        </div>
      </el-col>
    </el-row>

    <el-card v-if="clusterSummaries.length" class="content-card">
      <template #header>
        <div class="card-header card-header-wrap">
          <div>
            <span>导入风险提示与重复集群识别</span>
            <span class="result-subtitle">在导入前先看哪些集群已存在、哪些版本待确认，以及哪些入口更值得优先复核</span>
          </div>
        </div>
      </template>

      <div class="workbench-grid">
        <div class="workspace-panel">
          <h3>导入风险提示</h3>
          <p>根据当前扫描结果和导入状态，先确认最值得注意的风险点。</p>
          <div class="compact-list">
            <div class="compact-item">
              <div>
                <strong>待确认版本</strong>
                <span>当前共有 {{ discoveryRiskSummary.versionPending }} 个集群版本待确认，导入前建议人工核对版本。</span>
              </div>
            </div>
            <div class="compact-item">
              <div>
                <strong>访问入口混入</strong>
                <span>当前共有 {{ discoveryRiskSummary.accessEntryClusters }} 个集群同时识别到访问入口和 Broker 节点，导入时建议确认最终入口。</span>
              </div>
            </div>
            <div class="compact-item">
              <div>
                <strong>不可直接导入</strong>
                <span>当前共有 {{ discoveryRiskSummary.nonKafkaClusters }} 个分组未识别为 Kafka 集群，建议跳过或重新确认地址。</span>
              </div>
            </div>
          </div>
        </div>

        <div class="workspace-panel">
          <h3>重复集群识别</h3>
          <p>识别已导入或 bootstrap servers 重复的集群，避免重复接入同一组节点。</p>
          <div class="compact-list">
            <div v-for="item in duplicateClusterHints" :key="item.key" class="compact-item">
              <div>
                <strong>{{ item.title }}</strong>
                <span>{{ item.description }}</span>
              </div>
              <el-tag :type="item.type === 'imported' ? 'warning' : 'info'">
                {{ item.type === 'imported' ? '已导入' : '重复入口' }}
              </el-tag>
            </div>
          </div>
        </div>
      </div>
    </el-card>

    <el-card v-if="clusterSummaries.length" class="content-card">
      <template #header>
        <div class="card-header card-header-wrap">
          <div>
            <span>导入预检查清单</span>
            <span class="result-subtitle">导入前逐项确认版本、认证、TLS 和重复入口，减少把有风险的配置直接落到平台里</span>
          </div>
        </div>
      </template>

      <div class="workbench-grid">
        <div class="workspace-panel">
          <h3>当前批次检查状态</h3>
          <p>默认按当前可见扫描结果和已导入状态给出预检查结论。</p>
          <div class="compact-list">
            <div v-for="item in importPrecheckItems" :key="item.key" class="compact-item">
              <div>
                <strong>{{ item.label }}</strong>
                <span>{{ item.description }}</span>
              </div>
              <el-tag :type="item.passed ? 'success' : 'warning'">
                {{ item.passed ? '已通过' : '需确认' }}
              </el-tag>
            </div>
          </div>
        </div>

        <div class="workspace-panel">
          <h3>导入建议</h3>
          <p>建议先处理未通过项，再执行单个或批量导入。</p>
          <div class="compact-list">
            <div class="compact-item">
              <div>
                <strong>版本检查</strong>
                <span>对自动探测失败的集群先手工确认 Kafka 版本，再导入。</span>
              </div>
            </div>
            <div class="compact-item">
              <div>
                <strong>认证与 TLS</strong>
                <span>如果扫描或域名导入依赖认证/TLS，建议先检查用户名密码、CA 证书和证书校验策略。</span>
              </div>
            </div>
            <div class="compact-item">
              <div>
                <strong>重复入口</strong>
                <span>对已导入或重复 bootstrap servers 的结果，建议直接跳过，避免平台里出现重复集群。</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </el-card>

    <el-card v-if="clusterSummaries.length" class="content-card">
      <template #header>
        <div class="card-header card-header-wrap">
          <div>
            <span>集群卡片视图</span>
            <span class="result-subtitle">
              {{ filteredClusters.length }} / {{ clusterSummaries.length }} 个集群可见，已勾选 {{ selectedClusters.length }} 个
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
              <el-option label="只看 Kafka 集群" value="kafka" />
              <el-option label="只看自动探测成功" value="detected" />
              <el-option label="只看版本待确认" value="version-failed" />
            </el-select>
            <el-button @click="selectVisibleClusters" :disabled="!filteredClusters.length">勾选当前结果</el-button>
            <el-button @click="clearSelectedClusters" :disabled="!selectedClusters.length">清空勾选</el-button>
            <el-button type="primary" @click="openBatchImportDialog" :disabled="!selectedClusters.length">
              批量导入已选
            </el-button>
          </div>
        </div>
      </template>

      <div class="cluster-card-grid">
        <article
          v-for="row in filteredClusters"
          :key="row.key"
          class="cluster-card"
          :class="{
            'is-selected': isSelected(row.key),
            'is-warning': !!row.versionDetectError,
          }"
        >
          <div class="cluster-card-head">
            <el-checkbox
              :model-value="isSelected(row.key)"
              :disabled="!row.looksLikeKafka || isImportedCluster(row)"
              @change="(checked) => handleClusterSelection(row.key, checked)"
            >
              <span class="cluster-card-title">{{ row.clusterId || '未返回 Cluster ID' }}</span>
            </el-checkbox>
            <div class="cluster-card-tags">
              <el-tag :type="row.looksLikeKafka ? 'success' : 'info'" effect="plain">
                {{ row.looksLikeKafka ? 'Kafka 集群' : '非 Kafka' }}
              </el-tag>
              <el-tag v-if="isImportedCluster(row)" type="info" effect="plain">
                已导入{{ importedClusterMeta(row)?.name ? ` · ${importedClusterMeta(row).name}` : '' }}
              </el-tag>
              <el-tag v-if="row.kafkaVersion" type="success" effect="plain">{{ row.kafkaVersion }}</el-tag>
              <el-tag v-else-if="row.versionDetectError" type="warning" effect="plain">待确认版本</el-tag>
            </div>
          </div>

          <div class="cluster-card-metrics">
            <div class="metric-item">
              <span>Broker 节点</span>
              <strong>{{ row.brokerCount }}</strong>
            </div>
            <div class="metric-item">
              <span>Controller</span>
              <strong>{{ row.controllerId ?? '-' }}</strong>
            </div>
            <div class="metric-item">
              <span>listeners</span>
              <strong>{{ row.listenerCount || 0 }}</strong>
            </div>
            <div class="metric-item">
              <span>访问入口</span>
              <strong>{{ row.accessEntryCount }}</strong>
            </div>
          </div>

          <div class="cluster-card-block">
            <div class="block-label">Bootstrap Servers</div>
            <div class="block-value">{{ row.bootstrapServers || '-' }}</div>
          </div>

          <div class="cluster-card-block">
            <div class="block-label">状态说明</div>
            <div class="block-value muted">
              {{ row.versionDetectError || row.errorMessage || buildClusterHint(row) }}
            </div>
          </div>

          <div class="cluster-card-members">
            <span v-for="member in row.brokerMembers.slice(0, 4)" :key="member.address" class="member-chip">
              {{ member.address }}
            </span>
            <span v-if="row.brokerMembers.length > 4" class="member-chip more">+{{ row.brokerMembers.length - 4 }}</span>
          </div>

          <div v-if="row.accessEntries.length" class="cluster-card-block">
            <div class="block-label">访问入口</div>
            <div class="access-entry-list">
              <span v-for="entry in row.accessEntries" :key="entry.address" class="member-chip access">
                {{ entry.address }}
              </span>
            </div>
          </div>

          <div class="cluster-card-actions">
            <el-button link type="primary" :disabled="!row.looksLikeKafka || isImportedCluster(row)" @click="openImportDialog(row)">
              {{ isImportedCluster(row) ? '已导入' : row.versionDetectError ? '手动确认版本后导入' : '按集群导入' }}
            </el-button>
          </div>
        </article>
      </div>
    </el-card>

    <el-card class="content-card" v-loading="loading">
      <template #header>
        <div class="card-header">
          <span>集群明细表</span>
          <span class="result-subtitle">保留表格视图方便展开查看节点详情</span>
        </div>
      </template>

      <div v-if="!results.length" class="surface-muted discovery-empty-state">
        <strong>还没有发现结果</strong>
        <p>先执行网段扫描，或者使用上方“按域名 / Bootstrap Servers 补充发现入口”。识别完成后，这里会展开显示明细表、导入风险提示和批量导入入口。</p>
      </div>

      <el-table v-else :data="filteredClusters" empty-text="当前筛选条件下没有匹配集群">
        <el-table-column type="expand" width="56">
          <template #default="{ row }">
            <div class="member-panel">
              <div class="member-title">集群节点明细</div>
              <el-table :data="row.members" size="small">
                <el-table-column prop="ip" label="IP" width="150" />
                <el-table-column prop="port" label="端口" width="100" />
                <el-table-column label="角色" width="110">
                  <template #default="{ row: member }">
                    <el-tag :type="member.advertisedBroker ? 'success' : 'info'" effect="plain">
                      {{ member.advertisedBroker ? 'Broker' : '入口' }}
                    </el-tag>
                  </template>
                </el-table-column>
                <el-table-column prop="brokerId" label="Broker ID" width="110" />
                <el-table-column label="Kafka 版本" width="140">
                  <template #default="{ row: member }">
                    <el-tag v-if="member.kafkaVersion" type="success">{{ member.kafkaVersion }}</el-tag>
                    <el-tag v-else-if="member.versionDetectError" type="warning">待确认</el-tag>
                    <span v-else>-</span>
                  </template>
                </el-table-column>
                <el-table-column label="listeners" min-width="260" show-overflow-tooltip>
                  <template #default="{ row: member }">{{ (member.listeners || []).join(', ') || member.address }}</template>
                </el-table-column>
                <el-table-column label="状态说明" min-width="280" show-overflow-tooltip>
                  <template #default="{ row: member }">
                    {{ member.versionDetectError || member.errorMessage || buildMemberHint(member) }}
                  </template>
                </el-table-column>
              </el-table>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="Cluster ID" min-width="240" show-overflow-tooltip>
          <template #default="{ row }">{{ row.clusterId || '未返回 Cluster ID' }}</template>
        </el-table-column>
        <el-table-column label="节点数" width="100">
          <template #default="{ row }">{{ row.memberCount }}</template>
        </el-table-column>
        <el-table-column label="Kafka 版本" width="160">
          <template #default="{ row }">
            <el-tag v-if="row.kafkaVersion" type="success">{{ row.kafkaVersion }}</el-tag>
            <el-tag v-else-if="row.versionDetectError" type="warning">待确认</el-tag>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column label="Bootstrap Servers" min-width="320" show-overflow-tooltip>
          <template #default="{ row }">{{ row.bootstrapServers || '-' }}</template>
        </el-table-column>
        <el-table-column label="状态说明" min-width="320" show-overflow-tooltip>
          <template #default="{ row }">
            {{ row.versionDetectError || row.errorMessage || buildClusterHint(row) }}
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="importVisible" title="导入为 Kafka 集群" width="680px" destroy-on-close>
      <el-form label-position="top" :model="importForm">
        <el-alert
          v-if="importSource?.versionDetectError"
          type="warning"
          :closable="false"
          show-icon
          :title="`Kafka 版本自动探测失败：${importSource.versionDetectError}`"
          style="margin-bottom: 16px"
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

    <el-dialog v-model="batchImportVisible" title="批量导入 Kafka 集群" width="920px" destroy-on-close>
      <el-alert
        type="info"
        :closable="false"
        show-icon
        :title="`已选择 ${selectedClusters.length} 个集群，将按顺序逐个导入`"
        style="margin-bottom: 16px"
      />

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
            {{ item.versionDetectError || `将以 ${item.detectedVersion || item.version || '当前配置'} 导入` }}
          </div>
        </div>
      </div>

      <template #footer>
        <el-button @click="batchImportVisible = false">取消</el-button>
        <el-button type="primary" :loading="batchImporting" @click="batchImportClusters">
          批量导入
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { getKafkaClusters, importKafkaDiscoveryResult, probeKafkaBootstrapServers, scanKafkaNetwork } from '@/api/kafka.js'

const LAST_DISCOVERY_VERSION_KEY = 'kafka-console:last-discovery-version'
const DEFAULT_FALLBACK_VERSION = '3.9.0'

const loading = ref(false)
const importing = ref(false)
const batchImporting = ref(false)
const domainImporting = ref(false)
const authMode = ref('none')
const showAdvancedAuth = ref(false)
const portsInput = ref('9092,9093,29092')
const scanResults = ref([])
const manualProbeResults = ref([])
const results = ref([])
const importVisible = ref(false)
const batchImportVisible = ref(false)
const importSource = ref(null)
const selectedClusterKeys = ref([])
const batchImportItems = ref([])
const importedClusterMap = ref({})

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
  auth: {},
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

const dedupe = (items) => Array.from(new Set((items || []).filter(Boolean)))
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
    .sort((a, b) => a.localeCompare(b, 'zh-CN'))
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
const rebuildDiscoveryResults = () => {
  results.value = mergeDiscoveryResultList(scanResults.value, manualProbeResults.value)
}
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
        brokerCount: brokerMembers.length || dedupe(members.flatMap((member) => member.listeners || [])).length,
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

  return duplicates.slice(0, 6)
})

const importPrecheckItems = computed(() => {
  const visibleKafkaClusters = filteredClusters.value.filter((row) => row.looksLikeKafka)
  const versionPendingCount = visibleKafkaClusters.filter((row) => row.versionDetectError).length
  const duplicatedCount = visibleKafkaClusters.filter((row) => isImportedCluster(row)).length
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

const selectedClusters = computed(() =>
  clusterSummaries.value.filter((row) => selectedClusterKeys.value.includes(row.key)),
)

const collectClusterEndpointKeys = (row) =>
  buildEndpointKeys([
    row.bootstrapServers,
    ...(row.members || []).flatMap((member) => [member.address, ...(member.listeners || [])]),
  ])

const findImportedClusterMetaByEndpointKeys = (endpointKeys) => {
  if (!endpointKeys.length) return null
  return (
    Object.values(importedClusterMap.value).find((item) =>
      (item.endpointKeys || []).some((key) => endpointKeys.includes(key)),
    ) || null
  )
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

const applyAuthTemplate = () => {
  scanForm.auth.authType = authMode.value === 'tls' ? 'none' : authMode.value
  scanForm.auth.tlsEnabled = authMode.value === 'tls'
  if (authMode.value !== 'none') {
    showAdvancedAuth.value = true
  }
  if (authMode.value === 'none') {
    scanForm.auth.username = ''
    scanForm.auth.password = ''
    scanForm.auth.tlsEnabled = false
  }
}

const getRememberedKafkaVersion = () => {
  const remembered = localStorage.getItem(LAST_DISCOVERY_VERSION_KEY)?.trim()
  return remembered || DEFAULT_FALLBACK_VERSION
}

const rememberKafkaVersion = (version) => {
  const normalized = String(version || '').trim()
  if (!normalized) return
  localStorage.setItem(LAST_DISCOVERY_VERSION_KEY, normalized)
}

const refreshImportedClusters = async () => {
  try {
    const res = await getKafkaClusters({ page: 1, pageSize: 500 })
    const list = res?.data?.data?.list || []
    const nextMap = {}
    list.forEach((item) => {
      const key = normalizeBootstrapServers(item.bootstrapServers)
      if (!key) return
      nextMap[key] = {
        id: item.id,
        name: item.name,
        status: item.status,
        endpointKeys: buildEndpointKeys([item.bootstrapServers]),
      }
    })
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

const buildMemberHint = (row) => {
  if (!row.advertisedBroker && row.looksLikeKafka) {
    return '该地址返回了 Kafka 集群元数据，但不在 broker listeners 中，已识别为访问入口。'
  }
  if (row.looksLikeKafka && row.kafkaVersion) {
    return '版本已自动探测完成，可直接聚合到集群导入。'
  }
  if (row.looksLikeKafka) {
    return '已识别为 Kafka 节点，但版本仍需确认。'
  }
  return '当前节点未识别为 Kafka，可忽略。'
}

const buildClusterHint = (row) => {
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

const isSelected = (key) => selectedClusterKeys.value.includes(key)

const handleClusterSelection = (key, checked) => {
  if (checked) {
    if (!selectedClusterKeys.value.includes(key)) {
      selectedClusterKeys.value = [...selectedClusterKeys.value, key]
    }
    return
  }
  selectedClusterKeys.value = selectedClusterKeys.value.filter((item) => item !== key)
}

const selectVisibleClusters = () => {
  const visibleKeys = filteredClusters.value
    .filter((row) => row.looksLikeKafka && !isImportedCluster(row))
    .map((row) => row.key)
  selectedClusterKeys.value = dedupe([...selectedClusterKeys.value, ...visibleKeys])
}

const clearSelectedClusters = () => {
  selectedClusterKeys.value = []
}

const runScan = async () => {
  const ports = parsePorts()
  if (!scanForm.cidr || ports.length === 0) {
    ElMessage.warning('请填写 CIDR 和至少一个有效端口')
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
    rebuildDiscoveryResults()
    clearSelectedClusters()
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
  domainImportForm.auth = createEmptyAuthTemplate()
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
    rebuildDiscoveryResults()
    clearSelectedClusters()
    batchImportVisible.value = false
    await refreshImportedClusters()
    domainImportForm.address = ''
    ElMessage.success(`已识别 ${probedResults.length} 个入口，并合并到当前发现结果`)
  } catch (error) {
    ElMessage.error(error.message || '域名 / Bootstrap Servers 识别失败')
  } finally {
    domainImporting.value = false
  }
}

const importResult = async () => {
  if (!importForm.name || !importForm.address) {
    ElMessage.warning('请填写集群名称')
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
  batchImportVisible.value = true
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
  try {
    for (const item of batchImportItems.value) {
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
      })
      rememberKafkaVersion(item.version)
      const row = selectedClusters.value.find((cluster) => cluster.key === item.key)
      if (row) {
        markImportedCluster(row, {
          id: res?.data?.data?.id,
          name: item.name.trim(),
          status: res?.data?.data?.status,
        })
      }
    }
    ElMessage.success(`批量导入成功，共导入 ${batchImportItems.value.length} 个集群`)
    batchImportVisible.value = false
    clearSelectedClusters()
  } catch (error) {
    ElMessage.error(error.message || '批量导入失败')
  } finally {
    batchImporting.value = false
  }
}

onMounted(() => {
  refreshImportedClusters()
  domainImportForm.auth.version = getRememberedKafkaVersion()
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

.scan-actions-col {
  display: flex;
  align-items: flex-end;
}

.scan-actions {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 12px;
  min-height: 56px;
}

.scan-hint {
  color: #718096;
  font-size: 13px;
}

.advanced-auth {
  margin-top: 6px;
  padding: 18px 20px 4px;
  border-radius: 16px;
  background: linear-gradient(180deg, #f8fbff 0%, #ffffff 100%);
  border: 1px solid #e8eef5;
}

.advanced-title {
  margin-bottom: 16px;
  color: #1f2937;
  font-size: 14px;
  font-weight: 600;
}

.summary-row {
  margin: 20px 0;
}

.summary-panel {
  display: flex;
  flex-direction: column;
  gap: 10px;
  min-height: 124px;
  padding: 18px 20px;
  border: 1px solid #e9eef5;
  border-radius: 18px;
  background: linear-gradient(180deg, #ffffff 0%, #f8fbff 100%);
}

.summary-label {
  color: #64748b;
  font-size: 13px;
}

.summary-value {
  font-size: 30px;
  line-height: 1;
  color: #0f172a;
}

.summary-desc {
  margin-top: auto;
  color: #94a3b8;
  font-size: 12px;
}

.result-subtitle {
  display: inline-block;
  margin-left: 10px;
  color: #94a3b8;
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

.cluster-card-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(320px, 1fr));
  gap: 16px;
}

.cluster-card {
  display: flex;
  flex-direction: column;
  gap: 14px;
  min-height: 280px;
  padding: 18px;
  border: 1px solid #e7edf5;
  border-radius: 20px;
  background: linear-gradient(180deg, #ffffff 0%, #f8fbff 100%);
  transition: transform 0.2s ease, box-shadow 0.2s ease, border-color 0.2s ease;
}

.cluster-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 10px 26px rgba(15, 23, 42, 0.08);
}

.cluster-card.is-selected {
  border-color: rgba(64, 158, 255, 0.55);
  box-shadow: 0 0 0 3px rgba(64, 158, 255, 0.12);
}

.cluster-card.is-warning {
  background: linear-gradient(180deg, #ffffff 0%, #fffaf2 100%);
}

.cluster-card-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
}

.cluster-card-title {
  display: inline-block;
  max-width: 220px;
  overflow: hidden;
  color: #0f172a;
  font-weight: 700;
  text-overflow: ellipsis;
  white-space: nowrap;
  vertical-align: middle;
}

.cluster-card-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  justify-content: flex-end;
}

.cluster-card-metrics {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 10px;
}

.metric-item {
  padding: 12px 14px;
  border-radius: 14px;
  background: rgba(241, 245, 249, 0.8);
}

.metric-item span {
  display: block;
  margin-bottom: 6px;
  color: #64748b;
  font-size: 12px;
}

.metric-item strong {
  color: #0f172a;
  font-size: 20px;
}

.cluster-card-block {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.block-label {
  color: #64748b;
  font-size: 12px;
  font-weight: 600;
  letter-spacing: 0.02em;
}

.block-value {
  color: #0f172a;
  font-size: 13px;
  line-height: 1.7;
  word-break: break-word;
}

.block-value.muted {
  color: #475569;
}

.cluster-card-members {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-top: auto;
}

.member-chip {
  padding: 6px 10px;
  border-radius: 999px;
  background: #eff6ff;
  color: #2563eb;
  font-size: 12px;
}

.member-chip.more {
  background: #f1f5f9;
  color: #64748b;
}

.member-chip.access {
  background: #fff7ed;
  color: #c2410c;
}

.access-entry-list {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.cluster-card-actions {
  display: flex;
  justify-content: flex-end;
}

.member-panel {
  padding: 12px 8px 4px;
}

.member-title {
  margin-bottom: 12px;
  color: #334155;
  font-size: 13px;
  font-weight: 600;
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
  border: 1px solid #e7edf5;
  border-radius: 18px;
  background: linear-gradient(180deg, #ffffff 0%, #f8fbff 100%);
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
  color: #0f172a;
}

.batch-item-head span {
  color: #64748b;
  font-size: 12px;
}

.batch-address {
  margin-top: -4px;
  color: #0f172a;
  font-size: 13px;
  line-height: 1.7;
  word-break: break-word;
}

.batch-hint {
  margin-top: 8px;
  color: #64748b;
  font-size: 12px;
}

.discovery-empty-state strong {
  display: block;
  margin-bottom: 6px;
  color: #0f172a;
  font-size: 14px;
}

.discovery-empty-state p {
  margin: 0;
  color: #64748b;
  font-size: 13px;
  line-height: 1.7;
}

@media (max-width: 960px) {
  .card-header,
  .result-filters,
  .cluster-card-head,
  .batch-item-head {
    align-items: stretch;
    flex-direction: column;
  }

  .filter-input,
  .filter-select {
    width: 100%;
  }

  .scan-actions-col {
    margin-top: -6px;
  }
}
</style>
