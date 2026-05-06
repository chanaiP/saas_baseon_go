/**
 * NeuroAgent 主题管理系统
 * 提供统一的主题切换和管理功能
 */

import { ref, onMounted } from 'vue'

export type ThemeMode = 'dark' | 'light' | 'cyber' | 'neural' | 'matrix'
export type ThemeColor = 'blue' | 'purple' | 'green' | 'orange' | 'pink'

export interface ThemeConfig {
  mode: ThemeMode
  color: ThemeColor
  animations: boolean
  neuralEffects: boolean
  highContrast: boolean
}

export interface ThemeColors {
  primary: string
  secondary: string
  accent: string
  background: string
  surface: string
  text: string
  textSecondary: string
  border: string
  success: string
  warning: string
  error: string
  info: string
  neural: string
  matrix: string
}

// 默认主题配置
const defaultTheme: ThemeConfig = {
  mode: 'dark',
  color: 'blue',
  animations: true,
  neuralEffects: true,
  highContrast: false
}

// 主题颜色映射
const themeColorMap: Record<ThemeMode, Record<ThemeColor, ThemeColors>> = {
  dark: {
    blue: {
      primary: '#00f5d4',
      secondary: '#00b8a9',
      accent: '#8b5cf6',
      background: '#0f172a',
      surface: '#1e293b',
      text: '#f8fafc',
      textSecondary: '#94a3b8',
      border: '#334155',
      success: '#10b981',
      warning: '#f59e0b',
      error: '#ef4444',
      info: '#3b82f6',
      neural: '#00f5d4',
      matrix: '#00ff41'
    },
    purple: {
      primary: '#a855f7',
      secondary: '#7c3aed',
      accent: '#ec4899',
      background: '#0f172a',
      surface: '#1e293b',
      text: '#f8fafc',
      textSecondary: '#94a3b8',
      border: '#334155',
      success: '#10b981',
      warning: '#f59e0b',
      error: '#ef4444',
      info: '#3b82f6',
      neural: '#a855f7',
      matrix: '#ec4899'
    },
    green: {
      primary: '#10b981',
      secondary: '#059669',
      accent: '#0ea5e9',
      background: '#0f172a',
      surface: '#1e293b',
      text: '#f8fafc',
      textSecondary: '#94a3b8',
      border: '#334155',
      success: '#10b981',
      warning: '#f59e0b',
      error: '#ef4444',
      info: '#3b82f6',
      neural: '#10b981',
      matrix: '#0ea5e9'
    },
    orange: {
      primary: '#f97316',
      secondary: '#ea580c',
      accent: '#eab308',
      background: '#0f172a',
      surface: '#1e293b',
      text: '#f8fafc',
      textSecondary: '#94a3b8',
      border: '#334155',
      success: '#10b981',
      warning: '#f59e0b',
      error: '#ef4444',
      info: '#3b82f6',
      neural: '#f97316',
      matrix: '#eab308'
    },
    pink: {
      primary: '#ec4899',
      secondary: '#db2777',
      accent: '#a855f7',
      background: '#0f172a',
      surface: '#1e293b',
      text: '#f8fafc',
      textSecondary: '#94a3b8',
      border: '#334155',
      success: '#10b981',
      warning: '#f59e0b',
      error: '#ef4444',
      info: '#3b82f6',
      neural: '#ec4899',
      matrix: '#a855f7'
    }
  },
  light: {
    blue: {
      primary: '#0ea5e9',
      secondary: '#0284c7',
      accent: '#8b5cf6',
      background: '#f8fafc',
      surface: '#ffffff',
      text: '#1e293b',
      textSecondary: '#64748b',
      border: '#cbd5e1',
      success: '#10b981',
      warning: '#f59e0b',
      error: '#ef4444',
      info: '#3b82f6',
      neural: '#0ea5e9',
      matrix: '#8b5cf6'
    },
    purple: {
      primary: '#a855f7',
      secondary: '#7c3aed',
      accent: '#ec4899',
      background: '#f8fafc',
      surface: '#ffffff',
      text: '#1e293b',
      textSecondary: '#64748b',
      border: '#cbd5e1',
      success: '#10b981',
      warning: '#f59e0b',
      error: '#ef4444',
      info: '#3b82f6',
      neural: '#a855f7',
      matrix: '#ec4899'
    },
    green: {
      primary: '#10b981',
      secondary: '#059669',
      accent: '#0ea5e9',
      background: '#f8fafc',
      surface: '#ffffff',
      text: '#1e293b',
      textSecondary: '#64748b',
      border: '#cbd5e1',
      success: '#10b981',
      warning: '#f59e0b',
      error: '#ef4444',
      info: '#3b82f6',
      neural: '#10b981',
      matrix: '#0ea5e9'
    },
    orange: {
      primary: '#f97316',
      secondary: '#ea580c',
      accent: '#eab308',
      background: '#f8fafc',
      surface: '#ffffff',
      text: '#1e293b',
      textSecondary: '#64748b',
      border: '#cbd5e1',
      success: '#10b981',
      warning: '#f59e0b',
      error: '#ef4444',
      info: '#3b82f6',
      neural: '#f97316',
      matrix: '#eab308'
    },
    pink: {
      primary: '#ec4899',
      secondary: '#db2777',
      accent: '#a855f7',
      background: '#f8fafc',
      surface: '#ffffff',
      text: '#1e293b',
      textSecondary: '#64748b',
      border: '#cbd5e1',
      success: '#10b981',
      warning: '#f59e0b',
      error: '#ef4444',
      info: '#3b82f6',
      neural: '#ec4899',
      matrix: '#a855f7'
    }
  },
  cyber: {
    blue: {
      primary: '#00f5d4',
      secondary: '#00ff41',
      accent: '#ff00ff',
      background: '#000000',
      surface: '#0a0a0a',
      text: '#00ff41',
      textSecondary: '#00f5d4',
      border: '#00ff41',
      success: '#00ff41',
      warning: '#ffff00',
      error: '#ff0000',
      info: '#00ffff',
      neural: '#00f5d4',
      matrix: '#00ff41'
    },
    purple: {
      primary: '#ff00ff',
      secondary: '#a855f7',
      accent: '#00ffff',
      background: '#000000',
      surface: '#0a0a0a',
      text: '#ff00ff',
      textSecondary: '#a855f7',
      border: '#ff00ff',
      success: '#00ff41',
      warning: '#ffff00',
      error: '#ff0000',
      info: '#00ffff',
      neural: '#ff00ff',
      matrix: '#a855f7'
    },
    green: {
      primary: '#00ff41',
      secondary: '#00f5d4',
      accent: '#ffff00',
      background: '#000000',
      surface: '#0a0a0a',
      text: '#00ff41',
      textSecondary: '#00f5d4',
      border: '#00ff41',
      success: '#00ff41',
      warning: '#ffff00',
      error: '#ff0000',
      info: '#00ffff',
      neural: '#00ff41',
      matrix: '#00f5d4'
    },
    orange: {
      primary: '#ff6b00',
      secondary: '#ffd700',
      accent: '#ff00ff',
      background: '#000000',
      surface: '#0a0a0a',
      text: '#ff6b00',
      textSecondary: '#ffd700',
      border: '#ff6b00',
      success: '#00ff41',
      warning: '#ffff00',
      error: '#ff0000',
      info: '#00ffff',
      neural: '#ff6b00',
      matrix: '#ffd700'
    },
    pink: {
      primary: '#ff00ff',
      secondary: '#ff1493',
      accent: '#00ffff',
      background: '#000000',
      surface: '#0a0a0a',
      text: '#ff00ff',
      textSecondary: '#ff1493',
      border: '#ff00ff',
      success: '#00ff41',
      warning: '#ffff00',
      error: '#ff0000',
      info: '#00ffff',
      neural: '#ff00ff',
      matrix: '#ff1493'
    }
  },
  neural: {
    blue: {
      primary: '#00f5d4',
      secondary: '#00b8a9',
      accent: '#8b5cf6',
      background: '#1a1a2e',
      surface: '#16213e',
      text: '#e6f7ff',
      textSecondary: '#94d2ff',
      border: '#0ea5e9',
      success: '#00ff9d',
      warning: '#ffd700',
      error: '#ff4d6d',
      info: '#00f5d4',
      neural: '#00f5d4',
      matrix: '#8b5cf6'
    },
    purple: {
      primary: '#a855f7',
      secondary: '#7c3aed',
      accent: '#ec4899',
      background: '#1a1a2e',
      surface: '#16213e',
      text: '#f3e8ff',
      textSecondary: '#d8b4fe',
      border: '#a855f7',
      success: '#00ff9d',
      warning: '#ffd700',
      error: '#ff4d6d',
      info: '#a855f7',
      neural: '#a855f7',
      matrix: '#ec4899'
    },
    green: {
      primary: '#00ff9d',
      secondary: '#00cc7a',
      accent: '#0ea5e9',
      background: '#1a1a2e',
      surface: '#16213e',
      text: '#e6fff2',
      textSecondary: '#94ffc2',
      border: '#00ff9d',
      success: '#00ff9d',
      warning: '#ffd700',
      error: '#ff4d6d',
      info: '#00ff9d',
      neural: '#00ff9d',
      matrix: '#0ea5e9'
    },
    orange: {
      primary: '#ff9d00',
      secondary: '#ff7b00',
      accent: '#ffd700',
      background: '#1a1a2e',
      surface: '#16213e',
      text: '#fff7e6',
      textSecondary: '#ffd8a8',
      border: '#ff9d00',
      success: '#00ff9d',
      warning: '#ffd700',
      error: '#ff4d6d',
      info: '#ff9d00',
      neural: '#ff9d00',
      matrix: '#ffd700'
    },
    pink: {
      primary: '#ff4d6d',
      secondary: '#ff2e63',
      accent: '#ff8fab',
      background: '#1a1a2e',
      surface: '#16213e',
      text: '#ffe6ea',
      textSecondary: '#ffb3c1',
      border: '#ff4d6d',
      success: '#00ff9d',
      warning: '#ffd700',
      error: '#ff4d6d',
      info: '#ff4d6d',
      neural: '#ff4d6d',
      matrix: '#ff8fab'
    }
  },
  matrix: {
    blue: {
      primary: '#00ff41',
      secondary: '#00cc33',
      accent: '#00f5d4',
      background: '#000000',
      surface: '#0a0a0a',
      text: '#00ff41',
      textSecondary: '#00cc33',
      border: '#00ff41',
      success: '#00ff41',
      warning: '#ffff00',
      error: '#ff0000',
      info: '#00ffff',
      neural: '#00ff41',
      matrix: '#00f5d4'
    },
    purple: {
      primary: '#ff00ff',
      secondary: '#cc00cc',
      accent: '#a855f7',
      background: '#000000',
      surface: '#0a0a0a',
      text: '#ff00ff',
      textSecondary: '#cc00cc',
      border: '#ff00ff',
      success: '#00ff41',
      warning: '#ffff00',
      error: '#ff0000',
      info: '#00ffff',
      neural: '#ff00ff',
      matrix: '#a855f7'
    },
    green: {
      primary: '#00ff41',
      secondary: '#00cc33',
      accent: '#00f5d4',
      background: '#000000',
      surface: '#0a0a0a',
      text: '#00ff41',
      textSecondary: '#00cc33',
      border: '#00ff41',
      success: '#00ff41',
      warning: '#ffff00',
      error: '#ff0000',
      info: '#00ffff',
      neural: '#00ff41',
      matrix: '#00f5d4'
    },
    orange: {
      primary: '#ff6b00',
      secondary: '#cc5500',
      accent: '#ffd700',
      background: '#000000',
      surface: '#0a0a0a',
      text: '#ff6b00',
      textSecondary: '#cc5500',
      border: '#ff6b00',
      success: '#00ff41',
      warning: '#ffff00',
      error: '#ff0000',
      info: '#00ffff',
      neural: '#ff6b00',
      matrix: '#ffd700'
    },
    pink: {
      primary: '#ff1493',
      secondary: '#cc1075',
      accent: '#ff00ff',
      background: '#000000',
      surface: '#0a0a0a',
      text: '#ff1493',
      textSecondary: '#cc1075',
      border: '#ff1493',
      success: '#00ff41',
      warning: '#ffff00',
      error: '#ff0000',
      info: '#00ffff',
      neural: '#ff1493',
      matrix: '#ff00ff'
    }
  }
}

