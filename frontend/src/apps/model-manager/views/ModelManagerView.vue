<template>
  <div class="model-manager-page">
    <section class="model-manager-toolbar">
      <div>
        <p class="eyebrow">AI / Agent</p>
        <h1>模型管理</h1>
        <p>登记模型资产，维护发布状态，并把菜单、权限和套餐能力纳入底座治理。</p>
      </div>
      <button class="primary-action" type="button">
        <span>+</span>
        新增模型
      </button>
    </section>

    <section class="model-manager-filters" aria-label="模型筛选">
      <label>
        <span>模型名称</span>
        <input v-model="keyword" placeholder="搜索模型名称、编码或供应商" />
      </label>
      <label>
        <span>状态</span>
        <select v-model="status">
          <option value="">全部状态</option>
          <option value="online">已发布</option>
          <option value="draft">配置中</option>
          <option value="archived">已归档</option>
        </select>
      </label>
    </section>

    <section class="model-manager-grid">
      <article v-for="model in filteredModels" :key="model.code" class="model-card">
        <div class="model-card-main">
          <div class="model-icon">{{ model.shortName }}</div>
          <div>
            <h2>{{ model.name }}</h2>
            <p>{{ model.code }} / {{ model.provider }}</p>
          </div>
          <span :class="['status-pill', model.status]">{{ model.statusText }}</span>
        </div>
        <p class="model-desc">{{ model.description }}</p>
        <dl class="model-meta">
          <div>
            <dt>调用方式</dt>
            <dd>{{ model.invokeMode }}</dd>
          </div>
          <div>
            <dt>权限</dt>
            <dd>{{ model.permission }}</dd>
          </div>
          <div>
            <dt>版本</dt>
            <dd>{{ model.version }}</dd>
          </div>
        </dl>
      </article>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'

const keyword = ref('')
const status = ref('')

const models = [
  {
    code: 'base-chat',
    name: '基础对话模型',
    shortName: '基',
    provider: 'platform',
    status: 'online',
    statusText: '已发布',
    description: '面向通用问答、摘要和文本生成的默认模型能力。',
    invokeMode: 'API',
    permission: '租户内授权',
    version: '0.1.0',
  },
  {
    code: 'task-agent',
    name: '任务规划模型',
    shortName: '任',
    provider: 'platform',
    status: 'draft',
    statusText: '配置中',
    description: '用于项目任务拆解、步骤规划和执行建议的模型资产。',
    invokeMode: 'API / Agent',
    permission: '按套餐开通',
    version: '0.1.0',
  },
]

const filteredModels = computed(() => {
  const text = keyword.value.trim().toLowerCase()
  return models.filter((model) => {
    const matchText = !text || [model.name, model.code, model.provider].some((item) => item.toLowerCase().includes(text))
    const matchStatus = !status.value || model.status === status.value
    return matchText && matchStatus
  })
})
</script>

<style scoped>
.model-manager-page {
  min-height: 100%;
  padding: 32px;
  color: var(--text-primary, #f8fafc);
}

.model-manager-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 24px;
  padding: 28px;
  border: 1px solid rgba(45, 212, 191, 0.32);
  border-radius: 8px;
  background: linear-gradient(180deg, rgba(20, 184, 166, 0.14), rgba(15, 23, 42, 0.82));
}

.eyebrow,
.model-manager-toolbar p,
.model-card p,
.model-meta dt {
  margin: 0;
  color: var(--text-secondary, #94a3b8);
}

.eyebrow {
  margin-bottom: 8px;
  font-size: 13px;
  font-weight: 700;
  letter-spacing: 0;
  text-transform: uppercase;
}

h1,
h2 {
  margin: 0;
  letter-spacing: 0;
}

h1 {
  margin-bottom: 8px;
  font-size: 26px;
}

h2 {
  font-size: 18px;
}

.primary-action {
  flex: 0 0 auto;
  display: inline-flex;
  align-items: center;
  gap: 8px;
  height: 42px;
  padding: 0 18px;
  border: 0;
  border-radius: 8px;
  background: linear-gradient(135deg, #22d3ee, #14f1c8);
  color: #06131f;
  font-weight: 800;
  cursor: pointer;
}

.model-manager-filters {
  display: grid;
  grid-template-columns: minmax(260px, 1fr) 220px;
  gap: 16px;
  margin: 24px 0;
}

.model-manager-filters label {
  display: grid;
  gap: 8px;
  font-size: 14px;
  font-weight: 700;
}

.model-manager-filters input,
.model-manager-filters select {
  height: 42px;
  min-width: 0;
  border: 1px solid var(--form-border, rgba(148, 163, 184, 0.24));
  border-radius: 8px;
  background: var(--form-bg, rgba(30, 41, 59, 0.72));
  color: var(--text-primary, #f8fafc);
  font: inherit;
  padding: 0 14px;
}

.model-manager-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(320px, 1fr));
  gap: 18px;
}

.model-card {
  padding: 22px;
  border: 1px solid rgba(148, 163, 184, 0.22);
  border-radius: 8px;
  background: rgba(15, 23, 42, 0.74);
  transition: border-color 0.18s ease, transform 0.18s ease, background 0.18s ease;
}

.model-card:hover {
  transform: translateY(-2px);
  border-color: rgba(45, 212, 191, 0.72);
  background: rgba(20, 38, 58, 0.92);
}

.model-card-main {
  display: grid;
  grid-template-columns: 52px minmax(0, 1fr) auto;
  align-items: center;
  gap: 14px;
}

.model-icon {
  display: grid;
  place-items: center;
  width: 52px;
  height: 52px;
  border: 1px solid rgba(45, 212, 191, 0.52);
  border-radius: 8px;
  background: rgba(20, 184, 166, 0.2);
  color: #2dd4bf;
  font-size: 22px;
  font-weight: 900;
}

.status-pill {
  padding: 6px 12px;
  border-radius: 999px;
  font-size: 13px;
  font-weight: 800;
}

.status-pill.online {
  background: rgba(34, 197, 94, 0.18);
  color: #4ade80;
}

.status-pill.draft {
  background: rgba(245, 158, 11, 0.18);
  color: #fbbf24;
}

.model-desc {
  margin: 18px 0;
  line-height: 1.7;
}

.model-meta {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 12px;
  margin: 0;
}

.model-meta div {
  padding: 12px;
  border: 1px solid rgba(148, 163, 184, 0.18);
  border-radius: 8px;
}

.model-meta dd {
  margin: 6px 0 0;
  font-weight: 800;
}

@media (max-width: 760px) {
  .model-manager-page {
    padding: 18px;
  }

  .model-manager-toolbar {
    align-items: stretch;
    flex-direction: column;
    padding: 20px;
  }

  .model-manager-filters {
    grid-template-columns: 1fr;
  }

  .model-card-main,
  .model-meta {
    grid-template-columns: 1fr;
  }
}
</style>
