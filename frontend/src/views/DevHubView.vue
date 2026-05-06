<template>
  <div class="dev-hub" :data-theme="theme">
    <!-- 顶部导航栏 -->
    <header class="hub-header">
      <div class="header-left">
        <div class="brand-logo" @click="$router.push('/home')">
          <svg class="logo-svg" viewBox="0 0 48 48" fill="none">
            <rect x="2" y="2" width="44" height="44" rx="12" fill="rgba(0, 245, 212, 0.06)" stroke="rgba(0, 245, 212, 0.15)" stroke-width="1"/>
            <circle cx="14" cy="12" r="2.5" fill="#00f5d4" opacity="0.35"/>
            <circle cx="22" cy="12" r="2.5" fill="#00f5d4" opacity="0.35"/>
            <circle cx="30" cy="12" r="2.5" fill="#00f5d4" opacity="0.35"/>
            <circle cx="38" cy="12" r="2.5" fill="#00f5d4" opacity="0.35"/>
            <circle cx="14" cy="20" r="2.5" fill="#00f5d4" opacity="0.7" filter="url(#hubDotGlow)"/>
            <circle cx="22" cy="20" r="2.5" fill="#00f5d4" opacity="0.35"/>
            <circle cx="30" cy="20" r="2.5" fill="#00f5d4" opacity="0.7" filter="url(#hubDotGlow)"/>
            <circle cx="38" cy="20" r="2.5" fill="#00f5d4" opacity="0.35"/>
            <circle cx="14" cy="28" r="2.5" fill="#00f5d4" opacity="0.35"/>
            <circle cx="22" cy="28" r="2.5" fill="#00f5d4" opacity="0.7" filter="url(#hubDotGlow)"/>
            <circle cx="30" cy="28" r="2.5" fill="#00f5d4" opacity="0.35"/>
            <circle cx="38" cy="28" r="2.5" fill="#00f5d4" opacity="0.7" filter="url(#hubDotGlow)"/>
            <circle cx="14" cy="36" r="2.5" fill="#00f5d4" opacity="0.35"/>
            <circle cx="22" cy="36" r="2.5" fill="#00f5d4" opacity="0.35"/>
            <circle cx="30" cy="36" r="2.5" fill="#00f5d4" opacity="0.35"/>
            <circle cx="38" cy="36" r="2.5" fill="#00f5d4" opacity="0.35"/>
            <line x1="14" y1="20" x2="22" y2="28" stroke="#00f5d4" stroke-width="1" opacity="0.2"/>
            <line x1="22" y1="20" x2="30" y2="28" stroke="#00f5d4" stroke-width="1" opacity="0.2"/>
            <line x1="30" y1="20" x2="38" y2="28" stroke="#00f5d4" stroke-width="1" opacity="0.2"/>
            <line x1="14" y1="28" x2="22" y2="20" stroke="#00f5d4" stroke-width="1" opacity="0.2"/>
            <line x1="22" y1="28" x2="30" y2="20" stroke="#00f5d4" stroke-width="1" opacity="0.2"/>
            <defs>
              <filter id="hubDotGlow" x="-50%" y="-50%" width="200%" height="200%">
                <feGaussianBlur stdDeviation="2" result="blur"/>
                <feMerge>
                  <feMergeNode in="blur"/>
                  <feMergeNode in="SourceGraphic"/>
                </feMerge>
              </filter>
            </defs>
          </svg>
          <span class="brand-name">Ai DevOS</span>
        </div>

        <nav class="header-nav">
          <button class="nav-btn" :class="{ active: activeView === 'showcase' }" @click="activeView = 'showcase'">
            <span class="nav-icon">🧩</span>
            <span>组件展示</span>
          </button>
          <button class="nav-btn" :class="{ active: activeView === 'api' }" @click="activeView = 'api'">
            <span class="nav-icon">📡</span>
            <span>API 文档</span>
          </button>
        </nav>
      </div>

      <div class="header-right">
        <a class="header-btn header-link" href="/docs" target="_blank" rel="noopener" :title="'Swagger UI'">
          <span class="btn-icon">⚡</span>
          <span class="btn-label">Swagger</span>
        </a>
        <button class="header-btn" :title="theme === 'dark' ? '浅色模式' : '深色模式'" @click="toggleTheme">
          <span class="btn-icon">{{ theme === 'dark' ? '☀️' : '🌙' }}</span>
          <span class="btn-label">{{ theme === 'dark' ? '浅色模式' : '深色模式' }}</span>
        </button>
        <button class="header-btn" title="全屏" @click="toggleFullscreen">
          <span class="btn-icon">⛶</span>
        </button>
      </div>
    </header>

    <!-- 组件展示视图 -->
    <ShowcaseContent v-if="activeView === 'showcase'" :theme="theme" />

    <!-- API 文档视图 -->
    <ApiDocsContent v-if="activeView === 'api'" :theme="theme" />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import ShowcaseContent from './components/ShowcaseContent.vue'
