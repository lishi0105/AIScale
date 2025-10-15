import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { fileURLToPath, URL } from 'node:url'


export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
    },
  },
  server: {
    host: "0.0.0.0", //代理人地址
    cors: true, // 默认启用并允许任何源
    open: true, // 在服务器启动时自动在浏览器中打开应用程序
    //反向代理配置，注意rewrite写法，开始没看文档在这里踩了坑
    port:7280,
    proxy: {
      // 按需修改，转发到你的后端（比如 http://localhost:8080）
      '/accounts': {
        target: 'http://localhost:8080',
        changeOrigin: true
      }
    }
  }
})