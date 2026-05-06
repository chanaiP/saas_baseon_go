import { defineStore } from 'pinia'
import { ref, watch } from 'vue'

import type { PrimaryPresetId } from '@/constants/layoutPresets'
import { PRIMARY_PRESETS } from '@/constants/layoutPresets'

const THEME_KEY = 'ui_theme'
const WATERMARK_KEY = 'ui_watermark_text'
const SIDEBAR_COLLAPSED_KEY = 'ui_sidebar_collapsed'
const PLATFORM_NAME_KEY = 'ui_platform_name'
const PLATFORM_LOGO_KEY = 'ui_platform_logo_url'

const PRIMARY_PRESET_KEY = 'ui_primary_preset'
const NAV_MODE_KEY = 'ui_nav_mode'
const SIDEBAR_STYLE_KEY = 'ui_sidebar_style'
const CONTENT_WIDTH_KEY = 'ui_content_width'
const FIXED_HEADER_KEY = 'ui_fixed_header'
const FIXED_SIDEBAR_KEY = 'ui_fixed_sidebar'
const AUTO_SPLIT_MENU_KEY = 'ui_auto_split_menu'
const SHOW_TOPBAR_KEY = 'ui_show_topbar'
const SHOW_FOOTER_KEY = 'ui_show_footer'
const SHOW_MENU_KEY = 'ui_show_menu'
const SHOW_MENU_HEADER_KEY = 'ui_show_menu_header'
const COLOR_WEAK_KEY = 'ui_color_weak'
const MIX_MODULE_ID_KEY = 'ui_mix_module_id'

export const DEFAULT_PLATFORM_NAME = '管理台'

export type UiTheme = 'light' | 'dark'
export type NavMode = 'side' | 'top' | 'mix'
export type SidebarStyle = 'default' | 'light'
export type ContentWidthMode = 'fluid' | 'fixed'

function readStoredTheme(): UiTheme {
  const v = localStorage.getItem(THEME_KEY)
  return v === 'light' ? 'light' : 'dark'
}

function readSidebarCollapsed(): boolean {
  return localStorage.getItem(SIDEBAR_COLLAPSED_KEY) === '1'
}

function readPlatformName(): string {
  const v = localStorage.getItem(PLATFORM_NAME_KEY)?.trim()
  return v || DEFAULT_PLATFORM_NAME
}

function readPlatformLogoUrl(): string {
  return localStorage.getItem(PLATFORM_LOGO_KEY)?.trim() ?? ''
}

function readPrimaryPreset(): PrimaryPresetId {
  const v = localStorage.getItem(PRIMARY_PRESET_KEY)
  if (v && PRIMARY_PRESETS.some((p) => p.id === v)) return v as PrimaryPresetId
  return 'teal'
}

function readNavMode(): NavMode {
  const v = localStorage.getItem(NAV_MODE_KEY)
  if (v === 'top' || v === 'mix') return v
  return 'side'
}

function readSidebarStyle(): SidebarStyle {
  const v = localStorage.getItem(SIDEBAR_STYLE_KEY)
  if (v === 'default') return 'default'
  /* 默认浅色侧栏，对齐 Ant Design Pro */
  return 'light'
}

function readContentWidth(): ContentWidthMode {
  return localStorage.getItem(CONTENT_WIDTH_KEY) === 'fixed' ? 'fixed' : 'fluid'
}

function readBool(key: string, defaultVal: boolean): boolean {
  const v = localStorage.getItem(key)
  if (v === null) return defaultVal
  return v === '1'
}

function readMixModuleId(): string {
  return localStorage.getItem(MIX_MODULE_ID_KEY) || 'sys'
}

function applyDomTheme(mode: UiTheme) {
  const root = document.documentElement
  if (mode === 'dark') root.classList.add('dark')
  else root.classList.remove('dark')
}

function applyPrimaryPreset(id: PrimaryPresetId) {
  document.documentElement.setAttribute('data-ui-primary', id)
}

function applyColorWeak(on: boolean) {
  const root = document.documentElement
  if (on) root.classList.add('ui-color-weak')
  else root.classList.remove('ui-color-weak')
}

