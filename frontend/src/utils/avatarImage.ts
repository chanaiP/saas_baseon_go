/** 个人头像：压缩为 JPEG Data URL，控制体积以便写入库（Text） */

import { MAX_IMAGE_UPLOAD_BYTES, MAX_IMAGE_UPLOAD_LABEL } from '@/constants/uploadLimits'
const MAX_OUTPUT_CHARS = 55_000
const CANVAS_MAX_SIDE = 128

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

export async function fileToAvatarDataUrl(file: File): Promise<string> {
  if (!file.type.startsWith('image/')) {
    throw new Error('请选择图片文件（JPG、PNG、WebP、GIF 等）')
  }
  if (file.size > MAX_IMAGE_UPLOAD_BYTES) {
    throw new Error(`图片请小于 ${MAX_IMAGE_UPLOAD_LABEL}`)
  }
  const img = await loadImageFromFile(file)
  let w = img.naturalWidth
  let h = img.naturalHeight
  if (w < 1 || h < 1) {
    throw new Error('图片尺寸无效')
  }
  if (w > CANVAS_MAX_SIDE || h > CANVAS_MAX_SIDE) {
    if (w >= h) {
      h = Math.round((h * CANVAS_MAX_SIDE) / w)
      w = CANVAS_MAX_SIDE
    } else {
      w = Math.round((w * CANVAS_MAX_SIDE) / h)
      h = CANVAS_MAX_SIDE
    }
  }
  const canvas = document.createElement('canvas')
  canvas.width = w
  canvas.height = h
  const ctx = canvas.getContext('2d')
  if (!ctx) {
    throw new Error('浏览器不支持图片处理')
  }
  ctx.drawImage(img, 0, 0, w, h)
  let quality = 0.82
  let dataUrl = canvas.toDataURL('image/jpeg', quality)
  while (dataUrl.length > MAX_OUTPUT_CHARS && quality > 0.35) {
    quality -= 0.07
    dataUrl = canvas.toDataURL('image/jpeg', quality)
  }
  if (dataUrl.length > MAX_OUTPUT_CHARS) {
    throw new Error('图片仍过大，请换一张更简单或更小的图')
  }
  return dataUrl
}
