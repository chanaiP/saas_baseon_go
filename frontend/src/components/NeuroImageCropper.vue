<template>
  <div class="neuro-image-cropper">
    <!-- 独立按钮模式 -->
    <div v-if="props.interactionMode === 'standalone' && !isEditing" class="upload-trigger" @click="triggerFileInput">
      <slot name="trigger">
        <div class="default-trigger">
          <div class="trigger-icon">📷</div>
          <div class="trigger-text">上传图片</div>
        </div>
      </slot>
    </div>

    <!-- 悬停模式 -->
    <div
      v-else-if="props.interactionMode === 'hover' && !isEditing"
      class="hover-container"
      :style="hoverContainerStyle"
      @mouseenter="isHovering = true"
      @mouseleave="isHovering = false"
    >
      <!-- 默认内容插槽 -->
      <slot name="hover-content">
        <div class="default-hover-content">
          <!-- 平时显示的内容 -->
          <div class="hover-placeholder">
            <div class="placeholder-icon">🖼️</div>
            <!-- 移除"点击上传图片"文字 -->
          </div>
        </div>
      </slot>

      <!-- 悬停时显示的 + 号按钮（在父组件内部） -->
      <div v-if="isHovering" class="hover-plus-button" @click="triggerFileInput">
        <div class="plus-icon">+</div>
        <div class="plus-text">{{ props.hoverText }}</div>
      </div>
    </div>

    <!-- 编辑弹窗 - 使用Teleport挂载到body -->
    <Teleport to="body">
      <div v-if="isEditing" class="neuro-image-cropper-modal">
        <div class="neuro-image-cropper-modal-overlay" @click="cancelEdit"></div>

        <div class="neuro-image-cropper-modal-container">
          <!-- 弹窗头部 -->
          <div class="neuro-image-cropper-modal-header">
            <h3 class="neuro-image-cropper-modal-title">{{ editorTitle }}</h3>
            <button class="neuro-image-cropper-modal-close" @click="cancelEdit">×</button>
          </div>

          <!-- 弹窗内容 -->
          <div class="neuro-image-cropper-modal-content">
          <!-- 左侧：图片编辑区域 -->
          <div class="neuro-image-cropper-editor-left">
            <!-- 编辑容器（固定大小） -->
            <div class="neuro-image-cropper-editor-container" ref="editorContainer">
              <!-- 图片容器（可移动、缩放、旋转） -->
              <div
                class="image-container"
                ref="imageContainer"
                :style="imageContainerStyle"
                @mousedown="startImageDrag"
                @touchstart="startImageDrag"
              >
                <img
                  ref="imageElement"
                  :src="imageSrc"
                  alt="编辑图片"
                  class="editing-image"
                  :style="imageStyle"
                  @load="onImageLoad"
                />
              </div>

              <!-- 固定裁剪框（不可移动） -->
              <div class="fixed-crop-frame" :style="cropFrameStyle">
                <!-- 裁剪框形状 -->
                <div
                  class="crop-frame-inner"
                  :class="{
                    'crop-square': props.usageType === 'logo',
                    'crop-circle': props.usageType === 'avatar',
                    'crop-rectangle': props.usageType === 'banner',
                    'crop-free': props.usageType === 'custom'
                  }"
                >
                  <!-- 裁剪框网格线 -->
                  <div v-if="props.usageType !== 'avatar'" class="crop-grid">
                    <div class="grid-line vertical"></div>
                    <div class="grid-line horizontal"></div>
                  </div>

                  <!-- 圆形裁剪框的十字线 -->
                  <div v-if="props.usageType === 'avatar'" class="circle-cross">
                    <div class="cross-line vertical"></div>
                    <div class="cross-line horizontal"></div>
                  </div>

                  <!-- 裁剪框尺寸显示 -->
                  <div class="crop-size-display">
                    {{ getCropSizeDisplay() }}
                  </div>
                </div>

                <!-- 调试信息：显示裁剪框坐标 -->
                <!-- 暂时注释掉调试信息 -->
                <!--
                <div v-if="isEditing" class="debug-info">
                  <div class="debug-coord">框X: {{ cropFrame.value.x.toFixed(0) }}</div>
                  <div class="debug-coord">框Y: {{ cropFrame.value.y.toFixed(0) }}</div>
                  <div class="debug-size">{{ cropFrame.value.width }}×{{ cropFrame.value.height }}</div>
                </div>
                -->
              </div>

              <!-- 调试信息：显示图片位置 -->
              <!--
              <div v-if="isEditing" class="debug-image-info">
                <div class="debug-coord">图X: {{ imageState.value.x.toFixed(0) }}</div>
                <div class="debug-coord">图Y: {{ imageState.value.y.toFixed(0) }}</div>
                <div class="debug-coord">缩放: {{ (imageState.value.scale * 100).toFixed(0) }}%</div>
                <div class="debug-coord">原始: {{ imageState.value.naturalWidth }}×{{ imageState.value.naturalHeight }}</div>
              </div>
              -->

              <!-- 遮罩层（框外暗色区域） -->
              <div class="crop-mask">
                <div class="mask-outside top"></div>
                <div class="mask-outside bottom"></div>
                <div class="mask-outside left"></div>
                <div class="mask-outside right"></div>
              </div>
            </div>

            <!-- 编辑区域提示 -->
            <div class="editor-hint">
              <p>💡 提示：拖动图片调整位置，使用右侧工具栏缩放、旋转图片</p>
            </div>
          </div>

          <!-- 右侧：工具栏区域 -->
          <div class="toolbar-right">
            <div class="toolbar-section">
              <h4 class="toolbar-title">裁剪设置</h4>

              <!-- 裁剪尺寸信息 -->
              <div class="crop-info">
                <div class="info-item">
                  <span class="info-label">裁剪尺寸:</span>
                  <span class="info-value">{{ cropFrame.width }} × {{ cropFrame.height }} px</span>
                </div>
                <div class="info-item">
                  <span class="info-label">图片尺寸:</span>
                  <span class="info-value">{{ imageState.naturalWidth }} × {{ imageState.naturalHeight }} px</span>
                </div>
              </div>
            </div>

            <div class="toolbar-section">
              <h4 class="toolbar-title">缩放控制</h4>

              <div class="zoom-controls">
                <div class="zoom-display">
                  <span class="zoom-label">缩放:</span>
                  <span class="zoom-value">{{ Math.round(imageState.scale * 100) }}%</span>
                </div>

                <div class="zoom-slider-container">
                  <button class="zoom-btn zoom-out" @click="zoomOut">−</button>
                  <input
                    type="range"
                    v-model.number="imageState.scale"
                    :min="minScale"
                    :max="maxScale"
                    step="0.01"
                    class="zoom-slider"
                    @input="updateImageScale"
                  />
                  <button class="zoom-btn zoom-in" @click="zoomIn">+</button>
                </div>

                <div class="zoom-presets">
                  <button class="zoom-preset-btn" @click="setZoom(0.5)">50%</button>
                  <button class="zoom-preset-btn" @click="setZoom(1)">100%</button>
                  <button class="zoom-preset-btn" @click="setZoom(2)">200%</button>
                </div>
              </div>
            </div>

            <div class="toolbar-section">
              <h4 class="toolbar-title">图片操作</h4>

              <div class="image-operations">
                <div class="operation-row">
                  <button class="operation-btn" @click="rotateImage(-90)">
                    <span class="btn-icon">↺</span>
                    <span class="btn-text">左旋转</span>
                  </button>
                  <button class="operation-btn" @click="rotateImage(90)">
                    <span class="btn-icon">↻</span>
                    <span class="btn-text">右旋转</span>
                  </button>
                </div>

                <div class="operation-row">
                  <button class="operation-btn" @click="flipImage('horizontal')">
                    <span class="btn-icon">⇄</span>
                    <span class="btn-text">水平翻转</span>
                  </button>
                  <button class="operation-btn" @click="flipImage('vertical')">
                    <span class="btn-icon">⇅</span>
                    <span class="btn-text">垂直翻转</span>
                  </button>
                </div>

                <div class="operation-row">
                  <button class="operation-btn reset-btn" @click="resetImage">
                    <span class="btn-icon">↺</span>
                    <span class="btn-text">重置图片</span>
                  </button>
                </div>
              </div>
            </div>

            <div class="toolbar-section">
              <h4 class="toolbar-title">裁剪操作</h4>

              <div class="crop-operations">
                <div class="operation-row">
                  <button class="operation-btn crop-reset-btn" @click="resetCrop">
                    <span class="btn-icon">🗑️</span>
                    <span class="btn-text">重置裁剪</span>
                  </button>
                </div>

                <div class="operation-row">
                  <button class="operation-btn crop-center-btn" @click="centerImage">
                    <span class="btn-icon">◎</span>
                    <span class="btn-text">居中图片</span>
                  </button>
                </div>

                <!-- 调试按钮 -->
                <!--
                <div class="operation-row" v-if="isEditing">
                  <button class="operation-btn debug-btn" @click="testEmitEvent" style="background: #ff6b6b;">
                    <span class="btn-icon">🐛</span>
                    <span class="btn-text">测试事件触发</span>
                  </button>
                </div>

                <!-- 坐标测试按钮 -->
                <!--
                <div class="operation-row" v-if="isEditing">
                  <button class="operation-btn debug-btn" @click="testCoordinates" style="background: #ffa500;">
                    <span class="btn-icon">📐</span>
                    <span class="btn-text">测试坐标计算</span>
                  </button>
                </div>
                -->
              </div>
            </div>
          </div>
        </div>

        <!-- 弹窗底部 -->
        <div class="modal-footer">
          <button class="modal-btn cancel-btn" @click="cancelEdit">
            取消
          </button>
          <button class="modal-btn confirm-btn" @click="confirmCrop" :disabled="isProcessing">
            {{ isProcessing ? '裁剪中...' : '确认裁剪' }}
          </button>
        </div>
      </div>
    </div>
