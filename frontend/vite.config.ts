import { defineConfig, loadEnv } from 'vite'
import react from '@vitejs/plugin-react'

// Dev-сервер отдаёт клиент и проксирует API на бэкенд, поэтому в браузере
// оба адреса выглядят одним origin: cookie профиля работает без CORS.
export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '')
  const backend = env.VITE_DEV_BACKEND || 'http://127.0.0.1:8080'
  const proxy = { target: backend, changeOrigin: true }

  return {
    plugins: [react()],
    server: {
      host: '127.0.0.1',
      port: 5173,
      proxy: {
        '/api': proxy,
        '/healthz': proxy,
        '/readyz': proxy,
      },
    },
  }
})
