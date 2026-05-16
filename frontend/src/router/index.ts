import { createRouter, createWebHistory } from 'vue-router'

import { aiCapabilityCenterRoutes } from '@/apps/ai-capability-center/routes'
import { usePermissionStore } from '@/stores/permission'
import { useSidebarMenuStore } from '@/stores/sidebarMenu'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/login',
      component: () => import('@/views/LoginView.vue'),
      meta: { public: true, title: '登录' },
    },
    {
      path: '/developer',
      name: 'DevHub',
      component: () => import('@/views/DevHubView.vue'),
      meta: { public: true, title: 'DevHub' },
    },
    {
      path: '/home/ai-capability-center/:pathMatch(.*)*',
      redirect: (to) => {
        const suffix = Array.isArray(to.params.pathMatch) ? to.params.pathMatch.join('/') : String(to.params.pathMatch || '')
        return suffix ? `/ai-capability-center/${suffix}` : '/ai-capability-center'
      },
    },
    {
      path: '/',
      component: () => import('@/views/NeuroAgentAdminLayoutCommand.vue'),
      meta: { requiresAuth: true },
      children: [
        {
          path: '',
          redirect: { name: 'HomeDashboard' },
        },
        {
          path: 'home',
          name: 'HomeDashboard',
          meta: { title: '首页' },
          component: () => import('@/views/HomeDashboardView.vue'),
        },
        {
          path: 'apps',
          name: 'AppCenterListView',
          meta: { title: '应用列表', requiresPlatformAdmin: true },
          component: () => import('@/apps/app-center/views/AppCenterListView.vue'),
        },
        {
          path: 'apps/clients',
          name: 'AppCenterClientsView',
          meta: { title: '客户端中心', requiresPlatformAdmin: true },
          component: () => import('@/apps/app-center/views/AppCenterListView.vue'),
        },
        {
          path: 'apps/tenant-openings',
          name: 'AppCenterTenantOpeningsView',
          meta: { title: '租户开通总览', requiresPlatformAdmin: true },
          component: () => import('@/apps/app-center/views/AppCenterListView.vue'),
        },
        {
          path: 'apps/trial-invites',
          name: 'AppCenterTrialInvitesView',
          meta: { title: '体验邀请总览', requiresPlatformAdmin: true },
          component: () => import('@/apps/app-center/views/AppCenterListView.vue'),
        },
        {
          path: 'apps/manifests',
          name: 'AppCenterManifestLoadsView',
          meta: { title: 'Manifest 装载记录', requiresPlatformAdmin: true },
          component: () => import('@/apps/app-center/views/AppCenterListView.vue'),
        },
        {
          path: 'apps/audit-logs',
          name: 'AppCenterAuditLogsView',
          meta: { title: '应用审计日志', requiresPlatformAdmin: true },
          component: () => import('@/apps/app-center/views/AppCenterListView.vue'),
        },
        {
          path: 'model-manager',
          name: 'ModelManagerView',
          meta: { title: '模型管理' },
          component: () => import('@/apps/model-manager/views/ModelManagerView.vue'),
        },
        ...aiCapabilityCenterRoutes,
        {
          path: 'integration-center',
          name: 'IntegrationCenterOverview',
          meta: { title: '第三方集成中心', requiresPlatformAdmin: true },
          component: () => import('@/apps/integration-center/views/IntegrationCenterView.vue'),
        },
        {
          path: 'integration-center/platforms',
          name: 'IntegrationCenterPlatforms',
          meta: { title: '接入平台', requiresPlatformAdmin: true },
          component: () => import('@/apps/integration-center/views/IntegrationCenterView.vue'),
        },
        {
          path: 'integration-center/workspace',
          name: 'IntegrationCenterWorkspace',
          meta: { title: '集成工作台', requiresPlatformAdmin: true },
          component: () => import('@/apps/integration-center/views/IntegrationCenterView.vue'),
        },
        {
          path: 'integration-center/tenant-connections',
          name: 'IntegrationCenterTenantConnections',
          meta: { title: '租户连接', requiresPlatformAdmin: true },
          component: () => import('@/apps/integration-center/views/IntegrationCenterView.vue'),
        },
        {
          path: 'integration-center/my-connections',
          name: 'IntegrationCenterMyConnections',
          meta: { title: '我的第三方连接' },
          component: () => import('@/apps/integration-center/views/IntegrationCenterView.vue'),
        },
        {
          path: 'integration-center/sync-monitor',
          name: 'IntegrationCenterSyncMonitor',
          meta: { title: '同步监控', requiresPlatformAdmin: true },
          component: () => import('@/apps/integration-center/views/IntegrationCenterView.vue'),
        },
        {
          path: 'integration-center/quota',
          name: 'IntegrationCenterQuota',
          meta: { title: '配额与限流', requiresPlatformAdmin: true },
          component: () => import('@/apps/integration-center/views/IntegrationCenterView.vue'),
        },
        {
          path: 'integration-center/alerts',
          name: 'IntegrationCenterAlerts',
          meta: { title: '异常监控', requiresPlatformAdmin: true },
          component: () => import('@/apps/integration-center/views/IntegrationCenterView.vue'),
        },
        {
          path: 'integration-center/logs',
          name: 'IntegrationCenterLogs',
          meta: { title: '调用日志', requiresPlatformAdmin: true },
          component: () => import('@/apps/integration-center/views/IntegrationCenterView.vue'),
        },
        {
          path: 'tenants',
          name: 'TenantView',
          meta: { title: '主体管理', requiresPlatformAdmin: true },
          component: () => import('@/views/TenantView.vue'),
        },
        {
          path: 'plans',
          name: 'PlanManagementView',
          meta: { title: '套餐中心', requiresPlatformAdmin: true },
          component: () => import('@/views/PlanManagementView.vue'),
        },
        {
          path: 'organization',
          name: 'OrganizationView',
          meta: { title: '组织架构' },
          component: () => import('@/views/OrganizationView.vue'),
        },
        {
          path: 'business-units',
          name: 'BusinessUnitView',
          meta: { title: '业务单元' },
          component: () => import('@/views/BusinessUnitView.vue'),
        },
        {
          path: 'positions',
          name: 'PositionView',
          meta: { title: '岗位管理' },
          component: () => import('@/views/PositionView.vue'),
        },
        {
          path: 'users',
          name: 'UserView',
          meta: { title: '用户管理' },
          component: () => import('@/views/UserView.vue'),
        },
        {
          path: 'roles',
          name: 'RoleView',
          meta: { title: '角色权限' },
          component: () => import('@/views/RoleView.vue'),
        },
        {
          path: 'roles/:roleId/permission-config',
          name: 'RolePermissionConfigView',
          meta: { title: '权限配置' },
          component: () => import('@/views/RolePermissionConfigView.vue'),
        },
        {
          path: 'menus',
          name: 'MenuView',
          meta: { title: '菜单管理' },
          component: () => import('@/views/MenuView.vue'),
        },
        {
          path: 'dict',
          name: 'DictView',
          meta: { title: '数据字典' },
          component: () => import('@/views/DictView.vue'),
        },
        {
          path: 'params',
          name: 'ParamView',
          meta: { title: '系统参数' },
          component: () => import('@/views/ParamView.vue'),
        },
        {
          path: 'audit-logs',
          name: 'AuditLogsView',
          meta: { title: '操作日志' },
          component: () => import('@/views/AuditLogsView.vue'),
        },
        {
          path: 'login-logs',
          name: 'LoginLogsView',
          meta: { title: '登录日志' },
          component: () => import('@/views/LoginLogsView.vue'),
        },
        {
          path: 'monitor/health',
          name: 'MonitorHealthView',
          meta: { title: '健康检查', requiresPlatformAdmin: true },
          component: () => import('@/views/MonitorHealthView.vue'),
        },
        {
          path: 'monitor/server',
          name: 'MonitorServerView',
          meta: { title: '服务器信息', requiresPlatformAdmin: true },
          component: () => import('@/views/MonitorServerView.vue'),
        },
        {
          path: 'monitor/jobs',
          name: 'MonitorJobsView',
          meta: { title: '定时任务', requiresPlatformAdmin: true },
          component: () => import('@/views/MonitorJobsView.vue'),
        },
        {
          path: 'monitor/services',
          name: 'MonitorServicesView',
          meta: { title: '服务监控', requiresPlatformAdmin: true },
          component: () => import('@/views/MonitorServicesView.vue'),
        },
        {
          path: 'monitor/cache',
          name: 'MonitorCacheView',
          meta: { title: '缓存监控', requiresPlatformAdmin: true },
          component: () => import('@/views/MonitorCacheView.vue'),
        },
        {
          path: 'monitor/cache-keys',
          name: 'MonitorCacheKeysView',
          meta: { title: '缓存列表', requiresPlatformAdmin: true },
          component: () => import('@/views/MonitorCacheKeysView.vue'),
        },
        {
          path: 'profile',
          name: 'ProfileView',
          meta: { title: '个人中心' },
          component: () => import('@/views/ProfileView.vue'),
        },
      ],
    },
  ],
})

