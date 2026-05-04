import path from 'node:path'

import vue from '@vitejs/plugin-vue'
import { defineConfig, loadEnv } from 'vite'

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '')
  const devApiTarget = env.VITE_DEV_API_TARGET || 'http://127.0.0.1:8080'
  const buildId = new Date().toISOString().replace(/[-:.TZ]/g, '')

  return {
    plugins: [vue()],
    define: {
      __APP_BUILD_ID__: JSON.stringify(buildId)
    },
    resolve: {
      alias: {
        '@': path.resolve(__dirname, './src')
      }
    },
    build: {
      sourcemap: false,
      cssCodeSplit: true,
      assetsInlineLimit: 4096,
      chunkSizeWarningLimit: 1200,
      rollupOptions: {
        output: {
          manualChunks: {
            'vendor-vue': ['vue', 'vue-router', 'pinia'],
            'vendor-ui': ['element-plus', '@element-plus/icons-vue'],
            'vendor-net': ['axios'],
            'vendor-xlsx': ['xlsx']
          }
        }
      }
    },
    server: {
      host: '0.0.0.0',
      port: 5173,
      open: false,
      proxy: {
        '/api': {
          target: devApiTarget,
          changeOrigin: true
        }
      }
    }
  }
})
