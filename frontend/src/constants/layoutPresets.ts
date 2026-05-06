/** Element Plus 主题色预设（壳层默认 teal；选「蓝」为 AD Pro 经典蓝） */
export const PRIMARY_PRESETS = [
  { id: 'blue', label: '蓝', color: '#1677ff' },
  { id: 'sky', label: '天蓝', color: '#13c2c2' },
  { id: 'red', label: '红', color: '#f56c6c' },
  { id: 'orange', label: '橙', color: '#e6a23c' },
  { id: 'yellow', label: '黄', color: '#d4a017' },
  { id: 'teal', label: '青', color: '#0d9488' },
  { id: 'green', label: '绿', color: '#67c23a' },
  { id: 'royal', label: '宝蓝', color: '#4169e1' },
  { id: 'purple', label: '紫', color: '#9b59b6' },
] as const

export type PrimaryPresetId = (typeof PRIMARY_PRESETS)[number]['id']
