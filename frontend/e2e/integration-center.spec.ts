import { expect, test } from '@playwright/test'

async function doLogin(page: import('@playwright/test').Page, account = 'E10001') {
  await page.context().clearCookies()
  await page.goto('/login')
  await page.evaluate(() => localStorage.clear())
  await page.goto('/login')
  await page.getByPlaceholder('工号或手机号').fill(account)
  await page.getByLabel('密码').fill('112233')
  await page.getByRole('button', { name: '登录' }).click()
  await page.waitForURL('**/home', { timeout: 15000 })
}

function pagePayload(items: unknown[], total = items.length, skip = 0, limit = 20) {
  return { code: 0, message: 'ok', data: { items, total, skip, limit } }
}

async function mockIntegrationCenter(
  page: import('@playwright/test').Page,
  options: { failLogs?: () => boolean } = {},
) {
  await page.route('**/api/integration-center/**', async (route) => {
    const request = route.request()
    const url = new URL(request.url())
    const limit = Number(url.searchParams.get('limit') || 20)
    const skip = Number(url.searchParams.get('skip') || 0)
    const keyword = url.searchParams.get('keyword') || ''

    if (url.pathname === '/api/integration-center/logs') {
      if (options.failLogs?.()) {
        await route.fulfill({ status: 500, contentType: 'application/json', body: JSON.stringify({ code: 500, message: 'logs down' }) })
        return
      }
      const items = keyword
        ? []
        : [{
            id: skip >= 20 ? 22 : 1,
            request_id: skip >= 20 ? 'PAGE-2-REQUEST' : 'STALE-LOG-REQUEST',
            call_type: 'third_party_api',
            tenant_name: '演示租户',
            method: 'GET',
            endpoint: '/mocked/resource',
            http_status: 200,
            status: 'success',
            duration_ms: 36,
            called_at: '2026-05-15T12:00:00Z',
          }]
      await route.fulfill({ contentType: 'application/json', body: JSON.stringify(pagePayload(items, keyword ? 0 : 25, skip, limit)) })
      return
    }

    if (url.pathname === '/api/integration-center/my-connections') {
      await route.fulfill({
        contentType: 'application/json',
        body: JSON.stringify(pagePayload([{
          id: 101,
          tenant_name: '当前租户',
          platform_name: '京东',
          provider_app_name: '京东店铺应用',
          auth_subject_name: '旗舰店',
          auth_status: 'authorized',
          connection_status: 'active',
          final_capability_count: 3,
          calls_today: 18,
          last_sync_at: '2026-05-15T12:00:00Z',
        }], 1, skip, limit)),
      })
      return
    }

    if (url.pathname === '/api/integration-center/my-sync-jobs') {
      await route.fulfill({
        contentType: 'application/json',
        body: JSON.stringify(pagePayload([{
          id: 201,
          tenant_connection_id: 101,
          platform_name: '京东',
          tenant_name: '当前租户',
          job_type: '订单同步',
          capability_name: '订单',
          trigger_mode: 'schedule',
          status: 'success',
          total_count: 10,
          success_count: 10,
          finished_at: '2026-05-15T12:05:00Z',
        }], 1, skip, limit)),
      })
      return
    }

    await route.fulfill({ contentType: 'application/json', body: JSON.stringify(pagePayload([], 0, skip, limit)) })
  })
}

test.describe('Integration Center frontend states', () => {
  test.beforeEach(async ({ page }) => {
    await doLogin(page)
  })

  test('clears stale table rows when a paged API reload fails', async ({ page }) => {
    let failLogs = false
    await mockIntegrationCenter(page, { failLogs: () => failLogs })

    await page.goto('/integration-center/logs')
    await expect(page.getByText('STALE-LOG-REQUEST')).toBeVisible()

    failLogs = true
    await page.getByPlaceholder('搜索 request_id、平台、租户、Endpoint、错误码').fill('force reload')

    await expect(page.locator('.data-state-banner.error', { hasText: '调用日志加载失败，请检查后端接口与初始化数据' })).toBeVisible()
    await expect(page.getByText('STALE-LOG-REQUEST')).toHaveCount(0)
    await expect(page.getByText('暂无数据')).toBeVisible()
  })

  test('supports backend pagination, filtering and empty state on log table', async ({ page }) => {
    await mockIntegrationCenter(page)

    await page.goto('/integration-center/logs')
    await expect(page.getByText('共 25 条 · 第 1 / 2 页')).toBeVisible()
    await expect(page.getByText('STALE-LOG-REQUEST')).toBeVisible()

    await page.getByRole('button', { name: '下一页' }).click()
    await expect(page.getByText('PAGE-2-REQUEST')).toBeVisible()
    await expect(page.getByText('共 25 条 · 第 2 / 2 页')).toBeVisible()

    await page.getByPlaceholder('搜索 request_id、平台、租户、Endpoint、错误码').fill('empty')
    await expect(page.getByText('暂无数据')).toBeVisible()
    await expect(page.getByText('共 0 条 · 第 1 / 1 页')).toBeVisible()
  })

  test('platform admin can open all menus plus modal and drawer surfaces', async ({ page }) => {
    await mockIntegrationCenter(page)

    const routes = [
      ['/integration-center', '第三方集成中心总览'],
      ['/integration-center/platforms', '新增接入平台'],
      ['/integration-center/workspace', '集成工作台'],
      ['/integration-center/tenant-connections', '租户连接实例'],
      ['/integration-center/sync-monitor', '同步监控'],
      ['/integration-center/quota', '配额与限流策略'],
      ['/integration-center/alerts', '异常监控'],
      ['/integration-center/logs', 'Request ID'],
    ] as const

    for (const [route, text] of routes) {
      await page.goto(route)
      await expect(page.getByText(text).first()).toBeVisible()
    }

    await page.goto('/integration-center/platforms')
    await page.getByRole('button', { name: '新增接入平台' }).click()
    await expect(page.getByText('平台名称')).toBeVisible()
    await page.getByRole('button', { name: '取消' }).click()

    await page.goto('/integration-center/logs')
    await page.getByRole('button', { name: '详情' }).first().click()
    await expect(page.getByText('调用日志详情')).toBeVisible()

    await page.setViewportSize({ width: 390, height: 844 })
    await page.goto('/integration-center/logs')
    await expect(page.getByText('Request ID')).toBeVisible()
    await expect(page.locator('.data-pager').first()).toBeVisible()
  })

  test('normal tenant user cannot enter platform integration governance', async ({ page }) => {
    await doLogin(page, '00001')
    await page.goto('/integration-center')
    await expect(page).toHaveURL(/\/home$/)
    await expect(page.getByText('第三方集成中心总览')).toHaveCount(0)
  })

  test('normal tenant user can open tenant integration portal only', async ({ page }) => {
    await doLogin(page, '00001')
    await mockIntegrationCenter(page)

    await page.goto('/integration-center/my-connections')
    await expect(page.getByText('租户连接实例')).toBeVisible()
    await expect(page.getByRole('cell', { name: '当前租户 旗舰店' })).toBeVisible()
    await expect(page.getByText('旗舰店')).toBeVisible()
    await expect(page.getByRole('heading', { name: '同步任务' })).toBeVisible()
    await expect(page.getByText('订单同步')).toBeVisible()
    await expect(page.getByRole('heading', { name: '服务商应用' })).toHaveCount(0)
  })
})
