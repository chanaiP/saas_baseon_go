export const systemManagementManifest = {
  appCode: 'system-management',
  appName: '系统管理',
  backendManifest: 'internal/apps/system_management/app.manifest.yaml',
  routes: [
    '/tenants',
    '/plans',
    '/organization',
    '/positions',
    '/business-units',
    '/users',
    '/roles',
    '/menus',
    '/dict',
    '/params',
    '/audit-logs',
    '/login-logs',
  ],
}
