<template>
  <div class="api-docs-content-wrapper" :data-theme="theme">
    <!-- 左侧：API 模块列表 -->
    <aside class="docs-sidebar">
      <div class="sidebar-header">
        <h2 class="sidebar-title">API 文档</h2>
        <p class="sidebar-version">v{{ apiVersion }} · {{ totalEndpoints }} 个接口</p>
        <div class="sidebar-actions">
          <a class="swagger-link" href="/docs" target="_blank" rel="noopener">🔗 在 Swagger 中打开</a>
          <button class="refresh-btn" @click="refresh" title="刷新接口列表">↻ 刷新</button>
        </div>
      </div>

      <div class="search-box">
        <el-input v-model="searchQuery" placeholder="搜索接口..." clearable size="small" :prefix-icon="Search" />
      </div>

      <nav class="module-list">
        <button
          v-for="mod in filteredModules"
          :key="mod.name"
          class="module-item"
          :class="{ 'module-active': activeModule === mod.name }"
          @click="activeModule = mod.name"
        >
          <span class="mod-icon">{{ mod.icon }}</span>
          <div class="mod-info">
            <span class="mod-name">{{ mod.title }}</span>
            <span class="mod-count">{{ mod.endpoints.length }} 个接口</span>
          </div>
        </button>
      </nav>
    </aside>

    <!-- 右侧：接口详情 -->
    <main class="docs-main">
      <template v-if="currentModule">
        <div class="module-header">
          <h3 class="module-title">
            <span class="title-icon">{{ currentModule.icon }}</span>
            {{ currentModule.title }}
          </h3>
          <p class="module-desc">{{ currentModule.description }}</p>
        </div>

        <div class="endpoint-list">
          <div
            v-for="ep in currentModule.endpoints"
            :key="ep.path + ep.method"
            class="endpoint-card"
            :class="{ 'expanded': expandedEndpoint === ep.path + ep.method }"
            @click="toggleExpand(ep)"
          >
            <div class="endpoint-summary">
              <span class="method-badge" :class="'method-' + ep.method.toLowerCase()">{{ ep.method }}</span>
              <code class="endpoint-path">{{ ep.path }}</code>
              <span class="endpoint-summary-text">{{ ep.summary }}</span>
            </div>

            <div v-if="expandedEndpoint === ep.path + ep.method" class="endpoint-detail">
              <div v-if="ep.description" class="detail-section">
                <h4 class="detail-title">说明</h4>
                <p class="detail-text">{{ ep.description }}</p>
              </div>

              <div v-if="ep.parameters && ep.parameters.length" class="detail-section">
                <h4 class="detail-title">请求参数</h4>
                <table class="param-table">
                  <thead>
                    <tr>
                      <th>参数名</th>
                      <th>位置</th>
                      <th>类型</th>
                      <th>必填</th>
                      <th>说明</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr v-for="p in ep.parameters" :key="p.name">
                      <td><code>{{ p.name }}</code></td>
                      <td>{{ p.in }}</td>
                      <td>{{ p.schema?.type || p.schema?.format || 'string' }}</td>
                      <td>{{ p.required ? '是' : '否' }}</td>
                      <td>{{ p.description || '-' }}</td>
                    </tr>
                  </tbody>
                </table>
              </div>

              <div v-if="ep.requestBody" class="detail-section">
                <h4 class="detail-title">请求体</h4>
                <pre class="code-block"><code>{{ ep.requestBody }}</code></pre>
              </div>

              <div v-if="ep.responses && Object.keys(ep.responses).length" class="detail-section">
                <h4 class="detail-title">响应</h4>
                <pre class="code-block"><code>{{ ep.responses }}</code></pre>
              </div>

              <div class="detail-section">
                <h4 class="detail-title">调用示例</h4>
                <pre class="code-block"><code>{{ generateCurl(ep) }}</code></pre>
                <button class="copy-btn" @click.stop="copyCurl(ep)">复制</button>
              </div>
            </div>
          </div>
        </div>
      </template>

      <div v-else class="docs-empty">
        <div class="empty-icon">📡</div>
        <h3>API 接口文档</h3>
        <p>选择左侧模块查看接口详情</p>
        <p class="hint">所有接口均需携带 Authorization: Bearer &lt;token&gt; 请求头</p>
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { Search } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'

defineProps<{ theme: 'dark' | 'light' }>()

const apiVersion = ref('1.0.0')
const searchQuery = ref('')
const activeModule = ref('')
const expandedEndpoint = ref('')
const rawSpec = ref<any>(null)

