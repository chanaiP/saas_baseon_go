export const systemMonitorManifest = {
  appCode: 'system-monitor',
  appName: '系统监控',
  backendManifest: 'internal/apps/system_monitor/app.manifest.yaml',
  routes: [
    '/monitor/health',
    '/monitor/server',
    '/monitor/jobs',
    '/monitor/services',
    '/monitor/cache',
    '/monitor/cache-keys',
  ],
}
