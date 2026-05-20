<template>
  <div class="ai-geo-page">
    <section class="ai-geo-hero">
      <div>
        <p class="eyebrow">Generative Engine Optimization</p>
        <h1>{{ currentSection.title }}</h1>
        <p>{{ currentSection.description }}</p>
      </div>
      <div class="score-panel" aria-label="AI GEO 应用状态">
        <span>装载阶段</span>
        <strong>Manifest Ready</strong>
        <small>菜单、权限、套餐和配额已声明</small>
      </div>
    </section>

    <section class="ai-geo-nav" aria-label="AI GEO 模块">
      <RouterLink v-for="item in sections" :key="item.key" :to="item.path" :class="{ active: item.key === currentSection.key }">
        <span>{{ item.index }}</span>
        {{ item.title }}
      </RouterLink>
    </section>

    <section class="ai-geo-board">
      <article v-for="card in currentSection.cards" :key="card.title" class="metric-card">
        <span>{{ card.label }}</span>
        <strong>{{ card.value }}</strong>
        <p>{{ card.hint }}</p>
      </article>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { RouterLink, useRoute } from 'vue-router'

const route = useRoute()

const sections = [
  {
    key: 'dashboard',
    index: '01',
    title: 'GEO 总览',
    path: '/ai-geo/dashboard',
    description: '汇总品牌在 AI 问答、生成式搜索和引用来源中的整体可见度。',
    cards: [
      { label: '入口', title: '项目', value: '项目管理', hint: '定义品牌、站点、竞品和监测范围。' },
      { label: '指标', title: '可见度', value: 'Visibility', hint: '后续按项目、提示词和模型聚合得分。' },
      { label: '治理', title: '套餐', value: '3 quotas', hint: '项目数量、监测次数和内容优化次数已声明。' },
    ],
  },
  {
    key: 'projects',
    index: '02',
    title: '项目管理',
    path: '/ai-geo/projects',
    description: '维护租户内 GEO 项目，承载品牌、域名、行业和竞品边界。',
    cards: [
      { label: '权限', title: '维护', value: 'ai_geo:project:manage', hint: '创建和调整 GEO 项目。' },
      { label: '配额', title: '项目数', value: 'ai_geo_project_count', hint: '套餐可配置项目数量上限。' },
      { label: '数据', title: '租户隔离', value: 'TENANT', hint: '后续业务表必须写入 tenant_id。' },
    ],
  },
  {
    key: 'queries',
    index: '03',
    title: '提示词库',
    path: '/ai-geo/queries',
    description: '管理用于监测 AI 搜索结果的行业问题、品牌问题和竞品问题。',
    cards: [
      { label: '权限', title: '维护', value: 'ai_geo:query:manage', hint: '新增、编辑和归档提示词。' },
      { label: '来源', title: '导入', value: 'CSV / API', hint: '后续补充导入和标签分组。' },
      { label: '归属', title: '项目', value: 'Project scoped', hint: '提示词将归属具体 GEO 项目。' },
    ],
  },
  {
    key: 'monitoring',
    index: '04',
    title: '可见度监测',
    path: '/ai-geo/monitoring',
    description: '触发监测任务，采集不同 AI 渠道对品牌、竞品和内容来源的回答表现。',
    cards: [
      { label: '权限', title: '执行', value: 'ai_geo:monitoring:run', hint: '监测任务写操作需审计。' },
      { label: '配额', title: '月度次数', value: 'ai_geo_monthly_monitor_runs', hint: '套餐按月限制监测调用。' },
      { label: '接口', title: '运行记录', value: '/api/ai-geo/monitoring/runs', hint: '读写接口已在 Manifest 声明。' },
    ],
  },
  {
    key: 'citations',
    index: '05',
    title: '引用分析',
    path: '/ai-geo/citations',
    description: '分析 AI 回答中引用的网站、内容片段、品牌露出和来源可信度。',
    cards: [
      { label: '能力', title: '引用源', value: 'Sources', hint: '沉淀引用域名和内容片段。' },
      { label: '指标', title: '覆盖', value: 'Mention rate', hint: '统计品牌与竞品被提及比例。' },
      { label: '权限', title: '访问', value: '/ai-geo/citations', hint: '菜单访问权限已登记。' },
    ],
  },
  {
    key: 'competitors',
    index: '06',
    title: '竞品洞察',
    path: '/ai-geo/competitors',
    description: '比较品牌与竞品在 AI 回答中的排名、推荐理由和引用来源差异。',
    cards: [
      { label: '分析', title: '差距', value: 'Gap', hint: '发现竞品更易被 AI 推荐的原因。' },
      { label: '维度', title: '渠道', value: 'Model / query', hint: '按模型、提示词、主题聚合。' },
      { label: '权限', title: '访问', value: '/ai-geo/competitors', hint: '菜单访问权限已登记。' },
    ],
  },
  {
    key: 'content',
    index: '07',
    title: '内容优化',
    path: '/ai-geo/content',
    description: '基于监测结果生成内容改写、FAQ、结构化资料和引用机会建议。',
    cards: [
      { label: '权限', title: '生成', value: 'ai_geo:content:optimize', hint: '生成建议写操作需审计。' },
      { label: '配额', title: '月度次数', value: 'ai_geo_monthly_content_optimizations', hint: '套餐限制优化生成次数。' },
      { label: '接口', title: '建议', value: '/api/ai-geo/content/optimize', hint: '接口已绑定操作权限。' },
    ],
  },
  {
    key: 'reports',
    index: '08',
    title: '报告中心',
    path: '/ai-geo/reports',
    description: '输出项目周期报告，支持导出监测结果、引用来源和优化建议。',
    cards: [
      { label: '权限', title: '导出', value: 'ai_geo:report:export', hint: '报告导出写操作需审计。' },
      { label: '范围', title: '周期', value: 'Weekly / Monthly', hint: '后续支持项目周期报告。' },
      { label: '接口', title: '导出', value: '/api/ai-geo/reports/export', hint: '接口已绑定操作权限。' },
    ],
  },
]

