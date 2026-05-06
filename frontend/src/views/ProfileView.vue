<template>
  <div class="neuro-profile-view" :class="{ 'light-mode': isLight }">
    <!-- 页面头部 -->
    <div ref="headerRef" class="neuro-header">
      <div class="header-content">
        <h1 class="neuro-title">
          <span class="title-icon">🧠</span>
          个人资料中心
        </h1>
      </div>
      <div class="header-actions">
        <button class="neuro-btn back-btn" @click="goBack" title="返回控制台">
          <span class="btn-icon">←</span>
        </button>
      </div>
    </div>

    <div class="neuro-content">
      <!-- 左侧：个人信息卡片 -->
      <div class="neuro-card profile-card">
        <div class="card-header">
          <h2 class="card-title">
            <span class="title-icon">👤</span>
            身份信息
          </h2>
          <div class="card-subtitle">神经矩阵中的数字身份</div>
        </div>

        <div class="card-content">
          <!-- 头像区域 -->
          <div class="avatar-section">
            <div class="avatar-container">
              <div class="avatar-preview" :class="{ 'has-avatar': avatarSrc }">
                <img v-if="avatarSrc" :src="avatarSrc" alt="头像" class="avatar-image" />
                <div v-else class="avatar-placeholder">
                  <span class="placeholder-text">{{ avatarFallbackInitial }}</span>
                </div>
                <div class="avatar-glow"></div>
              </div>

              <div class="avatar-controls">
                <h3 class="controls-title">AI 头像配置</h3>
                <p class="controls-description">上传并裁剪您的数字身份头像</p>

                <div class="controls-buttons">
                  <!-- 使用NeuroImageCropper组件处理头像上传 -->
                  <NeuroImageCropper
                    usage-type="avatar"
                    :preset-width="150"
                    :preset-height="150"
                    editor-title="头像裁剪"
                    interaction-mode="standalone"
                    @update:image="handleAvatarCropped"
                    @upload-success="handleAvatarCroppedSuccess"
                    @upload-error="handleAvatarCroppedError"
                    @cancel="closeAvatarCropper"
                  >
                    <template #trigger>
                      <button class="neuro-btn primary-btn">
                        <span class="btn-icon">📷</span>
                        <span class="btn-text">上传新头像</span>
                      </button>
                    </template>
                  </NeuroImageCropper>

                  <button
                    class="neuro-btn secondary-btn"
                    @click="clearAvatar"
                    :disabled="!avatarSrc"
                  >
                    <span class="btn-icon">🗑️</span>
                    <span class="btn-text">清除头像</span>
                  </button>
                </div>

                <div class="avatar-info">
                  <p class="info-item">
                    <span class="info-icon">💡</span>
                    建议尺寸：150×150像素
                  </p>
                  <p class="info-item">
                    <span class="info-icon">⚙️</span>
                    支持格式：JPG、PNG、WebP
                  </p>
                  <p class="info-item">
                    <span class="info-icon">📏</span>
                    最大大小：{{ MAX_IMAGE_UPLOAD_LABEL }}
                  </p>
                </div>

                <div v-if="avatarUploadErr" class="error-message">
                  <span class="error-icon">❌</span>
                  {{ avatarUploadErr }}
                </div>
              </div>
            </div>
          </div>

          <!-- 基本信息表单 -->
          <div class="form-section">
            <h3 class="form-title">
              <span class="title-icon">📝</span>
              基本信息
            </h3>

            <div class="neuro-form">
              <div class="form-group">
                <label class="form-label">
                  <span class="label-icon">👤</span>
                  姓名
                </label>
                <div class="form-input-container">
                  <input
                    v-model="form.name"
                    type="text"
                    class="neuro-input"
                    placeholder="请输入您的姓名"
                  />
                  <div class="input-glow"></div>
                </div>
              </div>

              <div class="form-group">
                <label class="form-label">
                  <span class="label-icon">📱</span>
                  手机号码
                </label>
                <div class="form-input-container">
                  <input
                    v-model="form.phone"
                    type="tel"
                    inputmode="tel"
                    class="neuro-input"
                    placeholder="请输入手机号码"
                    @input="form.phone = sanitizePhoneInput(($event.target as HTMLInputElement).value)"
                  />
                  <div class="input-glow"></div>
                </div>
              </div>

              <div class="form-group">
                <label class="form-label">
                  <span class="label-icon">📧</span>
                  电子邮箱
                </label>
                <div class="form-input-container">
                  <input
                    v-model="form.email"
                    type="email"
                    class="neuro-input"
                    placeholder="请输入电子邮箱"
                  />
                  <div class="input-glow"></div>
                </div>
              </div>

              <div class="form-actions">
                <button class="neuro-btn save-btn" @click="saveProfile" :disabled="isSaving">
                  <span class="btn-icon" v-if="!isSaving">💾</span>
                  <span class="btn-icon loading" v-else>⏳</span>
                  <span class="btn-text">{{ isSaving ? '保存中...' : '保存资料' }}</span>
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- 右侧：安全设置卡片 -->
      <div class="neuro-card security-card">
        <div class="card-header">
          <h2 class="card-title">
            <span class="title-icon">🔒</span>
            安全设置
          </h2>
          <div class="card-subtitle">神经矩阵访问凭证管理</div>
        </div>

        <div class="card-content">
          <div class="security-info">
            <p class="info-text">
              <span class="info-icon">⚠️</span>
              修改密码需要验证当前密码与图形验证码
            </p>
            <p class="info-text">
              <span class="info-icon">🔑</span>
              新密码至少 8 位且同时包含字母与数字
            </p>
            <p class="info-text">
              <span class="info-icon">⏰</span>
              连续输错当前密码多次将暂时锁定 15 分钟
            </p>
          </div>

          <div class="neuro-form">
            <div class="form-group">
              <label class="form-label">
                <span class="label-icon">🔐</span>
                当前密码
              </label>
              <div class="form-input-container">
                <input
                  v-model="pwd.old"
                  type="password"
                  class="neuro-input"
                  placeholder="请输入当前密码"
                  autocomplete="current-password"
                />
                <div class="input-glow"></div>
              </div>
            </div>

            <div class="form-group">
              <label class="form-label">
                <span class="label-icon">🆕</span>
                新密码
              </label>
              <div class="form-input-container">
                <input
                  v-model="pwd.next"
                  type="password"
                  class="neuro-input"
                  placeholder="请输入新密码"
                  autocomplete="new-password"
                />
                <div class="input-glow"></div>
              </div>
            </div>

            <div class="form-group">
              <label class="form-label">
                <span class="label-icon">✅</span>
                确认新密码
              </label>
              <div class="form-input-container">
                <input
                  v-model="pwd.confirm"
                  type="password"
                  class="neuro-input"
                  placeholder="请再次输入新密码"
                  autocomplete="new-password"
                />
                <div class="input-glow"></div>
              </div>
            </div>

            <div class="form-group">
              <label class="form-label">
                <span class="label-icon">🖼️</span>
                图形验证码
              </label>
              <div class="captcha-container">
                <div class="form-input-container">
                  <input
                    v-model="pwd.captchaCode"
                    type="text"
                    class="neuro-input"
                    placeholder="请输入验证码"
                    maxlength="8"
                  />
                  <div class="input-glow"></div>
                </div>

                <div class="captcha-image-container">
                  <img
                    v-if="pwd.captchaImg"
                    :src="pwd.captchaImg"
                    alt="验证码"
                    class="captcha-image"
                    @click="refreshPwdCaptcha"
                    title="点击换一张"
                  />
                  <button
                    v-else
                    class="neuro-btn captcha-btn"
                    @click="refreshPwdCaptcha"
                  >
                    <span class="btn-icon">🔄</span>
                    <span class="btn-text">获取验证码</span>
                  </button>
                </div>
              </div>
            </div>

            <div class="form-actions">
              <button class="neuro-btn security-btn" @click="savePwd" :disabled="isChangingPassword">
                <span class="btn-icon" v-if="!isChangingPassword">🔐</span>
                <span class="btn-icon loading" v-else>⏳</span>
                <span class="btn-text">{{ isChangingPassword ? '修改中...' : '修改密码' }}</span>
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>


  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import NeuroImageCropper from '@/components/NeuroImageCropper.vue'
