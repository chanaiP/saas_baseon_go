<script setup lang="ts">
/**
 * 标准管理页骨架：第一层标题/说明/统计 + 右侧操作；第二层为内容区（通常内含搜索+卡片或表格）。
 * 与「主体管理」页眉 + 面板结构一致，供各业务页复用。
 */
defineOptions({ name: 'NeuroAgentPageShell' })

withDefaults(defineProps<{
  showHero?: boolean
}>(), {
  showHero: true,
})
</script>

<template>
  <div class="neuro-page-shell">
    <header v-if="showHero" class="neuro-page-shell__hero">
      <div class="neuro-page-shell__hero-main">
        <h2 v-if="$slots.title" class="neuro-page-shell__title">
          <slot name="title" />
        </h2>
        <p v-if="$slots.subtitle" class="neuro-page-shell__subtitle">
          <slot name="subtitle" />
        </p>
        <div v-if="$slots.meta" class="neuro-page-shell__meta">
          <slot name="meta" />
        </div>
      </div>
      <div v-if="$slots.actions" class="neuro-page-shell__actions">
        <slot name="actions" />
      </div>
    </header>

    <section class="neuro-page-shell__panel">
      <div class="neuro-page-shell__panel-body">
        <slot />
      </div>
    </section>
  </div>
</template>

<style scoped>
@import '@/styles/theme/neuro-theme.css';

.neuro-page-shell {
  flex: 1 1 auto;
  min-height: 0;
  width: 100%;
  max-width: 100%;
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding: 0;
}

.neuro-page-shell__hero {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  flex-shrink: 0;
  flex-wrap: wrap;
  padding: 16px 20px;
  border: none;
  border-radius: var(--neuro-radius-xl);
  backdrop-filter: blur(16px);
  background-color: color-mix(in srgb, var(--neuro-surface) 93%, var(--neuro-background));
  background-image:
    radial-gradient(120% 90% at 0% 0%, color-mix(in srgb, var(--neuro-primary) 16%, transparent), transparent 55%),
    radial-gradient(90% 70% at 100% 0%, color-mix(in srgb, var(--neuro-accent) 12%, transparent), transparent 50%);
}

.neuro-page-shell__hero-main {
  min-width: 0;
}

.neuro-page-shell__title {
  margin: 0 0 6px;
  font-size: 18px;
  font-weight: 700;
  color: var(--neuro-text);
  letter-spacing: 0.02em;
}

.neuro-page-shell__subtitle {
  margin: 0 0 10px;
  font-size: 13px;
  color: var(--neuro-text-secondary);
  line-height: 1.5;
  max-width: 640px;
}

.neuro-page-shell__meta {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 6px;
  font-size: 12px;
  color: var(--neuro-text-secondary);
}

.neuro-page-shell__actions {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: flex-end;
  gap: 10px;
  flex-shrink: 0;
}

.neuro-page-shell__panel {
  flex: 1 1 auto;
  min-height: 0;
  display: flex;
  flex-direction: column;
  border: none;
  border-radius: var(--neuro-radius-xl);
  overflow: visible;
  backdrop-filter: blur(16px);
  background-color: color-mix(in srgb, var(--neuro-surface) 86%, var(--neuro-background));
  background-image: linear-gradient(
    168deg,
    color-mix(in srgb, var(--neuro-primary) 6%, transparent) 0%,
    transparent 42%
  );
}

.neuro-page-shell__panel-body {
  flex: 1 1 auto;
  min-height: 0;
  display: flex;
  flex-direction: column;
  min-width: 0;
  overflow: visible;
}

/* 由内容自然撑高，交给浏览器主滚动条；勿 flex-basis:0 锁死高度导致裁切或内部滚动条 */
.neuro-page-shell__panel-body > * {
  flex: 0 1 auto;
  min-width: 0;
  min-height: 0;
}
</style>
