import ElementPlus from 'element-plus'
import { createApp } from 'vue'

import App from './App.vue'
import { loadRuntimeConfig } from './config/runtime'
import router from './router'
import pinia from './stores'
import { useLocaleStore } from './stores/locale'
import './styles/global.css'
import 'element-plus/dist/index.css'

async function bootstrap() {
  await loadRuntimeConfig()

  const app = createApp(App)
  app.use(pinia)
  useLocaleStore(pinia).setLocale(useLocaleStore(pinia).locale)
  app.use(router)
  app.use(ElementPlus)
  app.mount('#app')
}

void bootstrap()
