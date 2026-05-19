<template>
  <div class="app-shell no-page-shell" style="background: #0f172a;">
    <main class="main" style="background: #0f172a;">
      <header v-if="page === 'overview'" class="topbar">
        <div>
          <span class="overview-tag">Integration Admin Console</span>
          <h2>第三方集成中心总览</h2>
          <p>统一管理接入平台、服务商应用、租户连接、同步任务、配额限流与异常闭环。</p>
        </div>
        <div class="topbar-actions">
          <button class="ghost-btn" @click="globalSearchOpen = true">全局搜索</button>
          <button v-if="canRunHealthCheck" class="ghost-btn" @click="runHealthCheck('全局')">全局连通性检测</button>
        </div>
      </header>

      <div v-if="backendState.loading || backendState.error" :class="['data-state-banner', backendState.error ? 'error' : 'loading']">
        {{ backendState.error || '正在加载第三方集成中心数据...' }}
      </div>

      <section v-if="page === 'overview'" class="page-section">
        <div class="metric-grid">
          <article v-for="metric in overviewMetrics" :key="metric.label" class="metric-card">
            <span>{{ metric.label }}</span>
            <strong>{{ metric.value }}</strong>
            <p :class="metric.trend >= 0 ? 'up' : 'down'">{{ metric.trendText }}</p>
          </article>
        </div>

        <div class="content-grid two">
          <section class="panel">
            <div class="panel-head">
              <div>
                <h3>平台运行概览</h3>
                <p>按平台聚合应用、能力、连接实例、异常与调用量。</p>
              </div>
              <button class="text-btn" @click="switchPage('platforms')">查看接入平台</button>
            </div>
            <div class="platform-health-list">
              <EmptyState v-if="!platformsSorted.length" title="暂无平台运行数据" />
              <article v-for="platform in platformsSorted" :key="platform.id" class="health-row" @click="openPlatformDrawer(platform)">
                <div class="avatar">{{ platform.icon }}</div>
                <div class="health-main">
                  <div class="row-between">
                    <strong>{{ platform.name }}</strong>
                    <StatusBadge :status="platform.status" />
                  </div>
                  <div class="progress-line"><span :style="{ width: platform.successRate + '%' }"></span></div>
                  <p>{{ platform.appCount }} 个应用 · {{ platform.capabilityCount }} 项能力 · {{ platform.connectionCount }} 个连接 · 今日 {{ formatNumber(platform.callsToday) }} 次调用</p>
                </div>
              </article>
            </div>
          </section>

          <section class="panel">
            <div class="panel-head">
              <div>
                <h3>关系模型</h3>
                <p>界面以平台为单位聚合，第四层统一进入右侧抽屉。</p>
              </div>
              <button class="text-btn" @click="switchPage('workspace')">进入工作台</button>
            </div>
            <div class="relation-flow">
              <div><b>平台 - 应用 - 能力</b><span>应用能力连接，定义应用可开放能力</span></div>
              <div><b>平台 - 应用 - 连接实例</b><span>租户通过应用授权生成连接</span></div>
              <div><b>连接实例 - 最终能力</b><span>应用能力 + 授权范围 + 租户选择</span></div>
              <div><b>连接实例 - 同步 / 配额 / 异常</b><span>运行态全部挂到连接实例</span></div>
            </div>
          </section>
        </div>

        <div class="content-grid three">
          <section class="panel">
            <div class="panel-head"><h3>租户调用排行</h3></div>
            <RankingList :items="tenantRanking" />
            <EmptyState v-if="!tenantRanking.length" title="暂无租户调用数据" />
          </section>
          <section class="panel">
            <div class="panel-head"><h3>配额预警</h3></div>
            <div class="quota-mini-list">
              <EmptyState v-if="!quotaUsages.length" title="暂无配额用量数据" />
              <div v-for="usage in quotaUsages.slice(0, 4)" :key="usage.id" class="quota-mini-row">
                <div>
                  <strong>{{ usage.name }}</strong>
                  <p>{{ usage.scope }} · {{ usage.used.toLocaleString() }} / {{ usage.limit.toLocaleString() }}</p>
                </div>
                <StatusBadge :status="usage.rate >= 100 ? 'exceeded' : usage.rate >= 80 ? 'warning' : 'normal'" />
              </div>
            </div>
          </section>
          <section class="panel">
            <div class="panel-head"><h3>待处理异常</h3></div>
            <div class="alert-list compact">
              <EmptyState v-if="!alerts.length" title="暂无待处理异常" />
              <article v-for="alert in alerts.slice(0, 4)" :key="alert.id" class="alert-item">
                <div :class="['alert-dot', alert.level]"></div>
                <div>
                  <strong>{{ alert.title }}</strong>
                  <p>{{ alert.message }}</p>
                  <small>{{ alert.lastAt }} · {{ alert.count }} 次</small>
                </div>
                <StatusBadge :status="alert.status" />
              </article>
            </div>
          </section>
        </div>
      </section>

      <section v-if="page === 'platforms'" class="page-section">
        <div class="toolbar row-between">
          <div class="toolbar-query">
            <div class="search-box"><span>⌕</span><input v-model="platformKeyword" placeholder="搜索平台名称、code、类型" /></div>
            <select v-model="platformTypeFilter"><option value="all">全部类型</option><option value="协同办公">协同办公</option><option value="电商平台">电商平台</option><option value="ERP">ERP</option><option value="CRM">CRM</option><option value="WMS">WMS</option></select>
            <select v-model="platformStatusFilter"><option value="all">全部状态</option><option value="draft">草稿</option><option value="enabled">启用</option><option value="disabled">停用</option><option value="maintenance">维护中</option></select>
          </div>
          <div class="toolbar-actions">
            <button v-if="canManagePlatform" class="primary-btn" @click="openPlatformModal()">新增接入平台</button>
          </div>
        </div>

        <div class="platform-card-grid">
          <EmptyState v-if="!filteredPlatforms.length" title="暂无接入平台" />
          <article v-for="platform in filteredPlatforms" :key="platform.id" class="platform-card">
            <div class="platform-card-head">
              <div class="avatar large">{{ platform.icon }}</div>
              <div>
                <h3>{{ platform.name }}</h3>
                <code>{{ platform.code }}</code>
              </div>
              <StatusBadge :status="platform.status" />
            </div>
            <div class="platform-chip-row">
              <span class="meta-chip">{{ platform.type }}</span>
              <span class="meta-chip">{{ platform.accessType }}</span>
            </div>
            <p class="platform-desc">{{ platform.description }}</p>
            <div class="card-stats four">
              <div><span>应用</span><b>{{ platform.appCount }}</b></div>
              <div><span>能力</span><b>{{ platform.capabilityCount }}</b></div>
              <div><span>连接</span><b>{{ platform.connectionCount }}</b></div>
              <div><span>异常</span><b>{{ platform.alertCount }}</b></div>
            </div>
            <div class="usage-line">
              <div class="row-between">
                <span>成功率</span>
                <strong>{{ platform.callsToday > 0 ? `${platform.successRate}%` : '暂无调用' }}</strong>
              </div>
              <div class="progress-line" :class="{ empty: platform.callsToday <= 0 }"><span :style="{ width: platform.callsToday > 0 ? `${platform.successRate}%` : '0%' }"></span></div>
              <p>今日 {{ formatNumber(platform.callsToday) }} 次调用</p>
            </div>
            <div class="card-actions wrap">
              <button @click="openPlatformDrawer(platform)">查看详情</button>
              <button v-if="canManagePlatform" @click="openPlatformModal(platform)">编辑平台</button>
              <button v-if="canManageApp" @click="jumpWorkspace(platform.id, 'apps')">配置应用</button>
              <button v-if="canManagePlatform" @click="jumpWorkspace(platform.id, 'capabilities')">配置能力</button>
              <button v-if="canRunHealthCheck" @click="runHealthCheck(platform.name)">检测连通性</button>
            </div>
          </article>
        </div>
      </section>

      <section v-if="page === 'workspace'" class="page-section relation-layout">
        <PlatformRail
          :platforms="platformsSorted"
          :selected-id="workspacePlatformId"
          @select="selectWorkspacePlatform"
          @clear="workspacePlatformId = 'all'"
        />

        <div class="relation-main">
          <section class="panel workspace-head">
            <div>
              <p class="eyebrow">集成工作台</p>
              <h3>{{ selectedWorkspacePlatform ? selectedWorkspacePlatform.name : '全部平台' }}</h3>
              <p>管理平台下的应用、平台能力、应用能力连接与运行关系。</p>
            </div>
            <div class="header-metrics">
              <div><span>应用</span><b>{{ workspaceApps.length }}</b></div>
              <div><span>能力</span><b>{{ workspaceCapabilities.length }}</b></div>
              <div><span>连接</span><b>{{ workspaceConnections.length }}</b></div>
              <div><span>异常</span><b>{{ workspaceAlerts.length }}</b></div>
            </div>
          </section>

          <div class="tab-bar">
            <button v-for="tab in workspaceTabs" :key="tab.key" :class="{ active: workspaceTab === tab.key }" @click="workspaceTab = tab.key">{{ tab.label }}</button>
          </div>

          <template v-if="workspaceTab === 'apps'">
            <section class="panel relation-upper">
              <div class="panel-head">
                <div><h3>服务商应用</h3><p>应用基于平台创建，承载密钥、回调、suite_ticket 和能力连接。</p></div>
                <div class="panel-control-split"><div class="toolbar-query compact"><button class="ghost-btn" @click="clearSelectedApp">全部应用</button></div><div class="toolbar-actions"><button v-if="canManageApp" class="primary-btn" @click="openAppModal()">新增应用</button></div></div>
              </div>
              <div class="app-card-grid">
                <EmptyState v-if="!workspaceApps.length" title="暂无服务商应用" />
                <article v-for="app in workspaceApps" :key="app.id" :class="['mini-card selectable', { selected: workspaceSelectedAppId === app.id }]" @click="selectWorkspaceApp(app.id)">
                  <div class="row-between"><strong>{{ app.name }}</strong><StatusBadge :status="app.status" /></div>
                  <p>{{ app.type }} · {{ getPlatformName(app.platformId) }}</p>
                  <div class="mini-tags"><span>{{ app.capabilityCount }} 能力</span><span>{{ app.connectionCount }} 连接</span><span>{{ formatNumber(app.callsToday) }} 调用</span></div>
                </article>
              </div>
            </section>

            <section class="panel relation-lower">
              <div class="panel-head">
                <div><h3>{{ selectedWorkspaceApp ? selectedWorkspaceApp.name + ' 能力清单' : '全部应用 能力清单' }}</h3><p>{{ selectedWorkspaceApp ? '当前应用已接通、已授权、可开放给租户的能力。' : '未选择应用时，展示当前平台下全部应用能力连接。' }}</p></div>
                <button v-if="selectedWorkspaceApp || canRunHealthCheck" class="ghost-btn" @click="selectedWorkspaceApp ? openAppDrawer(selectedWorkspaceApp, 'capabilities') : runHealthCheck('当前应用能力连接')">{{ selectedWorkspaceApp ? '打开应用详情' : '批量检测' }}</button>
              </div>
              <DataTable :columns="appCapabilityColumns" :rows="workspaceAppCapabilityRows" :pagination="sectionPagination('appCapabilities')" min-width="1180px" @page="changeSectionPage('appCapabilities', $event)">
                <template #cell-name="{ row }"><strong>{{ row.name }}</strong><p class="muted">{{ row.code }}</p></template>
                <template #cell-app="{ row }">{{ row.appName }}</template>
                <template #cell-connected="{ row }"><StatusBadge :status="row.connected ? 'connected' : 'not_connected'" /></template>
                <template #cell-auth="{ row }"><StatusBadge :status="row.authStatus" /></template>
                <template #cell-open="{ row }"><SwitchToggle :model-value="row.openToTenant" :disabled="!canManageApp" @update:model-value="toggleAppCapability(row, 'openToTenant')" /></template>
                <template #cell-default="{ row }"><SwitchToggle :model-value="row.defaultEnabled" :disabled="!canManageApp" @update:model-value="toggleAppCapability(row, 'defaultEnabled')" /></template>
                <template #cell-configurable="{ row }"><SwitchToggle :model-value="row.tenantConfigurable" :disabled="!canManageApp" @update:model-value="toggleAppCapability(row, 'tenantConfigurable')" /></template>
                <template #cell-health="{ row }"><StatusBadge :status="row.health" /></template>
                <template #cell-actions="{ row }"><div class="actions"><button @click="openCapabilityDrawer(row)">详情</button><button v-if="canRunHealthCheck" @click="runHealthCheck(row.name)">检测</button></div></template>
              </DataTable>
            </section>
          </template>

          <template v-else-if="workspaceTab === 'capabilities'">
            <section class="panel table-panel">
              <div class="panel-head padded-head">
                <div><h3>平台能力定义</h3><p>这里只定义平台理论能力，不直接生成租户能力。</p></div>
                <button v-if="canManagePlatform" class="primary-btn" @click="openCapabilityModal()">新增平台能力</button>
              </div>
              <DataTable :columns="platformCapabilityColumns" :rows="workspaceCapabilities" :pagination="sectionPagination('capabilities')" min-width="1050px" @page="changeSectionPage('capabilities', $event)">
                <template #cell-name="{ row }"><strong>{{ row.name }}</strong><p class="muted">{{ row.code }}</p></template>
                <template #cell-type="{ row }"><span class="tag">{{ row.type }}</span></template>
                <template #cell-basic="{ row }">{{ row.isBasic ? '是' : '否' }}</template>
                <template #cell-status="{ row }"><StatusBadge :status="row.status" /></template>
                <template #cell-actions="{ row }"><div class="actions"><button v-if="canManagePlatform" @click="openCapabilityModal(row)">编辑</button><button @click="openCapabilityDrawer(row)">详情</button></div></template>
              </DataTable>
            </section>
          </template>

          <template v-else-if="workspaceTab === 'connections'">
            <section class="panel table-panel">
              <div class="panel-head padded-head">
                <div><h3>当前平台连接实例</h3><p>连接实例按应用建立，能力进入连接详情查看。</p></div>
                <button class="ghost-btn" @click="switchPage('tenantConnections')">进入租户连接</button>
              </div>
              <ConnectionTable :rows="workspaceConnections" :can-manage="canManageConnection" @open="openConnectionDrawer" @refresh="refreshConnection" @pause="pauseConnection" @retry="retryConnection" />
            </section>
          </template>
        </div>
      </section>

      <section v-if="page === 'tenantConnections'" :class="['page-section', 'tenant-layout', { 'tenant-portal': isTenantPortal }]">
        <PlatformRail
          v-if="!isTenantPortal"
          :platforms="platformsSorted"
          :selected-id="tenantPlatformId"
          :show-clear="false"
          @select="selectTenantPlatform"
          @clear="clearTenantPlatform"
        />

        <div class="tenant-main">
          <section v-if="!isTenantPortal" class="panel relation-upper fixed-upper">
            <div class="panel-head">
              <div>
                <h3>服务商应用</h3>
                <p>未选中平台时展示全部应用；选中平台后只展示当前平台应用。选中应用后，下方只显示该应用下的连接实例。</p>
              </div>
              <div class="panel-control-split"><div class="toolbar-query compact"><span class="tag">{{ tenantAppsCapabilityTotal }} 能力</span><button class="ghost-btn" @click="clearTenantPlatform">全部应用</button></div><div class="toolbar-actions"><button v-if="canManageApp" class="primary-btn" @click="openAppModal()">新增应用</button></div></div>
            </div>
            <div class="app-list-grid">
              <EmptyState v-if="!tenantApps.length" title="暂无服务商应用" />
              <article
                v-for="app in tenantApps"
                :key="app.id"
                :class="['app-row-card', { selected: tenantAppId === app.id }]"
                @click="selectTenantApp(app.id)"
              >
                <div class="row-between"><strong>{{ app.name }}</strong><StatusBadge :status="app.status" /></div>
                <p>{{ getPlatformName(app.platformId) }} · {{ app.type }}</p>
                <div class="mini-tags"><span>{{ app.connectionCount }} 连接</span><span>异常 {{ app.alertCount }}</span></div>
              </article>
            </div>
          </section>

          <section class="panel table-panel relation-lower">
            <div class="panel-head padded-head">
              <div>
                <h3>租户连接实例</h3>
                <p>{{ tenantConnectionScopeText }}</p>
              </div>
              <div class="toolbar-actions"><button class="ghost-btn" @click="clearTenantFilters">全部连接</button></div>
            </div>
            <ConnectionTable :rows="tenantConnections" :pagination="sectionPagination('connections')" :can-manage="canManageConnection" @page="changeSectionPage('connections', $event)" @open="openConnectionDrawer" @refresh="refreshConnection" @pause="pauseConnection" @retry="retryConnection" />
          </section>

          <section v-if="isTenantPortal" class="panel table-panel relation-lower">
            <div class="panel-head padded-head">
              <div>
                <h3>同步任务</h3>
                <p>只展示当前租户授权连接产生的同步任务。</p>
              </div>
            </div>
            <DataTable :columns="syncColumns" :rows="filteredSyncJobs" :pagination="sectionPagination('syncJobs')" min-width="980px" @page="changeSectionPage('syncJobs', $event)">
              <template #cell-job="{ row }"><strong>{{ row.job }}</strong><p class="muted">{{ row.mode }} · {{ row.cron }}</p></template>
              <template #cell-platform="{ row }">{{ row.platformName || getPlatformName(row.platformId) }}</template>
              <template #cell-status="{ row }"><StatusBadge :status="row.status" /></template>
              <template #cell-rate="{ row }"><div class="quota inline"><span :style="{ width: row.successRate + '%' }"></span></div><small>{{ row.successRate }}%</small></template>
              <template #cell-actions="{ row }"><div class="actions wide"><button @click="openSyncDrawer(row)">日志</button></div></template>
            </DataTable>
          </section>
        </div>
      </section>

      <section v-if="page === 'syncMonitor'" class="page-section">
        <div class="toolbar row-between">
          <div class="toolbar-query">
            <div class="search-box"><span>⌕</span><input v-model="syncKeyword" placeholder="搜索租户、平台、能力、任务" /></div>
            <select v-model="syncStatus"><option value="all">全部状态</option><option value="success">成功</option><option value="running">执行中</option><option value="failed">失败</option><option value="paused">暂停</option></select>
          </div>
          <div class="toolbar-actions"></div>
        </div>
        <section class="panel table-panel">
          <DataTable :columns="syncColumns" :rows="filteredSyncJobs" :pagination="sectionPagination('syncJobs')" min-width="1180px" @page="changeSectionPage('syncJobs', $event)">
            <template #cell-job="{ row }"><strong>{{ row.job }}</strong><p class="muted">{{ row.mode }} · {{ row.cron }}</p></template>
            <template #cell-platform="{ row }">{{ row.platformName || getPlatformName(row.platformId) }}</template>
            <template #cell-status="{ row }"><StatusBadge :status="row.status" /></template>
            <template #cell-rate="{ row }"><div class="quota inline"><span :style="{ width: row.successRate + '%' }"></span></div><small>{{ row.successRate }}%</small></template>
            <template #cell-actions="{ row }"><div class="actions wide"><button v-if="canManageConnection" @click="retrySync(row)">重试</button><button v-if="canManageConnection" @click="toggleSync(row)">{{ row.status === 'paused' ? '恢复' : '暂停' }}</button><button @click="openSyncDrawer(row)">日志</button></div></template>
          </DataTable>
        </section>
      </section>

      <section v-if="page === 'quota'" class="page-section quota-layout">
        <PlatformRail
          :platforms="platformsSorted"
          :selected-id="quotaPlatformId"
          :show-clear="false"
          @select="quotaPlatformId = $event"
          @clear="quotaPlatformId = 'all'"
        />
        <div class="quota-main">
          <section class="panel relation-upper">
            <div class="panel-head">
              <div>
                <h3>配额与限流策略</h3>
                <p>支持平台默认、应用默认、租户通用，以及特殊企业专属覆盖。</p>
              </div>
              <div class="panel-control-split"><div class="toolbar-query compact"><button class="ghost-btn" @click="resetQuotaFilters">全部</button></div><div class="toolbar-actions"><button v-if="canManageQuota" class="primary-btn" @click="openPolicyModal()">新增策略</button></div></div>
            </div>
            <div class="policy-grid">
              <EmptyState v-if="!filteredPolicies.length" title="暂无配额策略" />
              <article v-for="policy in filteredPolicies" :key="policy.id" :class="['policy-card selectable', { selected: selectedPolicyId === policy.id }]" @click="selectedPolicyId = policy.id">
                <div class="row-between"><strong>{{ policy.name }}</strong><StatusBadge :status="policy.status" /></div>
                <p>{{ policy.scopeLabel }} · {{ policy.isOverride ? '专属覆盖' : '通用策略' }} · 优先级 {{ policy.priority }}</p>
                <div class="policy-limits">
                  <span>日 {{ formatLimit(policy.dailyLimit) }}</span>
                  <span>月 {{ formatLimit(policy.monthlyLimit) }}</span>
                  <span>QPS {{ policy.qpsLimit || '-' }}</span>
                  <span>并发 {{ policy.concurrentLimit || '-' }}</span>
                </div>
                <div class="card-actions"><button v-if="canManageQuota" @click.stop="openPolicyModal(policy)">编辑</button><button v-if="canManageQuota" @click.stop="copyPolicy(policy)">复制为专属</button><button v-if="canManageQuota" @click.stop="toggleStatus(policy)">{{ policy.status === 'enabled' ? '停用' : '启用' }}</button></div>
              </article>
            </div>
          </section>

          <section class="panel relation-lower">
            <div class="panel-head">
              <div><h3>企业覆盖策略</h3><p>{{ selectedPolicy ? selectedPolicy.name : '请选择上方策略查看绑定对象、覆盖来源和实时用量。' }}</p></div>
              <button v-if="canManageQuota" class="ghost-btn" @click="selectedPolicy && openBindingDrawer(selectedPolicy)">添加租户覆盖策略</button>
            </div>
            <DataTable :columns="quotaUsageColumns" :rows="selectedPolicyUsage" :pagination="sectionPagination('quotaUsages')" min-width="1040px" @page="changeSectionPage('quotaUsages', $event)">
              <template #cell-object="{ row }"><strong>{{ row.name }}</strong><p class="muted">{{ row.objectPath }}</p></template>
              <template #cell-source="{ row }"><span :class="['tag', row.override ? 'danger-soft' : '']">{{ row.override ? '专属覆盖' : '通用策略' }}</span></template>
              <template #cell-used="{ row }"><div>{{ row.used.toLocaleString() }} / {{ row.limit.toLocaleString() }}</div><div class="quota inline"><span :style="{ width: Math.min(row.rate, 100) + '%' }"></span></div></template>
              <template #cell-status="{ row }"><StatusBadge :status="row.rate >= 100 ? 'exceeded' : row.rate >= 80 ? 'warning' : 'normal'" /></template>
              <template #cell-actions="{ row }"><div class="actions"><button v-if="canManageQuota" @click="openOverrideModal(row)">编辑</button><button @click="openUsageDrawer(row)">明细</button></div></template>
            </DataTable>
          </section>
        </div>
      </section>

      <section v-if="page === 'alerts'" class="page-section">
        <div class="toolbar row-between">
          <div class="toolbar-query">
            <div class="search-box"><span>⌕</span><input v-model="alertKeyword" placeholder="搜索异常、租户、平台、能力" /></div>
            <select v-model="alertStatus"><option value="all">全部状态</option><option value="open">待处理</option><option value="processing">处理中</option><option value="resolved">已恢复</option></select>
          </div>
          <div class="toolbar-actions"><button v-if="canManageConnection" class="ghost-btn" @click="exportLogs">导出日志</button></div>
        </div>
        <section class="panel">
          <div class="alert-list">
            <EmptyState v-if="!filteredAlerts.length" title="暂无异常记录" />
            <article v-for="alert in filteredAlerts" :key="alert.id" class="alert-item full">
              <div :class="['alert-dot', alert.level]"></div>
              <div class="alert-body">
                <div class="row-between"><strong>{{ alert.title }}</strong><StatusBadge :status="alert.status" /></div>
                <p>{{ alert.message }}</p>
                <small>{{ alert.path }} · {{ alert.lastAt }} · {{ alert.count }} 次</small>
              </div>
              <div class="actions wide"><button v-if="canManageConnection" @click="processAlert(alert)">处理</button><button v-if="canManageConnection" @click="resolveAlert(alert)">标记恢复</button><button v-if="canManageConnection" @click="ignoreAlert(alert)">忽略</button><button @click="openAlertDrawer(alert)">详情</button></div>
            </article>
          </div>
        </section>
      </section>

      <section v-if="page === 'logs'" class="page-section">
        <div class="toolbar row-between">
          <div class="toolbar-query">
            <div class="search-box"><span>⌕</span><input v-model="logKeyword" placeholder="搜索 request_id、平台、租户、Endpoint、错误码" /></div>
            <select v-model="logType"><option value="all">全部类型</option><option value="api">第三方 API</option><option value="token">Token 刷新</option><option value="callback">回调接收</option><option value="data_write">数据写入</option></select>
            <select v-model="logSuccess"><option value="all">全部结果</option><option value="success">成功</option><option value="failed">失败</option></select>
          </div>
          <div class="toolbar-actions"></div>
        </div>
        <section class="panel table-panel">
          <DataTable :columns="logColumns" :rows="filteredLogs" :pagination="sectionPagination('logs')" min-width="1220px" @page="changeSectionPage('logs', $event)">
            <template #cell-id="{ row }"><strong>{{ row.requestId }}</strong><p class="muted">{{ row.calledAt }}</p></template>
            <template #cell-type="{ row }"><span class="tag">{{ row.typeLabel }}</span></template>
            <template #cell-path="{ row }"><strong>{{ row.method }}</strong> {{ row.path }}<p class="muted">{{ row.endpoint }}</p></template>
            <template #cell-result="{ row }"><StatusBadge :status="row.success ? 'success' : 'failed'" /></template>
            <template #cell-actions="{ row }"><div class="actions"><button @click="openLogDrawer(row)">详情</button><button @click="copyText(row.requestId)">复制 ID</button></div></template>
          </DataTable>
        </section>
      </section>
    </main>

    <Teleport to="body">
      <div v-if="drawer.open" class="integration-portal integration-overlay-mask drawer-mask" @click="closeDrawer"></div>
      <aside v-if="drawer.open" class="integration-portal integration-drawer drawer">
      <div class="drawer-head">
        <div>
          <p class="eyebrow">{{ drawer.subtitle }}</p>
          <h3>{{ drawer.title }}</h3>
          <p>{{ drawer.desc }}</p>
        </div>
        <button class="icon-btn" @click="closeDrawer">×</button>
      </div>

      <div v-if="drawer.type === 'platform'" class="drawer-body">
        <div class="detail-grid three-cols">
          <div><span>平台简称</span><b>{{ drawer.data.shortName || drawer.data.name }}</b></div>
          <div><span>平台编码</span><b>{{ drawer.data.code }}</b></div>
          <div><span>平台类型</span><b>{{ drawer.data.type }}</b></div>
          <div><span>接入方式</span><b>{{ drawer.data.accessType }}</b></div>
          <div><span>平台状态</span><StatusBadge :status="drawer.data.status" /></div>
          <div><span>租户可见</span><b>{{ drawer.data.tenantVisible !== false ? '是' : '否' }}</b></div>
          <div><span>负责人</span><b>{{ drawer.data.owner }}</b></div>
          <div><span>排序权重</span><b>{{ drawer.data.sortWeight ?? '—' }}</b></div>
          <div class="span-2"><span>官方开放平台</span><b v-if="!drawer.data.officialUrl">—</b><a v-else :href="drawer.data.officialUrl" target="_blank" rel="noopener noreferrer" style="font-weight:800;color:var(--primary);word-break:break-all">{{ drawer.data.officialUrl }}</a></div>
          <div><span>服务商应用</span><b>{{ drawer.data.appCount }}</b></div>
          <div><span>平台能力</span><b>{{ drawer.data.capabilityCount }}</b></div>
          <div><span>租户连接</span><b>{{ drawer.data.connectionCount }}</b></div>
        </div>
        <section class="drawer-card"><h4>平台说明</h4><p>{{ drawer.data.description }}</p></section>
        <section v-if="canManageApp || canManagePlatform || canRunHealthCheck" class="drawer-card"><h4>快捷操作</h4><div class="actions wide"><button v-if="canManageApp" @click="jumpWorkspace(drawer.data.id, 'apps')">配置应用</button><button v-if="canManagePlatform" @click="jumpWorkspace(drawer.data.id, 'capabilities')">配置能力</button><button v-if="canRunHealthCheck" @click="runHealthCheck(drawer.data.name)">连通性检测</button></div></section>
        <section v-if="drawer.data.detail" class="drawer-card"><h4>后端详情</h4><pre>{{ pretty(drawer.data.detail) }}</pre></section>
      </div>

      <div v-if="drawer.type === 'app'" class="drawer-body">
        <div class="drawer-tabs"><button v-for="tab in appDrawerTabs" :key="tab.key" :class="{ active: drawerTab === tab.key }" @click="drawerTab = tab.key">{{ tab.label }}</button></div>
        <template v-if="drawerTab === 'basic'">
          <div class="detail-grid">
            <div><span>所属平台</span><b>{{ getPlatformName(drawer.data.platformId) }}</b></div>
            <div><span>应用类型</span><b>{{ drawer.data.type }}</b></div>
            <div><span>环境</span><b>{{ drawer.data.env }}</b></div>
            <div><span>状态</span><StatusBadge :status="drawer.data.status" /></div>
          </div>
        </template>
        <template v-if="drawerTab === 'secret'">
          <div class="detail-list">
            <div v-for="item in appSecretItems(drawer.data)" :key="item.label"><b>{{ item.label }}</b><span>{{ item.value }}</span></div>
          </div>
        </template>
        <template v-if="drawerTab === 'callback'">
          <div class="detail-list">
            <div><b>指令回调 URL</b><span>{{ drawer.data.commandCallback }}</span></div>
            <div><b>数据回调 URL</b><span>{{ drawer.data.dataCallback }}</span></div>
            <div><b>授权完成回调 URL</b><span>{{ drawer.data.authCallback }}</span></div>
            <div><b>suite_ticket 状态</b><span>{{ drawer.data.suiteTicketStatus }} · 最近 {{ drawer.data.suiteTicketAt }}</span></div>
            <div><b>suite_access_token</b><span>{{ drawer.data.suiteTokenStatus }} · {{ drawer.data.suiteTokenExpireAt }} 过期</span></div>
          </div>
        </template>
        <template v-if="drawerTab === 'capabilities'">
          <DataTable :columns="appCapabilityColumns" :rows="appCapabilitiesByApp(drawer.data.id)" min-width="980px">
            <template #cell-name="{ row }"><strong>{{ row.name }}</strong><p class="muted">{{ row.code }}</p></template>
            <template #cell-connected="{ row }"><StatusBadge :status="row.connected ? 'connected' : 'not_connected'" /></template>
            <template #cell-auth="{ row }"><StatusBadge :status="row.authStatus" /></template>
            <template #cell-open="{ row }"><SwitchToggle :model-value="row.openToTenant" :disabled="!canManageApp" @update:model-value="toggleAppCapability(row, 'openToTenant')" /></template>
            <template #cell-default="{ row }"><SwitchToggle :model-value="row.defaultEnabled" :disabled="!canManageApp" @update:model-value="toggleAppCapability(row, 'defaultEnabled')" /></template>
            <template #cell-configurable="{ row }"><SwitchToggle :model-value="row.tenantConfigurable" :disabled="!canManageApp" @update:model-value="toggleAppCapability(row, 'tenantConfigurable')" /></template>
            <template #cell-health="{ row }"><StatusBadge :status="row.health" /></template>
          </DataTable>
        </template>
        <template v-if="drawerTab === 'tenants'">
          <ConnectionTable :rows="connectionsByApp(drawer.data.id)" :can-manage="canManageConnection" @open="openConnectionDrawer" @refresh="refreshConnection" @pause="pauseConnection" @retry="retryConnection" />
        </template>
        <section v-if="drawer.data.detail" class="drawer-card"><h4>后端详情</h4><pre>{{ pretty(drawer.data.detail) }}</pre></section>
      </div>

      <div v-if="drawer.type === 'connection'" class="drawer-body">
        <div class="drawer-tabs"><button v-for="tab in connectionTabs" :key="tab.key" :class="{ active: drawerTab === tab.key }" @click="drawerTab = tab.key">{{ tab.label }}</button></div>
        <template v-if="drawerTab === 'basic'">
          <div class="chain-path">{{ getPlatformName(drawer.data.platformId) }} / {{ getAppName(drawer.data.appId) }} / {{ drawer.data.tenantName }} / {{ drawer.data.authSubject }}</div>
          <div class="detail-grid">
            <div><span>租户</span><b>{{ drawer.data.tenantName }}</b></div>
            <div><span>授权主体</span><b>{{ drawer.data.authSubject }}</b></div>
            <div><span>授权状态</span><StatusBadge :status="drawer.data.authStatus" /></div>
            <div><span>连接状态</span><StatusBadge :status="drawer.data.status" /></div>
            <div><span>最近同步</span><b>{{ drawer.data.lastSync }}</b></div>
            <div><span>今日调用</span><b>{{ formatNumber(drawer.data.callsToday) }}</b></div>
          </div>
        </template>
        <template v-if="drawerTab === 'scope'">
          <div class="detail-list">
            <div><b>授权范围</b><span>{{ drawer.data.authScope.join('、') }}</span></div>
            <div><b>可见范围</b><span>{{ drawer.data.visibleScope }}</span></div>
            <div><b>凭证</b><span>{{ drawer.data.credentialSummary }}</span></div>
          </div>
        </template>
        <template v-if="drawerTab === 'capabilities'">
          <div class="final-cap-list">
            <article v-for="cap in drawer.data.finalCapabilities" :key="cap.code" class="final-cap-card">
              <div class="row-between"><strong>{{ cap.name }}</strong><StatusBadge :status="cap.enabled ? 'enabled' : 'disabled'" /></div>
              <p>{{ cap.reason }}</p>
              <div class="mini-tags"><span>应用能力：{{ cap.appAllowed ? '通过' : '未通过' }}</span><span>授权范围：{{ cap.scopeAllowed ? '通过' : '未通过' }}</span><span>租户选择：{{ cap.tenantSelected ? '启用' : '关闭' }}</span></div>
            </article>
          </div>
        </template>
        <template v-if="drawerTab === 'quota'">
          <div class="quota-mini-list">
            <div v-for="usage in quotaUsages.filter(u => u.connectionId === drawer.data.id)" :key="usage.id" class="quota-mini-row">
              <div><strong>{{ usage.name }}</strong><p>{{ usage.used.toLocaleString() }} / {{ usage.limit.toLocaleString() }}</p><div class="quota"><span :style="{ width: Math.min(usage.rate, 100) + '%' }"></span></div></div>
              <button v-if="canManageQuota" @click="openOverrideModal(usage)">专属覆盖</button>
            </div>
          </div>
        </template>
        <template v-if="drawerTab === 'tasks'">
          <DataTable :columns="syncColumns" :rows="syncJobs.filter(j => j.connectionId === drawer.data.id)" min-width="980px">
            <template #cell-job="{ row }"><strong>{{ row.job }}</strong><p class="muted">{{ row.mode }} · {{ row.cron }}</p></template>
            <template #cell-platform="{ row }">{{ getPlatformName(row.platformId) }}</template>
            <template #cell-status="{ row }"><StatusBadge :status="row.status" /></template>
            <template #cell-rate="{ row }"><div class="quota inline"><span :style="{ width: row.successRate + '%' }"></span></div></template>
          </DataTable>
        </template>
        <section v-if="drawer.data.detail" class="drawer-card"><h4>后端详情</h4><pre>{{ pretty(drawer.data.detail) }}</pre></section>
      </div>

      <div v-if="['capability','policy','usage','alert','log','sync'].includes(drawer.type)" class="drawer-body">
        <section class="drawer-card"><h4>对象详情</h4><pre>{{ pretty(drawer.data) }}</pre></section>
      </div>
      </aside>

      <NeuroAgentDialog
        v-model="modal.open"
        :title="modal.title"
        icon="🔗"
        size="large"
        :width="modal.type === 'platform' ? '1040px' : '820px'"
        :height="modal.type === 'platform' ? 'min(760px, 86vh)' : 'auto'"
        :show-confirm="modal.type !== 'platform'"
        confirm-text="保存"
        @confirm="saveModal"
        @cancel="closeModal"
        @close="closeModal"
      >
        <div class="integration-standard-dialog-body">
          <p v-if="modal.subtitle" class="standard-dialog-subtitle">{{ modal.subtitle }}</p>
          <div v-if="modal.type === 'platform'" class="form-grid">
            <label :class="{ invalid: fieldErrors.name }">平台名称<input v-model="form.name" data-field="name" placeholder="如 企业微信、京东电商" /><small v-if="fieldErrors.name">{{ fieldErrors.name }}</small></label>
            <label>平台简称<input v-model="form.shortName" placeholder="左栏、卡片、表格展示" /></label>
            <label :class="{ invalid: fieldErrors.code }">平台编码<input v-model="form.code" data-field="code" :readonly="!!form.id" placeholder="wecom、jd… 创建后不建议修改" /><small v-if="fieldErrors.code">{{ fieldErrors.code }}</small></label>
            <label>平台类型<select v-model="form.type"><option>电商平台</option><option>协同办公</option><option>ERP</option><option>CRM</option><option>WMS</option></select></label>
            <label>接入方式<select v-model="form.accessType"><option>OAuth2</option><option>第三方服务商</option><option>API Key</option><option>Webhook</option><option>手动密钥</option><option>混合接入</option></select></label>
            <label>平台状态<select v-model="form.status"><option value="draft">草稿</option><option value="enabled">启用</option><option value="disabled">停用</option><option value="maintenance">维护中</option></select></label>
            <label>平台 Logo（卡片）<input v-model="form.icon" placeholder="单字或 emoji，用于头像位" /></label>
            <label>是否对租户可见<select v-model="form.tenantVisible"><option :value="true">是</option><option :value="false">否</option></select></label>
            <label>负责人 / 维护人<input v-model="form.owner" placeholder="异常归属、对接人" /></label>
            <label>排序权重<input type="number" v-model.number="form.sortWeight" placeholder="数字越小越靠前" /></label>
            <label class="span-2">官方开放平台地址<input v-model="form.officialUrl" placeholder="https://…" /></label>
            <label class="span-2">平台说明<textarea v-model="form.description" placeholder="能力边界、接入注意点等（密钥与回调不在此配置）"></textarea></label>
          </div>
          <div v-else-if="modal.type === 'app'" class="form-grid">
            <label :class="{ invalid: fieldErrors.name }">应用名称<input v-model="form.name" data-field="name" /><small v-if="fieldErrors.name">{{ fieldErrors.name }}</small></label>
            <label :class="{ invalid: fieldErrors.platformId }">所属平台<select v-model="form.platformId" data-field="platformId"><option v-for="p in platforms" :key="p.id" :value="p.id">{{ p.name }}</option></select><small v-if="fieldErrors.platformId">{{ fieldErrors.platformId }}</small></label>
            <label>应用类型<select v-model="form.type"><option>第三方服务商应用</option><option>正式应用</option><option>沙箱应用</option><option>专属应用</option></select></label>
            <label>环境<select v-model="form.env"><option>正式</option><option>沙箱</option><option>测试</option></select></label>
            <label class="span-2">Endpoint / 授权域名<input v-model="form.endpoint" /></label>
          </div>
          <div v-else-if="modal.type === 'policy'" class="form-grid">
            <label :class="{ invalid: fieldErrors.name }">策略名称<input v-model="form.name" data-field="name" /><small v-if="fieldErrors.name">{{ fieldErrors.name }}</small></label>
            <label>策略范围<select v-model="form.scope"><option value="platform">平台级</option><option value="provider_app">应用级</option><option value="tenant">租户级</option><option value="connection">连接实例级</option><option value="capability">能力级</option></select></label>
            <label>日配额<input type="number" v-model.number="form.dailyLimit" /></label>
            <label>月配额<input type="number" v-model.number="form.monthlyLimit" /></label>
            <label>QPS<input type="number" v-model.number="form.qpsLimit" /></label>
            <label>并发<input type="number" v-model.number="form.concurrentLimit" /></label>
            <label>超限策略<select v-model="form.exceedStrategy"><option>告警</option><option>排队</option><option>降频</option><option>暂停能力</option><option>暂停连接</option><option>拒绝调用</option></select></label>
            <label>专属覆盖<select v-model="form.isOverride"><option :value="false">否</option><option :value="true">是</option></select></label>
          </div>
          <div v-else-if="modal.type === 'override'" class="form-grid">
            <label class="span-2">覆盖对象<input v-model="form.name" readonly /></label>
            <label>日配额<input type="number" v-model.number="form.limit" /></label>
            <label>QPS<input type="number" v-model.number="form.qps" /></label>
            <label>并发<input type="number" v-model.number="form.concurrent" /></label>
            <label>有效期<select v-model="form.period"><option>长期有效</option><option>本月有效</option><option>7天临时提额</option></select></label>
            <label class="span-2">备注<textarea v-model="form.remark"></textarea></label>
          </div>
          <div v-else class="modal-empty">当前操作需要后端接口支持。</div>
        </div>
        <template v-if="modal.type === 'platform'" #footer-right>
          <div class="standard-dialog-actions">
            <button class="standard-dialog-btn standard-dialog-btn--ghost" @click="closeModal">取消</button>
            <button class="standard-dialog-btn standard-dialog-btn--ghost" @click="savePlatformFromForm(false)">保存</button>
            <button class="standard-dialog-btn standard-dialog-btn--primary" @click="savePlatformFromForm(true)">保存并配置应用</button>
          </div>
        </template>
      </NeuroAgentDialog>

      <NeuroAgentDialog
        v-model="globalSearchOpen"
        title="全局搜索"
        icon="⌕"
        size="large"
        width="1040px"
        height="70vh"
        :show-footer="false"
      >
        <div class="integration-standard-dialog-body">
          <div class="search-box full"><span>⌕</span><input v-model="globalKeyword" autofocus placeholder="搜索平台、应用、租户、连接实例、request_id" /></div>
          <div class="global-results">
            <article v-for="item in globalResults" :key="item.key" @click="openGlobalResult(item)">
              <span>{{ item.type }}</span><strong>{{ item.title }}</strong><p>{{ item.desc }}</p>
            </article>
          </div>
        </div>
      </NeuroAgentDialog>
    </Teleport>
  </div>
