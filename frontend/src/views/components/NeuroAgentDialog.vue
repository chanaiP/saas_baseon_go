<template>
  <Teleport to="body">
    <!-- Backdrop -->
    <div
      class="nm-dialog-backdrop"
      :class="{ 'is-visible': modelValue }"
      @click="handleOverlayClick"
      @wheel.prevent.stop
      @touchmove.prevent.stop
    >
      <!-- Dialog card -->
      <div
        class="nm-dialog"
        :class="[
          `nm-dialog--${resolvedSize}`,
          { 'is-visible': modelValue, 'is-loading': loading }
        ]"
        :style="{ width: customWidth, height: customHeight }"
        @click.stop
        @wheel.stop
        @touchmove.stop
      >
        <!-- Accent line -->
        <div class="nm-dialog__accent"></div>

        <!-- Header -->
        <header class="nm-dialog__header" v-if="showHeader">
          <div class="nm-dialog__header-left">
            <div class="nm-dialog__icon" v-if="icon">{{ icon }}</div>
            <h2 class="nm-dialog__title">{{ title }}</h2>
          </div>
          <div class="nm-dialog__header-right">
            <button
              v-if="showClose"
              class="nm-dialog__close"
              @click="handleClose"
              :disabled="loading"
              aria-label="Close"
            >
              <svg width="16" height="16" viewBox="0 0 16 16" fill="none">
                <path d="M4 4l8 8M12 4l-8 8" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/>
              </svg>
            </button>
          </div>
        </header>

        <!-- Body -->
        <div class="nm-dialog__body" :class="{ 'is-scrolling': bodyScrolling }" @scroll="handleBodyScroll">
          <slot></slot>
          <slot name="content"></slot>
        </div>

        <!-- Footer -->
        <footer class="nm-dialog__footer" v-if="showFooter">
          <div class="nm-dialog__footer-left">
            <slot name="footer-left"></slot>
          </div>
          <div class="nm-dialog__footer-right">
            <slot name="footer-right">
              <button
                v-if="showCancel"
                class="nm-btn nm-btn--ghost"
                @click="handleCancel"
                :disabled="loading"
              >
                {{ cancelText }}
              </button>
              <button
                v-if="showConfirm"
                class="nm-btn nm-btn--primary"
                @click="handleConfirm"
                :disabled="loading || confirmDisabled"
              >
                <span v-if="loading" class="nm-btn__spinner"></span>
                <span v-else>{{ confirmIcon }}</span>
                {{ confirmText }}
              </button>
            </slot>
          </div>
        </footer>

        <!-- Loading overlay -->
        <div class="nm-dialog__loading" v-if="loading">
          <div class="nm-dialog__loading-spinner"></div>
          <span class="nm-dialog__loading-text">{{ loadingText }}</span>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { computed, ref, watch, onMounted, onUnmounted } from 'vue'

interface Props {
  modelValue: boolean
  title?: string
  icon?: string
  size?: 'small' | 'medium' | 'large' | 'fullscreen'
  width?: string
  height?: string
  showHeader?: boolean
  showFooter?: boolean
  showClose?: boolean
  showCancel?: boolean
  showConfirm?: boolean
  cancelText?: string
  confirmText?: string
  confirmIcon?: string
  confirmDisabled?: boolean
  loadingText?: string
  loading?: boolean
  closeOnOverlayClick?: boolean
  closeOnEsc?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  modelValue: false,
  title: '弹窗',
  size: 'medium',
  showHeader: true,
  showFooter: true,
  showClose: true,
  showCancel: true,
  showConfirm: true,
  cancelText: '取消',
  confirmText: '确认',
  confirmIcon: '✓',
  loadingText: '处理中...',
  loading: false,
  confirmDisabled: false,
  closeOnOverlayClick: true,
  closeOnEsc: true,
})

const sizeMap: Record<string, { w: string; h: string }> = {
  small:    { w: '440px', h: 'auto' },
  medium:   { w: '580px', h: 'auto' },
  large:    { w: '760px', h: 'auto' },
  fullscreen: { w: '92vw', h: '92vh' },
}

const resolvedSize = computed(() => (props.size in sizeMap ? props.size : 'medium'))
const customWidth = computed(() => props.width || sizeMap[resolvedSize.value]?.w)
const customHeight = computed(() => props.height || sizeMap[resolvedSize.value]?.h)
const bodyScrolling = ref(false)
let scrollTimer: number | undefined

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  close: []
  cancel: []
  confirm: []
}>()

function handleOverlayClick() {
  if (props.closeOnOverlayClick && !props.loading) {
    emit('update:modelValue', false)
    emit('close')
  }
}

