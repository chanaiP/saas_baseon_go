import { expect, request, test } from '@playwright/test'

const apiPort = process.env.BASICP_API_PORT || '8081'
const apiBaseURL = `http://127.0.0.1:${apiPort}`

async function loginToken() {
  const api = await request.newContext({ baseURL: apiBaseURL })
  const resp = await api.post('/api/auth/login', {
    data: { account: 'E10001', password: '112233' },
  })
  expect(resp.ok()).toBeTruthy()
  const body = await resp.json()
  expect(body.code).toBe(0)
  expect(body.data.token).toBeTruthy()
  await api.dispose()
  return body.data.token as string
}

test.describe('API business regression', () => {
  test('rejects protected APIs without token', async ({ request }) => {
    const resp = await request.get(`${apiBaseURL}/api/users`)
    expect(resp.status()).toBe(401)
    const body = await resp.json()
    expect(body.code).not.toBe(0)
  })

  test('covers user CRUD, role CRUD, quota check and audit logs', async () => {
    const token = await loginToken()
    const api = await request.newContext({
      baseURL: apiBaseURL,
      extraHTTPHeaders: { Authorization: `Bearer ${token}` },
    })
    const stamp = Date.now()

    const createUser = await api.post('/api/users', {
      data: {
        employee_no: `T${stamp}`,
        password: '112233',
        name: `回归用户${stamp}`,
        status: 1,
        role_ids: [],
        position_ids: [],
      },
    })
    expect(createUser.ok()).toBeTruthy()
    const createdUser = await createUser.json()
    expect(createdUser.code).toBe(0)
    const userID = createdUser.data.id

    const updateUser = await api.put(`/api/users/${userID}`, {
      data: { name: `回归用户${stamp}-已更新`, role_ids: [], position_ids: [], department_ids: [] },
    })
    expect(updateUser.ok()).toBeTruthy()
    expect((await updateUser.json()).data.name).toContain('已更新')

    const resetPassword = await api.put(`/api/users/${userID}/password`)
    expect(resetPassword.ok()).toBeTruthy()
    expect((await resetPassword.json()).data.new_password).toBeTruthy()

    const roleCode = `regression_${stamp}`
    const createRole = await api.post('/api/roles', {
      data: { code: roleCode, name: `回归角色${stamp}`, permission_ids: [] },
    })
    expect(createRole.ok()).toBeTruthy()
    const role = await createRole.json()
    expect(role.code).toBe(0)
    const roleID = role.data.id

    const updateRole = await api.put(`/api/roles/${roleID}`, {
      data: { name: `回归角色${stamp}-已更新`, permission_ids: [] },
    })
    expect(updateRole.ok()).toBeTruthy()
    expect((await updateRole.json()).data.name).toContain('已更新')

    const quota = await api.get('/api/tenants/1/quota-check/max_users')
    expect(quota.ok()).toBeTruthy()
    const quotaBody = await quota.json()
    expect(quotaBody.code).toBe(0)
    expect(typeof quotaBody.data.allowed).toBe('boolean')

    const deleteRole = await api.delete(`/api/roles/${roleID}`)
    expect(deleteRole.ok()).toBeTruthy()
    expect((await deleteRole.json()).code).toBe(0)

    const deleteUser = await api.delete(`/api/users/${userID}`)
    expect(deleteUser.ok()).toBeTruthy()
    expect((await deleteUser.json()).code).toBe(0)

    const audit = await api.get('/api/logs/audit')
    expect(audit.ok()).toBeTruthy()
    const auditBody = await audit.json()
    expect(auditBody.code).toBe(0)
    const actions = auditBody.data.items.map((item: { module: string; action: string }) => `${item.module}:${item.action}`)
    expect(actions).toContain('user:create')
    expect(actions).toContain('role:create')

    await api.dispose()
  })
})

