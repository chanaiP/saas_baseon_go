import type { RouteRecordRaw } from 'vue-router'

export const dataCenterRoutes: RouteRecordRaw[] = [
  {
    path: 'data-center',
    name: 'DataCenterOverview',
    meta: { title: 'Ai经营决策中心' },
    component: () => import('./views/DataCenterView.vue'),
  },
  {
    path: 'data-center/:section',
    name: 'DataCenterView',
    meta: { title: 'Ai经营决策中心' },
    component: () => import('./views/DataCenterView.vue'),
  },
]
