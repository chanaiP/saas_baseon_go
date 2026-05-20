<script setup lang="ts">
import { Moon, Sunny } from '@element-plus/icons-vue'
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { resolveHealthUrl } from '@/api/http'
import {
  fetchCaptcha,
  fetchPhoneLoginTenants,
  login,
  type LoginOutcome,
  type LoginTenantOption,
} from '@/api/auth'
import { usePermissionStore } from '@/stores/permission'
import { DEFAULT_LOGIN_BRAND_NAME, useTenantBrandingStore } from '@/stores/tenantBranding'
import { useUiPreferencesStore } from '@/stores/uiPreferences'
import { isLikelyPhoneAccountInput } from '@/utils/phone'

const route = useRoute()
const router = useRouter()
const perm = usePermissionStore()
const uiPrefs = useUiPreferencesStore()
const tenantBrand = useTenantBrandingStore()

// 深色/浅色模式状态
const isLightMode = ref(false)

/** 种子演示账号：E10001 / 13800000000 与 E10100 / 13900000000，默认密码均为 112233 */
const account = ref('E10001')
const password = ref('112233')
const captchaId = ref('')
const captchaImg = ref('')
const captchaCode = ref('')
const needCaptcha = ref(false)
const loading = ref(false)
const err = ref('')
const tenantPickVisible = ref(false)
const tenantPickList = ref<LoginTenantOption[]>([])
const pickedTenantId = ref<number | null>(null)

/** 手机号多空间：接口探测结果与所选空间 */
const phoneTenantOptions = ref<LoginTenantOption[]>([])
const selectedPhoneTenantId = ref<number | null>(null)
const phoneTenantsLoading = ref(false)
let phoneTenantDebounce: ReturnType<typeof setTimeout> | null = null

const looksLikePhone = computed(() => isLikelyPhoneAccountInput(account.value))
const showPhoneTenantPicker = computed(() => phoneTenantOptions.value.length > 1)
/** 打开登录页时探测 Vite→后端代理是否可用（失败时 Network 里常为 login 红叉且无响应头） */
const backendHint = ref('')
const loginLogoErr = ref(false)
/** 仅来自上次登录后写入 localStorage 的后台品牌；首次访问无缓存则为系统默认 */
const loginLogoSrc = computed(() => tenantBrand.displayLogoSrc)
const loginTitle = DEFAULT_LOGIN_BRAND_NAME
const showLoginBrandImg = computed(() => !!loginLogoSrc.value && !loginLogoErr.value)

watch(loginLogoSrc, () => {
  loginLogoErr.value = false
})

watch(account, (v) => {
  if (phoneTenantDebounce) {
    clearTimeout(phoneTenantDebounce)
    phoneTenantDebounce = null
  }
  if (!isLikelyPhoneAccountInput(v)) {
    phoneTenantOptions.value = []
    selectedPhoneTenantId.value = null
    phoneTenantsLoading.value = false
    return
  }
  phoneTenantDebounce = setTimeout(() => {
    void (async () => {
      phoneTenantsLoading.value = true
      try {
        const rows = await fetchPhoneLoginTenants(v.trim())
        phoneTenantOptions.value = rows
        if (rows.length === 1) {
          selectedPhoneTenantId.value = rows[0].tenant_id
        } else if (rows.length > 1) {
          selectedPhoneTenantId.value = null
        } else {
          selectedPhoneTenantId.value = null
        }
      } catch {
        phoneTenantOptions.value = []
        selectedPhoneTenantId.value = null
      } finally {
        phoneTenantsLoading.value = false
      }
    })()
  }, 400)
})

const REM = 'remembered_account'

async function checkBackend() {
  try {
    const r = await fetch(resolveHealthUrl(), { method: 'GET' })
    if (r.ok) {
      backendHint.value = ''
      return
    }
    if (r.status === 503) {
      backendHint.value =
        'Go API 已启动，但 PostgreSQL 或 Redis 不可用。请在项目根目录执行 docker compose up -d，并检查 DATABASE_DSN 与 REDIS_ADDR。'
      return
    }
    if (r.status === 502 || r.status === 504) {
      backendHint.value =
        `健康检查返回 HTTP ${r.status}（多为 Vite 连不上 Go API 8081）。请确认后端已启动：DB_AUTO_MIGRATE=false go run ./cmd/api；并查看运行 npm run dev 的终端是否出现 [vite-proxy] … ECONNREFUSED。`
      return
    }
    backendHint.value = `健康检查返回 HTTP ${r.status}，请查看后端日志。`
  } catch {
    backendHint.value =
      '当前无法连上后端（Vite 会把 /api、/health 代理到本机 8081）。请先：① 项目根目录 docker compose up -d；② 启动 Go API：DB_AUTO_MIGRATE=false go run ./cmd/api。'
  }
}

onMounted(() => {
  void tenantBrand.loadPublicFooterIfEmpty()
  account.value = localStorage.getItem(REM)?.trim() || 'E10001'
  void checkBackend()

  // 同步管理后台主题模式（与 uiPrefsStore 保持一致）
  isLightMode.value = uiPrefs.theme === 'light'
})

async function refreshCaptcha() {
  const c = await fetchCaptcha()
  captchaId.value = c.captcha_id
  captchaImg.value = c.image_base64
}

async function handleLoginOutcome(outcome: LoginOutcome) {
  if (outcome.kind === 'success') {
    localStorage.setItem('access_token', outcome.token)
    localStorage.setItem(REM, account.value)
    try {
      await perm.load({ force: true })
    } catch (e) {
      localStorage.removeItem('access_token')
      perm.clear()
      throw e
    }
    try {
      await tenantBrand.load()
    } catch {
      /* 品牌失败不阻断登录 */
    }
    const redir = (route.query.redirect as string) || '/home'
    router.replace(redir)
    return
  }
  if (outcome.kind === 'captcha') {
    needCaptcha.value = true
    err.value = outcome.message
    await refreshCaptcha()
    return
  }
  tenantPickList.value = outcome.tenants
  pickedTenantId.value = outcome.tenants[0]?.tenant_id ?? null
  tenantPickVisible.value = true
  err.value = outcome.message
}

function resolvePhoneLoginTenantId(): number | undefined {
  if (!looksLikePhone.value || phoneTenantOptions.value.length === 0) return undefined
  if (phoneTenantOptions.value.length === 1) return phoneTenantOptions.value[0].tenant_id
  if (selectedPhoneTenantId.value != null) return selectedPhoneTenantId.value
  return undefined
}

async function submit() {
  err.value = ''
  if (showPhoneTenantPicker.value && selectedPhoneTenantId.value == null) {
    err.value = '请选择空间'
    return
  }

  loading.value = true
  try {
    await checkBackend()
    const outcome = await login({
      account: account.value,
      password: password.value,
      captcha_id: needCaptcha.value ? captchaId.value : undefined,
      captcha_code: needCaptcha.value ? captchaCode.value : undefined,
      tenant_id: resolvePhoneLoginTenantId(),
    })
    await handleLoginOutcome(outcome)
  } catch (e: unknown) {
    err.value = e instanceof Error ? e.message : '登录失败'
  } finally {
    loading.value = false
  }
}

async function confirmTenantPick() {
  if (pickedTenantId.value == null) {
    err.value = '请选择空间'
    return
  }
  err.value = ''
  loading.value = true
  try {
    await checkBackend()
    const outcome = await login({
      account: account.value,
      password: password.value,
      tenant_id: pickedTenantId.value,
      captcha_id: needCaptcha.value ? captchaId.value : undefined,
      captcha_code: needCaptcha.value ? captchaCode.value : undefined,
    })
    tenantPickVisible.value = false
    await handleLoginOutcome(outcome)
  } catch (e: unknown) {
    err.value = e instanceof Error ? e.message : '登录失败'
  } finally {
    loading.value = false
  }
}

// 切换深色/浅色模式
function toggleThemeMode() {
  isLightMode.value = !isLightMode.value
  localStorage.setItem('login-theme-mode', isLightMode.value ? 'light' : 'dark')

  // 同时更新全局主题（如果存在）
  if (uiPrefs) {
    uiPrefs.setTheme(isLightMode.value ? 'light' : 'dark')
  }
}

// 点击品牌区域：聚焦账号输入框，快速开始登录
const accountInputRef = ref()
function brandClickHandler() {
  accountInputRef.value?.focus()
}
</script>