// CSS 变量生成器
export function generateCSSVariables(colors: ThemeColors): Record<string, string> {
  return {
    '--neuro-primary': colors.primary,
    '--neuro-secondary': colors.secondary,
    '--neuro-accent': colors.accent,
    '--neuro-background': colors.background,
    '--neuro-surface': colors.surface,
    '--neuro-text': colors.text,
    '--neuro-text-secondary': colors.textSecondary,
    '--neuro-border': colors.border,
    '--neuro-success': colors.success,
    '--neuro-warning': colors.warning,
    '--neuro-error': colors.error,
    '--neuro-info': colors.info,
    '--neuro-neural': colors.neural,
    '--neuro-matrix': colors.matrix,

    // 渐变
    '--neuro-gradient-primary': `linear-gradient(135deg, ${colors.primary}, ${colors.accent})`,
    '--neuro-gradient-secondary': `linear-gradient(135deg, ${colors.secondary}, ${colors.primary})`,
    '--neuro-gradient-neural': `linear-gradient(135deg, ${colors.neural}, ${colors.matrix})`,

    // 透明度变体
    '--neuro-primary-10': `${colors.primary}1a`,
    '--neuro-primary-20': `${colors.primary}33`,
    '--neuro-primary-30': `${colors.primary}4d`,
    '--neuro-surface-90': `${colors.surface}e6`,

    // 阴影
    '--neuro-shadow-sm': `0 1px 2px 0 ${colors.primary}1a`,
    '--neuro-shadow-md': `0 4px 6px -1px ${colors.primary}1a, 0 2px 4px -1px ${colors.primary}0f`,
    '--neuro-shadow-lg': `0 10px 15px -3px ${colors.primary}1a, 0 4px 6px -2px ${colors.primary}0f`,
    '--neuro-shadow-xl': `0 20px 25px -5px ${colors.primary}1a, 0 10px 10px -5px ${colors.primary}0f`,

    // 神经特效
    '--neuro-glow': `0 0 20px ${colors.neural}`,
    '--neuro-pulse': `0 0 0 4px ${colors.neural}40`,
    '--neuro-connection': `2px solid ${colors.neural}80`
  }
}