export const useUiPreferencesStore = defineStore('uiPreferences', () => {
  const theme = ref<UiTheme>(readStoredTheme())
  const watermarkText = ref(localStorage.getItem(WATERMARK_KEY) ?? '')
  const sidebarCollapsed = ref(readSidebarCollapsed())
  const platformName = ref(readPlatformName())
  const platformLogoUrl = ref(readPlatformLogoUrl())

  const primaryPreset = ref<PrimaryPresetId>(readPrimaryPreset())
  const navMode = ref<NavMode>(readNavMode())
  const sidebarStyle = ref<SidebarStyle>(readSidebarStyle())
  const contentWidth = ref<ContentWidthMode>(readContentWidth())
  const fixedHeader = ref(readBool(FIXED_HEADER_KEY, false))
  const fixedSidebar = ref(readBool(FIXED_SIDEBAR_KEY, true))
  const autoSplitMenu = ref(readBool(AUTO_SPLIT_MENU_KEY, false))
  const showTopbar = ref(readBool(SHOW_TOPBAR_KEY, true))
  const showFooter = ref(readBool(SHOW_FOOTER_KEY, true))
  const showMenu = ref(readBool(SHOW_MENU_KEY, true))
  const showMenuHeader = ref(readBool(SHOW_MENU_HEADER_KEY, true))
  const colorWeak = ref(readBool(COLOR_WEAK_KEY, false))
  const mixModuleId = ref(readMixModuleId())

  applyDomTheme(theme.value)
  applyPrimaryPreset(primaryPreset.value)
  applyColorWeak(colorWeak.value)

  watch(theme, (m) => {
    localStorage.setItem(THEME_KEY, m)
    applyDomTheme(m)
  })

  watch(sidebarCollapsed, (c) => {
    localStorage.setItem(SIDEBAR_COLLAPSED_KEY, c ? '1' : '0')
  })

  watch(platformName, (n) => {
    const t = n.trim() || DEFAULT_PLATFORM_NAME
    localStorage.setItem(PLATFORM_NAME_KEY, t)
  })

  watch(platformLogoUrl, (u) => {
    const t = u.trim()
    if (t) localStorage.setItem(PLATFORM_LOGO_KEY, t)
    else localStorage.removeItem(PLATFORM_LOGO_KEY)
  })

  watch(primaryPreset, (id) => {
    localStorage.setItem(PRIMARY_PRESET_KEY, id)
    applyPrimaryPreset(id)
  })

  watch(navMode, (m) => {
    localStorage.setItem(NAV_MODE_KEY, m)
  })

  watch(sidebarStyle, (s) => {
    localStorage.setItem(SIDEBAR_STYLE_KEY, s)
  })

  watch(contentWidth, (w) => {
    localStorage.setItem(CONTENT_WIDTH_KEY, w)
  })

  watch(fixedHeader, (v) => localStorage.setItem(FIXED_HEADER_KEY, v ? '1' : '0'))
  watch(fixedSidebar, (v) => localStorage.setItem(FIXED_SIDEBAR_KEY, v ? '1' : '0'))
  watch(autoSplitMenu, (v) => localStorage.setItem(AUTO_SPLIT_MENU_KEY, v ? '1' : '0'))
  watch(showTopbar, (v) => localStorage.setItem(SHOW_TOPBAR_KEY, v ? '1' : '0'))
  watch(showFooter, (v) => localStorage.setItem(SHOW_FOOTER_KEY, v ? '1' : '0'))
  watch(showMenu, (v) => localStorage.setItem(SHOW_MENU_KEY, v ? '1' : '0'))
  watch(showMenuHeader, (v) => localStorage.setItem(SHOW_MENU_HEADER_KEY, v ? '1' : '0'))
  watch(colorWeak, (v) => {
    localStorage.setItem(COLOR_WEAK_KEY, v ? '1' : '0')
    applyColorWeak(v)
  })

  watch(mixModuleId, (id) => {
    localStorage.setItem(MIX_MODULE_ID_KEY, id)
  })

  function setTheme(m: UiTheme) {
    theme.value = m
  }

  function setWatermark(text: string) {
    const t = text.trim()
    watermarkText.value = t
    if (t) localStorage.setItem(WATERMARK_KEY, t)
    else localStorage.removeItem(WATERMARK_KEY)
  }

  function toggleSidebarCollapsed() {
    sidebarCollapsed.value = !sidebarCollapsed.value
  }

  function setPlatformBrand(name: string, logoUrl: string) {
    platformName.value = name.trim() || DEFAULT_PLATFORM_NAME
    platformLogoUrl.value = logoUrl.trim()
  }

  function setPrimaryPreset(id: PrimaryPresetId) {
    primaryPreset.value = id
  }

  function setNavMode(m: NavMode) {
    navMode.value = m
  }

  function setSidebarStyle(s: SidebarStyle) {
    sidebarStyle.value = s
  }

  function setContentWidth(w: ContentWidthMode) {
    contentWidth.value = w
  }

  function setMixModuleId(id: string) {
    mixModuleId.value = id
  }

  /** 供「拷贝设置」导出 */
  function layoutSettingsSnapshot() {
    return {
      theme: theme.value,
      primaryPreset: primaryPreset.value,
      navMode: navMode.value,
      sidebarStyle: sidebarStyle.value,
      contentWidth: contentWidth.value,
      fixedHeader: fixedHeader.value,
      fixedSidebar: fixedSidebar.value,
      autoSplitMenu: autoSplitMenu.value,
      showTopbar: showTopbar.value,
      showFooter: showFooter.value,
      showMenu: showMenu.value,
      showMenuHeader: showMenuHeader.value,
      colorWeak: colorWeak.value,
      mixModuleId: mixModuleId.value,
      sidebarCollapsed: sidebarCollapsed.value,
      watermarkText: watermarkText.value,
    }
  }

  return {
    theme,
    watermarkText,
    sidebarCollapsed,
    platformName,
    platformLogoUrl,
    primaryPreset,
    navMode,
    sidebarStyle,
    contentWidth,
    fixedHeader,
    fixedSidebar,
    autoSplitMenu,
    showTopbar,
    showFooter,
    showMenu,
    showMenuHeader,
    colorWeak,
    mixModuleId,
    setTheme,
    setWatermark,
    toggleSidebarCollapsed,
    setPlatformBrand,
    setPrimaryPreset,
    setNavMode,
    setSidebarStyle,
    setContentWidth,
    setMixModuleId,
    layoutSettingsSnapshot,
  }
})
