import type { ServerResponse } from 'node:http'
import { existsSync, readFileSync } from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

import vue from '@vitejs/plugin-vue'
import { defineConfig, type ProxyOptions } from 'vite'

const root = fileURLToPath(new URL('.', import.meta.url))
const repoRoot = path.join(root, '..')

/** 与项目根 .env 中 BASICP_* 一致，避免改端口后 Vite 仍代理到错误地址 */
function readRepoEnvInt(key: string, fallback: number): number {
  const ev = process.env[key]
  if (ev && /^\d+$/.test(ev)) return parseInt(ev, 10)

  const envFile = path.join(repoRoot, '.env')
  if (existsSync(envFile)) {
    const text = readFileSync(envFile, 'utf8')
    const m = text.match(new RegExp(`^\\s*${key}\\s*=\\s*(\\d+)\\s*$`, 'm'))
    if (m) return parseInt(m[1], 10)
  }
  return fallback
}

const apiPort = readRepoEnvInt('BASICP_API_PORT', 8081)
const vitePort = readRepoEnvInt('BASICP_VITE_PORT', 5173)
const proxyTarget = `http://127.0.0.1:${apiPort}`

/** 代理连不上后端时在运行 npm run dev 的终端打日志 */
function proxyToBackend(prefix: string, target: string): ProxyOptions {
  const proxyMs = 180_000
  return {
    target,
    changeOrigin: true,
    timeout: proxyMs,
    proxyTimeout: proxyMs,
    configure(proxy) {
      proxy.on('error', (err: NodeJS.ErrnoException, _req, res) => {
        console.error(`[vite-proxy] ${prefix} -> ${target}: ${err.message}${err.code ? ` (${err.code})` : ''}`)
        const r = res as ServerResponse | undefined
        if (r && 'writeHead' in r && !r.headersSent) {
          r.writeHead(502, { 'Content-Type': 'text/plain; charset=utf-8' })
          r.end(
            `Vite 代理无法连接 ${target}。请确认 Docker API 已起: 项目根执行 docker compose up -d，或 bash scripts/check-services.sh`,
          )
        }
      })
      proxy.on('proxyReq', (proxyReq) => {
        proxyReq.setTimeout(proxyMs)
      })
    },
  }
}

const apiProxy = {
  '/api': proxyToBackend('/api', proxyTarget),
  '/health': proxyToBackend('/health', proxyTarget),
  '/openapi.json': proxyToBackend('/openapi.json', proxyTarget),
} as const

// https://vite.dev/config/
export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      '@': path.join(root, 'src'),
    },
  },
  css: {
    transformer: 'postcss',
    devSourcemap: true,
  },
  build: {
    minify: false,
  },
  server: {
    port: vitePort,
    strictPort: false,
    proxy: { ...apiProxy },
  },
  preview: {
    proxy: { ...apiProxy },
  },
})
