import type { Directive, DirectiveBinding } from 'vue'
import { usePermissionStore } from '@/stores/permission'

export const vPermission: Directive = {
  mounted(el: HTMLElement, binding: DirectiveBinding<string>) {
    const permissionStore = usePermissionStore()
    const code = binding.value

    if (!code || !permissionStore.canUseAction(code)) {
      el.parentNode?.removeChild(el)
    }
  },
}
