<template>
  <div class="neuro-image-upload-editor">
    <!-- 上传触发器 -->
    <div v-if="!isEditing" class="upload-trigger" @click="triggerFileInput">
      <slot name="trigger">
        <div class="default-trigger">
          <div class="trigger-icon">📷</div>
          <div class="trigger-text">上传图片</div>
        </div>
      </slot>
    </div>

    <!-- 编辑模式 -->
    <div v-else class="image-editor">
      <div class="editor-header">
        <h3 class="editor-title">{{ editorTitle }}</h3>
        <button class="editor-close" @click="cancelEdit">×</button>
      </div>

      <div class="editor-content">
        <!-- 图片预览和裁剪区域 -->
        <div class="preview-container">
          <div class="crop-area" ref="cropArea">
            <img
              ref="imageElement"
              :src="imageSrc"
              alt="编辑图片"
              class="editing-image"
              @load="onImageLoad"
            />

            <!-- 裁剪框 -->
            <div
              v-if="showCropBox"
              class="crop-box"
              :style="cropBoxStyle"
              @mousedown="startDrag"
              @touchstart="startDrag"
            >
              <!-- 裁剪框控制点 -->
              <div class="crop-handle top-left" @mousedown="startResize('top-left')" @touchstart="startResize('top-left')"></div>
              <div class="crop-handle top-right" @mousedown="startResize('top-right')" @touchstart="startResize('top-right')"></div>
              <div class="crop-handle bottom-left" @mousedown="startResize('bottom-left')" @touchstart="startResize('bottom-left')"></div>
              <div class="crop-handle bottom-right" @mousedown="startResize('bottom-right')" @touchstart="startResize('bottom-right')"></div>

              <!-- 裁剪框网格线 -->
              <div class="crop-grid">
                <div class="grid-line vertical"></div>
                <div class="grid-line horizontal"></div>
              </div>
            </div>
          </div>

          <!-- 裁剪框尺寸提示 -->
          <div class="crop-info">
            <div class="crop-dimensions">
              <span class="dimension-label">裁剪尺寸:</span>
              <span class="dimension-value">{{ cropBox.width }} × {{ cropBox.height }} px</span>
            </div>
            <div class="crop-ratio">
              <span class="ratio-label">宽高比:</span>
              <span class="ratio-value">{{ cropAspectRatio }}</span>
            </div>
          </div>
        </div>

        <!-- 编辑控制面板 -->
        <div class="control-panel">
          <div class="control-group">
            <h4 class="control-title">裁剪设置</h4>

            <div class="ratio-presets">
              <button
                v-for="preset in aspectRatioPresets"
                :key="preset.value"
                class="ratio-preset-btn"
                :class="{ active: cropBox.aspectRatio === preset.value }"
                @click="setAspectRatio(preset.value)"
              >
                {{ preset.label }}
              </button>
            </div>

            <div class="size-controls">
              <div class="size-control">
                <label class="control-label">宽度 (px)</label>
                <input
                  type="number"
                  v-model.number="cropBox.width"
                  min="50"
                  :max="maxCropSize"
                  class="size-input"
                  @change="constrainCropBox"
                />
              </div>
              <div class="size-control">
                <label class="control-label">高度 (px)</label>
                <input
                  type="number"
                  v-model.number="cropBox.height"
                  min="50"
                  :max="maxCropSize"
                  class="size-input"
                  @change="constrainCropBox"
                />
              </div>
            </div>

            <div class="zoom-control">
              <label class="control-label">缩放: {{ zoomLevel }}%</label>
              <input
                type="range"
                v-model.number="zoomLevel"
                min="10"
                max="200"
                step="5"
                class="zoom-slider"
                @input="updateZoom"
              />
            </div>
          </div>

          <div class="control-group">
            <h4 class="control-title">操作</h4>
            <div class="action-buttons">
              <button class="action-btn reset-btn" @click="resetCropBox">
                ↺ 重置
              </button>
              <button class="action-btn rotate-btn" @click="rotateImage">
                ⟳ 旋转
              </button>
              <button class="action-btn flip-btn" @click="flipImage">
                ⇄ 翻转
              </button>
            </div>
          </div>
        </div>
      </div>

      <div class="editor-footer">
        <button class="neuro-btn neuro-btn-ghost" @click="cancelEdit">
          取消
        </button>
        <button class="neuro-btn neuro-btn-primary" @click="confirmCrop" :disabled="isProcessing">
          {{ isProcessing ? '处理中...' : '确认裁剪' }}
        </button>
      </div>
    </div>

    <!-- 隐藏的文件输入 -->
    <input
      ref="fileInput"
      type="file"
      accept="image/*"
      @change="handleFileChange"
      style="display: none"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onUnmounted, nextTick } from 'vue'

