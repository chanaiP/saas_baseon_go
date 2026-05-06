<script setup lang="ts">
defineOptions({ name: 'MonitorHealthView' })
import { onMounted, ref } from 'vue'

import { fetchMonitorHealth } from '@/api/monitor'
import type { MonitorHealthDetail } from '@/api/monitor'
import { usePageAction } from '@/composables/usePageAction'

const { pageAction } = usePageAction()

const loading = ref(false)
const data = ref<MonitorHealthDetail | null>(null)
const err = ref('')

async function load() {
  if (!pageAction('monhealth:view')) return
  loading.value = true
  err.value = ''
  try {
    data.value = await fetchMonitorHealth()
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
          <span>依赖健康</span>
          <el-button v-if="pageAction('monhealth:view')" type="primary" plain @click="load">刷新</el-button>
        </div>
      </template>
      <el-empty v-if="!pageAction('monhealth:view')" description="无查看权限" />
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
        </div>
        <p class="hint">与公开接口 <code>/health</code> 同源检测逻辑；此处需登录与菜单权限。</p>
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
.hint {
  margin-top: 20px;
  font-size: 13px;
  color: var(--el-text-color-secondary);
}
code {
  font-size: 12px;
  padding: 2px 6px;
  border-radius: 4px;
  background: rgba(0, 0, 0, 0.05);
}
</style>
