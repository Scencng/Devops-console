<template>
  <div class="pod-management">
    <el-card class="page-header-card">
      <div class="page-header">
        <div>
          <h2>Pod 管理</h2>
          <p>查看和管理 Kubernetes Pod</p>
        </div>
        <div class="header-actions">
          <div class="namespace-selector">
            <span class="selector-label">命名空间：</span>
            <el-select v-model="selectedNamespace" placeholder="选择命名空间" @change="handleNamespaceChange" style="width: 200px;">
              <el-option label="所有" value="all" />
              <el-option
                v-for="ns in namespaceList"
                :key="ns.name"
                :label="ns.name"
                :value="ns.name"
              />
            </el-select>
          </div>
          <el-button type="primary" @click="showCreateDialog = true" v-show="permStore.hasPerm('k8s:pod:showcreatedialogtrue')" >
            <el-icon><Plus /></el-icon>
            创建 Pod
          </el-button>

        </div>
      </div>
    </el-card>

    <el-card class="content-card">
      <div class="autoops-table-wrapper">
        <el-table 
          :data="podList" 
          style="width: 100%" class="autoops-table" 
          v-loading="loading"
          height="calc(100vh - 280px)"
          :empty-text="loading ? '加载中...' : '暂无数据'"
        >
          <el-table-column prop="name" label="名称" min-width="150" show-overflow-tooltip />
          <el-table-column prop="ready" label="就绪" width="100">
            <template #default="scope">
              <el-tag :type="scope.row.ready === '1/1' || scope.row.ready === '2/2' || scope.row.ready === '3/3' ? 'success' : 'warning'">
                {{ scope.row.ready || '0/0' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="status" label="状态" width="100">
            <template #default="scope">
              <el-tag :type="getStatusType(scope.row.status)">
                {{ scope.row.status }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="restarts" label="重启次数" width="100">
            <template #default="scope">
              <el-tag :type="scope.row.restarts > 0 ? 'warning' : 'success'">
                {{ scope.row.restarts || 0 }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="ip" label="IP 地址" width="150" />
          <el-table-column prop="node" label="节点" min-width="120" show-overflow-tooltip />
          <el-table-column prop="namespace" label="命名空间" width="120" />
          <el-table-column label="创建时间" width="180">
            <template #default="scope">
              {{ formatTimestamp(scope.row.age) }}
            </template>
          </el-table-column>
          <el-table-column label="操作" width="280" fixed="right">
            <template #default="scope">
              <el-button-group>
                <el-button size="small" @click="handleViewDetail(scope.row)" v-show="permStore.hasPerm('k8s:pod:handleviewdetail')" >
                  <el-icon><View /></el-icon>
                  详情
                </el-button>
                <el-button size="small" type="primary" @click="handleOpenTerminal(scope.row)" v-show="permStore.hasPerm('k8s:pod:handleopenterminal')" >
                  <el-icon><Monitor /></el-icon>
                  终端
                </el-button>
                <el-button size="small" type="danger" @click="handleDelete(scope.row)" v-show="permStore.hasPerm('k8s:pod:handledelete')" >
                  <el-icon><Delete /></el-icon>
                  删除
                </el-button>
              </el-button-group>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </el-card>

    <el-dialog v-model="showCreateDialog" title="创建 Pod" width="900px" :close-on-click-modal="false" destroy-on-close>
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
                <el-form-item label="Pod 名称" prop="podname">
                  <el-input v-model="createForm.podname" placeholder="请输入 Pod 名称" />
                </el-form-item>
              </el-col>
            </el-row>
            
            <el-divider content-position="left">容器配置</el-divider>
            
            <div v-for="(container, index) in createForm.containers" :key="index" class="container-config">
              <el-card class="container-card">
                <template #header>
                  <div class="container-header">
                    <span>容器 {{ index + 1 }}</span>
                    <el-button 
                      v-if="createForm.containers.length > 1"
                      type="danger" 
                      size="small" 
                      text
                      @click="removeContainer(index)"
                    >
                      <el-icon><Delete /></el-icon>
                      删除
                    </el-button>
                  </div>
                </template>
                
                <el-row :gutter="20">
                  <el-col :span="12">
                    <el-form-item :label="`容器名称`" :prop="`containers.${index}.name`" :rules="{ required: true, message: '请输入容器名称', trigger: 'blur' }">
                      <el-input v-model="container.name" placeholder="请输入容器名称" />
                    </el-form-item>
                  </el-col>
                  <el-col :span="12">
                    <el-form-item :label="`镜像`" :prop="`containers.${index}.image`" :rules="{ required: true, message: '请输入镜像', trigger: 'blur' }">
                      <el-input v-model="container.image" placeholder="例如: nginx:latest" />
                    </el-form-item>
                  </el-col>
                </el-row>
                
                <el-row :gutter="20">
                  <el-col :span="12">
                    <el-form-item label="镜像拉取策略">
                      <el-select v-model="container.imagePullPolicy" placeholder="选择策略">
                        <el-option label="Always" value="Always" />
                        <el-option label="IfNotPresent" value="IfNotPresent" />
                        <el-option label="Never" value="Never" />
                      </el-select>
                    </el-form-item>
                  </el-col>
                  <el-col :span="12">
                    <el-form-item label="重启策略">
                      <el-select v-model="createForm.restartPolicy" placeholder="选择策略">
                        <el-option label="Always" value="Always" />
                        <el-option label="OnFailure" value="OnFailure" />
                        <el-option label="Never" value="Never" />
                      </el-select>
                    </el-form-item>
                  </el-col>
                </el-row>
                
                <el-form-item label="端口配置">
                  <div v-for="(port, portIndex) in container.ports" :key="portIndex" class="port-config">
                    <el-row :gutter="10">
                      <el-col :span="6">
                        <el-input v-model="port.name" placeholder="端口名称" />
                      </el-col>
                      <el-col :span="6">
                        <el-input-number v-model="port.containerPort" :min="1" :max="65535" placeholder="容器端口" style="width: 100%" />
                      </el-col>
                      <el-col :span="6">
                        <el-select v-model="port.protocol" placeholder="协议">
                          <el-option label="TCP" value="TCP" />
                          <el-option label="UDP" value="UDP" />
                        </el-select>
                      </el-col>
                      <el-col :span="6">
                        <el-button type="danger" size="small" @click="removePort(container, portIndex)">删除</el-button>
                      </el-col>
                    </el-row>
                  </div>
                  <el-button type="primary" size="small" @click="addPort(container)">添加端口</el-button>
                </el-form-item>
                
                <el-form-item label="环境变量">
                  <div v-for="(env, envIndex) in container.env" :key="envIndex" class="env-config">
                    <el-row :gutter="10">
                      <el-col :span="8">
                        <el-input v-model="env.name" placeholder="变量名" />
                      </el-col>
                      <el-col :span="12">
                        <el-input v-model="env.value" placeholder="变量值" />
                      </el-col>
                      <el-col :span="4">
                        <el-button type="danger" size="small" @click="removeEnv(container, envIndex)">删除</el-button>
                      </el-col>
                    </el-row>
                  </div>
                  <el-button type="primary" size="small" @click="addEnv(container)">添加环境变量</el-button>
                </el-form-item>
                
                <el-form-item label="资源限制">
                  <el-row :gutter="20">
                    <el-col :span="12">
                      <el-form-item label="CPU请求">
                        <el-input v-model="container.resources.requests.cpu" placeholder="例如: 100m" />
                      </el-form-item>
                    </el-col>
                    <el-col :span="12">
                      <el-form-item label="内存请求">
                        <el-input v-model="container.resources.requests.memory" placeholder="例如: 128Mi" />
                      </el-form-item>
                    </el-col>
                  </el-row>
                  <el-row :gutter="20">
                    <el-col :span="12">
                      <el-form-item label="CPU限制">
                        <el-input v-model="container.resources.limits.cpu" placeholder="例如: 500m" />
                      </el-form-item>
                    </el-col>
                    <el-col :span="12">
                      <el-form-item label="内存限制">
                        <el-input v-model="container.resources.limits.memory" placeholder="例如: 512Mi" />
                      </el-form-item>
                    </el-col>
                  </el-row>
                </el-form-item>
              </el-card>
            </div>
            
            <el-button type="primary" @click="addContainer">添加容器</el-button>
            
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
            
            <el-form-item label="节点选择器">
              <div v-for="(selector, index) in createForm.nodeSelector" :key="index" class="selector-config">
                <el-row :gutter="10">
                  <el-col :span="10">
                    <el-input v-model="selector.key" placeholder="选择器键" />
                  </el-col>
                  <el-col :span="10">
                    <el-input v-model="selector.value" placeholder="选择器值" />
                  </el-col>
                  <el-col :span="4">
                    <el-button type="danger" size="small" @click="removeNodeSelector(index)">删除</el-button>
                  </el-col>
                </el-row>
              </div>
              <el-button type="primary" size="small" @click="addNodeSelector">添加节点选择器</el-button>
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
              placeholder="请输入Pod的YAML配置..."
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

    <el-dialog v-model="showDetailDialog" title="Pod 详情" width="1200px" :close-on-click-modal="false">
      <el-tabs v-model="detailTab" type="border-card">
        <el-tab-pane label="基本信息" name="info">
          <el-descriptions :column="2" border v-if="currentPod">
            <el-descriptions-item label="名称">{{ currentPod.name }}</el-descriptions-item>
            <el-descriptions-item label="命名空间">{{ currentPod.namespace }}</el-descriptions-item>
            <el-descriptions-item label="状态">
              <el-tag :type="getStatusType(currentPod.status)">
                {{ currentPod.status }}
              </el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="IP 地址">{{ currentPod.ip || '-' }}</el-descriptions-item>
            <el-descriptions-item label="节点">{{ currentPod.node || '-' }}</el-descriptions-item>
            <el-descriptions-item label="创建时间">{{ formatTimestamp(currentPod.age) }}</el-descriptions-item>
            <el-descriptions-item label="重启次数">{{ currentPod.restarts || 0 }}</el-descriptions-item>
            <el-descriptions-item label="UID">{{ currentPod.uid || '-' }}</el-descriptions-item>
          </el-descriptions>

          <el-divider content-position="left">标签</el-divider>
          <el-descriptions :column="1" border v-if="currentPod.labels && Object.keys(currentPod.labels).length > 0">
            <el-descriptions-item v-for="(value, key) in currentPod.labels" :key="key" :label="key">
              {{ value }}
            </el-descriptions-item>
          </el-descriptions>
          <div v-else class="no-data">无标签</div>

          <el-divider content-position="left">注解</el-divider>
          <el-descriptions :column="1" border v-if="currentPod.annotations && Object.keys(currentPod.annotations).length > 0">
            <el-descriptions-item v-for="(value, key) in currentPod.annotations" :key="key" :label="key">
              <div class="annotation-value">{{ value }}</div>
            </el-descriptions-item>
          </el-descriptions>
          <div v-else class="no-data">无注解</div>

          <el-divider content-position="left">容器信息</el-divider>
          <div v-if="currentPod.containers && currentPod.containers.length > 0">
            <el-collapse v-model="activeContainers">
              <el-collapse-item v-for="(container, index) in currentPod.containers" :key="index" :title="`容器: ${container.name}`" :name="index">
                <el-descriptions :column="2" border>
                  <el-descriptions-item label="镜像">{{ container.image }}</el-descriptions-item>
                  <el-descriptions-item label="镜像拉取策略">{{ container.imagePullPolicy || '-' }}</el-descriptions-item>
                  <el-descriptions-item label="重启策略">{{ container.restartPolicy || '-' }}</el-descriptions-item>
                  <el-descriptions-item label="端口">
                    <div v-if="container.ports && container.ports.length > 0">
                      <div v-for="port in container.ports" :key="port.containerPort">
                        {{ port.containerPort }}/{{ port.protocol }}{{ port.name ? ` (${port.name})` : '' }}
                      </div>
                    </div>
                    <div v-else>-</div>
                  </el-descriptions-item>
                  <el-descriptions-item label="环境变量">
                    <div v-if="container.env && container.env.length > 0">
                      <div v-for="env in container.env" :key="env.name">
                        {{ env.name }}={{ env.value }}
                      </div>
                    </div>
                    <div v-else>-</div>
                  </el-descriptions-item>
                  <el-descriptions-item label="资源请求">
                    <div v-if="container.resources">
                      CPU: {{ container.resources.requests?.cpu || '-' }}<br>
                      内存: {{ container.resources.requests?.memory || '-' }}
                    </div>
                    <div v-else>-</div>
                  </el-descriptions-item>
                  <el-descriptions-item label="资源限制">
                    <div v-if="container.resources">
                      CPU: {{ container.resources.limits?.cpu || '-' }}<br>
                      内存: {{ container.resources.limits?.memory || '-' }}
                    </div>
                    <div v-else>-</div>
                  </el-descriptions-item>
                </el-descriptions>
              </el-collapse-item>
            </el-collapse>
          </div>
        </el-tab-pane>

        <el-tab-pane label="事件" name="events">
          <el-timeline v-if="podEvents && podEvents.length > 0">
            <el-timeline-item
              v-for="(event, index) in podEvents"
              :key="index"
              :timestamp="formatEventTime(event.lastTimestamp)"
              placement="top"
            >
              <div class="event-item">
                <div class="event-header">
                  <el-tag 
                    :type="getEventTagType(event.type)" 
                    size="small"
                    effect="plain"
                  >
                    {{ event.type }}
                  </el-tag>
                  <span class="event-reason">{{ event.reason }}</span>
                </div>
                <div class="event-message">{{ event.message }}</div>
                <div class="event-source">
                  <el-icon><InfoFilled /></el-icon>
                  {{ event.source.component || event.source.host || 'unknown' }}
                </div>
              </div>
            </el-timeline-item>
          </el-timeline>
          <div v-else class="no-data">暂无事件</div>
        </el-tab-pane>


        <el-tab-pane label="日志" name="logs">
          <div class="log-container">
            <div class="log-toolbar">
              <el-select v-model="selectedContainer" placeholder="选择容器" style="width: 150px;" @change="handleContainerChange">
                <el-option
                  v-for="container in currentPod.containers || []"
                  :key="container.name"
                  :label="container.name"
                  :value="container.name"
                />
              </el-select>
              
              <el-select v-model="logTailLines" placeholder="显示行数" style="width: 120px;" @change="handleTailLinesChange">
                <el-option label="最后 10 行" :value="10" />
                <el-option label="最后 50 行" :value="50" />
                <el-option label="最后 100 行" :value="100" />
                <el-option label="最后 500 行" :value="500" />
                <el-option label="全部" :value="0" />
              </el-select>
              
              <el-switch
                v-model="followLogs"
                active-text="实时跟踪"
                inactive-text="停止跟踪"
              />
              <el-switch
                v-model="showTimestamps"
                active-text="显示时间戳"
                inactive-text="隐藏时间戳"
                @change="handleTimestampToggle"
              />
              
              <div style="flex: 1;"></div>
              
              <el-button @click="refreshLogs" :loading="loadingLogs" size="small">
                <el-icon><Refresh /></el-icon>
                刷新
              </el-button>
              <el-button @click="clearLogs" size="small" v-show="permStore.hasPerm('k8s:pod:clearlogs')" >
                <el-icon><Delete /></el-icon>
                清空
              </el-button>
              <el-button @click="downloadLogs" size="small">
                <el-icon><Download /></el-icon>
                下载
              </el-button>
            </div>
            <div class="log-content-wrapper">
              <div class="log-content" ref="logContentRef">
                <pre v-if="podLogs" class="kubectl-style-logs">{{ podLogs }}</pre>
                <div v-else class="no-logs">暂无日志</div>
              </div>
            </div>
          </div>
        </el-tab-pane>

        <el-tab-pane label="实时监控" name="monitoring">
          <div class="monitor-container" v-if="detailTab === 'monitoring'">
            <!-- 如果没有关联 Prometheus 实例，提示需要在配置里绑定，或者全局找一个 Prometheus 实例 proxy -->
            <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px;">
              <el-select v-model="selectedPrometheusId" placeholder="选择 Prometheus 实例" size="small" style="width: 200px;">
                <el-option v-for="inst in prometheusInstances" :key="inst.id" :label="inst.name" :value="inst.id" />
              </el-select>
              <el-button type="primary" size="small" @click="handleOpenCustomMonitorDialog(null)" v-show="permStore.hasPerm('k8s:pod:handleopencustommonitordialog')" >
                <el-icon><Plus /></el-icon> 新增自定义监控图表
              </el-button>
            </div>
            
            <div v-if="selectedPrometheusId">
              <!-- 自定义图表区 -->
              <el-row :gutter="20" style="margin-top: 10px;">
                <el-col :span="12" v-for="monitor in customMonitors" :key="monitor.id" style="margin-bottom: 20px;">
                  <DynamicPromQLChart 
                    :monitor-id="monitor.id"
                    :node-name="currentPod.node"
                    :pod-name="currentPod.name"
                    :prometheus-instance-id="parseInt(selectedPrometheusId)" 
                    :title="monitor.title" 
                    :chart-type="monitor.chart_type"
                    :unit="monitor.unit_suffix"
                    :color="monitor.color_theme"
                    :promql-template="monitor.promql_template"
                    :hours="1" 
                    @edit="handleOpenCustomMonitorDialog"
                    @delete="handleDeleteCustomMonitor"
                  />
                </el-col>
              </el-row>

            </div>
            <el-empty v-else description="请先选择用于监控的 Prometheus 实例数据源" />
          </div>
        </el-tab-pane>
      </el-tabs>

      <template #footer>
        <el-button @click="showDetailDialog = false">关闭</el-button>
      </template>
    </el-dialog>

    <!-- 新增/编辑自定义监控弹窗 -->
    <el-dialog v-model="showMonitorDialog" :title="monitorForm.id ? '编辑自定义图表' : '新增自定义图表'" width="600px" destroy-on-close>
      <el-form :model="monitorForm" :rules="monitorRules" ref="monitorFormRef" label-width="120px">
        <el-form-item label="图表标题" prop="title">
          <el-input v-model="monitorForm.title" placeholder="如：容器磁盘 I/O 速率" />
        </el-form-item>
        <el-form-item label="PromQL 模板" prop="promql_template">
          <el-input v-model="monitorForm.promql_template" type="textarea" :rows="3" placeholder="例如: rate(container_fs_io_time_seconds_total{pod='{{podName}}'}[5m])" />
          <div style="font-size: 12px; color: #999; margin-top: 5px;">
            可用插值变量: <code>{{podName}}</code>, <code>{{nodeName}}</code>
          </div>
        </el-form-item>
        <el-form-item label="图表类型" prop="chart_type">
          <el-radio-group v-model="monitorForm.chart_type">
            <el-radio label="line">折线图 (Line)</el-radio>
            <el-radio label="bar">柱状图 (Bar)</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="Y轴单位">
          <el-input v-model="monitorForm.unit_suffix" placeholder="如: %, MB, 秒 (可选)" />
        </el-form-item>
        <el-form-item label="主题颜色">
          <el-color-picker v-model="monitorForm.color_theme" show-alpha />
        </el-form-item>
      </el-form>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="showMonitorDialog = false">取消</el-button>
          <el-button type="primary" @click="handleSaveCustomMonitor" :loading="monitorSubmitting">保存</el-button>
        </div>
      </template>
    </el-dialog>

    <!-- 容器选择对话框 -->
    <el-dialog v-model="showContainerSelectDialog" title="选择容器" width="400px">
      <el-select 
        v-model="selectedContainerForTerminal" 
        placeholder="请选择要连接的容器"
        style="width: 100%;"
        size="large"
      >
        <el-option
          v-for="container in containerSelectList"
          :key="container"
          :label="container"
          :value="container"
        />
      </el-select>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="showContainerSelectDialog = false">取消</el-button>
          <el-button type="primary" @click="confirmContainerSelect" :disabled="!selectedContainerForTerminal">连接</el-button>
        </div>
      </template>
    </el-dialog>

    <!-- 容器终端 -->
    <PodTerminal
      v-model="showTerminalDialog"
      :namespace="terminalPod.namespace"
      :pod-name="terminalPod.podName"
      :container-name="terminalPod.containerName"
      :instance-id="getSelectedInstanceId()"
      @close="handleTerminalClose"
    />
  </div>
</template>

<script setup>
import { usePermissionStore } from '@/stores/permissionStore.js'
const permStore = usePermissionStore()

import { ref, onMounted, watch, nextTick } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, View, Delete, Refresh, Download, Monitor } from '@element-plus/icons-vue'
import { getPodList, getPodDetail, getPodEvents, getPodLogs, createPod, deletePod } from '@/api/k8s/pod'
import { getSelectedInstanceId } from '@/stores/instanceStore'
import { getNamespaceList } from '@/api/k8s/namespace'
import yaml from 'js-yaml'
import eventBus from '@/utils/eventBus'
import PodTerminal from '@/components/PodTerminal.vue'
import NodeMonitorChart from './components/NodeMonitorChart.vue'
import DynamicPromQLChart from './components/DynamicPromQLChart.vue'
import { getInstanceList } from '@/api/instance'
import { getCustomMonitors, createCustomMonitor, updateCustomMonitor, deleteCustomMonitor } from '@/api/monitor'

const loading = ref(false)
const submitting = ref(false)
const showCreateDialog = ref(false)
const showDetailDialog = ref(false)
const podList = ref([])
const namespaceList = ref([])
const selectedNamespace = ref('all')
const currentPod = ref(null)
const createFormRef = ref(null)
const activeTab = ref('form')
const yamlContent = ref('')

// 详情弹窗相关
const detailTab = ref('info')
const podEvents = ref([])
const podLogs = ref('')
const selectedContainer = ref('')
const loadingLogs = ref(false)
const followLogs = ref(false)
const showTimestamps = ref(false)
const logTailLines = ref(10)
const logContentRef = ref(null)
const activeContainers = ref([])
let logWebSocket = null

// 终端相关
const showTerminalDialog = ref(false)
const terminalPod = ref({
  namespace: '',
  podName: '',
  containerName: ''
})

// 容器选择对话框
const showContainerSelectDialog = ref(false)
const containerSelectPod = ref(null)
const containerSelectList = ref([])
const selectedContainerForTerminal = ref('')

const prometheusInstances = ref([])
const selectedPrometheusId = ref('')

// 自定义监控相关状态
const customMonitors = ref([])
const showMonitorDialog = ref(false)
const monitorSubmitting = ref(false)
const monitorFormRef = ref(null)
const monitorForm = ref({
  id: null,
  title: '',
  promql_template: '',
  chart_type: 'line',
  unit_suffix: '',
  color_theme: '#409EFF'
})
const monitorRules = {
  title: [{ required: true, message: '请输入图表标题', trigger: 'blur' }],
  promql_template: [{ required: true, message: '请输入 PromQL 查询模板', trigger: 'blur' }],
  chart_type: [{ required: true, message: '请选择图表类型', trigger: 'change' }]
}

const createForm = ref({
  namespace: 'default',
  podname: '',
  restartPolicy: 'Always',
  labels: [],
  annotations: [],
  nodeSelector: [],
  containers: [{
    name: '',
    image: '',
    imagePullPolicy: 'IfNotPresent',
    ports: [],
    env: [],
    resources: {
      requests: {
        cpu: '',
        memory: ''
      },
      limits: {
        cpu: '',
        memory: ''
      }
    }
  }]
})

const createRules = {
  namespace: [{ required: true, message: '请选择命名空间', trigger: 'change' }],
  podname: [{ required: true, message: '请输入 Pod 名称', trigger: 'blur' }]
}

// 容器管理方法
const addContainer = () => {
  createForm.value.containers.push({
    name: '',
    image: '',
    imagePullPolicy: 'IfNotPresent',
    ports: [],
    env: [],
    resources: {
      requests: {
        cpu: '',
        memory: ''
      },
      limits: {
        cpu: '',
        memory: ''
      }
    }
  })
}

const removeContainer = (index) => {
  createForm.value.containers.splice(index, 1)
}

// 端口管理方法
const addPort = (container) => {
  container.ports.push({
    name: '',
    containerPort: 80,
    protocol: 'TCP'
  })
}

const removePort = (container, index) => {
  container.ports.splice(index, 1)
}

// 环境变量管理方法
const addEnv = (container) => {
  container.env.push({
    name: '',
    value: ''
  })
}

const removeEnv = (container, index) => {
  container.env.splice(index, 1)
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

// 节点选择器管理方法
const addNodeSelector = () => {
  createForm.value.nodeSelector.push({
    key: '',
    value: ''
  })
}

const removeNodeSelector = (index) => {
  createForm.value.nodeSelector.splice(index, 1)
}

// YAML相关方法
const generateYAML = () => {
  // 过滤有效的容器
  const validContainers = createForm.value.containers.filter(container => container.name && container.image)
  
  if (validContainers.length === 0) {
    ElMessage.error('请至少配置一个有效的容器')
    return ''
  }
  
  // 手动构建YAML字符串，确保格式正确
  let yaml = `apiVersion: v1
kind: Pod
metadata:
  name: ${createForm.value.podname}
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
  restartPolicy: ${createForm.value.restartPolicy || 'Always'}
  containers:`
  
  // 添加容器
  validContainers.forEach(container => {
    yaml += `\n    - name: ${container.name}
      image: ${container.image}
      imagePullPolicy: ${container.imagePullPolicy || 'IfNotPresent'}`
    
    // 添加端口
    const validPorts = container.ports.filter(port => port.containerPort)
    if (validPorts.length > 0) {
      yaml += '\n      ports:'
      validPorts.forEach(port => {
        yaml += `\n        - containerPort: ${port.containerPort}`
        if (port.name) {
          yaml += `\n          name: ${port.name}`
        }
        yaml += `\n          protocol: ${port.protocol || 'TCP'}`
      })
    }
    
    // 添加环境变量
    const validEnv = container.env.filter(env => env.name)
    if (validEnv.length > 0) {
      yaml += '\n      env:'
      validEnv.forEach(env => {
        yaml += `\n        - name: ${env.name}
          value: "${env.value || ''}"`
      })
    }
    
    // 添加资源限制
    const hasRequests = container.resources.requests.cpu || container.resources.requests.memory
    const hasLimits = container.resources.limits.cpu || container.resources.limits.memory
    if (hasRequests || hasLimits) {
      yaml += '\n      resources:'
      if (hasRequests) {
        yaml += '\n        requests:'
        if (container.resources.requests.cpu) {
          yaml += `\n          cpu: ${container.resources.requests.cpu}`
        }
        if (container.resources.requests.memory) {
          yaml += `\n          memory: ${container.resources.requests.memory}`
        }
      }
      if (hasLimits) {
        yaml += '\n        limits:'
        if (container.resources.limits.cpu) {
          yaml += `\n          cpu: ${container.resources.limits.cpu}`
        }
        if (container.resources.limits.memory) {
          yaml += `\n          memory: ${container.resources.limits.memory}`
        }
      }
    }
  })
  
  // 添加节点选择器
  const validNodeSelectors = createForm.value.nodeSelector.filter(selector => selector.key && selector.value)
  if (validNodeSelectors.length > 0) {
    yaml += '\n  nodeSelector:'
    validNodeSelectors.forEach(selector => {
      yaml += `\n    ${selector.key}: ${selector.value}`
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
  a.download = `${createForm.value.podname || 'pod'}.yaml`
  document.body.appendChild(a)
  a.click()
  document.body.removeChild(a)
  URL.revokeObjectURL(url)
}

// 监听标签页切换
watch(activeTab, async (newTab) => {
  if (newTab === 'yaml') {
    // 先验证表单数据
    if (!createForm.value.podname) {
      ElMessage.warning('请先填写Pod名称')
      activeTab.value = 'form'
      return
    }
    
    // 验证容器配置
    const validContainers = createForm.value.containers.filter(container => container.name && container.image)
    if (validContainers.length === 0) {
      ElMessage.warning('请至少配置一个有效的容器（名称和镜像不能为空）')
      activeTab.value = 'form'
      return
    }
    
    // 从表单生成YAML
    yamlContent.value = generateYAML()
  }
})

const formatTimestamp = (timestamp) => {
  if (!timestamp) return '-'
  const date = new Date(timestamp * 1000)
  return date.toLocaleString('zh-CN')
}

const getStatusType = (status) => {
  switch (status) {
    case 'Running':
      return 'success'
    case 'Pending':
      return 'warning'
    case 'Failed':
    case 'Error':
      return 'danger'
    case 'Succeeded':
      return 'success'
    default:
      return 'info'
  }
}

const getEventType = (type) => {
  switch (type) {
    case 'Normal':
      return 'success'
    case 'Warning':
      return 'warning'
    default:
      return 'danger'
  }
}

const getEventTagType = (type) => {
  switch (type) {
    case 'Normal':
      return 'success'
    case 'Warning':
      return 'warning'
    default:
      return 'danger'
  }
}

const formatEventTime = (timestamp) => {
  if (!timestamp) return '-'
  return new Date(timestamp).toLocaleString('zh-CN')
}

// 获取Pod事件
const fetchPodEvents = async (namespace, podName, instanceId) => {
  try {
    const response = await getPodEvents(namespace, podName, instanceId)
    podEvents.value = response.data?.events || []
  } catch (error) {
    console.error('获取Pod事件失败:', error)
    podEvents.value = []
  }
}

// 获取Pod日志
const fetchPodLogs = async (namespace, podName, instanceId) => {
  if (!selectedContainer.value) return
  
  loadingLogs.value = true
  try {
    const options = {
      follow: followLogs.value,
      tailLines: logTailLines.value === 0 ? undefined : logTailLines.value,
      timestamps: showTimestamps.value
    }
    const response = await getPodLogs(namespace, podName, selectedContainer.value, instanceId, options)
    
    // 正确的路径是 response.data.logs.logs
    podLogs.value = response.data?.logs?.logs || response.data?.logs || ''
    
    // 如果是实时跟踪模式，滚动到底部
    if (followLogs.value && logContentRef.value) {
      nextTick(() => {
        const logElement = logContentRef.value
        if (logElement) {
          logElement.scrollTop = logElement.scrollHeight
        }
      })
    }
  } catch (error) {
    console.error('获取Pod日志失败:', error)
    podLogs.value = '获取日志失败: ' + (error.response?.data?.message || error.message)
  } finally {
    loadingLogs.value = false
  }
}

// 刷新日志
const refreshLogs = () => {
  podLogs.value = ''
  if (followLogs.value) {
    startLogStream()
  } else if (currentPod.value) {
    const instanceId = getSelectedInstanceId() || '1'
    fetchPodLogs(currentPod.value.namespace, currentPod.value.name, instanceId)
  }
}

// 下载日志
const downloadLogs = () => {
  if (!podLogs.value) return
  
  const blob = new Blob([podLogs.value], { type: 'text/plain' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `${currentPod.value.name}-${selectedContainer.value}-logs.txt`
  document.body.appendChild(a)
  a.click()
  document.body.removeChild(a)
  URL.revokeObjectURL(url)
}

// 监听容器切换
watch(selectedContainer, () => {
  if (currentPod.value && showDetailDialog.value) {
    refreshLogs()
  }
})

// 监听实时跟踪开关
watch(followLogs, (newVal) => {
  if (newVal && currentPod.value && showDetailDialog.value) {
    startLogStream()
  } else {
    stopLogStream()
  }
})

// 监听详情弹窗关闭
watch(showDetailDialog, (newVal) => {
  if (!newVal) {
    // 关闭弹窗时清理资源
    stopLogStream()
    followLogs.value = false
    podLogs.value = ''
    podEvents.value = []
  }
})

// 启动日志流
const startLogStream = () => {
  if (!currentPod.value || !selectedContainer.value) return
  
  stopLogStream() // 先停止现有的连接
  
  const instanceId = getSelectedInstanceId() || '1'
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  const tailParam = logTailLines.value === 0 ? 0 : logTailLines.value
  const wsUrl = `${protocol}//${window.location.hostname}:8081/ws/pod/${currentPod.value.name}/logs?namespace=${currentPod.value.namespace}&instance_id=${instanceId}&container=${selectedContainer.value}&follow=true&timestamps=${showTimestamps.value}&tail=${tailParam}`
  
  logWebSocket = new WebSocket(wsUrl)
  
  logWebSocket.onopen = () => {
    console.log('WebSocket日志连接已建立')
    podLogs.value = '' // 清空之前的日志
  }
  
  logWebSocket.onmessage = (event) => {
    // 调试：打印原始数据
    console.log('收到日志数据:', event.data)
    
    try {
      const data = JSON.parse(event.data)
      console.log('解析后的数据:', data)
      
      if (data.type === 'log' && data.content) {
        // 直接追加日志内容
        podLogs.value += data.content
      } else if (data.logs) {
        // 如果是 {logs: "..."} 格式
        podLogs.value += data.logs
      } else if (typeof data === 'string') {
        // 如果整个对象就是字符串
        podLogs.value += data
      }
      
      // 自动滚动到底部
      nextTick(() => {
        if (logContentRef.value) {
          logContentRef.value.scrollTop = logContentRef.value.scrollHeight
        }
      })
    } catch (e) {
      console.log('JSON解析失败，使用原始数据:', e)
      // 如果不是 JSON，直接追加原始内容
      podLogs.value += event.data
      nextTick(() => {
        if (logContentRef.value) {
          logContentRef.value.scrollTop = logContentRef.value.scrollHeight
        }
      })
    }
  }
  
  logWebSocket.onerror = (error) => {
    console.error('WebSocket错误:', error)
    ElMessage.error('日志连接失败')
    followLogs.value = false
  }
  
  logWebSocket.onclose = () => {
    console.log('WebSocket连接已关闭')
  }
}

// 停止日志流
const stopLogStream = () => {
  if (logWebSocket) {
    logWebSocket.close()
    logWebSocket = null
  }
}

// 清空日志
const clearLogs = () => {
  podLogs.value = ''
}

// 容器切换时重新加载日志
const handleContainerChange = () => {
  podLogs.value = ''
  if (followLogs.value) {
    startLogStream()
  }
}

// 日志行数切换时重新加载
const handleTailLinesChange = () => {
  podLogs.value = ''
  if (followLogs.value) {
    startLogStream()
  } else {
    refreshLogs()
  }
}

// 时间戳开关切换时重新连接
const handleTimestampToggle = () => {
  podLogs.value = ''
  if (followLogs.value) {
    // 实时跟踪模式：重新连接 WebSocket
    startLogStream()
  } else {
    // 非实时模式：重新获取日志
    refreshLogs()
  }
}


const fetchNamespaceList = async () => {
  try {
    const instanceId = getSelectedInstanceId() || '1'
    const response = await getNamespaceList(instanceId)
    namespaceList.value = response.data?.namespaceList || []
  } catch (error) {
    ElMessage.error('获取命名空间列表失败')
  }
}

const fetchPodList = async () => {
  loading.value = true
  try {
    const instanceId = getSelectedInstanceId() || '1'
    const response = await getPodList(selectedNamespace.value, instanceId)
    podList.value = response.data?.podList || []
  } catch (error) {
    ElMessage.error('获取 Pod 列表失败: ' + (error.response?.data?.message || error.message || '未知错误'))
  } finally {
    loading.value = false
  }
}

const handleNamespaceChange = () => {
  fetchPodList()
}

const handleCreate = async () => {
  submitting.value = true
  
  try {
    const instanceId = getSelectedInstanceId() || '1'
    let createData
    let createdNamespace = ''
    
    if (activeTab.value === 'yaml') {
      // 直接使用YAML创建
      try {
        const podSpec = yaml.load(yamlContent.value)
        createData = {
          namespace: podSpec.metadata.namespace,
          podname: podSpec.metadata.name,
          yaml: yamlContent.value
        }
        createdNamespace = podSpec.metadata.namespace
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
      
      // 验证容器配置
      if (!createForm.value.containers || createForm.value.containers.length === 0) {
        ElMessage.error('请至少配置一个容器')
        return
      }
      
      for (let i = 0; i < createForm.value.containers.length; i++) {
        const container = createForm.value.containers[i]
        if (!container.name || !container.image) {
          ElMessage.error(`容器 ${i + 1} 的名称和镜像不能为空`)
          return
        }
      }
      
      createData = createForm.value
      createdNamespace = createForm.value.namespace
    }
    
    await createPod(createData, instanceId)
    ElMessage.success('Pod 创建成功')
    showCreateDialog.value = false
    resetCreateForm()
    
    // 如果创建的Pod在当前选中的命名空间中，或者当前选择的是"所有"命名空间，则刷新列表
    if (selectedNamespace.value === 'all' || selectedNamespace.value === createdNamespace) {
      await fetchPodList()
    }
    
    // 发送Pod创建事件，通知其他组件刷新数据
    eventBus.emit('pod:created', { namespace: createdNamespace })
  } catch (error) {
    ElMessage.error('创建 Pod 失败: ' + (error.response?.data?.message || error.message || '未知错误'))
  } finally {
    submitting.value = false
  }
}

const resetCreateForm = () => {
  createForm.value = {
    namespace: 'default',
    podname: '',
    restartPolicy: 'Always',
    labels: [],
    annotations: [],
    nodeSelector: [],
    containers: [{
      name: '',
      image: '',
      imagePullPolicy: 'IfNotPresent',
      ports: [],
      env: [],
      resources: {
        requests: {
          cpu: '',
          memory: ''
        },
        limits: {
          cpu: '',
          memory: ''
        }
      }
    }]
  }
  yamlContent.value = ''
  activeTab.value = 'form'
}

const handleViewDetail = async (pod) => {
  try {
    const instanceId = getSelectedInstanceId() || '1'
    // 使用Pod的实际命名空间，而不是当前选中的命名空间
    const namespace = pod.namespace || 'default'
    const response = await getPodDetail(namespace, pod.name, instanceId)
    currentPod.value = response.data?.podDetail || pod
    
    // 设置默认选中的容器
    if (currentPod.value.containers && currentPod.value.containers.length > 0) {
      selectedContainer.value = currentPod.value.containers[0].name
    }
    
    // 获取Pod事件
    await fetchPodEvents(namespace, pod.name, instanceId)
    
    // 获取Pod日志
    await fetchPodLogs(namespace, pod.name, instanceId)
    
    // 获取 Prometheus 实例提供下拉选择
    if (prometheusInstances.value.length === 0) {
      const pRes = await getInstanceList({ page: 1, page_size: 100, type_name: 'prometheus' })
      if (pRes.data?.list?.data) {
        prometheusInstances.value = pRes.data.list.data
        if (prometheusInstances.value.length > 0) {
          selectedPrometheusId.value = prometheusInstances.value[0].id
        }
      }
    }

    // 获取当前用户的自定义监控配置 (Pod 维度)
    await fetchCustomMonitors()

    showDetailDialog.value = true
    detailTab.value = 'info'
  } catch (error) {
    const errorMsg = error.response?.data?.message || error.message || '获取 Pod 详情失败'
    ElMessage.error(errorMsg)
  }
}

// 打开容器终端
const handleOpenTerminal = async (pod) => {
  try {
    // 先获取Pod详情以获得容器信息
    const instanceId = getSelectedInstanceId() || '1'
    const namespace = pod.namespace || 'default'
    const res = await getPodDetail(namespace, pod.name, instanceId)
    
    if (res.status !== 200 || !res.data.podDetail) {
      ElMessage.error('获取 Pod 详情失败')
      return
    }
    
    const podDetail = res.data.podDetail
    const containers = podDetail.containers || []
    
    if (containers.length === 0) {
      ElMessage.error('Pod 没有可用的容器')
      return
    }
    
    // 如果Pod有多个容器，显示选择对话框
    if (containers.length > 1) {
      containerSelectPod.value = pod
      containerSelectList.value = containers.map(c => c.name)
      selectedContainerForTerminal.value = containers[0].name
      showContainerSelectDialog.value = true
    } else {
      // 单容器直接打开
      openTerminal(pod, containers[0].name)
    }
  } catch (error) {
    ElMessage.error('打开终端失败: ' + (error.message || '未知错误'))
  }
}

// 确认容器选择
const confirmContainerSelect = () => {
  if (containerSelectPod.value && selectedContainerForTerminal.value) {
    showContainerSelectDialog.value = false
    openTerminal(containerSelectPod.value, selectedContainerForTerminal.value)
  }
}

const openTerminal = (pod, containerName) => {
  terminalPod.value = {
    namespace: pod.namespace,
    podName: pod.name,
    containerName: containerName
  }
  showTerminalDialog.value = true
}

const handleTerminalClose = () => {
  showTerminalDialog.value = false
}

// ============== 自定义监控模块 ==============
const fetchCustomMonitors = async () => {
  try {
    const res = await getCustomMonitors('pod')
    if (res && res.data && res.data.list) {
      customMonitors.value = res.data.list
    } else {
      customMonitors.value = []
    }
  } catch (error) {
    console.error('Failed to fetch custom monitors:', error)
  }
}

const handleOpenCustomMonitorDialog = (monitorId = null) => {
  if (monitorFormRef.value) {
    monitorFormRef.value.resetFields()
  }
  if (monitorId) {
    const monitor = customMonitors.value.find(m => m.id === monitorId)
    if (monitor) {
      monitorForm.value = { ...monitor }
    }
  } else {
    monitorForm.value = {
      id: null,
      title: '',
      promql_template: '',
      chart_type: 'line',
      unit_suffix: '',
      color_theme: '#409EFF'
    }
  }
  showMonitorDialog.value = true
}

const handleSaveCustomMonitor = async () => {
  if (!monitorFormRef.value) return
  const valid = await monitorFormRef.value.validate().catch(() => false)
  if (!valid) return

  monitorSubmitting.value = true
  try {
    const payload = {
      target_type: 'pod',
      title: monitorForm.value.title,
      promql_template: monitorForm.value.promql_template,
      chart_type: monitorForm.value.chart_type,
      unit_suffix: monitorForm.value.unit_suffix,
      color_theme: monitorForm.value.color_theme
    }

    if (monitorForm.value.id) {
      await updateCustomMonitor(monitorForm.value.id, payload)
      ElMessage.success('更新自定义图表成功')
    } else {
      await createCustomMonitor(payload)
      ElMessage.success('创建自定义图表成功')
    }
    showMonitorDialog.value = false
    await fetchCustomMonitors()
  } catch (error) {
    ElMessage.error(monitorForm.value.id ? '更新图表失败' : '创建图表失败')
  } finally {
    monitorSubmitting.value = false
  }
}

const handleDeleteCustomMonitor = async (id) => {
  try {
    await ElMessageBox.confirm('确定要删除这个自定义监控图表吗？', '提示', { type: 'warning' })
    await deleteCustomMonitor(id)
    ElMessage.success('图表已删除')
    await fetchCustomMonitors()
  } catch (err) {
    if (err !== 'cancel') {
      ElMessage.error('删除图表失败')
    }
  }
}

const handleDelete = async (pod) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除 Pod "${pod.name}" 吗？`,
      '确认删除',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    const instanceId = getSelectedInstanceId() || '1'
    // 使用Pod的实际命名空间，而不是当前选中的命名空间
    const namespace = pod.namespace || 'default'
    await deletePod(namespace, pod.name, instanceId)
    ElMessage.success('Pod 删除成功')
    
    // 如果删除的Pod在当前选中的命名空间中，或者当前选择的是"所有"命名空间，则刷新列表
    if (selectedNamespace.value === 'all' || selectedNamespace.value === namespace) {
      await fetchPodList()
    }
    
    // 发送Pod删除事件，通知其他组件刷新数据
    eventBus.emit('pod:deleted', { namespace })
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除 Pod 失败')
    }
  }
}

onMounted(() => {
  fetchNamespaceList()
  fetchPodList()
})
</script>

<style scoped>
.pod-management {
  padding: 20px;
}

.page-header-card {
  margin-bottom: 20px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.page-header h2 {
  margin: 0 0 8px 0;
  font-size: 24px;
  font-weight: 600;
}

.page-header p {
  margin: 0;
  color: var(--el-text-color-secondary);
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 15px;
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

/* 容器配置样式 */
.container-config {
  margin-bottom: 20px;
}

.container-card {
  margin-bottom: 15px;
}

.container-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

/* 端口配置样式 */
.port-config {
  margin-bottom: 10px;
  padding: 10px;
  background-color: #f5f7fa;
  border-radius: 4px;
}

/* 环境变量配置样式 */
.env-config {
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
  margin-bottom: 10px;
}

.selector-config {
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

/* 按钮组样式 */
:deep(.el-button-group) {
  .el-button {
    padding: 5px 10px;
  }
}

/* 响应式调整 */
@media (max-width: 768px) {
  :deep(.el-col) {
    margin-bottom: 10px;
  }
}

/* 详情弹窗样式 */
:deep(.el-descriptions__label) {
  font-weight: 600;
  color: var(--el-text-color-primary);
}

:deep(.el-descriptions__content) {
  word-break: break-all;
}

.annotation-value {
  word-break: break-all;
  white-space: pre-wrap;
}

.no-data {
  text-align: center;
  color: var(--el-text-color-secondary);
  padding: 20px 0;
}

/* 事件时间线样式 */
:deep(.el-timeline-item__timestamp) {
  font-size: 12px;
  color: var(--el-text-color-secondary);
}

.event-item {
  padding: 12px;
  background-color: var(--el-fill-color-light);
  border-radius: 6px;
  border-left: 3px solid var(--el-color-primary);
}

.event-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}

.event-reason {
  font-weight: 600;
  color: var(--el-text-color-primary);
  font-size: 14px;
}

.event-message {
  color: var(--el-text-color-regular);
  font-size: 13px;
  line-height: 1.5;
  margin-bottom: 6px;
}

.event-source {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  display: flex;
  align-items: center;
  gap: 4px;
}

/* 日志容器样式 */
.log-container {
  height: 500px;
  display: flex;
  flex-direction: column;
}

.log-toolbar {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 10px;
  padding: 12px;
  background-color: #f5f7fa;
  border-radius: 4px;
  flex-wrap: wrap;
}

.log-content-wrapper {
  flex: 1;
  overflow: hidden;
  background-color: #0d1117;
  border-radius: 6px;
  border: 1px solid #30363d;
}

.log-content {
  height: 100%;
  overflow: auto;
  padding: 10px 14px;
}

.kubectl-style-logs {
  margin: 0;
  font-family: 'Consolas', 'Monaco', 'Courier New', 'SFMono-Regular', monospace;
  font-size: 13px;
  line-height: 1.6;
  color: #c9d1d9;
  white-space: pre;
  overflow-wrap: normal;
  word-break: normal;
}

.no-logs {
  text-align: center;
  color: #8b949e;
  padding: 40px 20px;
  font-size: 14px;
}

/* 折叠面板样式 */
:deep(.el-collapse-item__header) {
  font-weight: 600;
}

:deep(.el-collapse-item__content) {
  padding: 15px;
}
</style>