interface Props {
  // 使用场景类型
  usageType?: 'logo' | 'avatar' | 'banner' | 'custom'
  // 预设裁剪尺寸
  presetWidth?: number
  presetHeight?: number
  // 预设宽高比
  aspectRatio?: number
  // 最小尺寸
  minSize?: number
  // 最大尺寸
  maxSize?: number
  // 编辑器标题
  editorTitle?: string
}

const props = withDefaults(defineProps<Props>(), {
  usageType: 'custom',
  presetWidth: 200,
  presetHeight: 200,
  aspectRatio: 1,
  minSize: 50,
  maxSize: 2000,
  editorTitle: '图片编辑'
})

const emit = defineEmits<{
  'update:image': [imageData: string]
  'upload-success': [imageData: string]
  'upload-error': [error: string]
  'cancel': []
}>()

// 响应式数据
const fileInput = ref<HTMLInputElement | null>(null)
const imageElement = ref<HTMLImageElement | null>(null)

const imageSrc = ref('')
const isEditing = ref(false)
const isProcessing = ref(false)
const showCropBox = ref(false)

// 图片状态
const imageState = ref({
  naturalWidth: 0,
  naturalHeight: 0,
  rotation: 0,
  scaleX: 1,
  scaleY: 1
})

// 裁剪框状态
const cropBox = ref({
  x: 0,
  y: 0,
  width: props.presetWidth,
  height: props.presetHeight,
  aspectRatio: props.aspectRatio
})

// 交互状态
const isDragging = ref(false)
const isResizing = ref(false)
const dragStart = ref({ x: 0, y: 0 })
const resizeDirection = ref<string | null>(null)

// 缩放级别
const zoomLevel = ref(100)

// 根据使用场景设置预设
const aspectRatioPresets = computed(() => {
  const presets = [
    { label: '自由', value: 0 },
    { label: '1:1', value: 1 },
    { label: '4:3', value: 4/3 },
    { label: '16:9', value: 16/9 },
    { label: '3:4', value: 3/4 },
    { label: '9:16', value: 9/16 }
  ]

  // 根据使用场景添加特定预设
  if (props.usageType === 'logo') {
    return [
      { label: '正方形 (1:1)', value: 1 },
      { label: '长方形 (4:3)', value: 4/3 },
      { label: '自由', value: 0 }
    ]
  } else if (props.usageType === 'avatar') {
    return [
      { label: '圆形 (1:1)', value: 1 },
      { label: '自由', value: 0 }
    ]
  }

  return presets
})

// 计算属性
const cropBoxStyle = computed(() => ({
  left: `${cropBox.value.x}px`,
  top: `${cropBox.value.y}px`,
  width: `${cropBox.value.width}px`,
  height: `${cropBox.value.height}px`
}))

const cropAspectRatio = computed(() => {
  if (cropBox.value.aspectRatio === 0) {
    const ratio = cropBox.value.width / cropBox.value.height
    return ratio.toFixed(2) + ':1'
  }
  return '1:' + (1 / cropBox.value.aspectRatio).toFixed(2)
})

const maxCropSize = computed(() => {
  return Math.min(props.maxSize, Math.max(imageState.value.naturalWidth, imageState.value.naturalHeight))
})

// 触发文件选择
const triggerFileInput = () => {
  if (fileInput.value) {
    fileInput.value.click()
  }
}

// 公开方法：通过ref触发文件选择
defineExpose({
  triggerFileInput
})

// 处理文件选择
const handleFileChange = (event: Event) => {
  const target = event.target as HTMLInputElement
  const file = target.files?.[0]

  if (!file) return

  // 检查文件类型
  if (!file.type.startsWith('image/')) {
    emit('upload-error', '请选择图片文件')
    return
  }

  // 检查文件大小 (限制为10MB)
  if (file.size > 10 * 1024 * 1024) {
    emit('upload-error', '图片大小不能超过10MB')
    return
  }

  // 创建预览
  const reader = new FileReader()
  reader.onload = (e) => {
    imageSrc.value = e.target?.result as string
    isEditing.value = true
    showCropBox.value = false

    // 重置图片状态
    imageState.value = {
      naturalWidth: 0,
      naturalHeight: 0,
      rotation: 0,
      scaleX: 1,
      scaleY: 1
    }

    // 重置裁剪框
    resetCropBox()
  }
  reader.readAsDataURL(file)

  // 重置文件输入
  if (target) {
    target.value = ''
  }
}