</template>

<script setup>
import { ElMessage, ElMessageBox } from 'element-plus'
import { computed, defineComponent, h, nextTick, onMounted, reactive, ref, watch } from 'vue'
import { useRoute } from 'vue-router'

import './integrationCenter.css'
import { usePermissionStore } from '@/stores/permission'
import NeuroAgentDialog from '@/views/components/NeuroAgentDialog.vue'

import {
  checkIntegrationConnectivity,
  createIntegrationProviderApp,
  createIntegrationPlatform,
  createIntegrationQuotaPolicy,
  exportIntegrationLogs,
  fetchIntegrationLogDetail,
  fetchIntegrationAlerts,
  fetchIntegrationAppCapabilities,
  fetchIntegrationLogs,
  fetchIntegrationMyConnectionDetail,
  fetchIntegrationMyConnections,
  fetchIntegrationMySyncJobDetail,
  fetchIntegrationMySyncJobs,
  fetchIntegrationPlatformDetail,
  fetchIntegrationPlatformCapabilities,
  fetchIntegrationPlatforms,
  fetchIntegrationProviderAppDetail,
  fetchIntegrationQuota,
  fetchIntegrationQuotaUsages,
  fetchIntegrationSyncJobDetail,
  fetchIntegrationSyncMonitor,
  fetchIntegrationTenantConnectionDetail,
  fetchIntegrationTenantConnections,
  fetchIntegrationWorkspace,
  ignoreIntegrationAlert,
  pauseIntegrationSyncJob,
  pauseIntegrationTenantConnection,
  processIntegrationAlert,
  refreshIntegrationTenantConnection,
  resolveIntegrationAlert,
  resumeIntegrationSyncJob,
  resumeIntegrationTenantConnection,
  retryIntegrationSyncJob,
  retryIntegrationTenantConnection,
  updateIntegrationAppCapability,
  updateIntegrationProviderApp,
  updateIntegrationPlatform,
  updateIntegrationQuotaPolicy,
  updateIntegrationQuotaPolicyStatus,
} from '../api'