import { fetchCaptcha, fetchProfile, updateMyPassword, updateProfile } from '@/api/auth'
import type { Profile } from '@/api/auth'
import { usePermissionStore } from '@/stores/permission'
import { useAvatarStore } from '@/stores/avatar'
import { MAX_IMAGE_UPLOAD_LABEL } from '@/constants/uploadLimits'
import { useUiPreferencesStore } from '@/stores/uiPreferences'
import { isValidOptionalPhone, normalizePhoneInput, sanitizePhoneInput } from '@/utils/phone'

const router = useRouter()
const perm = usePermissionStore()
const avatarStore = useAvatarStore()
const uiPrefs = useUiPreferencesStore()

const isLight = computed(() => uiPrefs.theme === 'light')

// 个人信息状态
const p = ref<Profile | null>(null)
const form = ref({
  name: '',
  phone: '',
  email: '',
  avatar_url: ''
})

// 密码修改状态
const pwd = reactive({
  old: '',
  next: '',
  confirm: '',
  captchaId: '',
  captchaCode: '',
  captchaImg: ''
})

// 头像相关状态
const avatarUploadErr = ref('')

// 加载状态
const isSaving = ref(false)
const isChangingPassword = ref(false)

// 计算属性
const avatarSrc = computed(() => {
  // 优先使用全局存储的头像
  if (avatarStore.avatarUrl) {
    return avatarStore.avatarUrl
  }

  const u = form.value.avatar_url.trim()
  return u || undefined
})

