import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
  plugins: [vue()],
  server: {
    proxy: {
      '/v1/terminal': {
        target: 'http://127.0.0.1:8092',
        changeOrigin: true,
      },
      '/health': {
        target: 'http://127.0.0.1:8092',
        changeOrigin: true,
      },
    },
  },
})