</Teleport>

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
import { silentDebug } from '@/utils/debug'
import { ref, computed, onUnmounted } from 'vue'

interface Props {
  // 使用场景类型
  usageType?: 'logo' | 'avatar' | 'banner' | 'custom'
  // 预设裁剪尺寸
  presetWidth?: number
  presetHeight?: number
  // 编辑器标题
  editorTitle?: string
  // 交互模式：'standalone' - 独立按钮，'hover' - 悬停显示
  interactionMode?: 'standalone' | 'hover'
  // 悬停时显示的文本
  hoverText?: string
  // 父组件宽度（用于悬停模式）
  parentWidth?: number
  // 父组件高度（用于悬停模式）
  parentHeight?: number
}

const props = withDefaults(defineProps<Props>(), {
  usageType: 'custom',
  presetWidth: 200,
  presetHeight: 200,
  editorTitle: '图片裁剪',
  interactionMode: 'standalone',
  hoverText: '+上传图片',
  parentWidth: 200,
  parentHeight: 200
})

const emit = defineEmits<{
  'update:image': [imageData: string]
  'upload-success': [imageData: string]
  'upload-error': [error: string]
  'cancel': []
}>()

// DOM 引用
const fileInput = ref<HTMLInputElement | null>(null)
const imageElement = ref<HTMLImageElement | null>(null)
const editorContainer = ref<HTMLElement | null>(null)

// 状态
const imageSrc = ref('')
const isEditing = ref(false)
const isProcessing = ref(false)
const isHovering = ref(false)

// 图片状态
const imageState = ref({
  naturalWidth: 0,
  naturalHeight: 0,
  x: 0,
  y: 0,
  scale: 1,
  rotation: 0,
  flipX: 1,
  flipY: 1
})

// 裁剪框状态（固定位置和大小）
const cropFrame = ref({
  x: 0,
  y: 0,
  width: props.presetWidth,
  height: props.presetHeight
})

// 交互状态
const isDragging = ref(false)
const dragStart = ref({ x: 0, y: 0 })
const imageStart = ref({ x: 0, y: 0 })

// 缩放限制
const minScale = ref(0.01)  // 1%
const maxScale = ref(3)

// 容器尺寸
const containerSize = ref({ width: 0, height: 0 })

// 用于滑块缩放的状态
const lastScale = ref(1)  // 保存上一次的缩放值

// 计算属性
const imageContainerStyle = computed<Record<string, string>>(() => ({
  position: 'absolute', // 明确指定position
  top: '0px',           // 从左上角开始
  left: '0px',          // 从左上角开始
  transform: `
    translate(${imageState.value.x}px, ${imageState.value.y}px)
    scale(${imageState.value.scale})
    rotate(${imageState.value.rotation}deg)
    scaleX(${imageState.value.flipX})
    scaleY(${imageState.value.flipY})
  `,
  transformOrigin: '0 0' // 保持左上角为变换原点
}))