interface ModuleInfo {
  name: string
  title: string
  icon: string
  description: string
  endpoints: EndpointInfo[]
}

interface EndpointInfo {
  path: string
  method: string
  summary: string
  description: string
  parameters?: any[]
  requestBody?: string
  responses?: string
  tags?: string[]
}

const moduleIcons: Record<string, string> = {
  auth: '🔐', tenants: '🏢', users: '👤', roles: '🎭',
  permissions: '🔑', organizations: '🏗️', positions: '💼',
  'dict-param': '📖', logs: '📋', monitor: '📊',
  branding: '🎨', files: '📁', batch: '📦', health: '💓',
}

const moduleTitles: Record<string, string> = {
  auth: '认证与登录', tenants: '主体管理', users: '用户管理',
  roles: '角色管理', permissions: '权限管理', organizations: '组织架构',
  positions: '岗位管理', 'dict-param': '数据字典与系统参数',
  logs: '日志管理', monitor: '系统监控', branding: '品牌配置',
  files: '文件管理', batch: '批量操作', health: '健康检查',
}

const modules = computed<ModuleInfo[]>(() => {
  if (!rawSpec.value) return []
  const spec = rawSpec.value
  apiVersion.value = spec.info?.version || '1.0.0'

  const modMap: Record<string, ModuleInfo> = {}
  for (const [path, pathItem] of Object.entries(spec.paths || {})) {
    for (const method of ['get', 'post', 'put', 'delete', 'patch']) {
      const op = (pathItem as any)[method]
      if (!op || op.tags?.includes('health')) continue

      const tagName = (op.tags?.[0] || 'other')
      if (!modMap[tagName]) {
        modMap[tagName] = {
          name: tagName,
          title: moduleTitles[tagName] || tagName,
          icon: moduleIcons[tagName] || '📌',
          description: findTagDescription(spec, tagName),
          endpoints: [],
        }
      }
      modMap[tagName].endpoints.push({
        path, method: method.toUpperCase(),
        summary: op.summary || '', description: op.description || '',
        parameters: op.parameters || [],
        requestBody: op.requestBody ? formatRequestBody(op.requestBody) : undefined,
        responses: op.responses ? formatResponses(op.responses) : undefined,
        tags: op.tags || [],
      })
    }
  }

  const order = ['auth', 'tenants', 'users', 'roles', 'permissions', 'organizations', 'positions', 'dict-param', 'logs', 'monitor', 'branding', 'files', 'batch']
  return Object.values(modMap).sort((a, b) => {
    const ai = order.indexOf(a.name), bi = order.indexOf(b.name)
    return (ai === -1 ? 999 : ai) - (bi === -1 ? 999 : bi)
  })
})

const filteredModules = computed(() => {
  if (!searchQuery.value) return modules.value
  const q = searchQuery.value.toLowerCase()
  return modules.value.map(mod => ({
    ...mod,
    endpoints: mod.endpoints.filter(ep =>
      ep.path.toLowerCase().includes(q) || ep.summary.toLowerCase().includes(q) || ep.method.toLowerCase().includes(q)
    ),
  })).filter(mod => mod.endpoints.length > 0 || mod.title.toLowerCase().includes(q))
})

const currentModule = computed(() => {
  if (!activeModule.value) return null
  return modules.value.find(m => m.name === activeModule.value) || null
})

const totalEndpoints = computed(() => modules.value.reduce((s, m) => s + m.endpoints.length, 0))

function findTagDescription(spec: any, tagName: string): string {
  return spec.tags?.find((t: any) => t.name === tagName)?.description || ''
}

function formatRequestBody(rb: any): string {
  try {
    const json = rb.content?.['application/json']
    if (json?.schema) return JSON.stringify(json.schema, null, 2)
  } catch { /* ignore */ }
  return ''
}

function formatResponses(responses: any): string {
  try {
    const result: any = {}
    for (const [code, resp] of Object.entries(responses)) {
      const json = (resp as any).content?.['application/json']
      result[code] = json?.schema ? { schema: json.schema } : { description: (resp as any).description }
    }
    return JSON.stringify(result, null, 2)
  } catch { /* ignore */ }
  return ''
}

function generateCurl(ep: EndpointInfo): string {
  const baseUrl = window.location.origin
  let cmd = `curl -X ${ep.method.toUpperCase()} '${baseUrl}${ep.path}'`
  cmd += ` \\\n  -H 'Authorization: Bearer <your_token>'`
  if (ep.method === 'POST' || ep.method === 'PUT' || ep.method === 'PATCH') {
    cmd += ` \\\n  -H 'Content-Type: application/json' \\\n  -d '{}'`
  }
  return cmd
}

