<template>
  <div class="channel-management" :class="{ compact }">
    <div v-if="showSummary" class="grid four channel-summary">
      <div class="metric-card"><span>已启用渠道</span><strong>{{ summary.enabled }}</strong><p>可生成发布任务</p></div>
      <div class="metric-card"><span>可发布渠道</span><strong>{{ summary.publishable }}</strong><p>至少 1 个可用账号</p></div>
      <div class="metric-card"><span>待接入</span><strong>{{ summary.pending }}</strong><p>需要补充 URL / 授权</p></div>
      <div class="metric-card"><span>通用审核 Skill</span><strong>已启用</strong><p>按母稿语境动态检查</p></div>
    </div>

    <div class="panel channel-panel">
      <div class="panel-header">
        <div>
          <h3>渠道资料</h3>
          <p>维护渠道类型、发布入口、内容形态、支持方式与默认发布方式</p>
        </div>
        <button v-if="!compact" type="button" class="btn primary" @click="$emit('add-channel')">新增渠道</button>
        <button v-else type="button" class="btn ghost" @click="$emit('open-plans')">去渠道管理</button>
      </div>
      <div class="table-wrap">
        <table class="table channel-table">
          <colgroup>
            <col class="col-channel" />
            <col class="col-type" />
            <col class="col-url" />
            <col class="col-content" />
            <col class="col-support" />
            <col class="col-default" />
            <col class="col-account" />
            <col class="col-status" />
            <col v-if="!compact" class="col-actions" />
          </colgroup>
          <thead>
            <tr>
              <th class="col-channel">渠道</th>
              <th class="col-type">类型</th>
              <th class="col-url">URL / 发布入口</th>
              <th class="col-content">内容形态</th>
              <th class="col-support">支持方式</th>
              <th class="col-default">默认方式</th>
              <th class="col-account">账号</th>
              <th class="col-status">接入状态</th>
              <th v-if="!compact" class="col-actions">操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="channel in channels" :key="channel.name">
              <td class="col-channel">
                <div class="channel-cell">
                  <span class="channel-logo">{{ channel.icon }}</span>
                  <div>
                    <strong>{{ channel.name }}</strong>
                    <p>{{ channel.desc }}</p>
                  </div>
                </div>
              </td>
              <td class="col-type">{{ channel.type }}</td>
              <td class="col-url">
                <div class="channel-links">
                  <a v-if="channel.siteUrl" :href="channel.siteUrl" target="_blank" rel="noreferrer">站点</a>
                  <a v-if="channel.adminUrl" :href="channel.adminUrl" target="_blank" rel="noreferrer">发布后台</a>
                  <span v-if="!channel.siteUrl && !channel.adminUrl" class="muted">—</span>
                </div>
              </td>
              <td class="col-content">{{ channel.contentTypes }}</td>
              <td class="col-support">{{ channel.supportMethods }}</td>
              <td class="col-default">{{ channel.defaultMethod }}</td>
              <td class="col-account">{{ channel.accountCount }}</td>
              <td class="col-status"><span :class="['badge', statusClass(channel.status)]">{{ channel.status }}</span></td>
              <td v-if="!compact" class="col-actions row-actions-inline">
                <button type="button" class="btn small ghost" @click="$emit('configure', channel)">配置</button>
                <button type="button" class="btn small" @click="$emit('test', channel)">测试</button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<script setup>

defineProps({
  channels: { type: Array, required: true },
  summary: { type: Object, required: true },
  compact: { type: Boolean, default: false },
  showSummary: { type: Boolean, default: true }
})

defineEmits(['add-channel', 'open-plans', 'configure', 'test'])

function statusClass(status) {
  if (status === '可发布') return 'success'
  if (status === '待授权') return 'warning'
  if (status === '可生成素材') return 'warning'
  return 'muted-badge'
}
</script>
