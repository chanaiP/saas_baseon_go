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
          <MetricCard title="资料完整度" :value="currentBrand.completeness + '%'" desc="品牌 / 商品 / SKU 汇总" />
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
          :summary="channelSummary"
          compact
          @open-plans="goChannelManagement"
          @configure="configureChannel"
          @test="testChannel"
        />
      </section>

      <section v-if="activeMenu === 'workbench'" class="page workbench-page">
        <div class="workspace-layout">
          <div class="chat-panel panel">
            <div class="panel-header compact">
              <div>
                <h3>AI 创作对话</h3>
                <p>人工创作为主；品牌、商品、Skill、热点均为可选</p>
              </div>
            </div>
            <div class="loaders">
              <label>品牌<select v-model="workbench.brandId"><option value="">不选择</option><option v-for="b in brands" :value="b.id" :key="b.id">{{ b.name }}</option></select></label>
              <label>商品<select v-model="workbench.productId"><option value="">不选择</option><option v-for="p in currentBrand.products" :value="p.id" :key="p.id">{{ p.name }}</option></select></label>
              <label>Skill<select v-model="workbench.skill"><option value="">不选择</option><option v-for="s in skills" :key="s">{{ s }}</option></select></label>
              <label>热点<select v-model="selectedHotspotId">
                <option value="">不选择</option>
                <option v-for="h in hotspots" :key="h.id" :value="h.id">{{ h.title }} · {{ h.platform }}</option>
                <option value="__more__">热点库 / 手动添加…</option>
              </select></label>
            </div>
            <div class="chat-log">
              <div v-for="msg in chatMessages" :key="msg.id" :class="['bubble', msg.role]">
                <p>{{ msg.text }}</p>
              </div>
            </div>
            <div class="chat-input">
              <textarea v-model="workbench.prompt" placeholder="输入你的创作想法；不选资料和 Skill 时，AI 将自由发挥生成母稿"></textarea>
              <button class="btn primary" @click="generateDraftFromChat">生成母稿</button>
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
              <button class="btn ghost" @click="saveDraft">保存草稿</button>
              <button class="btn primary" @click="submitDraftAudit">提交审核</button>
              <button class="btn dark" @click="generateChannelsForDraft(editingDraft)">生成渠道版本</button>
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
                      <button class="btn small ghost" @click="approveDraft(draft)">审核通过</button>
                      <button class="btn small dark" @click="generateChannelsForDraft(draft)">生成渠道</button>
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
                      <button class="btn small ghost" @click="confirmChannel(channel)">确认</button>
                      <button class="btn small dark" @click="addChannelToPlan(draft, channel)">加入发布计划</button>
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
                  <button class="btn primary" @click="openNewPlanModal">新建发布计划</button>
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
                        <button type="button" class="btn small ghost" @click="runPlanAudit(plan)">运行审核</button>
                        <button type="button" class="btn small ghost" @click="preCheckPlan(plan)">前置检查</button>
                        <button type="button" class="btn small dark" @click="completePlan(plan)">完成</button>
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
          :summary="channelSummary"
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
            <button class="btn primary" @click="showToast('新增渠道账号（原型）')">新增账号</button>
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
                    <button class="btn small ghost" @click="showToast(`配置 ${account.accountName}`)">配置</button>
                    <button class="btn small" @click="showToast(`重新授权 ${account.accountName}`)">授权</button>
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
              <button class="btn ghost" @click="openImportModal">动态导入</button>
              <button class="btn primary" @click="openBrandModal">新增品牌</button>
            </div>
          </div>
          <div class="brand-grid">
            <div v-for="brand in brands" :key="brand.id" :class="['brand-card', { active: selectedBrandId === brand.id }]" @click="selectedBrandId = brand.id">
              <h3>{{ brand.name }}</h3>
              <p>{{ brand.position }}</p>
              <div class="brand-stats"><span>{{ brand.products.length }} 商品</span><span>{{ countSkus(brand) }} SKU</span><span>{{ brand.completeness }}%</span></div>
              <div class="progress"><span :style="{ width: brand.completeness + '%' }"></span></div>
            </div>
          </div>
        </div>

        <div class="panel brand-detail">
          <div class="panel-header compact">
            <div>
              <h3>{{ currentBrand.name }} 品牌资料卡</h3>
              <p>{{ currentBrand.position }} · {{ currentBrand.audience }} · {{ currentBrand.priceBand }}</p>
            </div>
            <div class="row-actions">
              <button class="btn small ghost" @click="openBrandModal(currentBrand)">编辑品牌</button>
              <button class="btn small dark" @click="generateBrandKeywords">生成关键词</button>
            </div>
          </div>
          <div class="tabs">
            <button v-for="tab in dataTabs" :key="tab" :class="{ active: dataTab === tab }" @click="dataTab = tab">{{ tab }}</button>
          </div>
          <div v-if="dataTab === '商品资料'" class="products-table">
            <table class="table">
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
          <button class="btn primary full" @click="createPlan">生成发布任务</button>
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
          </div>
        </div>
        <ChannelPreview v-else :mode="channelPreviewMode" :channel="selectedChannel" />
        <div class="drawer-actions">
          <button class="btn ghost" @click="regenerateChannel">重新生成</button>
          <button class="btn primary" @click="confirmSelectedChannel">确认渠道内容</button>
          <button class="btn dark" @click="addSelectedChannelToPlan">加入发布计划</button>
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
          <button class="btn ghost" @click="addCompetitor">新增竞品</button>
        </div>
        <div v-if="productTab === '关键词/内容'" class="keyword-cloud"><span v-for="kw in selectedProduct.keywords" :key="kw" class="keyword">{{ kw }}</span></div>
      </div>
    </aside>

    <div v-if="toast.show" class="toast">{{ toast.text }}</div>
  </div>
