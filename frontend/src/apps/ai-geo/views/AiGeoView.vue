<template>
  <div class="geo-growth-page">
    <main class="main">
      <div v-if="loading.page" class="page-loading">正在加载 AI GEO 业务数据...</div>
      <header class="topbar">
        <div>
          <h2>{{ activeTitle }}</h2>
          <p>{{ activeSubtitle }}</p>
        </div>
      </header>

      <section v-if="activeMenu === 'overview'" class="page overview-page">
        <div class="mode-strip">
          <div class="mode-card">
            <span>工作模式</span>
            <strong>{{ modeConfig.workMode }}</strong>
          </div>
          <div class="mode-card">
            <span>母稿审核模式</span>
            <strong>{{ modeConfig.draftAuditMode }}</strong>
          </div>
          <div class="mode-card">
            <span>渠道内容审核模式</span>
            <strong>{{ modeConfig.channelAuditMode }}</strong>
          </div>
        </div>

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
              <h3>按资料和 Skill 生成有指向性的 GEO 文章</h3>
              <p>直接说想法也可以；我会基于品牌、商品、Skill 和热点资料，把模糊想法整理成可生成的 GEO 母稿。</p>
            </div>

            <div v-if="!hasStartedWorkbenchChat" class="context-strip setup-strip">
              <div class="context-card strong">
                <span>品牌</span>
                <strong>{{ selectedWorkbenchBrand.name }}</strong>
                <p>{{ selectedWorkbenchBrand.position || '未选择品牌定位' }}</p>
                <select v-model="workbench.brandId">
                  <option value="">请选择</option>
                  <option v-for="b in brands" :value="b.id" :key="b.id">{{ b.name }}</option>
                </select>
              </div>
              <div class="context-card">
                <span>商品</span>
                <strong>{{ selectedWorkbenchProduct?.name || '未指定商品' }}</strong>
                <p>{{ selectedWorkbenchProduct?.sellingPoints || '可先用品牌资料自由生成' }}</p>
                <select v-model="workbench.productId">
                  <option value="">请选择</option>
                  <option v-for="p in selectedWorkbenchBrand.products" :value="p.id" :key="p.id">{{ p.name }}</option>
                </select>
              </div>
              <div class="context-card">
                <span>Skill</span>
                <strong>{{ selectedSkillProfile.name }}</strong>
                <p>{{ selectedSkillProfile.goal }} · {{ selectedSkillProfile.output }}</p>
                <select v-model="workbench.skill">
                  <option value="">请选择</option>
                  <option v-for="s in skills" :key="s">{{ s }}</option>
                </select>
              </div>
              <div class="context-card">
                <span>热点</span>
                <strong>{{ workbench.hotspot?.title || '不引用热点' }}</strong>
                <p>{{ workbench.hotspot?.summary || '可在对话中再决定是否借势' }}</p>
                <select v-model="selectedHotspotId">
                  <option value="">请选择</option>
                  <option v-for="h in hotspots" :key="h.id" :value="h.id">{{ h.title }} · {{ h.platform }}</option>
                  <option value="__more__">热点库 / 手动添加…</option>
                </select>
              </div>
            </div>

            <div v-else class="conversation-context-bar">
              <span>品牌 <strong>{{ selectedWorkbenchBrand.name }}</strong></span>
              <span>商品 <strong>{{ selectedWorkbenchProduct?.name || '未指定' }}</strong></span>
              <span>Skill <strong>{{ selectedSkillProfile.name }}</strong></span>
              <span>热点 <strong>{{ workbench.hotspot?.title || '未引用' }}</strong></span>
            </div>

            <div v-if="!hasStartedWorkbenchChat" class="evidence-list">
              <div v-for="item in workbenchEvidence" :key="item.label" class="evidence-item">
                <span>{{ item.label }}</span>
                <strong>{{ item.value }}</strong>
              </div>
            </div>

            <div v-if="hasStartedWorkbenchChat" class="chat-log geo-dialogue">
              <div v-for="msg in chatMessages" :key="msg.id" :class="['bubble', msg.role]">
                <div v-if="msg.role === 'ai'" class="message-rich" v-html="formatChatMessage(msg.text)"></div>
                <p v-else>{{ msg.text }}</p>
                <span v-if="msg.streaming" class="stream-cursor"></span>
              </div>
            </div>
            <div class="chat-input">
              <textarea
                v-model="workbench.prompt"
                :placeholder="generationPlaceholder"
                @keydown.enter.exact.prevent="sendWorkbenchMessage"
              ></textarea>
              <div class="chat-input-actions">
                <button class="btn send-btn" :disabled="!hasWorkbenchInput || chatStreaming" @click="sendWorkbenchMessage">
                  {{ chatStreaming ? '输出中' : '发送' }}
                </button>
                <button v-if="canGenerateDraft && hasStartedWorkbenchChat" class="btn primary" :disabled="loading.action || !workbenchReadiness.ready" @click="generateDraftFromChat">
                  {{ loading.action ? '生成中' : (workbenchReadiness.ready ? '生成母稿' : '继续沟通') }}
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
              <div class="segmented">
                <button :class="{ active: previewMode === 'edit' }" @click="previewMode = 'edit'">编辑</button>
                <button :class="{ active: previewMode === 'pc' }" @click="previewMode = 'pc'">PC 预览</button>
                <button :class="{ active: previewMode === 'mobile' }" @click="previewMode = 'mobile'">手机预览</button>
              </div>
            </div>
            <div v-if="previewMode === 'edit'" class="draft-editor">
              <div class="draft-hero">
                <div class="draft-cover-col">
                  <div v-if="editingDraft.coverImage" class="cover-frame has-image">
                    <img :src="editingDraft.coverImage" alt="封面预览" />
                    <div class="cover-frame-actions">
                      <button type="button" class="cover-icon-btn" title="更换封面" @click="pickCoverImage">+</button>
                      <button type="button" class="cover-icon-btn danger" title="移除封面" @click="editingDraft.coverImage = ''">×</button>
                    </div>
                  </div>
                  <button v-else type="button" class="cover-frame empty" title="上传封面" @click="pickCoverImage">
                    <span class="cover-icon-btn solo">+</span>
                    <span class="cover-empty-hint">封面图 <em>选填</em></span>
                  </button>
                </div>
                <div class="draft-meta-col">
                  <label>标题<input v-model="editingDraft.title" /></label>
                  <label>摘要<textarea v-model="editingDraft.summary" class="summary-editor" rows="2"></textarea></label>
                  <label>关键词<input v-model="editingDraft.keywordsText" placeholder="多个关键词用逗号分隔" /></label>
                </div>
              </div>
              <label>正文<textarea class="body-editor" v-model="editingDraft.body"></textarea></label>
            </div>
            <PreviewPane v-else :mode="previewMode" :title="editingDraft.title" :summary="editingDraft.summary" :body="editingDraft.body" :cover-image="editingDraft.coverImage" />
            <div class="editor-actions">
              <button v-if="canManageDraft" class="btn ghost" @click="saveDraft">保存草稿</button>
              <button v-if="canManageDraft" class="btn ghost" @click="generateDraftAudit(editingDraft)">AI审核建议</button>
              <button v-if="canManageDraft" class="btn primary" @click="submitDraftFlow">提交</button>
              <button v-if="canManageChannelContent" class="btn dark" @click="generateChannelsForDraft(editingDraft)">生成渠道版本</button>
            </div>
            <div v-if="draftAuditPanel.summary || draftAuditPanel.items.length" class="audit-result">
              <div class="audit-result-head">
                <span :class="['badge', draftAuditPanel.passed ? 'success' : 'warning']">{{ draftAuditPanel.risk }}</span>
                <strong>{{ draftAuditPanel.summary }}</strong>
              </div>
              <ul><li v-for="item in draftAuditPanel.items" :key="item">{{ item }}</li></ul>
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
              <select v-model="draftFilter"><option>全部状态</option><option>草稿</option><option>待审核</option><option>已通过</option><option>已生成渠道版本</option></select>
            </div>
          </div>
          <div v-if="!draftTimelineGroups.length" class="timeline-empty muted">当前筛选条件下暂无母稿</div>
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
                <div v-for="draft in group.drafts" :key="draft.id" class="draft-node">
                  <div class="draft-main">
                    <div>
                      <h3>{{ draft.title }}</h3>
                      <p>{{ draft.summary }}</p>
                      <div class="tagline"><span class="tag">{{ draft.source }}</span><span class="tag">{{ draft.status }}</span><span class="tag">{{ draft.audit }}</span></div>
                    </div>
                    <div class="row-actions">
                      <button class="btn small" @click="editDraftInWorkbench(draft)">进入编辑器</button>
                      <button v-if="canManageDraft" class="btn small ghost" @click="generateDraftAudit(draft)">审核建议</button>
                      <button v-if="canManageDraft" class="btn small ghost" @click="approveDraft(draft)">审核通过</button>
                      <button v-if="canManageDraft" class="btn small ghost danger-text" @click="rejectDraft(draft)">驳回</button>
                      <button v-if="canManageChannelContent" class="btn small dark" @click="generateChannelsForDraft(draft)">生成渠道</button>
                    </div>
                  </div>
                  <div v-if="draft.channels.length" class="channel-list">
                    <div class="channel-list-label">渠道内容</div>
                    <div v-for="channel in draft.channels" :key="channel.id" class="channel-row">
                      <div class="channel-row-leading">
                        <span class="channel-logo" :class="`channel-${channel.channel}`">{{ planChannelIcon(channel.channel) }}</span>
                        <div class="channel-row-text">
                          <span class="channel-name">{{ channel.channel }}</span>
                          <span class="channel-title">{{ channel.title }}</span>
                        </div>
                      </div>
                      <span :class="['badge', channel.status === '已确认' ? 'success' : 'warning']">{{ channel.status }}</span>
                      <button class="btn small" @click="openChannelEditor(draft, channel)">编辑/预览</button>
                      <button v-if="canManageChannelContent" class="btn small ghost" @click="confirmChannel(channel)">确认</button>
                      <button v-if="canManagePublishPlan" class="btn small dark" @click="addChannelToPlan(draft, channel)">加入发布计划</button>
                    </div>
                  </div>
                </div>
              </div>
            </div>
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
                  <div v-for="plan in day.plans" :key="plan.id" class="calendar-event">
                    <span>{{ plan.time }}</span>
                    <p>{{ plan.title }}</p>
                    <small>{{ plan.channel }} · {{ plan.status }}</small>
                  </div>
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
              <div class="metric-card"><span>可发布</span><strong>{{ planQueueSummary.ready }}</strong><p>审核、账号、素材均通过</p></div>
              <div class="metric-card"><span>待人工确认</span><strong>{{ planQueueSummary.manual }}</strong><p>Agent / 草稿箱发布需确认</p></div>
              <div class="metric-card"><span>阻断任务</span><strong>{{ planQueueSummary.blocked }}</strong><p>账号、审核或素材未通过</p></div>
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
                  <thead>
                    <tr>
                      <th>时间</th>
                      <th>渠道</th>
                      <th>账号</th>
                      <th>内容</th>
                      <th>发布方式</th>
                      <th>渠道审核</th>
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
                          <span class="channel-logo">{{ planChannelIcon(plan.channel) }}</span>
                          <strong>{{ plan.channel }}</strong>
                        </div>
                      </td>
                      <td>{{ plan.accountName }}</td>
                      <td class="queue-content">
                        <strong>{{ plan.title }}</strong>
                        <p>负责人：{{ plan.owner }}</p>
                      </td>
                      <td>{{ plan.method }}</td>
                      <td><span :class="['badge', queueStatusClass(plan.channelAudit)]">{{ plan.channelAudit }}</span></td>
                      <td><span :class="['badge', queueStatusClass(plan.accountStatus)]">{{ plan.accountStatus }}</span></td>
                      <td><span :class="['badge', queueStatusClass(plan.materialStatus)]">{{ plan.materialStatus }}</span></td>
                      <td><span :class="['badge', queueStatusClass(plan.publishStatus)]">{{ plan.publishStatus }}</span></td>
                      <td class="row-actions-inline">
                        <button v-if="canManagePublishPlan" type="button" class="btn small ghost" @click="runPlanAudit(plan)">运行审核</button>
                        <button v-if="canManagePublishPlan" type="button" class="btn small ghost" @click="preCheckPlan(plan)">前置检查</button>
                        <button v-if="canManagePublishPlan" type="button" class="btn small dark" @click="completePlan(plan)">完成</button>
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
                <tr v-for="account in channelAccounts" :key="account.id">
                  <td>
                    <div class="channel-cell compact">
                      <span class="channel-logo" :class="`channel-${account.channel}`">{{ planChannelIcon(account.channel) }}</span>
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
              <p>{{ brand.position }}</p>
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
            <InfoBlock title="内容口径" :value="currentBrand.tone" />
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
          <div v-else class="material-list">
            <div v-if="!currentBrand.materials.length" class="timeline-empty">
              暂无素材。后续可上传品牌图、商品主图、详情页、买家评价等内容资产。
            </div>
            <div v-for="m in currentBrand.materials" :key="m" class="material-card">{{ m }}</div>
          </div>
        </div>
      </section>
    </main>

    <div v-if="modal.type" class="modal-mask" @click.self="closeModal">
      <div class="modal" :class="modal.wide ? 'wide-modal' : ''">
        <div class="modal-header">
          <h3>{{ modal.title }}</h3>
          <button @click="closeModal">×</button>
        </div>

        <div v-if="modal.type === 'mode'" class="modal-body">
          <p class="mode-intro">两层审核相互独立：<strong>母稿</strong> → <strong>渠道内容</strong>。选择「AI 审核」时，系统按对应 Skill 自动检查合规性；「AI 审核 + 人工确认」用于高风险场景。</p>
          <label>工作模式<select v-model="modeConfig.workMode"><option>人工创作</option><option>智能生成</option></select></label>
          <label v-for="layer in auditLayers" :key="layer.key">
            {{ layer.label }}
            <select v-model="modeConfig[layer.key]">
              <option v-for="opt in auditOptionsFor(layer.key)" :key="opt.value" :value="opt.value">{{ opt.value }}</option>
            </select>
            <small class="field-hint">{{ auditModeHint(modeConfig[layer.key], layer) }}</small>
          </label>
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
          <label>内容口径<input v-model="brandForm.tone" placeholder="例如 专业、轻法式、可信、不硬广" /></label>
          <label>品牌关键词<textarea v-model="brandForm.keywordsText" rows="2" placeholder="多个关键词用逗号、顿号或换行分隔"></textarea></label>
          <div class="drawer-actions">
            <button class="btn ghost" @click="closeModal">取消</button>
            <button class="btn primary" :disabled="loading.action || !brandForm.brand_name.trim()" @click="saveBrand">
              {{ loading.action ? '保存中' : '保存品牌' }}
            </button>
          </div>
        </div>

        <div v-if="modal.type === 'import'" class="modal-body import-flow">
          <div class="steps"><span class="active">1 上传文件</span><span class="active">2 字段映射</span><span>3 校验预览</span><span>4 确认导入</span></div>
          <div class="upload-box">拖拽或选择 Excel / CSV 文件<br><small>系统读取表头后，映射到商品字段、SKU 字段、SKU 覆盖字段、竞品字段或自定义字段</small></div>
          <table class="table mapping-table">
            <thead><tr><th>原始字段</th><th>样例值</th><th>映射目标</th><th>字段层级</th><th>字段类型</th><th>动作</th></tr></thead>
            <tbody>
              <tr v-for="row in importMappings" :key="row.source">
                <td>{{ row.source }}</td><td>{{ row.sample }}</td>
                <td><input v-model="row.target" /></td>
                <td><select v-model="row.level"><option>商品字段</option><option>SKU字段</option><option>SKU覆盖字段</option><option>竞品字段</option><option>自定义字段</option><option>忽略</option></select></td>
                <td><select v-model="row.type"><option>文本</option><option>数字</option><option>枚举</option><option>图片</option><option>多图</option><option>链接</option></select></td>
                <td><button class="btn small ghost" @click="row.saved = true">保存映射</button></td>
              </tr>
            </tbody>
          </table>
          <button class="btn primary full" @click="mockImport">校验并导入</button>
        </div>

        <div v-if="modal.type === 'newPlan'" class="modal-body">
          <label>选择母稿<select v-model="newPlanDraftId"><option v-for="draft in drafts" :value="draft.id" :key="draft.id">{{ draft.title }}</option></select></label>
          <label>渠道<select v-model="newPlanChannel"><option v-for="c in channelProfiles" :key="c.name" :value="c.name">{{ c.name }}</option></select></label>
          <label>执行方式<select v-model="newPlanMethod"><option>渠道 API</option><option>Agent 执行</option><option>人工执行</option></select></label>
          <label>自动化级别<select v-model="newPlanLevel"><option>全自动</option><option>半自动</option><option>人工</option></select></label>
          <label>计划时间<input type="datetime-local" v-model="newPlanTime" /></label>
          <button v-if="canManagePublishPlan" class="btn primary full" @click="createPlan">生成发布任务</button>
        </div>
      </div>
    </div>

    <aside v-if="drawer.type" class="drawer">
      <div class="drawer-header">
        <h3>{{ drawer.title }}</h3>
        <button @click="closeDrawer">×</button>
      </div>

      <div v-if="drawer.type === 'hotspot'" class="drawer-body">
        <div class="hotspot-tools"><input v-model="hotspotSearch" placeholder="搜索热点关键词" /><button class="btn small" @click="addManualHotspot">手动添加</button></div>
        <div v-for="hot in filteredHotspots" :key="hot.id" class="hotspot-card">
          <div><h4>{{ hot.title }}</h4><p>{{ hot.summary }}</p><span>{{ hot.platform }} · 热度 {{ hot.heat }} · {{ hot.risk }}</span></div>
          <div class="row-actions"><button class="btn small" @click="useHotspot(hot, 'dialog')">引用到对话</button><button class="btn small ghost" @click="useHotspot(hot, 'draft')">生成借势角度</button></div>
        </div>
      </div>

      <div v-if="drawer.type === 'channelEditor'" class="drawer-body channel-editor">
        <div class="segmented full-width">
          <button :class="{ active: channelPreviewMode === 'edit' }" @click="channelPreviewMode = 'edit'">编辑</button>
          <button :class="{ active: channelPreviewMode === 'pc' }" @click="channelPreviewMode = 'pc'">PC预览</button>
          <button :class="{ active: channelPreviewMode === 'mobile' }" @click="channelPreviewMode = 'mobile'">手机预览</button>
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
            <button v-if="canManageChannelContent" class="btn small dark" @click="generateChannelAudit(selectedChannel)">AI审核建议</button>
          </div>
          <div v-if="channelAuditPanel.summary || channelAuditPanel.items.length" class="audit-result compact">
            <div class="audit-result-head">
              <span :class="['badge', channelAuditPanel.passed ? 'success' : 'warning']">{{ channelAuditPanel.risk }}</span>
              <strong>{{ channelAuditPanel.summary }}</strong>
            </div>
            <ul><li v-for="item in channelAuditPanel.items" :key="item">{{ item }}</li></ul>
          </div>
        </div>
        <ChannelPreview v-else :mode="channelPreviewMode" :channel="selectedChannel" />
        <div class="drawer-actions">
          <button v-if="canManageChannelContent" class="btn ghost" @click="regenerateChannel">重新生成</button>
          <button v-if="canManageChannelContent" class="btn primary" @click="confirmSelectedChannel">确认渠道内容</button>
          <button v-if="canManageChannelContent" class="btn ghost danger-text" @click="rejectSelectedChannel">驳回</button>
          <button v-if="canManagePublishPlan" class="btn dark" @click="addSelectedChannelToPlan">加入发布计划</button>
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
    </aside>

    <div v-if="toast.show" class="toast">{{ toast.text }}</div>
  </div>