<template>
  <div class="login-page-root login-tech" :class="{ 'light-mode': isLightMode }">
    <!--
      可选底图：覆盖 public/login-bg.png 后会以低透明度叠在动效背景上；无文件时仅显示内置科技风场景。
    -->
    <div class="login-bg-layer" aria-hidden="true">
      <!-- 高级科技AI背景 -->
      <div class="tech-ai-bg">
        <!-- 基础渐变背景 -->
        <div class="bg-base-gradient"></div>

        <!-- 网格系统 -->
        <div class="tech-grid tech-grid-1"></div>
        <div class="tech-grid tech-grid-2"></div>

        <!-- 数据流层 -->
        <div class="data-flow-layer">
          <div class="data-flow data-flow-1"></div>
          <div class="data-flow data-flow-2"></div>
          <div class="data-flow data-flow-3"></div>
          <div class="data-flow data-flow-4"></div>
        </div>

        <!-- 光晕效果 -->
        <div class="tech-glow tech-glow-1"></div>
        <div class="tech-glow tech-glow-2"></div>
        <div class="tech-glow tech-glow-3"></div>


        <!-- 粒子系统 -->
        <div class="particles-container">
          <div class="particle particle-1"></div>
          <div class="particle particle-2"></div>
          <div class="particle particle-3"></div>
          <div class="particle particle-4"></div>
          <div class="particle particle-5"></div>
          <div class="particle particle-6"></div>
          <div class="particle particle-7"></div>
          <div class="particle particle-8"></div>
          <div class="particle particle-9"></div>
          <div class="particle particle-10"></div>
        </div>

        <!-- 全息投影效果 -->
        <div class="hologram hologram-1"></div>
        <div class="hologram hologram-2"></div>

        <!-- 神经元网络层（保留但增强） -->
        <div class="enhanced-neural-network">
          <div class="neural-node neural-node-1"></div>
          <div class="neural-node neural-node-2"></div>
          <div class="neural-node neural-node-3"></div>
          <div class="neural-node neural-node-4"></div>
          <div class="neural-node neural-node-5"></div>
          <div class="neural-node neural-node-6"></div>

          <div class="neural-connection neural-conn-1"></div>
          <div class="neural-connection neural-conn-2"></div>
          <div class="neural-connection neural-conn-3"></div>
          <div class="neural-connection neural-conn-4"></div>
        </div>

        <!-- 浅色模式网格闪烁点 -->
        <div class="grid-sparkles" aria-hidden="true">
          <div class="grid-sparkle grid-sparkle-1"></div>
          <div class="grid-sparkle grid-sparkle-2"></div>
          <div class="grid-sparkle grid-sparkle-3"></div>
          <div class="grid-sparkle grid-sparkle-4"></div>
          <div class="grid-sparkle grid-sparkle-5"></div>
          <div class="grid-sparkle grid-sparkle-6"></div>
          <div class="grid-sparkle grid-sparkle-7"></div>
          <div class="grid-sparkle grid-sparkle-8"></div>
        </div>
      </div>
    </div>
    <div class="login-page-content">
    <div class="theme-bar">
      <el-tooltip :content="isLightMode ? '切换为深色模式' : '切换为浅色模式'" placement="left">
        <el-button
          class="theme-orbit-btn"
          :icon="isLightMode ? Moon : Sunny"
          circle
          @click="toggleThemeMode"
        />
      </el-tooltip>
    </div>
    <div class="login-main">
    <div class="panel">
      <div class="panel-edge panel-edge--tl" aria-hidden="true" />
      <div class="panel-edge panel-edge--tr" aria-hidden="true" />
      <div class="panel-edge panel-edge--bl" aria-hidden="true" />
      <div class="panel-edge panel-edge--br" aria-hidden="true" />
      <div class="panel-glow-ring" aria-hidden="true" />
      <div class="login-brand" @click="brandClickHandler">
        <div class="login-logo-wrap">
          <div v-if="!showLoginBrandImg" class="logo-orbit" aria-hidden="true" />
          <img
            v-if="showLoginBrandImg"
            :src="loginLogoSrc"
            alt=""
            class="login-logo-img custom-logo-img"
            @error="loginLogoErr = true"
          />
          <div v-else class="brand-mark" aria-hidden="true">
            <div class="neuron-axon"></div>
            <div class="neuron-synapse neuron-synapse-1"></div>
            <div class="neuron-synapse neuron-synapse-2"></div>
            <div class="neuron-synapse neuron-synapse-3"></div>
            <div class="neuron-synapse neuron-synapse-4"></div>
            <div class="neuron-synapse neuron-synapse-5"></div>
          </div>
        </div>
        <p class="eyebrow">SECURE ACCESS</p>
        <h1 class="title">{{ loginTitle }}</h1>
      </div>
      <p class="sub">使用工号或手机号登录</p>
      <p class="demo-hint">
        演示账号：<span class="mono">工号 E10001</span> 或 <span class="mono">手机 13800000000</span>，
        密码 <span class="mono">112233</span>（E10100 / 13900000000 同密码）
      </p>
      <el-alert
        v-if="backendHint"
        :title="backendHint"
        type="warning"
        show-icon
        :closable="false"
        class="hint-alert"
      />
      <el-form class="form" label-position="top" @submit.prevent="submit">
        <el-form-item label="账号">
          <el-input
            ref="accountInputRef"
            v-model="account"
            size="large"
            autocomplete="username"
            placeholder="工号或手机号"
            class="input-round"
            clearable
          />
        </el-form-item>
        <el-form-item v-if="showPhoneTenantPicker" label="选择空间">
          <el-select
            v-model="selectedPhoneTenantId"
            placeholder="请选择空间名称"
            size="large"
            class="input-round tenant-select"
            :loading="phoneTenantsLoading"
            clearable
          >
            <el-option
              v-for="t in phoneTenantOptions"
              :key="t.tenant_id"
              :label="`${t.name}（${t.code}）`"
              :value="t.tenant_id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="密码">
          <el-input
            v-model="password"
            size="large"
            type="password"
            show-password
            autocomplete="current-password"
            clearable
            class="input-round"
          />
        </el-form-item>
        <el-form-item v-if="needCaptcha" label="验证码">
          <div class="cap">
            <el-input v-model="captchaCode" size="large" placeholder="验证码" class="input-round cap-input" />
            <img v-if="captchaImg" class="cap-img" :src="captchaImg" alt="" @click="refreshCaptcha" />
          </div>
        </el-form-item>
        <el-alert v-if="err" :title="err" type="error" show-icon :closable="false" class="err-alert" />
        <el-button
          type="primary"
          size="large"
          class="submit-btn"
          :loading="loading"
          :disabled="loading"
          native-type="submit"
        >
          登录
        </el-button>
      </el-form>
    </div>
    </div>
    <footer v-if="tenantBrand.footerDisplay" class="shell-footer">
      {{ tenantBrand.footerDisplay }}
    </footer>

    <el-dialog
      v-model="tenantPickVisible"
      class="login-tech-dialog"
      title="选择登录空间"
      width="420px"
      append-to-body
      :close-on-click-modal="false"
    >
      <p class="tenant-pick-hint">请选择要进入的空间。</p>
      <el-radio-group v-model="pickedTenantId" class="tenant-pick-group">
        <el-radio v-for="t in tenantPickList" :key="t.tenant_id" :value="t.tenant_id" class="tenant-pick-row">
          {{ t.name }}（{{ t.code }}）
        </el-radio>
      </el-radio-group>
      <template #footer>
        <el-button @click="tenantPickVisible = false">取消</el-button>
        <el-button type="primary" :loading="loading" @click="confirmTenantPick">确定</el-button>
      </template>
    </el-dialog>
    </div>
  </div>
</template>

<style scoped>
/* 根布局占满 #app 高度；body 已 overflow:hidden，登录页自身负责纵向滚动（小屏/键盘） */
.login-page-root {
  height: 100%;
  min-height: 0;
  overflow-x: hidden;
  overflow-y: auto;
  box-sizing: border-box;
}

.login-page-root.login-tech {
  /* 深色模式变量（默认） */
  --neural-primary: #00ff9d;
  --neural-secondary: #00d4ff;
  --neural-accent: #9d4edd;
  --neural-glow: rgba(0, 255, 157, 0.45);
  --neural-bg-dark: #0a0a1a;
  --neural-bg-deep: #050510;
  --neural-fg: #f0f8ff;
  --neural-fg-muted: rgba(240, 248, 255, 0.65);
  --neural-panel: rgba(10, 15, 30, 0.85);
  --neural-border: rgba(0, 255, 157, 0.25);
  --neural-input-bg: rgba(5, 10, 20, 0.7);

  /* 浅色模式变量 - 生物数字融合主题（增强版） */
  --neural-primary-light: #2a8c6c;      /* 深蓝绿色 - 代表神经元核心 */
  --neural-secondary-light: #4a90e2;    /* 科技蓝 - 代表神经脉冲 */
  --neural-accent-light: #8a4db8;       /* 紫红色 - 代表突触连接 */
  --neural-glow-light: rgba(42, 140, 108, 0.35);

  /* 背景色调 - 有机细胞膜质感 */
  --neural-bg-light: #f0f8f5;           /* 浅蓝绿底色，与神经元和谐 */
  --neural-bg-light-deep: #e0f0ea;      /* 稍深的有机绿 */

  /* 文字颜色 - 深有机绿，与背景形成和谐对比 */
  --neural-fg-light: #1a3c2a;
  --neural-fg-muted-light: rgba(26, 60, 42, 0.7);

  /* 面板 - 有机玻璃质感，带轻微绿色调 */
  --neural-panel-light: rgba(255, 255, 255, 0.94);
  --neural-border-light: rgba(42, 140, 108, 0.3);
  --neural-input-bg-light: rgba(255, 255, 255, 0.92);

  font-family: 'IBM Plex Sans', system-ui, sans-serif;
  transition: background-color 0.5s ease, color 0.5s ease;
}

/* 浅色模式 */
.login-page-root.login-tech.light-mode {
  --neural-primary: var(--neural-primary-light);
  --neural-secondary: var(--neural-secondary-light);
  --neural-accent: var(--neural-accent-light);
  --neural-glow: var(--neural-glow-light);
  --neural-bg-dark: var(--neural-bg-light);
  --neural-bg-deep: var(--neural-bg-light-deep);
  --neural-fg: var(--neural-fg-light);
  --neural-fg-muted: var(--neural-fg-muted-light);
  --neural-panel: var(--neural-panel-light);
  --neural-border: var(--neural-border-light);
  --neural-input-bg: var(--neural-input-bg-light);
}