const navItems = [
  { key: 'overview', label: '总览', icon: '⌂', action: '' },
  { key: 'platforms', label: '接入平台', icon: '▣', action: '新增平台' },
  { key: 'workspace', label: '集成工作台', icon: '⌘', action: '新增应用' },
  { key: 'tenantConnections', label: '租户连接', icon: '⇄', action: '刷新列表' },
  { key: 'syncMonitor', label: '同步监控', icon: '↻', action: '创建任务' },
  { key: 'quota', label: '配额与限流', icon: '◴', action: '新增策略' },
  { key: 'alerts', label: '异常监控', icon: '!', action: '批量处理', badge: 5 },
  { key: 'logs', label: '调用日志', icon: '☷', action: '导出日志' },
]

const route = useRoute()
const permissionStore = usePermissionStore()
const routePageMap = {
  '/integration-center': 'overview',
  '/integration-center/platforms': 'platforms',
  '/integration-center/workspace': 'workspace',
  '/integration-center/tenant-connections': 'tenantConnections',
  '/integration-center/my-connections': 'tenantConnections',
  '/integration-center/sync-monitor': 'syncMonitor',
  '/integration-center/quota': 'quota',
  '/integration-center/alerts': 'alerts',
  '/integration-center/logs': 'logs',
}
const page = ref(routePageMap[route.path] || 'overview')
const workspacePlatformId = ref('all')
const workspaceTab = ref('apps')
const workspaceSelectedAppId = ref('all')
const tenantPlatformId = ref('all')
const tenantAppId = ref('all')
const quotaPlatformId = ref('all')
const quotaMode = ref('all')
const selectedPolicyId = ref('')
const drawerTab = ref('basic')
const globalSearchOpen = ref(false)
const globalKeyword = ref('')

