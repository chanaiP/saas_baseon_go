import { ElMessageBox } from 'element-plus'

interface ArchiveConfirmOptions {
  title?: string
  name?: string
  message?: string
  detail?: string
  confirmText?: string
}

export async function confirmArchiveAction(options: ArchiveConfirmOptions) {
  const subject = options.name ? `「${options.name}」` : '该记录'
  const message = options.message ?? `确认将${subject}归档？`
  const detail = options.detail ?? '归档后默认不再出现在业务列表中，历史记录、审计链路和关联追溯仍会保留。'
  await ElMessageBox.confirm(`${message}\n\n${detail}`, options.title ?? '归档确认', {
    type: 'warning',
    confirmButtonText: options.confirmText ?? '归档',
    cancelButtonText: '取消',
    distinguishCancelAndClose: true,
  })
}

export function archiveSuccessMessage(name?: string) {
  return name ? `已归档「${name}」` : '已归档'
}
