#!/usr/bin/env node
import fs from 'node:fs'
import path from 'node:path'

const root = process.cwd()
const manifestPath = path.join(root, 'internal/apps/ai_geo/app.manifest.yaml')
const routesPath = path.join(root, 'internal/bootstrap/api_routes.go')
const policyPath = path.join(root, 'internal/interfaces/http/handlers/auth_policy.go')

function read(file) {
  return fs.readFileSync(file, 'utf8')
}

function normalizeRoute(route) {
  return route.replaceAll('{id}', ':id').replace(/\/+/g, '/')
}

function manifestApis(source) {
  const apis = new Set()
  const lines = source.split('\n')
  for (let index = 0; index < lines.length; index += 1) {
    const methodMatch = lines[index].match(/^\s*-\s+method:\s+([A-Z]+)/)
    if (!methodMatch) continue
    const pathLine = lines.slice(index + 1, index + 5).find(line => /^\s+path:\s+/.test(line))
    if (!pathLine) continue
    const apiPath = pathLine.replace(/^\s+path:\s+/, '').trim()
    if (apiPath.startsWith('/api/ai-geo/')) {
      apis.add(`${methodMatch[1]} ${normalizeRoute(apiPath)}`)
    }
  }
  return apis
}

function registeredRoutes(source) {
  const routes = new Set()
  const pattern = /aiGeo\.(GET|POST|PUT|PATCH|DELETE)\("([^"]+)"/g
  for (const match of source.matchAll(pattern)) {
    routes.add(`${match[1]} ${normalizeRoute(`/api/ai-geo${match[2]}`)}`)
  }
  return routes
}

function policyMap(source, functionName) {
  const start = source.indexOf(`func ${functionName}() map[string]string`)
  if (start < 0) return new Map()
  const end = source.indexOf('\n}', start)
  const body = source.slice(start, end)
  const result = new Map()
  const pattern = /"([^"]+)":\s*"([^"]+)"/g
  for (const match of body.matchAll(pattern)) {
    result.set(normalizeRoute(match[1]), match[2])
  }
  return result
}

function featureCodes(source, sectionName) {
  const start = source.indexOf(`${sectionName}:`)
  if (start < 0) return new Set()
  const nextSection = source.slice(start + sectionName.length + 1).search(/\n[a-z_]+:/)
  const body = nextSection < 0 ? source.slice(start) : source.slice(start, start + sectionName.length + 1 + nextSection)
  return new Set([...body.matchAll(/feature_code:\s+([a-z0-9_:_-]+)/g)].map(match => match[1]))
}

const manifest = read(manifestPath)
const routeSource = read(routesPath)
const policySource = read(policyPath)
const manifestRouteSet = manifestApis(manifest)
const registeredRouteSet = registeredRoutes(routeSource)
const operationPolicy = policyMap(policySource, 'operationPermissionByRoute')
const menuPolicy = policyMap(policySource, 'menuPermissionByRoute')

const failures = []
for (const route of registeredRouteSet) {
  if (!manifestRouteSet.has(route)) {
    failures.push(`manifest missing API: ${route}`)
  }
  const [method, apiPath] = route.split(' ')
  if (method === 'GET') {
    if (!menuPolicy.has(apiPath) && !operationPolicy.has(route)) {
      failures.push(`auth policy missing read mapping: ${route}`)
    }
  } else if (!operationPolicy.has(route)) {
    failures.push(`auth policy missing write mapping: ${route}`)
  }
}
for (const route of manifestRouteSet) {
  if (!registeredRouteSet.has(route)) {
    failures.push(`manifest declares unregistered API: ${route}`)
  }
}

const menuFeatures = featureCodes(manifest, 'menus')
const packageFeatures = featureCodes(manifest, 'package_features')
for (const feature of menuFeatures) {
  if (!packageFeatures.has(feature)) {
    failures.push(`package_features missing menu feature_code: ${feature}`)
  }
}

if (failures.length) {
  console.error('AI GEO manifest parity failed:')
  for (const failure of failures) {
    console.error(`- ${failure}`)
  }
  process.exit(1)
}

console.log(`AI GEO manifest parity passed: ${registeredRouteSet.size} routes, ${operationPolicy.size} write policies, ${menuPolicy.size} menu mappings`)