router.onError((error, to) => {
  // 动态 chunk 加载失败（通常是构建后旧 chunk 已不存在），刷新页面获取最新资源
  if (
    error.message.includes('Failed to fetch dynamically imported module') ||
    error.message.includes('not a valid JavaScript MIME type')
  ) {
    window.location.href = to.fullPath
  }
})

router.beforeEach(async (to) => {
  if (to.meta.public) return true
  const token = localStorage.getItem('access_token')
  if (!token) {
    return { path: '/login', query: { redirect: to.fullPath } }
  }
  if (to.meta.requiresAuth) {
    const perm = usePermissionStore()
    const sidebarMenu = useSidebarMenuStore()
    const profileStale = perm.isStale()
    const needProfile = !perm.loaded || profileStale
    const needMenu = !sidebarMenu.overridesLoaded
    if (needProfile || needMenu) {
      try {
        const profileTask = needProfile
          ? perm.load({ force: profileStale })
          : Promise.resolve(perm.profile)
        const menuTask = needMenu
          ? // 平台管理员分支与 profile 并行触发；具体走向在 store 内部按 isPlatformAdmin 处理。
            // 首次或 force=true 时仍会真正请求菜单接口。
            (async () => {
              const p = await profileTask
              await sidebarMenu.loadTenantMenuRuntime({
                isPlatformAdmin: !!p?.is_platform_admin,
                force: profileStale || !sidebarMenu.overridesLoaded,
              })
            })()
          : Promise.resolve()
        await Promise.all([profileTask, menuTask])
      } catch (e: unknown) {
        /** 仅会话失效时退出登录；网络超时、5xx 等保留 token，避免误判「登录超时」。 */
        const st =
          typeof e === 'object' && e !== null && 'status' in e
            ? Number((e as { status?: unknown }).status)
            : NaN
        if (st === 401) {
          localStorage.removeItem('access_token')
          sidebarMenu.clearTenantMenuRuntime()
          perm.clear()
          return { path: '/login', query: { redirect: to.fullPath } }
        }
        if (!perm.loaded) {
          return { path: '/login', query: { redirect: to.fullPath } }
        }
      }
    }
    const platformRouteScope =
      !!(perm.profile?.is_platform_admin || perm.profile?.tenant_is_platform)
    if (to.meta.requiresPlatformAdmin && !platformRouteScope) {
      return { path: '/home', replace: true }
    }
    const routeMenu = sidebarMenu.menuForRoute(to.path)
    if (!perm.profile?.is_platform_admin && routeMenu?.enabled === false) {
      return { path: '/home', replace: true }
    }
    const permissionPath = routeMenu?.path ?? to.path
    if (!perm.profile?.is_platform_admin && !['/home', '/profile'].includes(to.path)) {
      const pathOk = perm.canUseMenuPath(permissionPath)
      const opsOk =
        !pathOk &&
        routeMenu?.type === 'menu' &&
        (routeMenu.children || []).some(
          (c) =>
            c.type === 'button' &&
            c.enabled !== false &&
            !!c.permissionCode &&
            perm.canUseAction(c.permissionCode),
        )
      if (!pathOk && !opsOk) {
        return { path: '/home', replace: true }
      }
    }
  }
  return true
})

export default router