const imageStyle = computed(() => ({
  width: `${imageState.value.naturalWidth}px`,
  height: `${imageState.value.naturalHeight}px`
}))

const cropFrameStyle = computed(() => ({
  left: `${cropFrame.value.x}px`,
  top: `${cropFrame.value.y}px`,
  width: `${cropFrame.value.width}px`,
  height: `${cropFrame.value.height}px`
}))

// 悬停容器样式
const hoverContainerStyle = computed(() => ({
  width: `${props.parentWidth}px`,
  height: `${props.parentHeight}px`
}))

// 触发文件选择
const triggerFileInput = () => {
  if (fileInput.value) {
    fileInput.value.click()
  }
}

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

    // 阻止body滚动，防止闪烁
    document.body.style.overflow = 'hidden'
    document.body.style.position = 'fixed'
    document.body.style.width = '100%'

    // 重置状态
    imageState.value = {
      naturalWidth: 0,
      naturalHeight: 0,
      x: 0,
      y: 0,
      scale: 1,
      rotation: 0,
      flipX: 1,
      flipY: 1
    }

    // 重置lastScale
    lastScale.value = 1

    // 重置裁剪框
    resetCrop()
  }
  reader.readAsDataURL(file)

  // 重置文件输入
  if (target) {
    target.value = ''
  }
}

// 图片加载完成
const onImageLoad = () => {
  if (!imageElement.value || !editorContainer.value) return

  const img = imageElement.value
  const container = editorContainer.value

  // 保存原始尺寸
  const naturalWidth = img.naturalWidth
  const naturalHeight = img.naturalHeight
  imageState.value.naturalWidth = naturalWidth
  imageState.value.naturalHeight = naturalHeight

  // 获取容器尺寸
  const containerWidth = container.clientWidth
  const containerHeight = container.clientHeight
  containerSize.value = {
    width: containerWidth,
    height: containerHeight
  }

  silentDebug('🖼️ 图片加载完成:', {
    图片尺寸: `${naturalWidth}×${naturalHeight}`,
    容器尺寸: `${containerWidth}×${containerHeight}`,
    裁剪框尺寸: `${cropFrame.value.width}×${cropFrame.value.height}`
  })

  // 初始化裁剪框位置（居中）
  cropFrame.value.x = (containerWidth - cropFrame.value.width) / 2
  cropFrame.value.y = (containerHeight - cropFrame.value.height) / 2

  silentDebug('🎯 裁剪框位置:', { x: cropFrame.value.x, y: cropFrame.value.y })

  // 智能居中图片：确保裁剪框在图片内部
  smartCenterImage()
}

// 开始拖动图片
const startImageDrag = (e: MouseEvent | TouchEvent) => {
  e.preventDefault()
  isDragging.value = true

  const clientX = 'touches' in e ? e.touches[0].clientX : e.clientX
  const clientY = 'touches' in e ? e.touches[0].clientY : e.clientY

  dragStart.value = { x: clientX, y: clientY }
  imageStart.value = { x: imageState.value.x, y: imageState.value.y }

  // 添加事件监听器
  document.addEventListener('mousemove', handleImageDrag)
  document.addEventListener('touchmove', handleImageDrag, { passive: false })
  document.addEventListener('mouseup', stopImageDrag)
  document.addEventListener('touchend', stopImageDrag)
}

// 处理图片拖动
const handleImageDrag = (e: MouseEvent | TouchEvent) => {
  if (!isDragging.value) return

  e.preventDefault()

  const clientX = 'touches' in e ? e.touches[0].clientX : e.clientX
  const clientY = 'touches' in e ? e.touches[0].clientY : e.clientY

  const deltaX = clientX - dragStart.value.x
  const deltaY = clientY - dragStart.value.y

  imageState.value.x = imageStart.value.x + deltaX
  imageState.value.y = imageStart.value.y + deltaY
}

// 停止拖动图片
const stopImageDrag = () => {
  isDragging.value = false

  // 移除事件监听器
  document.removeEventListener('mousemove', handleImageDrag)
  document.removeEventListener('touchmove', handleImageDrag)
  document.removeEventListener('mouseup', stopImageDrag)
  document.removeEventListener('touchend', stopImageDrag)
}

// 缩放控制 - 基于当前鼠标位置或图片中心进行缩放
const zoomIn = () => {
  // 基于图片中心进行缩放
  const oldScale = imageState.value.scale
  const newScale = Math.min(maxScale.value, oldScale + 0.01)  // +1%

  if (newScale !== oldScale) {
    // 当transformOrigin为0 0时，缩放基于左上角
    // 为了模拟中心缩放，需要调整位置

    // 计算图片中心点
    const centerX = imageState.value.x + (imageState.value.naturalWidth * oldScale) / 2
    const centerY = imageState.value.y + (imageState.value.naturalHeight * oldScale) / 2

    // 新的左上角位置 = 中心点 - (新尺寸/2)
    imageState.value.x = centerX - (imageState.value.naturalWidth * newScale) / 2
    imageState.value.y = centerY - (imageState.value.naturalHeight * newScale) / 2
    imageState.value.scale = newScale
    lastScale.value = newScale
  }
}

const zoomOut = () => {
  // 基于图片中心进行缩放
  const oldScale = imageState.value.scale
  const newScale = Math.max(minScale.value, oldScale - 0.01)  // -1%

  if (newScale !== oldScale) {
    // 计算图片中心点
    const centerX = imageState.value.x + (imageState.value.naturalWidth * oldScale) / 2
    const centerY = imageState.value.y + (imageState.value.naturalHeight * oldScale) / 2

    // 新的左上角位置 = 中心点 - (新尺寸/2)
    imageState.value.x = centerX - (imageState.value.naturalWidth * newScale) / 2
    imageState.value.y = centerY - (imageState.value.naturalHeight * newScale) / 2
    imageState.value.scale = newScale
    lastScale.value = newScale
  }
}

