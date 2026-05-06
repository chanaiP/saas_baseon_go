<script setup lang="ts">
import { computed, ref, watch } from 'vue'

import type { CapabilityQuotaValue, Plan, PlanCapabilityNode } from '@/api/plan'
import { collectFeatureIds, collectQuotaValues } from '@/composables/usePlanCapabilityMatrix'

const visible = defineModel<boolean>({ default: false })

const props = defineProps<{
  node: PlanCapabilityNode | null
  plan: Plan | null
  enabledFeatureIds: number[]
  quotaValues: Record<number, number>
}>()

const emit = defineEmits<{
  apply: [payload: { featureIds: number[]; quotaValues: Record<number, number> }]
}>()

const localFeatureIds = ref<number[]>([])
const localQuotaValues = ref<Record<number, number>>({})

const featureRows = computed(() => {
  if (!props.node) return []
  const rows: Array<{ id: number; name: string; code: string; type: string }> = []
  function walk(node: PlanCapabilityNode) {
    if (node.feature_id != null) {
      rows.push({
        id: node.feature_id,
        name: node.label,
        code: node.feature_code || '',
        type: node.feature_type || 'FEATURE',
      })
    }
    for (const child of node.children || []) walk(child)
  }
  walk(props.node)
  return rows
})

const quotaRows = computed<CapabilityQuotaValue[]>(() => {
  if (!props.node || !props.plan) return []
  return collectQuotaValues(props.node, props.plan.id)
})

watch(
  () => [props.node?.id, props.plan?.id, props.enabledFeatureIds, props.quotaValues] as const,
  () => {
    localFeatureIds.value = [...props.enabledFeatureIds]
    localQuotaValues.value = { ...props.quotaValues }
  },
  { immediate: true, deep: true },
)

function applyChanges() {
  emit('apply', {
    featureIds: [...localFeatureIds.value],
    quotaValues: { ...localQuotaValues.value },
  })
  visible.value = false
}

function toggleAll(enabled: boolean) {
  if (!props.node) return
  const ids = collectFeatureIds(props.node)
  const selected = new Set(localFeatureIds.value)
  for (const id of ids) {
    if (enabled) selected.add(id)
    else selected.delete(id)
  }
  localFeatureIds.value = [...selected].sort((a, b) => a - b)
}
</script>

<template>
  <el-drawer v-model="visible" size="520px" class="capability-drawer">
    <template #header>
      <div class="drawer-title">
        <span>{{ plan?.plan_name || '未选择套餐' }}</span>
        <strong>{{ node?.label || '能力详情' }}</strong>
      </div>
    </template>

    <div v-if="node && plan" class="drawer-body">
      <section class="drawer-card intro-card">
        <span class="node-type">{{ node.node_type === 'domain' ? '业务域' : node.feature_type || '能力' }}</span>
        <h3>{{ node.label }}</h3>
        <p>{{ node.description || '在这里集中维护该套餐下此能力范围的功能开关与配额。' }}</p>
        <div class="quick-actions">
          <el-button type="primary" plain @click="toggleAll(true)">全部启用</el-button>
          <el-button plain @click="toggleAll(false)">全部禁用</el-button>
        </div>
      </section>

      <section class="drawer-card">
        <div class="card-head">
          <strong>功能开关</strong>
          <small>{{ localFeatureIds.length }} 项已启用</small>
        </div>
        <el-checkbox-group v-model="localFeatureIds" class="feature-checks">
          <div v-for="feature in featureRows" :key="feature.id" class="feature-check">
            <el-checkbox :value="feature.id">
              <span class="feature-check-text">
                <strong>{{ feature.name }}</strong>
                <small>{{ feature.code }} · {{ feature.type }}</small>
              </span>
            </el-checkbox>
          </div>
        </el-checkbox-group>
      </section>

      <section v-if="quotaRows.length" class="drawer-card">
        <div class="card-head">
          <strong>关联配额</strong>
          <small>-1 表示无限制，0 表示不可用</small>
        </div>
        <div v-for="quota in quotaRows" :key="quota.quota_id" class="quota-edit-row">
          <div>
            <strong>{{ quota.quota_name }}</strong>
            <small>{{ quota.quota_code }} · {{ quota.period_type || 'NONE' }} · {{ quota.unit || 'COUNT' }}</small>
          </div>
          <el-input-number v-model="localQuotaValues[quota.quota_id]" :min="-1" controls-position="right" />
        </div>
      </section>
    </div>

    <template #footer>
      <el-button @click="visible = false">取消</el-button>
      <el-button type="primary" @click="applyChanges">应用到矩阵</el-button>
    </template>
  </el-drawer>
