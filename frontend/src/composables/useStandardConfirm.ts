import { createApp, defineComponent, h, nextTick, ref } from 'vue'

import NeuroAgentDialog from '@/views/components/NeuroAgentDialog.vue'

interface StandardConfirmOptions {
  title?: string
  icon?: string
  message: string
  detail?: string
  confirmText?: string
  cancelText?: string
  confirmIcon?: string
  size?: 'small' | 'medium' | 'large'
  width?: string
}

export function confirmStandardAction(options: StandardConfirmOptions): Promise<void> {
  return new Promise((resolve, reject) => {
    const host = document.createElement('div')
    document.body.appendChild(host)
    let settled = false
    let app: ReturnType<typeof createApp> | null = null

    const cleanup = () => {
      window.setTimeout(() => {
        app?.unmount()
        host.remove()
        app = null
      }, 260)
    }

    const finish = (ok: boolean) => {
      if (settled) return
      settled = true
      if (ok) resolve()
      else reject(new Error('cancel'))
      cleanup()
    }

    const DialogHost = defineComponent({
      setup() {
        const visible = ref(false)
        void nextTick(() => {
          visible.value = true
        })

        const close = (ok: boolean) => {
          visible.value = false
          finish(ok)
        }

        return () =>
          h(
            NeuroAgentDialog,
            {
              modelValue: visible.value,
              'onUpdate:modelValue': (value: boolean) => {
                visible.value = value
                if (!value) finish(false)
              },
              title: options.title ?? '操作确认',
              icon: options.icon ?? '!',
              size: options.size ?? 'small',
              width: options.width,
              confirmText: options.confirmText ?? '确认',
              cancelText: options.cancelText ?? '取消',
              confirmIcon: options.confirmIcon ?? '✓',
              closeOnOverlayClick: false,
              onConfirm: () => close(true),
              onCancel: () => close(false),
              onClose: () => close(false),
            },
            {
              default: () =>
                h('div', { style: { display: 'grid', gap: '12px', lineHeight: '1.75' } }, [
                  h('p', { style: { margin: '0', color: 'var(--nm-text, #f8fafc)', fontSize: '18px', fontWeight: '700' } }, options.message),
                  options.detail
                    ? h('p', { style: { margin: '0', color: 'var(--nm-text-secondary, #94a3b8)', fontSize: '15px' } }, options.detail)
                    : null,
                ]),
            },
          )
      },
    })

    app = createApp(DialogHost)
    app.mount(host)
  })
}
