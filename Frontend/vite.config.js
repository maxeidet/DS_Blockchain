import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    port: 3000,
    // ── Dev proxy ──────────────────────────────────────────────────────────
    // Forwards  /api/node/<port>/<path>  →  http://localhost:<port>/<path>
    // This avoids CORS issues when your blockchain nodes don't set CORS headers.
    //
    // Example: fetch('/api/node/5000/chain') → http://localhost:5000/chain
    //
    // To use it, prefix node URLs with /api/node/<port> in blockchain.js,
    // or keep calling nodes directly (they must then allow CORS themselves).
    proxy: {
      '/api/node': {
        target: 'http://localhost',          // base – overridden per-request
        changeOrigin: true,
        rewrite: (path) => {
          // /api/node/5001/chain  →  /chain  (target host is set dynamically below)
          return path.replace(/^\/api\/node\/\d+/, '')
        },
        configure: (proxy) => {
          proxy.on('proxyReq', (proxyReq, req) => {
            const match = req.url.match(/^\/api\/node\/(\d+)/)
            if (match) {
              proxyReq.host = `localhost:${match[1]}`
              proxyReq.setHeader('host', `localhost:${match[1]}`)
            }
          })
        },
      },
    },
  },
})

