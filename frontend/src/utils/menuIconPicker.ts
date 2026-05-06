import * as ElementPlusIcons from '@element-plus/icons-vue'
import type { Component } from 'vue'

/** 与侧栏一致：来自 @element-plus/icons-vue 的 PascalCase 组件名 */
export const MENU_PICKER_ICON_NAMES: readonly string[] = Object.keys(ElementPlusIcons)
  .filter((k) => k !== 'default' && /^[A-Z]/.test(k))
  .sort((a, b) => a.localeCompare(b))

export function menuIconComponent(name?: string | null): Component | null {
  if (!name) return null
  const c = (ElementPlusIcons as Record<string, Component>)[name]
  return c ?? null
}
