<template>
  <div class="logo-avatar-editor">
    <div class="editor-header">
      <h1>Logo与头像编辑器</h1>
      <p class="header-description">使用全新的图片裁剪组件修改Logo和头像</p>
    </div>

    <div class="editor-content">
      <!-- Logo编辑区域 -->
      <div class="editor-section">
        <h2 class="section-title">
          <span class="title-icon">🖼️</span>
          Logo编辑
        </h2>
        <p class="section-description">系统Logo，正方形裁剪，建议尺寸：128×128像素</p>

        <div class="section-content">
          <div class="preview-area">
            <h3>当前Logo预览</h3>
            <div class="logo-preview-container">
              <div v-if="currentLogo" class="logo-preview">
                <img :src="currentLogo" alt="Logo预览" class="preview-image" />
              </div>
              <div v-else class="logo-placeholder">
                <div class="placeholder-icon">🖼️</div>
                <p>暂无Logo</p>
              </div>
            </div>
            <div class="preview-info" v-if="currentLogo">
              <p>尺寸: 128×128像素</p>
              <p>格式: PNG (支持透明背景)</p>
            </div>
          </div>

          <div class="editor-area">
            <h3>上传新Logo</h3>
            <div class="editor-instructions">
              <p>💡 使用说明：</p>
              <ul>
                <li>点击下方按钮上传Logo图片</li>
                <li>在编辑器中拖动图片调整位置</li>
                <li>使用缩放功能调整Logo大小</li>
                <li>框内亮色区域为保留部分</li>
                <li>确认裁剪后Logo将自动更新</li>
              </ul>
            </div>

            <NeuroImageCropper
              usage-type="logo"
              :preset-width="128"
              :preset-height="128"
              editor-title="Logo裁剪编辑器"
              @update:image="handleLogoUpdate"
              @upload-success="handleLogoSuccess"
              @upload-error="handleLogoError"
              @cancel="handleLogoCancel"
            />

            <div class="editor-actions" v-if="currentLogo">
              <button class="action-btn reset-btn" @click="resetLogo">
                <span class="btn-icon">↺</span>
                重置Logo
              </button>
              <button class="action-btn download-btn" @click="downloadLogo" v-if="currentLogo">
                <span class="btn-icon">⬇️</span>
                下载Logo
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- 头像编辑区域 -->
      <div class="editor-section">
        <h2 class="section-title">
          <span class="title-icon">👤</span>
          头像编辑
        </h2>
        <p class="section-description">用户头像，圆形裁剪，建议尺寸：150×150像素</p>

        <div class="section-content">
          <div class="preview-area">
            <h3>当前头像预览</h3>
            <div class="avatar-preview-container">
              <div v-if="currentAvatar" class="avatar-preview">
                <img :src="currentAvatar" alt="头像预览" class="preview-image avatar-image" />
              </div>
              <div v-else class="avatar-placeholder">
                <div class="placeholder-icon">👤</div>
                <p>暂无头像</p>
              </div>
            </div>
            <div class="preview-info" v-if="currentAvatar">
              <p>尺寸: 150×150像素</p>
              <p>格式: JPEG</p>
            </div>
          </div>

          <div class="editor-area">
            <h3>上传新头像</h3>
            <div class="editor-instructions">
              <p>💡 使用说明：</p>
              <ul>
                <li>点击下方按钮上传头像图片</li>
                <li>圆形裁剪框，确保人脸在中心</li>
                <li>使用缩放功能调整细节</li>
                <li>框内亮色区域为保留部分</li>
                <li>确认裁剪后头像将自动更新</li>
              </ul>
            </div>

            <NeuroImageCropper
              usage-type="avatar"
              :preset-width="150"
              :preset-height="150"
              editor-title="头像裁剪编辑器"
              @update:image="handleAvatarUpdate"
              @upload-success="handleAvatarSuccess"
              @upload-error="handleAvatarError"
              @cancel="handleAvatarCancel"
            />

            <div class="editor-actions" v-if="currentAvatar">
              <button class="action-btn reset-btn" @click="resetAvatar">
                <span class="btn-icon">↺</span>
                重置头像
              </button>
              <button class="action-btn download-btn" @click="downloadAvatar" v-if="currentAvatar">
                <span class="btn-icon">⬇️</span>
                下载头像
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- 悬停模式演示 -->
      <div class="editor-section">
        <h2 class="section-title">
          <span class="title-icon">🖱️</span>
          悬停模式演示
        </h2>
        <p class="section-description">鼠标悬停时显示上传按钮，适合集成到现有UI中</p>

        <div class="hover-demo-area">
          <div class="demo-item">
            <h3>Logo悬停上传</h3>
            <div class="hover-demo-container">
              <p class="demo-hint">⬇️ 悬停下面的区域测试 ⬇️</p>
              <NeuroImageCropper
                usage-type="logo"
                :preset-width="100"
                :preset-height="100"
                interaction-mode="hover"
                hover-text="+更换Logo"
                :parent-width="200"
                :parent-height="200"
                editor-title="Logo悬停上传"
                @update:image="handleHoverLogoUpdate"
                @upload-success="handleHoverLogoSuccess"
              />
            </div>
          </div>

          <div class="demo-item">
            <h3>头像悬停上传</h3>
            <div class="hover-demo-container">
              <p class="demo-hint">⬇️ 悬停下面的区域测试 ⬇️</p>
              <NeuroImageCropper
                usage-type="avatar"
                :preset-width="120"
                :preset-height="120"
                interaction-mode="hover"
                hover-text="+更换头像"
                :parent-width="200"
                :parent-height="200"
                editor-title="头像悬停上传"
                @update:image="handleHoverAvatarUpdate"
                @upload-success="handleHoverAvatarSuccess"
              />
            </div>
          </div>
        </div>
      </div>
    </div>

    <div class="editor-footer">
      <div class="status-summary">
        <h3>当前状态</h3>
        <div class="status-grid">
          <div class="status-item" :class="{ active: currentLogo }">
            <div class="status-icon">🖼️</div>
            <div class="status-content">
              <h4>Logo状态</h4>
              <p>{{ currentLogo ? '已设置' : '未设置' }}</p>
            </div>
          </div>
          <div class="status-item" :class="{ active: currentAvatar }">
            <div class="status-icon">👤</div>
            <div class="status-content">
              <h4>头像状态</h4>
              <p>{{ currentAvatar ? '已设置' : '未设置' }}</p>
            </div>
          </div>
        </div>
      </div>

      <div class="footer-actions">
        <button class="footer-btn save-all-btn" @click="saveAll" :disabled="!hasChanges">
          <span class="btn-icon">💾</span>
          保存所有更改
        </button>
        <button class="footer-btn reset-all-btn" @click="resetAll">
          <span class="btn-icon">🗑️</span>
          重置所有
        </button>
        <button class="footer-btn test-btn" @click="testIntegration">
          <span class="btn-icon">🔗</span>
          测试集成
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { silentDebug } from '@/utils/debug'
import { ref, computed } from 'vue'
import NeuroImageCropper from './NeuroImageCropper.vue'