// 图片加载完成
const onImageLoad = () => {
  if (!imageElement.value) return

  const img = imageElement.value
  imageState.value.naturalWidth = img.naturalWidth
  imageState.value.naturalHeight = img.naturalHeight

  // 根据使用场景初始化裁剪框
  initCropBox()

  // 显示裁剪框
  nextTick(() => {
    showCropBox.value = true
    updateZoom()
  })
}

// 初始化裁剪框
const initCropBox = () => {
  const { naturalWidth, naturalHeight } = imageState.value

  // 根据使用场景设置默认裁剪框
  let width = props.presetWidth
  let height = props.presetHeight

  if (props.usageType === 'logo') {
    // Logo通常需要正方形
    const size = Math.min(naturalWidth, naturalHeight, 300)
    width = height = size
    cropBox.value.aspectRatio = 1
  } else if (props.usageType === 'avatar') {
    // 头像需要正方形
    const size = Math.min(naturalWidth, naturalHeight, 200)
    width = height = size
    cropBox.value.aspectRatio = 1
  } else if (props.usageType === 'banner') {
    // 横幅需要宽屏
    width = Math.min(naturalWidth, 800)
    height = Math.round(width / (16/9))
    cropBox.value.aspectRatio = 16/9
  }

  // 确保不超过图片尺寸
  width = Math.min(width, naturalWidth)
  height = Math.min(height, naturalHeight)

  // 居中裁剪框
  cropBox.value.width = width
  cropBox.value.height = height
  cropBox.value.x = (naturalWidth - width) / 2
  cropBox.value.y = (naturalHeight - height) / 2

  constrainCropBox()
}

// 约束裁剪框
const constrainCropBox = () => {
  const { naturalWidth, naturalHeight } = imageState.value

  // 确保最小尺寸
  cropBox.value.width = Math.max(props.minSize, cropBox.value.width)
  cropBox.value.height = Math.max(props.minSize, cropBox.value.height)

  // 确保不超过图片尺寸
  cropBox.value.width = Math.min(cropBox.value.width, naturalWidth)
  cropBox.value.height = Math.min(cropBox.value.height, naturalHeight)

  // 应用宽高比约束
  if (cropBox.value.aspectRatio > 0) {
    if (cropBox.value.aspectRatio === 1) {
      // 正方形
      const size = Math.min(cropBox.value.width, cropBox.value.height)
      cropBox.value.width = cropBox.value.height = size
    } else {
      // 固定宽高比
      cropBox.value.height = Math.round(cropBox.value.width / cropBox.value.aspectRatio)
    }
  }

  // 确保裁剪框在图片范围内
  cropBox.value.x = Math.max(0, Math.min(cropBox.value.x, naturalWidth - cropBox.value.width))
  cropBox.value.y = Math.max(0, Math.min(cropBox.value.y, naturalHeight - cropBox.value.height))
}

// 设置宽高比
const setAspectRatio = (ratio: number) => {
  cropBox.value.aspectRatio = ratio
  constrainCropBox()
}

// 重置裁剪框
const resetCropBox = () => {
  initCropBox()
  zoomLevel.value = 100
  updateZoom()
}

// 更新缩放
const updateZoom = () => {
  if (!imageElement.value) return

  const scale = zoomLevel.value / 100
  imageElement.value.style.transform = `
    scale(${scale})
    rotate(${imageState.value.rotation}deg)
    scaleX(${imageState.value.scaleX})
    scaleY(${imageState.value.scaleY})
  `
}

// 旋转图片
const rotateImage = () => {
  imageState.value.rotation = (imageState.value.rotation + 90) % 360
  updateZoom()

  // 旋转后可能需要调整裁剪框
  nextTick(() => {
    constrainCropBox()
  })
}

// 翻转图片
const flipImage = () => {
  imageState.value.scaleX = -imageState.value.scaleX
  updateZoom()
}

