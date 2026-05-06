import { test, expect } from '@playwright/test'

async function doLogin(page: import('@playwright/test').Page) {
  await page.context().clearCookies()
  await page.goto('/login')
  await page.getByPlaceholder('工号或手机号').fill('E10001')
  await page.getByLabel('密码').fill('112233')
  await page.getByRole('button', { name: '登录' }).click()
  await page.waitForURL('**/home', { timeout: 15000 })
}

/** 侧栏二级菜单项定位器 */
function sidebarItem(page: import('@playwright/test').Page, text: string) {
  return page.locator('.neuron-title', { hasText: text }).first()
}

function pageHeading(page: import('@playwright/test').Page, text: string) {
  return page.getByRole('heading', { name: new RegExp(text) }).first()
}

/** 点击一级目录展开侧栏 */
async function expandPrimary(page: import('@playwright/test').Page, text: string) {
  // 点击一级菜单（.stem-label 元素），忽略严格模式
  await page.locator('.stem-label', { hasText: text }).first().click()
}

test.describe('System Management Pages', () => {
  test.beforeEach(async ({ page }) => {
    await doLogin(page)
  })

  // --- 主体管理 ---
  test('should load 主体管理 page', async ({ page }) => {
    await expandPrimary(page, '系统管理')
    await sidebarItem(page, '主体管理').click()
    await page.waitForURL('**/tenants', { timeout: 10000 })
    await expect(pageHeading(page, '主体管理')).toBeVisible()
  })

  // --- 组织架构 ---
  test('should load 组织架构 page with tree table', async ({ page }) => {
    await expandPrimary(page, '系统管理')
    await sidebarItem(page, '组织架构').click()
    await page.waitForURL('**/organization', { timeout: 10000 })
    await expect(pageHeading(page, '组织架构')).toBeVisible()
    await expect(page.locator('.el-table').first()).toBeVisible()
  })

  // --- 岗位管理 ---
  test('should load 岗位管理 page with dual panels', async ({ page }) => {
    await expandPrimary(page, '系统管理')
    await sidebarItem(page, '岗位管理').click()
    await page.waitForURL('**/positions', { timeout: 10000 })
    await expect(page.getByRole('heading', { name: '岗位类型' })).toBeVisible()
  })

  // --- 用户管理 ---
  test('should load 用户管理 page with org tree and user list', async ({ page }) => {
    await expandPrimary(page, '系统管理')
    await sidebarItem(page, '用户管理').click()
    await page.waitForURL('**/users', { timeout: 10000 })
    await expect(page.getByText('用户列表')).toBeVisible()
    await expect(page.getByPlaceholder('模糊匹配')).toBeVisible()
    await expect(page.locator('.el-table').first()).toBeVisible()
  })

  // --- 角色权限 ---
  test('should load 角色权限 page with role list', async ({ page }) => {
    await expandPrimary(page, '系统管理')
    await sidebarItem(page, '角色权限').click()
    await page.waitForURL('**/roles', { timeout: 10000 })
    await expect(pageHeading(page, '角色权限')).toBeVisible()
    await expect(page.locator('table').first()).toBeVisible()
  })

  // --- 菜单管理 ---
  test('should load 菜单管理 page with tree table', async ({ page }) => {
    await expandPrimary(page, '系统管理')
    await sidebarItem(page, '菜单管理').click()
    await page.waitForURL('**/menus', { timeout: 10000 })
    await expect(page.getByText('恢复默认')).toBeVisible()
    await expect(page.locator('.el-table').first()).toBeVisible()
  })

  // --- 数据字典 ---
  test('should load 数据字典 page with dual panels', async ({ page }) => {
    await expandPrimary(page, '系统管理')
    await sidebarItem(page, '数据字典').click()
    await page.waitForURL('**/dict', { timeout: 10000 })
    await expect(page.getByText('字典类型').first()).toBeVisible()
  })

  // --- 系统参数 ---
  test('should load 系统参数 page with param list', async ({ page }) => {
    await expandPrimary(page, '系统管理')
    await sidebarItem(page, '参数管理').click()
    await page.waitForURL('**/params', { timeout: 10000 })
    await expect(pageHeading(page, '系统参数')).toBeVisible()
    await expect(page.locator('table').first()).toBeVisible()
  })

  // --- 操作日志 ---
  test('should load 操作日志 page with filters', async ({ page }) => {
    await expandPrimary(page, '系统管理')
    await sidebarItem(page, '操作日志').click()
    await page.waitForURL('**/audit-logs', { timeout: 10000 })
    await expect(pageHeading(page, '操作日志')).toBeVisible()
    await expect(page.getByPlaceholder('模块/动作/摘要/详情')).toBeVisible()
    await expect(page.getByRole('button', { name: '查询' })).toBeVisible()
  })

  // --- 登录日志 ---
  test('should load 登录日志 page with result filter', async ({ page }) => {
    await expandPrimary(page, '系统管理')
    await sidebarItem(page, '登录日志').click()
    await page.waitForURL('**/login-logs', { timeout: 10000 })
    await expect(pageHeading(page, '登录日志')).toBeVisible()
    await expect(page.getByRole('button', { name: '查询' })).toBeVisible()
  })
})

