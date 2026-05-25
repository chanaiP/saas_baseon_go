import type { RouteRecordRaw } from 'vue-router'

export const aiGeoRoutes: RouteRecordRaw[] = [
  {
    path: 'ai-geo',
    name: 'AiGeoRoot',
    redirect: '/ai-geo/dashboard',
  },
  {
    path: 'ai-geo/:section',
    name: 'AiGeoSection',
    meta: { title: 'GEO 内容增长应用' },
    component: () => import('./views/AiGeoView.vue'),
  },
  {
    path: 'ai-geo/:section/:subsection',
    name: 'AiGeoSubSection',
    meta: { title: 'GEO 内容增长应用' },
    component: () => import('./views/AiGeoView.vue'),
  },
]
