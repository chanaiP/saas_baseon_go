export function sanitizePhoneInput(value: string): string {
  return value.replace(/[^0-9\s\-()]/g, '')
}

export function normalizePhoneInput(value: string): string {
  return sanitizePhoneInput(value).replace(/[\s\-()]/g, '')
}

export function isValidOptionalPhone(value: string): boolean {
  const trimmed = value.trim()
  if (!trimmed) return true
  const digits = normalizePhoneInput(trimmed)
  return digits.length >= 10 && digits.length <= 15
}

export function isValidMobilePhone(value: string): boolean {
  return /^1[3-9]\d{9}$/.test(normalizePhoneInput(value))
}

/** 登录账号框：是否与后端「手机号登录」探测规则一致（仅数字与少量分隔符，且位数合法） */
export function isLikelyPhoneAccountInput(raw: string): boolean {
  const t = raw.trim()
  if (!t || /[a-zA-Z]/.test(t)) return false
  if (!/^[0-9\s\-()]+$/.test(t)) return false
  const digits = normalizePhoneInput(t)
  return digits.length >= 10 && digits.length <= 15
}
