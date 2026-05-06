<template>
  <div
    class="neuro-logo-editor"
    :class="{ 'is-readonly': !canEditBranding }"
    @click="openSystemSettings"
  >
    <div class="logo-name-container">
      <div class="logo-container">
        <div v-if="effectiveLogo" class="custom-logo">
          <img :src="effectiveLogo" alt="系统Logo" class="logo-image" />
          <div v-if="canEditBranding" class="logo-overlay">
            <span class="overlay-text">编辑系统设置</span>
          </div>
        </div>

        <div v-else class="default-neuro-logo">
          <div class="neuro-matrix-logo">
            <div class="matrix-grid">
              <div class="matrix-dot dot-1"></div>
              <div class="matrix-dot dot-2"></div>
              <div class="matrix-dot dot-3"></div>
              <div class="matrix-dot dot-4"></div>
              <div class="matrix-dot dot-5 active-dot"></div>
              <div class="matrix-dot dot-6"></div>
              <div class="matrix-dot dot-7 active-dot"></div>
              <div class="matrix-dot dot-8"></div>
              <div class="matrix-dot dot-9"></div>
              <div class="matrix-dot dot-10 active-dot"></div>
              <div class="matrix-dot dot-11"></div>
              <div class="matrix-dot dot-12 active-dot"></div>
              <div class="matrix-dot dot-13"></div>
              <div class="matrix-dot dot-14"></div>
              <div class="matrix-dot dot-15"></div>
              <div class="matrix-dot dot-16"></div>
            </div>
            <div class="matrix-connection conn-1"></div>
            <div class="matrix-connection conn-2"></div>
            <div class="matrix-connection conn-3"></div>
            <div class="matrix-connection conn-4"></div>
            <div class="neuro-glow"></div>
            <div class="neuro-pulse"></div>
          </div>
          <div v-if="canEditBranding" class="logo-overlay">
            <span class="overlay-text">编辑系统设置</span>
          </div>
        </div>
      </div>

      <div class="system-name-display">
        <h1 class="system-name">{{ displayName }}</h1>
      </div>
    </div>

    <SystemSettingsDialog
      v-model:visible="showSystemSettings"
      :current-logo="currentLogo"
      :current-system-name="displayName"
      :current-copyright="copyrightInfo"
      :is-platform-admin="props.isPlatformAdmin"
      @save="handleSettingsSave"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import SystemSettingsDialog from './SystemSettingsDialog.vue'

interface Props {
  currentLogo?: string
  currentName?: string
  currentCopyright?: string
  isPlatformAdmin?: boolean
  /** 具备 ``brand:edit`` 且套餐含 ``brand_config``（或平台管理员）；无则仅展示不可点开设置 */
  canEditBranding?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  currentLogo: '',
  currentName: 'Ai DevOS',
  currentCopyright: '© 2026 PMTools - AI协作开发系统',
  isPlatformAdmin: false,
  canEditBranding: false,
})

const emit = defineEmits<{
  'update:logo': [logoUrl: string]
  'update:name': [name: string]
  'update:copyright': [copyright: string]
  'settings-change': [settings: {
    logo: string
    name: string
    copyright: string
  }]
}>()

// 本地状态
const currentLogo = ref(props.currentLogo)
const displayName = ref(props.currentName)
const copyrightInfo = ref(props.currentCopyright)
const showSystemSettings = ref(false)

// 空字符串时不显示自定义 logo，使用默认神经元 logo
const effectiveLogo = computed(() => {
  const v = currentLogo.value.trim()
  return v || ''
})

// 同步父组件传入的 props 变化
watch(() => props.currentLogo, (v) => { currentLogo.value = v })
watch(() => props.currentName, (v) => { displayName.value = v })
watch(() => props.currentCopyright, (v) => { copyrightInfo.value = v })

// 打开系统设置（权限与套餐由父组件传入的 canEditBranding 统一门禁）
const openSystemSettings = () => {
  if (!props.canEditBranding) return
  showSystemSettings.value = true
}

// 处理设置保存
const handleSettingsSave = (settings: {
  logo: string
  systemName: string
  copyright: string
}) => {
  currentLogo.value = settings.logo
  displayName.value = settings.systemName
  copyrightInfo.value = settings.copyright

  emit('update:logo', settings.logo)
  emit('update:name', settings.systemName)
  emit('update:copyright', settings.copyright)
  emit('settings-change', {
    logo: settings.logo,
    name: settings.systemName,
    copyright: settings.copyright
  })
}
</script>

<style scoped>
.neuro-logo-editor {
  display: flex;
  align-items: center;
  gap: 20px;
  cursor: pointer;
  transition: all 0.3s ease;
  padding: 12px;
  border-radius: 12px;
  position: relative;
  overflow: hidden;
}

.neuro-logo-editor::before {
  content: '';
  position: absolute;
  top: 0;
  left: -100%;
  width: 100%;
  height: 100%;
  background: linear-gradient(90deg, transparent, rgba(0, 245, 212, 0.1), transparent);
  transition: left 0.5s ease;
}