function handleClose() {
  if (!props.loading) {
    emit('update:modelValue', false)
    emit('close')
  }
}

function handleCancel() {
  emit('cancel')
  if (!props.loading) {
    emit('update:modelValue', false)
  }
}

function handleConfirm() {
  if (!props.confirmDisabled && !props.loading) {
    emit('confirm')
  }
}

function onKeydown(e: KeyboardEvent) {
  if (props.closeOnEsc && e.key === 'Escape' && props.modelValue && !props.loading) {
    handleClose()
  }
}

function handleBodyScroll() {
  bodyScrolling.value = true
  if (scrollTimer) window.clearTimeout(scrollTimer)
  scrollTimer = window.setTimeout(() => {
    bodyScrolling.value = false
  }, 700)
}

onMounted(() => window.addEventListener('keydown', onKeydown))
onUnmounted(() => {
  window.removeEventListener('keydown', onKeydown)
  document.body.style.overflow = ''
  if (scrollTimer) window.clearTimeout(scrollTimer)
})

watch(() => props.modelValue, (v) => {
  document.body.style.overflow = v ? 'hidden' : ''
})
</script>

<style scoped>
/* ===== Backdrop ===== */
.nm-dialog-backdrop {
  position: fixed;
  inset: 0;
  z-index: 10000;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(0, 0, 0, 0);
  backdrop-filter: blur(0px);
  opacity: 0;
  visibility: hidden;
  transition: opacity 0.25s ease, visibility 0.25s ease;
  overscroll-behavior: none;
  touch-action: none;
}

.nm-dialog-backdrop.is-visible {
  background: rgba(10, 12, 20, 0.75);
  backdrop-filter: blur(8px);
  opacity: 1;
  visibility: visible;
}

/* ===== Dialog card ===== */
.nm-dialog {
  position: relative;
  background: #f8fafc;
  border: 1px solid rgba(0, 0, 0, 0.1);
  border-radius: 16px;
  box-shadow:
    0 24px 80px rgba(0, 0, 0, 0.12),
    0 0 1px rgba(0, 245, 212, 0.1) inset;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  max-width: 94vw;
  max-height: 94vh;
  transform: scale(0.96) translateY(12px);
  opacity: 0;
  transition: transform 0.3s cubic-bezier(0.22, 1, 0.36, 1),
              opacity 0.25s ease;
  overscroll-behavior: contain;
  touch-action: auto;
}

.dark .nm-dialog {
  background: #141926;
  border-color: rgba(45, 55, 72, 0.6);
  box-shadow:
    0 24px 80px rgba(0, 0, 0, 0.45),
    0 0 1px rgba(0, 245, 212, 0.15) inset;
}

.nm-dialog.is-visible {
  transform: scale(1) translateY(0);
  opacity: 1;
}

.nm-dialog.is-loading {
  pointer-events: none;
}

/* Accent gradient line at top */
.nm-dialog__accent {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 2px;
  background: linear-gradient(90deg, transparent 0%, #00f5d4 30%, #9d4edd 70%, transparent 100%);
  opacity: 0.9;
  z-index: 1;
}

/* ===== Header ===== */
.nm-dialog__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 20px 24px 16px;
  flex-shrink: 0;
}

.nm-dialog__header-left {
  display: flex;
  align-items: center;
  gap: 12px;
  min-width: 0;
}

.nm-dialog__icon {
  font-size: 22px;
  flex-shrink: 0;
  line-height: 1;
}

.nm-dialog__title {
  margin: 0;
  font-family: 'JetBrains Mono', monospace;
  font-size: 17px;
  font-weight: 600;
  color: #1e293b;
  letter-spacing: -0.01em;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.dark .nm-dialog__title {
  color: #f8fafc;
}

.nm-dialog__header-right {
  flex-shrink: 0;
  margin-left: 12px;
}

.nm-dialog__close {
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(0, 0, 0, 0.05);
  border: 1px solid rgba(0, 0, 0, 0.1);
  border-radius: 8px;
  color: #64748b;
  cursor: pointer;
  transition: all 0.2s ease;
  padding: 0;
}

.dark .nm-dialog__close {
  background: rgba(45, 55, 72, 0.4);
  border-color: rgba(45, 55, 72, 0.5);
  color: #94a3b8;
}

.nm-dialog__close:hover {
  background: rgba(239, 68, 68, 0.12);
  border-color: rgba(239, 68, 68, 0.3);
  color: #f87171;
}

/* ===== Body ===== */
.nm-dialog__body {
  flex: 1 1 auto;
  min-height: 0;
  overflow-y: auto;
  padding: 0 24px 24px;
  color: #475569;
  font-size: 14px;
  line-height: 1.7;
  overscroll-behavior: contain;
}

.dark .nm-dialog__body {
  color: #cbd5e1;
}

.nm-dialog__body::-webkit-scrollbar {
  width: 6px;
}

.nm-dialog__body::-webkit-scrollbar-track {
  background: transparent;
}

.nm-dialog__body::-webkit-scrollbar-thumb {
  background: transparent;
  border-radius: 3px;
}

.nm-dialog__body.is-scrolling::-webkit-scrollbar-thumb {
  background: rgba(0, 0, 0, 0.15);
}

.dark .nm-dialog__body::-webkit-scrollbar-thumb {
  background: transparent;
}

.dark .nm-dialog__body.is-scrolling::-webkit-scrollbar-thumb {
  background: rgba(45, 55, 72, 0.6);
}

/* ===== Footer ===== */
.nm-dialog__footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 24px 20px;
  border-top: 1px solid rgba(0, 0, 0, 0.08);
  flex-shrink: 0;
}