const platformKeyword = ref('')
const platformTypeFilter = ref('all')
const platformStatusFilter = ref('all')
const syncKeyword = ref('')
const syncStatus = ref('all')
const alertKeyword = ref('')
const alertStatus = ref('all')
const logKeyword = ref('')
const logType = ref('all')
const logSuccess = ref('all')

const drawer = reactive({ open: false, type: '', title: '', subtitle: '', desc: '', data: null })
const modal = reactive({ open: false, type: '', title: '', subtitle: '' })
const form = reactive({})
const fieldErrors = reactive({})
const backendState = reactive({ loading: false, loaded: false, error: '' })
const backendPaging = reactive({
  platforms: { skip: 0, limit: 24, total: 0 },
  apps: { skip: 0, limit: 24, total: 0 },
  capabilities: { skip: 0, limit: 20, total: 0 },
  appCapabilities: { skip: 0, limit: 20, total: 0 },
  connections: { skip: 0, limit: 20, total: 0 },
  syncJobs: { skip: 0, limit: 20, total: 0 },
  policies: { skip: 0, limit: 20, total: 0 },
  quotaUsages: { skip: 0, limit: 20, total: 0 },
  alerts: { skip: 0, limit: 20, total: 0 },
  logs: { skip: 0, limit: 20, total: 0 },
})

const platforms = reactive([])

const canManagePlatform = computed(() => permissionStore.canUseAction('integration_center:platform_manage'))
const canManageApp = computed(() => permissionStore.canUseAction('integration_center:app_manage'))
const canManageConnection = computed(() => permissionStore.canUseAction('integration_center:connection_manage'))
const canManageQuota = computed(() => permissionStore.canUseAction('integration_center:quota_manage'))
const canRunHealthCheck = computed(() => canManageConnection.value)
const isTenantPortal = computed(() => route.path === '/integration-center/my-connections')

const platformsSorted = computed(() => [...platforms].sort((a, b) => (a.sortWeight ?? 999) - (b.sortWeight ?? 999) || String(a.name).localeCompare(String(b.name), 'zh')))

const apps = reactive([])

const capabilities = reactive([])

const appCapabilities = reactive([])

const connections = reactive([])

const syncJobs = reactive([])

const policies = reactive([])

const quotaUsages = reactive([])

const alerts = reactive([])

const logs = reactive([])

const currentNav = computed(() => navItems.find(n => n.key === page.value) || navItems[0])
const overviewMetrics = computed(() => [
  { label: '接入平台', value: platforms.length, trend: 0, trendText: '后端实时统计' },
  { label: '服务商应用', value: apps.length, trend: 0, trendText: '后端实时统计' },
  { label: '平台能力', value: capabilities.length, trend: 0, trendText: '后端实时统计' },
  { label: '租户连接', value: connections.length, trend: 0, trendText: '后端实时统计' },
  { label: '今日调用', value: formatNumber(platforms.reduce((s, p) => s + p.callsToday, 0)), trend: 0, trendText: '来自调用日志' },
  { label: '待处理异常', value: alerts.filter(a => a.status !== 'resolved').length, trend: 0, trendText: '来自异常记录' },
])

const selectedWorkspacePlatform = computed(() => workspacePlatformId.value === 'all' ? null : platforms.find(p => p.id === workspacePlatformId.value))
const workspaceApps = computed(() => apps.filter(a => workspacePlatformId.value === 'all' || a.platformId === workspacePlatformId.value))
const workspaceCapabilities = computed(() => capabilities.filter(c => workspacePlatformId.value === 'all' || c.platformId === workspacePlatformId.value))
const workspaceConnections = computed(() => connections.filter(c => workspacePlatformId.value === 'all' || c.platformId === workspacePlatformId.value))
const workspaceAlerts = computed(() => alerts.filter(a => workspacePlatformId.value === 'all' || a.path.includes(selectedWorkspacePlatform.value?.name || '')))
const selectedWorkspaceApp = computed(() => apps.find(a => a.id === workspaceSelectedAppId.value))
const workspaceAppCapabilityRows = computed(() => {
  const allowedAppIds = selectedWorkspaceApp.value ? [selectedWorkspaceApp.value.id] : workspaceApps.value.map(a => a.id)
  return appCapabilities.filter(ac => allowedAppIds.includes(ac.appId)).map(enrichAppCapability)
})
const tenantApps = computed(() => apps.filter(a => tenantPlatformId.value === 'all' || a.platformId === tenantPlatformId.value))
const tenantAppsCapabilityTotal = computed(() => tenantApps.value.reduce((sum, a) => sum + (a.capabilityCount || 0), 0))
const tenantConnections = computed(() => connections.filter(c => (tenantPlatformId.value === 'all' || c.platformId === tenantPlatformId.value) && (tenantAppId.value === 'all' || c.appId === tenantAppId.value)))
const tenantConnectionScopeText = computed(() => {
  const platform = tenantPlatformId.value === 'all' ? '全部平台' : getPlatformName(tenantPlatformId.value)
  const app = tenantAppId.value === 'all' ? '全部应用' : getAppName(tenantAppId.value)
  return `${platform} / ${app} / ${tenantConnections.value.length} 个连接实例`
})
const filteredPlatforms = computed(() => [...platforms].filter(p => (platformTypeFilter.value === 'all' || p.type === platformTypeFilter.value) && (platformStatusFilter.value === 'all' || p.status === platformStatusFilter.value) && `${p.name}${p.shortName || ''}${p.code}${p.type}`.toLowerCase().includes(platformKeyword.value.toLowerCase())).sort((a, b) => (a.sortWeight ?? 999) - (b.sortWeight ?? 999) || String(a.name).localeCompare(String(b.name), 'zh')))
const filteredPolicies = computed(() => policies.filter(p => (quotaPlatformId.value === 'all' || p.platformId === quotaPlatformId.value) && (quotaMode.value === 'all' || p.isOverride)))
const selectedPolicy = computed(() => policies.find(p => p.id === selectedPolicyId.value))
const selectedPolicyUsage = computed(() => selectedPolicy.value ? quotaUsages.filter(u => u.policyId === selectedPolicy.value.id || selectedPolicy.value.scope === 'platform') : quotaUsages)
const filteredSyncJobs = computed(() => syncJobs.filter(j => (syncStatus.value === 'all' || j.status === syncStatus.value) && `${j.job}${j.tenant}${j.capability}`.toLowerCase().includes(syncKeyword.value.toLowerCase())))
const filteredAlerts = computed(() => alerts.filter(a => (alertStatus.value === 'all' || a.status === alertStatus.value) && `${a.title}${a.message}${a.path}`.toLowerCase().includes(alertKeyword.value.toLowerCase())))
const filteredLogs = computed(() => logs.filter(l => (logType.value === 'all' || l.type === logType.value) && (logSuccess.value === 'all' || (logSuccess.value === 'success' ? l.success : !l.success)) && `${l.requestId}${l.tenant}${l.path}${getPlatformName(l.platformId)}`.toLowerCase().includes(logKeyword.value.toLowerCase())))
const tenantRanking = computed(() => connections.slice().sort((a,b)=>b.callsToday-a.callsToday).slice(0,5).map(c => ({ name: c.tenantName, desc: `${getPlatformName(c.platformId)} · ${getAppName(c.appId)}`, value: formatNumber(c.callsToday) })))
const globalResults = computed(() => {
  const q = globalKeyword.value.toLowerCase()
  if (!q) return []
  return [
    ...platforms.map(p => ({ key: `p-${p.id}`, type: '平台', title: p.name, desc: `${p.code} · ${p.type}`, data: p, action: () => openPlatformDrawer(p) })),
    ...apps.map(a => ({ key: `a-${a.id}`, type: '应用', title: a.name, desc: `${getPlatformName(a.platformId)} · ${a.type}`, data: a, action: () => openAppDrawer(a) })),
    ...connections.map(c => ({ key: `c-${c.id}`, type: '连接', title: c.tenantName, desc: `${getPlatformName(c.platformId)} / ${c.authSubject}`, data: c, action: () => openConnectionDrawer(c) })),
    ...logs.map(l => ({ key: `l-${l.id}`, type: '日志', title: l.requestId, desc: `${getPlatformName(l.platformId)} · ${l.path}`, data: l, action: () => openLogDrawer(l) })),
  ].filter(i => `${i.title}${i.desc}`.toLowerCase().includes(q)).slice(0, 12)
})

