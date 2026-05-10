<script setup lang="ts">
defineOptions({ name: 'MonitorCacheView' })
import { onMounted, ref } from 'vue'

import { fetchMonitorCacheStats } from '@/api/monitor'
import type { MonitorCacheStats } from '@/api/monitor'
import { usePageAction } from '@/composables/usePageAction'

const { pageMenu } = usePageAction()

const loading = ref(false)
const data = ref<MonitorCacheStats | null>(null)
const err = ref('')

async function load() {
  if (!pageMenu()) return
  loading.value = true
  err.value = ''
  try {
    data.value = await fetchMonitorCacheStats()
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
          <span>缓存监控（Redis）</span>
          <el-button v-if="pageMenu()" type="primary" plain @click="load">刷新</el-button>
        </div>
      </template>
      <el-empty v-if="!pageMenu()" description="无查看权限" />
      <template v-else>
        <el-alert v-if="err" :title="err" type="error" show-icon :closable="false" style="margin-bottom: 16px" />
        <template v-if="data">
          <el-alert
            v-if="!data.ok"
            :title="data.message || 'Redis 不可用'"
            type="warning"
            show-icon
            :closable="false"
            style="margin-bottom: 16px"
          />
          <div v-else class="grid">
            <div class="tile">
              <span class="label">当前库键数量</span>
              <span class="val">{{ data.keys }}</span>
            </div>
            <div class="tile">
              <span class="label">内存占用（报告值）</span>
              <span class="val">{{ data.used_memory_human || '—' }}</span>
            </div>
            <div class="tile">
              <span class="label">连接客户端数</span>
              <span class="val">{{ data.connected_clients }}</span>
            </div>
          </div>
        </template>
        <p class="hint">键名列表请在「缓存列表」中按模式浏览（生产环境请控制 pattern 范围）。</p>
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
.val {
  font-size: 20px;
  font-weight: 600;
  font-variant-numeric: tabular-nums;
}
.hint {
  margin-top: 16px;
  font-size: 13px;
  color: var(--el-text-color-secondary);
}
</style>