const avatarFallbackInitial = computed(() => {
  const n = form.value.name.trim()
  return n ? n.slice(0, 1).toUpperCase() : '?'
})

// 生命周期
onMounted(() => {
  load()
  refreshPwdCaptcha()
})

// 数据加载
async function load() {
  try {
    p.value = await fetchProfile()
    if (p.value) {
      form.value = {
        name: p.value.name,
        phone: p.value.phone || '',
        email: p.value.email || '',
        avatar_url: p.value.avatar_url || ''
      }
    }
  } catch (error) {
    console.error('加载个人资料失败:', error)
    ElMessage.error('加载个人资料失败')
  }
}

// 刷新验证码
async function refreshPwdCaptcha() {
  try {
    const c = await fetchCaptcha()
    pwd.captchaId = c.captcha_id
    pwd.captchaImg = c.image_base64
    pwd.captchaCode = ''
  } catch {
    pwd.captchaImg = ''
    ElMessage.error('获取验证码失败，请稍后重试')
  }
}

// 保存个人资料
async function saveProfile() {
  if (!form.value.name.trim()) {
    ElMessage.warning('请输入姓名')
    return
  }
  if (!isValidOptionalPhone(form.value.phone)) {
    ElMessage.warning('手机号需为 10-15 位数字')
    return
  }

  isSaving.value = true
  try {
    await updateProfile({
      name: form.value.name,
      phone: form.value.phone ? normalizePhoneInput(form.value.phone) : undefined,
      email: form.value.email || undefined,
      avatar_url: form.value.avatar_url.trim()
    })

    await perm.load()
    await load()

    ElMessage.success('个人资料已保存')
  } catch (error) {
    console.error('保存个人资料失败:', error)
    ElMessage.error('保存个人资料失败')
  } finally {
    isSaving.value = false
  }
}

// 修改密码
async function savePwd() {
  // 基本验证
  if (!pwd.old) {
    ElMessage.warning('请输入当前密码')
    return
  }

  if (!pwd.next) {
    ElMessage.warning('请输入新密码')
    return
  }

  if (pwd.next.length < 8) {
    ElMessage.warning('新密码至少需要8位')
    return
  }

  if (!/[A-Za-z]/.test(pwd.next) || !/\d/.test(pwd.next)) {
    ElMessage.warning('新密码须同时包含英文字母与数字')
    return
  }

  if (pwd.next !== pwd.confirm) {
    ElMessage.warning('两次输入的新密码不一致')
    return
  }

  if (!pwd.captchaCode) {
    ElMessage.warning('请输入图形验证码')
    return
  }

  isChangingPassword.value = true
  try {
    await updateMyPassword({
      old_password: pwd.old,
      new_password: pwd.next,
      new_password_confirm: pwd.confirm,
      captcha_id: pwd.captchaId,
      captcha_code: pwd.captchaCode
    })

    // 重置表单
    pwd.old = ''
    pwd.next = ''
    pwd.confirm = ''
    pwd.captchaCode = ''

    await refreshPwdCaptcha()

    ElMessage.success('密码修改成功')
  } catch (error: any) {
    console.error('修改密码失败:', error)
    ElMessage.error(error.message || '修改密码失败')
    await refreshPwdCaptcha()
  } finally {
    isChangingPassword.value = false
  }
}