// Logo状态
const currentLogo = ref('')
const newLogo = ref('')

// 头像状态
const currentAvatar = ref('')
const newAvatar = ref('')

// 悬停模式状态
const hoverLogo = ref('')
const hoverAvatar = ref('')

// 计算是否有更改
const hasChanges = computed(() => {
  return newLogo.value !== '' || newAvatar.value !== ''
})

// Logo处理函数
const handleLogoUpdate = (imageData: string) => {
  silentDebug('Logo更新:', imageData.substring(0, 50) + '...')
  newLogo.value = imageData
  // 立即更新预览
  currentLogo.value = imageData
}

const handleLogoSuccess = (_imageData: string) => {
  silentDebug('Logo上传成功')
  // 这里可以添加保存到服务器的逻辑
}

const handleLogoError = (error: string) => {
  console.error('Logo上传错误:', error)
  alert(`Logo上传错误: ${error}`)
}

const handleLogoCancel = () => {
  silentDebug('Logo上传取消')
}

// 头像处理函数
const handleAvatarUpdate = (imageData: string) => {
  silentDebug('头像更新:', imageData.substring(0, 50) + '...')
  newAvatar.value = imageData
  // 立即更新预览
  currentAvatar.value = imageData
}

const handleAvatarSuccess = (_imageData: string) => {
  silentDebug('头像上传成功')
  // 这里可以添加保存到服务器的逻辑
}

const handleAvatarError = (error: string) => {
  console.error('头像上传错误:', error)
  alert(`头像上传错误: ${error}`)
}

const handleAvatarCancel = () => {
  silentDebug('头像上传取消')
}

// 悬停模式处理函数
const handleHoverLogoUpdate = (imageData: string) => {
  silentDebug('悬停Logo更新')
  hoverLogo.value = imageData
}

const handleHoverLogoSuccess = () => {
  silentDebug('悬停Logo上传成功')
}

const handleHoverAvatarUpdate = (imageData: string) => {
  silentDebug('悬停头像更新')
  hoverAvatar.value = imageData
}

const handleHoverAvatarSuccess = () => {
  silentDebug('悬停头像上传成功')
}

// 重置功能
const resetLogo = () => {
  currentLogo.value = ''
  newLogo.value = ''
  silentDebug('Logo已重置')
}

const resetAvatar = () => {
  currentAvatar.value = ''
  newAvatar.value = ''
  silentDebug('头像已重置')
}

