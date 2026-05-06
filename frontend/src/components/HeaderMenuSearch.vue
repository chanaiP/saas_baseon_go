<script setup lang="ts">
import { Search } from '@element-plus/icons-vue'
import type { ElAutocomplete } from 'element-plus'
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useRouter } from 'vue-router'

import { usePermissionStore } from '@/stores/permission'
import type { MenuNode } from '@/types/menu'
import { flattenAccessibleMenuItems, type FlatMenuSearchItem } from '@/utils/menuVisibility'

const props = withDefaults(
  defineProps<{
    menuTree: MenuNode[]
    /** toolbar：顶栏圆角；sidebar：侧栏全宽；drawer：抽屉内 */
    variant?: 'toolbar' | 'sidebar' | 'drawer'
  }>(),
  { variant: 'toolbar' },
)

const router = useRouter()
const perm = usePermissionStore()

const RECENT_KEY = 'header_menu_recent_v1'
const MAX_RECENT = 5

function loadRecent(): { path: string; title: string }[] {
  try {
    const raw = localStorage.getItem(RECENT_KEY)
    if (!raw) return []
    const parsed = JSON.parse(raw) as unknown
    if (!Array.isArray(parsed)) return []
    return parsed
      .filter((x): x is { path: string; title: string } => {
        return (
          x != null &&
          typeof x === 'object' &&
          typeof (x as { path?: string }).path === 'string' &&
          typeof (x as { title?: string }).title === 'string'
        )
      })
      .slice(0, MAX_RECENT)
  } catch {
    return []
  }
}

function pushRecent(path: string, title: string) {
  let list = loadRecent().filter((x) => x.path !== path)
  list.unshift({ path, title })
  list = list.slice(0, MAX_RECENT)
  try {
    localStorage.setItem(RECENT_KEY, JSON.stringify(list))
  } catch {
    /* ignore quota */
  }
}

type Suggestion = FlatMenuSearchItem & { value: string }

const flatItems = computed(() =>
  flattenAccessibleMenuItems(
    props.menuTree,
    props.menuTree,
    (code) => perm.canUseAction(code),
    (path) => perm.canUseMenuPath(path),
  ),
)

const query = ref('')
const autocompleteRef = ref<InstanceType<typeof ElAutocomplete> | null>(null)

const placeholderText = computed(() =>
  props.variant === 'toolbar' ? '搜索菜单、页面…' : '搜索页面、菜单…',
)

function toSuggestion(x: FlatMenuSearchItem, groupOverride?: string): Suggestion {
  return {
    ...x,
    value: x.title,
    group: groupOverride ?? x.group,
  }
}

function fetchSuggestions(queryString: string, cb: (rows: Suggestion[]) => void) {
  const all = flatItems.value
  const q = queryString.trim().toLowerCase()

  if (!q) {
    const recent = loadRecent()
    const seen = new Set<string>()
    const recentRows: Suggestion[] = []
    for (const r of recent) {
      if (seen.has(r.path)) continue
      seen.add(r.path)
      const hit = all.find((x) => x.path === r.path)
      recentRows.push(
        toSuggestion(
          hit ?? { path: r.path, title: r.title, group: '' },
          '最近访问',
        ),
      )
    }
    const rest = all.filter((x) => !seen.has(x.path)).slice(0, 12)
    cb([...recentRows, ...rest.map((x) => toSuggestion(x))])
    return
  }

  const filtered = all
    .filter(
      (x) =>
        x.title.toLowerCase().includes(q) ||
        x.group.toLowerCase().includes(q) ||
        x.path.toLowerCase().includes(q),
    )
    .slice(0, 24)
  cb(filtered.map((x) => toSuggestion(x)))
}

function onSelect(item: Suggestion) {
  if (!item.path) return
  void router.push(item.path)
  query.value = ''
  pushRecent(item.path, item.title)
}

function focusSearch(e: KeyboardEvent) {
  if ((e.metaKey || e.ctrlKey) && (e.key === 'k' || e.key === 'K')) {
    const t = e.target as HTMLElement | null
    if (t?.closest?.('input, textarea, [contenteditable=true]')) return
    e.preventDefault()
    autocompleteRef.value?.focus()
  }
}

onMounted(() => {
  window.addEventListener('keydown', focusSearch)
})
onUnmounted(() => {
  window.removeEventListener('keydown', focusSearch)
})
</script>