.login-bg-layer {
  position: fixed;
  inset: 0;
  z-index: 0;
  overflow: hidden;
  pointer-events: none;
}

.tech-ai-bg {
  position: absolute;
  inset: 0;
  overflow: hidden;
  background: #000;
}

/* 基础渐变背景 */
.bg-base-gradient {
  position: absolute;
  inset: 0;
  background:
    radial-gradient(ellipse 80% 60% at 20% 20%, rgba(0, 40, 80, 0.4) 0%, transparent 60%),
    radial-gradient(ellipse 60% 80% at 80% 30%, rgba(80, 0, 120, 0.3) 0%, transparent 60%),
    radial-gradient(ellipse 70% 70% at 50% 80%, rgba(0, 100, 100, 0.2) 0%, transparent 70%),
    linear-gradient(135deg, #0a0a1a 0%, #151530 50%, #0a0a1a 100%);
  animation: gradient-shift 20s ease-in-out infinite alternate;
}

.light-mode .bg-base-gradient {
  background:
    /* 有机细胞纹理 - 增强神经元主题 */
    radial-gradient(ellipse 60% 50% at 30% 40%, rgba(42, 140, 108, 0.12) 0%, transparent 70%),
    radial-gradient(ellipse 50% 60% at 70% 30%, rgba(74, 144, 226, 0.08) 0%, transparent 65%),
    radial-gradient(ellipse 55% 55% at 50% 70%, rgba(138, 77, 184, 0.06) 0%, transparent 60%),

    /* 生物膜质感渐变 - 与神经元颜色协调 */
    linear-gradient(
      135deg,
      rgba(240, 248, 245, 0.98) 0%,
      rgba(224, 240, 234, 0.95) 25%,
      rgba(210, 235, 225, 0.92) 50%,
      rgba(224, 240, 234, 0.95) 75%,
      rgba(240, 248, 245, 0.98) 100%
    ),

    /* 微妙的神经元纹理 */
    repeating-linear-gradient(
      45deg,
      transparent,
      transparent 3px,
      rgba(42, 140, 108, 0.03) 3px,
      rgba(42, 140, 108, 0.03) 6px
    ),

    /* 细胞膜效果 */
    repeating-radial-gradient(
      circle at 50% 50%,
      transparent,
      transparent 10px,
      rgba(74, 144, 226, 0.02) 10px,
      rgba(74, 144, 226, 0.02) 20px
    );

  animation: light-bg-pulse 20s ease-in-out infinite alternate;
}

@keyframes light-bg-pulse {
  0% {
    background-position: 0% 0%, 0% 0%, 0% 0%, 0% 0%, 0% 0%, 0% 0%;
    filter: hue-rotate(0deg) brightness(1);
  }
  100% {
    background-position: 10% 5%, -5% 10%, 5% -5%, 100% 100%, 20px 20px, 40px 40px;
    filter: hue-rotate(3deg) brightness(1.03);
  }
}

@keyframes gradient-shift {
  0% {
    filter: hue-rotate(0deg);
    background-position: 0% 0%;
  }
  100% {
    filter: hue-rotate(30deg);
    background-position: 100% 100%;
  }
}

/* 网格系统 */
.tech-grid {
  position: absolute;
  inset: 0;
  background-image:
    linear-gradient(rgba(0, 255, 157, 0.05) 1px, transparent 1px),
    linear-gradient(90deg, rgba(0, 255, 157, 0.05) 1px, transparent 1px);
  background-size: 60px 60px;
  opacity: 0.3;
}

.light-mode .tech-grid {
  background-image:
    linear-gradient(rgba(42, 140, 108, 0.08) 1px, transparent 1px),
    linear-gradient(90deg, rgba(42, 140, 108, 0.08) 1px, transparent 1px),
    /* 对角线网格 */
    linear-gradient(45deg, rgba(74, 144, 226, 0.03) 1px, transparent 1px),
    linear-gradient(-45deg, rgba(74, 144, 226, 0.03) 1px, transparent 1px),
    /* 细密网格 */
    linear-gradient(rgba(138, 77, 184, 0.02) 0.5px, transparent 0.5px),
    linear-gradient(90deg, rgba(138, 77, 184, 0.02) 0.5px, transparent 0.5px);
  background-size:
    60px 60px,  /* 主网格 */
    60px 60px,
    80px 80px,  /* 对角线网格 */
    80px 80px,
    20px 20px,  /* 细密网格 */
    20px 20px;
  opacity: 0.6;
  animation:
    light-grid-drift 40s linear infinite,
    grid-pulse 8s ease-in-out infinite;
  mix-blend-mode: multiply;
}

.light-mode .tech-grid-1 {
  background-image:
    linear-gradient(rgba(74, 144, 226, 0.04) 1px, transparent 1px),
    linear-gradient(90deg, rgba(74, 144, 226, 0.04) 1px, transparent 1px),
    /* 圆点网格 */
    radial-gradient(circle, rgba(42, 140, 108, 0.03) 1px, transparent 1px);
  background-size: 80px 80px, 80px 80px, 40px 40px;
  opacity: 0.4;
}

.light-mode .tech-grid-2 {
  background-image:
    linear-gradient(rgba(138, 77, 184, 0.03) 1px, transparent 1px),
    linear-gradient(90deg, rgba(138, 77, 184, 0.03) 1px, transparent 1px),
    /* 六边形网格 */
    repeating-linear-gradient(60deg, transparent, transparent 30px, rgba(74, 144, 226, 0.02) 30px, rgba(74, 144, 226, 0.02) 31px),
    repeating-linear-gradient(-60deg, transparent, transparent 30px, rgba(42, 140, 108, 0.02) 30px, rgba(42, 140, 108, 0.02) 31px);
  background-size: 120px 120px, 120px 120px, 60px 60px, 60px 60px;
  opacity: 0.3;
}

@keyframes light-grid-drift {
  0% {
    background-position:
      0 0,      /* 主网格水平 */
      0 0,      /* 主网格垂直 */
      0 0,      /* 对角线1 */
      0 0,      /* 对角线2 */
      0 0,      /* 细密网格水平 */
      0 0;      /* 细密网格垂直 */
    opacity: 0.6;
  }
  50% {
    background-position:
      30px 30px,
      30px 30px,
      40px 40px,
      40px 40px,
      10px 10px,
      10px 10px;
    opacity: 0.7;
  }
  100% {
    background-position:
      60px 60px,
      60px 60px,
      80px 80px,
      80px 80px,
      20px 20px,
      20px 20px;
    opacity: 0.6;
  }
}

/* 网格脉动效果 */
@keyframes grid-pulse {
  0%, 100% {
    filter: brightness(1) contrast(1);
  }
  50% {
    filter: brightness(1.05) contrast(1.1);
  }
}

/* 浅色模式额外动态效果 */
.light-mode .grid-sparkles {
  position: absolute;
  inset: 0;
  pointer-events: none;
  z-index: 1;
}

.light-mode .grid-sparkle {
  position: absolute;
  width: 2px;
  height: 2px;
  border-radius: 50%;
  background: var(--neural-primary);
  opacity: 0;
  animation: sparkle-twinkle 3s ease-in-out infinite;
  box-shadow: 0 0 6px var(--neural-glow);
}

.light-mode .grid-sparkle-1 { top: 15%; left: 20%; animation-delay: 0s; }
.light-mode .grid-sparkle-2 { top: 30%; left: 60%; animation-delay: 0.5s; }
.light-mode .grid-sparkle-3 { top: 50%; left: 10%; animation-delay: 1s; }
.light-mode .grid-sparkle-4 { top: 70%; left: 40%; animation-delay: 1.5s; }
.light-mode .grid-sparkle-5 { top: 20%; left: 80%; animation-delay: 2s; }
.light-mode .grid-sparkle-6 { top: 60%; left: 70%; animation-delay: 2.5s; }
.light-mode .grid-sparkle-7 { top: 40%; left: 30%; animation-delay: 3s; }
.light-mode .grid-sparkle-8 { top: 80%; left: 50%; animation-delay: 3.5s; }

@keyframes sparkle-twinkle {
  0%, 100% {
    opacity: 0;
    transform: scale(0.5);
  }
  50% {
    opacity: 0.8;
    transform: scale(1.2);
  }
}

.tech-grid-1 {
  animation: grid-drift-1 40s linear infinite;
}

.tech-grid-2 {
  background-size: 120px 120px;
  opacity: 0.15;
  animation: grid-drift-2 60s linear infinite reverse;
}

@keyframes grid-drift-1 {
  0% {
    background-position: 0 0;
  }
  100% {
    background-position: 60px 60px;
  }
}

@keyframes grid-drift-2 {
  0% {
    background-position: 0 0;
  }
  100% {
    background-position: 120px 120px;
  }
}

/* 数据流层 */
.data-flow-layer {
  position: absolute;
  inset: 0;
}

.data-flow {
  position: absolute;
  height: 2px;
  background: linear-gradient(90deg, transparent, var(--neural-primary), transparent);
  filter: blur(1px);
  opacity: 0;
  animation: data-flow-move 8s linear infinite;
}

.data-flow::before {
  content: '';
  position: absolute;
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--neural-primary);
  top: -2px;
  left: 0;
  box-shadow: 0 0 12px var(--neural-glow);
  animation: data-particle-move 8s linear infinite;
}

