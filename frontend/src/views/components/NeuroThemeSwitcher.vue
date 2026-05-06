<template>
  <div class="neuro-theme-switcher" :data-theme="theme.config.mode">
    <!-- 主题切换面板 -->
    <div class="theme-switcher-panel">
      <!-- 标题 -->
      <div class="panel-header">
        <div class="header-left">
          <span class="header-icon">🎨</span>
          <span class="header-title">神经主题控制中心</span>
        </div>
        <button class="header-close" @click="$emit('close')" title="关闭">
          <span class="close-icon">×</span>
        </button>
      </div>

      <!-- 主题模式选择 -->
      <div class="theme-section">
        <h3 class="section-title">
          <span class="title-icon">🌓</span>
          <span class="title-text">主题模式</span>
        </h3>
        <div class="theme-modes">
          <button
            v-for="mode in availableModes"
            :key="mode"
            class="theme-mode-btn"
            :class="{
              active: theme.config.mode === mode,
              'mode-dark': mode === 'dark',
              'mode-light': mode === 'light',
              'mode-cyber': mode === 'cyber',
              'mode-neural': mode === 'neural',
              'mode-matrix': mode === 'matrix'
            }"
            @click="theme.setMode(mode)"
            :title="getModeName(mode)"
          >
            <span class="mode-icon">{{ getModeIcon(mode) }}</span>
            <span class="mode-name">{{ getModeName(mode) }}</span>
            <span class="mode-preview" :style="getModePreviewStyle(mode)"></span>
          </button>
        </div>
      </div>

      <!-- 主题颜色选择 -->
      <div class="theme-section">
        <h3 class="section-title">
          <span class="title-icon">🎨</span>
          <span class="title-text">主题颜色</span>
        </h3>
        <div class="theme-colors">
          <button
            v-for="color in availableColors"
            :key="color"
            class="theme-color-btn"
            :class="{
              active: theme.config.color === color,
              'color-blue': color === 'blue',
              'color-purple': color === 'purple',
              'color-green': color === 'green',
              'color-orange': color === 'orange',
              'color-pink': color === 'pink'
            }"
            @click="theme.setColor(color)"
            :title="getColorName(color)"
          >
            <span class="color-preview" :style="{ backgroundColor: getColorValue(color) }"></span>
            <span class="color-name">{{ getColorName(color) }}</span>
          </button>
        </div>
      </div>

      <!-- 特效设置 -->
      <div class="theme-section">
        <h3 class="section-title">
          <span class="title-icon">✨</span>
          <span class="title-text">神经特效</span>
        </h3>
        <div class="theme-effects">
          <div class="effect-item">
            <label class="effect-label">
              <input
                type="checkbox"
                v-model="animationsEnabled"
                @change="theme.toggleAnimations(animationsEnabled)"
                class="effect-checkbox"
              />
              <span class="effect-icon">🌀</span>
              <span class="effect-text">启用动画效果</span>
            </label>
            <span class="effect-status">{{ animationsEnabled ? '已启用' : '已禁用' }}</span>
          </div>

          <div class="effect-item">
            <label class="effect-label">
              <input
                type="checkbox"
                v-model="neuralEffectsEnabled"
                @change="theme.toggleNeuralEffects(neuralEffectsEnabled)"
                class="effect-checkbox"
              />
              <span class="effect-icon">🧠</span>
              <span class="effect-text">启用神经特效</span>
            </label>
            <span class="effect-status">{{ neuralEffectsEnabled ? '已启用' : '已禁用' }}</span>
          </div>

          <div class="effect-item">
            <label class="effect-label">
              <input
                type="checkbox"
                v-model="highContrastEnabled"
                @change="theme.toggleHighContrast(highContrastEnabled)"
                class="effect-checkbox"
              />
              <span class="effect-icon">🔍</span>
              <span class="effect-text">高对比度模式</span>
            </label>
            <span class="effect-status">{{ highContrastEnabled ? '已启用' : '已禁用' }}</span>
          </div>
        </div>
      </div>

      <!-- 当前主题预览 -->
      <div class="theme-section">
        <h3 class="section-title">
          <span class="title-icon">👁️</span>
          <span class="title-text">当前主题预览</span>
        </h3>
        <div class="theme-preview">
          <div class="preview-card" :style="previewCardStyle">
            <div class="preview-header">
              <span class="preview-title">神经主题</span>
              <span class="preview-subtitle">{{ currentThemeName }}</span>
            </div>
            <div class="preview-content">
              <div class="preview-colors">
                <div
                  v-for="(colorValue, colorName) in previewColors"
                  :key="colorName"
                  class="preview-color-item"
                >
                  <span class="color-sample" :style="{ backgroundColor: colorValue }"></span>
                  <span class="color-label">{{ colorName }}</span>
                  <span class="color-value">{{ colorValue }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- 操作按钮 -->
      <div class="theme-actions">
        <button class="action-btn action-reset" @click="theme.resetToDefault">
          <span class="action-icon">🔄</span>
          <span class="action-text">重置为默认</span>
        </button>
        <button class="action-btn action-apply" @click="applyTheme">
          <span class="action-icon">✅</span>
          <span class="action-text">应用主题</span>
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
// Theme types - simplified for showcase
type ThemeMode = 'dark' | 'light' | 'cyber' | 'neural' | 'matrix'
type ThemeColor = string
// Stub theme manager for showcase
const theme = {
  config: { mode: 'dark' as ThemeMode, color: '#00f5d4' as ThemeColor, animations: true, neuralEffects: true, highContrast: false },
  colors: { primary: '#00f5d4', secondary: '#9d4edd', accent: '#ff6b6b', background: '#0a0c14', surface: '#111827', border: '#2d3748', text: '#f8fafc' },
  getAvailableModes: () => ['dark', 'light', 'cyber', 'neural', 'matrix'] as ThemeMode[],
  getAvailableColors: () => ['#00f5d4', '#9d4edd', '#ff6b6b', '#3b82f6', '#f59e0b'] as ThemeColor[],
  setMode: (_m: ThemeMode) => {},
  setColor: (_c: ThemeColor) => {},
  toggleAnimations: (_v: boolean) => {},
  toggleNeuralEffects: (_v: boolean) => {},
  toggleHighContrast: (_v: boolean) => {},
  resetToDefault: () => {},
  getThemePreview: (_m: ThemeMode, _c: string) => ({ gradient: 'linear-gradient(135deg, #00f5d4, #9d4edd)', colors: { primary: '#00f5d4', accent: '#ff6b6b' } }),
}

const emit = defineEmits<{
  close: []
}>()

// 可用选项
const availableModes = theme.getAvailableModes()
const availableColors = theme.getAvailableColors()

// 特效状态
const animationsEnabled = ref(theme.config.animations)
const neuralEffectsEnabled = ref(theme.config.neuralEffects)
const highContrastEnabled = ref(theme.config.highContrast)

// 获取模式名称
const getModeName = (mode: ThemeMode): string => {
  const names: Record<ThemeMode, string> = {
    dark: '暗黑模式',
    light: '明亮模式',
    cyber: '赛博朋克',
    neural: '神经美学',
    matrix: '数字矩阵'
  }
  return names[mode]
}

// 获取模式图标
const getModeIcon = (mode: ThemeMode): string => {
  const icons: Record<ThemeMode, string> = {
    dark: '🌙',
    light: '☀️',
    cyber: '🤖',
    neural: '🧠',
    matrix: '💻'
  }
  return icons[mode]
}

// 获取模式预览样式
const getModePreviewStyle = (mode: ThemeMode) => {
  const preview = theme.getThemePreview(mode, 'blue')
  return {
    background: `linear-gradient(135deg, ${preview.colors.primary}, ${preview.colors.accent})`
  }
}

// 获取颜色名称
const getColorName = (color: ThemeColor): string => {
  const names: Record<ThemeColor, string> = {
    blue: '蓝色',
    purple: '紫色',
    green: '绿色',
    orange: '橙色',
    pink: '粉色'
  }
  return names[color]
}

// 获取颜色值
const getColorValue = (color: ThemeColor): string => {
  const preview = theme.getThemePreview('dark', color)
  return preview.colors.primary
}

// 当前主题名称
const currentThemeName = computed(() => {
  const modeName = getModeName(theme.config.mode)
  const colorName = getColorName(theme.config.color)
  return `${modeName} · ${colorName}`
})

// 预览卡片样式
const previewCardStyle = computed(() => ({
  background: theme.colors.surface,
  borderColor: theme.colors.border,
  color: theme.colors.text
}))

// 预览颜色
const previewColors = computed(() => ({
  主色: theme.colors.primary,
  辅色: theme.colors.secondary,
  强调色: theme.colors.accent,
  背景: theme.colors.background,
  表面: theme.colors.surface,
  文字: theme.colors.text
}))

// 应用主题
const applyTheme = () => {
  // 主题已经通过响应式更新自动应用
  emit('close')
}
</script>

<style scoped>
@import '@/styles/theme/neuro-theme.css';
.neuro-theme-switcher {
  position: fixed;
  top: 0;
  right: 0;
  bottom: 0;
  width: 400px;
  background: var(--neuro-surface);
  border-left: 1px solid var(--neuro-border);
  box-shadow: var(--neuro-shadow-lg);
  z-index: 9999;
  animation: theme-slide-in 0.3s ease-out;
}

@keyframes theme-slide-in {
  from {
    transform: translateX(100%);
    opacity: 0;
  }
  to {
    transform: translateX(0);
    opacity: 1;
  }
}

.theme-switcher-panel {
  height: 100%;
  display: flex;
  flex-direction: column;
  overflow-y: auto;
}

.panel-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 20px;
  background: var(--neuro-background);
  border-bottom: 1px solid var(--neuro-border);
}