import ApiDocsContent from './components/ApiDocsContent.vue'

const activeView = ref<'showcase' | 'api'>('showcase')

// --- Theme ---
const theme = ref<'dark' | 'light'>('dark')
const THEME_KEY = 'devos-theme'
try {
  const saved = localStorage.getItem(THEME_KEY)
  if (saved === 'light' || saved === 'dark') theme.value = saved
} catch {}

function toggleTheme() {
  theme.value = theme.value === 'dark' ? 'light' : 'dark'
  try { localStorage.setItem(THEME_KEY, theme.value) } catch {}
}

function toggleFullscreen() {
  if (!document.fullscreenElement) {
    document.documentElement.requestFullscreen()
  } else {
    document.exitFullscreen()
  }
}
</script>

<style scoped>
.dev-hub {
  min-height: 100vh;
  background: var(--hub-bg, #0a0c14);
  color: var(--hub-text, #e2e8f0);
}

/* ===== 顶部导航栏 ===== */
.hub-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 90px;
  padding: 0 24px;
  background: var(--hub-header-bg, rgba(15, 23, 42, 0.9));
  backdrop-filter: blur(20px);
  border-bottom: 1px solid var(--hub-border, rgba(0, 245, 212, 0.15));
  position: sticky;
  top: 0;
  z-index: 100;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 140px;
}

.brand-logo {
  display: flex;
  align-items: center;
  gap: 10px;
  cursor: pointer;
}

.logo-svg { width: 52px; height: 52px; }

.brand-name {
  font-family: 'Orbitron', 'JetBrains Mono', monospace;
  font-size: 20px;
  font-weight: 700;
  background: linear-gradient(135deg, #00f5d4 0%, #4facfe 40%, #a855f7 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.header-nav {
  display: flex;
  gap: 4px;
}

.nav-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 24px;
  border-radius: 12px;
  font-size: 16px;
  font-weight: 500;
  color: var(--hub-text-muted, #64748b);
  background: transparent;
  border: 1px solid transparent;
  cursor: pointer;
  transition: all 0.2s ease;
  font-family: inherit;
}

.nav-btn:hover {
  color: var(--hub-text, #e2e8f0);
  background: var(--hub-hover, rgba(0, 245, 212, 0.05));
}

.nav-btn.active {
  color: var(--hub-primary, #00f5d4);
  background: var(--hub-active-bg, rgba(0, 245, 212, 0.1));
  border-color: rgba(0, 245, 212, 0.2);
}

.nav-icon { font-size: 18px; }

.header-right {
  display: flex;
  align-items: center;
  gap: 4px;
}

.header-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 12px;
  background: transparent;
  border: 1px solid transparent;
  border-radius: 8px;
  color: var(--hub-text-muted, #64748b);
  cursor: pointer;
  transition: all 0.2s ease;
  font-family: inherit;
  font-size: 13px;
}

.header-btn:hover {
  color: var(--hub-text, #e2e8f0);
  background: var(--hub-hover, rgba(0, 245, 212, 0.05));
  border-color: rgba(0, 245, 212, 0.15);
}

.btn-icon { font-size: 16px; }
.header-link { text-decoration: none; }

/* ===== 浅色模式 ===== */
.dev-hub[data-theme="light"] {
  --hub-bg: #f8fafc;
  --hub-header-bg: rgba(255, 255, 255, 0.9);
  --hub-border: rgba(14, 165, 233, 0.15);
  --hub-primary: #0ea5e9;
  --hub-text: #1e293b;
  --hub-text-muted: #64748b;
  --hub-hover: rgba(14, 165, 233, 0.05);
  --hub-active-bg: rgba(14, 165, 233, 0.1);
}

.dev-hub[data-theme="light"] .header-btn { color: #64748b; }
.dev-hub[data-theme="light"] .header-btn:hover { color: #1e293b; }
.dev-hub[data-theme="light"] .nav-btn { color: #64748b; }
.dev-hub[data-theme="light"] .nav-btn:hover { color: #1e293b; }
.dev-hub[data-theme="light"] .nav-btn.active { color: #0ea5e9; }
</style>
