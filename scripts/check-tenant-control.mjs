#!/usr/bin/env node

const base = process.env.API_BASE || 'http://127.0.0.1:8081'

const tenants = [
  {
    code: 'taohuadao',
    account: '100001',
    password: '112233',
    mustAllow: ['GET /api/users', 'GET /api/organizations/tree', 'GET /api/positions', 'GET /api/logs/login'],
    mustDeny: ['GET /api/logs/audit', 'GET /api/monitor/health-detail', 'GET /api/tenants', 'GET /api/plans', 'GET /api/batch/users/export'],
    actionChecks: [
      { name: 'POST /api/users', method: 'POST', path: '/api/users', body: { employee_no: '', name: '' }, expectNot403: true },
      { name: 'DELETE /api/users/:id', method: 'DELETE', path: '/api/users/999999', expect403: true },
      { name: 'POST /api/params', method: 'POST', path: '/api/params', body: {}, expect403: true },
    ],
  },
  {
    code: 'jianghu',
    account: '00001',
    password: '112233',
    mustAllow: ['GET /api/users', 'GET /api/organizations/tree', 'GET /api/positions', 'GET /api/logs/login'],
    mustDeny: ['GET /api/logs/audit', 'GET /api/monitor/health-detail', 'GET /api/tenants', 'GET /api/plans', 'GET /api/batch/users/export'],
    actionChecks: [
      { name: 'POST /api/users', method: 'POST', path: '/api/users', body: { employee_no: '', name: '' }, expect403: true },
      { name: 'POST /api/params', method: 'POST', path: '/api/params', body: {}, expectNot403: true },
    ],
  },
  {
    code: 'platform',
    account: 'E10001',
    password: '112233',
    mustAllow: ['GET /api/users', 'GET /api/logs/audit', 'GET /api/logs/login', 'GET /api/monitor/health-detail', 'GET /api/tenants', 'GET /api/plans'],
    mustDeny: [],
    actionChecks: [],
  },
]

async function request(path, { method = 'GET', token, body } = {}) {
  const response = await fetch(base + path, {
    method,
    headers: {
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
      ...(body ? { 'Content-Type': 'application/json' } : {}),
    },
    body: body ? JSON.stringify(body) : undefined,
  })
  const text = await response.text()
  let json = null
  try {
    json = JSON.parse(text)
  } catch {
    // CSV / binary endpoints are allowed to be non-JSON.
  }
  return { status: response.status, code: json?.code, message: json?.message, data: json?.data, text }
}

async function login(tenant) {
  const result = await request('/api/auth/login', {
    method: 'POST',
    body: { account: tenant.account, password: tenant.password, tenant_code: tenant.code },
  })
  if (!result.data?.token) {
    throw new Error(`[${tenant.code}] 登录失败：${result.message || result.text}`)
  }
  return result.data.token
}

function endpoint(spec) {
  const [method, path] = spec.split(' ')
  return { method, path }
}

function assert(condition, message) {
  if (!condition) throw new Error(message)
}

async function checkTenant(tenant) {
  const token = await login(tenant)
  const profile = await request('/api/users/me', { token })
  assert(profile.code === 0, `[${tenant.code}] profile 获取失败`)
  const permissionCodes = new Set(profile.data?.permission_codes || [])
  const features = new Set(profile.data?.features || [])
  console.log(`\n[${tenant.code}] permissions=${permissionCodes.size} features=${features.size}`)

  for (const spec of tenant.mustAllow) {
    const { method, path } = endpoint(spec)
    const result = await request(path, { method, token })
    assert(result.status !== 403 && result.code !== 40300, `[${tenant.code}] 应允许但被拒绝：${spec} -> ${result.message}`)
    console.log(`  allow ${spec}`)
  }

  for (const spec of tenant.mustDeny) {
    const { method, path } = endpoint(spec)
    const result = await request(path, { method, token })
    assert(result.status === 403 || result.code === 40300, `[${tenant.code}] 应拒绝但放行：${spec}`)
    console.log(`  deny  ${spec}`)
  }

  for (const check of tenant.actionChecks) {
    const result = await request(check.path, { method: check.method, token, body: check.body })
    if (check.expect403) {
      assert(result.status === 403 || result.code === 40300, `[${tenant.code}] 应拒绝但放行：${check.name}`)
    }
    if (check.expectNot403) {
      assert(result.status !== 403 && result.code !== 40300, `[${tenant.code}] 应通过权限闸但被拒绝：${check.name}`)
    }
    console.log(`  action ${check.name} -> ${result.status}/${result.code ?? 'non-json'}`)
  }
}

for (const tenant of tenants) {
  await checkTenant(tenant)
}

console.log('\n租户套餐/权限受控检查通过')
