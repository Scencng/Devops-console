<template>
  <div class="service-management">
    <div class="page-header">
      <h2>Service管理</h2>
      <div class="header-actions">
        <div class="namespace-selector">
          <span class="selector-label">命名空间：</span>
          <el-select v-model="selectedNamespace" placeholder="选择命名空间" @change="handleNamespaceChange" style="width: 200px;">
            <el-option
              key="all"
              label="所有"
              value="all"
            />
            <el-option
              v-for="ns in namespaceList"
              :key="ns.name"
              :label="ns.name"
              :value="ns.name"
            />
          </el-select>
        </div>
        <el-button type="primary" @click="showCreateDialog = true" v-show="permStore.hasPerm('k8s:service:showcreatedialogtrue')" >
          <el-icon><Plus /></el-icon>
          创建Service
        </el-button>
        <el-button @click="handleRefresh">
          <el-icon><Refresh /></el-icon>
          刷新
        </el-button>
      </div>
    </div>

    <!-- 搜索框 -->
    <el-card class="search-card">
      <el-row :gutter="20">
        <el-col :span="8">
          <el-input
            v-model="searchQuery"
            placeholder="搜索Service名称"
            clearable
            @input="handleSearch"
          >
            <template #prefix>
              <el-icon><Search /></el-icon>
            </template>
          </el-input>
        </el-col>
      </el-row>
    </el-card>

    <!-- Service列表 -->
    <el-card class="content-card">
      <div class="autoops-table-wrapper">
        <el-table
          v-loading="loading"
          :data="filteredServices"
          @selection-change="handleSelectionChange"
          stripe
          style="width: 100%"
          height="calc(100vh - 280px)"
          :empty-text="loading ? '加载中...' : '暂无数据'"
        >
        <el-table-column type="selection" width="55" />
        <el-table-column prop="name" label="名称" min-width="150" show-overflow-tooltip>
          <template #default="{ row }">
            <el-link type="primary" @click="handleViewDetail(row)">
              {{ row.name }}
            </el-link>
          </template>
        </el-table-column>
        <el-table-column prop="namespace" label="命名空间" width="120" />
        <el-table-column prop="type" label="类型" width="100">
          <template #default="{ row }">
            <el-tag :type="getServiceTypeTag(row.type)">
              {{ row.type }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="clusterIP" label="ClusterIP" width="140">
          <template #default="{ row }">
            <span v-if="row.clusterIP">{{ row.clusterIP }}</span>
            <el-tag v-else type="info" size="small">None</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="ports" label="端口" min-width="200">
          <template #default="{ row }">
            <span v-if="row.ports && row.ports.length > 0">
              {{ formatPorts(row.ports) }}
            </span>
            <span v-else class="text-muted">-</span>
          </template>
        </el-table-column>
        <el-table-column prop="selector" label="选择器" min-width="150" show-overflow-tooltip>
          <template #default="{ row }">
            <span v-if="row.selector">
              {{ formatSelector(row.selector) }}
            </span>
            <span v-else class="text-muted">-</span>
          </template>
        </el-table-column>
        <el-table-column prop="age" label="创建时间" width="120">
          <template #default="{ row }">
            {{ calculateAge(row.creationTimestamp) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button-group>
              <el-button size="small" @click="handleViewDetail(row)" v-show="permStore.hasPerm('k8s:service:handleviewdetail')" >
                <el-icon><View /></el-icon>
                详情
              </el-button>
              <el-button size="small" @click="handleEdit(row)" v-show="permStore.hasPerm('k8s:service:handleedit')" >
                <el-icon><Edit /></el-icon>
                编辑
              </el-button>
              <el-button size="small" type="danger" @click="handleDelete(row)" v-show="permStore.hasPerm('k8s:service:handledelete')" >
                <el-icon><Delete /></el-icon>
                删除
              </el-button>
            </el-button-group>
          </template>
        </el-table-column>
      </el-table>
      </div>
    </el-card>

    <!-- 批量操作 -->
    <div v-if="selectedServices.length > 0" class="batch-actions">
      <el-alert
        :title="`已选择 ${selectedServices.length} 个Service`"
        type="info"
        :closable="false"
      >
        <template #default>
          <el-button size="small" @click="handleBatchDelete" type="danger" v-show="permStore.hasPerm('k8s:service:handlebatchdelete')" >批量删除</el-button>
          <el-button size="small" @click="clearSelection">取消选择</el-button>
        </template>
      </el-alert>
    </div>

    <!-- 创建Service对话框 -->
    <el-dialog v-model="showCreateDialog" title="创建 Service" width="900px" :close-on-click-modal="false">
      <el-tabs v-model="activeTab" type="border-card">
        <el-tab-pane label="表单创建" name="form">
          <el-form :model="createForm" :rules="createRules" ref="createFormRef" label-width="120px">
            <el-row :gutter="20">
              <el-col :span="12">
                <el-form-item label="命名空间" prop="namespace">
                  <el-select v-model="createForm.namespace" placeholder="选择命名空间" style="width: 100%;">
                    <el-option
                      v-for="ns in namespaceList"
                      :key="ns.name"
                      :label="ns.name"
                      :value="ns.name"
                    />
                  </el-select>
                </el-form-item>
              </el-col>
              <el-col :span="12">
                <el-form-item label="Service 名称" prop="name">
                  <el-input v-model="createForm.name" placeholder="请输入 Service 名称" />
                </el-form-item>
              </el-col>
            </el-row>
            
            <el-row :gutter="20">
              <el-col :span="12">
                <el-form-item label="Service 类型" prop="type">
                  <el-select v-model="createForm.type" placeholder="选择类型" style="width: 100%;">
                    <el-option label="ClusterIP" value="ClusterIP" />
                    <el-option label="NodePort" value="NodePort" />
                    <el-option label="LoadBalancer" value="LoadBalancer" />
                    <el-option label="ExternalName" value="ExternalName" />
                  </el-select>
                </el-form-item>
              </el-col>
              <el-col :span="12">
                <el-form-item label="会话亲和性">
                  <el-select v-model="createForm.sessionAffinity" placeholder="选择亲和性" style="width: 100%;">
                    <el-option label="None" value="None" />
                    <el-option label="ClientIP" value="ClientIP" />
                  </el-select>
                </el-form-item>
              </el-col>
            </el-row>
            
            <el-form-item v-if="createForm.type === 'ExternalName'" label="外部名称">
              <el-input v-model="createForm.externalName" placeholder="请输入外部服务名称" />
            </el-form-item>
            
            <el-divider content-position="left">选择器配置</el-divider>
            
            <el-form-item label="选择器">
              <div v-for="(selector, index) in createForm.selector" :key="index" class="selector-config">
                <el-row :gutter="10">
                  <el-col :span="10">
                    <el-input v-model="selector.key" placeholder="选择器键（如：app）" />
                  </el-col>
                  <el-col :span="10">
                    <el-input v-model="selector.value" placeholder="选择器值（如：nginx）" />
                  </el-col>
                  <el-col :span="4">
                    <el-button type="danger" size="small" @click="removeSelector(index)">删除</el-button>
                  </el-col>
                </el-row>
              </div>
              <el-button type="primary" size="small" @click="addSelector">添加选择器</el-button>
            </el-form-item>
            
            <el-divider content-position="left">端口配置</el-divider>
            
            <div v-for="(port, index) in createForm.ports" :key="index" class="port-config">
              <el-card class="port-card">
                <template #header>
                  <div class="port-header">
                    <span>端口 {{ index + 1 }}</span>
                    <el-button 
                      v-if="createForm.ports.length > 1"
                      type="danger" 
                      size="small" 
                      text
                      @click="removePort(index)"
                    >
                      <el-icon><Delete /></el-icon>
                      删除
                    </el-button>
                  </div>
                </template>
                
                <el-row :gutter="20">
                  <el-col :span="8">
                    <el-form-item :label="`端口名称`">
                      <el-input v-model="port.name" placeholder="端口名称（可选）" />
                    </el-form-item>
                  </el-col>
                  <el-col :span="8">
                    <el-form-item :label="`服务端口`" :rules="{ required: true, message: '请输入服务端口', trigger: 'blur' }">
                      <el-input-number v-model="port.port" :min="1" :max="65535" style="width: 100%;" />
                    </el-form-item>
                  </el-col>
                  <el-col :span="8">
                    <el-form-item label="目标端口">
                      <el-input-number v-model="port.targetPort" :min="1" :max="65535" style="width: 100%;" />
                    </el-form-item>
                  </el-col>
                </el-row>
                
                <el-row :gutter="20" v-if="createForm.type === 'NodePort'">
                  <el-col :span="12">
                    <el-form-item label="节点端口">
                      <el-input-number v-model="port.nodePort" :min="30000" :max="32767" style="width: 100%;" />
                    </el-form-item>
                  </el-col>
                  <el-col :span="12">
                    <el-form-item label="协议">
                      <el-select v-model="port.protocol" placeholder="选择协议" style="width: 100%;">
                        <el-option label="TCP" value="TCP" />
                        <el-option label="UDP" value="UDP" />
                        <el-option label="SCTP" value="SCTP" />
                      </el-select>
                    </el-form-item>
                  </el-col>
                </el-row>
                
                <el-row :gutter="20" v-else>
                  <el-col :span="24">
                    <el-form-item label="协议">
                      <el-select v-model="port.protocol" placeholder="选择协议" style="width: 200px;">
                        <el-option label="TCP" value="TCP" />
                        <el-option label="UDP" value="UDP" />
                        <el-option label="SCTP" value="SCTP" />
                      </el-select>
                    </el-form-item>
                  </el-col>
                </el-row>
              </el-card>
            </div>
            
            <el-button type="primary" @click="addPort">添加端口</el-button>
            
            <el-divider content-position="left">高级配置</el-divider>
            
            <el-form-item label="标签">
              <div v-for="(label, index) in createForm.labels" :key="index" class="label-config">
                <el-row :gutter="10">
                  <el-col :span="10">
                    <el-input v-model="label.key" placeholder="标签键" />
                  </el-col>
                  <el-col :span="10">
                    <el-input v-model="label.value" placeholder="标签值" />
                  </el-col>
                  <el-col :span="4">
                    <el-button type="danger" size="small" @click="removeLabel(index)">删除</el-button>
                  </el-col>
                </el-row>
              </div>
              <el-button type="primary" size="small" @click="addLabel">添加标签</el-button>
            </el-form-item>
            
            <el-form-item label="注解">
              <div v-for="(annotation, index) in createForm.annotations" :key="index" class="annotation-config">
                <el-row :gutter="10">
                  <el-col :span="10">
                    <el-input v-model="annotation.key" placeholder="注解键" />
                  </el-col>
                  <el-col :span="10">
                    <el-input v-model="annotation.value" placeholder="注解值" />
                  </el-col>
                  <el-col :span="4">
                    <el-button type="danger" size="small" @click="removeAnnotation(index)">删除</el-button>
                  </el-col>
                </el-row>
              </div>
              <el-button type="primary" size="small" @click="addAnnotation">添加注解</el-button>
            </el-form-item>
            
            <el-form-item label="外部IP" v-if="createForm.type !== 'ExternalName'">
              <div v-for="(ip, index) in createForm.externalIPs" :key="index" class="ip-config">
                <el-row :gutter="10">
                  <el-col :span="20">
                    <el-input v-model="ip.value" placeholder="外部IP地址" />
                  </el-col>
                  <el-col :span="4">
                    <el-button type="danger" size="small" @click="removeExternalIP(index)" v-show="permStore.hasPerm('k8s:service:removeexternalip')" >删除</el-button>
                  </el-col>
                </el-row>
              </div>
              <el-button type="primary" size="small" @click="addExternalIP">添加外部IP</el-button>
            </el-form-item>
          </el-form>
        </el-tab-pane>
        
        <el-tab-pane label="YAML编辑" name="yaml">
          <div class="yaml-editor">
            <el-alert
              title="提示"
              type="info"
              description="您可以直接编辑YAML配置，或者先在表单中填写配置，然后切换到YAML查看生成的配置"
              show-icon
              :closable="false"
            />
            <div class="editor-toolbar">
              <el-button size="small" @click="formatYAML">格式化</el-button>
              <el-button size="small" @click="validateYAML">验证</el-button>
              <el-button size="small" @click="downloadYAML">下载</el-button>
            </div>
            <el-input
              v-model="yamlContent"
              type="textarea"
              :rows="20"
              placeholder="请输入Service的YAML配置..."
              class="yaml-textarea"
            />
          </div>
        </el-tab-pane>
      </el-tabs>
      
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="showCreateDialog = false">取消</el-button>
          <el-button type="primary" @click="handleCreate" :loading="submitting">创建</el-button>
        </div>
      </template>
    </el-dialog>

    <!-- Service详情对话框 -->
    <el-dialog
      v-model="detailDialogVisible"
      title="Service详情"
      width="80%"
      :before-close="handleCloseDetail"
    >
      <div v-if="selectedService" class="service-detail">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="名称">{{ selectedService.name || '-' }}</el-descriptions-item>
          <el-descriptions-item label="命名空间">{{ selectedService.namespace || '-' }}</el-descriptions-item>
          <el-descriptions-item label="类型">
            <el-tag :type="getServiceTypeTag(selectedService.type)">
              {{ selectedService.type || '-' }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="ClusterIP">
            <span v-if="selectedService.ip">{{ selectedService.ip }}</span>
            <el-tag v-else type="info" size="small">None</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="外部IP">
            <span v-if="selectedService.events && selectedService.events.length > 0">
              {{ selectedService.events.map(e => e.IP || e.Hostname).join(', ') }}
            </span>
            <span v-else>-</span>
          </el-descriptions-item>
          <el-descriptions-item label="会话亲和性">
            {{ selectedService.sessionAffinity || 'None' }}
          </el-descriptions-item>
          <el-descriptions-item label="外部流量策略">
            {{ selectedService.externalTrafficPolicy || 'Cluster' }}
          </el-descriptions-item>
          <el-descriptions-item label="创建时间">
            {{ formatAge(selectedService.age) }}
          </el-descriptions-item>
          <el-descriptions-item label="IP列表">
            <span v-if="selectedService.ips && selectedService.ips.length > 0">
              {{ selectedService.ips.join(', ') }}
            </span>
            <span v-else>-</span>
          </el-descriptions-item>
        </el-descriptions>

        <h3>端口配置</h3>
        <el-table :data="selectedService.ports" border v-if="selectedService.ports && selectedService.ports.length > 0">
          <el-table-column prop="name" label="名称">
            <template #default="{ row }">
              {{ row.name || '-' }}
            </template>
          </el-table-column>
          <el-table-column prop="protocol" label="协议">
            <template #default="{ row }">
              <el-tag :type="row.protocol === 'TCP' ? 'primary' : 'success'" size="small">
                {{ row.protocol }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="port" label="端口" />
          <el-table-column prop="targetPort" label="目标端口">
            <template #default="{ row }">
              {{ typeof row.targetPort === 'string' ? row.targetPort : row.targetPort }}
            </template>
          </el-table-column>
          <el-table-column prop="nodePort" label="NodePort">
            <template #default="{ row }">
              {{ row.nodePort || '-' }}
            </template>
          </el-table-column>
        </el-table>
        <div v-else class="no-data">
          <span>暂无端口配置</span>
        </div>

        <h3>选择器</h3>
        <div v-if="selectedService.selector && Object.keys(selectedService.selector).length > 0">
          <el-tag
            v-for="(value, key) in selectedService.selector"
            :key="key"
            size="small"
            class="tag-margin"
          >
            {{ key }}: {{ value }}
          </el-tag>
        </div>
        <div v-else class="no-data">
          <span>暂无选择器配置</span>
        </div>

        <h3>标签</h3>
        <div v-if="selectedService.labels && Object.keys(selectedService.labels).length > 0">
          <el-tag
            v-for="(value, key) in selectedService.labels"
            :key="key"
            size="small"
            class="tag-margin"
          >
            {{ key }}: {{ value }}
          </el-tag>
        </div>
        <div v-else class="no-data">
          <span>暂无标签</span>
        </div>

        <h3>注解</h3>
        <div v-if="selectedService.annotations && Object.keys(selectedService.annotations).length > 0">
          <el-tag
            v-for="(value, key) in selectedService.annotations"
            :key="key"
            size="small"
            class="tag-margin annotation-tag"
          >
            {{ key }}: {{ value }}
          </el-tag>
        </div>
        <div v-else class="no-data">
          <span>暂无注解</span>
        </div>

        <h3>端点信息</h3>
        <div v-if="selectedService.endpoints">
          <pre class="endpoints-info">{{ formatEndpoints(selectedService.endpoints) }}</pre>
        </div>
        <div v-else class="no-data">
          <span>暂无端点信息</span>
        </div>
      </div>
    </el-dialog>

    <!-- 编辑Service对话框 -->
    <el-dialog v-model="showEditDialog" title="编辑 Service" width="900px" :close-on-click-modal="false">
      <el-form :model="editForm" :rules="editRules" ref="editFormRef" label-width="120px">
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="Service 名称" prop="name">
              <el-input v-model="editForm.name" disabled placeholder="Service名称不可修改" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="命名空间" prop="namespace">
              <el-input v-model="editForm.namespace" disabled placeholder="命名空间不可修改" />
            </el-form-item>
          </el-col>
        </el-row>
        
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="Service类型" prop="type">
              <el-select v-model="editForm.type" placeholder="选择Service类型" style="width: 100%;">
                <el-option label="ClusterIP" value="ClusterIP" />
                <el-option label="NodePort" value="NodePort" />
                <el-option label="LoadBalancer" value="LoadBalancer" />
                <el-option label="ExternalName" value="ExternalName" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="会话亲和性">
              <el-select v-model="editForm.sessionAffinity" placeholder="选择会话亲和性" style="width: 100%;">
                <el-option label="None" value="None" />
                <el-option label="ClientIP" value="ClientIP" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
        
        <el-divider content-position="left">选择器配置</el-divider>
        
        <el-form-item label="选择器标签">
          <div v-for="(selector, index) in editForm.selector" :key="index" class="selector-config">
            <el-row :gutter="10">
              <el-col :span="10">
                <el-input v-model="selector.key" placeholder="选择器键（如：app）" />
              </el-col>
              <el-col :span="10">
                <el-input v-model="selector.value" placeholder="选择器值（如：nginx）" />
              </el-col>
              <el-col :span="4">
                <el-button type="danger" size="small" @click="removeEditSelector(index)">删除</el-button>
              </el-col>
            </el-row>
          </div>
          <el-button type="primary" size="small" @click="addEditSelector">添加选择器</el-button>
        </el-form-item>
        
        <el-divider content-position="left">Service标签</el-divider>
        
        <el-form-item label="标签">
          <div v-for="(label, index) in editForm.labels" :key="index" class="label-config">
            <el-row :gutter="10">
              <el-col :span="10">
                <el-input v-model="label.key" placeholder="标签键" />
              </el-col>
              <el-col :span="10">
                <el-input v-model="label.value" placeholder="标签值" />
              </el-col>
              <el-col :span="4">
                <el-button type="danger" size="small" @click="removeEditLabel(index)">删除</el-button>
              </el-col>
            </el-row>
          </div>
          <el-button type="primary" size="small" @click="addEditLabel">添加标签</el-button>
        </el-form-item>
        
        <el-divider content-position="left">端口配置</el-divider>
        
        <div v-for="(port, index) in editForm.ports" :key="index" class="port-config">
          <el-card class="port-card">
            <template #header>
              <div class="port-header">
                <span>端口 {{ index + 1 }}</span>
                <el-button 
                  v-if="editForm.ports.length > 1"
                  type="danger" 
                  size="small" 
                  text
                  @click="removeEditPort(index)"
                >
                  <el-icon><Delete /></el-icon>
                  删除
                </el-button>
              </div>
            </template>
            
            <el-row :gutter="20">
              <el-col :span="8">
                <el-form-item :label="`端口名称`">
                  <el-input v-model="port.name" placeholder="端口名称（可选）" />
                </el-form-item>
              </el-col>
              <el-col :span="8">
                <el-form-item :label="`服务端口`" :rules="{ required: true, message: '服务端口不能为空', trigger: 'blur' }">
                  <el-input-number v-model="port.port" :min="1" :max="65535" placeholder="服务端口" style="width: 100%" />
                </el-form-item>
              </el-col>
              <el-col :span="8">
                <el-form-item :label="`目标端口`">
                  <el-input v-model="port.targetPort" placeholder="目标端口" />
                </el-form-item>
              </el-col>
            </el-row>
            
            <el-row :gutter="20">
              <el-col :span="8">
                <el-form-item label="NodePort">
                  <el-input-number v-model="port.nodePort" :min="30000" :max="32767" placeholder="NodePort（仅NodePort类型）" style="width: 100%" />
                </el-form-item>
              </el-col>
              <el-col :span="8">
                <el-form-item label="协议">
                  <el-select v-model="port.protocol" placeholder="选择协议">
                    <el-option label="TCP" value="TCP" />
                    <el-option label="UDP" value="UDP" />
                  </el-select>
                </el-form-item>
              </el-col>
            </el-row>
          </el-card>
        </div>
        
        <el-button type="primary" @click="addEditPort">添加端口</el-button>
      </el-form>
      
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="showEditDialog = false">取消</el-button>
          <el-button type="primary" @click="handleUpdate" :loading="submitting">更新</el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { usePermissionStore } from '@/stores/permissionStore.js'
const permStore = usePermissionStore()

import { ref, computed, onMounted, watch, nextTick } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Plus, Refresh, Search, View, Edit, Delete
} from '@element-plus/icons-vue'
import {
  getServiceList,
  getServiceDetail,
  deleteService,
  deleteMultipleServices,
  createService
} from '@/api/k8s/service.js'
import { getNamespaceList } from '@/api/k8s/namespace.js'
import { getSelectedInstanceId } from '@/stores/instanceStore'
import yaml from 'js-yaml'

// 响应式数据
const loading = ref(false)
const submitting = ref(false)
const services = ref([])
const searchQuery = ref('')
const selectedServices = ref([])
const detailDialogVisible = ref(false)
const selectedService = ref(null)
const namespaceList = ref([])
const selectedNamespace = ref('all')
const showCreateDialog = ref(false)
const showEditDialog = ref(false)
const createFormRef = ref(null)
const editFormRef = ref(null)
const activeTab = ref('form')
const yamlContent = ref('')

const createForm = ref({
  namespace: 'default',
  name: '',
  type: 'ClusterIP',
  sessionAffinity: 'None',
  externalName: '',
  selector: [],
  ports: [{
    name: '',
    port: 80,
  }]
})

const editForm = ref({
  name: '',
  namespace: '',
  type: 'ClusterIP',
  sessionAffinity: 'None',
  selector: [],
  labels: [],
  ports: [{
    name: '',
    port: 80,
  }]
})

const editRules = {
  name: [{ required: true, message: '请输入Service名称', trigger: 'blur' }],
  namespace: [{ required: true, message: '请选择命名空间', trigger: 'change' }],
  type: [{ required: true, message: '请选择Service类型', trigger: 'change' }]
}

const createRules = {
  namespace: [{ required: true, message: '请选择命名空间', trigger: 'change' }],
  name: [{ required: true, message: '请输入Service名称', trigger: 'blur' }],
  type: [{ required: true, message: '请选择Service类型', trigger: 'change' }]
}

// 计算属性
const filteredServices = computed(() => {
  if (!searchQuery.value) return services.value
  return services.value.filter(service =>
    service.name.toLowerCase().includes(searchQuery.value.toLowerCase())
  )
})

// 方法
const fetchNamespaces = async () => {
  try {
    const instanceId = getSelectedInstanceId() || '1'
    const response = await getNamespaceList(instanceId)
    namespaceList.value = response.data?.namespaceList || []
  } catch (error) {
    ElMessage.error('获取命名空间列表失败: ' + error.message)
  }
}

const handleNamespaceChange = () => {
  fetchServices()
}

const fetchServices = async () => {
  loading.value = true
  try {
    const instanceId = getSelectedInstanceId() || '1'
    const response = await getServiceList(selectedNamespace.value, instanceId)
    services.value = response.data?.services || []
  } catch (error) {
    ElMessage.error('获取Service列表失败: ' + error.message)
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  // 搜索逻辑已在计算属性中实现
}

const handleRefresh = () => {
  fetchServices()
}

const handleSelectionChange = (selection) => {
  selectedServices.value = selection
}

const clearSelection = () => {
  selectedServices.value = []
}

const handleViewDetail = async (service) => {
  try {
    const instanceId = getSelectedInstanceId() || '1'
    const response = await getServiceDetail(service.namespace, service.name, instanceId)
    selectedService.value = response.data?.serviceDetail
    detailDialogVisible.value = true
  } catch (error) {
    ElMessage.error('获取Service详情失败: ' + error.message)
  }
}

const handleCloseDetail = () => {
  detailDialogVisible.value = false
  selectedService.value = null
}

const handleEdit = async (service) => {
  try {
    const instanceId = getSelectedInstanceId() || '1'
    const response = await getServiceDetail(service.namespace, service.name, instanceId)
    selectedService.value = response.data?.serviceDetail
    
    // 初始化编辑表单数据
    editForm.value = {
      name: selectedService.value.name,
      namespace: selectedService.value.namespace,
      type: selectedService.value.type,
      sessionAffinity: selectedService.value.sessionAffinity || 'None',
      selector: selectedService.value.selector ? Object.entries(selectedService.value.selector).map(([key, value]) => ({ key, value })) : [],
      labels: selectedService.value.labels ? Object.entries(selectedService.value.labels).map(([key, value]) => ({ key, value })) : [],
      ports: selectedService.value.ports ? selectedService.value.ports.map(port => ({
        name: port.name || '',
        port: port.port,
        targetPort: typeof port.targetPort === 'string' ? port.targetPort : port.targetPort.toString(),
        nodePort: port.nodePort || '',
        protocol: port.protocol
      })) : []
    }
    
    showEditDialog.value = true
  } catch (error) {
    ElMessage.error('获取Service详情失败: ' + error.message)
  }
}

const handleDelete = async (service) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除Service "${service.name}" 吗？`,
      '确认删除',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    const instanceId = getSelectedInstanceId() || '1'
    await deleteService(service.namespace, service.name, instanceId)
    ElMessage.success('Service删除成功')
    await fetchServices()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除Service失败: ' + error.message)
    }
  }
}

const handleBatchDelete = async () => {
  try {
    await ElMessageBox.confirm(
      `确定要删除选中的 ${selectedServices.value.length} 个Service吗？`,
      '确认批量删除',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    const instanceId = getSelectedInstanceId() || '1'
    // 逐个删除，因为每个Service可能属于不同的命名空间
    for (const service of selectedServices.value) {
      await deleteService(service.namespace, service.name, instanceId)
    }
    ElMessage.success('批量删除成功')
    selectedServices.value = []
    await fetchServices()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('批量删除失败: ' + error.message)
    }
  }
}

// 端口管理方法
const addPort = () => {
  createForm.value.ports.push({
    name: '',
    port: 80,
    targetPort: 80,
    protocol: 'TCP',
    nodePort: null
  })
}

const removePort = (index) => {
  createForm.value.ports.splice(index, 1)
}

// 选择器管理方法
const addSelector = () => {
  createForm.value.selector.push({
    key: '',
    value: ''
  })
}

const removeSelector = (index) => {
  createForm.value.selector.splice(index, 1)
}

// 标签管理方法
const addLabel = () => {
  createForm.value.labels.push({
    key: '',
    value: ''
  })
}

const removeLabel = (index) => {
  createForm.value.labels.splice(index, 1)
}

// 注解管理方法
const addAnnotation = () => {
  createForm.value.annotations.push({
    key: '',
    value: ''
  })
}

const removeAnnotation = (index) => {
  createForm.value.annotations.splice(index, 1)
}

// 外部IP管理方法
const addExternalIP = () => {
  createForm.value.externalIPs.push({
    value: ''
  })
}

const removeExternalIP = (index) => {
  createForm.value.externalIPs.splice(index, 1)
}

// YAML相关方法
const generateYAML = () => {
  // 验证端口配置
  const validPorts = createForm.value.ports.filter(port => port.port)
  
  if (validPorts.length === 0) {
    ElMessage.error('请至少配置一个有效的端口')
    return ''
  }
  
  // 手动构建YAML字符串，确保格式正确
  let yaml = `apiVersion: v1
kind: Service
metadata:
  name: ${createForm.value.name}
  namespace: ${createForm.value.namespace}`
  
  // 添加标签
  const validLabels = createForm.value.labels.filter(label => label.key && label.value)
  if (validLabels.length > 0) {
    yaml += '\n  labels:'
    validLabels.forEach(label => {
      yaml += `\n    ${label.key}: ${label.value}`
    })
  }
  
  // 添加注解
  const validAnnotations = createForm.value.annotations.filter(annotation => annotation.key && annotation.value)
  if (validAnnotations.length > 0) {
    yaml += '\n  annotations:'
    validAnnotations.forEach(annotation => {
      yaml += `\n    ${annotation.key}: ${annotation.value}`
    })
  }
  
  yaml += `\nspec:
  type: ${createForm.value.type}
  sessionAffinity: ${createForm.value.sessionAffinity}`
  
  // 添加外部名称（ExternalName类型）
  if (createForm.value.type === 'ExternalName' && createForm.value.externalName) {
    yaml += `\n  externalName: ${createForm.value.externalName}`
  }
  
  // 添加选择器
  const validSelectors = createForm.value.selector.filter(selector => selector.key && selector.value)
  if (validSelectors.length > 0) {
    yaml += '\n  selector:'
    validSelectors.forEach(selector => {
      yaml += `\n    ${selector.key}: ${selector.value}`
    })
  }
  
  // 添加端口
  yaml += '\n  ports:'
  validPorts.forEach(port => {
    yaml += `\n    - port: ${port.port}`
    if (port.name) {
      yaml += `\n      name: ${port.name}`
    }
    if (port.targetPort) {
      yaml += `\n      targetPort: ${port.targetPort}`
    }
    yaml += `\n      protocol: ${port.protocol || 'TCP'}`
    if (createForm.value.type === 'NodePort' && port.nodePort) {
      yaml += `\n      nodePort: ${port.nodePort}`
    }
  })
  
  // 添加外部IP
  const validExternalIPs = createForm.value.externalIPs.filter(ip => ip.value)
  if (validExternalIPs.length > 0) {
    yaml += '\n  externalIPs:'
    validExternalIPs.forEach(ip => {
      yaml += `\n    - ${ip.value}`
    })
  }
  
  return yaml
}

const formatYAML = () => {
  try {
    // 如果当前在YAML标签页，尝试解析并重新生成
    if (activeTab.value === 'yaml' && yamlContent.value) {
      try {
        const parsed = yaml.load(yamlContent.value, {
          schema: yaml.FAILSAFE_SCHEMA
        })
        yamlContent.value = yaml.dump(parsed, {
          indent: 2,
          lineWidth: -1,
          noRefs: true,
          sortKeys: false
        })
        ElMessage.success('YAML格式化成功')
      } catch (parseError) {
        // 如果解析失败，至少清理一下多余的空行
        const lines = yamlContent.value.split('\n')
        const cleaned = lines.filter(line => line.trim() !== '' || line === '').join('\n')
        yamlContent.value = cleaned
        ElMessage.warning('YAML已清理（部分格式可能需要手动调整）')
      }
    } else {
      // 如果在表单页，直接重新生成
      yamlContent.value = generateYAML()
      ElMessage.success('YAML重新生成成功')
    }
  } catch (error) {
    ElMessage.error('YAML格式化失败: ' + error.message)
  }
}

const validateYAML = () => {
  try {
    // 如果是手动生成的YAML，尝试解析基本结构
    if (yamlContent.value) {
      const lines = yamlContent.value.split('\n')
      let hasApiVersion = false
      let hasKind = false
      let hasMetadata = false
      let hasSpec = false
      
      for (const line of lines) {
        const trimmed = line.trim()
        if (trimmed.startsWith('apiVersion:')) hasApiVersion = true
        if (trimmed.startsWith('kind:')) hasKind = true
        if (trimmed.startsWith('metadata:')) hasMetadata = true
        if (trimmed.startsWith('spec:')) hasSpec = true
      }
      
      if (hasApiVersion && hasKind && hasMetadata && hasSpec) {
        // 尝试解析，但不严格要求格式完美
        yaml.load(yamlContent.value, {
          schema: yaml.FAILSAFE_SCHEMA
        })
        ElMessage.success('YAML验证通过')
      } else {
        ElMessage.warning('YAML缺少必要字段（apiVersion, kind, metadata, spec）')
      }
    } else {
      ElMessage.warning('YAML内容为空')
    }
  } catch (error) {
    ElMessage.error('YAML验证失败: ' + error.message)
  }
}

const downloadYAML = () => {
  const blob = new Blob([yamlContent.value], { type: 'text/yaml' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `${createForm.value.name || 'service'}.yaml`
  document.body.appendChild(a)
  a.click()
  document.body.removeChild(a)
  URL.revokeObjectURL(url)
}

// 监听标签页切换
watch(activeTab, async (newTab) => {
  if (newTab === 'yaml') {
    // 先验证表单数据
    if (!createForm.value.name) {
      ElMessage.warning('请先填写Service名称')
      activeTab.value = 'form'
      return
    }
    
    // 验证端口配置
    const validPorts = createForm.value.ports.filter(port => port.port)
    if (validPorts.length === 0) {
      ElMessage.warning('请至少配置一个有效的端口')
      activeTab.value = 'form'
      return
    }
    
    // 从表单生成YAML
    yamlContent.value = generateYAML()
  }
})

const handleCreate = async () => {
  submitting.value = true
  
  try {
    const instanceId = getSelectedInstanceId() || '1'
    let createData
    
    if (activeTab.value === 'yaml') {
      // 直接使用YAML创建
      try {
        const serviceSpec = yaml.load(yamlContent.value)
        createData = {
          namespace: serviceSpec.metadata.namespace,
          name: serviceSpec.metadata.name,
          yaml: yamlContent.value
        }
      } catch (error) {
        ElMessage.error('YAML解析失败: ' + error.message)
        return
      }
    } else {
      // 使用表单数据创建
      if (!createFormRef.value) return
      
      // 表单验证
      const valid = await createFormRef.value.validate().catch(() => false)
      if (!valid) return
      
      // 验证端口配置
      if (!createForm.value.ports || createForm.value.ports.length === 0) {
        ElMessage.error('请至少配置一个端口')
        return
      }
      
      for (let i = 0; i < createForm.value.ports.length; i++) {
        const port = createForm.value.ports[i]
        if (!port.port) {
          ElMessage.error(`端口 ${i + 1} 的服务端口不能为空`)
          return
        }
      }
      
      // 验证ExternalName类型
      if (createForm.value.type === 'ExternalName' && !createForm.value.externalName) {
        ElMessage.error('ExternalName类型的Service必须提供外部名称')
        return
      }
      
      createData = createForm.value
    }
    
    await createService(createData.namespace, createData, instanceId)
    ElMessage.success('Service 创建成功')
    showCreateDialog.value = false
    resetCreateForm()
    await fetchServices()
  } catch (error) {
    ElMessage.error('创建 Service 失败: ' + (error.response?.data?.message || error.message || '未知错误'))
  } finally {
    submitting.value = false
  }
}

const resetCreateForm = () => {
  createForm.value = {
    namespace: selectedNamespace.value || 'default',
    name: '',
    type: 'ClusterIP',
    sessionAffinity: 'None',
    externalName: '',
    selector: [],
    ports: [{
      name: '',
      port: 80,
      targetPort: 80,
      protocol: 'TCP',
      nodePort: null
    }],
    labels: [],
    annotations: [],
    externalIPs: []
  }
  yamlContent.value = ''
  activeTab.value = 'form'
}

// 工具方法
const getServiceTypeTag = (type) => {
  const typeMap = {
    'ClusterIP': 'primary',
    'NodePort': 'success',
    'LoadBalancer': 'warning',
    'ExternalName': 'info'
  }
  return typeMap[type] || 'info'
}

const formatPorts = (ports) => {
  return ports.map(port => {
    const parts = [port.port]
    if (port.targetPort) parts.push(port.targetPort)
    if (port.nodePort) parts.push(port.nodePort)
    if (port.protocol) parts.push(port.protocol)
    return parts.join(':')
  }).join(', ')
}

const formatSelector = (selector) => {
  return Object.entries(selector).map(([key, value]) => `${key}=${value}`).join(', ')
}

const formatTime = (timestamp) => {
  if (!timestamp) return '-'
  return new Date(timestamp).toLocaleString()
}

const calculateAge = (timestamp) => {
  if (!timestamp) return '-'
  const now = new Date()
  const created = new Date(timestamp)
  const diff = now - created
  const days = Math.floor(diff / (1000 * 60 * 60 * 24))
  const hours = Math.floor((diff % (1000 * 60 * 60 * 24)) / (1000 * 60 * 60))
  
  if (days > 0) return `${days}天`
  if (hours > 0) return `${hours}小时`
  return '刚刚'
}

const formatAge = (ageTimestamp) => {
  if (!ageTimestamp) return '-'
  const now = new Date()
  const created = new Date(ageTimestamp * 1000)
  const diff = now - created
  const days = Math.floor(diff / (1000 * 60 * 60 * 24))
  const hours = Math.floor((diff % (1000 * 60 * 60 * 24)) / (1000 * 60 * 60))
  
  if (days > 0) return `${days}天`
  if (hours > 0) return `${hours}小时`
  return '刚刚'
}

const formatEndpoints = (endpoints) => {
  if (!endpoints) return '-'
  
  // 如果是字符串（错误消息）
  if (typeof endpoints === 'string') {
    return endpoints
  }
  
  // 如果是对象，格式化显示
  if (typeof endpoints === 'object') {
    return JSON.stringify(endpoints, null, 2)
  }
  
  return '-'
}

// 编辑表单处理方法
const addEditSelector = () => {
  editForm.value.selector.push({ key: '', value: '' })
}

const removeEditSelector = (index) => {
  editForm.value.selector.splice(index, 1)
}

const addEditLabel = () => {
  editForm.value.labels.push({ key: '', value: '' })
}

const removeEditLabel = (index) => {
  editForm.value.labels.splice(index, 1)
}

const addEditPort = () => {
  editForm.value.ports.push({
    name: '',
    port: 80,
    targetPort: '',
    nodePort: '',
    protocol: 'TCP'
  })
}

const removeEditPort = (index) => {
  editForm.value.ports.splice(index, 1)
}

const handleUpdate = async () => {
  submitting.value = true
  
  try {
    if (!editFormRef.value) return
    
    // 表单验证
    const valid = await editFormRef.value.validate().catch(() => false)
    if (!valid) return
    
    // 验证端口配置
    if (!editForm.value.ports || editForm.value.ports.length === 0) {
      ElMessage.error('请至少配置一个端口')
      return
    }
    
    for (let i = 0; i < editForm.value.ports.length; i++) {
      const port = editForm.value.ports[i]
      if (!port.port) {
        ElMessage.error(`端口 ${i + 1} 的服务端口不能为空`)
        return
      }
    }
    
    // 构建更新数据
    const updateData = {
      labels: editForm.value.labels.reduce((acc, label) => {
        if (label.key && label.value) {
          acc[label.key] = label.value
        }
        return acc
      }, {}),
      selector: editForm.value.selector.reduce((acc, selector) => {
        if (selector.key && selector.value) {
          acc[selector.key] = selector.value
        }
        return acc
      }, {}),
      ports: editForm.value.ports.filter(port => port.port).map(port => ({
        name: port.name,
        port: port.port,
        targetPort: port.targetPort,
        nodePort: port.nodePort,
        protocol: port.protocol
      }))
    }
    
    const instanceId = getSelectedInstanceId() || '1'
    await updateService(editForm.value.namespace, editForm.value.name, updateData, instanceId)
    ElMessage.success('Service更新成功')
    showEditDialog.value = false
    await fetchServices()
  } catch (error) {
    ElMessage.error('更新Service失败: ' + (error.response?.data?.message || error.message || '未知错误'))
  } finally {
    submitting.value = false
  }
}

// 生命周期
onMounted(() => {
  fetchNamespaces()
  fetchServices()
})
</script>

<style scoped>
.service-management {
  padding: 24px;
  background: linear-gradient(135deg, #f5f7fa 0%, #e4e7eb 100%);
  min-height: 100vh;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}

.page-header h2 {
  margin: 0;
  color: #1a1a1a;
  font-size: 28px;
  font-weight: 600;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

.namespace-selector {
  display: flex;
  align-items: center;
  gap: 8px;
}

.selector-label {
  font-size: 14px;
  color: var(--el-text-color-regular);
  white-space: nowrap;
}

.search-card, .content-card {
  margin-bottom: 24px;
  border-radius: 12px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.05);
}

.batch-actions {
  margin-bottom: 24px;
}

.service-detail {
  padding: 20px 0;
}

.service-detail h3 {
  margin: 30px 0 15px 0;
  color: #1a1a1a;
  font-size: 18px;
  font-weight: 600;
}

.tag-margin {
  margin-right: 8px;
  margin-bottom: 8px;
}

.annotation-tag {
  background-color: #f0f9ff;
  color: #0369a1;
  border-color: #bae6fd;
}

.no-data {
  color: #999;
  font-style: italic;
  padding: 10px;
  text-align: center;
  background-color: #f5f5f5;
  border-radius: 4px;
  margin: 10px 0;
}

.endpoints-info {
  background-color: #f8f9fa;
  border: 1px solid #e9ecef;
  border-radius: 4px;
  padding: 15px;
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
  font-size: 12px;
  line-height: 1.5;
  color: #495057;
  white-space: pre-wrap;
  word-break: break-all;
  max-height: 200px;
  overflow-y: auto;
}

/* 编辑表单样式 */
.selector-config, .label-config {
  margin-bottom: 10px;
  padding: 10px;
  background-color: #f5f7fa;
  border-radius: 4px;
}

.port-config {
  margin-bottom: 20px;
}

.port-card {
  margin-bottom: 15px;
}

.port-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}

.text-muted {
  color: #999;
}

:deep(.el-table) {
  border-radius: 8px;
}

:deep(.el-descriptions) {
  margin-bottom: 20px;
}

/* 端口配置样式 */
.port-config {
  margin-bottom: 20px;
}

.port-card {
  margin-bottom: 15px;
}

.port-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

/* 选择器配置样式 */
.selector-config {
  margin-bottom: 10px;
  padding: 10px;
  background-color: #f5f7fa;
  border-radius: 4px;
}

/* 标签配置样式 */
.label-config {
  margin-bottom: 10px;
}

.annotation-config {
  margin-bottom: 10px}

/* 外部IP配置样式 */
.ip-config {
  margin-bottom: 10px;
}

/* YAML编辑器样式 */
.yaml-editor {
  padding: 20px;
}

.editor-toolbar {
  margin: 15px 0;
  display: flex;
  gap: 10px;
}

.yaml-textarea {
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
  font-size: 13px;
}

:deep(.yaml-textarea .el-textarea__inner) {
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
  font-size: 13px;
  line-height: 1.5;
}

/* 对话框底部样式 */
.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}

/* 表单分隔线样式 */
:deep(.el-divider__text) {
  font-weight: 600;
  color: var(--el-text-color-primary);
}

/* 卡片间距调整 */
:deep(.el-card__body) {
  padding: 15px;
}

/* 表单项间距调整 */
:deep(.el-form-item) {
  margin-bottom: 15px;
}

.content-card {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-height: 0;
}

.table-container {
  flex: 1;
  overflow: hidden;
}

/* 表格滚动条样式优化 */
:deep(.el-table__body-wrapper) {
  scrollbar-width: thin;
  scrollbar-color: #c1c1c1 transparent;
}

:deep(.el-table__body-wrapper::-webkit-scrollbar) {
  width: 6px;
  height: 6px;
}

:deep(.el-table__body-wrapper::-webkit-scrollbar-track) {
  background: transparent;
}

:deep(.el-table__body-wrapper::-webkit-scrollbar-thumb) {
  background-color: #c1c1c1;
  border-radius: 3px;
}

:deep(.el-table__body-wrapper::-webkit-scrollbar-thumb:hover) {
  background-color: #a8a8a8;
}
</style>