const workspaceTabs = [{ key: 'apps', label: '服务商应用' }, { key: 'capabilities', label: '平台能力' }, { key: 'connections', label: '租户连接实例' }]
const appDrawerTabs = [{ key: 'basic', label: '基础信息' }, { key: 'secret', label: '密钥凭证' }, { key: 'callback', label: '回调与票据' }, { key: 'capabilities', label: '能力连接' }, { key: 'tenants', label: '授权租户' }]
const connectionTabs = [{ key: 'basic', label: '基础信息' }, { key: 'scope', label: '授权范围' }, { key: 'capabilities', label: '最终能力' }, { key: 'tasks', label: '同步任务' }, { key: 'quota', label: '配额用量' }]

const appCapabilityColumns = [
  { key: 'name', label: '能力' }, { key: 'app', label: '所属应用' }, { key: 'connected', label: '接通' }, { key: 'auth', label: '权限状态' }, { key: 'open', label: '开放给租户' }, { key: 'default', label: '默认启用' }, { key: 'configurable', label: '允许关闭' }, { key: 'health', label: '连通性' }, { key: 'actions', label: '操作' }
]
const platformCapabilityColumns = [{ key: 'name', label: '能力' }, { key: 'type', label: '类型' }, { key: 'basic', label: '基础能力' }, { key: 'status', label: '状态' }, { key: 'actions', label: '操作' }]
const syncColumns = [{ key: 'job', label: '任务' }, { key: 'tenant', label: '租户' }, { key: 'platform', label: '平台' }, { key: 'capability', label: '能力' }, { key: 'status', label: '状态' }, { key: 'rate', label: '成功率' }, { key: 'lastRun', label: '最近执行' }, { key: 'nextRun', label: '下次执行' }, { key: 'actions', label: '操作' }]
const quotaUsageColumns = [{ key: 'object', label: '绑定对象' }, { key: 'scope', label: '范围' }, { key: 'source', label: '策略来源' }, { key: 'used', label: '今日用量' }, { key: 'status', label: '状态' }, { key: 'actions', label: '操作' }]
const logColumns = [{ key: 'id', label: 'Request ID' }, { key: 'type', label: '类型' }, { key: 'tenant', label: '租户' }, { key: 'path', label: '调用路径' }, { key: 'status', label: 'HTTP' }, { key: 'cost', label: '耗时(ms)' }, { key: 'result', label: '结果' }, { key: 'actions', label: '操作' }]

watch(
  () => route.path,
  (path) => {
    page.value = routePageMap[path] || 'overview'
    if (backendState.loaded) loadBackendSnapshots()
  },
)

watch([platformKeyword, platformStatusFilter], () => resetSectionPage('platforms', true))
watch([syncKeyword, syncStatus], () => resetSectionPage('syncJobs', true))
watch([alertKeyword, alertStatus], () => resetSectionPage('alerts', true))
watch([logKeyword, logSuccess], () => resetSectionPage('logs', true))
watch([tenantPlatformId, tenantAppId], () => resetSectionPage('connections', true))
watch([quotaPlatformId, selectedPolicyId], () => {
  resetSectionPage('policies', false)
  resetSectionPage('quotaUsages', true)
})

onMounted(() => {
  loadBackendSnapshots()
})

function switchPage(key) { page.value = key }
function primaryAction() {
  if (page.value === 'platforms') return openPlatformModal()
  if (page.value === 'workspace') return openAppModal()
  if (page.value === 'tenantConnections') return loadBackendSnapshots()
  if (page.value === 'quota') return openPolicyModal()
  if (page.value === 'logs') return exportLogs()
  if (page.value === 'alerts') return notify('批量处理需要选择具体异常', 'error')
  loadBackendSnapshots()
}
function selectWorkspacePlatform(id) { workspacePlatformId.value = id; workspaceSelectedAppId.value = workspaceApps.value[0]?.id || 'all' }
function jumpWorkspace(platformId, tab) { workspacePlatformId.value = platformId; workspaceTab.value = tab; page.value = 'workspace'; workspaceSelectedAppId.value = apps.find(a => a.platformId === platformId)?.id || 'all' }
function selectWorkspaceApp(id) { workspaceSelectedAppId.value = id }
function clearSelectedApp() { workspaceSelectedAppId.value = 'all' }
function selectTenantPlatform(id) { tenantPlatformId.value = id; tenantAppId.value = 'all' }
function clearTenantPlatform() { tenantPlatformId.value = 'all'; tenantAppId.value = 'all' }
function selectTenantApp(id) { tenantAppId.value = id; const app = apps.find(a => a.id === id); if (app) tenantPlatformId.value = app.platformId }
function clearTenantFilters() { tenantPlatformId.value = 'all'; tenantAppId.value = 'all' }
function resetQuotaFilters() { quotaPlatformId.value = 'all'; quotaMode.value = 'all' }
function openGlobalResult(item) { globalSearchOpen.value = false; item.action() }
function getPlatformName(id) { return platforms.find(p => p.id === id)?.name || '-' }
function getAppName(id) { return apps.find(a => a.id === id)?.name || '-' }
function enrichAppCapability(ac) { const cap = capabilities.find(c => c.id === ac.capabilityId) || {}; const app = apps.find(a => a.id === ac.appId) || {}; return { ...ac, name: cap.name, code: cap.code, platformId: cap.platformId, appName: app.name } }
function appCapabilitiesByApp(appId) { return appCapabilities.filter(ac => ac.appId === appId).map(enrichAppCapability) }
function connectionsByApp(appId) { return connections.filter(c => c.appId === appId) }
function formatNumber(n) { return Number(n || 0).toLocaleString('zh-CN') }
function formatLimit(n) { return n ? formatNumber(n) : '-' }
function pretty(data) { return JSON.stringify(data, null, 2) }
function notify(message, type = 'success') {
  const options = {
    message,
    duration: 2200,
    showClose: false,
    grouping: true,
    customClass: 'integration-platform-message',
  }
  if (type === 'error') {
    ElMessage.error(options)
    return
  }
  ElMessage.success(options)
}
function clearFieldErrors() {
  Object.keys(fieldErrors).forEach(key => delete fieldErrors[key])
}
async function focusFirstFieldError() {
  await nextTick()
  const field = Object.keys(fieldErrors)[0]
  if (!field) return
  document.querySelector(`[data-field="${field}"]`)?.focus?.()
}
function setFieldErrors(errors) {
  clearFieldErrors()
  Object.assign(fieldErrors, errors)
  focusFirstFieldError()
}
function readableError(error, fallback = '操作失败') {
  const raw = error instanceof Error ? error.message : String(error || '')
  const message = raw.replace(/（HTTP \d+）$/, '').trim()
  if (!message) return fallback
  if (/Network Error|无法连接后端|timeout|ECONN/i.test(message)) return '后端服务不可用，请稍后重试或联系管理员'
  if (/请求参数错误|invalid json|invalid request/i.test(message)) return '提交内容格式不正确，请检查必填项和字段格式'
  if (/forbidden|permission|无权限|unauthorized/i.test(message)) return '当前账号无权执行该操作'
  if (/not found|不存在/i.test(message)) return '目标数据不存在或已被更新，请刷新后重试'
  if (/配额|套餐|quota/i.test(message)) return message
  return message || fallback
}
function ensureAction(allowed, message = '当前账号无权执行该操作') {
  if (allowed) return true
  notify(message, 'error')
  return false
}
function closeDrawer() { drawer.open = false; drawer.type = ''; drawer.data = null }
function openDrawer(type, title, subtitle, desc, data, tab = 'basic') { drawer.type = type; drawer.title = title; drawer.subtitle = subtitle; drawer.desc = desc; drawer.data = data; drawer.open = true; drawerTab.value = tab }
function openPlatformDrawer(platform) {
  openDrawer('platform', platform.name, '接入平台详情', '平台是聚合根，应用、能力、连接实例都围绕平台展开。', platform)
  loadDrawerDetail(platform, () => fetchIntegrationPlatformDetail(platform.id))
}
function openAppDrawer(app, tab = 'basic') {
  openDrawer('app', app.name, '服务商应用详情', '应用承载密钥、回调、应用能力连接与授权租户。', app, tab)
  loadDrawerDetail(app, () => fetchIntegrationProviderAppDetail(app.id))
}
function openConnectionDrawer(conn) {
  openDrawer('connection', `${conn.tenantName} · ${conn.authSubject}`, '租户连接实例详情', '租户通过应用授权后生成的运行实例。', conn)
  loadDrawerDetail(conn, () => (isTenantPortal.value ? fetchIntegrationMyConnectionDetail(conn.id) : fetchIntegrationTenantConnectionDetail(conn.id)))
}
function openCapabilityDrawer(row) { openDrawer('capability', row.name, '能力详情', '能力定义、应用接通情况与租户开放策略。', row) }
function openPolicyDrawer(policy) { openDrawer('policy', policy.name, '配额策略详情', '查看策略范围、优先级、超限动作与绑定情况。', policy) }
function openBindingDrawer(policy) { openDrawer('policy', policy.name, '策略绑定对象', '策略可绑定平台、应用、租户、连接实例或能力。', { policy, bindings: selectedPolicyUsage.value }) }
function openUsageDrawer(row) { openDrawer('usage', row.name, '配额用量明细', '展示用量、来源、覆盖关系与超限记录。', row) }
function openAlertDrawer(row) { openDrawer('alert', row.title, '异常详情', '异常聚合、影响对象、处理动作与恢复状态。', row) }
function openLogDrawer(row) {
  openDrawer('log', row.requestId, '调用日志详情', '请求、响应、耗时、错误码与调用链路。', row)
  loadDrawerDetail(row, () => fetchIntegrationLogDetail(row.id))
}
function openSyncDrawer(row) {
  openDrawer('sync', row.job, '同步任务日志', '查看任务执行记录、失败原因与重试链路。', row)
  loadDrawerDetail(row, () => (isTenantPortal.value ? fetchIntegrationMySyncJobDetail(row.id) : fetchIntegrationSyncJobDetail(row.id)))
}
async function loadDrawerDetail(target, loader) {
  if (!target || !isNumericId(target.id) && !target.id) return
  try {
    const detail = await loader()
    Object.assign(target, { detail })
  } catch (error) {
    notify(readableError(error, '详情加载失败'), 'error')
  }
}
async function loadBackendSnapshots() {
  if (isTenantPortal.value) {
    return loadBackendSections(['connections', 'syncJobs'])
  }
  return loadBackendSections()
}

async function loadBackendSections(keys) {
  backendState.loading = true
  backendState.error = ''
  try {
    const wanted = Array.isArray(keys) && keys.length ? new Set(keys) : null
    const sections = buildBackendSections().filter(section => !wanted || wanted.has(section.key))
    const results = await Promise.allSettled(sections.map(section => loadPagedSection(section)))
    const failures = []
    results.forEach((result, index) => {
      const section = sections[index]
      try {
        if (result.status === 'fulfilled') {
          applyBackendSection(result.value, section)
          return
        }
      } catch {
        // Treat malformed backend payloads as load failures so the page never shows stale local state.
      }
      replaceRows(section.target, [])
      failures.push(section.label)
    })
    if (failures.length) {
      backendState.error = `${failures.join('、')}加载失败，请检查后端接口与初始化数据`
      notify(backendState.error, 'error')
    }
    reconcileSelections()
  } catch (error) {
    backendState.error = readableError(error, '第三方集成中心数据加载失败')
    notify(backendState.error, 'error')
  } finally {
    backendState.loaded = true
    backendState.loading = false
  }
}

function buildBackendSections() {
  const connectionLoader = isTenantPortal.value ? fetchIntegrationMyConnections : fetchIntegrationTenantConnections
  const syncLoader = isTenantPortal.value ? fetchIntegrationMySyncJobs : fetchIntegrationSyncMonitor
  return [
    { key: 'platforms', label: '接入平台', loader: fetchIntegrationPlatforms, target: platforms, map: mapBackendPlatform, params: platformListParams },
    { key: 'apps', label: '服务商应用', loader: fetchIntegrationWorkspace, target: apps, map: mapBackendApp, params: appListParams },
    { key: 'capabilities', label: '平台能力', loader: fetchIntegrationPlatformCapabilities, target: capabilities, map: mapBackendCapability, params: platformScopedParams },
    { key: 'appCapabilities', label: '应用能力', loader: fetchIntegrationAppCapabilities, target: appCapabilities, map: mapBackendAppCapability, params: appCapabilityParams },
    { key: 'connections', label: '租户连接', loader: connectionLoader, target: connections, map: mapBackendConnection, params: connectionListParams },
    { key: 'syncJobs', label: '同步任务', loader: syncLoader, target: syncJobs, map: mapBackendSyncJob, params: syncListParams },
    { key: 'policies', label: '配额策略', loader: fetchIntegrationQuota, target: policies, map: mapBackendPolicy, params: platformScopedParams },
    { key: 'quotaUsages', label: '配额用量', loader: fetchIntegrationQuotaUsages, target: quotaUsages, map: mapBackendQuotaUsage, params: platformScopedParams },
    { key: 'alerts', label: '异常监控', loader: fetchIntegrationAlerts, target: alerts, map: mapBackendAlert, params: alertListParams },
    { key: 'logs', label: '调用日志', loader: fetchIntegrationLogs, target: logs, map: mapBackendLog, params: logListParams },
  ]
}

