<template>
  <Teleport to="body">
    <div v-if="visible" class="system-settings-dialog" :class="{ 'light-mode': isLight }">
    <div class="dialog-overlay" @click="closeDialog">
      <div class="dialog-container" @click.stop>
        <!-- 对话框头部 -->
        <div class="dialog-header">
          <div class="header-content">
            <h2 class="dialog-title">
              <span class="title-icon">⚙️</span>
              系统设置
            </h2>
            <p class="dialog-subtitle">配置系统Logo{{ isPlatformAdmin ? '、名称和版权信息' : '与名称' }}</p>
          </div>
          <button class="close-btn" @click="closeDialog">
            <span class="close-icon">×</span>
          </button>
        </div>

        <!-- 对话框内容 -->
        <div class="dialog-content">
          <!-- Logo设置 -->
          <div class="settings-section">
            <div class="section-header">
              <h3 class="section-title">
                <span class="section-icon">🖼️</span>
                系统Logo设置
              </h3>
              <p class="section-description">上传并裁剪系统Logo，建议尺寸：128×128像素</p>
            </div>

            <div class="section-content">
              <div class="logo-preview-area">
                <div class="current-logo">
                  <h4>当前Logo</h4>
                  <div class="logo-container">
                    <img v-if="currentLogo" :src="currentLogo" alt="系统Logo" class="logo-image" />
                    <div v-else class="logo-placeholder">
                      <span class="placeholder-icon">🖼️</span>
                      <p>暂无Logo</p>
                    </div>
                  </div>
                  <div class="logo-info" v-if="currentLogo">
                    <p>尺寸: 128×128像素</p>
                    <p>格式: PNG</p>
                  </div>
                </div>

                <div class="logo-editor">
                  <h4>上传新Logo</h4>
                  <div class="editor-instructions">
                    <p class="instruction">
                      <span class="instruction-icon">💡</span>
                      使用说明：正方形裁剪框，框内亮色区域为保留部分
                    </p>
                  </div>

                  <NeuroImageCropper
                    usage-type="logo"
                    :preset-width="128"
                    :preset-height="128"
                    editor-title="系统Logo裁剪"
                    @update:image="handleLogoUpdate"
                    @upload-success="handleLogoSuccess"
                    @upload-error="handleLogoError"
                    @cancel="handleLogoCancel"
                  />

                  <div class="logo-actions" v-if="currentLogo">
                    <button class="action-btn reset-btn" @click="resetLogo">
                      <span class="btn-icon">↺</span>
                      重置Logo
                    </button>
                    <button class="action-btn download-btn" @click="downloadLogo">
                      <span class="btn-icon">⬇️</span>
                      下载Logo
                    </button>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- 系统名称设置 -->
          <div class="settings-section">
            <div class="section-header">
              <h3 class="section-title">
                <span class="section-icon">🏷️</span>
                系统名称设置
              </h3>
              <p class="section-description">设置系统显示名称，将显示在登录页和系统顶部</p>
            </div>

            <div class="section-content">
              <div class="system-name-editor">
                <div class="form-group">
                  <label class="form-label">
                    <span class="label-icon">📝</span>
                    系统名称
                  </label>
                  <div class="input-container">
                    <input
                      v-model="systemName"
                      type="text"
                      class="neuro-input"
                      placeholder="请输入系统名称"
                      maxlength="50"
                    />
                    <div class="input-glow"></div>
                  </div>
                  <p class="input-hint">最多50个字符</p>
                </div>

                <div class="name-preview">
                  <h4>预览效果</h4>
                  <div class="preview-container">
                    <div class="preview-header">
                      <div class="preview-logo" v-if="currentLogo">
                        <img :src="currentLogo" alt="Logo预览" />
                      </div>
                      <div class="preview-logo placeholder" v-else>
                        <span>🖼️</span>
                      </div>
                      <div class="preview-name">{{ systemName || '系统名称' }}</div>
                    </div>
                    <div class="preview-footer" v-if="isPlatformAdmin">
                      <p class="preview-copyright">{{ copyrightInfo || '© 2026 版权所有' }}</p>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- 版权信息设置（仅平台管理员可见） -->
          <div v-if="isPlatformAdmin" class="settings-section">
            <div class="section-header">
              <h3 class="section-title">
                <span class="section-icon">©️</span>
                版权信息设置
              </h3>
              <p class="section-description">设置登录页底部的版权信息</p>
            </div>

            <div class="section-content">
              <div class="copyright-editor">
                <div class="form-group">
                  <label class="form-label">
                    <span class="label-icon">📄</span>
                    版权信息
                  </label>
                  <div class="input-container">
                    <textarea
                      v-model="copyrightInfo"
                      class="neuro-textarea"
                      placeholder="请输入版权信息"
                      rows="3"
                      maxlength="200"
                    ></textarea>
                    <div class="input-glow"></div>
                  </div>
                  <p class="input-hint">最多200个字符</p>
                </div>

                <div class="copyright-preview">
                  <h4>登录页预览</h4>
                  <div class="login-preview">
                    <div class="preview-login-header">
                      <div class="preview-login-logo" v-if="currentLogo">
                        <img :src="currentLogo" alt="Logo" />
                      </div>
                      <div class="preview-login-logo placeholder" v-else>
                        <span>🖼️</span>
                      </div>
                      <div class="preview-login-name">{{ systemName || '系统名称' }}</div>
                    </div>
                    <div class="preview-login-footer">
                      <p class="preview-login-copyright">{{ copyrightInfo || '© 2026 版权所有' }}</p>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- 对话框底部 -->
        <div class="dialog-footer">
          <div class="footer-actions">
            <button class="footer-btn cancel-btn" @click="closeDialog">
              <span class="btn-icon">←</span>
              取消
            </button>
            <button class="footer-btn reset-btn" @click="resetAll">
              <span class="btn-icon">🗑️</span>
              重置所有
            </button>
            <button class="footer-btn save-btn" @click="saveSettings" :disabled="!hasChanges">
              <span class="btn-icon">💾</span>
              保存设置
            </button>
          </div>

          <div class="footer-info">
            <p class="info-text">
              <span class="info-icon">💡</span>
              提示：设置将立即生效，建议在非工作时间进行修改
            </p>
          </div>
        </div>
      </div>
    </div>
  </div>
  </Teleport>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { ElMessage } from 'element-plus'
