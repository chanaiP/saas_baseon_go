<template>
  <div class="neuro-avatar-upload">
    <!-- 头像容器 -->
    <div class="avatar-container">
      <div class="avatar-wrapper">
        <!-- 显示上传的头像或默认头像 -->
        <div class="avatar-display">
          <div v-if="avatarUrl" class="avatar-image">
            <img :src="avatarUrl" alt="用户头像" class="avatar-img" />
          </div>
          <div v-else class="avatar-default">
            <div class="default-initials">{{ userInitials }}</div>
          </div>
        </div>

        <!-- 脉冲效果层（始终显示，覆盖在头像上方） -->
        <div class="avatar-pulse-layer">
          <div class="default-glow"></div>
          <div class="default-pulse"></div>
        </div>

        <!-- 上传指示器 -->
        <div v-if="isUploading" class="upload-indicator">
          <div class="upload-spinner"></div>
          <span class="upload-text">上传中...</span>
        </div>
      </div>

      <!-- 图片裁剪组件 -->
      <NeuroImageCropper
        usage-type="avatar"
        :preset-width="150"
        :preset-height="150"
        :aspect-ratio="1"
        :min-size="100"
        :max-size="500"
        editor-title="头像裁剪"
        @update:image="handleAvatarUpdate"
        @upload-success="handleAvatarUploadSuccess"
        @upload-error="handleAvatarUploadError"
        @cancel="handleAvatarUploadCancel"
        style="display: none"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import NeuroImageCropper from './NeuroImageCropper.vue'
import { useAvatarStore } from '@/stores/avatar'

interface Props {
  userId?: string
  userName?: string
  currentAvatar?: string
}

const props = withDefaults(defineProps<Props>(), {
  userId: '',
  userName: '用户',
  currentAvatar: ''
})

const emit = defineEmits<{
  'update:avatar': [avatarUrl: string]
  'upload-success': [avatarUrl: string]
  'upload-error': [error: string]
}>()

// 响应式数据
const avatarStore = useAvatarStore()
const isUploading = ref(false)

// 计算头像URL，优先使用props，然后使用store
const avatarUrl = computed(() => {
  return props.currentAvatar || avatarStore.avatarUrl
})

// 计算用户 initials
const userInitials = computed(() => {
  if (!props.userName) return 'U'
  const names = props.userName.split(' ')
  if (names.length >= 2) {
    return (names[0][0] + names[1][0]).toUpperCase()
  }
  return names[0][0].toUpperCase()
})

// 处理头像更新
const handleAvatarUpdate = (imageData: string) => {
  // 保存到全局存储
  avatarStore.saveAvatarToStorage(props.userId, imageData)

  // 发出事件
  emit('update:avatar', imageData)
  emit('upload-success', imageData)

  console.log('头像已更新:', imageData)
}

// 处理头像上传成功
const handleAvatarUploadSuccess = (imageData: string) => {
  // 保存到全局存储
  avatarStore.saveAvatarToStorage(props.userId, imageData)

  // 发出事件
  emit('update:avatar', imageData)
  emit('upload-success', imageData)

  console.log('头像上传成功:', imageData)
}

// 处理头像上传错误
const handleAvatarUploadError = (error: string) => {
  emit('upload-error', error)
  console.error('头像上传错误:', error)
}

// 处理头像上传取消
const handleAvatarUploadCancel = () => {
  console.log('头像上传已取消')
}

// 组件挂载时加载头像
onMounted(() => {
  if (props.userId) {
    avatarStore.loadAvatarFromStorage(props.userId)
  }
})
</script>

<style scoped>
.neuro-avatar-upload {
  display: inline-block;
}

.avatar-container {
  position: relative;
  cursor: pointer;
  transition: all 0.3s ease;
}

.avatar-container:hover {
  transform: scale(1.05);
}


.avatar-wrapper {
  position: relative;
  width: 80px;
  height: 80px;
  border-radius: 50%;
  overflow: hidden;
  background: var(--nm-bg-card);
  border: 2px solid var(--nm-border-glow);
  transition: all 0.3s ease;
}

.avatar-container:hover .avatar-wrapper {
  border-color: var(--nm-primary);
  box-shadow: 0 0 20px var(--nm-primary-glow);
}

.avatar-display {
  width: 100%;
  height: 100%;
  position: relative;
  z-index: 2;
}

.avatar-image {
  width: 100%;
  height: 100%;
  position: relative;
  z-index: 1;
}

.avatar-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  border-radius: 50%;
}

.avatar-pulse-layer {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  z-index: 3;
  pointer-events: none;
}


.avatar-default {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
  background: linear-gradient(135deg, var(--nm-primary), var(--nm-secondary));
}

.default-initials {
  font-family: 'JetBrains Mono', monospace;
  font-size: 24px;
  font-weight: 600;
  color: var(--nm-text-on-dark);
  z-index: 2;
}

.default-glow {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: radial-gradient(circle, var(--nm-primary-glow) 0%, transparent 70%);
  opacity: 0.5;
  animation: glow-pulse 3s infinite;
}

.default-pulse {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  width: 100%;
  height: 100%;
  border: 2px solid var(--nm-primary);
  border-radius: 50%;
  animation: pulse-ring 2s infinite;
}

@keyframes glow-pulse {
  0%, 100% { opacity: 0.3; }
  50% { opacity: 0.6; }
}

@keyframes pulse-ring {
  0% { transform: translate(-50%, -50%) scale(0.8); opacity: 1; }
  100% { transform: translate(-50%, -50%) scale(1.2); opacity: 0; }
}


.upload-indicator {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.7);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  z-index: 20;
}

.upload-spinner {
  width: 24px;
  height: 24px;
  border: 2px solid var(--nm-border);
  border-top-color: var(--nm-primary);
  border-radius: 50%;
  animation: spin 1s linear infinite;
  margin-bottom: 8px;
}

.upload-text {
  font-size: 12px;
  color: var(--nm-text-primary);
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

/* 预览对话框样式 */
.preview-dialog-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.7);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  backdrop-filter: blur(10px);
}

.preview-dialog {
  background: var(--nm-bg-card);
  border-radius: 16px;
  border: 1px solid var(--nm-border-glow);
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.5);
  width: 90%;
  max-width: 400px;
  overflow: hidden;
}

.preview-header {
  padding: 20px 24px;
  border-bottom: 1px solid var(--nm-border);
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.preview-title {
  font-family: 'JetBrains Mono', monospace;
  font-size: 18px;
  font-weight: 600;
  color: var(--nm-text-primary);
  margin: 0;
}

.preview-close {
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

.preview-close:hover {
  background: var(--nm-bg-surface);
  color: var(--nm-primary);
}

.preview-content {
  padding: 24px;
}

.preview-image-container {
  width: 200px;
  height: 200px;
  margin: 0 auto 24px;
  border-radius: 50%;
  overflow: hidden;
  border: 3px solid var(--nm-border-glow);
}

.preview-image {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.preview-actions {
  display: flex;
  gap: 12px;
  justify-content: center;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .avatar-wrapper {
    width: 60px;
    height: 60px;
  }

  .default-initials {
    font-size: 18px;
  }

  .preview-dialog {
    width: 95%;
  }

  .preview-image-container {
    width: 150px;
    height: 150px;
  }
}
</style>