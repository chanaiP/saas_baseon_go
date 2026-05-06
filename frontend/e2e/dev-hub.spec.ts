import { test, expect } from '@playwright/test'

test.describe('DevHub - Unified Developer Portal', () => {

  test.beforeEach(async ({ page }) => {
    await page.goto('/developer')
    // Wait for Vue hydration
    await expect(page.locator('.dev-hub')).toBeVisible()
  })

  test('should load DevHub page with correct title', async ({ page }) => {
    await expect(page.locator('.brand-name')).toHaveText('Ai DevOS')
    await expect(page.locator('.hub-header')).toBeVisible()
  })

  test('should show component showcase by default', async ({ page }) => {
    // Default tab is showcase
    const activeTab = page.locator('.nav-btn.active')
    await expect(activeTab).toContainText('组件展示')

    // Sidebar should show component list
    await expect(page.locator('.showcase-sidebar')).toBeVisible()
    await expect(page.locator('.sidebar-title')).toHaveText('组件库')
  })

  test('should switch to API docs when clicking tab', async ({ page }) => {
    await page.getByRole('button', { name: 'API 文档' }).click()

    const activeTab = page.locator('.nav-btn.active')
    await expect(activeTab).toContainText('API 文档')

    // API docs sidebar should be visible
    await expect(page.locator('.docs-sidebar')).toBeVisible()
    await expect(page.locator('.sidebar-title')).toHaveText('API 文档')
  })

  test('should switch back to showcase', async ({ page }) => {
    // Go to API docs first
    await page.getByRole('button', { name: 'API 文档' }).click()
    await expect(page.locator('.docs-sidebar')).toBeVisible()

    // Switch back
    await page.getByRole('button', { name: '组件展示' }).click()
    await expect(page.locator('.showcase-sidebar')).toBeVisible()
  })

  test('should toggle theme', async ({ page }) => {
    // Default is dark
    const hub = page.locator('.dev-hub')
    await expect(hub).toHaveAttribute('data-theme', 'dark')

    // Toggle to light
    await page.getByTitle('浅色模式').click()
    await expect(hub).toHaveAttribute('data-theme', 'light')

    // Toggle back to dark
    await page.getByTitle('深色模式').click()
    await expect(hub).toHaveAttribute('data-theme', 'dark')
  })

  test('should have Swagger link in header', async ({ page }) => {
    const swaggerBtn = page.locator('a.header-btn')
    await expect(swaggerBtn).toBeVisible()
    await expect(swaggerBtn).toContainText('Swagger')
  })

  test('should have Swagger link in API docs sidebar', async ({ page }) => {
    // Switch to API docs
    await page.getByRole('button', { name: 'API 文档' }).click()

    const link = page.locator('.swagger-link')
    await expect(link).toBeVisible()
    await expect(link).toHaveAttribute('href', '/docs')
  })

  test('should have refresh button in API docs sidebar', async ({ page }) => {
    await page.getByRole('button', { name: 'API 文档' }).click()

    const refreshBtn = page.locator('.refresh-btn')
    await expect(refreshBtn).toBeVisible()
    await expect(refreshBtn).toContainText('刷新')

    // Click and verify success message
    await refreshBtn.click()
    await expect(page.getByText('API 文档已更新')).toBeVisible()
  })

  test('should expand endpoint card when clicked', async ({ page }) => {
    await page.getByRole('button', { name: 'API 文档' }).click()

    // Wait for API spec to load
    await expect(page.locator('.endpoint-card').first()).toBeVisible()

    // Click first endpoint card
    const firstCard = page.locator('.endpoint-card').first()
    await firstCard.click()

    // Should have expanded state
    await expect(firstCard).toHaveClass(/expanded/)
  })

  test('should show component list with expected components', async ({ page }) => {
    await expect(page.locator('.sidebar-title')).toHaveText('组件库')

    // Verify some expected components exist
    await expect(page.locator('button.component-item').filter({ hasText: '智能列表页' })).toBeVisible()
    await expect(page.locator('button.component-item').filter({ hasText: '全息弹窗' })).toBeVisible()
    await expect(page.locator('button.component-item').filter({ hasText: '智能数据表' })).toBeVisible()
  })

  test('should select different component when clicking sidebar item', async ({ page }) => {
    // Default is 智能列表页
    await expect(page.locator('.component-active .comp-name')).toHaveText('智能列表页')

    // Click 全息弹窗
    await page.getByRole('button', { name: '全息弹窗' }).click()
    await expect(page.locator('.component-active .comp-name')).toHaveText('全息弹窗')

    // Should see dialog demo buttons
    await expect(page.getByRole('button', { name: /小尺寸/ })).toBeVisible()
  })
})