.light-mode .data-flow {
  filter: blur(0.3px);
  opacity: 0.9;
  background: linear-gradient(90deg, transparent, var(--neural-primary), var(--neural-secondary), transparent);
}

.light-mode .data-flow::before {
  box-shadow: 0 0 15px var(--neural-glow), 0 0 25px rgba(255, 255, 255, 0.3);
  background: var(--neural-secondary);
}

.data-flow-1 {
  top: 20%; left: 5%; width: 30%; transform: rotate(15deg);
  animation-delay: 0s;
}
.data-flow-2 {
  top: 40%; left: 60%; width: 25%; transform: rotate(-10deg);
  animation-delay: 2s;
}
.data-flow-3 {
  top: 70%; left: 20%; width: 35%; transform: rotate(5deg);
  animation-delay: 4s;
}
.data-flow-4 {
  top: 30%; left: 70%; width: 20%; transform: rotate(-20deg);
  animation-delay: 6s;
}

@keyframes data-flow-move {
  0% { opacity: 0; }
  10% { opacity: 0.8; }
  90% { opacity: 0.8; }
  100% { opacity: 0; }
}

@keyframes data-particle-move {
  0% { left: 0; opacity: 0; }
  10% { opacity: 1; }
  90% { opacity: 1; }
  100% { left: 100%; opacity: 0; }
}

/* 光晕效果 */
.tech-glow {
  position: absolute;
  border-radius: 50%;
  background: radial-gradient(circle, var(--neural-primary), transparent);
  filter: blur(60px);
  opacity: 0.1;
  animation: glow-pulse 15s ease-in-out infinite alternate;
}

.light-mode .tech-glow {
  opacity: 0.15;
  filter: blur(70px);
  background: radial-gradient(circle, var(--neural-primary), var(--neural-secondary), transparent);
  mix-blend-mode: screen;
}

.light-mode .tech-glow-1 {
  background: radial-gradient(circle, var(--neural-primary), transparent);
}

.light-mode .tech-glow-2 {
  background: radial-gradient(circle, var(--neural-secondary), transparent);
}

.light-mode .tech-glow-3 {
  background: radial-gradient(circle, var(--neural-accent), transparent);
}

.tech-glow-1 {
  width: 400px; height: 400px;
  top: 10%; left: 10%;
  animation-delay: 0s;
}
.tech-glow-2 {
  width: 300px; height: 300px;
  top: 60%; left: 70%;
  animation-delay: 5s;
}
.tech-glow-3 {
  width: 500px; height: 500px;
  top: 70%; left: 20%;
  animation-delay: 10s;
}

@keyframes glow-pulse {
  0% {
    transform: translate(0, 0) scale(1);
    opacity: 0.08;
  }
  100% {
    transform: translate(100px, -50px) scale(1.3);
    opacity: 0.15;
  }
}



/* 粒子系统 */
.particles-container {
  position: absolute;
  inset: 0;
}

.particle {
  position: absolute;
  border-radius: 50%;
  background: var(--neural-primary);
  filter: blur(1px);
  animation: particle-float 10s ease-in-out infinite;
}

.light-mode .particle {
  filter: blur(0.3px);
  opacity: 0.8;
  box-shadow: 0 0 10px var(--neural-glow), 0 0 20px rgba(255, 255, 255, 0.3);
}

.light-mode .particle-1,
.light-mode .particle-4,
.light-mode .particle-7,
.light-mode .particle-10 {
  background: var(--neural-primary);
}

.light-mode .particle-2,
.light-mode .particle-5,
.light-mode .particle-8 {
  background: var(--neural-secondary);
}

.light-mode .particle-3,
.light-mode .particle-6,
.light-mode .particle-9 {
  background: var(--neural-accent);
}

.particle-1 { width: 3px; height: 3px; top: 15%; left: 20%; animation-delay: 0s; }
.particle-2 { width: 2px; height: 2px; top: 30%; left: 60%; animation-delay: 1s; }
.particle-3 { width: 4px; height: 4px; top: 50%; left: 10%; animation-delay: 2s; }
.particle-4 { width: 3px; height: 3px; top: 70%; left: 40%; animation-delay: 3s; }
.particle-5 { width: 2px; height: 2px; top: 20%; left: 80%; animation-delay: 4s; }
.particle-6 { width: 3px; height: 3px; top: 60%; left: 70%; animation-delay: 5s; }
.particle-7 { width: 4px; height: 4px; top: 40%; left: 30%; animation-delay: 6s; }
.particle-8 { width: 2px; height: 2px; top: 80%; left: 50%; animation-delay: 7s; }
.particle-9 { width: 3px; height: 3px; top: 25%; left: 40%; animation-delay: 8s; }
.particle-10 { width: 2px; height: 2px; top: 65%; left: 90%; animation-delay: 9s; }

@keyframes particle-float {
  0%, 100% {
    transform: translate(0, 0) scale(1);
    opacity: 0.3;
  }
  50% {
    transform: translate(20px, -20px) scale(1.5);
    opacity: 0.8;
  }
}

/* 全息投影效果 */
.hologram {
  position: absolute;
  border: 1px solid rgba(0, 255, 157, 0.2);
  border-radius: 50%;
  animation: hologram-pulse 6s ease-in-out infinite;
  opacity: 0;
}

.light-mode .hologram {
  border-color: rgba(42, 140, 108, 0.4);
  box-shadow: 0 0 40px rgba(42, 140, 108, 0.2), inset 0 0 20px rgba(255, 255, 255, 0.1);
}

.light-mode .hologram-1 {
  border-color: rgba(42, 140, 108, 0.4);
}

.light-mode .hologram-2 {
  border-color: rgba(74, 144, 226, 0.4);
}

.hologram-1 {
  width: 200px; height: 200px;
  top: 30%; left: 30%;
  animation-delay: 0s;
}
.hologram-2 {
  width: 150px; height: 150px;
  top: 60%; left: 60%;
  animation-delay: 3s;
}

@keyframes hologram-pulse {
  0%, 100% {
    transform: scale(0.5);
    opacity: 0;
  }
  50% {
    transform: scale(1.2);
    opacity: 0.3;
  }
}

/* 增强的神经元网络 */
.enhanced-neural-network {
  position: absolute;
  inset: 0;
}

.neural-node {
  position: absolute;
  width: 16px;
  height: 16px;
  border-radius: 50%;
  background: radial-gradient(circle at 30% 30%, var(--neural-primary), var(--neural-accent));
  box-shadow:
    0 0 25px var(--neural-glow),
    inset 0 0 10px rgba(255, 255, 255, 0.2);
  animation: neural-node-pulse 4s ease-in-out infinite;
}

.light-mode .neural-node {
  box-shadow:
    0 0 30px var(--neural-glow),
    inset 0 0 12px rgba(255, 255, 255, 0.5);
}

.neural-node-1 { top: 25%; left: 15%; animation-delay: 0s; }
.neural-node-2 { top: 35%; left: 40%; animation-delay: 0.5s; }
.neural-node-3 { top: 20%; left: 65%; animation-delay: 1s; }
.neural-node-4 { top: 50%; left: 25%; animation-delay: 1.5s; }
.neural-node-5 { top: 45%; left: 75%; animation-delay: 2s; }
.neural-node-6 { top: 70%; left: 50%; animation-delay: 2.5s; }

@keyframes neural-node-pulse {
  0%, 100% {
    transform: scale(1);
    filter: brightness(1);
  }
  50% {
    transform: scale(1.4);
    filter: brightness(1.5);
  }
}

.neural-connection {
  position: absolute;
  height: 1px;
  background: linear-gradient(90deg, transparent, var(--neural-primary), transparent);
  filter: blur(0.5px);
  opacity: 0.4;
  animation: connection-glow 5s ease-in-out infinite;
}

.light-mode .neural-connection {
  filter: blur(0.3px);
  opacity: 0.5;
}

.neural-conn-1 {
  top: 27%; left: 17%; width: 25%; transform: rotate(20deg);
  animation-delay: 0s;
}
.neural-conn-2 {
  top: 22%; left: 42%; width: 20%; transform: rotate(-15deg);
  animation-delay: 1s;
}
.neural-conn-3 {
  top: 48%; left: 27%; width: 30%; transform: rotate(10deg);
  animation-delay: 2s;
}
.neural-conn-4 {
  top: 47%; left: 40%; width: 35%; transform: rotate(-5deg);
  animation-delay: 3s;
}

@keyframes connection-glow {
  0%, 100% {
    opacity: 0.2;
    filter: blur(0.5px) brightness(1);
  }
  50% {
    opacity: 0.6;
    filter: blur(1px) brightness(1.5);
  }
}

/* DNA双螺旋结构 */
.dna-helix {
  position: absolute;
  width: 2px;
  background: linear-gradient(to bottom, transparent, var(--neural-primary), transparent);
  opacity: 0.4;
  filter: blur(1px);
}

.dna-helix::before,
.dna-helix::after {
  content: '';
  position: absolute;
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: var(--neural-primary);
  box-shadow: 0 0 12px var(--neural-glow);
  animation: dna-node-pulse 2s ease-in-out infinite;
}