import NeuroImageCropper from './NeuroImageCropper.vue'

interface Props {
  visible: boolean
  currentLogo?: string
  currentSystemName?: string
  currentCopyright?: string
  isPlatformAdmin?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  visible: false,
  currentLogo: '',
  currentSystemName: 'PMTools',
  currentCopyright: '© 2026 PMTools - AI协作开发系统',
  isPlatformAdmin: false
})

// 检测浅色模式（从 localStorage 读取，与登录页一致）
const isLight = computed(() => localStorage.getItem('login-theme-mode') === 'light')

const emit = defineEmits<{
  'update:visible': [visible: boolean]
  'save': [settings: {
    logo: string
    systemName: string
    copyright: string
  }]
  'logo-update': [logoData: string]
}>()

// 本地编辑状态
const localVisible = ref(props.visible)
const currentLogo = ref(props.currentLogo)
const systemName = ref(props.currentSystemName)
const copyrightInfo = ref(props.currentCopyright)

// 对话框打开时的初始值快照（用于判断是否有修改）
const snapshotLogo = ref(props.currentLogo)
const snapshotName = ref(props.currentSystemName)
const snapshotCopyright = ref(props.currentCopyright)

// 计算属性：对比当前编辑值与打开对话框时的快照
const hasChanges = computed(() => {
  const logoChanged = currentLogo.value !== snapshotLogo.value
  const nameChanged = systemName.value !== snapshotName.value
  const copyrightChanged = props.isPlatformAdmin && copyrightInfo.value !== snapshotCopyright.value
  return logoChanged || nameChanged || copyrightChanged
})