async function loadPagedSection(section) {
  const paging = backendPaging[section.key]
  const params = { skip: paging.skip, limit: paging.limit, ...(section.params?.() || {}) }
  let payload = await section.loader(params)
  let page = normalizePagePayload(payload, paging)
  if (!page.items.length && page.total > 0 && page.skip > 0) {
    paging.skip = Math.max(0, Math.floor((page.total - 1) / page.limit) * page.limit)
    payload = await section.loader({ ...params, skip: paging.skip, limit: paging.limit })
  }
  return payload
}

function applyBackendSection(payload, section) {
  const { target, map, key } = section
  const page = normalizePagePayload(payload, backendPaging[key])
  Object.assign(backendPaging[key], { skip: page.skip, limit: page.limit, total: page.total })
  const items = Array.isArray(payload?.items) ? payload.items : []
  replaceRows(target, items.map(map))
}

function normalizePagePayload(payload, fallback) {
  const items = Array.isArray(payload?.items) ? payload.items : []
  const limit = positiveNumber(payload?.limit, fallback.limit)
  const skip = positiveNumber(payload?.skip, fallback.skip)
  const total = positiveNumber(payload?.total, items.length)
  return { items, skip, limit, total }
}

function positiveNumber(value, fallback) {
  const n = Number(value)
  return Number.isFinite(n) && n >= 0 ? n : fallback
}

function sectionPagination(key) {
  return backendPaging[key]
}

function changeSectionPage(key, skip) {
  if (!backendPaging[key]) return
  backendPaging[key].skip = Math.max(0, Number(skip) || 0)
  loadBackendSections([key])
}

function resetSectionPage(key, reload = false) {
  if (!backendPaging[key]) return
  backendPaging[key].skip = 0
  if (reload && backendState.loaded) loadBackendSections([key])
}

function platformListParams() {
  return compactParams({
    keyword: platformKeyword.value,
    status: platformStatusFilter.value === 'all' ? '' : platformStatusFilter.value,
  })
}

function platformScopedParams() {
  return compactParams({
    platform_code: quotaPlatformId.value === 'all' ? '' : quotaPlatformId.value,
  })
}

function appListParams() {
  return compactParams({
    platform_code: workspacePlatformId.value === 'all' ? '' : workspacePlatformId.value,
  })
}

function appCapabilityParams() {
  return compactParams({
    platform_code: workspacePlatformId.value === 'all' ? '' : workspacePlatformId.value,
    provider_app_code: workspaceSelectedAppId.value === 'all' ? '' : workspaceSelectedAppId.value,
  })
}

function connectionListParams() {
  return compactParams({
    platform_code: tenantPlatformId.value === 'all' ? '' : tenantPlatformId.value,
    provider_app_code: tenantAppId.value === 'all' ? '' : tenantAppId.value,
  })
}

function syncListParams() {
  return compactParams({
    keyword: syncKeyword.value,
    status: syncStatus.value === 'all' ? '' : syncStatus.value,
  })
}

function alertListParams() {
  return compactParams({
    keyword: alertKeyword.value,
    status: alertStatus.value === 'all' ? '' : alertStatus.value,
  })
}

function logListParams() {
  return compactParams({
    keyword: logKeyword.value,
    status: logSuccess.value === 'all' ? '' : logSuccess.value,
  })
}

function compactParams(params) {
  return Object.fromEntries(Object.entries(params).filter(([, value]) => value !== '' && value !== undefined && value !== null))
}

function replaceRows(target, rows) {
  target.splice(0, target.length, ...rows)
}

function reconcileSelections() {
  if (workspacePlatformId.value !== 'all' && !platforms.some(p => p.id === workspacePlatformId.value)) {
    workspacePlatformId.value = 'all'
  }
  if (workspaceSelectedAppId.value !== 'all' && !apps.some(a => a.id === workspaceSelectedAppId.value)) {
    workspaceSelectedAppId.value = 'all'
  }
  if (tenantPlatformId.value !== 'all' && !platforms.some(p => p.id === tenantPlatformId.value)) {
    tenantPlatformId.value = 'all'
  }
  if (tenantAppId.value !== 'all' && !apps.some(a => a.id === tenantAppId.value)) {
    tenantAppId.value = 'all'
  }
  if (quotaPlatformId.value !== 'all' && !platforms.some(p => p.id === quotaPlatformId.value)) {
    quotaPlatformId.value = 'all'
  }
  if (!selectedPolicyId.value || !policies.some(p => p.id === selectedPolicyId.value)) {
    selectedPolicyId.value = policies[0]?.id || ''
  }
}

function mapBackendPlatform(row) {
  const code = row.code || row.platform_code || row.PlatformCode || String(row.ID || row.id || '')
  return {
    id: code,
    rawId: row.id || row.ID || null,
    icon: String(row.name || row.platform_name || row.PlatformName || code).slice(0, 1),
    shortName: row.short_name || row.platform_short_name || row.PlatformShortName || row.name || row.platform_name || row.PlatformName,
    name: row.name || row.platform_name || row.PlatformName || code,
    code,
    type: row.platform_type || row.PlatformType || '-',
    accessType: row.access_mode || row.AccessMode || '-',
    owner: row.owner_name || row.OwnerName || '-',
    status: mapStatus(row.status || row.Status),
    tenantVisible: row.tenant_visible ?? row.TenantVisible ?? false,
    officialUrl: row.official_url || row.OfficialURL || '',
    sortWeight: Number(row.sort_order || row.SortOrder || row.id || row.ID || 999),
    appCount: Number(row.app_count || row.AppCount || 0),
    capabilityCount: Number(row.capability_count || row.CapabilityCount || 0),
    connectionCount: Number(row.connection_count || row.ConnectionCount || 0),
    alertCount: Number(row.open_alert_count || row.OpenAlertCount || 0),
    callsToday: Number(row.calls_today || row.CallsToday || 0),
    successRate: Number(row.success_rate || row.SuccessRate || 0),
    description: row.description || row.Description || '',
  }
}

function mapBackendApp(row) {
  const id = row.app_code || row.AppCode || String(row.id || row.ID || '')
  const platformId = findPlatformId(row.platform_id || row.PlatformID, row.platform_name || row.PlatformName)
  return {
    id,
    rawId: row.id || row.ID || null,
    platformId,
    name: row.app_name || row.AppName || id,
    type: row.app_type || row.AppType || '服务商应用',
    env: row.environment || row.Environment || 'prod',
    status: mapStatus(row.status || row.Status),
    capabilityCount: Number(row.capability_count || row.CapabilityCount || 0),
    connectionCount: Number(row.connection_count || row.ConnectionCount || 0),
    alertCount: Number(row.alert_count || row.AlertCount || 0),
    callsToday: Number(row.calls_today || row.CallsToday || 0),
    endpoint: row.endpoint || row.Endpoint || row.callback_url || row.CallbackURL || '',
    authMode: row.auth_mode || row.AuthMode || '-',
    credentialRef: row.credential_ref || row.CredentialRef || '',
    callbackUrl: row.callback_url || row.CallbackURL || '',
    webhookUrl: row.webhook_url || row.WebhookURL || '',
    owner: row.owner_name || row.OwnerName || '',
    description: row.description || row.Description || '',
  }
}

function mapBackendCapability(row) {
  const platformId = findPlatformId(row.platform_id || row.PlatformID, row.platform_name || row.PlatformName)
  const code = row.capability_code || row.CapabilityCode || String(row.id || row.ID || '')
  const id = `${platformId}:${code}`
  const capabilityType = row.capability_type || row.CapabilityType || '-'
  return {
    id,
    rawId: row.id || row.ID || null,
    platformId,
    name: row.capability_name || row.CapabilityName || code,
    code,
    type: capabilityType,
    isBasic: ['authorization', '基础能力', 'api'].includes(capabilityType),
    status: mapStatus(row.status || row.Status),
    description: row.description || row.Description || '',
  }
}

function mapBackendAppCapability(row) {
  const platformId = findPlatformId(row.platform_id || row.PlatformID, row.platform_name || row.PlatformName)
  const appId = findAppId(row.provider_app_id || row.ProviderAppID, row.provider_app_name || row.ProviderAppName, platformId)
  const capabilityCode = row.capability_code || row.CapabilityCode || String(row.platform_capability_id || row.PlatformCapabilityID || '')
  return {
    id: String(row.id || row.ID || ''),
    appId,
    capabilityId: `${platformId}:${capabilityCode}`,
    connected: (row.connection_status || row.ConnectionStatus) === 'connected',
    authStatus: row.review_status || row.ReviewStatus || 'pending',
    openToTenant: row.open_to_tenant ?? row.OpenToTenant ?? false,
    defaultEnabled: row.default_enabled ?? row.DefaultEnabled ?? false,
    tenantConfigurable: row.tenant_configurable ?? row.TenantConfigurable ?? true,
    health: mapConnectionStatus(row.connection_status || row.ConnectionStatus),
  }
}

function mapBackendConnection(row) {
  const platformId = findPlatformId(row.platform_id || row.PlatformID, row.platform_name || row.PlatformName)
  const appId = findAppId(row.provider_app_id || row.ProviderAppID, row.provider_app_name || row.ProviderAppName, platformId)
  return {
    id: String(row.id || row.ID || row.auth_subject_id || row.AuthSubjectID),
    tenantName: row.tenant_name || row.TenantName || `租户 ${row.tenant_id || row.TenantID || '-'}`,
    platformId,
    platformName: row.platform_name || row.PlatformName || '',
    appId,
    appName: row.provider_app_name || row.ProviderAppName || '',
    authSubject: row.auth_subject_name || row.AuthSubjectName || '-',
    authStatus: mapAuthStatus(row.auth_status || row.AuthStatus),
    status: mapConnectionStatus(row.connection_status || row.ConnectionStatus),
    finalCapabilityCount: Number(row.final_capability_count || row.FinalCapabilityCount || 0),
    callsToday: Number(row.calls_today || row.CallsToday || 0),
    lastSync: formatBackendTime(row.last_sync_at || row.LastSyncAt),
    authScope: parseBackendList(row.auth_scope || row.AuthScope),
    visibleScope: row.auth_subject_type || row.AuthSubjectType || '-',
    credentialSummary: row.credential_summary || row.CredentialSummary || '凭证由后端密钥引用托管',
    finalCapabilities: parseBackendList(row.final_capabilities || row.FinalCapabilities),
  }
}

function mapBackendSyncJob(row) {
  const total = Number(row.total_count || row.TotalCount || 0)
  const success = Number(row.success_count || row.SuccessCount || 0)
  const platformId = findPlatformId(row.platform_id || row.PlatformID, row.platform_name || row.PlatformName)
  return {
    id: String(row.id || row.ID),
    connectionId: String(row.tenant_connection_id || row.TenantConnectionID || ''),
    platformId,
    platformName: row.platform_name || row.PlatformName || '',
    job: row.job_type || row.JobType || '同步任务',
    tenant: row.tenant_name || row.TenantName || `租户 ${row.tenant_id || row.TenantID || '-'}`,
    capability: row.capability_name || row.CapabilityName || row.capability_code || row.CapabilityCode || '-',
    mode: row.trigger_mode || row.TriggerMode || '-',
    cron: '-',
    status: mapStatus(row.status || row.Status),
    successRate: total ? Math.round((success / total) * 100) : 0,
    lastRun: formatBackendTime(row.finished_at || row.FinishedAt || row.started_at || row.StartedAt),
    nextRun: '-',
  }
}

function mapBackendPolicy(row) {
  return {
    id: row.policy_code || row.PolicyCode || String(row.id || row.ID),
    platformId: findPlatformId(row.platform_id || row.PlatformID, row.platform_name || row.PlatformName),
    name: row.policy_name || row.PolicyName || '-',
    quotaCode: row.quota_code || row.QuotaCode || '',
    scope: 'tenant',
    scopeLabel: '租户级',
    dailyLimit: Number(row.default_limit || row.DefaultLimit || 0),
    monthlyLimit: 0,
    qpsLimit: 0,
    concurrentLimit: 0,
    exceedStrategy: row.over_limit_action || row.OverLimitAction || '-',
    status: mapStatus(row.status || row.Status),
    priority: 300,
    isOverride: false,
  }
}

