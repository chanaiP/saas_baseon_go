import type { RouteRecordRaw } from 'vue-router'

export const aiCapabilityCenterRoutes: RouteRecordRaw[] = [
  {
    path: 'ai-capability-center',
    name: 'AiDashboardView',
    meta: { title: 'AI 能力中心', requiresPlatformAdmin: true },
    component: () => import('./views/AiDashboardView.vue'),
  },
  {
    path: 'ai-capability-center/providers',
    name: 'AiProvidersView',
    meta: { title: '供应商', requiresPlatformAdmin: true },
    component: () => import('./views/AiProvidersView.vue'),
  },
  {
    path: 'ai-capability-center/models',
    name: 'AiModelsView',
    meta: { title: '模型目录', requiresPlatformAdmin: true },
    component: () => import('./views/AiModelsView.vue'),
  },
  {
    path: 'ai-capability-center/scenarios',
    name: 'AiScenariosView',
    meta: { title: 'AI 场景', requiresPlatformAdmin: true },
    component: () => import('./views/AiScenariosView.vue'),
  },
  {
    path: 'ai-capability-center/routes',
    name: 'AiRoutesView',
    meta: { title: '基础路由', requiresPlatformAdmin: true },
    component: () => import('./views/AiRoutesView.vue'),
  },
  {
    path: 'ai-capability-center/strategy',
    name: 'AiStrategyView',
    meta: { title: '策略中心', requiresPlatformAdmin: true },
    component: () => import('./views/AiStrategyView.vue'),
  },
  {
    path: 'ai-capability-center/usage-logs',
    name: 'AiUsageLogsView',
    meta: { title: '调用日志', requiresPlatformAdmin: true },
    component: () => import('./views/AiUsageLogsView.vue'),
  },
  {
    path: 'ai-capability-center/test-console',
    name: 'AiTestConsoleView',
    meta: { title: '测试窗口', requiresPlatformAdmin: true },
    component: () => import('./views/AiTestConsoleView.vue'),
  },
  {
    path: 'ai-capability-center/settings',
    name: 'AiSettingsView',
    meta: { title: '系统设置', requiresPlatformAdmin: true },
    component: () => import('./views/AiSettingsView.vue'),
  },
]