const setZoom = (scale: number) => {
  const oldScale = imageState.value.scale
  const newScale = Math.max(minScale.value, Math.min(maxScale.value, scale))

  silentDebug('🔍 setZoom调用:', { oldScale, newScale, currentX: imageState.value.x, currentY: imageState.value.y })

  if (newScale !== oldScale) {
    // 计算图片中心点
    const centerX = imageState.value.x + (imageState.value.naturalWidth * oldScale) / 2
    const centerY = imageState.value.y + (imageState.value.naturalHeight * oldScale) / 2

    silentDebug('📐 中心点计算:', { centerX, centerY, naturalWidth: imageState.value.naturalWidth, naturalHeight: imageState.value.naturalHeight })

    // 新的左上角位置 = 中心点 - (新尺寸/2)
    const newX = centerX - (imageState.value.naturalWidth * newScale) / 2
    const newY = centerY - (imageState.value.naturalHeight * newScale) / 2

    silentDebug('📍 新位置计算:', { newX, newY, oldX: imageState.value.x, oldY: imageState.value.y })

    imageState.value.x = newX
    imageState.value.y = newY
    imageState.value.scale = newScale
    lastScale.value = newScale

    silentDebug('✅ 缩放完成:', { x: imageState.value.x, y: imageState.value.y, scale: imageState.value.scale })
  }
}

const updateImageScale = (event: Event) => {
  const target = event.target as HTMLInputElement
  const newScale = parseFloat(target.value)
  const oldScale = lastScale.value

  silentDebug('🔍 updateImageScale调用:', { oldScale, newScale, currentX: imageState.value.x, currentY: imageState.value.y })

  if (Math.abs(newScale - oldScale) > 0.0001) { // 使用容差比较
    // 计算图片中心点
    const centerX = imageState.value.x + (imageState.value.naturalWidth * oldScale) / 2
    const centerY = imageState.value.y + (imageState.value.naturalHeight * oldScale) / 2

    silentDebug('📐 中心点计算:', { centerX, centerY })

    // 新的左上角位置 = 中心点 - (新尺寸/2)
    const newX = centerX - (imageState.value.naturalWidth * newScale) / 2
    const newY = centerY - (imageState.value.naturalHeight * newScale) / 2

    silentDebug('📍 新位置计算:', { newX, newY, oldX: imageState.value.x, oldY: imageState.value.y })

    imageState.value.x = newX
    imageState.value.y = newY
    imageState.value.scale = newScale
    lastScale.value = newScale

    silentDebug('✅ 滑块缩放完成:', { x: imageState.value.x, y: imageState.value.y, scale: imageState.value.scale })
  }
}

// 旋转图片
const rotateImage = (degrees: number) => {
  imageState.value.rotation = (imageState.value.rotation + degrees) % 360
}

// 翻转图片
const flipImage = (direction: 'horizontal' | 'vertical') => {
  if (direction === 'horizontal') {
    imageState.value.flipX = -imageState.value.flipX
  } else {
    imageState.value.flipY = -imageState.value.flipY
  }
}

// 重置图片
const resetImage = () => {
  imageState.value = {
    ...imageState.value,
    x: 0,
    y: 0,
    scale: 1,
    rotation: 0,
    flipX: 1,
    flipY: 1
  }

  // 重置lastScale
  lastScale.value = 1

  // 重置后重新居中图片
  smartCenterImage()
}

// 重置裁剪框
const resetCrop = () => {
  if (!editorContainer.value) return

  // 根据业务场景设置裁剪框尺寸
  if (props.usageType === 'logo') {
    // Logo场景：正方形
    const size = Math.min(props.presetWidth, props.presetHeight, 128)
    cropFrame.value.width = cropFrame.value.height = size
  } else if (props.usageType === 'avatar') {
    // 头像场景：圆形（正方形显示为圆形）
    const size = Math.min(props.presetWidth, props.presetHeight, 150)
    cropFrame.value.width = cropFrame.value.height = size
  } else if (props.usageType === 'banner') {
    // 横幅场景：16:9长方形
    cropFrame.value.width = Math.min(props.presetWidth, 800)
    cropFrame.value.height = Math.round(cropFrame.value.width / (16/9))
  } else {
    // 自定义场景：使用预设尺寸
    cropFrame.value.width = props.presetWidth
    cropFrame.value.height = props.presetHeight
  }

  // 居中裁剪框
  if (editorContainer.value) {
    cropFrame.value.x = (editorContainer.value.clientWidth - cropFrame.value.width) / 2
    cropFrame.value.y = (editorContainer.value.clientHeight - cropFrame.value.height) / 2
  }
}

// 简单居中图片：确保图片覆盖裁剪框
const smartCenterImage = () => {
  if (!editorContainer.value) return

  const containerWidth = editorContainer.value.clientWidth
  const containerHeight = editorContainer.value.clientHeight
  const naturalWidth = imageState.value.naturalWidth
  const naturalHeight = imageState.value.naturalHeight

  silentDebug('🖼️ 简单居中开始:', {
    容器: `${containerWidth}×${containerHeight}`,
    图片: `${naturalWidth}×${naturalHeight}`,
    裁剪框: `${cropFrame.value.width}×${cropFrame.value.height}`
  })

  // 1. 确保裁剪框在容器内（居中）
  cropFrame.value.x = (containerWidth - cropFrame.value.width) / 2
  cropFrame.value.y = (containerHeight - cropFrame.value.height) / 2

  silentDebug('🎯 裁剪框位置:', { x: cropFrame.value.x, y: cropFrame.value.y })

  // 2. 计算需要的缩放，让图片至少覆盖裁剪框
  const neededScaleX = cropFrame.value.width / naturalWidth
  const neededScaleY = cropFrame.value.height / naturalHeight
  const neededScale = Math.max(neededScaleX, neededScaleY) * 1.5 // 增加50%边距

  if (imageState.value.scale < neededScale) {
    imageState.value.scale = neededScale
    silentDebug(`📏 设置缩放: ${neededScale.toFixed(3)}`)
  }

  // 3. 计算显示尺寸
  const displayWidth = naturalWidth * imageState.value.scale
  const displayHeight = naturalHeight * imageState.value.scale

  // 4. 简单位置：让裁剪框在图片中心
  // imageState.value.x和y是图片左上角的位置
  imageState.value.x = cropFrame.value.x + cropFrame.value.width / 2 - displayWidth / 2
  imageState.value.y = cropFrame.value.y + cropFrame.value.height / 2 - displayHeight / 2

  silentDebug('✅ 最终设置:', {
    图片位置: `(${imageState.value.x.toFixed(1)}, ${imageState.value.y.toFixed(1)})`,
    显示尺寸: `${displayWidth.toFixed(1)}×${displayHeight.toFixed(1)}`,
    缩放: imageState.value.scale.toFixed(3),
    裁剪框位置: `(${cropFrame.value.x}, ${cropFrame.value.y})`,
    裁剪框尺寸: `${cropFrame.value.width}×${cropFrame.value.height}`,
    图片中心点: `(${imageState.value.x + displayWidth / 2}, ${imageState.value.y + displayHeight / 2})`,
    裁剪框中心点: `(${cropFrame.value.x + cropFrame.value.width / 2}, ${cropFrame.value.y + cropFrame.value.height / 2})`
  })
}