.dna-helix::before {
  top: 0;
  left: -3px;
}

.dna-helix::after {
  bottom: 0;
  left: -3px;
}

.dna-helix-1 {
  left: 15%;
  height: 120%;
  top: -10%;
  transform: rotate(15deg);
  animation: dna-drift-1 25s linear infinite;
}

.dna-helix-2 {
  left: 50%;
  height: 140%;
  top: -20%;
  transform: rotate(-10deg);
  animation: dna-drift-2 30s linear infinite reverse;
}

.dna-helix-3 {
  left: 85%;
  height: 110%;
  top: -5%;
  transform: rotate(5deg);
  animation: dna-drift-3 35s linear infinite;
}

@keyframes dna-drift-1 {
  0% { transform: rotate(15deg) translateY(0); }
  100% { transform: rotate(15deg) translateY(100px); }
}

@keyframes dna-drift-2 {
  0% { transform: rotate(-10deg) translateY(0); }
  100% { transform: rotate(-10deg) translateY(-80px); }
}

@keyframes dna-drift-3 {
  0% { transform: rotate(5deg) translateY(0); }
  100% { transform: rotate(5deg) translateY(120px); }
}

@keyframes dna-node-pulse {
  0%, 100% { opacity: 0.3; transform: scale(0.8); }
  50% { opacity: 0.8; transform: scale(1.2); }
}

/* 神经元节点 */
.neuron-node {
  position: absolute;
  width: 12px;
  height: 12px;
  border-radius: 50%;
  background: radial-gradient(circle at 30% 30%, var(--neural-primary), var(--neural-secondary));
  box-shadow: 0 0 20px var(--neural-glow);
  filter: blur(0.5px);
  animation: neuron-pulse 3s ease-in-out infinite;
}

.neuron-node-1 { top: 20%; left: 10%; animation-delay: 0s; }
.neuron-node-2 { top: 30%; left: 25%; animation-delay: 0.5s; }
.neuron-node-3 { top: 15%; left: 40%; animation-delay: 1s; }
.neuron-node-4 { top: 40%; left: 15%; animation-delay: 1.5s; }
.neuron-node-5 { top: 25%; left: 60%; animation-delay: 2s; }
.neuron-node-6 { top: 50%; left: 35%; animation-delay: 2.5s; }
.neuron-node-7 { top: 35%; left: 75%; animation-delay: 3s; }
.neuron-node-8 { top: 60%; left: 50%; animation-delay: 3.5s; }
.neuron-node-9 { top: 45%; left: 85%; animation-delay: 4s; }
.neuron-node-10 { top: 70%; left: 65%; animation-delay: 4.5s; }

@keyframes neuron-pulse {
  0%, 100% { transform: scale(1); opacity: 0.6; }
  50% { transform: scale(1.3); opacity: 1; }
}

/* 浅色模式神经元节点增强 */
.light-mode .neuron-node {
  box-shadow: 0 0 25px var(--neural-glow), 0 0 40px rgba(255, 255, 255, 0.3);
  filter: blur(0.3px);
  opacity: 0.8;
}

.light-mode .neuron-node::after {
  content: '';
  position: absolute;
  inset: -4px;
  border-radius: 50%;
  border: 1px solid rgba(255, 255, 255, 0.4);
  animation: neuron-halo 3s ease-in-out infinite;
}

@keyframes neuron-halo {
  0%, 100% { transform: scale(1); opacity: 0.3; }
  50% { transform: scale(1.2); opacity: 0.6; }
}

/* 神经连接线 */
.neural-connection {
  position: absolute;
  height: 2px;
  background: linear-gradient(90deg, transparent, var(--neural-primary), transparent);
  border-radius: 1px;
  filter: blur(1px);
  opacity: 0.4;
  animation: connection-flow 4s linear infinite;
}

.neural-connection::before {
  content: '';
  position: absolute;
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--neural-primary);
  top: -2px;
  left: 0;
  animation: pulse-travel 4s linear infinite;
}

.neural-connection-1 {
  top: 25%; left: 12%; width: 15%; transform: rotate(20deg);
  animation-delay: 0s;
}
.neural-connection-2 {
  top: 20%; left: 28%; width: 12%; transform: rotate(-15deg);
  animation-delay: 0.5s;
}
.neural-connection-3 {
  top: 38%; left: 18%; width: 18%; transform: rotate(10deg);
  animation-delay: 1s;
}
.neural-connection-4 {
  top: 30%; left: 45%; width: 14%; transform: rotate(-5deg);
  animation-delay: 1.5s;
}
.neural-connection-5 {
  top: 50%; left: 38%; width: 16%; transform: rotate(15deg);
  animation-delay: 2s;
}
.neural-connection-6 {
  top: 42%; left: 70%; width: 13%; transform: rotate(-10deg);
  animation-delay: 2.5s;
}

@keyframes connection-flow {
  0% { opacity: 0.2; }
  50% { opacity: 0.6; }
  100% { opacity: 0.2; }
}

@keyframes pulse-travel {
  0% { left: 0; opacity: 0; }
  10% { opacity: 1; }
  90% { opacity: 1; }
  100% { left: 100%; opacity: 0; }
}

/* 浅色模式神经连接线增强 */
.light-mode .neural-connection {
  opacity: 0.6;
  filter: blur(0.5px);
  background: linear-gradient(90deg, transparent, var(--neural-primary), var(--neural-secondary), transparent);
}

.light-mode .neural-connection::before {
  box-shadow: 0 0 15px var(--neural-glow), 0 0 25px rgba(255, 255, 255, 0.4);
}

/* 脉冲动画 */
.neural-pulse {
  position: absolute;
  width: 100px;
  height: 100px;
  border-radius: 50%;
  border: 2px solid var(--neural-primary);
  opacity: 0;
  animation: pulse-expand 3s ease-out infinite;
}

.neural-pulse-1 {
  top: 25%; left: 25%;
  animation-delay: 0s;
}
.neural-pulse-2 {
  top: 40%; left: 60%;
  animation-delay: 1s;
}
.neural-pulse-3 {
  top: 60%; left: 40%;
  animation-delay: 2s;
}

@keyframes pulse-expand {
  0% {
    transform: scale(0.1);
    opacity: 0.8;
  }
  100% {
    transform: scale(3);
    opacity: 0;
  }
}

/* 浅色模式脉冲动画增强 */
.light-mode .neural-pulse {
  border: 2px solid var(--neural-primary);
  box-shadow: 0 0 40px var(--neural-glow), inset 0 0 20px rgba(255, 255, 255, 0.2);
}

/* 背景光晕 */
.neural-glow {
  position: absolute;
  border-radius: 50%;
  background: radial-gradient(circle, var(--neural-primary), transparent);
  filter: blur(40px);
  opacity: 0.15;
  animation: glow-drift 20s ease-in-out infinite alternate;
}

.neural-glow-1 {
  width: 300px;
  height: 300px;
  top: 10%;
  left: 10%;
  animation-delay: 0s;
}
.neural-glow-2 {
  width: 400px;
  height: 400px;
  top: 50%;
  left: 60%;
  animation-delay: 5s;
}
.neural-glow-3 {
  width: 250px;
  height: 250px;
  top: 70%;
  left: 20%;
  animation-delay: 10s;
}

@keyframes glow-drift {
  0% {
    transform: translate(0, 0) scale(1);
  }
  100% {
    transform: translate(50px, -30px) scale(1.2);
  }
}

/* AI Agent 数据流效果 */
.data-stream {
  position: absolute;
  height: 1px;
  background: linear-gradient(90deg, transparent, var(--neural-primary), transparent);
  filter: blur(0.5px);
  opacity: 0;
  animation: data-stream-flow 6s linear infinite;
}

.data-stream::before {
  content: '';
  position: absolute;
  width: 4px;
  height: 4px;
  border-radius: 50%;
  background: var(--neural-primary);
  top: -1.5px;
  left: 0;
  box-shadow: 0 0 8px var(--neural-glow);
  animation: data-particle-travel 6s linear infinite;
}

.data-stream-1 {
  top: 30%;
  left: 5%;
  width: 25%;
  transform: rotate(25deg);
  animation-delay: 0s;
}

.data-stream-2 {
  top: 50%;
  left: 70%;
  width: 20%;
  transform: rotate(-15deg);
  animation-delay: 2s;
}

.data-stream-3 {
  top: 70%;
  left: 20%;
  width: 30%;
  transform: rotate(10deg);
  animation-delay: 4s;
}

@keyframes data-stream-flow {
  0% {
    opacity: 0;
  }
  10% {
    opacity: 0.7;
  }
  90% {
    opacity: 0.7;
  }
  100% {
    opacity: 0;
  }
}

@keyframes data-particle-travel {
  0% {
    left: 0;
    opacity: 0;
  }
  10% {
    opacity: 1;
  }
  90% {
    opacity: 1;
  }
  100% {
    left: 100%;
    opacity: 0;
  }
}

.login-bg-img {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  object-fit: cover;
  object-position: center center;
  display: block;
  image-rendering: auto;
  opacity: 0.18;
  mix-blend-mode: screen;
  filter: saturate(0.85) contrast(1.05);
}

.login-page-content {
  position: relative;
  z-index: 1;
  min-height: 100vh;
  min-height: 100dvh;
  display: flex;
  flex-direction: column;
  align-items: stretch;
  padding: 24px 0 0;
  box-sizing: border-box;
}