// 开始拖动
const startDrag = (event: MouseEvent | TouchEvent) => {
  event.preventDefault()
  isDragging.value = true

  const clientX = 'touches' in event ? event.touches[0].clientX : event.clientX
  const clientY = 'touches' in event ? event.touches[0].clientY : event.clientY

  dragStart.value = {
    x: clientX - cropBox.value.x,
    y: clientY - cropBox.value.y
  }

  document.addEventListener('mousemove', handleDrag)
  document.addEventListener('touchmove', handleDrag, { passive: false })
  document.addEventListener('mouseup', stopDrag)
  document.addEventListener('touchend', stopDrag)
}

// 处理拖动
const handleDrag = (event: MouseEvent | TouchEvent) => {
  if (!isDragging.value) return

  event.preventDefault()

  const clientX = 'touches' in event ? event.touches[0].clientX : event.clientX
  const clientY = 'touches' in event ? event.touches[0].clientY : event.clientY

  cropBox.value.x = clientX - dragStart.value.x
  cropBox.value.y = clientY - dragStart.value.y

  constrainCropBox()
}

// 停止拖动
const stopDrag = () => {
  isDragging.value = false
  document.removeEventListener('mousemove', handleDrag)
  document.removeEventListener('touchmove', handleDrag)
  document.removeEventListener('mouseup', stopDrag)
  document.removeEventListener('touchend', stopDrag)
}

// 开始调整大小
const startResize = (direction: string) => (event: MouseEvent | TouchEvent) => {
  event.preventDefault()
  event.stopPropagation()

  isResizing.value = true
  resizeDirection.value = direction

  const clientX = 'touches' in event ? event.touches[0].clientX : event.clientX
  const clientY = 'touches' in event ? event.touches[0].clientY : event.clientY

  dragStart.value = {
    x: clientX,
    y: clientY
  }

  document.addEventListener('mousemove', handleResize)
  document.addEventListener('touchmove', handleResize, { passive: false })
  document.addEventListener('mouseup', stopResize)
  document.addEventListener('touchend', stopResize)
}

// 处理调整大小
const handleResize = (event: MouseEvent | TouchEvent) => {
  if (!isResizing.value || !resizeDirection.value) return

  event.preventDefault()

  const clientX = 'touches' in event ? event.touches[0].clientX : event.clientX
  const clientY = 'touches' in event ? event.touches[0].clientY : event.clientY

  const deltaX = clientX - dragStart.value.x
  const deltaY = clientY - dragStart.value.y

  const direction = resizeDirection.value

  // 根据方向调整裁剪框
  if (direction.includes('right')) {
    cropBox.value.width = Math.max(props.minSize, cropBox.value.width + deltaX)
  }
  if (direction.includes('left')) {
    const newWidth = Math.max(props.minSize, cropBox.value.width - deltaX)
    cropBox.value.x += cropBox.value.width - newWidth
    cropBox.value.width = newWidth
  }
  if (direction.includes('bottom')) {
    cropBox.value.height = Math.max(props.minSize, cropBox.value.height + deltaY)
  }
  if (direction.includes('top')) {
    const newHeight = Math.max(props.minSize, cropBox.value.height - deltaY)
    cropBox.value.y += cropBox.value.height - newHeight
    cropBox.value.height = newHeight
  }

  dragStart.value = { x: clientX, y: clientY }
  constrainCropBox()
}

// 停止调整大小
const stopResize = () => {
  isResizing.value = false
  resizeDirection.value = null
  document.removeEventListener('mousemove', handleResize)
  document.removeEventListener('touchmove', handleResize)
  document.removeEventListener('mouseup', stopResize)
  document.removeEventListener('touchend', stopResize)
}

// 确认裁剪
const confirmCrop = async () => {
  if (!imageElement.value) return

  isProcessing.value = true

  try {
    // 创建Canvas进行裁剪
    const canvas = document.createElement('canvas')
    const ctx = canvas.getContext('2d')

    if (!ctx) {
      throw new Error('无法创建Canvas上下文')
    }

    // 设置Canvas尺寸为裁剪框尺寸
    canvas.width = cropBox.value.width
    canvas.height = cropBox.value.height

    // 计算缩放比例
    const scale = zoomLevel.value / 100

    // 绘制裁剪区域
    ctx.save()

    // 应用变换
    ctx.translate(canvas.width / 2, canvas.height / 2)
    ctx.rotate((imageState.value.rotation * Math.PI) / 180)
    ctx.scale(imageState.value.scaleX, imageState.value.scaleY)
    ctx.translate(-canvas.width / 2, -canvas.height / 2)

    // 绘制图片
    ctx.drawImage(
      imageElement.value,
      cropBox.value.x / scale,
      cropBox.value.y / scale,
      cropBox.value.width / scale,
      cropBox.value.height / scale,
      0,
      0,
      canvas.width,
      canvas.height
    )

    ctx.restore()

    // 转换为DataURL
    const croppedImageData = canvas.toDataURL('image/png')

    // 发出事件
    emit('update:image', croppedImageData)
    emit('upload-success', croppedImageData)

    // 关闭编辑器
    cancelEdit()

  } catch (error) {
    emit('upload-error', error instanceof Error ? error.message : '裁剪失败')
  } finally {
    isProcessing.value = false
  }
}