// 居中图片（旧版本，保持兼容性）
const centerImage = () => {
  smartCenterImage()
}

// 获取裁剪尺寸显示
const getCropSizeDisplay = () => {
  if (props.usageType === 'logo') {
    return `${cropFrame.value.width}×${cropFrame.value.width} (正方形)`
  } else if (props.usageType === 'avatar') {
    return `${cropFrame.value.width}×${cropFrame.value.width} (圆形)`
  } else if (props.usageType === 'banner') {
    return `${cropFrame.value.width}×${cropFrame.value.height} (16:9)`
  } else {
    return `${cropFrame.value.width}×${cropFrame.value.height}`
  }
}

// 确认裁剪 - 优化版本，提高图片质量
const confirmCrop = async () => {
  if (!imageElement.value || isProcessing.value) return

  isProcessing.value = true

  try {
    silentDebug('🔄 开始裁剪，优化图片质量...')

    // 获取图片原始尺寸
    const imgNaturalWidth = imageState.value.naturalWidth
    const imgNaturalHeight = imageState.value.naturalHeight
    const scale = imageState.value.scale

    silentDebug('📐 图片信息:', {
      原始尺寸: `${imgNaturalWidth}×${imgNaturalHeight}`,
      当前缩放: scale,
      裁剪框尺寸: `${cropFrame.value.width}×${cropFrame.value.height}`
    })

    // 简单直接的计算方法
    const imgLeft = imageState.value.x
    const imgTop = imageState.value.y

    // 裁剪框在原始图片中的位置和尺寸
    const imgX = (cropFrame.value.x - imgLeft) / scale
    const imgY = (cropFrame.value.y - imgTop) / scale
    const imgWidth = cropFrame.value.width / scale
    const imgHeight = cropFrame.value.height / scale

    silentDebug('🎯 简单裁剪计算:', {
      图片位置: `(${imgLeft.toFixed(1)}, ${imgTop.toFixed(1)})`,
      显示尺寸: `${(imgNaturalWidth * scale).toFixed(1)}×${(imgNaturalHeight * scale).toFixed(1)}`,
      裁剪框: `(${cropFrame.value.x}, ${cropFrame.value.y}) ${cropFrame.value.width}×${cropFrame.value.height}`,
      源区域: `(${imgX.toFixed(1)}, ${imgY.toFixed(1)}) ${imgWidth.toFixed(1)}×${imgHeight.toFixed(1)}`,
      变换原点: 'center center',
      图片中心点: `(${imgLeft + (imgNaturalWidth * scale) / 2}, ${imgTop + (imgNaturalHeight * scale) / 2})`
    })

    // 计算实际可用的裁剪区域（允许部分超出）
    const sourceX = Math.max(0, imgX)
    const sourceY = Math.max(0, imgY)
    const sourceWidth = Math.min(imgWidth, imgNaturalWidth - sourceX)
    const sourceHeight = Math.min(imgHeight, imgNaturalHeight - sourceY)

    // 计算裁剪框超出图片的部分
    const overflowLeft = Math.max(0, -imgX)
    const overflowTop = Math.max(0, -imgY)
    const overflowRight = Math.max(0, imgX + imgWidth - imgNaturalWidth)
    const overflowBottom = Math.max(0, imgY + imgHeight - imgNaturalHeight)

    silentDebug('📏 裁剪区域分析:', {
      源起点: `(${sourceX.toFixed(2)}, ${sourceY.toFixed(2)})`,
      源尺寸: `${sourceWidth.toFixed(2)}×${sourceHeight.toFixed(2)}`,
      超出部分: `左:${overflowLeft.toFixed(2)}, 上:${overflowTop.toFixed(2)}, 右:${overflowRight.toFixed(2)}, 下:${overflowBottom.toFixed(2)}`,
      是否完全超出: sourceWidth <= 0 || sourceHeight <= 0 ? '是' : '否'
    })

    // 检查是否有有效的裁剪区域
    if (sourceWidth <= 0 || sourceHeight <= 0) {
      console.error('❌ 没有有效的裁剪区域（裁剪框完全超出图片）')
      throw new Error('裁剪框完全超出图片范围，请调整图片位置')
    }

    // 创建高质量Canvas
    const canvas = document.createElement('canvas')
    const ctx = canvas.getContext('2d', {
      alpha: false, // 禁用alpha通道，提高性能
      willReadFrequently: false
    })

    if (!ctx) {
      throw new Error('无法创建Canvas上下文')
    }

    // 关键优化1：提高Canvas分辨率（防止锯齿）
    // 根据裁剪框尺寸和设备像素比调整Canvas大小
    const devicePixelRatio = window.devicePixelRatio || 1
    const targetWidth = cropFrame.value.width
    const targetHeight = cropFrame.value.height

    // 设置Canvas为高分辨率
    canvas.width = targetWidth * devicePixelRatio
    canvas.height = targetHeight * devicePixelRatio

    // 设置Canvas显示尺寸为原始尺寸
    canvas.style.width = `${targetWidth}px`
    canvas.style.height = `${targetHeight}px`

    silentDebug('🎨 Canvas设置:', {
      逻辑尺寸: `${targetWidth}×${targetHeight}`,
      实际尺寸: `${canvas.width}×${canvas.height}`,
      设备像素比: devicePixelRatio
    })

    // 关键优化2：设置高质量绘制选项
    ctx.imageSmoothingEnabled = true
    ctx.imageSmoothingQuality = 'high' // 使用高质量图像平滑
    ctx.globalCompositeOperation = 'copy'

    // 关键优化3：缩放Canvas上下文以匹配高分辨率
    ctx.scale(devicePixelRatio, devicePixelRatio)

    // 应用图片变换（旋转、翻转）
    ctx.save()
    ctx.translate(targetWidth / 2, targetHeight / 2)
    ctx.rotate((imageState.value.rotation * Math.PI) / 180)
    ctx.scale(imageState.value.flipX, imageState.value.flipY)
    ctx.translate(-targetWidth / 2, -targetHeight / 2)

    try {
      // 关键优化4：使用整数坐标绘制，避免亚像素渲染
      const drawX = 0
      const drawY = 0
      const drawWidth = targetWidth
      const drawHeight = targetHeight

      silentDebug('🖼️ 开始绘制图片...')
      silentDebug('源区域:', `(${sourceX.toFixed(2)}, ${sourceY.toFixed(2)}) ${sourceWidth.toFixed(2)}×${sourceHeight.toFixed(2)}`)
      silentDebug('目标区域:', `(${drawX}, ${drawY}) ${drawWidth}×${drawHeight}`)

      // 先填充白色背景（如果使用JPEG格式，需要不透明背景）
      ctx.fillStyle = '#ffffff'
      ctx.fillRect(drawX, drawY, drawWidth, drawHeight)

      // 计算目标位置（考虑裁剪框超出图片的部分）
      // 如果裁剪框左侧超出图片，图片在目标Canvas中应该右移
      const targetX = drawX + (overflowLeft / imgWidth) * drawWidth
      const targetY = drawY + (overflowTop / imgHeight) * drawHeight

      // 计算目标尺寸（按比例缩放）
      const finalTargetWidth = (sourceWidth / imgWidth) * drawWidth
      const finalTargetHeight = (sourceHeight / imgHeight) * drawHeight

      silentDebug('🎯 绘制参数:', {
        目标位置: `(${targetX.toFixed(2)}, ${targetY.toFixed(2)})`,
        目标尺寸: `${finalTargetWidth.toFixed(2)}×${finalTargetHeight.toFixed(2)}`,
        超出处理: `左移:${overflowLeft.toFixed(2)}px, 上移:${overflowTop.toFixed(2)}px`
      })

      // 绘制裁剪区域
      ctx.drawImage(
        imageElement.value,
        sourceX,      // 源X（图片内的起点）
        sourceY,      // 源Y（图片内的起点）
        sourceWidth,  // 源宽度（图片内的宽度）
        sourceHeight, // 源高度（图片内的高度）
        targetX,      // 目标X（考虑超出部分）
        targetY,      // 目标Y（考虑超出部分）
        finalTargetWidth,  // 目标宽度（按比例缩放）
        finalTargetHeight  // 目标高度（按比例缩放）
      )

      ctx.restore()
      silentDebug('✅ 图片绘制成功')

    } catch (drawError) {
      ctx.restore()
      console.error('❌ 图片绘制失败:', drawError)
      throw new Error(`图片绘制失败: ${drawError}`)
    }

    // 关键优化5：使用高质量JPEG格式（文件更小，质量更好）
    // 对于照片类图片，JPEG通常比PNG质量更好且文件更小
    let croppedImageData
    try {
      // 尝试使用JPEG格式，质量为0.92（高质量）
      croppedImageData = canvas.toDataURL('image/jpeg', 0.92)
      silentDebug('✅ 使用JPEG格式，质量: 0.92')
    } catch (jpegError) {
      // 如果JPEG失败，回退到PNG
      silentDebug('🔄 JPEG失败，回退到PNG格式')
      croppedImageData = canvas.toDataURL('image/png')
    }

    silentDebug('🔥 裁剪完成，图片数据长度:', croppedImageData.length)
    silentDebug('📊 图片格式:', croppedImageData.substring(0, 20))

    // 发出事件
    emit('update:image', croppedImageData)
    silentDebug('✅ update:image 事件已触发')

    emit('upload-success', croppedImageData)
    silentDebug('✅ upload-success 事件已触发')

    // 等待一小段时间，确保事件被处理
    setTimeout(() => {
      silentDebug('⏰ 关闭编辑器')
      cancelEdit()
    }, 100)

  } catch (error) {
    console.error('❌ 裁剪失败:', error)
    emit('upload-error', error instanceof Error ? error.message : '裁剪失败')
  } finally {
    isProcessing.value = false
  }
}

