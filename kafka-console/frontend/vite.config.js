import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { resolve } from 'path'
import Components from 'unplugin-vue-components/vite'
import { ElementPlusResolver } from 'unplugin-vue-components/resolvers'

// https://vite.dev/config/
export default defineConfig({
  plugins: [
    vue(),
    Components({
      dts: false,
      resolvers: [
        ElementPlusResolver({
          importStyle: 'css',
        }),
      ],
    }),
  ],
  build: {
    rollupOptions: {
      output: {
        manualChunks(id) {
          if (!id.includes('node_modules')) return

          const isElementPlusComponent = (patterns) =>
            patterns.some((pattern) => id.includes(`/element-plus/es/components/${pattern}/`) || id.includes(`\\element-plus\\es\\components\\${pattern}\\`))

          if (id.includes('monaco-editor')) return 'vendor-monaco'
          if (id.includes('xterm')) return 'vendor-terminal'
          if (id.includes('crypto-js')) return 'vendor-crypto'
          if (id.includes('@element-plus/icons-vue')) return 'vendor-icons'
          if (id.includes('echarts') || id.includes('vue-echarts')) return 'vendor-charts'
          if (isElementPlusComponent([
            'button',
            'checkbox',
            'date-picker',
            'form',
            'form-item',
            'icon',
            'input',
            'input-number',
            'option',
            'option-group',
            'select',
            'switch',
            'time-picker',
          ])) {
            return 'vendor-ui-forms'
          }
          if (isElementPlusComponent([
            'card',
            'descriptions',
            'dialog',
            'drawer',
            'empty',
            'loading',
            'menu',
            'menu-item',
            'message',
            'message-box',
            'overlay',
            'pagination',
            'scrollbar',
            'skeleton',
            'sub-menu',
            'table',
            'table-column',
            'tag',
          ])) {
            return 'vendor-ui-data'
          }
          if (id.includes('element-plus')) return 'vendor-ui-core'
          if (id.includes('vue-router') || id.includes('pinia') || id.includes('/vue/')) return 'vendor-framework'
        },
      },
    },
  },
  resolve: {
    alias: {
      '@': resolve(__dirname, 'src')
    }
  },
  server: {
    proxy: {
      '/api': {
        // target: 'http://172.20.0.3:8081',
        target: 'http://127.0.0.1:8081',
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api/, '/api')
      },
      '/ws': {
        // target: 'http://172.20.0.3:8081',
        target: 'http://svc-backend:8081',
        changeOrigin: true,
        ws: true
      }
    }
  }
})