// 主题管理器类
export class NeuroThemeManager {
  private config: ThemeConfig = defaultTheme
  private subscribers: Array<(config: ThemeConfig) => void> = []

  constructor() {
    this.loadFromStorage()
    this.applyToDocument()
  }

  // 获取当前配置
  getConfig(): ThemeConfig {
    return { ...this.config }
  }

  // 获取当前颜色
  getColors(): ThemeColors {
    return themeColorMap[this.config.mode][this.config.color]
  }

  // 获取CSS变量
  getCSSVariables(): Record<string, string> {
    return generateCSSVariables(this.getColors())
  }

  // 设置主题模式
  setMode(mode: ThemeMode): void {
    this.config.mode = mode
    this.saveToStorage()
    this.applyToDocument()
    this.notifySubscribers()
  }

  // 设置主题颜色
  setColor(color: ThemeColor): void {
    this.config.color = color
    this.saveToStorage()
    this.applyToDocument()
    this.notifySubscribers()
  }

  // 切换动画
  toggleAnimations(enabled: boolean): void {
    this.config.animations = enabled
    this.saveToStorage()
    this.applyToDocument()
    this.notifySubscribers()
  }

  // 切换神经特效
  toggleNeuralEffects(enabled: boolean): void {
    this.config.neuralEffects = enabled
    this.saveToStorage()
    this.applyToDocument()
    this.notifySubscribers()
  }