// 取消编辑
const cancelEdit = () => {
  isEditing.value = false
  imageSrc.value = ''

  // 恢复body滚动
  document.body.style.overflow = ''
  document.body.style.position = ''
  document.body.style.width = ''

  emit('cancel')
}



// 清理事件监听器
onUnmounted(() => {
  stopImageDrag()

  // 确保在组件卸载时恢复body滚动
  document.body.style.overflow = ''
  document.body.style.position = ''
  document.body.style.width = ''
})

// 公开方法
defineExpose({
  triggerFileInput
})
</script>

<style scoped>
.neuro-image-cropper {
  font-family: 'JetBrains Mono', monospace;
}

/* 上传触发器 */
.upload-trigger {
  cursor: pointer;
  transition: all 0.3s ease;
}

.upload-trigger:hover {
  opacity: 0.8;
}

.default-trigger {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 16px;
  background: var(--nm-bg-surface);
  border: 2px dashed var(--nm-border);
  border-radius: 12px;
  transition: all 0.3s ease;
}

.default-trigger:hover {
  border-color: var(--nm-primary);
  background: rgba(0, 245, 212, 0.05);
}

/* 悬停模式 */
.hover-container {
  position: relative;
  cursor: pointer;
  overflow: hidden;
  border-radius: 12px;
  background: var(--nm-bg-surface);
  border: 2px solid var(--nm-border);
  transition: all 0.3s ease;
  /* 移除测试背景，使用正常背景 */
}

.hover-container:hover {
  border-color: var(--nm-primary);
}