<template>
  <div class="header-menu-search header-menu-search--tech" :class="[`header-menu-search--${variant}`]">
    <el-autocomplete
      ref="autocompleteRef"
      v-model="query"
      aria-label="搜索菜单"
      :fetch-suggestions="fetchSuggestions"
      :trigger-on-focus="true"
      clearable
      :placeholder="placeholderText"
      value-key="value"
      placement="bottom-start"
      popper-class="header-menu-search-popper"
      @select="onSelect"
    >
      <template #prefix>
        <el-icon class="header-menu-search__icon"><Search /></el-icon>
      </template>
      <template v-if="variant === 'toolbar'" #suffix>
        <span class="header-menu-search__shortcut" title="快捷键：⌘K 或 Ctrl+K">⌘K / Ctrl+K</span>
      </template>
      <template #default="{ item }">
        <div class="header-menu-search__row">
          <span class="header-menu-search__title">{{ item.title }}</span>
          <span v-if="item.group" class="header-menu-search__group">{{ item.group }}</span>
        </div>
      </template>
    </el-autocomplete>
  </div>
</template>

<style scoped>
.header-menu-search {
  min-width: 0;
}
.header-menu-search--toolbar {
  width: min(300px, 100%);
  flex-shrink: 0;
  align-self: center;
}
.header-menu-search--toolbar :deep(.el-input__wrapper) {
  border-radius: 999px;
  min-height: 32px;
  padding-top: 1px;
  padding-bottom: 1px;
  box-shadow: 0 0 0 1px var(--el-border-color-lighter) inset;
  background: var(--el-fill-color-blank);
}
.header-menu-search__shortcut {
  font-size: 11px;
  font-weight: 500;
  color: var(--el-text-color-placeholder);
  white-space: nowrap;
  user-select: none;
  margin-right: 2px;
  letter-spacing: -0.02em;
}
.header-menu-search--toolbar :deep(.el-input__wrapper:hover) {
  box-shadow: 0 0 0 1px var(--el-border-color) inset;
}
.header-menu-search--toolbar :deep(.el-input__wrapper.is-focus) {
  box-shadow:
    0 0 0 1px var(--el-color-primary) inset,
    0 0 24px rgba(13, 148, 136, 0.2);
}

.header-menu-search--tech.header-menu-search--toolbar :deep(.el-input__wrapper) {
  transition:
    box-shadow 0.34s cubic-bezier(0.33, 1, 0.32, 1),
    transform 0.34s cubic-bezier(0.33, 1, 0.32, 1),
    background-color 0.34s cubic-bezier(0.33, 1, 0.32, 1);
}

.header-menu-search--tech.header-menu-search--toolbar :deep(.el-input__wrapper.is-focus) {
  transform: translateY(-1px);
}
.header-menu-search--sidebar,
.header-menu-search--drawer {
  width: 100%;
}
.header-menu-search--sidebar :deep(.el-input__wrapper),
.header-menu-search--drawer :deep(.el-input__wrapper) {
  border-radius: 10px;
  box-shadow: 0 0 0 1px var(--el-border-color-lighter) inset;
  background: var(--el-bg-color);
  transition: box-shadow 0.15s ease, background 0.15s ease;
}
.header-menu-search--sidebar :deep(.el-input__wrapper:hover),
.header-menu-search--drawer :deep(.el-input__wrapper:hover) {
  box-shadow: 0 0 0 1px var(--el-border-color) inset;
}
.header-menu-search--sidebar :deep(.el-input__wrapper.is-focus),
.header-menu-search--drawer :deep(.el-input__wrapper.is-focus) {
  box-shadow:
    0 0 0 1px var(--el-color-primary) inset,
    0 0 20px rgba(13, 148, 136, 0.16);
}
.header-menu-search__icon {
  color: var(--el-text-color-secondary);
}
.header-menu-search__row {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 2px;
  line-height: 1.35;
  padding: 2px 0;
}
.header-menu-search__title {
  font-size: 14px;
  color: var(--el-text-color-primary);
}
.header-menu-search__group {
  font-size: 12px;
  color: var(--el-text-color-secondary);
}
</style>

<style>
.header-menu-search-popper.el-autocomplete__popper.el-popper {
  max-width: min(360px, 92vw);
}
</style>
