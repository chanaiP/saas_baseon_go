<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Delete, Edit, Plus, Refresh, Search } from '@element-plus/icons-vue'

import NeuroAgentPageShell from '@/views/components/NeuroAgentPageShell.vue'

import { createAiResource, deleteAiResource, fetchAiOverview, fetchAiResource, updateAiResource } from '../api'
import type { AiOverview, AiPage, AiResource, AiSectionConfig } from '../types'

defineOptions({ name: 'AiCapabilityCenterView' })

const route = useRoute()
const router = useRouter()

const sections: AiSectionConfig[] = [
  { key: 'dashboard', title: '总览', route: '/ai-capability-center', description: '平台调用、成本、成功率、租户排行榜和健康检查。', columns: [] },
  { key: 'providers', title: '供应商', route: '/ai-capability-center/providers', resource: 'providers', description: '维护供应商、接入资源、预算、区域和负责人。', writable: true, columns: [
    { key: 'name', label: '供应商', width: 180 }, { key: 'code', label: '编码', width: 160 }, { key: 'type', label: '类型', width: 120 }, { key: 'base_url', label: 'Endpoint' }, { key: 'auth_type', label: '鉴权', width: 120 }, { key: 'qps_limit', label: 'QPS', width: 90 }, { key: 'monthly_budget', label: '月预算', width: 120 }, { key: 'owner', label: '负责人', width: 120 }, { key: 'status', label: '状态', width: 100 },
  ] },
  { key: 'models', title: '模型目录', route: '/ai-capability-center/models', resource: 'models', description: '按供应商维护模型、能力标签、质量指标和默认场景。', writable: true, columns: [
    { key: 'model_name', label: '模型', width: 200 }, { key: 'model_code', label: '模型编码', width: 180 }, { key: 'model_type', label: '类型', width: 120 }, { key: 'capabilities', label: '能力' }, { key: 'context_window', label: '上下文', width: 110 }, { key: 'success_rate', label: '成功率', width: 100 }, { key: 'latency_p95', label: 'P95', width: 100 }, { key: 'status', label: '状态', width: 100 },
  ] },
  { key: 'scenarios', title: 'AI 场景', route: '/ai-capability-center/scenarios', resource: 'scenarios', description: '强制注册 app_code + ai_scenario_code，并绑定默认基础路由。', writable: true, columns: [
    { key: 'ai_scenario_name', label: '场景', width: 200 }, { key: 'ai_scenario_code', label: '场景编码', width: 200 }, { key: 'app_name', label: '应用', width: 160 }, { key: 'app_code', label: '应用编码', width: 160 }, { key: 'scenario_type', label: '场景类型', width: 120 }, { key: 'capability_code', label: '能力', width: 160 }, { key: 'default_base_route_id', label: '默认基础路由', width: 240 }, { key: 'owner', label: '负责人', width: 120 }, { key: 'status', label: '状态', width: 100 },
  ] },
  { key: 'routes', title: '基础路由', route: '/ai-capability-center/routes', resource: 'base-routes', description: '维护可复用模型池和路由策略，不绑定租户和具体场景。', writable: true, columns: [
    { key: 'route_name', label: '路由', width: 220 }, { key: 'route_code', label: '路由编码', width: 200 }, { key: 'capability_code', label: '能力', width: 160 }, { key: 'model_type', label: '模型类型', width: 120 }, { key: 'strategy', label: '策略', width: 140 }, { key: 'timeout_ms', label: '超时', width: 100 }, { key: 'max_retry', label: '重试', width: 90 }, { key: 'status', label: '状态', width: 100 },
  ] },
  { key: 'usage', title: '用量统计', route: '/ai-capability-center/usage', resource: 'usage-records', description: '按租户、应用、场景、模型、供应商统计用量、成本和销售额。', columns: [
    { key: 'called_at', label: '调用时间', width: 180 }, { key: 'tenant_name', label: '租户', width: 160 }, { key: 'app_name', label: '应用', width: 140 }, { key: 'ai_scenario_name', label: '场景', width: 180 }, { key: 'usage_amount', label: '用量', width: 100 }, { key: 'usage_unit', label: '单位', width: 120 }, { key: 'calls', label: '调用', width: 90 }, { key: 'cost_amount', label: '成本', width: 100 }, { key: 'billing_amount', label: '销售额', width: 110 }, { key: 'status', label: '状态', width: 100 },
  ] },
  { key: 'strategy', title: '策略中心', route: '/ai-capability-center/strategy', resource: 'tenant-strategies', description: '一条租户策略统一表达路由覆盖、多维配额、多维限流和超限动作。', writable: true, columns: [
    { key: 'policy_name', label: '策略', width: 220 }, { key: 'tenant_scope', label: '租户范围', width: 130 }, { key: 'tenant_ids', label: '租户', width: 180 }, { key: 'app_code', label: '应用', width: 150 }, { key: 'ai_scenario_code', label: 'AI 场景', width: 190 }, { key: 'override_base_route_id', label: '覆盖路由', width: 240 }, { key: 'status', label: '状态', width: 100 },
  ] },
  { key: 'settings', title: '系统设置', route: '/ai-capability-center/settings', resource: 'settings', description: '维护网关参数、能力字典基线和安全归属。', writable: true, columns: [
    { key: 'setting_key', label: '配置项', width: 220 }, { key: 'setting_value', label: '配置值' }, { key: 'description', label: '说明', width: 300 }, { key: 'status', label: '状态', width: 100 },
  ] },
]

