import { test, expect } from '@playwright/test'

const BASE_URL = 'http://localhost:5174'

async function doLogin(page: import('@playwright/test').Page) {
  await page.context().clearCookies()
  await page.goto('/login')
  await page.getByPlaceholder('工号或手机号').fill('E10001')
  await page.getByLabel('密码').fill('112233')
  await page.getByRole('button', { name: '登录' }).click()
  await page.waitForURL('**/home', { timeout: 15000 })
}

/** 侧栏菜单项定位器（排除内容区卡片） */
function sidebarItem(page: import('@playwright/test').Page, text: string) {
  return page.locator('.neuron-title', { hasText: text }).first()
}

/** 内容区卡片标题定位器 */
function contentCard(page: import('@playwright/test').Page, text: string) {
  return page.locator('.card-title', { hasText: text }).first()
}

test.describe('Framework E2E', () => {

  test.describe('Login Page', () => {
    test('should redirect to login when accessing protected route without token', async ({ page }) => {
      await page.context().clearCookies()
      await page.goto('/home')
      await expect(page).toHaveURL(/\/login/)
    })

    test('should show login form with branding', async ({ page }) => {
      await page.goto('/login')
      await expect(page.locator('.login-page-root')).toBeVisible()
      await expect(page.getByPlaceholder('工号或手机号')).toBeVisible()
      await expect(page.getByRole('button', { name: '登录' })).toBeVisible()
    })

    test('should login successfully and redirect to home', async ({ page }) => {
      await page.context().clearCookies()
      await page.goto('/login')

      await page.getByPlaceholder('工号或手机号').fill('E10001')
      await page.getByLabel('密码').fill('112233')
      await page.getByRole('button', { name: '登录' }).click()

      await page.waitForURL('**/home', { timeout: 15000 })
      await expect(page).toHaveURL(/\/home/)
      await expect(page.locator('.neuro-command-layout')).toBeVisible()
    })

    test('should remember account after logout', async ({ page }) => {
      await doLogin(page)

      // Logout — handle confirm dialog
      page.on('dialog', d => d.accept())
      await page.locator('.logout-btn').click()
      await page.waitForURL('**/login', { timeout: 10000 })

      // Account should be remembered
      const input = page.getByPlaceholder('工号或手机号')
      await expect(input).toHaveValue('E10001')
    })
  })

  test.describe('Admin Shell', () => {
    test.beforeEach(async ({ page }) => {
      await doLogin(page)
    })

    test('should navigate to different pages via sidebar', async ({ page }) => {
      // Click 系统管理 (top-level menu)
      await page.getByText('系统管理').click()

      // Navigate to 用户管理
      await sidebarItem(page, '用户管理').click()
      await page.waitForURL('**/users', { timeout: 10000 })
      await expect(page).toHaveURL(/\/users/)

      // Breadcrumb should update
      await expect(page.locator('.breadcrumb-item.active')).toContainText('用户管理')
    })

    test('should toggle theme', async ({ page }) => {
      const layout = page.locator('.neuro-command-layout')
      await expect(layout).toHaveAttribute('data-theme', 'dark')

      // Toggle to light
      await page.getByTitle('主题切换').click()
      await expect(layout).toHaveAttribute('data-theme', 'light')

      // Toggle back
      await page.getByTitle('主题切换').click()
      await expect(layout).toHaveAttribute('data-theme', 'dark')
    })

    test('should show correct breadcrumb', async ({ page }) => {
      await page.getByText('系统管理').click()
      await sidebarItem(page, '组织架构').click()
      await page.waitForURL('**/organization', { timeout: 10000 })

      await expect(page.locator('.breadcrumb-item').first()).toContainText('系统管理')
      await expect(page.locator('.breadcrumb-item.active')).toContainText('组织架构')
    })

    test('should keep pages cached (keep-alive)', async ({ page }) => {
      // Go to 组织架构
      await page.getByText('系统管理').click()
      await sidebarItem(page, '组织架构').click()
      await page.waitForURL('**/organization', { timeout: 10000 })

      // Go to 用户管理
      await sidebarItem(page, '用户管理').click()
      await page.waitForURL('**/users', { timeout: 10000 })

      // Go back to 组织架构
      await sidebarItem(page, '组织架构').click()
      await page.waitForURL('**/organization', { timeout: 10000 })
      await expect(page).toHaveURL(/\/organization/)
    })

    test('should refresh current page', async ({ page }) => {
      await page.getByText('系统管理').click()
      await sidebarItem(page, '用户管理').click()
      await page.waitForURL('**/users', { timeout: 10000 })

      const refreshBtn = page.locator('.page-refresh-btn')
      await expect(refreshBtn).toBeVisible()
      await refreshBtn.click()

      await expect(page).toHaveURL(/\/users/)
    })

    test('should logout and redirect to login', async ({ page }) => {
      page.on('dialog', d => d.accept())
      await page.locator('.logout-btn').click()
      await page.waitForURL('**/login', { timeout: 10000 })
      await expect(page).toHaveURL(/\/login/)

      // Accessing protected route should still redirect
      await page.goto('/home')
      await expect(page).toHaveURL(/\/login/)
    })
  })

  test.describe('Public Routes', () => {
    test('should access /developer without login', async ({ page }) => {
      await page.goto('/developer')
      await expect(page.locator('.dev-hub')).toBeVisible()
    })

    test('should show component showcase in developer portal without login', async ({ page }) => {
      await page.goto('/developer')
      await expect(page.locator('.showcase-sidebar')).toBeVisible()
    })
  })
})