// 头像相关函数

function handleAvatarCropped(dataUrl: string) {
  form.value.avatar_url = dataUrl
  avatarUploadErr.value = ''

  // 保存到全局存储
  const userId = String(perm.profile?.id || '')
  avatarStore.saveAvatarToStorage(userId, dataUrl)

  ElMessage.success('头像已更新')
}

function handleAvatarCroppedSuccess(dataUrl: string) {
  handleAvatarCropped(dataUrl)
}

function handleAvatarCroppedError(error: string) {
  avatarUploadErr.value = error
  ElMessage.error(`头像裁剪失败: ${error}`)
}

function clearAvatar() {
  form.value.avatar_url = ''
  avatarUploadErr.value = ''
  ElMessage.info('头像已清除')
}

function closeAvatarCropper() {
  // 不需要特殊处理，NeuroImageCropper组件会自己处理
}

// 导航
function goBack() {
  router.back()
}

// 滚动激活标题栏 sticky：只有滚动超过标题栏自身高度时才锁定
const headerRef = ref<HTMLElement | null>(null)

function handleScroll() {
  if (!headerRef.value) return
  const headerBottom = headerRef.value.getBoundingClientRect().bottom
  if (headerBottom <= 0) {
    headerRef.value.classList.add('sticky')
  } else {
    headerRef.value.classList.remove('sticky')
  }
}

onMounted(() => {
  load()
  refreshPwdCaptcha()
  window.addEventListener('scroll', handleScroll, { passive: true })
})

onUnmounted(() => {
  window.removeEventListener('scroll', handleScroll)
})
</script>

