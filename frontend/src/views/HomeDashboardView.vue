<script setup lang="ts">
defineOptions({ name: 'HomeDashboardView' })
import { ArrowRight, House, Key, Monitor, Setting, Share, User } from '@element-plus/icons-vue'
import { computed } from 'vue'
import { useRouter } from 'vue-router'

import { usePermissionStore } from '@/stores/permission'

const router = useRouter()
const perm = usePermissionStore()

const greeting = computed(() => {
  const h = new Date().getHours()
  if (h < 12) return '上午好'
  if (h < 18) return '下午好'
  return '晚上好'
})

const name = computed(() => perm.profile?.name || perm.profile?.employee_no || '用户')

type QuickItem = { title: string; path: string; desc: string; icon: typeof House }
const quickLinks = computed<QuickItem[]>(() => {
  const items: QuickItem[] = []
  if (perm.canUseMenuPath('/organization')) {
    items.push({ title: '组织架构', path: '/organization', desc: '公司与部门', icon: Share })
  }
  if (perm.canUseMenuPath('/users')) {
    items.push({ title: '用户管理', path: '/users', desc: '账号与权限入口', icon: User })
  }
  if (perm.canUseMenuPath('/roles')) {
    items.push({ title: '角色权限', path: '/roles', desc: '角色与授权', icon: Key })
  }
  if (perm.canUseMenuPath('/monitor/health')) {
    items.push({ title: '健康检查', path: '/monitor/health', desc: 'PostgreSQL / Redis', icon: Monitor })
  }
  if (perm.canUseMenuPath('/tenants')) {
    items.push({ title: '主体管理', path: '/tenants', desc: '租户与配额', icon: Setting })
  }
  return items
})

function go(path: string) {
  void router.push(path)
}
</script>

<template>
  <div class="home pro-page">
    <!-- 仪表盘页：Ant Design Pro 式页头 + 白底块 -->
    <header class="pro-page-header">
      <div>
        <h1 class="pro-page-header__title">{{ greeting }}，{{ name }}</h1>
        <p class="pro-page-header__desc">
          常用入口在下方卡片；更多能力在侧栏。顶栏页签便于在已打开页面间切换。
        </p>
      </div>
    </header>

    <section class="pro-card-block">
      <h2 class="pro-card-block__title">快捷入口</h2>
      <el-row :gutter="16">
        <el-col v-for="item in quickLinks" :key="item.path" :xs="24" :sm="12" :md="8" :lg="6">
          <div class="card" role="button" tabindex="0" @click="go(item.path)" @keydown.enter="go(item.path)">
            <el-icon class="card-icon" :size="28">
              <component :is="item.icon" />
            </el-icon>
            <div class="card-body">
              <div class="card-title">{{ item.title }}</div>
              <div class="card-desc">{{ item.desc }}</div>
            </div>
            <el-icon class="card-arrow"><ArrowRight /></el-icon>
          </div>
        </el-col>
      </el-row>
      <el-empty v-if="!quickLinks.length" description="暂无可用快捷入口（可能受菜单权限限制）" />
    </section>

    <section class="pro-card-block tips">
      <h2 class="pro-card-block__title">使用提示</h2>
      <ul>
        <li>登录后进入首页；常用入口会按当前账号的菜单权限自动展示。</li>
        <li>顶栏支持主题切换、内容区全屏和界面设置，已打开页面可通过页签快速切换。</li>
        <li>侧栏支持收起为图标；在界面设置中切换混合导航后，顶级模块在顶栏展示，子菜单保留在侧栏。</li>
        <li>主体管理、套餐中心、系统监控等平台能力仅对具备对应权限的账号开放。</li>
      </ul>
    </section>
  </div>
</template>

<style scoped>
.home {
  max-width: min(1120px, 100%);
  margin: 0 auto;
}
.card {
  display: flex;
  align-items: center;
  gap: var(--nm-space-md);
  padding: var(--nm-space-lg) var(--nm-space-xl);
  margin-bottom: var(--nm-space-md);
  border-radius: var(--nm-radius-lg);
  border: 1px solid var(--nm-border);
  background: var(--nm-bg-card);
  cursor: pointer;
  transition: all var(--nm-transition-normal);
  box-shadow: var(--nm-shadow-md);
  position: relative;
  overflow: hidden;
}
.card::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 1px;
  background: linear-gradient(90deg, transparent, var(--nm-primary), transparent);
  opacity: 0;
  transition: opacity var(--nm-transition-normal);
}
.card:hover {
  border-color: var(--nm-primary);
  background: var(--nm-bg-elevated);
  box-shadow: var(--nm-shadow-glow);
  transform: translateY(-4px);
}
.card:hover::before {
  opacity: 1;
}
.card:active {
  transform: translateY(-2px);
  box-shadow: var(--nm-shadow-lg);
}
.card-icon {
  color: var(--nm-primary);
  flex-shrink: 0;
  opacity: 0.92;
}
.card-body {
  flex: 1;
  min-width: 0;
}
.card-title {
  font-family: var(--nm-font-ui);
  font-weight: 600;
  font-size: 15px;
  letter-spacing: 0.03em;
  margin-bottom: 4px;
  color: var(--nm-text-primary);
}
.card-desc {
  font-size: 13px;
  color: var(--nm-text-secondary);
  line-height: 1.5;
  font-family: var(--nm-font-body);
}
.card-arrow {
  flex-shrink: 0;
  color: var(--nm-text-muted);
  transition: all var(--nm-transition-normal);
}
.card:hover .card-arrow {
  color: var(--nm-primary);
  transform: translateX(4px);
  transform: translateX(3px);
}
.tips ul {
  margin: 0;
  padding-left: 22px;
  font-size: 14px;
  line-height: 1.72;
  color: var(--el-text-color-secondary);
}
</style>