function mapBackendQuotaUsage(row) {
  const used = Number(row.used_amount || row.UsedAmount || 0)
  const limit = Number(row.limit_amount || row.LimitAmount || 0)
  const rate = limit > 0 ? Math.min(100, Math.round((used / limit) * 100)) : 0
  const policyId = row.policy_code || row.PolicyCode || policies.find(policy => policy.quotaCode === (row.quota_code || row.QuotaCode))?.id || ''
  const path = [row.platform_name || row.PlatformName, row.provider_app_name || row.ProviderAppName, row.connection_name || row.ConnectionName].filter(Boolean).join(' / ')
  return {
    id: String(row.id || row.ID || ''),
    policyId,
    connectionId: row.tenant_connection_id || row.TenantConnectionID || null,
    name: row.tenant_name || row.TenantName || '-',
    scope: row.tenant_connection_id || row.TenantConnectionID ? '连接实例级' : '租户级',
    objectPath: path || row.quota_code || row.QuotaCode || '-',
    used,
    limit,
    rate,
    override: false,
    limitedCount: Number(row.limited_count || row.LimitedCount || 0),
    lastUsedAt: formatBackendTime(row.last_used_at || row.LastUsedAt),
  }
}

function mapBackendAlert(row) {
  return {
    id: String(row.id || row.ID),
    level: row.severity || row.Severity || 'warning',
    title: row.title || row.Title || '-',
    message: row.message || row.Message || '',
    path: row.object_path || row.ObjectPath || row.alert_type || row.AlertType || '-',
    status: mapAlertStatus(row.status || row.Status),
    count: Number(row.alert_count || row.AlertCount || 1),
    lastAt: formatBackendTime(row.last_seen_at || row.LastSeenAt),
  }
}

function mapBackendLog(row) {
  const status = Number(row.http_status || row.HTTPStatus || 0)
  return {
    id: String(row.id || row.ID),
    requestId: row.request_id || row.RequestID || '-',
    type: normalizeLogType(row.call_type || row.CallType),
    typeLabel: row.call_type || row.CallType || '第三方 API',
    platformId: findPlatformId(row.platform_id || row.PlatformID, row.platform_name || row.PlatformName),
    tenant: row.tenant_name || row.TenantName || (row.tenant_id || row.TenantID ? `租户 ${row.tenant_id || row.TenantID}` : '-'),
    method: row.method || row.Method || '-',
    path: row.endpoint || row.Endpoint || '-',
    endpoint: row.endpoint || row.Endpoint || '-',
    success: (row.status || row.Status) === 'success',
    status,
    cost: Number(row.duration_ms || row.DurationMS || 0),
    calledAt: formatBackendTime(row.called_at || row.CalledAt),
  }
}

function findPlatformId(platformID, platformName) {
  const byName = platforms.find(p => p.name === platformName || p.shortName === platformName)
  if (byName) return byName.id
  const byIndex = platforms.find(p => String(p.rawId) === String(platformID) || String(p.id) === String(platformID))
  return byIndex?.id || 'all'
}

function findAppId(appID, appName, platformId) {
  const app = apps.find(a => (String(a.rawId) === String(appID) || a.name === appName) && (platformId === 'all' || a.platformId === platformId))
  return app?.id || 'all'
}

function mapStatus(status) {
  if (status === 'online') return 'enabled'
  if (status === 'connected') return 'enabled'
  if (status === 'beta') return 'testing'
  return status || 'draft'
}

function mapAuthStatus(status) {
  if (status === 'authorized') return 'valid'
  return status || 'unknown'
}

function mapConnectionStatus(status) {
  if (status === 'connected') return 'connected'
  if (status === 'inactive') return 'disabled'
  return status || 'warning'
}

function mapAlertStatus(status) {
  if (status === 'open') return 'pending'
  return status || 'pending'
}

function normalizeLogType(type) {
  if (type === 'third_party_api') return 'api'
  if (type === 'token_refresh') return 'token'
  if (type === 'webhook') return 'callback'
  return type || 'api'
}

function formatBackendTime(value) {
  if (!value) return '-'
  return String(value).replace('T', ' ').slice(0, 16)
}
function parseBackendList(value) {
  if (Array.isArray(value)) return value
  if (!value) return []
  if (typeof value === 'string') {
    try {
      const parsed = JSON.parse(value)
      if (Array.isArray(parsed)) return parsed
    } catch (_) {
      return value.split(/[,，]/).map(item => item.trim()).filter(Boolean)
    }
  }
  return []
}
function openPlatformModal(platform) {
  if (!ensureAction(canManagePlatform.value, '缺少接入平台管理权限')) return
  const subtitle = '轻量建档 · 密钥、回调、suite_id、能力、租户授权请在平台详情与工作台配置'
  const template = () => ({
    name: '',
    shortName: '',
    code: '',
    type: '电商平台',
    accessType: 'OAuth2',
    icon: '',
    status: 'draft',
    tenantVisible: true,
    owner: '',
    officialUrl: '',
    sortWeight: 100,
    description: '',
  })
  const data = platform ? JSON.parse(JSON.stringify(platform)) : template()
  if (data.tenantVisible === undefined) data.tenantVisible = true
  if (data.sortWeight === undefined) data.sortWeight = 100
  if (data.shortName === undefined) data.shortName = data.name
  openModal('platform', platform ? '编辑接入平台' : '新增接入平台', subtitle, data)
}
async function savePlatformFromForm(goWorkspace) {
  if (!ensureAction(canManagePlatform.value, '缺少接入平台管理权限')) return
  const name = String(form.name || '').trim()
  const code = String(form.code || '').trim()
  if (!name || !code) {
    setFieldErrors({
      ...(name ? {} : { name: '请填写平台名称' }),
      ...(code ? {} : { code: '请填写平台编码' }),
    })
    notify('请先补全平台必填字段', 'error')
    return
  }
  const existing = form.id && platforms.find(p => p.id === form.id)
  const shortName = String(form.shortName || '').trim() || name
  const slug = code.replace(/[^a-z0-9_-]/gi, '_').toLowerCase()
  let savedId = form.id
  const payload = {
    name,
    short_name: shortName,
    code: slug,
    platform_type: form.type,
    access_mode: form.accessType,
    status: form.status,
    tenant_visible: form.tenantVisible !== false,
    owner_name: String(form.owner || '').trim(),
    official_url: String(form.officialUrl || '').trim(),
    sort_order: Number(form.sortWeight) || 100,
    description: String(form.description || ''),
  }
  try {
    const saved = existing
      ? await updateIntegrationPlatform(existing.code || existing.id, payload)
      : await createIntegrationPlatform(payload)
    const mapped = mapBackendPlatform(saved)
    if (existing) {
      Object.assign(existing, mapped)
      savedId = existing.id
    } else {
      savedId = mapped.id
      platforms.push(mapped)
    }
  } catch (error) {
    backendState.error = readableError(error, '接入平台保存失败')
    notify(backendState.error, 'error')
    return
  }
  closeModal()
  notify(existing ? '平台档案已更新' : '平台档案已创建')
  if (goWorkspace) jumpWorkspace(savedId, 'apps')
}
function openAppModal(app) {
  if (!ensureAction(canManageApp.value, '缺少服务商应用管理权限')) return
  openModal('app', app ? '编辑服务商应用' : '新增服务商应用', '服务商应用', app || { name: '', platformId: workspacePlatformId.value === 'all' ? (platforms[0]?.id || 'all') : workspacePlatformId.value, type: '正式应用', env: '正式', endpoint: '' })
}
function openCapabilityModal(cap) {
  if (!ensureAction(canManagePlatform.value, '缺少平台能力管理权限')) return
  openModal('capability', cap ? '编辑平台能力' : '新增平台能力', '平台能力', cap || {})
}
function openPolicyModal(policy) {
  if (!ensureAction(canManageQuota.value, '缺少配额策略管理权限')) return
  openModal('policy', policy ? '编辑配额策略' : '新增配额策略', '配额与限流', policy || { name: '', scope: 'tenant', dailyLimit: 100000, monthlyLimit: 3000000, qpsLimit: 20, concurrentLimit: 5, exceedStrategy: '告警', isOverride: false })
}
function openOverrideModal(row) {
  if (!ensureAction(canManageQuota.value, '缺少配额覆盖管理权限')) return
  openModal('override', '配置专属覆盖', '特殊企业配额覆盖', { name: row.name, limit: row.limit || 300000, qps: 50, concurrent: 10, period: '长期有效', remark: '大客户专属提额' })
}
function openModal(type, title, subtitle, data) { clearFieldErrors(); Object.keys(form).forEach(k => delete form[k]); Object.assign(form, JSON.parse(JSON.stringify(data || {}))); modal.type = type; modal.title = title; modal.subtitle = subtitle; modal.open = true }
function closeModal() { clearFieldErrors(); modal.open = false }
async function saveModal() {
  if (modal.type === 'app') return saveAppFromForm()
  if (modal.type === 'policy') return savePolicyFromForm()
  notify(`${modal.title}已保存`)
  closeModal()
}
async function saveAppFromForm() {
  if (!ensureAction(canManageApp.value, '缺少服务商应用管理权限')) return
  const name = String(form.name || '').trim()
  const platformId = String(form.platformId || workspacePlatformId.value || '').trim()
  if (!name || !platformId || platformId === 'all') {
    setFieldErrors({
      ...(name ? {} : { name: '请填写应用名称' }),
      ...(!platformId || platformId === 'all' ? { platformId: '请选择所属平台' } : {}),
    })
    notify('请先补全应用必填字段', 'error')
    return
  }
  const existing = form.id && apps.find(a => a.id === form.id)
  const code = existing?.id || slugify(form.code || name)
  const payload = {
    platform_code: platformId,
    code,
    name,
    app_type: form.type || 'provider_app',
    auth_mode: form.authMode || form.accessType || 'OAuth2',
    environment: form.env || 'prod',
    status: form.status || 'enabled',
    tenant_visible: form.tenantVisible !== false,
    callback_url: form.callbackUrl || form.endpoint || '',
    webhook_url: form.webhookUrl || '',
    credential_ref: form.credentialRef || '',
    owner_name: form.owner || '',
    description: form.description || '',
  }
  try {
    const saved = existing
      ? await updateIntegrationProviderApp(existing.id, payload)
      : await createIntegrationProviderApp(payload)
    const mapped = mapBackendApp(saved)
    if (existing) Object.assign(existing, mapped)
    else apps.push(mapped)
    closeModal()
    notify(existing ? '服务商应用已更新' : '服务商应用已创建')
  } catch (error) {
    notify(readableError(error, '服务商应用保存失败'), 'error')
  }
}
async function savePolicyFromForm() {
  if (!ensureAction(canManageQuota.value, '缺少配额策略管理权限')) return
  const name = String(form.name || '').trim()
  if (!name) {
    setFieldErrors({ name: '请填写策略名称' })
    notify('请先补全策略必填字段', 'error')
    return
  }
  const existing = form.id && policies.find(p => p.id === form.id)
  const code = existing?.id || slugify(form.code || name)
  const payload = {
    code,
    name,
    quota_code: form.quotaCode || 'integration_api_calls_daily',
    quota_unit: form.quotaUnit || 'CALL',
    period_type: form.periodType || 'DAY',
    default_limit: Number(form.dailyLimit || form.defaultLimit || 0),
    over_limit_action: form.exceedStrategy || form.overLimitAction || 'reject',
    status: form.status || 'enabled',
    description: form.description || '',
  }
  try {
    const saved = existing
      ? await updateIntegrationQuotaPolicy(existing.id, payload)
      : await createIntegrationQuotaPolicy(payload)
    const mapped = mapBackendPolicy(saved)
    if (existing) Object.assign(existing, mapped)
    else policies.push(mapped)
    selectedPolicyId.value = mapped.id
    closeModal()
    notify(existing ? '配额策略已更新' : '配额策略已创建')
  } catch (error) {
    notify(readableError(error, '配额策略保存失败'), 'error')
  }
}
async function toggleAppCapability(row, key) {
  if (!ensureAction(canManageApp.value, '缺少服务商应用管理权限')) return
  const target = appCapabilities.find(ac => ac.id === row.id)
  if (!target) return
  const next = !target[key]
  if (!requireNumericId(row.id, '应用能力')) return
  const payloadMap = {
    openToTenant: { open_to_tenant: next },
    defaultEnabled: { enabled: next, default_enabled: next },
    tenantConfigurable: { tenant_configurable: next },
  }
  try {
    await updateIntegrationAppCapability(row.id, payloadMap[key] || { config: { [key]: next } })
  } catch (error) {
    notify(readableError(error, '应用能力更新失败'), 'error')
    return
  }
  target[key] = next
  notify(`${row.name}：${key} 已更新`)
}
async function runHealthCheck(target) {
  if (!ensureAction(canRunHealthCheck.value, '缺少连接治理权限')) return
  try {
    const result = await checkIntegrationConnectivity(target)
    notify(result?.message || `${target}连通性检测已完成`)
  } catch (error) {
    notify(readableError(error, '连通性检测失败'), 'error')
  }
}
async function exportLogs() {
  if (!ensureAction(canManageConnection.value, '缺少连接治理权限')) return
  try {
    const result = await exportIntegrationLogs()
    const url = URL.createObjectURL(result.blob)
    const link = document.createElement('a')
    link.href = url
    link.download = result.filename
    link.click()
    URL.revokeObjectURL(url)
    notify(`已导出 ${result.rowCount} 条调用日志`)
  } catch (error) {
    notify(readableError(error, '调用日志导出失败'), 'error')
  }
}
async function refreshConnection(row) {
  if (!ensureAction(canManageConnection.value, '缺少连接治理权限')) return
  if (!requireNumericId(row.id, '租户连接')) return
  try { await refreshIntegrationTenantConnection(row.id) } catch (error) { notify(readableError(error, '授权刷新失败'), 'error'); return }
  row.authStatus = 'valid'; row.status = 'connected'; notify(`${row.tenantName} 授权状态已刷新`)
}
async function pauseConnection(row) {
  if (!ensureAction(canManageConnection.value, '缺少连接治理权限')) return
  const willPause = row.status !== 'paused'
  if (!requireNumericId(row.id, '租户连接')) return
  const confirmed = await confirmDanger(`${willPause ? '暂停' : '恢复'}租户连接`, `确认${willPause ? '暂停' : '恢复'}「${row.tenantName || row.name || row.id}」的第三方连接？`)
  if (!confirmed) return
  try {
    await (willPause ? pauseIntegrationTenantConnection(row.id) : resumeIntegrationTenantConnection(row.id))
  } catch (error) {
    notify(readableError(error, '连接状态更新失败'), 'error')
    return
  }
  row.status = willPause ? 'paused' : 'connected'; notify(`${row.tenantName} 连接已${row.status === 'paused' ? '暂停' : '恢复'}`)
}
async function retryConnection(row) {
  if (!ensureAction(canManageConnection.value, '缺少连接治理权限')) return
  if (!requireNumericId(row.id, '租户连接')) return
  try { await retryIntegrationTenantConnection(row.id) } catch (error) { notify(readableError(error, '连接重试失败'), 'error'); return }
  notify(`${row.tenantName} 同步任务已加入重试队列`)
}
async function retrySync(row) {
  if (!ensureAction(canManageConnection.value, '缺少连接治理权限')) return
  if (!requireNumericId(row.id, '同步任务')) return
  try { await retryIntegrationSyncJob(row.id) } catch (error) { notify(readableError(error, '同步任务重试失败'), 'error'); return }
  row.status = 'running'; notify(`${row.job} 已开始重试`)
}
async function toggleSync(row) {
  if (!ensureAction(canManageConnection.value, '缺少连接治理权限')) return
  const willPause = row.status !== 'paused'
  if (!requireNumericId(row.id, '同步任务')) return
  const confirmed = await confirmDanger(`${willPause ? '暂停' : '恢复'}同步任务`, `确认${willPause ? '暂停' : '恢复'}「${row.job || row.id}」？`)
  if (!confirmed) return
  try {
    await (willPause ? pauseIntegrationSyncJob(row.id) : resumeIntegrationSyncJob(row.id))
  } catch (error) {
    notify(readableError(error, '同步任务状态更新失败'), 'error')
    return
  }
  row.status = willPause ? 'paused' : 'running'; notify(`${row.job} 已${row.status === 'paused' ? '暂停' : '恢复'}`)
}
async function copyPolicy(policy) {
  if (!ensureAction(canManageQuota.value, '缺少配额策略管理权限')) return
  const copy = { ...policy, id: `${policy.id}-copy-${Date.now()}`, name: `${policy.name} - 专属覆盖`, isOverride: true, priority: policy.priority + 200, dailyLimit: Math.round(policy.dailyLimit * 2) }
  try {
    const saved = await createIntegrationQuotaPolicy({
      code: copy.id,
      name: copy.name,
      quota_code: 'integration_api_calls_daily',
      quota_unit: 'CALL',
      period_type: 'DAY',
      default_limit: copy.dailyLimit,
      over_limit_action: copy.exceedStrategy || 'reject',
      status: 'enabled',
      description: '由通用策略复制生成的专属覆盖策略。',
    })
    const mapped = mapBackendPolicy(saved)
    policies.push(mapped)
    selectedPolicyId.value = mapped.id
    notify('已复制为专属覆盖策略')
  } catch (error) {
    notify(readableError(error, '复制策略失败'), 'error')
    return
  }
}
async function toggleStatus(obj) {
  if (!ensureAction(canManageQuota.value, '缺少配额策略管理权限')) return
  const enabled = obj.status !== 'enabled'
  if (!enabled) {
    const confirmed = await confirmDanger('停用配额策略', `确认停用「${obj.name || obj.title || obj.id}」？`)
    if (!confirmed) return
  }
  if (policies.includes(obj)) {
    try { await updateIntegrationQuotaPolicyStatus(obj.id, enabled) } catch (error) { notify(readableError(error, '策略状态更新失败'), 'error'); return }
  }
  obj.status = enabled ? 'enabled' : 'disabled'; notify(`${obj.name || obj.title} 已${obj.status === 'enabled' ? '启用' : '停用'}`)
}
async function processAlert(alert) {
  if (!ensureAction(canManageConnection.value, '缺少连接治理权限')) return
  if (!requireNumericId(alert.id, '异常记录')) return
  try { await processIntegrationAlert(alert.id) } catch (error) { notify(readableError(error, '异常处理失败'), 'error'); return }
  alert.status = 'processing'; notify('异常已标记处理中')
}
async function resolveAlert(alert) {
  if (!ensureAction(canManageConnection.value, '缺少连接治理权限')) return
  if (!requireNumericId(alert.id, '异常记录')) return
  const confirmed = await confirmDanger('标记异常恢复', `确认将「${alert.title || alert.id}」标记为已恢复？`)
  if (!confirmed) return
  try { await resolveIntegrationAlert(alert.id) } catch (error) { notify(readableError(error, '异常恢复失败'), 'error'); return }
  alert.status = 'resolved'; notify('异常已标记恢复')
}
async function ignoreAlert(alert) {
  if (!ensureAction(canManageConnection.value, '缺少连接治理权限')) return
  if (!requireNumericId(alert.id, '异常记录')) return
  const confirmed = await confirmDanger('忽略异常', `确认忽略「${alert.title || alert.id}」？`)
  if (!confirmed) return
  try { await ignoreIntegrationAlert(alert.id) } catch (error) { notify(readableError(error, '异常忽略失败'), 'error'); return }
  alert.status = 'ignored'; notify('异常已忽略')
}
async function confirmDanger(title, message) {
  try {
    await ElMessageBox.confirm(message, title, {
      confirmButtonText: '确认',
      cancelButtonText: '取消',
      type: 'warning',
      distinguishCancelAndClose: true,
    })
    return true
  } catch {
    return false
  }
}
function isNumericId(id) { return /^\d+$/.test(String(id || '')) }
function requireNumericId(id, label) {
  if (isNumericId(id)) return true
  notify(`${label}缺少后端记录 ID，请刷新或完成初始化后重试`, 'error')
  return false
}
function slugify(value) { return String(value || '').trim().toLowerCase().replace(/[^a-z0-9_-]+/g, '_').replace(/^_+|_+$/g, '') || `item_${Date.now()}` }
function copyText(text) { navigator?.clipboard?.writeText(text); notify('已复制') }
function appSecretItems(app) { return [{ label: 'suite_id / app_key', value: app.suiteId || app.credentialRef || '未配置' }, { label: 'suite_secret / app_secret', value: app.suiteSecret || '由密钥引用托管' }, { label: 'Token', value: app.token || '由密钥引用托管' }, { label: 'EncodingAESKey', value: app.encodingAesKey || '由密钥引用托管' }, { label: 'Endpoint', value: app.endpoint || '-' }] }