const activeSection = computed(() => sections.find((item) => item.route === route.path) ?? sections[0])
const overview = ref<AiOverview | null>(null)
const page = ref<AiPage<Record<string, unknown>>>({ items: [], total: 0, skip: 0, limit: 20 })
const currentPage = computed(() => Math.floor(page.value.skip / Math.max(page.value.limit, 1)) + 1)
const keyword = ref('')
const loading = ref(false)
const errorText = ref('')
const editorVisible = ref(false)
const editorMode = ref<'create' | 'edit'>('create')
const editorJson = ref('{}')
const editingId = ref('')

const canWrite = computed(() => Boolean(activeSection.value.resource && activeSection.value.writable))

function displayCell(row: Record<string, unknown>, key: string) {
  const value = row[key]
  if (Array.isArray(value)) return value.join('、') || '-'
  if (value === null || value === undefined || value === '') return '-'
  if (typeof value === 'object') return JSON.stringify(value)
  return String(value)
}

function statusType(value: unknown) {
  if (value === 'active' || value === 'success') return 'success'
  if (value === 'warning') return 'warning'
  if (value === 'error' || value === 'failed' || value === 'timeout') return 'danger'
  return 'info'
}

async function loadData() {
  loading.value = true
  errorText.value = ''
  try {
    if (!activeSection.value.resource) {
      overview.value = await fetchAiOverview()
      return
    }
    page.value = await fetchAiResource(activeSection.value.resource, {
      skip: page.value.skip,
      limit: page.value.limit,
      keyword: keyword.value.trim(),
    })
  } catch (error) {
    errorText.value = error instanceof Error ? error.message : '数据加载失败'
  } finally {
    loading.value = false
  }
}

function switchSection(path: string) {
  if (path !== route.path) router.push(path)
}

function openCreate() {
  editorMode.value = 'create'
  editingId.value = ''
  editorJson.value = JSON.stringify(defaultPayload(activeSection.value.resource), null, 2)
  editorVisible.value = true
}

function openEdit(row: Record<string, unknown>) {
  editorMode.value = 'edit'
  editingId.value = String(row.id || '')
  editorJson.value = JSON.stringify(row, null, 2)
  editorVisible.value = true
}

async function saveEditor() {
  if (!activeSection.value.resource) return
  let payload: Record<string, unknown>
  try {
    payload = JSON.parse(editorJson.value) as Record<string, unknown>
  } catch {
    ElMessage.error('JSON 格式不合法')
    return
  }
  if (editorMode.value === 'edit' && editingId.value) {
    await updateAiResource(activeSection.value.resource, editingId.value, payload)
    ElMessage.success('已更新')
  } else {
    await createAiResource(activeSection.value.resource, payload)
    ElMessage.success('已创建')
  }
  editorVisible.value = false
  await loadData()
}

