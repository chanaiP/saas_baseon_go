<template>
  <div
    class="neuro-command-layout"
    :class="{ 'is-integration-center-route': isIntegrationCenterRoute }"
    :data-theme="theme"
  >
    <!-- 神经背景层 -->
    <div class="neuro-bg-layer">
      <div class="neuro-particle-field">
        <div v-for="n in 20" :key="n" class="neuro-particle" :style="{
          left: `${Math.random() * 100}%`,
          top: `${Math.random() * 100}%`,
          animationDelay: `${Math.random() * 5}s`
        }"></div>
      </div>
      <div class="neuro-connection-grid"></div>
    </div>

    <!-- 主界面层 -->
    <div class="neuro-interface">
      <!-- 顶部神经主干菜单 -->
      <header class="neuro-stem-menu">
        <div class="stem-brand">
          <!-- 可编辑的Logo和系统名称 -->
          <NeuroLogoEditor
            :current-logo="customLogoUrl"
            :current-name="systemName"
            :current-copyright="copyrightInfo"
            :is-platform-admin="platformUiScope"
            :can-edit-branding="canEditStemBranding"
            :can-edit-footer="tenantBrand.canEditFooter"
            @update:logo="handleLogoUpdate"
            @update:name="handleNameUpdate"
            @update:copyright="handleCopyrightUpdate"
            @upload-success="handleLogoUploadSuccess"
            @upload-error="handleLogoUploadError"
            @settings-change="handleSettingsChange"
          />
        </div>
        
        <nav class="stem-nav">
          <div v-for="item in primaryMenu" :key="item.id"
               class="stem-item-wrapper">
            <div class="stem-item"
                 :class="{ 'stem-item-active': activePrimary === item.id }"
                 @click="activatePrimary(item.id)">
              <span class="stem-label">{{ item.title }}</span>
              <div class="stem-glow" v-if="activePrimary === item.id"></div>
              <div class="stem-connection" v-if="activePrimary === item.id"></div>
            </div>
          </div>
        </nav>

        <div class="stem-utils">
          <button class="neuro-btn neuro-btn-ghost theme-toggle" title="主题切换" @click="toggleTheme">
            <span class="btn-icon theme-icon">{{ theme === 'dark' ? '☀️' : '🌙' }}</span>
            <span class="theme-label">{{ theme === 'dark' ? '浅色模式' : '深色模式' }}</span>
          </button>
          <button class="neuro-btn neuro-btn-ghost" title="搜索">
            <span class="btn-icon">🔍</span>
          </button>
          <button class="neuro-btn neuro-btn-ghost" title="通知">
            <span class="btn-icon">🔔</span>
            <span class="notification-dot"></span>
          </button>
          <button class="neuro-btn neuro-btn-ghost" title="全屏" @click="toggleFullscreen">
            <span class="btn-icon">⛶</span>
          </button>
        </div>
      </header>

      <!-- 左侧神经突触面板 -->
      <aside class="neuro-synapse-panel" :class="{ 'panel-expanded': activePrimary }">
        <!-- 左侧神经元链 -->
        <div class="neuron-chain">
          <div v-for="(neuron, index) in neuronChain"
               :key="index"
               class="neuron-node"
               :class="{
                 'neuron-active': neuron.isActive,
                 'neuron-connected': neuron.isConnected
               }"
               @click="handleNeuronClick(neuron)"
               :title="neuron.title">
            <div class="neuron-core"></div>
            <div class="neuron-pulse" v-if="neuron.isActive"></div>
            <div class="neuron-connection" v-if="index < neuronChain.length - 1"></div>
          </div>
        </div>

        <div class="synapse-header">
          <h3 class="synapse-title">
            <span v-if="activeSidebarPrimary">{{ sidebarPrimaryTitle }}</span>
            <span v-else>快捷入口</span>
          </h3>
          <div class="synapse-header-right">
            <button v-if="activeSidebarPrimary"
                    class="neuro-btn neuro-btn-ghost synapse-back"
                    @click="activePrimary = ''"
                    title="返回快捷入口">
              <span class="btn-icon">←</span>
              <span class="back-label">返回</span>
            </button>
            <template v-else>
              <button
                v-if="!isShortcutEditing"
                class="neuro-btn neuro-btn-ghost shortcut-edit-trigger"
                :disabled="!shortcutMenuItems.length"
                title="编辑快捷入口顺序"
                aria-label="编辑快捷入口顺序"
                @click="startShortcutEdit"
              >
                <span class="btn-icon">✎</span>
              </button>
              <div v-else class="shortcut-edit-actions">
                <button
                  class="neuro-btn neuro-btn-ghost shortcut-edit-cancel"
                  :disabled="isSavingShortcutOrder"
                  title="取消编辑快捷入口"
                  aria-label="取消编辑快捷入口"
                  @click="cancelShortcutEdit"
                >
                  取消
                </button>
                <button
                  class="neuro-btn neuro-btn-ghost shortcut-edit-save"
                  :disabled="isSavingShortcutOrder"
                  title="保存快捷入口顺序"
                  aria-label="保存快捷入口顺序"
                  @click="saveShortcutOrder"
                >
                  {{ isSavingShortcutOrder ? '保存中...' : '完成' }}
                </button>
              </div>
            </template>
          </div>
        </div>

        <div
          class="synapse-content"
          :class="{ 'is-scrolling': isSynapseScrolling }"
          @scroll.passive="handleSynapseScroll"
        >
          <!-- 快捷入口（未选中一级菜单时） -->
          <div v-if="!activeSidebarPrimary" class="neuron-menu">
            <div class="neuron-menu-container">
              <div v-if="isShortcutEditing" class="shortcut-edit-tip">拖动菜单调整顺序</div>
              <div v-if="isShortcutEditing" class="shortcut-draggable-list">
                <div
                  v-for="(item, index) in draftShortcutMenuItems"
                  :key="item.id"
                  class="neuron-menu-item shortcut-editing-item"
                  :class="{
                    'neuron-current': currentPage?.id === item.id,
                    'shortcut-drag-over': dragOverIndex === index
                  }"
                  :style="{ '--neuron-delay': index * 0.06 + 's' }"
                  :draggable="!isSavingShortcutOrder"
                  @dragstart="onShortcutDragStart($event, index)"
                  @dragover.prevent="onShortcutDragOver(index)"
                  @drop.prevent="onShortcutDrop(index)"
                  @dragend="onShortcutDragEnd"
                >
                  <div class="neuron-core-wrapper">
                    <div class="neuron-core">
                      <div class="core-inner"></div>
                      <div class="core-glow" v-if="currentPage?.id === item.id"></div>
                      <div class="core-pulse" v-if="currentPage?.id === item.id"></div>
                    </div>
                    <div class="neuron-connection" v-if="index < draftShortcutMenuItems.length - 1">
                      <div class="connection-line"></div>
                      <div class="connection-glow"></div>
                    </div>
                  </div>
                  <div class="neuron-info">
                    <div class="neuron-title">{{ item.title }}</div>
                    <div class="neuron-desc" v-if="item.description">{{ item.description }}</div>
                  </div>
                  <div class="shortcut-drag-hint" aria-hidden="true">⋮⋮</div>
                </div>
              </div>
              <template v-else>
                <div v-for="(item, index) in shortcutMenuItems" :key="item.id"
                     class="neuron-menu-item"
                     :class="{
                       'neuron-current': currentPage?.id === item.id
                     }"
                     @click="openFromShortcut(item.path, item.title, item.id)"
                     :style="{ '--neuron-delay': index * 0.1 + 's' }">
                  <div class="neuron-core-wrapper">
                    <div class="neuron-core">
                      <div class="core-inner"></div>
                      <div class="core-glow" v-if="currentPage?.id === item.id"></div>
                      <div class="core-pulse" v-if="currentPage?.id === item.id"></div>
                    </div>
                    <div class="neuron-connection" v-if="index < shortcutMenuItems.length - 1">
                      <div class="connection-line"></div>
                      <div class="connection-glow"></div>
                    </div>
                  </div>
                  <div class="neuron-info">
                    <div class="neuron-title">{{ item.title }}</div>
                    <div class="neuron-desc" v-if="item.description">{{ item.description }}</div>
                    <div v-if="isPageOpen(item.path)"
                         class="dock-indicator"
                         :class="{ 'is-active': currentPage?.id === item.id }"
                         @click.stop="closePage(item.path)"
                         :title="currentPage?.id === item.id ? '点击关闭当前页面' : '点击切换到此页面'">
                      <span class="indicator-dot"></span>
                    </div>
                  </div>
                </div>
              </template>
              <el-empty v-if="!shortcutMenuItems.length" description="打开页面，点击「添加快捷」" :image-size="80" />
            </div>
          </div>

          <!-- 二级神经元菜单（选中一级菜单时） -->
          <div v-else class="neuron-menu">
            <div class="neuron-menu-container">
              <div v-for="(item, index) in sidebarSecondaryMenu" :key="item.id"
                   class="neuron-menu-item"
                  :class="{
                    'neuron-current': isCurrentSecondaryMenu(item.path)
                  }"
                   @click="navigateTo(item.path, item.title, item.id)"
                   :style="{ '--neuron-delay': index * 0.1 + 's' }">

                <!-- 神经元核心 -->
                <div class="neuron-core-wrapper">
                  <div class="neuron-core">
                    <div class="core-inner"></div>
                    <div class="core-glow" v-if="isCurrentSecondaryMenu(item.path)"></div>
                    <div class="core-pulse" v-if="isCurrentSecondaryMenu(item.path)"></div>
                  </div>

                  <!-- 连接线（除了最后一个） -->
                  <div class="neuron-connection" v-if="index < sidebarSecondaryMenu.length - 1">
                    <div class="connection-line"></div>
                    <div class="connection-glow"></div>
                  </div>
                </div>

                <!-- 神经元信息 -->
                <div class="neuron-info">
                  <div class="neuron-title">{{ item.title }}</div>
                  <div class="neuron-desc" v-if="item.description">{{ item.description }}</div>

                  <!-- Dock 指示器：页面已打开时显示 -->
                  <div v-if="isPageOpen(item.path)"
                       class="dock-indicator"
                       :class="{ 'is-active': isCurrentSecondaryMenu(item.path) }"
                       @click.stop="closePage(item.path)"
                       :title="isCurrentSecondaryMenu(item.path) ? '点击关闭当前页面' : '点击切换到此页面'">
                    <span class="indicator-dot"></span>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- 神经身份节点（整块可点击） -->
        <div class="neuro-identity-node clickable" @click="goToProfileSettings">
          <!-- 可编辑的用户头像 -->
          <NeuroAvatarUpload
            :user-id="userId"
            :user-name="userName"
            :current-avatar="userAvatarUrl"
            @update:avatar="handleAvatarUpdate"
            @upload-success="handleAvatarUploadSuccess"
            @upload-error="handleAvatarUploadError"
          />
          <div class="identity-info">
            <div class="identity-name">{{ userName }}</div>
            <div class="identity-role">{{ userRole }}</div>
          </div>
          <div class="identity-settings">
            <button class="logout-btn" @click.stop="logout" title="退出登录">
              <svg class="logout-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"/>
                <polyline points="16 17 21 12 16 7"/>
                <line x1="21" y1="12" x2="9" y2="12"/>
              </svg>
            </button>
          </div>
        </div>
      </aside>

      <!-- 内容区域 -->
      <main class="neuro-content-field">
        <div class="content-header">
          <div class="content-breadcrumb">
            <span class="home-pill" @click="navigateHome" title="返回首页">
              <svg class="home-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"/>
                <polyline points="9 22 9 12 15 12 15 22"/>
              </svg>
              <span class="home-label">首页</span>
            </span>
            <span v-if="currentPrimaryTitle || (currentPage && currentPage.id !== 'home')" class="breadcrumb-divider">|</span>
            <template v-if="currentPrimaryTitle">
              <span class="breadcrumb-item">{{ currentPrimaryTitle }}</span>
            </template>
            <template v-if="currentPage && currentPage.id !== 'home'">
              <span class="breadcrumb-separator">›</span>
              <span class="breadcrumb-item active">{{ currentPage.title }}</span>
            </template>
          </div>
          <div class="content-actions">
            <div v-if="currentPage && currentPage.id !== 'home'" class="page-controls">
              <button class="neuro-btn neuro-btn-ghost page-refresh-btn"
                      @click="refreshCurrentPage"
                      title="刷新当前页面">
                <span class="btn-icon">⟳</span>
                <span class="btn-text">刷新</span>
              </button>
              <button class="neuro-btn neuro-btn-ghost shortcut-toggle-btn"
                      @click="toggleShortcut"
                      :title="isCurrentPageShortcut ? '取消快捷入口' : '添加快捷入口'">
                <span class="btn-icon">{{ isCurrentPageShortcut ? '★' : '☆' }}</span>
                <span class="btn-text">{{ isCurrentPageShortcut ? '取消快捷' : '添加快捷' }}</span>
              </button>
              <div class="close-control-wrapper">
                <button
                  class="neuro-btn neuro-btn-ghost close-control-btn"
                  :class="{ 'has-multiple': nonHomePageCount > 1 }"
                  @click="handleCloseAction"
                  :title="getCloseButtonTitle"
                >
                  <span class="btn-icon">×</span>
                  <span class="btn-text">{{ getCloseButtonText }}</span>
                </button>
                <div class="close-dropdown">
                  <button class="close-dropdown-item" @click="closeCurrentPage">
                    <span class="dropdown-icon">×</span>
                    关闭当前页面
                  </button>
                  <button class="close-dropdown-item" @click="closeOtherPages">
                    <span class="dropdown-icon">✕</span>
                    关闭其他页面
                  </button>
                  <button class="close-dropdown-item" @click="closeAllPages">
                    <span class="dropdown-icon">✕</span>
                    关闭所有页面
                  </button>
                </div>
              </div>
            </div>
            <button class="neuro-btn neuro-btn-primary matrix-entry-btn" @click="openMenuMatrix">
              <span class="btn-icon">▦</span>
              功能矩阵全景图
            </button>
          </div>
        </div>

        <div
          class="content-main"
          :style="isIntegrationCenterRoute ? { background: '#111827', backgroundImage: 'none' } : undefined"
        >
          <router-view v-slot="{ Component }" :key="routerViewKey">
            <keep-alive :include="cachedPages">
              <component :is="Component" />
            </keep-alive>
          </router-view>
        </div>
      </main>
      <div v-if="matrixVisible" class="matrix-overlay" @click.self="closeMenuMatrix">
        <div class="matrix-panel">
          <div class="matrix-header">
            <div class="matrix-title-wrap">
              <h3 class="matrix-title">功能矩阵全景图</h3>
              <p class="matrix-subtitle">点击可快速访问；无权限菜单点击后将提示“没有权限”</p>
            </div>
            <button class="neuro-btn neuro-btn-ghost matrix-close-btn" @click="closeMenuMatrix">关闭</button>
          </div>
          <div class="matrix-toolbar">
            <input v-model.trim="matrixQuery" class="matrix-search" placeholder="搜索菜单名称或路由..." />
          </div>
          <div class="matrix-body">
            <div v-if="!matrixGroups.length" class="matrix-empty">未匹配到菜单</div>
            <section v-for="group in matrixGroups" :key="group.name" class="matrix-group">
              <h4 class="matrix-group-title">{{ group.name }}（{{ group.count }}）</h4>
              <div class="matrix-grid">
                <button
                  v-for="item in group.items"
                  :key="item.id"
                  class="matrix-card"
                  :class="{
                    'is-accessible': item.canAccess,
                    'is-locked': !item.canAccess,
                    'is-current': currentPage?.id === item.id,
                  }"
                  :title="item.path"
                  @click="onMatrixMenuClick(item)"
                >
                  <div class="matrix-card-head">
                    <span class="matrix-card-title">{{ item.title }}</span>
                    <span class="matrix-card-badge">{{ item.canAccess ? '可访问' : '无权限' }}</span>
                  </div>
                  <div class="matrix-card-path">{{ item.path }}</div>
                </button>
              </div>
            </section>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { silentDebug } from '@/utils/debug'