<style scoped>
.neuro-profile-view {
  min-height: 100vh;
  background: var(--nm-bg-surface, linear-gradient(135deg, #0a0a0f 0%, #1a1a2e 100%));
  font-family: 'JetBrains Mono', monospace;
  color: var(--nm-text-primary, #e0e0ff);
}

/* 头部样式 */
.neuro-header {
  background: var(--nm-bg-card, rgba(16, 16, 32, 0.9));
  border-bottom: 1px solid var(--nm-border, rgba(0, 245, 212, 0.3));
  padding: 0 32px;
  height: 70px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  backdrop-filter: blur(10px);
  position: relative;
  z-index: 100;
}

/* 滚动后标题栏锁定到顶部 */
.neuro-header.sticky {
  position: sticky;
  top: 0;
}

.header-content .neuro-title {
  font-size: 15px;
  font-weight: 700;
  margin: 0;
  background: linear-gradient(90deg, #00f5d4, #00b4d8);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  display: flex;
  align-items: center;
  gap: 6px;
}

.title-icon {
  font-size: 18px;
}

.neuro-subtitle {
  font-size: 11px;
  color: #8a8aff;
  opacity: 0.8;
  letter-spacing: 0.5px;
}

.header-actions {
  display: flex;
  gap: 10px;
}

/* 按钮样式 */
.neuro-btn {
  padding: 8px 16px;
  border: none;
  border-radius: 6px;
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

.neuro-btn::before {
  content: '';
  position: absolute;
  top: 0;
  left: -100%;
  width: 100%;
  height: 100%;
  background: linear-gradient(90deg, transparent, rgba(255, 255, 255, 0.1), transparent);
  transition: left 0.5s ease;
}

.neuro-btn:hover::before {
  left: 100%;
}

.back-btn {
  background: rgba(255, 255, 255, 0.05);
  color: #e0e0ff;
  border: 1px solid rgba(0, 245, 212, 0.3);
  padding: 4px 10px;
  border-radius: 8px;
}

.back-btn:hover {
  background: rgba(0, 245, 212, 0.1);
  border-color: #00f5d4;
  transform: translateX(-4px);
}

.primary-btn {
  background: linear-gradient(135deg, #00f5d4, #00b4d8);
  color: #0a0a0f;
  border: none;
}

.primary-btn:hover {
  background: linear-gradient(135deg, #00e6c7, #0099cc);
  transform: translateY(-2px);
  box-shadow: 0 4px 20px rgba(0, 245, 212, 0.4);
}

.secondary-btn {
  background: rgba(255, 255, 255, 0.05);
  color: #e0e0ff;
  border: 1px solid rgba(255, 255, 255, 0.1);
}

.secondary-btn:hover:not(:disabled) {
  background: rgba(255, 255, 255, 0.1);
  border-color: rgba(255, 255, 255, 0.3);
}

.secondary-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.save-btn, .security-btn {
  background: linear-gradient(135deg, #9d4edd, #560bad);
  color: white;
  border: none;
}

.save-btn:hover:not(:disabled), .security-btn:hover:not(:disabled) {
  background: linear-gradient(135deg, #8a2be2, #4a0080);
  transform: translateY(-2px);
  box-shadow: 0 4px 20px rgba(157, 78, 221, 0.4);
}

.save-btn:disabled, .security-btn:disabled {
  opacity: 0.7;
  cursor: not-allowed;
  transform: none !important;
}

.btn-icon.loading {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

/* 内容区域 */
.neuro-content {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 32px;
  padding: 32px;
  max-width: 1400px;
  margin: 0 auto;
}

@media (max-width: 1024px) {
  .neuro-content {
    grid-template-columns: 1fr;
  }
}

/* 卡片样式 */
.neuro-card {
  background: var(--nm-bg-card, rgba(16, 16, 32, 0.7));
  border-radius: 12px;
  border: 1px solid var(--nm-border, rgba(0, 245, 212, 0.2));
  overflow: hidden;
  backdrop-filter: blur(10px);
  position: relative;
}

.neuro-card::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 1px;
  background: linear-gradient(90deg, transparent, #00f5d4, transparent);
}

.card-header {
  padding: 24px 28px;
  border-bottom: 1px solid var(--nm-border, rgba(0, 245, 212, 0.1));
}

.card-title {
  font-size: 18px;
  font-weight: 600;
  margin-bottom: 4px;
  color: var(--nm-primary, #00f5d4);
  display: flex;
  align-items: center;
  gap: 8px;
}

.card-subtitle {
  font-size: 13px;
  color: var(--nm-text-secondary, #8a8aff);
  opacity: 0.8;
}

.card-content {
  padding: 24px 28px;
}

/* 头像区域 */
.avatar-section {
  margin-bottom: 32px;
}

.avatar-container {
  display: flex;
  align-items: flex-start;
  gap: 32px;
}

@media (max-width: 768px) {
  .avatar-container {
    flex-direction: column;
    align-items: flex-start;
  }
}

.avatar-preview {
  width: 130px;
  height: 130px;
  border-radius: 50%;
  position: relative;
  overflow: hidden;
  border: 3px solid transparent;
  background: linear-gradient(135deg, #00f5d4, #9d4edd) border-box;
  flex-shrink: 0;
}

.avatar-preview.has-avatar {
  border: 3px solid #00f5d4;
}

.avatar-image {
  width: 100%;
  height: 100%;
  object-fit: cover;
  border-radius: 50%;
}

.avatar-placeholder {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #1a1a2e, #16213e);
}

.placeholder-text {
  font-size: 44px;
  font-weight: bold;
  color: #00f5d4;
}

.avatar-glow {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  border-radius: 50%;
  box-shadow: inset 0 0 20px rgba(0, 245, 212, 0.3);
  pointer-events: none;
}

.avatar-controls {
  flex: 1;
}

.controls-title {
  font-size: 17px;
  margin-bottom: 6px;
  color: var(--nm-text-primary, #e0e0ff);
}

.controls-description {
  font-size: 13px;
  color: var(--nm-text-secondary, #8a8aff);
  margin-bottom: 16px;
  opacity: 0.8;
}

.controls-buttons {
  display: flex;
  gap: 12px;
  margin-bottom: 16px;
}

.avatar-info {
  background: rgba(0, 245, 212, 0.05);
  border: 1px solid rgba(0, 245, 212, 0.1);
  border-radius: 8px;
  padding: 12px;
  margin-bottom: 12px;
}

.info-item {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 6px;
  font-size: 13px;
  color: var(--nm-text-secondary, #8a8aff);
}

.info-item:last-child {
  margin-bottom: 0;
}

.info-icon {
  font-size: 12px;
}

.error-message {
  background: rgba(255, 87, 87, 0.1);
  border: 1px solid rgba(255, 87, 87, 0.3);
  border-radius: 6px;
  padding: 8px;
  color: #ff5757;
  font-size: 12px;
  display: flex;
  align-items: center;
  gap: 6px;
}

.error-icon {
  font-size: 14px;
}

/* 表单样式 */
.form-section {
  margin-top: 32px;
}

.form-title {
  font-size: 17px;
  margin-bottom: 20px;
  color: var(--nm-primary, #00f5d4);
  display: flex;
  align-items: center;
  gap: 8px;
}

.neuro-form {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.form-label {
  font-size: 14px;
  font-weight: 600;
  color: var(--nm-text-primary, #e0e0ff);
  display: flex;
  align-items: center;
  gap: 8px;
}

.label-icon {
  font-size: 14px;
}

.form-input-container {
  position: relative;
}

.neuro-input {
  width: 100%;
  max-width: 100%;
  box-sizing: border-box;
  padding: 10px 14px;
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid var(--nm-border, rgba(0, 245, 212, 0.2));
  border-radius: 8px;
  color: var(--nm-text-primary, #e0e0ff);
  font-family: 'JetBrains Mono', monospace;
  font-size: 14px;
  transition: all 0.3s ease;
}

.neuro-input:focus {
  outline: none;
  border-color: var(--nm-primary, #00f5d4);
  background: rgba(0, 245, 212, 0.05);
  box-shadow: 0 0 0 2px rgba(0, 245, 212, 0.1);
}

.neuro-input::placeholder {
  color: var(--nm-text-secondary, #8a8aff);
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

.neuro-input:focus + .input-glow {
  opacity: 1;
}

.form-actions {
  margin-top: 24px;
}

/* 安全信息 */
.security-info {
  background: rgba(157, 78, 221, 0.05);
  border: 1px solid rgba(157, 78, 221, 0.1);
  border-radius: 8px;
  padding: 16px;
  margin-bottom: 24px;
}

.info-text {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
  font-size: 13px;
  color: var(--nm-text-secondary, #8a8aff);
}

.info-text:last-child {
  margin-bottom: 0;
}

/* 验证码容器 */
.captcha-container {
  display: flex;
  gap: 10px;
  align-items: center;
}

.captcha-container .form-input-container {
  flex: 1;
}

.captcha-image-container {
  flex-shrink: 0;
}

.captcha-image {
  height: 36px;
  border-radius: 6px;
  cursor: pointer;
  border: 1px solid rgba(0, 245, 212, 0.3);
  transition: all 0.3s ease;
}

.captcha-image:hover {
  border-color: #00f5d4;
  transform: scale(1.05);
}

.captcha-btn {
  padding: 8px 12px;
  white-space: nowrap;
  font-size: 12px;
}

/* 模态框样式 */

/* ===== 浅色模式覆盖 ===== */
.neuro-profile-view.light-mode {
  --nm-bg-surface: #f8fafc;
  --nm-bg-card: #ffffff;
  --nm-bg-elevated: #f1f5f9;
  --nm-primary: #0ea5e9;
  --nm-text-primary: #0f172a;
  --nm-text-secondary: #475569;
  --nm-border: #cbd5e1;
}

.neuro-profile-view.light-mode .neuro-header {
  background: var(--nm-bg-card);
  border-bottom-color: var(--nm-border);
}

.neuro-profile-view.light-mode .neuro-card {
  background: var(--nm-bg-card);
  border-color: var(--nm-border);
}

.neuro-profile-view.light-mode .neuro-card::before {
  background: linear-gradient(90deg, transparent, var(--nm-primary), transparent);
}

.neuro-profile-view.light-mode .card-title {
  color: var(--nm-primary);
}

.neuro-profile-view.light-mode .card-subtitle {
  color: var(--nm-text-secondary);
}

.neuro-profile-view.light-mode .form-title {
  color: var(--nm-primary);
}

.neuro-profile-view.light-mode .form-label {
  color: var(--nm-text-primary);
}

.neuro-profile-view.light-mode .controls-title {
  color: var(--nm-text-primary);
}

.neuro-profile-view.light-mode .controls-description {
  color: var(--nm-text-secondary);
}

.neuro-profile-view.light-mode .info-item {
  color: var(--nm-text-secondary);
}

.neuro-profile-view.light-mode .info-text {
  color: var(--nm-text-secondary);
}

.neuro-profile-view.light-mode .neuro-input {
  background: #f8fafc;
  border-color: var(--nm-border);
  color: var(--nm-text-primary);
}

.neuro-profile-view.light-mode .neuro-input:focus {
  border-color: var(--nm-primary);
  background: #f0f9ff;
}

.neuro-profile-view.light-mode .neuro-input::placeholder {
  color: #94a3b8;
}

.neuro-profile-view.light-mode .back-btn {
  background: #f1f5f9;
  color: var(--nm-text-primary);
  border-color: var(--nm-border);
}

.neuro-profile-view.light-mode .back-btn:hover {
  background: var(--nm-primary);
  color: #ffffff;
  border-color: var(--nm-primary);
}

.neuro-profile-view.light-mode .secondary-btn {
  background: #f1f5f9;
  color: var(--nm-text-primary);
  border-color: var(--nm-border);
}

.neuro-profile-view.light-mode .secondary-btn:hover:not(:disabled) {
  background: #e2e8f0;
  border-color: var(--nm-border);
}

.neuro-profile-view.light-mode .security-info {
  background: rgba(139, 92, 246, 0.06);
  border-color: rgba(139, 92, 246, 0.15);
}

.neuro-profile-view.light-mode .avatar-info {
  background: rgba(14, 165, 233, 0.05);
  border-color: rgba(14, 165, 233, 0.12);
}

.neuro-profile-view.light-mode .error-message {
  background: rgba(239, 68, 68, 0.06);
  border-color: rgba(239, 68, 68, 0.25);
  color: #dc2626;
}

.neuro-profile-view.light-mode .captcha-image {
  border-color: rgba(14, 165, 233, 0.3);
}

.neuro-profile-view.light-mode .captcha-image:hover {
  border-color: var(--nm-primary);
}

.neuro-profile-view.light-mode .neuro-btn::before {
  background: linear-gradient(90deg, transparent, rgba(0, 0, 0, 0.05), transparent);
}

.neuro-profile-view.light-mode .title-icon {
  /* emoji 在浅色下保持原样 */
}

.neuro-profile-view.light-mode .label-icon,
.neuro-profile-view.light-mode .info-icon,
.neuro-profile-view.light-mode .error-icon {
  /* emoji 保持原样 */
}

.neuro-profile-view.light-mode .captcha-btn {
  background: #f1f5f9;
  color: var(--nm-text-primary);
  border: 1px solid var(--nm-border);
}

.neuro-profile-view.light-mode .captcha-btn:hover {
  background: #e2e8f0;
}

.neuro-profile-view.light-mode .neuro-btn.primary-btn {
  background: linear-gradient(135deg, #0ea5e9, #0284c7);
  color: #ffffff;
}

.neuro-profile-view.light-mode .neuro-btn.primary-btn:hover {
  background: linear-gradient(135deg, #0284c7, #0369a1);
  box-shadow: 0 4px 20px rgba(14, 165, 233, 0.3);
}

.neuro-profile-view.light-mode .neuro-btn.save-btn,
.neuro-profile-view.light-mode .neuro-btn.security-btn {
  background: linear-gradient(135deg, #8b5cf6, #7c3aed);
  color: #ffffff;
}

.neuro-profile-view.light-mode .neuro-btn.save-btn:hover:not(:disabled),
.neuro-profile-view.light-mode .neuro-btn.security-btn:hover:not(:disabled) {
  background: linear-gradient(135deg, #7c3aed, #6d28d9);
  box-shadow: 0 4px 20px rgba(139, 92, 246, 0.3);
}
</style>