async function removeRow(row: Record<string, unknown>) {
  if (!activeSection.value.resource) return
  const id = String(row.id || '')
  if (!id) return
  await ElMessageBox.confirm('确认删除该配置？删除会写入底座操作日志。', '删除确认', { type: 'warning' })
  await deleteAiResource(activeSection.value.resource, id)
  ElMessage.success('已删除')
  await loadData()
}

function defaultPayload(resource?: AiResource): Record<string, unknown> {
  const status = 'active'
  if (resource === 'providers') return { name: '', code: '', type: 'public_cloud', base_url: '', auth_type: 'api_key', status, priority: 80, region: 'CN', qps_limit: 100, monthly_budget: 10000, owner: '' }
  if (resource === 'models') return { provider_id: '', model_code: '', model_name: '', model_type: 'text', capabilities: ['text_generation'], context_window: 32000, unit: 'tokens', latency_p95: 0, success_rate: 0, status, default_for: [] }
  if (resource === 'scenarios') return { app_code: '', app_name: '', ai_scenario_code: '', ai_scenario_name: '', scenario_type: 'text', capability_code: 'text_generation', model_type: 'text', default_base_route_id: '', owner: '', version: 'v1.0', status }
  if (resource === 'base-routes') return { route_code: '', route_name: '', capability_code: 'text_generation', model_type: 'text', strategy: 'fallback', timeout_ms: 30000, max_retry: 2, status }
  if (resource === 'tenant-strategies') return { policy_name: '', tenant_scope: 'include', tenant_ids: [], app_code: '', app_name: '', ai_scenario_code: '', ai_scenario_name: '', default_base_route_id: '', override_base_route_id: '', status }
  if (resource === 'settings') return { setting_key: '', setting_value: {}, description: '', status }
  return { status }
}

watch(() => route.path, () => {
  page.value.skip = 0
  keyword.value = ''
  loadData()
})

onMounted(loadData)
</script>

