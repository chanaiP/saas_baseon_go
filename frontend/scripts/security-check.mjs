import { readdir, readFile } from 'node:fs/promises'
import { fileURLToPath } from 'node:url'
import { join, relative } from 'node:path'

const root = fileURLToPath(new URL('../src', import.meta.url))
const allowedTokenFiles = new Set([
  'api/http.ts',
  'router/index.ts',
  'views/LoginView.vue',
  'views/NeuroAgentAdminLayoutCommand.vue',
])

const findings = []

async function walk(dir) {
  const entries = await readdir(dir, { withFileTypes: true })
  for (const entry of entries) {
    const path = join(dir, entry.name)
    if (entry.isDirectory()) {
      await walk(path)
      continue
    }
    if (!/\.(ts|vue|js)$/.test(entry.name)) continue
    await checkFile(path)
  }
}

async function checkFile(path) {
  const rel = relative(root, path)
  const source = await readFile(path, 'utf8')
  const lines = source.split(/\r?\n/)
  lines.forEach((line, index) => {
    const location = `${rel}:${index + 1}`
    if (/\bv-html\b|\.innerHTML\b|\.outerHTML\b/.test(line)) {
      findings.push(`${location} unsafe HTML injection is not allowed`)
    }
    if (/console\.(log|debug|info|warn|error)\(.*(access_token|Authorization|Bearer|token)/i.test(line)) {
      findings.push(`${location} token or Authorization value must not be logged`)
    }
    if (/access_token/.test(line) && !allowedTokenFiles.has(rel)) {
      findings.push(`${location} access_token is only allowed in auth/http boundary files`)
    }
  })
}

await walk(root)

await checkPermissionGuardContract()

if (findings.length > 0) {
  console.error(findings.join('\n'))
  process.exit(1)
}

console.log('frontend security check passed')

async function checkPermissionGuardContract() {
  const router = await readFile(join(root, 'router/index.ts'), 'utf8')
  const directive = await readFile(join(root, 'directives/permission.ts'), 'utf8')
  const sidebar = await readFile(join(root, 'stores/sidebarMenu.ts'), 'utf8')

  const routerRequirements = [
    ['router.beforeEach', 'router must keep a navigation guard'],
    ['perm.load', 'router guard must refresh backend profile permissions'],
    ['sidebarMenu.loadTenantMenuRuntime', 'router guard must refresh backend menu config'],
    ['perm.canUseMenuPath', 'router guard must check backend menu permission codes'],
    ['perm.canUseAction', 'router guard must allow menu entry only through backend action codes when needed'],
  ]
  for (const [needle, message] of routerRequirements) {
    if (!router.includes(needle)) findings.push(`router/index.ts ${message}`)
  }

  if (!directive.includes('permissionStore.canUseAction')) {
    findings.push('directives/permission.ts button visibility must use backend action permission codes')
  }
  if (!sidebar.includes('loadTenantMenuRuntime') || !sidebar.includes('fetchMenuBundles')) {
    findings.push('stores/sidebarMenu.ts menu refresh must use backend menu bundles instead of localStorage only')
  }
}
