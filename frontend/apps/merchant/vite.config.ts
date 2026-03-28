import { fileURLToPath, URL } from 'node:url'
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
    },
  },
  server: {
    // 更长前缀在前，避免被 /v1/merchant、/v1/callback 误匹配到错误端口
    proxy: {
      '/v1/merchant/balance/query': {
        target: 'http://127.0.0.1:8090',
        changeOrigin: true,
      },
      '/v1/callback': {
        target: 'http://127.0.0.1:8090',
        changeOrigin: true,
      },
      '/v1/payin': {
        target: 'http://127.0.0.1:8090',
        changeOrigin: true,
      },
      '/v1/payout': {
        target: 'http://127.0.0.1:8090',
        changeOrigin: true,
      },
      '/v1/merchant': {
        target: 'http://127.0.0.1:8088',
        changeOrigin: true,
      },
    },
  },
})