const resetAll = () => {
  resetLogo()
  resetAvatar()
  hoverLogo.value = ''
  hoverAvatar.value = ''
  silentDebug('所有图片已重置')
}

// 下载功能
const downloadLogo = () => {
  if (!currentLogo.value) return

  const link = document.createElement('a')
  link.href = currentLogo.value
  link.download = 'logo.png'
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
}

const downloadAvatar = () => {
  if (!currentAvatar.value) return

  const link = document.createElement('a')
  link.href = currentAvatar.value
  link.download = 'avatar.jpg'
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
}

// 保存所有更改
const saveAll = () => {
  if (!hasChanges.value) {
    alert('没有需要保存的更改')
    return
  }

  // 这里可以添加保存到服务器的逻辑
  const changes = []
  if (newLogo.value) changes.push('Logo')
  if (newAvatar.value) changes.push('头像')

  alert(`已保存更改: ${changes.join(', ')}`)
  silentDebug('保存所有更改:', { logo: newLogo.value ? '已更新' : '未更新', avatar: newAvatar.value ? '已更新' : '未更新' })

  // 重置新图片状态
  newLogo.value = ''
  newAvatar.value = ''
}

// 测试集成
const testIntegration = () => {
  silentDebug('测试集成功能')

  // 使用内置示例图片
  const sampleLogo = 'data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iMTI4IiBoZWlnaHQ9IjEyOCIgdmlld0JveD0iMCAwIDEyOCAxMjgiIGZpbGw9Im5vbmUiIHhtbG5zPSJodHRwOi8vd3d3LnczLm9yZy8yMDAwL3N2ZyI+CjxyZWN0IHdpZHRoPSIxMjgiIGhlaWdodD0iMTI4IiBmaWxsPSIjMDA5OUZGIi8+Cjx0ZXh0IHg9IjY0IiB5PSI3MCIgZm9udC1mYW1pbHk9IkFyaWFsIiBmb250LXNpemU9IjI0IiBmaWxsPSJ3aGl0ZSIgdGV4dC1hbmNob3I9Im1pZGRsZSI+TE9HTzwvdGV4dD4KPC9zdmc+'
  const sampleAvatar = 'data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iMTUwIiBoZWlnaHQ9IjE1MCIgdmlld0JveD0iMCAwIDE1MCAxNTAiIGZpbGw9Im5vbmUiIHhtbG5zPSJodHRwOi8vd3d3LnczLm9yZy8yMDAwL3N2ZyI+CjxyZWN0IHdpZHRoPSIxNTAiIGhlaWdodD0iMTUwIiByeD0iNzUiIGZpbGw9IiM2NjY2NjYiLz4KPHRleHQgeD0iNzUiIHk9IjgwIiBmb250LWZhbWlseT0iQXJpYWwiIGZvbnQtc2l6ZT0iMzIiIGZpbGw9IndoaXRlIiB0ZXh0LWFuY2hvcj0ibWlkZGxlIj5BPC90ZXh0Pgo8L3N2Zz4='

  currentLogo.value = sampleLogo
  currentAvatar.value = sampleAvatar

  alert('测试集成完成：已加载示例Logo和头像')
}
</script>

<style scoped>
.logo-avatar-editor {
  font-family: 'JetBrains Mono', monospace;
  max-width: 1200px;
  margin: 0 auto;
  padding: 24px;
  background: var(--nm-bg-base);
  color: var(--nm-text-primary);
}

.editor-header {
  text-align: center;
  margin-bottom: 40px;
  padding-bottom: 20px;
  border-bottom: 2px solid var(--nm-border);
}

.editor-header h1 {
  font-size: 32px;
  font-weight: 700;
  margin-bottom: 12px;
  color: var(--nm-primary);
}

.header-description {
  font-size: 16px;
  color: var(--nm-text-secondary);
  opacity: 0.8;
}

.editor-section {
  background: var(--nm-bg-surface);
  border-radius: 16px;
  padding: 28px;
  margin-bottom: 32px;
  border: 1px solid var(--nm-border);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
}

.section-title {
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 24px;
  margin-bottom: 12px;
  color: var(--nm-primary);
}

.title-icon {
  font-size: 28px;
}

.section-description {
  color: var(--nm-text-secondary);
  margin-bottom: 24px;
  font-size: 15px;
}

.section-content {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 32px;
}

@media (max-width: 768px) {
  .section-content {
    grid-template-columns: 1fr;
  }
}

.preview-area h3,
.editor-area h3 {
  font-size: 18px;
  margin-bottom: 16px;
  color: var(--nm-text-primary);
}

.logo-preview-container,
.avatar-preview-container {
  width: 200px;
  height: 200px;
  border: 2px solid var(--nm-border);
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 16px;
  background: var(--nm-bg-base);
  overflow: hidden;
}