.login-main {
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 0 16px 24px;
  box-sizing: border-box;
}

.theme-bar {
  position: absolute;
  top: 20px;
  right: 20px;
  display: flex;
  gap: 8px;
  z-index: 2;
}

.theme-orbit-btn {
  --el-button-bg-color: rgba(10, 15, 30, 0.7);
  --el-button-border-color: var(--neural-border);
  --el-button-text-color: var(--neural-primary);
  --el-button-hover-bg-color: rgba(0, 255, 157, 0.15);
  --el-button-hover-border-color: rgba(0, 255, 157, 0.5);
  --el-button-hover-text-color: var(--neural-secondary);
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
  box-shadow:
    0 0 0 1px rgba(0, 255, 157, 0.2),
    0 0 24px rgba(0, 255, 157, 0.15);
}

.light-mode .theme-orbit-btn {
  --el-button-bg-color: rgba(255, 255, 255, 0.85);
  --el-button-border-color: var(--neural-border);
  --el-button-text-color: var(--neural-primary);
  --el-button-hover-bg-color: rgba(42, 140, 108, 0.15);
  --el-button-hover-border-color: rgba(42, 140, 108, 0.5);
  --el-button-hover-text-color: var(--neural-secondary);
  box-shadow:
    0 0 0 1px rgba(42, 140, 108, 0.2),
    0 0 24px rgba(42, 140, 108, 0.15);
}

.panel {
  position: relative;
  width: 100%;
  max-width: 440px;
  padding: 48px 40px 40px;
  border-radius: 16px;
  background: linear-gradient(
    145deg,
    rgba(20, 25, 45, 0.9) 0%,
    rgba(15, 20, 40, 0.95) 50%,
    rgba(10, 15, 35, 0.9) 100%
  );
  backdrop-filter: blur(32px) saturate(1.8);
  -webkit-backdrop-filter: blur(32px) saturate(1.8);
  border: 1px solid rgba(0, 255, 157, 0.3);
  box-shadow:
    0 0 0 1px rgba(255, 255, 255, 0.08) inset,
    0 8px 32px rgba(0, 0, 0, 0.4),
    0 32px 96px rgba(0, 0, 0, 0.8),
    0 0 80px rgba(0, 255, 157, 0.15),
    inset 0 1px 0 rgba(255, 255, 255, 0.1);
  animation: neural-panel-in 1s cubic-bezier(0.22, 1, 0.36, 1) both;
  overflow: hidden;
  transition: all 0.5s ease;
}

.panel::before {
  content: '';
  position: absolute;
  inset: 0;
  background: linear-gradient(
    135deg,
    transparent 0%,
    rgba(0, 255, 157, 0.05) 30%,
    rgba(0, 212, 255, 0.03) 70%,
    transparent 100%
  );
  pointer-events: none;
  z-index: 1;
  transition: all 0.5s ease;
}

.light-mode .panel {
  background: linear-gradient(
    145deg,
    rgba(255, 255, 255, 0.98) 0%,
    rgba(250, 252, 255, 0.99) 50%,
    rgba(245, 250, 255, 0.98) 100%
  );
  border: 1px solid rgba(0, 90, 60, 0.4);
  box-shadow:
    0 0 0 1px rgba(255, 255, 255, 0.95) inset,
    0 8px 32px rgba(0, 0, 0, 0.15),
    0 32px 96px rgba(0, 0, 0, 0.2),
    0 0 80px rgba(0, 90, 60, 0.15),
    inset 0 1px 0 rgba(255, 255, 255, 0.9);
}

.light-mode .panel::before {
  background: linear-gradient(
    135deg,
    transparent 0%,
    rgba(0, 140, 90, 0.03) 30%,
    rgba(0, 136, 204, 0.02) 70%,
    transparent 100%
  );
}

@keyframes neural-panel-in {
  from {
    opacity: 0;
    transform: translateY(16px) scale(0.98);
    filter: blur(6px);
  }
  to {
    opacity: 1;
    transform: translateY(0) scale(1);
    filter: blur(0);
  }
}

.panel-glow-ring {
  position: absolute;
  inset: -2px;
  border-radius: 18px;
  pointer-events: none;
  background: conic-gradient(
    from 0deg at 50% 50%,
    rgba(0, 255, 157, 0.8) 0deg,
    rgba(0, 212, 255, 0.6) 120deg,
    rgba(157, 78, 221, 0.4) 240deg,
    rgba(0, 255, 157, 0.8) 360deg
  );
  opacity: 0.5;
  z-index: 0;
  mask:
    linear-gradient(#fff 0 0) content-box,
    linear-gradient(#fff 0 0);
  mask-composite: xor;
  -webkit-mask-composite: xor;
  padding: 2px;
  animation: neural-ring-pulse 4s ease-in-out infinite, neural-ring-rotate 20s linear infinite;
  filter: blur(1px);
}

@keyframes neural-ring-pulse {
  0%, 100% {
    opacity: 0.4;
    filter: blur(1px) brightness(1);
  }
  50% {
    opacity: 0.7;
    filter: blur(1.5px) brightness(1.3);
  }
}

@keyframes neural-ring-rotate {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}

.panel-edge {
  position: absolute;
  width: 24px;
  height: 24px;
  border-color: var(--neural-primary);
  border-style: solid;
  opacity: 0.9;
  z-index: 2;
  pointer-events: none;
  filter: drop-shadow(0 0 8px var(--neural-glow));
}
.panel-edge--tl {
  top: 16px;
  left: 16px;
  border-width: 3px 0 0 3px;
  border-radius: 4px 0 0 0;
}
.panel-edge--tr {
  top: 16px;
  right: 16px;
  border-width: 3px 3px 0 0;
  border-radius: 0 4px 0 0;
}
.panel-edge--bl {
  bottom: 16px;
  left: 16px;
  border-width: 0 0 3px 3px;
  border-radius: 0 0 0 4px;
}
.panel-edge--br {
  bottom: 16px;
  right: 16px;
  border-width: 0 3px 3px 0;
  border-radius: 0 0 4px 0;
}

.panel-edge::after {
  content: '';
  position: absolute;
  width: 6px;
  height: 6px;
  background: var(--neural-primary);
  border-radius: 50%;
  animation: corner-dot-pulse 2s ease-in-out infinite;
}

.panel-edge--tl::after {
  top: -3px;
  left: -3px;
}
.panel-edge--tr::after {
  top: -3px;
  right: -3px;
}
.panel-edge--bl::after {
  bottom: -3px;
  left: -3px;
}
.panel-edge--br::after {
  bottom: -3px;
  right: -3px;
}

@keyframes corner-dot-pulse {
  0%, 100% {
    opacity: 0.6;
    transform: scale(1);
  }
  50% {
    opacity: 1;
    transform: scale(1.3);
  }
}

.login-brand,
.sub,
.demo-hint,
.form,
.hint-alert,
.err-alert {
  position: relative;
  z-index: 1;
}

.login-brand {
  margin-bottom: 12px;
  position: relative;
  z-index: 2;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  width: 100%;
  cursor: pointer;
  padding: 8px 0;
  border-radius: 12px;
  transition: background 0.3s ease;
  outline: none;
}

.login-brand:hover {
  background: rgba(0, 255, 157, 0.04);
}

.light-mode .login-brand:hover {
  background: rgba(42, 140, 108, 0.06);
}

.login-logo-wrap {
  position: relative;
  width: 88px;
  height: 88px;
  margin: 0 auto 20px;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  border-radius: 22px;
  transition: all 0.4s cubic-bezier(0.22, 1, 0.36, 1);
}

.login-logo-wrap:hover {
  transform: scale(1.08);
  box-shadow: 0 0 30px rgba(0, 255, 157, 0.25);
}

.light-mode .login-logo-wrap:hover {
  box-shadow: 0 0 30px rgba(42, 140, 108, 0.2);
}

.logo-orbit {
  position: absolute;
  inset: -12px;
  border-radius: 50%;
  border: 2px dashed rgba(0, 255, 157, 0.4);
  animation: neural-orbit-spin 10s linear infinite;
  filter: drop-shadow(0 0 8px rgba(0, 255, 157, 0.3));
}

.logo-orbit::before {
  content: '';
  position: absolute;
  inset: -4px;
  border-radius: 50%;
  border: 1px solid rgba(0, 212, 255, 0.2);
  animation: neural-orbit-spin 15s linear infinite reverse;
}

.logo-orbit::after {
  content: '';
  position: absolute;
  width: 10px;
  height: 10px;
  top: 0;
  left: 50%;
  margin-left: -5px;
  border-radius: 50%;
  background: var(--neural-primary);
  box-shadow: 0 0 20px var(--neural-glow);
  animation: orbit-node-pulse 3s ease-in-out infinite, orbit-node-rotate 10s linear infinite;
}

@keyframes neural-orbit-spin {
  to {
    transform: rotate(360deg);
  }
}

@keyframes orbit-node-pulse {
  0%, 100% {
    transform: scale(1) rotate(0deg);
    opacity: 0.7;
  }
  50% {
    transform: scale(1.4) rotate(180deg);
    opacity: 1;
  }
}

@keyframes orbit-node-rotate {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(-360deg);
  }
}

.login-logo-img {
  width: 100%;
  height: 100%;
  object-fit: contain;
  display: block;
  filter: drop-shadow(0 0 12px rgba(46, 230, 214, 0.25));
}

/* 自定义 logo — 圆角矩阵框 */
.custom-logo-img {
  width: 72px;
  height: 72px;
  object-fit: contain;
  border-radius: 16px;
  background: rgba(10, 10, 30, 0.8);
  padding: 8px;
  box-shadow:
    0 0 20px rgba(0, 255, 157, 0.15),
    inset 0 0 10px rgba(0, 255, 157, 0.1);
  border: 1px solid rgba(0, 255, 157, 0.25);
}

/* 默认神经元Logo — 细胞体 + 树突 + 轴突 */
.brand-mark {
  width: 72px;
  height: 72px;
  position: relative;
  animation: neuron-logo-float 6s ease-in-out infinite;
}

/* 神经元细胞体 */
.brand-mark::before {
  content: '';
  position: absolute;
  width: 48px;
  height: 48px;
  top: 12px;
  left: 12px;
  border-radius: 50%;
  background: radial-gradient(
    circle at 30% 30%,
    var(--neural-primary) 0%,
    var(--neural-accent) 50%,
    transparent 70%
  );
  box-shadow:
    0 0 40px var(--neural-glow),
    inset 0 0 20px rgba(255, 255, 255, 0.3);
  animation: neuron-core-pulse 3s ease-in-out infinite;
}

/* 树突（输入分支） */
.brand-mark::after {
  content: '';
  position: absolute;
  width: 60px;
  height: 60px;
  top: 6px;
  left: 6px;
  border-radius: 50%;
  border: 2px solid rgba(0, 255, 157, 0.4);
  clip-path: polygon(
    50% 0%, 60% 20%, 80% 30%, 70% 50%,
    80% 70%, 60% 80%, 50% 100%,
    40% 80%, 20% 70%, 30% 50%,
    20% 30%, 40% 20%
  );
  animation: dendrites-rotate 20s linear infinite;
}

/* 轴突（输出分支） */
.neuron-axon {
  position: absolute;
  width: 40px;
  height: 4px;
  background: linear-gradient(90deg, var(--neural-primary), transparent);
  top: 50%;
  left: 70%;
  border-radius: 2px;
  transform: translateY(-50%);
  animation: axon-pulse 2s ease-in-out infinite;
}

.neuron-axon::before {
  content: '';
  position: absolute;
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: var(--neural-primary);
  right: -4px;
  top: -2px;
  box-shadow: 0 0 12px var(--neural-glow);
  animation: axon-terminal-pulse 1.5s ease-in-out infinite;
}

/* 突触连接点 */
.neuron-synapse {
  position: absolute;
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--neural-secondary);
  box-shadow: 0 0 8px var(--neural-glow);
  animation: synapse-pulse 2s ease-in-out infinite;
}