</template>

<script setup>
import { computed, defineComponent, h, onMounted, reactive, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { usePermissionStore } from '@/stores/permission'
import ChannelManagementPanel from '../components/ChannelManagementPanel.vue'
import {
  approveAiGeoChannelContent,
  approveAiGeoDraft,
  createAiGeoBrand,
  createAiGeoCompetitor,
  createAiGeoDraft,
  createAiGeoHotspot,
  createAiGeoPublishPlan,
  fetchAiGeoBrands,
  fetchAiGeoChannels,
  fetchAiGeoChannelContents,
  fetchAiGeoChannelContentAuditSuggestions,
  fetchAiGeoCompetitors,
  fetchAiGeoDraftAuditSuggestions,
  fetchAiGeoDrafts,
  fetchAiGeoHotspots,
  fetchAiGeoKeywords,
  fetchAiGeoMaterialAssets,
  fetchAiGeoOverview,
  fetchAiGeoProducts,
  fetchAiGeoPublishPlanCalendar,
  fetchAiGeoPublishPlans,
  fetchAiGeoSKUs,
  generateAiGeoChannelContentAuditSuggestion,
  generateAiGeoChannelContent,
  generateAiGeoDraftAuditSuggestion,
  generateAiGeoDraft,
  importAiGeoMaterials,
  rejectAiGeoDraft,
  rejectAiGeoChannelContent,
  streamAiGeoGatewayInvoke,
  submitAiGeoDraft,
  updateAiGeoChannelContent,
  updateAiGeoBrand,
  updateAiGeoPublishPlan,
  updateAiGeoPublishPlanStatus,
} from '../api'

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

const PreviewPane = defineComponent({
  props: ['mode', 'title', 'summary', 'body', 'coverImage'],
  setup(props) {
    return () => h('div', { class: ['preview-pane', props.mode === 'mobile' ? 'mobile-frame' : 'pc-frame'] }, [
      props.coverImage ? h('img', { class: 'preview-cover', src: props.coverImage, alt: '封面' }) : null,
      h('h1', props.title),
      h('p', { class: 'summary' }, props.summary),
      ...String(props.body || '').split('\n').map(p => h('p', p))
    ])
  }
})

const ChannelPreview = defineComponent({
  props: ['mode', 'channel'],
  setup(props) {
    return () => h('div', { class: ['channel-preview', props.mode === 'mobile' ? 'mobile-frame' : 'pc-frame', `channel-${props.channel.channel}`] }, [
      h('div', { class: 'preview-channel-name' }, `${props.channel.channel} 预览`),
      h('h2', props.channel.title),
      h('p', { class: 'summary' }, props.channel.tags || props.channel.seoTitle || ''),
      ...String(props.channel.body || '').split('\n').map(p => h('p', p)),
      props.channel.script ? h('pre', props.channel.script) : null
    ])
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
}

const menuRouteMap = {
  overview: '/ai-geo/dashboard',
  workbench: '/ai-geo/workbench',
  drafts: '/ai-geo/drafts',
  plans: '/ai-geo/plans/queue',
  channels: '/ai-geo/channels',
  data: '/ai-geo/data/products',
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
const dataTabs = ['品牌信息', '商品资料', '关键词', '素材']
const productTab = ref('公共资料')
const productTabs = ['公共资料', 'SKU明细', 'SKU覆盖资料', '竞品信息', '关键词/内容']
const previewMode = ref('edit')
const channelPreviewMode = ref('edit')
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
  const parts = [currentBrand.value.position, currentBrand.value.audience, currentBrand.value.priceBand].filter(Boolean)
  return parts.length ? parts.join(' · ') : '品牌定位、目标人群、价格带待维护'
})
const productCount = computed(() => overview.product_count || brands.reduce((sum, b) => sum + b.products.length, 0))
const skuCount = computed(() => overview.sku_count || brands.reduce((sum, b) => sum + countSkus(b), 0))
function countSkus(brand) { return brand.products.reduce((sum, p) => sum + p.skus.length, 0) }
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

const auditModeCatalog = [
  { value: '无需审核', hint: '跳过本环节，直接进入下一步' },
  { value: '人工审核', hint: '由运营人工逐条确认' },
  { value: 'AI审核', hint: '按绑定的 Skill 自动检查合规性（口径、禁词、渠道规范）' },
  { value: 'AI审核 + 人工确认', hint: 'Skill 先审并标记风险项，人工复核后通过' },
  { value: '按渠道配置', hint: '读取各渠道单独配置的审核策略' }
]

const auditLayers = [
  { key: 'draftAuditMode', label: '母稿审核模式', skills: '品牌介绍母稿 Skill、商品种草母稿 Skill' },
  { key: 'channelAuditMode', label: '渠道内容审核模式', skills: '小红书改写 Skill、知乎问答改写 Skill' }
]

function auditOptionsFor(key) {
  if (key === 'draftAuditMode') return auditModeCatalog.filter(o => o.value !== '按渠道配置')
  return auditModeCatalog
}

function auditModeHint(value, layer) {
  const item = auditModeCatalog.find(o => o.value === value)
  if (!item) return ''
  if (value === 'AI审核' || value === 'AI审核 + 人工确认') {
    return `${item.hint}。关联 Skill：${layer.skills}`
  }
  return item.hint
}

const modeConfig = reactive({
  workMode: '人工创作',
  draftAuditMode: 'AI审核 + 人工确认',
  channelAuditMode: 'AI审核'
})

const channelProfiles = reactive([
  { name: '独立站', icon: '🌐', desc: '品牌独立站', type: '自有站点', siteUrl: 'https://mardiladin.com', adminUrl: 'https://admin.mardiladin.com', contentTypes: 'SEO 长文 / 商品详情', supportMethods: '渠道 API / 人工', defaultMethod: 'API 自动发布', accountCount: 1, status: '可发布', method: '渠道 API', level: '全自动', skill: '', risk: '接口校验 + Sitemap 更新', enabled: true },
  { name: '小红书', icon: '📕', desc: '图文种草平台', type: '社交平台', siteUrl: 'https://www.xiaohongshu.com', adminUrl: 'https://creator.xiaohongshu.com', contentTypes: '图文笔记', supportMethods: '渠道 API / Agent / 人工', defaultMethod: 'API 草稿 + 人工确认', accountCount: 1, status: '可发布', method: 'Agent 执行', level: '半自动', skill: '小红书图文发布 Skill', risk: '随机停顿 / 频率限制 / 人工确认', enabled: true },
  { name: '抖音', icon: '🎵', desc: '短视频平台', type: '视频平台', siteUrl: 'https://www.douyin.com', adminUrl: 'https://creator.douyin.com', contentTypes: '短视频 / 图文', supportMethods: '渠道 API / Agent / 人工', defaultMethod: 'API 草稿 + 人工确认', accountCount: 0, status: '待授权', method: 'Agent 执行', level: '半自动', skill: '抖音短视频发布 Skill', risk: '频率限制 + 人工确认', enabled: true },
  { name: '微信公众号', icon: '💬', desc: '微信内容平台', type: '内容平台', siteUrl: 'https://mp.weixin.qq.com', adminUrl: 'https://mp.weixin.qq.com', contentTypes: '长图文', supportMethods: '渠道 API / 人工', defaultMethod: 'API 草稿 + 人工确认', accountCount: 1, status: '可发布', method: '渠道 API', level: '半自动', skill: '', risk: '草稿箱接口 + 发布前确认', enabled: true },
  { name: '微博', icon: '📢', desc: '社交媒体平台', type: '社交平台', siteUrl: 'https://weibo.com', adminUrl: 'https://weibo.com', contentTypes: '短图文 / 话题', supportMethods: '渠道 API / 人工', defaultMethod: 'API 自动发布', accountCount: 0, status: '待授权', method: '渠道 API', level: '半自动', skill: '', risk: '敏感词检查 + 发布前确认', enabled: true },
  { name: '百家号', icon: '📰', desc: '百度内容平台', type: '内容平台', siteUrl: 'https://baijiahao.baidu.com', adminUrl: 'https://baijiahao.baidu.com', contentTypes: '资讯 / 图文', supportMethods: '渠道 API / 人工', defaultMethod: 'API 自动发布', accountCount: 0, status: '未配置', method: '渠道 API', level: '半自动', skill: '', risk: '原创检测 + 接口校验', enabled: true },
  { name: '知乎', icon: '💡', desc: '问答与专栏', type: '问答平台', siteUrl: 'https://www.zhihu.com', adminUrl: 'https://www.zhihu.com/creator', contentTypes: '回答 / 文章', supportMethods: '渠道 API / Agent / 人工', defaultMethod: 'API 自动发布', accountCount: 1, status: '待授权', method: 'Agent 执行', level: '人工', skill: '知乎回答发布 Skill', risk: '人工接管 / 敏感词检查', enabled: true }
])

const planTabs = ['发布日历', '发布队列']
const planTab = ref('发布队列')

watch(
  () => [route.params.section, route.params.subsection],
  ([section, subsection]) => {
    if (section === 'plans') {
      planTab.value = subsection === 'calendar' ? '发布日历' : '发布队列'
    }
    if (section === 'data') {
      dataTab.value = subsection === 'brands' ? '品牌信息' : '商品资料'
    }
  },
  { immediate: true }
)

const channelAccounts = reactive([
  { id: 1, channel: '独立站', accountName: 'Mardi 官网主站', login: 'admin@mardi.com', status: '已授权', updatedAt: '2026-05-18' },
  { id: 2, channel: '小红书', accountName: 'Mardi 小红书官方', login: 'xhs_mardi_official', status: '已授权', updatedAt: '2026-05-19' },
  { id: 3, channel: '微信公众号', accountName: 'Mardi 品牌公众号', login: 'mardi_wechat', status: '已授权', updatedAt: '2026-05-17' },
  { id: 4, channel: '知乎', accountName: 'Mardi 品牌知乎', login: '—', status: '待授权', updatedAt: '2026-05-15' },
  { id: 5, channel: '抖音', accountName: 'Mardi 抖音官方', login: '—', status: '待授权', updatedAt: '2026-05-16' },
  { id: 6, channel: '微博', accountName: 'Mardi 官方微博', login: '—', status: '待授权', updatedAt: '2026-05-14' },
  { id: 7, channel: '百家号', accountName: 'Mardi 百家号', login: '—', status: '未配置', updatedAt: '2026-05-13' }
])

const planCalendarDays = ref([])

const calendarDays = computed(() => {
  if (planCalendarDays.value.length) {
    return planCalendarDays.value.map(day => ({
      date: day.date,
      label: day.date.slice(5).replace('-', '/'),
      plans: (day.items || []).map(apiPlanToPlan),
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
  { label: '今日母稿', value: overview.draft_count_today || drafts.filter(d => d.date === '2026-05-20').length, desc: '含草稿和已审核' },
  { label: '待审核', value: overview.pending_draft_count || drafts.filter(d => d.status === '待审核').length, desc: '需人工确认' },
  { label: '渠道内容', value: overview.channel_content_count || drafts.reduce((s, d) => s + d.channels.length, 0), desc: '已生成版本' },
  { label: '今日发布', value: overview.publish_plan_today || plans.filter(p => p.date === '2026-05-20').length, desc: '发布计划任务' }
])

const pendingTasks = computed(() => {
  if (overview.pending_tasks?.length) {
    return overview.pending_tasks.map((task, index) => ({ id: `api-${index}`, tag: task.tag, title: task.title, menu: index === 0 ? 'drafts' : 'data' }))
  }
  const pendingDraftCount = drafts.filter(d => d.status === '待审核').length
  const failedPlanCount = plans.filter(p => p.status === '发布失败').length
  return [
    { id: 'create', tag: '人工 AI 创作', title: '根据运营想法生成母稿', menu: 'workbench' },
    { id: 'audit', tag: '母稿审核', title: pendingDraftCount ? `${pendingDraftCount} 篇待审核` : '暂无待审核母稿', menu: 'drafts' },
    { id: 'plan', tag: '发布计划', title: failedPlanCount ? `${failedPlanCount} 个失败任务需处理` : '暂无失败任务', menu: 'plans' }
  ]
})

const skills = ['品牌介绍母稿 Skill', '商品种草母稿 Skill', '场景攻略母稿 Skill', 'FAQ问答母稿 Skill', '小红书改写 Skill', '知乎问答改写 Skill']

const workbench = reactive({ brandId: '', productId: '', skill: '', hotspot: null, prompt: '' })
const chatMessages = reactive([])
const chatStreaming = ref(false)
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
})
const skillProfiles = {
  '品牌介绍母稿 Skill': { name: '品牌介绍母稿 Skill', goal: '建立品牌认知', output: '品牌定位长文', desc: '适合生成品牌介绍、品牌故事、品牌优势和 GEO 搜索入口文章。', tags: ['品牌定位', '目标人群', '搜索心智'] },
  '商品种草母稿 Skill': { name: '商品种草母稿 Skill', goal: '推动商品种草', output: '商品推荐文章', desc: '适合围绕单品卖点、适用人群、选购理由生成有转化指向的内容。', tags: ['卖点提炼', '适用场景', '购买理由'] },
  '场景攻略母稿 Skill': { name: '场景攻略母稿 Skill', goal: '占领场景搜索', output: '场景解决方案', desc: '适合回答“怎么选、怎么搭、适合谁”这类 GEO 问题。', tags: ['场景问题', '解决方案', '对比建议'] },
  'FAQ问答母稿 Skill': { name: 'FAQ问答母稿 Skill', goal: '承接长尾问答', output: '问答型文章', desc: '适合生成知乎、搜索问答、独立站 FAQ 可复用内容。', tags: ['FAQ', '长尾词', '可信回答'] },
  '小红书改写 Skill': { name: '小红书改写 Skill', goal: '生成种草笔记', output: '小红书渠道文', desc: '适合把母稿改成轻口语、强场景、带话题标签的图文笔记。', tags: ['口语化', '话题标签', '种草感'] },
  '知乎问答改写 Skill': { name: '知乎问答改写 Skill', goal: '生成可信回答', output: '知乎问答文', desc: '适合把母稿改成解释充分、逻辑清晰、可信度更高的回答。', tags: ['问题拆解', '理性表达', '可信证据'] },
}
const defaultSkillProfile = { name: '未选择 Skill', goal: '自由创作', output: '通用 GEO 母稿', desc: '选择 Skill 后，AI 会按固定任务框架组织标题、摘要、正文和关键词。', tags: ['自由提示', '资料驱动', '人工确认'] }
const sampleCoverImage = 'https://images.unsplash.com/photo-1595777457583-95e059d581b8?auto=format&fit=crop&w=800&q=80'

const editingDraft = reactive({
  id: 0,
  title: '小个子女生怎么选通勤连衣裙？',
  summary: '围绕小个子、通勤场景和轻法式穿搭，生成一篇可改写到多渠道的母稿。',
  body: '很多小个子女生选通勤连衣裙时，最怕压身高、显拖沓。\n选择收腰线清晰、裙长不过分压低比例的款式，会更容易穿出利落感。\n法式通勤连衣裙的优势，是在正式感和松弛感之间取得平衡。',
  keywordsText: '小个子连衣裙, 通勤穿搭, 法式通勤',
  coverImage: sampleCoverImage,
  status: '草稿',
  source: '人工创作'
})

const selectedWorkbenchBrand = computed(() => brands.find(b => b.id === Number(workbench.brandId)) || currentBrand.value)
const selectedWorkbenchProduct = computed(() => selectedWorkbenchBrand.value.products.find(p => p.id === Number(workbench.productId)) || null)
const selectedSkillProfile = computed(() => skillProfiles[workbench.skill] || defaultSkillProfile)
const workbenchKeywords = computed(() => {
  const brandKeywords = selectedWorkbenchBrand.value.keywordGroups.flatMap(group => group.keywords || [])
  const productKeywords = selectedWorkbenchProduct.value?.keywords || []
  return [...new Set([...productKeywords, ...brandKeywords])]
})
const workbenchEvidence = computed(() => [
  { label: '人群', value: selectedWorkbenchProduct.value?.audience || selectedWorkbenchBrand.value.audience || '未维护' },
  { label: '卖点', value: selectedWorkbenchProduct.value?.sellingPoints || selectedWorkbenchBrand.value.position || '未维护' },
  { label: '关键词', value: workbenchKeywords.value.slice(0, 4).join('、') || '未维护' },
  { label: '热点', value: workbench.hotspot?.title || '不引用热点' },
])
const generationPlaceholder = computed(() => {
  const product = selectedWorkbenchProduct.value?.name || '当前资料'
  return `例如：围绕「${product}」写一篇回答“小个子通勤怎么穿”的 GEO 文章，强调适合人群、选择理由、场景建议。`
})
const hasWorkbenchInput = computed(() => Boolean(String(workbench.prompt || '').trim()))
const workbenchUserMessages = computed(() => chatMessages.filter(msg => msg.role === 'user'))
const hasStartedWorkbenchChat = computed(() => workbenchUserMessages.value.length > 0)
const latestWorkbenchUserPrompt = computed(() => [...chatMessages].reverse().find(msg => msg.role === 'user')?.text || '')
const workbenchReadiness = computed(() => evaluateWorkbenchReadiness(String(workbench.prompt || '').trim()))
watch(() => workbench.brandId, () => {
  workbench.productId = ''
})

const drafts = reactive([
  {
    id: 201,
    date: '2026-05-20',
    title: '小个子女生怎么选通勤连衣裙？',
    summary: '基于法式通勤连衣裙生成的选购建议母稿。',
    body: editingDraft.body,
    source: '人工创作',
    status: '已通过',
    audit: 'AI审核 + 人工确认',
    channels: [
      { id: 301, channel: '独立站', title: '小个子女生怎么选通勤连衣裙？', body: editingDraft.body, tags: '', seoTitle: '小个子通勤连衣裙选购指南', script: '', status: '已确认' },
      { id: 302, channel: '小红书', title: '小个子通勤裙别乱买！', body: '小个子选通勤裙，重点看腰线、长度和版型。\n这条轻法式连衣裙属于不夸张但很显比例的类型。', tags: '#小个子穿搭 #通勤穿搭 #法式穿搭', seoTitle: '', script: '', status: '已编辑' }
    ]
  },
  {
    id: 203,
    date: '2026-05-19',
    title: '梨形身材春夏怎么选半身裙？',
    summary: '围绕 A 字半身裙与收腰版型生成的穿搭建议母稿。',
    body: '梨形身材选裙时，重点看腰臀过渡与裙摆垂感，避免胯部膨胀感。',
    source: '人工创作',
    status: '草稿',
    audit: '人工审核',
    channels: [
      { id: 303, channel: '小红书', title: '梨形身材半身裙避雷', body: 'A 字版型更友好，面料要有垂感。', tags: '#梨形身材 #半身裙', seoTitle: '', script: '', status: '已编辑' }
    ]
  },
  {
    id: 202,
    date: '2026-05-20',
    title: '春夏通勤衬衫怎么穿不普通？',
    summary: '围绕飘带法式衬衫生成通勤穿搭母稿。',
    body: '春夏通勤衬衫不一定要很正式。飘带领、轻薄面料和干净剪裁，可以让基础款更有细节。',
    source: '智能生成',
    status: '待审核',
    audit: 'AI审核 + 人工确认',
    channels: []
  }
])

const plans = reactive([
  {
    id: 401, date: '2026-05-20', time: '10:00', title: '小个子通勤裙别乱买！', channel: '小红书',
    accountName: 'Mardi 官方号', method: 'Agent 执行', level: '半自动', skill: '小红书图文发布 Skill',
    risk: '随机停顿 + 人工确认', owner: '运营A', status: '待发布', link: '',
    channelAudit: '待审核', accountStatus: '待人工确认', materialStatus: '凭证异常', publishStatus: '可发布'
  },
  {
    id: 402, date: '2026-05-20', time: '14:00', title: '小个子女生怎么选通勤连衣裙？', channel: '独立站',
    accountName: 'Mardi 官网主站', method: '渠道 API', level: '全自动', skill: '', risk: '接口校验',
    owner: '系统', status: '已排期', link: '',
    channelAudit: '已通过', accountStatus: '可发布', materialStatus: '完整', publishStatus: '可发布'
  },
  {
    id: 404, date: '2026-05-20', time: '17:00', title: '春夏通勤衬衫怎么穿不普通？', channel: '微信公众号',
    accountName: 'Mardi 公众号', method: '渠道 API', level: '半自动', skill: '', risk: '草稿箱接口 + 发布前确认',
    owner: '运营A', status: '待发布', link: '',
    channelAudit: '已通过', accountStatus: '阻断', materialStatus: '缺封面', publishStatus: '阻断'
  }
])

const activeTitle = computed(() => menus.find(m => m.key === activeMenu.value)?.label || '')
const activeSubtitle = computed(() => ({
  overview: '整个应用总览：资料、母稿、发布计划、渠道管理',
  workbench: '人工 AI 创作空间：左侧对话，右侧母稿编辑与预览',
  drafts: '母稿与渠道内容：按日期管理和编辑预览',
  plans: '发布日历与发布队列',
  channels: '渠道资料与账号授权',
  data: '资料中心：品牌 → 商品资料卡 → SKU / 竞品信息'
}[activeMenu.value]))

const draftDateOptions = computed(() => [...new Set(drafts.map(d => d.date))].sort((a, b) => b.localeCompare(a)))
const filteredDrafts = computed(() => drafts.filter(d => {
  if (draftDate.value && d.date !== draftDate.value) return false
  if (draftFilter.value !== '全部状态' && d.status !== draftFilter.value) return false
  return true
}))
const draftTimelineGroups = computed(() => {
  const map = new Map()
  for (const draft of filteredDrafts.value) {
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
const filteredPlans = computed(() => plans.filter(p => p.date === planDate.value))

const planQueueSummary = computed(() => {
  const list = filteredPlans.value
  const isReady = p => p.channelAudit === '已通过' && p.accountStatus === '可发布' && p.materialStatus === '完整' && p.publishStatus === '可发布'
  return {
    total: list.length,
    ready: list.filter(p => p.publishStatus === '可发布').length,
    manual: list.filter(p => p.accountStatus === '待人工确认').length,
    blocked: list.filter(p => p.publishStatus === '阻断' || p.accountStatus === '阻断').length
  }
})

function planChannelIcon(channel) {
  return channelProfiles.find(c => c.name === channel)?.icon || '📣'
}

function queueStatusClass(status) {
  if (['已通过', '可发布', '完整'].includes(status)) return 'success'
  if (status === '阻断') return 'danger'
  return 'warning'
}

const modal = reactive({ type: '', title: '', wide: false })
const drawer = reactive({ type: '', title: '' })
const toast = reactive({ show: false, text: '' })
const selectedChannel = reactive({})
const selectedDraft = ref(null)
const selectedProduct = reactive({})
const brandForm = reactive({
  id: 0,
  brand_code: '',
  brand_name: '',
  positioning: '',
  target_audience: '',
  price_band: '',
  tone: '',
  keywordsText: '',
})
const draftAuditPanel = reactive({ summary: '', risk: '未审核', passed: false, items: [] })
const channelAuditPanel = reactive({ summary: '', risk: '未审核', passed: false, items: [] })

const hotspots = reactive([])
const hotspotSearch = ref('')
const filteredHotspots = computed(() => hotspots.filter(h => !hotspotSearch.value || h.title.includes(hotspotSearch.value) || h.summary.includes(hotspotSearch.value)))

const selectedHotspotId = computed({
  get: () => workbench.hotspot?.id ?? '',
  set(id) {
    if (id === '__more__') {
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

const importMappings = reactive([
  { source: '款式编号', sample: 'SPU-DRESS-001', target: '商品编码', level: '商品字段', type: '文本', saved: false },
  { source: '颜色', sample: '黑色', target: '颜色', level: 'SKU字段', type: '枚举', saved: false },
  { source: '黑色款卖点', sample: '显瘦通勤', target: 'SKU卖点覆盖', level: 'SKU覆盖字段', type: '文本', saved: false },
  { source: '竞品品牌', sample: 'Lily', target: '竞品品牌', level: '竞品字段', type: '文本', saved: false },
  { source: '竞品链接', sample: 'https://...', target: '竞品链接', level: '竞品字段', type: '链接', saved: false }
])

const newPlanDraftId = ref(201)
const newPlanChannel = ref('小红书')
const newPlanMethod = ref('Agent 执行')
const newPlanLevel = ref('半自动')
const newPlanTime = ref('2026-05-20T18:00')

onMounted(async () => {
  try {
    await permissionStore.load()
  } catch {
    showToast('权限信息加载失败，按钮将以后端校验为准')
  }
  loadAiGeoData()
})

watch(planDate, () => {
  loadPlanCalendar()
})

async function loadAiGeoData() {
  loading.page = true
  try {
    const calendarRange = publishCalendarRange()
    const [overviewData, brandPage, productPage, skuPage, competitorPage, keywordPage, assetPage, hotspotPage, channelPage, draftPage, channelContentPage, planPage, planCalendar] = await Promise.all([
      fetchAiGeoOverview(),
      fetchAiGeoBrands({ limit: 200 }),
      fetchAiGeoProducts({ limit: 200 }),
      fetchAiGeoSKUs({ limit: 500 }),
      fetchAiGeoCompetitors({ limit: 500 }),
      fetchAiGeoKeywords({ limit: 500, status: 'active' }),
      fetchAiGeoMaterialAssets({ limit: 500 }),
      fetchAiGeoHotspots({ limit: 200, status: 'active' }),
      fetchAiGeoChannels({ limit: 200 }),
      fetchAiGeoDrafts({ limit: 200 }),
      fetchAiGeoChannelContents({ limit: 500 }),
      fetchAiGeoPublishPlans({ limit: 200 }),
      fetchAiGeoPublishPlanCalendar(calendarRange),
    ])
    Object.assign(overview, overviewData)
    hydrateBrands(brandPage.items || [], productPage.items || [], skuPage.items || [], competitorPage.items || [], assetPage.items || [], keywordPage.items || [])
    hydrateHotspots(hotspotPage.items || [])
    hydrateChannels(channelPage.items || [])
    hydrateDrafts(draftPage.items || [], channelContentPage.items || [])
    hydratePlans(planPage.items || [])
    planCalendarDays.value = planCalendar.days || []
  } catch (error) {
    showToast(error?.message || 'AI GEO 数据加载失败')
  } finally {
    loading.page = false
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
    name: apiField(channel, 'channel_name', 'ChannelName') || '',
    icon: channelIcon(apiField(channel, 'channel_name', 'ChannelName')),
    desc: apiField(channel, 'channel_type', 'ChannelType') || '',
    type: apiField(channel, 'channel_type', 'ChannelType') || '',
    siteUrl: apiField(channel, 'entry_url', 'EntryURL') || '',
    adminUrl: apiField(channel, 'entry_url', 'EntryURL') || '',
    contentTypes: parseJsonArray(apiField(channel, 'content_forms', 'ContentForms')).join(' / '),
    supportMethods: parseJsonArray(apiField(channel, 'support_modes', 'SupportModes')).join(' / '),
    defaultMethod: apiField(channel, 'default_publish_mode', 'DefaultPublishMode'),
    accountCount: 0,
    status: apiField(channel, 'status', 'Status') === 'active' ? '可发布' : '未配置',
    method: apiField(channel, 'default_publish_mode', 'DefaultPublishMode'),
    level: apiField(channel, 'default_publish_mode', 'DefaultPublishMode') === 'manual' ? '人工' : '半自动',
    skill: '',
    risk: '发布前校验 + 异常人工接管',
    enabled: apiField(channel, 'status', 'Status') === 'active',
  })))
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
    id: apiField(item, 'id', 'ID'),
    title: apiField(item, 'title', 'Title') || '',
    summary: apiField(item, 'source_url', 'SourceURL') || '来自热点资料库，可作为选题借势上下文。',
    platform: apiField(item, 'platform', 'Platform') || '热点',
    heat: Number(apiField(item, 'heat_score', 'HeatScore') || 0),
    risk: '待判断',
  })))
  if (workbench.hotspot && !hotspots.find(h => Number(h.id) === Number(workbench.hotspot.id))) {
    workbench.hotspot = null
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
  drafts.splice(0, drafts.length, ...apiDrafts.map(draft => ({
    id: apiField(draft, 'id', 'ID'),
    date: String(apiField(draft, 'created_at', 'CreatedAt') || new Date().toISOString()).slice(0, 10),
    title: apiField(draft, 'title', 'Title') || '',
    summary: apiField(draft, 'summary', 'Summary') || '',
    body: apiField(draft, 'body', 'Body') || '',
    source: sourceLabel(apiField(draft, 'source', 'Source')),
    status: draftStatusLabel(apiField(draft, 'audit_status', 'AuditStatus')),
    audit: modeConfig.draftAuditMode,
    channels: apiChannelContents
      .filter(content => Number(apiField(content, 'draft_id', 'DraftID')) === Number(apiField(draft, 'id', 'ID')))
      .map(channelContentFromApi),
    rawStatus: apiField(draft, 'audit_status', 'AuditStatus'),
  })))
}

function channelContentFromApi(content) {
  const channelId = apiField(content, 'channel_id', 'ChannelID')
  const channel = channelProfiles.find(item => Number(item.id) === Number(channelId))
  return {
    id: apiField(content, 'id', 'ID'),
    channel: channel?.name || `渠道 ${channelId}`,
    channelId,
    title: apiField(content, 'title', 'Title') || '',
    body: apiField(content, 'body', 'Body') || '',
    tags: '',
    seoTitle: '',
    script: '',
    status: channelContentStatusLabel(apiField(content, 'audit_status', 'AuditStatus')),
    rawAuditStatus: apiField(content, 'audit_status', 'AuditStatus'),
    publishStatus: apiField(content, 'publish_status', 'PublishStatus'),
  }
}

function apiPlanToPlan(plan) {
  const scheduledAt = apiField(plan, 'scheduled_at', 'ScheduledAt')
  const channelId = apiField(plan, 'channel_id', 'ChannelID')
  const planStatus = apiField(plan, 'status', 'Status')
  const scheduled = new Date(scheduledAt)
  return {
    id: apiField(plan, 'id', 'ID'),
    date: Number.isNaN(scheduled.getTime()) ? String(scheduledAt || '').slice(0, 10) : scheduled.toISOString().slice(0, 10),
    time: Number.isNaN(scheduled.getTime()) ? '' : scheduled.toTimeString().slice(0, 5),
    title: apiField(plan, 'plan_code', 'PlanCode') || '',
    channel: channelProfiles.find(c => Number(c.id) === Number(channelId))?.name || `渠道 ${channelId}`,
    channelId,
    channelContentId: apiField(plan, 'channel_content_id', 'ChannelContentID'),
    accountName: '',
    method: apiField(plan, 'publish_method', 'PublishMethod'),
    level: apiField(plan, 'automation_level', 'AutomationLevel'),
    skill: '',
    risk: '发布前校验 + 异常人工接管',
    owner: '系统',
    status: publishStatusLabel(planStatus),
    link: apiField(plan, 'published_url', 'PublishedURL') || '',
    channelAudit: '已通过',
    accountStatus: '可发布',
    materialStatus: '完整',
    publishStatus: planStatus === 'failed' ? '阻断' : '可发布',
  }
}

function hydratePlans(apiPlans) {
  plans.splice(0, plans.length, ...apiPlans.map(apiPlanToPlan))
}

function showToast(text) {
  toast.text = text
  toast.show = true
  setTimeout(() => { toast.show = false }, 1800)
}
function closeModal() { modal.type = ''; modal.title = ''; modal.wide = false }
function closeDrawer() { drawer.type = ''; drawer.title = '' }
function goChannelManagement() { activeMenu.value = 'channels' }
function configureChannel(channel) { showToast(`配置 ${channel.name}`) }
function testChannel(channel) { showToast(`测试 ${channel.name} 连通性`) }
function addChannel() { showToast('新增渠道（原型）') }
function openPendingTask(task) {
  activeMenu.value = task.menu
  if (task.id === 'audit') draftFilter.value = '待审核'
}
function saveModeConfig() { closeModal(); showToast('模式配置已保存') }
function openImportModal() { modal.type = 'import'; modal.title = '动态字段导入：商品 / SKU / 竞品信息'; modal.wide = true }
function openBrandModal(brand = null) {
  const isEdit = Boolean(brand?.id)
  Object.assign(brandForm, {
    id: isEdit ? brand.id : 0,
    brand_code: isEdit ? String(brand.code || brand.brand_code || brand.name || '') : '',
    brand_name: isEdit ? String(brand.name || brand.brand_name || '') : '',
    positioning: isEdit ? String(brand.position || brand.positioning || '') : '',
    target_audience: isEdit ? String(brand.audience || brand.target_audience || '') : '',
    price_band: isEdit ? String(brand.priceBand || brand.price_band || '') : '',
    tone: isEdit ? String(brand.tone || '') : '',
    keywordsText: isEdit ? (brand.keywordGroups || []).flatMap(group => group.keywords || []).join('、') : '',
  })
  modal.type = 'brand'
  modal.title = isEdit ? '编辑品牌资料' : '新增品牌资料'
  modal.wide = false
}
function openNewPlanModal() { modal.type = 'newPlan'; modal.title = '新建发布计划' }
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
    tone: String(brandForm.tone || '').trim(),
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
async function mockImport() {
  loading.action = true
  try {
    await importAiGeoMaterials({
      import_type: 'product_material',
      mapping_config: Object.fromEntries(importMappings.map(item => [item.source, item.target])),
      records: [],
    })
    closeModal()
    await loadAiGeoData()
    showToast('资料导入批次已记录')
  } catch (error) {
    showToast(error?.message || '资料导入失败')
  } finally {
    loading.action = false
  }
}
function generateBrandKeywords() { showToast('已基于品牌、商品、竞品信息生成关键词') }
function pickCoverImage() {
  editingDraft.coverImage = sampleCoverImage
  showToast('已选择封面图（原型演示）')
}
function openHotspotDrawer() { drawer.type = 'hotspot'; drawer.title = '引用热点' }
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
  const userAcceptedDraft = isWorkbenchDraftDecision(intentText)
  const hasSearchQuestion = /写|生成|文章|攻略|问答|种草|小红书|知乎|通勤|怎么|如何|适合|推荐|选择|对比|人群|场景|卖点|关键词|解决|回答|内容|母稿/.test(intentText)
  const hasSpecificBrief = intentText.replace(/\s/g, '').length >= 12
  const hasAudienceOrScene = Boolean(ideaSession.audience || ideaSession.scene || /(目标|人群|用户|场景|口吻|语气|通勤|办公|职场|日常|约会|小个子|女性)/.test(intentText))
  const hasMaterialContext = Boolean(selectedWorkbenchBrand.value?.id || workbench.brandId)
  const ready = ideaSession.stage === 'ready' || userAcceptedDraft || (hasMaterialContext && hasSpecificBrief && (hasSearchQuestion || hasAudienceOrScene))
  const missing = []
  if (!hasSpecificBrief) missing.push('你的核心想法')
  if (!hasSearchQuestion && !hasAudienceOrScene) missing.push('文章要解决的问题或使用场景')
  return {
    ready,
    missing,
    hint: ready ? '可以生成母稿；生成会带上选择资料、Skill 和本轮对话收敛出的 brief。' : `还需要补充：${missing.join('、')}。`,
  }
}
function inferWorkbenchSkill(text) {
  if (/小红书|种草|笔记|探店|穿搭分享/.test(text)) return '小红书改写 Skill'
  if (/知乎|问答|为什么|如何|怎么|对比|避坑|选购/.test(text)) return '知乎问答改写 Skill'
  if (/场景|通勤|约会|职场|旅行|攻略|怎么穿|怎么选/.test(text)) return '场景攻略母稿 Skill'
  if (/商品|单品|卖点|推荐|种草|购买|转化/.test(text)) return '商品种草母稿 Skill'
  if (/品牌|定位|介绍|故事|认知/.test(text)) return '品牌介绍母稿 Skill'
  return workbench.skill || '场景攻略母稿 Skill'
}
function inferWorkbenchProduct(text) {
  const normalized = text.toLowerCase()
  return selectedWorkbenchBrand.value.products.find(product => {
    const haystack = [product.name, product.sellingPoints, ...(product.keywords || [])].join(' ').toLowerCase()
    return haystack && haystack.split(/\s+|、|,|，/).some(token => token && token.length > 1 && normalized.includes(token))
  }) || null
}
function extractIdeaSlots(text) {
  const audienceByLabel = text.match(/(?:目标用户|目标人群|写给|面向)[:：是为给]*([\s\S]*?)(?=使用场景|场景|口吻|语气|$|[，。,.；;\n])/)?.[1]?.trim() || ''
  const sceneByLabel = text.match(/(?:使用场景|场景|用于|适合)[:：是为]*([\s\S]*?)(?=目标用户|目标人群|口吻|语气|$|[，。,.；;\n])/)?.[1]?.trim() || ''
  const audience = audienceByLabel || text.match(/(\d{2}岁[^，。,.；;\n]*|小个子|梨形|职场新人|通勤党|18-35岁[^，。,.；;\n]*)/)?.[0] || selectedWorkbenchProduct.value?.audience || selectedWorkbenchBrand.value.audience || ''
  const scene = sceneByLabel || text.match(/(通勤|上班|职场|办公室|办公场所|高级办公场所|约会|轻正式|旅行|面试|日常|春夏|秋冬)/g)?.join('、') || ''
  const tone = text.match(/(专业问答|种草|轻松|理性|高级|口语|真实|避坑)/g)?.join('、') || ''
  const searchProblem = text.match(/(怎么[\s\S]*?|如何[\s\S]*?|适合[\s\S]*?|为什么[\s\S]*?|选[\s\S]*?)(?=目标用户|目标人群|使用场景|场景|口吻|语气|$|[，。,.；;\n])/)?.[1]?.trim() || ''
  return { audience, scene, tone, searchProblem }
}
function isWorkbenchCorrection(text) {
  return /不是.*补充|已经补充|刚才.*说了|不是给了|你没看到|同样的话|重复问/.test(text)
}
function isWorkbenchDraftDecision(text) {
  return /不补了|不用补|先生成|直接生成|测试一轮|试一轮|看下?是否生成|看能不能生成|可以生成|按这个方向生成|生成母稿/.test(text)
}
function buildIdeaAnalysis(prompt) {
  const messages = workbenchUserMessages.value.map(msg => msg.text)
  if (messages[messages.length - 1] !== prompt) messages.push(prompt)
  const userAcceptedDraft = isWorkbenchDraftDecision(prompt)
  const allText = messages.filter(text => text && !isWorkbenchCorrection(text)).join('\n')
  const recommendedSkill = inferWorkbenchSkill(allText)
  const matchedProduct = inferWorkbenchProduct(allText)
  if (!workbench.skill && recommendedSkill) workbench.skill = recommendedSkill
  if (!workbench.productId && matchedProduct?.id) workbench.productId = String(matchedProduct.id)

  const slots = extractIdeaSlots(allText)
  Object.assign(ideaSession, {
    intent: recommendedSkill,
    recommendedSkill,
    recommendedProductId: matchedProduct?.id || ideaSession.recommendedProductId,
    searchProblem: slots.searchProblem || ideaSession.searchProblem,
    audience: slots.audience || ideaSession.audience,
    scene: slots.scene || ideaSession.scene,
    tone: slots.tone || ideaSession.tone,
  })
  const brand = selectedWorkbenchBrand.value.name
  const product = matchedProduct?.name || selectedWorkbenchProduct.value?.name || '当前品牌资料'
  const missing = []
  if (!ideaSession.searchProblem) missing.push('这篇文章要回答的具体搜索问题')
  if (!ideaSession.audience) missing.push('目标人群')
  if (!ideaSession.scene) missing.push('使用场景')
  const enoughToGenerate = userAcceptedDraft || allText.replace(/\s/g, '').length >= 12 || (isWorkbenchCorrection(prompt) && (ideaSession.audience || ideaSession.scene))
  ideaSession.stage = enoughToGenerate ? 'ready' : 'shaping'
  ideaSession.brief = `围绕「${brand}」${selectedWorkbenchProduct.value || matchedProduct ? `和「${product}」` : '的现有资料'}，用「${recommendedSkill}」生成一篇${ideaSession.tone || '清晰可信'}的 GEO 母稿；核心问题是「${ideaSession.searchProblem || prompt}」，目标读者为「${ideaSession.audience || '潜在目标用户'}」，场景聚焦「${ideaSession.scene || '由资料和对话推断'}」。`
  const correctionReading = isWorkbenchCorrection(prompt) && ideaSession.stage === 'ready'
    ? `你说得对，目标用户和使用场景已经补充了。我已把目标用户识别为「${ideaSession.audience}」，使用场景识别为「${ideaSession.scene}」，现在可以进入生成或继续细化口吻。`
    : ''
  const decisionReading = userAcceptedDraft
    ? `收到，你已经决定先不继续补充。当前信息会作为一次可生成 brief，缺失的人群、场景或口吻由 AI 按品牌资料和 Skill 兜底，不再阻塞生成。`
    : ''
  return {
    intent: recommendedSkill,
    reading: correctionReading || decisionReading || `我会把「${prompt}」先转成可发布母稿的创作 brief。当前资料底座是「${brand}」，${matchedProduct ? `已匹配商品「${matchedProduct.name}」` : selectedWorkbenchProduct.value ? `使用已选商品「${selectedWorkbenchProduct.value.name}」` : `先用「${product}」承接`}，输出框架按「${recommendedSkill}」收敛。`,
    expand: [
      `品牌/资料：从「${selectedWorkbenchBrand.value.position || brand}」里提取可信卖点，不写空泛介绍。`,
      `用户问题：把想法收敛成“${ideaSession.searchProblem || prompt}”这个可被搜索和 AI 问答引用的入口。`,
      `输出增强：母稿会同时保留标题、摘要、正文、关键词和可改写到渠道的论点。`,
    ],
    converge: ideaSession.brief,
    questions: ideaSession.stage === 'ready' ? ['按这个方向生成母稿', '强化人群和场景', '强化商品卖点'] : missing.slice(0, 2).map(item => `补充${item}`),
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
  const brand = selectedWorkbenchBrand.value.name
  const product = selectedWorkbenchProduct.value?.name
  const audience = selectedWorkbenchProduct.value?.audience || selectedWorkbenchBrand.value.audience || '目标用户'
  const scene = ideaSession.scene || '日常/通勤场景'
  if (/品牌介绍/.test(selectedSkillProfile.value.name)) {
    return [
      `「${brand} 是什么风格，适合哪些人」`,
      `「${brand} 为什么适合 ${audience}」`,
      `「${brand} 和同类品牌的差异在哪里」`,
    ]
  }
  if (/商品|种草/.test(selectedSkillProfile.value.name) && product) {
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
      structure: ['先回答“这个品牌适合谁/是什么风格”', '再用定位、人群、价格带和内容口径建立可信度', '最后给出适用场景与选择建议'],
      sample: '适合写成“某类人为什么会选择这个品牌”的认知型文章，而不是品牌自夸介绍。',
    }
  }
  if (/商品|种草/.test(skillName)) {
    return {
      entry: '商品选购入口',
      structure: ['先定义用户痛点或购买犹豫', '再拆商品卖点和适用人群', '最后给场景化购买理由与避坑提醒'],
      sample: '适合写成“这件单品适合谁、解决什么问题、为什么值得选”的种草母稿。',
    }
  }
  if (/场景攻略/.test(skillName)) {
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
  if (/小红书/.test(skillName)) {
    return {
      entry: '生活场景种草入口',
      structure: ['先给真实场景', '再说穿搭/使用感受', '补充适合人群和避雷点', '最后给话题化标题'],
      sample: '适合先生成母稿逻辑，再改成更口语、可种草的小红书笔记。',
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

function buildMentorReply(prompt, idea, readiness) {
  const brand = selectedWorkbenchBrand.value.name
  const product = selectedWorkbenchProduct.value?.name
  const skill = selectedSkillProfile.value.name
  const framework = skillWritingFramework(skill)
  const searchProblem = ideaSession.searchProblem || prompt
  const audience = ideaSession.audience || selectedWorkbenchProduct.value?.audience || selectedWorkbenchBrand.value.audience || '潜在目标用户'
  const scene = ideaSession.scene || '由资料和对话推断的使用场景'
  const structureText = framework.structure.map((item, index) => `${index + 1}. ${item}`).join('；')

  if (isWorkbenchDraftDecision(prompt) || readiness.ready) {
    return `这个想法可以进入母稿了。我建议按「${framework.entry}」来写：核心问题是「${searchProblem}」，资料底座用「${brand}」${product ? `和「${product}」` : ''}，目标读者先按「${audience}」，场景按「${scene}」。\n\n写作思路：${structureText}。\n\n范本方向：标题可以围绕「${searchProblem}」展开，正文先给明确答案，再把「${selectedMaterialSummary()}」转成选择理由和场景建议。这样生成出来不是品牌介绍，而是一篇能被搜索和 AI 问答引用的 GEO 母稿。`
  }

  return `我会先把这个想法当成「${framework.entry}」来处理。基于当前资料，最强的写法不是直接写“${brand} 很好”，而是把用户问题写清楚：例如「${searchProblem}」。\n\n可用资料：${selectedMaterialSummary()}。\n\n建议结构：${structureText}。\n\n现在还差一个决定文章质量的点：${idea?.questions?.[0]?.replace(/^补充/, '') || '明确这篇文章要回答的用户问题'}。你可以直接补一句目标人群、场景或想解决的问题，我会继续把它整理成可生成母稿的范本。`
}

function buildWorkbenchReply(idea) {
  const latestPrompt = latestWorkbenchUserPrompt.value
  const brand = selectedWorkbenchBrand.value.name
  const skill = selectedSkillProfile.value.name
  const product = selectedWorkbenchProduct.value?.name
  const angles = workbenchAngleOptions()
  if (isGreetingMessage(latestPrompt)) {
    return `你好，我在。你可以告诉我想写什么 GEO 母稿，我会基于当前资料和「${skill}」帮你拆思路、定文章入口，再整理成可生成的母稿方向。\n\n如果你还没想清楚，我也可以先帮你从「${brand}」的资料里找选题。比如可以写：${angles.join('；')}。`
  }
  if (isLowSignalIdea(latestPrompt)) {
    return `收到。当前资料底座是「${brand}」${product ? `，商品是「${product}」` : ''}，Skill 是「${skill}」。我不会直接套模板生成，建议先把文章入口定成一个用户真的会搜索/提问的问题。可以从这三个方向里选一个：${angles.join('；')}。你也可以直接说你的目标人群、使用场景或想解决的问题。`
  }
  const readiness = evaluateWorkbenchReadiness()
  return buildMentorReply(latestPrompt, idea, readiness)
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

async function streamAiMessage(text, idea) {
  const message = reactive({ id: Date.now() + 1, role: 'ai', text: '', idea, streaming: true })
  chatMessages.push(message)
  chatStreaming.value = true
  try {
    for (let index = 0; index < text.length;) {
      const chunk = nextStreamChunk(text, index)
      message.text += chunk
      index += chunk.length
      await wait(/[。！？；\n]$/.test(chunk) ? 90 : 22)
    }
  } finally {
    message.streaming = false
    chatStreaming.value = false
  }
}

function buildGatewayMentorMessages(prompt, idea) {
  const recentMessages = chatMessages.slice(-8).map(msg => ({
    role: msg.role === 'user' ? 'user' : 'assistant',
    content: msg.text,
  }))
  const context = {
    brand: {
      name: selectedWorkbenchBrand.value.name,
      positioning: selectedWorkbenchBrand.value.position,
      audience: selectedWorkbenchBrand.value.audience,
      keywords: selectedWorkbenchBrand.value.keywordGroups?.flatMap(group => group.keywords || []) || [],
    },
    product: selectedWorkbenchProduct.value ? {
      name: selectedWorkbenchProduct.value.name,
      audience: selectedWorkbenchProduct.value.audience,
      selling_points: selectedWorkbenchProduct.value.sellingPoints,
      keywords: selectedWorkbenchProduct.value.keywords || [],
    } : null,
    skill: selectedSkillProfile.value,
    hotspot: workbench.hotspot,
    inferred_brief: ideaSession.brief,
    inferred_slots: {
      search_problem: ideaSession.searchProblem,
      audience: ideaSession.audience,
      scene: ideaSession.scene,
      tone: ideaSession.tone,
    },
    latest_prompt: prompt,
    local_readiness: evaluateWorkbenchReadiness(prompt),
    local_reference_reply: buildWorkbenchReply(idea),
  }
  return [
    {
      role: 'system',
      content: [
        '你是 AI GEO 写作导师，不是模板生成器。',
        '你的任务是根据用户想法、品牌资料、商品资料、Skill 和热点，边聊边把模糊想法拆解、放大、澄清、收敛成可生成母稿的方向。',
        '用户打招呼时要自然回应，并主动询问要写什么，或基于资料猜测可写方向。',
        '不要重复机械追问；如果用户已经给了目标人群或场景，要承认并继续推进。',
        '回复要像专家：给具体文章入口、搜索/AI 问答问题、结构建议、资料如何使用。不要输出 JSON，不要说自己基于规则。',
        '回复必须有清晰格式：短段落、加粗小标题、编号或项目符号列表；每个要点独立换行，不要把所有内容挤成一整段。',
        '如果信息足够，明确告诉用户可以生成母稿；如果不足，只问一个最关键的澄清点。',
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
  const message = reactive({ id: Date.now() + 1, role: 'ai', text: '', idea, streaming: true })
  chatMessages.push(message)
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
        messages: buildGatewayMentorMessages(prompt, idea),
      },
    }, {
      onDelta(delta) {
        message.text += delta
      },
      onFinal(event) {
        if (!message.text && event.text) message.text = event.text
      },
      onError(event) {
        throw new Error(event.error_message || 'AI Gateway 流式调用失败')
      },
    })
  } catch (error) {
    const index = chatMessages.findIndex(item => item.id === message.id)
    if (index >= 0) chatMessages.splice(index, 1)
    await streamAiMessage(buildWorkbenchReply(idea), idea)
    showToast(error?.message || 'AI 流式沟通失败，已切回本地兜底')
  } finally {
    message.streaming = false
    chatStreaming.value = false
  }
}

async function sendWorkbenchMessage() {
  const prompt = String(workbench.prompt || '').trim()
  if (!prompt || chatStreaming.value) return
  chatMessages.push({ id: Date.now(), role: 'user', text: prompt })
  const idea = buildIdeaAnalysis(prompt)
  workbench.prompt = ''
  await streamGatewayMentorReply(prompt, idea)
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
    .replace(/\*\*([^*]+)\*\*/g, '<strong>$1</strong>')
    .replace(/`([^`]+)`/g, '<code>$1</code>')
}

function formatChatMessage(text) {
  const source = String(text || '').trim()
  if (!source) return '<p></p>'
  const blocks = source
    .replace(/([。！？])\s*(?=(?:\d+[.、]|[-•])\s*)/g, '$1\n')
    .replace(/\s+(?=(?:\d+[.、]|[-•])\s*[^\n])/g, '\n')
    .split(/\n{2,}/)
    .map(block => block.trim())
    .filter(Boolean)
  return blocks.map(block => {
    const lines = block.split('\n').map(line => line.trim()).filter(Boolean)
    const numbered = lines.every(line => /^\d+[.、]\s*/.test(line))
    const bulleted = lines.every(line => /^[-•]\s*/.test(line))
    if (numbered || bulleted) {
      const tag = numbered ? 'ol' : 'ul'
      const items = lines.map(line => `<li>${inlineMarkdown(line.replace(/^(?:\d+[.、]|[-•])\s*/, ''))}</li>`).join('')
      return `<${tag}>${items}</${tag}>`
    }
    return `<p>${inlineMarkdown(lines.join('\n')).replace(/\n/g, '<br>')}</p>`
  }).join('')
}
async function generateDraftFromChat() {
  if (!workbenchReadiness.value.ready) {
    await streamAiMessage(buildWorkbenchReply(), null)
    return
  }
  loading.action = true
  const product = selectedWorkbenchProduct.value
  const skill = workbench.skill || '通用 GEO 母稿 Skill'
  const typedPrompt = String(workbench.prompt || '').trim()
  const latestUserPrompt = latestWorkbenchUserPrompt.value
  const prompt = typedPrompt || latestUserPrompt || `请基于 ${selectedWorkbenchBrand.value.name}${product ? ` 的 ${product.name}` : ''}，生成一篇面向「${selectedSkillProfile.value.goal}」的 GEO 文章。`
  if (typedPrompt || !latestUserPrompt) {
    chatMessages.push({ id: Date.now(), role: 'user', text: prompt })
  }
  try {
    const draft = await generateAiGeoDraft({
      brand_id: workbench.brandId ? Number(workbench.brandId) : undefined,
      product_id: workbench.productId ? Number(workbench.productId) : undefined,
      skill,
      hotspot_id: workbench.hotspot?.id,
      prompt: [
        '你是 AI GEO 母稿协作编辑。请根据用户选择的资料、Skill 和聊天中收敛出的 brief 生成一篇可作为多渠道源稿的 GEO 母稿。',
        '生成目标：先回答用户真实搜索/AI 问答问题，再自然带出品牌与商品资料；避免空泛品牌介绍和硬广。',
        `收敛 brief：${ideaSession.brief || prompt}`,
        `用户原始想法：${prompt}`,
        `本轮对话：${workbenchUserMessages.value.map(msg => msg.text).join(' / ') || prompt}`,
        `使用 Skill：${skill}`,
        `品牌定位：${selectedWorkbenchBrand.value.position || '未维护'}`,
        product ? `商品卖点：${product.sellingPoints || product.name}` : '',
        ideaSession.audience ? `目标人群：${ideaSession.audience}` : '',
        ideaSession.scene ? `场景：${ideaSession.scene}` : '',
        ideaSession.tone ? `口吻：${ideaSession.tone}` : '',
        `关键词：${workbenchKeywords.value.join('、') || '未维护'}`,
        workbench.hotspot ? `引用热点：${workbench.hotspot.title}` : '',
        '输出要求：标题明确、摘要可发布、正文有问题拆解/选择理由/场景建议/结论，关键词可供渠道改写复用。',
      ].filter(Boolean).join('\n'),
    })
    Object.assign(editingDraft, {
      id: draft.id,
      title: draft.title,
      summary: draft.summary || '',
      body: draft.body,
      keywordsText: parseJsonArray(draft.keywords).join(', '),
      status: draftStatusLabel(draft.audit_status),
      source: sourceLabel(draft.source),
    })
    await loadAiGeoData()
    chatMessages.push({ id: Date.now() + 1, role: 'ai', text: `已按「${selectedSkillProfile.value.name}」生成母稿，并写入右侧编辑器。你可以继续要求我强化人群、卖点、FAQ 或渠道语气。` })
    workbench.prompt = ''
  } catch (error) {
    showToast(error?.message || '母稿生成失败')
  } finally {
    loading.action = false
  }
}
async function saveDraft() {
  loading.action = true
  try {
    const draft = await createAiGeoDraft({
      brand_id: workbench.brandId ? Number(workbench.brandId) : undefined,
      product_id: workbench.productId ? Number(workbench.productId) : undefined,
      title: editingDraft.title,
      summary: editingDraft.summary,
      body: editingDraft.body,
      keywords: String(editingDraft.keywordsText || '').split(',').map(s => s.trim()).filter(Boolean),
      source: 'manual',
    })
    editingDraft.id = draft.id
    await loadAiGeoData()
    showToast('母稿草稿已保存')
  } catch (error) {
    showToast(error?.message || '母稿保存失败')
  } finally {
    loading.action = false
  }
}
async function submitDraftFlow() {
  if (!editingDraft.id) {
    await saveDraft()
  }
  if (!editingDraft.id) return
  loading.action = true
  try {
    const mode = modeConfig.draftAuditMode
    const draftId = Number(editingDraft.id)
    if (mode === 'AI审核 + 人工确认') {
      const suggestion = await runDraftAudit(draftId)
      await submitDraftIfNeeded(draftId, editingDraft)
      showToast(`AI审核报告已生成，${auditOpinion(suggestion)}，待人工确认`)
    } else if (mode === 'AI审核') {
      const suggestion = await runDraftAudit(draftId)
      await submitDraftIfNeeded(draftId, editingDraft)
      const opinion = auditOpinion(suggestion)
      if (Boolean(suggestion.passed ?? suggestion.Passed)) {
        const approved = await approveAiGeoDraft(draftId, { opinion: `AI审核通过：${opinion}` })
        editingDraft.status = draftStatusLabel(approved.audit_status)
        await generateChannelsForDraft({ ...editingDraft, id: draftId, channels: [] }, { silent: true, throwOnError: true })
        showToast('AI审核通过，已自动生成渠道内容')
      } else {
        const rejected = await rejectAiGeoDraft(draftId, { opinion: `AI审核未通过：${opinion}` })
        editingDraft.status = draftStatusLabel(rejected.audit_status)
        showToast('AI审核未通过，已留下审核意见')
      }
    } else if (mode === '人工审核') {
      await submitDraftIfNeeded(draftId, editingDraft)
      showToast('母稿已提交，等待人工审核')
    } else {
      await submitDraftIfNeeded(draftId, editingDraft)
      const approved = await approveAiGeoDraft(draftId, { opinion: '无需审核，提交后自动通过。' })
      editingDraft.status = draftStatusLabel(approved.audit_status)
      await generateChannelsForDraft({ ...editingDraft, id: draftId, channels: [] }, { silent: true, throwOnError: true })
      showToast('已提交并跳过审核，渠道内容已生成')
    }
    await loadAiGeoData()
  } catch (error) {
    showToast(error?.message || '提交失败')
  } finally {
    loading.action = false
  }
}
function editDraftInWorkbench(draft) { Object.assign(editingDraft, { ...draft, keywordsText: '通勤穿搭, 商品种草' }); activeMenu.value = 'workbench'; previewMode.value = 'edit' }
async function approveDraft(draft) {
  loading.action = true
  try {
    await submitDraftIfNeeded(Number(draft.id), draft)
    const opinion = window.prompt('请输入人工审核意见', '人工确认通过。') || '人工确认通过。'
    const updated = await approveAiGeoDraft(Number(draft.id), { opinion })
    draft.status = draftStatusLabel(updated.audit_status)
    await loadAiGeoData()
    showToast('母稿审核通过')
  } catch (error) {
    showToast(error?.message || '母稿审核失败')
  } finally {
    loading.action = false
  }
}
async function rejectDraft(draft) {
  loading.action = true
  try {
    await submitDraftIfNeeded(Number(draft.id), draft)
    const opinion = window.prompt('请输入驳回原因', '内容需要调整，请修改后再提交。') || '内容需要调整，请修改后再提交。'
    const updated = await rejectAiGeoDraft(Number(draft.id), { opinion })
    draft.status = draftStatusLabel(updated.audit_status)
    await loadAiGeoData()
    showToast('母稿已驳回并留下审核意见')
  } catch (error) {
    showToast(error?.message || '母稿驳回失败')
  } finally {
    loading.action = false
  }
}
async function generateDraftAudit(draft) {
  const targetId = Number(draft?.id || editingDraft.id)
  if (!targetId) {
    showToast('请先保存母稿再生成审核建议')
    return
  }
  loading.action = true
  try {
    await runDraftAudit(targetId)
    showToast('母稿审核建议已生成')
  } catch (error) {
    showToast(error?.message || '母稿审核建议生成失败')
  } finally {
    loading.action = false
  }
}
async function runDraftAudit(targetId) {
  const suggestion = await generateAiGeoDraftAuditSuggestion(targetId)
  await fetchAiGeoDraftAuditSuggestions(targetId, { limit: 5 })
  applyAuditPanel(draftAuditPanel, suggestion)
  return suggestion
}
async function submitDraftIfNeeded(targetId, draft) {
  const status = draft?.rawStatus || draft?.audit_status || draft?.status || editingDraft.status
  if (['pending', 'approved', '待审核', '已通过'].includes(status)) return draft
  const submitted = await submitAiGeoDraft(Number(targetId))
  if (draft) {
    draft.rawStatus = submitted.audit_status
    draft.status = draftStatusLabel(submitted.audit_status)
  }
  return submitted
}
function auditOpinion(suggestion) {
  return suggestion?.summary || suggestion?.Summary || '审核报告已生成'
}
async function generateChannelsForDraft(draft, options = {}) {
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
    target.status = '已生成渠道版本'
    if (!options.silent) showToast('已生成渠道版本')
  } catch (error) {
    if (options.throwOnError) throw error
    showToast(error?.message || '生成渠道内容失败')
  } finally {
    loading.action = false
  }
}
function openChannelEditor(draft, channel) {
  selectedDraft.value = draft
  Object.assign(selectedChannel, channel)
  Object.assign(channelAuditPanel, { summary: '', risk: '未审核', passed: false, items: [] })
  drawer.type = 'channelEditor'
  drawer.title = `${channel.channel} 内容编辑`
  channelPreviewMode.value = 'edit'
}
async function confirmChannel(channel) {
  const contentId = Number(channel?.id)
  if (!contentId) {
    showToast('请先生成渠道内容')
    return
  }
  loading.action = true
  try {
    const approved = await approveAiGeoChannelContent(contentId, { opinion: '人工确认渠道内容通过。' })
    Object.assign(channel, channelContentFromApi(approved))
    syncSelectedChannel(channel)
    showToast('渠道内容已确认')
  } catch (error) {
    showToast(error?.message || '渠道内容确认失败')
  } finally {
    loading.action = false
  }
}

async function confirmSelectedChannel() {
  const saved = await persistSelectedChannel()
  if (!saved) return
  await confirmChannel(selectedChannel)
}
function optimizeChannel(type) { selectedChannel.body += `\n\nAI 局部优化：${type}。`; if (type === '生成话题标签') selectedChannel.tags = '#通勤穿搭 #法式穿搭 #小个子穿搭'; showToast(type + '完成') }
async function rejectSelectedChannel() {
  const contentId = Number(selectedChannel?.id)
  if (!contentId) {
    showToast('请先生成渠道内容')
    return
  }
  loading.action = true
  try {
    const rejected = await rejectAiGeoChannelContent(contentId, { opinion: '人工驳回：需要继续优化渠道表达。' })
    Object.assign(selectedChannel, channelContentFromApi(rejected))
    syncSelectedChannel(selectedChannel)
    showToast('渠道内容已驳回，审核意见已保留')
  } catch (error) {
    showToast(error?.message || '渠道内容驳回失败')
  } finally {
    loading.action = false
  }
}

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
async function generateChannelAudit(channel) {
  const contentId = Number(channel?.id)
  if (!contentId) {
    showToast('请先生成渠道内容再运行审核')
    return
  }
  loading.action = true
  try {
    const suggestion = await generateAiGeoChannelContentAuditSuggestion(contentId)
    await fetchAiGeoChannelContentAuditSuggestions(contentId, { limit: 5 })
    applyAuditPanel(channelAuditPanel, suggestion)
    showToast('渠道内容审核建议已生成')
  } catch (error) {
    showToast(error?.message || '渠道审核建议生成失败')
  } finally {
    loading.action = false
  }
}
async function addChannelToPlan(draft, channel) {
  const profile = channelProfiles.find(item => item.name === channel.channel) || channelProfiles[0]
  if (!profile?.id) {
    showToast('请先配置可发布渠道')
    return
  }
  loading.action = true
  try {
    await createAiGeoPublishPlan({
      channel_content_id: Number(channel.id || draft?.channelContentId || 1),
      channel_id: Number(profile.id),
      scheduled_at: new Date(`${planDate.value || '2026-05-20'}T18:00:00+08:00`).toISOString(),
      publish_method: profile.method || 'manual',
      automation_level: profile.level || 'manual',
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
function addSelectedChannelToPlan() { addChannelToPlan(selectedDraft.value || drafts[0], selectedChannel) }
async function createPlan() {
  const draft = drafts.find(d => d.id === Number(newPlanDraftId.value)) || drafts[0]
  const channel = channelProfiles.find(c => c.name === newPlanChannel.value) || channelProfiles[0]
  if (!draft || !channel?.id) {
    showToast('请先准备母稿和渠道')
    return
  }
  loading.action = true
  try {
    await createAiGeoPublishPlan({
      channel_content_id: Number(draft.channels?.[0]?.id || 1),
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
function runPlanAudit(plan) {
  plan.channelAudit = '已通过'
  if (plan.materialStatus === '凭证异常') plan.materialStatus = '完整'
  showToast(`已完成 ${plan.channel} 渠道审核`)
}
function preCheckPlan(plan) {
  if (plan.accountStatus === '阻断') {
    showToast('账号状态阻断，请先处理账号授权')
    return
  }
  if (plan.materialStatus === '缺封面') {
    showToast('素材缺封面，请补齐后重试')
    return
  }
  plan.accountStatus = plan.method === 'Agent 执行' ? '待人工确认' : '可发布'
  plan.publishStatus = plan.accountStatus === '可发布' && plan.channelAudit === '已通过' && plan.materialStatus === '完整' ? '可发布' : '待审核'
  showToast('前置检查完成')
}
function completePlan(plan) {
  if (plan.publishStatus === '阻断' || plan.accountStatus === '阻断') {
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
  loading.action = true
  try {
    const updated = await updateAiGeoPublishPlanStatus(Number(plan.id), payload)
    plan.status = publishStatusLabel(updated.status)
    plan.link = updated.published_url || plan.link || ''
    await loadAiGeoData()
    showToast(successText)
  } catch (error) {
    showToast(error?.message || '发布计划状态更新失败')
  } finally {
    loading.action = false
  }
}

function applyAuditPanel(panel, suggestion) {
  const rawSuggestions = suggestion.suggestion_json || suggestion.SuggestionJSON || '[]'
  const items = parseJsonArray(rawSuggestions)
    .map(item => {
      if (typeof item === 'string') return item
      return item?.content || item?.text || item?.suggestion || item?.message || item?.title || ''
    })
    .filter(Boolean)
  panel.summary = suggestion.summary || suggestion.Summary || '审核完成，暂无额外摘要'
  panel.risk = auditRiskLabel(suggestion.risk_level || suggestion.RiskLevel)
  panel.passed = Boolean(suggestion.passed ?? suggestion.Passed)
  panel.items = items.length ? items : ['未发现需要阻断的问题，建议进入人工确认。']
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

function auditRiskLabel(risk) {
  return {
    low: '低风险',
    medium: '中风险',
    high: '高风险',
  }[risk] || risk || '未评级'
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

function channelIcon(name) {
  if (name?.includes('小红书')) return '📕'
  if (name?.includes('抖音')) return '🎵'
  if (name?.includes('微信')) return '💬'
  if (name?.includes('知乎')) return '💡'
  if (name?.includes('微博')) return '📢'
  if (name?.includes('百家')) return '📰'
  if (name?.includes('站')) return '🌐'
  return '📣'
}

function sourceLabel(source) {
  if (source === 'ai_workbench') return '智能生成'
  if (source === 'manual') return '人工创作'
  return source || '人工创作'
}

function draftStatusLabel(status) {
  return {
    draft: '草稿',
    pending: '待审核',
    approved: '已通过',
    rejected: '已驳回',
  }[status] || status || '草稿'
}

function channelContentStatusLabel(status) {
  return {
    pending: '待确认',
    approved: '已确认',
    rejected: '已驳回',
  }[status] || status || '待确认'
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
