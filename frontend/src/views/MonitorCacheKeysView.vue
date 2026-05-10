<script setup lang="ts">
defineOptions({ name: 'MonitorCacheKeysView' })
import { computed, onMounted, ref } from 'vue'

import { fetchMonitorCacheKeys } from '@/api/monitor'
import type { MonitorCacheKeyRow } from '@/api/monitor'
import type { TableColumn } from '@/views/components/NeuroAgentListPage.vue'
import NeuroAgentListPage from '@/views/components/NeuroAgentListPage.vue'
import { usePageAction } from '@/composables/usePageAction'

const { pageMenu } = usePageAction()

const loading = ref(false)
const err = ref('')
const pattern = ref('*')
const rows = ref<MonitorCacheKeyRow[]>([])
const cursor = ref(0)
const limit = 50

const hasMore = computed(() => cursor.value !== 0)

function ttlLabel(ttl: number) {
  if (ttl === -2) return '不存在'
  if (ttl === -1) return '无过期'
  if (ttl < 0) return String(ttl)
  return `${ttl}s`
}

async function load(reset: boolean) {
  if (!pageMenu()) return
  loading.value = true
  err.value = ''
  try {
    const cur = reset ? 0 : cursor.value
    const res = await fetchMonitorCacheKeys(cur, limit, pattern.value.trim() || '*')
    if (reset) rows.value = res.items
    else rows.value = [...rows.value, ...res.items]
    cursor.value = res.cursor
  } catch (e) {
    err.value = e instanceof Error ? e.message : '加载失败'
  } finally {
    loading.value = false
  }
}

const columns: TableColumn[] = [
  { key: 'key', title: '键', minWidth: 280 },
  { key: 'ttl', title: 'TTL', width: 120 },
]

onMounted(() => void load(true))
</script>

<template>
  <div class="page">
    <el-empty v-if="!pageMenu()" description="无查看权限" />
    <template v-else>
      <el-alert v-if="err" :title="err" type="error" show-icon :closable="false" class="err-alert" />
      <NeuroAgentListPage
        mode="el-table"
        title="缓存列表"
        :columns="columns"
        :data="rows"
        :loading="loading"
        :show-create="false"
        :show-selection="false"
        :show-pagination="false"
        :has-filters="true"
      >
        <template #toolbar>
          <div class="tools">
            <el-input
              v-model="pattern"
              placeholder="SCAN 匹配，如 session:*"
              clearable
              style="width: 220px"
              @keyup.enter="load(true)"
            />
          </div>
        </template>
        <template #actions>
          <el-button type="primary" plain @click="load(true)">查询</el-button>
        </template>
        <template #col-ttl="{ row }">
          {{ ttlLabel(row.ttl) }}
        </template>
      </NeuroAgentListPage>
      <div v-if="rows.length && hasMore" class="more">
        <el-button :loading="loading" @click="load(false)">加载更多</el-button>
      </div>
      <p class="hint">使用 Redis SCAN 分页遍历，避免阻塞；仅展示键名与 TTL，不读取值。</p>
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
.tools {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  align-items: center;
}
.more {
  margin-top: 16px;
  display: flex;
  justify-content: center;
}
.hint {
  margin-top: 16px;
  font-size: 13px;
  color: var(--el-text-color-secondary);
}
</style>