// 取消编辑
const cancelEdit = () => {
  isEditing.value = false
  imageSrc.value = ''
  showCropBox.value = false
  emit('cancel')
}

// 清理事件监听器
onUnmounted(() => {
  stopDrag()
  stopResize()
})
</script>

<style scoped>
.neuro-image-upload-editor {
  font-family: 'JetBrains Mono', monospace;
}

/* 上传触发器 */
.upload-trigger {
  cursor: pointer;
  transition: all 0.3s ease;
}

.upload-trigger:hover {
  transform: scale(1.05);
}

.default-trigger {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 20px;
  background: var(--nm-bg-surface);
  border: 2px dashed var(--nm-border);
  border-radius: 12px;
  transition: all 0.3s ease;
}

.default-trigger:hover {
  border-color: var(--nm-primary);
  background: var(--nm-bg-card);
}

.trigger-icon {
  font-size: 32px;
  margin-bottom: 8px;
}

.trigger-text {
  font-size: 14px;
  color: var(--nm-text-secondary);
}

/* 编辑器样式 */
.image-editor {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.8);
  display: flex;
  flex-direction: column;
  z-index: 1000;
  backdrop-filter: blur(10px);
}

.editor-header {
  padding: 20px 24px;
  background: var(--nm-bg-card);
  border-bottom: 1px solid var(--nm-border);
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.editor-title {
  font-family: 'JetBrains Mono', monospace;
  font-size: 18px;
  font-weight: 600;
  color: var(--nm-text-primary);
  margin: 0;
}

.editor-close {
  background: transparent;
  border: none;
  color: var(--nm-text-secondary);
  font-size: 24px;
  cursor: pointer;
  padding: 0;
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 8px;
  transition: all 0.2s ease;
}

.editor-close:hover {
  background: var(--nm-bg-surface);
  color: var(--nm-primary);
}

.editor-content {
  flex: 1;
  display: flex;
  padding: 24px;
  gap: 24px;
  overflow: hidden;
}

/* 预览容器 */
.preview-container {
  flex: 1;
  display: flex;
  flex-direction: column;
  background: var(--nm-bg-deep);
  border-radius: 12px;
  overflow: hidden;
  position: relative;
}

.crop-area {
  flex: 1;
  position: relative;
  overflow: hidden;
  display: flex;
  align-items: center;
  justify-content: center;
  background: repeating-conic-gradient(#1a1a1a 0% 25%, #2a2a2a 0% 50%) 50% / 20px 20px;
}

.editing-image {
  max-width: 100%;
  max-height: 100%;
  object-fit: contain;
  transition: transform 0.3s ease;
}

/* 裁剪框样式 */
.crop-box {
  position: absolute;
  border: 2px solid var(--nm-primary);
  background: rgba(0, 245, 212, 0.1);
  cursor: move;
  box-shadow: 0 0 0 9999px rgba(0, 0, 0, 0.5);
}

.crop-handle {
  position: absolute;
  width: 12px;
  height: 12px;
  background: var(--nm-primary);
  border: 2px solid var(--nm-bg-card);
  border-radius: 2px;
}

.crop-handle.top-left {
  top: -6px;
  left: -6px;
  cursor: nwse-resize;
}

.crop-handle.top-right {
  top: -6px;
  right: -6px;
  cursor: nesw-resize;
}

.crop-handle.bottom-left {
  bottom: -6px;
  left: -6px;
  cursor: nesw-resize;
}

.crop-handle.bottom-right {
  bottom: -6px;
  right: -6px;
  cursor: nwse-resize;
}

.crop-grid {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  pointer-events: none;
}

.grid-line {
  position: absolute;
  background: rgba(255, 255, 255, 0.3);
}

.grid-line.vertical {
  top: 0;
  bottom: 0;
  left: 50%;
  width: 1px;
  transform: translateX(-50%);
}

.grid-line.horizontal {
  left: 0;
  right: 0;
  top: 50%;
  height: 1px;
  transform: translateY(-50%);
}

/* 裁剪信息 */
.crop-info {
  padding: 16px;
  background: var(--nm-bg-card);
  border-top: 1px solid var(--nm-border);
  display: flex;
  justify-content: space-between;
  font-size: 14px;
}

.dimension-label,
.ratio-label {
  color: var(--nm-text-secondary);
  margin-right: 8px;
}

.dimension-value,
.ratio-value {
  color: var(--nm-text-primary);
  font-weight: 600;
}

/* 控制面板 */
.control-panel {
  width: 300px;
  background: var(--nm-bg-card);
  border-radius: 12px;
  padding: 20px;
  display: flex;
  flex-direction: column;
  gap: 24px;
  overflow-y: auto;
}

.control-group {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.control-title {
  font-size: 16px;
  font-weight: 600;
  color: var(--nm-text-primary);
  margin: 0;
}

/* 宽高比预设 */
.ratio-presets {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.ratio-preset-btn {
  padding: 8px 12px;
  background: var(--nm-bg-surface);
  border: 1px solid var(--nm-border);
  border-radius: 6px;
  color: var(--nm-text-secondary);
  cursor: pointer;
  transition: all 0.2s ease;
  font-size: 13px;
}

.ratio-preset-btn:hover {
  border-color: var(--nm-primary);
  color: var(--nm-primary);
}

.ratio-preset-btn.active {
  background: var(--nm-primary);
  border-color: var(--nm-primary);
  color: var(--nm-text-on-dark);
}

/* 尺寸控制 */
.size-controls {
  display: flex;
  gap: 12px;
}

.size-control {
  flex: 1;
}

.control-label {
  display: block;
  font-size: 13px;
  color: var(--nm-text-secondary);
  margin-bottom: 6px;
}

.size-input {
  width: 100%;
  padding: 8px 12px;
  background: var(--nm-bg-surface);
  border: 1px solid var(--nm-border);
  border-radius: 6px;
  color: var(--nm-text-primary);
  font-family: 'JetBrains Mono', monospace;
  font-size: 14px;
}

.size-input:focus {
  outline: none;
  border-color: var(--nm-primary);
}

/* 缩放控制 */
.zoom-control {
  margin-top: 8px;
}

.zoom-slider {
  width: 100%;
  height: 6px;
  margin-top: 8px;
  background: var(--nm-bg-surface);
  border-radius: 3px;
  outline: none;
  -webkit-appearance: none;
}

.zoom-slider::-webkit-slider-thumb {
  -webkit-appearance: none;
  width: 18px;
  height: 18px;
  background: var(--nm-primary);
  border-radius: 50%;
  cursor: pointer;
}

.zoom-slider::-moz-range-thumb {
  width: 18px;
  height: 18px;
  background: var(--nm-primary);
  border-radius: 50%;
  cursor: pointer;
  border: none;
}

/* 操作按钮 */
.action-buttons {
  display: flex;
  gap: 8px;
}

.action-btn {
  flex: 1;
  padding: 10px 12px;
  background: var(--nm-bg-surface);
  border: 1px solid var(--nm-border);
  border-radius: 6px;
  color: var(--nm-text-primary);
  cursor: pointer;
  transition: all 0.2s ease;
  font-size: 14px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
}

.action-btn:hover {
  background: var(--nm-bg-card);
  border-color: var(--nm-primary);
}

.reset-btn:hover {
  color: var(--nm-warning);
  border-color: var(--nm-warning);
}

.rotate-btn:hover {
  color: var(--nm-primary);
  border-color: var(--nm-primary);
}

.flip-btn:hover {
  color: var(--nm-info);
  border-color: var(--nm-info);
}

/* 编辑器底部 */
.editor-footer {
  padding: 20px 24px;
  background: var(--nm-bg-card);
  border-top: 1px solid var(--nm-border);
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}

/* 响应式设计 */
@media (max-width: 1024px) {
  .editor-content {
    flex-direction: column;
  }

  .control-panel {
    width: 100%;
    max-height: 300px;
  }
}

@media (max-width: 768px) {
  .editor-content {
    padding: 16px;
  }

  .size-controls {
    flex-direction: column;
  }

  .action-buttons {
    flex-direction: column;
  }
}
</style>