function copyCurl(ep: EndpointInfo) {
  navigator.clipboard.writeText(generateCurl(ep))
    .then(() => ElMessage.success('已复制到剪贴板'))
    .catch(() => ElMessage.error('复制失败'))
}

function toggleExpand(ep: EndpointInfo) {
  const key = ep.path + ep.method
  expandedEndpoint.value = expandedEndpoint.value === key ? '' : key
}

onMounted(async () => {
  await loadSpec()
})

async function loadSpec() {
  try {
    const resp = await fetch('/openapi.json')
    rawSpec.value = await resp.json()
    if (modules.value.length > 0 && !activeModule.value) activeModule.value = modules.value[0].name
  } catch {
    ElMessage.error('无法加载 OpenAPI 规范')
  }
}

async function refresh() {
  await loadSpec()
  ElMessage.success('API 文档已更新')
}

defineExpose({ refresh })
</script>

<style scoped>
.api-docs-content-wrapper {
  display: flex;
  gap: 20px;
  padding: 20px;
  min-height: calc(100vh - 90px);
  background: var(--docs-bg, #0a0c14);
  color: var(--docs-text, #e2e8f0);
}

/* 左侧 */
.docs-sidebar {
  width: 260px;
  flex-shrink: 0;
  background: var(--docs-card-bg, rgba(15, 23, 42, 0.7));
  backdrop-filter: blur(20px);
  border: 1px solid var(--docs-border, rgba(0, 245, 212, 0.2));
  border-radius: 16px;
  padding: 16px;
  align-self: flex-start;
  position: sticky;
  top: 110px;
  max-height: calc(100vh - 130px);
  display: flex;
  flex-direction: column;
}

.sidebar-header {
  padding: 8px 8px 16px;
  border-bottom: 1px solid var(--docs-border, rgba(100, 116, 139, 0.2));
  margin-bottom: 12px;
}

.sidebar-title {
  font-family: 'Orbitron', 'JetBrains Mono', monospace;
  font-size: 18px;
  font-weight: 700;
  background: linear-gradient(135deg, var(--docs-primary, #00f5d4), var(--docs-secondary, #9d4edd));
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  margin: 0;
}

.sidebar-version {
  color: var(--docs-text-muted, #64748b);
  font-size: 12px;
  margin: 4px 0 0;
  font-family: 'JetBrains Mono', monospace;
}

.swagger-link {
  display: block;
  font-size: 12px;
  color: var(--docs-primary, #00f5d4);
  text-decoration: none;
  padding: 4px 8px;
  background: rgba(0, 245, 212, 0.06);
  border: 1px solid rgba(0, 245, 212, 0.15);
  border-radius: 6px;
  text-align: center;
  transition: all 0.2s ease;
}

.swagger-link:hover {
  background: rgba(0, 245, 212, 0.12);
  border-color: rgba(0, 245, 212, 0.3);
}

.sidebar-actions {
  display: flex;
  gap: 6px;
  margin-top: 8px;
}

.refresh-btn {
  flex-shrink: 0;
  font-size: 12px;
  color: var(--docs-primary, #00f5d4);
  padding: 4px 8px;
  background: rgba(0, 245, 212, 0.06);
  border: 1px solid rgba(0, 245, 212, 0.15);
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s ease;
  font-family: inherit;
}

.refresh-btn:hover {
  background: rgba(0, 245, 212, 0.12);
  border-color: rgba(0, 245, 212, 0.3);
}

.search-box { margin-bottom: 8px; }
.search-box :deep(.el-input__wrapper) {
  background: rgba(0, 245, 212, 0.05);
  border-color: rgba(0, 245, 212, 0.15);
  border-radius: 8px;
  box-shadow: none;
}

.module-list {
  display: flex;
  flex-direction: column;
  gap: 4px;
  overflow-y: auto;
  flex: 1;
}

.module-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 12px;
  background: transparent;
  border: 1px solid transparent;
  border-radius: 10px;
  color: var(--docs-text-secondary, #94a3b8);
  cursor: pointer;
  transition: all 0.2s ease;
  text-align: left;
  width: 100%;
  font-family: inherit;
}

.module-item:hover {
  background: var(--docs-hover, rgba(0, 245, 212, 0.05));
  border-color: rgba(0, 245, 212, 0.15);
}

.module-active {
  background: rgba(0, 245, 212, 0.1);
  border-color: rgba(0, 245, 212, 0.3);
}

.mod-icon { font-size: 18px; flex-shrink: 0; }

.mod-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 0;
}

.mod-name {
  font-size: 13px;
  font-weight: 600;
  color: var(--docs-text, #e2e8f0);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.module-active .mod-name { color: var(--docs-primary, #00f5d4); }
.mod-count { font-size: 11px; color: var(--docs-text-muted, #64748b); }

/* 右侧 */
.docs-main { flex: 1; min-width: 0; }

.module-header {
  background: var(--docs-card-bg, rgba(15, 23, 42, 0.7));
  backdrop-filter: blur(20px);
  border: 1px solid var(--docs-border, rgba(0, 245, 212, 0.2));
  border-radius: 16px;
  padding: 24px;
  margin-bottom: 16px;
}

.module-title {
  font-family: 'Orbitron', 'JetBrains Mono', monospace;
  font-size: 20px;
  font-weight: 600;
  color: var(--docs-primary, #00f5d4);
  margin: 0 0 8px;
  display: flex;
  align-items: center;
  gap: 10px;
}

.title-icon { font-size: 22px; }
.module-desc { color: var(--docs-text-secondary, #94a3b8); font-size: 14px; margin: 0; line-height: 1.5; }

.endpoint-list { display: flex; flex-direction: column; gap: 8px; }

.endpoint-card {
  background: var(--docs-card-bg, rgba(15, 23, 42, 0.7));
  backdrop-filter: blur(20px);
  border: 1px solid var(--docs-border, rgba(0, 245, 212, 0.15));
  border-radius: 12px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.endpoint-card:hover { border-color: rgba(0, 245, 212, 0.3); }
.endpoint-card.expanded { border-color: rgba(0, 245, 212, 0.3); background: rgba(0, 245, 212, 0.03); }

.endpoint-summary { display: flex; align-items: center; gap: 12px; padding: 14px 18px; flex-wrap: wrap; }

.method-badge {
  font-size: 11px;
  font-weight: 700;
  padding: 3px 10px;
  border-radius: 6px;
  font-family: 'JetBrains Mono', monospace;
  min-width: 50px;
  text-align: center;
  flex-shrink: 0;
}

.method-get { color: #10b981; background: rgba(16, 185, 129, 0.12); border: 1px solid rgba(16, 185, 129, 0.3); }
.method-post { color: #3b82f6; background: rgba(59, 130, 246, 0.12); border: 1px solid rgba(59, 130, 246, 0.3); }
.method-put { color: #f59e0b; background: rgba(245, 158, 11, 0.12); border: 1px solid rgba(245, 158, 11, 0.3); }
.method-delete { color: #ef4444; background: rgba(239, 68, 68, 0.12); border: 1px solid rgba(239, 68, 68, 0.3); }
.method-patch { color: #a855f7; background: rgba(168, 85, 247, 0.12); border: 1px solid rgba(168, 85, 247, 0.3); }

.endpoint-path { font-family: 'JetBrains Mono', monospace; font-size: 13px; color: var(--docs-text, #e2e8f0); flex-shrink: 0; }
.endpoint-summary-text { font-size: 13px; color: var(--docs-text-secondary, #94a3b8); flex: 1; min-width: 0; }

.endpoint-detail {
  padding: 0 18px 18px;
  border-top: 1px solid var(--docs-border, rgba(100, 116, 139, 0.15));
  animation: fadeIn 0.3s ease;
}

@keyframes fadeIn {
  from { opacity: 0; transform: translateY(-4px); }
  to { opacity: 1; transform: translateY(0); }
}

.detail-section { margin-top: 16px; }

.detail-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--docs-primary, #00f5d4);
  margin: 0 0 10px;
  padding-bottom: 6px;
  border-bottom: 1px solid var(--docs-border, rgba(100, 116, 139, 0.15));
}

.detail-text { font-size: 13px; color: var(--docs-text-secondary, #94a3b8); line-height: 1.6; margin: 0; }

.param-table { width: 100%; border-collapse: collapse; font-size: 13px; }

.param-table th {
  text-align: left;
  padding: 8px 12px;
  background: rgba(0, 245, 212, 0.05);
  color: var(--docs-text, #e2e8f0);
  font-weight: 600;
  border-bottom: 1px solid var(--docs-border, rgba(100, 116, 139, 0.2));
}

.param-table td {
  padding: 8px 12px;
  color: var(--docs-text-secondary, #94a3b8);
  border-bottom: 1px solid rgba(100, 116, 139, 0.1);
}

.param-table code {
  font-family: 'JetBrains Mono', monospace;
  background: rgba(0, 245, 212, 0.06);
  padding: 2px 6px;
  border-radius: 4px;
  font-size: 12px;
  color: var(--docs-primary, #00f5d4);
}

.code-block {
  background: rgba(0, 0, 0, 0.3);
  border: 1px solid var(--docs-border, rgba(100, 116, 139, 0.15));
  border-radius: 8px;
  padding: 14px;
  overflow-x: auto;
  font-family: 'JetBrains Mono', monospace;
  font-size: 12px;
  line-height: 1.6;
  color: #e2e8f0;
  margin: 0;
  position: relative;
}

.copy-btn {
  position: absolute;
  right: 24px;
  margin-top: -36px;
  padding: 4px 12px;
  background: rgba(0, 245, 212, 0.1);
  border: 1px solid rgba(0, 245, 212, 0.3);
  border-radius: 6px;
  color: var(--docs-primary, #00f5d4);
  font-size: 12px;
  cursor: pointer;
  transition: all 0.2s ease;
  font-family: inherit;
}

.copy-btn:hover { background: rgba(0, 245, 212, 0.2); }

.docs-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 60vh;
  text-align: center;
  background: var(--docs-card-bg, rgba(15, 23, 42, 0.5));
  backdrop-filter: blur(20px);
  border: 1px solid var(--docs-border, rgba(0, 245, 212, 0.15));
  border-radius: 16px;
  padding: 40px;
}

.empty-icon { font-size: 48px; margin-bottom: 16px; }

.docs-empty h3 {
  font-family: 'Orbitron', 'JetBrains Mono', monospace;
  font-size: 20px;
  color: var(--docs-primary, #00f5d4);
  margin: 0 0 8px;
}

.docs-empty p { color: var(--docs-text-secondary, #94a3b8); font-size: 14px; margin: 0; }
.hint { margin-top: 12px; font-size: 12px; font-family: 'JetBrains Mono', monospace; color: var(--docs-text-muted, #64748b); }

@media (max-width: 900px) {
  .api-docs-content-wrapper { flex-direction: column; }
  .docs-sidebar { width: 100%; position: static; max-height: none; }
  .module-list { flex-direction: row; flex-wrap: wrap; }
  .module-item { width: auto; }
  .endpoint-summary { flex-direction: column; align-items: flex-start; gap: 6px; }
}

/* 浅色模式 */
.api-docs-content-wrapper[data-theme="light"] {
  --docs-bg: #f8fafc;
  --docs-border: rgba(14, 165, 233, 0.15);
  --docs-card-bg: rgba(248, 250, 252, 0.9);
  --docs-primary: #0ea5e9;
  --docs-secondary: #8b5cf6;
  --docs-text: #1e293b;
  --docs-text-secondary: #475569;
  --docs-text-muted: #64748b;
  --docs-hover: rgba(14, 165, 233, 0.05);
  --docs-active-bg: rgba(14, 165, 233, 0.1);
}

.api-docs-content-wrapper[data-theme="light"] .code-block {
  background: rgba(0, 0, 0, 0.04);
  color: #1e293b;
  border-color: rgba(14, 165, 233, 0.15);
}

.api-docs-content-wrapper[data-theme="light"] .copy-btn {
  color: #0ea5e9;
  background: rgba(14, 165, 233, 0.1);
  border-color: rgba(14, 165, 233, 0.3);
}

.api-docs-content-wrapper[data-theme="light"] .swagger-link {
  color: #0ea5e9;
  background: rgba(14, 165, 233, 0.06);
  border-color: rgba(14, 165, 233, 0.15);
}

.api-docs-content-wrapper[data-theme="light"] .swagger-link:hover {
  background: rgba(14, 165, 233, 0.12);
  border-color: rgba(14, 165, 233, 0.3);
}

.api-docs-content-wrapper[data-theme="light"] .refresh-btn {
  color: #0ea5e9;
  background: rgba(14, 165, 233, 0.06);
  border-color: rgba(14, 165, 233, 0.15);
}

.api-docs-content-wrapper[data-theme="light"] .refresh-btn:hover {
  background: rgba(14, 165, 233, 0.12);
  border-color: rgba(14, 165, 233, 0.3);
}
</style>
