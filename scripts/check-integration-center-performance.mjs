#!/usr/bin/env node

const apiBase = process.env.API_BASE_URL || `http://127.0.0.1:${process.env.BASICP_API_PORT || '8081'}`
const account = process.env.BASICP_E2E_ACCOUNT || 'E10001'
const password = process.env.BASICP_E2E_PASSWORD || '112233'
const concurrency = Number(process.env.INTEGRATION_PERF_CONCURRENCY || 6)
const rounds = Number(process.env.INTEGRATION_PERF_ROUNDS || 8)
const p95LimitMs = Number(process.env.INTEGRATION_PERF_P95_MS || 800)

const targets = [
  { name: 'overview', path: '/api/integration-center/overview' },
  { name: 'tenant-connections', path: '/api/integration-center/tenant-connections?skip=0&limit=20&sort_by=updated_at&sort_order=desc' },
  { name: 'sync-monitor', path: '/api/integration-center/sync-monitor?skip=0&limit=20&sort_by=started_at&sort_order=desc' },
  { name: 'logs', path: '/api/integration-center/logs?skip=0&limit=20&sort_by=called_at&sort_order=desc' },
]

async function requestJSON(path, init = {}) {
  const res = await fetch(`${apiBase}${path}`, {
    ...init,
    headers: { 'content-type': 'application/json', ...(init.headers || {}) },
  })
  const body = await res.json().catch(() => ({}))
  if (!res.ok || body.code !== 0) {
    throw new Error(`${path} failed status=${res.status} code=${body.code} message=${body.message || ''}`)
  }
  return body.data
}

function percentile(values, ratio) {
  const sorted = [...values].sort((a, b) => a - b)
  const index = Math.min(sorted.length - 1, Math.ceil(sorted.length * ratio) - 1)
  return sorted[index] || 0
}

async function timed(target, token) {
  const start = performance.now()
  await requestJSON(target.path, { headers: { Authorization: `Bearer ${token}` } })
  return performance.now() - start
}

async function runTarget(target, token) {
  const samples = []
  await timed(target, token)
  for (let round = 0; round < rounds; round += 1) {
    const batch = await Promise.all(Array.from({ length: concurrency }, () => timed(target, token)))
    samples.push(...batch)
  }
  const p95 = percentile(samples, 0.95)
  const avg = samples.reduce((sum, value) => sum + value, 0) / Math.max(samples.length, 1)
  return { name: target.name, samples: samples.length, avg, p95, pass: p95 <= p95LimitMs }
}

async function main() {
  const login = await requestJSON('/api/auth/login', {
    method: 'POST',
    body: JSON.stringify({ account, password }),
  })
  const token = login.token
  if (!token) throw new Error('login did not return token')

  const results = []
  for (const target of targets) {
    results.push(await runTarget(target, token))
  }

  for (const result of results) {
    const status = result.pass ? 'PASS' : 'FAIL'
    console.log(`check=${result.name} samples=${result.samples} avg_ms=${result.avg.toFixed(1)} p95_ms=${result.p95.toFixed(1)} threshold_ms=${p95LimitMs} status=${status}`)
  }
  if (results.some(result => !result.pass)) {
    console.error('integration_center_performance_check=FAIL')
    process.exit(1)
  }
  console.log('integration_center_performance_check=PASS')
}

main().catch((error) => {
  console.error(error.message || error)
  process.exit(1)
})