const currentSection = computed(() => {
  const key = String(route.params.section || 'dashboard')
  return sections.find((item) => item.key === key) || sections[0]
})
</script>

<style scoped>
.ai-geo-page {
  min-height: 100%;
  padding: 28px;
  color: var(--text-primary, #e5edf6);
}

.ai-geo-hero {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 280px;
  gap: 24px;
  align-items: stretch;
  padding: 28px;
  border: 1px solid rgba(34, 197, 94, 0.26);
  border-radius: 8px;
  background:
    linear-gradient(135deg, rgba(9, 27, 34, 0.96), rgba(18, 37, 43, 0.9)),
    radial-gradient(circle at 82% 10%, rgba(74, 222, 128, 0.18), transparent 36%);
}

.eyebrow,
.ai-geo-hero p,
.score-panel span,
.score-panel small,
.metric-card span,
.metric-card p {
  margin: 0;
  color: var(--text-secondary, #9fb2c7);
}

.eyebrow {
  margin-bottom: 10px;
  font-size: 12px;
  font-weight: 800;
  letter-spacing: 0;
  text-transform: uppercase;
}

h1 {
  margin: 0 0 10px;
  font-size: 28px;
  letter-spacing: 0;
}

.score-panel {
  display: grid;
  align-content: center;
  gap: 8px;
  padding: 22px;
  border: 1px solid rgba(148, 163, 184, 0.22);
  border-radius: 8px;
  background: rgba(3, 15, 20, 0.66);
}

.score-panel strong {
  font-size: 24px;
  letter-spacing: 0;
}

.ai-geo-nav {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(160px, 1fr));
  gap: 10px;
  margin: 20px 0;
}

.ai-geo-nav a {
  display: flex;
  align-items: center;
  gap: 10px;
  min-height: 44px;
  padding: 0 14px;
  border: 1px solid rgba(148, 163, 184, 0.2);
  border-radius: 8px;
  color: var(--text-primary, #e5edf6);
  text-decoration: none;
  background: rgba(15, 23, 42, 0.52);
}

.ai-geo-nav a.active {
  border-color: rgba(34, 197, 94, 0.78);
  background: rgba(20, 83, 45, 0.28);
}

.ai-geo-nav span {
  color: #86efac;
  font-size: 12px;
  font-weight: 800;
}

.ai-geo-board {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(260px, 1fr));
  gap: 16px;
}

.metric-card {
  min-height: 152px;
  padding: 22px;
  border: 1px solid rgba(148, 163, 184, 0.2);
  border-radius: 8px;
  background: rgba(15, 23, 42, 0.62);
}

.metric-card strong {
  display: block;
  margin: 10px 0;
  font-size: 20px;
  letter-spacing: 0;
  overflow-wrap: anywhere;
}

@media (max-width: 760px) {
  .ai-geo-page {
    padding: 16px;
  }

  .ai-geo-hero {
    grid-template-columns: 1fr;
  }
}
</style>