.default-hover-content {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.hover-placeholder {
  text-align: center;
  color: var(--nm-text-secondary);
}

.placeholder-icon {
  font-size: 32px;
  opacity: 0.6;
}

/* 悬停时显示的 + 号按钮（在父组件内部） */
.hover-plus-button {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  background: rgba(0, 245, 212, 0.9); /* 半透明背景 */
  color: var(--nm-text-on-dark);
  border-radius: 50%;
  width: 60px;
  height: 60px;
  cursor: pointer;
  z-index: 10;
  animation: fadeInScale 0.3s ease forwards;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
  transition: all 0.3s ease;
}

.hover-plus-button:hover {
  transform: translate(-50%, -50%) scale(1.1);
  background: rgba(0, 245, 212, 1);
  box-shadow: 0 6px 16px rgba(0, 0, 0, 0.4);
}

@keyframes fadeInScale {
  from {
    opacity: 0;
    transform: translate(-50%, -50%) scale(0.8);
  }
  to {
    opacity: 1;
    transform: translate(-50%, -50%) scale(1);
  }
}

.plus-icon {
  font-size: 28px;
  font-weight: bold;
  line-height: 1;
}

.plus-text {
  font-size: 11px;
  font-weight: 600;
  margin-top: 2px;
  text-align: center;
  line-height: 1.2;
}

.trigger-icon {
  font-size: 32px;
  margin-bottom: 8px;
}

.trigger-text {
  color: var(--nm-text-primary);
  font-size: 14px;
  font-weight: 600;
}

/* 编辑弹窗 - 修复闪烁和居中问题 */
.neuro-image-cropper-modal {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  z-index: 9999;
  display: flex;
  align-items: center;
  justify-content: center;
  pointer-events: none;
}

.neuro-image-cropper-modal-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.85);
  backdrop-filter: blur(8px);
  pointer-events: auto;
}

.neuro-image-cropper-modal-container {
  position: relative;
  width: 90%;
  max-width: 1000px;
  max-height: 85vh;
  background: #0a0a1a;
  border-radius: 16px;
  overflow: hidden;
  box-shadow: 0 25px 50px rgba(0, 0, 0, 0.6);
  border: 1px solid rgba(0, 245, 212, 0.3);
  display: flex;
  flex-direction: column;
  z-index: 10000;
  pointer-events: auto;
  transform: translateZ(0); /* 硬件加速，减少闪烁 */
}

/* 弹窗头部 - 压缩版 */
.neuro-image-cropper-modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 20px;
  background: #0f0f1f;
  border-bottom: 1px solid rgba(0, 245, 212, 0.2);
  flex-shrink: 0;
}

.neuro-image-cropper-modal-title {
  color: #00f5d4;
  font-size: 18px;
  font-weight: 600;
  margin: 0;
  font-family: 'JetBrains Mono', monospace;
}

.neuro-image-cropper-modal-close {
  width: 32px;
  height: 32px;
  border: none;
  background: transparent;
  color: #8a8aff;
  font-size: 24px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 6px;
  transition: all 0.2s ease;
}

.neuro-image-cropper-modal-close:hover {
  background: rgba(0, 245, 212, 0.1);
  color: #00f5d4;
}

/* 弹窗内容 */
.neuro-image-cropper-modal-content {
  flex: 1;
  display: flex;
  overflow: hidden;
  min-height: 0; /* 防止内容溢出 */
}

/* 左侧编辑区域 - 压缩版 */
.neuro-image-cropper-editor-left {
  flex: 3;
  display: flex;
  flex-direction: column;
  padding: 20px;
  border-right: 1px solid rgba(0, 245, 212, 0.2);
  min-height: 0; /* 防止内容溢出 */
}

.neuro-image-cropper-editor-container {
  flex: 1;
  position: relative;
  background: #0a0a15;
  border-radius: 8px;
  overflow: hidden;
  border: 1px solid rgba(0, 245, 212, 0.1);
}

/* 图片容器 */
.image-container {
  position: absolute;
  cursor: grab;
  user-select: none;
  touch-action: none;
}

.image-container:active {
  cursor: grabbing;
}

.editing-image {
  display: block;
  pointer-events: none;
}

/* 固定裁剪框 */
.fixed-crop-frame {
  position: absolute;
  pointer-events: none;
  z-index: 10;
}

.crop-frame-inner {
  width: 100%;
  height: 100%;
  border: 2px solid var(--nm-primary);
  box-shadow: 0 0 0 9999px rgba(0, 0, 0, 0.7);
  position: relative;
  overflow: hidden;
}

/* 裁剪框形状 */
.crop-frame-inner.crop-square {
  /* 正方形样式 */
}

.crop-frame-inner.crop-circle {
  border-radius: 50%;
}

.crop-frame-inner.crop-rectangle {
  aspect-ratio: 16/9;
}

.crop-frame-inner.crop-free {
  /* 自由形状样式 */
}

/* 裁剪框网格线 */
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

/* 圆形裁剪框十字线 */
.circle-cross {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  pointer-events: none;
}

.cross-line {
  position: absolute;
  background: rgba(255, 255, 255, 0.3);
}

.cross-line.vertical {
  top: 0;
  bottom: 0;
  left: 50%;
  width: 1px;
  transform: translateX(-50%);
}

.cross-line.horizontal {
  left: 0;
  right: 0;
  top: 50%;
  height: 1px;
  transform: translateY(-50%);
}

/* 裁剪框尺寸显示 */
.crop-size-display {
  position: absolute;
  bottom: 8px;
  right: 8px;
  background: rgba(0, 0, 0, 0.7);
  color: white;
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 12px;
  font-family: 'JetBrains Mono', monospace;
}

/* 遮罩层 */
.crop-mask {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  pointer-events: none;
}

.mask-outside {
  position: absolute;
  background: rgba(0, 0, 0, 0.7);
  backdrop-filter: blur(2px);
}

.mask-outside.top {
  top: 0;
  left: 0;
  right: 0;
}

.mask-outside.bottom {
  left: 0;
  right: 0;
  bottom: 0;
}

.mask-outside.left {
  top: 0;
  left: 0;
  bottom: 0;
}

.mask-outside.right {
  top: 0;
  right: 0;
  bottom: 0;
}

/* 编辑区域提示 */
.editor-hint {
  margin-top: 12px;
  padding: 12px;
  background: var(--nm-bg-surface);
  border-radius: 8px;
  border-left: 4px solid var(--nm-primary);
}

.editor-hint p {
  margin: 0;
  color: var(--nm-text-secondary);
  font-size: 13px;
}

/* 右侧工具栏 */
.toolbar-right {
  flex: 1;
  min-width: 300px;
  padding: 24px;
  background: var(--nm-bg-card);
  overflow-y: auto;
}

