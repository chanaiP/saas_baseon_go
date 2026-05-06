<script setup lang="ts">
import { ElIcon } from 'element-plus'
import { computed } from 'vue'

import { MENU_PICKER_ICON_NAMES, menuIconComponent } from '@/utils/menuIconPicker'

const model = defineModel<string>({ default: '' })

const props = withDefaults(
  defineProps<{
    placeholder?: string
    disabled?: boolean
    clearable?: boolean
  }>(),
  { placeholder: '选择图标', disabled: false, clearable: true },
)

const currentComp = computed(() => menuIconComponent(model.value))
</script>

<template>
  <el-select
    v-model="model"
    class="menu-icon-select"
    filterable
    :clearable="props.clearable"
    :disabled="props.disabled"
    :placeholder="props.placeholder"
  >
    <template #prefix>
      <span v-if="currentComp" class="menu-icon-select__prefix">
        <el-icon><component :is="currentComp" /></el-icon>
      </span>
    </template>
    <el-option v-for="name in MENU_PICKER_ICON_NAMES" :key="name" :label="name" :value="name">
      <span class="menu-icon-select__option">
        <el-icon v-if="menuIconComponent(name)"><component :is="menuIconComponent(name)!" /></el-icon>
        <span class="menu-icon-select__option-name">{{ name }}</span>
      </span>
    </el-option>
  </el-select>
</template>

<style scoped>
.menu-icon-select {
  width: 100%;
}
.menu-icon-select__prefix {
  display: inline-flex;
  align-items: center;
  margin-right: 4px;
  color: var(--el-text-color-regular);
}
.menu-icon-select__option {
  display: inline-flex;
  align-items: center;
  gap: 8px;
}
.menu-icon-select__option-name {
  font-size: 13px;
  color: var(--el-text-color-primary);
}
</style>