// 对话框打开时记录快照
watch(() => props.visible, (newVal) => {
  localVisible.value = newVal
  if (newVal) {
    snapshotLogo.value = props.currentLogo
    snapshotName.value = props.currentSystemName
    snapshotCopyright.value = props.currentCopyright
    currentLogo.value = props.currentLogo
    systemName.value = props.currentSystemName
    copyrightInfo.value = props.currentCopyright
  }
})

// 对话框关闭期间同步外部 props 变化
watch(() => props.currentLogo, (newVal) => {
  if (!localVisible.value) {
    currentLogo.value = newVal
  }
})

watch(() => props.currentSystemName, (newVal) => {
  if (!localVisible.value) {
    systemName.value = newVal
  }
})

watch(() => props.currentCopyright, (newVal) => {
  if (!localVisible.value) {
    copyrightInfo.value = newVal
  }
})

// Logo 裁剪回调
const handleLogoUpdate = (imageData: string) => {
  currentLogo.value = imageData
  emit('logo-update', imageData)
}

const handleLogoSuccess = () => {}
const handleLogoError = (error: string) => {
  alert(`Logo上传错误: ${error}`)
}
const handleLogoCancel = () => {}

// 重置功能
const resetLogo = () => {
  currentLogo.value = ''
}

const resetAll = () => {
  currentLogo.value = ''
  systemName.value = 'PMTools'
  if (props.isPlatformAdmin) {
    copyrightInfo.value = '© 2026 PMTools - AI协作开发系统'
  }
}

// 下载功能
const downloadLogo = () => {
  if (!currentLogo.value) return
  const link = document.createElement('a')
  link.href = currentLogo.value
  link.download = 'system-logo.png'
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
}

// 保存设置
const saveSettings = () => {
  if (!hasChanges.value) return

  emit('save', {
    logo: currentLogo.value,
    systemName: systemName.value,
    copyright: props.isPlatformAdmin ? copyrightInfo.value : props.currentCopyright
  })

  // 更新快照
  snapshotLogo.value = currentLogo.value
  snapshotName.value = systemName.value
  snapshotCopyright.value = copyrightInfo.value

  ElMessage.success('设置已保存')

  // 延迟关闭对话框
  setTimeout(() => {
    localVisible.value = false
    emit('update:visible', false)
  }, 800)
}

// 关闭对话框
const closeDialog = () => {
  localVisible.value = false
  emit('update:visible', false)
}
</script>

<style scoped>
.system-settings-dialog {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  z-index: 9999;
}

.dialog-overlay {
  width: 100%;
  height: 100%;
  background: rgba(0, 0, 0, 0.8);
  display: flex;
  align-items: center;
  justify-content: center;
  backdrop-filter: blur(10px);
}

.dialog-container {
  background: linear-gradient(135deg, #0a0a0f 0%, #1a1a2e 100%);
  border-radius: 12px;
  border: 1px solid rgba(0, 245, 212, 0.3);
  width: 90%;
  max-width: 860px;
  max-height: 85vh;
  overflow-y: auto;
  box-shadow: 0 25px 50px rgba(0, 0, 0, 0.5);
  font-family: 'JetBrains Mono', monospace;
}

/* 对话框头部 */
.dialog-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  border-bottom: 1px solid rgba(0, 245, 212, 0.2);
  background: rgba(16, 16, 32, 0.8);
  border-radius: 12px 12px 0 0;
}

