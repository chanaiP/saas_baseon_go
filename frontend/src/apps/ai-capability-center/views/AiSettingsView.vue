<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { Refresh } from '@element-plus/icons-vue'

import NeuroAgentPageShell from '@/views/components/NeuroAgentPageShell.vue'

import { fetchAiResource } from '../api'
import { emptyPage, settingJSON, statusType, text } from './viewHelpers'
import AiResourceActions from './AiResourceActions.vue'
import './aiPrototype.css'

defineOptions({ name: 'AiSettingsView' })

const loading = ref(false)
const settings = ref(emptyPage())
const capabilities = ref(emptyPage())

const gatewayRuntime = computed(() => settingJSON(settings.value.items.find((row) => row.setting_key === 'gateway_runtime')))
const security = computed(() => settingJSON(settings.value.items.find((row) => row.setting_key === 'security')))

async function loadData() {
  loading.value = true
  try {
    const [settingPage, capabilityPage] = await Promise.all([
      fetchAiResource('settings', { skip: 0, limit: 200 }),
      fetchAiResource('capabilities', { skip: 0, limit: 200 }),
    ])
    settings.value = settingPage
    capabilities.value = capabilityPage
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '系统设置加载失败')
  } finally {
    loading.value = false
  }
}

onMounted(loadData)
</script>

<template>
  <NeuroAgentPageShell class="ai-prototype">
    <template #title>系统设置</template>
    <template #subtitle>维护网关参数、能力字典基线和安全归属。</template>
    <template #actions>
      <el-button :icon="Refresh" :loading="loading" @click="loadData">刷新</el-button>
    </template>

    <div class="ai-settings-grid">
      <section class="ai-card">
        <header class="ai-card__header"><div><h3>网关配置</h3><p class="ai-card__description">统一入口、超时、重试和降级策略。</p></div></header>
        <div class="ai-card__body ai-setting-list">
          <div><span class="ai-muted">默认超时</span><strong> {{ text(gatewayRuntime.default_timeout_ms) }}ms</strong></div>
          <div><span class="ai-muted">失败重试</span><strong> {{ text(gatewayRuntime.max_retry) }} 次</strong></div>
          <div><span class="ai-muted">异常告警</span><strong> {{ text(gatewayRuntime.alert_channel) }}</strong></div>
          <div><span class="ai-muted">异步用量日志</span><strong> {{ gatewayRuntime.usage_log_async ? '开启' : '关闭' }}</strong></div>
        </div>
      </section>

      <section class="ai-card">
        <header class="ai-card__header"><div><h3>安全与日志归属</h3><p class="ai-card__description">模型能力中心不单独设计安全审计页面，统一写入 SaaS 底座操作日志。</p></div></header>
        <div class="ai-card__body ai-setting-list">
          <div><span class="ai-muted">API Key 加密</span><strong> {{ text(security.key_encryption) }}</strong></div>
          <div><span class="ai-muted">Prompt 明文存储</span><strong> {{ security.prompt_plaintext_store ? '开启' : '关闭' }}</strong></div>
          <div><span class="ai-muted">操作审计</span><strong> {{ text(security.operation_log_target) }}</strong></div>
          <div><span class="ai-muted">归属边界</span><strong> 平台能力，非售卖，非订阅</strong></div>
        </div>
      </section>
    </div>

    <section class="ai-card ai-table">
      <header class="ai-card__header">
        <div>
          <h3>AI 能力字典</h3>
          <p class="ai-card__description">AI 场景、基础路由、模型能力和价格策略都从这里选择能力编码。</p>
        </div>
      </header>
      <div class="ai-card__body">
        <el-table :data="capabilities.items" border v-loading="loading">
          <el-table-column label="能力" min-width="220"><template #default="{ row }"><span class="ai-table-cell-main"><strong>{{ text(row.capability_name) }}</strong><small>{{ text(row.capability_code) }}</small></span></template></el-table-column>
          <el-table-column label="场景 / 模型类型" min-width="160"><template #default="{ row }">{{ text(row.scenario_type) }} / {{ text(row.model_type) }}</template></el-table-column>
          <el-table-column label="默认用量单位" width="140"><template #default="{ row }">{{ text(row.default_billing_unit) }}</template></el-table-column>
          <el-table-column label="分档价格" width="110"><template #default="{ row }">{{ row.supports_tier_pricing ? '支持' : '不支持' }}</template></el-table-column>
          <el-table-column label="状态" width="100"><template #default="{ row }"><el-tag :type="statusType(row.status)">{{ text(row.status) }}</el-tag></template></el-table-column>
          <el-table-column label="操作" width="140"><template #default="{ row }"><AiResourceActions resource="capabilities" :row="row" @saved="loadData" /></template></el-table-column>
        </el-table>
      </div>
    </section>

    <section class="ai-card">
      <header class="ai-card__header">
        <div>
          <h3>环境变量示例</h3>
          <p class="ai-card__description">真实生产环境由后端配置中心和密钥管理服务下发。</p>
        </div>
      </header>
      <div class="ai-card__body">
        <pre class="ai-code-block">AI_GATEWAY_TIMEOUT_MS=30000
AI_GATEWAY_RETRY_TIMES=2
AI_KEY_ENCRYPTION_KMS_ALIAS=prod/ai-capability-center
AI_USAGE_LOG_ASYNC=true
AI_OPERATION_LOG_TARGET=SAAS_BASE_OPERATION_LOG
AI_PROMPT_STORE_PLAINTEXT=false</pre>
      </div>
    </section>
  </NeuroAgentPageShell>
</template>
