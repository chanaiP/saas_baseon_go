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
        const detailLines = (options.detail || '')
          .split('\n')
          .map(line => line.trim())
          .filter(Boolean)
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
                  detailLines.length
                    ? h(
                        'div',
                        { style: { display: 'grid', gap: '8px', color: 'var(--nm-text-secondary, #94a3b8)', fontSize: '15px' } },
                        detailLines.map((line) => {
                          const numbered = line.match(/^(\d+)[.、]\s*(.+)$/)
                          if (!numbered) {
                            return h('p', { style: { margin: '0' } }, line)
                          }
                          return h('div', { style: { display: 'grid', gridTemplateColumns: '26px minmax(0, 1fr)', gap: '8px', alignItems: 'start' } }, [
                            h('span', {
                              style: {
                                display: 'inline-grid',
                                placeItems: 'center',
                                width: '22px',
                                height: '22px',
                                borderRadius: '999px',
                                background: 'rgba(45, 212, 191, .14)',
                                color: 'var(--nm-accent, #2dd4bf)',
                                fontSize: '12px',
                                fontWeight: '800',
                                lineHeight: '1',
                              },
                            }, numbered[1]),
                            h('span', { style: { minWidth: '0' } }, numbered[2]),
                          ])
                        }),
                      )
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
