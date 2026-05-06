<script setup lang="ts">
import { CopyDocument } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { ref, watch } from 'vue'

import { PRIMARY_PRESETS, type PrimaryPresetId } from '@/constants/layoutPresets'
import type { ContentWidthMode } from '@/stores/uiPreferences'
import { useUiPreferencesStore } from '@/stores/uiPreferences'

const props = defineProps<{ modelValue: boolean }>()
const emit = defineEmits<{ 'update:modelValue': [v: boolean] }>()

const ui = useUiPreferencesStore()
const wmDraft = ref('')

watch(
  () => props.modelValue,
  (open) => {
    if (open) wmDraft.value = ui.watermarkText
  },
)

function saveWatermark() {
  ui.setWatermark(wmDraft.value)
  ElMessage.success('水印已保存')
}

function clearWatermark() {
  wmDraft.value = ''
  ui.setWatermark('')
  ElMessage.success('已清除水印')
}

async function copySettings() {
  try {
    const text = JSON.stringify(ui.layoutSettingsSnapshot(), null, 2)
    await navigator.clipboard.writeText(text)
    ElMessage.success('已复制到剪贴板')
  } catch {
    ElMessage.error('复制失败，请手动选择文本')
  }
}

const isDev = import.meta.env.DEV
</script>

<template>
  <el-drawer
    :model-value="modelValue"
    title="界面设置"
    direction="rtl"
    size="320px"
    append-to-body
    destroy-on-close
    class="layout-settings-drawer"
    @update:model-value="emit('update:modelValue', $event)"
  >
    <div class="ls-body">
      <section class="ls-section">
        <div class="ls-label">主题色</div>
        <div class="ls-colors">
          <button
            v-for="p in PRIMARY_PRESETS"
            :key="p.id"
            type="button"
            class="ls-color-swatch"
            :class="{ 'is-active': ui.primaryPreset === p.id }"
            :style="{ background: p.color }"
            :title="p.label"
            @click="ui.setPrimaryPreset(p.id as PrimaryPresetId)"
          >
            <span v-if="ui.primaryPreset === p.id" class="ls-check">✓</span>
          </button>
        </div>
      </section>

      <section class="ls-section">
        <div class="ls-label">导航模式</div>
        <div class="ls-nav-modes">
          <button
            type="button"
            class="ls-nav-card"
            :class="{ 'is-active': ui.navMode === 'side' }"
            @click="ui.setNavMode('side')"
          >
            <span class="ls-nav-ico ls-nav-ico--side" aria-hidden="true" />
            <span class="ls-nav-cap">侧栏</span>
          </button>
          <button
            type="button"
            class="ls-nav-card"
            :class="{ 'is-active': ui.navMode === 'top' }"
            @click="ui.setNavMode('top')"
          >
            <span class="ls-nav-ico ls-nav-ico--top" aria-hidden="true" />
            <span class="ls-nav-cap">顶栏</span>
          </button>
          <button
            type="button"
            class="ls-nav-card"
            :class="{ 'is-active': ui.navMode === 'mix' }"
            @click="ui.setNavMode('mix')"
          >
            <span class="ls-nav-ico ls-nav-ico--mix" aria-hidden="true" />
            <span class="ls-nav-cap">混合</span>
          </button>
        </div>
        <p class="ls-hint">混合：顶栏切换一级模块，侧栏仅显示当前模块下菜单。</p>
      </section>

      <section class="ls-section">
        <div class="ls-label">侧边菜单风格</div>
        <div class="ls-nav-modes">
          <button
            type="button"
            class="ls-nav-card ls-nav-card--sm"
            :class="{ 'is-active': ui.sidebarStyle === 'default' }"
            @click="ui.setSidebarStyle('default')"
          >
            <span class="ls-side-ico ls-side-ico--dark" aria-hidden="true" />
            <span class="ls-nav-cap">深色</span>
          </button>
          <button
            type="button"
            class="ls-nav-card ls-nav-card--sm"
            :class="{ 'is-active': ui.sidebarStyle === 'light' }"
            @click="ui.setSidebarStyle('light')"
          >
            <span class="ls-side-ico ls-side-ico--light" aria-hidden="true" />
            <span class="ls-nav-cap">浅色</span>
          </button>
        </div>
      </section>

      <section class="ls-section">
        <div class="ls-label">内容区域宽度</div>
        <el-select :model-value="ui.contentWidth" class="ls-select" @update:model-value="ui.setContentWidth($event as ContentWidthMode)">
          <el-option label="流式" value="fluid" />
          <el-option label="定宽（居中）" value="fixed" />
        </el-select>
      </section>

      <section class="ls-section ls-toggles">
        <div class="ls-row">
          <span>固定 Header</span>
          <el-switch v-model="ui.fixedHeader" />
        </div>
        <div class="ls-row">
          <span>固定侧边菜单</span>
          <el-switch v-model="ui.fixedSidebar" />
        </div>
        <div class="ls-row">
          <span>自动分割菜单</span>
          <el-switch v-model="ui.autoSplitMenu" />
        </div>
        <p class="ls-hint">开启后允许多个侧栏子菜单同时展开。</p>
      </section>

      <section class="ls-section ls-toggles">
        <div class="ls-label">内容区域</div>
        <div class="ls-row">
          <span>顶栏（菜单搜索）</span>
          <el-switch v-model="ui.showTopbar" />
        </div>
        <div class="ls-row">
          <span>页脚</span>
          <el-switch v-model="ui.showFooter" />
        </div>
        <div class="ls-row">
          <span>侧栏菜单</span>
          <el-switch v-model="ui.showMenu" />
        </div>
        <div class="ls-row">
          <span>侧栏品牌区</span>
          <el-switch v-model="ui.showMenuHeader" />
        </div>
      </section>

      <section class="ls-section ls-toggles">
        <div class="ls-row">
          <span>色弱模式</span>
          <el-switch v-model="ui.colorWeak" />
        </div>
      </section>

      <section class="ls-section">
        <div class="ls-label">全局水印</div>
        <p class="ls-desc">斜向平铺；留空并保存可关闭。与顶栏水印入口合并至此。</p>
        <el-input v-model="wmDraft" type="textarea" :rows="2" maxlength="48" show-word-limit placeholder="例如：内部资料 禁止外传" />
        <div class="ls-wm-actions">
          <el-button size="small" @click="clearWatermark">清除</el-button>
          <el-button size="small" type="primary" @click="saveWatermark">保存水印</el-button>
        </div>
      </section>

      <el-alert
        v-if="isDev"
        type="warning"
        show-icon
        :closable="false"
        class="ls-dev-alert"
        title="配置栏仅在开发环境用于预览时提示；生产环境请按需将拷贝的 JSON 合并到业务配置或持久化策略中。"
      />

      <el-button class="ls-copy-btn" type="primary" plain block :icon="CopyDocument" @click="copySettings">拷贝设置</el-button>
    </div>
  </el-drawer>
