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
    proxy: {
      '/v1/merchant': {
        target: 'http://127.0.0.1:8088',
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
      '/v1/callback': {
        target: 'http://127.0.0.1:8090',
        changeOrigin: true,
      },
    },
  },
})
