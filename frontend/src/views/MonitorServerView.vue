<script setup lang="ts">
defineOptions({ name: 'MonitorServerView' })
import { onMounted, ref } from 'vue'

import { fetchMonitorServer } from '@/api/monitor'
import type { MonitorServerInfo } from '@/api/monitor'
import { usePageAction } from '@/composables/usePageAction'

const { pageAction } = usePageAction()

const loading = ref(false)
const data = ref<MonitorServerInfo | null>(null)
const err = ref('')

async function load() {
  if (!pageAction('monserver:view')) return
  loading.value = true
  err.value = ''
  try {
    data.value = await fetchMonitorServer()
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
          <span>服务器进程</span>
          <el-button v-if="pageAction('monserver:view')" type="primary" plain @click="load">刷新</el-button>
        </div>
      </template>
      <el-empty v-if="!pageAction('monserver:view')" description="无查看权限" />
      <template v-else>
        <el-alert v-if="err" :title="err" type="error" show-icon :closable="false" style="margin-bottom: 16px" />
        <el-descriptions v-if="data" :column="1" border size="default" class="desc">
          <el-descriptions-item label="Python">{{ data.python_version }}</el-descriptions-item>
          <el-descriptions-item label="进程 PID">{{ data.pid }}</el-descriptions-item>
          <el-descriptions-item label="CPU %">
            {{ data.cpu_percent != null && data.cpu_percent !== undefined ? data.cpu_percent.toFixed(1) : '—' }}
          </el-descriptions-item>
          <el-descriptions-item label="内存 (MB)">
            {{ data.memory_mb != null && data.memory_mb !== undefined ? data.memory_mb : '—' }}
          </el-descriptions-item>
          <el-descriptions-item v-if="data.note" label="说明">{{ data.note }}</el-descriptions-item>
        </el-descriptions>
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
.desc {
  max-width: 560px;
}
</style>
