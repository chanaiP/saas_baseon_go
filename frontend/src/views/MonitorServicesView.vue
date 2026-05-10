<script setup lang="ts">
defineOptions({ name: 'MonitorServicesView' })
import { onMounted, ref } from 'vue'

import { fetchMonitorServicesOverview } from '@/api/monitor'
import type { MonitorServicesOverview } from '@/api/monitor'
import { usePageAction } from '@/composables/usePageAction'

const { pageMenu } = usePageAction()

const loading = ref(false)
const data = ref<MonitorServicesOverview | null>(null)
const err = ref('')

async function load() {
  if (!pageMenu()) return
  loading.value = true
  err.value = ''
  try {
    data.value = await fetchMonitorServicesOverview()
  } catch (e) {
    err.value = e instanceof Error ? e.message : '加载失败'
  } finally {
    loading.value = false
  }
}

onMounted(() => void load())
</script>

<template>
  <div class="page">
    <el-card v-loading="loading" shadow="never" class="shell-card">
      <template #header>
        <div class="hdr">
          <span>服务监控</span>
          <el-button v-if="pageMenu()" type="primary" plain @click="load">刷新</el-button>
        </div>
      </template>
      <el-empty v-if="!pageMenu()" description="无查看权限" />
      <template v-else>
        <el-alert v-if="err" :title="err" type="error" show-icon :closable="false" style="margin-bottom: 16px" />
        <div v-if="data" class="grid">
          <div class="tile">
            <span class="label">MySQL</span>
            <el-tag :type="data.mysql ? 'success' : 'danger'" size="large" effect="light" round>
              {{ data.mysql ? '可用' : '不可用' }}
            </el-tag>
          </div>
          <div class="tile">
            <span class="label">Redis</span>
            <el-tag :type="data.redis ? 'success' : 'danger'" size="large" effect="light" round>
              {{ data.redis ? '可用' : '不可用' }}
            </el-tag>
          </div>
          <div class="tile">
            <span class="label">API 进程</span>
            <span class="mono">Python {{ data.python_version }} · PID {{ data.pid }}</span>
          </div>
          <div v-if="data.cpu_percent != null" class="tile">
            <span class="label">CPU</span>
            <span class="mono">{{ data.cpu_percent }}%</span>
          </div>
          <div v-if="data.memory_mb != null" class="tile">
            <span class="label">内存</span>
            <span class="mono">{{ data.memory_mb }} MB</span>
          </div>
        </div>
        <p v-if="data?.note" class="hint">{{ data.note }}</p>
        <p class="hint">详细分项见「健康检查」「服务器信息」菜单。</p>
      </template>
    </el-card>
  </div>
</template>

<style scoped>
.page {
  padding: 0;
}
.shell-card {
  border-radius: 12px;
  border: 1px solid rgba(0, 0, 0, 0.06);
}
.hdr {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: 16px;
}
.tile {
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 20px;
  border-radius: 12px;
  background: rgba(0, 0, 0, 0.02);
}
.label {
  font-size: 13px;
  color: var(--el-text-color-secondary);
}
.mono {
  font-size: 14px;
  font-variant-numeric: tabular-nums;
}
.hint {
  margin-top: 16px;
  font-size: 13px;
  color: var(--el-text-color-secondary);
}
</style>