import { ref, reactive, watch, onMounted, onBeforeUnmount, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { useRoute, useRouter } from 'vue-router'
import { useUiPreferencesStore } from '@/stores/uiPreferences'
import { usePermissionStore } from '@/stores/permission'
import { useTenantBrandingStore } from '@/stores/tenantBranding'
import { useSidebarMenuStore, filterPlatformOnlyMenus, buildDefaultMenuTreeSnapshot } from '@/stores/sidebarMenu'
import { useShortcutStore } from '@/stores/shortcut'
import type { MenuNode } from '@/types/menu'
import { confirmStandardAction } from '@/composables/useStandardConfirm'
import { filterVisibleMenuTree } from '@/utils/menuVisibility'
import NeuroLogoEditor from '@/components/NeuroLogoEditor.vue'
import NeuroAvatarUpload from '@/components/NeuroAvatarUpload.vue'

const route = useRoute()
const router = useRouter()
const uiPrefs = useUiPreferencesStore()
const perm = usePermissionStore()
const tenantBrand = useTenantBrandingStore()
const sidebarMenu = useSidebarMenuStore()
const shortcut = useShortcutStore()

// 缓存的页面组件名称列表（keep-alive 匹配组件 name，不是菜单 ID）
const cachedPages = computed(() => {
  return openPages.value.map(p => p.componentName).filter(n => n && n !== 'HomeDashboardView')
})

// 主题管理
const theme = computed(() => uiPrefs.theme)

function toggleTheme() {
  const newTheme = theme.value === 'dark' ? 'light' : 'dark'
  uiPrefs.setTheme(newTheme)
  updateThemeStyles()
}

// Logo 和系统名称
const customLogoUrl = ref('')
const systemName = ref('Ai DevOS')
const copyrightInfo = ref('© 2026 SaaS - AI协作开发系统')
const userAvatarUrl = ref('')

const userName = computed(() => perm.profile?.name || '管理员')
const userId = computed(() => String(perm.profile?.id ?? ''))
const platformUiScope = computed(
  () => !!(perm.profile?.is_platform_admin || perm.profile?.tenant_is_platform),
)

/** 侧栏品牌：与后端 ``brand:edit`` + 套餐 ``brand_config`` 一致；平台管理员可直通 */
const canEditStemBranding = computed(
  () =>
    !!perm.profile?.is_platform_admin ||
    tenantBrand.canEdit ||
    perm.canUseAction('brand:edit'),
)
const userRole = computed(() => {
  if (perm.profile?.is_platform_admin) return '平台管理员'
  if (perm.profile?.tenant_is_platform) return '平台运维'
  const codes = perm.profile?.role_codes ?? []
  if (codes.includes('admin')) return '超级管理员'
  if (perm.profile?.role_ids?.length) return `角色ID: ${perm.profile.role_ids[0]}`
  return '系统管理员'
})

const isSynapseScrolling = ref(false)
let synapseScrollTimer: number | undefined

function handleSynapseScroll() {
  isSynapseScrolling.value = true
  if (synapseScrollTimer !== undefined) {
    window.clearTimeout(synapseScrollTimer)
  }
  synapseScrollTimer = window.setTimeout(() => {
    isSynapseScrolling.value = false
    synapseScrollTimer = undefined
  }, 900)
}

onBeforeUnmount(() => {
  if (synapseScrollTimer !== undefined) {
    window.clearTimeout(synapseScrollTimer)
  }
})

const baseMenuTree = computed(() =>
  platformUiScope.value ? sidebarMenu.adminTree : filterPlatformOnlyMenus(sidebarMenu.tenantTree),
)

const finalMenuTree = computed(() =>
  filterVisibleMenuTree(baseMenuTree.value, (code) => perm.canUseAction(code), (path) => perm.canUseMenuPath(path)),
)

const matrixSourceTree = computed(() =>
  platformUiScope.value ? buildDefaultMenuTreeSnapshot() : filterPlatformOnlyMenus(buildDefaultMenuTreeSnapshot()),
)

const handleLogoUpdate = (logoUrl: string) => { customLogoUrl.value = logoUrl }
const handleNameUpdate = (name: string) => { systemName.value = name }
const handleCopyrightUpdate = (copyright: string) => { copyrightInfo.value = copyright }
const handleLogoUploadSuccess = (_logoUrl: string) => { /* persisted */ }
const handleLogoUploadError = (_error: string) => { /* show toast */ }
const handleSettingsChange = async (settings: { logo: string; name: string; copyright: string }) => {
  try {
    await tenantBrand.save(
      settings.name,
      settings.logo,
      tenantBrand.canEditFooter ? settings.copyright : undefined,
    )
    // store 状态已由 save() 自动更新，同步到本地
    customLogoUrl.value = settings.logo
    systemName.value = settings.name
    if (tenantBrand.canEditFooter) {
      copyrightInfo.value = settings.copyright
    }
  } catch (e) {
    console.error('品牌设置保存失败:', e)
    ElMessage.error('保存失败：' + (e instanceof Error ? e.message : '未知错误'))
  }
}
const handleAvatarUpdate = (avatarUrl: string) => { userAvatarUrl.value = avatarUrl }
const handleAvatarUploadSuccess = (_avatarUrl: string) => { /* persisted */ }
const handleAvatarUploadError = (_error: string) => { /* show toast */ }

// const userInitials = computed(() => {
//   const name = userName.value
//   if (name.length >= 2) {
//     return name.substring(0, 2).toUpperCase()
//   }
//   return name.charAt(0).toUpperCase()
// })

// const userRole = computed(() => {
//   return perm.profile?.role_ids?.[0] ? `角色ID: ${perm.profile.role_ids[0]}` : '系统管理员'
// })

// 一级菜单数据 - 从菜单管理配置读取，只展示目录
const primaryMenu = computed(() => {
  const items: { id: string; title: string }[] = []
  for (const n of finalMenuTree.value) {
    if (n.enabled === false) continue
    if (n.type === 'directory') {
      items.push({ id: n.id, title: n.title })
    }
  }
  return items
})

// 当前页面对应的一级菜单标题
const currentPrimaryTitle = computed(() => {
  if (!currentPage.value || currentPage.value.id === 'home') return ''
  return findRouteMenuMatch(currentPage.value.path)?.primaryTitle ?? ''
})

// 快捷入口数据 - 保留原有的首页等快捷入口
// 二级菜单数据 - 从菜单管理配置动态生成
const secondaryMenus = computed(() => {
  const menus: Record<string, Array<{ id: string; title: string; description: string; path: string }>> = {}

  finalMenuTree.value.forEach((node) => {
    if (node.enabled === false) return
    if (node.type === 'directory' && node.children?.length) {
      menus[node.id] = node.children
        .filter((c): c is MenuNode & { path: string } =>
          c.type === 'menu' && c.enabled !== false && !!c.path && c.path !== directoryRootPath(node),
        )
        .map((c) => ({
          id: c.id,
          title: c.title,
          description: '',
          path: c.path,
        }))
    }
  })

  return menus
})

function directoryRootPath(node: MenuNode) {
  if (node.id.startsWith('manifest-app-')) return `/${node.id.slice('manifest-app-'.length)}`
  return ''
}

function collectMatrixMenuItems(
  nodes: MenuNode[],
  rootGroup: string | null,
  out: MatrixMenuItem[],
): void {
  for (const node of nodes) {
    if (node.type === 'directory') {
      const nextGroup = rootGroup ?? node.title
      if (node.children?.length) collectMatrixMenuItems(node.children, nextGroup, out)
      continue
    }
    if (node.type === 'menu' && node.path && node.path !== '/home') {
      out.push({
        id: node.id,
        title: node.title,
        path: node.path,
        group: rootGroup ?? '未分组',
        canAccess: node.enabled !== false && perm.canUseMenuPath(node.path),
      })
      if (node.children?.length) collectMatrixMenuItems(node.children, rootGroup ?? '未分组', out)
    }
  }
}

const matrixMenuItems = computed(() => {
  const items: MatrixMenuItem[] = []
  collectMatrixMenuItems(matrixSourceTree.value, null, items)
  return items
})

const matrixGroups = computed(() => {
  const query = matrixQuery.value.toLowerCase()
  const filtered = matrixMenuItems.value.filter((item) => {
    if (!query) return true
    return item.title.toLowerCase().includes(query) || item.path.toLowerCase().includes(query)
  })
  const grouped = new Map<string, MatrixMenuItem[]>()
  for (const item of filtered) {
    if (!grouped.has(item.group)) grouped.set(item.group, [])
    grouped.get(item.group)!.push(item)
  }
  return Array.from(grouped.entries()).map(([name, items]) => ({
    name,
    items,
    count: items.length,
  }))
})

// 页面管理状态
interface PageState {
  id: string        // 菜单树节点 ID（UUID），用于菜单高亮
  componentName: string // 组件名（对应 defineOptions name），用于 keep-alive 缓存
  path: string
  title: string
  isActive: boolean
  isOpen: boolean
}

interface MatrixMenuItem {
  id: string
  title: string
  path: string
  group: string
  canAccess: boolean
}

// 路由路径 → 组件名映射（与 router/index.ts 中 route.name 一致）
const pathToComponentName: Record<string, string> = {
  '/home': 'HomeDashboardView',
  '/apps': 'AppCenterListView',
  '/apps/clients': 'AppCenterListView',
  '/apps/tenant-openings': 'AppCenterListView',
  '/apps/trial-invites': 'AppCenterListView',
  '/apps/manifests': 'AppCenterListView',
  '/apps/audit-logs': 'AppCenterListView',
  '/ai-capability-center': 'AiDashboardView',
  '/ai-capability-center/providers': 'AiProvidersView',
  '/ai-capability-center/models': 'AiModelsView',
  '/ai-capability-center/scenarios': 'AiScenariosView',
  '/ai-capability-center/routes': 'AiRoutesView',
  '/ai-capability-center/strategy': 'AiStrategyView',
  '/ai-capability-center/test-console': 'AiTestConsoleView',
  '/ai-capability-center/settings': 'AiSettingsView',
  '/data-center': 'DataCenterView',
  '/data-center/dashboard': 'DataCenterView',
  '/data-center/overview': 'DataCenterView',
  '/data-center/raw': 'DataCenterView',
  '/data-center/standard': 'DataCenterView',
  '/data-center/metrics': 'DataCenterView',
  '/data-center/anomalies': 'DataCenterView',
  '/data-center/rules': 'DataCenterView',
  '/data-center/tasks': 'DataCenterView',
  '/data-center/reviews': 'DataCenterView',
  '/integration-center': 'IntegrationCenterOverview',
  '/integration-center/platforms': 'IntegrationCenterPlatforms',
  '/integration-center/workspace': 'IntegrationCenterWorkspace',
  '/integration-center/tenant-connections': 'IntegrationCenterTenantConnections',
  '/integration-center/sync-monitor': 'IntegrationCenterSyncMonitor',
  '/integration-center/quota': 'IntegrationCenterQuota',
  '/integration-center/alerts': 'IntegrationCenterAlerts',
  '/integration-center/logs': 'IntegrationCenterLogs',
  '/tenants': 'TenantView',
  '/plans': 'PlanManagementView',
  '/organization': 'OrganizationView',
  '/positions': 'PositionView',
  '/users': 'UserView',
  '/roles': 'RoleView',
  '/roles/:roleId/permission-config': 'RolePermissionConfigView',
  '/menus': 'MenuView',
  '/dict': 'DictView',
  '/params': 'ParamView',
  '/audit-logs': 'AuditLogsView',
  '/login-logs': 'LoginLogsView',
  '/monitor/health': 'MonitorHealthView',
  '/monitor/server': 'MonitorServerView',
  '/monitor/jobs': 'MonitorJobsView',
  '/monitor/services': 'MonitorServicesView',
  '/monitor/cache': 'MonitorCacheView',
  '/monitor/cache-keys': 'MonitorCacheKeysView',
  '/profile': 'ProfileView',
}

function getComponentNameByPath(path: string): string {
  if (pathToComponentName[path]) return pathToComponentName[path]
  if (/^\/roles\/[^/]+\/permission-config$/.test(path)) return pathToComponentName['/roles/:roleId/permission-config']
  return ''
}

interface RouteMenuMatch {
  menu: MenuNode & { path: string }
  primaryId: string
  primaryTitle: string
}

function routeMatchesMenuPath(routePath: string, menuPath?: string): boolean {
  if (!menuPath) return false
  if (routePath === menuPath) return true
  if (menuPath === '/home') return false
  return routePath.startsWith(menuPath + '/')
}

function findBestMenuInTree(nodes: MenuNode[], routePath: string): (MenuNode & { path: string }) | null {
  let best: (MenuNode & { path: string }) | null = null
  let bestLength = -1

  function visit(list: MenuNode[]) {
    for (const node of list) {
      if (node.enabled === false) continue
      if (node.type === 'menu' && node.path && routeMatchesMenuPath(routePath, node.path)) {
        if (node.path.length > bestLength) {
          best = node as MenuNode & { path: string }
          bestLength = node.path.length
        }
      }
      if (node.children?.length) visit(node.children)
    }
  }

  visit(nodes)
  return best
}

function findRouteMenuMatch(routePath: string): RouteMenuMatch | null {
  for (const root of finalMenuTree.value) {
    if (root.enabled === false) continue
    if (root.type === 'directory') {
      const menu = findBestMenuInTree(root.children || [], routePath)
      if (menu) {
        return {
          menu,
          primaryId: root.id,
          primaryTitle: root.title,
        }
      }
    }
  }
  return null
}

function isRootStandaloneMenuRoute(routePath: string): boolean {
  return finalMenuTree.value.some((root) =>
    root.enabled !== false && root.type === 'menu' && root.path && routeMatchesMenuPath(routePath, root.path),
  )
}

function firstPrimaryDefaultMenu() {
  for (const item of primaryMenu.value) {
    const target = defaultMenuForPrimary(item.id)
    if (target) return target
  }
  return null
}

const integrationRouteTitles: Record<string, string> = {
  '/integration-center': '总览',
  '/integration-center/platforms': '接入平台',
  '/integration-center/workspace': '集成工作台',
  '/integration-center/tenant-connections': '租户连接',
  '/integration-center/sync-monitor': '同步监控',
  '/integration-center/quota': '配额与限流',
  '/integration-center/alerts': '异常监控',
  '/integration-center/logs': '调用日志',
}

function fallbackPrimaryIdForRoute(routePath: string): string {
  if (routePath.startsWith('/integration-center')) {
    return primaryMenu.value.find((item) => item.title === '第三方集成中心')?.id ?? ''
  }
  return ''
}

function syncNavigationWithRoute(path: string) {
  openPages.value.forEach(p => { p.isActive = false })

  if (!path || path === '/' || path === '/home') {
    activePrimary.value = ''
    const homePage = openPages.value.find(p => p.id === 'home')
    if (homePage) {
      homePage.isActive = true
      currentPage.value = homePage
    } else {
      const newHome: PageState = {
        id: 'home',
        componentName: 'HomeDashboardView',
        path: '/home',
        title: '首页',
        isActive: true,
        isOpen: true,
      }
      openPages.value.unshift(newHome)
      currentPage.value = newHome
    }
    return
  }

  if (isRootStandaloneMenuRoute(path)) {
    const target = firstPrimaryDefaultMenu()
    if (target && target.path !== path) {
      navigateTo(target.path, target.title, target.id)
      return
    }
  }

  const match = findRouteMenuMatch(path)
  if (match) {
    activePrimary.value = match.primaryId
  } else {
    const fallbackPrimaryId = fallbackPrimaryIdForRoute(path)
    if (fallbackPrimaryId) activePrimary.value = fallbackPrimaryId
  }

  const existingPage = openPages.value.find(p => p.path === path)
  if (existingPage) {
    existingPage.isActive = true
    if (match) {
      existingPage.id = match.menu.id
      existingPage.title = match.menu.title
      existingPage.componentName = getComponentNameByPath(path)
    }
    currentPage.value = existingPage
    return
  }

  const title = match?.menu.title || integrationRouteTitles[path] || String(route.meta?.title || '未命名页面')
  const page: PageState = {
    id: match?.menu.id || path,
    componentName: getComponentNameByPath(path),
    path,
    title,
    isActive: true,
    isOpen: true,
  }
  openPages.value.push(page)
  currentPage.value = page
}

// 状态
// 路由视图 key，改变时强制 keep-alive 组件重新挂载（实现刷新）
const routerViewKey = ref(0)

const activePrimary = ref<string>('')
const isFullscreen = ref(false)
const matrixVisible = ref(false)
const matrixQuery = ref('')
const currentPage = ref<PageState | null>(null)
const openPages = ref<PageState[]>([
  { id: 'home', componentName: 'HomeDashboardView', path: '/home', title: '首页', isActive: true, isOpen: true }
])
const isShortcutEditing = ref(false)
const isSavingShortcutOrder = ref(false)
const originalShortcutIds = ref<string[]>([])
type ShortcutMenuItem = { id: string; title: string; path: string; description: string }
const draftShortcutMenuItems = ref<ShortcutMenuItem[]>([])
const dragFromIndex = ref<number | null>(null)
const dragOverIndex = ref<number | null>(null)

// 神经元链数据
interface NeuronNode {
  id: string
  title: string
  action: string
  isActive: boolean
  isConnected: boolean
  target?: string
}

const neuronChain = ref<NeuronNode[]>([
  { id: 'neuron1', title: '刷新数据', action: 'refresh', isActive: false, isConnected: true },
  { id: 'neuron2', title: '切换主题', action: 'toggleTheme', isActive: false, isConnected: true },
  { id: 'neuron3', title: '全屏模式', action: 'toggleFullscreen', isActive: false, isConnected: true },
  { id: 'neuron4', title: '新建用户', action: 'navigate', isActive: false, isConnected: false, target: '/users' },
  { id: 'neuron5', title: '系统监控', action: 'navigate', isActive: false, isConnected: true, target: '/monitor/health' }
])

// 方法
const activatePrimary = (id: string) => {
  activePrimary.value = id
  const target = defaultMenuForPrimary(id)
  if (target) {
    navigateTo(target.path, target.title, target.id)
  }
}

const getPrimaryTitle = (id: string) => primaryMenu.value.find((item) => item.id === id)?.title ?? ''

const getSecondaryMenu = (id: string) => {
  return secondaryMenus.value[id as keyof typeof secondaryMenus.value] || []
}

function defaultMenuForPrimary(id: string) {
  const items = visibleSecondaryMenus(getSecondaryMenu(id))
  if (!items.length) return null
  const openPaths = new Set(openPages.value.filter((page) => page.id !== 'home').map((page) => page.path))
  return items.find((item) => openPaths.has(item.path)) ?? items[0]
}

const integrationSidebarItems = Object.entries(integrationRouteTitles).map(([path, title]) => ({
  id: path,
  title,
  description: '',
  path,
}))

const isIntegrationCenterRoute = computed(() => route.path.startsWith('/integration-center'))
const activeSidebarPrimary = computed(() =>
  activePrimary.value || (isIntegrationCenterRoute.value ? '__integration_center_fallback__' : ''),
)
const sidebarPrimaryTitle = computed(() =>
  activePrimary.value ? getPrimaryTitle(activePrimary.value) : (isIntegrationCenterRoute.value ? '第三方集成中心' : ''),
)
const sidebarSecondaryMenu = computed(() =>
  {
    const menu = activePrimary.value ? getSecondaryMenu(activePrimary.value) : []
    if (activePrimary.value) return visibleSecondaryMenus(menu)
    if (isIntegrationCenterRoute.value) return integrationSidebarItems
    return visibleSecondaryMenus(menu)
  },
)

function canonicalSecondaryMenuPath(path: string): string {
  if (path === '/data-center' || path === '/data-center/overview') return '/data-center/raw'
  return path
}

function visibleSecondaryMenus(items: Array<{ id: string; title: string; description: string; path: string }>) {
  return items.filter((item) => item.path !== '/data-center' && item.path !== '/data-center/overview')
}

const currentRouteMenuPath = computed(() => {
  const matched = findRouteMenuMatch(route.path)?.menu.path ?? currentPage.value?.path ?? ''
  return canonicalSecondaryMenuPath(matched)
})

function isCurrentSecondaryMenu(path: string): boolean {
  return currentRouteMenuPath.value === canonicalSecondaryMenuPath(path)
}

// 导航到页面
const navigateTo = (path: string, title: string, id: string) => {
  const existingPage = openPages.value.find(page => page.path === path)

  if (!existingPage) {
    const newPage: PageState = {
      id,
      componentName: getComponentNameByPath(path),
      path,
      title,
      isActive: true,
      isOpen: true,
    }
    openPages.value.push(newPage)
    updateSecondaryMenuOpenState(id, true)
  }

  router.push(path)
}

// 从快捷入口打开页面 — 不切换左侧菜单面板
const openFromShortcut = (path: string, title: string, id: string) => {
  navigateTo(path, title, id)
}

const openMenuMatrix = () => {
  matrixVisible.value = true
}

const closeMenuMatrix = () => {
  matrixVisible.value = false
  matrixQuery.value = ''
}

const onMatrixMenuClick = (item: MatrixMenuItem) => {
  if (!item.canAccess) {
    ElMessage.warning('没有权限')
    return
  }
  navigateTo(item.path, item.title, item.id)
  closeMenuMatrix()
}

// 导航到首页
const navigateHome = () => {
  openPages.value.forEach(p => { p.isActive = false })
  const homePage = openPages.value.find(p => p.id === 'home')
  if (homePage) {
    homePage.isActive = true
    currentPage.value = homePage
  } else {
    const newHome: PageState = { id: 'home', componentName: 'HomeDashboardView', path: '/home', title: '首页', isActive: true, isOpen: true }
    openPages.value.unshift(newHome)
    currentPage.value = newHome
  }
  router.push('/home')
}

// 关闭当前页面
const closeCurrentPage = () => {
  if (!currentPage.value) return
  if (currentPage.value.id === 'home') return

  const pageId = currentPage.value.id
  openPages.value = openPages.value.filter(page => page.id !== pageId)
  updateSecondaryMenuOpenState(pageId, false)

  // 切到最后一个剩余的非首页页面，没有则回首页
  const remaining = openPages.value.filter(p => p.id !== 'home')
  if (remaining.length > 0) {
    const last = remaining[remaining.length - 1]
    last.isActive = true
    currentPage.value = last
    router.push(last.path)
  } else {
    navigateHome()
  }
}

// 智能关闭处理 — 1 个页面直接关闭，多个页面展开下拉菜单
const handleCloseAction = () => {
  if (nonHomePageCount.value <= 1) {
    closeCurrentPage()
  }
  // > 1 时靠 CSS :hover 展开下拉菜单，无需额外逻辑
}

// 非首页的打开页面数
const nonHomePageCount = computed(() => {
  return openPages.value.filter(p => p.id !== 'home').length
})

// 获取关闭按钮文本
const getCloseButtonText = computed(() => {
  const count = nonHomePageCount.value
  return count > 1 ? `关闭(${count})` : '关闭'
})

// 获取关闭按钮标题
const getCloseButtonTitle = computed(() => {
  return '关闭已打开的页面'
})

// 关闭所有页面，回到首页
const closeAllPages = () => {
  // 关闭除首页外的所有页面
  const pagesToClose = openPages.value.filter(page => page.id !== 'home')
  pagesToClose.forEach(page => {
    updateSecondaryMenuOpenState(page.id, false)
  })

  // 只保留首页
  openPages.value = openPages.value.filter(page => page.id === 'home')
  navigateHome()
}

// 关闭其他页面（除了当前页面和首页）
const closeOtherPages = () => {
  if (!currentPage.value) return

  const currentId = currentPage.value.id
  const pagesToClose = openPages.value.filter(page => page.id !== currentId && page.id !== 'home')

  // 关闭其他页面
  pagesToClose.forEach(page => {
    updateSecondaryMenuOpenState(page.id, false)
  })

  // 只保留当前页面和首页
  openPages.value = openPages.value.filter(page => page.id === currentId || page.id === 'home')

  // 导航到当前页面（确保 keep-alive 生效）
  router.push(currentPage.value.path)
}

// 刷新当前页面 — 改变 routerViewKey 强制 keep-alive 组件重新挂载
const refreshCurrentPage = () => {
  routerViewKey.value++
}

// 刷新所有页面
const refreshAllPages = () => {
  routerViewKey.value++
}

// 神经元点击处理
const handleNeuronClick = (neuron: NeuronNode) => {
  // 激活当前神经元
  neuronChain.value.forEach(n => {
    n.isActive = n.id === neuron.id
  })

  // 执行对应操作
  switch (neuron.action) {
    case 'refresh':
      refreshAllPages()
      break
    case 'toggleTheme':
      toggleTheme()
      break
    case 'toggleFullscreen':
      toggleFullscreen()
      break
    case 'navigate':
      if (neuron.target) {
        // 这里可以添加导航逻辑
        silentDebug('导航到:', neuron.target)
      }
      break
  }

  // 3秒后取消激活状态
  setTimeout(() => {
    neuronChain.value.forEach(n => {
      n.isActive = false
    })
  }, 3000)
}


// 固定页面状态
const pinnedPages = ref<Set<string>>(new Set())

// 检查当前页面是否被固定
const isCurrentPagePinned = computed(() => {
  return currentPage.value ? pinnedPages.value.has(currentPage.value.id) : false
})

// 固定/取消固定当前页面
const pinCurrentPage = () => {
  if (!currentPage.value) return

  const pageId = currentPage.value.id
  if (pinnedPages.value.has(pageId)) {
    pinnedPages.value.delete(pageId)
    silentDebug('取消固定页面:', currentPage.value.title)
  } else {
    pinnedPages.value.add(pageId)
    silentDebug('固定页面:', currentPage.value.title)
  }
}

// 新建相关页面
const createRelatedPage = () => {
  if (!currentPage.value) return

  // 根据当前页面类型创建相关页面
  const relatedPages: Record<string, { title: string; path: string; icon: string }> = {
    'dashboard': { title: '新建仪表板', path: '/dashboard/new', icon: '📊' },
    'ai': { title: '新建AI工具', path: '/ai-tools/new', icon: '🤖' },
    'defects': { title: '新建缺陷', path: '/defects/new', icon: '🐛' },
  }

  const related = relatedPages[currentPage.value.id] || { title: '新建页面', path: '/new', icon: '📄' }
  silentDebug('创建相关页面:', related.title)

  // 这里可以触发页面创建逻辑
  // 例如: router.push(related.path)
}

// 更新二级菜单的打开状态（不再需要，open state 从 openPages 推导）
const updateSecondaryMenuOpenState = (_pageId: string, _isOpen: boolean) => {}

// 检查页面是否已打开
const isPageOpen = (path: string): boolean => {
  return openPages.value.some(p => p.path === path)
}

// 关闭指定页面
const closePage = (path: string) => {
  const page = openPages.value.find(p => p.path === path)
  if (!page || page.id === 'home') return

  const wasActive = currentPage.value?.path === path
  openPages.value = openPages.value.filter(p => p.path !== path)

  if (wasActive) {
    // 关闭的是当前页：切到最后一个剩余的非首页页面，没有则回首页
    const remaining = openPages.value.filter(p => p.id !== 'home')
    if (remaining.length > 0) {
      const last = remaining[remaining.length - 1]
      last.isActive = true
      currentPage.value = last
      router.push(last.path)
    } else {
      navigateHome()
    }
  }
  // 关闭非当前页：不做任何路由操作，只移除记录
}

// 智能控制按钮处理（保留兼容）
const handleSmartControl = (_item: any) => {}

// 快捷入口管理（服务端持久化）
const shortcutMenuItems = computed(() => {
  const idMap = new Map<string, ShortcutMenuItem>()

  finalMenuTree.value.forEach((node) => {
    if (node.type === 'directory' && node.enabled !== false && node.children?.length) {
      node.children.forEach(c => {
        if (c.type === 'menu' && c.enabled !== false && c.path) {
          idMap.set(c.id, { id: c.id, title: c.title, path: c.path, description: '' })
        }
      })
    }
  })

  return shortcut.shortcutIds
    .map(id => idMap.get(normalizeShortcutMenuId(id, idMap)))
    .filter(Boolean) as ShortcutMenuItem[]
})

function normalizeShortcutMenuId(id: string, idMap: Map<string, ShortcutMenuItem>): string {
  if (idMap.has(id)) return id
  const aliases: Record<string, string> = {
    org_manage: 'org',
    position_manage: 'pos',
    role_manage: 'role',
    user_manage: 'user',
    menu_manage: 'menu',
    dict_manage: 'dict',
    param_manage: 'param',
    business_unit_manage: 'business-unit',
    audit_log: 'audit-log',
    login_log: 'login-log',
  }
  const alias = aliases[id]
  if (alias && idMap.has(alias)) return alias
  return id
}

const startShortcutEdit = () => {
  if (!shortcutMenuItems.value.length) return
  originalShortcutIds.value = [...shortcut.shortcutIds]
  draftShortcutMenuItems.value = shortcutMenuItems.value.map(item => ({ ...item }))
  isShortcutEditing.value = true
}

const cancelShortcutEdit = () => {
  if (isSavingShortcutOrder.value) return
  isShortcutEditing.value = false
  draftShortcutMenuItems.value = []
  originalShortcutIds.value = []
  onShortcutDragEnd()
}

const saveShortcutOrder = async () => {
  if (!isShortcutEditing.value || isSavingShortcutOrder.value) return
  isSavingShortcutOrder.value = true

  const dedupedDraftIds = Array.from(new Set(draftShortcutMenuItems.value.map(item => item.id)))
  const visibleIds = new Set(shortcutMenuItems.value.map(item => item.id))
  const validDraftIds = dedupedDraftIds.filter(id => visibleIds.has(id))
  const rollbackIds = [...originalShortcutIds.value]

  shortcut.shortcutIds = [...validDraftIds]
  try {
    await shortcut.persist()
    ElMessage.success('快捷入口顺序已更新')
  } catch (error) {
    shortcut.shortcutIds = [...rollbackIds]
    ElMessage.error('保存失败，已恢复原顺序')
    console.error('快捷入口排序保存失败:', error)
  } finally {
    isSavingShortcutOrder.value = false
    isShortcutEditing.value = false
    draftShortcutMenuItems.value = []
    originalShortcutIds.value = []
    onShortcutDragEnd()
  }
}

const onShortcutDragStart = (event: DragEvent, index: number) => {
  if (isSavingShortcutOrder.value) return
  event.dataTransfer?.setData('text/plain', String(index))
  if (event.dataTransfer) event.dataTransfer.effectAllowed = 'move'
  dragFromIndex.value = index
  dragOverIndex.value = index
}

const onShortcutDragOver = (index: number) => {
  if (isSavingShortcutOrder.value) return
  dragOverIndex.value = index
}

const onShortcutDrop = (index: number) => {
  if (isSavingShortcutOrder.value) return
  if (dragFromIndex.value === null || dragFromIndex.value === index) return
  const list = [...draftShortcutMenuItems.value]
  const [moved] = list.splice(dragFromIndex.value, 1)
  list.splice(index, 0, moved)
  draftShortcutMenuItems.value = list
  dragFromIndex.value = index
}

const onShortcutDragEnd = () => {
  dragFromIndex.value = null
  dragOverIndex.value = null
}

// 当前页面是否已在快捷入口中
const isCurrentPageShortcut = computed(() => {
  if (!currentPage.value) return false
  for (const node of finalMenuTree.value) {
    if (node.type === 'directory' && node.enabled !== false && node.children?.length) {
      const match = node.children.find(c => c.type === 'menu' && c.path === currentPage.value?.path)
      if (match) return shortcut.shortcutIds.includes(match.id)
    }
  }
  return false
})

// 切换快捷入口（添加/移除）
const toggleShortcut = async () => {
  if (!currentPage.value) return
  let menuId = ''
  for (const node of finalMenuTree.value) {
    if (node.type === 'directory' && node.enabled !== false && node.children?.length) {
      const match = node.children.find(c => c.type === 'menu' && c.path === currentPage.value?.path)
      if (match) {
        menuId = match.id
        break
      }
    }
  }
  if (!menuId) return

  const rollbackIds = [...shortcut.shortcutIds]
  const idx = shortcut.shortcutIds.indexOf(menuId)
  if (idx >= 0) {
    shortcut.shortcutIds.splice(idx, 1)
  } else {
    shortcut.shortcutIds.push(menuId)
  }
  try {
    await shortcut.persist()
  } catch (error) {
    shortcut.shortcutIds = rollbackIds
    ElMessage.error('快捷入口保存失败，请稍后重试')
    console.error('快捷入口保存失败:', error)
  }
}


// 更新主题样式
const updateThemeStyles = () => {
  const root = document.documentElement

  if (theme.value === 'light') {
    // 浅色模式 - 灰色画布、白色块面、清晰边界
    root.style.setProperty('--nm-bg-deep', '#f3f6fb')
    root.style.setProperty('--nm-bg-surface', '#f6f8fc')
    root.style.setProperty('--nm-bg-card', '#ffffff')
    root.style.setProperty('--nm-bg-elevated', '#eef3f9')

    // 鲜明但不刺眼的主题色
    root.style.setProperty('--nm-primary', '#2563eb')
    root.style.setProperty('--nm-primary-glow', 'rgba(37, 99, 235, 0.16)')
    root.style.setProperty('--nm-secondary', '#22c7ba')
    root.style.setProperty('--nm-secondary-glow', 'rgba(34, 199, 186, 0.14)')
    root.style.setProperty('--nm-accent', '#ef4444')
    root.style.setProperty('--nm-accent-glow', 'rgba(239, 68, 68, 0.2)')

    // 高对比度的文本颜色 - 确保在浅色背景下清晰可见
    root.style.setProperty('--nm-text-primary', '#0f172a')
    root.style.setProperty('--nm-text-secondary', '#334155')
    root.style.setProperty('--nm-text-muted', '#64748b')
    root.style.setProperty('--nm-text-on-dark', '#ffffff')

    // 清晰的边框
    root.style.setProperty('--nm-border', '#d7e0ec')
    root.style.setProperty('--nm-border-glow', 'rgba(100, 116, 139, 0.20)')

    // 鲜明的状态颜色
    root.style.setProperty('--nm-success', '#10b981')
    root.style.setProperty('--nm-warning', '#f59e0b')
    root.style.setProperty('--nm-error', '#ef4444')
    root.style.setProperty('--nm-info', '#3b82f6')
  } else {
    // 恢复暗色主题 - 保持原有的神经矩阵风格
    root.style.setProperty('--nm-bg-deep', '#0a0c14')
    root.style.setProperty('--nm-bg-surface', '#111827')
    root.style.setProperty('--nm-bg-card', '#1a1f2e')
    root.style.setProperty('--nm-bg-elevated', '#242a3a')

    // 恢复原有的荧光主题色
    root.style.setProperty('--nm-primary', '#00f5d4')
    root.style.setProperty('--nm-primary-glow', 'rgba(0, 245, 212, 0.3)')
    root.style.setProperty('--nm-secondary', '#9d4edd')
    root.style.setProperty('--nm-secondary-glow', 'rgba(157, 78, 221, 0.3)')
    root.style.setProperty('--nm-accent', '#ff6b6b')
    root.style.setProperty('--nm-accent-glow', 'rgba(255, 107, 107, 0.3)')

    root.style.setProperty('--nm-text-primary', '#f8fafc')
    root.style.setProperty('--nm-text-secondary', '#94a3b8')
    root.style.setProperty('--nm-text-muted', '#64748b')
    root.style.setProperty('--nm-text-on-dark', '#0a0c14')

    root.style.setProperty('--nm-border', '#2d3748')
    root.style.setProperty('--nm-border-glow', 'rgba(0, 245, 212, 0.1)')

    // 恢复原有的状态颜色
    root.style.setProperty('--nm-success', '#10b981')
    root.style.setProperty('--nm-warning', '#f59e0b')
    root.style.setProperty('--nm-error', '#ef4444')
    root.style.setProperty('--nm-info', '#3b82f6')
  }
}

// 监听主题变化
watch(theme, () => {
  updateThemeStyles()
})

// 监听路由变化，同步 currentPage
watch(
  () => route.path,
  (newPath) => {
    syncNavigationWithRoute(newPath)
  }
)

// 初始化
onMounted(async () => {
  updateThemeStyles()
  sidebarMenu.normalizeBuiltinTree()
  syncNavigationWithRoute(route.path)

  try {
    await tenantBrand.load()
    if (tenantBrand.displayLogoSrc) customLogoUrl.value = tenantBrand.displayLogoSrc
    if (tenantBrand.displayName) systemName.value = tenantBrand.displayName
    if (tenantBrand.footerText) copyrightInfo.value = tenantBrand.footerText
  } catch {
    /* 品牌加载失败保留默认值 */
  }

  await shortcut.init(perm.profile?.shortcut_ids)
  await sidebarMenu.loadTenantMenuRuntime({
    isPlatformAdmin: !!perm.profile?.is_platform_admin,
    tenantType: perm.profile?.tenant_type,
    force: true,
  })

  syncNavigationWithRoute(route.path)
})

// 跳转到个人中心
const goToProfileSettings = () => {
  router.push('/profile')
}

// 退出登录
const logout = async () => {
  try {
    await confirmStandardAction({
      title: '退出登录',
      icon: '!',
      message: '确定要退出登录吗？',
      detail: '退出后将需要重新登录。',
      confirmText: '退出',
    })
  } catch {
    return
  }
  localStorage.removeItem('access_token')
  usePermissionStore().clear()
  sidebarMenu.clearTenantMenuRuntime()
  router.push('/login')
}

const toggleFullscreen = () => {
  isFullscreen.value = !isFullscreen.value
  if (!document.fullscreenElement) {
    document.documentElement.requestFullscreen()
  } else {
    document.exitFullscreen()
  }
}
</script>

<style scoped>
/* CSS变量现在通过JavaScript动态设置，确保全局可用 */

.neuro-command-layout {
  position: relative;
  width: 100%;
  /* 块级：高度 = max(一屏保底, 实际内容)。勿内设 flex 子项 flex:1，否则所有路由会被同一套拉伸规则拖长 scrollHeight */
  min-height: 100vh;
  min-height: 100dvh;
  height: auto;
  max-height: none;
  overflow: visible;
  background: var(--nm-bg-deep);
  color: var(--nm-text-primary);
  font-family: 'Inter', -apple-system, BlinkMacSystemFont, sans-serif;
  transition: background-color 0.3s ease, color 0.3s ease;
}

.neuro-command-layout[data-theme="light"] {
  background: #f3f6fb;
  color: #0f172a;
}

.neuro-command-layout[data-theme="light"] .neuro-bg-layer {
  display: none;
}

/* 神经背景层：固定铺满视口，不参与 neuro-command-layout 的文档高度；裁剪粒子动画避免 transform 把可滚区域撑大 */
.neuro-bg-layer {
  position: fixed;
  inset: 0;
  z-index: 0;
  pointer-events: none;
  overflow: hidden;
}

.neuro-particle-field {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
}

.neuro-particle {
  position: absolute;
  width: 1px;
  height: 1px;
  background: var(--nm-primary);
  border-radius: 50%;
  filter: blur(0.5px);
  opacity: 0;
  animation: particle-float 15s infinite linear;
}


@keyframes particle-float {
  0% {
    transform: translateY(100vh) translateX(0);
    opacity: 0;
  }
  10% {
    opacity: 0.6;
  }
  90% {
    opacity: 0.6;
  }
  100% {
    transform: translateY(-100px) translateX(100px);
    opacity: 0;
  }
}

.neuro-connection-grid {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-image:
    linear-gradient(90deg, var(--nm-primary-glow) 1px, transparent 1px),
    linear-gradient(0deg, var(--nm-primary-glow) 1px, transparent 1px);
  background-size: 40px 40px;
  opacity: 0.3;
}


/* 主界面层：块级流式布局；fixed 顶栏/侧栏不占文档流高度，由 padding-top 预留顶栏占位 */
.neuro-interface {
  position: relative;
  z-index: 1;
  box-sizing: border-box;
  padding-top: 80px;
  padding-left: 280px;
  background: var(--nm-bg-deep);
}

.neuro-command-layout[data-theme="light"] .neuro-interface {
  background: #f3f6fb;
}


/* 顶部神经主干菜单：横跨整个视口，固定在顶部 */
.neuro-stem-menu {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  display: flex;
  align-items: center;
  background: var(--nm-bg-surface);
  backdrop-filter: blur(20px);
  border-bottom: 1px solid var(--nm-border-glow);
  padding: 0 24px;
  height: 80px;
  min-height: 80px;
  z-index: 100;
  transition: background-color 0.3s ease, border-color 0.3s ease;
}

.neuro-command-layout[data-theme="light"] .neuro-stem-menu {
  background: rgba(255, 255, 255, 0.96);
  border-bottom-color: #d7e0ec;
  box-shadow: 0 12px 30px rgba(15, 23, 42, 0.06);
}

.neuro-command-layout[data-theme="light"] .neuro-stem-menu::before {
  content: '';
  position: absolute;
  inset: 0 auto 0 0;
  width: 280px;
  background:
    radial-gradient(circle at 24% 22%, rgba(0, 245, 212, 0.12), transparent 34%),
    linear-gradient(180deg, #111827 0%, #0f172a 100%);
  border-right: 1px solid rgba(0, 245, 212, 0.16);
  pointer-events: none;
}

.neuro-command-layout[data-theme="light"] .neuro-stem-menu::after {
  content: '';
  position: absolute;
  left: 0;
  bottom: -1px;
  width: 280px;
  height: 2px;
  background: linear-gradient(180deg, #111827 0%, #151b2a 100%);
  pointer-events: none;
  z-index: 3;
}


.stem-brand {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-right: 32px;
  position: relative;
  z-index: 1;
}

.stem-nav,
.stem-utils {
  position: relative;
  z-index: 1;
}


/* 全息矩阵Logo：悬浮光点矩阵 */
.hologram-matrix-logo {
  position: relative;
  width: 40px;
  height: 40px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.hologram-matrix-logo .matrix-grid {
  position: relative;
  width: 32px;
  height: 32px;
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  grid-template-rows: repeat(4, 1fr);
  gap: 4px;
  z-index: 2;
}

.matrix-dot {
  width: 4px;
  height: 4px;
  background: var(--nm-text-secondary);
  border-radius: 50%;
  opacity: 0.4;
  transition: all 0.3s ease;
}

.active-dot {
  background: var(--nm-primary);
  opacity: 0.8;
  box-shadow: 0 0 6px var(--nm-primary-glow);
  animation: matrix-dot-pulse 2s infinite;
}

@keyframes matrix-dot-pulse {
  0%, 100% {
    opacity: 0.8;
    transform: scale(1);
  }
  50% {
    opacity: 1;
    transform: scale(1.2);
  }
}

/* 连接线 */
.matrix-connection {
  position: absolute;
  background: var(--nm-primary-glow);
  height: 1px;
  opacity: 0.4;
  z-index: 1;
  animation: connection-glow 3s infinite;
}

.conn-1 {
  /* 连接点5和点7 */
  top: 25%;
  left: 25%;
  width: 25%;
  transform: rotate(0deg);
  animation-delay: 0s;
}

.conn-2 {
  /* 连接点7和点12 */
  top: 25%;
  left: 50%;
  width: 25%;
  transform: rotate(45deg);
  animation-delay: 0.5s;
}

.conn-3 {
  /* 连接点10和点12 */
  top: 50%;
  left: 50%;
  width: 25%;
  transform: rotate(0deg);
  animation-delay: 1s;
}

.conn-4 {
  /* 连接点5和点10 */
  top: 25%;
  left: 25%;
  width: 35%;
  transform: rotate(90deg);
  animation-delay: 1.5s;
}

@keyframes connection-glow {
  0%, 100% {
    opacity: 0.3;
    box-shadow: 0 0 2px var(--nm-primary-glow);
  }
  50% {
    opacity: 0.6;
    box-shadow: 0 0 6px var(--nm-primary-glow);
  }
}

/* 全息光晕效果 */
.hologram-glow {
  position: absolute;
  top: 50%;
  left: 50%;
  width: 48px;
  height: 48px;
  background: radial-gradient(circle, var(--nm-primary-glow) 0%, transparent 70%);
  border-radius: 50%;
  transform: translate(-50%, -50%);
  opacity: 0.3;
  filter: blur(4px);
  z-index: 0;
  animation: hologram-glow-pulse 4s infinite;
}

.hologram-pulse {
  position: absolute;
  top: 50%;
  left: 50%;
  width: 56px;
  height: 56px;
  border: 1px solid var(--nm-primary-glow);
  border-radius: 50%;
  transform: translate(-50%, -50%);
  opacity: 0;
  z-index: 0;
  animation: hologram-ring 4s infinite;
}

@keyframes hologram-glow-pulse {
  0%, 100% {
    opacity: 0.2;
    transform: translate(-50%, -50%) scale(1);
  }
  50% {
    opacity: 0.4;
    transform: translate(-50%, -50%) scale(1.1);
  }
}

@keyframes hologram-ring {
  0% {
    opacity: 0;
    transform: translate(-50%, -50%) scale(0.8);
  }
  50% {
    opacity: 0.3;
    transform: translate(-50%, -50%) scale(1);
  }
  100% {
    opacity: 0;
    transform: translate(-50%, -50%) scale(1.2);
  }
}

/* 悬停效果 */
.hologram-matrix-logo:hover .matrix-dot {
  opacity: 0.6;
}

.hologram-matrix-logo:hover .active-dot {
  opacity: 1;
  animation-duration: 1s;
}

.hologram-matrix-logo:hover .matrix-connection {
  opacity: 0.6;
  animation-duration: 2s;
}

/* 浅色模式优化 */
.neuro-command-layout[data-theme="light"] .matrix-dot {
  background: #94a3b8;
}

.neuro-command-layout[data-theme="light"] .active-dot {
  background: #0ea5e9;
  box-shadow: 0 0 6px rgba(14, 165, 233, 0.3);
}

.neuro-command-layout[data-theme="light"] .matrix-connection {
  background: rgba(14, 165, 233, 0.2);
}

.neuro-command-layout[data-theme="light"] .hologram-glow {
  background: radial-gradient(circle, rgba(14, 165, 233, 0.1) 0%, transparent 70%);
}

.neuro-command-layout[data-theme="light"] .hologram-pulse {
  border-color: rgba(14, 165, 233, 0.2);
}


.brand-text {
  font-family: 'JetBrains Mono', monospace;
  font-size: 22px;
  font-weight: 700;
  background: linear-gradient(135deg, var(--nm-primary), var(--nm-secondary));
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.stem-nav {
  flex: 1;
  display: flex;
  align-items: center;
  margin: 0 24px 0 48px;
}

.stem-item-wrapper {
  display: flex;
  align-items: center;
  position: relative;
}

.stem-item {
  display: flex;
  align-items: center;
  padding: 12px 16px;
  cursor: pointer;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  position: relative;
  min-width: 100px; /* 增大最小宽度以适应更大的字体 */
  height: 56px;
  justify-content: center;
  font-weight: 700;
  font-size: 28px;
  color: var(--nm-text-primary);
  border-bottom: 3px solid transparent;
}

.stem-item:hover {
  color: var(--nm-primary);
  background: rgba(0, 245, 212, 0.05);
}

.stem-item-active {
  color: var(--nm-primary);
  border-bottom-color: var(--nm-primary);
  background: rgba(0, 245, 212, 0.08);
}


.stem-icon {
  position: relative;
  width: 28px;
  height: 28px;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 6px;
}

.stem-icon-symbol {
  font-size: 16px;
  font-weight: 600;
  color: var(--nm-text-secondary);
  z-index: 1;
  transition: all 0.3s ease;
}

.stem-item-active .stem-icon-symbol {
  color: var(--nm-primary);
  text-shadow: 0 0 8px var(--nm-primary-glow);
  transform: scale(1.1);
}

.stem-glow {
  position: absolute;
  bottom: -3px;
  left: 50%;
  transform: translateX(-50%);
  width: 40px;
  height: 3px;
  background: var(--nm-primary);
  filter: blur(2px);
  opacity: 0.6;
  animation: stem-glow-pulse 2s infinite;
}

@keyframes stem-glow-pulse {
  0%, 100% { opacity: 0.3; transform: translate(-50%, -50%) scale(1); }
  50% { opacity: 0.5; transform: translate(-50%, -50%) scale(1.2); }
}

.stem-label {
  font-size: 14px;
  font-weight: 500;
  color: var(--nm-text-secondary);
  transition: color 0.3s;
}

.stem-item-active .stem-label {
  color: var(--nm-primary);
  font-weight: 600;
}

/* 在浅色模式下优化标签颜色 */
.neuro-command-layout[data-theme="light"] .stem-label {
  color: #475569;
}

.neuro-command-layout[data-theme="light"] .stem-item-active .stem-label {
  color: #0ea5e9;
  font-weight: 600;
}

.stem-connection {
  position: absolute;
  bottom: -1px;
  left: 50%;
  transform: translateX(-50%);
  width: 60%;
  height: 2px;
  background: linear-gradient(90deg, transparent, var(--nm-primary), transparent);
  border-radius: 1px;
}

.stem-utils {
  display: flex;
  gap: 8px;
}

/* 按钮样式 */
.neuro-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  padding: 8px 12px;
  font-family: 'Inter', sans-serif;
  font-size: 13px;
  font-weight: 500;
  border: none;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s ease;
  outline: none;
  background: transparent;
  color: var(--nm-text-secondary);
  position: relative; /* 为通知红点定位 */
}

.neuro-btn:hover {
  background: var(--nm-primary-glow);
  color: var(--nm-primary);
}

/* 在浅色模式下优化按钮文字颜色 */
.neuro-command-layout[data-theme="light"] .neuro-btn {
  color: #475569;
}

.neuro-command-layout[data-theme="light"] .neuro-btn:hover {
  color: #0ea5e9;
}

.neuro-btn-ghost {
  border: 1px solid var(--nm-border-glow);
}

.neuro-btn-primary {
  background: linear-gradient(135deg, var(--nm-primary), var(--nm-secondary));
  color: var(--nm-text-on-dark);
  font-weight: 600;
  border: none;
}

.neuro-btn-primary:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 20px var(--nm-primary-glow);
}


.btn-icon {
  font-size: 14px;
}

/* 主题切换按钮特殊样式 */
.theme-toggle {
  position: relative;
  overflow: hidden;
}

.theme-toggle:hover {
  background: var(--nm-primary-glow);
  border-color: var(--nm-primary);
}

.theme-icon {
  font-size: 16px;
  transition: transform 0.3s ease;
}

.theme-toggle:hover .theme-icon {
  transform: rotate(15deg);
}

.theme-label {
  font-size: 12px;
  font-weight: 500;
  opacity: 0.8;
}

.notification-dot {
  position: absolute;
  top: 6px;
  right: 6px;
  width: 6px;
  height: 6px;
  background: var(--nm-accent);
  border-radius: 50%;
  animation: notification-pulse 2s infinite;
}

@keyframes notification-pulse {
  0%, 100% { opacity: 1; transform: scale(1); }
  50% { opacity: 0.7; transform: scale(1.2); }
}

/* 左侧神经突触面板 */
/* 神经元链样式 */
.neuron-chain {
  position: absolute;
  left: -40px;
  top: 50%;
  transform: translateY(-50%);
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 24px;
  z-index: 10;
}

.neuron-node {
  position: relative;
  width: 24px;
  height: 24px;
  cursor: pointer;
  transition: all 0.3s ease;
}

.neuron-node:hover {
  transform: scale(1.2);
}

.neuron-core {
  width: 100%;
  height: 100%;
  background: var(--nm-text-secondary);
  border-radius: 50%;
  position: relative;
  z-index: 2;
  transition: all 0.3s ease;
}

.neuron-node:hover .neuron-core {
  background: var(--nm-primary);
  box-shadow: 0 0 12px var(--nm-primary-glow);
}

.neuron-active .neuron-core {
  /* 已打开的神经元链节点：无特殊样式 */
  /* 背景色和阴影保持默认 */
}

.neuron-connected .neuron-core {
  background: var(--nm-secondary);
}

.neuron-pulse {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  width: 32px;
  height: 32px;
  border: 2px solid var(--nm-primary);
  border-radius: 50%;
  animation: neuron-ring 2s infinite;
  z-index: 1;
}

.neuron-connection {
  position: absolute;
  bottom: -28px;
  left: 50%;
  transform: translateX(-50%);
  width: 2px;
  height: 20px;
  background: linear-gradient(to bottom, var(--nm-primary-glow), transparent);
}

@keyframes neuron-pulse {
  0%, 100% { transform: scale(1); }
  50% { transform: scale(1.1); }
}

@keyframes neuron-ring {
  0% { transform: translate(-50%, -50%) scale(0.8); opacity: 1; }
  100% { transform: translate(-50%, -50%) scale(1.5); opacity: 0; }
}

.neuro-synapse-panel {
  position: fixed;
  top: 80px;
  left: 0;
  bottom: 0;
  width: 280px;
  background: var(--nm-bg-card);
  backdrop-filter: blur(20px);
  border-right: 1px solid var(--nm-border-glow);
  display: flex;
  flex-direction: column;
  transition: background-color 0.3s ease, border-color 0.3s ease;
  z-index: 50;
}

.neuro-command-layout[data-theme="light"] .neuro-synapse-panel {
  --nm-bg-deep: #0a0c14;
  --nm-bg-surface: #111827;
  --nm-bg-card: #151b2a;
  --nm-bg-elevated: #1f2937;
  --nm-primary: #00f5d4;
  --nm-primary-glow: rgba(0, 245, 212, 0.3);
  --nm-secondary: #9d4edd;
  --nm-secondary-glow: rgba(157, 78, 221, 0.3);
  --nm-accent: #ff6b6b;
  --nm-accent-glow: rgba(255, 107, 107, 0.3);
  --nm-text-primary: #f8fafc;
  --nm-text-secondary: #94a3b8;
  --nm-text-muted: #64748b;
  --nm-text-on-dark: #0a0c14;
  --nm-border: #2d3748;
  --nm-border-glow: rgba(0, 245, 212, 0.14);
  background:
    radial-gradient(circle at 16% 12%, rgba(0, 245, 212, 0.10), transparent 30%),
    linear-gradient(180deg, #151b2a 0%, #111827 100%);
  color: #f8fafc;
  border-right-color: rgba(0, 245, 212, 0.16);
}

.neuro-command-layout[data-theme="light"] .neuro-synapse-panel .neuro-btn {
  color: var(--nm-text-secondary);
}

.neuro-command-layout[data-theme="light"] .neuro-synapse-panel .neuro-btn:hover {
  color: var(--nm-primary);
}

.panel-expanded {
  background: var(--nm-bg-card);
}


.synapse-header {
  padding: 20px 20px 16px;
  border-bottom: 1px solid var(--nm-border-glow);
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.neuro-command-layout[data-theme="light"] .synapse-header {
  border-top: 0;
  border-bottom-color: transparent;
  box-shadow: none;
}

.synapse-header-right {
  flex-shrink: 0;
}

.synapse-back {
  padding: 6px 10px;
  font-size: 12px;
  display: flex;
  align-items: center;
  gap: 4px;
}

.synapse-back .back-label {
  font-size: 12px;
  font-weight: 500;
}

.shortcut-edit-trigger {
  min-width: 32px;
  min-height: 32px;
  padding: 6px 8px;
}

.shortcut-edit-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.shortcut-edit-cancel,
.shortcut-edit-save {
  min-height: 32px;
  padding: 6px 10px;
  font-size: 12px;
}

.shortcut-edit-save {
  border-color: rgba(0, 245, 212, 0.45);
  color: var(--nm-primary);
}

.synapse-title {
  font-family: 'JetBrains Mono', monospace;
  font-size: 15px;
  font-weight: 600;
  color: var(--nm-primary);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.synapse-content {
  flex: 1;
  min-height: 0;
  padding: 16px 0;
  overflow-y: auto;
  overflow-x: hidden;
  scrollbar-color: transparent transparent;
  scrollbar-width: thin;
}

/* 侧栏内部滚动条默认隐身，滚动中再浮现，避免抢主区视觉 */
.synapse-content::-webkit-scrollbar {
  width: 8px;
}

.synapse-content::-webkit-scrollbar-thumb {
  background: transparent;
  border-radius: 4px;
  transition: background-color 0.18s ease;
}

.synapse-content::-webkit-scrollbar-thumb:hover {
  background: rgba(0, 245, 212, 0.78);
}

.synapse-content::-webkit-scrollbar-track {
  background: transparent;
}

.synapse-content.is-scrolling {
  scrollbar-color: rgba(0, 245, 212, 0.62) transparent;
}

.synapse-content.is-scrolling::-webkit-scrollbar-thumb {
  background: rgba(0, 245, 212, 0.62);
}

/* 快捷入口 */
.quick-entries {
  display: flex;
  flex-direction: column;
  gap: 4px;
  padding: 0 12px;
}

.quick-entry {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  border-radius: 10px;
  cursor: pointer;
  transition: all 0.2s ease;
  position: relative;
}

.quick-entry:hover {
  background: var(--nm-primary-glow);
  transform: translateX(4px);
}

/* 快捷入口选中状态 - 当前页面 */
.quick-entry-active {
  background: var(--nm-primary-glow);
  border-left: 3px solid var(--nm-primary);
  padding-left: 9px; /* 12px - 3px边框 */
}

.quick-entry-active .entry-title {
  color: var(--nm-primary);
  font-weight: 600;
}

.quick-entry-active .entry-icon {
  color: var(--nm-primary);
}

.entry-icon {
  font-size: 14px;
  font-weight: 600;
  width: 28px;
  height: 28px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--nm-primary-glow);
  border-radius: 6px;
  border: 1px solid var(--nm-border-glow);
  color: var(--nm-text-secondary);
}

.entry-info {
  flex: 1;
  min-width: 0;
}

.entry-title {
  font-size: 13px;
  font-weight: 500;
  color: var(--nm-text-primary);
  margin-bottom: 2px;
}

.entry-desc {
  font-size: 11px;
  color: var(--nm-text-secondary);
  line-height: 1.3;
}

.entry-status {
  padding-left: 8px;
}

/* 二级菜单 */
/* 二级神经元菜单样式 */
.neuron-menu {
  padding: 16px 12px;
}

.neuron-menu-container {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.shortcut-draggable-list {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.shortcut-edit-tip {
  margin: 0 0 10px;
  padding: 8px 10px 8px 12px;
  font-size: 12px;
  font-weight: 500;
  color: var(--nm-text-primary);
  background: rgba(0, 245, 212, 0.08);
  border: 1px solid rgba(0, 245, 212, 0.28);
  border-left: 3px solid var(--nm-primary);
  border-radius: 8px;
  box-shadow: inset 0 0 0 1px rgba(0, 245, 212, 0.08);
}

.neuron-menu-item {
  display: flex;
  align-items: center;
  padding: 12px;
  border-radius: 12px;
  cursor: pointer;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  position: relative;
  overflow: hidden;
  animation: neuron-appear 0.5s ease-out var(--neuron-delay) both;
}

.neuron-menu-item:hover {
  background: rgba(0, 245, 212, 0.05);
  transform: translateX(4px);
}

.shortcut-editing-item {
  cursor: grab;
  border: 1px solid transparent;
}

.shortcut-editing-item:hover {
  background: rgba(0, 245, 212, 0.08);
  border-color: rgba(0, 245, 212, 0.2);
  transform: none;
}

.shortcut-editing-item:active {
  cursor: grabbing;
}

.shortcut-drag-over {
  border-color: rgba(0, 245, 212, 0.45);
  box-shadow: 0 0 0 1px rgba(0, 245, 212, 0.18), inset 0 0 0 1px rgba(0, 245, 212, 0.1);
}

.shortcut-drag-hint {
  margin-left: auto;
  color: var(--nm-text-secondary);
  opacity: 0.7;
  letter-spacing: 1px;
  font-size: 12px;
}

.shortcut-drag-ghost {
  opacity: 0.45;
}

.shortcut-drag-chosen {
  border: 1px solid rgba(0, 245, 212, 0.45);
  box-shadow: 0 0 0 1px rgba(0, 245, 212, 0.16), 0 6px 16px rgba(0, 245, 212, 0.2);
}

.shortcut-dragging {
  border: 1px solid rgba(0, 245, 212, 0.6);
  box-shadow: 0 10px 20px rgba(0, 245, 212, 0.2);
}

.neuron-active {
  /* 已打开页面：无括号（无左边框），无底纹 */
  /* 什么都不加 */
}

.neuron-current {
  /* 当前激活的神经元：有括号（左边框）和底纹（背景色） */
  background: rgba(0, 245, 212, 0.12);
}

@keyframes neuron-appear {
  from {
    opacity: 0;
    transform: translateX(-20px);
  }
  to {
    opacity: 1;
    transform: translateX(0);
  }
}

/* 神经元核心样式 - 缩小尺寸 */
.neuron-core-wrapper {
  position: relative;
  margin-right: 12px;
  flex-shrink: 0;
}

.neuron-core {
  position: relative;
  width: 28px;
  height: 28px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.core-inner {
  width: 12px;
  height: 12px;
  background: var(--nm-text-secondary);
  border-radius: 50%;
  transition: all 0.3s ease;
  z-index: 2;
  position: relative;
}

.neuron-menu-item:hover .core-inner {
  background: var(--nm-primary);
  transform: scale(1.3);
  box-shadow:
    0 0 8px var(--nm-primary-glow),
    inset 0 0 4px rgba(255, 255, 255, 0.3);
}

.neuron-active .core-inner {
  /* 打开的神经元：核心保持普通状态，无特殊样式 */
  /* 背景色和阴影保持默认 */
}

.neuron-current .core-inner {
  background: var(--nm-primary);
  box-shadow:
    0 0 16px var(--nm-primary-glow),
    0 0 24px var(--nm-primary-glow),
    inset 0 0 12px rgba(255, 255, 255, 0.6);
  animation: brain-expand 2s infinite cubic-bezier(0.4, 0, 0.2, 1);
}

@keyframes brain-expand {
  0%, 100% {
    transform: scale(1);
    box-shadow:
      0 0 12px var(--nm-primary-glow),
      0 0 20px var(--nm-primary-glow),
      inset 0 0 8px rgba(255, 255, 255, 0.5);
  }
  50% {
    transform: scale(1.4);
    box-shadow:
      0 0 20px var(--nm-primary-glow),
      0 0 32px var(--nm-primary-glow),
      inset 0 0 16px rgba(255, 255, 255, 0.7);
  }
}

.core-glow {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  width: 20px;
  height: 20px;
  background: var(--nm-primary);
  border-radius: 50%;
  filter: blur(4px);
  opacity: 0.4;
}

.core-pulse {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  width: 32px;
  height: 32px;
  border: 1px solid var(--nm-primary);
  border-radius: 50%;
  animation: pulse-ring 2s infinite;
}

@keyframes core-pulse {
  0%, 100% { transform: scale(1); }
  50% { transform: scale(1.1); }
}

@keyframes pulse-ring {
  0% { transform: translate(-50%, -50%) scale(0.8); opacity: 1; }
  100% { transform: translate(-50%, -50%) scale(1.5); opacity: 0; }
}

/* 神经元连接线 - 增强连接感 */
.neuron-connection {
  position: absolute;
  bottom: -10px;
  left: 50%;
  transform: translateX(-50%);
  width: 1px;
  height: 10px;
  z-index: 1;
}

.connection-line {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background: linear-gradient(to bottom,
    var(--nm-primary) 0%,
    rgba(0, 245, 212, 0.6) 30%,
    rgba(0, 245, 212, 0.3) 70%,
    transparent 100%
  );
}

.connection-glow {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background: var(--nm-primary);
  filter: blur(1px);
  opacity: 0.4;
  animation: connection-flow 2s infinite;
}

/* 激活状态下的连接线 */
.neuron-current + .neuron-menu-item .neuron-connection .connection-line {
  background: linear-gradient(to bottom,
    var(--nm-primary) 0%,
    rgba(0, 245, 212, 0.8) 50%,
    rgba(0, 245, 212, 0.4) 80%,
    transparent 100%
  );
}

.neuron-current + .neuron-menu-item .neuron-connection .connection-glow {
  opacity: 0.6;
  animation: connection-flow 1.5s infinite;
}

@keyframes connection-flow {
  0%, 100% {
    opacity: 0.3;
    transform: scaleY(1);
  }
  50% {
    opacity: 0.7;
    transform: scaleY(1.1);
  }
}

/* 神经元信息 */
.neuron-info {
  flex: 1;
  min-width: 0;
  display: flex;
  align-items: center;
  gap: 4px;
}

.neuron-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--nm-text-primary);
  transition: all 0.3s ease;
}

.neuron-menu-item:hover .neuron-title {
  color: var(--nm-primary);
}

.neuron-current .neuron-title {
  color: var(--nm-primary);
  text-shadow: 0 0 6px var(--nm-primary-glow);
}

.neuron-desc {
  font-size: 12px;
  color: var(--nm-text-secondary);
  line-height: 1.3;
}

/* Dock 指示器 — 已打开页面的标记点 */
.dock-indicator {
  width: 24px;
  height: 24px;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  position: relative;
  transition: all 0.2s ease;
  user-select: none;
  flex-shrink: 0;
  margin-left: auto;
}

.indicator-dot {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  transition: all 0.2s ease;
  position: relative;
}

/* 已打开但未激活：灰色静止点 */
.dock-indicator:not(.is-active) .indicator-dot {
  background: var(--nm-text-muted);
  opacity: 0.5;
}

/* 当前激活页面：主题色跳动点 */
.dock-indicator.is-active .indicator-dot {
  background: var(--nm-primary);
  opacity: 1;
  box-shadow: 0 0 8px var(--nm-primary-glow);
  animation: dock-bounce 2s ease-in-out infinite;
}

@keyframes dock-bounce {
  0%, 100% { transform: scale(1); }
  50% { transform: scale(1.3); }
}

/* 悬停：点 → 红色关闭 X */
.dock-indicator:hover .indicator-dot {
  opacity: 0;
  transform: scale(0);
}

.dock-indicator::after {
  content: '×';
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%) scale(0);
  font-size: 16px;
  font-weight: 600;
  line-height: 1;
  opacity: 0;
  transition: all 0.2s ease;
}

.dock-indicator:hover::after {
  opacity: 1;
  transform: translate(-50%, -50%) scale(1);
}

/* 非激活页面悬停：浅灰 X */
.dock-indicator:not(.is-active):hover::after {
  color: var(--nm-text-secondary);
}

/* 激活页面悬停：红色 X */
.dock-indicator.is-active:hover::after {
  color: var(--nm-error);
}

/* 浅色模式 */
.neuro-command-layout[data-theme="light"] .dock-indicator:not(.is-active) .indicator-dot {
  background: #94a3b8;
}

.neuro-command-layout[data-theme="light"] .dock-indicator.is-active .indicator-dot {
  background: #0ea5e9;
  box-shadow: 0 0 8px rgba(14, 165, 233, 0.2);
}

.neuro-command-layout[data-theme="light"] .dock-indicator:not(.is-active):hover::after {
  color: #64748b;
}

.neuro-command-layout[data-theme="light"] .dock-indicator.is-active:hover::after {
  color: #ef4444;
}

/* 神经身份节点 */
.neuro-identity-node {
  padding: 16px 20px;
  border-top: 1px solid rgba(0, 245, 212, 0.1);
  display: flex;
  align-items: center;
  gap: 14px;
  background: var(--nm-bg-card);
  transition: all 0.3s ease;
}

/* 可点击的神经身份节点 */
.neuro-identity-node.clickable {
  cursor: pointer;
}

.neuro-identity-node.clickable:hover {
  background: rgba(0, 245, 212, 0.05);
  transform: translateX(4px);
}

.identity-avatar {
  position: relative;
  width: 60px;
  height: 60px;
}

.avatar-glow {
  position: absolute;
  top: -4px;
  left: -4px;
  right: -4px;
  bottom: -4px;
  background: linear-gradient(135deg, var(--nm-primary), var(--nm-secondary));
  border-radius: 50%;
  filter: blur(6px);
  opacity: 0.4;
  animation: avatar-brain-pulse 2.5s infinite cubic-bezier(0.4, 0, 0.2, 1);
}

@keyframes avatar-brain-pulse {
  0%, 100% {
    opacity: 0.3;
    transform: scale(1);
    filter: blur(6px);
  }
  25% {
    opacity: 0.5;
    transform: scale(1.05);
    filter: blur(8px);
  }
  50% {
    opacity: 0.6;
    transform: scale(1.1);
    filter: blur(10px);
  }
  75% {
    opacity: 0.5;
    transform: scale(1.05);
    filter: blur(8px);
  }
}

.avatar-initial {
  position: relative;
  width: 100%;
  height: 100%;
  background: linear-gradient(135deg, var(--nm-primary), var(--nm-secondary));
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-family: 'JetBrains Mono', monospace;
  font-size: 14px;
  font-weight: 600;
  color: var(--nm-text-on-dark);
  z-index: 2;
  animation: avatar-inner-pulse 3s infinite;
}

@keyframes avatar-inner-pulse {
  0%, 100% {
    transform: scale(1);
    box-shadow:
      0 0 0 0 rgba(0, 245, 212, 0.4),
      inset 0 0 10px rgba(255, 255, 255, 0.1);
  }
  50% {
    transform: scale(1.05);
    box-shadow:
      0 0 20px 5px rgba(0, 245, 212, 0.6),
      inset 0 0 15px rgba(255, 255, 255, 0.2);
  }
}

.identity-info {
  flex: 1;
  min-width: 0;
}

.identity-name {
  font-size: 14px;
  font-weight: 600;
  color: var(--nm-text-primary);
  margin-bottom: 2px;
}

.identity-role {
  font-size: 12px;
  color: var(--nm-text-secondary);
}

.identity-settings {
  margin-left: auto;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
}

/* 侧边栏内头像缩小 */
.neuro-identity-node :deep(.avatar-wrapper) {
  width: 60px !important;
  height: 60px !important;
  border-width: 2px;
}

.neuro-identity-node :deep(.default-initials) {
  font-size: 20px;
}

.settings-arrow {
  font-size: 16px;
  color: var(--nm-text-secondary);
  transition: all 0.3s ease;
}

.neuro-identity-node.clickable:hover .settings-arrow {
  color: var(--nm-primary);
  transform: translateX(4px);
}

.logout-btn {
  background: transparent;
  border: 1px solid rgba(255, 87, 87, 0.3);
  border-radius: 50%;
  width: 24px;
  height: 24px;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: all 0.3s ease;
  padding: 0;
  color: #ff5757;
}

.logout-btn:hover {
  background: rgba(255, 87, 87, 0.15);
  border-color: rgba(255, 87, 87, 0.6);
  transform: scale(1.1);
}

.logout-icon {
  width: 12px;
  height: 12px;
}

/* 内容区域：至少占满「可视高度 − 顶栏占位」保证短页不镂空；长页由内容继续撑高 command-layout */
.neuro-content-field {
  min-height: calc(100vh - 80px);
  min-height: calc(100dvh - 80px);
  background: var(--nm-bg-surface);
  backdrop-filter: blur(10px);
  display: flex;
  flex-direction: column;
}

.neuro-command-layout[data-theme="light"] .neuro-content-field,
.neuro-command-layout[data-theme="light"] .content-main {
  background: #f3f6fb;
  backdrop-filter: none;
}

.neuro-command-layout[data-theme="light"] .content-header {
  background: rgba(255, 255, 255, 0.86);
  border-bottom-color: #d7e0ec;
  box-shadow: 0 1px 0 rgba(15, 23, 42, 0.03);
}

.neuro-command-layout[data-theme="light"] .breadcrumb-divider {
  background: #d7e0ec;
}

.neuro-command-layout.is-integration-center-route:not([data-theme="light"]) .neuro-bg-layer {
  display: none;
}

.neuro-command-layout.is-integration-center-route:not([data-theme="light"]) .neuro-content-field {
  background: #0f172a;
  backdrop-filter: none;
}

.neuro-command-layout.is-integration-center-route:not([data-theme="light"]) .content-main {
  background: #0f172a;
  background-image: none;
}

.neuro-command-layout.is-integration-center-route[data-theme="light"] .neuro-content-field,
.neuro-command-layout.is-integration-center-route[data-theme="light"] .content-main {
  background: #f3f6fb !important;
  background-image: none !important;
}

.neuro-content-field > .content-header {
  flex-shrink: 0;
}

.content-header {
  padding: 16px 24px;
  border-bottom: 1px solid var(--nm-border-glow);
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.content-breadcrumb {
  display: flex;
  align-items: center;
  gap: 8px;
  font-family: 'JetBrains Mono', monospace;
  font-size: 13px;
}

.breadcrumb-item {
  color: var(--nm-text-secondary);
  transition: color 0.2s;
}

.breadcrumb-item.breadcrumb-home:hover {
  color: var(--nm-primary);
  cursor: pointer;
}

/* 首页药丸按钮 — 呼吸发光，一眼可点 */
.home-pill {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 6px 14px;
  background: linear-gradient(135deg, rgba(0, 245, 212, 0.12), rgba(157, 78, 221, 0.12));
  border: 1px solid rgba(0, 245, 212, 0.35);
  border-radius: 20px;
  cursor: pointer;
  font-family: 'JetBrains Mono', monospace;
  font-size: 12px;
  font-weight: 600;
  color: var(--nm-primary);
  letter-spacing: 0.05em;
  transition: all 0.25s cubic-bezier(0.4, 0, 0.2, 1);
  position: relative;
  animation: home-pill-breathe 3s ease-in-out infinite;
  user-select: none;
}

.home-pill:hover {
  background: linear-gradient(135deg, rgba(0, 245, 212, 0.25), rgba(157, 78, 221, 0.25));
  border-color: var(--nm-primary);
  box-shadow: 0 0 16px rgba(0, 245, 212, 0.3), inset 0 0 8px rgba(0, 245, 212, 0.1);
  transform: translateY(-1px);
  animation-play-state: paused;
}

.home-pill:active {
  transform: translateY(0) scale(0.97);
  box-shadow: 0 0 8px rgba(0, 245, 212, 0.2);
}

.home-icon {
  width: 14px;
  height: 14px;
  flex-shrink: 0;
  transition: transform 0.25s ease;
}

.home-pill:hover .home-icon {
  transform: scale(1.15) rotate(-5deg);
}

.home-label {
  white-space: nowrap;
}

@keyframes home-pill-breathe {
  0%, 100% {
    border-color: rgba(0, 245, 212, 0.3);
    box-shadow: 0 0 0 rgba(0, 245, 212, 0);
  }
  50% {
    border-color: rgba(0, 245, 212, 0.55);
    box-shadow: 0 0 12px rgba(0, 245, 212, 0.2), inset 0 0 6px rgba(0, 245, 212, 0.05);
  }
}

/* 浅色模式 */
.neuro-command-layout[data-theme="light"] .home-pill {
  background: linear-gradient(135deg, rgba(14, 165, 233, 0.1), rgba(139, 92, 246, 0.1));
  border-color: rgba(14, 165, 233, 0.3);
  color: #0ea5e9;
}

.neuro-command-layout[data-theme="light"] .home-pill:hover {
  border-color: #0ea5e9;
  box-shadow: 0 0 12px rgba(14, 165, 233, 0.2), inset 0 0 6px rgba(14, 165, 233, 0.05);
}

@keyframes home-pill-breathe {
  0%, 100% {
    border-color: rgba(14, 165, 233, 0.25);
    box-shadow: 0 0 0 rgba(14, 165, 233, 0);
  }
  50% {
    border-color: rgba(14, 165, 233, 0.5);
    box-shadow: 0 0 10px rgba(14, 165, 233, 0.15), inset 0 0 4px rgba(14, 165, 233, 0.03);
  }
}

.breadcrumb-item:hover {
  color: var(--nm-primary);
  cursor: pointer;
}

.breadcrumb-item.active {
  color: var(--nm-primary);
  font-weight: 600;
}

.breadcrumb-separator {
  color: var(--nm-text-muted);
}

.breadcrumb-divider {
  color: var(--nm-text-muted);
  margin: 0 12px;
  display: inline-block;
  width: 1px;
  height: 18px;
  background: rgba(0, 245, 212, 0.2);
  flex-shrink: 0;
  font-size: 0;
  line-height: 0;
}

.content-actions {
  display: flex;
  gap: 8px;
  align-items: center;
}

.matrix-entry-btn .btn-icon {
  font-size: 18px;
  font-weight: 800;
  line-height: 1;
}

.matrix-overlay {
  position: fixed;
  inset: 0;
  background: rgba(5, 10, 24, 0.62);
  backdrop-filter: blur(6px);
  z-index: 210;
  display: flex;
  justify-content: flex-end;
  align-items: stretch;
}

.matrix-panel {
  width: min(980px, 92vw);
  height: 100dvh;
  min-height: 0;
  background: var(--nm-bg-card);
  border-left: 1px solid var(--nm-border-glow);
  display: flex;
  flex-direction: column;
  box-shadow: -18px 0 42px rgba(0, 0, 0, 0.45);
  animation: matrix-drawer-enter 220ms cubic-bezier(0.2, 0.8, 0.2, 1);
}

@keyframes matrix-drawer-enter {
  from {
    transform: translateX(40px);
    opacity: 0.8;
  }
  to {
    transform: translateX(0);
    opacity: 1;
  }
}

.matrix-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  padding: 18px 20px 12px;
  border-bottom: 1px solid var(--nm-border-glow);
}

.matrix-title {
  margin: 0;
  font-size: 20px;
  color: var(--nm-text-primary);
}

.matrix-subtitle {
  margin: 6px 0 0;
  font-size: 13px;
  color: var(--nm-text-secondary);
}

.matrix-toolbar {
  padding: 12px 20px;
  border-bottom: 1px solid var(--nm-border-glow);
}

.matrix-search {
  width: 100%;
  border: 1px solid var(--nm-border);
  background: var(--nm-bg-surface);
  color: var(--nm-text-primary);
  border-radius: 10px;
  padding: 10px 12px;
  font-size: 14px;
  outline: none;
}

.matrix-search:focus {
  border-color: var(--nm-primary);
}

.matrix-body {
  flex: 1;
  min-height: 0;
  overflow: auto;
  padding: 16px 20px 20px;
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.matrix-group-title {
  margin: 0 0 10px;
  font-size: 14px;
  color: var(--nm-text-secondary);
}

.matrix-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));
  gap: 10px;
}

.matrix-card {
  border: 1px solid var(--nm-border);
  border-radius: 12px;
  background: var(--nm-bg-surface);
  padding: 12px;
  text-align: left;
  color: var(--nm-text-primary);
  cursor: pointer;
  transition: all 0.2s ease;
}

.matrix-card:hover {
  transform: translateY(-1px);
  border-color: var(--nm-primary);
  box-shadow: 0 8px 20px rgba(0, 245, 212, 0.15);
}

.matrix-card.is-current {
  border-color: var(--nm-primary);
  box-shadow: inset 0 0 0 1px rgba(0, 245, 212, 0.35);
}

.matrix-card.is-locked {
  cursor: not-allowed;
  background: rgba(148, 163, 184, 0.08);
  color: var(--nm-text-secondary);
  border-color: rgba(148, 163, 184, 0.18);
}

.matrix-card.is-locked:hover {
  transform: none;
  box-shadow: none;
  border-color: rgba(148, 163, 184, 0.24);
}

.matrix-card-head {
  display: flex;
  justify-content: space-between;
  gap: 8px;
  align-items: center;
}

.matrix-card-title {
  font-size: 14px;
  font-weight: 600;
}

.matrix-card-badge {
  font-size: 11px;
  border-radius: 999px;
  padding: 2px 8px;
  background: rgba(0, 245, 212, 0.14);
  color: var(--nm-primary);
  border: 1px solid rgba(0, 245, 212, 0.2);
}

.matrix-card.is-locked .matrix-card-badge {
  background: rgba(148, 163, 184, 0.15);
  color: #94a3b8;
  border-color: rgba(148, 163, 184, 0.25);
}

.matrix-card-path {
  margin-top: 8px;
  font-size: 12px;
  color: var(--nm-text-secondary);
  word-break: break-all;
}

.matrix-empty {
  color: var(--nm-text-secondary);
  font-size: 14px;
  padding: 20px 8px;
}

.page-controls {
  display: flex;
  gap: 8px;
  margin-right: 16px;
  padding-right: 16px;
  position: relative;
}

.page-controls::after {
  content: '';
  position: absolute;
  right: 0;
  top: 50%;
  transform: translateY(-50%);
  width: 1px;
  height: 75%;
  background: rgba(0, 245, 212, 0.2);
}

.content-main {
  /* 高度随路由页真实内容；不参与 flex-grow，避免在「已很高的内容」下再叠一段空白可滚区 */
  flex: 0 1 auto;
  min-width: 0;
  padding: 24px;
  display: flex;
  flex-direction: column;
  background: var(--nm-bg-surface);
}

/* 与顶栏一致：业务页根节点勿再叠一层横向内边距 */
.content-main :deep(.page) {
  flex: 0 1 auto;
  display: flex;
  flex-direction: column;
  padding-left: 0 !important;
  padding-right: 0 !important;
}

/* 主体管理等使用 NeuroAgentPageShell 的根容器（非 .page） */
.content-main :deep(.tenant-page) {
  flex: 0 1 auto;
  min-width: 0;
  min-height: 0;
  display: flex;
  flex-direction: column;
}

.content-main :deep(.home) {
  max-width: none;
  margin-left: 0;
  margin-right: 0;
}

/* 页面内容样式 */
.page-content {
  max-width: none;
  margin: 0;
}

.page-header {
  margin-bottom: 32px;
}

.page-title {
  font-family: 'JetBrains Mono', monospace;
  font-size: 32px;
  font-weight: 600;
  color: var(--nm-text-primary);
  margin-bottom: 8px;
}

.page-subtitle {
  font-size: 16px;
  color: var(--nm-text-secondary);
}

.page-body {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 24px;
}

.page-card {
  background: var(--nm-bg-card);
  border: 1px solid var(--nm-border-glow);
  border-radius: 16px;
  padding: 24px;
  transition: all 0.3s ease;
}

.page-card:hover {
  border-color: var(--nm-primary);
  transform: translateY(-4px);
  box-shadow: 0 8px 24px rgba(0, 245, 212, 0.15);
}

.page-card h3 {
  font-size: 18px;
  font-weight: 600;
  color: var(--nm-primary);
  margin-bottom: 12px;
}

.page-card p {
  font-size: 14px;
  color: var(--nm-text-secondary);
  line-height: 1.6;
  margin-bottom: 8px;
}

.page-card ul {
  padding-left: 20px;
  margin: 12px 0;
}

.page-card li {
  font-size: 14px;
  color: var(--nm-text-secondary);
  line-height: 1.6;
  margin-bottom: 6px;
}

.page-card li::marker {
  color: var(--nm-primary);
}

/* 关闭控制样式 */
/* 刷新按钮 — 关闭按钮左侧 */
.page-refresh-btn .btn-icon {
  font-size: 16px;
  transition: color 0.2s ease;
}

.page-refresh-btn:hover .btn-icon {
  color: var(--nm-primary);
}

.shortcut-toggle-btn .btn-icon {
  font-size: 14px;
  transition: all 0.2s ease;
}

.shortcut-toggle-btn:hover .btn-icon {
  color: var(--nm-primary);
  transform: scale(1.15);
}

.close-control-wrapper {
  position: relative;
  display: inline-block;
}

.close-control-btn {
  position: relative;
  transition: all 0.2s ease;
}

.close-control-btn.has-multiple {
  background: linear-gradient(135deg, var(--nm-primary), var(--nm-secondary)) !important;
  color: var(--nm-text-on-dark) !important;
  font-weight: 600;
  box-shadow: 0 2px 8px rgba(0, 245, 212, 0.2);
}

.close-control-btn.has-multiple:hover {
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(0, 245, 212, 0.3);
}

.close-control-btn.has-multiple .btn-icon {
  animation: pulse-glow 1.5s infinite;
}

@keyframes pulse-glow {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.7; }
}

.close-dropdown {
  position: absolute;
  top: 100%;
  right: 0;
  margin-top: 8px;
  min-width: 200px;
  background: var(--nm-bg-card);
  border: 1px solid var(--nm-border);
  border-radius: 12px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.3);
  opacity: 0;
  visibility: hidden;
  transform: translateY(-10px);
  transition: all 0.2s ease;
  z-index: 1000;
  overflow: hidden;
}

.close-control-wrapper:hover .close-dropdown {
  opacity: 1;
  visibility: visible;
  transform: translateY(0);
}

.close-dropdown-item {
  display: flex;
  align-items: center;
  width: 100%;
  padding: 12px 16px;
  background: transparent;
  border: none;
  color: var(--nm-text-primary);
  font-size: 14px;
  text-align: left;
  cursor: pointer;
  transition: all 0.2s ease;
}

.close-dropdown-item:hover {
  background: var(--nm-bg-surface);
  color: var(--nm-primary);
}

.close-dropdown-item.disabled {
  opacity: 0.5;
  cursor: not-allowed;
  pointer-events: none;
}

.close-dropdown-item.disabled:hover {
  background: transparent;
  color: var(--nm-text-primary);
}

.close-dropdown-item .dropdown-icon {
  margin-right: 10px;
  font-size: 16px;
  min-width: 20px;
  text-align: center;
}

.close-dropdown-divider {
  height: 1px;
  background: var(--nm-border);
  margin: 4px 0;
}

/* 响应式设计：固定侧栏宽度与主体 padding-left 必须同步收窄 */
@media (max-width: 1200px) {
  .neuro-interface {
    padding-left: 240px;
  }

  .neuro-command-layout[data-theme="light"] .neuro-stem-menu::before {
    width: 240px;
  }

  .neuro-command-layout[data-theme="light"] .neuro-stem-menu::after {
    width: 240px;
  }

  .neuro-synapse-panel {
    width: 240px;
  }

  .stem-nav {
    margin: 0 16px 0 36px;
  }

  .stem-item {
    min-width: 60px;
    padding: 8px 8px;
  }
}

@media (max-width: 1024px) {
  .neuro-interface {
    padding-left: 200px;
  }

  .neuro-command-layout[data-theme="light"] .neuro-stem-menu::before {
    width: 200px;
  }

  .neuro-command-layout[data-theme="light"] .neuro-stem-menu::after {
    width: 200px;
  }

  .neuro-synapse-panel {
    width: 200px;
  }

  .stem-brand {
    margin-right: 16px;
  }

  .stem-nav {
    margin: 0 12px 0 24px;
    gap: 2px;
  }

  .stem-item {
    min-width: 50px;
  }

  .stem-label {
    font-size: 11px;
  }
}

@media (max-width: 768px) {
  .neuro-interface {
    padding-left: 0;
  }

  .neuro-synapse-panel {
    display: none;
  }

  .content-header {
    align-items: flex-start;
    flex-direction: column;
    gap: 12px;
    padding: 12px 14px;
  }

  .content-breadcrumb,
  .content-actions {
    max-width: 100%;
    min-width: 0;
  }

  .content-breadcrumb,
  .content-actions {
    flex-wrap: wrap;
  }

  .content-main {
    padding: 14px;
  }
}
</style>

<style>
/* ============================================
   全局按钮系统 — 作用于 .neuro-command-layout 内容区
   必须用非 scoped 样式才能穿透 Element Plus
   ============================================ */

/* ── 1. 基础 ────────────────────────────── */
.neuro-command-layout .el-button {
  font-family: 'Inter', -apple-system, BlinkMacSystemFont, 'PingFang SC', 'Microsoft YaHei', sans-serif;
  font-weight: 500;
  letter-spacing: 0.02em;
  transition:
    background-color 0.25s ease,
    border-color 0.25s ease,
    color 0.25s ease,
    box-shadow 0.25s ease,
    transform 0.25s ease;
}

/* ── 2. 实心按钮尺寸 ────────────────────── */
.neuro-command-layout .el-button:not(.is-circle):not(.is-link):not(.is-text) {
  border-radius: 9px;
  padding: 10px 22px;
  min-height: 38px;
  box-sizing: border-box;
}
.neuro-command-layout .el-button--small:not(.is-circle):not(.is-link):not(.is-text) {
  font-size: 13px;
  padding: 8px 18px;
  min-height: 34px;
  border-radius: 8px;
}
.neuro-command-layout .el-button--large:not(.is-circle):not(.is-link):not(.is-text) {
  font-size: 15px;
  padding: 12px 28px;
  min-height: 44px;
  border-radius: 10px;
}

/* ── 3. 非表格区域按钮 ──────────────────── */

/* Primary — 青到紫渐变 */
.neuro-command-layout .el-button--primary:not(.is-link):not(.is-text) {
  --el-button-text-color: #fff;
  --el-button-hover-text-color: #fff;
  --el-button-active-text-color: #fff;
  --el-button-hover-bg-color: transparent;
  --el-button-active-bg-color: transparent;
  border: none;
  background: linear-gradient(135deg, #00d4aa 0%, #7c3aed 100%);
  box-shadow: 0 4px 16px 0 rgba(0, 212, 170, 0.3);
  color: #fff;
  border-radius: 12px;
}
.neuro-command-layout .el-button--primary:not(.is-link):not(.is-text):hover {
  border-color: transparent;
  box-shadow: 0 6px 24px 0 rgba(0, 212, 170, 0.4);
  transform: translateY(-1px);
  background: linear-gradient(135deg, #00e4ba 0%, #8b5cf6 100%);
}
.neuro-command-layout .el-button--primary:not(.is-link):not(.is-text):active {
  transform: translateY(0);
  box-shadow: 0 2px 8px 0 rgba(0, 212, 170, 0.3);
  background: linear-gradient(135deg, #00c49a 0%, #6d28d9 100%);
}

/* Default — 暗色表面 + 细边框 */
.neuro-command-layout .el-button--default:not(.is-link):not(.is-text) {
  background: var(--nm-bg-elevated);
  border: 1px solid var(--nm-border);
  color: var(--nm-text-primary);
  border-radius: 12px;
}
.neuro-command-layout .el-button--default:not(.is-link):not(.is-text):hover {
  border-color: var(--nm-primary);
  background: rgba(0, 245, 212, 0.06);
  color: var(--nm-primary);
}

/* Danger — 渐变红（非表格区域） */
.neuro-command-layout .el-button--danger:not(.is-link):not(.is-text) {
  background: linear-gradient(135deg, var(--nm-error), #dc2626);
  border: 1px solid rgba(239, 68, 68, 0.3);
  box-shadow: 0 4px 14px 0 rgba(239, 68, 68, 0.25);
  color: #fff;
  border-radius: 12px;
}
.neuro-command-layout .el-button--danger:not(.is-link):not(.is-text):hover {
  border-color: var(--nm-error);
  box-shadow: 0 6px 20px 0 rgba(239, 68, 68, 0.35);
  transform: translateY(-2px);
}

/* Success / Warning / Info */
.neuro-command-layout .el-button--success:not(.is-link):not(.is-text) {
  background: linear-gradient(135deg, var(--nm-success), #059669);
  border: 1px solid rgba(16, 185, 129, 0.3);
  box-shadow: 0 4px 14px 0 rgba(16, 185, 129, 0.25);
  color: #fff;
}
.neuro-command-layout .el-button--success:not(.is-link):not(.is-text):hover {
  box-shadow: 0 6px 20px 0 rgba(16, 185, 129, 0.35);
  transform: translateY(-2px);
}
.neuro-command-layout .el-button--warning:not(.is-link):not(.is-text) {
  background: linear-gradient(135deg, var(--nm-warning), #d97706);
  border: 1px solid rgba(245, 158, 11, 0.3);
  box-shadow: 0 4px 14px 0 rgba(245, 158, 11, 0.25);
  color: #fff;
}
.neuro-command-layout .el-button--warning:not(.is-link):not(.is-text):hover {
  box-shadow: 0 6px 20px 0 rgba(245, 158, 11, 0.35);
  transform: translateY(-2px);
}
.neuro-command-layout .el-button--info:not(.is-link):not(.is-text) {
  background: linear-gradient(135deg, var(--nm-info), #2563eb);
  border: 1px solid rgba(59, 130, 246, 0.3);
  box-shadow: 0 4px 14px 0 rgba(59, 130, 246, 0.25);
  color: #fff;
}
.neuro-command-layout .el-button--info:not(.is-link):not(.is-text):hover {
  box-shadow: 0 6px 20px 0 rgba(59, 130, 246, 0.35);
  transform: translateY(-2px);
}

/* Plain / Link */
.neuro-command-layout .el-button.is-plain {
  border: 1px solid var(--nm-primary);
  background: transparent;
  color: var(--nm-primary);
}
.neuro-command-layout .el-button.is-plain:hover {
  background: rgba(0, 245, 212, 0.08);
}
.neuro-command-layout .el-button.is-link {
  color: var(--nm-primary);
  padding: 4px 8px;
  min-height: auto;
}
.neuro-command-layout .el-button.is-link:hover { color: #00c9a7; }
.neuro-command-layout .el-button.is-link.el-button--primary { color: var(--nm-primary); }
.neuro-command-layout .el-button.is-link.el-button--primary:hover { color: #00c9a7; }
.neuro-command-layout .el-button.is-link.el-button--danger { color: var(--nm-error); }
.neuro-command-layout .el-button.is-link.el-button--danger:hover { color: #ff4444; }

/* ── 4. 表格内按钮（含固定列）────────────── */
/* 表格内 default 按钮：暗色表面，统一 28px 高 */
div.neuro-command-layout .el-table__body-wrapper .el-button--default:not(.is-link):not(.is-text),
div.neuro-command-layout .el-table__body .el-button--default:not(.is-link):not(.is-text),
div.neuro-command-layout .el-table__fixed .el-button--default:not(.is-link):not(.is-text),
div.neuro-command-layout .el-table__fixed-right .el-button--default:not(.is-link):not(.is-text),
div.neuro-command-layout .el-table__fixed-left .el-button--default:not(.is-link):not(.is-text) {
  --el-button-bg-color: var(--nm-bg-elevated) !important;
  --el-button-hover-bg-color: rgba(0, 245, 212, 0.06) !important;
  --el-button-text-color: var(--nm-text-primary) !important;
  --el-button-hover-text-color: var(--nm-primary) !important;
  --el-button-border-color: var(--nm-border) !important;
  --el-button-hover-border-color: var(--nm-primary) !important;
  background: var(--nm-bg-elevated) !important;
  border: 1px solid var(--nm-border) !important;
  color: var(--nm-text-primary) !important;
  box-shadow: none !important;
  transform: none !important;
  padding: 0 10px !important;
  height: 28px !important;
  line-height: 28px !important;
  border-radius: 12px !important;
  font-size: 12px !important;
}
div.neuro-command-layout .el-table__body-wrapper .el-button--default:not(.is-link):not(.is-text):hover,
div.neuro-command-layout .el-table__body .el-button--default:not(.is-link):not(.is-text):hover,
div.neuro-command-layout .el-table__fixed .el-button--default:not(.is-link):not(.is-text):hover,
div.neuro-command-layout .el-table__fixed-right .el-button--default:not(.is-link):not(.is-text):hover,
div.neuro-command-layout .el-table__fixed-left .el-button--default:not(.is-link):not(.is-text):hover {
  background: rgba(0, 245, 212, 0.06) !important;
  border-color: var(--nm-primary) !important;
  color: var(--nm-primary) !important;
}

/* 表格内 danger 按钮：默认暗色表面，hover 时边框变红 */
div.neuro-command-layout .el-table__body-wrapper .el-button--danger:not(.is-link):not(.is-text),
div.neuro-command-layout .el-table__body .el-button--danger:not(.is-link):not(.is-text),
div.neuro-command-layout .el-table__fixed .el-button--danger:not(.is-link):not(.is-text),
div.neuro-command-layout .el-table__fixed-right .el-button--danger:not(.is-link):not(.is-text),
div.neuro-command-layout .el-table__fixed-left .el-button--danger:not(.is-link):not(.is-text) {
  --el-button-bg-color: var(--nm-bg-elevated) !important;
  --el-button-hover-bg-color: rgba(239, 68, 68, 0.06) !important;
  --el-button-text-color: var(--nm-text-primary) !important;
  --el-button-hover-text-color: var(--nm-error) !important;
  --el-button-border-color: var(--nm-border) !important;
  --el-button-hover-border-color: var(--nm-error) !important;
  background: var(--nm-bg-elevated) !important;
  border: 1px solid var(--nm-border) !important;
  color: var(--nm-text-primary) !important;
  box-shadow: none !important;
  transform: none !important;
  padding: 0 10px !important;
  height: 28px !important;
  line-height: 28px !important;
  border-radius: 12px !important;
  font-size: 12px !important;
}
div.neuro-command-layout .el-table__body-wrapper .el-button--danger:not(.is-link):not(.is-text):hover,
div.neuro-command-layout .el-table__body .el-button--danger:not(.is-link):not(.is-text):hover,
div.neuro-command-layout .el-table__fixed .el-button--danger:not(.is-link):not(.is-text):hover,
div.neuro-command-layout .el-table__fixed-right .el-button--danger:not(.is-link):not(.is-text):hover,
div.neuro-command-layout .el-table__fixed-left .el-button--danger:not(.is-link):not(.is-text):hover {
  background: rgba(239, 68, 68, 0.06) !important;
  border-color: var(--nm-error) !important;
  color: var(--nm-error) !important;
}

/* 表格内 success / warning / info 按钮：统一暗色表面 */
div.neuro-command-layout .el-table__body-wrapper .el-button--success:not(.is-link):not(.is-text),
div.neuro-command-layout .el-table__body .el-button--success:not(.is-link):not(.is-text),
div.neuro-command-layout .el-table__fixed .el-button--success:not(.is-link):not(.is-text),
div.neuro-command-layout .el-table__fixed-right .el-button--success:not(.is-link):not(.is-text),
div.neuro-command-layout .el-table__fixed-left .el-button--success:not(.is-link):not(.is-text),
div.neuro-command-layout .el-table__body-wrapper .el-button--warning:not(.is-link):not(.is-text),
div.neuro-command-layout .el-table__body .el-button--warning:not(.is-link):not(.is-text),
div.neuro-command-layout .el-table__fixed .el-button--warning:not(.is-link):not(.is-text),
div.neuro-command-layout .el-table__fixed-right .el-button--warning:not(.is-link):not(.is-text),
div.neuro-command-layout .el-table__fixed-left .el-button--warning:not(.is-link):not(.is-text),
div.neuro-command-layout .el-table__body-wrapper .el-button--info:not(.is-link):not(.is-text),
div.neuro-command-layout .el-table__body .el-button--info:not(.is-link):not(.is-text),
div.neuro-command-layout .el-table__fixed .el-button--info:not(.is-link):not(.is-text),
div.neuro-command-layout .el-table__fixed-right .el-button--info:not(.is-link):not(.is-text),
div.neuro-command-layout .el-table__fixed-left .el-button--info:not(.is-link):not(.is-text) {
  --el-button-bg-color: var(--nm-bg-elevated) !important;
  --el-button-hover-bg-color: rgba(0, 245, 212, 0.06) !important;
  --el-button-text-color: var(--nm-text-primary) !important;
  --el-button-hover-text-color: var(--nm-primary) !important;
  --el-button-border-color: var(--nm-border) !important;
  --el-button-hover-border-color: var(--nm-primary) !important;
  background: var(--nm-bg-elevated) !important;
  border: 1px solid var(--nm-border) !important;
  color: var(--nm-text-primary) !important;
  box-shadow: none !important;
  transform: none !important;
  padding: 0 10px !important;
  height: 28px !important;
  line-height: 28px !important;
  border-radius: 12px !important;
  font-size: 12px !important;
}
div.neuro-command-layout .el-table__body-wrapper .el-button--success:not(.is-link):not(.is-text):hover,
div.neuro-command-layout .el-table__body .el-button--success:not(.is-link):not(.is-text):hover,
div.neuro-command-layout .el-table__fixed .el-button--success:not(.is-link):not(.is-text):hover,
div.neuro-command-layout .el-table__fixed-right .el-button--success:not(.is-link):not(.is-text):hover,
div.neuro-command-layout .el-table__fixed-left .el-button--success:not(.is-link):not(.is-text):hover,
div.neuro-command-layout .el-table__body-wrapper .el-button--warning:not(.is-link):not(.is-text):hover,
div.neuro-command-layout .el-table__body .el-button--warning:not(.is-link):not(.is-text):hover,
div.neuro-command-layout .el-table__fixed .el-button--warning:not(.is-link):not(.is-text):hover,
div.neuro-command-layout .el-table__fixed-right .el-button--warning:not(.is-link):not(.is-text):hover,
div.neuro-command-layout .el-table__fixed-left .el-button--warning:not(.is-link):not(.is-text):hover,
div.neuro-command-layout .el-table__body-wrapper .el-button--info:not(.is-link):not(.is-text):hover,
div.neuro-command-layout .el-table__body .el-button--info:not(.is-link):not(.is-text):hover,
div.neuro-command-layout .el-table__fixed .el-button--info:not(.is-link):not(.is-text):hover,
div.neuro-command-layout .el-table__fixed-right .el-button--info:not(.is-link):not(.is-text):hover,
div.neuro-command-layout .el-table__fixed-left .el-button--info:not(.is-link):not(.is-text):hover {
  background: rgba(0, 245, 212, 0.06) !important;
  border-color: var(--nm-primary) !important;
  color: var(--nm-primary) !important;
}

/* 表格内 primary 按钮：保留渐变（不受表格暗色规则覆盖） */
div.neuro-command-layout .el-table__body-wrapper .el-button--primary:not(.is-link):not(.is-text),
div.neuro-command-layout .el-table__body .el-button--primary:not(.is-link):not(.is-text),
div.neuro-command-layout .el-table__fixed .el-button--primary:not(.is-link):not(.is-text),
div.neuro-command-layout .el-table__fixed-right .el-button--primary:not(.is-link):not(.is-text),
div.neuro-command-layout .el-table__fixed-left .el-button--primary:not(.is-link):not(.is-text) {
  background: linear-gradient(135deg, #00d4aa 0%, #7c3aed 100%) !important;
  border: none !important;
  color: #fff !important;
  box-shadow: 0 4px 16px 0 rgba(0, 212, 170, 0.3) !important;
  transform: none !important;
  padding: 0 12px !important;
  height: 28px !important;
  line-height: 28px !important;
  border-radius: 12px !important;
  font-size: 12px !important;
}

/* ── 5. Dialog / Drawer 内按钮 ──────────── */
.neuro-command-layout .el-dialog .el-button--primary:not(.is-link):not(.is-text),
.neuro-command-layout .el-drawer .el-button--primary:not(.is-link):not(.is-text) {
  --el-button-text-color: #fff;
  --el-button-hover-text-color: #fff;
  --el-button-active-text-color: #fff;
  --el-button-hover-bg-color: transparent;
  --el-button-active-bg-color: transparent;
  border: none;
  background: linear-gradient(135deg, #00d4aa 0%, #7c3aed 100%);
  box-shadow: 0 4px 16px 0 rgba(0, 212, 170, 0.3);
  color: #fff;
  border-radius: 12px;
}
.neuro-command-layout .el-dialog .el-button--primary:not(.is-link):not(.is-text):hover,
.neuro-command-layout .el-drawer .el-button--primary:not(.is-link):not(.is-text):hover {
  border-color: transparent;
  box-shadow: 0 6px 24px 0 rgba(0, 212, 170, 0.4);
  transform: translateY(-1px);
  background: linear-gradient(135deg, #00e4ba 0%, #8b5cf6 100%);
}
.neuro-command-layout .el-dialog .el-button--default:not(.is-link):not(.is-text),
.neuro-command-layout .el-drawer .el-button--default:not(.is-link):not(.is-text) {
  background: var(--nm-bg-elevated);
  border: 1px solid var(--nm-border);
  color: var(--nm-text-primary);
  border-radius: 12px;
}
.neuro-command-layout .el-dialog .el-button--default:not(.is-link):not(.is-text):hover,
.neuro-command-layout .el-drawer .el-button--default:not(.is-link):not(.is-text):hover {
  border-color: var(--nm-primary);
  background: rgba(0, 245, 212, 0.06);
  color: var(--nm-primary);
}

/* ── 6. 浅色模式 ───────────────────────── */

/* 非表格 primary */
.neuro-command-layout[data-theme="light"] .el-button--primary:not(.is-link):not(.is-text) {
  border: none;
  background: linear-gradient(135deg, #00d4aa 0%, #7c3aed 100%);
  box-shadow: 0 4px 16px 0 rgba(0, 212, 170, 0.25);
  color: #fff;
  border-radius: 12px;
}
.neuro-command-layout[data-theme="light"] .el-button--primary:not(.is-link):not(.is-text):hover {
  box-shadow: 0 6px 24px 0 rgba(0, 212, 170, 0.35);
  transform: translateY(-1px);
  background: linear-gradient(135deg, #00e4ba 0%, #8b5cf6 100%);
}

/* 非表格 default */
.neuro-command-layout[data-theme="light"] .el-button--default:not(.is-link):not(.is-text) {
  background: #f8fafc;
  border-color: #cbd5e1;
  color: #1e293b;
  border-radius: 12px;
}
.neuro-command-layout[data-theme="light"] .el-button--default:not(.is-link):not(.is-text):hover {
  border-color: #0ea5e9;
  background: rgba(14, 165, 233, 0.06);
  color: #0ea5e9;
}

/* 非表格 danger */
.neuro-command-layout[data-theme="light"] .el-button--danger:not(.is-link):not(.is-text) {
  background: linear-gradient(135deg, #ef4444, #dc2626);
  border-color: rgba(239, 68, 68, 0.4);
  box-shadow: 0 4px 14px 0 rgba(239, 68, 68, 0.2);
  color: #fff;
}

/* Link / Plain 浅色模式 */
.neuro-command-layout[data-theme="light"] .el-button.is-link { color: #0ea5e9; }
.neuro-command-layout[data-theme="light"] .el-button.is-link:hover { color: #0284c7; }
.neuro-command-layout[data-theme="light"] .el-button.is-plain { border-color: #0ea5e9; color: #0ea5e9; }
.neuro-command-layout[data-theme="light"] .el-button.is-plain:hover { background: rgba(14, 165, 233, 0.08); }

/* 表格内 default 浅色模式 */
.neuro-command-layout[data-theme="light"] .el-table__body-wrapper .el-button--default:not(.is-link):not(.is-text),
.neuro-command-layout[data-theme="light"] .el-table__body .el-button--default:not(.is-link):not(.is-text),
.neuro-command-layout[data-theme="light"] .el-table__fixed .el-button--default:not(.is-link):not(.is-text),
.neuro-command-layout[data-theme="light"] .el-table__fixed-right .el-button--default:not(.is-link):not(.is-text),
.neuro-command-layout[data-theme="light"] .el-table__fixed-left .el-button--default:not(.is-link):not(.is-text) {
  background: #f8fafc !important;
  border: 1px solid #cbd5e1 !important;
  color: #1e293b !important;
  box-shadow: none !important;
  transform: none !important;
}
.neuro-command-layout[data-theme="light"] .el-table__body-wrapper .el-button--default:not(.is-link):not(.is-text):hover,
.neuro-command-layout[data-theme="light"] .el-table__body .el-button--default:not(.is-link):not(.is-text):hover,
.neuro-command-layout[data-theme="light"] .el-table__fixed .el-button--default:not(.is-link):not(.is-text):hover,
.neuro-command-layout[data-theme="light"] .el-table__fixed-right .el-button--default:not(.is-link):not(.is-text):hover,
.neuro-command-layout[data-theme="light"] .el-table__fixed-left .el-button--default:not(.is-link):not(.is-text):hover {
  border-color: #0ea5e9 !important;
  color: #0ea5e9 !important;
  background: rgba(14, 165, 233, 0.06) !important;
}

/* 表格内 danger 浅色模式 */
.neuro-command-layout[data-theme="light"] .el-table__body-wrapper .el-button--danger:not(.is-link):not(.is-text),
.neuro-command-layout[data-theme="light"] .el-table__body .el-button--danger:not(.is-link):not(.is-text),
.neuro-command-layout[data-theme="light"] .el-table__fixed .el-button--danger:not(.is-link):not(.is-text),
.neuro-command-layout[data-theme="light"] .el-table__fixed-right .el-button--danger:not(.is-link):not(.is-text),
.neuro-command-layout[data-theme="light"] .el-table__fixed-left .el-button--danger:not(.is-link):not(.is-text) {
  background: #f8fafc !important;
  border: 1px solid #cbd5e1 !important;
  color: #1e293b !important;
}
.neuro-command-layout[data-theme="light"] .el-table__body-wrapper .el-button--danger:not(.is-link):not(.is-text):hover,
.neuro-command-layout[data-theme="light"] .el-table__body .el-button--danger:not(.is-link):not(.is-text):hover,
.neuro-command-layout[data-theme="light"] .el-table__fixed .el-button--danger:not(.is-link):not(.is-text):hover,
.neuro-command-layout[data-theme="light"] .el-table__fixed-right .el-button--danger:not(.is-link):not(.is-text):hover,
.neuro-command-layout[data-theme="light"] .el-table__fixed-left .el-button--danger:not(.is-link):not(.is-text):hover {
  border-color: #ef4444 !important;
  color: #ef4444 !important;
  background: rgba(239, 68, 68, 0.06) !important;
}

/* ── 7. 操作列布局 ─────────────────────── */
.neuro-command-layout .op-btns {
  display: inline-flex;
  flex-wrap: nowrap;
  gap: 6px;
  align-items: center;
}
.neuro-command-layout .el-table__body .cell { white-space: nowrap; }
.neuro-command-layout .el-table__body .cell .el-button { white-space: nowrap; }
.neuro-command-layout .el-table__body-wrapper .el-table__row > td:last-child .cell,
.neuro-command-layout .el-table__body .el-table__row > td:last-child .cell {
  display: flex;
  flex-wrap: nowrap;
  gap: 6px;
  align-items: center;
  justify-content: flex-start;
  padding: 4px 12px;
}

/* ── 8. el-tag 统一样式 ────────────────── */
.neuro-command-layout .el-tag {
  padding: 2px 8px;
  border-radius: 8px;
  font-size: 11px;
  line-height: 1.4;
  border: 1px solid;
}
.neuro-command-layout .el-tag.el-tag--small { padding: 2px 8px; font-size: 11px; }
.neuro-command-layout .el-tag--plain {
  background: rgba(0, 245, 212, 0.06);
  border-color: rgba(0, 245, 212, 0.15);
  color: var(--nm-primary, #00f5d4);
}
.neuro-command-layout .el-tag--plain.el-tag--success { background: rgba(16, 185, 129, 0.08); border-color: rgba(16, 185, 129, 0.2); color: #10b981; }
.neuro-command-layout .el-tag--plain.el-tag--warning { background: rgba(245, 158, 11, 0.08); border-color: rgba(245, 158, 11, 0.2); color: #f59e0b; }
.neuro-command-layout .el-tag--plain.el-tag--danger { background: rgba(239, 68, 68, 0.08); border-color: rgba(239, 68, 68, 0.2); color: #ef4444; }
.neuro-command-layout .el-tag--plain.el-tag--info { background: rgba(100, 116, 139, 0.08); border-color: rgba(100, 116, 139, 0.2); color: #94a3b8; }
.neuro-command-layout .el-tag--light {
  background: rgba(0, 245, 212, 0.04);
  border-color: rgba(0, 245, 212, 0.1);
  color: var(--nm-primary, #00f5d4);
}

/* 浅色模式 el-tag */
.neuro-command-layout[data-theme="light"] .el-tag--plain { background: rgba(14, 165, 233, 0.06); border-color: rgba(14, 165, 233, 0.15); color: #0ea5e9; }
.neuro-command-layout[data-theme="light"] .el-tag--plain.el-tag--success { background: rgba(16, 185, 129, 0.08); border-color: rgba(16, 185, 129, 0.2); color: #10b981; }
.neuro-command-layout[data-theme="light"] .el-tag--plain.el-tag--warning { background: rgba(245, 158, 11, 0.08); border-color: rgba(245, 158, 11, 0.2); color: #f59e0b; }
.neuro-command-layout[data-theme="light"] .el-tag--plain.el-tag--danger { background: rgba(239, 68, 68, 0.08); border-color: rgba(239, 68, 68, 0.2); color: #ef4444; }
.neuro-command-layout[data-theme="light"] .el-tag--plain.el-tag--info { background: rgba(100, 116, 139, 0.06); border-color: rgba(100, 116, 139, 0.15); color: #64748b; }

/* ── 9. Focus / 禁用态 ──────────────────── */
.neuro-command-layout .el-button:focus-visible {
  outline: 2px solid var(--nm-primary);
  outline-offset: 2px;
}
.neuro-command-layout .el-button.is-disabled {
  opacity: 0.45;
  cursor: not-allowed;
  transform: none !important;
  box-shadow: none !important;
}
</style>