<template>
  <NeuroAgentPageShell class="ai-center-page">
    <template #title>AI 能力中心</template>
    <template #subtitle>{{ activeSection.description }}</template>
    <template #actions>
      <el-button :icon="Refresh" @click="loadData">刷新</el-button>
      <el-button v-if="canWrite" type="primary" :icon="Plus" @click="openCreate">新增配置</el-button>
    </template>

    <div class="ai-center-tabs">
      <button v-for="item in sections" :key="item.key" :class="{ active: item.route === route.path }" @click="switchSection(item.route)">
        {{ item.title }}
      </button>
    </div>

    <div v-if="activeSection.key === 'dashboard'" class="ai-dashboard" v-loading="loading">
      <el-alert v-if="errorText" :title="errorText" type="error" show-icon />
      <div class="ai-metrics">
        <article v-for="metric in overview?.metrics || []" :key="metric.label">
          <span>{{ metric.label }}</span>
          <strong>{{ metric.value }}</strong>
          <small>{{ metric.trend }}</small>
        </article>
      </div>
      <section class="ai-panel">
        <header><strong>平台健康检查</strong><span>供应商、路由、租户策略和日志写入状态</span></header>
        <div class="ai-health">
          <el-tag v-for="item in overview?.health_checks || []" :key="item.name" :type="statusType(item.status)">
            {{ item.name }}：{{ item.message }}
          </el-tag>
        </div>
      </section>
      <section class="ai-panel">
        <header><strong>核心基础路由</strong><span>AI 场景默认绑定，租户策略可覆盖</span></header>
        <el-table :data="overview?.core_base_routes || []" border>
          <el-table-column prop="route_name" label="路由" min-width="180" />
          <el-table-column prop="route_code" label="编码" min-width="180" />
          <el-table-column prop="capability_code" label="能力" width="160" />
          <el-table-column prop="strategy" label="策略" width="140" />
          <el-table-column prop="status" label="状态" width="110" />
        </el-table>
      </section>
    </div>

    <div v-else class="ai-resource">
      <div class="ai-toolbar">
        <el-input v-model="keyword" clearable placeholder="搜索当前页面数据" @keyup.enter="loadData">
          <template #prefix><el-icon><Search /></el-icon></template>
        </el-input>
        <el-button :icon="Search" @click="loadData">查询</el-button>
      </div>
      <el-alert v-if="errorText" :title="errorText" type="error" show-icon />
      <el-table v-loading="loading" :data="page.items" border class="ai-table" empty-text="暂无数据">
        <el-table-column v-for="column in activeSection.columns" :key="column.key" :label="column.label" :min-width="column.width || 160">
          <template #default="{ row }">
            <el-tag v-if="column.key === 'status'" :type="statusType(row[column.key])">{{ displayCell(row, column.key) }}</el-tag>
            <span v-else>{{ displayCell(row, column.key) }}</span>
          </template>
        </el-table-column>
        <el-table-column v-if="canWrite" fixed="right" label="操作" width="150">
          <template #default="{ row }">
            <el-button link type="primary" :icon="Edit" @click="openEdit(row)">编辑</el-button>
            <el-button link type="danger" :icon="Delete" @click="removeRow(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
      <el-pagination
        :current-page="currentPage"
        class="ai-pagination"
        layout="total, prev, pager, next"
        :page-size="page.limit"
        :total="page.total"
        @current-change="(pageNo: number) => { page.skip = (pageNo - 1) * page.limit; loadData() }"
      />
    </div>

    <el-dialog v-model="editorVisible" :title="editorMode === 'create' ? '新增配置' : '编辑配置'" width="720px">
      <el-alert title="按后端字段提交 JSON；保存后会写入底座操作日志。" type="info" show-icon />
      <el-input v-model="editorJson" class="json-editor" type="textarea" :rows="18" spellcheck="false" />
      <template #footer>
        <el-button @click="editorVisible = false">取消</el-button>
        <el-button type="primary" @click="saveEditor">保存</el-button>
      </template>
    </el-dialog>
  </NeuroAgentPageShell>
</template>

<style scoped>
.ai-center-page :deep(.neuro-page-shell__panel-body) {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.ai-center-tabs {
  display: flex;
  gap: 8px;
  overflow-x: auto;
  padding-bottom: 2px;
}

.ai-center-tabs button {
  border: 1px solid var(--neuro-border);
  background: var(--neuro-surface);
  color: var(--neuro-text);
  border-radius: 8px;
  padding: 8px 12px;
  white-space: nowrap;
  cursor: pointer;
}

.ai-center-tabs button.active {
  border-color: var(--neuro-primary);
  color: var(--neuro-primary);
  background: color-mix(in srgb, var(--neuro-primary) 10%, var(--neuro-surface));
}

.ai-dashboard,
.ai-resource {
  display: flex;
  flex-direction: column;
  gap: 12px;
  min-width: 0;
}

.ai-metrics {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 12px;
}

.ai-metrics article,
.ai-panel {
  border: 1px solid var(--neuro-border);
  background: var(--neuro-surface);
  border-radius: 8px;
  padding: 14px;
}

.ai-metrics span,
.ai-metrics small,
.ai-panel header span {
  display: block;
  color: var(--neuro-text-muted);
}

.ai-metrics strong {
  display: block;
  margin: 8px 0;
  font-size: 24px;
}

.ai-panel header {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 12px;
}

.ai-health {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.ai-toolbar {
  display: flex;
  gap: 8px;
  align-items: center;
}

.ai-toolbar .el-input {
  max-width: 360px;
}

.ai-table {
  width: 100%;
}

.ai-pagination {
  justify-content: flex-end;
}

.json-editor {
  margin-top: 12px;
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
}

@media (max-width: 900px) {
  .ai-metrics {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .ai-toolbar {
    align-items: stretch;
    flex-direction: column;
  }

  .ai-toolbar .el-input {
    max-width: none;
  }
}
</style>