.dark .nm-dialog__footer {
  border-top-color: rgba(45, 55, 72, 0.5);
}

.nm-dialog__footer-left,
.nm-dialog__footer-right {
  display: flex;
  align-items: center;
  gap: 10px;
}

/* ===== Buttons ===== */
.nm-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  padding: 9px 20px;
  font-family: 'JetBrains Mono', monospace;
  font-size: 13px;
  font-weight: 500;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s ease;
  border: 1px solid transparent;
  outline: none;
  white-space: nowrap;
  user-select: none;
}

.nm-btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.nm-btn--primary {
  background: linear-gradient(135deg, #00f5d4 0%, #00c9a7 100%);
  color: #0a0c14;
  border-color: transparent;
  box-shadow: 0 2px 12px rgba(0, 245, 212, 0.25);
}

.nm-btn--primary:hover:not(:disabled) {
  box-shadow: 0 4px 20px rgba(0, 245, 212, 0.35);
  transform: translateY(-1px);
}

.nm-btn--primary:active:not(:disabled) {
  transform: translateY(0);
}

.nm-btn--ghost {
  background: transparent;
  color: #64748b;
  border-color: rgba(0, 0, 0, 0.15);
}

.dark .nm-btn--ghost {
  color: #94a3b8;
  border-color: rgba(45, 55, 72, 0.6);
}

.nm-btn--ghost:hover:not(:disabled) {
  color: #1e293b;
  border-color: rgba(0, 0, 0, 0.25);
  background: rgba(0, 0, 0, 0.03);
}

.dark .nm-btn--ghost:hover:not(:disabled) {
  color: #e2e8f0;
  border-color: #2d3748;
  background: rgba(45, 55, 72, 0.3);
}

.nm-btn__spinner {
  width: 14px;
  height: 14px;
  border: 2px solid rgba(255, 255, 255, 0.3);
  border-top-color: #0a0c14;
  border-radius: 50%;
  animation: nm-spin 0.6s linear infinite;
}

@keyframes nm-spin {
  to { transform: rotate(360deg); }
}

/* ===== Loading overlay ===== */
.nm-dialog__loading {
  position: absolute;
  inset: 0;
  background: rgba(248, 250, 252, 0.92);
  backdrop-filter: blur(6px);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 16px;
  z-index: 10;
}

.dark .nm-dialog__loading {
  background: rgba(20, 25, 38, 0.92);
}

.nm-dialog__loading-spinner {
  width: 40px;
  height: 40px;
  border: 3px solid rgba(0, 0, 0, 0.1);
  border-top-color: #00c9a7;
  border-radius: 50%;
  animation: nm-spin 0.8s linear infinite;
}

.dark .nm-dialog__loading-spinner {
  border-color: rgba(45, 55, 72, 0.4);
  border-top-color: #00f5d4;
}

.nm-dialog__loading-text {
  font-family: 'JetBrains Mono', monospace;
  font-size: 13px;
  color: #00f5d4;
  letter-spacing: 0.02em;
}

/* ===== Sizes ===== */
.nm-dialog--small    { width: 440px; }
.nm-dialog--medium   { width: 580px; }
.nm-dialog--large    { width: 760px; }
.nm-dialog--fullscreen { width: 92vw !important; height: 92vh !important; border-radius: 16px; }

/* ===== Responsive ===== */
@media (max-width: 768px) {
  .nm-dialog,
  .nm-dialog--small,
  .nm-dialog--medium,
  .nm-dialog--large {
    width: 94vw !important;
    max-height: 94vh !important;
  }
  .nm-dialog__header { padding: 16px 20px 14px; }
  .nm-dialog__body { padding: 0 20px 20px; }
  .nm-dialog__footer { padding: 14px 20px 18px; }
}
</style>