</template>

<script setup>
import { computed, defineComponent, h, reactive, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import ChannelManagementPanel from '../components/ChannelManagementPanel.vue'

const MetricCard = defineComponent({
  props: ['title', 'value', 'desc'],
  setup(props) {
    return () => h('div', { class: 'metric-card' }, [h('span', props.title), h('strong', props.value), h('p', props.desc)])
  }
})

const InfoBlock = defineComponent({
  props: ['title', 'value'],
  setup(props) {
    return () => h('div', { class: 'info-block' }, [h('span', props.title), h('strong', props.value)])
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
const dataTab = ref('商品资料')
const dataTabs = ['品牌信息', '商品资料', '关键词', '素材']
const productTab = ref('公共资料')
const productTabs = ['公共资料', 'SKU明细', 'SKU覆盖资料', '竞品信息', '关键词/内容']
const previewMode = ref('edit')
const channelPreviewMode = ref('edit')
const draftDate = ref('')
const planDate = ref('2026-05-20')
const draftFilter = ref('全部状态')

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

const currentBrand = computed(() => brands.find(b => b.id === Number(selectedBrandId.value)) || brands[0])
const productCount = computed(() => brands.reduce((sum, b) => sum + b.products.length, 0))
const skuCount = computed(() => brands.reduce((sum, b) => sum + countSkus(b), 0))
function countSkus(brand) { return brand.products.reduce((sum, p) => sum + p.skus.length, 0) }

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

const channelSummary = computed(() => ({
  enabled: channelProfiles.filter(c => c.enabled).length,
  publishable: channelProfiles.filter(c => c.status === '可发布').length,
  pending: channelProfiles.filter(c => ['待授权', '未配置'].includes(c.status)).length
}))

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

const calendarDays = computed(() => {
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
  { label: '今日母稿', value: drafts.filter(d => d.date === '2026-05-20').length, desc: '含草稿和已审核' },
  { label: '待审核', value: drafts.filter(d => d.status === '待审核').length, desc: '需人工确认' },
  { label: '渠道内容', value: drafts.reduce((s, d) => s + d.channels.length, 0), desc: '已生成版本' },
  { label: '今日发布', value: plans.filter(p => p.date === '2026-05-20').length, desc: '发布计划任务' }
])

const pendingTasks = computed(() => {
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
const chatMessages = reactive([
  { id: 1, role: 'ai', text: '可以直接输入想法生成母稿；品牌、商品、Skill、热点都不是必选。' }
])
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

const hotspots = reactive([
  { id: 1, title: '通勤穿搭回归轻量化', summary: '春夏职场穿搭更强调舒适、轻正式和可复用单品。', platform: '小红书', heat: 86, risk: '低风险' },
  { id: 2, title: '小个子显高穿搭讨论升温', summary: '多平台讨论小个子女生如何通过腰线和裙长优化比例。', platform: '知乎', heat: 74, risk: '低风险' },
  { id: 3, title: '轻法式风格持续走热', summary: '轻法式关键词在穿搭内容里持续出现，适合做品牌风格承接。', platform: '抖音', heat: 69, risk: '中风险' }
])
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
function openBrandModal() { showToast('新增/编辑品牌弹窗已触发') }
function openNewPlanModal() { modal.type = 'newPlan'; modal.title = '新建发布计划' }
function mockImport() { closeModal(); showToast('已完成字段映射校验，导入 2 个商品、6 个 SKU、3 条竞品信息') }
function generateBrandKeywords() { showToast('已基于品牌、商品、竞品信息生成关键词') }
function pickCoverImage() {
  editingDraft.coverImage = sampleCoverImage
  showToast('已选择封面图（原型演示）')
}
function openHotspotDrawer() { drawer.type = 'hotspot'; drawer.title = '引用热点' }
function addManualHotspot() { hotspots.unshift({ id: Date.now(), title: '手动录入热点', summary: '运营手动录入的热点摘要，可用于轻引用或选题引用。', platform: '手动', heat: 50, risk: '待判断' }); showToast('已添加手动热点') }
function useHotspot(hot, mode) {
  workbench.hotspot = hot
  chatMessages.push({ id: Date.now(), role: 'ai', text: mode === 'dialog' ? `已引用热点：${hot.title}` : `借势角度：围绕「${hot.title}」做轻引用，不夸大热点关系。` })
  closeDrawer()
}
function generateDraftFromChat() {
  const product = currentBrand.value.products.find(p => p.id === Number(workbench.productId))
  const skill = workbench.skill || '无 Skill，自由发挥'
  editingDraft.title = product ? `${product.name}怎么写出种草感？` : '根据你的想法生成的新母稿'
  editingDraft.summary = `生成方式：人工创作；Skill：${skill}；${workbench.hotspot ? '引用热点：' + workbench.hotspot.title : '未引用热点'}。`
  editingDraft.body = `这是一篇基于当前输入生成的母稿。\n${product ? `关联商品：${product.name}，核心卖点：${product.sellingPoints}` : '未装载商品资料，因此内容会以用户想法为主。'}\n${workbench.hotspot ? `热点轻引用：${workbench.hotspot.summary}` : '没有引用热点，内容保持常规表达。'}\n后续可以继续通过 AI 重构、人工编辑，并生成渠道版本。`
  editingDraft.status = '草稿'
  editingDraft.source = '人工创作'
  chatMessages.push({ id: Date.now(), role: 'ai', text: '母稿已生成到右侧编辑器，可以编辑、预览或提交审核。' })
}
function saveDraft() { showToast('母稿草稿已保存') }
function submitDraftAudit() { editingDraft.status = '待审核'; showToast('母稿已提交审核') }
function editDraftInWorkbench(draft) { Object.assign(editingDraft, { ...draft, keywordsText: '通勤穿搭, 商品种草' }); activeMenu.value = 'workbench'; previewMode.value = 'edit' }
function approveDraft(draft) { draft.status = '已通过'; showToast('母稿审核通过') }
function generateChannelsForDraft(draft) {
  const target = draft.id ? draft : drafts[0]
  if (!target.channels.find(c => c.channel === '微信公众号')) target.channels.push({ id: Date.now(), channel: '微信公众号', title: target.title, body: `${target.body}\n\n公众号版本：结构更完整，适合阅读和转发。`, tags: '', seoTitle: '', script: '', status: '已生成' })
  if (!target.channels.find(c => c.channel === '知乎')) target.channels.push({ id: Date.now() + 1, channel: '知乎', title: target.title, body: `${target.body}\n\n知乎版本：增加理性分析，减少促销语气。`, tags: '', seoTitle: '', script: '', status: '已生成' })
  target.status = '已生成渠道版本'
  showToast('已生成渠道版本')
}
function openChannelEditor(draft, channel) { selectedDraft.value = draft; Object.assign(selectedChannel, channel); drawer.type = 'channelEditor'; drawer.title = `${channel.channel} 内容编辑`; channelPreviewMode.value = 'edit' }
function confirmChannel(channel) { channel.status = '已确认'; showToast('渠道内容已确认') }
function confirmSelectedChannel() { selectedChannel.status = '已确认'; showToast('渠道内容已确认') }
function optimizeChannel(type) { selectedChannel.body += `\n\nAI 局部优化：${type}。`; if (type === '生成话题标签') selectedChannel.tags = '#通勤穿搭 #法式穿搭 #小个子穿搭'; showToast(type + '完成') }
function regenerateChannel() { selectedChannel.body = selectedChannel.body + '\n\n已基于最新母稿重新生成渠道表达。'; showToast('渠道内容已重新生成') }
function addChannelToPlan(draft, channel) { plans.push({ id: Date.now(), date: '2026-05-20', time: '18:00', title: channel.title, channel: channel.channel, method: channel.channel === '独立站' ? '渠道 API' : 'Agent 执行', level: channel.channel === '独立站' ? '全自动' : '半自动', skill: channel.channel === '小红书' ? '小红书图文发布 Skill' : '', risk: '发布前校验 + 异常人工接管', owner: '运营A', status: '已排期', link: '' }); channel.status = '已加入发布计划'; showToast('已加入发布计划') }
function addSelectedChannelToPlan() { addChannelToPlan(selectedDraft.value || drafts[0], selectedChannel) }
function createPlan() {
  const draft = drafts.find(d => d.id === Number(newPlanDraftId.value)) || drafts[0]
  plans.push({ id: Date.now(), date: newPlanTime.value.slice(0, 10), time: newPlanTime.value.slice(11, 16), title: draft.title, channel: newPlanChannel.value, method: newPlanMethod.value, level: newPlanLevel.value, skill: newPlanMethod.value === 'Agent 执行' ? `${newPlanChannel.value}发布 Skill` : '', risk: newPlanMethod.value === 'Agent 执行' ? '频率限制 + 人工接管' : '发布前校验', owner: '运营A', status: '已排期', link: '' })
  closeModal(); showToast('发布计划已创建')
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
function adjustPlan(plan) { plan.time = '20:00'; showToast('已调整计划时间为 20:00') }
function markPublished(plan) { plan.status = '已发布'; plan.link = plan.link || 'https://example.com/post'; showToast('已标记为已发布') }
function fillPublishLink(plan) { plan.link = 'https://example.com/published-link'; showToast('发布链接已回填') }
function markFailed(plan) { plan.status = '发布失败'; showToast('已标记失败，等待重试或人工接管') }
function openProductDrawer(product) { Object.keys(selectedProduct).forEach(k => delete selectedProduct[k]); Object.assign(selectedProduct, product); drawer.type = 'product'; drawer.title = '商品资料卡'; productTab.value = '公共资料' }
function addCompetitor() { selectedProduct.competitors.push({ brand: '新增竞品', name: '竞品商品', price: '待录入', point: '待录入', diff: '待分析', angle: '待生成', link: 'https://example.com' }); showToast('已新增竞品信息') }
</script>

<style src="../styles.css"></style>
