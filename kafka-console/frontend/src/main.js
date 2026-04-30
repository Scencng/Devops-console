import { createApp } from 'vue'
import { ElLoadingDirective } from 'element-plus'
import 'element-plus/es/components/loading/style/css'
import 'element-plus/es/components/message/style/css'
import 'element-plus/es/components/message-box/style/css'
import { createPinia } from 'pinia'
import './style.css'
import './styles/global.css'
import './styles/enhancements.css'
import './styles/page-layout.css'
import './style/autoops.css' // AutoOps 清新主题
import App from './App.vue'
import router from './router'

const app = createApp(App)
const pinia = createPinia()

app.use(pinia)
app.directive('loading', ElLoadingDirective)
app.use(router)
app.mount('#app')
