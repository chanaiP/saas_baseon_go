/** 串末是否含 Z 或 ±偏移（含 +08:00），有时区则交给 `Date` 原生解析 */
function hasExplicitTimezone(iso: string): boolean {
  return /(Z|[+-]\d{2}:?\d{2}(:\d{2})?)$/i.test(iso.trim())
}

/**
 * 将日志相关时间串解析为「绝对时刻」的 Date，供 `formatDateTimeChina` 使用。
 *
 * - 带 `Z` / `±offset`：交给 `Date` 解析。
 * - **`YYYY-MM-DD HH:mm:ss`（空格、无 T）**：视为 **北京时间墙钟**（避免与其它展示串混用 UTC 规则）。
 * - **`YYYY-MM-DDTHH:mm:ss` 且无后缀**：视为 **UTC 墙钟**（常见 JSON 省略 Z），再转上海展示。
 */
export function parseBackendUtcToDate(value: string | Date | number | null | undefined): Date | null {
  if (value == null || value === '') return null
  if (value instanceof Date) return Number.isNaN(value.getTime()) ? null : value
  if (typeof value === 'number') {
    const d = new Date(value)
    return Number.isNaN(d.getTime()) ? null : d
  }
  const s = String(value).trim()
  if (!s) return null
  if (hasExplicitTimezone(s)) {
    const d = new Date(s)
    return Number.isNaN(d.getTime()) ? null : d
  }
  // `YYYY-MM-DD HH:mm:ss`：按东八区墙钟解析（与 `...T...` 无后缀按 UTC 的规则区分）
  const mSpace = s.match(/^(\d{4}-\d{2}-\d{2}) (\d{2}:\d{2}:\d{2})(\.\d+)?$/)
  if (mSpace) {
    const d = new Date(`${mSpace[1]}T${mSpace[2]}${mSpace[3] ?? ''}+08:00`)
    return Number.isNaN(d.getTime()) ? null : d
  }
  const m = s.match(/^(\d{4}-\d{2}-\d{2})T(\d{2}:\d{2}:\d{2})(\.\d+)?/)
  if (!m) return null
  const d = new Date(`${m[1]}T${m[2]}Z`)
  return Number.isNaN(d.getTime()) ? null : d
}

/**
 * 格式化为北京时间（Asia/Shanghai），形如 `2026-05-04 18:31:18`（`YYYY-MM-DD HH:mm:ss`），用于日志等列表展示。
 */
export function formatDateTimeChina(value: string | Date | number | null | undefined): string {
  if (value == null || value === '') return '—'
  const d = parseBackendUtcToDate(value)
  if (!d) {
    if (typeof value === 'string' && value.includes('T')) {
      return value.trim().replace('T', ' ').replace(/\.\d{3}/, '').slice(0, 19)
    }
    return '—'
  }
  const parts = new Intl.DateTimeFormat('en-GB', {
    timeZone: 'Asia/Shanghai',
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
    hourCycle: 'h23',
  }).formatToParts(d)
  const get = (type: Intl.DateTimeFormatPart['type']) => parts.find((p) => p.type === type)?.value ?? ''
  return `${get('year')}-${get('month')}-${get('day')} ${get('hour')}:${get('minute')}:${get('second')}`
}
