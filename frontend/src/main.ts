import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import 'element-plus/theme-chalk/dark/css-vars.css'
import { createPinia, setActivePinia } from 'pinia'
import { createApp, watch } from 'vue'

import App from './App.vue'
import router from './router'
import { useTenantBrandingStore } from './stores/tenantBranding'
import { useUiPreferencesStore } from './stores/uiPreferences'
import { buildDocumentTitle } from './utils/documentTitle'
import { vPermission } from './directives/permission'
import './style.css'

const pinia = createPinia()
setActivePinia(pinia)
useUiPreferencesStore()
useTenantBrandingStore()
const app = createApp(App)
app.use(pinia)
app.use(router)
app.use(ElementPlus, { zIndex: 12000 })
app.directive('permission', vPermission)

const brandStore = useTenantBrandingStore()
function syncDocumentTitle() {
  document.title = buildDocumentTitle(router.currentRoute.value, brandStore.displayName)
}
router.afterEach(() => {
  syncDocumentTitle()
})
watch(
  () => brandStore.displayName,
  () => {
    syncDocumentTitle()
  },
)

app.mount('#app')
syncDocumentTitle()