.logo-preview,
.avatar-preview {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.preview-image {
  max-width: 100%;
  max-height: 100%;
  object-fit: contain;
}

.avatar-image {
  border-radius: 50%;
}

.logo-placeholder,
.avatar-placeholder {
  text-align: center;
  color: var(--nm-text-secondary);
}

.placeholder-icon {
  font-size: 48px;
  margin-bottom: 12px;
  opacity: 0.6;
}

.preview-info {
  font-size: 14px;
  color: var(--nm-text-secondary);
}

.preview-info p {
  margin: 4px 0;
}

.editor-instructions {
  background: rgba(0, 245, 212, 0.05);
  border: 1px solid var(--nm-primary);
  border-radius: 8px;
  padding: 16px;
  margin-bottom: 20px;
}

.editor-instructions p {
  font-weight: 600;
  margin-bottom: 8px;
  color: var(--nm-primary);
}

.editor-instructions ul {
  margin: 0;
  padding-left: 20px;
}

.editor-instructions li {
  margin-bottom: 6px;
  font-size: 14px;
  color: var(--nm-text-secondary);
}

.editor-actions {
  display: flex;
  gap: 12px;
  margin-top: 20px;
}

.action-btn {
  padding: 10px 20px;
  border: none;
  border-radius: 8px;
  font-family: 'JetBrains Mono', monospace;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 8px;
  transition: all 0.3s ease;
}

.reset-btn {
  background: var(--nm-bg-base);
  color: var(--nm-text-primary);
  border: 1px solid var(--nm-border);
}

.reset-btn:hover {
  background: var(--nm-bg-hover);
  border-color: var(--nm-primary);
}

.download-btn {
  background: var(--nm-primary);
  color: var(--nm-text-on-dark);
}

.download-btn:hover {
  background: var(--nm-primary-hover);
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 245, 212, 0.3);
}

/* 悬停模式演示 */
.hover-demo-area {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 32px;
  margin-top: 20px;
}

@media (max-width: 768px) {
  .hover-demo-area {
    grid-template-columns: 1fr;
  }
}

.demo-item h3 {
  font-size: 18px;
  margin-bottom: 16px;
  color: var(--nm-text-primary);
}

.hover-demo-container {
  border: 2px dashed var(--nm-border);
  border-radius: 12px;
  padding: 24px;
  background: var(--nm-bg-base);
}

.demo-hint {
  text-align: center;
  color: var(--nm-text-secondary);
  margin-bottom: 20px;
  font-size: 14px;
}

/* 页脚 */
.editor-footer {
  margin-top: 40px;
  padding-top: 28px;
  border-top: 2px solid var(--nm-border);
}

.status-summary {
  margin-bottom: 32px;
}

.status-summary h3 {
  font-size: 20px;
  margin-bottom: 20px;
  color: var(--nm-primary);
}

.status-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 20px;
}

.status-item {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 16px;
  background: var(--nm-bg-surface);
  border-radius: 12px;
  border: 1px solid var(--nm-border);
  transition: all 0.3s ease;
}

.status-item.active {
  border-color: var(--nm-primary);
  background: rgba(0, 245, 212, 0.05);
}

.status-icon {
  font-size: 32px;
}

.status-content h4 {
  font-size: 16px;
  margin-bottom: 4px;
  color: var(--nm-text-primary);
}

.status-content p {
  font-size: 14px;
  color: var(--nm-text-secondary);
}

.footer-actions {
  display: flex;
  gap: 16px;
  justify-content: center;
}

.footer-btn {
  padding: 14px 28px;
  border: none;
  border-radius: 10px;
  font-family: 'JetBrains Mono', monospace;
  font-size: 16px;
  font-weight: 600;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 10px;
  transition: all 0.3s ease;
}

.save-all-btn {
  background: var(--nm-primary);
  color: var(--nm-text-on-dark);
}

.save-all-btn:hover:not(:disabled) {
  background: var(--nm-primary-hover);
  transform: translateY(-2px);
  box-shadow: 0 6px 16px rgba(0, 245, 212, 0.4);
}

.save-all-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.reset-all-btn {
  background: var(--nm-bg-base);
  color: var(--nm-text-primary);
  border: 1px solid var(--nm-border);
}

.reset-all-btn:hover {
  background: var(--nm-bg-hover);
  border-color: var(--nm-error);
  color: var(--nm-error);
}

.test-btn {
  background: var(--nm-bg-base);
  color: var(--nm-text-primary);
  border: 1px solid var(--nm-border);
}

.test-btn:hover {
  background: var(--nm-bg-hover);
  border-color: var(--nm-warning);
  color: var(--nm-warning);
}
</style>