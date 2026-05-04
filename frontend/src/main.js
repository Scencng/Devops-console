import { createApp } from 'vue'
import ElementPlus from 'element-plus'
import zhCn from 'element-plus/es/locale/lang/zh-cn'
import 'element-plus/dist/index.css'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'
import { createPinia } from 'pinia'
import './style.css'
import './styles/global.css'
import './styles/enhancements.css'
import './styles/page-layout.css'
import './style/autoops.css' // AutoOps 娓呮柊涓婚
import './mysql/styles/global.css'
import App from './App.vue'
import router from './router'

const app = createApp(App)
const pinia = createPinia()

// 娉ㄥ唽Element Plus鍥炬爣
for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
  app.component(key, component)
}

app.use(pinia)
app.use(ElementPlus, {
  locale: zhCn
})
app.use(router)
app.mount('#app')