.header-content .dialog-title {
  font-size: 18px;
  font-weight: 700;
  margin-bottom: 0;
  background: linear-gradient(90deg, #00f5d4, #9d4edd);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  display: flex;
  align-items: center;
  gap: 8px;
}

.title-icon {
  font-size: 20px;
}

.dialog-subtitle {
  font-size: 12px;
  color: #8a8aff;
  opacity: 0.8;
  margin: 2px 0 0;
}

.close-btn {
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 50%;
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: all 0.3s ease;
}

.close-btn:hover {
  background: rgba(255, 87, 87, 0.1);
  border-color: #ff5757;
  transform: rotate(90deg);
}

.close-icon {
  font-size: 20px;
  color: #e0e0ff;
}

.close-btn:hover .close-icon {
  color: #ff5757;
}

/* 对话框内容 */
.dialog-content {
  padding: 16px 20px;
}

/* 设置区域 */
.settings-section {
  margin-bottom: 16px;
  background: rgba(16, 16, 32, 0.5);
  border-radius: 8px;
  border: 1px solid rgba(0, 245, 212, 0.1);
  overflow: hidden;
}

.section-header {
  padding: 12px 16px;
  border-bottom: 1px solid rgba(0, 245, 212, 0.1);
  background: rgba(0, 245, 212, 0.05);
}

.section-title {
  font-size: 14px;
  font-weight: 600;
  margin-bottom: 0;
  color: #00f5d4;
  display: flex;
  align-items: center;
  gap: 8px;
}

.section-icon {
  font-size: 16px;
}

.section-description {
  font-size: 12px;
  color: #8a8aff;
  opacity: 0.8;
  margin: 2px 0 0;
}

.section-content {
  padding: 16px;
}

/* Logo设置 */
.logo-preview-area {
  display: grid;
  grid-template-columns: 1fr 2fr;
  gap: 20px;
}

@media (max-width: 768px) {
  .logo-preview-area {
    grid-template-columns: 1fr;
  }
}

.current-logo h4,
.logo-editor h4 {
  font-size: 14px;
  margin-bottom: 10px;
  color: #e0e0ff;
}

.logo-container {
  width: 100px;
  height: 100px;
  border: 2px solid rgba(0, 245, 212, 0.3);
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 10px;
  background: rgba(255, 255, 255, 0.05);
  overflow: hidden;
}

.logo-image {
  max-width: 100%;
  max-height: 100%;
  object-fit: contain;
}

.logo-placeholder {
  text-align: center;
  color: #8a8aff;
}

.placeholder-icon {
  font-size: 32px;
  margin-bottom: 6px;
  opacity: 0.6;
}

.logo-info {
  font-size: 12px;
  color: #8a8aff;
}

.logo-info p {
  margin: 2px 0;
}

.editor-instructions {
  background: rgba(0, 245, 212, 0.05);
  border: 1px solid rgba(0, 245, 212, 0.1);
  border-radius: 6px;
  padding: 8px;
  margin-bottom: 10px;
}

.instruction {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: #8a8aff;
}

.instruction-icon {
  font-size: 12px;
}

.logo-actions {
  display: flex;
  gap: 8px;
  margin-top: 12px;
}

.action-btn {
  padding: 6px 14px;
  border: none;
  border-radius: 6px;
  font-family: 'JetBrains Mono', monospace;
  font-size: 12px;
  font-weight: 600;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 6px;
  transition: all 0.3s ease;
}

.reset-btn {
  background: rgba(255, 255, 255, 0.05);
  color: #e0e0ff;
  border: 1px solid rgba(255, 255, 255, 0.1);
}

.reset-btn:hover {
  background: rgba(255, 255, 255, 0.1);
  border-color: rgba(255, 255, 255, 0.3);
}

.download-btn {
  background: linear-gradient(135deg, #00f5d4, #00b4d8);
  color: #0a0a0f;
  border: none;
}

.download-btn:hover {
  background: linear-gradient(135deg, #00e6c7, #0099cc);
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 245, 212, 0.3);
}

/* 系统名称设置 */
.system-name-editor {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 20px;
}

@media (max-width: 768px) {
  .system-name-editor {
    grid-template-columns: 1fr;
  }
}

.form-group {
  margin-bottom: 14px;
}

.form-label {
  display: block;
  font-size: 13px;
  font-weight: 600;
  margin-bottom: 6px;
  color: #e0e0ff;
  display: flex;
  align-items: center;
  gap: 6px;
}

.label-icon {
  font-size: 14px;
}

.input-container {
  position: relative;
}

.neuro-input,
.neuro-textarea {
  width: 100%;
  max-width: 100%;
  box-sizing: border-box;
  padding: 8px 12px;
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(0, 245, 212, 0.2);
  border-radius: 6px;
  color: #e0e0ff;
  font-family: 'JetBrains Mono', monospace;
  font-size: 13px;
  transition: all 0.3s ease;
}

.neuro-textarea {
  resize: vertical;
  min-height: 50px;
}

.neuro-input:focus,
.neuro-textarea:focus {
  outline: none;
  border-color: #00f5d4;
  background: rgba(0, 245, 212, 0.05);
  box-shadow: 0 0 0 2px rgba(0, 245, 212, 0.1);
}

.neuro-input::placeholder,
.neuro-textarea::placeholder {
  color: #8a8aff;
  opacity: 0.5;
}

.input-glow {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  border-radius: 6px;
  box-shadow: inset 0 0 10px rgba(0, 245, 212, 0.1);
  pointer-events: none;
  opacity: 0;
  transition: opacity 0.3s ease;
}

.neuro-input:focus + .input-glow,
.neuro-textarea:focus + .input-glow {
  opacity: 1;
}

.input-hint {
  font-size: 11px;
  color: #8a8aff;
  margin-top: 2px;
  opacity: 0.7;
}

/* 预览效果 */
.name-preview h4,
.copyright-preview h4 {
  font-size: 14px;
  margin-bottom: 10px;
  color: #e0e0ff;
}

.preview-container,
.login-preview {
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(0, 245, 212, 0.2);
  border-radius: 8px;
  padding: 14px;
}

.preview-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 12px;
  padding-bottom: 12px;
  border-bottom: 1px solid rgba(0, 245, 212, 0.1);
}

.preview-logo {
  width: 36px;
  height: 36px;
  border-radius: 6px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(255, 255, 255, 0.1);
  overflow: hidden;
}

.preview-logo img {
  max-width: 100%;
  max-height: 100%;
  object-fit: contain;
}

.preview-logo.placeholder {
  font-size: 18px;
}

.preview-name {
  font-size: 16px;
  font-weight: 600;
  color: #00f5d4;
}

.preview-footer {
  text-align: center;
}

.preview-copyright {
  font-size: 12px;
  color: #8a8aff;
  opacity: 0.8;
}

/* 登录页预览 */
.login-preview {
  background: linear-gradient(135deg, #1a1a2e 0%, #16213e 100%);
}

.preview-login-header {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 10px;
  margin-bottom: 16px;
}

.preview-login-logo {
  width: 56px;
  height: 56px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(255, 255, 255, 0.1);
  overflow: hidden;
}

.preview-login-logo img {
  max-width: 100%;
  max-height: 100%;
  object-fit: contain;
}

.preview-login-logo.placeholder {
  font-size: 28px;
}

.preview-login-name {
  font-size: 18px;
  font-weight: 700;
  color: #00f5d4;
  text-align: center;
}

.preview-login-footer {
  text-align: center;
  padding-top: 10px;
  border-top: 1px solid rgba(0, 245, 212, 0.1);
}

.preview-login-copyright {
  font-size: 12px;
  color: #8a8aff;
  opacity: 0.8;
}

/* 对话框底部 */
.dialog-footer {
  padding: 14px 20px;
  border-top: 1px solid rgba(0, 245, 212, 0.2);
  background: rgba(16, 16, 32, 0.8);
  border-radius: 0 0 12px 12px;
}

.footer-actions {
  display: flex;
  gap: 10px;
  justify-content: center;
  margin-bottom: 10px;
}

.footer-btn {
  padding: 8px 18px;
  border: none;
  border-radius: 8px;
  font-family: 'JetBrains Mono', monospace;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 6px;
  transition: all 0.3s ease;
  position: relative;
  overflow: hidden;
}

.footer-btn::before {
  content: '';
  position: absolute;
  top: 0;
  left: -100%;
  width: 100%;
  height: 100%;
  background: linear-gradient(90deg, transparent, rgba(255, 255, 255, 0.1), transparent);
  transition: left 0.5s ease;
}

.footer-btn:hover::before {
  left: 100%;
}

.cancel-btn {
  background: rgba(255, 255, 255, 0.05);
  color: #e0e0ff;
  border: 1px solid rgba(255, 255, 255, 0.1);
}

.cancel-btn:hover {
  background: rgba(255, 255, 255, 0.1);
  border-color: rgba(255, 255, 255, 0.3);
  transform: translateX(-4px);
}

.footer-btn.reset-btn {
  background: rgba(255, 87, 87, 0.1);
  color: #ff5757;
  border: 1px solid rgba(255, 87, 87, 0.3);
}

.footer-btn.reset-btn:hover {
  background: rgba(255, 87, 87, 0.2);
  border-color: #ff5757;
  transform: translateY(-2px);
}

.save-btn {
  background: linear-gradient(135deg, #00f5d4, #9d4edd);
  color: #0a0a0f;
  border: none;
}

.save-btn:hover:not(:disabled) {
  background: linear-gradient(135deg, #00e6c7, #8a2be2);
  transform: translateY(-2px);
  box-shadow: 0 6px 16px rgba(0, 245, 212, 0.4);
}

.save-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
  transform: none !important;
}

.footer-info {
  text-align: center;
}

.info-text {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  font-size: 12px;
  color: #8a8aff;
  opacity: 0.8;
}

.info-icon {
  font-size: 12px;
}

/* ===== 浅色模式 ===== */
.system-settings-dialog.light-mode .dialog-overlay {
  background: rgba(0, 0, 0, 0.4);
}

.system-settings-dialog.light-mode .dialog-container {
  background: linear-gradient(135deg, #f8fafc 0%, #ffffff 100%);
  border-color: rgba(14, 165, 233, 0.3);
  color: #0f172a;
}

.system-settings-dialog.light-mode .dialog-header {
  border-bottom-color: rgba(14, 165, 233, 0.2);
  background: rgba(255, 255, 255, 0.95);
}

.system-settings-dialog.light-mode .dialog-subtitle {
  color: #475569;
}

.system-settings-dialog.light-mode .close-btn {
  background: #f1f5f9;
  border-color: #cbd5e1;
}

.system-settings-dialog.light-mode .close-btn:hover {
  background: rgba(239, 68, 68, 0.1);
  border-color: #ef4444;
}

.system-settings-dialog.light-mode .close-icon {
  color: #0f172a;
}

.system-settings-dialog.light-mode .close-btn:hover .close-icon {
  color: #ef4444;
}

.system-settings-dialog.light-mode .settings-section {
  background: #ffffff;
  border-color: #cbd5e1;
}

.system-settings-dialog.light-mode .section-header {
  border-bottom-color: #e2e8f0;
  background: rgba(14, 165, 233, 0.04);
}

.system-settings-dialog.light-mode .section-title {
  color: #0ea5e9;
}

.system-settings-dialog.light-mode .section-description {
  color: #475569;
}

.system-settings-dialog.light-mode .current-logo h4,
.system-settings-dialog.light-mode .logo-editor h4 {
  color: #0f172a;
}

.system-settings-dialog.light-mode .logo-container {
  border-color: rgba(14, 165, 233, 0.3);
  background: #f8fafc;
}

.system-settings-dialog.light-mode .logo-placeholder {
  color: #64748b;
}

.system-settings-dialog.light-mode .logo-info {
  color: #475569;
}

.system-settings-dialog.light-mode .editor-instructions {
  background: rgba(14, 165, 233, 0.04);
  border-color: rgba(14, 165, 233, 0.12);
}

.system-settings-dialog.light-mode .instruction {
  color: #475569;
}

.system-settings-dialog.light-mode .reset-btn {
  background: #f1f5f9;
  color: #0f172a;
  border-color: #cbd5e1;
}

.system-settings-dialog.light-mode .reset-btn:hover {
  background: #e2e8f0;
  border-color: #94a3b8;
}

.system-settings-dialog.light-mode .download-btn {
  background: linear-gradient(135deg, #0ea5e9, #0284c7);
  color: #ffffff;
}

.system-settings-dialog.light-mode .download-btn:hover {
  background: linear-gradient(135deg, #0284c7, #0369a1);
  box-shadow: 0 4px 12px rgba(14, 165, 233, 0.3);
}

.system-settings-dialog.light-mode .form-label {
  color: #0f172a;
}

.system-settings-dialog.light-mode .neuro-input,
.system-settings-dialog.light-mode .neuro-textarea {
  background: #f8fafc;
  border-color: #cbd5e1;
  color: #0f172a;
}

.system-settings-dialog.light-mode .neuro-input:focus,
.system-settings-dialog.light-mode .neuro-textarea:focus {
  border-color: #0ea5e9;
  background: #f0f9ff;
  box-shadow: 0 0 0 2px rgba(14, 165, 233, 0.1);
}

.system-settings-dialog.light-mode .neuro-input::placeholder,
.system-settings-dialog.light-mode .neuro-textarea::placeholder {
  color: #94a3b8;
}

.system-settings-dialog.light-mode .input-hint {
  color: #64748b;
}

.system-settings-dialog.light-mode .name-preview h4,
.system-settings-dialog.light-mode .copyright-preview h4 {
  color: #0f172a;
}

.system-settings-dialog.light-mode .preview-container,
.system-settings-dialog.light-mode .login-preview {
  background: #f8fafc;
  border-color: #cbd5e1;
}

.system-settings-dialog.light-mode .preview-header {
  border-bottom-color: #e2e8f0;
}

.system-settings-dialog.light-mode .preview-name {
  color: #0ea5e9;
}

.system-settings-dialog.light-mode .preview-copyright {
  color: #475569;
}

.system-settings-dialog.light-mode .login-preview {
  background: linear-gradient(135deg, #f8fafc 0%, #f1f5f9 100%);
}

.system-settings-dialog.light-mode .preview-login-name {
  color: #0ea5e9;
}

.system-settings-dialog.light-mode .preview-login-copyright {
  color: #475569;
}

.system-settings-dialog.light-mode .preview-login-footer {
  border-top-color: #e2e8f0;
}

.system-settings-dialog.light-mode .preview-logo {
  background: #e2e8f0;
}

.system-settings-dialog.light-mode .preview-login-logo {
  background: #e2e8f0;
}

.system-settings-dialog.light-mode .dialog-footer {
  border-top-color: rgba(14, 165, 233, 0.2);
  background: rgba(255, 255, 255, 0.95);
}

.system-settings-dialog.light-mode .cancel-btn {
  background: #f1f5f9;
  color: #0f172a;
  border-color: #cbd5e1;
}

.system-settings-dialog.light-mode .cancel-btn:hover {
  background: #e2e8f0;
  border-color: #94a3b8;
}

.system-settings-dialog.light-mode .footer-btn.reset-btn {
  background: rgba(239, 68, 68, 0.08);
  color: #dc2626;
  border-color: rgba(239, 68, 68, 0.3);
}

.system-settings-dialog.light-mode .footer-btn.reset-btn:hover {
  background: rgba(239, 68, 68, 0.15);
  border-color: #ef4444;
}

.system-settings-dialog.light-mode .save-btn {
  background: linear-gradient(135deg, #0ea5e9, #8b5cf6);
  color: #ffffff;
}

.system-settings-dialog.light-mode .save-btn:hover:not(:disabled) {
  background: linear-gradient(135deg, #0284c7, #7c3aed);
  box-shadow: 0 6px 16px rgba(14, 165, 233, 0.3);
}

.system-settings-dialog.light-mode .info-text {
  color: #475569;
}

.system-settings-dialog.light-mode .save-success-toast {
  background: rgba(14, 165, 233, 0.12);
  border-color: rgba(14, 165, 233, 0.5);
  box-shadow: 0 8px 32px rgba(14, 165, 233, 0.15);
}

.system-settings-dialog.light-mode .toast-icon {
  background: #0ea5e9;
  color: #ffffff;
}

.system-settings-dialog.light-mode .toast-text {
  color: #0ea5e9;
}

.system-settings-dialog.light-mode .footer-btn::before {
  background: linear-gradient(90deg, transparent, rgba(0, 0, 0, 0.04), transparent);
}
</style>
