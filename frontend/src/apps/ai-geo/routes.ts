import type { RouteRecordRaw } from 'vue-router'

export const aiGeoRoutes: RouteRecordRaw[] = [
  {
    path: 'ai-geo',
    name: 'AiGeoRoot',
    meta: { title: 'AI GEO' },
    component: () => import('./views/AiGeoView.vue'),
  },
  {
    path: 'ai-geo/:section',
    name: 'AiGeoSection',
    meta: { title: 'AI GEO' },
    component: () => import('./views/AiGeoView.vue'),
  },
]