</template>

<style scoped>
/* 与全局主题一致：浅色默认，html.dark 时覆盖变量（见 uiPreferences applyDomTheme） */
.capability-drawer {
  --cap-card-bg: rgba(255, 255, 255, 0.88);
  --cap-card-border: rgba(120, 140, 190, 0.2);
  --cap-intro-bg:
    radial-gradient(circle at 0 0, rgba(71, 120, 255, 0.18), transparent 38%),
    linear-gradient(135deg, #ffffff, #f6f9ff);
  --cap-title-muted: #667085;
  --cap-title-strong: #111827;
  --cap-text-muted: #667085;
  --cap-heading: #101828;
  --cap-node-bg: color-mix(in srgb, #1eb7a6 26%, #ffffff);
  --cap-node-color: #0f172a;
  --cap-node-border: color-mix(in srgb, #1eb7a6 42%, transparent);
  --cap-feature-row-bg: rgba(17, 24, 39, 0.035);
  --cap-row-divider: rgba(120, 140, 190, 0.14);
}

.dark .capability-drawer {
  --cap-card-bg: rgba(20, 37, 55, 0.96);
  --cap-card-border: rgba(0, 245, 212, 0.2);
  --cap-intro-bg:
    radial-gradient(circle at 0 0, rgba(71, 120, 255, 0.28), transparent 42%),
    linear-gradient(135deg, rgba(24, 38, 54, 0.98), rgba(18, 30, 44, 0.98));
  --cap-title-muted: #92a6b4;
  --cap-title-strong: #f8fdff;
  --cap-text-muted: #92a6b4;
  --cap-heading: #e7f3f7;
  --cap-node-bg: rgba(0, 245, 212, 0.16);
  --cap-node-color: #ecfeff;
  --cap-node-border: rgba(0, 245, 212, 0.35);
  --cap-feature-row-bg: rgba(255, 255, 255, 0.06);
  --cap-row-divider: rgba(0, 245, 212, 0.14);
}

.capability-drawer :deep(.el-drawer__header) {
  margin-bottom: 0;
}

.capability-drawer :deep(.el-drawer__body) {
  color: var(--cap-heading);
}

.drawer-title {
  display: flex;
  flex-direction: column;
  gap: 3px;
}

.drawer-title span {
  color: var(--cap-title-muted);
  font-size: 12px;
}

.drawer-title strong {
  color: var(--cap-title-strong);
  font-size: 20px;
}

.drawer-body {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.drawer-card {
  padding: 18px;
  border: 1px solid var(--cap-card-border);
  border-radius: var(--neuro-radius-xl, 16px);
  background: var(--cap-card-bg);
}

.intro-card {
  background: var(--cap-intro-bg);
}

.node-type {
  display: inline-flex;
  padding: 4px 9px;
  border-radius: 999px;
  border: 1px solid var(--cap-node-border);
  background: var(--cap-node-bg);
  color: var(--cap-node-color);
  font-size: 12px;
  font-weight: 800;
}

.drawer-card h3 {
  margin: 10px 0 6px;
  color: var(--cap-heading);
}

.drawer-card p,
.card-head small,
.feature-check-text small,
.quota-edit-row small {
  color: var(--cap-text-muted);
}

.drawer-card .card-head strong,
.feature-check-text strong,
.quota-edit-row strong {
  color: var(--cap-heading);
}

.quick-actions {
  display: flex;
  gap: 8px;
  margin-top: 14px;
}

.card-head,
.quota-edit-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}

.feature-checks {
  display: flex;
  flex-direction: column;
  gap: 10px;
  margin-top: 14px;
}

.feature-check {
  padding: 12px;
  border-radius: var(--neuro-radius-lg, 12px);
  background: var(--cap-feature-row-bg);
}

.feature-check :deep(.el-checkbox) {
  width: 100%;
  height: auto;
  align-items: flex-start;
  white-space: normal;
}

.feature-check :deep(.el-checkbox__input) {
  padding-top: 2px;
}

.feature-check :deep(.el-checkbox__label) {
  flex: 1;
  min-width: 0;
  padding-left: 10px;
  line-height: 1.35;
  color: var(--cap-heading);
  white-space: normal;
}

.feature-check-text {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.quota-edit-row div {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.quota-edit-row {
  padding: 12px 0;
  border-top: 1px solid var(--cap-row-divider);
}

.quota-edit-row:first-of-type {
  margin-top: 10px;
}
</style>