  // 切换高对比度
  toggleHighContrast(enabled: boolean): void {
    this.config.highContrast = enabled
    this.saveToStorage()
    this.applyToDocument()
    this.notifySubscribers()
  }

  // 应用主题到文档
  applyToDocument(): void {
    const root = document.documentElement
    const variables = this.getCSSVariables()

    // 应用CSS变量
    Object.entries(variables).forEach(([key, value]) => {
      root.style.setProperty(key, value)
    })

    // 应用主题类
    root.setAttribute('data-neuro-theme', this.config.mode)
    root.setAttribute('data-neuro-color', this.config.color)

    if (this.config.animations) {
      root.classList.add('neuro-animations-enabled')
    } else {
      root.classList.remove('neuro-animations-enabled')
    }

    if (this.config.neuralEffects) {
      root.classList.add('neuro-effects-enabled')
    } else {
      root.classList.remove('neuro-effects-enabled')
    }

    if (this.config.highContrast) {
      root.classList.add('neuro-high-contrast')
    } else {
      root.classList.remove('neuro-high-contrast')
    }

    // 添加主题特定的类
    root.classList.add(`neuro-theme-${this.config.mode}`)
    root.classList.add(`neuro-color-${this.config.color}`)
  }

  // 订阅主题变化
  subscribe(callback: (config: ThemeConfig) => void): () => void {
    this.subscribers.push(callback)
    return () => {
      const index = this.subscribers.indexOf(callback)
      if (index > -1) {
        this.subscribers.splice(index, 1)
      }
    }
  }

