<script setup lang="ts">
defineOptions({ name: 'MonitorJobsView' })
import { ref } from 'vue'

import { fetchMonitorScheduledJobs } from '@/api/monitor'
import type { MonitorJobsData } from '@/api/monitor'
import type { TableColumn } from '@/views/components/NeuroAgentListPage.vue'
import NeuroAgentListPage from '@/views/components/NeuroAgentListPage.vue'
import { usePageAction } from '@/composables/usePageAction'

const { pageAction } = usePageAction()

const loading = ref(false)
const data = ref<MonitorJobsData | null>(null)
const err = ref('')

async function load() {
  if (!pageAction('monjobs:view')) return
  loading.value = true
  err.value = ''
  try {
    data.value = await fetchMonitorScheduledJobs()
  } catch (e) {
    err.value = e instanceof Error ? e.message : '加载失败'
  } finally {
    loading.value = false
  }
}

const columns: TableColumn[] = [
  { key: 'name', title: '名称', minWidth: 200 },
  { key: 'schedule', title: '调度说明', minWidth: 160 },
  { key: 'status', title: '状态', width: 100 },
]
</script>

<template>
  <div class="page">
    <el-empty v-if="!pageAction('monjobs:view')" description="无查看权限" />
    <template v-else>
      <el-alert v-if="err" :title="err" type="error" show-icon :closable="false" class="err-alert" />
      <NeuroAgentListPage
        mode="el-table"
        title="定时任务与周期行为"
        :columns="columns"
        :data="data?.items ?? []"
        :loading="loading"
        :show-create="false"
        :show-selection="false"
        :show-pagination="false"
      >
        <template #actions>
          <el-button type="primary" plain @click="load">刷新</el-button>
        </template>
      </NeuroAgentListPage>
      <p v-if="data?.note" class="hint">{{ data.note }}</p>
    </template>
  </div>
</template>

<style scoped>
.page {
  padding: 0;
}
.err-alert {
  margin: 0 16px 16px;
}
.hint {
  margin: 16px;
  font-size: 13px;
  color: var(--el-text-color-secondary);
  line-height: 1.5;
}
</style>
