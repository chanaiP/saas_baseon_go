<script setup lang="ts">
import { RouterView } from 'vue-router'

import WatermarkLayer from '@/components/WatermarkLayer.vue'

/** 表格单元格默认单行省略 + Tooltip，避免列内换行；需在自定义列（操作/开关/标签等）上单独 :show-overflow-tooltip="false" */
const tableConfig = { showOverflowTooltip: true as const }
const elementPlusZIndex = 12000
</script>

<template>
  <div class="app-viewport">
    <!-- ElConfigProvider 根节点非单一 DOM，不能挂 class；布局类放在内部包裹层 -->
    <el-config-provider :table="tableConfig" :z-index="elementPlusZIndex">
      <div class="app-config-root">
        <RouterView />
        <WatermarkLayer />
      </div>
    </el-config-provider>
  </div>
</template>

<style scoped>
/* 与 #app 一致：纯块级，高度完全由路由根组件（如 neuro-command-layout）决定，避免 flex 拉伸链 */
.app-viewport,
.app-config-root {
  display: block;
  width: 100%;
  box-sizing: border-box;
}
</style>
