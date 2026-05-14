import { readdir, readFile } from 'node:fs/promises'
import { fileURLToPath } from 'node:url'
import { dirname, join, relative, resolve } from 'node:path'

const frontendRoot = resolve(fileURLToPath(new URL('..', import.meta.url)))
const repoRoot = resolve(frontendRoot, '..')
const appsRoot = join(frontendRoot, 'src/apps')
const routerPath = join(frontendRoot, 'src/router/index.ts')

const findings = []
const warnings = []

const appManifests = await readFrontendAppManifests()
const routerRoutes = parseRouterRoutes(await readFile(routerPath, 'utf8'))

for (const manifest of appManifests) {
  const backendPath = resolve(repoRoot, manifest.backendManifest)
  const backend = parseBackendManifest(await readFile(backendPath, 'utf8'))
  const backendMenuPaths = new Set(backend.menus.map((item) => item.path).filter(Boolean))
  const frontendRoutes = new Set(manifest.routes)

  for (const route of frontendRoutes) {
    if (!backendMenuPaths.has(route)) {
      findings.push(`${manifest.appCode}: frontend route ${route} is missing from ${manifest.backendManifest} menus[].path`)
    }
  }
  for (const menu of backend.menus) {
    if (!frontendRoutes.has(menu.path)) {
      findings.push(`${manifest.appCode}: backend menu ${menu.path} is missing from frontend manifest routes`)
    }
    const routerRoute = routerRoutes.get(menu.path)
    if (!routerRoute) {
      findings.push(`${manifest.appCode}: backend menu ${menu.path} is missing from frontend router`)
      continue
    }
    if (routerRoute.title && menu.name && routerRoute.title !== menu.name) {
      warnings.push(`${manifest.appCode}: route title "${routerRoute.title}" differs from backend menu "${menu.name}" for ${menu.path}`)
    }
  }

  const routePrefixes = routePrefixesFor(frontendRoutes)
  for (const [routePath, route] of routerRoutes.entries()) {
    if (!routePrefixes.some((prefix) => routePath === prefix || routePath.startsWith(`${prefix}/`))) {
      continue
    }
    if (route.internalPage) {
      continue
    }
    if (!frontendRoutes.has(routePath)) {
      findings.push(`${manifest.appCode}: router route ${routePath} must be declared in frontend manifest routes or marked meta.internalPage=true`)
    }
  }
}

if (warnings.length > 0) {
  console.warn(warnings.join('\n'))
}
if (findings.length > 0) {
  console.error(findings.join('\n'))
  process.exit(1)
}

console.log('frontend manifest parity check passed')

async function readFrontendAppManifests() {
  const entries = await readdir(appsRoot, { withFileTypes: true })
  const manifests = []
  for (const entry of entries) {
    if (!entry.isDirectory()) continue
    const manifestPath = join(appsRoot, entry.name, 'manifest.ts')
    let source = ''
    try {
      source = await readFile(manifestPath, 'utf8')
    } catch {
      continue
    }
    const appCode = matchString(source, /appCode:\s*'([^']+)'/)
    const appName = matchString(source, /appName:\s*'([^']+)'/)
    const backendManifest = matchString(source, /backendManifest:\s*'([^']+)'/)
    const routesBlock = matchString(source, /routes:\s*\[([\s\S]*?)\]/)
    const routes = [...routesBlock.matchAll(/'([^']+)'/g)].map((match) => normalizePath(match[1]))
    if (!appCode || !backendManifest || routes.length === 0) {
      findings.push(`${relative(repoRoot, manifestPath)} must declare appCode, backendManifest and routes`)
      continue
    }
    manifests.push({ appCode, appName, backendManifest, routes })
  }
  return manifests
}

function parseBackendManifest(source) {
  const menus = []
  let inMenus = false
  let current = null
  for (const rawLine of source.split(/\r?\n/)) {
    const line = rawLine.replace(/\s+#.*$/, '')
    if (/^\S/.test(line)) {
      if (current) menus.push(current)
      current = null
      inMenus = line.trim() === 'menus:'
      continue
    }
    if (!inMenus) continue
    const itemMatch = line.match(/^\s{2}-\s+code:\s*(.+?)\s*$/)
    if (itemMatch) {
      if (current) menus.push(current)
      current = { code: unquote(itemMatch[1]), name: '', path: '' }
      continue
    }
    if (!current) continue
    const fieldMatch = line.match(/^\s{4}(name|path):\s*(.*?)\s*$/)
    if (!fieldMatch) continue
    current[fieldMatch[1]] = fieldMatch[1] === 'path' ? normalizePath(unquote(fieldMatch[2])) : unquote(fieldMatch[2])
  }
  if (current) menus.push(current)
  return { menus }
}

function parseRouterRoutes(source) {
  const routes = new Map()
  const routePattern = /\{\s*path:\s*'([^']+)'([\s\S]*?)(?=\n\s*\},?\n\s*(?:\{|]|,))/g
  for (const match of source.matchAll(routePattern)) {
    const rawPath = match[1]
    if (rawPath === '' || rawPath.includes(':')) continue
    const block = match[2]
    if (/redirect:\s*/.test(block)) continue
    const title = matchString(block, /title:\s*'([^']+)'/)
    const internalPage = /internalPage:\s*true/.test(block)
    routes.set(normalizePath(rawPath), { title, internalPage })
  }
  return routes
}

function routePrefixesFor(routes) {
  const prefixes = new Set()
  for (const route of routes) {
    const parts = route.split('/').filter(Boolean)
    if (parts.length > 0) {
      prefixes.add(`/${parts[0]}`)
    }
  }
  return [...prefixes]
}

function matchString(source, pattern) {
  return source.match(pattern)?.[1]?.trim() || ''
}

function normalizePath(value) {
  const path = value.trim()
  if (path === '') return '/'
  return path.startsWith('/') ? path : `/${path}`
}

function unquote(value) {
  return value.trim().replace(/^['"]|['"]$/g, '')
}
