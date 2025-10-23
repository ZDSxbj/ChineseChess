import { fileURLToPath, URL } from 'node:url'

import vue from '@vitejs/plugin-vue'
import vueJsx from '@vitejs/plugin-vue-jsx'
import UnoCSS from 'unocss/vite'
import { defineConfig } from 'vite'
// import { analyzer } from 'vite-bundle-analyzer'

// https://vite.dev/config/
export default defineConfig({
  base: '/chess',
  plugins: [
    vue(),
    vueJsx(),
    UnoCSS(),
    // analyzer({
    //   analyzerMode: 'server', // 默认值，启动本地服务器
    //   openAnalyzer: true, // 自动打开浏览器
    //   analyzerPort: 8888, // 分析服务器端口
    // }),
  ],
  build: {
    rollupOptions: {
      output: {
        manualChunks(id) {
          if (!id.includes('node_modules')) {
            return
          }
          id = id.split('node_modules/')[1]

          if (id.includes('vue')) {
            return 'vue'
          }

          if (id.includes('axios')) {
            return 'axios'
          }

          if (id.includes('crypto-js')) {
            return 'cryptoJs'
          }

          return 'vendor'
        },
      },
    },
  },
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
    },
  },
})
