import { aiCapabilityCenterRoutes } from './routes'

export const aiCapabilityCenterManifest = {
  appCode: 'ai-capability-center',
  appName: 'AI 能力中心',
  backendManifest: 'internal/apps/ai_capability_center/app.manifest.yaml',
  routes: aiCapabilityCenterRoutes.map((route) => `/${String(route.path)}`),
}
