# DevOps Console 前端

基于 Vue 3 + Element Plus 的现代化运维管理平台前端界面。

## 技术栈

- **Vue 3.5.22**：渐进式 JavaScript 框架
- **Element Plus 2.4.4**：企业级 UI 组件库
- **Vite 7.1.7**：下一代前端构建工具
- **Vue Router 4.2.5**：Vue.js 官方路由
- **Axios 1.6.2**：基于 Promise 的 HTTP 客户端
- **ECharts 5.4.3**：数据可视化图表库
- **@element-plus/icons-vue 2.3.1**：Element Plus 图标库

## 功能特性

- 🎨 **现代化界面**：基于 Element Plus 的响应式设计
- 📊 **数据可视化**：ECharts 图表展示系统监控数据
- 🔄 **实时更新**：WebSocket 实时数据推送
- 📱 **响应式布局**：支持桌面端和移动端
- 🌐 **多语言支持**：中英文界面切换
- 🎯 **路由管理**：基于 Vue Router 的单页应用

## 快速开始

### 环境要求

- Node.js 22.x 或更高版本
- npm 11.x 或更高版本

### 安装依赖

```bash
npm install
```

### 开发环境

```bash
# 启动开发服务器
npm run dev

# 服务器将在 http://localhost:5173 启动
```

### 构建生产版本

```bash
# 构建生产版本
npm run build

# 预览生产版本
npm run preview
```

## 项目结构

```
src/
├── api/                    # API 接口封装
│   ├── index.js           # 通用 API 配置
│   ├── instance.js        # 实例管理 API
│   ├── es/                # Elasticsearch 相关 API
│   └── k8s/               # Kubernetes 相关 API
├── assets/                # 静态资源
├── components/            # 公共组件
├── layouts/               # 布局组件
├── router/                # 路由配置
├── stores/                # 状态管理
├── styles/                # 样式文件
├── utils/                 # 工具函数
└── views/                 # 页面组件
    ├── Dashboard.vue      # 仪表板
    ├── es/                # ES 相关页面
    └── k8s/               # K8s 相关页面
```

## 开发指南

### 添加新页面

1. 在 `src/views/` 下创建页面组件
2. 在 `src/api/` 中添加对应的 API 调用
3. 在 `src/router/index.js` 中注册路由
4. 在菜单中添加导航项

### API 接口封装

```javascript
// src/api/example.js
import request from './index'

export const getExampleData = (params) => {
  return request({
    url: '/api/example',
    method: 'get',
    params
  })
}
```

### 组件开发

```vue
<template>
  <div class="example-component">
    <el-card>
      <template #header>
        <span>组件标题</span>
      </template>
      <!-- 组件内容 -->
    </el-card>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'

// 响应式数据
const data = ref([])

// 生命周期
onMounted(() => {
  // 初始化逻辑
})
</script>

<style scoped>
.example-component {
  /* 组件样式 */
}
</style>
```

### 状态管理

使用 Vue 3 的 Composition API 进行状态管理：

```javascript
// src/stores/exampleStore.js
import { ref, computed } from 'vue'

const state = ref({
  data: [],
  loading: false
})

export const useExampleStore = () => {
  const setData = (newData) => {
    state.value.data = newData
  }

  const isLoading = computed(() => state.value.loading)

  return {
    state,
    setData,
    isLoading
  }
}
```

## 部署配置

### 环境变量

创建 `.env.development` 和 `.env.production` 文件：

```bash
# .env.development
VITE_API_BASE_URL=http://localhost:8081
VITE_APP_TITLE=DevOps Console (开发环境)

# .env.production
VITE_API_BASE_URL=/api
VITE_APP_TITLE=DevOps Console
```

### Nginx 配置

```nginx
server {
    listen 80;
    server_name your-domain.com;
    root /path/to/dist;
    index index.html;

    location / {
        try_files $uri $uri/ /index.html;
    }

    location /api {
        proxy_pass http://backend-service:8081;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

## 常见问题

### 1. 开发环境跨域问题

在 `vite.config.js` 中配置代理：

```javascript
export default defineConfig({
  server: {
    proxy: {
      '/api': {
        target: 'http://localhost:8081',
        changeOrigin: true
      }
    }
  }
})
```

### 2. Element Plus 按需引入

```javascript
// main.js
import { createApp } from 'vue'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'

const app = createApp(App)
app.use(ElementPlus)
```

### 3. 图表不显示

确保 ECharts 容器有明确的宽高：

```css
.chart-container {
  width: 100%;
  height: 400px;
}
```

## 性能优化

### 1. 路由懒加载

```javascript
const routes = [
  {
    path: '/dashboard',
    component: () => import('@/views/Dashboard.vue')
  }
]
```

### 2. 组件懒加载

```vue
<template>
  <el-skeleton v-if="loading" />
  <AsyncComponent v-else />
</template>

<script setup>
import { defineAsyncComponent, ref } from 'vue'

const AsyncComponent = defineAsyncComponent(() => 
  import('./AsyncComponent.vue')
)
</script>
```

### 3. 图片优化

```javascript
// 使用 WebP 格式
import imgWebp from '@/assets/image.webp'
import imgFallback from '@/assets/image.jpg'
```

## 贡献指南

1. Fork 项目
2. 创建特性分支
3. 提交更改
4. 推送到分支
5. 创建 Pull Request

## 许可证

MIT License

## 联系方式

如有问题或建议，请提交 Issue 或联系开发团队。
