<script setup lang="ts">
import Cropper from 'cropperjs'
import 'cropperjs/dist/cropper.css'
import { onMounted, onUnmounted, ref, watch } from 'vue'
import NeuroAgentDialog from '@/views/components/NeuroAgentDialog.vue'

const props = defineProps<{
  modelValue: boolean
  src: string
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', v: boolean): void
  (e: 'done', dataUrl: string): void
}>()

const imgRef = ref<HTMLImageElement | null>(null)
let cropper: Cropper | null = null

function close() {
  emit('update:modelValue', false)
}

function destroyCropper() {
  if (cropper) {
    cropper.destroy()
    cropper = null
  }
}

function initCropper() {
  const img = imgRef.value
  if (!img) return
  destroyCropper()
  cropper = new Cropper(img, {
    viewMode: 1,
    aspectRatio: 1,
    autoCropArea: 1,
    dragMode: 'move',
    background: false,
    movable: true,
    zoomable: true,
    scalable: false,
    rotatable: false,
    responsive: true,
  })
}

function confirm() {
  if (!cropper) return
  const canvas = cropper.getCroppedCanvas({
    width: 256,
    height: 256,
    imageSmoothingQuality: 'high',
    /** 不先铺底色，避免把透明 PNG 裁成实色底 */
    fillColor: 'transparent',
  })
  /** JPEG 无透明通道，透明区会变成黑底；Logo/透明素材必须用 PNG */
  const dataUrl = canvas.toDataURL('image/png')
  emit('done', dataUrl)
  close()
}

watch(
  () => props.modelValue,
  (v) => {
    if (!v) destroyCropper()
  },
)

onMounted(() => {
  if (props.modelValue) initCropper()
})

onUnmounted(() => {
  destroyCropper()
})
</script>

<template>
  <NeuroAgentDialog
    :model-value="modelValue"
    title="裁剪 Logo"
    icon="✂️"
    size="medium"
    @close="close"
  >
    <div class="cropper-wrap">
      <img v-if="src" ref="imgRef" :src="src" alt="" class="cropper-img" @load="initCropper" />
    </div>
    <template #footer-right>
      <button class="nm-btn nm-btn--primary" @click="confirm">确定</button>
    </template>
  </NeuroAgentDialog>
</template>

<style scoped>
.cropper-wrap {
  width: 100%;
  max-height: 360px;
  border-radius: 8px;
  overflow: hidden;
}
.cropper-img {
  max-width: 100%;
  display: block;
}
</style>

