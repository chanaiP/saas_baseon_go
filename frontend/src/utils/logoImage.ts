/** 侧栏/登录 Logo：优先保留透明底，输出 PNG Data URL；不足时退回 JPEG 以控制体积，便于写入 localStorage */

import { MAX_IMAGE_UPLOAD_BYTES, MAX_IMAGE_UPLOAD_LABEL } from '@/constants/uploadLimits'
const MAX_OUTPUT_CHARS = 750_000
const CANVAS_MAX_SIDE = 256

function loadImageFromFile(file: File): Promise<HTMLImageElement> {
  return new Promise((resolve, reject) => {
    const url = URL.createObjectURL(file)
    const img = new Image()
    img.onload = () => {
      URL.revokeObjectURL(url)
      resolve(img)
    }
    img.onerror = () => {
      URL.revokeObjectURL(url)
      reject(new Error('无法读取该图片'))
    }
    img.src = url
  })
}

export interface LogoTransformOptions {
  /** 以图片适配画布为 1.0 的基础缩放倍数，用于放大裁剪中心区域 */
  scale?: number
}

/**
 * 将用户选择的图片转为适合本地持久化的 Data URL：
 * - 优先输出 PNG（可保留透明底）
 * - 若体积超出限制，则退回 JPEG 并逐步降低质量
 */
export async function fileToLogoDataUrl(file: File, opts: LogoTransformOptions = {}): Promise<string> {
  if (!file.type.startsWith('image/')) {
    throw new Error('请选择图片文件（JPG、PNG、WebP、GIF 等）')
  }
  if (file.size > MAX_IMAGE_UPLOAD_BYTES) {
    throw new Error(`图片请小于 ${MAX_IMAGE_UPLOAD_LABEL}`)
  }
  const img = await loadImageFromFile(file)
  const iw = img.naturalWidth
  const ih = img.naturalHeight
  if (iw < 1 || ih < 1) {
    throw new Error('图片尺寸无效')
  }
  const canvas = document.createElement('canvas')
  canvas.width = CANVAS_MAX_SIDE
  canvas.height = CANVAS_MAX_SIDE
  const ctx = canvas.getContext('2d')
  if (!ctx) {
    throw new Error('浏览器不支持图片处理')
  }
  // 先将图片按短边铺满画布，再根据 opts.scale 额外放大，居中绘制，实现简单缩放裁剪
  const baseScale = Math.min(CANVAS_MAX_SIDE / iw, CANVAS_MAX_SIDE / ih)
  const extraScale = Math.min(Math.max(opts.scale ?? 1, 1), 2) // 限制在 [1, 2]
  const s = baseScale * extraScale
  const dw = iw * s
  const dh = ih * s
  const dx = (CANVAS_MAX_SIDE - dw) / 2
  const dy = (CANVAS_MAX_SIDE - dh) / 2
  ctx.clearRect(0, 0, CANVAS_MAX_SIDE, CANVAS_MAX_SIDE)
  // PNG：保留透明像素，适合无底纹 Logo
  ctx.drawImage(img, dx, dy, dw, dh)
  let dataUrl = canvas.toDataURL('image/png')
  if (dataUrl.length <= MAX_OUTPUT_CHARS) {
    return dataUrl
  }
  // 体积过大时，退回 JPEG 并控制质量
  let quality = 0.88
  dataUrl = canvas.toDataURL('image/jpeg', quality)
  while (dataUrl.length > MAX_OUTPUT_CHARS && quality > 0.45) {
    quality -= 0.08
    dataUrl = canvas.toDataURL('image/jpeg', quality)
  }
  if (dataUrl.length > MAX_OUTPUT_CHARS) {
    throw new Error('图片仍过大，请换一张更简单或更小的图')
  }
  return dataUrl
}