.toolbar-section {
  margin-bottom: 24px;
}

.toolbar-section:last-child {
  margin-bottom: 0;
}

.toolbar-title {
  color: var(--nm-text-primary);
  font-size: 16px;
  font-weight: 600;
  margin: 0 0 16px 0;
  padding-bottom: 8px;
  border-bottom: 1px solid var(--nm-border);
}

/* 裁剪信息 */
.crop-info {
  background: var(--nm-bg-surface);
  border-radius: 8px;
  padding: 16px;
}

.info-item {
  display: flex;
  justify-content: space-between;
  margin-bottom: 8px;
}

.info-item:last-child {
  margin-bottom: 0;
}

.info-label {
  color: var(--nm-text-secondary);
  font-size: 14px;
}

.info-value {
  color: var(--nm-text-primary);
  font-size: 14px;
  font-weight: 600;
  font-family: 'JetBrains Mono', monospace;
}

/* 缩放控制 */
.zoom-controls {
  background: var(--nm-bg-surface);
  border-radius: 8px;
  padding: 16px;
}

.zoom-display {
  display: flex;
  justify-content: space-between;
  margin-bottom: 12px;
}

.zoom-label {
  color: var(--nm-text-secondary);
  font-size: 14px;
}

.zoom-value {
  color: var(--nm-text-primary);
  font-size: 14px;
  font-weight: 600;
  font-family: 'JetBrains Mono', monospace;
}

.zoom-slider-container {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 12px;
}

.zoom-btn {
  width: 32px;
  height: 32px;
  border: none;
  background: var(--nm-bg-deep);
  color: var(--nm-text-primary);
  border-radius: 6px;
  font-size: 18px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s ease;
}

.zoom-btn:hover {
  background: var(--nm-primary);
  color: var(--nm-text-on-dark);
}

.zoom-slider {
  flex: 1;
  height: 6px;
  -webkit-appearance: none;
  appearance: none;
  background: var(--nm-bg-deep);
  border-radius: 3px;
  outline: none;
}

.zoom-slider::-webkit-slider-thumb {
  -webkit-appearance: none;
  appearance: none;
  width: 20px;
  height: 20px;
  background: var(--nm-primary);
  border-radius: 50%;
  cursor: pointer;
  border: 2px solid white;
}

.zoom-slider::-moz-range-thumb {
  width: 20px;
  height: 20px;
  background: var(--nm-primary);
  border-radius: 50%;
  cursor: pointer;
  border: 2px solid white;
}

.zoom-presets {
  display: flex;
  gap: 8px;
}

.zoom-preset-btn {
  flex: 1;
  padding: 8px 12px;
  border: 1px solid var(--nm-border);
  background: var(--nm-bg-deep);
  color: var(--nm-text-secondary);
  border-radius: 6px;
  font-size: 13px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.zoom-preset-btn:hover {
  border-color: var(--nm-primary);
  color: var(--nm-text-primary);
}

/* 图片操作 */
.image-operations {
  background: var(--nm-bg-surface);
  border-radius: 8px;
  padding: 16px;
}

.operation-row {
  display: flex;
  gap: 8px;
  margin-bottom: 12px;
}

.operation-row:last-child {
  margin-bottom: 0;
}

.operation-btn {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 12px;
  border: 1px solid var(--nm-border);
  background: var(--nm-bg-deep);
  color: var(--nm-text-primary);
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.operation-btn:hover {
  border-color: var(--nm-primary);
  background: rgba(0, 245, 212, 0.1);
}

.operation-btn.reset-btn:hover {
  border-color: var(--nm-error);
  background: rgba(239, 68, 68, 0.1);
}

.btn-icon {
  font-size: 20px;
  margin-bottom: 4px;
}

.btn-text {
  font-size: 12px;
  font-weight: 500;
}

/* 裁剪操作 */
.crop-operations {
  background: var(--nm-bg-surface);
  border-radius: 8px;
  padding: 16px;
}

.operation-btn.crop-reset-btn:hover {
  border-color: var(--nm-error);
  background: rgba(239, 68, 68, 0.1);
}

.operation-btn.crop-center-btn:hover {
  border-color: var(--nm-info);
  background: rgba(59, 130, 246, 0.1);
}

/* 弹窗底部 */
.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding: 20px 24px;
  background: var(--nm-bg-card);
  border-top: 1px solid var(--nm-border);
}

.modal-btn {
  padding: 10px 24px;
  border: none;
  border-radius: 8px;
  font-family: 'JetBrains Mono', monospace;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s ease;
}

.cancel-btn {
  background: var(--nm-bg-surface);
  color: var(--nm-text-primary);
  border: 1px solid var(--nm-border);
}

.cancel-btn:hover {
  background: var(--nm-bg-deep);
}

.confirm-btn {
  background: var(--nm-primary);
  color: var(--nm-text-on-dark);
}

.confirm-btn:hover:not(:disabled) {
  background: #00d4b4;
}

.confirm-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* 响应式设计 */
@media (max-width: 1024px) {
  .neuro-image-cropper-modal-content {
    flex-direction: column;
  }

  .neuro-image-cropper-editor-left {
    border-right: none;
    border-bottom: 1px solid var(--nm-border);
  }

  .toolbar-right {
    min-width: auto;
  }
}

@media (max-width: 768px) {
  .modal-container {
    width: 95%;
    max-height: 95vh;
  }

  .operation-row {
    flex-direction: column;
  }
}

/* 调试信息 */
/*
.debug-info {
  position: absolute;
  top: -35px;
  left: 0;
  background: rgba(255, 0, 0, 0.9);
  color: white;
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 11px;
  font-family: 'JetBrains Mono', monospace;
  z-index: 100;
  display: flex;
  gap: 8px;
  white-space: nowrap;
}

.debug-image-info {
  position: absolute;
  top: 10px;
  right: 10px;
  background: rgba(0, 0, 255, 0.9);
  color: white;
  padding: 6px 10px;
  border-radius: 4px;
  font-size: 11px;
  font-family: 'JetBrains Mono', monospace;
  z-index: 100;
  display: flex;
  flex-direction: column;
  gap: 3px;
  line-height: 1.3;
}

.debug-coord {
  white-space: nowrap;
}

.debug-size {
  white-space: nowrap;
  font-weight: bold;
}
*/
</style>