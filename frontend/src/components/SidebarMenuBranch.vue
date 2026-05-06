<script setup lang="ts">
import { computed } from 'vue'

import { usePermissionStore } from '@/stores/permission'
import type { MenuNode } from '@/types/menu'
import { menuIconComponent } from '@/utils/menuIconPicker'
import { filterVisibleMenuTree } from '@/utils/menuVisibility'

const props = defineProps<{
  nodes: MenuNode[]
}>()

const perm = usePermissionStore()

const visibleNodes = computed(() => {
  if (!perm.loaded) return props.nodes
  return filterVisibleMenuTree(props.nodes, (code) => perm.canUseAction(code), (path) => perm.canUseMenuPath(path))
})
</script>

<template>
  <template v-for="node in visibleNodes" :key="node.id">
    <el-sub-menu v-if="node.type === 'directory'" :index="node.id">
      <template #title>
        <el-icon v-if="menuIconComponent(node.icon)"><component :is="menuIconComponent(node.icon)!" /></el-icon>
        <span>{{ node.title }}</span>
      </template>
      <template v-for="child in node.children || []" :key="child.id">
        <el-menu-item v-if="child.type === 'menu' && child.path" :index="child.path">
          <el-icon v-if="menuIconComponent(child.icon)"><component :is="menuIconComponent(child.icon)!" /></el-icon>
          <template #title>{{ child.title }}</template>
        </el-menu-item>
      </template>
    </el-sub-menu>
    <el-menu-item v-else-if="node.type === 'menu' && node.path" :index="node.path">
      <el-icon v-if="menuIconComponent(node.icon)"><component :is="menuIconComponent(node.icon)!" /></el-icon>
      <span>{{ node.title }}</span>
    </el-menu-item>
  </template>
</template>