.header-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.header-icon {
  font-size: 24px;
}

.header-title {
  font-size: 18px;
  font-weight: 600;
  background: var(--neuro-gradient-primary);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.header-close {
  width: 32px;
  height: 32px;
  border-radius: var(--neuro-radius-md);
  background: var(--neuro-surface);
  border: 1px solid var(--neuro-border);
  color: var(--neuro-text);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
}

.header-close:hover {
  background: var(--neuro-error);
  border-color: var(--neuro-error);
  color: var(--neuro-text);
  transform: scale(1.1);
}

.close-icon {
  font-size: 20px;
  font-weight: bold;
}

.theme-section {
  padding: 20px;
  border-bottom: 1px solid var(--neuro-border);
}

.section-title {
  display: flex;
  align-items: center;
  gap: 10px;
  margin: 0 0 16px 0;
  color: var(--neuro-text);
}

.title-icon {
  font-size: 20px;
}

.title-text {
  font-size: 16px;
  font-weight: 600;
}

.theme-modes {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 12px;
}

.theme-mode-btn {
  padding: var(--neuro-spacing-md);
  border-radius: var(--neuro-radius-lg);
  background: var(--neuro-surface);
  border: 2px solid var(--neuro-border);
  color: var(--neuro-text);
  cursor: pointer;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  transition: all var(--neuro-transition-normal);
  position: relative;
  overflow: hidden;
}

.theme-mode-btn:hover {
  transform: translateY(-2px);
  box-shadow: var(--neuro-shadow-md);
  border-color: var(--neuro-primary);
}

.theme-mode-btn.active {
  border-color: var(--neuro-primary);
  background: var(--neuro-primary-10);
  box-shadow: var(--neuro-shadow-lg);
}

.mode-icon {
  font-size: 24px;
  margin-bottom: 4px;
}

.mode-name {
  font-size: 14px;
  font-weight: 500;
}

.mode-preview {
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  height: 4px;
  opacity: 0.6;
}

/* 模式特定样式 */
.mode-dark .mode-preview {
  background: linear-gradient(90deg, var(--neuro-surface), #475569);
}

.mode-light .mode-preview {
  background: linear-gradient(90deg, #f1f5f9, #cbd5e1);
}

.mode-cyber .mode-preview {
  background: linear-gradient(90deg, var(--neuro-primary), #ff00ff);
}

.mode-neural .mode-preview {
  background: linear-gradient(90deg, var(--neuro-primary), #8b5cf6);
}

.mode-matrix .mode-preview {
  background: linear-gradient(90deg, #00ff41, #000000);
}

.theme-colors {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 12px;
}

.theme-color-btn {
  padding: 12px;
  border-radius: 10px;
  background: var(--neuro-surface);
  border: 2px solid var(--neuro-border);
  color: var(--neuro-text);
  cursor: pointer;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  transition: all var(--neuro-transition-normal);
}

.theme-color-btn:hover {
  transform: translateY(-2px);
  box-shadow: var(--neuro-shadow-md);
}

.theme-color-btn.active {
  border-color: var(--neuro-primary);
  background: var(--neuro-primary-10);
  box-shadow: var(--neuro-shadow-lg);
}

.color-preview {
  width: 40px;
  height: 40px;
  border-radius: var(--neuro-radius-md);
  border: 2px solid var(--neuro-border);
}

.color-name {
  font-size: 12px;
  font-weight: 500;
}

.theme-effects {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.effect-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px;
  background: var(--neuro-surface);
  border: 1px solid var(--neuro-border);
  border-radius: 10px;
  transition: all 0.2s;
}

.effect-item:hover {
  border-color: var(--neuro-primary);
  background: var(--neuro-primary-10);
}

.effect-label {
  display: flex;
  align-items: center;
  gap: 12px;
  cursor: pointer;
  flex: 1;
}

.effect-checkbox {
  width: 18px;
  height: 18px;
  border-radius: var(--neuro-radius-sm);
  border: 2px solid var(--neuro-border);
  background: var(--neuro-surface);
  cursor: pointer;
  appearance: none;
  position: relative;
}

.effect-checkbox:checked {
  background: var(--neuro-primary);
  border-color: var(--neuro-primary);
}

.effect-checkbox:checked::after {
  content: '✓';
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  color: var(--neuro-text);
  font-size: 12px;
  font-weight: bold;
}

.effect-icon {
  font-size: 18px;
}

.effect-text {
  font-size: 14px;
  color: var(--neuro-text);
}

.effect-status {
  font-size: 12px;
  padding: 4px 8px;
  background: var(--neuro-surface);
  border: 1px solid var(--neuro-border);
  border-radius: 6px;
  color: var(--neuro-text-secondary);
}

.theme-preview {
  margin-top: 12px;
}

.preview-card {
  padding: 20px;
  border-radius: var(--neuro-radius-lg);
  border: 1px solid;
  box-shadow: var(--neuro-shadow-md);
}

.preview-header {
  margin-bottom: 16px;
  padding-bottom: 12px;
  border-bottom: 1px solid var(--neuro-border);
}

.preview-title {
  display: block;
  font-size: 18px;
  font-weight: 600;
  background: var(--neuro-gradient-primary);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  margin-bottom: 4px;
}

.preview-subtitle {
  font-size: 14px;
  color: var(--neuro-text-secondary);
}

.preview-colors {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 12px;
}

.preview-color-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: var(--neuro-spacing-sm);
  background: var(--neuro-background);
  border-radius: var(--neuro-radius-md);
}

.color-sample {
  width: 24px;
  height: 24px;
  border-radius: var(--neuro-radius-sm);
  border: 1px solid var(--neuro-border);
}

.color-label {
  font-size: 12px;
  color: var(--neuro-text);
  flex: 1;
}

.color-value {
  font-size: 10px;
  font-family: monospace;
  color: var(--neuro-text-secondary);
}

.theme-actions {
  padding: 20px;
  display: flex;
  gap: 12px;
  background: var(--neuro-background);
  border-top: 1px solid var(--neuro-border);
  margin-top: auto;
}

.action-btn {
  flex: 1;
  padding: 14px;
  border-radius: 10px;
  border: none;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  font-size: 14px;
  font-weight: 500;
  transition: all var(--neuro-transition-normal);
}

.action-reset {
  background: var(--neuro-surface);
  border: 1px solid var(--neuro-border);
  color: var(--neuro-text);
}

.action-reset:hover {
  background: var(--neuro-warning);
  border-color: var(--neuro-warning);
  color: var(--neuro-text);
  transform: translateY(-2px);
}

.action-apply {
  background: var(--neuro-gradient-primary);
  color: var(--neuro-text);
}

.action-apply:hover {
  transform: translateY(-2px);
  box-shadow: var(--neuro-shadow-lg);
}

.action-icon {
  font-size: 16px;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .neuro-theme-switcher {
    width: 100%;
    max-width: 400px;
  }

  .theme-modes {
    grid-template-columns: 1fr;
  }

  .theme-colors {
    grid-template-columns: repeat(2, 1fr);
  }

  .preview-colors {
    grid-template-columns: 1fr;
  }
}

/* 主题特定样式 */
.neuro-theme-switcher[data-theme="light"] {
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(10px);
}

.neuro-theme-switcher[data-theme="cyber"] {
  background: rgba(0, 0, 0, 0.95);
  border-left-color: var(--neuro-primary);
}

.neuro-theme-switcher[data-theme="neural"] {
  background: rgba(26, 26, 46, 0.95);
  backdrop-filter: blur(10px);
}

.neuro-theme-switcher[data-theme="matrix"] {
  background: rgba(0, 0, 0, 0.95);
  border-left-color: var(--neuro-matrix);
}
</style>