.neuron-synapse-1 { top: 20%; left: 15%; animation-delay: 0s; }
.neuron-synapse-2 { top: 15%; left: 50%; animation-delay: 0.3s; }
.neuron-synapse-3 { top: 30%; left: 80%; animation-delay: 0.6s; }
.neuron-synapse-4 { top: 60%; left: 20%; animation-delay: 0.9s; }
.neuron-synapse-5 { top: 70%; left: 60%; animation-delay: 1.2s; }

@keyframes neuron-logo-float {
  0%, 100% { transform: translateY(0) rotate(0deg); }
  33% { transform: translateY(-4px) rotate(2deg); }
  66% { transform: translateY(4px) rotate(-2deg); }
}

@keyframes neuron-core-pulse {
  0%, 100% { transform: scale(1); filter: brightness(1) blur(0); }
  50% { transform: scale(1.1); filter: brightness(1.3) blur(1px); }
}

@keyframes dendrites-rotate {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

@keyframes axon-pulse {
  0%, 100% { opacity: 0.7; width: 40px; }
  50% { opacity: 1; width: 44px; }
}

@keyframes axon-terminal-pulse {
  0%, 100% { transform: scale(1); opacity: 0.8; }
  50% { transform: scale(1.3); opacity: 1; }
}

@keyframes synapse-pulse {
  0%, 100% { transform: scale(1); opacity: 0.6; }
  50% { transform: scale(1.4); opacity: 1; }
}

.eyebrow {
  margin: 0 auto 8px;
  text-align: center;
  font-family: 'Exo 2', sans-serif;
  font-size: 12px;
  font-weight: 700;
  letter-spacing: 0.4em;
  color: var(--neural-primary);
  text-shadow: 0 0 24px var(--neural-glow);
  animation: eyebrow-pulse 3s ease-in-out infinite;
  position: relative;
  display: block;
  width: fit-content;
  padding: 6px 20px;
  background: rgba(0, 255, 157, 0.08);
  border-radius: 6px;
  border: 1px solid rgba(0, 255, 157, 0.2);
  box-sizing: border-box;
}

@keyframes eyebrow-pulse {
  0%, 100% {
    opacity: 0.9;
    text-shadow: 0 0 24px var(--neural-glow);
  }
  50% {
    opacity: 1;
    text-shadow: 0 0 32px var(--neural-glow);
  }
}

.title {
  margin: 12px 0 0;
  text-align: center;
  white-space: pre-wrap;
  font-family: 'Exo 2', sans-serif;
  font-size: 32px;
  font-weight: 800;
  letter-spacing: 0.06em;
  line-height: 1.1;
  background: linear-gradient(
    135deg,
    #ffffff 0%,
    var(--neural-secondary) 25%,
    var(--neural-primary) 50%,
    var(--neural-accent) 75%,
    #ffffff 100%
  );
  -webkit-background-clip: text;
  background-clip: text;
  color: transparent;
  background-size: 200% 200%;
  animation: title-glow 4s ease-in-out infinite, title-gradient-shift 8s ease-in-out infinite;
  position: relative;
  padding-bottom: 8px;
}

.title::after {
  content: '';
  position: absolute;
  bottom: 0;
  left: 25%;
  right: 25%;
  height: 2px;
  background: linear-gradient(90deg, transparent, var(--neural-primary), transparent);
  border-radius: 1px;
  animation: title-underline 3s ease-in-out infinite;
}

@keyframes title-glow {
  0%, 100% {
    filter: drop-shadow(0 0 40px rgba(0, 255, 157, 0.4));
  }
  50% {
    filter: drop-shadow(0 0 60px rgba(0, 255, 157, 0.6));
  }
}

@keyframes title-gradient-shift {
  0%, 100% {
    background-position: 0% 50%;
  }
  50% {
    background-position: 100% 50%;
  }
}

@keyframes title-underline {
  0%, 100% {
    opacity: 0.5;
    width: 50%;
    left: 25%;
  }
  50% {
    opacity: 0.8;
    width: 60%;
    left: 20%;
  }
}

.sub {
  margin: 10px 0 8px;
  text-align: center;
  font-size: 14px;
  color: var(--neural-fg-muted);
  letter-spacing: 0.02em;
  animation: sub-fade 5s ease-in-out infinite;
}

@keyframes sub-fade {
  0%, 100% { opacity: 0.8; }
  50% { opacity: 1; }
}

.demo-hint {
  margin: 0 0 22px;
  text-align: center;
  font-size: 12px;
  line-height: 1.55;
  color: var(--neural-fg-muted);
  padding: 12px 14px;
  border-radius: 4px;
  background: rgba(0, 255, 157, 0.08);
  border: 1px solid rgba(0, 255, 157, 0.15);
  animation: demo-hint-pulse 6s ease-in-out infinite;
}

.demo-hint .mono {
  font-family: 'IBM Plex Mono', ui-monospace, monospace;
  font-variant-numeric: tabular-nums;
  letter-spacing: 0.04em;
  color: var(--neural-secondary);
}

@keyframes demo-hint-pulse {
  0%, 100% { border-color: rgba(0, 255, 157, 0.15); }
  50% { border-color: rgba(0, 255, 157, 0.25); }
}

.form :deep(.el-form-item__label) {
  font-weight: 500;
  font-size: 13px;
  letter-spacing: 0.06em;
  color: rgba(232, 244, 255, 0.78);
}

.input-round :deep(.el-input__wrapper),
.input-round :deep(.el-select__wrapper) {
  border-radius: 8px;
  background: linear-gradient(
    145deg,
    rgba(10, 20, 40, 0.8) 0%,
    rgba(5, 15, 35, 0.9) 100%
  ) !important;
  box-shadow:
    0 0 0 1px rgba(0, 255, 157, 0.25) inset,
    0 4px 16px rgba(0, 0, 0, 0.4) inset,
    0 2px 8px rgba(0, 0, 0, 0.2) !important;
  transition:
    box-shadow 0.4s cubic-bezier(0.22, 1, 0.36, 1),
    transform 0.3s cubic-bezier(0.22, 1, 0.36, 1),
    border-color 0.4s ease;
  position: relative;
  overflow: hidden;
  user-select: text;
  -webkit-user-select: text;
}

/* 浅色模式输入框 */
.light-mode .input-round :deep(.el-input__wrapper),
.light-mode .input-round :deep(.el-select__wrapper) {
  background: linear-gradient(
    145deg,
    rgba(255, 255, 255, 0.95) 0%,
    rgba(245, 250, 255, 0.98) 100%
  ) !important;
  box-shadow:
    0 0 0 1px rgba(0, 90, 60, 0.3) inset,
    0 4px 16px rgba(0, 0, 0, 0.1) inset,
    0 2px 8px rgba(0, 0, 0, 0.08) !important;
}

.light-mode .input-round :deep(.el-input__wrapper:hover),
.light-mode .input-round :deep(.el-select__wrapper:hover) {
  box-shadow:
    0 0 0 1px rgba(0, 90, 60, 0.5) inset,
    0 6px 24px rgba(0, 90, 60, 0.1) inset,
    0 4px 12px rgba(0, 0, 0, 0.15) !important;
}

.light-mode .input-round :deep(.el-input__wrapper.is-focus),
.light-mode .input-round :deep(.el-select__wrapper.is-focused) {
  box-shadow:
    0 0 0 2px rgba(0, 90, 60, 0.7) inset,
    0 8px 32px rgba(0, 90, 60, 0.2) inset,
    0 0 40px rgba(0, 90, 60, 0.15) !important;
  animation: input-focus-pulse-light 2s ease-in-out infinite;
}

@keyframes input-focus-pulse-light {
  0%, 100% {
    box-shadow:
      0 0 0 2px rgba(0, 90, 60, 0.7) inset,
      0 8px 32px rgba(0, 90, 60, 0.2) inset,
      0 0 40px rgba(0, 90, 60, 0.15);
  }
  50% {
    box-shadow:
      0 0 0 2px rgba(0, 90, 60, 0.8) inset,
      0 8px 32px rgba(0, 90, 60, 0.3) inset,
      0 0 50px rgba(0, 90, 60, 0.25);
  }
}

.input-round :deep(.el-input__wrapper::before),
.input-round :deep(.el-select__wrapper::before) {
  content: '';
  position: absolute;
  inset: 0;
  background: linear-gradient(135deg, transparent, rgba(0, 255, 157, 0.1), transparent);
  opacity: 0;
  transition: opacity 0.4s ease;
  pointer-events: none;
}

.input-round :deep(.el-input__wrapper:hover),
.input-round :deep(.el-select__wrapper:hover) {
  box-shadow:
    0 0 0 1px rgba(0, 255, 157, 0.4) inset,
    0 6px 24px rgba(0, 255, 157, 0.15) inset,
    0 4px 12px rgba(0, 0, 0, 0.3) !important;
  transform: translateY(-2px) scale(1.01);
}

.input-round :deep(.el-input__wrapper:hover::before),
.input-round :deep(.el-select__wrapper:hover::before) {
  opacity: 1;
}

.input-round :deep(.el-input__wrapper.is-focus),
.input-round :deep(.el-select__wrapper.is-focused) {
  box-shadow:
    0 0 0 2px rgba(0, 255, 157, 0.7) inset,
    0 8px 32px rgba(0, 255, 157, 0.25) inset,
    0 0 40px rgba(0, 255, 157, 0.15) !important;
  transform: translateY(-3px) scale(1.02);
  animation: input-focus-pulse 2s ease-in-out infinite;
}

@keyframes input-focus-pulse {
  0%, 100% {
    box-shadow:
      0 0 0 2px rgba(0, 255, 157, 0.7) inset,
      0 8px 32px rgba(0, 255, 157, 0.25) inset,
      0 0 40px rgba(0, 255, 157, 0.15);
  }
  50% {
    box-shadow:
      0 0 0 2px rgba(0, 255, 157, 0.8) inset,
      0 8px 32px rgba(0, 255, 157, 0.35) inset,
      0 0 50px rgba(0, 255, 157, 0.25);
  }
}

.input-round :deep(.el-input__inner),
.input-round :deep(.el-select__placeholder),
.input-round :deep(.el-select__selected-item) {
  color: var(--neural-fg) !important;
  font-family: 'IBM Plex Sans', sans-serif;
  user-select: text;
  -webkit-user-select: text;
}

.input-round :deep(.el-input__inner::placeholder) {
  color: rgba(240, 248, 255, 0.4);
}

.light-mode .input-round :deep(.el-input__inner::placeholder) {
  color: rgba(5, 5, 20, 0.6);
}

/* 增强输入框文字对比度 */
.light-mode .input-round :deep(.el-input__inner),
.light-mode .input-round :deep(.el-select__placeholder),
.light-mode .input-round :deep(.el-select__selected-item) {
  color: var(--neural-fg) !important;
  font-weight: 500;
}

.submit-btn {
  width: 100%;
  margin-top: 12px;
  height: 48px;
  border-radius: 6px;
  font-family: 'Exo 2', sans-serif;
  font-weight: 700;
  font-size: 15px;
  letter-spacing: 0.14em;
  text-indent: 0.14em;
  border: none;
  background: linear-gradient(100deg, #003d33 0%, var(--neural-primary) 45%, var(--neural-secondary) 100%);
  color: #001018 !important;
  box-shadow:
    0 0 0 1px rgba(255, 255, 255, 0.15) inset,
    0 12px 40px rgba(0, 255, 157, 0.3);
  transition:
    transform 0.3s ease,
    box-shadow 0.3s ease,
    filter 0.3s ease;
  position: relative;
  overflow: hidden;
}

.submit-btn::before {
  content: '';
  position: absolute;
  top: 0;
  left: -100%;
  width: 100%;
  height: 100%;
  background: linear-gradient(90deg, transparent, rgba(255, 255, 255, 0.2), transparent);
  animation: btn-shine 3s ease-in-out infinite;
}

.submit-btn:hover:not(:disabled) {
  filter: brightness(1.1);
  box-shadow:
    0 0 0 1px rgba(255, 255, 255, 0.2) inset,
    0 16px 48px rgba(0, 255, 157, 0.4);
  transform: translateY(-2px);
}

.submit-btn:disabled {
  background: linear-gradient(100deg, #333, #666) !important;
  cursor: not-allowed;
  opacity: 0.7;
}

@keyframes btn-shine {
  0% { left: -100%; }
  100% { left: 100%; }
}

.hint-alert,
.err-alert {
  margin-bottom: 16px;
  border-radius: 2px;
  --el-alert-border-radius-base: 2px;
}

:deep(.hint-alert.el-alert) {
  background: rgba(255, 186, 60, 0.1) !important;
  border: 1px solid rgba(255, 186, 60, 0.35);
}

:deep(.err-alert.el-alert) {
  background: rgba(255, 80, 100, 0.1) !important;
  border: 1px solid rgba(255, 100, 120, 0.4);
}

.tenant-pick-hint {
  margin: 0 0 12px;
  font-size: 14px;
  line-height: 1.5;
}

.tenant-pick-group {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 10px;
}

.tenant-pick-row {
  margin-right: 0;
  height: auto;
  white-space: normal;
  line-height: 1.4;
}

.tenant-select {
  width: 100%;
}

.cap {
  display: flex;
  gap: 10px;
  width: 100%;
  align-items: center;
}

.cap-input {
  flex: 1;
}

.cap-img {
  height: 40px;
  border-radius: 4px;
  cursor: pointer;
  border: 1px solid rgba(0, 255, 157, 0.3);
  box-shadow: 0 0 20px rgba(0, 255, 157, 0.15);
  transition: all 0.3s ease;
}

.cap-img:hover {
  border-color: rgba(0, 255, 157, 0.5);
  box-shadow: 0 0 24px rgba(0, 255, 157, 0.25);
  transform: scale(1.02);
}

.shell-footer {
  flex-shrink: 0;
  padding: 12px 16px 20px;
  text-align: center;
  font-size: 13px;
  color: rgba(240, 248, 255, 0.5);
  border-top: none;
  background: transparent;
  letter-spacing: 0.04em;
  animation: footer-fade 8s ease-in-out infinite;
}

@keyframes footer-fade {
  0%, 100% { opacity: 0.7; }
  50% { opacity: 1; }
}
</style>

<style>
/* el-dialog 挂载到 body，需非 scoped */
.login-tech-dialog.el-dialog {
  --el-dialog-bg-color: rgba(10, 15, 30, 0.95);
  border: 1px solid rgba(0, 255, 157, 0.25);
  border-radius: 8px;
  box-shadow:
    0 0 0 1px rgba(255, 255, 255, 0.08) inset,
    0 24px 80px rgba(0, 0, 0, 0.7);
  backdrop-filter: blur(20px);
}

.login-tech-dialog .el-dialog__title {
  font-family: 'Exo 2', sans-serif;
  letter-spacing: 0.12em;
  font-size: 15px;
  color: var(--neural-fg);
  background: linear-gradient(90deg, var(--neural-fg), var(--neural-secondary));
  -webkit-background-clip: text;
  background-clip: text;
  color: transparent;
}

.login-tech-dialog .el-dialog__headerbtn .el-dialog__close {
  color: rgba(0, 255, 157, 0.8);
}

.login-tech-dialog .tenant-pick-hint {
  color: rgba(240, 248, 255, 0.7);
}

.login-tech-dialog .el-radio {
  --el-radio-text-color: rgba(240, 248, 255, 0.9);
}

.login-tech-dialog .el-button--default {
  --el-button-bg-color: rgba(10, 20, 40, 0.7);
  --el-button-border-color: rgba(0, 255, 157, 0.3);
  --el-button-text-color: var(--neural-fg);
}

.login-tech-dialog .el-button--primary {
  background: linear-gradient(100deg, #003d33, var(--neural-primary));
  border: none;
  color: #001018;
  font-weight: 700;
}
</style>