</template>

<style scoped>
.layout-settings-drawer :deep(.el-drawer__body) {
  padding: 0 16px 24px;
}
.ls-body {
  display: flex;
  flex-direction: column;
  gap: 20px;
}
.ls-section {
  padding-bottom: 16px;
  border-bottom: 1px solid var(--el-border-color-lighter);
}
.ls-section:last-of-type {
  border-bottom: none;
}
.ls-label {
  font-size: 13px;
  font-weight: 600;
  color: var(--el-text-color-primary);
  margin-bottom: 10px;
}
.ls-colors {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}
.ls-color-swatch {
  width: 28px;
  height: 28px;
  border-radius: 6px;
  border: 2px solid transparent;
  cursor: pointer;
  padding: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  box-sizing: border-box;
}
.ls-color-swatch.is-active {
  border-color: var(--el-text-color-primary);
  box-shadow: 0 0 0 1px var(--el-bg-color);
}
.ls-check {
  color: #fff;
  font-size: 14px;
  font-weight: 700;
  text-shadow: 0 0 2px rgba(0, 0, 0, 0.5);
}
.ls-nav-modes {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
}
.ls-nav-card {
  flex: 1;
  min-width: 72px;
  max-width: 96px;
  padding: 10px 8px;
  border-radius: 8px;
  border: 1px solid var(--el-border-color);
  background: var(--el-fill-color-blank);
  cursor: pointer;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
  transition:
    border-color 0.15s,
    background 0.15s;
}
.ls-nav-card--sm {
  max-width: 120px;
}
.ls-nav-card:hover {
  border-color: var(--el-color-primary-light-5);
}
.ls-nav-card.is-active {
  border-color: var(--el-color-primary);
  background: var(--el-color-primary-light-9);
}
.ls-nav-ico {
  display: block;
  width: 28px;
  height: 22px;
  border: 2px solid currentColor;
  border-radius: 2px;
  color: var(--el-text-color-regular);
  box-sizing: border-box;
}
.ls-nav-card.is-active .ls-nav-ico {
  color: var(--el-color-primary);
}
.ls-nav-ico--side {
  box-shadow: inset 6px 0 0 currentColor;
}
.ls-nav-ico--top {
  box-shadow: inset 0 6px 0 currentColor;
}
.ls-nav-ico--mix {
  box-shadow:
    inset 6px 0 0 currentColor,
    inset 0 6px 0 rgba(64, 158, 255, 0.35);
}
.ls-side-ico {
  display: block;
  width: 32px;
  height: 22px;
  border-radius: 4px;
  border: 1px solid var(--el-border-color);
}
.ls-side-ico--dark {
  background: linear-gradient(90deg, #2c2c2e 40%, #3a3a3c 40%);
}
.ls-side-ico--light {
  background: linear-gradient(90deg, #f5f5f7 40%, #ffffff 40%);
}
.ls-nav-cap {
  font-size: 11px;
  color: var(--el-text-color-secondary);
}
.ls-nav-card.is-active .ls-nav-cap {
  color: var(--el-color-primary);
}
.ls-hint {
  margin: 8px 0 0;
  font-size: 12px;
  color: var(--el-text-color-secondary);
  line-height: 1.4;
}
.ls-select {
  width: 100%;
}
.ls-toggles .ls-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 6px 0;
  font-size: 13px;
  color: var(--el-text-color-regular);
}
.ls-desc {
  margin: 0 0 8px;
  font-size: 12px;
  color: var(--el-text-color-secondary);
  line-height: 1.4;
}
.ls-wm-actions {
  margin-top: 10px;
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}
.ls-dev-alert {
  margin-top: 4px;
}
.ls-copy-btn {
  margin-top: 8px;
}
</style>