test.describe('Monitor Pages', () => {
  test.beforeEach(async ({ page }) => {
    await doLogin(page)
  })

  // --- 健康检查 ---
  test('should load 健康检查 page', async ({ page }) => {
    await expandPrimary(page, '系统监控')
    await sidebarItem(page, '健康检查').click()
    await page.waitForURL('**/monitor/health', { timeout: 10000 })
    await expect(page.locator('.hdr > span').filter({ hasText: '依赖健康' })).toBeVisible()
  })

  // --- 服务器信息 ---
  test('should load 服务器信息 page', async ({ page }) => {
    await expandPrimary(page, '系统监控')
    await sidebarItem(page, '服务器信息').click()
    await page.waitForURL('**/monitor/server', { timeout: 10000 })
    await expect(page.locator('.hdr > span').filter({ hasText: '服务器进程' })).toBeVisible()
  })

  // --- 定时任务 ---
  test('should load 定时任务 page', async ({ page }) => {
    await expandPrimary(page, '系统监控')
    await sidebarItem(page, '定时任务').click()
    await page.waitForURL('**/monitor/jobs', { timeout: 10000 })
    await expect(pageHeading(page, '定时任务与周期行为')).toBeVisible()
  })

  // --- 服务监控 ---
  test('should load 服务监控 page', async ({ page }) => {
    await expandPrimary(page, '系统监控')
    await sidebarItem(page, '服务监控').click()
    await page.waitForURL('**/monitor/services', { timeout: 10000 })
    await expect(page.locator('.hdr > span').filter({ hasText: '服务监控' })).toBeVisible()
  })

  // --- 缓存监控 ---
  test('should load 缓存监控 page', async ({ page }) => {
    await expandPrimary(page, '系统监控')
    await sidebarItem(page, '缓存监控').click()
    await page.waitForURL('**/monitor/cache', { timeout: 10000 })
    await expect(page.locator('.hdr > span').filter({ hasText: '缓存监控' })).toBeVisible()
  })

  // --- 缓存列表 ---
  test('should load 缓存列表 page with search', async ({ page }) => {
    await expandPrimary(page, '系统监控')
    await sidebarItem(page, '缓存列表').click()
    await page.waitForURL('**/monitor/cache-keys', { timeout: 10000 })
    await expect(pageHeading(page, '缓存列表')).toBeVisible()
    await expect(page.getByPlaceholder(/SCAN/)).toBeVisible()
  })
})

test.describe('Profile Page', () => {
  test.beforeEach(async ({ page }) => {
    await doLogin(page)
  })

  test('should load 个人中心 page', async ({ page }) => {
    await page.goto('/profile')
    await page.waitForURL('**/profile', { timeout: 10000 })
    await expect(page.locator('.neuro-header .neuro-title')).toContainText('个人资料中心')
    await expect(page.locator('.card-title', { hasText: '身份信息' })).toBeVisible()
    await expect(page.locator('.card-title', { hasText: '安全设置' })).toBeVisible()
  })

  test('should show password change form', async ({ page }) => {
    await page.goto('/profile')
    await page.waitForURL('**/profile', { timeout: 10000 })
    await expect(page.locator('input[placeholder*="当前密码"]')).toBeVisible()
    await expect(page.locator('input[placeholder*="新密码"]').first()).toBeVisible()
    await expect(page.getByRole('button', { name: '修改密码' })).toBeVisible()
  })
})
