<template>
  <div class="geo-growth-page">
    <main class="main">
      <header class="topbar">
        <div>
          <h2>{{ activeTitle }}</h2>
          <p>{{ activeSubtitle }}</p>
        </div>
      </header>

      <section v-if="activeMenu === 'overview'" class="page overview-page">
        <div class="grid four">
          <MetricCard title="品牌数" :value="brands.length" desc="当前租户已维护品牌" />
          <MetricCard title="商品数" :value="productCount" desc="SPU 商品资料卡" />
          <MetricCard title="SKU 数" :value="skuCount" desc="可售卖规格明细" />
          <MetricCard title="资料完整度" :value="brandOverallCompleteness(currentBrand) + '%'" desc="品牌 / 商品 / SKU 汇总" />
        </div>

        <div class="grid overview-row">
          <div class="panel overview-card">
            <div class="panel-header compact">
              <div>
                <h3>今日生产进度</h3>
                <p>母稿、渠道内容、发布计划状态</p>
              </div>
              <button class="btn text" @click="activeMenu = 'drafts'">看母稿</button>
            </div>
            <div class="overview-card-body status-grid">
              <div v-for="item in progressStats" :key="item.label" class="status-card">
                <span>{{ item.label }}</span>
                <strong>{{ item.value }}</strong>
                <p>{{ item.desc }}</p>
              </div>
            </div>
          </div>
          <div class="panel overview-card">
            <div class="panel-header compact">
              <div>
                <h3>待处理事项</h3>
                <p>运营今天需要优先处理的动作。</p>
              </div>
            </div>
            <div class="overview-card-body pending-task-list">
              <button
                v-for="task in pendingTasks"
                :key="task.id"
                type="button"
                class="pending-task-item"
                @click="openPendingTask(task)"
              >
                <span>{{ task.tag }}</span>
                <strong>{{ task.title }}</strong>
              </button>
            </div>
          </div>
          <div class="panel overview-card">
            <div class="panel-header">
              <div>
                <h3>资料缺失提醒</h3>
                <p>优先补齐会影响内容生成的商品字段</p>
              </div>
              <button class="btn text" @click="activeMenu = 'data'">去资料中心</button>
            </div>
            <div class="overview-card-body issue-list">
              <div v-for="issue in dataIssues" :key="issue.id" class="issue-row">
                <span class="badge warning">{{ issue.level }}</span>
                <div>
                  <strong>{{ issue.title }}</strong>
                  <p>{{ issue.desc }}</p>
                </div>
                <button class="btn small" @click="activeMenu = 'data'">处理</button>
              </div>
            </div>
          </div>
        </div>

        <ChannelManagementPanel
          :channels="channelProfiles"
          compact
          @open-plans="goChannelManagement"
          @configure="configureChannel"
          @test="testChannel"
        />
      </section>

      <section v-if="activeMenu === 'workbench'" class="page workbench-page">
        <div class="workspace-layout">
          <div class="chat-panel panel geo-command-panel" :class="{ 'is-chatting': hasStartedWorkbenchChat }">
            <div v-if="!hasStartedWorkbenchChat" class="workbench-brief">
              <span class="badge muted-badge">GEO Article Console</span>
              <h3>按资料和内容类型生成有指向性的 GEO 文章</h3>
              <p>直接说想法也可以；我会基于品牌、商品、内容类型和热点资料，把模糊想法整理成可生成的 GEO 母稿。</p>
            </div>

            <div v-if="!hasStartedWorkbenchChat" class="context-strip setup-strip">
              <div class="context-card strong">
                <span>品牌</span>
                <strong>{{ selectedWorkbenchBrand?.name || '未选择品牌' }}</strong>
                <p>{{ selectedWorkbenchBrand?.position || '请选择品牌资料作为生成底座' }}</p>
                <select v-model="workbench.brandId">
                  <option value="">请选择</option>
                  <option v-for="b in brands" :value="b.id" :key="b.id">{{ b.name }}</option>
                </select>
              </div>
              <div class="context-card">
                <span>商品</span>
                <strong>{{ selectedWorkbenchProduct?.name || '未指定商品' }}</strong>
                <p>{{ selectedWorkbenchProduct?.sellingPoints || (selectedWorkbenchBrand ? '可先用品牌资料自由生成' : '先选择品牌后再指定商品') }}</p>
                <select v-model="workbench.productId">
                  <option value="">请选择</option>
                  <option v-for="p in selectedWorkbenchBrand?.products || []" :value="p.id" :key="p.id">{{ p.name }}</option>
                </select>
              </div>
              <div class="context-card">
                <span>内容类型</span>
                <strong>{{ selectedSkillProfile.name }}</strong>
                <p>{{ selectedSkillProfile.goal }} · {{ selectedSkillProfile.output }}</p>
                <select v-model="workbench.contentType">
                  <option v-for="type in contentTypeOptions" :key="type.value" :value="type.value">{{ type.label }}</option>
                </select>
              </div>
              <div class="context-card">
                <span>热点</span>
                <strong>{{ workbench.hotspot?.title || '不引用热点' }}</strong>
                <p>{{ workbench.hotspot?.summary || '可在对话中再决定是否借势' }}</p>
                <select v-model="selectedHotspotId">
                  <option value="">请选择</option>
                  <option v-for="h in hotspots" :key="h.id" :value="h.id">{{ h.title }} · {{ h.platform }}</option>
                  <option value="__extract__">输入 URL 提炼热点…</option>
                  <option value="__manage__">热点管理…</option>
                </select>
              </div>
              <div class="context-card">
                <span>参考写作风格</span>
                <strong>{{ workbench.styleTemplate?.templateName || '不套用风格' }}</strong>
                <p>{{ selectedStyleTemplateSummary || '可选择已保存模板，或输入 URL 提炼写法结构' }}</p>
                <select v-model="selectedStyleTemplateId">
                  <option value="">请选择</option>
                  <option v-for="style in styleTemplates" :key="style.id" :value="style.id">{{ style.templateName }} · {{ style.platform || '通用' }}</option>
                  <option value="__extract__">输入 URL 提炼风格…</option>
                  <option value="__manage__">风格管理…</option>
                </select>
              </div>
            </div>

            <div v-if="hasStartedWorkbenchChat" class="conversation-context-bar">
              <span>品牌 <strong>{{ selectedWorkbenchBrand?.name || '未选择' }}</strong></span>
              <span>商品 <strong>{{ selectedWorkbenchProduct?.name || '未指定' }}</strong></span>
              <span>内容类型 <strong>{{ selectedSkillProfile.name }}</strong></span>
              <span>热点 <strong>{{ workbench.hotspot?.title || '未引用' }}</strong></span>
              <span>风格 <strong>{{ workbench.styleTemplate?.templateName || '未引用' }}</strong></span>
              <button type="button" class="btn small primary reset-workbench-btn context-reset-btn" :disabled="chatStreaming || loading.action" @click="resetWorkbenchSession">新建母稿</button>
            </div>

            <div v-if="!hasStartedWorkbenchChat && hasWorkbenchContext" class="evidence-list">
              <div v-for="item in workbenchEvidence" :key="item.label" class="evidence-item">
                <span>{{ item.label }}</span>
                <strong>{{ item.value }}</strong>
              </div>
            </div>

            <div
              v-if="hasStartedWorkbenchChat"
            ref="workbenchChatLogRef"
            class="chat-log geo-dialogue"
            @scroll="handleWorkbenchChatScroll"
            @change="handleMessageInlineChoice"
          >
              <div v-if="decisionCard.recommendedDirection || aiWorkStage !== 'idle' || decisionCard.nextStepLabel || hasEditingDraftContent" class="conversation-workflow-row">
                <span v-if="directionDisplayTitle" class="wide">方向 <strong>{{ directionDisplayTitle }}</strong></span>
                <span v-if="aiWorkStage !== 'idle'">阶段 <strong>{{ aiStageLabel }}</strong></span>
                <span v-if="decisionCard.nextStepLabel">下一步 <strong>{{ decisionCard.nextStepLabel }}</strong></span>
                <div v-if="hasEditingDraftContent" class="draft-quick-actions">
                  <button
                    v-for="action in draftQuickEditActions"
                    :key="action.label"
                    type="button"
                    class="btn small ghost"
                    :disabled="chatStreaming || loading.action"
                    @click="requestDraftQuickEdit(action)"
                  >{{ action.label }}</button>
                </div>
              </div>
              <div v-for="msg in visibleChatMessages" :key="msg.id" :class="['bubble', msg.role]">
                <template v-if="msg.role === 'ai'">
                  <div
                    v-if="clarificationChoiceOptions(msg).length"
                    class="message-direction-choices message-clarification-choices"
                    role="radiogroup"
                    aria-label="选择一个回答"
                  >
                    <label
                      v-for="option in clarificationChoiceOptions(msg)"
                      :key="option.key"
                      :class="['message-direction-choice', 'message-clarification-choice', { selected: selectedDirectionChoiceByMessage[msg.id] === option.key }]"
                    >
                      <input
                        type="radio"
                        :name="`clarification-choice-${msg.id}`"
                        :value="option.key"
                        :checked="selectedDirectionChoiceByMessage[msg.id] === option.key"
                        @change="selectClarificationChoice(msg, option)"
                      />
                      <span class="direction-choice-text">
                        <strong>{{ option.displayLabel }}</strong>
                        <em>{{ option.title }}</em>
                      </span>
                    </label>
                  </div>
                  <div v-if="msg.text" class="message-rich" v-html="formatChatMessage(msg.text, msg)"></div>
                  <div v-else-if="msg.streaming" class="ai-waiting-indicator" aria-live="polite">
                    <span class="ai-waiting-orb"></span>
                    <span class="ai-waiting-text">AI 正在组织内容</span>
                    <span class="ai-waiting-dots"><i></i><i></i><i></i></span>
                  </div>
                </template>
                <p v-else-if="msg.role === 'system'">{{ msg.text }}</p>
                <p v-else>{{ msg.text }}</p>
                <span v-if="msg.streaming" class="stream-cursor"></span>
                <div v-if="shouldShowReplyActionsAfterMessage(msg)" class="ai-reply-actions">
                  <button
                    v-for="action in messageReplyActions(msg)"
                    :key="action.id || action.type"
                    :class="['btn', 'small', action.variant || 'ghost']"
                    :disabled="loading.action || chatStreaming"
                    @click="handleReplyAction(action, msg)"
                  >
                    {{ action.label }}
                  </button>
                </div>
              </div>
            </div>
            <div class="chat-input">
              <textarea
                ref="workbenchPromptRef"
                v-model="workbench.prompt"
                :placeholder="generationPlaceholder"
                @keydown.enter.exact.prevent="sendWorkbenchMessage"
              ></textarea>
              <div class="chat-input-actions">
                <button class="btn send-btn" :disabled="!hasWorkbenchInput || chatStreaming" @click="sendWorkbenchMessage">
                  {{ chatStreaming ? '输出中' : '发送' }}
                </button>
              </div>
            </div>
          </div>

          <div class="editor-panel panel">
            <div class="editor-toolbar">
              <div>
                <h3>母稿编辑器</h3>
                <p>{{ editingDraft.status }} · {{ editingDraft.source }}</p>
              </div>
              <div class="editor-toolbar-actions">
                <div class="segmented">
                  <button :class="{ active: previewMode === 'mobile' }" @click="previewMode = 'mobile'">手机预览</button>
                  <button :class="{ active: previewMode === 'pc' }" @click="previewMode = 'pc'">PC 预览</button>
                  <button :class="{ active: previewMode === 'edit' }" @click="previewMode = 'edit'">编辑</button>
                </div>
              </div>
            </div>
            <div v-if="previewMode === 'edit'" class="draft-editor">
              <div class="draft-hero">
                <div class="draft-meta-col">
                  <label>标题<input v-model="editingDraft.title" /></label>
                  <label>摘要<textarea v-model="editingDraft.summary" class="summary-editor" rows="2"></textarea></label>
                  <label>关键词<input v-model="editingDraft.keywordsText" placeholder="多个关键词用逗号分隔" /></label>
                </div>
              </div>
              <label>正文<textarea class="body-editor" v-model="editingDraft.body"></textarea></label>
            </div>
            <PreviewPane v-else :mode="previewMode" :title="editingDraft.title" :summary="editingDraft.summary" :body="editingDraft.body" :loading="draftPreviewLoading" />
            <div class="editor-footer">
              <div v-if="draftSourceRecords.length && draftSourceExpanded" class="draft-source-record expanded">
                <div class="source-record-grid">
                  <div v-for="record in draftSourceRecords" :key="record.label" class="source-record-item">
                    <span>{{ record.label }}</span>
                    <strong>{{ record.title }}</strong>
                    <p>{{ record.desc }}</p>
                  </div>
                </div>
              </div>
              <div class="editor-footer-actions">
                <button
                  v-if="draftSourceRecords.length"
                  type="button"
                  class="source-record-toggle"
                  @click="draftSourceExpanded = !draftSourceExpanded"
                >
                  <span>母稿生成依据</span>
                  <em>{{ draftSourceSummary }}</em>
                  <strong>{{ draftSourceExpanded ? '收起' : '展开' }}</strong>
                </button>
                <div class="editor-actions">
                  <button v-if="canShowSaveDraft" class="btn ghost" :disabled="!canEditDraftActions" @click="saveDraft">保存草稿</button>
                  <button v-if="canShowApprovedChannelAction" class="btn primary" :disabled="loading.action" @click="saveDraftAndOpenChannelGeneration">生成渠道内容</button>
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>

      <section v-if="activeMenu === 'drafts'" class="page drafts-page">
        <div class="panel">
          <div class="panel-header">
            <div>
              <h3>母稿 / 渠道内容</h3>
              <p>按日期管理；母稿可回到工作台深度 AI 编辑，渠道内容轻量编辑并预览</p>
            </div>
            <div class="filters">
              <select v-model="draftDate" class="date-filter">
                <option value="">全部日期</option>
                <option v-for="d in draftDateOptions" :key="d" :value="d">{{ d }}</option>
              </select>
              <select v-model="draftFilter"><option>全部状态</option><option>草稿</option><option>已通过</option><option>已生成渠道内容</option></select>
            </div>
          </div>
          <div v-if="loading.page" class="timeline-empty muted">正在加载母稿与渠道内容...</div>
          <div v-else-if="!draftTimelineGroups.length" class="timeline-empty muted">当前筛选条件下暂无母稿</div>
          <div v-else class="draft-timeline">
            <div v-for="(group, groupIndex) in draftTimelineGroups" :key="group.date" class="draft-timeline-group" :class="{ last: groupIndex === draftTimelineGroups.length - 1 }">
              <div class="draft-timeline-axis">
                <time class="timeline-date">
                  <strong>{{ formatTimelineDay(group.date) }}</strong>
                  <span>{{ formatTimelineWeekday(group.date) }}</span>
                </time>
                <span class="timeline-dot" aria-hidden="true"></span>
                <span v-if="groupIndex !== draftTimelineGroups.length - 1" class="timeline-line" aria-hidden="true"></span>
              </div>
              <div class="draft-timeline-items">
                <div v-for="draft in group.drafts" :key="draft.id" class="draft-node" :class="{ expanded: isDraftExpanded(draft) }">
                  <div class="draft-main" role="button" tabindex="0" @click="toggleDraftChannels(draft)" @keydown.enter.prevent="toggleDraftChannels(draft)">
                    <div>
                      <div class="draft-title-line">
                        <h3>{{ draft.title }}</h3>
                        <span class="draft-status-chip" :class="draftStatusTone(draft.status)">{{ draft.status }}</span>
                      </div>
                      <p>{{ draft.summary }}</p>
                      <p class="draft-generated-at">生成时间：{{ draft.generatedAtLabel }}</p>
                      <div class="tagline"><span class="tag">{{ draft.source }}</span><span class="tag">{{ draft.flow }}</span></div>
                      <div class="draft-channel-summary" :class="{ empty: !draftChannelStats(draft).total }">
                        <div class="draft-channel-summary-head">
                          <span class="draft-channel-summary-label">渠道内容</span>
                          <strong>{{ draftChannelSummaryText(draft) }}</strong>
                          <span v-if="draftChannelPlatformText(draft)" class="draft-channel-platforms">{{ draftChannelPlatformText(draft) }}</span>
                        </div>
                        <div v-if="draftChannelStats(draft).total" class="draft-channel-stat-grid">
                          <span
                            v-for="item in draftChannelStatItems(draft)"
                            :key="item.key"
                            class="draft-channel-stat"
                            :class="item.tone"
                          >
                            <em>{{ item.label }}</em>
                            <strong>{{ item.value }}</strong>
                          </span>
                        </div>
                        <div v-else class="draft-channel-empty-hint">展开后可生成知乎、小红书等渠道内容稿。</div>
                      </div>
                    </div>
                    <div class="row-actions" @click.stop>
                      <button class="btn small ghost" @click="viewDraft(draft)">查看</button>
                      <button class="btn small ghost editor-entry-btn" @click="editDraftInWorkbench(draft)">进入编辑器</button>
                      <button v-if="canManageChannelContent && canGenerateChannelFromDraft(draft)" class="btn small dark" @click="generateChannelsForDraft(draft)">生成渠道内容</button>
                    </div>
                  </div>
                  <div v-if="isDraftExpanded(draft) && draft.channels.length" class="channel-list">
                    <div class="channel-list-label">渠道内容</div>
                    <div v-for="channel in draft.channels" :key="channel.id" class="channel-row">
                      <div class="channel-row-leading">
                        <ChannelLogo :name="channel.channel" />
                        <div class="channel-row-text">
                          <span class="channel-name">{{ channel.channel }}</span>
                          <span class="channel-title-line">
                            <strong class="channel-title">{{ channel.title }}</strong>
                            <span v-if="!isChannelConfirmed(channel) && !isChannelInPublishPlan(channel)" class="badge warning">待确认</span>
                            <span v-if="isChannelInPublishPlan(channel)" class="badge success">已加入发布计划</span>
                          </span>
                        </div>
                      </div>
                      <button v-if="!isChannelInPublishPlan(channel)" class="btn small ghost editor-entry-btn" @click="openChannelContentEditor(draft, channel)">编辑</button>
                      <button class="btn small ghost" @click="openChannelPreview(draft, channel)">预览</button>
                      <button v-if="canShowChannelConfirm(channel)" class="btn small ghost" @click="confirmChannel(channel)">确认</button>
                      <button v-if="canShowChannelAddPlan(channel)" class="btn small dark" @click="addChannelToPlan(draft, channel)">加入发布计划</button>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
          <div v-if="hasMoreDrafts" ref="draftLoadMoreRef" class="draft-load-more">
            <span>继续向下滚动加载更多母稿</span>
          </div>
        </div>
      </section>

      <section v-if="activeMenu === 'plans'" class="page plans-page">
        <div class="tabs plan-tabs">
          <button v-for="tab in planTabs" :key="tab" :class="{ active: planTab === tab }" @click="planTab = tab">{{ tab }}</button>
        </div>

        <template v-if="planTab === '发布日历'">
          <div class="panel">
            <div class="panel-header compact">
              <div>
                <h3>发布日历</h3>
                <p>按日期查看已排期与待发布任务</p>
              </div>
              <input type="date" v-model="planDate" class="date-input" />
            </div>
            <div class="publish-calendar">
              <div v-for="day in calendarDays" :key="day.date" class="calendar-day" :class="{ active: day.date === planDate }">
                <div class="calendar-day-head">
                  <strong>{{ day.label }}</strong>
                  <span>{{ day.plans.length }} 项</span>
                </div>
                <div v-if="day.plans.length" class="calendar-events">
                  <button
                    v-for="plan in day.plans"
                    :key="plan.id"
                    type="button"
                    class="calendar-event"
                    @click="openPublishPlanDrawer(plan)"
                  >
                    <span>{{ plan.time }}</span>
                    <p>{{ plan.title }}</p>
                    <small>{{ plan.channel }} · {{ plan.status }}</small>
                  </button>
                </div>
                <p v-else class="muted calendar-empty">暂无计划</p>
              </div>
            </div>
          </div>
        </template>

        <template v-else-if="planTab === '发布队列'">
          <div class="publish-queue">
            <div class="grid four queue-summary">
              <div class="metric-card"><span>今日任务</span><strong>{{ planQueueSummary.total }}</strong><p>按计划日期筛选</p></div>
              <div class="metric-card"><span>可发布</span><strong>{{ planQueueSummary.ready }}</strong><p>账号和素材均通过</p></div>
              <div class="metric-card"><span>待人工确认</span><strong>{{ planQueueSummary.manual }}</strong><p>Agent / 草稿箱发布需确认</p></div>
              <div class="metric-card"><span>阻断任务</span><strong>{{ planQueueSummary.blocked }}</strong><p>账号或素材未通过</p></div>
            </div>

            <div class="panel">
              <div class="panel-header compact">
                <div>
                  <h3>发布队列</h3>
                  <p>按时间查看各渠道发布任务的前置检查与发布状态</p>
                </div>
                <div class="filters">
                  <input type="date" v-model="planDate" class="date-input" />
                  <button v-if="canManagePublishPlan" class="btn primary" @click="openNewPlanModal">新建发布计划</button>
                </div>
              </div>
              <div class="table-wrap">
                <table class="table queue-table">
                  <colgroup>
                    <col class="queue-col-time" />
                    <col class="queue-col-channel" />
                    <col class="queue-col-account" />
                    <col class="queue-col-content" />
                    <col class="queue-col-method" />
                    <col class="queue-col-account-status" />
                    <col class="queue-col-material" />
                    <col class="queue-col-publish" />
                    <col class="queue-col-actions" />
                  </colgroup>
                  <thead>
                    <tr>
                      <th>时间</th>
                      <th>渠道</th>
                      <th>账号</th>
                      <th>内容</th>
                      <th>发布方式</th>
                      <th>账号状态</th>
                      <th>素材</th>
                      <th>发布状态</th>
                      <th>操作</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr v-for="plan in filteredPlans" :key="plan.id">
                      <td class="queue-time">{{ plan.time }}</td>
                      <td>
                        <div class="channel-cell compact">
                          <ChannelLogo :name="plan.channel" />
                          <strong>{{ plan.channel }}</strong>
                        </div>
                      </td>
                      <td>{{ plan.accountName }}</td>
                      <td class="queue-content">
                        <strong>{{ plan.title }}</strong>
                        <p>负责人：{{ plan.owner }}</p>
                      </td>
                      <td class="queue-method">{{ publishModeLabel(plan.method) }}</td>
                      <td><span :class="['badge', queueStatusClass(plan.accountStatus)]">{{ plan.accountStatus }}</span></td>
                      <td><span :class="['badge', queueStatusClass(plan.materialStatus)]">{{ plan.materialStatus }}</span></td>
                      <td><span :class="['badge', queueStatusClass(plan.publishStatus)]">{{ plan.publishStatus }}</span></td>
                      <td>
                        <div class="queue-actions">
                          <button v-if="canManagePublishPlan && canPreCheckPlan(plan)" type="button" class="btn small ghost" @click="preCheckPlan(plan)">前置检查</button>
                          <button v-if="canManagePublishPlan && canCompletePlan(plan)" type="button" class="btn small dark" @click="completePlan(plan)">完成</button>
                        </div>
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </div>
          </div>
        </template>

      </section>

      <section v-if="activeMenu === 'channels'" class="page channels-page">
        <ChannelManagementPanel
          :channels="channelProfiles"
          @add-channel="addChannel"
          @configure="configureChannel"
          @test="testChannel"
        />
        <div class="panel channel-accounts-panel">
          <div class="panel-header">
            <div>
              <h3>渠道账号</h3>
              <p>管理各渠道的授权账号与登录态</p>
            </div>
            <button v-if="canManageChannelAccount" class="btn primary" @click="showToast('新增渠道账号（原型）')">新增账号</button>
          </div>
          <div class="table-wrap">
            <table class="table">
              <thead>
                <tr><th>渠道</th><th>账号名称</th><th>登录标识</th><th>状态</th><th>最近更新</th><th>操作</th></tr>
              </thead>
              <tbody>
                <tr v-if="!channelAccounts.length">
                  <td colspan="6" class="timeline-empty muted">暂无渠道账号。请通过“新增账号”接入真实授权账号。</td>
                </tr>
                <tr v-for="account in channelAccounts" :key="account.id">
                  <td>
                    <div class="channel-cell compact">
                      <ChannelLogo :name="account.channel" :code="account.channelCode" />
                      <strong>{{ account.channel }}</strong>
                    </div>
                  </td>
                  <td>{{ account.accountName }}</td>
                  <td>{{ account.login }}</td>
                  <td><span :class="['badge', account.status === '已授权' ? 'success' : 'warning']">{{ account.status }}</span></td>
                  <td>{{ account.updatedAt }}</td>
                  <td class="row-actions-inline">
                    <button v-if="canManageChannelAccount" class="btn small ghost" @click="showToast(`配置 ${account.accountName}`)">配置</button>
                    <button v-if="canManageChannelAccount" class="btn small" @click="showToast(`重新授权 ${account.accountName}`)">授权</button>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </section>

      <section v-if="activeMenu === 'data'" class="page data-page">
        <div class="panel">
          <div class="panel-header">
            <div>
              <h3>资料中心</h3>
              <p>品牌第一维度；品牌下包含商品资料卡、SKU 资料、SKU 覆盖、竞品信息</p>
            </div>
            <div class="top-actions">
              <button v-if="canImportData" class="btn ghost" @click="openImportModal">动态导入</button>
              <button v-if="canImportData" class="btn primary" @click="openBrandModal">新增品牌</button>
            </div>
          </div>
          <div class="brand-grid">
            <div v-for="brand in brands" :key="brand.id" :class="['brand-card', { active: selectedBrandId === brand.id }]" @click="selectedBrandId = brand.id">
              <h3>{{ brand.name }}</h3>
              <p>{{ brandCardMetaText(brand) }}</p>
              <div class="brand-stats"><span>{{ brand.products.length }} 商品</span><span>{{ countSkus(brand) }} SKU</span><span>综合 {{ brandOverallCompleteness(brand) }}%</span></div>
              <div class="progress" :title="brandCompletenessBreakdownText(brand)"><span :style="{ width: brandOverallCompleteness(brand) + '%' }"></span></div>
            </div>
            <div v-if="!brands.length" class="timeline-empty data-empty-state">
              <strong>暂无品牌资料</strong>
              <p>请先新增品牌，或通过动态导入创建品牌、商品、SKU 和竞品资料。</p>
            </div>
          </div>
        </div>

        <div class="panel brand-detail">
          <div class="panel-header compact">
            <div>
              <h3>{{ brandDetailTitle }}</h3>
              <p>{{ brandMetaText }}</p>
            </div>
            <div class="row-actions">
              <button v-if="canImportData && hasBrandData" class="btn small ghost" @click="openBrandModal(currentBrand)">编辑品牌</button>
              <button v-if="canImportData && hasBrandData" class="btn small dark" @click="generateBrandKeywords">生成关键词</button>
              <button v-if="canImportData && !hasBrandData" class="btn small primary" @click="openBrandModal">新增品牌</button>
            </div>
          </div>
          <div class="tabs">
            <button v-for="tab in dataTabs" :key="tab" :class="{ active: dataTab === tab }" @click="dataTab = tab">{{ tab }}</button>
          </div>
          <div v-if="dataTab === '商品资料'" class="products-table">
            <div v-if="!currentBrand.products.length" class="timeline-empty">
              当前品牌暂无商品资料。新增商品后，可继续维护 SKU、竞品和内容关键词。
            </div>
            <table v-else class="table">
              <thead><tr><th>商品</th><th>类目/系列</th><th>SKU</th><th>资料完整度</th><th>竞品</th><th>缺失项</th><th>操作</th></tr></thead>
              <tbody>
                <tr v-for="product in currentBrand.products" :key="product.id">
                  <td><strong>{{ product.name }}</strong><p>{{ product.code }}</p></td>
                  <td>{{ product.category }} / {{ product.series }}</td>
                  <td>{{ product.skus.length }}</td>
                  <td><div class="inline-progress"><span :style="{ width: product.completeness + '%' }"></span></div>{{ product.completeness }}%</td>
                  <td>{{ product.competitors.length }} 个</td>
                  <td>{{ product.missing.join('、') || '无' }}</td>
                  <td><button class="btn small" @click="openProductDrawer(product)">查看资料卡</button></td>
                </tr>
              </tbody>
            </table>
          </div>
          <div v-else-if="dataTab === '品牌信息'" class="info-grid">
            <InfoBlock title="品牌定位" :value="currentBrand.position" />
            <InfoBlock title="目标人群" :value="currentBrand.audience" />
            <InfoBlock title="价格带" :value="currentBrand.priceBand" />
          </div>
          <div v-else-if="dataTab === '关键词'" class="keyword-group-list">
            <div v-if="!currentBrand.keywordGroups.length" class="timeline-empty">
              暂无关键词组。可先维护品牌资料，或用“生成关键词”从品牌定位中提取。
            </div>
            <div v-for="group in currentBrand.keywordGroups" :key="group.id" class="keyword-group-card">
              <div class="keyword-group-head">
                <h4>{{ group.name }}</h4>
                <span class="muted">{{ group.keywords.length }} 个关键词</span>
              </div>
              <div class="keyword-cloud">
                <span v-for="kw in group.keywords" :key="kw" class="keyword">{{ kw }}</span>
              </div>
            </div>
          </div>
          <div v-else-if="dataTab === '素材'" class="material-list">
            <div v-if="!currentBrand.materials.length" class="timeline-empty">
              暂无素材。后续可上传品牌图、商品主图、详情页、买家评价等内容资产。
            </div>
            <div v-for="m in currentBrand.materials" :key="m" class="material-card">{{ m }}</div>
          </div>
          <div v-else class="hotspot-materials">
            <div class="section-toolbar">
              <div>
                <h4>热点文章</h4>
                <p>用于工作台选题借势，可手动录入平台热文、搜索问题或趋势标题。</p>
              </div>
              <button v-if="canImportData" class="btn small primary" @click="openHotspotModal">新建热点文章</button>
            </div>
            <div v-if="!hotspots.length" class="timeline-empty">暂无热点文章。新建后可以在工作台引用。</div>
            <div v-for="hot in hotspots" :key="hot.id" class="hotspot-row-card">
              <div>
                <strong>{{ hot.title }}</strong>
                <p>{{ hot.platform }} · 热度 {{ hot.heat || 0 }}<span v-if="hot.summary"> · {{ hot.summary }}</span></p>
              </div>
              <div class="row-actions">
                <button class="btn small" @click="useHotspot(hot, 'dialog')">引用</button>
                <button v-if="canImportData" class="btn small ghost" @click="openHotspotModal(hot)">编辑</button>
              </div>
            </div>
          </div>
        </div>
      </section>

      <section v-if="activeMenu === 'channelGeneration'" class="page channel-generation-page">
        <div class="channel-generation-console">
          <section class="channel-console-col master-readonly">
            <div class="console-panel-head">
              <div>
                <h4>母稿输入</h4>
                <p>已保存，作为渠道内容生成的唯一主输入</p>
              </div>
              <span class="badge success">{{ selectedDraft?.status || '已通过' }}</span>
            </div>
            <div class="readonly-draft">
              <span>母稿标题</span>
              <h2>{{ selectedDraft?.title }}</h2>
              <span>摘要</span>
              <p>{{ selectedDraft?.summary || '未填写摘要' }}</p>
              <span>正文</span>
              <article class="readonly-article" v-html="formatArticlePreview(selectedDraft?.body)"></article>
            </div>
            <div class="source-record-grid compact">
              <div class="source-record-item">
                <span>关联品牌</span>
                <strong>{{ channelGenerationMaster.brandName }}</strong>
              </div>
              <div class="source-record-item">
                <span>关联商品</span>
                <strong>{{ channelGenerationMaster.productName }}</strong>
              </div>
              <div class="source-record-item">
                <span>关键词</span>
                <strong>{{ channelGenerationMaster.keywords.join(' / ') || '未记录' }}</strong>
              </div>
              <div class="source-record-item">
                <span>生成来源</span>
                <strong>{{ selectedDraft?.source || '母稿' }}</strong>
              </div>
            </div>
          </section>

          <section class="channel-console-col channel-chain">
            <div class="console-panel-head">
              <div>
                <h4>渠道内容链式生成</h4>
                <p>点击左侧节点选择需要生成的平台，已选择的平台会打勾并高亮</p>
              </div>
              <span :class="['badge', channelGenerationStatus === 'completed' ? 'success' : 'warning']">{{ channelGenerationStatusLabel }}</span>
            </div>
            <div class="channel-chain-layout">
              <ol class="chain-node-list">
                <li
                  v-for="(node, index) in channelGeneration.nodes"
                  :key="node.channelName"
                  :class="['chain-node', { active: node.channelName === channelGeneration.activeChannel, selected: isGenerationChannelSelected(node.channelName), muted: !isGenerationChannelSelected(node.channelName), locked: isGenerationChannelLocked(node.channelName) }]"
                  @click="selectChannelNode(node.channelName)"
                >
                  <span class="chain-dot"></span>
                  <button
                    type="button"
                    class="chain-check"
                    :disabled="isGenerationChannelLocked(node.channelName) || channelGeneration.status === 'generating'"
                    :title="isGenerationChannelLocked(node.channelName) ? '已生成，不能取消选择' : '选择/取消生成'"
                    @click.stop="toggleChainNodeSelection(node.channelName)"
                  >{{ isGenerationChannelSelected(node.channelName) ? '✓' : '' }}</button>
                  <ChannelLogo :name="node.channelName" :code="channelCodeByName(node.channelName)" />
                  <strong>{{ index + 1 }}. {{ node.channelName }}</strong>
                  <em>{{ isGenerationChannelLocked(node.channelName) ? '已生成锁定' : (isGenerationChannelSelected(node.channelName) ? generationNodeStatusLabel(node.status) : '未选择') }}</em>
                </li>
              </ol>
              <div class="current-platform-pane">
                <div class="platform-pane-head">
                  <div class="channel-pane-toolbar">
                    <div class="channel-content-tabs">
                      <button :disabled="channelGenerationButtonsLocked" :class="{ active: channelGeneration.detailTab === 'body' }" @click="channelGeneration.detailTab = 'body'">正文</button>
                      <button :disabled="channelGenerationButtonsLocked" :class="{ active: channelGeneration.detailTab === 'basic' }" @click="channelGeneration.detailTab = 'basic'">基本信息</button>
                    </div>
                  </div>
                  <div v-if="channelGeneration.detailTab === 'body'" class="segmented">
                    <button :disabled="channelGenerationButtonsLocked" :class="{ active: channelGeneration.previewMode === 'mobile' }" @click="channelGeneration.previewMode = 'mobile'">手机预览</button>
                    <button :disabled="channelGenerationButtonsLocked" :class="{ active: channelGeneration.previewMode === 'pc' }" @click="channelGeneration.previewMode = 'pc'">PC预览</button>
                    <button :disabled="channelGenerationButtonsLocked" :class="{ active: channelGeneration.previewMode === 'source' }" @click="channelGeneration.previewMode = 'source'">编辑</button>
                  </div>
                </div>
                <div v-if="activeChannelNode?.status === 'generating'" class="channel-generating-state">
                  <ChannelLogo :name="activeChannelNode.channelName" :code="channelCodeByName(activeChannelNode.channelName)" />
                  <strong>{{ activeChannelNode.channelName }}渠道内容生成中</strong>
                  <p>正在读取母稿、渠道标准和素材要求，生成完成后会自动展示当前平台内容。</p>
                  <div class="channel-generation-loader"><span></span></div>
                </div>
                <div v-else-if="!activeChannelNode?.contentPayload?.body" class="channel-empty-state">当前平台还没有生成渠道内容。</div>
                <div v-else-if="channelGeneration.detailTab === 'basic'" class="channel-basic-panel">
                  <label class="channel-basic-title">标题<input v-model="activeChannelNode.contentPayload.title" /></label>
                  <div class="channel-basic-grid">
                    <div
                      v-for="field in activeChannelBasicFields"
                      :key="field.label"
                      :class="['channel-basic-card', { 'channel-basic-card--wide': field.wide }]"
                    >
                      <span>{{ field.label }}</span>
                      <p>{{ field.value || '待维护' }}</p>
                    </div>
                  </div>
                </div>
                <template v-else>
                  <div v-if="channelGeneration.previewMode === 'source'" class="channel-source-editor">
                    <label>正文<textarea v-model="activeChannelNode.contentPayload.body"></textarea></label>
                  </div>
                  <ChannelPreview v-else :mode="channelGeneration.previewMode" :channel="activeChannelPreview" />
                </template>
                <div class="platform-actions">
                  <div class="platform-actions-left">
                    <button class="btn ghost" :disabled="channelGenerationButtonsLocked" @click="showToast('选择资料库素材功能稍后接入资料中心素材弹窗')">选择资料库素材</button>
                  </div>
                  <div class="platform-actions-right">
                    <button class="btn ghost" :disabled="channelGenerationButtonsLocked || !channelGeneration.selectedChannels.length" @click="startChannelGeneration">{{ isChannelGenerationRunning ? '生成中' : '开始生成' }}</button>
                    <button class="btn primary save-plan-btn" :disabled="channelGenerationButtonsLocked || !generatedChannelPlanOptions.length" @click="openActiveChannelPlanModal">保存并添加发布计划</button>
                  </div>
                </div>
              </div>
            </div>
          </section>
        </div>
      </section>
    </main>

    <NeuroAgentDialog
      v-model="modalVisible"
      :title="modal.title"
      :width="modalDialogWidth"
      :height="modalDialogHeight"
      :show-footer="false"
      :close-on-overlay-click="false"
    >
      <div class="geo-growth-page ai-geo-dialog-scope">

        <div v-if="modal.type === 'mode'" class="modal-body">
          <p class="mode-intro">当前流程为母稿保存后生成渠道内容，渠道内容可直接加入发布计划。</p>
          <label>工作模式<select v-model="modeConfig.workMode"><option>人工创作</option><option>智能生成</option></select></label>
          <div class="mini-table">
            <div v-for="channel in channelProfiles" :key="channel.name" class="mini-row">
              <strong>{{ channel.name }}</strong>
              <select v-model="channel.method"><option>渠道 API</option><option>Agent 执行</option><option>人工执行</option></select>
              <select v-model="channel.level"><option>全自动</option><option>半自动</option><option>人工</option></select>
            </div>
          </div>
          <button class="btn primary full" @click="saveModeConfig">保存配置</button>
        </div>

        <div v-if="modal.type === 'brand'" class="modal-body brand-form">
          <div class="grid two">
            <label>品牌名称<input v-model="brandForm.brand_name" placeholder="例如 Mardi Ladin" /></label>
            <label>品牌编码<input v-model="brandForm.brand_code" :disabled="Boolean(brandForm.id)" placeholder="例如 mardi-ladin" /></label>
          </div>
          <label>品牌定位<textarea v-model="brandForm.positioning" rows="2" placeholder="品牌风格、核心品类、差异化定位"></textarea></label>
          <div class="grid two">
            <label>目标人群<input v-model="brandForm.target_audience" placeholder="例如 25-35 岁都市女性" /></label>
            <label>价格带<input v-model="brandForm.price_band" placeholder="例如 299-899 元" /></label>
          </div>
          <label>品牌关键词<textarea v-model="brandForm.keywordsText" rows="2" placeholder="多个关键词用逗号、顿号或换行分隔"></textarea></label>
          <div class="drawer-actions">
            <button class="btn ghost" @click="closeModal">取消</button>
            <button class="btn primary" :disabled="loading.action || !brandForm.brand_name.trim()" @click="saveBrand">
              {{ loading.action ? '保存中' : '保存品牌' }}
            </button>
          </div>
        </div>

        <div v-if="modal.type === 'channelProfile'" class="modal-body channel-profile-form">
          <div class="grid two">
            <label>渠道名称<input v-model="channelForm.channel_name" placeholder="例如 小红书" /></label>
            <label>渠道编码<input v-model="channelForm.channel_code" :disabled="Boolean(channelForm.id)" placeholder="例如 xiaohongshu" /></label>
          </div>
          <div class="grid two">
            <label>渠道类型<input v-model="channelForm.channel_type" placeholder="例如 social_seed" /></label>
            <label>接入状态
              <select v-model="channelForm.status">
                <option value="active">已启用</option>
                <option value="pending">待授权</option>
                <option value="disabled">停用</option>
              </select>
            </label>
          </div>
          <label>URL / 发布入口<input v-model="channelForm.entry_url" placeholder="https://www.xiaohongshu.com" /></label>
          <label>内容形态<textarea v-model="channelForm.content_forms" rows="2" placeholder="多个内容形态用逗号、顿号或换行分隔"></textarea></label>
          <label>支持方式<textarea v-model="channelForm.support_modes" rows="2" placeholder="例如 渠道 API、Agent 执行、人工发布"></textarea></label>
          <label>默认发布方式
            <select v-model="channelForm.default_publish_mode">
              <option value="api">渠道 API</option>
              <option value="agent">Agent 执行</option>
              <option value="manual">人工发布</option>
              <option value="api_draft_manual_confirm">API 草稿 + 人工确认</option>
              <option value="agent_manual_confirm">Agent 执行 + 人工确认</option>
            </select>
          </label>
          <div class="drawer-actions">
            <button class="btn ghost" @click="closeModal">取消</button>
            <button class="btn primary" :disabled="loading.action || !channelForm.channel_name.trim() || !channelForm.channel_code.trim()" @click="saveChannelProfile">
              {{ loading.action ? '保存中' : '保存渠道配置' }}
            </button>
          </div>
        </div>

        <div v-if="modal.type === 'import'" class="modal-body import-flow import-simple">
          <div class="import-type-grid">
            <button v-for="option in importTypeOptions" :key="option.value" :class="['import-type-card', { active: importForm.type === option.value }]" @click="selectImportType(option.value)">
              <strong>{{ option.label }}</strong>
              <span>{{ option.desc }}</span>
            </button>
          </div>
          <div class="import-help">
            <div>
              <h4>{{ currentImportOption.label }}字段</h4>
              <p>{{ currentImportOption.help }}</p>
            </div>
            <div class="import-help-actions">
              <button class="btn small ghost" @click="downloadImportTemplate">下载 Excel 模板</button>
              <button class="btn small ghost" @click="fillImportExample">填入示例</button>
            </div>
          </div>
          <div class="template-table-wrap">
            <table class="table template-table">
              <thead><tr><th>中文列名</th><th>系统字段</th><th>是否必填</th><th>示例</th></tr></thead>
              <tbody>
                <tr v-for="field in currentImportFields" :key="field.key">
                  <td>{{ field.label }}</td>
                  <td>
                    <div class="import-field-name">
                      <code>{{ field.key }}</code>
                      <span>{{ field.label }}</span>
                    </div>
                  </td>
                  <td><span :class="['badge', field.required ? 'danger' : 'info']">{{ field.required ? '必填' : '可选' }}</span></td>
                  <td>{{ field.sample }}</td>
                </tr>
              </tbody>
            </table>
          </div>
          <div class="excel-upload-box">
            <div class="import-file-picker">
              <input ref="importFileInput" type="file" accept=".xlsx,.xls,.csv,.tsv" @change="handleImportFileChange" />
              <span :class="{ active: selectedImportFileName }">{{ selectedImportFileName || '未选择文件' }}</span>
            </div>
            <div>
              <strong>上传 Excel / CSV 文件</strong>
              <p>建议先下载模板填写，再选择文件导入。系统会读取第一个工作表。</p>
              <p class="import-field-hint">当前字段名：{{ importFieldHeaderText }}</p>
            </div>
          </div>
          <label>粘贴 CSV / 表格数据
            <textarea v-model="importForm.rawText" rows="8" :placeholder="currentImportOption.placeholder"></textarea>
          </label>
          <div v-if="importPreviewRows.length" class="import-preview">
            <strong>将导入 {{ importPreviewRows.length }} 行</strong>
            <span>{{ importPreviewRows.slice(0, 2).map(row => Object.values(row).slice(0, 3).join(' / ')).join('；') }}</span>
          </div>
          <div v-if="importResult" class="import-result">
            导入结果：成功 {{ importSuccessCount }} 条，失败 {{ importFailedCount }} 条
            <button v-if="importFailedCount > 0" class="btn small ghost" @click="loadImportErrors(importBatchID)">查看失败原因</button>
          </div>
          <div v-if="importErrors.length" class="import-errors">
            <div v-for="err in importErrors" :key="err.id || err.ID" class="import-error-row">
              第 {{ err.row_number || err.RowNumber }} 行：{{ err.error_message || err.ErrorMessage }}
            </div>
          </div>
          <div class="drawer-actions">
            <button class="btn ghost" @click="closeModal">取消</button>
            <button class="btn primary" :disabled="loading.action || !importPreviewRows.length" @click="submitMaterialImport">
              {{ loading.action ? '导入中' : '开始导入' }}
            </button>
          </div>
        </div>

        <div v-if="modal.type === 'hotspotForm'" class="modal-body hotspot-form">
          <div class="grid two">
            <label>文章日期 *<input v-model="hotspotForm.article_date" type="date" /></label>
            <label>热点主题 *<input v-model="hotspotForm.hotspot_topic" placeholder="如：法式通勤穿搭热度" /></label>
          </div>
          <label>文章标题 *<input v-model="hotspotForm.title" placeholder="输入文章标题" /></label>
          <div class="grid two">
            <label>文章来源 / 平台 *<input v-model="hotspotForm.platform" placeholder="小红书 / 行业媒体 / 平台公告" /></label>
            <label>文章类型
              <select v-model="hotspotForm.article_type">
                <option v-for="type in hotspotArticleTypes" :key="type" :value="type">{{ type }}</option>
              </select>
            </label>
          </div>
          <div class="grid two">
            <label>原文链接<input v-model="hotspotForm.source_url" placeholder="https://" /></label>
            <label>作者 / 账号<input v-model="hotspotForm.author" placeholder="来源账号或维护组" /></label>
          </div>
          <label>文章摘要 *<textarea v-model="hotspotForm.summary" rows="3" placeholder="AI 可读摘要，建议 100-300 字"></textarea></label>
          <label>核心观点 / 运营启发 *<textarea v-model="hotspotForm.core_viewpoint" rows="3" placeholder="这篇热点可以如何转化为标题、卖点、详情页、脚本或投放角度"></textarea></label>
          <label>场景标签 *</label>
          <div class="scene-tag-grid">
            <label v-for="tag in hotspotSceneTags" :key="tag" class="check-card">
              <input type="checkbox" :value="tag" v-model="hotspotForm.scene_tags" />
              <span>{{ tag }}</span>
            </label>
          </div>
          <div class="grid two">
            <label>关联品牌
              <select v-model.number="hotspotForm.brand_id">
                <option :value="0">不关联品牌</option>
                <option v-for="brand in brands" :key="brand.id" :value="brand.id">{{ brand.name }}</option>
              </select>
            </label>
            <label>关联商品 / SPU
              <select v-model.number="hotspotForm.product_id">
                <option :value="0">不关联商品</option>
                <option v-for="product in hotspotProductOptions" :key="product.id" :value="product.id">{{ product.name }} · {{ product.code }}</option>
              </select>
            </label>
          </div>
          <div class="grid two">
            <label>关联 SKU
              <select v-model="hotspotForm.sku_text">
                <option value="">不关联 SKU</option>
                <option value="全 SKU">全 SKU</option>
                <option v-for="sku in hotspotSkuOptions" :key="sku.code" :value="skuLabel(sku)">{{ skuLabel(sku) }}</option>
              </select>
            </label>
            <label>关联竞品<input v-model="hotspotForm.competitor_text" placeholder="竞品品牌或竞品商品" /></label>
          </div>
          <div class="grid two">
            <label>相关关键词<input v-model="hotspotForm.keywords_text" placeholder="多个关键词用逗号分隔" /></label>
            <label>维护人<input v-model="hotspotForm.owner" placeholder="运营同学 / 团队" /></label>
          </div>
          <div class="grid three">
            <label>热度评分<input v-model.number="hotspotForm.heat_score" type="number" min="0" max="100" /></label>
            <label>优先级
              <select v-model="hotspotForm.priority">
                <option v-for="priority in hotspotPriorities" :key="priority" :value="priority">{{ priority }}</option>
              </select>
            </label>
            <label>内容状态
              <select v-model="hotspotForm.content_status">
                <option v-for="status in hotspotContentStatuses" :key="status" :value="status">{{ status }}</option>
              </select>
            </label>
          </div>
          <div class="drawer-actions">
            <button class="btn ghost" @click="closeModal">取消</button>
            <button class="btn primary" :disabled="loading.action || !canSaveHotspotForm" @click="saveHotspot">
              {{ loading.action ? '保存中' : '保存热点文章' }}
            </button>
          </div>
        </div>

        <div v-if="modal.type === 'newPlan'" class="modal-body">
          <label>选择母稿<select v-model="newPlanDraftId"><option v-for="draft in drafts" :value="draft.id" :key="draft.id">{{ draft.title }}</option></select></label>
          <label>渠道<select v-model="newPlanChannel"><option v-for="c in channelProfiles" :key="c.name" :value="c.name">{{ c.name }}</option></select></label>
          <label>执行方式<select v-model="newPlanMethod"><option>渠道 API</option><option>Agent 执行</option><option>人工执行</option></select></label>
          <label>自动化级别<select v-model="newPlanLevel"><option>全自动</option><option>半自动</option><option>人工</option></select></label>
          <label>计划时间<input type="datetime-local" v-model="newPlanTime" /></label>
          <button v-if="canManagePublishPlan" class="btn primary full" @click="createPlan">生成发布任务</button>
        </div>

        <div v-if="modal.type === 'activeChannelPlan'" class="modal-body channel-plan-modal">
          <p class="mode-intro">确认后会先保存所选渠道内容，再把它添加到发布计划。</p>
          <label class="channel-plan-time-field">
            发布时间
            <input type="datetime-local" v-model="channelPlanForm.scheduledAt" />
          </label>
          <div class="channel-plan-card-grid">
            <button
              v-for="item in channelPlanCards"
              :key="item.channelName"
              type="button"
              :class="['channel-plan-card', { selected: isChannelPlanSelected(item.channelName) }]"
              :disabled="loading.action"
              @click="toggleChannelPlanSelection(item.channelName)"
            >
              <span class="channel-plan-card-check">{{ isChannelPlanSelected(item.channelName) ? '✓' : '' }}</span>
              <span class="channel-plan-card-head">
                <ChannelLogo :name="item.channelName" :code="item.channelCode" />
                <strong>{{ item.channelName }}</strong>
              </span>
              <span>执行方式：{{ item.methodLabel }}</span>
              <span>自动化级别：{{ item.level }}</span>
              <span>发布时间：{{ item.scheduledLabel }}</span>
            </button>
          </div>
          <div class="drawer-actions">
            <button class="btn ghost" :disabled="loading.action" @click="closeModal">取消</button>
            <button class="btn primary" :disabled="loading.action || !channelPlanForm.selectedChannels.length" @click="saveChannelAndCreatePlan">
              {{ loading.action ? '保存中' : '确认保存并添加' }}
            </button>
          </div>
        </div>

        <div v-if="modal.type === 'draftSubmitSuccess'" class="modal-body submit-success-modal">
          <div class="submit-success-icon">✓</div>
          <h4>母稿已保存成功</h4>
          <p>母稿已保存，可以直接进入渠道内容生成页面。</p>
          <div class="drawer-actions">
            <button class="btn ghost" @click="closeModal">稍后处理</button>
            <button class="btn primary" @click="goGenerateChannelsAfterSubmit">去生成渠道内容</button>
          </div>
        </div>

        <div v-if="modal.type === 'draftView'" class="modal-body draft-view-modal">
          <div class="draft-view-meta">
            <span>生成时间：{{ selectedDraft?.generatedAtLabel || '-' }}</span>
            <span>状态：{{ selectedDraft?.status || '-' }}</span>
            <span>来源：{{ selectedDraft?.source || '-' }}</span>
          </div>
          <h2>{{ selectedDraft?.title }}</h2>
          <p class="draft-view-summary">{{ selectedDraft?.summary }}</p>
          <article class="draft-view-body">
            <p v-for="(paragraph, paragraphIndex) in draftViewParagraphs" :key="paragraphIndex">{{ paragraph }}</p>
          </article>
          <div v-if="selectedDraft?.channels?.length" class="draft-view-channels">
            <h4>已生成渠道内容</h4>
            <span v-for="channel in selectedDraft.channels" :key="channel.id" class="tag">{{ channel.channel }} · {{ channel.status }}</span>
          </div>
        </div>
      </div>
    </NeuroAgentDialog>

    <div v-if="drawer.type" class="drawer-mask" @click="closeDrawer"></div>
    <aside v-if="drawer.type" class="drawer" role="dialog" aria-modal="true">
      <div class="drawer-header">
        <h3>{{ drawer.title }}</h3>
        <button type="button" aria-label="关闭抽屉" @click="closeDrawer">×</button>
      </div>

      <div v-if="drawer.type === 'hotspot'" class="drawer-body">
        <div class="hotspot-tools">
          <input v-model="hotspotSearch" placeholder="搜索热点关键词" />
          <button class="btn small primary" @click="openHotspotExtractDrawer({ returnTo: 'manage' })">输入 URL 提炼热点</button>
          <button class="btn small" @click="openHotspotModal">新建热点文章</button>
        </div>
        <div v-for="hot in filteredHotspots" :key="hot.id" class="hotspot-card">
          <div><h4>{{ hot.title }}</h4><p>{{ hot.summary }}</p><span>{{ hot.platform }} · 热度 {{ hot.heat }} · {{ hot.risk }}</span></div>
          <div class="row-actions"><button class="btn small" @click="useHotspot(hot, 'dialog')">引用到对话</button><button class="btn small ghost" @click="useHotspot(hot, 'draft')">生成借势角度</button><button v-if="canImportData" class="btn small ghost" @click="openHotspotModal(hot)">编辑</button></div>
        </div>
      </div>

      <div v-if="drawer.type === 'hotspotExtract'" class="drawer-body extract-drawer">
        <div class="extract-form">
          <label>热点文章 URL<input v-model="hotspotExtractForm.url" placeholder="https://" /></label>
          <div class="extract-standard-head">
            <div>
              <strong>标准化提炼</strong>
              <span>默认提炼主题、搜索意图、关键词、借势角度、热度建议和风险提示。</span>
            </div>
          </div>
          <div class="grid two">
            <label class="extract-meta-field">来源平台（可选）<input v-model="hotspotExtractForm.platform" placeholder="自动识别或手动标记" /></label>
            <label class="extract-meta-field">内容类型（可选）
              <select v-model="hotspotExtractForm.contentType">
                <option value="">跟随当前工作台</option>
                <option v-for="type in contentTypeOptions" :key="type.value" :value="type.label">{{ type.label }}</option>
              </select>
            </label>
          </div>
          <div class="drawer-actions">
            <button class="btn ghost" @click="hotspotExtractReturnTo === 'manage' ? openHotspotDrawer() : closeDrawer()">关闭</button>
            <button class="btn primary" :disabled="loading.action || !hotspotExtractForm.url" @click="extractHotspotFromURL">{{ loading.action ? '提炼中' : '提炼热点' }}</button>
          </div>
        </div>
        <div v-if="hotspotExtractResult" class="extract-result">
          <section class="extract-section">
            <div class="section-toolbar"><h4>热点结构</h4><button class="btn small primary" @click="saveExtractedHotspot">{{ hotspotExtractReturnTo === 'manage' ? '保存并返回' : '保存并选中' }}</button></div>
            <div class="extract-card-grid">
              <InfoBlock title="主题" :value="hotspotExtractPreview.title" />
              <InfoBlock title="搜索意图" :value="hotspotExtractPreview.searchIntent" />
              <InfoBlock title="关键词" :value="hotspotExtractPreview.keywordsText" />
              <InfoBlock title="热度与风险" :value="hotspotExtractPreview.riskText" />
            </div>
            <article class="prompt-fragment">{{ hotspotExtractPreview.anglesText }}</article>
          </section>
          <section class="extract-section raw-source-section">
            <h4>原文</h4>
            <p class="muted">{{ hotspotExtractPreview.sourceTitle }} · {{ hotspotExtractPreview.sourceUrl }}</p>
            <article>{{ hotspotExtractPreview.cleanText }}</article>
          </section>
        </div>
      </div>

      <div v-if="drawer.type === 'styleManage'" class="drawer-body style-manage-drawer">
        <div class="style-manage-toolbar">
          <div>
            <h4>已保存风格</h4>
            <p>管理可复用的参考写法；从 URL 提炼的新风格保存后会回到这里。</p>
          </div>
          <button class="btn primary" @click="openStyleExtractDrawer({ returnTo: 'manage' })">输入 URL 提炼风格</button>
        </div>
        <div v-if="styleTemplates.length" class="style-template-list">
          <article v-for="style in styleTemplates" :key="style.id" class="style-template-card">
            <div>
              <h4>{{ style.templateName }}</h4>
              <p>{{ style.promptFragment || style.description || '已保存的参考写作风格' }}</p>
              <span>{{ [style.platform || '通用平台', style.contentType || '通用内容'].join(' · ') }}</span>
            </div>
            <div class="row-actions">
              <button class="btn small ghost" @click="openStyleDetailDrawer(style)">详情</button>
              <button class="btn small" @click="useStyleTemplate(style)">选中</button>
            </div>
          </article>
        </div>
        <div v-else class="empty-state">
          <strong>暂无风格模板</strong>
          <p>输入参考文章 URL 后，可提炼并保存为可复用风格。</p>
        </div>
      </div>

      <div v-if="drawer.type === 'styleDetail'" class="drawer-body extract-drawer">
        <section class="extract-section">
          <div class="section-toolbar">
            <h4>{{ styleDetailPreview.templateName }}</h4>
            <div class="row-actions">
              <button class="btn small ghost" @click="openStyleManageDrawer">返回管理</button>
              <button class="btn small primary" @click="useStyleTemplate(selectedStyleDetail)">选中</button>
            </div>
          </div>
          <div class="extract-card-grid">
            <InfoBlock title="语气" :value="styleDetailPreview.tone" />
            <InfoBlock title="结构" :value="styleDetailPreview.structure" />
            <InfoBlock title="手法" :value="styleDetailPreview.techniques" />
            <InfoBlock title="标签" :value="styleDetailPreview.meta" />
          </div>
          <article class="prompt-fragment">{{ styleDetailPreview.promptFragment }}</article>
        </section>
        <section class="extract-section">
          <h4>禁止复制规则</h4>
          <article class="prompt-fragment">{{ styleDetailPreview.negativeRules }}</article>
        </section>
        <section v-if="styleDetailPreview.sourceExcerpt" class="extract-section raw-source-section">
          <h4>提炼来源摘录</h4>
          <article>{{ styleDetailPreview.sourceExcerpt }}</article>
        </section>
      </div>

      <div v-if="drawer.type === 'styleExtract'" class="drawer-body extract-drawer">
        <div class="extract-form">
          <label>参考文章 URL<input v-model="styleExtractForm.url" placeholder="https://" /></label>
          <div class="extract-standard-head">
            <div>
              <strong>标准化提炼</strong>
              <span>默认提炼可复用的结构、语气和写法边界，平台与内容类型只作为归档标签。</span>
            </div>
          </div>
          <div class="style-focus-options">
            <button
              v-for="option in styleExtractFocusOptions"
              :key="option.value"
              type="button"
              :class="{ active: styleExtractForm.focuses.includes(option.value) }"
              @click="toggleStyleExtractFocus(option.value)"
            >
              <strong>{{ option.label }}</strong>
              <span>{{ option.desc }}</span>
            </button>
          </div>
          <div class="grid two">
            <label class="extract-meta-field">来源平台（可选）<input v-model="styleExtractForm.platform" placeholder="自动识别或手动标记" /></label>
            <label class="extract-meta-field">内容类型（可选）
              <select v-model="styleExtractForm.contentType">
                <option value="">跟随当前工作台</option>
                <option v-for="type in contentTypeOptions" :key="type.value" :value="type.label">{{ type.label }}</option>
              </select>
            </label>
          </div>
          <div class="drawer-actions">
            <button class="btn ghost" @click="closeDrawer">关闭</button>
            <button class="btn primary" :disabled="loading.action || !styleExtractForm.url" @click="extractStyleFromURL">{{ loading.action ? '提炼中' : '提炼写作风格' }}</button>
          </div>
        </div>
        <div v-if="styleExtractResult" class="extract-result">
          <section class="extract-section">
            <div class="section-toolbar"><h4>提炼结构</h4><button class="btn small primary" @click="saveExtractedStyle">{{ styleExtractReturnTo === 'manage' ? '保存并返回' : '保存并选中' }}</button></div>
            <div class="extract-card-grid">
              <InfoBlock title="模板名称" :value="styleExtractPreview.templateName" />
              <InfoBlock title="语气" :value="styleExtractPreview.tone" />
              <InfoBlock title="结构" :value="styleExtractPreview.structure" />
              <InfoBlock title="手法" :value="styleExtractPreview.techniques" />
            </div>
            <article class="prompt-fragment">{{ styleExtractPreview.promptFragment }}</article>
          </section>
          <section class="extract-section raw-source-section">
            <h4>原文</h4>
            <p class="muted">{{ styleExtractPreview.sourceTitle }} · {{ styleExtractPreview.sourceUrl }}</p>
            <article>{{ styleExtractPreview.cleanText }}</article>
          </section>
        </div>
      </div>

      <div v-if="drawer.type === 'channelEditor'" class="drawer-body channel-editor">
        <div class="segmented full-width">
          <button :class="{ active: channelPreviewMode === 'mobile' }" @click="channelPreviewMode = 'mobile'">手机预览</button>
          <button :class="{ active: channelPreviewMode === 'pc' }" @click="channelPreviewMode = 'pc'">PC预览</button>
          <button v-if="!selectedChannel.previewOnly" :class="{ active: channelPreviewMode === 'edit' }" @click="channelPreviewMode = 'edit'">编辑</button>
        </div>
        <div v-if="channelPreviewMode === 'edit'" class="draft-editor">
          <label>渠道标题<input v-model="selectedChannel.title" /></label>
          <label>渠道正文<textarea class="body-editor" v-model="selectedChannel.body"></textarea></label>
          <label v-if="selectedChannel.channel === '小红书'">话题标签<input v-model="selectedChannel.tags" /></label>
          <label v-if="selectedChannel.channel === '独立站'">SEO 标题<input v-model="selectedChannel.seoTitle" /></label>
          <label v-if="selectedChannel.channel === '抖音'">分镜脚本<textarea v-model="selectedChannel.script"></textarea></label>
          <div class="ai-tools">
            <button class="btn small ghost" @click="optimizeChannel('标题更抓人')">优化标题</button>
            <button class="btn small ghost" @click="optimizeChannel('减少营销感')">减少营销感</button>
            <button class="btn small ghost" @click="optimizeChannel('生成话题标签')">生成话题标签</button>
            <button class="btn small ghost" @click="optimizeChannel('规避敏感词')">规避敏感词</button>
          </div>
        </div>
        <ChannelPreview v-else :mode="channelPreviewMode" :channel="selectedChannel" />
        <div v-if="!selectedChannel.previewOnly" class="drawer-actions">
          <button class="btn ghost" @click="closeDrawer">关闭</button>
          <button v-if="canManageChannelContent && !selectedChannel.previewOnly" class="btn ghost" @click="regenerateChannel">重新生成</button>
          <button v-if="canShowChannelConfirm(selectedChannel)" class="btn primary" @click="confirmSelectedChannel">确认</button>
          <button v-if="canShowChannelAddPlan(selectedChannel)" class="btn dark" @click="addSelectedChannelToPlan">加入发布计划</button>
        </div>
      </div>

      <div v-if="drawer.type === 'product'" class="drawer-body product-drawer">
        <div class="product-head">
          <div>
            <h3>{{ selectedProduct.name }}</h3>
            <p>{{ selectedProduct.category }} · {{ selectedProduct.series }} · {{ selectedProduct.priceRange }}</p>
          </div>
          <span class="badge success">{{ selectedProduct.completeness }}% 完整</span>
        </div>
        <div class="tabs small-tabs"><button v-for="tab in productTabs" :key="tab" :class="{ active: productTab === tab }" @click="productTab = tab">{{ tab }}</button></div>
        <div v-if="productTab === '公共资料'" class="info-grid">
          <InfoBlock title="面料" :value="selectedProduct.fabric" />
          <InfoBlock title="版型" :value="selectedProduct.fit" />
          <InfoBlock title="风格标签" :value="selectedProduct.styleTags.join('、')" />
          <InfoBlock title="场景标签" :value="selectedProduct.sceneTags.join('、')" />
          <InfoBlock title="适合人群" :value="selectedProduct.audience" />
          <InfoBlock title="核心卖点" :value="selectedProduct.sellingPoints" />
        </div>
        <div v-if="productTab === 'SKU明细'">
          <table class="table"><thead><tr><th>SKU</th><th>颜色</th><th>尺码</th><th>价格</th><th>状态</th><th>覆盖资料</th></tr></thead><tbody><tr v-for="sku in selectedProduct.skus" :key="sku.code"><td>{{ sku.code }}</td><td>{{ sku.color }}</td><td>{{ sku.size }}</td><td>¥{{ sku.price }}</td><td>{{ sku.status }}</td><td>{{ sku.overrideTitle ? '已配置' : '未配置' }}</td></tr></tbody></table>
        </div>
        <div v-if="productTab === 'SKU覆盖资料'" class="override-list">
          <div v-for="sku in selectedProduct.skus" :key="sku.code" class="override-card"><h4>{{ sku.color }} / {{ sku.size }}</h4><p>标题覆盖：{{ sku.overrideTitle || '沿用商品标题' }}</p><p>卖点覆盖：{{ sku.overridePoint || '沿用商品卖点' }}</p><p>展示规则：SKU覆盖字段 > 商品公共字段</p></div>
        </div>
        <div v-if="productTab === '竞品信息'" class="competitor-list">
          <div v-for="comp in selectedProduct.competitors" :key="comp.name" class="competitor-card">
            <div><h4>{{ comp.brand }} - {{ comp.name }}</h4><p>{{ comp.link }}</p></div>
            <div class="competitor-grid"><span>价格：{{ comp.price }}</span><span>主卖点：{{ comp.point }}</span><span>差异点：{{ comp.diff }}</span><span>可借鉴内容：{{ comp.angle }}</span></div>
          </div>
          <button v-if="canImportData" class="btn ghost" @click="addCompetitor">新增竞品</button>
        </div>
        <div v-if="productTab === '关键词/内容'" class="keyword-cloud"><span v-for="kw in selectedProduct.keywords" :key="kw" class="keyword">{{ kw }}</span></div>
      </div>

      <div v-if="drawer.type === 'publishPlan'" class="drawer-body publish-plan-drawer">
        <div class="plan-detail-head">
          <span class="badge info">{{ selectedPlan.time || '待排期' }}</span>
          <span :class="['badge', statusClass(selectedPlan.publishStatus || selectedPlan.status)]">{{ selectedPlan.publishStatus || selectedPlan.status }}</span>
        </div>
        <h3>{{ selectedPlan.title || selectedPlan.planCode }}</h3>
        <p class="muted">{{ selectedPlan.channel }} · {{ selectedPlan.accountName || '未绑定账号' }}</p>
        <div class="info-grid compact">
          <InfoBlock title="计划编号" :value="selectedPlan.planCode || '-'" />
          <InfoBlock title="发布日期" :value="selectedPlan.date || '-'" />
          <InfoBlock title="发布方式" :value="selectedPlan.method || '-'" />
          <InfoBlock title="账号状态" :value="selectedPlan.accountStatus || '-'" />
          <InfoBlock title="素材状态" :value="selectedPlan.materialStatus || '-'" />
        </div>
        <div class="plan-detail-content">
          <h4>渠道内容</h4>
          <p v-if="selectedPlan.summary" class="muted">{{ selectedPlan.summary }}</p>
          <article v-if="selectedPlan.body">
            <p v-for="(paragraph, index) in publishPlanBodyParagraphs" :key="index">{{ paragraph }}</p>
          </article>
          <p v-else class="muted">当前计划暂未关联可预览的渠道正文。</p>
        </div>
        <div v-if="selectedPlan.tags" class="plan-detail-content">
          <h4>关键词 / 标签</h4>
          <p>{{ selectedPlan.tags }}</p>
        </div>
      </div>
    </aside>

  </div>
</template>

<script setup>
import { computed, defineComponent, h, nextTick, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import * as XLSX from 'xlsx'
import { useRoute, useRouter } from 'vue-router'
import { confirmStandardAction } from '@/composables/useStandardConfirm'
import { fetchDictItemsByCode } from '@/api/dict'
import { usePermissionStore } from '@/stores/permission'
import NeuroAgentDialog from '@/views/components/NeuroAgentDialog.vue'
import ChannelManagementPanel from '../components/ChannelManagementPanel.vue'
import ChannelLogo from '../components/ChannelLogo.vue'
import {
  createAiGeoBrand,
  createAiGeoCompetitor,
  createAiGeoChannel,
  createAiGeoDraft,
  createAiGeoHotspot,
  createAiGeoKeyword,
  createAiGeoPublishPlan,
  extractAiGeoExternalSource,
  fetchAiGeoImportErrors,
  fetchAiGeoBrands,
  fetchAiGeoChannels,
  fetchAiGeoChannelAccounts,
  fetchAiGeoChannelContents,
  fetchAiGeoCompetitors,
  fetchAiGeoDrafts,
  fetchAiGeoHotspots,
  fetchAiGeoKeywords,
  fetchAiGeoMaterialAssets,
  fetchAiGeoOverview,
  fetchAiGeoProducts,
  fetchAiGeoPublishPlanCalendar,
  fetchAiGeoPublishPlans,
  fetchAiGeoSKUs,
  fetchAiGeoStyleTemplates,
  generateAiGeoChannelContent,
  generateAiGeoDraft,
  importAiGeoMaterials,
  streamAiGeoGatewayInvoke,
  testAiGeoChannel,
  updateAiGeoChannel,
  updateAiGeoChannelContent,
  updateAiGeoBrand,
  updateAiGeoDraft,
  updateAiGeoHotspot,
  updateAiGeoPublishPlan,
  updateAiGeoPublishPlanStatus,
} from '../api'
import workbenchSemanticRouterSkill from '../skills/workbenchSemanticRouterSkill.md?raw'

const MetricCard = defineComponent({
  props: ['title', 'value', 'desc'],
  setup(props) {
    return () => h('div', { class: 'metric-card' }, [h('span', props.title), h('strong', props.value), h('p', props.desc)])
  }
})

const InfoBlock = defineComponent({
  props: ['title', 'value'],
  setup(props) {
    return () => h('div', { class: 'info-block' }, [h('span', props.title), h('strong', props.value || '待维护')])
  }
})

function isPreviewSectionTitle(line) {
  const value = String(line || '').trim()
  if (!value || value.length > 34) return false
  if (/^(#{1,4})\s+/.test(value)) return true
  if (/^\d+[.、]\s*\S+/.test(value)) return true
  if (/^(第[一二三四五六七八九十]+部分|[一二三四五六七八九十]+、)/.test(value)) return true
  return /^(职场|约会|周末|通勤|日常|旅行|办公室|场景|选择|搭配|避坑|总结|结论|建议|人群|风格|材质|版型|价格|品牌|商品).{0,14}[：:]\s*\S+/.test(value)
}

function previewSectionTitleText(line) {
  return String(line || '')
    .trim()
    .replace(/^#{1,4}\s+/, '')
    .replace(/^\d+[.、]\s*/, '')
}

function isConclusionHeading(line) {
  const value = previewSectionTitleText(line).replace(/\s+/g, '')
  return /^(?:[一二三四五六七八九十]+、)?(?:结论|总结|结语)$/.test(value)
}

function fallbackConclusionText() {
  return '整体来看，是否适合选择，不能只看单一卖点，而要回到具体人群、使用场景、预算和选择边界。只要把这些判断标准写清楚，内容就更容易被用户理解，也更适合被搜索和 AI 问答引用。'
}

function articlePreviewNodes(body) {
  const source = sanitizeGeneratedDraftBody(body)
  if (!source) return [h('p', { class: 'preview-empty' }, '暂无正文')]
  return source
    .split(/\n{2,}/)
    .map(block => block.trim())
    .filter(Boolean)
    .flatMap((block, blockIndex) => {
      const lines = block.split('\n').map(line => line.trim()).filter(Boolean)
      if (!lines.length) return []
      if (isConclusionHeading(lines[lines.length - 1])) {
        lines.push(fallbackConclusionText())
      }
      return lines.map((line, lineIndex) => {
        const key = `${blockIndex}-${lineIndex}`
        if (isPreviewSectionTitle(line)) {
          return h('h3', { key, class: 'preview-section-title' }, previewSectionTitleText(line))
        }
        return h('p', { key }, line)
      })
    })
}

function renderIphoneShell(children, title = '', options = {}) {
  const cleanTitle = cleanDraftTitle(title)
  const titlePinned = Boolean(options.titlePinned)
  const onScroll = options.onScroll
  return h('div', { class: 'iphone-preview-stage' }, [
    h('div', { class: 'iphone-device iphone-17-pro-max', 'aria-label': 'iPhone 17 Pro Max 样机预览' }, [
      h('div', { class: 'iphone-metal-edge' }),
      h('div', { class: 'iphone-side-button iphone-side-button-left' }),
      h('div', { class: 'iphone-side-button iphone-side-button-right' }),
      h('div', { class: 'iphone-screen' }, [
        h('div', { class: 'iphone-statusbar' }, [
          h('span', { class: 'iphone-time' }, '9:41'),
          h('span', { class: 'iphone-status-icons' }, [
            h('span', { class: 'iphone-signal' }),
            h('span', { class: 'iphone-wifi' }),
            h('span', { class: 'iphone-battery' }),
          ]),
        ]),
        h('div', { class: 'iphone-dynamic-island' }),
        h('div', { class: ['iphone-content-scroll', titlePinned ? 'is-title-pinned' : ''], onScroll }, [
          titlePinned && cleanTitle ? h('div', { class: 'iphone-sticky-title', title: cleanTitle }, cleanTitle) : null,
          cleanTitle ? h('h1', { class: 'iphone-article-title' }, cleanTitle) : null,
          ...children,
        ]),
        h('div', { class: 'iphone-home-indicator' }),
      ]),
    ]),
  ])
}

const PreviewPane = defineComponent({
  props: ['mode', 'title', 'summary', 'body', 'loading'],
  setup(props) {
    const mobileTitlePinned = ref(false)
    const displayTitle = () => cleanDraftTitle(props.title)
    const handleMobileScroll = event => {
      mobileTitlePinned.value = Number(event?.currentTarget?.scrollTop || 0) > 32
    }
    watch(() => [props.mode, props.title, props.loading], () => {
      mobileTitlePinned.value = false
    })
    const renderArticleContent = (includeTitle = true) => props.loading
      ? [h('div', { class: 'draft-preview-loading', 'aria-live': 'polite' }, [
          h('span', { class: 'draft-preview-spinner' }),
          h('strong', '正在生成中，请稍后'),
        ])]
      : [
          includeTitle ? h('h1', displayTitle()) : null,
          h('article', { class: 'preview-article' }, articlePreviewNodes(props.body))
        ].filter(Boolean)
    return () => h('div', { class: ['preview-pane', props.mode === 'mobile' ? 'mobile-frame' : 'pc-frame'] },
      props.mode === 'mobile'
        ? renderIphoneShell(renderArticleContent(false), props.loading ? '' : displayTitle(), {
            titlePinned: mobileTitlePinned.value,
            onScroll: handleMobileScroll,
          })
        : renderArticleContent(true)
    )
  }
})

const ChannelPreview = defineComponent({
  props: ['mode', 'channel'],
  setup(props) {
    const renderContent = () => [
      h('div', { class: 'preview-channel-name' }, `${props.channel.channel} 预览`),
      h('h2', props.channel.title),
      h('p', { class: 'summary' }, props.channel.tags || props.channel.seoTitle || ''),
      ...String(props.channel.body || '').split('\n').map(p => h('p', p)),
      props.channel.script ? h('pre', props.channel.script) : null
    ]
    return () => h('div', { class: ['channel-preview', props.mode === 'mobile' ? 'mobile-frame' : 'pc-frame', `channel-${props.channel.channel}`] },
      props.mode === 'mobile' ? renderIphoneShell(renderContent()) : renderContent()
    )
  }
})

const menus = [
  { key: 'overview', label: '总览', icon: '◎' },
  { key: 'workbench', label: '工作台', icon: '✦' },
  { key: 'drafts', label: '母稿', icon: '▦' },
  { key: 'plans', label: '发布计划', icon: '◷' },
  { key: 'channels', label: '渠道管理', icon: '◇' },
  { key: 'data', label: '资料中心', icon: '◫' }
]

const route = useRoute()
const router = useRouter()
const permissionStore = usePermissionStore()

const canGenerateDraft = computed(() => permissionStore.canUseAction('ai_geo:workbench:generate'))
const canManageDraft = computed(() => permissionStore.canUseAction('ai_geo:draft:manage'))
const canManageChannelContent = computed(() => permissionStore.canUseAction('ai_geo:channel_content:manage'))
const canManagePublishPlan = computed(() => permissionStore.canUseAction('ai_geo:publish_plan:manage'))
const canManageChannelAccount = computed(() => permissionStore.canUseAction('ai_geo:channel_account:manage'))
const canImportData = computed(() => permissionStore.canUseAction('ai_geo:data:import'))

const routeMenuMap = {
  overview: 'overview',
  dashboard: 'overview',
  workbench: 'workbench',
  drafts: 'drafts',
  plans: 'plans',
  calendar: 'plans',
  queue: 'plans',
  channels: 'channels',
  accounts: 'channels',
  data: 'data',
  brands: 'data',
  products: 'data',
  'channel-generation': 'channelGeneration',
}

const menuRouteMap = {
  overview: '/ai-geo/dashboard',
  workbench: '/ai-geo/workbench',
  drafts: '/ai-geo/drafts',
  plans: '/ai-geo/plans/queue',
  channels: '/ai-geo/channels',
  data: '/ai-geo/data/brands',
  channelGeneration: '/ai-geo/channel-generation',
}

const activeMenu = computed({
  get() {
    const section = String(route.params.section || 'dashboard')
    return routeMenuMap[section] || 'overview'
  },
  set(value) {
    router.push(menuRouteMap[value] || '/ai-geo/dashboard')
  },
})
const selectedBrandId = ref(1)
const dataTab = ref('品牌信息')
const dataTabs = ['品牌信息', '商品资料', '关键词', '素材', '热点文章']
const productTab = ref('公共资料')
const productTabs = ['公共资料', 'SKU明细', 'SKU覆盖资料', '竞品信息', '关键词/内容']
const previewMode = ref('mobile')
const draftPreviewLoading = ref(false)
const channelPreviewMode = ref('mobile')
const draftSourceExpanded = ref(false)
const draftDate = ref('')
const planDate = ref('2026-05-20')
const draftFilter = ref('全部状态')
const loading = reactive({ page: false, action: false })
const overview = reactive({
  brand_count: 0,
  product_count: 0,
  sku_count: 0,
  channel_count: 0,
  channel_account_count: 0,
  draft_count_today: 0,
  pending_draft_count: 0,
  channel_content_count: 0,
  publish_plan_today: 0,
  average_completeness: 0,
  pending_tasks: [],
  quota_usage: {},
})

const brands = reactive([
  {
    id: 1,
    name: 'Mardi Ladin',
    position: '轻法式通勤女装品牌',
    audience: '18-35 岁都市女性',
    priceBand: '299-899 元',
    tone: '自然、克制、质感，避免夸大和绝对化表达',
    completeness: 82,
    keywordGroups: [
      { id: 1, name: '通勤场景', keywords: ['轻法式通勤', '通勤穿搭', '小个子连衣裙'] },
      { id: 2, name: '单品词', keywords: ['法式衬衫', '梨形身材穿搭'] }
    ],
    materials: ['品牌 Lookbook', '商品主图', '模特图', '历史小红书笔记', '买家评价'],
    products: [
      {
        id: 101,
        code: 'SPU-DRESS-001',
        name: '法式通勤连衣裙',
        category: '女装/连衣裙',
        series: '春夏通勤系列',
        priceRange: '399-459 元',
        fabric: '醋酸混纺',
        fit: '收腰 A 字版型',
        styleTags: ['轻法式', '显瘦', '通勤'],
        sceneTags: ['上班', '约会', '周末出街'],
        audience: '小个子、梨形身材、通勤女性',
        sellingPoints: '收腰显比例，面料垂顺，适合办公室和轻正式场合。',
        completeness: 86,
        missing: ['部分 SKU 模特图'],
        keywords: ['小个子连衣裙', '通勤连衣裙怎么选', '轻法式穿搭'],
        competitors: [
          { brand: 'Ochirly', name: '通勤收腰连衣裙', price: '¥599', point: '职场通勤、面料挺括', diff: '我们的价格更友好，风格更轻法式', angle: '小个子通勤连衣裙对比', link: 'https://example.com/ochirly-dress' },
          { brand: 'Lily', name: '法式 A 字裙', price: '¥499', point: '年轻通勤、版型修身', diff: '我们的场景覆盖更丰富', angle: '春夏通勤穿搭清单', link: 'https://example.com/lily-dress' }
        ],
        skus: [
          { code: 'SKU-D001-BLK-S', color: '黑色', size: 'S', price: 399, status: '在售', overrideTitle: '黑色显瘦通勤款', overridePoint: '黑色更显瘦，适合正式通勤。' },
          { code: 'SKU-D001-BEI-M', color: '杏色', size: 'M', price: 429, status: '在售', overrideTitle: '杏色温柔约会款', overridePoint: '杏色更柔和，适合约会和春夏出街。' }
        ]
      },
      {
        id: 102,
        code: 'SPU-SHIRT-006',
        name: '飘带法式衬衫',
        category: '女装/衬衫',
        series: '基础通勤系列',
        priceRange: '259-299 元',
        fabric: '雪纺混纺',
        fit: '微宽松',
        styleTags: ['法式', '知性', '百搭'],
        sceneTags: ['通勤', '面试', '会议'],
        audience: '职场新人、通勤女性',
        sellingPoints: '飘带领设计提升精致感，可单穿也可内搭。',
        completeness: 69,
        missing: ['FAQ', '竞品信息', '详情图'],
        keywords: ['法式衬衫', '通勤衬衫', '面试穿搭'],
        competitors: [],
        skus: [
          { code: 'SKU-S006-WHT-M', color: '白色', size: 'M', price: 259, status: '在售', overrideTitle: '', overridePoint: '' }
        ]
      }
    ]
  },
  {
    id: 2,
    name: 'Vela Bag',
    position: '都市通勤包袋品牌',
    audience: '25-40 岁职场女性',
    priceBand: '499-1299 元',
    tone: '专业、实用、质感',
    completeness: 74,
    keywordGroups: [
      { id: 1, name: '通勤场景', keywords: ['通勤包', '大容量托特包', '职场包包'] },
      { id: 2, name: '功能需求', keywords: ['电脑包女'] }
    ],
    materials: ['品牌图', '包袋场景图', '详情图'],
    products: []
  }
])

const emptyBrand = { id: 0, name: '暂无品牌', position: '', audience: '', priceBand: '', tone: '', completeness: 0, keywordGroups: [], materials: [], products: [] }
const currentBrand = computed(() => brands.find(b => b.id === Number(selectedBrandId.value)) || brands[0] || emptyBrand)
const hasBrandData = computed(() => Boolean(currentBrand.value.id))
const brandDetailTitle = computed(() => hasBrandData.value ? `${currentBrand.value.name} 品牌资料卡` : '暂无品牌资料')
const brandMetaText = computed(() => {
  if (!hasBrandData.value) return '新增品牌后，可继续维护品牌信息、商品资料、关键词和素材。'
  const parts = [currentBrand.value.audience, currentBrand.value.priceBand].filter(Boolean)
  if (parts.length) return parts.join(' · ')
  return currentBrand.value.position ? '品牌定位已维护，目标人群和价格带待补充。' : '品牌定位、目标人群、价格带待维护'
})
const productCount = computed(() => overview.product_count || brands.reduce((sum, b) => sum + b.products.length, 0))
const skuCount = computed(() => overview.sku_count || brands.reduce((sum, b) => sum + countSkus(b), 0))
function countSkus(brand) { return brand.products.reduce((sum, p) => sum + p.skus.length, 0) }
function brandCardMetaText(brand) {
  const parts = [brand?.audience, brand?.priceBand].filter(Boolean)
  if (parts.length) return parts.join(' · ')
  return brand?.position ? '品牌定位已维护' : '品牌资料待完善'
}
function clampPercent(value) {
  const number = Number(value || 0)
  if (Number.isNaN(number)) return 0
  return Math.max(0, Math.min(100, Math.round(number)))
}
function averagePercent(values) {
  const valid = values.map(value => Number(value || 0)).filter(value => !Number.isNaN(value))
  if (!valid.length) return 0
  return clampPercent(valid.reduce((sum, value) => sum + value, 0) / valid.length)
}
function skuCompletenessScore(brand) {
  const products = brand.products || []
  if (!products.length) return 0
  const skus = products.flatMap(product => product.skus || [])
  if (!skus.length) return 0
  const productCoverage = products.filter(product => (product.skus || []).length > 0).length / products.length
  const skuDetailScore = averagePercent(skus.map(sku => {
    const fields = [sku.code, sku.name, sku.color || sku.size, sku.price > 0 ? sku.price : '', sku.overridePoint || sku.overrideTitle]
    return fields.filter(value => String(value || '').trim()).length * 100 / fields.length
  }))
  return clampPercent(productCoverage * 60 + skuDetailScore * 0.4)
}
function contentAssetScore(brand) {
  const products = brand.products || []
  const hasKeyword = Boolean((brand.keywordGroups || []).length || products.some(product => (product.keywords || []).length))
  const hasMaterial = Boolean((brand.materials || []).length)
  const hasCompetitor = products.some(product => (product.competitors || []).length)
  return [hasKeyword, hasMaterial, hasCompetitor].filter(Boolean).length * 100 / 3
}
function brandCompletenessParts(brand) {
  const products = brand.products || []
  return {
    brand: clampPercent(brand.completeness),
    product: products.length ? averagePercent(products.map(product => product.completeness)) : 0,
    sku: skuCompletenessScore(brand),
    asset: clampPercent(contentAssetScore(brand)),
  }
}
function brandOverallCompleteness(brand) {
  const parts = brandCompletenessParts(brand || emptyBrand)
  return clampPercent(parts.brand * 0.4 + parts.product * 0.3 + parts.sku * 0.2 + parts.asset * 0.1)
}
function brandCompletenessBreakdownText(brand) {
  const parts = brandCompletenessParts(brand || emptyBrand)
  return `品牌资料 ${parts.brand}% · 商品资料 ${parts.product}% · SKU 覆盖 ${parts.sku}% · 关键词/素材/竞品 ${parts.asset}%`
}

const modeConfig = reactive({
  workMode: '人工创作'
})

const channelProfiles = reactive([])
const channelGenerationOrder = ['小红书', '知乎', '微信公众号', '抖音', '微博', '百家号', '独立站']
const draftQuickEditActions = [
  { label: '降低营销感', prompt: '降低营销感，语气更自然，减少品牌自夸和硬广表达，保留核心 GEO 关键词。' },
  { label: '强化 GEO 表达', prompt: '强化 GEO 表达，补充问题词、人群词和场景词，让内容更适合搜索和 AI 问答引用。' },
  { label: '优化标题', prompt: '优化标题，让标题更像平台用户会点击的问题或入口，降低生硬营销感。' },
  { label: '优化素材说明', prompt: '优化素材说明，明确需要哪些真实商品图、场景图、封面图或视频素材。' },
  { label: '调整内容结构', prompt: '调整内容结构，让层次更清晰，先回答问题，再说明理由，最后给出选择建议。' },
]
const channelGeneration = reactive({
  selectedChannels: ['小红书', '知乎', '微信公众号'],
  activeChannel: '小红书',
  detailTab: 'body',
  previewMode: 'mobile',
  status: 'not_started',
  nodes: [],
})
const channelEditorMessages = reactive([])
const channelEditPrompt = ref('')
const channelGenerationPlatforms = computed(() => channelGenerationOrder.map(name => channelProfiles.find(channel => channel.name === name)).filter(Boolean))
const activeChannelNode = computed(() => channelGeneration.nodes.find(node => node.channelName === channelGeneration.activeChannel) || channelGeneration.nodes[0] || null)
const channelGenerationStatus = computed(() => channelGeneration.status)
const isChannelGenerationRunning = computed(() => channelGeneration.status === 'generating')
const channelGenerationButtonsLocked = computed(() => loading.action || isChannelGenerationRunning.value)
const channelGenerationStatusLabel = computed(() => ({
  not_started: '未开始',
  generating: '生成中',
  partial_completed: '部分完成',
  completed: '全部完成',
  failed: '生成失败',
  cancelled: '已取消',
}[channelGeneration.status] || '未开始'))
const generatedChannelPlanOptions = computed(() => channelGeneration.nodes.filter(node => node.contentId && node.contentPayload?.body))
const channelPlanCards = computed(() => generatedChannelPlanOptions.value.map(node => {
  const profile = channelProfiles.find(item => item.name === node.channelName) || {}
  return {
    channelName: node.channelName,
    channelCode: profile.code || channelCodeByName(node.channelName),
    channelId: Number(profile.id || node.channelId || 0),
    method: profile.method || 'manual',
    methodLabel: publishModeLabel(profile.method || 'manual'),
    level: profile.level || '人工',
    scheduledAt: channelPlanForm.scheduledAt,
    scheduledLabel: channelPlanTimeLabel(channelPlanForm.scheduledAt),
  }
}))
const channelGenerationMaster = computed(() => {
  const snapshot = parseJsonObject(selectedDraft.value?.sourceSnapshot || selectedDraft.value?.raw?.source_snapshot || selectedDraft.value?.raw?.SourceSnapshot)
  const brand = parseJsonObject(snapshot.brand)
  const product = parseJsonObject(snapshot.product)
  return {
    brandName: brand.name || brand.brand_name || brand.BrandName || currentBrand.value.name || '未记录',
    productName: product.name || product.product_name || product.ProductName || '未指定',
    keywords: parseJsonArray(selectedDraft.value?.keywords).concat(Array.isArray(snapshot.keywords) ? snapshot.keywords : []).filter(Boolean).slice(0, 8),
  }
})
const activeChannelExtraFields = computed(() => {
  const payload = activeChannelNode.value?.contentPayload || {}
  const extra = payload.extraFields || {}
  const rows = []
  Object.entries(extra).forEach(([key, value]) => rows.push({ label: channelFieldLabel(key), value: arrayOrObjectText(value) }))
  ;['tags', 'hashtags', 'key_points', 'section_titles', 'faq', 'keywords'].forEach(key => {
    if (payload[key]) rows.push({ label: channelFieldLabel(key), value: arrayOrObjectText(payload[key]) })
  })
  return rows.slice(0, 8)
})
const activeChannelAssetSummary = computed(() => {
  const asset = activeChannelNode.value?.assetPayload || {}
  const required = (asset.requiredAssets || []).map(item => item.slot || item.requirement).filter(Boolean)
  const missing = (asset.missingAssets || []).map(item => item.slot || item.requirement).filter(Boolean)
  if (missing.length) return `缺失：${missing.join('、')}。请从资料中心补充或进入渠道图片生成。`
  if (required.length) return `需要：${required.join('、')}。`
  return '当前平台无强制素材，可按渠道需要补充。'
})
const activeChannelBasicFields = computed(() => {
  const node = activeChannelNode.value || {}
  const payload = node.contentPayload || {}
  const asset = node.assetPayload || {}
  const rows = []
  const add = (label, value, wide = false) => {
    const text = arrayOrObjectText(value)
    if (text) rows.push({ label, value: text, wide })
  }
  add('平台', node.channelName)
  add('内容类型', payload.note_type || payload.content_type || channelContentTypeForName(node.channelName))
  add('封面风格', asset.coverStyle || asset.cover_style || payload.cover_style, true)
  add('图片数量', asset.imageCount || asset.image_count || payload.image_count)
  add('话题标签', payload.tags || payload.hashtags || payload.topic_tags || payload.keywords, true)
  add('图片建议', asset.imageSuggestions || asset.image_suggestions || payload.image_suggestions, true)
  add('素材需求', activeChannelAssetSummary.value, true)
  return rows
})
const activeChannelPreview = computed(() => {
  const node = activeChannelNode.value || {}
  const payload = node.contentPayload || {}
  return {
    channel: node.channelName || '',
    title: payload.title || payload.question_title || payload.answer_title || payload.video_title || payload.seo_title || payload.page_title || '',
    body: payload.body || payload.answer_body || payload.script || payload.post_text || '',
    tags: arrayOrObjectText(payload.tags || payload.hashtags || payload.keywords || node.geoPayload?.keywords || []),
    seoTitle: payload.seo_title || payload.meta_description || '',
    script: payload.storyboard || payload.subtitles ? [arrayOrObjectText(payload.storyboard), arrayOrObjectText(payload.subtitles)].filter(Boolean).join('\n\n') : '',
  }
})

const planTabs = ['发布队列', '发布日历']
const planTab = ref('发布队列')

watch(
  () => [route.params.section, route.params.subsection],
  ([section, subsection]) => {
    if (section === 'plans') {
      planTab.value = subsection === 'calendar' ? '发布日历' : '发布队列'
    }
    if (section === 'data') {
      dataTab.value = subsection === 'products' ? '商品资料' : '品牌信息'
    }
  },
  { immediate: true }
)

const channelAccounts = reactive([])
const channelContentIndex = reactive([])

const planCalendarDays = ref([])

const calendarDays = computed(() => {
  if (planCalendarDays.value.length) {
    return planCalendarDays.value.map(day => ({
      date: day.date,
      label: day.date.slice(5).replace('-', '/'),
      plans: dedupePlans((day.items || []).map(apiPlanToPlan)),
    }))
  }
  const base = new Date(planDate.value || '2026-05-20')
  const days = []
  for (let i = -2; i <= 4; i++) {
    const d = new Date(base)
    d.setDate(base.getDate() + i)
    const date = d.toISOString().slice(0, 10)
    const label = `${d.getMonth() + 1}/${d.getDate()}`
    days.push({ date, label, plans: plans.filter(p => p.date === date) })
  }
  return days
})
const dataIssues = computed(() => [
  { id: 1, level: '高', title: '飘带法式衬衫缺少竞品信息', desc: '会影响选题差异化和竞品对比内容生成。' },
  { id: 2, level: '中', title: '法式通勤连衣裙部分 SKU 缺少模特图', desc: '小红书、独立站商品页预览素材不足。' },
  { id: 3, level: '中', title: '部分商品 FAQ 未补齐', desc: '会影响知乎问答和独立站 FAQ 页面生成。' }
])

const progressStats = computed(() => [
  { label: '今日母稿', value: overview.draft_count_today || drafts.filter(d => d.date === '2026-05-20').length, desc: '含草稿和已生成' },
  { label: '渠道内容', value: overview.channel_content_count || drafts.reduce((s, d) => s + d.channels.length, 0), desc: '已生成版本' },
  { label: '今日发布', value: overview.publish_plan_today || plans.filter(p => p.date === '2026-05-20').length, desc: '发布计划任务' }
])

const pendingTasks = computed(() => {
  if (overview.pending_tasks?.length) {
    return overview.pending_tasks.map((task, index) => ({ id: `api-${index}`, tag: task.tag, title: task.title, menu: index === 0 ? 'drafts' : 'data' }))
  }
  const failedPlanCount = plans.filter(p => p.status === '发布失败').length
  return [
    { id: 'create', tag: '人工 AI 创作', title: '根据运营想法生成母稿', menu: 'workbench' },
    { id: 'plan', tag: '发布计划', title: failedPlanCount ? `${failedPlanCount} 个失败任务需处理` : '暂无失败任务', menu: 'plans' }
  ]
})

const INTERNAL_DRAFT_SKILL = '通用母稿生成 Skill'
const DEFAULT_CONTENT_TYPE = '通用母稿'
const DEFAULT_CONTENT_TYPE_OPTIONS = [
  { value: '通用母稿', label: '通用母稿' },
  { value: '商品种草文', label: '商品种草文' },
  { value: '商品测评文', label: '商品测评文' },
  { value: '商品对比文', label: '商品对比文' },
  { value: '场景解决方案', label: '场景解决方案' },
  { value: '品牌介绍文', label: '品牌介绍文' },
  { value: 'FAQ 问答文', label: 'FAQ 问答文' },
  { value: '榜单推荐文', label: '榜单推荐文' },
  { value: '热点借势文', label: '热点借势文' },
  { value: '活动营销文', label: '活动营销文' },
]
const contentTypeOptions = ref([...DEFAULT_CONTENT_TYPE_OPTIONS])

const workbench = reactive({ brandId: '', productId: '', contentType: DEFAULT_CONTENT_TYPE, skill: INTERNAL_DRAFT_SKILL, hotspot: null, styleTemplate: null, prompt: '' })
const chatMessages = reactive([])
const chatStreaming = ref(false)
const workbenchChatLogRef = ref(null)
const workbenchPromptRef = ref(null)
const shouldFollowWorkbenchChat = ref(true)
const selectedDirectionChoiceByMessage = reactive({})
let workbenchChatScrollFrame = 0
let draftLoadObserver = null
const WORKBENCH_DRAFT_STORAGE_KEY = 'ai_geo_workbench_draft_id'
const WORKBENCH_RESET_STORAGE_KEY = 'ai_geo_workbench_reset'
const WORKBENCH_SESSION_STORAGE_KEY = 'ai_geo_workbench_session'
let restoringWorkbenchContext = false
let workbenchSessionSaveTimer = 0
const ideaSession = reactive({
  stage: 'collecting',
  intent: '等待想法',
  recommendedSkill: '',
  recommendedProductId: '',
  searchProblem: '',
  audience: '',
  scene: '',
  tone: '',
  brief: '',
  directionHistory: [],
})
const aiWorkStage = ref('idle')
const aiIntent = ref('unknown')
const decisionCard = reactive({
  brandName: '',
  productName: '',
  skillName: '',
  contentType: '',
  recommendedDirection: '',
  reason: '',
  stageLabel: '未开始',
  nextStepLabel: '',
})
const stageLabels = {
  idle: '未开始',
  understanding: '理解需求中',
  direction_recommended: '待生成母稿',
  ready_to_generate: '可生成母稿',
  draft_generating: '生成中',
  draft_generated: '已生成草稿',
  draft_editing: '修改中',
  review_ready: '可生成渠道',
  approved: '已通过',
  channel_ready: '可生成渠道内容',
}
const skillProfiles = {
  '通用母稿': { name: '通用母稿', goal: '自由创作', output: '通用 GEO 母稿', desc: '适合未明确内容方向时，生成可继续精修的 GEO 母稿。', tags: ['通用结构', '资料驱动', '人工确认'] },
  '商品种草文': { name: '商品种草文', goal: '推动商品种草', output: '商品推荐文章', desc: '强调商品卖点、适用人群和购买理由。', tags: ['卖点提炼', '适用场景', '购买理由'] },
  '商品测评文': { name: '商品测评文', goal: '帮助用户判断是否值得买', output: '商品测评文章', desc: '强调体验、优缺点、适用边界和真实感。', tags: ['体验维度', '优缺点', '适用边界'] },
  '商品对比文': { name: '商品对比文', goal: '辅助选购决策', output: '对比型文章', desc: '强调同类产品或品牌差异、选择建议。', tags: ['差异对比', '怎么选', '边界说明'] },
  '场景解决方案': { name: '场景解决方案', goal: '占领场景搜索', output: '场景解决方案', desc: '围绕通勤、约会、旅行、办公等具体场景给出解决方案。', tags: ['场景问题', '解决方案', '对比建议'] },
  '品牌介绍文': { name: '品牌介绍文', goal: '建立品牌认知', output: '品牌介绍文章', desc: '解释品牌定位、适合人群和品牌价值。', tags: ['品牌定位', '目标人群', '搜索心智'] },
  'FAQ 问答文': { name: 'FAQ 问答文', goal: '承接长尾问答', output: '问答型文章', desc: '适合搜索问答和 AI 问答引用。', tags: ['FAQ', '长尾词', '可信回答'] },
  '榜单推荐文': { name: '榜单推荐文', goal: '沉淀推荐清单', output: '榜单推荐文章', desc: '适合推荐清单、排行榜、选购合集。', tags: ['榜单主题', '入选标准', '推荐理由'] },
  '热点借势文': { name: '热点借势文', goal: '结合热点做轻引用', output: '热点借势文章', desc: '结合热点做轻引用，不强行蹭热点。', tags: ['热点关系', '观点切入', '品牌承接'] },
  '活动营销文': { name: '活动营销文', goal: '承接活动与节点营销', output: '活动营销文章', desc: '围绕活动、上新、促销、节日节点生成内容。', tags: ['活动信息', '行动建议', '注意事项'] },
}
const defaultSkillProfile = skillProfiles[DEFAULT_CONTENT_TYPE]

const editingDraft = reactive({
  id: 0,
  title: '',
  summary: '',
  body: '',
  keywordsText: '',
  status: '草稿',
  source: '人工创作',
  sourceSnapshot: {}
})
const savedEditingDraftSnapshot = ref('')
function currentEditingDraftSnapshot() {
  return JSON.stringify({
    id: Number(editingDraft.id || 0),
    title: String(editingDraft.title || '').trim(),
    summary: String(editingDraft.summary || '').trim(),
    body: String(editingDraft.body || '').trim(),
    keywordsText: String(editingDraft.keywordsText || '').trim(),
    brandId: String(workbench.brandId || ''),
    productId: String(workbench.productId || ''),
    contentType: String(workbench.contentType || DEFAULT_CONTENT_TYPE),
    hotspotId: String(workbench.hotspot?.id || ''),
  })
}
function markEditingDraftSaved() {
  savedEditingDraftSnapshot.value = currentEditingDraftSnapshot()
}
function clearEditingDraftSavedSnapshot() {
  savedEditingDraftSnapshot.value = ''
}
const hasEditingDraftContent = computed(() => Boolean(String(editingDraft.title || editingDraft.summary || editingDraft.body || '').trim()))
const hasUnsavedEditingDraftChanges = computed(() => hasEditingDraftContent.value && currentEditingDraftSnapshot() !== savedEditingDraftSnapshot.value)
const canEditDraftActions = computed(() => hasEditingDraftContent.value && hasUnsavedEditingDraftChanges.value && !loading.action)
const canShowSaveDraft = computed(() => canManageDraft.value)
const canShowApprovedChannelAction = computed(() => canManageChannelContent.value && hasEditingDraftContent.value)

const selectedWorkbenchBrand = computed(() => workbench.brandId ? (brands.find(b => b.id === Number(workbench.brandId)) || null) : null)
const selectedWorkbenchProduct = computed(() => selectedWorkbenchBrand.value?.products?.find(p => p.id === Number(workbench.productId)) || null)
const selectedSkillProfile = computed(() => skillProfiles[workbench.contentType] || {
  ...defaultSkillProfile,
  name: workbench.contentType || DEFAULT_CONTENT_TYPE,
  goal: '按内容类型生成',
  output: 'GEO 母稿',
  desc: '由数据字典维护的内容类型。',
})
const hasWorkbenchContext = computed(() => Boolean(workbench.brandId || workbench.productId || workbench.hotspot?.id || workbench.contentType !== DEFAULT_CONTENT_TYPE))
function workbenchBrandName(fallback = '未选择品牌') {
  return selectedWorkbenchBrand.value?.name || fallback
}
function workbenchBrandForGeneration() {
  return selectedWorkbenchBrand.value || emptyBrand
}
function setAiStage(stage) {
  aiWorkStage.value = stage
  decisionCard.stageLabel = stageLabels[stage] || stageLabels.idle
}
function inferContentType(typeName = selectedSkillProfile.value.name) {
  return typeName || DEFAULT_CONTENT_TYPE
}
function updateDecisionCard(patch = {}) {
  Object.assign(decisionCard, {
    brandName: selectedWorkbenchBrand.value?.name || '',
    productName: selectedWorkbenchProduct.value?.name || '',
    skillName: INTERNAL_DRAFT_SKILL,
    contentType: inferContentType(selectedSkillProfile.value.name),
    stageLabel: stageLabels[aiWorkStage.value] || '未开始',
    ...patch,
  })
}
const workbenchKeywords = computed(() => {
  const brandKeywords = selectedWorkbenchBrand.value?.keywordGroups?.flatMap(group => group.keywords || []) || []
  const productKeywords = selectedWorkbenchProduct.value?.keywords || []
  return [...new Set([...productKeywords, ...brandKeywords])]
})
const workbenchEvidence = computed(() => [
  { label: '人群', value: selectedWorkbenchProduct.value?.audience || selectedWorkbenchBrand.value?.audience || '未维护' },
  { label: '关键词', value: workbenchKeywords.value.slice(0, 4).join('、') || '未维护' },
])
const draftSourceRecords = computed(() => sourceRecordsFromSnapshot(editingDraft.sourceSnapshot))
const draftSourceSummary = computed(() => draftSourceRecords.value.slice(0, 3).map(record => `${record.label}：${record.title}`).join(' / '))
const generationPlaceholder = computed(() => {
  const product = selectedWorkbenchProduct.value?.name || '当前资料'
  return `例如：围绕「${product}」写一篇回答“小个子通勤怎么穿”的 GEO 文章，强调适合人群、选择理由、场景建议。`
})
const hasWorkbenchInput = computed(() => Boolean(String(workbench.prompt || '').trim()))
const workbenchUserMessages = computed(() => chatMessages.filter(msg => msg.role === 'user'))
const visibleChatMessages = computed(() => chatMessages.filter(msg => !shouldHideChatMessage(msg)))
const hasStartedWorkbenchChat = computed(() => workbenchUserMessages.value.length > 0)
const latestWorkbenchUserPrompt = computed(() => [...chatMessages].reverse().find(msg => msg.role === 'user')?.text || '')
const workbenchReadiness = computed(() => evaluateWorkbenchReadiness(String(workbench.prompt || '').trim()))
const aiStageLabel = computed(() => decisionCard.stageLabel || stageLabels[aiWorkStage.value] || '未开始')
const showDecisionCard = computed(() => hasWorkbenchContext.value || hasStartedWorkbenchChat.value || hasEditingDraftContent.value || aiWorkStage.value !== 'idle')
const showDecisionActions = computed(() => ['direction_recommended', 'ready_to_generate'].includes(aiWorkStage.value) && Boolean(decisionCard.recommendedDirection) && !hasEditingDraftContent.value)
const directionDisplayTitle = computed(() => directionTitleOnly(decisionCard.recommendedDirection))
const latestVisibleAiMessageId = computed(() => {
  const messages = visibleChatMessages.value
  for (let index = messages.length - 1; index >= 0; index -= 1) {
    if (messages[index]?.role === 'ai') return messages[index].id
  }
  return ''
})
const WORKBENCH_REPLY_ACTION_META = {
  generate_draft: { label: '生成母稿', variant: 'primary' },
  apply_to_draft: { label: '应用到母稿', variant: 'primary' },
  apply_title: { label: '应用标题', variant: 'primary' },
  regenerate_draft: { label: '重新生成母稿', variant: 'ghost' },
  recommend_other_directions: { label: '更多 3 个其他方向', variant: 'ghost' },
  copy_reply: { label: '复制回复', variant: 'ghost' },
  retry: { label: '重试', variant: 'ghost' },
}
const WORKBENCH_REPLY_ACTION_TYPES = new Set(Object.keys(WORKBENCH_REPLY_ACTION_META))
const WORKBENCH_REPLY_ACTION_MARKER = /<!--\s*(?:suggested_)?actions\s*:\s*([\s\S]*?)\s*-->/gi

function normalizeReplyActions(actions) {
  const list = Array.isArray(actions) ? actions : (actions ? [actions] : [])
  const normalized = []
  const seen = new Set()
  for (const item of list) {
    const rawType = typeof item === 'string' ? item : (item?.type || item?.action || item?.name)
    const type = String(rawType || '').trim()
    if (!WORKBENCH_REPLY_ACTION_TYPES.has(type) || seen.has(type)) continue
    const meta = WORKBENCH_REPLY_ACTION_META[type]
    normalized.push({
      ...(typeof item === 'object' && item ? item : {}),
      id: `${type}-${normalized.length}`,
      type,
      label: String((typeof item === 'object' && item?.label) || meta.label),
      variant: meta.variant,
    })
    seen.add(type)
  }
  return normalized
}

function parseReplyActionMarker(text) {
  let parsedActions = []
  const visibleText = String(text || '').replace(WORKBENCH_REPLY_ACTION_MARKER, (_match, json) => {
    try {
      const actions = normalizeReplyActions(JSON.parse(json))
      if (actions.length) parsedActions = actions
    } catch (error) {
      console.warn('[ai-geo] invalid suggested action marker:', error)
    }
    return ''
  }).trim()
  return { text: visibleText, actions: parsedActions }
}

function applyAiMessageText(message, text, options = {}) {
  const nextText = options.replace ? String(text || '') : `${message.text || ''}${text || ''}`
  const parsed = parseReplyActionMarker(nextText)
  message.text = parsed.text
  if (parsed.actions.length) message.suggestedActions = parsed.actions
}

function eventReplyActions(event) {
  return normalizeReplyActions(
    event?.data?.suggested_actions ||
    event?.data?.suggestedActions ||
    event?.data?.actions ||
    event?.suggested_actions ||
    event?.suggestedActions,
  )
}

function replyRequiresDirectionChoice(text) {
  const value = String(text || '').replace(/\s+/g, '')
  if (!value) return false
  const asksChoice = /(请|先)?(告诉我|选择|选|确认).{0,10}(哪个|哪一个|方案|方向|倾向)|倾向于哪个方案|选方案[一二三四五六七八九十\d]|如果.*选方案[一二三四五六七八九十\d]/.test(value)
  const hasMultipleOptions = /(方案一|方向一|1[.、]).{0,80}(方案二|方向二|2[.、])/.test(value)
  const notReady = /(一旦确认|确认后|确定后|选定后|再进入|再生成|才可以|才会)/.test(value)
  return (asksChoice && hasMultipleOptions) || (hasMultipleOptions && notReady)
}

function replyActionApplyContent(action) {
  const payload = action?.payload && typeof action.payload === 'object' ? action.payload : {}
  const content = action?.content || payload.content || payload.body || payload.draft_body || payload.draftBody
  if (typeof content === 'string' && content.trim()) return content.trim()
  const title = typeof payload.title === 'string' ? payload.title.trim() : ''
  const summary = typeof payload.summary === 'string' ? payload.summary.trim() : ''
  const body = typeof payload.body === 'string' ? payload.body.trim() : ''
  const parts = [
    title ? `标题：${title}` : '',
    summary ? `摘要：${summary}` : '',
    body ? `正文：${body}` : '',
  ].filter(Boolean)
  return parts.join('\n\n')
}

function impliedReplyActionsFromAiText(message) {
  const text = String(message?.text || '').trim()
  if (!text || message?.streaming) return []
  const hasDirectionOptions = directionChoiceOptions(message).length > 0
  if (replyRequiresDirectionChoice(text)) {
    return normalizeReplyActions([
      { type: 'recommend_other_directions', label: '更多 3 个其他方向' },
      { type: 'copy_reply', label: '复制回复' },
    ])
  }
  const actions = []
  if (extractTitleCandidateFromAiReply(text)) {
    actions.push({ type: 'apply_title', label: '应用标题' })
  }
  if (replyContainsApplicableDraft(text)) {
    actions.push({ type: 'apply_to_draft', label: '应用到母稿', content: text })
  }
  if (/(生成完整母稿|生成母稿|进入生成母稿|进入生成完整母稿|直接进入生成)/.test(text)) {
    actions.push({ type: 'generate_draft', label: '生成母稿' })
  }
  if (
    !actions.some(action => action.type === 'generate_draft') &&
    shouldOfferGenerateDraftAction(text, hasDirectionOptions) &&
    decisionCard.recommendedDirection
  ) {
    actions.push({ type: 'generate_draft', label: hasDirectionOptions ? '按推荐方向生成母稿' : '生成母稿' })
  }
  if (/(其他方向|换方向|换个方向|另一个方向|别的方向|方向参考|标题参考|继续细化|先定一个方向|这个方向可以吗|这个方向你觉得|方向是否符合|哪个方向更好)/.test(text)) {
    actions.push({ type: 'recommend_other_directions', label: '更多 3 个其他方向' })
  }
  if (hasDirectionOptions && !actions.some(action => action.type === 'recommend_other_directions')) {
    actions.push({ type: 'recommend_other_directions', label: '换一组方向' })
  }
  if (hasEditingDraftContent.value && !actions.some(action => action.type === 'regenerate_draft')) {
    actions.push({ type: 'regenerate_draft', label: '按当前方向重生成' })
  }
  if (!actions.some(action => action.type === 'copy_reply')) {
    actions.push({ type: 'copy_reply', label: '复制回复' })
  }
  return normalizeReplyActions(actions)
}

function shouldOfferGenerateDraftAction(text, hasDirectionOptions = false) {
  if (!decisionCard.recommendedDirection) return false
  if (['ready_to_generate', 'direction_recommended'].includes(aiWorkStage.value)) return true
  if (/生成\s*(?:GEO\s*)?母稿/.test(String(decisionCard.nextStepLabel || ''))) return true
  if (hasDirectionOptions) return true
  return /(确认后|倾向哪个|你倾向|哪个方案|哪个方向|完整的标题|正文改写|高质量GEO|软文)/.test(String(text || ''))
}

function replyContainsApplicableDraft(text) {
  const value = String(text || '')
  if (value.length < 180) return false
  const hasDraftFields = /(标题|摘要|正文|核心问题|核心答案|用户痛点|解决方案)[:：]/.test(value)
  const isAnalysisOnly = /(方案[一二三四五六七八九十\d]+|方向[一二三四五六七八九十\d]+).{0,80}(方案[一二三四五六七八九十\d]+|方向[一二三四五六七八九十\d]+)/s.test(value)
  const asksChoice = /(你倾向|请选择|选哪个|哪个方案|哪个方向|确认后)/.test(value)
  return hasDraftFields && !isAnalysisOnly && !asksChoice
}

function messageReplyActions(message) {
  const explicitActions = normalizeReplyActions(message?.suggestedActions || message?.suggested_actions || message?.actions)
  const text = String(message?.text || '')
  if (explicitActions.length) {
    const safeActions = replyRequiresDirectionChoice(text)
      ? explicitActions.filter(action => !['generate_draft', 'regenerate_draft', 'apply_to_draft', 'apply_title'].includes(action.type))
      : explicitActions
    if (replyRequiresDirectionChoice(text) && !safeActions.some(action => action.type === 'recommend_other_directions')) {
      safeActions.push({ type: 'recommend_other_directions', label: '更多 3 个其他方向' })
    }
    const payloadSafeActions = safeActions.filter(action => action.type !== 'apply_to_draft' || replyActionApplyContent(action))
    const title = extractTitleCandidateFromAiReply(text)
    if (title && payloadSafeActions.some(action => action.type === 'apply_to_draft')) {
      return normalizeReplyActions(payloadSafeActions.map(action => action.type === 'apply_to_draft' ? { ...action, type: 'apply_title', label: '应用标题' } : action))
    }
    return payloadSafeActions
  }
  return impliedReplyActionsFromAiText(message)
}

function shouldShowReplyActionsAfterMessage(message) {
  return hasStartedWorkbenchChat.value &&
    message?.role === 'ai' &&
    message.id === latestVisibleAiMessageId.value &&
    !message.streaming &&
    !loading.action &&
    messageReplyActions(message).length > 0
}
const workbenchPrimaryActionText = computed(() => {
  if (loading.action) return hasEditingDraftContent.value ? '应用中' : '生成中'
  return hasEditingDraftContent.value ? '应用到母稿' : '生成母稿'
})
watch(() => workbench.brandId, () => {
  if (restoringWorkbenchContext) return
  workbench.productId = ''
})

watch(
  () => ({
    workbench: {
      brandId: workbench.brandId,
      productId: workbench.productId,
      contentType: workbench.contentType,
      skill: workbench.skill,
      hotspotId: workbench.hotspot?.id || '',
      prompt: workbench.prompt,
    },
    messages: chatMessages.map(message => [message.role, message.text, message.streaming ? 1 : 0]),
    draft: {
      id: editingDraft.id,
      title: editingDraft.title,
      summary: editingDraft.summary,
      body: editingDraft.body,
      keywordsText: editingDraft.keywordsText,
      status: editingDraft.status,
      source: editingDraft.source,
      sourceSnapshot: editingDraft.sourceSnapshot,
      rawStatus: editingDraft.rawStatus,
    },
    idea: { ...ideaSession, directionHistory: [...(ideaSession.directionHistory || [])] },
    stage: aiWorkStage.value,
    intent: aiIntent.value,
    decision: { ...decisionCard },
    previewMode: previewMode.value,
    draftSourceExpanded: draftSourceExpanded.value,
  }),
  scheduleWorkbenchSessionSave,
  { deep: true }
)

const drafts = reactive([])

const plans = reactive([])

const activeTitle = computed(() => menus.find(m => m.key === activeMenu.value)?.label || ({ channelGeneration: '渠道内容生成' }[activeMenu.value] || ''))
const activeSubtitle = computed(() => ({
  overview: '整个应用总览：资料、母稿、发布计划、渠道管理',
  workbench: '人工 AI 创作空间：左侧对话，右侧母稿编辑与预览',
  drafts: '母稿与渠道内容：按日期管理和编辑预览',
  plans: '发布日历与发布队列',
  channels: '渠道资料与账号授权',
  data: '资料中心：品牌 → 商品资料卡 → SKU / 竞品信息',
  channelGeneration: '基于已通过母稿生成各平台渠道版本'
}[activeMenu.value]))

const draftDateOptions = computed(() => [...new Set(drafts.map(d => d.date))].sort((a, b) => b.localeCompare(a)))
const filteredDrafts = computed(() => drafts.filter(d => {
  if (draftDate.value && d.date !== draftDate.value) return false
  if (draftFilter.value !== '全部状态' && d.status !== draftFilter.value) return false
  return true
}))
const DRAFT_PAGE_SIZE = 5
const visibleDraftLimit = ref(DRAFT_PAGE_SIZE)
const draftLoadMoreRef = ref(null)
const visibleDrafts = computed(() => filteredDrafts.value.slice(0, visibleDraftLimit.value))
const hasMoreDrafts = computed(() => visibleDraftLimit.value < filteredDrafts.value.length)
const draftTimelineGroups = computed(() => {
  const map = new Map()
  for (const draft of visibleDrafts.value) {
    if (!map.has(draft.date)) map.set(draft.date, [])
    map.get(draft.date).push(draft)
  }
  return [...map.entries()]
    .sort((a, b) => b[0].localeCompare(a[0]))
    .map(([date, items]) => ({ date, drafts: items }))
})
const timelineWeekdays = ['周日', '周一', '周二', '周三', '周四', '周五', '周六']
function formatTimelineDay(dateStr) {
  const d = new Date(`${dateStr}T12:00:00`)
  return `${d.getMonth() + 1}/${d.getDate()}`
}
function formatTimelineWeekday(dateStr) {
  return timelineWeekdays[new Date(`${dateStr}T12:00:00`).getDay()]
}
function formatGeneratedTime(value) {
  if (!value) return '-'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return String(value)
  const pad = number => String(number).padStart(2, '0')
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())} ${pad(date.getHours())}:${pad(date.getMinutes())}:${pad(date.getSeconds())}`
}
function loadMoreDrafts() {
  if (!hasMoreDrafts.value) return
  visibleDraftLimit.value = Math.min(visibleDraftLimit.value + DRAFT_PAGE_SIZE, filteredDrafts.value.length)
  nextTick(observeDraftLoadMore)
}
function observeDraftLoadMore() {
  if (draftLoadObserver) draftLoadObserver.disconnect()
  const el = draftLoadMoreRef.value
  if (!el || typeof IntersectionObserver === 'undefined') return
  draftLoadObserver = new IntersectionObserver(entries => {
    if (entries.some(entry => entry.isIntersecting)) loadMoreDrafts()
  }, { root: null, rootMargin: '240px 0px', threshold: 0.01 })
  draftLoadObserver.observe(el)
}
const filteredPlans = computed(() => plans.filter(p => p.date === planDate.value))

const planQueueSummary = computed(() => {
  const list = filteredPlans.value
  const canPublish = p => ['已排期', '发布中'].includes(p.publishStatus) && p.accountStatus === '可发布' && p.materialStatus === '完整'
  return {
    total: list.length,
    ready: list.filter(canPublish).length,
    manual: list.filter(p => p.accountStatus === '待人工确认').length,
    blocked: list.filter(p => ['发布失败', '已取消'].includes(p.publishStatus) || ['阻断', '待授权'].includes(p.accountStatus) || p.materialStatus !== '完整').length
  }
})

function queueStatusClass(status) {
  if (['已通过', '可发布', '完整', '已发布'].includes(status)) return 'success'
  if (['阻断', '发布失败', '缺渠道内容', '缺正文'].includes(status)) return 'danger'
  if (['发布中', '已排期'].includes(status)) return 'info'
  return 'warning'
}

const modal = reactive({ type: '', title: '', wide: false })
const modalVisible = computed({
  get: () => Boolean(modal.type),
  set: value => {
    if (!value) closeModal()
  },
})
const modalDialogWidth = computed(() => {
  if (modal.wide || ['import', 'draftView'].includes(modal.type)) return 'min(1080px, 96vw)'
  if (modal.type === 'hotspotForm') return 'min(920px, 94vw)'
  return 'min(640px, 92vw)'
})
const modalDialogHeight = computed(() => {
  if (['import', 'hotspotForm', 'draftView'].includes(modal.type)) return 'min(88vh, 860px)'
  return 'auto'
})
const draftGenerationGuard = computed(() => buildDraftGenerationGuard())
const drawer = reactive({ type: '', title: '' })
const selectedChannel = reactive({})
const selectedDraft = ref(null)
const submittedDraftForChannel = ref(null)
const expandedDraftId = ref(0)
const draftViewParagraphs = computed(() => String(selectedDraft.value?.body || '').split(/\n+/).map(item => item.trim()).filter(Boolean))
const selectedPlan = reactive({})
const publishPlanBodyParagraphs = computed(() => String(selectedPlan.body || '').split(/\n+/).map(item => item.trim()).filter(Boolean))
const selectedProduct = reactive({})
const brandForm = reactive({
  id: 0,
  brand_code: '',
  brand_name: '',
  positioning: '',
  target_audience: '',
  price_band: '',
  keywordsText: '',
})
const hotspotForm = reactive({
  id: 0,
  article_date: '',
  hotspot_topic: '',
  platform: '',
  title: '',
  article_type: '行业趋势',
  author: '',
  summary: '',
  core_viewpoint: '',
  scene_tags: [],
  brand_id: 0,
  product_id: 0,
  sku_text: '',
  competitor_text: '',
  keywords_text: '',
  heat_score: 50,
  priority: 'P1',
  content_status: '待确认',
  owner: '',
  source_url: '',
  captured_at: '',
})
const channelForm = reactive({
  id: 0,
  channel_code: '',
  channel_name: '',
  channel_type: 'content',
  entry_url: '',
  content_forms: '',
  support_modes: '',
  default_publish_mode: 'manual',
  status: 'active',
})
const hotspots = reactive([])
const hotspotSearch = ref('')
const filteredHotspots = computed(() => hotspots.filter(h => !hotspotSearch.value || h.title.includes(hotspotSearch.value) || h.summary.includes(hotspotSearch.value)))
const hotspotExtractForm = reactive({ url: '', platform: '', contentType: '' })
const hotspotExtractResult = ref(null)
const hotspotExtractReturnTo = ref('select')
const styleTemplates = reactive([])
const styleExtractFocusOptions = [
  { value: 'structure', label: '结构路径', desc: '开头、承接、展开、收束' },
  { value: 'tone', label: '语气人设', desc: '亲疏感、专业度、表达密度' },
  { value: 'technique', label: '句式手法', desc: '对比、设问、清单、节奏' },
  { value: 'reuse', label: '复用规则', desc: '可用片段与禁止复制边界' },
]
const styleExtractForm = reactive({
  url: '',
  platform: '',
  contentType: '',
  focuses: styleExtractFocusOptions.map(option => option.value),
})
const styleExtractResult = ref(null)
const styleExtractReturnTo = ref('select')
const selectedStyleDetail = ref(null)
const hotspotArticleTypes = ['行业趋势', '平台热点', '选品趋势', '商品上新', '爆款复盘', '消费趋势', '供应链趋势', '标题优化', '五点描述']
const hotspotSceneTags = ['热点趋势', '选品趋势', '新品开发', '商品上新', '爆款复盘', '跨境趋势', '消费趋势', '供应链趋势', '商品卖点', '卖点提炼', '标题优化', '五点描述']
const hotspotPriorities = ['P0', 'P1', 'P2', 'P3']
const hotspotContentStatuses = ['待确认', '可召回', '已采用', '已过期']
const hotspotProductOptions = computed(() => {
  const brand = brands.find(item => Number(item.id) === Number(hotspotForm.brand_id))
  return brand?.products || brands.flatMap(brandItem => brandItem.products || [])
})
const hotspotSkuOptions = computed(() => {
  const product = hotspotProductOptions.value.find(item => Number(item.id) === Number(hotspotForm.product_id))
  return product?.skus || []
})
const canSaveHotspotForm = computed(() => {
  return Boolean(
    String(hotspotForm.article_date || '').trim() &&
    String(hotspotForm.hotspot_topic || '').trim() &&
    String(hotspotForm.title || '').trim() &&
    String(hotspotForm.platform || '').trim() &&
    String(hotspotForm.summary || '').trim() &&
    String(hotspotForm.core_viewpoint || '').trim() &&
    hotspotForm.scene_tags.length,
  )
})

watch(
  () => [route.params.section, route.query.draft_id, drafts.length],
  ([section]) => {
    if (section === 'channel-generation') restoreChannelGenerationAfterLoad()
  }
)

const selectedHotspotId = computed({
  get: () => workbench.hotspot?.id ?? '',
  set(id) {
    if (id === '__extract__') {
      openHotspotExtractDrawer({ returnTo: 'select' })
      return
    }
    if (id === '__manage__') {
      openHotspotDrawer()
      return
    }
    const next = id ? hotspots.find(h => String(h.id) === String(id)) ?? null : null
    if (next && next.id !== workbench.hotspot?.id) {
      chatMessages.push({ id: Date.now(), role: 'ai', text: `已引用热点：${next.title}` })
    }
    workbench.hotspot = next
  }
})

const selectedStyleTemplateId = computed({
  get: () => workbench.styleTemplate?.id ?? '',
  set(id) {
    if (id === '__extract__') {
      openStyleExtractDrawer({ returnTo: 'select' })
      return
    }
    if (id === '__manage__') {
      openStyleManageDrawer()
      return
    }
    const next = id ? styleTemplates.find(style => String(style.id) === String(id)) ?? null : null
    if (next && next.id !== workbench.styleTemplate?.id) {
      chatMessages.push({ id: Date.now(), role: 'ai', text: `已选择参考写作风格：${next.templateName}` })
    }
    workbench.styleTemplate = next
  }
})

const selectedStyleTemplateSummary = computed(() => {
  const style = workbench.styleTemplate
  if (!style) return ''
  return style.promptFragment || [style.contentType, style.platform].filter(Boolean).join(' · ') || '已选择写作风格模板'
})

const styleDetailPreview = computed(() => styleTemplateDetailPreview(selectedStyleDetail.value || {}))

const hotspotExtractPreview = computed(() => {
  const result = hotspotExtractResult.value || {}
  const draft = result.hotspot_draft || result.hotspotDraft || {}
  const source = result.source || {}
  const keywords = parseJsonArray(apiField(draft, 'keywords', 'Keywords'))
  const angles = parseJsonArray(apiField(draft, 'topic_angles', 'topicAngles', 'TopicAngles'))
  const riskNotes = parseJsonArray(apiField(draft, 'risk_notes', 'riskNotes', 'RiskNotes'))
  const heat = Number(apiField(draft, 'heat_score', 'heatScore', 'HeatScore') || 0)
  return {
    title: apiField(draft, 'title', 'Title') || '外部文章热点',
    platform: apiField(draft, 'platform', 'Platform') || apiField(source, 'source_site', 'sourceSite', 'SourceSite') || '外部来源',
    summary: apiField(draft, 'summary', 'Summary') || '已从外部文章提炼热点摘要',
    searchIntent: apiField(draft, 'search_intent', 'searchIntent', 'SearchIntent') || '已提炼搜索意图',
    keywords,
    keywordsText: keywords.length ? keywords.join('、') : '已提炼关键词',
    angles,
    anglesText: angles.length ? `借势角度：${angles.join(' / ')}` : '已提炼借势角度',
    heat,
    riskNotes,
    riskText: [`热度 ${heat || 60}`, ...(riskNotes.length ? riskNotes : ['需人工确认真实性和时效性'])].join(' · '),
    sourceId: Number(apiField(draft, 'source_id', 'sourceId', 'SourceID') || apiField(source, 'id', 'ID') || 0),
    sourceTitle: apiField(source, 'source_title', 'sourceTitle', 'SourceTitle') || '外部文章',
    sourceUrl: apiField(draft, 'source_url', 'sourceUrl', 'SourceURL') || apiField(source, 'source_url', 'sourceUrl', 'SourceURL') || '',
    capturedAt: apiField(draft, 'captured_at', 'capturedAt', 'CapturedAt') || apiField(source, 'captured_at', 'capturedAt', 'CapturedAt') || '',
    cleanText: apiField(source, 'clean_text', 'cleanText', 'CleanText') || apiField(source, 'raw_text', 'rawText', 'RawText') || '',
  }
})

const styleExtractPreview = computed(() => {
  const result = styleExtractResult.value || {}
  const style = normalizeStyleTemplate(result.style_template || result.styleTemplate || {})
  const source = result.source || {}
  const detail = styleTemplateDetailPreview(style)
  return {
    ...detail,
    sourceTitle: apiField(source, 'source_title', 'sourceTitle', 'SourceTitle') || '外部文章',
    sourceUrl: apiField(source, 'source_url', 'sourceUrl', 'SourceURL') || '',
    cleanText: apiField(source, 'clean_text', 'cleanText', 'CleanText') || apiField(source, 'raw_text', 'rawText', 'RawText') || '',
  }
})

const importForm = reactive({ type: 'brand', rawText: '' })
const importFileInput = ref(null)
const selectedImportFileName = ref('')
const importResult = ref(null)
const importErrors = reactive([])
const importFieldSets = {
  brand: [
    { label: '品牌编码', key: 'brand_code', required: true, sample: 'BRAND-001' },
    { label: '品牌名称', key: 'brand_name', required: true, sample: 'Mardi Ladin' },
    { label: '品牌定位', key: 'positioning', required: false, sample: '轻法式通勤女装品牌' },
    { label: '目标人群', key: 'target_audience', required: false, sample: '18-35 岁都市女性' },
    { label: '价格带', key: 'price_band', required: false, sample: '299-899 元' },
    { label: '品牌关键词', key: 'keywords', required: false, sample: 'Mardi Ladin、玛尔迪、轻法式' },
  ],
  product: [
    { label: '品牌编码', key: 'brand_code', required: true, sample: 'BRAND-001' },
    { label: '商品编码', key: 'product_code', required: true, sample: 'SPU-DRESS-001' },
    { label: '商品名称', key: 'product_name', required: true, sample: '法式通勤连衣裙' },
    { label: '类目', key: 'category_name', required: true, sample: '女装/连衣裙' },
    { label: '系列', key: 'series', required: false, sample: '春夏通勤系列' },
    { label: '适合人群', key: 'target_audience', required: false, sample: '小个子、梨形身材、通勤女性' },
    { label: '商品卖点', key: 'selling_points', required: false, sample: '收腰显比例、面料垂顺' },
    { label: '常见问题', key: 'faq', required: false, sample: '小个子能穿吗、通勤合适吗' },
    { label: '内容角度', key: 'content_angles', required: false, sample: '轻法式通勤、小个子穿搭' },
  ],
  sku: [
    { label: '商品编码', key: 'product_code', required: true, sample: 'SPU-DRESS-001' },
    { label: 'SKU 编码', key: 'sku_code', required: true, sample: 'SKU-D001-BLK-S' },
    { label: 'SKU 名称', key: 'sku_name', required: false, sample: '黑色 S 码' },
    { label: '颜色', key: 'color', required: false, sample: '黑色' },
    { label: '尺码', key: 'size', required: false, sample: 'S' },
    { label: '价格', key: 'price', required: false, sample: '399' },
    { label: '面料', key: 'fabric', required: false, sample: '醋酸混纺' },
    { label: '版型', key: 'fit', required: false, sample: '收腰 A 字版型' },
    { label: '风格标签', key: 'style_tags', required: false, sample: '轻法式、显瘦、通勤' },
    { label: '场景标签', key: 'scene_tags', required: false, sample: '上班、约会、周末出街' },
    { label: '产品图', key: 'product_image_url', required: false, sample: 'https://example.com/product.jpg' },
    { label: '模特图', key: 'model_image_url', required: false, sample: 'https://example.com/model.jpg' },
  ],
  keyword: [
    { label: '品牌编码', key: 'brand_code', required: true, sample: 'BRAND-001' },
    { label: '商品编码', key: 'product_code', required: false, sample: 'SPU-DRESS-001' },
    { label: '关键词包名称', key: 'package_name', required: false, sample: '首批关键词分组' },
    { label: '关键词分组', key: 'keyword_group', required: true, sample: '场景词' },
    { label: '关键词', key: 'keywords', required: true, sample: '法式通勤穿搭、约会穿搭、旅行穿搭' },
    { label: '意图', key: 'intent', required: false, sample: '场景搜索' },
    { label: '权重', key: 'weight', required: false, sample: '80' },
    { label: '来源', key: 'source', required: false, sample: 'import' },
    { label: '合规禁用', key: 'forbidden_terms', required: false, sample: '法国品牌、源自法国' },
  ],
  competitor: [
    { label: '商品编码', key: 'product_code', required: true, sample: 'SPU-DRESS-001' },
    { label: '竞品品牌', key: 'brand_name', required: true, sample: 'Lily' },
    { label: '竞品商品名称', key: 'product_name', required: true, sample: '法式 A 字裙' },
    { label: '竞品价格', key: 'price_text', required: false, sample: '299-399' },
    { label: '竞品卖点', key: 'point', required: false, sample: '年轻通勤、版型修身' },
    { label: '与我方差异', key: 'difference', required: false, sample: '我们的场景覆盖更丰富' },
    { label: '内容角度', key: 'angle', required: false, sample: '春夏通勤穿搭清单' },
    { label: '竞品链接', key: 'link_url', required: false, sample: 'https://example.com/item' },
  ],
}
const importTypeOptions = [
  { value: 'brand', label: '导入品牌', desc: '品牌编码、名称、定位和关键词', help: '先导入品牌，商品资料会通过品牌编码归属到对应品牌。', placeholder: '品牌编码,品牌名称,品牌定位,目标人群,价格带,品牌关键词\nBRAND-001,Mardi Ladin,轻法式通勤女装品牌,18-35 岁都市女性,299-899 元,Mardi Ladin、玛尔迪、轻法式' },
  { value: 'product', label: '导入商品', desc: '商品 SPU、类目、系列、卖点', help: '商品必须带品牌编码；系列、适合人群等会进入内容角度和卖点上下文。', placeholder: '品牌编码,商品编码,商品名称,类目,系列,适合人群,商品卖点,常见问题,内容角度\nBRAND-001,SPU-DRESS-001,法式通勤连衣裙,女装/连衣裙,春夏通勤系列,小个子通勤女性,收腰显比例、面料垂顺,小个子能穿吗,轻法式通勤、小个子穿搭' },
  { value: 'sku', label: '导入 SKU', desc: 'SKU 编码、颜色、尺码、价格和图片', help: 'SKU 必须带商品编码；颜色、尺码、面料、版型等会保存为 SKU 属性。', placeholder: '商品编码,SKU编码,SKU名称,颜色,尺码,价格,面料,版型,风格标签,场景标签,产品图,模特图\nSPU-DRESS-001,SKU-D001-BLK-S,黑色 S 码,黑色,S,399,醋酸混纺,收腰 A 字版型,轻法式、显瘦、通勤,上班、约会,https://example.com/product.jpg,https://example.com/model.jpg' },
  { value: 'keyword', label: '导入关键词', desc: '品牌词、场景词、人群词和禁用词', help: '关键词会按品牌编码、商品编码归属；一格里多个关键词会拆成多条记录。', placeholder: '品牌编码,商品编码,关键词包名称,关键词分组,关键词,意图,权重,来源,合规禁用\nBRAND-001,,首批关键词分组,品牌词,Mardi Ladin、玛尔迪、玛尔迪女装,品牌认知,90,import,法国品牌、源自法国\nBRAND-001,SPU-DRESS-001,法国概念关键词包,场景词,巴黎日常穿搭、法式通勤穿搭、法式约会穿搭,场景搜索,80,import,' },
  { value: 'competitor', label: '导入竞品', desc: '竞品品牌、商品、差异和链接', help: '竞品必须关联本方商品编码，用于生成对比内容和避坑角度。', placeholder: '商品编码,竞品品牌,竞品商品名称,竞品价格,竞品卖点,与我方差异,内容角度,竞品链接\nSPU-DRESS-001,Lily,法式 A 字裙,299-399,年轻通勤、版型修身,我们的场景覆盖更丰富,春夏通勤穿搭清单,https://example.com/item' },
]
const currentImportOption = computed(() => importTypeOptions.find(item => item.value === importForm.type) || importTypeOptions[0])
const currentImportFields = computed(() => importFieldSets[importForm.type] || importFieldSets.brand)
const importFieldHeaderText = computed(() => currentImportFields.value.map(field => field.key).join(', '))
const importPreviewRows = computed(() => parseImportText(importForm.rawText, currentImportFields.value))
const importSuccessCount = computed(() => Number(apiField(importResult.value, 'success_count', 'SuccessCount') || 0))
const importFailedCount = computed(() => Number(apiField(importResult.value, 'failed_count', 'FailedCount') || 0))
const importBatchID = computed(() => Number(apiField(importResult.value, 'id', 'ID') || 0))

const newPlanDraftId = ref(201)
const newPlanChannel = ref('小红书')
const newPlanMethod = ref('Agent 执行')
const newPlanLevel = ref('半自动')
const newPlanTime = ref('2026-05-20T18:00')
const channelPlanForm = reactive({
  selectedChannels: [],
  scheduledAt: '',
})

onMounted(async () => {
  try {
    await permissionStore.load()
  } catch {
    showToast('权限信息加载失败，按钮将以后端校验为准')
  }
  loadContentTypeDictionary()
  loadAiGeoData()
  nextTick(observeDraftLoadMore)
})

onBeforeUnmount(() => {
  persistWorkbenchSessionNow()
  if (workbenchChatScrollFrame) window.cancelAnimationFrame(workbenchChatScrollFrame)
  if (workbenchSessionSaveTimer) window.clearTimeout(workbenchSessionSaveTimer)
  if (draftLoadObserver) draftLoadObserver.disconnect()
})

async function loadContentTypeDictionary() {
  try {
    const data = await fetchDictItemsByCode('ai_geo_content_type')
    const items = (data.items || [])
      .filter(item => item.enabled !== false)
      .sort((a, b) => Number(a.sort_order || 0) - Number(b.sort_order || 0))
      .map(item => ({ value: item.label || item.value, label: item.label || item.value }))
      .filter(item => item.value)
    if (items.length) {
      contentTypeOptions.value = items
      if (!items.some(item => item.value === workbench.contentType)) {
        workbench.contentType = items[0].value || DEFAULT_CONTENT_TYPE
      }
    }
  } catch {
    contentTypeOptions.value = [...DEFAULT_CONTENT_TYPE_OPTIONS]
  }
}

watch(draftLoadMoreRef, () => nextTick(observeDraftLoadMore))

watch([draftDate, draftFilter], () => {
  visibleDraftLimit.value = DRAFT_PAGE_SIZE
  expandedDraftId.value = 0
  nextTick(observeDraftLoadMore)
})

watch(activeMenu, menu => {
  if (menu === 'drafts') nextTick(observeDraftLoadMore)
  if (menu === 'workbench') queueWorkbenchChatScrollAfterRender(true)
})

watch(planDate, () => {
  loadPlanCalendar()
})

watch(() => hotspotForm.product_id, () => {
  if (!hotspotForm.sku_text || hotspotForm.sku_text === '全 SKU') return
  const exists = hotspotSkuOptions.value.some(sku => skuLabel(sku) === hotspotForm.sku_text)
  if (!exists) hotspotForm.sku_text = ''
})

async function loadAiGeoData() {
  loading.page = true
  try {
    const calendarRange = publishCalendarRange()
    const [overviewData, brandPage, productPage, skuPage, competitorPage, keywordPage, assetPage, hotspotPage, channelPage, channelAccountPage, draftPage, channelContentPage, planPage, planCalendar] = await Promise.all([
      fetchAiGeoOverview(),
      fetchAiGeoBrands({ limit: 200 }),
      fetchAiGeoProducts({ limit: 200 }),
      fetchAiGeoSKUs({ limit: 500 }),
      fetchAiGeoCompetitors({ limit: 500 }),
      fetchAiGeoKeywords({ limit: 500, status: 'active' }),
      fetchAiGeoMaterialAssets({ limit: 500 }),
      fetchAiGeoHotspots({ limit: 200, status: 'active' }),
      fetchAiGeoChannels({ limit: 200 }),
      fetchAiGeoChannelAccounts({ limit: 500 }),
      fetchAiGeoDrafts({ limit: 200 }),
      fetchAiGeoChannelContents({ limit: 500 }),
      fetchAiGeoPublishPlans({ limit: 200 }),
      fetchAiGeoPublishPlanCalendar(calendarRange),
    ])
    Object.assign(overview, overviewData)
    hydrateBrands(brandPage.items || [], productPage.items || [], skuPage.items || [], competitorPage.items || [], assetPage.items || [], keywordPage.items || [])
    hydrateHotspots(hotspotPage.items || [])
    hydrateChannels(channelPage.items || [])
    hydrateChannelAccounts(channelAccountPage.items || [])
    hydrateDrafts(draftPage.items || [], channelContentPage.items || [])
    restoreWorkbenchDraftAfterLoad()
    restoreChannelGenerationAfterLoad()
    hydratePlans(planPage.items || [])
    planCalendarDays.value = planCalendar.days || []
    loadStyleTemplates()
  } catch (error) {
    showToast(error?.message || 'AI GEO 数据加载失败')
  } finally {
    loading.page = false
  }
}

async function loadStyleTemplates() {
  try {
    const styleTemplatePage = await fetchAiGeoStyleTemplates({ limit: 200, status: 'active' })
    hydrateStyleTemplates(styleTemplatePage.items || [])
  } catch (error) {
    console.warn('[ai-geo] style template API unavailable', error)
    hydrateStyleTemplates([])
  }
}

async function loadPlanCalendar() {
  try {
    const calendar = await fetchAiGeoPublishPlanCalendar(publishCalendarRange())
    planCalendarDays.value = calendar.days || []
  } catch (error) {
    showToast(error?.message || '发布日历加载失败')
  }
}

function publishCalendarRange() {
  const base = new Date(planDate.value || new Date())
  if (Number.isNaN(base.getTime())) {
    return {}
  }
  const start = new Date(base)
  start.setDate(base.getDate() - 2)
  const end = new Date(base)
  end.setDate(base.getDate() + 4)
  return {
    start_date: start.toISOString().slice(0, 10),
    end_date: end.toISOString().slice(0, 10),
  }
}

function apiField(obj, ...keys) {
  for (const key of keys) {
    if (obj?.[key] !== undefined && obj?.[key] !== null) return obj[key]
  }
  return undefined
}

function hydrateBrands(apiBrands, apiProducts, apiSkus = [], apiCompetitors = [], apiAssets = [], apiKeywords = []) {
  brands.splice(0, brands.length, ...apiBrands.map(brand => ({
    id: apiField(brand, 'id', 'ID'),
    code: apiField(brand, 'brand_code', 'BrandCode') || '',
    name: apiField(brand, 'brand_name', 'BrandName') || '',
    position: apiField(brand, 'positioning', 'Positioning') || '',
    audience: apiField(brand, 'target_audience', 'TargetAudience') || '',
    priceBand: apiField(brand, 'price_band', 'PriceBand') || '',
    tone: apiField(brand, 'tone', 'Tone') || '',
    completeness: apiField(brand, 'completeness', 'Completeness') || 0,
    keywordGroups: keywordGroupsForBrand(brand, apiKeywords),
    materials: apiAssets
      .filter(asset => Number(apiField(asset, 'brand_id', 'BrandID')) === Number(apiField(brand, 'id', 'ID')))
      .map(asset => apiField(asset, 'asset_name', 'AssetName') || apiField(asset, 'asset_type', 'AssetType'))
      .filter(Boolean),
    products: apiProducts.filter(product => Number(apiField(product, 'brand_id', 'BrandID')) === Number(apiField(brand, 'id', 'ID'))).map(product => {
      const productId = apiField(product, 'id', 'ID')
      const productSkus = apiSkus.filter(sku => Number(apiField(sku, 'product_id', 'ProductID')) === Number(productId)).map(sku => skuFromApi(sku))
      const productCompetitors = apiCompetitors.filter(comp => Number(apiField(comp, 'product_id', 'ProductID')) === Number(productId)).map(comp => competitorFromApi(comp))
      const productKeywords = apiKeywords
        .filter(keyword => Number(apiField(keyword, 'product_id', 'ProductID')) === Number(productId))
        .map(keyword => apiField(keyword, 'keyword', 'Keyword'))
        .filter(Boolean)
      return {
        id: productId,
        code: apiField(product, 'product_code', 'ProductCode') || '',
        name: apiField(product, 'product_name', 'ProductName') || '',
        category: apiField(product, 'category_name', 'CategoryName') || '',
        series: '',
        priceRange: productSkus.length ? priceRangeFromSkus(productSkus) : '',
        fabric: '',
        fit: '',
        styleTags: parseJsonArray(apiField(product, 'content_angles', 'ContentAngles')),
        sceneTags: [],
        audience: '',
        sellingPoints: parseJsonArray(apiField(product, 'selling_points', 'SellingPoints')).join('、'),
        completeness: apiField(product, 'completeness', 'Completeness') || 0,
        missing: [],
        keywords: productKeywords,
        competitors: productCompetitors,
        skus: productSkus,
      }
    }),
  })))
  if (!brands.find(b => b.id === Number(selectedBrandId.value))) {
    selectedBrandId.value = brands[0]?.id || 0
  }
}

function keywordGroupsForBrand(brand, apiKeywords = []) {
  const brandId = Number(apiField(brand, 'id', 'ID'))
  const scoped = apiKeywords.filter(keyword => Number(apiField(keyword, 'brand_id', 'BrandID')) === brandId && !apiField(keyword, 'product_id', 'ProductID'))
  if (!scoped.length) {
    const fallback = parseJsonArray(apiField(brand, 'keywords', 'Keywords'))
    return fallback.length ? [{ id: `brand-${brandId}`, name: '品牌关键词', keywords: fallback }] : []
  }
  const groups = new Map()
  scoped.forEach(keyword => {
    const name = apiField(keyword, 'keyword_group', 'KeywordGroup') || '通用关键词'
    if (!groups.has(name)) groups.set(name, [])
    const value = apiField(keyword, 'keyword', 'Keyword')
    if (value) groups.get(name).push(value)
  })
  return Array.from(groups.entries()).map(([name, keywords], index) => ({ id: `${brandId}-${name}-${index}`, name, keywords }))
}

function hydrateChannels(apiChannels) {
  channelProfiles.splice(0, channelProfiles.length, ...apiChannels.map(channel => ({
    id: apiField(channel, 'id', 'ID'),
    code: apiField(channel, 'channel_code', 'ChannelCode') || '',
    name: apiField(channel, 'channel_name', 'ChannelName') || '',
    desc: channelDescription(apiField(channel, 'channel_code', 'ChannelCode'), apiField(channel, 'channel_name', 'ChannelName')),
    type: apiField(channel, 'channel_type', 'ChannelType') || '',
    siteUrl: apiField(channel, 'entry_url', 'EntryURL') || '',
    adminUrl: channelAdminUrl(apiField(channel, 'channel_code', 'ChannelCode'), apiField(channel, 'entry_url', 'EntryURL')),
    contentTypes: parseJsonArray(apiField(channel, 'content_forms', 'ContentForms')).join(' / '),
    supportMethods: parseJsonArray(apiField(channel, 'support_modes', 'SupportModes')).join(' / '),
    defaultMethod: publishModeLabel(apiField(channel, 'default_publish_mode', 'DefaultPublishMode')),
    accountCount: 0,
    status: channelAccessStatus(apiField(channel, 'status', 'Status')),
    method: apiField(channel, 'default_publish_mode', 'DefaultPublishMode'),
    level: apiField(channel, 'default_publish_mode', 'DefaultPublishMode') === 'manual' ? '人工' : '半自动',
    skill: channelSkill(apiField(channel, 'channel_code', 'ChannelCode'), apiField(channel, 'channel_name', 'ChannelName')),
    risk: channelRisk(apiField(channel, 'channel_code', 'ChannelCode'), apiField(channel, 'channel_name', 'ChannelName')),
    enabled: apiField(channel, 'status', 'Status') === 'active',
  })))
}

function hydrateChannelAccounts(apiAccounts) {
  channelAccounts.splice(0, channelAccounts.length, ...apiAccounts.map(account => {
    const channelID = Number(apiField(account, 'channel_id', 'ChannelID'))
    const channel = channelProfiles.find(item => Number(item.id) === channelID)
    return {
      id: apiField(account, 'id', 'ID'),
      channelId: channelID,
      channel: channel?.name || `渠道 ${channelID}`,
      channelCode: channel?.code || '',
      accountName: apiField(account, 'account_name', 'AccountName') || '',
      login: apiField(account, 'external_account_id', 'ExternalAccountID') || '未绑定',
      status: accountAuthLabel(apiField(account, 'auth_status', 'AuthStatus'), apiField(account, 'publish_status', 'PublishStatus')),
      updatedAt: formatGeneratedTime(apiField(account, 'updated_at', 'UpdatedAt')),
      raw: account,
    }
  }))
  const accountCountByChannel = channelAccounts.reduce((map, account) => {
    map.set(account.channelId, (map.get(account.channelId) || 0) + 1)
    return map
  }, new Map())
  channelProfiles.forEach(channel => {
    channel.accountCount = accountCountByChannel.get(Number(channel.id)) || 0
    if (channel.accountCount === 0 && channel.status === '可发布') channel.status = '待授权'
  })
}

function skuFromApi(sku) {
  const attributes = parseJsonObject(apiField(sku, 'attributes', 'Attributes'))
  return {
    id: apiField(sku, 'id', 'ID'),
    code: apiField(sku, 'sku_code', 'SKUCode') || '',
    name: apiField(sku, 'sku_name', 'SKUName') || '',
    color: attributes.color || attributes.颜色 || '',
    size: attributes.size || attributes.尺码 || '',
    price: Number(apiField(sku, 'price', 'Price') || 0),
    status: skuStatusLabel(apiField(sku, 'stock_status', 'StockStatus') || apiField(sku, 'status', 'Status')),
    overrideTitle: attributes.override_title || attributes.title || '',
    overridePoint: attributes.override_point || attributes.selling_point || '',
  }
}

function hydrateHotspots(apiHotspots) {
  hotspots.splice(0, hotspots.length, ...apiHotspots.map(item => ({
    ...hotspotFromApi(item),
  })))
  if (workbench.hotspot && !hotspots.find(h => Number(h.id) === Number(workbench.hotspot.id))) {
    workbench.hotspot = null
  }
}

function hydrateStyleTemplates(apiStyles) {
  styleTemplates.splice(0, styleTemplates.length, ...apiStyles.map(normalizeStyleTemplate))
  if (workbench.styleTemplate && !styleTemplates.find(style => Number(style.id) === Number(workbench.styleTemplate.id))) {
    workbench.styleTemplate = null
  }
}

function normalizeStyleTemplate(item) {
  const extraction = parseJsonObject(apiField(item, 'extraction_summary', 'extractionSummary', 'ExtractionSummary'))
  return {
    id: apiField(item, 'id', 'ID'),
    sourceId: apiField(item, 'source_id', 'sourceId', 'SourceID'),
    templateCode: apiField(item, 'template_code', 'templateCode', 'TemplateCode') || '',
    templateName: apiField(item, 'template_name', 'templateName', 'TemplateName') || extraction.source_title || '参考写作风格',
    description: apiField(item, 'description', 'Description') || '',
    contentType: apiField(item, 'content_type', 'contentType', 'ContentType') || '',
    platform: apiField(item, 'platform', 'Platform') || '',
    tone_profile: apiField(item, 'tone_profile', 'toneProfile', 'ToneProfile') || {},
    structure_profile: apiField(item, 'structure_profile', 'structureProfile', 'StructureProfile') || {},
    technique_profile: apiField(item, 'technique_profile', 'techniqueProfile', 'TechniqueProfile') || {},
    promptFragment: apiField(item, 'prompt_fragment', 'promptFragment', 'PromptFragment') || '',
    negativeRules: apiField(item, 'negative_rules', 'negativeRules', 'NegativeRules') || [],
    styleKeywords: apiField(item, 'style_keywords', 'styleKeywords', 'StyleKeywords') || [],
    extractionSummary: extraction,
    status: apiField(item, 'status', 'Status') || 'active',
  }
}

function styleTemplateDetailPreview(style = {}) {
  const extraction = parseJsonObject(style.extractionSummary || style.extraction_summary)
  const tone = parseJsonObject(style.tone_profile || style.toneProfile)
  const structure = parseJsonObject(style.structure_profile || style.structureProfile)
  const technique = parseJsonObject(style.technique_profile || style.techniqueProfile)
  const negativeRules = parseJsonArray(style.negativeRules || style.negative_rules)
  const keywords = parseJsonArray(style.styleKeywords || style.style_keywords)
  return {
    templateName: style.templateName || style.template_name || '参考写作风格',
    tone: [tone.voice, tone.sentence_density, tone.point_of_view].filter(Boolean).join(' · ') || '已提炼语气画像',
    structure: [structure.opening, ...(Array.isArray(structure.flow) ? structure.flow : [])].filter(Boolean).join(' / ') || '已提炼结构',
    techniques: Array.isArray(technique.techniques) ? technique.techniques.join('、') : '已提炼写作手法',
    promptFragment: style.promptFragment || style.prompt_fragment || '已形成可复用风格提示词。',
    negativeRules: negativeRules.length ? negativeRules.join('\n') : '不得复制原文句子\n不得照搬外部事实',
    meta: [
      style.platform || '通用平台',
      style.contentType || style.content_type || '通用内容',
      keywords.length ? keywords.join('、') : '',
    ].filter(Boolean).join(' · '),
    sourceExcerpt: extraction.source_excerpt || '',
  }
}

function styleTemplateSourceSnapshot(style = null) {
  if (!style) return null
  const tone = parseJsonObject(style.toneProfile || style.tone_profile)
  const structure = parseJsonObject(style.structureProfile || style.structure_profile)
  const technique = parseJsonObject(style.techniqueProfile || style.technique_profile)
  return {
    id: style.id,
    template_code: style.templateCode || style.template_code || '',
    template_name: style.templateName || style.template_name || '参考写作风格',
    platform: style.platform || '通用平台',
    content_type: style.contentType || style.content_type || '通用内容',
    tone_profile: tone,
    structure_profile: structure,
    technique_profile: technique,
    style_keywords: parseJsonArray(style.styleKeywords || style.style_keywords).slice(0, 12),
    prompt_fragment: style.promptFragment || style.prompt_fragment || '',
    negative_rules: parseJsonArray(style.negativeRules || style.negative_rules),
  }
}

function hotspotFromApi(item) {
  const metadata = parseJsonObject(apiField(item, 'metadata', 'Metadata'))
  return {
    id: apiField(item, 'id', 'ID'),
    title: apiField(item, 'title', 'Title') || '',
    summary: metadata.summary || apiField(item, 'source_url', 'SourceURL') || '来自热点资料库，可作为选题借势上下文。',
    platform: apiField(item, 'platform', 'Platform') || '热点',
    heat: Number(apiField(item, 'heat_score', 'HeatScore') || 0),
    sourceUrl: apiField(item, 'source_url', 'SourceURL') || '',
    capturedAt: apiField(item, 'captured_at', 'CapturedAt') || '',
    metadata,
    articleDate: metadata.article_date || dateOnly(apiField(item, 'captured_at', 'CapturedAt')),
    hotspotTopic: metadata.hotspot_topic || '',
    articleType: metadata.article_type || '行业趋势',
    author: metadata.author || '',
    coreViewpoint: metadata.core_viewpoint || '',
    sceneTags: Array.isArray(metadata.scene_tags) ? metadata.scene_tags : [],
    brandId: Number(metadata.brand_id || 0),
    productId: Number(metadata.product_id || 0),
    skuText: metadata.sku_text || '',
    competitorText: metadata.competitor_text || '',
    keywordsText: Array.isArray(metadata.keywords) ? metadata.keywords.join('、') : String(metadata.keywords || ''),
    priority: metadata.priority || 'P1',
    contentStatus: metadata.content_status || '待确认',
    owner: metadata.owner || '',
    risk: '待判断',
  }
}

function competitorFromApi(comp) {
  return {
    id: apiField(comp, 'id', 'ID'),
    brand: apiField(comp, 'brand_name', 'BrandName') || '',
    name: apiField(comp, 'product_name', 'ProductName') || '',
    price: apiField(comp, 'price_text', 'PriceText') || '待录入',
    point: apiField(comp, 'point', 'Point') || '待录入',
    diff: apiField(comp, 'difference', 'Difference') || '待分析',
    angle: apiField(comp, 'angle', 'Angle') || '待生成',
    link: apiField(comp, 'link_url', 'LinkURL') || '',
  }
}

function priceRangeFromSkus(skus) {
  const prices = skus.map(sku => Number(sku.price || 0)).filter(price => price > 0)
  if (!prices.length) return ''
  const min = Math.min(...prices)
  const max = Math.max(...prices)
  return min === max ? `${min} 元` : `${min}-${max} 元`
}

function hydrateDrafts(apiDrafts, apiChannelContents = []) {
  channelContentIndex.splice(0, channelContentIndex.length, ...apiChannelContents.map(channelContentFromApi))
  drafts.splice(0, drafts.length, ...apiDrafts.map(draft => ({
    raw: draft,
    id: apiField(draft, 'id', 'ID'),
    brandId: apiField(draft, 'brand_id', 'BrandID') || '',
    productId: apiField(draft, 'product_id', 'ProductID') || '',
    generatedAt: apiField(draft, 'created_at', 'CreatedAt') || '',
    generatedAtLabel: formatGeneratedTime(apiField(draft, 'created_at', 'CreatedAt')),
    date: String(apiField(draft, 'created_at', 'CreatedAt') || new Date().toISOString()).slice(0, 10),
    title: apiField(draft, 'title', 'Title') || '',
    summary: apiField(draft, 'summary', 'Summary') || '',
    body: apiField(draft, 'body', 'Body') || '',
    keywords: apiField(draft, 'keywords', 'Keywords') || '[]',
    conversation: parseJsonArray(apiField(draft, 'conversation', 'Conversation')),
    sourceSnapshot: parseJsonObject(apiField(draft, 'source_snapshot', 'SourceSnapshot')),
    source: sourceLabel(apiField(draft, 'source', 'Source')),
    status: draftStatusLabel(apiField(draft, 'audit_status', 'AuditStatus')),
    flow: '直接流转',
    channels: channelContentIndex
      .filter(content => Number(content.draftId) === Number(apiField(draft, 'id', 'ID'))),
    rawStatus: apiField(draft, 'audit_status', 'AuditStatus'),
  })))
}

function rememberWorkbenchDraft(draftId) {
  const id = Number(draftId || editingDraft.id || 0)
  if (!id) return
  window.localStorage?.removeItem(WORKBENCH_RESET_STORAGE_KEY)
  window.localStorage?.setItem(WORKBENCH_DRAFT_STORAGE_KEY, String(id))
  if (activeMenu.value !== 'workbench') return
  const currentId = String(route.query.draft_id || '')
  if (currentId === String(id)) return
  router.replace({
    path: menuRouteMap.workbench,
    query: { ...route.query, draft_id: String(id) },
  }).catch(() => {})
}

function findDraftById(draftId) {
  const id = Number(draftId || 0)
  if (!id) return null
  return drafts.find(draft => Number(draft.id) === id) || null
}

function hasWorkbenchDraftState() {
  return Boolean(
    editingDraft.id ||
    String(editingDraft.title || editingDraft.summary || editingDraft.body || '').trim() ||
    String(workbench.prompt || '').trim() ||
    chatMessages.some(message => String(message.text || '').trim())
  )
}

function workbenchSessionSnapshot() {
  return {
    version: 1,
    savedAt: new Date().toISOString(),
    workbench: {
      brandId: workbench.brandId,
      productId: workbench.productId,
      contentType: workbench.contentType,
      skill: workbench.skill,
      hotspotId: workbench.hotspot?.id || null,
      prompt: workbench.prompt,
    },
    chatMessages: chatMessages
      .filter(message => String(message.text || '').trim())
      .slice(-40)
      .map(message => ({
        id: message.id,
        role: message.role,
        text: message.text,
        hidden: Boolean(message.hidden),
        createdAt: message.createdAt || '',
      })),
    editingDraft: {
      id: editingDraft.id,
      title: editingDraft.title,
      summary: editingDraft.summary,
      body: editingDraft.body,
      keywordsText: editingDraft.keywordsText,
      status: editingDraft.status,
      source: editingDraft.source,
      sourceSnapshot: editingDraft.sourceSnapshot,
      rawStatus: editingDraft.rawStatus,
    },
    ideaSession: { ...ideaSession, directionHistory: [...(ideaSession.directionHistory || [])] },
    aiWorkStage: aiWorkStage.value,
    aiIntent: aiIntent.value,
    decisionCard: { ...decisionCard },
    previewMode: previewMode.value,
    draftSourceExpanded: draftSourceExpanded.value,
  }
}

function persistWorkbenchSessionNow() {
  if (restoringWorkbenchContext) return
  if (!hasWorkbenchDraftState()) {
    window.localStorage?.removeItem(WORKBENCH_SESSION_STORAGE_KEY)
    return
  }
  try {
    window.localStorage?.removeItem(WORKBENCH_RESET_STORAGE_KEY)
    window.localStorage?.setItem(WORKBENCH_SESSION_STORAGE_KEY, JSON.stringify(workbenchSessionSnapshot()))
  } catch {
    // localStorage can be full or disabled; failing to cache should not block editing.
  }
}

function scheduleWorkbenchSessionSave() {
  if (restoringWorkbenchContext) return
  if (workbenchSessionSaveTimer) window.clearTimeout(workbenchSessionSaveTimer)
  workbenchSessionSaveTimer = window.setTimeout(() => {
    workbenchSessionSaveTimer = 0
    persistWorkbenchSessionNow()
  }, 250)
}

function clearWorkbenchSessionCache() {
  if (workbenchSessionSaveTimer) {
    window.clearTimeout(workbenchSessionSaveTimer)
    workbenchSessionSaveTimer = 0
  }
  window.localStorage?.removeItem(WORKBENCH_SESSION_STORAGE_KEY)
}

function restoreWorkbenchLocalSession() {
  if (hasWorkbenchDraftState()) return false
  const raw = window.localStorage?.getItem(WORKBENCH_SESSION_STORAGE_KEY)
  if (!raw) return false
  let snapshot
  try {
    snapshot = JSON.parse(raw)
  } catch {
    clearWorkbenchSessionCache()
    return false
  }
  if (!snapshot || snapshot.version !== 1) return false
  restoringWorkbenchContext = true
  const cachedWorkbench = snapshot.workbench || {}
  workbench.brandId = cachedWorkbench.brandId || ''
  selectedBrandId.value = Number(cachedWorkbench.brandId || selectedBrandId.value || 0)
  workbench.productId = cachedWorkbench.productId || ''
  workbench.contentType = cachedWorkbench.contentType || DEFAULT_CONTENT_TYPE
  workbench.skill = cachedWorkbench.skill || INTERNAL_DRAFT_SKILL
  workbench.hotspot = cachedWorkbench.hotspotId ? hotspots.find(hot => String(hot.id) === String(cachedWorkbench.hotspotId)) || null : null
  workbench.prompt = cachedWorkbench.prompt || ''
  chatMessages.splice(0, chatMessages.length, ...(snapshot.chatMessages || []).map((message, index) => ({
    id: message.id || Date.now() + index,
    role: message.role,
    text: message.text || '',
    hidden: Boolean(message.hidden),
    createdAt: message.createdAt || '',
  })).filter(message => message.text && ['user', 'ai', 'assistant', 'system'].includes(message.role) && !isLegacyExpertUnavailableMessage(message.text)))
  Object.assign(editingDraft, {
    id: 0,
    title: '',
    summary: '',
    body: '',
    keywordsText: '',
    status: '草稿',
    source: '人工创作',
    sourceSnapshot: {},
    ...(snapshot.editingDraft || {}),
  })
  if (editingDraft.id) markEditingDraftSaved()
  else clearEditingDraftSavedSnapshot()
  Object.assign(ideaSession, {
    stage: 'collecting',
    intent: '等待想法',
    recommendedSkill: '',
    recommendedProductId: '',
    searchProblem: '',
    audience: '',
    scene: '',
    tone: '',
    brief: '',
    directionHistory: [],
    ...(snapshot.ideaSession || {}),
  })
  aiWorkStage.value = snapshot.aiWorkStage || 'idle'
  aiIntent.value = snapshot.aiIntent || 'unknown'
  Object.assign(decisionCard, {
    brandName: '',
    productName: '',
    skillName: '',
    contentType: '',
    recommendedDirection: '',
    reason: '',
    nextStepLabel: '',
    stageLabel: stageLabels.idle,
    ...(snapshot.decisionCard || {}),
  })
  previewMode.value = snapshot.previewMode || 'mobile'
  draftSourceExpanded.value = Boolean(snapshot.draftSourceExpanded)
  queueWorkbenchChatScrollAfterRender(true)
  window.setTimeout(() => {
    restoringWorkbenchContext = false
  }, 0)
  return true
}

function resetWorkbenchSession() {
  if (chatStreaming.value || loading.action) return
  clearWorkbenchDraftState({ clearSelection: true })
  showToast('已新建母稿，可以重新开始')
}

function clearWorkbenchDraftState(options = {}) {
  chatMessages.splice(0, chatMessages.length)
  workbench.prompt = ''
  Object.assign(editingDraft, {
    id: 0,
    title: '',
    summary: '',
    body: '',
    keywordsText: '',
    status: '草稿',
    source: '人工创作',
    sourceSnapshot: {},
  })
  clearEditingDraftSavedSnapshot()
  Object.assign(ideaSession, {
    stage: 'collecting',
    intent: '等待想法',
    recommendedSkill: '',
    recommendedProductId: '',
    searchProblem: '',
    audience: '',
    scene: '',
    tone: '',
    brief: '',
    directionHistory: [],
  })
  aiIntent.value = 'unknown'
  setAiStage('idle')
  Object.assign(decisionCard, {
    brandName: '',
    productName: '',
    skillName: '',
    contentType: '',
    recommendedDirection: '',
    reason: '',
    nextStepLabel: '',
    stageLabel: stageLabels.idle,
  })
  draftSourceExpanded.value = false
  previewMode.value = 'mobile'
  clearWorkbenchSessionCache()
  window.localStorage?.removeItem(WORKBENCH_DRAFT_STORAGE_KEY)
  window.localStorage?.setItem(WORKBENCH_RESET_STORAGE_KEY, '1')
  if (activeMenu.value === 'workbench') {
    router.replace({ path: menuRouteMap.workbench }).catch(() => {})
  }
  if (options.clearSelection) {
    workbench.brandId = ''
    workbench.productId = ''
    workbench.contentType = DEFAULT_CONTENT_TYPE
    workbench.skill = INTERNAL_DRAFT_SKILL
    workbench.hotspot = null
  }
}

function latestRecoverableDraft() {
  return [...drafts]
    .filter(draft => ['draft', 'approved', '草稿', '已通过'].includes(draft.rawStatus || draft.status))
    .sort((a, b) => String(b.generatedAt || '').localeCompare(String(a.generatedAt || '')))
    .find(draft => draft.conversation?.length || draft.body || draft.title) || null
}

function restoreWorkbenchDraftAfterLoad() {
  if (activeMenu.value !== 'workbench' || hasWorkbenchDraftState()) return
  if (restoreWorkbenchLocalSession()) return
  const routeDraft = findDraftById(route.query.draft_id)
  const cachedDraft = findDraftById(window.localStorage?.getItem(WORKBENCH_DRAFT_STORAGE_KEY))
  if (window.localStorage?.getItem(WORKBENCH_RESET_STORAGE_KEY) && !routeDraft && !cachedDraft) return
  const target = routeDraft || cachedDraft || latestRecoverableDraft()
  if (!target) return
  restoreDraftToWorkbench(target, { navigate: false })
}

function restoreChannelGenerationAfterLoad() {
  if (activeMenu.value !== 'channelGeneration') return
  const routeDraft = findDraftById(route.query.draft_id)
  const target = routeDraft || selectedDraft.value || drafts.find(canGenerateChannelFromDraft)
  if (!target?.id) return
  selectedDraft.value = target
  setupChannelGenerationNodes(target)
}

function channelContentFromApi(content) {
  const channelId = apiField(content, 'channel_id', 'ChannelID')
  const channel = channelProfiles.find(item => Number(item.id) === Number(channelId))
  const body = apiField(content, 'body', 'Body') || ''
  const pkg = parseChannelPackage(body, channel?.name || `渠道 ${channelId}`)
  return {
    id: apiField(content, 'id', 'ID'),
    draftId: apiField(content, 'draft_id', 'DraftID'),
    channel: channel?.name || `渠道 ${channelId}`,
    channelId,
    title: apiField(content, 'title', 'Title') || pkg.contentPayload.title || '',
    body: pkg.contentPayload.body || body,
    package: pkg,
    tags: arrayOrObjectText(pkg.contentPayload.tags || pkg.contentPayload.hashtags || pkg.geoPayload.keywords || []),
    seoTitle: '',
    script: '',
    status: channelContentStatusLabel(apiField(content, 'audit_status', 'AuditStatus')),
    rawAuditStatus: 'approved',
    publishStatus: apiField(content, 'publish_status', 'PublishStatus'),
  }
}

function parseChannelPackage(body, channelName = '') {
  const fallback = {
    channel: channelName,
    status: 'generated',
    contentType: channelContentTypeForName(channelName),
    contentPayload: { title: '', body: String(body || ''), tags: [], extraFields: {} },
    assetPayload: { requiredAssets: [], matchedAssets: [], missingAssets: [], aiGenerateSuggestions: [] },
    geoPayload: { keywords: [], geoSuggestions: [] },
    riskNotes: [],
    nextAction: '待编辑',
  }
  const parsed = parseJsonObject(body)
  if (!Object.keys(parsed).length) return fallback
  const contentPayload = parseJsonObject(parsed.contentPayload || parsed.content_payload)
  const assetPayload = parseJsonObject(parsed.assetPayload || parsed.asset_payload)
  const geoPayload = parseJsonObject(parsed.geoPayload || parsed.geo_payload)
  return {
    channel: parsed.channel || channelName,
    status: parsed.status || 'generated',
    contentType: parsed.contentType || parsed.content_type || channelContentTypeForName(channelName),
    contentPayload: {
      title: firstText(contentPayload, ['title', 'question_title', 'answer_title', 'video_title', 'seo_title', 'page_title', 'post_text']),
      summary: firstText(contentPayload, ['summary', 'meta_description', 'intro']),
      body: firstText(contentPayload, ['body', 'answer_body', 'script', 'post_text']) || String(body || ''),
      tags: contentPayload.tags || contentPayload.hashtags || contentPayload.keywords || [],
      extraFields: contentPayload.extraFields || contentPayload.extra_fields || Object.fromEntries(Object.entries(contentPayload).filter(([key]) => !['title', 'question_title', 'answer_title', 'video_title', 'seo_title', 'page_title', 'summary', 'meta_description', 'intro', 'body', 'answer_body', 'script', 'post_text', 'tags', 'hashtags', 'keywords'].includes(key))),
      ...contentPayload,
    },
    assetPayload: {
      requiredAssets: assetPayload.requiredAssets || assetPayload.required_assets || [],
      matchedAssets: assetPayload.matchedAssets || assetPayload.matched_assets || [],
      missingAssets: assetPayload.missingAssets || assetPayload.missing_assets || [],
      aiGenerateSuggestions: assetPayload.aiGenerateSuggestions || assetPayload.ai_generate_suggestions || [],
    },
    geoPayload: {
      brandEntityIncluded: Boolean(geoPayload.brandEntityIncluded ?? geoPayload.brand_entity_included),
      productEntityIncluded: Boolean(geoPayload.productEntityIncluded ?? geoPayload.product_entity_included),
      keywords: geoPayload.keywords || [],
      geoSuggestions: geoPayload.geoSuggestions || geoPayload.geo_suggestions || [],
    },
    riskNotes: parsed.riskNotes || parsed.risk_notes || [],
    nextAction: parsed.nextAction || parsed.next_action || '待编辑',
  }
}

function channelNodeFromContent(content) {
  const item = channelContentFromApi(content)
  return {
    channelName: item.channel,
    channelId: item.channelId,
    contentId: item.id,
    contentType: item.package.contentType,
    status: item.package.assetPayload.missingAssets?.length ? 'need_assets' : 'completed',
    startedAt: '',
    completedAt: apiField(content, 'updated_at', 'UpdatedAt') || new Date().toISOString(),
    updatedLabel: formatGeneratedTime(apiField(content, 'updated_at', 'UpdatedAt') || new Date().toISOString()),
    contentPayload: item.package.contentPayload,
    assetPayload: item.package.assetPayload,
    geoPayload: item.package.geoPayload,
    riskNotes: item.package.riskNotes,
    raw: content,
  }
}

function serializeChannelNode(node) {
  return JSON.stringify({
    channel: node.channelName,
    status: 'generated',
    contentType: node.contentType,
    contentPayload: node.contentPayload,
    assetPayload: node.assetPayload,
    geoPayload: node.geoPayload,
    riskNotes: node.riskNotes,
    nextAction: node.status === 'need_assets' ? '待补充素材' : '待编辑',
  }, null, 2)
}

function channelContentTypeForName(name) {
  return {
    小红书: '图文笔记',
    知乎: '问答回答',
    微信公众号: '图文文章',
    抖音: '视频脚本',
    微博: '短帖',
    百家号: '图文文章',
    独立站: 'SEO文章',
  }[name] || '渠道内容'
}

function firstText(obj, keys) {
  for (const key of keys) {
    const value = obj?.[key]
    if (typeof value === 'string' && value.trim()) return value.trim()
  }
  return ''
}

function arrayOrObjectText(value) {
  if (Array.isArray(value)) return value.map(item => arrayOrObjectText(item)).filter(Boolean).join('、')
  if (value && typeof value === 'object') return Object.entries(value).map(([key, item]) => `${channelFieldLabel(key)}：${arrayOrObjectText(item)}`).join('；')
  return String(value || '').trim()
}

function normalizeFieldKey(key) {
  return String(key || '')
    .replace(/([a-z0-9])([A-Z])/g, '$1_$2')
    .replace(/[-\s]+/g, '_')
    .replace(/^_+|_+$/g, '')
    .toLowerCase()
}

function channelFieldLabel(key) {
  const labels = {
    cover_title: '封面标题',
    cover_style: '封面风格',
    image_count: '图片数量',
    image_suggestions: '图片建议',
    note_type: '笔记类型',
    content_type: '内容类型',
    platform_notes: '平台说明',
    topic_tags: '话题标签',
    material_notes: '素材说明',
    publish_notes: '发布备注',
    question_title: '问题标题',
    answer_title: '回答标题',
    title: '标题',
    body: '正文',
    summary: '摘要',
    caption: '发布文案',
    cta: '行动引导',
    key_points: '核心观点',
    argument_structure: '论证结构',
    product_mention_strategy: '商品露出方式',
    section_titles: '小标题结构',
    intro: '引导语',
    ending_cta: '结尾 CTA',
    original_link_suggestion: '原文链接建议',
    hook: '开头钩子',
    storyboard: '分镜脚本',
    subtitles: '字幕文案',
    shooting_notes: '拍摄建议',
    publish_caption: '发布文案',
    image_plan: '配图需求',
    seo_title: 'SEO 标题',
    meta_description: 'Meta 描述',
    url_slug: 'URL Slug',
    faq: 'FAQ',
    internal_links: '内链建议',
    schema_json: 'Schema 结构化数据',
    tags: '话题标签',
    hashtags: '话题标签',
    keywords: '关键词',
  }
  const normalized = normalizeFieldKey(key)
  return labels[normalized] || labels[key] || key
}

function isLongChannelMetaField(field) {
  const value = String(field?.value || '')
  return value.length > 44 || ['图片建议', '素材需求', '核心观点', '小标题结构', 'FAQ', '平台说明', '素材说明'].includes(field?.label)
}

function apiPlanToPlan(plan) {
  const scheduledAt = apiField(plan, 'scheduled_at', 'ScheduledAt')
  const channelId = apiField(plan, 'channel_id', 'ChannelID')
  const planStatus = apiField(plan, 'status', 'Status')
  const scheduled = new Date(scheduledAt)
  const channelContentId = apiField(plan, 'channel_content_id', 'ChannelContentID')
  const channelContent = findChannelContentById(channelContentId)
  const account = channelAccounts.find(item => Number(item.channelId) === Number(channelId))
  const channelName = channelProfiles.find(c => Number(c.id) === Number(channelId))?.name || `渠道 ${channelId}`
  return {
    id: apiField(plan, 'id', 'ID'),
    date: Number.isNaN(scheduled.getTime()) ? String(scheduledAt || '').slice(0, 10) : scheduled.toISOString().slice(0, 10),
    time: Number.isNaN(scheduled.getTime()) ? '' : scheduled.toTimeString().slice(0, 5),
    planCode: apiField(plan, 'plan_code', 'PlanCode') || '',
    title: publishPlanContentTitle(channelContent, channelName),
    summary: channelContent?.package?.contentPayload?.summary || '',
    body: channelContent?.body || '',
    tags: channelContent?.tags || '',
    channel: channelName,
    channelId,
    channelContentId,
    accountName: account?.accountName || '未绑定账号',
    method: apiField(plan, 'publish_method', 'PublishMethod'),
    level: apiField(plan, 'automation_level', 'AutomationLevel'),
    skill: '',
    risk: '发布前校验 + 异常人工接管',
    owner: '系统',
    rawStatus: planStatus,
    status: publishStatusLabel(planStatus),
    link: apiField(plan, 'published_url', 'PublishedURL') || '',
    flowStatus: '直接流转',
    accountStatus: planAccountStatusLabel(account),
    materialStatus: planMaterialStatusLabel(channelContent),
    publishStatus: publishStatusLabel(planStatus),
  }
}

function publishPlanContentTitle(channelContent, channelName = '渠道') {
  const title = String(channelContent?.title || '').trim()
  if (title) return title
  if (channelContent) return `${channelName}内容未命名`
  return '未生成渠道内容'
}

function hydratePlans(apiPlans) {
  plans.splice(0, plans.length, ...dedupePlans(apiPlans.map(apiPlanToPlan)))
}

function dedupePlans(planItems) {
  const statusRank = {
    published: 5,
    publishing: 4,
    scheduled: 3,
    failed: 2,
    cancelled: 1,
  }
  const byContentChannel = new Map()
  planItems.forEach(plan => {
    const contentId = Number(plan.channelContentId || 0)
    const channelId = Number(plan.channelId || 0)
    const key = contentId && channelId ? `${contentId}:${channelId}` : `id:${plan.id}`
    const current = byContentChannel.get(key)
    if (!current) {
      byContentChannel.set(key, plan)
      return
    }
    const currentRank = statusRank[current.rawStatus] || 0
    const nextRank = statusRank[plan.rawStatus] || 0
    if (nextRank > currentRank || (nextRank === currentRank && Number(plan.id || 0) > Number(current.id || 0))) {
      byContentChannel.set(key, plan)
    }
  })
  return Array.from(byContentChannel.values())
}

function findChannelContentById(contentId) {
  const id = Number(contentId || 0)
  if (!id) return null
  for (const draft of drafts) {
    const content = draft.channels?.find(item => Number(item.id) === id)
    if (content) return content
  }
  return channelContentIndex.find(item => Number(item.id) === id) || null
}

function planAccountStatusLabel(account) {
  if (!account) return '待授权'
  return account.status === '已授权' ? '可发布' : '阻断'
}

function planMaterialStatusLabel(content) {
  if (!content) return '缺渠道内容'
  return String(content.body || content.title || '').trim() ? '完整' : '缺正文'
}

function showToast(text) {
  const rawMessage = String(text || '').trim()
  if (/^资源不存在(?:（HTTP\s*404）)?$/.test(rawMessage)) return
  const isNetworkFailure = /(load|fetch|network).{0,20}(fail|failed|failure|error)|(fail|failed|failure|error).{0,20}(load|fetch|network)|networkerror|aborterror/i.test(rawMessage)
  const message = isNetworkFailure
    ? 'AI 专家服务连接失败，请检查网络或稍后重试'
    : rawMessage
  if (!message) return
  const type = /失败|错误|异常|无法|阻断|failed|failure|error|network|abort|fetch|load/i.test(message)
    ? 'error'
    : /请先|请填写|未找到|不存在|不能|需要|告警|风险/.test(message) ? 'warning' : 'success'
  ElMessage({ message, type, grouping: true, showClose: true })
}
function closeModal() {
  modal.type = ''
  modal.title = ''
  modal.wide = false
}
function closeDrawer() { drawer.type = ''; drawer.title = '' }
function openPublishPlanDrawer(plan) {
  Object.keys(selectedPlan).forEach(key => delete selectedPlan[key])
  Object.assign(selectedPlan, plan || {})
  drawer.type = 'publishPlan'
  drawer.title = '发布计划详情'
}
function isDraftExpanded(draft) {
  return Number(expandedDraftId.value) === Number(draft?.id)
}
function toggleDraftChannels(draft) {
  const id = Number(draft?.id || 0)
  if (!id) return
  expandedDraftId.value = isDraftExpanded(draft) ? 0 : id
}
function goChannelManagement() { activeMenu.value = 'channels' }
function channelStatusValue(channel) {
  if (!channel) return 'active'
  if (channel.enabled || channel.status === '可发布') return 'active'
  if (channel.status === '停用') return 'disabled'
  return 'pending'
}
function openChannelProfileModal(channel = null) {
  const isEdit = Boolean(channel?.id)
  Object.assign(channelForm, {
    id: isEdit ? Number(channel.id) : 0,
    channel_code: isEdit ? String(channel.code || '') : '',
    channel_name: isEdit ? String(channel.name || '') : '',
    channel_type: isEdit ? String(channel.type || 'content') : 'content',
    entry_url: isEdit ? String(channel.siteUrl || '') : '',
    content_forms: isEdit ? String(channel.contentTypes || '').replace(/\s*\/\s*/g, '、') : '',
    support_modes: isEdit ? String(channel.supportMethods || '').replace(/\s*\/\s*/g, '、') : '',
    default_publish_mode: isEdit ? String(channel.method || 'manual') : 'manual',
    status: isEdit ? channelStatusValue(channel) : 'active',
  })
  modal.type = 'channelProfile'
  modal.title = isEdit ? `配置渠道：${channel.name}` : '新增渠道'
  modal.wide = false
}
function configureChannel(channel) { openChannelProfileModal(channel) }
async function refreshChannels() {
  const [channelPage, channelAccountPage] = await Promise.all([
    fetchAiGeoChannels({ limit: 200 }),
    fetchAiGeoChannelAccounts({ limit: 500 }),
  ])
  hydrateChannels(channelPage.items || [])
  hydrateChannelAccounts(channelAccountPage.items || [])
}
async function testChannel(channel) {
  if (!channel?.id) {
    showToast('请先保存渠道资料后再测试')
    return
  }
  loading.action = true
  try {
    const result = await testAiGeoChannel(Number(channel.id))
    showToast(String(result?.message || `${channel.name} 测试完成`))
  } catch (error) {
    showToast(error?.message || '渠道测试失败')
  } finally {
    loading.action = false
  }
}
function addChannel() { openChannelProfileModal() }
function openPendingTask(task) {
  activeMenu.value = task.menu
}
function saveModeConfig() { closeModal(); showToast('模式配置已保存') }
function openImportModal() {
  importResult.value = null
  importErrors.splice(0, importErrors.length)
  modal.type = 'import'
  modal.title = '导入资料'
  modal.wide = true
}
function openBrandModal(brand = null) {
  const isEdit = Boolean(brand?.id)
  Object.assign(brandForm, {
    id: isEdit ? brand.id : 0,
    brand_code: isEdit ? String(brand.code || brand.brand_code || brand.name || '') : '',
    brand_name: isEdit ? String(brand.name || brand.brand_name || '') : '',
    positioning: isEdit ? String(brand.position || brand.positioning || '') : '',
    target_audience: isEdit ? String(brand.audience || brand.target_audience || '') : '',
    price_band: isEdit ? String(brand.priceBand || brand.price_band || '') : '',
    keywordsText: isEdit ? (brand.keywordGroups || []).flatMap(group => group.keywords || []).join('、') : '',
  })
  modal.type = 'brand'
  modal.title = isEdit ? '编辑品牌资料' : '新增品牌资料'
  modal.wide = false
}
function openHotspotModal(hotspot = null) {
  const isEdit = Boolean(hotspot?.id)
  const metadata = isEdit ? (hotspot.metadata || {}) : {}
  Object.assign(hotspotForm, {
    id: isEdit ? Number(hotspot.id) : 0,
    article_date: isEdit ? String(hotspot.articleDate || metadata.article_date || dateOnly(hotspot.capturedAt)) : dateOnly(new Date()),
    hotspot_topic: isEdit ? String(hotspot.hotspotTopic || metadata.hotspot_topic || '') : '',
    platform: isEdit ? String(hotspot.platform || '') : '手动',
    title: isEdit ? String(hotspot.title || '') : '',
    article_type: isEdit ? String(hotspot.articleType || metadata.article_type || '行业趋势') : '行业趋势',
    author: isEdit ? String(hotspot.author || metadata.author || '') : '',
    summary: isEdit ? String(hotspot.summary || metadata.summary || '') : '',
    core_viewpoint: isEdit ? String(hotspot.coreViewpoint || metadata.core_viewpoint || '') : '',
    scene_tags: isEdit ? [...(hotspot.sceneTags || metadata.scene_tags || [])] : [],
    brand_id: isEdit ? Number(hotspot.brandId || metadata.brand_id || 0) : Number(currentBrand.value?.id || 0),
    product_id: isEdit ? Number(hotspot.productId || metadata.product_id || 0) : 0,
    sku_text: isEdit ? String(hotspot.skuText || metadata.sku_text || '') : '',
    competitor_text: isEdit ? String(hotspot.competitorText || metadata.competitor_text || '') : '',
    keywords_text: isEdit ? String(hotspot.keywordsText || (Array.isArray(metadata.keywords) ? metadata.keywords.join('、') : metadata.keywords || '')) : '',
    heat_score: isEdit ? Number(hotspot.heat || hotspot.heat_score || 50) : 50,
    priority: isEdit ? String(hotspot.priority || metadata.priority || 'P1') : 'P1',
    content_status: isEdit ? String(hotspot.contentStatus || metadata.content_status || '待确认') : '待确认',
    owner: isEdit ? String(hotspot.owner || metadata.owner || '') : '',
    source_url: isEdit ? String(hotspot.sourceUrl || hotspot.source_url || hotspot.summary || '') : '',
    captured_at: isEdit && hotspot.capturedAt ? toLocalDateTime(hotspot.capturedAt) : toLocalDateTime(new Date().toISOString()),
  })
  modal.type = 'hotspotForm'
  modal.title = isEdit ? '编辑热点文章' : '新建热点文章'
  modal.wide = false
}
function openNewPlanModal() { modal.type = 'newPlan'; modal.title = '新建发布计划' }
function defaultChannelPlanTime() {
  const baseDate = planDate.value || dateOnly(new Date())
  return toLocalDateTime(new Date(`${baseDate}T18:00:00`))
}
function channelPlanTimeLabel(value) {
  return String(value || '').replace('T', ' ') || '-'
}
function isChannelPlanSelected(name) {
  return channelPlanForm.selectedChannels.includes(name)
}
function toggleChannelPlanSelection(name) {
  const index = channelPlanForm.selectedChannels.indexOf(name)
  if (index >= 0) channelPlanForm.selectedChannels.splice(index, 1)
  else channelPlanForm.selectedChannels.push(name)
}
function openActiveChannelPlanModal() {
  const node = activeChannelNode.value?.contentId ? activeChannelNode.value : generatedChannelPlanOptions.value[0]
  if (!node?.contentId) {
    showToast('请先生成当前平台渠道内容')
    return
  }
  channelPlanForm.scheduledAt = defaultChannelPlanTime()
  channelPlanForm.selectedChannels.splice(0, channelPlanForm.selectedChannels.length, ...generatedChannelPlanOptions.value.map(item => item.channelName))
  modal.type = 'activeChannelPlan'
  modal.title = '添加发布计划'
  modal.wide = false
}
function viewDraft(draft) {
  selectedDraft.value = draft
  modal.type = 'draftView'
  modal.title = '查看母稿'
  modal.wide = true
}
function openDraftSubmitSuccess(draft) {
  submittedDraftForChannel.value = draft ? { ...draft } : null
  modal.type = 'draftSubmitSuccess'
  modal.title = '保存成功'
  modal.wide = false
}
function openDraftSubmitConfirm() {
  if (!hasEditingDraftContent.value) {
    showToast('请先生成或填写母稿内容')
    return
  }
  confirmStandardAction({
    title: '保存母稿',
    icon: '!',
    message: '确认保存当前母稿？',
    detail: '保存后可直接进入渠道内容生成页面。',
    confirmText: '确认保存',
  })
    .then(() => submitDraftFlow())
    .catch(() => {})
}
async function openDraftGenerateConfirm() {
  if (chatStreaming.value || loading.action) return
  const guard = draftGenerationGuard.value
  if (!guard.requiresConfirm) {
    generateDraftFromChat({ skipConfirm: true })
    return
  }
  const issueCount = guard.items?.filter(item => item.level === 'warning').length || 0
  confirmStandardAction({
    title: '生成前确认',
    icon: '!',
    message: issueCount ? `发现 ${issueCount} 个生成风险，确认仍要生成？` : '信息较完整，确认生成母稿？',
    detail: draftGenerationConfirmDetail(guard),
    confirmText: '确认生成',
    cancelText: '继续补充',
    size: 'medium',
  })
    .then(() => generateDraftFromChat())
    .catch(() => {})
}

async function generateFromDecision() {
  if (chatStreaming.value || loading.action) return
  await generateDraftFromChat({ promptOverride: decisionCard.recommendedDirection, appendPrompt: false })
}

async function retryLastWorkbenchMessage() {
  const prompt = latestWorkbenchUserPrompt.value
  if (!prompt || chatStreaming.value || loading.action) return
  await streamGatewayMentorReply(prompt, { route: 'retry', next_action: 'none' })
}

async function copyReplyText(message) {
  const text = String(message?.text || '').trim()
  if (!text) return
  try {
    await navigator.clipboard.writeText(text)
    showToast('已复制 AI 回复')
  } catch (error) {
    showToast('复制失败，请手动选择文本复制')
  }
}

function extractTitleCandidateFromAiReply(text) {
  const value = normalizeGeneratedDraftText(text)
  if (!value) return ''
  const patterns = [
    /示例标题[:：]\s*[「《“"]?([^\n。；;]+?)[」》”"]?(?=\s*(?:[-－]\s*)?(?:适用情况|改法|方案|我的建议|$))/,
    /标题(?:变为|改为|改成|为|是)[:：]?\s*[「《“"]?([^\n。；;]+?)[」》”"]?(?=\s*(?:[-－]\s*)?(?:适用情况|改法|方案|我的建议|$))/,
    /(?:完整标题|最终标题)[:：]\s*[「《“"]?([^\n。；;]+?)[」》”"]?(?=\s*(?:[-－]\s*)?(?:适用情况|改法|方案|我的建议|$))/,
  ]
  for (const pattern of patterns) {
    const matched = value.match(pattern)?.[1]
    const title = directionTitleOnly(matched || '')
    if (title && title.length >= 6) return title
  }
  return ''
}

function applyTitleFromReply(text) {
  const title = extractTitleCandidateFromAiReply(text)
  if (!title) return false
  editingDraft.title = title
  editingDraft.status = '草稿'
  updateDecisionCard({
    recommendedDirection: title,
    stageLabel: stageLabels.draft_editing,
    nextStepLabel: '',
    reason: '已按 AI 回复应用标题，正文未变更。',
  })
  setAiStage('draft_editing')
  showToast('已应用标题')
  return true
}

function generatePromptFromReplyAction(action, message) {
  const explicit = String(action?.prompt || action?.payload?.prompt || '').trim()
  if (explicit) return explicit
  const text = String(message?.text || '').trim()
  if (text) {
    return [
      '请严格基于下面这条 AI 回复里已经确认的方向生成右侧母稿，不要沿用旧方向：',
      text,
    ].join('\n')
  }
  return latestWorkbenchUserPrompt.value || decisionCard.recommendedDirection
}

async function handleReplyAction(action, message) {
  const type = String(action?.type || '')
  if (!WORKBENCH_REPLY_ACTION_TYPES.has(type) || chatStreaming.value || loading.action) return
  if (type === 'copy_reply') {
    await copyReplyText(message)
    return
  }
  if (type === 'retry') {
    await retryLastWorkbenchMessage()
    return
  }
  if (type === 'recommend_other_directions') {
    await recommendAnotherDirection()
    return
  }
  if (type === 'apply_title') {
    if (!applyTitleFromReply(String(action.content || action.payload?.content || message?.text || ''))) {
      showToast('AI 回复里没有可应用的标题')
    }
    return
  }
  if (type === 'generate_draft' || type === 'regenerate_draft') {
    await generateDraftFromChat({
      promptOverride: generatePromptFromReplyAction(action, message),
      appendPrompt: false,
      skipConfirm: true,
    })
    return
  }
  if (type === 'apply_to_draft') {
    const applyContent = replyActionApplyContent(action)
    if (!applyContent) {
      showToast('这条回复只是修改建议，不能直接应用到母稿')
      return
    }
    if (applyTitleFromReply(applyContent)) return
    const applied = applyCoCreatedDraftIfPresent(applyContent, latestWorkbenchUserPrompt.value, message?.idea || {})
    if (!applied) showToast('AI 回复里没有可应用到母稿的完整内容')
  }
}

async function recommendAnotherDirection() {
  if (chatStreaming.value) return
  await streamExpertCoCreationReply('请基于当前上下文重新给我 3 个其他且不重复的 GEO 母稿方向。每个方向必须使用同样结构：方向定位、母稿标题、内容逻辑、GEO价值；标题必须和方向定位一致，不要把核心钩子和标题混在同一行。', { route: 'direction_consultation', next_action: 'recommend_direction' })
}

async function diagnoseCurrentDirection() {
  if (chatStreaming.value) return
  await streamExpertCoCreationReply(`请诊断当前方向「${decisionCard.recommendedDirection || latestWorkbenchUserPrompt.value}」哪里弱，并给出更专家的推进建议。`, { route: 'direction_consultation', next_action: 'recommend_direction' })
}

function focusCustomDirection() {
  workbench.prompt = decisionCard.recommendedDirection ? `我想改成：${decisionCard.recommendedDirection}` : '我想改成：'
}
function goGenerateChannelsAfterSubmit() {
  const target = submittedDraftForChannel.value || selectedDraft.value || editingDraft
  closeModal()
  openChannelGenerationConsole(target)
}
function normalizeBrandCode(name) {
  const raw = String(name || '').trim().toLowerCase()
  const ascii = raw
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-+|-+$/g, '')
  return ascii || `brand-${Date.now()}`
}
function splitKeywords(value) {
  return String(value || '').split(/[\n,，、;；]+/).map(item => item.trim()).filter(Boolean)
}
function toLocalDateTime(value) {
  const date = value instanceof Date ? value : new Date(value || Date.now())
  if (Number.isNaN(date.getTime())) return ''
  const offset = date.getTimezoneOffset() * 60000
  return new Date(date.getTime() - offset).toISOString().slice(0, 16)
}
function dateOnly(value) {
  const date = value instanceof Date ? value : new Date(value || Date.now())
  if (Number.isNaN(date.getTime())) return ''
  const offset = date.getTimezoneOffset() * 60000
  return new Date(date.getTime() - offset).toISOString().slice(0, 10)
}
function fromLocalDateTime(value) {
  if (!value) return new Date().toISOString()
  const date = new Date(value)
  return Number.isNaN(date.getTime()) ? new Date().toISOString() : date.toISOString()
}
function selectImportType(type) {
  importForm.type = type
  importResult.value = null
  selectedImportFileName.value = ''
  if (importFileInput.value) importFileInput.value.value = ''
  importErrors.splice(0, importErrors.length)
}
function fillImportExample() {
  importForm.rawText = currentImportOption.value.placeholder
  selectedImportFileName.value = ''
}
function downloadImportTemplate() {
  const headers = currentImportFields.value.map(field => field.label)
  const sample = currentImportFields.value.map(field => field.sample || '')
  const workbook = XLSX.utils.book_new()
  const sheet = XLSX.utils.aoa_to_sheet([headers, sample])
  sheet['!cols'] = currentImportFields.value.map(field => ({ wch: Math.max(12, field.label.length * 2, String(field.sample || '').length + 2) }))
  XLSX.utils.book_append_sheet(workbook, sheet, '导入模板')
  XLSX.writeFile(workbook, `AI-GEO-${currentImportOption.value.label}.xlsx`)
}
async function handleImportFileChange(event) {
  const file = event.target?.files?.[0]
  if (!file) return
  selectedImportFileName.value = file.name
  try {
    if (/\.(csv|tsv)$/i.test(file.name)) {
      importForm.rawText = await file.text()
    } else {
      const buffer = await file.arrayBuffer()
      const workbook = XLSX.read(buffer, { type: 'array' })
      const sheetName = workbook.SheetNames[0]
      if (!sheetName) {
        showToast('Excel 文件里没有可读取的工作表')
        return
      }
      const rows = XLSX.utils.sheet_to_json(workbook.Sheets[sheetName], { header: 1, defval: '' })
      importForm.rawText = aoaToCsv(rows)
    }
    importResult.value = null
    importErrors.splice(0, importErrors.length)
    showToast(`已读取文件：${file.name}`)
  } catch (error) {
    showToast(error?.message || '文件读取失败，请确认是 Excel / CSV 模板文件')
  } finally {
    if (importFileInput.value) importFileInput.value.value = ''
  }
}
function aoaToCsv(rows) {
  return rows
    .filter(row => Array.isArray(row) && row.some(cell => String(cell || '').trim()))
    .map(row => row.map(cell => csvCell(cell)).join(','))
    .join('\n')
}
function csvCell(value) {
  const text = String(value ?? '')
  return /[",\n\r]/.test(text) ? `"${text.replace(/"/g, '""')}"` : text
}
function parseImportText(rawText, fields) {
  const lines = String(rawText || '').split(/\r?\n/).map(line => line.trim()).filter(Boolean)
  if (lines.length < 2) return []
  const delimiter = lines[0].includes('\t') ? '\t' : ','
  const headerLabels = splitImportLine(lines[0], delimiter)
  const fieldByHeader = new Map()
  fields.forEach(field => {
    fieldByHeader.set(field.label, field.key)
    fieldByHeader.set(normalizeImportHeader(field.label), field.key)
    fieldByHeader.set(field.key, field.key)
  })
  return lines.slice(1).map(line => {
    const values = splitImportLine(line, delimiter)
    const row = {}
    headerLabels.forEach((header, index) => {
      const key = fieldByHeader.get(header) || fieldByHeader.get(normalizeImportHeader(header)) || header
      row[key] = values[index] ?? ''
    })
    return normalizeImportRow(importForm.type, row)
  }).filter(row => Object.values(row).some(value => String(value || '').trim()))
}
function normalizeImportHeader(value) {
  return String(value || '').replace(/\s+/g, '').toLowerCase()
}
function splitImportLine(line, delimiter) {
  if (delimiter === '\t') return line.split('\t').map(item => item.trim())
  const cells = []
  let current = ''
  let quoted = false
  for (const char of line) {
    if (char === '"') {
      quoted = !quoted
    } else if (char === ',' && !quoted) {
      cells.push(current.trim())
      current = ''
    } else {
      current += char
    }
  }
  cells.push(current.trim())
  return cells
}
function normalizeImportRow(type, row) {
  const next = { ...row }
  if (type === 'brand') {
    next.keywords = splitKeywords(next.keywords)
  }
  if (type === 'product') {
    const selling = splitKeywords(next.selling_points)
    const angles = [
      ...splitKeywords(next.content_angles),
      ...splitKeywords(next.series),
      ...splitKeywords(next.target_audience),
    ]
    next.selling_points = selling
    next.faq = splitKeywords(next.faq)
    next.content_angles = [...new Set(angles)]
  }
  if (type === 'sku') {
    const attrs = {}
    ;['color', 'size', 'fabric', 'fit', 'style_tags', 'scene_tags', 'model_image_url'].forEach(key => {
      if (next[key]) attrs[key] = key.endsWith('_tags') ? splitKeywords(next[key]) : next[key]
    })
    next.attributes = attrs
    next.image_url = next.product_image_url || next.image_url || ''
    next.price = Number(next.price || 0)
    next.sku_name = next.sku_name || [next.color, next.size].filter(Boolean).join(' ') || next.sku_code
  }
  if (type === 'keyword') {
    next.keywords = splitKeywords(next.keywords)
    next.weight = Number(next.weight || 0)
    next.source = next.source || 'import'
  }
  return next
}
async function submitMaterialImport() {
  const rows = importPreviewRows.value
  if (!rows.length) {
    showToast('请先粘贴包含表头和数据的 CSV / 表格内容')
    return
  }
  loading.action = true
  importErrors.splice(0, importErrors.length)
  try {
    if (importForm.type === 'keyword') {
      const result = await submitKeywordImport(rows)
      importResult.value = result
      await loadAiGeoData()
      showToast(importResultMessage(result))
      return
    }
    const result = await importAiGeoMaterials({
      import_type: importForm.type,
      mapping_config: Object.fromEntries(currentImportFields.value.map(field => [field.label, field.key])),
      records: rows,
    })
    importResult.value = result
    await loadAiGeoData()
    showToast(importResultMessage(result))
  } catch (error) {
    showToast(error?.message || '资料导入失败')
  } finally {
    loading.action = false
  }
}
function importResultMessage(result) {
  return `导入完成：成功 ${Number(apiField(result, 'success_count', 'SuccessCount') || 0)} 条，失败 ${Number(apiField(result, 'failed_count', 'FailedCount') || 0)} 条`
}
async function submitKeywordImport(rows) {
  let success = 0
  const failures = []
  for (let index = 0; index < rows.length; index += 1) {
    const row = rows[index]
    const brand = findBrandByCode(row.brand_code)
    const product = row.product_code ? findProductByCode(row.product_code) : null
    if (!brand || (row.product_code && !product)) {
      failures.push({ id: `keyword-${index}`, row_number: index + 1, error_message: !brand ? '品牌编码不存在，请先导入品牌' : '商品编码不存在，请先导入商品' })
      continue
    }
    const keywords = Array.isArray(row.keywords) ? row.keywords : splitKeywords(row.keywords)
    if (!keywords.length || !row.keyword_group) {
      failures.push({ id: `keyword-${index}`, row_number: index + 1, error_message: '关键词分组和关键词不能为空' })
      continue
    }
    for (const keyword of keywords) {
      await createAiGeoKeyword({
        brand_id: brand.id,
        product_id: product?.id || undefined,
        keyword_group: row.keyword_group,
        keyword,
        intent: row.intent || row.package_name || '',
        source: row.source || 'import',
        weight: Number(row.weight || 0),
        status: 'active',
      })
      success += 1
    }
  }
  importErrors.splice(0, importErrors.length, ...failures)
  return { id: 0, success_count: success, failed_count: failures.length }
}
function findBrandByCode(code) {
  const target = String(code || '').trim().toLowerCase()
  return brands.find(brand => String(brand.code || '').trim().toLowerCase() === target)
}
function findProductByCode(code) {
  const target = String(code || '').trim().toLowerCase()
  for (const brand of brands) {
    const product = (brand.products || []).find(item => String(item.code || '').trim().toLowerCase() === target)
    if (product) return product
  }
  return null
}
function skuLabel(sku) {
  return [sku.code, sku.color, sku.size].filter(Boolean).join(' / ')
}
async function loadImportErrors(batchID) {
  if (!batchID) return
  try {
    const data = await fetchAiGeoImportErrors(Number(batchID), { limit: 100 })
    importErrors.splice(0, importErrors.length, ...(data.items || []))
  } catch (error) {
    showToast(error?.message || '导入错误明细加载失败')
  }
}
async function saveChannelProfile() {
  const name = String(channelForm.channel_name || '').trim()
  const code = String(channelForm.channel_code || '').trim() || normalizeBrandCode(name)
  if (!name || !code) {
    showToast('请填写渠道名称和渠道编码')
    return
  }
  const payload = {
    channel_code: code,
    channel_name: name,
    channel_type: String(channelForm.channel_type || '').trim() || 'content',
    entry_url: String(channelForm.entry_url || '').trim(),
    content_forms: splitKeywords(channelForm.content_forms),
    support_modes: splitKeywords(channelForm.support_modes),
    default_publish_mode: String(channelForm.default_publish_mode || 'manual').trim(),
    status: String(channelForm.status || 'active').trim(),
  }
  loading.action = true
  try {
    if (channelForm.id) {
      await updateAiGeoChannel(Number(channelForm.id), payload)
    } else {
      await createAiGeoChannel(payload)
    }
    await refreshChannels()
    closeModal()
    showToast('渠道配置已保存')
  } catch (error) {
    showToast(error?.message || '渠道配置保存失败')
  } finally {
    loading.action = false
  }
}
async function saveBrand() {
  const name = String(brandForm.brand_name || '').trim()
  if (!name) {
    showToast('请填写品牌名称')
    return
  }
  const payload = {
    brand_code: brandForm.id ? brandForm.brand_code : (String(brandForm.brand_code || '').trim() || normalizeBrandCode(name)),
    brand_name: name,
    positioning: String(brandForm.positioning || '').trim(),
    target_audience: String(brandForm.target_audience || '').trim(),
    price_band: String(brandForm.price_band || '').trim(),
    keywords: splitKeywords(brandForm.keywordsText),
    status: 'active',
  }
  loading.action = true
  try {
    const saved = brandForm.id
      ? await updateAiGeoBrand(Number(brandForm.id), payload)
      : await createAiGeoBrand(payload)
    selectedBrandId.value = saved.id
    closeModal()
    await loadAiGeoData()
    showToast(brandForm.id ? '品牌资料已更新' : '品牌资料已创建')
  } catch (error) {
    showToast(error?.message || '品牌资料保存失败')
  } finally {
    loading.action = false
  }
}
async function saveHotspot() {
  const title = String(hotspotForm.title || '').trim()
  if (!canSaveHotspotForm.value) {
    showToast('请补齐热点文章必填信息')
    return
  }
  const capturedAt = hotspotForm.article_date
    ? new Date(`${hotspotForm.article_date}T00:00:00`).toISOString()
    : fromLocalDateTime(hotspotForm.captured_at)
  const metadata = {
    article_date: hotspotForm.article_date,
    hotspot_topic: String(hotspotForm.hotspot_topic || '').trim(),
    article_type: String(hotspotForm.article_type || '').trim(),
    author: String(hotspotForm.author || '').trim(),
    summary: String(hotspotForm.summary || '').trim(),
    core_viewpoint: String(hotspotForm.core_viewpoint || '').trim(),
    scene_tags: [...hotspotForm.scene_tags],
    brand_id: Number(hotspotForm.brand_id || 0) || undefined,
    brand_name: brands.find(brand => Number(brand.id) === Number(hotspotForm.brand_id))?.name || '',
    product_id: Number(hotspotForm.product_id || 0) || undefined,
    product_name: hotspotProductOptions.value.find(product => Number(product.id) === Number(hotspotForm.product_id))?.name || '',
    sku_text: String(hotspotForm.sku_text || '').trim(),
    competitor_text: String(hotspotForm.competitor_text || '').trim(),
    keywords: splitKeywords(hotspotForm.keywords_text),
    priority: String(hotspotForm.priority || 'P1').trim(),
    content_status: String(hotspotForm.content_status || '待确认').trim(),
    owner: String(hotspotForm.owner || '').trim(),
  }
  const payload = {
    platform: String(hotspotForm.platform || '手动').trim(),
    title,
    heat_score: Number(hotspotForm.heat_score || 0),
    source_url: String(hotspotForm.source_url || '').trim(),
    captured_at: capturedAt,
    metadata,
    status: 'active',
  }
  loading.action = true
  try {
    const saved = hotspotForm.id
      ? await updateAiGeoHotspot(Number(hotspotForm.id), payload)
      : await createAiGeoHotspot(payload)
    workbench.hotspot = {
      id: saved.id || saved.ID,
      title: saved.title || saved.Title || title,
      platform: saved.platform || saved.Platform || payload.platform,
      heat: saved.heat_score || saved.HeatScore || payload.heat_score,
      summary: saved.source_url || saved.SourceURL || payload.source_url,
      metadata,
    }
    closeModal()
    await loadAiGeoData()
    showToast(hotspotForm.id ? '热点文章已更新' : '热点文章已创建')
  } catch (error) {
    showToast(error?.message || '热点文章保存失败')
  } finally {
    loading.action = false
  }
}
function generateBrandKeywords() { showToast('已基于品牌、商品、竞品信息生成关键词') }
function openHotspotDrawer() { drawer.type = 'hotspot'; drawer.title = '引用热点' }
function openHotspotExtractDrawer(options = {}) {
  hotspotExtractReturnTo.value = options.returnTo || 'select'
  hotspotExtractResult.value = null
  hotspotExtractForm.platform = hotspotExtractForm.platform || ''
  hotspotExtractForm.contentType = hotspotExtractForm.contentType || ''
  drawer.type = 'hotspotExtract'
  drawer.title = '提炼引用热点'
}
function openStyleManageDrawer() {
  drawer.type = 'styleManage'
  drawer.title = '风格管理'
}
function openStyleDetailDrawer(style) {
  selectedStyleDetail.value = style
  drawer.type = 'styleDetail'
  drawer.title = '风格详情'
}
function openStyleExtractDrawer(options = {}) {
  styleExtractReturnTo.value = options.returnTo || 'select'
  styleExtractResult.value = null
  styleExtractForm.platform = styleExtractForm.platform || ''
  styleExtractForm.contentType = styleExtractForm.contentType || ''
  if (!styleExtractForm.focuses.length) {
    styleExtractForm.focuses = styleExtractFocusOptions.map(option => option.value)
  }
  drawer.type = 'styleExtract'
  drawer.title = '提炼参考写作风格'
}
function toggleStyleExtractFocus(value) {
  const index = styleExtractForm.focuses.indexOf(value)
  if (index >= 0) {
    if (styleExtractForm.focuses.length === 1) {
      showToast('至少保留一个提炼维度')
      return
    }
    styleExtractForm.focuses.splice(index, 1)
  } else {
    styleExtractForm.focuses.push(value)
  }
}
function buildStyleExtractInstruction() {
  const selected = styleExtractFocusOptions
    .filter(option => styleExtractForm.focuses.includes(option.value))
    .map(option => `${option.label}：${option.desc}`)
  return [
    '按标准风格模板提炼，不要求用户补充自由文本。',
    ...selected,
    '只总结可复用的写作结构、语气和手法，不复制原文句子，不把原文事实作为后续生成依据。',
  ].join('\n')
}
async function extractStyleFromURL() {
  const url = String(styleExtractForm.url || '').trim()
  if (!url) {
    showToast('请先输入参考文章 URL')
    return
  }
  loading.action = true
  try {
    styleExtractResult.value = await extractAiGeoExternalSource({
      url,
      extract_type: 'style',
      platform: String(styleExtractForm.platform || '').trim(),
      content_type: String(styleExtractForm.contentType || selectedSkillProfile.value.name || '').trim(),
      instruction: buildStyleExtractInstruction(),
    })
    showToast('写作风格已提炼')
  } catch (error) {
    const message = error?.status === 404
      ? '提炼接口未加载，请重启后端 API 后再试'
      : error?.message || '写作风格提炼失败'
    showToast(message)
  } finally {
    loading.action = false
  }
}
function upsertStyleTemplate(style) {
  const index = styleTemplates.findIndex(item => Number(item.id) === Number(style.id))
  if (index >= 0) {
    styleTemplates.splice(index, 1, style)
  } else {
    styleTemplates.unshift(style)
  }
}
function useStyleTemplate(style) {
  workbench.styleTemplate = style
  closeDrawer()
  chatMessages.push({ id: Date.now(), role: 'ai', text: `已选择参考写作风格：${style.templateName}` })
}
async function saveExtractedStyle() {
  const rawStyle = styleExtractResult.value?.style_template
  if (!rawStyle) {
    showToast('请先完成风格提炼')
    return
  }
  const style = normalizeStyleTemplate(rawStyle)
  upsertStyleTemplate(style)
  if (styleExtractReturnTo.value === 'manage') {
    drawer.type = 'styleManage'
    drawer.title = '风格管理'
    showToast('写作风格已保存')
    return
  }
  workbench.styleTemplate = style
  closeDrawer()
  showToast('已选中参考写作风格')
}
function buildHotspotExtractInstruction() {
  return [
    '按标准热点模板提炼，不要求用户补充自由文本。',
    '提炼主题、搜索意图、关键词、可借势角度、热度建议和风险提示。',
    '只保存外部文章作为热点参考，生成时仍需以品牌资料为准，避免照搬外部事实。',
  ].join('\n')
}
async function extractHotspotFromURL() {
  const url = String(hotspotExtractForm.url || '').trim()
  if (!url) {
    showToast('请先输入热点文章 URL')
    return
  }
  loading.action = true
  try {
    hotspotExtractResult.value = await extractAiGeoExternalSource({
      url,
      extract_type: 'hotspot',
      platform: String(hotspotExtractForm.platform || '').trim(),
      content_type: String(hotspotExtractForm.contentType || selectedSkillProfile.value.name || '').trim(),
      instruction: buildHotspotExtractInstruction(),
    })
    showToast('热点已提炼')
  } catch (error) {
    const message = error?.status === 404
      ? '提炼接口未加载，请重启后端 API 后再试'
      : error?.message || '热点提炼失败'
    showToast(message)
  } finally {
    loading.action = false
  }
}
function upsertHotspot(hotspot) {
  const index = hotspots.findIndex(item => Number(item.id) === Number(hotspot.id))
  if (index >= 0) {
    hotspots.splice(index, 1, hotspot)
  } else {
    hotspots.unshift(hotspot)
  }
}
async function saveExtractedHotspot() {
  const preview = hotspotExtractPreview.value
  if (!hotspotExtractResult.value?.hotspot_draft && !hotspotExtractResult.value?.hotspotDraft) {
    showToast('请先完成热点提炼')
    return
  }
  const metadata = {
    summary: preview.summary,
    search_intent: preview.searchIntent,
    topic_angles: preview.angles,
    keywords: preview.keywords,
    risk_notes: preview.riskNotes,
    article_type: 'URL 提炼',
    content_type: hotspotExtractForm.contentType || selectedSkillProfile.value.name,
    content_status: '待确认',
    priority: 'P1',
  }
  loading.action = true
  try {
    const saved = await createAiGeoHotspot({
      source_id: preview.sourceId || undefined,
      platform: String(hotspotExtractForm.platform || preview.platform || '外部来源').trim(),
      title: preview.title,
      heat_score: preview.heat || 60,
      source_url: preview.sourceUrl,
      captured_at: preview.capturedAt || new Date().toISOString(),
      metadata,
      status: 'active',
    })
    const hotspot = hotspotFromApi(saved)
    upsertHotspot(hotspot)
    if (hotspotExtractReturnTo.value === 'manage') {
      drawer.type = 'hotspot'
      drawer.title = '引用热点'
      showToast('热点已保存')
      return
    }
    workbench.hotspot = hotspot
    closeDrawer()
    showToast('已选中引用热点')
  } catch (error) {
    showToast(error?.message || '热点保存失败')
  } finally {
    loading.action = false
  }
}
async function addManualHotspot() {
  try {
    const title = hotspotSearch.value || '手动录入热点'
    const created = await createAiGeoHotspot({ platform: '手动', title, heat_score: 50 })
    hydrateHotspots([created, ...hotspots])
    workbench.hotspot = hotspots.find(h => Number(h.id) === Number(created.id)) || null
    showToast('已添加手动热点')
  } catch (error) {
    showToast(error?.message || '添加热点失败')
  }
}
function useHotspot(hot, mode) {
  workbench.hotspot = hot
  chatMessages.push({ id: Date.now(), role: 'ai', text: mode === 'dialog' ? `已引用热点：${hot.title}` : `借势角度：围绕「${hot.title}」做轻引用，不夸大热点关系。` })
  closeDrawer()
}
function evaluateWorkbenchReadiness(extraPrompt = '') {
  const intentText = [...workbenchUserMessages.value.map(msg => msg.text), extraPrompt].filter(Boolean).join('\n')
  const knownSlots = knownWorkbenchSlots()
  const userAcceptedDraft = isWorkbenchDraftDecision(intentText)
  const hasSearchQuestion = /写|生成|文章|攻略|问答|种草|小红书|知乎|通勤|怎么|如何|适合|推荐|选择|对比|人群|场景|卖点|关键词|解决|回答|内容|母稿/.test(intentText)
  const hasSpecificBrief = intentText.replace(/\s/g, '').length >= 12
  const hasAudienceOrScene = Boolean(knownSlots.audience || knownSlots.scene || /(目标|人群|用户|场景|口吻|语气|通勤|办公|职场|日常|约会|小个子|女性)/.test(intentText))
  const hasMaterialContext = Boolean(selectedWorkbenchBrand.value?.id || workbench.brandId)
  const ready = ideaSession.stage === 'ready' || userAcceptedDraft || (hasMaterialContext && hasSpecificBrief && (hasSearchQuestion || hasAudienceOrScene))
  const missing = []
  if (!hasSpecificBrief) missing.push('你的核心想法')
  if (!hasSearchQuestion && !hasAudienceOrScene) missing.push('文章要解决的问题或使用场景')
  return {
    ready,
    missing,
    hint: ready ? '可以生成母稿；生成会带上选择资料、内容类型和本轮对话收敛出的 brief。' : `还需要补充：${missing.join('、')}。`,
  }
}

function buildDraftGenerationGuard() {
  const typedPrompt = String(workbench.prompt || '').trim()
  const intentText = [...workbenchUserMessages.value.map(msg => msg.text), typedPrompt].filter(Boolean).join('\n')
  const compactIntent = intentText.replace(/\s/g, '')
  const brand = selectedWorkbenchBrand.value || {}
  const contentType = selectedSkillProfile.value.name || DEFAULT_CONTENT_TYPE
  const hasDirection = Boolean(decisionCard.recommendedDirection || ideaSession.searchProblem)
  const highRiskText = /(真实案例|用户案例|竞品|对标|价格承诺|最低价|保证|功效|产地|销量|认证|第一|最强|唯一|官方背书)/.test(intentText)
  const items = []
  if ((!compactIntent || compactIntent.length < 12 || isGreetingMessage(intentText) || isLowSignalIdea(intentText)) && !hasDirection) {
    items.push({ level: 'warning', label: '缺少核心想法', message: '当前输入还不足以判断文章要回答什么问题，生成结果可能偏品牌通用介绍。' })
  }
  if (!selectedWorkbenchBrand.value?.id && !brand.name) {
    items.push({ level: 'warning', label: '未选择品牌', message: '缺少品牌实体时，AI 只能生成通用内容，品牌定位和资料引用会失真。' })
  } else if (!brand.position && !brand.positioning && !(Array.isArray(workbenchKeywords.value) && workbenchKeywords.value.length)) {
    items.push({ level: 'warning', label: '品牌资料偏少', message: '缺少定位或关键词时，AI 可能用更安全但更通用的表达。' })
  }
  if (highRiskText) {
    items.push({ level: 'warning', label: '存在事实/品牌风险', message: '涉及竞品、承诺、功效、产地、销量、认证或真实案例时，需要确认资料是否可信。' })
  }
  const hasWarnings = items.some(item => item.level === 'warning')
  return {
    requiresConfirm: hasWarnings,
    summary: hasWarnings
      ? `当前可以继续生成，但仍有 ${items.filter(item => item.level === 'warning').length} 个关键缺项；如果继续，AI 会按「${contentType}」和已有资料兜底。`
      : `当前信息较完整，可以按「${contentType}」生成母稿。生成后仍会自动保存为草稿，可继续精修。`,
    riskTitle: hasWarnings ? '可能结果' : '生成说明',
    riskMessage: hasWarnings
      ? '缺项越多，结果越可能偏通用、偏安全表达；确认生成后仍会自动保存草稿，你可以再精修标题、结构和细节。'
      : '确认后会直接生成并自动保存草稿。',
    items,
  }
}

function draftGenerationConfirmDetail(guard) {
  const issues = guard.items || []
  const issueLines = issues.map((item, index) => {
    const prefix = item.level === 'warning' ? '关键缺项' : '提醒'
    return `${index + 1}. ${prefix}｜${item.label}：${item.message}`
  })
  const nextIndex = issueLines.length + 1
  return [
    ...issueLines,
    `${nextIndex}. 继续生成：AI 会按当前内容类型和已有资料兜底，草稿会自动保存，后续仍可精修标题、结构和细节。`,
  ].join('\n')
}

function inferWorkbenchContentType(text) {
  if (/测评|体验|优缺点|值得买吗|好不好/.test(text)) return '商品测评文'
  if (/对比|区别|差异|哪个好|怎么选|避坑|选购/.test(text)) return '商品对比文'
  if (/榜单|排行|推荐清单|合集|top/i.test(text)) return '榜单推荐文'
  if (/热点|热搜|趋势|借势/.test(text)) return '热点借势文'
  if (/活动|促销|上新|节日|营销/.test(text)) return '活动营销文'
  if (/知乎|问答|为什么|如何|FAQ|常见问题/.test(text)) return 'FAQ 问答文'
  if (/场景|通勤|约会|职场|旅行|攻略|怎么穿|怎么选/.test(text)) return '场景解决方案'
  if (/商品|单品|卖点|推荐|种草|购买|转化|小红书|笔记/.test(text)) return '商品种草文'
  if (/品牌|定位|介绍|故事|认知/.test(text)) return '品牌介绍文'
  return workbench.contentType || DEFAULT_CONTENT_TYPE
}

function contentTypeFromLegacySkill(skill = '') {
  if (/品牌介绍/.test(skill)) return '品牌介绍文'
  if (/商品种草|小红书/.test(skill)) return '商品种草文'
  if (/场景攻略/.test(skill)) return '场景解决方案'
  if (/FAQ|知乎|问答/.test(skill)) return 'FAQ 问答文'
  return ''
}
function inferWorkbenchProduct(text) {
  const normalized = text.toLowerCase()
  return (selectedWorkbenchBrand.value?.products || []).find(product => {
    const haystack = [product.name, product.sellingPoints, ...(product.keywords || [])].join(' ').toLowerCase()
    return haystack && haystack.split(/\s+|、|,|，/).some(token => token && token.length > 1 && normalized.includes(token))
  }) || null
}
function extractIdeaSlots(text) {
  const audienceByLabel = text.match(/(?:目标用户|目标人群|写给|面向)[:：是为给]*([\s\S]*?)(?=使用场景|场景|口吻|语气|$|[，。,.；;\n])/)?.[1]?.trim() || ''
  const sceneByLabel = text.match(/(?:使用场景|场景|用于|适合)[:：是为]*([\s\S]*?)(?=目标用户|目标人群|口吻|语气|$|[，。,.；;\n])/)?.[1]?.trim() || ''
  const audience = audienceByLabel || text.match(/(\d{2}岁[^，。,.；;\n]*|小个子|梨形|职场新人|通勤党|18-35岁[^，。,.；;\n]*)/)?.[0] || selectedWorkbenchProduct.value?.audience || selectedWorkbenchBrand.value?.audience || ''
  const scene = sceneByLabel || text.match(/(通勤|上班|职场|办公室|办公场所|高级办公场所|约会|轻正式|旅行|面试|日常|春夏|秋冬)/g)?.join('、') || ''
  const tone = text.match(/(专业问答|种草|轻松|理性|高级|口语|真实|避坑)/g)?.join('、') || ''
  const searchProblem = text.match(/(怎么[\s\S]*?|如何[\s\S]*?|适合[\s\S]*?|为什么[\s\S]*?|选[\s\S]*?)(?=目标用户|目标人群|使用场景|场景|口吻|语气|$|[，。,.；;\n])/)?.[1]?.trim() || ''
  return { audience, scene, tone, searchProblem }
}
function isWorkbenchCorrection(text) {
  return /不是.*补充|已经补充|刚才.*说了|不是给了|你没看到|同样的话|重复问/.test(text)
}
function isWorkbenchDraftDecision(text) {
  return /不补了|不用补|先生成|直接生成|测试一轮|试一轮|看下?是否生成|看能不能生成|可以生成|按这个方向生成|生成母稿|你来写|你帮我写|你决定|你看着写|交给你|按你判断|直接写|帮我生成|替我写/.test(text)
}
function buildIdeaAnalysis(prompt) {
  const messages = workbenchUserMessages.value.map(msg => msg.text)
  if (messages[messages.length - 1] !== prompt) messages.push(prompt)
  const userAcceptedDraft = isWorkbenchDraftDecision(prompt)
  const allText = messages.filter(text => text && !isWorkbenchCorrection(text)).join('\n')
  const recommendedContentType = inferWorkbenchContentType(allText)
  const matchedProduct = inferWorkbenchProduct(allText)
  if ((!workbench.contentType || workbench.contentType === DEFAULT_CONTENT_TYPE) && recommendedContentType) workbench.contentType = recommendedContentType
  if (!workbench.productId && matchedProduct?.id) workbench.productId = String(matchedProduct.id)

  const slots = extractIdeaSlots(allText)
  Object.assign(ideaSession, {
    intent: recommendedContentType,
    recommendedSkill: recommendedContentType,
    recommendedProductId: matchedProduct?.id || ideaSession.recommendedProductId,
    searchProblem: slots.searchProblem || ideaSession.searchProblem,
    audience: slots.audience || ideaSession.audience,
    scene: slots.scene || ideaSession.scene,
    tone: slots.tone || ideaSession.tone,
  })
  const brand = workbenchBrandName()
  const product = matchedProduct?.name || selectedWorkbenchProduct.value?.name || '当前品牌资料'
  const missing = []
  if (!ideaSession.searchProblem) missing.push('这篇文章要回答的具体搜索问题')
  if (!ideaSession.audience) missing.push('目标人群')
  if (!ideaSession.scene) missing.push('使用场景')
  const enoughToGenerate = userAcceptedDraft || allText.replace(/\s/g, '').length >= 12 || (isWorkbenchCorrection(prompt) && (ideaSession.audience || ideaSession.scene))
  ideaSession.stage = enoughToGenerate ? 'ready' : 'shaping'
  ideaSession.brief = `围绕「${brand}」${selectedWorkbenchProduct.value || matchedProduct ? `和「${product}」` : '的现有资料'}，用「${recommendedContentType}」生成一篇${ideaSession.tone || '清晰可信'}的 GEO 母稿；核心问题是「${ideaSession.searchProblem || prompt}」，目标读者为「${ideaSession.audience || '潜在目标用户'}」，场景聚焦「${ideaSession.scene || '由资料和对话推断'}」。`
  const correctionReading = isWorkbenchCorrection(prompt) && ideaSession.stage === 'ready'
    ? `你说得对，目标用户和使用场景已经补充了。我已把目标用户识别为「${ideaSession.audience}」，使用场景识别为「${ideaSession.scene}」，现在可以进入生成或继续细化口吻。`
    : ''
  const decisionReading = userAcceptedDraft
    ? `收到，你已经决定先不继续补充。当前信息会作为一次可生成 brief，缺失的人群、场景或口吻由 AI 按品牌资料和内容类型兜底，不再阻塞生成。`
    : ''
  return {
    intent: recommendedContentType,
    reading: correctionReading || decisionReading || `我会把「${prompt}」先转成可发布母稿的创作 brief。当前资料底座是「${brand}」，${matchedProduct ? `已匹配商品「${matchedProduct.name}」` : selectedWorkbenchProduct.value ? `使用已选商品「${selectedWorkbenchProduct.value.name}」` : `先用「${product}」承接`}，输出框架按「${recommendedContentType}」收敛。`,
    expand: [
      `品牌/资料：从「${selectedWorkbenchBrand.value?.position || brand}」里提取可信卖点，不写空泛介绍。`,
      `用户问题：把想法收敛成“${ideaSession.searchProblem || prompt}”这个可被搜索和 AI 问答引用的入口。`,
      `输出增强：母稿会同时保留标题、摘要、正文、关键词和可改写到渠道的论点。`,
    ],
    converge: ideaSession.brief,
    questions: ideaSession.stage === 'ready' ? ['按这个方向生成母稿', '强化人群和场景', '强化商品卖点'] : missing.slice(0, 2).map(item => `补充${item}`),
  }
}

function knownWorkbenchSlots() {
  const brand = selectedWorkbenchBrand.value || {}
  const product = selectedWorkbenchProduct.value || {}
  return {
    search_problem: ideaSession.searchProblem || decisionCard.recommendedDirection || '',
    audience: ideaSession.audience || product.audience || brand.audience || brand.target_audience || '',
    scene: ideaSession.scene || '',
    tone: ideaSession.tone || brand.tone || '',
  }
}

function isLowSignalIdea(text) {
  const compact = String(text || '').replace(/\s/g, '')
  return compact.length < 8 || /^(我想写一篇文章|写文章|生成文章)$/.test(compact)
}

function isGreetingMessage(text) {
  const compact = String(text || '').replace(/\s/g, '').toLowerCase()
  return /^(你好|你好啊|你哈|hello|hi|嗨|在吗|在不在)$/.test(compact)
}

function workbenchAngleOptions() {
  const brand = workbenchBrandName()
  const product = selectedWorkbenchProduct.value?.name
  const audience = selectedWorkbenchProduct.value?.audience || selectedWorkbenchBrand.value?.audience || '目标用户'
  const scene = ideaSession.scene || '日常/通勤场景'
  if (/品牌介绍/.test(selectedSkillProfile.value.name)) {
    return [
      `「${brand} 是什么风格，适合哪些人」`,
      `「${brand} 为什么适合 ${audience}」`,
      `「${brand} 和同类品牌的差异在哪里」`,
    ]
  }
  if (/商品|种草|测评|对比/.test(selectedSkillProfile.value.name) && product) {
    return [
      `「${product} 适合什么人买」`,
      `「${product} 在${scene}怎么搭」`,
      `「${product} 的卖点和避坑点」`,
    ]
  }
  if (/知乎|问答|FAQ/.test(selectedSkillProfile.value.name)) {
    return [
      `「${brand} 适合 ${audience} 吗」`,
      `「${scene}应该怎么选这类单品」`,
      `「这类风格和普通通勤装有什么区别」`,
    ]
  }
  return [
    `「${scene}怎么穿更合适」`,
    `「${brand} 适合哪些人」`,
    `「怎么根据身材和场景选择单品」`,
  ]
}

function skillWritingFramework(skillName) {
  if (/品牌介绍/.test(skillName)) {
    return {
      entry: '品牌认知入口',
      structure: ['先回答“这个品牌适合谁/是什么风格”', '再用定位、人群和价格带建立可信度', '最后给出适用场景与选择建议'],
      sample: '适合写成“某类人为什么会选择这个品牌”的认知型文章，而不是品牌自夸介绍。',
    }
  }
  if (/商品.*测评/.test(skillName)) {
    return {
      entry: '商品测评入口',
      structure: ['先给一句话结论', '再拆体验维度、优点和不足', '补充适合/不适合人群与 FAQ'],
      sample: '适合写成“这件商品值不值得买、适合什么需求”的测评型内容。',
    }
  }
  if (/商品.*对比/.test(skillName)) {
    return {
      entry: '商品对比入口',
      structure: ['先给对比结论', '再拆核心差异和适合人群差异', '最后给怎么选和注意事项'],
      sample: '适合写成“同类商品/品牌怎么选”的决策型内容。',
    }
  }
  if (/商品|种草/.test(skillName)) {
    return {
      entry: '商品选购入口',
      structure: ['先定义用户痛点或购买犹豫', '再拆商品卖点和适用人群', '最后给场景化购买理由与避坑提醒'],
      sample: '适合写成“这件单品适合谁、解决什么问题、为什么值得选”的种草母稿。',
    }
  }
  if (/场景|解决方案/.test(skillName)) {
    return {
      entry: '场景解决方案入口',
      structure: ['先锁定场景和用户问题', '再给判断标准', '再落到品牌/商品资料里的可用方案', '最后总结适合/不适合人群'],
      sample: '适合写成“某个场景怎么选/怎么穿/怎么搭”的攻略型 GEO 文章。',
    }
  }
  if (/FAQ|知乎|问答/.test(skillName)) {
    return {
      entry: '问答引用入口',
      structure: ['先用一句话回答问题', '再解释判断依据', '补充适用边界', '最后给可执行建议'],
      sample: '适合写成 AI 问答和知乎都能引用的解释型内容，重点是可信、克制、有边界。',
    }
  }
  if (/榜单|推荐文/.test(skillName)) {
    return {
      entry: '榜单推荐入口',
      structure: ['先说明榜单主题和入选标准', '再给推荐对象与适合人群', '最后总结怎么选'],
      sample: '适合写成“推荐清单/排行榜/选购合集”。',
    }
  }
  if (/热点/.test(skillName)) {
    return {
      entry: '热点借势入口',
      structure: ['先解释热点与用户问题的关系', '再用品牌/商品做轻切入', '最后沉淀观点和 FAQ'],
      sample: '适合轻引用热点，不夸大品牌和热点关系。',
    }
  }
  if (/活动|营销/.test(skillName)) {
    return {
      entry: '活动营销入口',
      structure: ['先说明活动核心信息', '再解释适合谁关注', '最后给行动建议和注意事项'],
      sample: '适合活动、上新、促销、节日节点内容。',
    }
  }
  return {
    entry: 'GEO 内容入口',
    structure: ['先回答用户问题', '再引用资料证据', '再给场景建议', '最后沉淀关键词'],
    sample: '适合生成一篇可继续改写到多渠道的通用母稿。',
  }
}

function selectedMaterialSummary() {
  const brand = selectedWorkbenchBrand.value
  const product = selectedWorkbenchProduct.value
  const parts = [
    brand?.position ? `品牌定位是「${brand.position}」` : '',
    brand?.audience ? `目标人群是「${brand.audience}」` : '',
    product?.name ? `当前商品是「${product.name}」` : '',
    product?.sellingPoints ? `商品卖点可用「${product.sellingPoints}」` : '',
    workbenchKeywords.value.length ? `关键词可用「${workbenchKeywords.value.slice(0, 5).join('、')}」` : '',
  ].filter(Boolean)
  return parts.length ? parts.join('，') : `当前只拿到了「${brand?.name || '品牌'}」的基础资料`
}

function compactText(text, limit = 34) {
  const value = String(text || '').replace(/\s+/g, '').trim()
  if (!value) return ''
  return value.length > limit ? `${value.slice(0, limit)}...` : value
}

function isLegacyExpertUnavailableMessage(text) {
  const value = String(text || '')
  return /专家共创服务.{0,12}没有返回结果|固定模板.{0,8}假装判断|专家共创可用后/.test(value)
}

function buildLocalExpertConversationReply(prompt = '') {
  if (isDirectionOptionsRequest(prompt) || /换.*方向|方向/.test(String(prompt || ''))) return buildDirectionAlternativesReply(prompt)
  if (isWritingHowQuestion(prompt)) return buildWritingPlanReply(prompt)
  if (isGeoMetaTopic(prompt)) return buildGeoTopicExpertReply()
  if (isExplicitGenerateRequest(prompt) || isExplicitProceedRequest(prompt) || isExplicitDelegateDecision(prompt)) {
    return buildChatOnlyGenerationReply(prompt)
  }
  const known = knownWorkbenchSlots()
  const direction = decisionCard.recommendedDirection || known.search_problem || chooseRecommendedDirection(prompt)
  if (direction) {
    updateDecisionCard({
      recommendedDirection: direction,
      reason: '根据当前品牌、目标人群和聊天内容收敛出的沟通方向。',
      nextStepLabel: '直接生成 GEO 母稿',
    })
    ideaSession.searchProblem = direction
    setAiStage('direction_recommended')
  }
  const audience = known.audience ? `目标人群我已经按「${known.audience}」处理，不再让你重复补。` : ''
  return [
    '我先按沟通来处理，不会从聊天里自动生成母稿。',
    audience,
    direction ? `当前更适合收敛到这个方向：「${direction}」。` : '',
    '如果你想继续聊，可以让我换方向、做方向诊断、给 3 个其他方向，或调整人群/场景/语气。',
    '如果要写入右侧母稿，只点「生成母稿」按钮。',
  ].filter(Boolean).join('\n\n')
}
function wait(ms) {
  return new Promise(resolve => window.setTimeout(resolve, ms))
}

function nextStreamChunk(text, index) {
  const punctuation = '。！？；\n'
  const current = text[index] || ''
  if (punctuation.includes(current)) return current
  const size = /[A-Za-z0-9]/.test(current) ? 4 : 2
  return text.slice(index, index + size)
}

function isWorkbenchChatNearBottom(el = workbenchChatLogRef.value) {
  if (!el) return true
  return el.scrollHeight - el.scrollTop - el.clientHeight <= 72
}

function handleWorkbenchChatScroll() {
  shouldFollowWorkbenchChat.value = isWorkbenchChatNearBottom()
}

function scrollWorkbenchChatToBottom(force = false) {
  const el = workbenchChatLogRef.value
  if (!el || (!force && !shouldFollowWorkbenchChat.value)) return
  if (workbenchChatScrollFrame) window.cancelAnimationFrame(workbenchChatScrollFrame)
  workbenchChatScrollFrame = window.requestAnimationFrame(() => {
    const current = workbenchChatLogRef.value
    if (current && (force || shouldFollowWorkbenchChat.value)) {
      current.scrollTop = current.scrollHeight
    }
    workbenchChatScrollFrame = 0
  })
}

function queueWorkbenchChatScroll(force = false) {
  nextTick(() => scrollWorkbenchChatToBottom(force))
}

function queueWorkbenchChatScrollAfterRender(force = false) {
  shouldFollowWorkbenchChat.value = true
  nextTick(() => {
    scrollWorkbenchChatToBottom(force)
    window.setTimeout(() => scrollWorkbenchChatToBottom(force), 0)
    window.setTimeout(() => scrollWorkbenchChatToBottom(force), 80)
  })
}

function resumeWorkbenchChatFollow() {
  shouldFollowWorkbenchChat.value = true
  queueWorkbenchChatScroll(true)
}

async function streamAiMessage(text, idea) {
  const message = reactive({ id: Date.now() + 1, role: 'ai', text: '', idea, streaming: true })
  chatMessages.push(message)
  resumeWorkbenchChatFollow()
  chatStreaming.value = true
  try {
    for (let index = 0; index < text.length;) {
      const chunk = nextStreamChunk(text, index)
      message.text += chunk
      queueWorkbenchChatScroll()
      index += chunk.length
      await wait(/[。！？；\n]$/.test(chunk) ? 90 : 22)
    }
  } finally {
    message.streaming = false
    queueWorkbenchChatScroll()
    chatStreaming.value = false
  }
}

function buildGatewayMentorMessages(prompt, idea) {
  const recentMessages = chatMessages.slice(-8).map(msg => ({
    role: msg.role === 'user' ? 'user' : 'assistant',
    content: msg.text,
  }))
  const knownSlots = knownWorkbenchSlots()
  const context = {
    brand: selectedWorkbenchBrand.value ? {
      name: selectedWorkbenchBrand.value.name,
      positioning: selectedWorkbenchBrand.value.position,
      audience: selectedWorkbenchBrand.value.audience,
      keywords: selectedWorkbenchBrand.value.keywordGroups?.flatMap(group => group.keywords || []) || [],
    } : null,
    product: selectedWorkbenchProduct.value ? {
      name: selectedWorkbenchProduct.value.name,
      audience: selectedWorkbenchProduct.value.audience,
      selling_points: selectedWorkbenchProduct.value.sellingPoints,
      keywords: selectedWorkbenchProduct.value.keywords || [],
    } : null,
    content_type: selectedSkillProfile.value.name,
    internal_skill: INTERNAL_DRAFT_SKILL,
    hotspot: workbench.hotspot,
    style_template: styleTemplateSourceSnapshot(workbench.styleTemplate),
    inferred_brief: ideaSession.brief,
    inferred_slots: {
      search_problem: ideaSession.searchProblem,
      audience: ideaSession.audience,
      scene: ideaSession.scene,
      tone: ideaSession.tone,
    },
    known_slots: knownSlots,
    forbidden_reask: [
      knownSlots.audience ? `目标人群已知：${knownSlots.audience}，不要再让用户补充目标人群。` : '',
      knownSlots.scene ? `使用场景已知：${knownSlots.scene}，不要再让用户补充使用场景。` : '',
      knownSlots.search_problem ? `核心问题已知：${knownSlots.search_problem}，不要再让用户补充核心问题。` : '',
    ].filter(Boolean),
    direction_history: ideaSession.directionHistory || [],
    latest_prompt: prompt,
    local_readiness: evaluateWorkbenchReadiness(prompt),
  }
  return [
    {
      role: 'system',
      content: [
        '你是 GEO 母稿共创专家，不是模板生成器、客服或命令执行器。',
        '左侧聊天框只处理母稿共创：母稿生成、母稿修改、改标题/摘要/结构、调整语气/人群/场景、范文转母稿、把 AI 回复应用到母稿。',
        '你的任务是理解用户真实意图，判断当前方向是否有 GEO 价值，并主动把模糊表达升级成可生成、可拆分的内容母版方向。',
        '不要套固定话术，不要复读上一轮，不要把所有回复写成同一个结构。',
        '上下文中的 known_slots 是已识别事实，不是待确认草稿；如果某个槽位已存在，禁止再要求用户补充同一项。',
        '如果已知目标人群，就直接使用它推进；如果已知使用场景，就直接使用它推进；如果已知核心问题，就直接围绕它推进。',
        '当用户问“怎么写、写什么、给几个方向、换方向、内容是什么”时，必须做专家判断和方向发散，不能说“我现在生成母稿”。',
        '聊天输入只用于交流沟通：即使用户说“开始、继续、你写、帮我生成”，也只做判断、诊断、方向推荐或写法建议，不执行生成。',
        '真正生成母稿只能由界面上的“生成母稿”动作触发；真正写入右侧母稿只能由“应用到母稿”动作触发。',
        '不要在聊天框里引导渠道内容、发布计划、保存渠道内容、加入发布计划或渠道改写；这些不属于当前聊天框职责。',
        '如果用户要多个方向，给出新的、互相区分的方向，并避开 direction_history 中已经出现过的方向。',
        '当输出 2-3 个方向/方案时，必须让每个方案使用完全一致的字段层级：方向定位、母稿标题、内容逻辑、GEO价值。不要把“核心钩子、标题参考、方案名”混在同一行。',
        '每个方案的“方向定位”必须是内容策略类型，例如“场景穿搭指南型、价值观共鸣型、风格诊断型”；“母稿标题”必须是可直接生成文章的标题，并且和该方向定位一致。',
        '方案之间必须有真实差异：目标问题、搜索意图、内容角度和适合渠道不能只是换词。不要让方向二和方向三重复。',
        '回复要像资深内容策略顾问：指出问题、给判断、给更优方向和下一步。少问废话；必要时只问一个关键问题。',
        '如果你的回复后适合出现操作按钮，可在回复末尾追加一个 HTML 注释动作标记，格式为：<!-- actions: [{"type":"generate_draft","label":"生成母稿"}] -->。',
        '动作 type 只允许 generate_draft、apply_to_draft、apply_title、regenerate_draft、copy_reply、retry；不得输出任何渠道内容或发布计划动作。',
        '如果你的回复只是给出标题修改方案，只能给 apply_title，不能给 apply_to_draft。',
        '如果你的回复要求用户先选择方案/方向，或包含“请告诉我倾向哪个方案”“选方案一/二/三”等表达，禁止返回 generate_draft 或 regenerate_draft 动作。',
        '只有当你已经给出明确可生成方向时才给 generate_draft；只有当你的回复中包含可写入右侧母稿的明确内容时才给 apply_to_draft；普通解释可只给 copy_reply 或不给动作。',
        '不要把动作标记解释给用户。除可选的 HTML 注释动作标记外，不要输出 JSON，不要说自己基于规则，不要写“固定结构”这类说明。',
      ].join('\n'),
    },
    {
      role: 'user',
      content: `当前 AI GEO 工作台上下文：\n${JSON.stringify(context, null, 2)}`,
    },
    ...recentMessages,
    { role: 'user', content: prompt },
  ]
}

async function streamGatewayMentorReply(prompt, idea) {
  const messages = buildGatewayMentorMessages(prompt, idea)
  const message = reactive({ id: Date.now() + 1, role: 'ai', text: '', idea, streaming: true, suggestedActions: [] })
  chatMessages.push(message)
  resumeWorkbenchChatFollow()
  chatStreaming.value = true
  try {
    await streamAiGeoGatewayInvoke({
      app_code: 'ai-geo',
      app_name: 'AI GEO',
      ai_scenario_code: 'ai_geo_draft_generation',
      params: {
        usage_amount: 1,
        usage_unit: 'calls',
        temperature: 0.45,
        max_tokens: 900,
      },
      input: {
        messages,
      },
    }, {
      onDelta(delta) {
        applyAiMessageText(message, delta)
        queueWorkbenchChatScroll()
      },
      onFinal(event) {
        if (!message.text && event.text) applyAiMessageText(message, event.text, { replace: true })
        else applyAiMessageText(message, '', { replace: false })
        const actions = eventReplyActions(event)
        if (actions.length) message.suggestedActions = actions
        queueWorkbenchChatScroll()
      },
      onError(event) {
        throw new Error(event.error_message || 'AI Gateway 流式调用失败')
      },
    })
  } catch (error) {
    if (!String(message.text || '').trim()) {
      const index = chatMessages.findIndex(item => item.id === message.id)
      if (index >= 0) chatMessages.splice(index, 1)
      appendSystemStatus(error?.message || 'AI 回复失败，请检查模型服务后重试。')
    }
    showToast(error?.message || 'AI 专家共创暂不可用')
  } finally {
    message.streaming = false
    queueWorkbenchChatScroll()
    chatStreaming.value = false
  }
}

function buildChatExampleMessages(prompt) {
  const known = knownWorkbenchSlots()
  const selectedOption = selectedDirectionOptionFromPrompt(prompt)
  const direction = selectedOption?.title || decisionCard.recommendedDirection || known.search_problem || chooseRecommendedDirection(prompt)
  const brand = selectedWorkbenchBrand.value || {}
  const product = selectedWorkbenchProduct.value || {}
  return [
    {
      role: 'system',
      content: [
        '你是 GEO 母稿共创专家。用户要的是聊天里先看一篇范例，不是写入右侧母稿。',
        '你必须直接输出一篇可阅读的中文范例稿，不能说“请点击生成母稿”，不能只解释规则。',
        '这篇范例用于沟通预览：要像真实文章，有标题、摘要、正文小标题和结尾，但不要输出 JSON。',
        '内容要克制可信，不要硬广，不要夸大事实；没有资料支撑的信息用“可以理解为/更适合从...判断”这类表达。',
        '目标是让用户看见写法、语气和结构。右侧编辑器不会被自动更新。',
      ].join('\n'),
    },
    {
      role: 'user',
      content: JSON.stringify({
        user_request: prompt,
        direction,
        brand: {
          name: brand.name || workbenchBrandName('当前品牌'),
          positioning: brand.position || brand.positioning || '',
          target_audience: known.audience || brand.audience || brand.target_audience || '',
          tone: known.tone || brand.tone || '',
        },
        product: product.name ? {
          name: product.name,
          selling_points: product.sellingPoints || '',
          audience: product.audience || '',
        } : null,
        known_slots: known,
        content_type: selectedSkillProfile.value.name,
      }, null, 2),
    },
  ]
}

function buildLocalChatExampleReply(prompt = '') {
  const known = knownWorkbenchSlots()
  const direction = selectedDirectionOptionFromPrompt(prompt)?.title || decisionCard.recommendedDirection || known.search_problem || chooseRecommendedDirection(prompt)
  const brand = workbenchBrandName('Mardi Ladin')
  const audience = known.audience || selectedWorkbenchBrand.value?.audience || '25-34 岁追求品质与格调的都市女性'
  updateDecisionCard({
    recommendedDirection: direction,
    reason: '用户希望先在聊天里查看范例，当前只输出沟通样稿，不写入右侧母稿。',
    nextStepLabel: '直接生成 GEO 母稿',
  })
  ideaSession.searchProblem = direction
  setAiStage('direction_recommended')
  return [
    `以下是一篇聊天范例，只用于先看写法，不会写入右侧母稿。`,
    `# ${direction}`,
    `摘要：这篇内容围绕「${brand} 适合哪些人」展开，不把品牌写成自夸介绍，而是从风格、人群、场景和选择理由出发，帮助${audience}判断它是否符合自己的日常穿搭需求。`,
    `## 先给结论`,
    `${brand} 更适合喜欢克制、耐看、带一点轻法式气质的人。它不是那种第一眼很强烈、靠夸张设计制造记忆点的品牌，而是更适合放进通勤、约会、周末出行等日常场景里，慢慢体现质感和风格稳定性。`,
    `如果你的穿搭需求是“看起来有品位，但不要太用力”，这个方向就比较适合继续深入。`,
    `## 为什么适合这类人`,
    `对${audience}来说，穿衣往往不只是好不好看，还包括是否得体、是否舒适、是否能覆盖多个场景。${brand} 这类轻法式通勤风的价值，正在于它不强迫用户进入某一种很窄的风格标签，而是提供一种比较温和的表达：有女性感，但不过分甜；有设计感，但不太抢戏；适合日常，但不显随便。`,
    `## 适合的使用场景`,
    `它更适合办公室通勤、朋友小聚、周末看展、日常约会这类场景。尤其是当用户不想每天重新思考“今天怎么穿才不出错”时，稳定的版型、低饱和色彩和细节设计，会比强趋势单品更容易被反复穿。`,
    `## 不适合哪些人`,
    `如果你更喜欢强视觉冲击、明显潮流符号、极度修身或很高街的表达，${brand} 可能不是最直接的选择。它的优势不是制造强烈存在感，而是让人看起来更松弛、更干净、更有秩序。`,
    `## 可以怎么写成 GEO 母稿`,
    `正式母稿可以继续补三块内容：第一，品牌风格定义；第二，适合/不适合人群；第三，通勤和日常场景里的选择理由。这样后续既能拆成小红书种草，也能改成知乎问答或公众号长文。`,
  ].join('\n\n')
}

async function streamChatExampleReply(prompt, routeResult = {}) {
  const message = reactive({ id: Date.now() + 1, role: 'ai', text: '', idea: routeResult, streaming: true })
  chatMessages.push(message)
  resumeWorkbenchChatFollow()
  chatStreaming.value = true
  try {
    await streamAiGeoGatewayInvoke({
      app_code: 'ai-geo',
      app_name: 'AI GEO',
      ai_scenario_code: 'ai_geo_draft_generation',
      params: {
        usage_amount: 1,
        usage_unit: 'calls',
        temperature: 0.5,
        max_tokens: 1500,
      },
      input: {
        messages: buildChatExampleMessages(prompt),
      },
    }, {
      onDelta(delta) {
        message.text += delta
        queueWorkbenchChatScroll()
      },
      onFinal(event) {
        if (!message.text && event.text) message.text = event.text
        queueWorkbenchChatScroll()
      },
      onError(event) {
        throw new Error(event.error_message || 'AI 范例生成失败')
      },
    })
  } catch (error) {
    if (!String(message.text || '').trim()) {
      const index = chatMessages.findIndex(item => item.id === message.id)
      if (index >= 0) chatMessages.splice(index, 1)
      appendSystemStatus('AI 范例生成失败，请检查模型配置或稍后重试。')
    }
    showToast(error?.message || 'AI 范例生成失败')
  } finally {
    message.streaming = false
    queueWorkbenchChatScroll()
    chatStreaming.value = false
  }
}

function isCompactDirectionSelectionMessage(text) {
  const value = String(text || '').replace(/\s+/g, '')
  if (!value || value.length > 8) return false
  if (/(生成|写|范文|范例|样稿|母稿|文章|诊断|方向|换)/.test(value) && !/^(方向|方案)[一二三四五六七八九十123456789]$/.test(value)) return false
  return /^(?:好|可以|行|就)?(?:选|用|按)?(?:第)?[一二三四五六七八九十123456789](?:个|号)?$/.test(value) ||
    /^(?:方向|方案)[一二三四五六七八九十123456789]$/.test(value)
}

function shouldHideChatMessage(message) {
  if (!message) return true
  if (message.hidden) return true
  return false
}

function appendUserMessage(text, options = {}) {
  const value = String(text || '').trim()
  if (!value) return false
  const lastUser = [...chatMessages].reverse().find(message => message.role === 'user')
  if (lastUser && String(lastUser.text || '').trim() === value) return false
  chatMessages.push({ id: Date.now(), role: 'user', text: value, hidden: Boolean(options.hidden) })
  return true
}

function appendSystemStatus(text) {
  const value = String(text || '').trim()
  if (!value) return
  const last = chatMessages[chatMessages.length - 1]
  if (last?.role === 'system' && String(last.text || '').trim() === value) return
  if (last?.role === 'system') {
    last.text = value
    return
  }
  chatMessages.push({ id: Date.now() + 2, role: 'system', text: value })
}

function normalizeIntentText(text) {
  return String(text || '').replace(/\s+/g, '').toLowerCase()
}

function classifyChatIntent(text) {
  const value = normalizeIntentText(text)
  if (!value) return 'unknown'
  if (isGreetingMessage(value)) return 'greeting'
  if (isExplicitDelegateDecision(text)) return 'delegate_decision'
  if (isExplicitGenerateRequest(text)) return 'confirm_generate'
  if (isExplicitProceedRequest(text)) return 'proceed_command'
  if (/^(继续|可以|就这个|用这个|按这个|确认)$/.test(value) && decisionCard.recommendedDirection) return 'confirm_generate'
  if (/然后呢|然后.*(怎么|如何|干嘛|做什么|推进)|下一步|接下来|怎么推进|如何推进|怎么做|如何做|怎么办|咋办|怎么下去|往下走/.test(value)) return 'ask_next_step'
  if (isDirectionOptionsRequest(text)) return 'direction_consultation'
  if (/提交/.test(value) && hasEditingDraftContent.value) return 'submit_review'
  if (/重新生成|重写|再生成|换一版|重新来/.test(value)) return 'regenerate_request'
  if (/为什么|原因|解释|为啥|怎么判断/.test(value)) return 'explain_request'
  if (/标题.*(改|优化|自然|短|长|口语|专业)|改.*标题/.test(value)) return 'title_edit_request'
  if (/改|优化|调整|润色|更自然|更专业|更短|更口语|降低营销|增强geo|强化/.test(value) && hasEditingDraftContent.value) return 'draft_edit_request'
  if (/(第?[一二三123]个?|选第[一二三123]|入口[一二三123]?|就这个|用这个|按这个)/.test(value) && decisionCard.recommendedDirection) return 'direction_selection'
  if (/换个方向|换方向|另一个方向/.test(value)) return 'change_direction'
  if (/自定义方向|我想改成|方向改成/.test(value)) return 'custom_direction'
  if (/写|生成|文章|母稿|适合|推荐|怎么|如何|为什么|对比|攻略|问答|种草/.test(value)) return 'new_draft_request'
  return 'unknown'
}

const detectWorkbenchIntent = classifyChatIntent

const workbenchRoutes = new Set([
  'greeting',
  'capability_question',
  'new_draft_request',
  'direction_consultation',
  'delegate_decision',
  'next_step_question',
  'proceed_command',
  'generate_master_draft',
  'edit_current_draft',
  'title_edit_request',
  'review_request',
  'channel_generation_request',
  'context_change',
  'explain_reason',
  'ambiguous',
  'off_topic',
])

const workbenchActions = new Set([
  'none',
  'recommend_direction',
  'choose_best_direction',
  'generate_master_draft',
  'update_master_draft',
  'review_master_draft',
  'generate_channel_content',
  'update_context',
  'ask_clarification',
])

function fallbackWorkbenchIntentResult(message) {
  const localIntent = classifyChatIntent(message)
  if (localIntent === 'confirm_generate' || localIntent === 'proceed_command') {
    return normalizeWorkbenchIntentResult({
      route: 'direction_consultation',
      confidence: 0.68,
      user_goal: '用户在聊天框表达了生成或继续意图，但聊天输入只用于交流沟通。',
      should_execute: false,
      next_action: 'none',
      target: routeTarget('direction_consultation'),
      reasoning_summary: '语义路由失败；按产品原则，聊天框不触发生成，只提示用户使用生成母稿按钮。',
      reply_strategy: 'recommend',
      clarification_question: '',
    })
  }
  if (localIntent === 'delegate_decision') {
    return normalizeWorkbenchIntentResult({
      route: 'direction_consultation',
      confidence: 0.68,
      user_goal: '用户委托系统判断，但聊天输入只用于交流沟通。',
      should_execute: false,
      next_action: 'none',
      target: routeTarget('direction_consultation'),
      reasoning_summary: '语义路由失败；按产品原则，委托表达只做方向沟通，不自动生成。',
      reply_strategy: 'recommend',
      clarification_question: '',
    })
  }
  if (localIntent !== 'unknown') {
    return normalizeWorkbenchIntentResult({
      route: localIntent,
      confidence: 0.58,
      user_goal: `语义路由不可用，采用本地已识别意图继续推进：${localIntent}`,
      should_execute: false,
      next_action: routeNextAction(localIntent),
      target: routeTarget(localIntent),
      reasoning_summary: 'AI 语义路由调用失败或返回异常，使用本地语义判断作为兜底；不输出固定兜底话术。',
      reply_strategy: routeReplyStrategy(localIntent),
      clarification_question: '',
    })
  }
  return normalizeWorkbenchIntentResult({
    route: 'direction_consultation',
    confidence: 0.45,
    user_goal: `AI 语义路由不可用时，不做本地意图枚举；把用户输入交给专家共创能力判断：${String(message || '').slice(0, 80)}`,
    should_execute: false,
    next_action: 'recommend_direction',
    target: routeTarget('direction_consultation'),
    reasoning_summary: 'AI 语义路由调用失败或返回异常，进入通用专家共创模式；兜底层不触发生成。',
    reply_strategy: 'expert_co_create',
    clarification_question: '',
  })
}

function routeNextAction(route) {
  if (route === 'delegate_decision' || route === 'generate_master_draft') return 'generate_master_draft'
  if (route === 'new_draft_request' || route === 'direction_consultation') return 'recommend_direction'
  if (route === 'edit_current_draft' || route === 'title_edit_request') return 'update_master_draft'
  if (route === 'review_request') return 'review_master_draft'
  if (route === 'channel_generation_request') return 'generate_channel_content'
  if (route === 'context_change') return 'update_context'
  if (route === 'ambiguous') return 'recommend_direction'
  return 'none'
}

function routeReplyStrategy(route) {
  if (route === 'greeting') return 'greet'
  if (route === 'capability_question') return 'explain_capability'
  if (route === 'next_step_question') return 'explain_next_step'
  if (route === 'edit_current_draft' || route === 'title_edit_request') return 'edit'
  if (route === 'ambiguous') return 'recommend'
  if (route === 'delegate_decision' || route === 'generate_master_draft' || route === 'proceed_command') return 'execute'
  return 'recommend'
}

function routeTarget(route) {
  if (route === 'title_edit_request') return { object: 'title', field: 'title' }
  if (route === 'edit_current_draft' || route === 'review_request' || route === 'generate_master_draft') return { object: 'master_draft', field: '' }
  if (route === 'channel_generation_request') return { object: 'channel_content', field: '' }
  if (route === 'context_change') return { object: 'context', field: '' }
  return { object: 'conversation', field: '' }
}

function normalizeWorkbenchIntentResult(raw) {
  const route = workbenchRoutes.has(raw?.route) ? raw.route : 'ambiguous'
  const nextAction = workbenchActions.has(raw?.next_action) ? raw.next_action : routeNextAction(route)
  const executableActions = new Set(['generate_master_draft', 'update_master_draft', 'review_master_draft', 'generate_channel_content', 'update_context'])
  return {
    route,
    confidence: Number(raw?.confidence || 0),
    user_goal: String(raw?.user_goal || ''),
    should_execute: Boolean(raw?.should_execute || executableActions.has(nextAction)),
    next_action: route === 'greeting' || route === 'next_step_question' ? 'none' : (route === 'ambiguous' ? 'recommend_direction' : nextAction),
    target: raw?.target && typeof raw.target === 'object' ? raw.target : routeTarget(route),
    reasoning_summary: String(raw?.reasoning_summary || ''),
    reply_strategy: raw?.reply_strategy || routeReplyStrategy(route),
    clarification_question: String(raw?.clarification_question || ''),
  }
}

function isGeoMetaTopic(text) {
  const value = String(text || '').trim()
  if (!value) return false
  const mentionsGeo = /(geo|GEO|生成引擎优化|AI\s*搜索优化|AI\s*问答优化)/i.test(value)
  const mentionsCurrentBrand = new RegExp(`${workbenchBrandName('') || 'Mardi Ladin'}|mardi|ladin|当前品牌|这个品牌|母稿|品牌介绍|商品`, 'i').test(value)
  return mentionsGeo && /(关于|写|介绍|讲|解释|科普|文章|内容)/.test(value) && !mentionsCurrentBrand
}

function isWritingHowQuestion(text) {
  return /(怎么写|如何写|怎么展开|怎么下笔|写法|结构|框架|大纲|怎么组织|怎么开头|从哪.*写|怎么切入)/.test(String(text || ''))
}

function isChatExampleRequest(text) {
  const value = String(text || '')
  return /(范例|范文|示例|样稿|样例|试写|试稿|试下|试一下|试一试|试试|先试|先跑一版|跑一版看看|先写一版|先来一版|写一段看看|写一版看看|给我一篇|来一篇|给一篇|先给我看|先看看)/.test(value)
}

function isExplicitGenerateRequest(text) {
  const value = String(text || '')
  if (isChatExampleRequest(value)) return false
  return /(直接|现在|开始|确认|就这样|按这个|按当前|可以|帮我|替我|马上)?.{0,6}(生成|写出|写一版|出一版).{0,6}(母稿|GEO母稿|geo母稿|正文|文章)|((生成|写).{0,4}(吧|出来|一版|母稿))/.test(value)
}

function isExplicitDelegateDecision(text) {
  return /(你选|你决定|你看着办|按你建议|帮我选|随便你|你来定|你觉得哪个好就用哪个|你定|交给你|按你判断|你写|你来写|你帮我写|帮我写|替我写)/.test(String(text || ''))
}

function isExplicitProceedRequest(text) {
  const value = String(text || '').replace(/\s+/g, '')
  return /^(继续|开始|开始吧|可以|就这个|用这个|按这个|确认|下一步执行)$/.test(value) || /(就按这个做|继续往下做|继续生成|开始生成|确认生成)/.test(value)
}

function isDirectionOptionsRequest(text) {
  const value = String(text || '')
  return /(给|来|出|想要|能不能|可以|帮我|再给|多给).{0,8}([三3几多]个|[三3几多]种|几个|多个).{0,8}(方向|方案|选题|角度|入口|思路|建议)|([三3]个|[三3]种|几个|多个).{0,8}(方向|方案|选题|角度|入口|思路|建议)|(方向|方案|选题|角度|入口|思路|建议).{0,8}(吗|么|呢|有哪些|有几个|给我|再给|多给|吧)/.test(value)
}

function isDirectionRejection(text) {
  return /(换个方向|换方向|换一.*方向|重新.*方向|再给.*方向|方向.*(烂|差|不行|不好|太大|太泛|太宽|太空|不满意|不对|没意思|不准)|不想写这个|不是这个|别写这个|这个方向.*(不|差|烂|泛))/.test(String(text || ''))
}

function refineWorkbenchIntentResult(result, message) {
  const normalized = normalizeWorkbenchIntentResult(result)
  const wantsDirectionOptions = isDirectionOptionsRequest(message)
  const wantsWritingAdvice = isWritingHowQuestion(message)
  const wantsDirectionChange = isDirectionRejection(message)
  const explicitGenerate = isExplicitGenerateRequest(message)
  const explicitDelegate = isExplicitDelegateDecision(message)
  const explicitProceed = isExplicitProceedRequest(message)

  if (
    ['generate_master_draft', 'delegate_decision', 'proceed_command'].includes(normalized.route) &&
    !explicitGenerate &&
    !explicitDelegate &&
    !explicitProceed
  ) {
    return {
      ...normalized,
      route: 'direction_consultation',
      should_execute: false,
      next_action: 'recommend_direction',
      target: routeTarget('direction_consultation'),
      reply_strategy: 'recommend',
      reasoning_summary: '路由试图执行生成，但用户没有明确生成、继续或委托决策；降级为专家方向共创。',
    }
  }
  if (wantsDirectionOptions) {
    return {
      ...normalized,
      route: 'direction_consultation',
      should_execute: false,
      next_action: 'recommend_direction',
      target: routeTarget('direction_consultation'),
      reply_strategy: 'recommend',
      reasoning_summary: '用户在请求多个方向或选题建议，不能触发生成。',
    }
  }
  if (isGeoMetaTopic(message)) {
    return {
      ...normalized,
      route: 'direction_consultation',
      should_execute: false,
      next_action: 'recommend_direction',
      target: routeTarget('direction_consultation'),
      reply_strategy: 'recommend',
      clarification_question: '',
      reasoning_summary: '用户提到 GEO 泛主题，进入专家共创模式，先给可选方向和推荐判断。',
    }
  }
  if (wantsDirectionChange) {
    return {
      ...normalized,
      route: 'direction_consultation',
      should_execute: false,
      next_action: 'recommend_direction',
      target: routeTarget('direction_consultation'),
      reply_strategy: 'recommend',
      reasoning_summary: '用户在否定或要求更换方向，不能触发生成。',
    }
  }
  if (wantsWritingAdvice) {
    return {
      ...normalized,
      route: 'direction_consultation',
      should_execute: false,
      next_action: 'none',
      target: routeTarget('direction_consultation'),
      reply_strategy: 'explain_next_step',
      reasoning_summary: '用户在问写法和展开方式，不是执行生成。',
    }
  }
  if ((normalized.route === 'proceed_command' || normalized.route === 'generate_master_draft') && !normalized.should_execute) {
    return { ...normalized, next_action: 'none' }
  }
  return normalized
}

function extractJsonObject(text) {
  const source = String(text || '').trim().replace(/^```json\s*/i, '').replace(/```$/i, '').trim()
  try {
    return JSON.parse(source)
  } catch (_) {
    const match = source.match(/\{[\s\S]*\}/)
    if (!match) throw new Error('语义路由未返回 JSON')
    return JSON.parse(match[0])
  }
}

function workbenchRouteState(message) {
  return {
    user_message: message,
    current_stage: aiWorkStage.value,
    stage_label: decisionCard.stageLabel || aiStageLabel.value,
    decision_card: {
      brand_name: selectedWorkbenchBrand.value?.name || '',
      product_name: selectedWorkbenchProduct.value?.name || '',
      content_type: selectedSkillProfile.value.name,
      recommended_direction: decisionCard.recommendedDirection || '',
      next_step_label: decisionCard.nextStepLabel || '',
      hotspot: workbench.hotspot?.title || '',
      style_template: workbench.styleTemplate?.templateName || '',
    },
    current_draft: hasEditingDraftContent.value ? {
      id: editingDraft.id,
      title: editingDraft.title,
      summary: editingDraft.summary,
      status: editingDraft.status,
      source: editingDraft.source,
      has_body: Boolean(editingDraft.body),
    } : null,
    conversation: chatMessages.slice(-8).map(message => ({
      role: message.role,
      text: message.text,
    })),
  }
}

async function classifyWorkbenchIntent(message) {
  let text = ''
  try {
    await streamAiGeoGatewayInvoke({
      app_code: 'ai-geo',
      app_name: 'AI GEO',
      ai_scenario_code: 'ai_geo_draft_generation',
      params: {
        usage_amount: 1,
        usage_unit: 'calls',
        temperature: 0,
        max_tokens: 700,
      },
      input: {
        messages: [
          { role: 'system', content: workbenchSemanticRouterSkill },
          { role: 'user', content: JSON.stringify(workbenchRouteState(message), null, 2) },
        ],
      },
    }, {
      onDelta(delta) { text += delta },
      onFinal(event) { if (!text && event.text) text = event.text },
      onError(event) { throw new Error(event.error_message || '语义路由失败') },
    })
    return refineWorkbenchIntentResult(extractJsonObject(text), message)
  } catch (error) {
    console.warn('[ai-geo] semantic route fallback:', error)
    return refineWorkbenchIntentResult(fallbackWorkbenchIntentResult(message), message)
  }
}

function isSimilarPrompt(a, b) {
  const left = normalizeIntentText(a)
  const right = normalizeIntentText(b)
  if (!left || !right) return false
  if (left === right) return true
  const shorter = left.length < right.length ? left : right
  const longer = left.length < right.length ? right : left
  return shorter.length >= 8 && longer.includes(shorter)
}

function hasRepeatedWorkbenchPrompt(prompt) {
  const users = workbenchUserMessages.value
  const previous = users.length > 1 ? users[users.length - 2]?.text : ''
  return isSimilarPrompt(prompt, previous)
}

function chooseRecommendedDirection(prompt = latestWorkbenchUserPrompt.value) {
  const brand = workbenchBrandName('当前品牌')
  const product = selectedWorkbenchProduct.value?.name
  const text = String(prompt || '')
  const contentType = selectedSkillProfile.value.name
  if (isDirectionRejection(text)) return expertDirectionOptions()[0]
  if (isGeoMetaTopic(text)) return ''
  if (/品牌介绍|品牌认知|品牌文|介绍文/.test(text) || /品牌介绍/.test(contentType)) return `${brand} 是什么品牌，适合哪些人？`
  if (/商品测评|测评文|体验|好不好|值得买吗/.test(text) || /商品测评/.test(contentType)) return product ? `${product} 值得买吗？真实测评和适合人群` : `${brand} 的单品值得买吗？测评思路`
  if (/商品对比|对比文|区别|差异|哪个好/.test(text) || /商品对比/.test(contentType)) return product ? `${product} 和同类单品怎么选？` : `${brand} 和同类品牌的差异在哪里？`
  if (/场景解决|场景方案|解决方案/.test(text) || /场景解决/.test(contentType)) return product ? `${product} 在日常场景怎么用？` : `${brand} 适合哪些生活场景？`
  if (/FAQ|问答|常见问题/.test(text) || /FAQ|问答/.test(contentType)) return `${brand} 常见问题：适合哪些人，怎么选择？`
  if (/小个子|适合哪些人|适合谁|人群/.test(text)) return `${brand} 适合哪些人？`
  if (/怎么穿|通勤|场景|上班|职场|约会/.test(text)) return product ? `${product} 在通勤场景怎么穿？` : `${brand} 通勤场景怎么选？`
  if (/对比|差异|区别/.test(text)) return `${brand} 和同类品牌的差异在哪里？`
  if (/商品|单品|推荐|种草/.test(text) && product) return `${product} 适合什么人买？`
  return `${brand} 适合哪些人？`
}

function shouldRecomputeDirection(prompt, intent) {
  const text = String(prompt || '')
  if (intent === 'new_draft_request' || intent === 'regenerate_request' || intent === 'unknown') return true
  if (/品牌介绍|品牌认知|商品测评|商品对比|场景解决|FAQ|问答|种草|榜单|热点|活动/.test(text)) return true
  if (/不知道.*(写|做)|不会.*(写|做)|怎么写|给.*方向|给.*建议|想写/.test(text)) return true
  return false
}

function resolveWorkbenchDirection(prompt, intent) {
  if (shouldRecomputeDirection(prompt, intent)) return chooseRecommendedDirection(prompt)
  return decisionCard.recommendedDirection || chooseRecommendedDirection(prompt)
}

function buildRecommendationMessage(prompt, direction, contentType) {
  const text = String(prompt || '')
  if (isGeoMetaTopic(text)) {
    return buildGeoTopicExpertReply()
  }
  if (/不知道.*(写|做)|不会.*(写|做)|怎么写/.test(text)) {
    return buildWritingPlanReply(prompt)
  }
  if (/品牌介绍|品牌认知|品牌文|介绍文/.test(text) || /品牌介绍/.test(contentType)) {
    return `我理解你的目标是做品牌认知，但这个方向如果只写“品牌介绍”，容易变成百科式资料复述，GEO 价值不够高。\n\n我建议升级成「${direction}」。\n\n原因是：\n1. 它先回答用户真的会搜索/提问的问题。\n2. 它能同时覆盖品牌是什么、适合谁、适合什么场景。\n3. 后续可以从同一篇母稿里拆出小红书种草、知乎问答和公众号深度稿。\n\n下一步建议：先按这个方向生成 GEO 母稿；如果想更商业化，我也可以再给 3 个转化型方向。`
  }
  return `我理解你的目标是：把这个想法变成一篇能被搜索、AI 问答和内容平台共同复用的 GEO 母稿。\n\n当前方向的问题是：如果直接写，很可能只是普通文章，缺少“问题入口”和“可引用答案”。\n\n我建议改成「${direction}」。\n\n原因是：\n1. 它是用户会真实搜索或提问的表达。\n2. 它能承接品牌实体、目标人群、场景和选择理由。\n3. 它适合作为内容母版，后续再转成小红书、知乎、抖音或公众号。\n\n我建议下一步：先生成这篇 GEO 母稿；如果你觉得这个方向还不够强，我可以先给 3 个更商业化/更专业/更种草的方向。`
}

function expertDirectionOptions() {
  const brand = workbenchBrandName('当前品牌')
  const audience = selectedWorkbenchBrand.value?.target_audience || ideaSession.audience || '25-34 岁都市女性'
  const scene = ideaSession.scene || '通勤和日常场景'
  const product = selectedWorkbenchProduct.value?.name
  if (product) {
    return [
      `${product} 适合什么场景，怎么判断值不值得买？`,
      `${product} 和同类单品怎么选？`,
      `${product} 更适合哪些人，哪些情况不建议选？`,
      `${product} 解决了哪些穿搭痛点，适合什么预算和场景？`,
      `${product} 为什么适合${audience}，选择时要看哪些细节？`,
      `${product} 在${scene}怎么搭更自然？`,
      `${product} 是不是适合第一次了解 ${brand} 的用户？`,
      `${product} 和普通通勤单品的差异在哪里？`,
      `${product} 适合做小红书种草还是知乎问答？`,
    ]
  }
  return [
    `${brand} 为什么适合${audience}的通勤和日常穿搭？`,
    `${brand} 和快时尚、网红风品牌有什么不同？`,
    `${brand} 的轻法式风格适合哪些人，不适合哪些人？`,
    `${brand} 是什么风格，为什么适合${scene}？`,
    `${brand} 适合第一次尝试轻法式穿搭的人吗？`,
    `${brand} 的品牌价值应该从风格、人群还是场景切入？`,
    `${brand} 为什么值得被目标用户了解，而不只是被种草？`,
    `${brand} 和同价位通勤品牌相比，选择标准是什么？`,
    `${brand} 适合做品牌认知、用户决策还是渠道种草？`,
  ]
}

function normalizeDirectionKey(direction) {
  return String(direction || '').replace(/[？?，,。.、\s]/g, '').toLowerCase()
}

function rememberDirectionOptions(options = []) {
  const history = Array.isArray(ideaSession.directionHistory) ? ideaSession.directionHistory : []
  options.filter(Boolean).forEach(option => {
    const key = normalizeDirectionKey(option)
    if (!key || history.some(item => normalizeDirectionKey(item) === key)) return
    history.push(option)
  })
  ideaSession.directionHistory = history.slice(-18)
}

function uniqueDirectionOptions(options = [], count = 3, optionsConfig = {}) {
  const blocked = new Set([
    normalizeDirectionKey(optionsConfig.current || decisionCard.recommendedDirection),
    ...(optionsConfig.excludeHistory === false ? [] : (ideaSession.directionHistory || []).map(normalizeDirectionKey)),
  ].filter(Boolean))
  const unique = []
  for (const option of options) {
    const key = normalizeDirectionKey(option)
    if (!key || blocked.has(key) || unique.some(item => normalizeDirectionKey(item) === key)) continue
    unique.push(option)
    if (unique.length >= count) break
  }
  if (unique.length < count) {
    for (const option of options) {
      const key = normalizeDirectionKey(option)
      if (!key || key === normalizeDirectionKey(optionsConfig.current || decisionCard.recommendedDirection) || unique.some(item => normalizeDirectionKey(item) === key)) continue
      unique.push(option)
      if (unique.length >= count) break
    }
  }
  return unique.slice(0, count)
}

function buildDirectionAlternativesReply(prompt = '') {
  const options = expertDirectionOptions()
  const current = decisionCard.recommendedDirection
  const nextOptions = uniqueDirectionOptions(options, 3, { current })
  const next = nextOptions[0]
  rememberDirectionOptions(nextOptions)
  updateDecisionCard({
    recommendedDirection: next,
    reason: '用户认为原方向不够好，已改成更具体、更像真实搜索问题的 GEO 入口。',
    nextStepLabel: '确认后生成母稿',
  })
  ideaSession.searchProblem = next
  setAiStage('direction_recommended')
  const prefix = isDirectionRejection(prompt) ? '认同，原方向确实偏泛。我先做方向诊断，再换成更具体的入口。' : '我给你换一组更强的 GEO 母稿方向。'
  return `${prefix}\n\n原方向的问题：\n1. 问题入口不够具体，容易写成普通介绍。\n2. 用户决策场景不明显，后续渠道转写会缺少抓手。\n3. 可被 AI 摘要引用的短答案不够清晰。\n\n其他方向：\n1. ${nextOptions[0]}（推荐）\n2. ${nextOptions[1] || options[1]}\n3. ${nextOptions[2] || options[2]}\n\n我已先把推荐方向设为第 1 个，不会自动生成；你确认后再生成。`
}

function buildDirectionDiagnosisReply(prompt = '') {
  const direction = prompt || decisionCard.recommendedDirection || chooseRecommendedDirection()
  const options = expertDirectionOptions()
  const upgraded = options.find(option => option !== direction) || options[0] || direction
  updateDecisionCard({
    recommendedDirection: upgraded,
    reason: '已从百科式表达升级为更适合 GEO 的问题入口和内容母版方向。',
    nextStepLabel: '直接生成 GEO 母稿',
  })
  ideaSession.searchProblem = upgraded
  setAiStage('direction_recommended')
  return `我先诊断这个方向：「${direction}」。\n\n它的问题是：\n1. 偏标题或百科介绍，用户为什么要看还不够明确。\n2. 缺少决策逻辑，后续转小红书/知乎/抖音时容易像洗稿。\n3. 没有明确“可被 AI 引用的短答案”，GEO 价值会偏弱。\n\n我建议升级成：「${upgraded}」。\n\n这个方向更好，因为它同时覆盖：品牌/产品定义、适合人群、使用场景、选择理由和 FAQ。它不是一篇普通文章，而是后续多渠道内容的母版。\n\n下一步可以直接生成 GEO 母稿，也可以让我再给 3 个不同目标的方向。`
}

function buildWritingPlanReply(prompt = '') {
  const direction = decisionCard.recommendedDirection || chooseRecommendedDirection(prompt) || expertDirectionOptions()[0]
  updateDecisionCard({
    recommendedDirection: direction,
    reason: '先用真实问题收敛写法，再进入母稿生成。',
    nextStepLabel: '确认后生成母稿',
  })
  setAiStage('direction_recommended')
  return `可以，不急着生成。我建议按「${direction}」来写。\n\n结构用 4 段就够：1. 直接回答问题；2. 解释品牌/商品是什么；3. 讲适合人群和场景；4. 给选择理由、注意事项和 FAQ。\n\n这样比普通介绍更像用户会搜的问题，也更容易被 AI 摘取。`
}

function buildGeoTopicClarificationReply() {
  const brand = workbenchBrandName('当前品牌')
  return `你说的“关于 GEO 的文章”至少有三种可能方向，我不会直接套成普通科普。\n\n1. 科普型：GEO 是什么，适合刚了解的人。\n2. 策略型：品牌为什么要做 GEO，适合老板/市场负责人。\n3. 实操型：如何用一篇母稿覆盖搜索、AI 问答和多渠道内容。\n\n如果目标是后续分发，我建议不要写纯科普，而是写「品牌为什么要重视 GEO，以及如何用一篇母稿覆盖多个渠道」。这个方向更有转化价值。\n\n也可以切回当前品牌，写「${brand} 如何通过 GEO 母稿讲清楚品牌、人群和场景」。`
}

function buildGeoTopicExpertReply() {
  const brand = workbenchBrandName('当前品牌')
  const direction = '品牌为什么要重视 GEO，以及如何用一篇母稿覆盖多个渠道？'
  updateDecisionCard({
    recommendedDirection: direction,
    reason: '用户提出 GEO 泛主题，已从科普文章升级为更有商业价值的策略型母稿方向。',
    nextStepLabel: '直接生成 GEO 母稿',
  })
  ideaSession.searchProblem = direction
  setAiStage('direction_recommended')
  return `我理解你想写“GEO”这个主题，但这个表达目前太泛。\n\n它可能有三种方向：\n1. 科普型：GEO 是什么，适合刚了解的人。\n2. 策略型：品牌为什么要做 GEO，适合老板/市场负责人。\n3. 实操型：怎么写一篇适合 GEO 的母稿，适合内容运营。\n\n我的判断：如果你后面要分发到小红书、知乎、抖音或公众号，不建议写纯科普。更好的方向是「${direction}」。\n\n原因是：\n1. 它有明确商业场景，不只是解释概念。\n2. 它能自然讲到母稿、AI 搜索、问答引用和渠道转译。\n3. 它可以作为内容母版，后续拆成多个平台版本。\n\n下一步可以：\n1. 直接生成这篇 GEO 母稿。\n2. 先生成 3 个其他方向。\n3. 切回当前品牌，写「${brand} 如何用 GEO 母稿讲清品牌价值」。`
}

function buildExpertCoCreationReply(prompt, routeResult = {}) {
  if (isGeoMetaTopic(prompt)) return buildGeoTopicExpertReply()
  const direction = resolveWorkbenchDirection(prompt, 'unknown') || expertDirectionOptions()[0]
  const options = expertDirectionOptions()
  updateDecisionCard({
    recommendedDirection: direction,
    reason: routeResult.reasoning_summary || '无法明确路由时进入专家共创判断，先把模糊输入升级成可生成的 GEO 母稿方向。',
    nextStepLabel: '直接生成 GEO 母稿',
  })
  ideaSession.searchProblem = direction
  ideaSession.brief = `围绕「${direction}」生成一篇 GEO 内容母版，覆盖核心问题、核心答案、目标用户、定位、场景、痛点、选择理由、FAQ 和渠道改写建议。`
  setAiStage('direction_recommended')
  return `我先按专家共创来判断，不把这句话当成简单命令。\n\n我理解你可能想做的是：把「${prompt}」变成一篇能承接搜索、AI 问答和多渠道分发的 GEO 母稿。\n\n当前关键点是：这个表达还缺少清晰的问题入口。如果直接写，容易变成泛泛文章。\n\n我建议先升级成：「${direction}」。\n\n另外两个可选方向：\n1. ${options[0] || direction}\n2. ${options[1] || direction}\n3. ${options[2] || direction}\n\n我的建议：先按推荐方向生成 GEO 母稿；如果你要更种草、更专业或更商业化，我可以继续换一组方向。`
}

function updateDecisionFromPrompt(prompt, reason = '') {
  const direction = chooseRecommendedDirection(prompt)
  updateDecisionCard({
    recommendedDirection: direction,
    reason: reason || '这是更像真实搜索和 AI 问答的问题入口，能自然覆盖品牌、人群和场景。',
    contentType: inferContentType(selectedSkillProfile.value.name),
  })
  ideaSession.searchProblem = direction
  ideaSession.brief = `围绕「${direction}」生成一篇 ${decisionCard.contentType || 'GEO 母稿'}，用品牌实体、目标人群、场景解释和可信资料回答真实问题。`
}

function buildDirectionRecommendationReply(prompt) {
  updateDecisionFromPrompt(prompt)
  setAiStage('ready_to_generate')
  updateDecisionCard({ nextStepLabel: '直接生成 GEO 母稿' })
  return buildRecommendationMessage(prompt, decisionCard.recommendedDirection, selectedSkillProfile.value.name)
}

function buildDelegateDecisionReply(prompt) {
  updateDecisionFromPrompt(prompt || latestWorkbenchUserPrompt.value, '用户已授权 AI 决策，按当前资料选择最稳妥的 GEO 问答入口。')
  setAiStage('ready_to_generate')
  updateDecisionCard({ nextStepLabel: '直接生成 GEO 母稿' })
  return `我来判断。\n\n我不建议写成普通介绍，最稳妥的 GEO 母稿方向是：「${decisionCard.recommendedDirection}」。\n\n原因是：\n1. 它先回答真实问题，不是堆品牌资料。\n2. 它能覆盖品牌定位、目标人群、使用场景和选择理由。\n3. 它后续能拆成小红书种草、知乎问答和公众号深度内容。\n\n我会按这个方向生成内容母版。`
}

function buildDuplicatePromptReply() {
  return `这个方向已经生成或推荐过了。\n\n你可以继续要求：标题更自然、正文更短、更专业、增强 GEO、降低营销感。`
}

function buildGreetingReply() {
  const hasBrand = Boolean(selectedWorkbenchBrand.value?.name || workbench.brandId)
  const hasDraft = hasEditingDraftContent.value
  const currentDirection = decisionCard.recommendedDirection
  if (hasDraft) {
    return '你好，我在。\n\n你可以告诉我想怎么改右侧母稿，比如标题更自然、正文更短、降低营销感或增强 GEO 表达。'
  }
  if (currentDirection) {
    return `你好，我是你的 GEO 写作助手。\n\n我可以帮你生成母稿、优化当前方向，或继续处理「${currentDirection}」。`
  }
  if (hasBrand) {
    return '你好，我在。\n\n你可以直接告诉我要写什么，我会帮你收敛方向并生成右侧母稿。'
  }
  return '你好，我在。\n\n你可以先选择品牌，或直接说一个想写的主题，我会帮你整理成可生成的 GEO 母稿方向。'
}

function buildNextStepReply() {
  const direction = decisionCard.recommendedDirection
  const stage = decisionCard.stageLabel || stageLabels[aiWorkStage.value] || '待明确方向'
  if (hasEditingDraftContent.value) {
    return `当前已有右侧母稿，阶段是「${stage}」。\n\n下一步可以：\n1. 修改标题或正文\n2. 保存草稿\n3. 生成渠道内容\n\n如果你想让我改，直接说具体修改要求。`
  }
  if (direction) {
    return `当前方向已经确定：「${direction}」。\n\n下一步可以：\n1. 直接生成 GEO 母稿\n2. 先做方向诊断\n3. 生成 3 个其他方向\n\n我的建议：如果这个方向还偏泛，先诊断；如果已经明确，直接生成母稿。`
  }
  return '现在还没有明确母稿方向。\n\n下一步可以：\n1. 先告诉我想写的问题\n2. 选择品牌和内容类型\n3. 让我推荐一个方向'
}

function buildCapabilityReply() {
  const direction = decisionCard.recommendedDirection
  const draftPart = hasEditingDraftContent.value ? '也可以直接优化右侧已有母稿。' : '也可以先帮你把想法收敛成母稿方向。'
  const directionPart = direction ? `当前方向是「${direction}」，可以继续生成或换方向。` : '当前还没有固定方向，可以先说一个主题或问题。'
  return `我可以帮你做四件事：\n1. 判断主题有没有 GEO 价值\n2. 把模糊想法升级成母稿方向\n3. 生成可拆分到多渠道的 GEO 母稿\n4. 保存后转成小红书、知乎、抖音或公众号内容\n\n${directionPart}${draftPart}`
}

function buildChatOnlyGenerationReply(prompt = '') {
  const direction = decisionCard.recommendedDirection || chooseRecommendedDirection(prompt)
  if (direction) {
    updateDecisionCard({
      recommendedDirection: direction,
      reason: '聊天输入只用于交流沟通；生成动作由生成母稿按钮统一触发。',
      nextStepLabel: '直接生成 GEO 母稿',
    })
    ideaSession.searchProblem = direction
    setAiStage('direction_recommended')
  }
  return [
    '收到，这里我先把它当成沟通确认，不会从聊天输入里自动生成母稿。',
    direction ? `当前可生成方向是：「${direction}」。` : '',
    '如果要真正写入右侧母稿，请点击「生成母稿」或「直接生成 GEO 母稿」。',
    '在聊天框里你可以继续让我换方向、做诊断、给 3 个方向，或调整人群和语气。',
  ].filter(Boolean).join('\n\n')
}

function buildChatOnlyOperationReply(operation, adviceFocus) {
  const direction = decisionCard.recommendedDirection || editingDraft.title || chooseRecommendedDirection()
  return [
    `我先按沟通处理，不会在聊天里直接执行「${operation}」。`,
    direction ? `当前内容方向是：「${direction}」。` : '',
    `如果你想先聊，我可以继续帮你${adviceFocus}，并给出可选修改方案。`,
    '真正执行写入或渠道生成时，用右侧或底部对应按钮触发。',
  ].filter(Boolean).join('\n\n')
}

function buildDirectionSelectionReply(prompt = '') {
  const selectedOption = selectedDirectionOptionFromPrompt(prompt)
  const direction = selectedOption?.title || decisionCard.recommendedDirection || chooseRecommendedDirection(prompt)
  if (direction) {
    updateDecisionCard({
      recommendedDirection: direction,
      reason: selectedOption ? `已选择方向${selectedOption.label}，后续生成会按这个方向展开。` : '已确认当前推荐方向，后续生成会按这个方向展开。',
      nextStepLabel: '直接生成 GEO 母稿',
    })
    ideaSession.searchProblem = direction
    if (selectedOption?.block) ideaSession.brief = selectedOption.block
    setAiStage('direction_recommended')
  }
  return [
    selectedOption ? `已选方向${selectedOption.label}：「${direction}」。` : `已确认当前方向：「${direction}」。`,
    '我不会从聊天里自动写入右侧母稿。',
    '如果你想先看写法，可以说“给我一篇范例”；如果要正式写入右侧，点击「生成母稿」。',
  ].join('\n\n')
}

function buildOffTopicReply() {
  return '我先把这个理解为非工作台任务。\n\n如果要继续当前工作，可以说：生成母稿、换个方向、优化标题、检查母稿，或生成渠道内容。'
}

async function executeProceedByStage(routeResult) {
  if (aiWorkStage.value === 'ready_to_generate' || decisionCard.recommendedDirection) {
    const result = buildWorkbenchOrchestration(decisionCard.recommendedDirection || latestWorkbenchUserPrompt.value, 'confirm_generate')
    await executeWorkbenchOrchestration(result, latestWorkbenchUserPrompt.value)
    return
  }
  if (hasEditingDraftContent.value) {
    await streamAiMessage('当前已有右侧母稿。\n\n下一步可以继续修改，或保存后直接生成渠道内容。', routeResult)
    return
  }
  await streamAiMessage(buildNextStepReply(), routeResult)
}

async function handleReviewRequest(routeResult) {
  if (!hasEditingDraftContent.value) {
    await streamAiMessage('现在还没有母稿内容。\n\n当前流程为直接流转，你可以先生成或填写母稿，保存后直接进入渠道内容生成。', routeResult)
    return
  }
  await streamAiMessage('当前流程已调整为直接流转。\n\n你可以直接保存母稿，然后生成小红书、知乎、公众号等渠道内容。', routeResult)
}

async function handleChannelGenerationRequest(routeResult) {
  if (editingDraft.id || hasEditingDraftContent.value) {
    if (!editingDraft.id) await persistCurrentDraft(editingDraft.source === '智能生成' ? 'ai_workbench' : 'manual')
    openChannelGenerationConsole({ ...editingDraft })
    await streamAiMessage('已进入渠道内容生成。你可以选择要生成的平台。', routeResult)
    return
  }
  await streamAiMessage('请先生成或填写母稿内容，保存后即可生成渠道内容。', routeResult)
}

async function handleContextChangeRequest(prompt, routeResult) {
  if (isDirectionRejection(prompt)) {
    await streamExpertCoCreationReply(prompt, routeResult)
    return
  }
  await streamAiMessage('我理解你想调整当前上下文。\n\n请在上方切换品牌、商品、内容类型或热点；切换后我会重新计算推荐方向，不会沿用旧方向直接生成。', routeResult)
  if (/方向/.test(prompt)) focusCustomDirection()
}

function prepareExpertCoCreation(prompt, routeResult = {}) {
  const idea = buildIdeaAnalysis(prompt)
  const direction = routeResult?.decision?.recommended_direction || decisionCard.recommendedDirection || chooseRecommendedDirection(prompt)
  if (direction) {
    updateDecisionCard({
      recommendedDirection: direction,
      reason: routeResult.reasoning_summary || '专家共创模式根据当前资料和用户输入收敛出的推荐方向。',
      nextStepLabel: '直接生成 GEO 母稿',
    })
    ideaSession.searchProblem = direction
  }
  setAiStage('direction_recommended')
  return idea
}

async function streamExpertCoCreationReply(prompt, routeResult = {}) {
  const idea = prepareExpertCoCreation(prompt, routeResult)
  await streamGatewayMentorReply(prompt, { ...routeResult, idea })
}

async function requestDraftQuickEdit(action) {
  if (chatStreaming.value || loading.action) return
  if (!hasEditingDraftContent.value) {
    showToast('请先生成或填写母稿内容')
    return
  }
  const label = String(action?.label || '').trim()
  const prompt = String(action?.prompt || label || '').trim()
  if (!prompt) return
  appendUserMessage(label || prompt)
  await streamDraftCoCreationReply(prompt, {
    route: 'edit_current_draft',
    next_action: 'none',
    source: 'draft_quick_action',
    action_label: label,
  })
}

function shouldUseDraftCoCreation(prompt, routeResult = {}) {
  if (!hasEditingDraftContent.value) return false
  const value = normalizeIntentText(prompt)
  if (/^(保存|提交|生成渠道)$/.test(value)) return false
  if (routeResult.route === 'title_edit_request' || routeResult.route === 'edit_current_draft') return true
  if (/太|过于|味道|感觉|不像|不对|不够|有点|不要|别|降低|弱化|更自然|更克制|更松弛|更高级|更像|少一点|浓烈|营销|硬广|广告|销售|客服|套路|油腻|浮夸|尴尬/.test(value)) return true
  return routeResult.route !== 'title_edit_request' && routeResult.route !== 'edit_current_draft'
}

async function streamDraftCoCreationReply(prompt, routeResult = {}) {
  const message = reactive({ id: Date.now() + 1, role: 'ai', text: '', idea: routeResult, streaming: true, suggestedActions: [] })
  chatMessages.push(message)
  resumeWorkbenchChatFollow()
  chatStreaming.value = true
  setAiStage('draft_editing')
  try {
    await streamAiGeoGatewayInvoke({
      app_code: 'ai-geo',
      app_name: 'AI GEO',
      ai_scenario_code: 'ai_geo_draft_generation',
      params: {
        usage_amount: 1,
        usage_unit: 'calls',
        temperature: 0.5,
        max_tokens: 1200,
      },
      input: {
        messages: [
          {
            role: 'system',
            content: [
              '你是 GEO 母稿共创编辑专家，不是状态播报机器人。',
              '用户正在对已经生成的范文/母稿提出反馈。聊天区只用于共创沟通：总结、解释、诊断、提出方案、给示例；不得声称已经更新、已经应用、已经写入右侧母稿。',
              '只有用户点击界面按钮时，系统才会真正写入右侧编辑器或生成母稿。',
              '如果用户要求改标题、改正文、降低营销感、换风格，你要先判断反馈真正指向的问题，再给出 2-3 个可选修改方案；每个方案要说明改法、示例标题或片段、适用情况。',
              '如果用户明确说“给我一篇范例”，可以在聊天区输出范例，但必须说明这只是聊天范例，不写入右侧母稿。',
              '不要只说“已更新右侧母稿”。不要说“没有改动字段”。不要机械确认。',
              '如果用户说“营销味太浓、太硬广、太套路”，你要把表达降到更克制、更像真实经验或内容编辑判断，减少品牌自夸和煽动式形容。',
              '回复结构建议：我理解你想调整什么、当前问题、方案一、方案二、方案三、建议怎么选。每个方案之间要有真实差异。',
              '如果你的回复中已经包含可直接写入右侧母稿的明确改稿内容，可在末尾追加 HTML 注释：<!-- actions: [{"type":"apply_to_draft","label":"应用到母稿","payload":{"content":"这里放完整可写入母稿的标题/摘要/正文"}}] -->。',
              '动作 type 只允许 generate_draft、apply_to_draft、apply_title、regenerate_draft、copy_reply、retry；不得输出渠道内容或发布计划动作。',
              '如果你的回复只是分析问题、给修改建议、给多个方案、询问选择，不能给 apply_to_draft。如果只是给出标题修改方案，只能给 apply_title，不能给 apply_to_draft。',
            ].join('\n'),
          },
          {
            role: 'user',
            content: JSON.stringify({
              user_feedback: prompt,
              current_direction: decisionCard.recommendedDirection,
              current_draft: {
                title: editingDraft.title,
                summary: editingDraft.summary,
                body: editingDraft.body,
                keywords: editingDraft.keywordsText,
              },
              recent_conversation: chatMessages.slice(-8).filter(item => item.id !== message.id).map(item => ({ role: item.role, text: item.text })),
              known_slots: knownWorkbenchSlots(),
            }, null, 2),
          },
        ],
      },
    }, {
      onDelta(delta) {
        applyAiMessageText(message, delta)
        queueWorkbenchChatScroll()
      },
      onFinal(event) {
        if (!message.text && event.text) applyAiMessageText(message, event.text, { replace: true })
        else applyAiMessageText(message, '', { replace: false })
        const actions = eventReplyActions(event)
        if (actions.length) message.suggestedActions = actions
        queueWorkbenchChatScroll()
      },
      onError(event) {
        throw new Error(event.error_message || 'AI 共创修稿失败')
      },
    })
  } catch (error) {
    if (!String(message.text || '').trim()) {
      const index = chatMessages.findIndex(item => item.id === message.id)
      if (index >= 0) chatMessages.splice(index, 1)
      appendSystemStatus(error?.message || 'AI 共创修稿失败，请检查模型服务后重试。')
    }
    showToast(error?.message || 'AI 共创修稿暂不可用')
  } finally {
    message.streaming = false
    chatStreaming.value = false
  }
}

function shouldApplyCoCreatedDraft(text, prompt) {
  const body = extractArticleBody(text)
  if (body.length < 80) return false
  if (/^(好的|明白|收到|可以|建议你|我建议|下一步)/.test(body) && body.length < 180) return false
  return true
}

function applyCoCreatedDraftIfPresent(text, prompt, routeResult = {}) {
  if (!shouldApplyCoCreatedDraft(text, prompt)) return false
  const selectedOption = selectedDirectionOptionFromPrompt(prompt)
  const body = sanitizeGeneratedDraftBody(extractArticleBody(text) || text)
  const title = selectedOption?.title || extractArticleTitle(text, directionDisplayTitle.value || editingDraft.title || latestWorkbenchUserPrompt.value)
  const summary = extractArticleSummary('', body)
  const sourceSnapshot = workbenchSourceSnapshot({
    draft_brief: selectedOption?.draftBrief || workbenchDraftBrief(title),
    selected_direction_option: selectedOption ? {
      index: selectedOption.index,
      label: `方向${selectedOption.label}`,
      title: selectedOption.title,
      block: selectedOption.block,
    } : undefined,
    co_created_from_chat: true,
    route: routeResult.route || '',
  })
  applyDraftToEditor({
    ...editingDraft,
    title,
    summary,
    body,
    keywords: String(editingDraft.keywordsText || '').split(',').map(item => item.trim()).filter(Boolean),
    audit_status: 'approved',
    source: 'ai_workbench',
    source_snapshot: sourceSnapshot,
  }, {
    title,
    summary,
    body,
    keywordsText: editingDraft.keywordsText || workbenchKeywords.value.slice(0, 8).join(', '),
    sourceSnapshot,
  })
  updateDecisionCard({
    recommendedDirection: title,
    reason: selectedOption ? `已按用户选择的方向${selectedOption.label}应用到右侧母稿。` : '已把左侧共创结果应用到右侧母稿。',
    stageLabel: stageLabels.draft_generated,
    nextStepLabel: '',
  })
  setAiStage('draft_generated')
  return true
}

function buildDraftCoCreationFallback(prompt) {
  const title = editingDraft.title || decisionCard.recommendedDirection || '当前母稿'
  if (/标题|题目|headline|title/i.test(String(prompt || ''))) {
    return [
      `我理解你想先优化标题，但这里先只给修改方案，不直接改右侧母稿。`,
      `当前标题可以围绕「${title}」继续收窄，重点是让用户一眼知道：这篇内容解决什么问题、适合谁、为什么要看。`,
      `方案一：搜索问题型。把标题改成用户会直接搜索的问题，例如「${naturalizeTitle(title)}」。适合做 GEO 问答和搜索入口。`,
      `方案二：场景判断型。标题突出具体使用场景，例如「${workbenchBrandName('这个品牌')}适合哪些日常场景，怎么判断值不值得选？」。适合承接小红书、知乎和公众号。`,
      `方案三：人群决策型。标题先锁定目标人群，例如「${workbenchBrandName('这个品牌')}更适合哪些人，不适合哪些人？」。适合解决“我适不适合”的决策问题。`,
      '你可以选 1、2、3，或直接给我你想要的标题语气，我再继续细化。',
    ].join('\n\n')
  }
  return [
    `我理解你的反馈是：${prompt}`,
    '这里先按共创沟通处理，不直接改右侧母稿。',
    `当前「${title}」的问题可能不只是局部措辞，而是表达策略需要重新收窄。`,
    '方案一：降营销感。减少“值得买、推荐、完美”等判断词，改成场景、边界和选择理由。',
    '方案二：增强真实感。把品牌自夸改成用户具体感受，例如穿着场景、搭配限制、适合和不适合的人。',
    '方案三：提高 GEO 密度。用问题式小标题、短答案和 FAQ 承接搜索与 AI 摘要。',
    '你可以选一个方案，我再给你对应的改写片段或完整聊天范例。',
  ].join('\n\n')
}

function isGuidanceRequest(text) {
  return /先聊|聊清楚|引导|帮我梳理|怎么定方向|怎么选方向|还不清楚|不确定/.test(String(text || ''))
}

function buildGuidanceReply() {
  const direction = decisionCard.recommendedDirection || chooseRecommendedDirection()
  updateDecisionCard({
    recommendedDirection: direction,
    reason: '先把母稿入口、人群和场景收窄，再进入生成。',
    nextStepLabel: '直接生成 GEO 母稿',
  })
  setAiStage('ready_to_generate')
  return `可以，我们先把方向定清楚。\n\n1. 文章入口：先回答「${direction}」这个真实问题。\n2. 读者焦点：写给正在判断品牌是否适合自己的用户。\n3. 内容重心：讲清适合人群、使用场景和选择理由。\n\n你只需要补一句：更想强调人群、场景，还是品牌差异？`
}

function isDirectionProgressQuestion(text) {
  return Boolean(decisionCard.recommendedDirection) && /(然后|下一步|接下来|怎么推进|如何推进|怎么做|如何做|怎么办|咋办|怎么写|继续|下去|往下走)/.test(String(text || ''))
}

function buildDirectionProgressReply() {
  return buildNextStepReply()
}

function workbenchDraftBrief(direction = decisionCard.recommendedDirection || chooseRecommendedDirection()) {
  const contentType = selectedSkillProfile.value.name
  return {
    title_direction: direction,
    writing_focus: ['适合人群', '使用场景', '品牌风格', '选择理由', '适用边界', 'FAQ'],
    geo_goal: ['品牌实体', '人群词', '场景词', '问题词', contentType],
    risk_flags: draftGenerationGuard.value.items?.filter(item => item.level === 'warning').map(item => item.label) || [],
  }
}

function buildWorkbenchOrchestration(prompt, intent) {
  const direction = resolveWorkbenchDirection(prompt, intent)
  const contentType = selectedSkillProfile.value.name
  const brandName = selectedWorkbenchBrand.value?.name || workbenchBrandName('当前品牌')
  const productName = selectedWorkbenchProduct.value?.name || '未指定'
  const baseDecision = {
    content_type: contentType,
    recommended_direction: direction,
    reason: '问题型入口更容易被搜索和 AI 问答引用。',
    needs_user_confirm: false,
  }
  const decision_card = {
    brand_name: brandName,
    product_name: productName,
    content_type: contentType,
    recommended_direction: direction,
    stage_label: stageLabels.ready_to_generate,
    next_step_label: '直接生成 GEO 母稿',
  }
  const brief = workbenchDraftBrief(direction)

  if (intent === 'ask_next_step') {
    return {
      intent,
      stage: 'ready_to_generate',
      message: buildNextStepReply(),
      next_action: 'none',
      decision: baseDecision,
      draft_brief: brief,
      decision_card,
    }
  }
  if (intent === 'delegate_decision' || intent === 'confirm_generate') {
    return {
      intent,
      stage: 'ready_to_generate',
      message: intent === 'confirm_generate'
        ? `收到，我按「${direction}」生成 GEO 母稿。\n\n这次会按内容母版来写：先回答核心问题，再补品牌/产品定义、用户决策逻辑、FAQ 和渠道改写素材。`
        : `我选择「${direction}」作为母稿方向。\n\n这个方向比普通介绍更适合 GEO：它能回答真实问题，也能承接品牌定位、目标人群、使用场景和选择理由。接下来我会生成内容母版。`,
      next_action: 'generate_master_draft',
      decision: baseDecision,
      draft_brief: brief,
      decision_card,
    }
  }
  if (intent === 'direction_selection') {
    return {
      intent,
      stage: 'ready_to_generate',
      message: `已按当前方向「${direction}」继续。\n\n你可以直接说“生成母稿”，我再开始写入右侧编辑器。`,
      next_action: 'none',
      decision: baseDecision,
      draft_brief: brief,
      decision_card,
    }
  }
  if (intent === 'new_draft_request' || intent === 'regenerate_request' || intent === 'unknown') {
    return {
      intent,
      stage: 'ready_to_generate',
      message: buildRecommendationMessage(prompt, direction, contentType),
      next_action: 'recommend_direction',
      decision: baseDecision,
      draft_brief: brief,
      decision_card,
    }
  }
  return {
    intent,
    stage: aiWorkStage.value,
    message: '',
    next_action: 'none',
    decision: baseDecision,
    draft_brief: brief,
    decision_card,
  }
}

function applyWorkbenchOrchestrationState(result) {
  if (!result) return
  if (result.stage) setAiStage(result.stage)
  const card = result.decision_card || {}
  updateDecisionCard({
    brandName: card.brand_name || selectedWorkbenchBrand.value?.name || '',
    productName: card.product_name || selectedWorkbenchProduct.value?.name || '',
    contentType: card.content_type || result.decision?.content_type || selectedSkillProfile.value.name,
    recommendedDirection: card.recommended_direction || result.decision?.recommended_direction || decisionCard.recommendedDirection,
    reason: result.decision?.reason || decisionCard.reason,
    stageLabel: card.stage_label || stageLabels[result.stage] || decisionCard.stageLabel,
    nextStepLabel: card.next_step_label || '',
  })
  if (result.draft_brief?.title_direction) {
    ideaSession.searchProblem = result.draft_brief.title_direction
    ideaSession.brief = `围绕「${result.draft_brief.title_direction}」生成一篇 ${result.decision?.content_type || selectedSkillProfile.value.name}，重点覆盖${(result.draft_brief.writing_focus || []).join('、')}。`
  }
}

async function executeWorkbenchOrchestration(result, prompt) {
  applyWorkbenchOrchestrationState(result)
  if (result.message) await streamAiMessage(result.message, result)
  if (result.next_action === 'generate_master_draft') {
    const selectedOption = selectedDirectionOptionFromPrompt(prompt)
    const titleDirection = selectedOption?.title || result.draft_brief?.title_direction || result.decision?.recommended_direction || prompt
    if (selectedOption) {
      updateDecisionCard({
        recommendedDirection: selectedOption.title,
        reason: `用户已选择方向${selectedOption.label}，本次生成必须按该方向展开。`,
      })
      ideaSession.searchProblem = selectedOption.title
      ideaSession.brief = selectedOption.block
    }
    await generateDraftFromChat({
      promptOverride: titleDirection,
      draftBrief: selectedOption?.draftBrief || result.draft_brief,
      appendPrompt: false,
      skipConfirm: true,
      selectedDirectionOption: selectedOption,
    })
  }
}

async function sendWorkbenchMessage() {
  const prompt = String(workbench.prompt || '').trim()
  if (!prompt || chatStreaming.value || loading.action) return
  appendUserMessage(prompt)
  resumeWorkbenchChatFollow()
  workbench.prompt = ''
  const routeResult = {
    route: hasEditingDraftContent.value ? 'draft_co_creation' : 'draft_generation_conversation',
    next_action: 'none',
    should_execute: false,
  }
  aiIntent.value = routeResult.route
  if (hasEditingDraftContent.value) {
    await streamDraftCoCreationReply(prompt, routeResult)
    return
  }
  await streamGatewayMentorReply(prompt, routeResult)
}
function useClarifyQuestion(question) {
  workbench.prompt = question.replace(/^补充/, '')
}

function escapeHtml(text) {
  return String(text || '')
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#39;')
}

function inlineMarkdown(text) {
  return escapeHtml(text)
    .replace(/\*{2,3}([^*\n]+?)\*{2,3}/g, '<strong>$1</strong>')
    .replace(/\*([^*\n]{2,80})\*/g, '<strong>$1</strong>')
    .replace(/`([^`]+)`/g, '<code>$1</code>')
    .replace(/\*{2,3}/g, '')
}

const CHAT_FIELD_HEADING_PATTERN = '方向定位|内容逻辑|GEO价值|改法|适用情况|示例标题|示例片段(?:（正文）)?|示例片段|母稿标题|标题参考|推荐理由|适合渠道|内容目标'
const CHAT_FIELD_HEADING_RE = new RegExp(`^(?:${CHAT_FIELD_HEADING_PATTERN})$`)
const CHAT_FIELD_WITH_BODY_RE = new RegExp(`^(?:${CHAT_FIELD_HEADING_PATTERN})[:：]`)

function normalizeChatMarkdown(text) {
  return String(text || '')
    .replace(/\r\n/g, '\n')
    .replace(/(^|\n)\s*#{1,6}\s*/g, '$1')
    .replace(/\s+#{1,6}\s*(?=(?:我的建议|方案[一二三四五六七八九十\d]+|方向[一二三四五六七八九十\d]+|改法|适用情况|示例标题|示例片段))/g, '\n\n')
    .replace(/(^|\n)\s*>\s*/g, '$1')
    .replace(/\s*>\s*(?=(?:周[一二三四五六日天]|星期[一二三四五六日天]|第[一二三四五六七八九十\d]+|[一二三四五六七八九十\d]+[.、]))/g, '\n')
    .replace(/(^|\n)\s*\|?\s*:?-{2,}:?\s*(?:\|\s*:?-{2,}:?\s*)+\|?\s*(?=\n|$)/g, '\n')
    .replace(/(^|\n)\s*\|?\s*:?\s*\|?\s*(?=\n|$)/g, '\n')
    .replace(/(^|\n)\s*母稿\s*\n+\s*(标题[:：])/g, '$1母稿$2')
    .replace(/(^|\n)\s*文章\s*\n+\s*(标题[:：])/g, '$1文章$2')
    .replace(/(^|\n)\s*(母稿标题|文章标题|标题)\s*\n+\s*[:：]\s*/g, '$1$2：')
    .replace(/\s*---+\s*/g, '\n\n')
    .replace(/(^|\n)\s*\*\s*(?=\n|$)/g, '\n')
    .replace(/(^|\n)\s*\*\*\s*(?=\n|$)/g, '\n')
    .replace(/\s+\*\s*(?=\n|$)/g, '')
    .replace(/\s+\*\*\s*(?=\n|$)/g, '')
    .replace(/\s*[-－]\s*入口([一二三四五六七八九十\d]+)[:：]\s*/g, '\n$1. ')
    .replace(/\s*[-－]\s*(?=(?:标题|标题参考|示例标题|示例片段(?:（正文）)?|母稿标题|方向定位|改法|适用情况|GEO价值|核心逻辑|适合渠道|内容目标|推荐理由|写作说明)[:：])/g, '\n')
    .replace(/(^|\n)\s*(我的建议)\s*[-－]\s*/g, '$1$2：')
    .replace(/\s*[-－]\s*(?=(?:我的建议)[:：]?)/g, '\n')
    .replace(/([。！？；])\s*(?=(?:第[一二三四五六七八九十]+步|注意|请确认|生成方向|范本方向|写作思路与结构建议|结论|信息已经足够|已识别要点|推荐入口|母稿方向|当前上下文|下一步|下一步建议|我的建议|方案[一二三四五六七八九十\d]+|方向[一二三四五六七八九十\d]+|入口问题|对比问题|适配问题|方向定位|标题|标题参考|示例标题|母稿标题|GEO价值|写作说明|第一[剂层段步]|第二[剂层段步]|第三[剂层段步]|第四[剂层段步]|最后)[:：,，])/g, '$1\n\n')
    .replace(/\s*(?=(?:第[一二三四五六七八九十]+步|注意|请确认|生成方向|范本方向|写作思路与结构建议|结论|信息已经足够|已识别要点|推荐入口|母稿方向|当前上下文|下一步|下一步建议|我的建议|方案[一二三四五六七八九十\d]+|方向[一二三四五六七八九十\d]+|方向定位|标题|母稿标题|写作说明|第一[剂层段步]|第二[剂层段步]|第三[剂层段步]|第四[剂层段步]|最后)[:：,，])/g, '\n\n')
    .replace(/\s*(?=(?:入口问题|对比问题|适配问题|核心问题|核心逻辑|内容逻辑|对目标人群的吸引力|价值主张|搜索问题示例|标题参考|示例标题|示例片段(?:（正文）)?|母稿标题|改法|适用情况|GEO价值|适合渠道|内容目标|推荐理由)[:：])/g, '\n')
    .replace(/([:：])\s*(?=\d+[.、]\s*)/g, '$1\n')
    .replace(/([:：])\s*(?=\d+\s*[“"「《])/g, '$1\n')
    .replace(/([。！？；])\s*(?=\d+[.、]\s*)/g, '$1\n')
    .replace(/([。！？；）)])\s*(?=\d+[.、]?\s*[“"「《])/g, '$1\n')
    .replace(/([。！？；）)])\s*(?=(?:这[三两\d]+个方向|总结|我的建议|你更倾向|或者你有|告诉我))/g, '$1\n\n')
    .replace(/\s*(?=(?:周[一二三四五六日天]|星期[一二三四五六日天])[:：])/g, '\n')
    .replace(/\s+(?=\d+[.、]\s*)/g, '\n')
    .replace(/([^\n])\s*(?=\d+[.、]?\s*[“"「《])/g, '$1\n')
    .replace(/([^\n])\s+(?=\d+[.、]\s*\*\*)/g, '$1\n')
    .replace(/\s+(?=[-*•]\s*)/g, '\n')
    .replace(/\n{3,}/g, '\n\n')
    .trim()
}

function plainChatHeadingText(line) {
  return String(line || '')
    .trim()
    .replace(/^#{1,6}\s*/, '')
    .replace(/^\*{1,3}/, '')
    .replace(/\*{1,3}$/, '')
    .replace(/[:：]\s*$/, '')
    .replace(/\s*[-－]\s*$/, '')
    .trim()
}

function inlineDirectionChoiceMarkup(message, headingText) {
  if (!message) return ''
  const options = directionChoiceOptions(message)
  if (!options.length) return ''
  const text = plainChatHeadingText(headingText)
  const option = options.find(item => text.includes(`${item.marker}${item.label}`) || text.includes(`${item.marker}${item.index}`))
  if (!option) return ''
  const checked = selectedDirectionChoiceByMessage[message.id] === option.key ? ' checked' : ''
  return `<label class="message-inline-choice" title="选择${escapeHtml(option.displayLabel)}"><input type="radio" name="direction-choice-${escapeHtml(message.id)}" value="${escapeHtml(option.key)}" data-message-id="${escapeHtml(message.id)}" data-choice-type="direction" data-choice-key="${escapeHtml(option.key)}" data-choice-prompt="${escapeHtml(option.prompt)}"${checked}><span>选择</span></label>`
}

function renderChatLine(line, message = null) {
  const plainHeading = plainChatHeadingText(line)
  if (/^(?:方案|方向)[一二三四五六七八九十\d]+$/.test(plainHeading)) {
    return `<h4 class="message-direction-title">${inlineMarkdown(plainHeading)}${inlineDirectionChoiceMarkup(message, plainHeading)}</h4>`
  }
  if (CHAT_FIELD_HEADING_RE.test(plainHeading)) {
    return `<h5 class="message-field-title">${inlineMarkdown(plainHeading)}</h5>`
  }
  if (/^\*\*[^*]+[:：]?\*\*\s*$/.test(line)) {
    return `<h4>${inlineMarkdown(line)}</h4>`
  }
  if (/^(?:第[一二三四五六七八九十]+步|注意|请确认|生成方向|方案[一二三四五六七八九十\d]+|方向[一二三四五六七八九十\d]+)[:：]/.test(line)) {
    const [title, ...rest] = line.split(/[:：]/)
    const titleText = plainChatHeadingText(title)
    return `<h4 class="message-direction-title">${inlineMarkdown(titleText)}${inlineDirectionChoiceMarkup(message, titleText)}</h4>${rest.join('：').trim() ? `<p>${inlineMarkdown(rest.join('：').trim())}</p>` : ''}`
  }
  if (CHAT_FIELD_WITH_BODY_RE.test(line)) {
    const [title, ...rest] = line.split(/[:：]/)
    return `<h5 class="message-field-title">${inlineMarkdown(title)}</h5>${renderMessageFieldBody(rest.join('：'))}`
  }
  if (/^(?:范本方向|写作思路|建议结构|写作思路与结构建议|结论|信息已经足够|已识别要点|推荐入口|母稿方向|当前上下文|下一步|下一步建议|我的建议|方向定位|入口问题|对比问题|适配问题|核心问题|核心逻辑|内容逻辑|对目标人群的吸引力|价值主张|搜索问题示例|标题参考|示例标题|母稿标题|改法|适用情况|GEO价值|适合渠道|内容目标|推荐理由|可用资料|还差一个关键点|初步判断|生成判断)[:：]/.test(line)) {
    const [title, ...rest] = line.split(/[:：]/)
    return `<h4>${inlineMarkdown(title)}</h4>${renderMessageFieldBody(rest.join('：'))}`
  }
  return splitLongChatParagraphs(line).map(item => `<p>${inlineMarkdown(item)}</p>`).join('')
}

function renderMessageFieldBody(value) {
  const body = String(value || '').trim()
  if (!body) return ''
  const lines = splitReadableBodyLines(body)
  if (lines.length >= 2 && lines.every(line => /^(?:周[一二三四五六日天]|星期[一二三四五六日天])[:：]/.test(line))) {
    return `<ul class="message-field-list">${lines.map(line => `<li>${inlineMarkdown(line)}</li>`).join('')}</ul>`
  }
  return lines.map(line => `<p class="message-field-body">${inlineMarkdown(line)}</p>`).join('')
}

function splitReadableBodyLines(value) {
  const normalized = String(value || '')
    .replace(/\s*>\s*/g, '\n')
    .replace(/\s*(?=(?:周[一二三四五六日天]|星期[一二三四五六日天])[:：])/g, '\n')
    .replace(/([。！？；])\s*(?=(?:如果|但|同时|适合|用户|标题|正文|结论))/g, '$1\n')
    .replace(/\n{2,}/g, '\n')
    .trim()
  const lines = normalized.split('\n').map(line => line.trim()).filter(Boolean)
  if (lines.length > 1) return lines
  return splitLongChatParagraphs(normalized)
}

function shouldRenderInlineListChoices(message) {
  const text = String(message?.text || '')
  if (!text || message?.streaming || message?.role !== 'ai' || message.id !== latestVisibleAiMessageId.value) return false
  if (directionChoiceOptions(message).length) return false
  return /(选择|选一个|选哪|哪个|哪种|倾向|请告诉|确认|更适合|还是|或者|或)/.test(text)
}

function listChoicePrompt(text) {
  return `我选择：${stripMarkdownForDraft(text).replace(/\s+/g, ' ').trim()}。请按这个选择继续。`
}

function renderChoiceListItem(message, text, index) {
  const cleanText = String(text || '').trim()
  const key = `${message.id}-list-${index}`
  const checked = selectedDirectionChoiceByMessage[message.id] === key ? ' checked' : ''
  const prompt = listChoicePrompt(cleanText)
  return `<li class="message-choice-list-item"><label class="message-list-choice"><input type="radio" name="list-choice-${escapeHtml(message.id)}" value="${escapeHtml(key)}" data-message-id="${escapeHtml(message.id)}" data-choice-type="list" data-choice-key="${escapeHtml(key)}" data-choice-prompt="${escapeHtml(prompt)}"${checked}><span>选择</span></label><span class="message-choice-list-content">${inlineMarkdown(cleanText)}</span></li>`
}

function splitLongChatParagraphs(line) {
  const value = String(line || '').trim()
  if (!value) return []
  const withBreakHints = value
    .replace(/\s+(?=(?:第一[剂层段步]|第二[剂层段步]|第三[剂层段步]|第四[剂层段步])[:：])/g, '\n')
    .replace(/\s+(?=(?:最后|写作说明|标题|示例标题|改法|适用情况|正文|完整范文)[:：,，])/g, '\n')
  const hinted = withBreakHints.split('\n').map(item => item.trim()).filter(Boolean)
  if (hinted.length > 1) return hinted.flatMap(item => splitLongChatParagraphs(item))
  if (value.length <= 180) return [value]
  const sentences = value.match(/[^。！？；]+[。！？；]?/g) || [value]
  const paragraphs = []
  let current = ''
  sentences.forEach(sentence => {
    const next = sentence.trim()
    if (!next) return
    if (current && (current.length + next.length > 180 || /^(?:第一|第二|第三|第四|最后|写作说明|标题|正文)/.test(next))) {
      paragraphs.push(current)
      current = next
      return
    }
    current += next
  })
  if (current) paragraphs.push(current)
  return paragraphs.length ? paragraphs : [value]
}

function formatChatMessage(text, message = null) {
  const source = normalizeChatMarkdown(text)
  if (!source) return '<p></p>'
  const blocks = source
    .split(/\n{2,}/)
    .map(block => block.trim())
    .filter(Boolean)
  return blocks.map(block => {
    const lines = block.split('\n').map(line => line.trim()).filter(Boolean)
    const rendered = []
    let listTag = ''
    let listItems = []
    let listHasChoices = false
    const flushList = () => {
      if (!listTag) return
      const listClass = listHasChoices ? ' class="message-choice-list"' : ''
      rendered.push(`<${listTag}${listClass}>${listItems.join('')}</${listTag}>`)
      listTag = ''
      listItems = []
      listHasChoices = false
    }
    for (const line of lines) {
      const numbered = line.match(/^(\d+)(?:[.、]\s*|\s+(?=[“"「《])|(?=[“"「《]))(.+)$/)
      const bulleted = line.match(/^[-•]\s*(.+)$/)
      const splitNumbered = !numbered && !bulleted ? splitInlineNumberedItems(line) : []
      if (splitNumbered.length > 1) {
        flushList()
        const canChooseList = shouldRenderInlineListChoices(message)
        rendered.push(`<ol${canChooseList ? ' class="message-choice-list"' : ''}>${splitNumbered.map((item, index) => canChooseList ? renderChoiceListItem(message, item, index) : `<li>${inlineMarkdown(item)}</li>`).join('')}</ol>`)
        continue
      }
      if (numbered || bulleted) {
        const tag = numbered ? 'ol' : 'ul'
        if (listTag && listTag !== tag) flushList()
        listTag = tag
        const itemText = (numbered?.[2] || bulleted?.[1] || '').trim()
        if (shouldRenderInlineListChoices(message)) {
          listHasChoices = true
          listItems.push(renderChoiceListItem(message, itemText, listItems.length))
        } else {
          listItems.push(`<li>${inlineMarkdown(itemText)}</li>`)
        }
        continue
      }
      flushList()
      rendered.push(renderChatLine(line, message))
    }
    flushList()
    return rendered.join('')
  }).join('')
}

function splitInlineNumberedItems(line) {
  const value = String(line || '').trim()
  if (!/\d+(?:[.、]\s*|\s+(?=[“"「《])|(?=[“"「《]))\S+/.test(value)) return []
  return value
    .split(/(?:^|\s)(?:\d+)(?:[.、]\s*|\s+(?=[“"「《])|(?=[“"「《]))/)
    .map(item => item.trim().replace(/[；;]\s*$/, ''))
    .filter(Boolean)
}

function renderArticleLine(line) {
  const heading = line.match(/^(#{1,4})\s+(.+)$/)
  if (heading) {
    const level = heading[1].length >= 3 ? 'h3' : 'h2'
    return `<${level}>${inlineMarkdown(heading[2].trim())}</${level}>`
  }
  return `<p>${inlineMarkdown(line)}</p>`
}

function formatArticlePreview(text) {
  const source = String(text || '')
    .replace(/\r\n/g, '\n')
    .replace(/\n{3,}/g, '\n\n')
    .trim()
  if (!source) return '<p>暂无正文</p>'
  const blocks = source.split(/\n{2,}/).map(block => block.trim()).filter(Boolean)
  return blocks.map(block => {
    const lines = block.split('\n').map(line => line.trim()).filter(Boolean)
    const rendered = []
    let listTag = ''
    let listItems = []
    const flushList = () => {
      if (!listTag) return
      rendered.push(`<${listTag}>${listItems.join('')}</${listTag}>`)
      listTag = ''
      listItems = []
    }
    for (const line of lines) {
      const numbered = line.match(/^(\d+)[.、]\s*(.+)$/)
      const bulleted = line.match(/^[-*•]\s*(.+)$/)
      if (numbered || bulleted) {
        const tag = numbered ? 'ol' : 'ul'
        if (listTag && listTag !== tag) flushList()
        listTag = tag
        listItems.push(`<li>${inlineMarkdown((numbered?.[2] || bulleted?.[1] || '').trim())}</li>`)
        continue
      }
      flushList()
      rendered.push(renderArticleLine(line))
    }
    flushList()
    return rendered.join('')
  }).join('')
}

function stripMarkdownForDraft(text) {
  return String(text || '')
    .replace(/\*\*/g, '')
    .replace(/```[\s\S]*?```/g, '')
    .replace(/`([^`]+)`/g, '$1')
    .trim()
}

function normalizeDraftArticleText(text) {
  return String(text || '')
    .replace(/\r\n/g, '\n')
    .replace(/\s*---+\s*/g, '\n\n')
    .replace(/\s+(?=(?:标题|文章标题|正文|文章正文|完整母稿正文|写作说明)[:：])/g, '\n')
    .replace(/\s+(?=(?:第一[剂层段步]|第二[剂层段步]|第三[剂层段步]|第四[剂层段步])[:：])/g, '\n\n')
    .replace(/\s+(?=(?:一、|二、|三、|四、|五、|六、))/g, '\n\n')
    .replace(/\s+(?=(?:最后|结尾|总结)[:：,，])/g, '\n\n')
    .replace(/\n{3,}/g, '\n\n')
    .trim()
}

function latestAiWorkbenchText() {
  return [...chatMessages].reverse().find(msg => msg.role === 'ai' && String(msg.text || '').trim())?.text || ''
}

function isMentorTalkLine(line) {
  const value = String(line || '').trim()
  if (!value) return false
  return /^(好的|明白|收到|我会先|我先|我建议|建议先|现在可以|现在，我可以|你可以|请确认|在生成前|如果你|这决定了|我不会)/.test(value)
    || value.includes('补充这篇文章要回答')
    || value.includes('继续帮你放大和收敛')
    || value.includes('生成出来不是品牌介绍')
}

function isDraftMetaLine(line) {
  return /(母稿结构|母稿使用说明|替换占位符|补充细节|调整语气|生成方向|输出要求|输入上下文|收敛 brief|用户原始想法|本轮对话|请根据|严格返回 JSON)/.test(String(line || ''))
}

function extractArticleBody(text) {
  let value = normalizeDraftArticleText(stripMarkdownForDraft(text))
  const markerMatch = value.match(/(?:正文(?:（[^）]*）)?|文章正文|完整母稿正文)[:：]?\s*([\s\S]*)/i)
  if (markerMatch?.[1]) value = markerMatch[1]
  return value
    .split('\n')
    .map(line => line.trim().replace(/^[-#\s]+/, '').trim())
    .filter(line => line && !isMentorTalkLine(line) && !isDraftMetaLine(line))
    .join('\n')
    .replace(/\n{3,}/g, '\n\n')
    .trim()
}

function latestWorkbenchReferenceArticle() {
  const raw = latestAiWorkbenchText()
  if (!raw) return null
  const body = extractArticleBody(raw)
  const hasArticleSignal = /(范文|完整正文|完整母稿|文章标题|标题[:：]|写作说明|第一[剂层段步]|一、)/.test(raw)
  const looksLikeDirectionList = /方向[一二三四五六七八九十\d]+/.test(raw) && !/(范文|完整正文|文章标题|写作说明)/.test(raw)
  if (!hasArticleSignal || looksLikeDirectionList || body.length < 280) return null
  const title = extractArticleTitle(raw, decisionCard.recommendedDirection || ideaSession.searchProblem || latestWorkbenchUserPrompt.value || 'AI GEO 母稿')
  const summary = extractArticleSummary('', body)
  return {
    title,
    summary,
    body,
    keywordsText: workbenchKeywords.value.slice(0, 8).join(', '),
    sourceSnapshot: {
      reference_article: {
        title,
        summary,
        body,
        source: 'latest_ai_workbench_message',
      },
      generation_mode: 'apply_reference_article',
    },
  }
}

function selectedDirectionNumberFromPrompt(prompt) {
  const value = String(prompt || '')
  const compact = value.trim()
  const directMap = { 一: 1, 二: 2, 三: 3, 四: 4, 五: 5, 六: 6, 七: 7, 八: 8, 九: 9, 十: 10 }
  if (/^[一二三四五六七八九十1-9]$/.test(compact)) {
    return Number(directMap[compact] || compact || 0)
  }
  const match = value.match(/(?:方向|方案|第)\s*([一二三四五六七八九十123456789])|([一二三四五六七八九十123456789])\s*(?:号|个)/)
  const raw = match?.[1] || match?.[2] || ''
  return Number(directMap[raw] || raw || 0)
}

function directionNumberFromLabel(label) {
  const value = String(label || '').trim()
  const directMap = { 一: 1, 二: 2, 三: 3, 四: 4, 五: 5, 六: 6, 七: 7, 八: 8, 九: 9, 十: 10 }
  return Number(directMap[value] || value || 0)
}

function chineseDirectionLabel(index) {
  return ['零', '一', '二', '三', '四', '五', '六', '七', '八', '九', '十'][index] || String(index)
}

const DIRECTION_OPTION_FIELD_NAMES = '方向定位|母稿标题|标题参考|文章标题|标题|内容逻辑|GEO价值|核心逻辑|推荐理由|适用情况|改法|适合渠道|内容目标|价值主张'
const DIRECTION_OPTION_FIELD_BOUNDARY_RE = new RegExp(`[\\s\\n\\-－]*(?:${DIRECTION_OPTION_FIELD_NAMES})[:：]`)

function cleanDirectionOptionSnippet(value) {
  return stripMarkdownForDraft(value)
    .split(DIRECTION_OPTION_FIELD_BOUNDARY_RE)[0]
    .replace(/^[:：\-－\s]+/, '')
    .replace(/[《》「」“”"]/g, '')
    .replace(/\s+/g, ' ')
    .replace(/[。；;，,]\s*$/, '')
    .trim()
}

function extractDirectionOptionField(block, fieldName) {
  const fieldPattern = new RegExp(`(?:^|\\n|\\s|[-－])(?:${fieldName})[:：]\\s*([\\s\\S]*?)(?=(?:\\n|\\s|[-－])(?:${DIRECTION_OPTION_FIELD_NAMES})[:：]|$)`)
  return cleanDirectionOptionSnippet(block.match(fieldPattern)?.[1] || '')
}

function directionOptionTitleFromBlock(block, marker, label, optionIndex, fallback = '') {
  const directionTitle = extractDirectionOptionField(block, '方向定位')
  if (directionTitle) return directionTitle.slice(0, 56)
  const draftTitle = extractDirectionOptionField(block, '母稿标题|标题参考|文章标题|标题')
  if (draftTitle) return draftTitle.slice(0, 72)
  const firstLineTitle = cleanDirectionOptionSnippet(
    block
      .split('\n')
      .find(line => line.trim()) || '',
  )
    .replace(new RegExp(`^(?:${marker}|方向|方案)?\\s*(?:${label}|${optionIndex})[.、:：\\s-]*`), '')
    .replace(/[（(]\s*推荐\s*[）)]/g, '')
    .trim()
  return (firstLineTitle || fallback).slice(0, 72)
}

function directionChoiceOptions(message) {
  const text = String(message?.text || '')
  if (!text || message?.streaming || message?.role !== 'ai' || message.id !== latestVisibleAiMessageId.value) return []
  const source = normalizeChatMarkdown(text)
  const seen = new Set()
  const matches = [...source.matchAll(/(?:^|\n)\s*(?:(方向|方案)\s*([一二三四五六七八九十\d]+)|([1-9]\d*)[.、])\s*[:：]?\s*([^\n]*)/g)]
    .filter(match => {
      if (/^如果/.test(String(match[4] || '').trim())) return false
      const optionIndex = directionNumberFromLabel(match[2] || match[3])
      if (!optionIndex || seen.has(optionIndex)) return false
      seen.add(optionIndex)
      return true
    })
  if (matches.length < 2) return []
  return matches
    .map((match, index) => {
      const next = matches[index + 1]
      const start = match.index + (match[0].startsWith('\n') ? 1 : 0)
      const end = next ? next.index : source.length
      const block = normalizeDraftArticleText(stripMarkdownForDraft(source.slice(start, end)))
      const marker = match[1] || '方向'
      const optionIndex = directionNumberFromLabel(match[2] || match[3])
      if (!optionIndex) return null
      const label = chineseDirectionLabel(optionIndex)
      const displayLabel = `${marker}${label}`
      const title = directionOptionTitleFromBlock(block, marker, label, optionIndex, displayLabel)
      return {
        key: `${message.id}-${optionIndex}`,
        index: optionIndex,
        marker,
        label,
        displayLabel,
        title,
        prompt: `我选择${displayLabel}：${title}。请按这个方向继续生成母稿。`,
      }
    })
    .filter(Boolean)
}

function compactChoiceText(value) {
  return String(value || '')
    .replace(/\s+/g, ' ')
    .replace(/^(?:还是|或者|或)\s*/, '')
    .replace(/[？?。；;，,]\s*$/, '')
    .trim()
}

function clarificationChoiceOptions(message) {
  const text = String(message?.text || '')
  if (!text || message?.streaming || message?.role !== 'ai' || message.id !== latestVisibleAiMessageId.value) return []
  if (directionChoiceOptions(message).length) return []
  const source = stripMarkdownForDraft(normalizeChatMarkdown(text)).replace(/\n+/g, ' ')
  const pair = source.match(/(?:是|选择|主要吸引)\s*([^？?。]+?)\s*(?:，|,)?\s*(?:还是|或者|或)\s*([^？?。]+?)(?:[？?。]|$)/)
  if (!pair) return []
  const rawOptions = [compactChoiceText(pair[1]), compactChoiceText(pair[2])]
    .map(option => option.replace(/^(?:用户|人群|目标人群)[:：]\s*/, '').trim())
    .filter(option => option.length >= 4 && option.length <= 90)
  if (rawOptions.length < 2) return []
  return rawOptions.map((title, index) => ({
    key: `${message.id}-clarify-${index}`,
    index,
    displayLabel: index === 0 ? '选项 A' : '选项 B',
    title,
    prompt: `我选择：${title}。请按这个选择继续。`,
  }))
}

function selectDirectionChoice(message, option) {
  if (!message?.id || !option) return
  selectedDirectionChoiceByMessage[message.id] = option.key
  workbench.prompt = option.prompt
  nextTick(() => {
    workbenchPromptRef.value?.focus?.()
  })
}

function selectClarificationChoice(message, option) {
  selectDirectionChoice(message, option)
}

function handleMessageInlineChoice(event) {
  const input = event?.target
  if (!(input instanceof HTMLInputElement) || !['direction', 'list'].includes(input.dataset.choiceType || '')) return
  const messageId = input.dataset.messageId || ''
  const choiceKey = input.dataset.choiceKey || ''
  const prompt = input.dataset.choicePrompt || ''
  if (!messageId || !choiceKey || !prompt) return
  selectedDirectionChoiceByMessage[messageId] = choiceKey
  workbench.prompt = prompt
  nextTick(() => {
    workbenchPromptRef.value?.focus?.()
  })
}

function selectedDirectionOptionFromPrompt(prompt) {
  const index = selectedDirectionNumberFromPrompt(prompt)
  if (!index) return null
  const raw = latestAiWorkbenchText()
  if (!raw) return null
  const label = chineseDirectionLabel(index)
  const markers = ['方向', '方案']
  let start = -1
  let markerText = ''
  for (const marker of markers) {
    const patterns = [
      `${marker}${label}`,
      `${marker}${index}`,
      `${index}.`,
      `${index}、`,
    ]
    for (const pattern of patterns) {
      const found = raw.indexOf(pattern)
      if (found >= 0 && (start < 0 || found < start)) {
        start = found
        markerText = pattern
      }
    }
  }
  if (start < 0) return null
  let end = raw.length
  for (let next = index + 1; next <= 10; next += 1) {
    const nextLabel = chineseDirectionLabel(next)
    for (const marker of markers) {
      for (const pattern of [`${marker}${nextLabel}`, `${marker}${next}`, `${next}.`, `${next}、`]) {
        const found = raw.indexOf(pattern, start + markerText.length)
        if (found > start && found < end) end = found
      }
    }
  }
  const block = normalizeDraftArticleText(stripMarkdownForDraft(raw.slice(start, end)))
  const title = directionOptionTitleFromBlock(block, '方向', label, index)
  const fallbackTitle = `${workbenchBrandName('当前品牌')} ${block.split('\n').find(line => line && !/^方向/.test(line)) || `方向${label}`}`.slice(0, 42)
  const titleDirection = (title || fallbackTitle).slice(0, 60)
  return {
    index,
    label,
    title: titleDirection,
    block,
    draftBrief: {
      title_direction: titleDirection,
      writing_focus: block.split('\n').filter(Boolean).slice(0, 6),
      selected_direction_index: index,
      selected_direction_label: `方向${label}`,
    },
  }
}

function extractArticleTitle(text, fallback = '') {
  const source = stripMarkdownForDraft(text)
  const explicit = source.match(/(?:^|\n)\s*(?:标题|文章标题)[:：]\s*([^\n]+)/)?.[1]
  const firstLine = extractArticleBody(source).split('\n').find(line => line.trim() && !/^\d+[.、]/.test(line.trim()))
  const value = cleanDraftTitle(explicit || fallback || firstLine || 'AI GEO 母稿')
    .split(/[>\n]/)[0]
    .trim()
  return value.slice(0, 42)
}

function extractArticleSummary(summary, body) {
  const raw = stripMarkdownForDraft(summary)
  if (raw && !isMentorTalkLine(raw) && !isDraftMetaLine(raw)) return raw.slice(0, 160)
  const sentence = String(body || '').replace(/\n/g, ' ').match(/^(.+?[。！？])/)
  return (sentence?.[1] || String(body || '').slice(0, 120) || '基于当前资料和对话生成的 GEO 母稿。').slice(0, 160)
}

function directionTitleOnly(value) {
  let text = normalizeGeneratedDraftText(value)
  if (!text) return ''
  const structured = parseGeneratedDraftJson(text)
  if (structured.title) text = normalizeGeneratedDraftText(structured.title)
  const titleMatch = text.match(/(?:母稿标题|文章标题|标题|title)[:：]\s*([^\n]+)/i)
  if (titleMatch?.[1]) text = titleMatch[1]
  text = cleanDraftTitle(text)
    .split(/\n/)[0]
    .split(/\s*[-－]\s*(?:内容逻辑|GEO价值|方向定位|核心逻辑|适合渠道|推荐理由)[:：]?/)[0]
    .replace(/^方向[一二三四五六七八九十\d]+[:：\s-]*/, '')
    .trim()
  return text.length > 42 ? `${text.slice(0, 42)}...` : text
}

function draftFallbackFromChat(prompt) {
  const source = [
    ideaSession.brief,
    latestAiWorkbenchText(),
    prompt,
    selectedMaterialSummary(),
  ].filter(Boolean).join('\n')
  const body = extractArticleBody(source) || `围绕「${ideaSession.searchProblem || prompt}」，结合「${workbenchBrandName('当前资料')}」的资料，输出一篇先回答用户问题、再给出选择理由和场景建议的 GEO 母稿。`
  const title = extractArticleTitle(source, ideaSession.searchProblem || prompt || 'AI GEO 母稿')
  const summary = extractArticleSummary(ideaSession.brief, body)
  return {
    title: String(title).slice(0, 42),
    summary: String(summary).slice(0, 160),
    body,
    keywordsText: workbenchKeywords.value.slice(0, 6).join(', '),
  }
}

function buildAiCoCreateDraft(prompt) {
  const brand = workbenchBrandForGeneration()
  const product = selectedWorkbenchProduct.value
  const brandName = brand.name || '当前品牌'
  const productName = product?.name || ''
  const audience = ideaSession.audience || product?.audience || brand.audience || '目标用户'
  const scene = ideaSession.scene || '日常使用场景'
  const problem = ideaSession.searchProblem || prompt || `${brandName}适合哪些人`
  const sellingPoint = product?.sellingPoints || brand.position || selectedMaterialSummary()
  const priceBand = product?.priceBand || brand.priceBand || ''
  const keywords = workbenchKeywords.value.length
    ? workbenchKeywords.value
    : [brandName, productName, audience, scene, problem].filter(Boolean).slice(0, 6)
  const title = extractArticleTitle(problem, `${brandName}适合哪些人？一篇看懂选择思路`)
  const summary = `围绕「${problem}」，结合${brandName}${productName ? `和${productName}` : ''}的资料，拆解适合人群、选择理由和使用场景，帮助读者快速判断是否值得选择。`
  const body = [
    `${title}`,
    '',
    `很多人在搜索「${problem}」时，真正想知道的不是品牌自夸，而是它是否适合自己的需求、场景和预算。${brandName}${productName ? `的${productName}` : ''}可以先放在「${audience}」和「${scene}」这两个维度里判断。`,
    '',
    `一、先看它解决什么问题`,
    `${sellingPoint}。这部分信息适合转化成用户能理解的选择理由：它不是简单告诉你“值得买”，而是说明它在哪些场景里更顺手，哪些需求下更匹配。`,
    '',
    `二、再看适合哪些人`,
    `${brandName}更适合关注品质、风格稳定和实际使用感的人群。${priceBand ? `如果你的预算大致在「${priceBand}」，可以把它作为同类选择中的候选项。` : '如果你更看重长期使用和搭配稳定性，可以把它作为同类选择中的候选项。'}如果你只追求一次性低价或极强功能参数，则需要再和其他选择对比。`,
    '',
    `三、放到具体场景里判断`,
    `在${scene}中，用户通常会同时在意是否好用、是否自然、是否容易搭配，以及是否会带来额外负担。围绕这些问题写清楚，文章会更像一篇可被搜索和 AI 问答引用的内容，而不是普通广告文案。`,
    '',
    `四、结论`,
    `如果你的需求接近「${audience}」在「${scene}」中的真实问题，${brandName}${productName ? `的${productName}` : ''}可以作为一个值得了解的选择。建议先看自己的核心需求，再对照品牌资料、商品卖点和使用边界做判断。`,
  ].join('\n')
  return {
    title: String(title).slice(0, 42),
    summary: String(summary).slice(0, 160),
    body,
    keywordsText: keywords.slice(0, 8).join(', '),
  }
}

function draftLooksLikeClarification(draft) {
  const text = [draft?.title, draft?.summary, draft?.body].filter(Boolean).join('\n')
  return /请确认|请补充|你希望|是否可以|还差|建议先回答|可以从.*选一个|再告诉我|生成前|现在可以点击|篇幅|深度干货|快速种草/.test(text)
}

function ensureDraftConclusionBody(body, title = '', summary = '') {
  const source = String(body || '').replace(/\r\n/g, '\n').trim()
  if (!source) return source
  const lines = source.split('\n').map(line => line.trim()).filter(Boolean)
  if (!lines.length || !isConclusionHeading(lines[lines.length - 1])) return source
  const subject = String(title || summary || ideaSession.searchProblem || workbenchBrandName('当前内容')).trim()
  const brand = workbenchBrandName('')
  const product = selectedWorkbenchProduct.value?.name || ''
  const target = product || brand || subject || '这类内容'
  return `${source}\n\n整体来看，${target}是否值得选择，关键不在于单一卖点，而在于它是否匹配用户的人群特征、使用场景、预算区间和选择边界。围绕这些标准继续展开，内容会更像一篇可被搜索和 AI 问答引用的决策参考，而不是普通广告介绍。`
}

function normalizeGeneratedDraftText(text) {
  return String(text || '')
    .replace(/\\r\\n/g, '\n')
    .replace(/\\n/g, '\n')
    .replace(/\\"/g, '"')
    .replace(/\\t/g, ' ')
    .trim()
}

function cleanDraftTitle(value) {
  return normalizeGeneratedDraftText(value)
    .replace(/^#{1,6}\s*/, '')
    .replace(/\*\*/g, '')
    .replace(/__([^_]+)__/g, '$1')
    .replace(/`([^`]+)`/g, '$1')
    .replace(/^(?:标题|文章标题|母稿标题|示例标题|完整标题|最终标题|title)[:：]\s*/i, '')
    .replace(/^[-*•]\s*/, '')
    .replace(/^["“”《》「」]+|["“”《》「」]+$/g, '')
    .replace(/\s+/g, ' ')
    .trim()
}

function sanitizeGeneratedDraftBody(text) {
  let value = normalizeGeneratedDraftText(text)
  const structured = parseGeneratedDraftJson(value)
  if (structured.body) value = normalizeGeneratedDraftText(structured.body)
  value = value
    .replace(/^\s*(?:标题|文章标题)[:：][^\n]+\n+/i, '')
    .replace(/^\s*(?:摘要|summary)[:：][^\n]+\n+/i, '')
    .replace(/^\s*(?:正文|内容|body)[:：]\s*/i, '')
    .replace(/\n{3,}/g, '\n\n')
    .trim()
  const channelStart = value.search(/(?:^|\n)\s*(?:渠道改写建议|渠道改写|渠道拆解|平台分发|小红书|知乎|抖音|公众号)[:：]?\s*(?:\n|$)/)
  if (channelStart > 0) value = value.slice(0, channelStart).trim()
  return value
}

function parseGeneratedDraftJson(text) {
  const value = String(text || '').trim()
  if (!value || !/[{]/.test(value) || !/"(?:title|summary|body)"\s*:/.test(value)) return {}
  const start = value.indexOf('{')
  const end = value.lastIndexOf('}')
  if (start < 0 || end <= start) return {}
  try {
    const parsed = JSON.parse(value.slice(start, end + 1))
    return parsed && typeof parsed === 'object' ? parsed : {}
  } catch (_) {
    return {}
  }
}

function normalizedDraftFieldsFromResponse(draft, fallback) {
  const bodyValue = apiField(draft, 'body', 'Body')
  const titleValue = apiField(draft, 'title', 'Title')
  const summaryValue = apiField(draft, 'summary', 'Summary')
  const structured = {
    ...parseGeneratedDraftJson(summaryValue),
    ...parseGeneratedDraftJson(bodyValue),
  }
  return {
    title: cleanDraftTitle(structured.title || titleValue || fallback.title),
    summary: normalizeGeneratedDraftText(structured.summary || summaryValue || fallback.summary),
    body: sanitizeGeneratedDraftBody(structured.body || bodyValue || fallback.body),
  }
}

function applyDraftToEditor(draft, fallback) {
  const keywords = parseJsonArray(apiField(draft, 'keywords', 'Keywords'))
  const fields = normalizedDraftFieldsFromResponse(draft, fallback)
  const rawBody = extractArticleBody(fields.body) || fields.body
  const title = extractArticleTitle(fields.title || rawBody, fallback.title)
  const body = ensureDraftConclusionBody(rawBody, title, fallback.summary)
  const summary = extractArticleSummary(fields.summary || fallback.summary, body)
  const sourceSnapshot = parseJsonObject(apiField(draft, 'source_snapshot', 'SourceSnapshot', 'sourceSnapshot'))
  const draftId = Number(apiField(draft, 'id', 'ID') || editingDraft.id || 0)
  Object.assign(editingDraft, {
    id: draftId,
    title,
    summary,
    body,
    keywordsText: keywords.length ? keywords.join(', ') : fallback.keywordsText,
    status: draftStatusLabel(apiField(draft, 'audit_status', 'AuditStatus') || 'draft'),
    source: sourceLabel(apiField(draft, 'source', 'Source') || 'ai_workbench'),
    sourceSnapshot: Object.keys(sourceSnapshot).length ? sourceSnapshot : (fallback.sourceSnapshot || workbenchSourceSnapshot()),
  })
  previewMode.value = 'mobile'
}

function restoreDraftToWorkbench(draft, options = {}) {
  const rawKeywords = parseJsonArray(draft.raw?.keywords || draft.raw?.Keywords || draft.keywords)
  const sourceSnapshot = parseJsonObject(draft.raw?.source_snapshot || draft.raw?.SourceSnapshot || draft.sourceSnapshot)
  restoringWorkbenchContext = true
  if (draft.brandId) {
    workbench.brandId = String(draft.brandId)
    selectedBrandId.value = Number(draft.brandId)
  }
  workbench.productId = draft.productId ? String(draft.productId) : ''
  Object.assign(editingDraft, {
    ...draft,
    body: ensureDraftConclusionBody(sanitizeGeneratedDraftBody(draft.body), draft.title, draft.summary),
    keywordsText: rawKeywords.length ? rawKeywords.join(', ') : String(draft.keywordsText || ''),
    sourceSnapshot,
  })
  markEditingDraftSaved()
  chatMessages.splice(0, chatMessages.length, ...(draft.conversation || []).map((message, index) => ({
    id: Date.now() + index,
    role: message.role === 'assistant' ? 'ai' : message.role,
    text: message.text || '',
    createdAt: message.created_at || message.createdAt || '',
  })).filter(message => message.text))
  window.setTimeout(() => {
    restoringWorkbenchContext = false
  }, 0)
  previewMode.value = 'mobile'
  setAiStage('draft_generated')
  updateDecisionCard({
    recommendedDirection: editingDraft.title || ideaSession.searchProblem || '',
    reason: '已从母稿列表恢复，可继续编辑或生成渠道内容。',
  })
  if (options.navigate !== false) {
    const id = Number(apiField(draft, 'id', 'ID') || 0)
    window.localStorage?.removeItem(WORKBENCH_RESET_STORAGE_KEY)
    if (id) window.localStorage?.setItem(WORKBENCH_DRAFT_STORAGE_KEY, String(id))
    router.push({ path: menuRouteMap.workbench, query: id ? { draft_id: String(id) } : {} }).catch(() => {})
  } else {
    rememberWorkbenchDraft(apiField(draft, 'id', 'ID'))
  }
  queueWorkbenchChatScrollAfterRender(true)
}

function draftConversationPayload(extraMessages = []) {
  return [...chatMessages, ...extraMessages]
    .filter(message => message && ['user', 'ai', 'assistant'].includes(message.role) && String(message.text || '').trim())
    .map(message => ({
      role: message.role === 'assistant' ? 'ai' : message.role,
      text: String(message.text || '').trim(),
      created_at: message.createdAt || new Date().toISOString(),
    }))
}

function workbenchSourceSnapshot(extra = {}) {
  const brand = selectedWorkbenchBrand.value
  const product = selectedWorkbenchProduct.value
  return {
    brand: brand?.id ? {
      id: brand.id,
      code: brand.code,
      name: brand.name,
      positioning: brand.position,
      target_audience: brand.audience,
      price_band: brand.priceBand,
      tone: brand.tone,
      keywords: brand.keywordGroups?.flatMap(group => group.keywords || []) || [],
      completeness: brand.completeness,
    } : null,
    product: product?.id ? {
      id: product.id,
      code: product.code,
      name: product.name,
      category: product.category,
      selling_points: product.sellingPoints,
      target_audience: product.audience,
      keywords: product.keywords || [],
      sku_count: product.skus?.length || 0,
      completeness: product.completeness,
    } : null,
    skill: INTERNAL_DRAFT_SKILL,
    internal_skill: INTERNAL_DRAFT_SKILL,
    content_type: selectedSkillProfile.value.name,
    content_type_goal: selectedSkillProfile.value.goal,
    hotspot: workbench.hotspot ? {
      id: workbench.hotspot.id,
      title: workbench.hotspot.title,
      platform: workbench.hotspot.platform || workbench.hotspot.source || '',
      heat_score: workbench.hotspot.score || workbench.hotspot.heat_score || '',
    } : null,
    style_template: styleTemplateSourceSnapshot(workbench.styleTemplate),
    prompt: latestWorkbenchUserPrompt.value || String(workbench.prompt || '').trim(),
    brief: ideaSession.brief,
    inferred_slots: {
      search_problem: ideaSession.searchProblem,
      audience: ideaSession.audience,
      scene: ideaSession.scene,
      tone: ideaSession.tone,
    },
    keywords: workbenchKeywords.value.slice(0, 12),
    references: workbenchEvidence.value.map(item => ({ label: item.label, value: item.value })),
    ...extra,
  }
}

function sourceRecordsFromSnapshot(snapshotValue) {
  const snapshot = parseJsonObject(snapshotValue)
  if (!Object.keys(snapshot).length) return []
  const brand = parseJsonObject(snapshot.brand)
  const product = parseJsonObject(snapshot.product)
  const hotspot = parseJsonObject(snapshot.hotspot)
  const styleTemplate = parseJsonObject(snapshot.style_template)
  const inferred = parseJsonObject(snapshot.inferred_slots)
  const keywords = Array.isArray(snapshot.keywords) ? snapshot.keywords : parseJsonArray(snapshot.keywords)
  const records = []
  if (brand.name) {
    records.push({
      label: '品牌',
      title: brand.name,
      desc: [brand.positioning, brand.target_audience, brand.price_band].filter(Boolean).join(' · ') || '已记录品牌资料',
    })
  }
  if (product.name) {
    records.push({
      label: '商品',
      title: product.name,
      desc: [product.category, product.selling_points, product.target_audience].filter(Boolean).join(' · ') || '未指定商品资料',
    })
  }
  records.push({
    label: '内容类型',
    title: snapshot.content_type || contentTypeFromLegacySkill(snapshot.skill) || selectedSkillProfile.value.name,
    desc: [snapshot.content_type_goal, inferred.search_problem || snapshot.brief || snapshot.prompt].filter(Boolean).join(' · ') || '通用 GEO 母稿',
  })
  if (hotspot.title) {
    records.push({
      label: '热点',
      title: hotspot.title,
      desc: [hotspot.platform, hotspot.heat_score ? `热度 ${hotspot.heat_score}` : ''].filter(Boolean).join(' · ') || '已记录引用热点',
    })
  }
  if (styleTemplate.template_name) {
    records.push({
      label: '风格',
      title: styleTemplate.template_name,
      desc: [styleTemplate.platform, styleTemplate.content_type, styleTemplate.prompt_fragment].filter(Boolean).join(' · ') || '已记录参考写作风格',
    })
  }
  if (keywords.length) {
    records.push({
      label: '关键词',
      title: keywords.slice(0, 4).join('、'),
      desc: keywords.slice(4, 12).join('、') || '用于母稿与渠道改写',
    })
  }
  return records.filter(record => record.title)
}

function draftPayload(source = 'manual', extraMessages = []) {
  const existingSnapshot = parseJsonObject(editingDraft.sourceSnapshot)
  return {
    brand_id: workbench.brandId ? Number(workbench.brandId) : undefined,
    product_id: workbench.productId ? Number(workbench.productId) : undefined,
    title: editingDraft.title,
    summary: editingDraft.summary,
    body: editingDraft.body,
    keywords: String(editingDraft.keywordsText || '').split(',').map(s => s.trim()).filter(Boolean),
    content_type: selectedSkillProfile.value.name,
    conversation: draftConversationPayload(extraMessages),
    source_snapshot: { ...workbenchSourceSnapshot(), ...existingSnapshot },
    source,
  }
}

async function persistCurrentDraft(source = 'manual') {
  const id = Number(editingDraft.id || 0)
  const payload = draftPayload(source)
  const draft = id > 0
    ? await updateAiGeoDraft(id, payload)
    : await createAiGeoDraft(payload)
  applyDraftToEditor(draft, {
    title: payload.title,
    summary: payload.summary,
    body: payload.body,
    keywordsText: String(editingDraft.keywordsText || ''),
  })
  rememberWorkbenchDraft(apiField(draft, 'id', 'ID'))
  markEditingDraftSaved()
  return draft
}

function naturalizeTitle(title) {
  const value = String(title || decisionCard.recommendedDirection || ideaSession.searchProblem || '这篇内容适合哪些人？')
    .replace(/一篇看懂选择思路/g, '怎么判断更清楚')
    .replace(/到底/g, '')
    .replace(/\s+/g, ' ')
    .trim()
  if (/适合哪些人[？?]?$/.test(value)) return value.replace(/[？?]?$/, '？')
  return value.length > 28 ? `${value.slice(0, 28)}？` : value
}

function applyWorkbenchDraftEdit(prompt, intent = 'draft_edit_request') {
  if (!hasEditingDraftContent.value) {
    appendSystemStatus('右侧还没有母稿，请先生成母稿。')
    return
  }
  setAiStage('draft_editing')
  if (intent === 'title_edit_request') {
    editingDraft.title = naturalizeTitle(editingDraft.title)
  } else {
    const value = String(prompt || '')
    if (/降低营销|不硬广|自然/.test(value)) {
      editingDraft.body = String(editingDraft.body || '')
        .replace(/强烈推荐/g, '可以作为参考')
        .replace(/必买/g, '值得了解')
        .replace(/完美/g, '相对合适')
    } else if (/更短|精简/.test(value)) {
      editingDraft.body = String(editingDraft.body || '').split('\n').filter(Boolean).slice(0, 12).join('\n')
    } else if (/增强geo|GEO|问答/i.test(value)) {
      const faq = `\n\nFAQ\n1. ${decisionCard.recommendedDirection || editingDraft.title}\n可以先看人群、场景和实际需求是否匹配，再结合品牌资料判断。\n2. 选择时最该注意什么？\n不要只看风格标签，也要看版型、场景和预算是否匹配。`
      if (!/FAQ/.test(editingDraft.body)) editingDraft.body = `${editingDraft.body}${faq}`
    } else {
      editingDraft.summary = editingDraft.summary || `围绕「${editingDraft.title}」提炼适合人群、使用场景和选择理由。`
    }
  }
  editingDraft.status = '草稿'
  updateDecisionCard({ stageLabel: stageLabels.draft_editing })
}

async function generateDraftFromChat(options = {}) {
  loading.action = true
  draftPreviewLoading.value = true
  previewMode.value = 'mobile'
  setAiStage('draft_generating')
  const product = selectedWorkbenchProduct.value
  const skill = INTERNAL_DRAFT_SKILL
  const contentType = selectedSkillProfile.value.name
  const typedPrompt = String(workbench.prompt || '').trim()
  const latestUserPrompt = latestWorkbenchUserPrompt.value
  const selectedOption = options.selectedDirectionOption || selectedDirectionOptionFromPrompt(options.promptOverride || typedPrompt || latestUserPrompt)
  const rawPrompt = selectedOption?.title || options.promptOverride || typedPrompt || latestUserPrompt || decisionCard.recommendedDirection
  const defaultPrompt = `请基于 ${workbenchBrandName('当前资料')}${product ? ` 的 ${product.name}` : ''}，按「${contentType}」生成一篇面向「${selectedSkillProfile.value.goal}」的 GEO 内容母版；缺失的目标人群、场景和口吻由 AI 根据资料合理推断。`
  const prompt = rawPrompt && !isGreetingMessage(rawPrompt) && !isLowSignalIdea(rawPrompt) ? rawPrompt : defaultPrompt
  const referenceArticle = options.referenceArticle || null
  const selectedStyleSnapshot = styleTemplateSourceSnapshot(workbench.styleTemplate)
  const selectedStylePreview = selectedStyleSnapshot ? styleTemplateDetailPreview(workbench.styleTemplate) : null
  const generationSnapshot = workbenchSourceSnapshot({
    draft_brief: selectedOption?.draftBrief || options.draftBrief || workbenchDraftBrief(prompt),
    selected_direction_option: selectedOption ? {
      index: selectedOption.index,
      label: `方向${selectedOption.label}`,
      title: selectedOption.title,
      block: selectedOption.block,
    } : undefined,
    reference_article: referenceArticle?.sourceSnapshot?.reference_article,
    generation_mode: referenceArticle ? 'apply_reference_article' : 'generate_from_brief',
  })
  if (typedPrompt && options.appendPrompt !== false) appendUserMessage(typedPrompt)
  try {
    updateDecisionFromPrompt(prompt, decisionCard.reason)
    const draft = await generateAiGeoDraft({
      brand_id: workbench.brandId ? Number(workbench.brandId) : undefined,
      product_id: workbench.productId ? Number(workbench.productId) : undefined,
      skill,
      content_type: contentType,
      hotspot_id: workbench.hotspot?.id,
      style_template_id: workbench.styleTemplate?.id ? Number(workbench.styleTemplate.id) : undefined,
      conversation: draftConversationPayload(),
      source_snapshot: generationSnapshot,
      prompt: [
        '你是 AI GEO 母稿共创专家，不是普通文章生成器。请根据用户选择的资料、内容类型和聊天中收敛出的 brief 生成一份可作为多渠道源文件的 GEO 内容母版。',
        '当前触发方式：用户已经点击“生成母稿”，这代表用户选择 AI 共创/委托生成。你必须直接写完整母稿，禁止继续追问、要求确认或让用户选择篇幅。',
        referenceArticle ? '最高优先级：左侧聊天中已经产出了一篇范文。你必须沿用这篇范文的标题、结构、语气、叙事角度和核心内容，只做必要的母稿字段整理，不得重新改写成另一篇“选择指南/思路拆解”。' : '',
        selectedOption ? `最高优先级：用户明确选择了方向${selectedOption.label}。本次标题、摘要和正文必须按这个方向生成，不得沿用旧方向或顶部 direction_chip。` : '',
        '如果缺少目标人群、场景、篇幅、口吻、商品或热点，请基于品牌资料、商品资料、内容类型、关键词和对话自行合理推断并补齐，不得阻塞生成。',
        '生成目标：先回答用户真实搜索/AI 问答问题，再自然带出品牌与商品资料；避免空泛品牌介绍和硬广。',
        '重要边界：这是写入右侧母稿编辑器的文章资产，不是聊天回复。禁止出现“好的、明白、我建议、请确认、现在可以生成、母稿使用说明、替换占位符、请补充、你希望、是否可以、还差一个点”等沟通过程或操作说明。',
        '正文必须是给目标用户阅读的完整文章母稿，不是内容方案、脚本方案、渠道拆解或运营 brief。',
        '正文需要包含：核心问题、可被 AI 引用的核心答案、目标用户、品牌/产品定位、使用场景、用户痛点、选择理由、对比逻辑、证据与论点、FAQ、GEO 关键词结构。',
        'FAQ 至少 6 个问题，每个回答要短、准、可引用。',
        '禁止输出小红书、知乎、抖音、公众号、平台分发、短视频脚本、图文笔记、渠道改写建议、发布计划、素材方案等渠道内容。',
        selectedStyleSnapshot ? '参考写作风格：本次必须学习所选风格的结构路径、语气人设和句式手法；只迁移写法，不复制原文句子，不照搬外部事实，品牌/商品资料仍是事实依据。' : '',
        '标题、摘要、正文分别返回；正文只写可发布的母稿正文和必要的 FAQ，不要写创作过程。',
        selectedOption ? `用户选择的方向块：\n${selectedOption.block}` : '',
        `收敛 brief：${ideaSession.brief || prompt}`,
        (selectedOption?.draftBrief || options.draftBrief) ? `结构化 draft_brief：${JSON.stringify(selectedOption?.draftBrief || options.draftBrief)}` : '',
        referenceArticle ? `左侧范文标题：${referenceArticle.title}` : '',
        referenceArticle ? `左侧范文正文（必须作为母稿主体沿用）：\n${referenceArticle.body}` : '',
        `用户原始想法：${prompt}`,
        `本轮对话：${workbenchUserMessages.value.map(msg => msg.text).join(' / ') || prompt}`,
        `内容类型：${contentType}`,
        `内部生成能力：${skill}`,
        selectedStyleSnapshot ? `参考写作风格：${selectedStyleSnapshot.template_name}` : '',
        selectedStylePreview ? `风格结构：${selectedStylePreview.structure}` : '',
        selectedStylePreview ? `风格语气：${selectedStylePreview.tone}` : '',
        selectedStylePreview ? `句式手法：${selectedStylePreview.techniques}` : '',
        selectedStyleSnapshot?.prompt_fragment ? `可复用风格提示词：${selectedStyleSnapshot.prompt_fragment}` : '',
        selectedStyleSnapshot?.negative_rules?.length ? `风格禁止规则：${selectedStyleSnapshot.negative_rules.join('；')}` : '',
        `品牌定位：${selectedWorkbenchBrand.value?.position || '未维护'}`,
        product ? `商品卖点：${product.sellingPoints || product.name}` : '',
        ideaSession.audience ? `目标人群：${ideaSession.audience}` : '',
        ideaSession.scene ? `场景：${ideaSession.scene}` : '',
        ideaSession.tone ? `口吻：${ideaSession.tone}` : '',
        `关键词：${workbenchKeywords.value.join('、') || '未维护'}`,
        workbench.hotspot ? `引用热点：${workbench.hotspot.title}` : '',
	        '输出要求：标题明确、摘要可发布、正文有问题拆解/选择理由/场景建议/FAQ，关键词可供后续使用。只输出文章本身。',
	      ].filter(Boolean).join('\n'),
    })
    if (!referenceArticle && draftLooksLikeClarification(draft)) {
      throw new Error('AI 未返回有效母稿内容，请检查模型输出或重试')
    }
    const draftToApply = referenceArticle
      ? { ...draft, title: referenceArticle.title, summary: referenceArticle.summary, body: referenceArticle.body, keywords: referenceArticle.keywordsText.split(',').map(item => item.trim()).filter(Boolean), source_snapshot: generationSnapshot }
      : selectedOption
        ? { ...draft, title: selectedOption.title, source_snapshot: generationSnapshot }
        : draft
    const aiDraftFallback = {
      title: apiField(draftToApply, 'title', 'Title') || referenceArticle?.title || prompt,
      summary: apiField(draftToApply, 'summary', 'Summary') || referenceArticle?.summary || '',
      body: apiField(draftToApply, 'body', 'Body') || referenceArticle?.body || '',
      keywordsText: referenceArticle?.keywordsText || workbenchKeywords.value.slice(0, 8).join(', '),
      sourceSnapshot: generationSnapshot,
    }
    if (!String(aiDraftFallback.body || '').trim()) {
      throw new Error('AI 未返回有效母稿正文，请重试')
    }
    applyDraftToEditor(draftToApply, aiDraftFallback)
    const generatedDraftId = Number(apiField(draft, 'id', 'ID') || editingDraft.id || 0)
    rememberWorkbenchDraft(generatedDraftId)
    appendSystemStatus('已生成并自动保存草稿。接下来你可以进行精修。')
    setAiStage('draft_generated')
    updateDecisionCard({
      recommendedDirection: selectedOption?.title || prompt,
      reason: selectedOption ? `已按用户选择的方向${selectedOption.label}生成右侧草稿。` : (decisionCard.reason || '已按当前资料生成右侧草稿，可继续局部修改。'),
    })
    if (generatedDraftId) {
      await updateAiGeoDraft(generatedDraftId, draftPayload('ai_workbench'))
      markEditingDraftSaved()
    }
    await loadAiGeoData()
    workbench.prompt = ''
  } catch (error) {
    setAiStage('direction_recommended')
    showToast(error?.message || '母稿生成失败')
  } finally {
    loading.action = false
    draftPreviewLoading.value = false
  }
}
async function saveDraft() {
  if (!hasEditingDraftContent.value) {
    showToast('请先生成或填写母稿内容')
    return
  }
  loading.action = true
  try {
    await persistCurrentDraft(editingDraft.source === '智能生成' ? 'ai_workbench' : 'manual')
    await loadAiGeoData()
    showToast('母稿草稿已保存')
  } catch (error) {
    showToast(error?.message || '母稿保存失败')
  } finally {
    loading.action = false
  }
}
async function saveCurrentDraftForChannelGeneration() {
  if (!hasEditingDraftContent.value) {
    showToast('请先生成或填写母稿内容')
    return null
  }
  loading.action = true
  try {
    const saved = await persistCurrentDraft(editingDraft.source === '智能生成' ? 'ai_workbench' : 'manual')
    const draftId = Number(apiField(saved, 'id', 'ID') || editingDraft.id || 0)
    if (!draftId) {
      showToast('母稿保存失败，无法继续')
      return null
    }
    await loadAiGeoData()
    return {
      ...editingDraft,
      id: draftId,
      rawStatus: editingDraft.rawStatus || 'draft',
      status: editingDraft.status || '草稿',
      channels: editingDraft.channels || [],
    }
  } catch (error) {
    showToast(error?.message || '母稿保存失败')
    return null
  } finally {
    loading.action = false
  }
}
async function submitDraftFlow(options = {}) {
  if (!hasEditingDraftContent.value) {
    showToast('请先生成或填写母稿内容')
    return null
  }
  loading.action = true
  try {
    await persistCurrentDraft(editingDraft.source === '智能生成' ? 'ai_workbench' : 'manual')
    if (!editingDraft.id) {
      showToast('母稿保存失败，无法继续')
      return null
    }
    const draftId = Number(editingDraft.id)
    editingDraft.status = '已通过'
    editingDraft.rawStatus = 'approved'
    const submitSuccessDraft = { ...editingDraft, id: draftId, rawStatus: 'approved', status: '已通过', channels: [] }
    await loadAiGeoData()
    clearWorkbenchDraftState({ clearSelection: true })
    if (!options.silentSuccess) openDraftSubmitSuccess(submitSuccessDraft)
    return submitSuccessDraft
  } catch (error) {
    showToast(error?.message || '母稿保存失败')
    return null
  } finally {
    loading.action = false
  }
}
async function saveDraftAndOpenChannelGeneration() {
  if (loading.action) return
  const savedDraft = await saveCurrentDraftForChannelGeneration()
  if (!savedDraft?.id) return
  showToast('母稿已保存，正在进入渠道内容生成')
  openChannelGenerationConsole(savedDraft)
}
function editDraftInWorkbench(draft) {
  restoreDraftToWorkbench(draft)
}

function canGenerateChannelFromDraft(draft) {
  return Boolean(draft?.id)
}

async function submitDraftFromList(draft) {
  loading.action = true
  try {
    const draftId = Number(draft.id)
    draft.rawStatus = 'approved'
    draft.status = '已通过'
    await loadAiGeoData()
    openDraftSubmitSuccess({ ...draft, id: draftId, rawStatus: 'approved', status: '已通过', channels: draft.channels || [] })
  } catch (error) {
    showToast(error?.message || '母稿保存失败')
  } finally {
    loading.action = false
  }
}
async function generateChannelsForDraft(draft, options = {}) {
  if (!options.silent) {
    openChannelGenerationConsole(draft)
    return
  }
  const target = draft.id ? draft : drafts[0]
  const channel = channelProfiles[0]
  if (!target?.id || !channel?.id) {
    showToast('请先准备母稿和渠道资料')
    return
  }
  loading.action = true
  try {
    if (!Array.isArray(target.channels)) target.channels = []
    const content = await generateAiGeoChannelContent(Number(target.id), {
      channel_id: Number(channel.id),
      title: target.title,
      body: `${target.body}\n\n${channel.name}版本：按渠道语境完成表达适配。`,
    })
    target.channels.push(channelContentFromApi(content))
    target.status = '已生成渠道内容'
    if (!options.silent) showToast('已生成渠道内容')
  } catch (error) {
    if (options.throwOnError) throw error
    showToast(error?.message || '生成渠道内容失败')
  } finally {
    loading.action = false
  }
}

function openChannelGenerationConsole(draft) {
  const target = draft?.id ? draft : drafts.find(item => Number(item.id) === Number(editingDraft.id)) || drafts[0]
  if (!target?.id) {
    showToast('请先保存母稿')
    return
  }
  if (!canGenerateChannelFromDraft(target)) {
    showToast('请先保存母稿')
    return
  }
  selectedDraft.value = target
  setupChannelGenerationNodes(target)
  closeModal()
  router.push({ path: menuRouteMap.channelGeneration, query: { draft_id: String(target.id) } }).catch(() => {})
}

function returnToDrafts() {
  router.push(menuRouteMap.drafts).catch(() => {})
}

function setupChannelGenerationNodes(draft) {
  const generatedByName = new Map((draft.channels || []).map(content => [content.channel, content]))
  const lockedNames = Array.from(generatedByName.keys()).filter(Boolean)
  lockedNames.forEach(name => {
    if (!channelGeneration.selectedChannels.includes(name)) channelGeneration.selectedChannels.push(name)
  })
  channelGeneration.selectedChannels.sort((a, b) => channelGenerationOrder.indexOf(a) - channelGenerationOrder.indexOf(b))
  const platformNames = channelGenerationPlatforms.value.map(platform => platform.name)
  const names = platformNames.length ? platformNames : channelGenerationOrder
  channelGeneration.nodes.splice(0, channelGeneration.nodes.length, ...names.map(name => {
    const existing = generatedByName.get(name)
    if (existing?.id) {
      const pkg = existing.package || parseChannelPackage(existing.raw?.body || existing.raw?.Body || existing.body, name)
      return {
        channelName: name,
        channelId: existing.channelId || channelProfiles.find(channel => channel.name === name)?.id,
        contentId: existing.id,
        contentType: pkg.contentType,
        status: existing.status === '已通过' ? 'completed' : 'edited',
        startedAt: '',
        completedAt: existing.raw?.updated_at || existing.raw?.UpdatedAt || '',
        updatedLabel: existing.raw?.updated_at ? formatGeneratedTime(existing.raw.updated_at) : '已保存',
        contentPayload: pkg.contentPayload,
        assetPayload: pkg.assetPayload,
        geoPayload: pkg.geoPayload,
        riskNotes: pkg.riskNotes,
        raw: existing.raw,
      }
    }
    const channel = channelProfiles.find(item => item.name === name)
    return emptyChannelNode(name, channel?.id)
  }))
  const currentActive = channelGeneration.nodes.find(node => node.channelName === channelGeneration.activeChannel)
  const firstSelected = channelGeneration.nodes.find(node => channelGeneration.selectedChannels.includes(node.channelName))
  channelGeneration.activeChannel = currentActive?.channelName || firstSelected?.channelName || channelGeneration.nodes[0]?.channelName || ''
  channelGeneration.detailTab = 'body'
  channelGeneration.previewMode = 'mobile'
  channelGeneration.status = channelGeneration.nodes.some(node => node.contentId && channelGeneration.selectedChannels.includes(node.channelName)) ? 'partial_completed' : 'not_started'
}

function emptyChannelNode(channelName, channelId) {
  return {
    channelName,
    channelId,
    contentId: 0,
    contentType: channelContentTypeForName(channelName),
    status: 'waiting',
    startedAt: '',
    completedAt: '',
    updatedLabel: '',
    errorMessage: '',
    contentPayload: { title: '', summary: '', body: '', tags: [], extraFields: {} },
    assetPayload: { requiredAssets: [], matchedAssets: [], missingAssets: [], aiGenerateSuggestions: [] },
    geoPayload: { keywords: [], geoSuggestions: [] },
    riskNotes: [],
    raw: null,
  }
}

function toggleGenerationChannel(name) {
  if (channelGeneration.status === 'generating') return
  const index = channelGeneration.selectedChannels.indexOf(name)
  if (index >= 0) {
    channelGeneration.selectedChannels.splice(index, 1)
  } else {
    channelGeneration.selectedChannels.push(name)
    channelGeneration.selectedChannels.sort((a, b) => channelGenerationOrder.indexOf(a) - channelGenerationOrder.indexOf(b))
  }
  setupChannelGenerationNodes(selectedDraft.value)
}

function isGenerationChannelSelected(name) {
  return channelGeneration.selectedChannels.includes(name) || isGenerationChannelLocked(name)
}

function isGenerationChannelLocked(name) {
  const node = channelGeneration.nodes.find(item => item.channelName === name)
  return Boolean(node?.contentId)
}

function toggleChainNodeSelection(name) {
  if (channelGeneration.status === 'generating') return
  if (isGenerationChannelLocked(name)) {
    showToast('已生成的渠道内容不能取消选择')
    return
  }
  const index = channelGeneration.selectedChannels.indexOf(name)
  if (index >= 0) {
    channelGeneration.selectedChannels.splice(index, 1)
  } else {
    channelGeneration.selectedChannels.push(name)
    channelGeneration.selectedChannels.sort((a, b) => channelGenerationOrder.indexOf(a) - channelGenerationOrder.indexOf(b))
  }
}

function selectChannelNode(name) {
  channelGeneration.activeChannel = name
  channelGeneration.detailTab = 'body'
}

function channelCodeByName(name) {
  return channelProfiles.find(item => item.name === name)?.code || ''
}

function generationNodeStatusLabel(status) {
  return {
    waiting: '等待中',
    generating: '生成中',
    completed: '已完成',
    edited: '已编辑',
    need_assets: '需补充素材',
    failed: '生成失败',
    skipped: '已跳过',
  }[status] || '等待中'
}

async function startChannelGeneration() {
  if (!selectedDraft.value?.id) return
  const selectedNodes = channelGeneration.nodes.filter(node => channelGeneration.selectedChannels.includes(node.channelName))
  if (!selectedNodes.length) {
    showToast('请先选择需要生成的平台')
    return
  }
  channelGeneration.status = 'generating'
  loading.action = true
  try {
    for (const node of selectedNodes) {
      if (node.contentId && ['completed', 'edited', 'need_assets'].includes(node.status)) continue
      channelGeneration.activeChannel = node.channelName
      node.status = 'generating'
      node.startedAt = new Date().toISOString()
      const content = await generateAiGeoChannelContent(Number(selectedDraft.value.id), {
        channel_id: Number(node.channelId),
      })
      Object.assign(node, channelNodeFromContent(content))
      if (!Array.isArray(selectedDraft.value.channels)) selectedDraft.value.channels = []
      const listIndex = selectedDraft.value.channels.findIndex(item => Number(item.id) === Number(node.contentId))
      const listItem = channelContentFromApi(content)
      if (listIndex >= 0) selectedDraft.value.channels.splice(listIndex, 1, listItem)
      else selectedDraft.value.channels.push(listItem)
      await wait(160)
    }
    channelGeneration.status = 'completed'
    showToast('本次渠道内容生成完成')
    await loadAiGeoData()
  } catch (error) {
    const node = activeChannelNode.value
    if (node) {
      node.status = 'failed'
      node.errorMessage = error?.message || '生成失败'
    }
    channelGeneration.status = 'failed'
    showToast(error?.message || '渠道内容生成失败')
  } finally {
    loading.action = false
  }
}

function resetChannelGeneration() {
  setupChannelGenerationNodes(selectedDraft.value)
  channelGeneration.status = 'not_started'
}

async function regenerateActiveChannel() {
  const node = activeChannelNode.value
  if (!node) return
  node.contentId = 0
  node.status = 'waiting'
  await startChannelGeneration()
}

async function saveActiveChannel() {
  const node = activeChannelNode.value
  if (!node?.contentId) return
  loading.action = true
  try {
    await saveChannelNode(node)
    showToast('当前平台内容已保存')
  } catch (error) {
    showToast(error?.message || '保存当前平台失败')
  } finally {
    loading.action = false
  }
}

async function saveChannelNode(node) {
  if (!node?.contentId) throw new Error('请先生成渠道内容')
  const saved = await updateAiGeoChannelContent(Number(node.contentId), {
    channel_id: Number(node.channelId),
    title: node.contentPayload.title || activeChannelPreview.value.title,
    body: serializeChannelNode(node),
  })
  Object.assign(node, channelNodeFromContent(saved))
  const listItem = channelContentFromApi(saved)
  const draft = selectedDraft.value || drafts.find(item => Number(item.id) === Number(listItem.draftId))
  if (draft) {
    if (!Array.isArray(draft.channels)) draft.channels = []
    const index = draft.channels.findIndex(item => Number(item.id) === Number(listItem.id))
    if (index >= 0) draft.channels.splice(index, 1, listItem)
    else draft.channels.push(listItem)
  }
  return saved
}

async function saveAllGeneratedChannels() {
  for (const node of channelGeneration.nodes.filter(item => item.contentId)) {
    channelGeneration.activeChannel = node.channelName
    await saveActiveChannel()
  }
  showToast('全部渠道内容已保存')
}

async function saveActiveChannelContent() {
  const node = activeChannelNode.value
  if (!node?.contentId) return
  await saveActiveChannel()
  showToast('渠道内容已保存')
}

function returnAfterChannelGeneration() {
  showToast('渠道内容已生成，可在母稿列表中查看')
  returnToDrafts()
}

function submitChannelEditPrompt(event) {
  if (event?.isComposing) return
  event?.preventDefault()
  if (loading.action || !activeChannelNode.value || !String(channelEditPrompt.value || '').trim()) return
  applyChannelAiEdit()
}

function confirmChannelQuickEdit(action) {
  const node = activeChannelNode.value
  if (!node?.contentPayload?.body) {
    showToast('请先生成渠道内容')
    return
  }
  confirmStandardAction({
    title: '确认 AI 修改',
    icon: '!',
    message: `确认对「${node.channelName}」执行「${action.label}」？`,
    detail: '确认后会直接调用 AI 修改当前平台内容，并自动应用到当前渠道稿；不影响母稿和其他平台。',
    confirmText: '确认修改',
    cancelText: '取消',
    size: 'medium',
  })
    .then(() => applyChannelAiEdit({ promptOverride: action.prompt, autoApply: true, silentUserMessage: true, actionLabel: action.label }))
    .catch(() => {})
}

async function applyChannelAiEdit(options = {}) {
  const node = activeChannelNode.value
  const prompt = String(options.promptOverride || channelEditPrompt.value || '').trim()
  if (!node || !prompt) return
  removeChannelEditorIntro()
  const userMessage = { id: Date.now(), role: 'user', text: options.actionLabel || prompt }
  if (!hasChannelEditIntent(prompt)) {
    if (!options.silentUserMessage) channelEditorMessages.push(userMessage, {
      id: Date.now() + 1,
      role: 'ai',
      text: buildChannelEditGuidance(prompt, node),
    })
    channelEditPrompt.value = ''
    return
  }
  const aiMessage = reactive({
    id: Date.now() + 1,
    role: 'ai',
    text: '正在根据当前平台内容生成修改方案...',
    rawText: '',
    streaming: true,
    pendingEdit: null,
  })
  if (!options.silentUserMessage) channelEditorMessages.push(userMessage, aiMessage)
  loading.action = true
  try {
    await streamAiGeoGatewayInvoke({
      app_code: 'ai-geo',
      app_name: 'AI GEO',
      ai_scenario_code: 'ai_geo_channel_content_editor',
      params: { usage_amount: 1, usage_unit: 'calls', temperature: 0.35, max_tokens: 1400 },
      input: {
        messages: [
          { role: 'system', content: '你是渠道内容 AI 编辑。只修改当前平台渠道内容，不修改母稿和其他平台。必须返回纯 JSON，不要 Markdown，不要解释文本。结构：{"contentPayload":{"title":"...","summary":"...","body":"...","tags":[]},"assetPayload":{},"geoPayload":{},"riskNotes":[],"editorNote":"用中文说明本次改了什么","preview":"用中文给运营看的改稿预览，说明标题、摘要、正文和素材会怎么变"}。' },
          { role: 'user', content: JSON.stringify({ request: prompt, channel: node.channelName, currentPackage: serializeChannelNode(node), masterDraft: selectedDraft.value }, null, 2) },
        ],
      },
    }, {
      onDelta(delta) { aiMessage.rawText += delta },
      onFinal(event) { if (!aiMessage.rawText && event.text) aiMessage.rawText = event.text },
      onError(event) { throw new Error(event.error_message || 'AI 编辑失败') },
    })
    const proposal = buildChannelEditorProposal(node, aiMessage.rawText, prompt)
    if (proposal) {
      if (options.autoApply) {
        applyChannelEditorResult(node, proposal.parsed, options.actionLabel || prompt, { silentMessage: true })
        node.status = 'edited'
        node.updatedLabel = '刚刚编辑'
        await saveActiveChannel()
        showToast(`已完成「${options.actionLabel || 'AI 修改'}」`)
      } else {
        aiMessage.text = proposal.previewText
        aiMessage.pendingEdit = proposal
      }
    } else {
      aiMessage.text = `这次没有形成可直接应用的修改包。\n\n${cleanAssistantText(aiMessage.rawText) || '请换一种说法再发送一次。'}`
      if (options.autoApply) showToast('AI 没有返回可应用的修改')
    }
    channelEditPrompt.value = ''
  } catch (error) {
    aiMessage.text = `AI 编辑失败：${error?.message || '请稍后重试'}`
    showToast(error?.message || 'AI 编辑失败')
  } finally {
    aiMessage.streaming = false
    loading.action = false
  }
}

function hasChannelEditIntent(text) {
  const compact = String(text || '').replace(/\s/g, '').toLowerCase()
  if (!compact || isGreetingMessage(compact)) return false
  if (compact.length <= 6 && !/(改|调|换|删|加|补|降|升|优|标题|正文|摘要|图|标签|关键词)/.test(compact)) return false
  return /(修改|改成|改为|调整|优化|替换|删除|增加|新增|补充|强化|弱化|降低|提高|重写|换个|更自然|更柔和|更专业|更短|更长|标题|正文|摘要|关键词|话题|标签|素材|图片|封面|营销感|口吻|语气|结构|表达|平台|小红书|知乎|公众号|抖音|微博|百家号|独立站)/.test(compact)
}

function buildChannelEditGuidance(prompt, node) {
  const channel = node?.channelName || '当前平台'
  const title = node?.contentPayload?.title || '当前渠道内容'
  if (isGreetingMessage(prompt)) {
    return `你好，我在。现在选中的是「${channel}」的渠道内容：${title}。\n\n你可以直接告诉我想怎么改，比如：\n1. 标题更柔和，不要太营销；\n2. 正文更像真实分享，减少品牌自夸；\n3. 补充适合人群、使用场景或 GEO 关键词。\n\n我会先生成修改建议，确认后再应用到当前平台。`
  }
  return `我还没有拿到明确的修改方向，所以先不进入改稿流程。\n\n请补一句具体目标，例如「把标题改得更柔和」「正文增加通勤场景」「降低营销感」「补充小红书话题标签」。我会基于「${channel}」当前内容生成可确认的修改方案。`
}

function removeChannelEditorIntro() {
  const introIndex = channelEditorMessages.findIndex(message => message.kind === 'intro')
  if (introIndex >= 0) channelEditorMessages.splice(introIndex, 1)
}

async function applyPendingChannelEdit(message) {
  const node = activeChannelNode.value
  const proposal = message?.pendingEdit
  if (!node || !proposal) return
  loading.action = true
  try {
    applyChannelEditorResult(node, proposal.parsed, proposal.prompt)
    node.status = 'edited'
    node.updatedLabel = '刚刚编辑'
    message.pendingEdit = null
    message.text = `${proposal.previewText}\n\n已应用到「${node.channelName}」。`
    await saveActiveChannel()
  } catch (error) {
    showToast(error?.message || '应用修改失败')
  } finally {
    loading.action = false
  }
}

function buildChannelEditorProposal(node, text, prompt) {
  const parsed = parseJsonObject(extractJsonText(text))
  const contentPayload = parseJsonObject(parsed.contentPayload || parsed.content_payload)
  const assetPayload = parseJsonObject(parsed.assetPayload || parsed.asset_payload)
  const geoPayload = parseJsonObject(parsed.geoPayload || parsed.geo_payload)
  const hasPatch = [contentPayload, assetPayload, geoPayload].some(item => Object.keys(item).length)
  if (!hasPatch) return null
  const preview = parsed.preview || parsed.editorPreview || parsed.editor_preview || parsed.editorNote || parsed.editor_note || `已按「${prompt}」形成修改方案。`
  const parts = [
    '已生成当前平台的修改方案，先不自动应用。',
    preview,
  ]
  if (contentPayload.title) parts.push(`标题：${contentPayload.title}`)
  if (contentPayload.summary) parts.push(`摘要：${contentPayload.summary}`)
  if (contentPayload.tags || contentPayload.hashtags) parts.push(`话题/关键词：${arrayOrObjectText(contentPayload.tags || contentPayload.hashtags)}`)
  const riskNotes = parsed.riskNotes || parsed.risk_notes
  if (Array.isArray(riskNotes) && riskNotes.length) parts.push(`风险提示：${arrayOrObjectText(riskNotes)}`)
  return {
    channelName: node.channelName,
    prompt,
    parsed,
    previewText: parts.filter(Boolean).join('\n\n'),
  }
}

function applyChannelEditorResult(node, parsed, prompt, options = {}) {
  const contentPayload = parseJsonObject(parsed.contentPayload || parsed.content_payload)
  const assetPayload = parseJsonObject(parsed.assetPayload || parsed.asset_payload)
  const geoPayload = parseJsonObject(parsed.geoPayload || parsed.geo_payload)
  if (Object.keys(contentPayload).length) Object.assign(node.contentPayload, contentPayload)
  if (Object.keys(assetPayload).length) Object.assign(node.assetPayload, assetPayload)
  if (Object.keys(geoPayload).length) Object.assign(node.geoPayload, geoPayload)
  if (Array.isArray(parsed.riskNotes || parsed.risk_notes)) node.riskNotes = parsed.riskNotes || parsed.risk_notes
  const note = parsed.editorNote || parsed.editor_note || `已按「${prompt}」更新当前平台内容。`
  if (!options.silentMessage) channelEditorMessages.push({ id: Date.now() + 2, role: 'ai', text: note })
}

function normalizeJsonText(text) {
  return String(text || '').trim().replace(/^```json/i, '').replace(/^```/, '').replace(/```$/, '').trim()
}

function extractJsonText(text) {
  const normalized = normalizeJsonText(text)
  if (!normalized) return ''
  const start = normalized.indexOf('{')
  const end = normalized.lastIndexOf('}')
  if (start >= 0 && end > start) return normalized.slice(start, end + 1)
  return normalized
}

function cleanAssistantText(text) {
  const parsed = parseJsonObject(extractJsonText(text))
  if (parsed.preview || parsed.editorNote || parsed.editor_note) return [parsed.preview, parsed.editorNote || parsed.editor_note].filter(Boolean).join('\n\n')
  return normalizeJsonText(text).replace(/\\n/g, '\n').replace(/[{}"]/g, '').trim()
}

function messageParagraphs(message) {
  const text = String(message?.text || '').trim()
  if (!text) return []
  return text.split(/\n{2,}|\n/).map(item => item.trim()).filter(Boolean)
}
function openChannelEditor(draft, channel) {
  selectedDraft.value = draft
  Object.assign(selectedChannel, channel, { previewOnly: false })
  drawer.type = 'channelEditor'
  drawer.title = `${channel.channel} 内容编辑`
  channelPreviewMode.value = 'edit'
}

function openChannelPreview(draft, channel) {
  selectedDraft.value = draft
  Object.assign(selectedChannel, channel, { previewOnly: true })
  drawer.type = 'channelEditor'
  drawer.title = `${channel.channel} 内容预览`
  channelPreviewMode.value = 'mobile'
}

function openChannelContentEditor(draft, channel) {
  const target = draft?.id ? draft : drafts.find(item => Number(item.id) === Number(channel?.draftId))
  if (!target?.id) {
    showToast('请先选择母稿')
    return
  }
  selectedDraft.value = target
  if (channel?.channel && !channelGeneration.selectedChannels.includes(channel.channel)) {
    channelGeneration.selectedChannels.push(channel.channel)
    channelGeneration.selectedChannels.sort((a, b) => channelGenerationOrder.indexOf(a) - channelGenerationOrder.indexOf(b))
  }
  setupChannelGenerationNodes(target)
  channelGeneration.activeChannel = channel.channel
  channelGeneration.detailTab = 'body'
  channelGeneration.previewMode = 'mobile'
  closeDrawer()
  router.push({ path: menuRouteMap.channelGeneration, query: { draft_id: String(target.id) } }).catch(() => {})
}
async function confirmChannel(channel) {
  const contentId = Number(channel?.id)
  if (!contentId) {
    showToast('请先生成渠道内容')
    return false
  }
  channel.status = '已确认'
  channel.rawAuditStatus = 'approved'
  syncSelectedChannel(channel)
  showToast('渠道内容已确认')
  return true
}

async function confirmSelectedChannel() {
  const saved = await persistSelectedChannel()
  if (!saved) return
  await confirmChannel(selectedChannel)
}

function isChannelConfirmed(channel) {
  return ['已确认', '已通过'].includes(channel?.status) || ['approved', 'pass'].includes(channel?.rawAuditStatus)
}

function isChannelInPublishPlan(channel) {
  const contentId = Number(channel?.id)
  return channel?.status === '已加入发布计划'
    || ['scheduled', 'publishing', 'published'].includes(channel?.publishStatus)
    || (contentId > 0 && plans.some(plan => Number(plan.channelContentId) === contentId))
}

function canShowChannelConfirm(channel) {
  return canManageChannelContent.value && !channel?.previewOnly && !isChannelConfirmed(channel) && !isChannelInPublishPlan(channel)
}

function canShowChannelConfirmAndPlan(channel) {
  return canManageChannelContent.value && canManagePublishPlan.value && !channel?.previewOnly && !isChannelConfirmed(channel) && !isChannelInPublishPlan(channel)
}

function canShowChannelAddPlan(channel) {
  return canManagePublishPlan.value && !channel?.previewOnly && isChannelConfirmed(channel) && !isChannelInPublishPlan(channel)
}

function channelStatusBadgeClass(channel) {
  if (isChannelInPublishPlan(channel)) return 'success'
  if (isChannelConfirmed(channel)) return 'success'
  return 'warning'
}

function channelDisplayStatus(channel) {
  if (isChannelInPublishPlan(channel)) return '已加入发布计划'
  return channel?.status || '待确认'
}

function draftChannelStats(draft) {
  const channels = Array.isArray(draft?.channels) ? draft.channels : []
  return channels.reduce((stats, channel) => {
    stats.total += 1
    if (isChannelInPublishPlan(channel)) {
      stats.planned += 1
    } else if (isChannelConfirmed(channel)) {
      stats.confirmed += 1
    } else {
      stats.pending += 1
    }
    return stats
  }, { total: 0, pending: 0, confirmed: 0, planned: 0, rejected: 0 })
}

function draftChannelSummaryText(draft) {
  const stats = draftChannelStats(draft)
  if (!stats.total) return '未生成渠道稿'
  const activeCount = stats.pending + stats.confirmed + stats.planned
  if (stats.planned === stats.total) return `${stats.total} 篇均已进计划`
  if (stats.pending === stats.total) return `${stats.total} 篇待确认`
  return `${stats.total} 篇，${activeCount} 篇可继续推进`
}

function draftChannelStatItems(draft) {
  const stats = draftChannelStats(draft)
  return [
    { key: 'total', label: '总数', value: stats.total, tone: 'neutral' },
    { key: 'pending', label: '待确认', value: stats.pending, tone: 'warning' },
    { key: 'confirmed', label: '已确认', value: stats.confirmed, tone: 'info' },
    { key: 'planned', label: '发布计划', value: stats.planned, tone: 'success' },
  ].filter(item => item.key === 'total' || item.value > 0)
}

function draftChannelPlatformText(draft) {
  const names = [...new Set((draft?.channels || []).map(channel => channel.channel).filter(Boolean))]
  if (!names.length) return ''
  if (names.length <= 3) return names.join('、')
  return `${names.slice(0, 3).join('、')} 等 ${names.length} 个平台`
}

function optimizeChannel(type) { selectedChannel.body += `\n\nAI 局部优化：${type}。`; if (type === '生成话题标签') selectedChannel.tags = '#通勤穿搭 #法式穿搭 #小个子穿搭'; showToast(type + '完成') }

async function regenerateChannel() {
  selectedChannel.body = `${selectedChannel.body || ''}\n\n已基于最新母稿重新生成渠道表达。`
  const saved = await persistSelectedChannel()
  if (saved) showToast('渠道内容已重新生成并保存')
}

async function persistSelectedChannel() {
  const contentId = Number(selectedChannel?.id)
  if (!contentId) {
    showToast('请先生成渠道内容')
    return null
  }
  loading.action = true
  try {
    const saved = await updateAiGeoChannelContent(contentId, {
      channel_id: Number(selectedChannel.channelId),
      title: selectedChannel.title,
      body: selectedChannel.body,
    })
    Object.assign(selectedChannel, channelContentFromApi(saved))
    syncSelectedChannel(selectedChannel)
    return saved
  } catch (error) {
    showToast(error?.message || '渠道内容保存失败')
    return null
  } finally {
    loading.action = false
  }
}

function syncSelectedChannel(channel) {
  const draft = selectedDraft.value || drafts.find(item => item.channels?.some(content => Number(content.id) === Number(channel.id)))
  const target = draft?.channels?.find(content => Number(content.id) === Number(channel.id))
  if (target) Object.assign(target, channel)
}
async function addChannelToPlan(draft, channel) {
  const profile = channelProfiles.find(item => item.name === channel.channel) || channelProfiles[0]
  if (!profile?.id) {
    showToast('请先配置可发布渠道')
    return
  }
  const channelContentId = Number(channel?.id || 0)
  if (!channelContentId) {
    showToast('请先生成并确认该渠道内容，再加入发布计划')
    return
  }
  loading.action = true
  try {
    await createPublishPlanForChannelContent({
      channelContentId,
      channelId: Number(profile.id),
      scheduledAt: new Date(`${planDate.value || '2026-05-20'}T18:00:00+08:00`).toISOString(),
      method: profile.method || 'manual',
      level: profile.level || 'manual',
    })
    channel.status = '已加入发布计划'
    await loadAiGeoData()
    showToast('已加入发布计划')
  } catch (error) {
    showToast(error?.message || '加入发布计划失败')
  } finally {
    loading.action = false
  }
}

async function createPublishPlanForChannelContent({ channelContentId, channelId, scheduledAt, method, level }) {
  return createAiGeoPublishPlan({
    channel_content_id: Number(channelContentId),
    channel_id: Number(channelId),
    scheduled_at: scheduledAt,
    publish_method: method,
    automation_level: level,
  })
}

async function saveChannelAndCreatePlan() {
  const selectedCards = channelPlanCards.value.filter(item => channelPlanForm.selectedChannels.includes(item.channelName))
  if (!selectedCards.length) {
    showToast('请先选择已生成的渠道内容')
    return
  }
  const missingProfile = selectedCards.find(item => !item.channelId)
  if (missingProfile) {
    showToast('请先配置可发布渠道')
    return
  }
  loading.action = true
  try {
    for (const item of selectedCards) {
      const node = channelGeneration.nodes.find(channelNode => channelNode.channelName === item.channelName)
      if (!node?.contentId) continue
      channelGeneration.activeChannel = node.channelName
      const saved = await saveChannelNode(node)
      await createPublishPlanForChannelContent({
        channelContentId: Number(apiField(saved, 'id', 'ID') || node.contentId),
        channelId: item.channelId,
        scheduledAt: fromLocalDateTime(item.scheduledAt),
        method: item.method,
        level: item.level,
      })
    }
    closeModal()
    await loadAiGeoData()
    showToast('已保存并添加到发布计划')
  } catch (error) {
    showToast(error?.message || '保存并添加发布计划失败')
  } finally {
    loading.action = false
  }
}

async function confirmAndAddChannelToPlan(draft, channel) {
  if (!isChannelConfirmed(channel)) {
    const confirmed = await confirmChannel(channel)
    if (!confirmed) return
  }
  if (!isChannelInPublishPlan(channel)) {
    await addChannelToPlan(draft || selectedDraft.value || drafts[0], channel)
  }
}

async function addSelectedChannelToPlan() {
  await addChannelToPlan(selectedDraft.value || drafts[0], selectedChannel)
}

async function confirmAndAddSelectedChannelToPlan() {
  const saved = await persistSelectedChannel()
  if (!saved) return
  await confirmAndAddChannelToPlan(selectedDraft.value || drafts[0], selectedChannel)
}

async function createPlan() {
  const draft = drafts.find(d => d.id === Number(newPlanDraftId.value)) || drafts[0]
  const channel = channelProfiles.find(c => c.name === newPlanChannel.value) || channelProfiles[0]
  if (!draft || !channel?.id) {
    showToast('请先准备母稿和渠道')
    return
  }
  const channelContent = (draft.channels || []).find(item => Number(item.channelId) === Number(channel.id))
  if (!channelContent?.id) {
    showToast(`请先生成「${channel.name}」渠道内容，再创建发布计划`)
    return
  }
  loading.action = true
  try {
    await createAiGeoPublishPlan({
      channel_content_id: Number(channelContent.id),
      channel_id: Number(channel.id),
      scheduled_at: new Date(`${newPlanTime.value}:00+08:00`).toISOString(),
      publish_method: newPlanMethod.value,
      automation_level: newPlanLevel.value,
    })
    closeModal()
    await loadAiGeoData()
    showToast('发布计划已创建')
  } catch (error) {
    showToast(error?.message || '发布计划创建失败')
  } finally {
    loading.action = false
  }
}
function canPreCheckPlan(plan) {
  return plan?.rawStatus === 'scheduled'
}

function canCompletePlan(plan) {
  return ['scheduled', 'publishing'].includes(plan?.rawStatus) && plan?.publishStatus !== '已发布'
}

async function preCheckPlan(plan) {
  if (plan.accountStatus === '阻断') {
    showToast('账号状态阻断，请先处理账号授权')
    return
  }
  if (plan.materialStatus !== '完整') {
    showToast('素材或渠道内容不完整，请补齐后重试')
    return
  }
  await updatePlanStatus(plan, { status: 'publishing' }, '前置检查完成，任务已进入发布中')
}
function completePlan(plan) {
  if (plan.materialStatus !== '完整' || ['阻断', '待授权'].includes(plan.accountStatus)) {
    showToast('任务仍被阻断，无法完成发布')
    return
  }
  markPublished(plan)
}
function viewPublishMaterial(plan) { showToast(`打开 ${plan.channel} 发布素材包`) }
async function adjustPlan(plan) {
  const nextTime = '20:00'
  await updatePlanSchedule(plan, nextTime, `已调整计划时间为 ${nextTime}`)
}
async function markPublished(plan) {
  await updatePlanStatus(plan, { status: 'published', published_url: plan.link || 'https://example.com/post' }, '已标记为已发布')
}
async function fillPublishLink(plan) {
  await updatePlanStatus(plan, { status: 'published', published_url: 'https://example.com/published-link' }, '发布链接已回填')
}
async function markFailed(plan) {
  await updatePlanStatus(plan, { status: 'failed', fail_reason: '人工标记失败' }, '已标记失败，等待重试或人工接管')
}
async function updatePlanSchedule(plan, time, successMessage) {
  if (!plan?.id || !plan?.channelContentId || !plan?.channelId) {
    showToast('发布计划缺少渠道内容，无法调整')
    return
  }
  loading.action = true
  try {
    const updated = await updateAiGeoPublishPlan(Number(plan.id), {
      channel_content_id: Number(plan.channelContentId),
      channel_id: Number(plan.channelId),
      scheduled_at: new Date(`${plan.date || planDate.value}T${time}:00+08:00`).toISOString(),
      publish_method: plan.method,
      automation_level: plan.level,
    })
    plan.time = time
    plan.date = String(updated.scheduled_at || updated.ScheduledAt || plan.date).slice(0, 10)
    showToast(successMessage)
  } catch (error) {
    showToast(error?.message || '发布计划调整失败')
  } finally {
    loading.action = false
  }
}
function openProductDrawer(product) { Object.keys(selectedProduct).forEach(k => delete selectedProduct[k]); Object.assign(selectedProduct, product); drawer.type = 'product'; drawer.title = '商品资料卡'; productTab.value = '公共资料' }
async function addCompetitor() {
  if (!selectedProduct.id) {
    showToast('请先选择商品')
    return
  }
  loading.action = true
  try {
    const created = await createAiGeoCompetitor({
      product_id: selectedProduct.id,
      brand_name: '新增竞品',
      product_name: '竞品商品',
      price_text: '待录入',
      point: '待录入',
      difference: '待分析',
      angle: '待生成',
      link_url: 'https://example.com',
    })
    selectedProduct.competitors.push(competitorFromApi(created))
    await loadAiGeoData()
    showToast('已新增竞品信息')
  } catch (error) {
    showToast(error?.message || '新增竞品失败')
  } finally {
    loading.action = false
  }
}

async function updatePlanStatus(plan, payload, successText) {
  if (!plan?.id) {
    showToast('发布计划缺少 ID，无法更新状态')
    return
  }
  loading.action = true
  try {
    const updated = await updateAiGeoPublishPlanStatus(Number(plan.id), payload)
    plan.rawStatus = updated.status
    plan.status = publishStatusLabel(updated.status)
    plan.publishStatus = publishStatusLabel(updated.status)
    plan.link = updated.published_url || plan.link || ''
    await loadAiGeoData()
    showToast(successText)
  } catch (error) {
    showToast(error?.message || '发布计划状态更新失败')
  } finally {
    loading.action = false
  }
}

function parseJsonArray(value) {
  if (Array.isArray(value)) return value
  if (!value) return []
  try {
    const parsed = JSON.parse(value)
    return Array.isArray(parsed) ? parsed : []
  } catch {
    return []
  }
}

function parseJsonObject(value) {
  if (value && typeof value === 'object' && !Array.isArray(value)) return value
  if (!value) return {}
  try {
    const parsed = JSON.parse(value)
    return parsed && typeof parsed === 'object' && !Array.isArray(parsed) ? parsed : {}
  } catch {
    return {}
  }
}

function skuStatusLabel(status) {
  const map = {
    active: '在售',
    inactive: '停用',
    unknown: '未知',
    in_stock: '在售',
    out_of_stock: '缺货',
  }
  return map[status] || status || '未知'
}

function channelKey(code, name) {
  const raw = `${code || ''} ${name || ''}`.toLowerCase()
  if (raw.includes('xiaohongshu') || raw.includes('小红书')) return 'xiaohongshu'
  if (raw.includes('douyin') || raw.includes('tiktok') || raw.includes('抖音')) return 'douyin'
  if (raw.includes('wechat') || raw.includes('weixin') || raw.includes('微信')) return 'wechat'
  if (raw.includes('zhihu') || raw.includes('知乎')) return 'zhihu'
  if (raw.includes('weibo') || raw.includes('微博')) return 'weibo'
  if (raw.includes('baijia') || raw.includes('baidu') || raw.includes('百家') || raw.includes('百度')) return 'baijiahao'
  return 'website'
}

function channelDescription(code, name) {
  return {
    website: '品牌自有站点 / SEO 内容入口',
    xiaohongshu: '生活方式与种草图文平台',
    douyin: '短视频与图文内容平台',
    wechat: '公众号长图文内容平台',
    zhihu: '问答与专栏内容平台',
    weibo: '开放社交媒体平台',
    baijiahao: '百度内容生态平台',
  }[channelKey(code, name)] || '渠道资料'
}

function channelAdminUrl(code, fallback) {
  return {
    website: fallback || '',
    xiaohongshu: 'https://creator.xiaohongshu.com',
    douyin: 'https://creator.douyin.com',
    wechat: 'https://mp.weixin.qq.com',
    zhihu: 'https://www.zhihu.com/creator',
    weibo: 'https://weibo.com',
    baijiahao: 'https://baijiahao.baidu.com',
  }[channelKey(code, '')] || fallback || ''
}

function publishModeLabel(mode) {
  return {
    api_auto: 'API 自动发布',
    api_draft_manual_confirm: 'API 草稿 + 人工确认',
    agent_manual_confirm: 'Agent 执行 + 人工确认',
    manual: '人工发布',
  }[mode] || mode || '人工发布'
}

function channelAccessStatus(status) {
  if (status === 'active') return '可发布'
  if (status === 'pending_auth') return '待授权'
  if (status === 'inactive') return '未启用'
  return status || '未配置'
}

function accountAuthLabel(authStatus, publishStatus) {
  if (authStatus === 'authorized' && publishStatus === 'available') return '已授权'
  if (authStatus === 'authorized') return '已授权'
  if (authStatus === 'pending') return '待授权'
  return '未授权'
}

function channelSkill(code, name) {
  return {
    xiaohongshu: '小红书图文发布 Skill',
    douyin: '抖音短视频发布 Skill',
    wechat: '微信公众号图文发布 Skill',
    zhihu: '知乎回答发布 Skill',
    weibo: '微博短帖发布 Skill',
    baijiahao: '百家号图文发布 Skill',
    website: '独立站 SEO 发布 Skill',
  }[channelKey(code, name)] || ''
}

function channelRisk(code, name) {
  return {
    xiaohongshu: '平台节奏控制 / 草稿人工确认',
    douyin: '视频素材校验 / 发布频率控制',
    wechat: '草稿箱接口 / 发布前确认',
    zhihu: '问答语境检查 / 敏感词检查',
    weibo: '话题与敏感词检查',
    baijiahao: '原创检测 / 接口校验',
    website: 'Sitemap 更新 / SEO 字段校验',
  }[channelKey(code, name)] || '发布前校验 + 异常人工接管'
}

function sourceLabel(source) {
  if (source === 'ai_workbench') return '智能生成'
  if (source === 'manual') return '人工创作'
  return source || '人工创作'
}

function draftStatusLabel(status) {
  return {
    draft: '草稿',
    pending: '已通过',
    approved: '已通过',
    rejected: '已通过',
  }[status] || status || '草稿'
}

function draftStatusTone(status) {
  if (status === '已通过' || status === 'approved') return 'success'
  return 'neutral'
}

function channelContentStatusLabel(status) {
  return '已确认'
}

function publishStatusLabel(status) {
  return {
    scheduled: '已排期',
    publishing: '发布中',
    published: '已发布',
    failed: '发布失败',
    cancelled: '已取消',
  }[status] || status || '已排期'
}
</script>

<style src="../styles.css"></style>