  // 通知订阅者
  private notifySubscribers(): void {
    this.subscribers.forEach(callback => callback(this.getConfig()))
  }

  // 保存到本地存储
  private saveToStorage(): void {
    try {
      localStorage.setItem('neuro-theme-config', JSON.stringify(this.config))
    } catch (error) {
      console.warn('Failed to save theme config to localStorage:', error)
    }
  }

  // 从本地存储加载
  private loadFromStorage(): void {
    try {
      const saved = localStorage.getItem('neuro-theme-config')
      if (saved) {
        const parsed = JSON.parse(saved)
        this.config = { ...defaultTheme, ...parsed }
      }
    } catch (error) {
      console.warn('Failed to load theme config from localStorage:', error)
    }
  }

  // 重置为默认主题
  resetToDefault(): void {
    this.config = { ...defaultTheme }
    this.saveToStorage()
    this.applyToDocument()
    this.notifySubscribers()
  }

  // 获取可用主题模式
  static getAvailableModes(): ThemeMode[] {
    return ['dark', 'light', 'cyber', 'neural', 'matrix']
  }

  // 获取可用颜色
  static getAvailableColors(): ThemeColor[] {
    return ['blue', 'purple', 'green', 'orange', 'pink']
  }

  // 获取主题预览
  static getThemePreview(mode: ThemeMode, color: ThemeColor): { colors: ThemeColors, name: string } {
    const colors = themeColorMap[mode][color]
    const modeNames: Record<ThemeMode, string> = {
      dark: '暗黑',
      light: '明亮',
      cyber: '赛博',
      neural: '神经',
      matrix: '矩阵'
    }
    const colorNames: Record<ThemeColor, string> = {
      blue: '蓝色',
      purple: '紫色',
      green: '绿色',
      orange: '橙色',
      pink: '粉色'
    }

    return {
      colors,
      name: `${modeNames[mode]} · ${colorNames[color]}`
    }
  }
}

// 创建全局主题管理器实例
export const neuroThemeManager = new NeuroThemeManager()

// Vue composable
export function useNeuroTheme() {
  const config = ref<ThemeConfig>(neuroThemeManager.getConfig())
  const colors = ref<ThemeColors>(neuroThemeManager.getColors())

  // 订阅主题变化
  onMounted(() => {
    neuroThemeManager.subscribe((newConfig) => {
      config.value = newConfig
      colors.value = neuroThemeManager.getColors()
    })
  })

  return {
    config,
    colors,
    setMode: neuroThemeManager.setMode.bind(neuroThemeManager),
    setColor: neuroThemeManager.setColor.bind(neuroThemeManager),
    toggleAnimations: neuroThemeManager.toggleAnimations.bind(neuroThemeManager),
    toggleNeuralEffects: neuroThemeManager.toggleNeuralEffects.bind(neuroThemeManager),
    toggleHighContrast: neuroThemeManager.toggleHighContrast.bind(neuroThemeManager),
    resetToDefault: neuroThemeManager.resetToDefault.bind(neuroThemeManager),
    getAvailableModes: NeuroThemeManager.getAvailableModes,
    getAvailableColors: NeuroThemeManager.getAvailableColors,
    getThemePreview: NeuroThemeManager.getThemePreview
  }
}