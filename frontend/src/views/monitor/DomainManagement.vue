<template>
  <div class="domain-management">
    <!-- 统计卡片行 -->
    <div class="stats-row">
      <div class="stat-card">
        <div class="stat-icon-wrap" style="background:#6c5ce7;">
          <el-icon size="22" color="#fff"><Monitor /></el-icon>
        </div>
        <div class="stat-info">
          <div class="stat-label">域名总数</div>
          <div class="stat-value">{{ domainStats.total }}</div>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon-wrap" style="background:#00b894;">
          <el-icon size="22" color="#fff"><CircleCheck /></el-icon>
        </div>
        <div class="stat-info">
          <div class="stat-label">正常域名</div>
          <div class="stat-value" style="color:#00b894">{{ domainStats.normal }}</div>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon-wrap" style="background:#d63031;">
          <el-icon size="22" color="#fff"><CircleClose /></el-icon>
        </div>
        <div class="stat-info">
          <div class="stat-label">异常域名</div>
          <div class="stat-value" style="color:#d63031">{{ domainStats.abnormal }}</div>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon-wrap" style="background:#e17055;">
          <el-icon size="22" color="#fff"><Warning /></el-icon>
        </div>
        <div class="stat-info">
          <div class="stat-label">即将过期</div>
          <div class="stat-value" style="color:#e17055">{{ domainStats.expiringSoon }}</div>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon-wrap" style="background:#b2bec3;">
          <el-icon size="22" color="#fff"><Remove /></el-icon>
        </div>
        <div class="stat-info">
          <div class="stat-label">已过期</div>
          <div class="stat-value" style="color:#636e72">{{ domainStats.expired }}</div>
        </div>
      </div>
    </div>

    <!-- 主内容卡片 -->
    <div class="content-card">
      <div class="section-title">域名监控</div>

      <el-tabs v-model="activeTab" class="domain-tabs">
        <!-- ==================== Tab1: SSL证书监控 ==================== -->
        <el-tab-pane name="monitor">
          <template #label>
            <span><el-icon><Monitor /></el-icon> SSL证书监控</span>
          </template>

          <!-- 搜索 -->
          <div class="bar">
            <el-input v-model="monitorSearch.domain" placeholder="输入域名" clearable style="width:160px" />
            <el-input v-model="monitorSearch.tag" placeholder="输入标签" clearable style="width:140px" />
            <el-select v-model="monitorSearch.status" placeholder="状态" clearable style="width:110px">
              <el-option label="正常" value="normal" />
              <el-option label="异常" value="abnormal" />
            </el-select>
            <el-select v-model="monitorSearch.aliveStatus" placeholder="存活状态" clearable style="width:110px">
              <el-option label="存活" value="alive" />
              <el-option label="不存活" value="dead" />
            </el-select>
            <el-button type="primary" :icon="Search" @click="fetchDomains">搜索</el-button>
            <el-button :icon="Refresh" @click="resetMonitorSearch">重置</el-button>
          </div>
          <div class="bar">
            <el-button type="primary" :icon="Plus" @click="openDomainDialog()">添加域名</el-button>
            <el-button :icon="Upload" @click="handleBatchUpload">批量上传</el-button>
            <el-button :icon="Download" @click="handleCloudImport">云导入</el-button>
            <el-button type="warning" @click="handleAutoMonitor">自动监控</el-button>
            <el-button :type="filter30Days ? 'warning' : 'default'" @click="filter30Days = !filter30Days">
              {{ filter30Days ? '已选30天（点击取消）' : '⏱ 30天到期' }}
            </el-button>
            <el-button :type="filterAbnormal ? 'danger' : 'default'" @click="filterAbnormal = !filterAbnormal">
              {{ filterAbnormal ? '取消看异常' : '● 异常域名' }}
            </el-button>
          </div>

          <el-table :data="domainTableData" v-loading="domainLoading" border stripe>
            <el-table-column prop="id" label="ID" width="60" />
            <el-table-column prop="domain" label="域名" min-width="180" show-overflow-tooltip />
            <el-table-column label="状态" width="80">
              <template #default="{ row }">
                <el-tag :type="row.status === 'normal' ? 'success' : 'danger'" size="small">
                  {{ row.status === 'normal' ? '正常' : '异常' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="statusCode" label="状态码" width="80" />
            <el-table-column label="响应" width="80">
              <template #default="{ row }">
                <span>{{ row.responseTime ? row.responseTime + 'ms' : '-' }}</span>
              </template>
            </el-table-column>
            <el-table-column label="SSL证书" min-width="140">
              <template #default="{ row }">
                <div v-if="row.sslExpiry">
                  <div class="ssl-date">{{ row.sslExpiry }}</div>
                  <el-tag :type="getSslTagType(row.sslDaysLeft)" size="small">
                    剩余 {{ row.sslDaysLeft }} 天
                  </el-tag>
                </div>
                <span v-else>-</span>
              </template>
            </el-table-column>
            <el-table-column prop="lastCheck" label="最后检查" min-width="150" show-overflow-tooltip />
            <el-table-column label="启用" width="70">
              <template #default="{ row }">
                <el-switch :model-value="row.enabled" @change="(v) => toggleDomain(row, v)" />
              </template>
            </el-table-column>
            <el-table-column prop="remark" label="备注" min-width="100" show-overflow-tooltip />
            <el-table-column label="证书提供商" width="110">
              <template #default="{ row }">
                <el-tag v-if="row.certProvider" size="small" type="info">{{ row.certProvider }}</el-tag>
                <span v-else>-</span>
              </template>
            </el-table-column>
            <el-table-column label="操作" width="120" fixed="right">
              <template #default="{ row }">
                <el-button link type="primary" size="small" @click="openDomainDialog(row)">编辑</el-button>
                <el-button link type="warning" size="small" @click="handleCheckDomain(row)">检查</el-button>
                <el-button link type="danger" size="small" @click="handleDeleteDomain(row)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>
          <el-pagination class="pagination" v-model:current-page="domainPage.page" v-model:page-size="domainPage.pageSize"
            :total="domainPage.total" layout="total, sizes, prev, pager, next, jumper"
            :page-sizes="[10, 20, 50]" @size-change="fetchDomains" @current-change="fetchDomains" />
        </el-tab-pane>

        <!-- ==================== Tab2: SSL证书申请 ==================== -->
        <el-tab-pane name="cert">
          <template #label>
            <span><el-icon><Document /></el-icon> SSL证书申请</span>
          </template>

          <div class="bar">
            <el-input v-model="certSearch.domain" placeholder="搜索域名" clearable style="width:180px" />
            <el-select v-model="certSearch.status" placeholder="证书状态" clearable style="width:130px">
              <el-option label="申请中" :value="-1" />
              <el-option label="已签发" :value="1" />
              <el-option label="申请失败" :value="0" />
              <el-option label="已过期" :value="2" />
            </el-select>
            <el-button type="primary" :icon="Search" @click="fetchCerts">查询</el-button>
            <el-button :icon="Refresh" @click="resetCertSearch">重置</el-button>
          </div>
          <div class="bar">
            <el-button type="primary" :icon="Plus" @click="openApplyDialog">申请证书</el-button>
            <el-button :icon="Upload" @click="openUploadDialog">上传证书</el-button>
            <el-button :icon="Refresh" @click="fetchCerts">刷新</el-button>
          </div>

          <el-table :data="certTableData" v-loading="certLoading" border stripe>
            <el-table-column prop="id" label="ID" width="60" />
            <el-table-column prop="domain" label="域名" min-width="160" show-overflow-tooltip />
            <el-table-column label="云配置" min-width="100">
              <template #default="{ row }">
                <el-tag size="small" type="info">{{ row.dnsConfig || '-' }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="证书来源" width="110">
              <template #default="{ row }">
                <el-tag size="small" :type="row.certSource === 'ACME协议' ? 'success' : ''">{{ row.certSource }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="CA提供商" width="110">
              <template #default="{ row }">
                <span>{{ row.caProvider }}</span>
              </template>
            </el-table-column>
            <el-table-column label="加密算法" width="100">
              <template #default="{ row }">
                <el-tag size="small" type="warning">{{ row.keyAlgorithm }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="状态" width="90">
              <template #default="{ row }">
                <el-tag :type="getCertStatusType(row.status)" size="small">{{ getCertStatusLabel(row.status) }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="expireAt" label="过期时间" min-width="120" show-overflow-tooltip />
            <el-table-column label="操作" width="150" fixed="right">
              <template #default="{ row }">
                <el-button link type="primary" size="small" :icon="Download" @click="handleDownloadCert(row)">下载</el-button>
                <el-button link type="success" size="small" @click="handleDeployCert(row)">部署</el-button>
                <el-button link type="danger" size="small" @click="handleDeleteCert(row)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>
          <el-pagination class="pagination" v-model:current-page="certPage.page" v-model:page-size="certPage.pageSize"
            :total="certPage.total" layout="total, sizes, prev, pager, next"
            :page-sizes="[10, 20, 50]" @size-change="fetchCerts" @current-change="fetchCerts" />
        </el-tab-pane>

        <!-- ==================== Tab3: DNS云厂商配置 ==================== -->
        <el-tab-pane name="dns">
          <template #label>
            <span><el-icon><Setting /></el-icon> DNS云厂商配置</span>
          </template>

          <div class="bar">
            <el-input v-model="dnsSearch.name" placeholder="搜索配置名称" clearable style="width:180px" />
            <el-select v-model="dnsSearch.status" placeholder="配置状态" clearable style="width:130px">
              <el-option label="启用" value="active" />
              <el-option label="停用" value="inactive" />
            </el-select>
            <el-button type="primary" :icon="Search" @click="fetchDnsProviders">查询</el-button>
            <el-button :icon="Refresh" @click="resetDnsSearch">重置</el-button>
          </div>
          <div class="bar">
            <el-button type="primary" :icon="Plus" @click="openDnsDialog()">添加配置</el-button>
            <el-button :icon="Refresh" @click="fetchDnsProviders">刷新</el-button>
          </div>

          <el-table :data="dnsTableData" v-loading="dnsLoading" border stripe>
            <el-table-column prop="id" label="ID" width="60" />
            <el-table-column prop="name" label="配置名称" min-width="140" show-overflow-tooltip />
            <el-table-column label="云厂商" width="120">
              <template #default="{ row }">
                <el-tag :type="getProviderTagType(row.provider)" size="small">{{ getProviderLabel(row.provider) }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="AccessKey" min-width="180" show-overflow-tooltip>
              <template #default="{ row }">
                <span class="access-key">{{ maskAccessKey(row.accessKey) }}</span>
              </template>
            </el-table-column>
            <el-table-column label="状态" width="90">
              <template #default="{ row }">
                <el-tag :type="row.status === 'active' ? 'success' : 'info'" size="small">
                  {{ row.status === 'active' ? '启用' : '停用' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="email" label="邮箱" min-width="160" show-overflow-tooltip />
            <el-table-column prop="phone" label="电话" min-width="120" show-overflow-tooltip />
            <el-table-column label="操作" width="160" fixed="right">
              <template #default="{ row }">
                <el-button link type="primary" size="small" @click="handleTestDns(row)">测试</el-button>
                <el-button link type="primary" size="small" @click="openDnsDialog(row)">编辑</el-button>
                <el-button link type="danger" size="small" @click="handleDeleteDns(row)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>
          <el-pagination class="pagination" v-model:current-page="dnsPage.page" v-model:page-size="dnsPage.pageSize"
            :total="dnsPage.total" layout="total, sizes, prev, pager, next"
            :page-sizes="[10, 20, 50]" @size-change="fetchDnsProviders" @current-change="fetchDnsProviders" />
        </el-tab-pane>
      </el-tabs>
    </div>

    <!-- ==================== Dialog: 添加/编辑域名 ==================== -->
    <el-dialog v-model="domainDialogVisible" :title="domainDialogTitle" width="540px" destroy-on-close>
      <el-form ref="domainFormRef" :model="domainForm" :rules="domainFormRules" label-width="90px">
        <el-form-item label="域名" prop="domain">
          <el-input v-model="domainForm.domain" placeholder="example.com" />
        </el-form-item>
        <el-form-item label="标签">
          <el-select v-model="domainForm.tags" multiple allow-create filterable placeholder="输入后回车添加" style="width:100%" />
        </el-form-item>
        <el-form-item label="监控协议">
          <el-radio-group v-model="domainForm.protocol">
            <el-radio value="https">HTTPS</el-radio>
            <el-radio value="http">HTTP</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="检查间隔">
          <el-select v-model="domainForm.checkInterval" style="width:100%">
            <el-option label="1分钟" :value="60" />
            <el-option label="5分钟" :value="300" />
            <el-option label="10分钟" :value="600" />
            <el-option label="30分钟" :value="1800" />
          </el-select>
        </el-form-item>
        <el-form-item label="启用">
          <el-switch v-model="domainForm.enabled" />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="domainForm.remark" type="textarea" :rows="2" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="domainDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleDomainSubmit">确认</el-button>
      </template>
    </el-dialog>

    <!-- ==================== Dialog: 申请SSL证书 ==================== -->
    <el-dialog v-model="applyDialogVisible" title="申请SSL证书" width="560px" destroy-on-close>
      <el-form ref="applyFormRef" :model="applyForm" :rules="applyFormRules" label-width="100px">
        <el-form-item label="域名" prop="domain">
          <el-input v-model="applyForm.domain" placeholder="example.com 或 *.example.com" />
        </el-form-item>
        <el-form-item label="DNS配置" prop="dnsConfigId">
          <el-select v-model="applyForm.dnsConfigId" placeholder="选择DNS云厂商配置" style="width:100%">
            <el-option v-for="item in dnsTableData" :key="item.id" :label="item.name" :value="item.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="证书来源" prop="certSource">
          <el-radio-group v-model="applyForm.certSource">
            <el-radio value="ACME">ACME协议（免费）</el-radio>
            <el-radio value="CAS">CA商业证书</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="CA提供商" prop="caProvider">
          <el-select v-model="applyForm.caProvider" style="width:100%">
            <el-option v-if="applyForm.certSource === 'ACME'" label="Let's Encrypt（免费）" value="LetsEncrypt" />
            <el-option v-if="applyForm.certSource === 'ACME'" label="ZeroSSL（免费）" value="ZeroSSL" />
            <el-option v-if="applyForm.certSource === 'CAS'" label="DigiCert" value="DigiCert" />
            <el-option v-if="applyForm.certSource === 'CAS'" label="GlobalSign" value="GlobalSign" />
            <el-option v-if="applyForm.certSource === 'CAS'" label="Sectigo" value="Sectigo" />
          </el-select>
        </el-form-item>
        <el-form-item label="加密算法" prop="keyAlgorithm">
          <el-select v-model="applyForm.keyAlgorithm" style="width:100%">
            <el-option label="RSA 2048" value="RSA2048" />
            <el-option label="RSA 4096" value="RSA4096" />
            <el-option label="EC 256（推荐）" value="EC256" />
          </el-select>
        </el-form-item>
        <el-form-item label="联系邮箱" prop="email">
          <el-input v-model="applyForm.email" placeholder="用于证书通知" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="applyDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleApplyCert">申请</el-button>
      </template>
    </el-dialog>

    <!-- ==================== Dialog: 上传SSL证书 ==================== -->
    <el-dialog v-model="uploadDialogVisible" title="上传SSL证书" width="540px" destroy-on-close>
      <el-form ref="uploadFormRef" :model="uploadForm" :rules="uploadFormRules" label-width="100px">
        <el-form-item label="域名" prop="domain">
          <el-input v-model="uploadForm.domain" placeholder="example.com" />
        </el-form-item>
        <el-form-item label="证书内容" prop="certPem">
          <el-input v-model="uploadForm.certPem" type="textarea" :rows="5" placeholder="------BEGIN CERTIFICATE------" />
        </el-form-item>
        <el-form-item label="私钥内容" prop="keyPem">
          <el-input v-model="uploadForm.keyPem" type="textarea" :rows="5" placeholder="------BEGIN PRIVATE KEY------" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="uploadDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleUploadCert">上传</el-button>
      </template>
    </el-dialog>

    <!-- ==================== Dialog: 添加/编辑DNS配置 ==================== -->
    <el-dialog v-model="dnsDialogVisible" :title="dnsDialogTitle" width="560px" destroy-on-close>
      <el-form ref="dnsFormRef" :model="dnsForm" :rules="dnsFormRules" label-width="100px">
        <el-form-item label="配置名称" prop="name">
          <el-input v-model="dnsForm.name" placeholder="如：公司阿里云DNS" />
        </el-form-item>
        <el-form-item label="云厂商" prop="provider">
          <el-select v-model="dnsForm.provider" placeholder="选择云厂商" style="width:100%">
            <el-option v-for="p in CLOUD_PROVIDERS" :key="p.value" :label="p.label" :value="p.value" />
          </el-select>
        </el-form-item>
        <!-- 动态字段根据厂商显示 -->
        <template v-if="dnsForm.provider">
          <el-form-item :label="getProviderFieldLabel(dnsForm.provider, 'key')" prop="accessKey">
            <el-input v-model="dnsForm.accessKey" :placeholder="getProviderFieldPlaceholder(dnsForm.provider, 'key')" />
          </el-form-item>
          <el-form-item :label="getProviderFieldLabel(dnsForm.provider, 'secret')" prop="accessSecret">
            <el-input v-model="dnsForm.accessSecret" type="password" show-password
              :placeholder="getProviderFieldPlaceholder(dnsForm.provider, 'secret')" />
          </el-form-item>
          <!-- CloudFlare 额外字段 -->
          <el-form-item v-if="dnsForm.provider === 'cloudflare'" label="Zone ID">
            <el-input v-model="dnsForm.zoneId" placeholder="CloudFlare Zone ID（可选）" />
          </el-form-item>
          <!-- AWS 额外字段 -->
          <el-form-item v-if="dnsForm.provider === 'aws'" label="区域">
            <el-select v-model="dnsForm.region" placeholder="AWS Region" style="width:100%">
              <el-option label="us-east-1" value="us-east-1" />
              <el-option label="us-west-2" value="us-west-2" />
              <el-option label="ap-southeast-1" value="ap-southeast-1" />
              <el-option label="ap-northeast-1" value="ap-northeast-1" />
            </el-select>
          </el-form-item>
        </template>
        <el-form-item label="联系邮箱" prop="email">
          <el-input v-model="dnsForm.email" placeholder="联系邮箱" />
        </el-form-item>
        <el-form-item label="联系电话">
          <el-input v-model="dnsForm.phone" placeholder="联系电话" />
        </el-form-item>
        <el-form-item label="状态">
          <el-radio-group v-model="dnsForm.status">
            <el-radio value="active">启用</el-radio>
            <el-radio value="inactive">停用</el-radio>
          </el-radio-group>
        </el-form-item>
        <!-- 厂商说明提示 -->
        <el-alert v-if="dnsForm.provider" :title="getProviderHint(dnsForm.provider)" type="info" :closable="false" show-icon style="margin-top:4px" />
      </el-form>
      <template #footer>
        <el-button @click="dnsDialogVisible = false">取消</el-button>
        <el-button type="success" :loading="testingDns" @click="handleTestDnsForm">测试连接</el-button>
        <el-button type="primary" :loading="submitting" @click="handleDnsSubmit">确认</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import {onMounted, reactive, ref} from 'vue'
import {
  CircleCheck,
  CircleClose,
  Document,
  Download,
  Monitor,
  Plus,
  Refresh,
  Remove,
  Search,
  Setting,
  Upload,
  Warning
} from '@element-plus/icons-vue'
import {ElMessage, ElMessageBox} from 'element-plus'
import {
  applySslCert,
  createDnsProvider,
  createDomain,
  deleteDnsProvider,
  deleteDomain,
  deleteSslCert,
  downloadSslCert,
  getDnsProviderList,
  getDomainList,
  getDomainStats,
  getSslCertList,
  testDnsProvider,
  toggleDomainEnabled,
  updateDnsProvider,
  updateDomain,
  uploadSslCert
} from '@/api/domain.js'

// ===================== 云厂商配置 =====================
const CLOUD_PROVIDERS = [
  { value: 'aliyun', label: '阿里云 DNS', keyLabel: 'AccessKey ID', secretLabel: 'AccessKey Secret' },
  { value: 'tencent', label: '腾讯云 DNS（DNSPod）', keyLabel: 'SecretId', secretLabel: 'SecretKey' },
  { value: 'huawei', label: '华为云 DNS', keyLabel: 'Access Key ID', secretLabel: 'Secret Access Key' },
  { value: 'aws', label: 'AWS Route53', keyLabel: 'Access Key ID', secretLabel: 'Secret Access Key' },
  { value: 'cloudflare', label: 'CloudFlare', keyLabel: '邮箱', secretLabel: 'Global API Key / Token' },
  { value: 'godaddy', label: 'GoDaddy', keyLabel: 'API Key', secretLabel: 'API Secret' },
  { value: 'dnspod', label: 'DNSPod（独立版）', keyLabel: 'API Token ID', secretLabel: 'API Token' },
  { value: 'baidu', label: '百度云 DNS', keyLabel: 'Access Key', secretLabel: 'Secret Key' },
  { value: 'volcengine', label: '火山引擎 DNS', keyLabel: 'Access Key ID', secretLabel: 'Secret Access Key' },
]

const PROVIDER_HINTS = {
  aliyun: '需在阿里云 RAM 控制台创建具有 DNS 权限的子用户，获取 AccessKey ID 和 Secret。',
  tencent: '在腾讯云控制台 → 访问管理 → 新建密钥，获取 SecretId 和 SecretKey，并确保具有 QcloudDNSPodFullAccess 权限。',
  huawei: '在华为云 IAM → 安全凭证 → 访问密钥，创建并下载 AK/SK 文件，需要 DNS 服务权限。',
  aws: '在 AWS IAM 控制台创建用户，附加 AmazonRoute53FullAccess 策略，生成访问密钥。',
  cloudflare: '在 CloudFlare 控制台 → 个人资料 → API 令牌，可使用 Global API Key 或创建具有 DNS 权限的 API Token。',
  godaddy: '登录 GoDaddy 开发者中心 (developer.godaddy.com) 创建生产环境 API Key 和 Secret。',
  dnspod: '在 DNSPod 控制台 → 密钥管理 → 创建密钥，获取 API Token ID 和 Token。',
  baidu: '在百度智能云 → 访问控制 → 用户 → 创建访问密钥，需要云 DNS 相关权限。',
  volcengine: '在火山引擎 → IAM → 访问密钥管理，创建并获取 AK/SK，需要具有 DNS 服务权限。',
}

// ===================== 顶部统计 =====================
const domainStats = reactive({ total: 0, normal: 0, abnormal: 0, expiringSoon: 0, expired: 0 })

// ===================== Tab 状态 =====================
const activeTab = ref('monitor')

// ===================== Tab1: 域名监控 =====================
const domainLoading = ref(false)
const domainTableData = ref([])
const filter30Days = ref(false)
const filterAbnormal = ref(false)
const monitorSearch = reactive({ domain: '', tag: '', status: '', aliveStatus: '' })
const domainPage = reactive({ page: 1, pageSize: 10, total: 0 })

// ===================== Tab2: SSL证书 =====================
const certLoading = ref(false)
const certTableData = ref([])
const certSearch = reactive({ domain: '', status: '' })
const certPage = reactive({ page: 1, pageSize: 10, total: 0 })

// ===================== Tab3: DNS配置 =====================
const dnsLoading = ref(false)
const dnsTableData = ref([])
const dnsSearch = reactive({ name: '', status: '' })
const dnsPage = reactive({ page: 1, pageSize: 10, total: 0 })

const submitting = ref(false)
const testingDns = ref(false)

// Dialog refs
const domainDialogVisible = ref(false)
const domainDialogTitle = ref('添加域名')
const domainFormRef = ref(null)
const domainForm = reactive({ id: null, domain: '', tags: [], protocol: 'https', checkInterval: 300, enabled: true, remark: '' })
const domainFormRules = { domain: [{ required: true, message: '请输入域名', trigger: 'blur' }] }

const applyDialogVisible = ref(false)
const applyFormRef = ref(null)
const applyForm = reactive({ domain: '', dnsConfigId: null, certSource: 'ACME', caProvider: 'LetsEncrypt', keyAlgorithm: 'EC256', email: '' })
const applyFormRules = {
  domain: [{ required: true, message: '请输入域名' }],
  dnsConfigId: [{ required: true, message: '请选择DNS配置' }],
  caProvider: [{ required: true, message: '请选择CA提供商' }],
  email: [{ required: true, message: '请输入联系邮箱' }],
}

const uploadDialogVisible = ref(false)
const uploadFormRef = ref(null)
const uploadForm = reactive({ domain: '', certPem: '', keyPem: '' })
const uploadFormRules = {
  domain: [{ required: true, message: '请输入域名' }],
  certPem: [{ required: true, message: '请填写证书内容' }],
  keyPem: [{ required: true, message: '请填写私钥内容' }],
}

const dnsDialogVisible = ref(false)
const dnsDialogTitle = ref('添加配置')
const dnsFormRef = ref(null)
const dnsForm = reactive({ id: null, name: '', provider: '', accessKey: '', accessSecret: '', zoneId: '', region: '', email: '', phone: '', status: 'active' })
const dnsFormRules = {
  name: [{ required: true, message: '请输入配置名称' }],
  provider: [{ required: true, message: '请选择云厂商' }],
  accessKey: [{ required: true, message: '请填写 AccessKey' }],
  accessSecret: [{ required: true, message: '请填写 AccessSecret' }],
  email: [{ required: true, message: '请填写联系邮箱' }],
}

// ===================== 生命周期 =====================
onMounted(() => {
  fetchStats()
  fetchDomains()
  fetchCerts()
  fetchDnsProviders()
})

// ===================== 数据获取 =====================
async function fetchStats() {
  try {
    const res = await getDomainStats()
    Object.assign(domainStats, res?.data?.data || {})
  } catch {}
}

async function fetchDomains() {
  domainLoading.value = true
  try {
    const params = { page: domainPage.page, pageSize: domainPage.pageSize, ...monitorSearch }
    if (filter30Days.value) params.expireWithin = 30
    if (filterAbnormal.value) params.status = 'abnormal'
    const res = await getDomainList(params)
    const d = res?.data?.data || {}
    domainTableData.value = d.list || []
    domainPage.total = d.total || 0
  } catch { domainTableData.value = [] } finally { domainLoading.value = false }
}

async function fetchCerts() {
  certLoading.value = true
  try {
    const params = { page: certPage.page, pageSize: certPage.pageSize, ...certSearch }
    const res = await getSslCertList(params)
    const d = res?.data?.data || {}
    certTableData.value = d.list || []
    certPage.total = d.total || 0
  } catch { certTableData.value = [] } finally { certLoading.value = false }
}

async function fetchDnsProviders() {
  dnsLoading.value = true
  try {
    const params = { page: dnsPage.page, pageSize: dnsPage.pageSize, ...dnsSearch }
    const res = await getDnsProviderList(params)
    const d = res?.data?.data || {}
    dnsTableData.value = d.list || []
    dnsPage.total = d.total || 0
  } catch { dnsTableData.value = [] } finally { dnsLoading.value = false }
}

// ===================== 重置 =====================
function resetMonitorSearch() { Object.assign(monitorSearch, { domain: '', tag: '', status: '', aliveStatus: '' }); filter30Days.value = false; filterAbnormal.value = false; fetchDomains() }
function resetCertSearch() { Object.assign(certSearch, { domain: '', status: '' }); fetchCerts() }
function resetDnsSearch() { Object.assign(dnsSearch, { name: '', status: '' }); fetchDnsProviders() }

// ===================== 域名 CRUD =====================
function openDomainDialog(row = null) {
  Object.assign(domainForm, { id: null, domain: '', tags: [], protocol: 'https', checkInterval: 300, enabled: true, remark: '' })
  if (row) {
    Object.assign(domainForm, { id: row.id, domain: row.domain, tags: row.tags || [], protocol: row.protocol || 'https', checkInterval: row.checkInterval || 300, enabled: row.enabled !== false, remark: row.remark || '' })
    domainDialogTitle.value = '编辑域名'
  } else { domainDialogTitle.value = '添加域名' }
  domainDialogVisible.value = true
}

async function handleDomainSubmit() {
  await domainFormRef.value?.validate()
  submitting.value = true
  try {
    if (domainForm.id) { await updateDomain(domainForm.id, domainForm) } else { await createDomain(domainForm) }
    ElMessage.success('操作成功'); domainDialogVisible.value = false; fetchDomains(); fetchStats()
  } catch (e) { ElMessage.error(e?.message || '操作失败') } finally { submitting.value = false }
}

async function handleDeleteDomain(row) {
  await ElMessageBox.confirm(`确认删除域名「${row.domain}」？`, '提示', { type: 'warning' })
  try { await deleteDomain(row.id); ElMessage.success('删除成功'); fetchDomains(); fetchStats() }
  catch (e) { ElMessage.error(e?.message || '删除失败') }
}

async function toggleDomain(row, val) {
  try { await toggleDomainEnabled(row.id, { enabled: val }); row.enabled = val } catch { ElMessage.error('操作失败') }
}

function handleCheckDomain(row) { ElMessage.info(`正在检查 ${row.domain}...`) }
function handleBatchUpload() { ElMessage.info('批量上传功能待实现') }
function handleCloudImport() { ElMessage.info('云导入功能待实现') }
function handleAutoMonitor() { ElMessage.info('自动监控设置待实现') }

// ===================== SSL证书 =====================
function openApplyDialog() {
  Object.assign(applyForm, { domain: '', dnsConfigId: null, certSource: 'ACME', caProvider: 'LetsEncrypt', keyAlgorithm: 'EC256', email: '' })
  applyDialogVisible.value = true
}

async function handleApplyCert() {
  await applyFormRef.value?.validate()
  submitting.value = true
  try { await applySslCert(applyForm); ElMessage.success('证书申请已提交'); applyDialogVisible.value = false; fetchCerts() }
  catch (e) { ElMessage.error(e?.message || '申请失败') } finally { submitting.value = false }
}

function openUploadDialog() {
  Object.assign(uploadForm, { domain: '', certPem: '', keyPem: '' })
  uploadDialogVisible.value = true
}

async function handleUploadCert() {
  await uploadFormRef.value?.validate()
  submitting.value = true
  try { await uploadSslCert(uploadForm); ElMessage.success('证书上传成功'); uploadDialogVisible.value = false; fetchCerts() }
  catch (e) { ElMessage.error(e?.message || '上传失败') } finally { submitting.value = false }
}

async function handleDeleteCert(row) {
  await ElMessageBox.confirm(`确认删除 ${row.domain} 的证书？`, '提示', { type: 'warning' })
  try { await deleteSslCert(row.id); ElMessage.success('删除成功'); fetchCerts() }
  catch (e) { ElMessage.error(e?.message || '删除失败') }
}

async function handleDownloadCert(row) {
  try {
    const res = await downloadSslCert(row.id)
    const url = URL.createObjectURL(new Blob([res.data]))
    const a = document.createElement('a'); a.href = url; a.download = `${row.domain}.zip`; a.click(); URL.revokeObjectURL(url)
  } catch { ElMessage.error('下载失败') }
}

function handleDeployCert(row) { ElMessage.info(`部署证书: ${row.domain}`) }

// ===================== DNS配置 CRUD =====================
function openDnsDialog(row = null) {
  Object.assign(dnsForm, { id: null, name: '', provider: '', accessKey: '', accessSecret: '', zoneId: '', region: '', email: '', phone: '', status: 'active' })
  if (row) {
    Object.assign(dnsForm, { id: row.id, name: row.name, provider: row.provider, accessKey: row.accessKey, accessSecret: '', email: row.email || '', phone: row.phone || '', status: row.status || 'active', zoneId: row.zoneId || '', region: row.region || '' })
    dnsDialogTitle.value = '编辑配置'
  } else { dnsDialogTitle.value = '添加配置' }
  dnsDialogVisible.value = true
}

async function handleDnsSubmit() {
  await dnsFormRef.value?.validate()
  submitting.value = true
  try {
    if (dnsForm.id) { await updateDnsProvider(dnsForm.id, dnsForm) } else { await createDnsProvider(dnsForm) }
    ElMessage.success('操作成功'); dnsDialogVisible.value = false; fetchDnsProviders()
  } catch (e) { ElMessage.error(e?.message || '操作失败') } finally { submitting.value = false }
}

async function handleTestDns(row) {
  try { await testDnsProvider(row.id); ElMessage.success('连接测试成功') } catch (e) { ElMessage.error(e?.message || '测试失败') }
}

async function handleTestDnsForm() {
  testingDns.value = true
  try { await testDnsProvider(dnsForm.id); ElMessage.success('连接测试成功') } catch (e) { ElMessage.error(e?.message || '测试失败') } finally { testingDns.value = false }
}

async function handleDeleteDns(row) {
  await ElMessageBox.confirm(`确认删除配置「${row.name}」？`, '提示', { type: 'warning' })
  try { await deleteDnsProvider(row.id); ElMessage.success('删除成功'); fetchDnsProviders() }
  catch (e) { ElMessage.error(e?.message || '删除失败') }
}

// ===================== 工具函数 =====================
const getSslTagType = (days) => {
  if (days == null) return 'info'
  if (days < 0) return 'danger'
  if (days <= 15) return 'danger'
  if (days <= 30) return 'warning'
  return 'success'
}

const getCertStatusType = (s) => ({ '-1': 'info', 1: 'success', 0: 'danger', 2: 'warning' }[String(s)] || 'info')
const getCertStatusLabel = (s) => ({ '-1': '申请中', 1: '已签发', 0: '申请失败', 2: '已过期' }[String(s)] || '-')

const getProviderLabel = (v) => CLOUD_PROVIDERS.find(p => p.value === v)?.label || v
const getProviderTagType = (v) => {
  const types = { aliyun: '', tencent: 'success', huawei: 'danger', aws: 'warning', cloudflare: 'info', godaddy: '', dnspod: 'success', baidu: 'warning', volcengine: 'danger' }
  return types[v] || 'info'
}

const getProviderFieldLabel = (provider, field) => {
  const p = CLOUD_PROVIDERS.find(x => x.value === provider)
  return field === 'key' ? (p?.keyLabel || 'AccessKey') : (p?.secretLabel || 'AccessSecret')
}
const getProviderFieldPlaceholder = (provider, field) => {
  const p = CLOUD_PROVIDERS.find(x => x.value === provider)
  return field === 'key' ? `请输入 ${p?.keyLabel || 'AccessKey'}` : `请输入 ${p?.secretLabel || 'AccessSecret'}`
}

const getProviderHint = (v) => PROVIDER_HINTS[v] || ''

const maskAccessKey = (key) => {
  if (!key) return '-'
  if (key.length <= 8) return '****'
  return key.slice(0, 4) + '****' + key.slice(-4)
}
</script>

<style scoped>
.domain-management {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

/* 统计卡片 */
.stats-row {
  display: flex;
  gap: 12px;
}

.stat-card {
  flex: 1;
  background: #fff;
  border-radius: 10px;
  border: 1px solid #e4e7ed;
  padding: 16px 20px;
  display: flex;
  align-items: center;
  gap: 14px;
  transition: box-shadow 0.2s;
}

.stat-card:hover {
  box-shadow: 0 4px 16px rgba(0,0,0,0.08);
}

.stat-icon-wrap {
  width: 48px;
  height: 48px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.stat-info {
  flex: 1;
}

.stat-label {
  font-size: 12px;
  color: #909399;
  margin-bottom: 4px;
}

.stat-value {
  font-size: 28px;
  font-weight: 700;
  color: #303133;
  line-height: 1;
}

/* 内容卡片 */
.content-card {
  background: #fff;
  border-radius: 10px;
  border: 1px solid #e4e7ed;
  padding: 20px;
}

.section-title {
  font-size: 16px;
  font-weight: 600;
  color: #303133;
  margin-bottom: 16px;
}

/* Tab 样式 */
.domain-tabs :deep(.el-tabs__header) {
  margin-bottom: 16px;
}

/* 工具栏行 */
.bar {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
  margin-bottom: 10px;
}

/* SSL 日期 */
.ssl-date {
  font-size: 12px;
  color: #606266;
  margin-bottom: 2px;
}

/* AccessKey 遮码 */
.access-key {
  font-family: monospace;
  color: #606266;
}

/* 分页 */
.pagination {
  margin-top: 12px;
  display: flex;
  justify-content: flex-end;
}
</style>
