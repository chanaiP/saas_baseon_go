<script setup lang="ts">
import { ArrowDown, Close, DArrowLeft, DArrowRight, RefreshRight } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { useTabsStore } from '@/stores/tabs'

type TabMenuCmd = 'refresh' | 'left' | 'right' | 'others' | 'all'

const tabs = useTabsStore()
const route = useRoute()
const router = useRouter()

const currentIndex = computed(() => tabs.visited.findIndex((t) => t.path === route.path))

const canCloseLeft = computed(() => currentIndex.value > 0)
const canCloseRight = computed(
  () => currentIndex.value >= 0 && currentIndex.value < tabs.visited.length - 1,
)

function isActive(path: string) {
  return route.path === path
}

function onTabClick(path: string, fullPath: string) {
  if (route.path !== path) void router.push(fullPath)
}

function onCloseOthers() {
  tabs.closeOthers(route.path)
}

function onCloseAll() {
  tabs.closeAll()
}

function onTabMenu(cmd: TabMenuCmd) {
  const p = route.path
  switch (cmd) {
    case 'refresh':
      tabs.refreshCurrentTab()
      ElMessage.success('已刷新当前页')
      break
    case 'left':
      tabs.closeLeft(p)
      break
    case 'right':
      tabs.closeRight(p)
      break
    case 'others':
      onCloseOthers()
      break
    case 'all':
      onCloseAll()
      break
    default:
      break
  }
}
</script>

<template>
  <div v-if="tabs.visited.length" class="tabs-wrap">
    <div class="tabs-scroll">
      <button
        v-for="t in tabs.visited"
        :key="t.path"
        type="button"
        class="tab"
        :class="{ active: isActive(t.path) }"
        @click="onTabClick(t.path, t.fullPath)"
      >
        <span class="tab-title">{{ t.title }}</span>
        <span
          class="tab-close"
          role="button"
          tabindex="0"
          title="关闭"
          @click.stop="tabs.closeTab(t.path)"
          @keydown.enter.prevent.stop="tabs.closeTab(t.path)"
        >
          <el-icon :size="12"><Close /></el-icon>
        </span>
      </button>
    </div>
    <el-dropdown trigger="click" @command="(c: TabMenuCmd) => onTabMenu(c)">
      <el-button class="tabs-more" size="small" text type="primary">
        页签操作
        <el-icon class="el-icon--right"><ArrowDown /></el-icon>
      </el-button>
      <template #dropdown>
        <el-dropdown-menu class="tabs-dropdown-menu">
          <el-dropdown-item command="refresh">
            <el-icon class="tabs-item-icon"><RefreshRight /></el-icon>
            刷新当前页
          </el-dropdown-item>
          <el-dropdown-item command="left" :disabled="!canCloseLeft">
            <el-icon class="tabs-item-icon"><DArrowLeft /></el-icon>
            关闭左侧页签
          </el-dropdown-item>
          <el-dropdown-item command="right" :disabled="!canCloseRight">
            <el-icon class="tabs-item-icon"><DArrowRight /></el-icon>
            关闭右侧页签
          </el-dropdown-item>
          <el-dropdown-item divided command="others" :disabled="tabs.visited.length <= 1">
            关闭其他
          </el-dropdown-item>
          <el-dropdown-item command="all">关闭全部</el-dropdown-item>
        </el-dropdown-menu>
      </template>
    </el-dropdown>
  </div>
</template>

<style scoped>
.tabs-wrap {
  display: flex;
  align-items: stretch;
  gap: 0;
  flex-shrink: 0;
  background: transparent;
  padding: 8px 16px 0 24px;
  min-height: 44px;
  transition: background-color var(--shell-t-slow, 0.35s) var(--shell-ease-standard, cubic-bezier(0.2, 0, 0, 1));
}
.tabs-scroll {
  display: flex;
  align-items: flex-end;
  gap: 2px;
  flex: 1;
  min-width: 0;
  overflow-x: auto;
  overflow-y: hidden;
  scrollbar-width: thin;
}
.tabs-more {
  flex-shrink: 0;
  align-self: center;
  margin-left: 4px;
}
.tabs-item-icon {
  margin-right: 6px;
  vertical-align: middle;
}
.tab {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  max-width: 200px;
  padding: 9px 14px 10px;
  margin-bottom: -1px;
  border: none;
  border-bottom: 2px solid transparent;
  border-radius: 8px 8px 0 0;
  background: transparent;
  color: var(--el-text-color-regular);
  font-size: 13px;
  font-weight: 500;
  letter-spacing: 0.01em;
  cursor: pointer;
  white-space: nowrap;
  flex-shrink: 0;
  transition:
    color var(--shell-t-med, 0.24s) var(--shell-ease-standard, cubic-bezier(0.2, 0, 0, 1)),
    border-color var(--shell-t-slow, 0.35s) var(--shell-ease-emphasized, cubic-bezier(0, 0, 0, 1)),
    background-color var(--shell-t-med, 0.24s) var(--shell-ease-standard, cubic-bezier(0.2, 0, 0, 1));
}
.tab:hover {
  color: var(--el-color-primary);
  background: color-mix(in srgb, var(--el-fill-color-light) 88%, transparent);
}
.tab.active {
  color: var(--el-color-primary);
  border-bottom-color: var(--el-color-primary);
  font-weight: 600;
  background: color-mix(in srgb, var(--el-fill-color-light) 72%, transparent);
  box-shadow: none;
}

.dark .tab.active {
  box-shadow: none;
}
.tab-title {
  overflow: hidden;
  text-overflow: ellipsis;
}
.tab-close {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 2px;
  border-radius: 2px;
  color: var(--el-text-color-secondary);
}
.tab-close:hover {
  color: var(--el-color-danger);
  background: var(--el-fill-color);
}
</style>

<style>
/* 下拉项内图标与文字对齐（下拉挂载到 body 时不受 scoped 影响） */
.tabs-dropdown-menu .el-dropdown-menu__item {
  display: flex;
  align-items: center;
}
</style>