const StatusBadge = defineComponent({
  props: { status: { type: [String, Boolean], default: 'normal' } },
  setup(props) {
    const labels = { enabled: '启用', disabled: '停用', draft: '草稿', maintenance: '维护中', connected: '正常', normal: '正常', success: '成功', valid: '有效', approved: '已授权', warning: '预警', testing: '测试中', expiring: '即将过期', running: '执行中', open: '待处理', processing: '处理中', error: '异常', failed: '失败', expired: '已过期', exceeded: '超限', not_connected: '未接通', not_applied: '未申请', paused: '暂停', ignored: '已忽略', resolved: '已恢复' }
    return () => h('span', { class: ['status-badge', props.status] }, labels[props.status] || props.status)
  }
})

const SwitchToggle = defineComponent({
  props: { modelValue: Boolean, disabled: Boolean },
  emits: ['update:modelValue'],
  setup(props, { emit }) {
    return () => h('button', {
      class: ['switch', props.modelValue ? 'on' : '', props.disabled ? 'disabled' : ''],
      disabled: props.disabled,
      onClick: () => { if (!props.disabled) emit('update:modelValue', !props.modelValue) },
    }, [h('i')])
  }
})

const DataTable = defineComponent({
  props: { columns: Array, rows: Array, minWidth: { type: String, default: '960px' }, pagination: Object },
  emits: ['page'],
  setup(props, { slots, emit }) {
    const emptyRows = () => h('tr', { class: 'empty-row' }, [h('td', { colspan: props.columns.length }, [h(EmptyState, { title: '暂无数据' })])])
    const pager = () => props.pagination ? h(DataPager, { pagination: props.pagination, onPage: skip => emit('page', skip) }) : null
    return () => h('div', { class: 'data-table-wrap' }, [
      h('table', { style: { minWidth: props.minWidth } }, [
        h('thead', [h('tr', props.columns.map(c => h('th', c.label)))]),
        h('tbody', props.rows.length ? props.rows.map(row => h('tr', { key: row.id }, props.columns.map(c => h('td', slots[`cell-${c.key}`] ? slots[`cell-${c.key}`]({ row }) : row[c.key])))) : [emptyRows()])
      ]),
      pager(),
    ])
  }
})

const DataPager = defineComponent({
  props: { pagination: Object },
  emits: ['page'],
  setup(props, { emit }) {
    const go = skip => emit('page', skip)
    return () => {
      const total = Number(props.pagination?.total || 0)
      if (total <= 0) return null
      const limit = Math.max(1, Number(props.pagination?.limit || 20))
      const skip = Math.max(0, Number(props.pagination?.skip || 0))
      const current = Math.floor(skip / limit) + 1
      const pages = Math.max(1, Math.ceil(total / limit))
      return h('div', { class: 'data-pager' }, [
        h('span', `共 ${total.toLocaleString('zh-CN')} 条 · 第 ${current} / ${pages} 页`),
        h('div', { class: 'pager-actions' }, [
          h('button', { disabled: skip <= 0, onClick: () => go(Math.max(0, skip - limit)) }, '上一页'),
          h('button', { disabled: skip + limit >= total, onClick: () => go(skip + limit) }, '下一页'),
        ]),
      ])
    }
  },
})

const RankingList = defineComponent({
  props: { items: Array },
  setup(props) { return () => h('div', { class: 'ranking-list' }, props.items.map((item, index) => h('div', { class: 'ranking-row', key: item.name }, [h('span', { class: 'rank-index' }, index + 1), h('div', [h('strong', item.name), h('p', item.desc)]), h('b', item.value)]))) }
})

const EmptyState = defineComponent({
  props: { title: { type: String, default: '暂无数据' } },
  setup(props) {
    return () => h('div', { class: 'empty-state' }, props.title)
  },
})

const PlatformRail = defineComponent({
  props: { platforms: Array, selectedId: String, showClear: { type: Boolean, default: true } },
  emits: ['select', 'clear'],
  setup(props, { emit }) {
    return () => h('aside', { class: 'platform-rail panel' }, [
      h('div', { class: 'rail-head' }, [h('div', [h('h3', '接入平台'), h('p', '左栏始终是平台')]), ...(props.showClear ? [h('button', { class: ['small-btn', props.selectedId === 'all' ? 'active' : ''], onClick: () => emit('clear') }, '全部')] : [])]),
      h('div', { class: 'rail-list' }, props.platforms.map(p => h('button', { class: ['rail-item', props.selectedId === p.id ? 'active' : ''], key: p.id, onClick: () => emit('select', p.id) }, [h('span', { class: 'avatar small' }, p.icon), h('div', [h('strong', p.shortName || p.name), h('small', `${p.appCount} 应用 · ${p.connectionCount} 连接`)]), h(StatusBadge, { status: p.status })])))
    ])
  }
})

const ConnectionTable = defineComponent({
  props: { rows: Array, canManage: Boolean, pagination: Object },
  emits: ['open', 'refresh', 'pause', 'retry', 'page'],
  setup(props, { emit }) {
    const cols = [
      { key: 'tenant', label: '租户 / 授权主体' }, { key: 'platform', label: '平台' }, { key: 'app', label: '服务商应用' }, { key: 'auth', label: '授权状态' }, { key: 'status', label: '连接状态' }, { key: 'caps', label: '最终能力' }, { key: 'calls', label: '今日调用' }, { key: 'last', label: '最近同步' }, { key: 'actions', label: '操作' }
    ]
    return () => h(DataTable, { columns: cols, rows: props.rows, pagination: props.pagination, minWidth: '1260px', onPage: skip => emit('page', skip) }, {
      'cell-tenant': ({ row }) => h('div', [h('strong', row.tenantName), h('p', { class: 'muted' }, row.authSubject)]),
      'cell-platform': ({ row }) => row.platformName || getPlatformName(row.platformId),
      'cell-app': ({ row }) => row.appName || getAppName(row.appId),
      'cell-auth': ({ row }) => h(StatusBadge, { status: row.authStatus }),
      'cell-status': ({ row }) => h(StatusBadge, { status: row.status }),
      'cell-caps': ({ row }) => `${row.finalCapabilityCount} 项`,
      'cell-calls': ({ row }) => formatNumber(row.callsToday),
      'cell-last': ({ row }) => row.lastSync,
      'cell-actions': ({ row }) => h('div', { class: 'actions wide' }, [
        h('button', { onClick: () => emit('open', row) }, '详情'),
        ...(props.canManage ? [
          h('button', { onClick: () => emit('refresh', row) }, '刷新授权'),
          h('button', { onClick: () => emit('pause', row) }, row.status === 'paused' ? '恢复' : '暂停'),
          h('button', { onClick: () => emit('retry', row) }, '重试'),
        ] : []),
      ])
    })
  }
})
</script>