.neuro-logo-editor:hover::before {
  left: 100%;
}

.neuro-logo-editor:hover {
  background: rgba(0, 245, 212, 0.05);
  transform: translateY(-2px);
  box-shadow: 0 4px 20px rgba(0, 245, 212, 0.2);
}

.neuro-logo-editor.is-readonly {
  cursor: default;
}

.neuro-logo-editor.is-readonly:hover {
  background: transparent;
  transform: none;
  box-shadow: none;
}

.neuro-logo-editor.is-readonly::before {
  display: none;
}

.logo-name-container {
  display: flex;
  align-items: center;
  gap: 20px;
  width: 100%;
}

.logo-container {
  position: relative;
  width: 60px;
  height: 60px;
  flex-shrink: 0;
}

.custom-logo,
.default-neuro-logo {
  width: 100%;
  height: 100%;
  border-radius: 12px;
  overflow: hidden;
  position: relative;
  background: rgba(16, 16, 32, 0.8);
  transition: all 0.3s ease;
}

.neuro-logo-editor:hover .custom-logo,
.neuro-logo-editor:hover .default-neuro-logo {
  transform: scale(1.05);
}

.logo-image {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.logo-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.7);
  display: flex;
  align-items: center;
  justify-content: center;
  opacity: 0;
  transition: opacity 0.3s ease;
}

.neuro-logo-editor:hover .logo-overlay {
  opacity: 1;
}

.overlay-text {
  color: #00f5d4;
  font-size: 11px;
  font-weight: 600;
  text-align: center;
  padding: 4px 8px;
  background: rgba(0, 0, 0, 0.8);
  border-radius: 4px;
  border: 1px solid rgba(0, 245, 212, 0.3);
}

/* 默认神经元Logo — 4x4 矩阵 */
.neuro-matrix-logo {
  width: 100%;
  height: 100%;
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
}

.matrix-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  grid-template-rows: repeat(4, 1fr);
  gap: 3px;
  width: 40px;
  height: 40px;
  position: relative;
  z-index: 2;
}

.matrix-dot {
  width: 4px;
  height: 4px;
  border-radius: 50%;
  background: rgba(0, 245, 212, 0.3);
  transition: all 0.3s ease;
}

.active-dot {
  background: #00f5d4;
  box-shadow: 0 0 8px #00f5d4;
  animation: pulse 2s infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 0.3; transform: scale(1); }
  50% { opacity: 1; transform: scale(1.2); }
}

.matrix-connection {
  position: absolute;
  background: rgba(0, 245, 212, 0.2);
  z-index: 1;
}

.conn-1 {
  top: 25%; left: 25%;
  width: 20px; height: 1px;
  transform: rotate(45deg);
}
.conn-2 {
  top: 25%; right: 25%;
  width: 20px; height: 1px;
  transform: rotate(-45deg);
}
.conn-3 {
  bottom: 25%; left: 25%;
  width: 20px; height: 1px;
  transform: rotate(-45deg);
}
.conn-4 {
  bottom: 25%; right: 25%;
  width: 20px; height: 1px;
  transform: rotate(45deg);
}

.neuro-glow {
  position: absolute;
  top: 0; left: 0; right: 0; bottom: 0;
  border-radius: 12px;
  box-shadow: inset 0 0 20px rgba(0, 245, 212, 0.2);
  pointer-events: none;
}

.neuro-pulse {
  position: absolute;
  top: 50%; left: 50%;
  transform: translate(-50%, -50%);
  width: 30px; height: 30px;
  border-radius: 50%;
  border: 1px solid rgba(0, 245, 212, 0.3);
  animation: ripple 3s infinite;
}

@keyframes ripple {
  0% { width: 30px; height: 30px; opacity: 1; }
  100% { width: 60px; height: 60px; opacity: 0; }
}

/* 编辑徽章 */
.edit-badge {
  position: absolute;
  top: -6px;
  right: -6px;
  width: 24px;
  height: 24px;
  background: linear-gradient(135deg, #00f5d4, #9d4edd);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 2px solid #0a0a0f;
  z-index: 10;
  opacity: 0;
  transform: scale(0.8);
  transition: all 0.3s ease;
}

.neuro-logo-editor:hover .edit-badge {
  opacity: 1;
  transform: scale(1);
}

.badge-icon {
  font-size: 12px;
  color: #0a0a0f;
  font-weight: bold;
}

/* 系统名称显示 */
.system-name-display {
  flex: 1;
}

.system-name {
  font-size: 20px;
  font-weight: 700;
  margin: 0;
  white-space: pre-wrap;
  letter-spacing: 0.02em;
  background: linear-gradient(90deg, #00f5d4, #9d4edd);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  transition: all 0.3s ease;
}

.neuro-logo-editor:hover .system-name {
  background: linear-gradient(90deg, #00e6c7, #8a2be2);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}
</style>
