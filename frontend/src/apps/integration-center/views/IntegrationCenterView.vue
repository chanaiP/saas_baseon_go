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
          <button class="ghost-btn" @click="runHealthCheck('全局')">全局连通性检测</button>
          <button class="primary-btn" @click="primaryAction">{{ currentNav.action }}</button>
        </div>
      </header>

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
          </section>
          <section class="panel">
            <div class="panel-head"><h3>配额预警</h3></div>
            <div class="quota-mini-list">
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
          <div class="search-box"><span>⌕</span><input v-model="platformKeyword" placeholder="搜索平台名称、code、类型" /></div>
          <div class="toolbar-actions">
            <select v-model="platformTypeFilter"><option value="all">全部类型</option><option value="协同办公">协同办公</option><option value="电商平台">电商平台</option><option value="ERP">ERP</option><option value="CRM">CRM</option><option value="WMS">WMS</option></select>
            <select v-model="platformStatusFilter"><option value="all">全部状态</option><option value="draft">草稿</option><option value="enabled">启用</option><option value="disabled">停用</option><option value="maintenance">维护中</option></select>
            <button class="primary-btn" @click="openPlatformModal()">新增接入平台</button>
          </div>
        </div>

        <div class="platform-card-grid">
          <article v-for="platform in filteredPlatforms" :key="platform.id" class="platform-card">
            <div class="platform-card-head">
              <div class="avatar large">{{ platform.icon }}</div>
              <div>
                <h3>{{ platform.shortName || platform.name }}</h3>
                <p>{{ platform.name }} · {{ platform.code }} · {{ platform.type }} · {{ platform.accessType }}</p>
              </div>
              <StatusBadge :status="platform.status" />
            </div>
            <p class="platform-desc">{{ platform.description }}</p>
            <div class="card-stats four">
              <div><span>应用</span><b>{{ platform.appCount }}</b></div>
              <div><span>能力</span><b>{{ platform.capabilityCount }}</b></div>
              <div><span>连接</span><b>{{ platform.connectionCount }}</b></div>
              <div><span>异常</span><b>{{ platform.alertCount }}</b></div>
            </div>
            <div class="usage-line">
              <div class="row-between"><span>今日调用</span><strong>{{ formatNumber(platform.callsToday) }}</strong></div>
              <div class="progress-line"><span :style="{ width: platform.successRate + '%' }"></span></div>
            </div>
            <div class="card-actions wrap">
              <button @click="openPlatformDrawer(platform)">查看详情</button>
              <button @click="openPlatformModal(platform)">编辑平台</button>
              <button @click="jumpWorkspace(platform.id, 'apps')">配置应用</button>
              <button @click="jumpWorkspace(platform.id, 'capabilities')">配置能力</button>
              <button @click="runHealthCheck(platform.name)">检测连通性</button>
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
                <div class="toolbar-actions"><button class="ghost-btn" @click="clearSelectedApp">全部应用</button><button class="primary-btn" @click="openAppModal()">新增应用</button></div>
              </div>
              <div class="app-card-grid">
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
                <button class="ghost-btn" @click="selectedWorkspaceApp ? openAppDrawer(selectedWorkspaceApp, 'capabilities') : runHealthCheck('当前应用能力连接')">{{ selectedWorkspaceApp ? '打开应用详情' : '批量检测' }}</button>
              </div>
              <DataTable :columns="appCapabilityColumns" :rows="workspaceAppCapabilityRows" min-width="1180px">
                <template #cell-name="{ row }"><strong>{{ row.name }}</strong><p class="muted">{{ row.code }}</p></template>
                <template #cell-app="{ row }">{{ row.appName }}</template>
                <template #cell-connected="{ row }"><StatusBadge :status="row.connected ? 'connected' : 'not_connected'" /></template>
                <template #cell-auth="{ row }"><StatusBadge :status="row.authStatus" /></template>
                <template #cell-open="{ row }"><SwitchToggle :model-value="row.openToTenant" @update:model-value="toggleAppCapability(row, 'openToTenant')" /></template>
                <template #cell-default="{ row }"><SwitchToggle :model-value="row.defaultEnabled" @update:model-value="toggleAppCapability(row, 'defaultEnabled')" /></template>
                <template #cell-configurable="{ row }"><SwitchToggle :model-value="row.tenantConfigurable" @update:model-value="toggleAppCapability(row, 'tenantConfigurable')" /></template>
                <template #cell-health="{ row }"><StatusBadge :status="row.health" /></template>
                <template #cell-actions="{ row }"><div class="actions"><button @click="openCapabilityDrawer(row)">详情</button><button @click="runHealthCheck(row.name)">检测</button></div></template>
              </DataTable>
            </section>
          </template>

          <template v-else-if="workspaceTab === 'capabilities'">
            <section class="panel table-panel">
              <div class="panel-head padded-head">
                <div><h3>平台能力定义</h3><p>这里只定义平台理论能力，不直接生成租户能力。</p></div>
                <button class="primary-btn" @click="openCapabilityModal()">新增平台能力</button>
              </div>
              <DataTable :columns="platformCapabilityColumns" :rows="workspaceCapabilities" min-width="1050px">
                <template #cell-name="{ row }"><strong>{{ row.name }}</strong><p class="muted">{{ row.code }}</p></template>
                <template #cell-type="{ row }"><span class="tag">{{ row.type }}</span></template>
                <template #cell-basic="{ row }">{{ row.isBasic ? '是' : '否' }}</template>
                <template #cell-status="{ row }"><StatusBadge :status="row.status" /></template>
                <template #cell-actions="{ row }"><div class="actions"><button @click="openCapabilityModal(row)">编辑</button><button @click="openCapabilityDrawer(row)">详情</button></div></template>
              </DataTable>
            </section>
          </template>

          <template v-else-if="workspaceTab === 'connections'">
            <section class="panel table-panel">
              <div class="panel-head padded-head">
                <div><h3>当前平台连接实例</h3><p>连接实例按应用建立，能力进入连接详情查看。</p></div>
                <button class="ghost-btn" @click="switchPage('tenantConnections')">进入租户连接</button>
              </div>
              <ConnectionTable :rows="workspaceConnections" @open="openConnectionDrawer" @refresh="refreshConnection" @pause="pauseConnection" @retry="retryConnection" />
            </section>
          </template>
        </div>
      </section>

      <section v-if="page === 'tenantConnections'" class="page-section tenant-layout">
        <PlatformRail
          :platforms="platformsSorted"
          :selected-id="tenantPlatformId"
          :show-clear="false"
          @select="selectTenantPlatform"
          @clear="clearTenantPlatform"
        />

        <div class="tenant-main">
          <section class="panel relation-upper fixed-upper">
            <div class="panel-head">
              <div>
                <h3>服务商应用</h3>
                <p>未选中平台时展示全部应用；选中平台后只展示当前平台应用。选中应用后，下方只显示该应用下的连接实例。</p>
              </div>
              <div class="toolbar-actions"><span class="tag">{{ tenantAppsCapabilityTotal }} 能力</span><button class="ghost-btn" @click="clearTenantPlatform">全部应用</button><button class="primary-btn" @click="openAppModal()">新增应用</button></div>
            </div>
            <div class="app-list-grid">
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
            <ConnectionTable :rows="tenantConnections" @open="openConnectionDrawer" @refresh="refreshConnection" @pause="pauseConnection" @retry="retryConnection" />
          </section>
        </div>
      </section>

      <section v-if="page === 'syncMonitor'" class="page-section">
        <div class="toolbar row-between">
          <div class="search-box"><span>⌕</span><input v-model="syncKeyword" placeholder="搜索租户、平台、能力、任务" /></div>
          <div class="toolbar-actions"><select v-model="syncStatus"><option value="all">全部状态</option><option value="success">成功</option><option value="running">执行中</option><option value="failed">失败</option><option value="paused">暂停</option></select></div>
        </div>
        <section class="panel table-panel">
          <DataTable :columns="syncColumns" :rows="filteredSyncJobs" min-width="1180px">
            <template #cell-job="{ row }"><strong>{{ row.job }}</strong><p class="muted">{{ row.mode }} · {{ row.cron }}</p></template>
            <template #cell-platform="{ row }">{{ getPlatformName(row.platformId) }}</template>
            <template #cell-status="{ row }"><StatusBadge :status="row.status" /></template>
            <template #cell-rate="{ row }"><div class="quota inline"><span :style="{ width: row.successRate + '%' }"></span></div><small>{{ row.successRate }}%</small></template>
            <template #cell-actions="{ row }"><div class="actions wide"><button @click="retrySync(row)">重试</button><button @click="toggleSync(row)">{{ row.status === 'paused' ? '恢复' : '暂停' }}</button><button @click="openSyncDrawer(row)">日志</button></div></template>
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
              <div class="toolbar-actions"><button class="ghost-btn" @click="resetQuotaFilters">全部</button><button class="primary-btn" @click="openPolicyModal()">新增策略</button></div>
            </div>
            <div class="policy-grid">
              <article v-for="policy in filteredPolicies" :key="policy.id" :class="['policy-card selectable', { selected: selectedPolicyId === policy.id }]" @click="selectedPolicyId = policy.id">
                <div class="row-between"><strong>{{ policy.name }}</strong><StatusBadge :status="policy.status" /></div>
                <p>{{ policy.scopeLabel }} · {{ policy.isOverride ? '专属覆盖' : '通用策略' }} · 优先级 {{ policy.priority }}</p>
                <div class="policy-limits">
                  <span>日 {{ formatLimit(policy.dailyLimit) }}</span>
                  <span>月 {{ formatLimit(policy.monthlyLimit) }}</span>
                  <span>QPS {{ policy.qpsLimit || '-' }}</span>
                  <span>并发 {{ policy.concurrentLimit || '-' }}</span>
                </div>
                <div class="card-actions"><button @click.stop="openPolicyModal(policy)">编辑</button><button @click.stop="copyPolicy(policy)">复制为专属</button><button @click.stop="toggleStatus(policy)">{{ policy.status === 'enabled' ? '停用' : '启用' }}</button></div>
              </article>
            </div>
          </section>

          <section class="panel relation-lower">
            <div class="panel-head">
              <div><h3>企业覆盖策略</h3><p>{{ selectedPolicy ? selectedPolicy.name : '请选择上方策略查看绑定对象、覆盖来源和实时用量。' }}</p></div>
              <button class="ghost-btn" @click="selectedPolicy && openBindingDrawer(selectedPolicy)">添加租户覆盖策略</button>
            </div>
            <DataTable :columns="quotaUsageColumns" :rows="selectedPolicyUsage" min-width="1040px">
              <template #cell-object="{ row }"><strong>{{ row.name }}</strong><p class="muted">{{ row.objectPath }}</p></template>
              <template #cell-source="{ row }"><span :class="['tag', row.override ? 'danger-soft' : '']">{{ row.override ? '专属覆盖' : '通用策略' }}</span></template>
              <template #cell-used="{ row }"><div>{{ row.used.toLocaleString() }} / {{ row.limit.toLocaleString() }}</div><div class="quota inline"><span :style="{ width: Math.min(row.rate, 100) + '%' }"></span></div></template>
              <template #cell-status="{ row }"><StatusBadge :status="row.rate >= 100 ? 'exceeded' : row.rate >= 80 ? 'warning' : 'normal'" /></template>
              <template #cell-actions="{ row }"><div class="actions"><button @click="openOverrideModal(row)">编辑</button><button @click="openUsageDrawer(row)">明细</button></div></template>
            </DataTable>
          </section>
        </div>
      </section>

      <section v-if="page === 'alerts'" class="page-section">
        <div class="toolbar row-between">
          <div class="search-box"><span>⌕</span><input v-model="alertKeyword" placeholder="搜索异常、租户、平台、能力" /></div>
          <div class="toolbar-actions"><select v-model="alertStatus"><option value="all">全部状态</option><option value="open">待处理</option><option value="processing">处理中</option><option value="resolved">已恢复</option></select></div>
        </div>
        <section class="panel">
          <div class="alert-list">
            <article v-for="alert in filteredAlerts" :key="alert.id" class="alert-item full">
              <div :class="['alert-dot', alert.level]"></div>
              <div class="alert-body">
                <div class="row-between"><strong>{{ alert.title }}</strong><StatusBadge :status="alert.status" /></div>
                <p>{{ alert.message }}</p>
                <small>{{ alert.path }} · {{ alert.lastAt }} · {{ alert.count }} 次</small>
              </div>
              <div class="actions wide"><button @click="processAlert(alert)">处理</button><button @click="resolveAlert(alert)">标记恢复</button><button @click="ignoreAlert(alert)">忽略</button><button @click="openAlertDrawer(alert)">详情</button></div>
            </article>
          </div>
        </section>
      </section>

      <section v-if="page === 'logs'" class="page-section">
        <div class="toolbar row-between">
          <div class="search-box"><span>⌕</span><input v-model="logKeyword" placeholder="搜索 request_id、平台、租户、Endpoint、错误码" /></div>
          <div class="toolbar-actions"><select v-model="logType"><option value="all">全部类型</option><option value="api">第三方 API</option><option value="token">Token 刷新</option><option value="callback">回调接收</option><option value="data_write">数据写入</option></select><select v-model="logSuccess"><option value="all">全部结果</option><option value="success">成功</option><option value="failed">失败</option></select></div>
        </div>
        <section class="panel table-panel">
          <DataTable :columns="logColumns" :rows="filteredLogs" min-width="1220px">
            <template #cell-id="{ row }"><strong>{{ row.requestId }}</strong><p class="muted">{{ row.calledAt }}</p></template>
            <template #cell-type="{ row }"><span class="tag">{{ row.typeLabel }}</span></template>
            <template #cell-path="{ row }"><strong>{{ row.method }}</strong> {{ row.path }}<p class="muted">{{ row.endpoint }}</p></template>
            <template #cell-result="{ row }"><StatusBadge :status="row.success ? 'success' : 'failed'" /></template>
            <template #cell-actions="{ row }"><div class="actions"><button @click="openLogDrawer(row)">详情</button><button @click="copyText(row.requestId)">复制 ID</button></div></template>
          </DataTable>
        </section>
      </section>
    </main>

    <div v-if="drawer.open" class="drawer-mask" @click="closeDrawer"></div>
    <aside v-if="drawer.open" class="drawer">
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
        <section class="drawer-card"><h4>快捷操作</h4><div class="actions wide"><button @click="jumpWorkspace(drawer.data.id, 'apps')">配置应用</button><button @click="jumpWorkspace(drawer.data.id, 'capabilities')">配置能力</button><button @click="runHealthCheck(drawer.data.name)">连通性检测</button></div></section>
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
            <template #cell-open="{ row }"><SwitchToggle :model-value="row.openToTenant" @update:model-value="toggleAppCapability(row, 'openToTenant')" /></template>
            <template #cell-default="{ row }"><SwitchToggle :model-value="row.defaultEnabled" @update:model-value="toggleAppCapability(row, 'defaultEnabled')" /></template>
            <template #cell-configurable="{ row }"><SwitchToggle :model-value="row.tenantConfigurable" @update:model-value="toggleAppCapability(row, 'tenantConfigurable')" /></template>
            <template #cell-health="{ row }"><StatusBadge :status="row.health" /></template>
          </DataTable>
        </template>
        <template v-if="drawerTab === 'tenants'">
          <ConnectionTable :rows="connectionsByApp(drawer.data.id)" @open="openConnectionDrawer" @refresh="refreshConnection" @pause="pauseConnection" @retry="retryConnection" />
        </template>
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
              <button @click="openOverrideModal(usage)">专属覆盖</button>
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
      </div>

      <div v-if="['capability','policy','usage','alert','log','sync'].includes(drawer.type)" class="drawer-body">
        <section class="drawer-card"><h4>对象详情</h4><pre>{{ pretty(drawer.data) }}</pre></section>
      </div>
    </aside>

    <div v-if="modal.open" class="modal-mask" @click.self="closeModal">
      <section class="modal-card" :class="{ large: modal.type === 'platform' }">
        <div class="modal-head">
          <div><p class="eyebrow">{{ modal.subtitle }}</p><h3>{{ modal.title }}</h3></div>
          <button class="icon-btn" @click="closeModal">×</button>
        </div>
        <div v-if="modal.type === 'platform'" class="form-grid">
          <label>平台名称<input v-model="form.name" placeholder="如 企业微信、京东电商" /></label>
          <label>平台简称<input v-model="form.shortName" placeholder="左栏、卡片、表格展示" /></label>
          <label>平台编码<input v-model="form.code" :readonly="!!form.id" placeholder="wecom、jd… 创建后不建议修改" /></label>
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
          <label>应用名称<input v-model="form.name" /></label>
          <label>所属平台<select v-model="form.platformId"><option v-for="p in platforms" :key="p.id" :value="p.id">{{ p.name }}</option></select></label>
          <label>应用类型<select v-model="form.type"><option>第三方服务商应用</option><option>正式应用</option><option>沙箱应用</option><option>专属应用</option></select></label>
          <label>环境<select v-model="form.env"><option>正式</option><option>沙箱</option><option>测试</option></select></label>
          <label class="span-2">Endpoint / 授权域名<input v-model="form.endpoint" /></label>
        </div>
        <div v-else-if="modal.type === 'policy'" class="form-grid">
          <label>策略名称<input v-model="form.name" /></label>
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
        <div v-else class="modal-empty">该操作会写入模拟数据并触发页面刷新。</div>
        <div v-if="modal.type === 'platform'" class="modal-actions">
          <button class="ghost-btn" @click="closeModal">取消</button>
          <button class="ghost-btn" @click="savePlatformFromForm(false)">保存</button>
          <button class="primary-btn" @click="savePlatformFromForm(true)">保存并配置应用</button>
        </div>
        <div v-else class="modal-actions"><button class="ghost-btn" @click="closeModal">取消</button><button class="primary-btn" @click="saveModal">保存</button></div>
      </section>
    </div>

    <div v-if="globalSearchOpen" class="modal-mask" @click.self="globalSearchOpen = false">
      <section class="modal-card large">
        <div class="modal-head"><div><p class="eyebrow">Global Search</p><h3>全局搜索</h3></div><button class="icon-btn" @click="globalSearchOpen = false">×</button></div>
        <div class="search-box full"><span>⌕</span><input v-model="globalKeyword" autofocus placeholder="搜索平台、应用、租户、连接实例、request_id" /></div>
        <div class="global-results">
          <article v-for="item in globalResults" :key="item.key" @click="openGlobalResult(item)">
            <span>{{ item.type }}</span><strong>{{ item.title }}</strong><p>{{ item.desc }}</p>
          </article>
        </div>
      </section>
    </div>

    <Transition name="toast"><div v-if="toast.show" :class="['toast-box', toast.type]">{{ toast.message }}</div></Transition>
  </div>
</template>

<script setup>
import { computed, defineComponent, h, onMounted, reactive, ref, watch } from 'vue'
import { useRoute } from 'vue-router'

import {
  createIntegrationPlatform,
  fetchIntegrationAlerts,
  fetchIntegrationLogs,
  fetchIntegrationPlatforms,
  fetchIntegrationQuota,
  fetchIntegrationSyncMonitor,
  fetchIntegrationTenantConnections,
  fetchIntegrationWorkspace,
  updateIntegrationPlatform,
} from '../api'

const navItems = [
  { key: 'overview', label: '总览', icon: '⌂', action: '刷新总览' },
  { key: 'platforms', label: '接入平台', icon: '▣', action: '新增平台' },
  { key: 'workspace', label: '集成工作台', icon: '⌘', action: '新增应用' },
  { key: 'tenantConnections', label: '租户连接', icon: '⇄', action: '刷新列表' },
  { key: 'syncMonitor', label: '同步监控', icon: '↻', action: '创建任务' },
  { key: 'quota', label: '配额与限流', icon: '◴', action: '新增策略' },
  { key: 'alerts', label: '异常监控', icon: '!', action: '批量处理', badge: 5 },
  { key: 'logs', label: '调用日志', icon: '☷', action: '导出日志' },
]

const route = useRoute()
const routePageMap = {
  '/integration-center': 'overview',
  '/integration-center/platforms': 'platforms',
  '/integration-center/workspace': 'workspace',
  '/integration-center/tenant-connections': 'tenantConnections',
  '/integration-center/sync-monitor': 'syncMonitor',
  '/integration-center/quota': 'quota',
  '/integration-center/alerts': 'alerts',
  '/integration-center/logs': 'logs',
}
const page = ref(routePageMap[route.path] || 'overview')
const workspacePlatformId = ref('wecom')
const workspaceTab = ref('apps')
const workspaceSelectedAppId = ref('wecom-suite-main')
const tenantPlatformId = ref('all')
const tenantAppId = ref('all')
const quotaPlatformId = ref('all')
const quotaMode = ref('all')
const selectedPolicyId = ref('policy-tenant-standard')
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

const toast = reactive({ show: false, message: '', type: 'success' })
const drawer = reactive({ open: false, type: '', title: '', subtitle: '', desc: '', data: null })
const modal = reactive({ open: false, type: '', title: '', subtitle: '' })
const form = reactive({})
const backendState = reactive({ loading: false, loaded: false, error: '' })

const platforms = reactive([
  { id: 'wecom', icon: '企', shortName: '企微', name: '企业微信', code: 'wecom', type: '协同办公', accessType: '第三方服务商', owner: '平台集成组', status: 'enabled', tenantVisible: true, officialUrl: 'https://developer.work.weixin.qq.com/', sortWeight: 10, appCount: 2, capabilityCount: 9, connectionCount: 128, alertCount: 5, callsToday: 32680, successRate: 98, description: '企业微信第三方服务商模式，支持组织架构、成员、部门、消息推送、应用安装授权和事件回调。' },
  { id: 'jd', icon: '京', shortName: '京东', name: '京东电商', code: 'jd', type: '电商平台', accessType: 'OAuth2', owner: '电商集成组', status: 'enabled', tenantVisible: true, officialUrl: 'https://open.jd.com/', sortWeight: 20, appCount: 3, capabilityCount: 12, connectionCount: 86, alertCount: 2, callsToday: 52100, successRate: 96, description: '京东开放平台服务商应用，支持店铺、商品、订单、售后、库存等数据同步。' },
  { id: 'douyin', icon: '抖', shortName: '抖店', name: '抖音电商', code: 'douyin', type: '电商平台', accessType: 'OAuth2', owner: '电商集成组', status: 'maintenance', tenantVisible: true, officialUrl: 'https://partner.open.jinritemai.com/', sortWeight: 30, appCount: 2, capabilityCount: 10, connectionCount: 64, alertCount: 8, callsToday: 44210, successRate: 88, description: '抖店开放平台应用授权，支持店铺、商品、订单、售后、素材与消息订阅。' },
  { id: 'taobao', icon: '淘', shortName: '淘天', name: '淘宝/天猫', code: 'taobao', type: '电商平台', accessType: 'OAuth2', owner: '开放平台组', status: 'enabled', tenantVisible: true, officialUrl: 'https://open.taobao.com/', sortWeight: 40, appCount: 2, capabilityCount: 11, connectionCount: 72, alertCount: 1, callsToday: 38120, successRate: 97, description: '淘宝开放平台应用，支持商品、交易、退款、评价与店铺数据同步。' },
])

const platformsSorted = computed(() => [...platforms].sort((a, b) => (a.sortWeight ?? 999) - (b.sortWeight ?? 999) || String(a.name).localeCompare(String(b.name), 'zh')))

const apps = reactive([
  { id: 'wecom-suite-main', platformId: 'wecom', name: '企业微信第三方标准应用', type: '第三方服务商应用', env: '正式', status: 'enabled', capabilityCount: 8, connectionCount: 118, alertCount: 4, callsToday: 30210, endpoint: 'https://api.weixin.qq.com/cgi-bin', suiteId: 'wwsuite_8f91a0d6', suiteSecret: '******5f2a', token: '******token', encodingAesKey: '******AESKey', commandCallback: 'https://api.ec-aios.com/callback/wecom/command', dataCallback: 'https://api.ec-aios.com/callback/wecom/data', authCallback: 'https://console.ec-aios.com/oauth/wecom/callback', suiteTicketStatus: '已接收', suiteTicketAt: '2026-05-14 09:42', suiteTokenStatus: '有效', suiteTokenExpireAt: '2026-05-14 11:42' },
  { id: 'wecom-suite-test', platformId: 'wecom', name: '企业微信沙箱测试应用', type: '第三方服务商应用', env: '测试', status: 'testing', capabilityCount: 5, connectionCount: 10, alertCount: 1, callsToday: 2470, endpoint: 'https://api.weixin.qq.com/cgi-bin', suiteId: 'wwsuite_test_29b2', suiteSecret: '******test', token: '******token', encodingAesKey: '******AESKey', commandCallback: 'https://staging.ec-aios.com/callback/wecom/command', dataCallback: 'https://staging.ec-aios.com/callback/wecom/data', authCallback: 'https://staging.ec-aios.com/oauth/wecom/callback', suiteTicketStatus: '已接收', suiteTicketAt: '2026-05-14 09:31', suiteTokenStatus: '有效', suiteTokenExpireAt: '2026-05-14 11:31' },
  { id: 'jd-main', platformId: 'jd', name: '京东正式服务商应用 A', type: '正式应用', env: '正式', status: 'enabled', capabilityCount: 11, connectionCount: 72, alertCount: 1, callsToday: 43100, endpoint: 'https://api.jd.com/routerjson' },
  { id: 'jd-backup', platformId: 'jd', name: '京东备用应用 B', type: '正式应用', env: '正式', status: 'enabled', capabilityCount: 9, connectionCount: 14, alertCount: 1, callsToday: 9000, endpoint: 'https://api.jd.com/routerjson' },
  { id: 'douyin-main', platformId: 'douyin', name: '抖音电商正式应用', type: '正式应用', env: '正式', status: 'maintenance', capabilityCount: 8, connectionCount: 64, alertCount: 8, callsToday: 44210, endpoint: 'https://openapi-fxg.jinritemai.com' },
  { id: 'taobao-main', platformId: 'taobao', name: '淘宝天猫正式应用', type: '正式应用', env: '正式', status: 'enabled', capabilityCount: 10, connectionCount: 72, alertCount: 1, callsToday: 38120, endpoint: 'https://eco.taobao.com/router/rest' },
])

const capabilities = reactive([
  { id: 'wecom-auth-check', platformId: 'wecom', name: '授权状态检测', code: 'wecom.auth.check', type: '基础能力', isBasic: true, status: 'enabled' },
  { id: 'wecom-token-refresh', platformId: 'wecom', name: 'Token 自动刷新', code: 'wecom.token.refresh', type: '基础能力', isBasic: true, status: 'enabled' },
  { id: 'wecom-dept-sync', platformId: 'wecom', name: '部门同步', code: 'wecom.department.sync', type: '数据同步', isBasic: false, status: 'enabled' },
  { id: 'wecom-user-sync', platformId: 'wecom', name: '成员同步', code: 'wecom.user.sync', type: '数据同步', isBasic: false, status: 'enabled' },
  { id: 'wecom-org-sync', platformId: 'wecom', name: '组织架构同步', code: 'wecom.org.sync', type: '数据同步', isBasic: false, status: 'enabled' },
  { id: 'wecom-message-push', platformId: 'wecom', name: '消息推送', code: 'wecom.message.push', type: '消息推送', isBasic: false, status: 'enabled' },
  { id: 'wecom-event-callback', platformId: 'wecom', name: '事件回调', code: 'wecom.event.callback', type: '回调能力', isBasic: true, status: 'enabled' },
  { id: 'wecom-app-install', platformId: 'wecom', name: '应用安装授权', code: 'wecom.app.install', type: '授权能力', isBasic: true, status: 'enabled' },
  { id: 'wecom-customer-sync', platformId: 'wecom', name: '客户联系同步', code: 'wecom.customer.sync', type: '数据同步', isBasic: false, status: 'maintenance' },
  { id: 'jd-order-sync', platformId: 'jd', name: '订单同步', code: 'jd.order.sync', type: '数据同步', isBasic: false, status: 'enabled' },
  { id: 'jd-product-sync', platformId: 'jd', name: '商品同步', code: 'jd.product.sync', type: '数据同步', isBasic: false, status: 'enabled' },
  { id: 'jd-refund-sync', platformId: 'jd', name: '售后同步', code: 'jd.refund.sync', type: '数据同步', isBasic: false, status: 'enabled' },
  { id: 'douyin-order-sync', platformId: 'douyin', name: '订单同步', code: 'douyin.order.sync', type: '数据同步', isBasic: false, status: 'maintenance' },
  { id: 'taobao-trade-sync', platformId: 'taobao', name: '交易同步', code: 'taobao.trade.sync', type: '数据同步', isBasic: false, status: 'enabled' },
])

const appCapabilities = reactive([
  ...['wecom-auth-check','wecom-token-refresh','wecom-dept-sync','wecom-user-sync','wecom-org-sync','wecom-event-callback','wecom-app-install'].map((capabilityId) => ({ id: `ac-${capabilityId}`, appId: 'wecom-suite-main', capabilityId, connected: true, authStatus: 'approved', openToTenant: true, defaultEnabled: true, tenantConfigurable: !['wecom-auth-check','wecom-token-refresh','wecom-event-callback','wecom-app-install'].includes(capabilityId), health: 'normal' })),
  { id: 'ac-wecom-msg-main', appId: 'wecom-suite-main', capabilityId: 'wecom-message-push', connected: true, authStatus: 'testing', openToTenant: false, defaultEnabled: false, tenantConfigurable: true, health: 'testing' },
  { id: 'ac-wecom-customer-main', appId: 'wecom-suite-main', capabilityId: 'wecom-customer-sync', connected: false, authStatus: 'not_applied', openToTenant: false, defaultEnabled: false, tenantConfigurable: false, health: 'not_connected' },
  ...['wecom-auth-check','wecom-token-refresh','wecom-dept-sync','wecom-user-sync','wecom-org-sync'].map((capabilityId) => ({ id: `ac-test-${capabilityId}`, appId: 'wecom-suite-test', capabilityId, connected: true, authStatus: 'approved', openToTenant: true, defaultEnabled: true, tenantConfigurable: !['wecom-auth-check','wecom-token-refresh'].includes(capabilityId), health: 'normal' })),
  ...['jd-order-sync','jd-product-sync','jd-refund-sync'].map((capabilityId) => ({ id: `ac-${capabilityId}`, appId: 'jd-main', capabilityId, connected: true, authStatus: 'approved', openToTenant: true, defaultEnabled: true, tenantConfigurable: true, health: 'normal' })),
  { id: 'ac-douyin-order-sync', appId: 'douyin-main', capabilityId: 'douyin-order-sync', connected: true, authStatus: 'approved', openToTenant: true, defaultEnabled: true, tenantConfigurable: true, health: 'error' },
  { id: 'ac-taobao-trade-sync', appId: 'taobao-main', capabilityId: 'taobao-trade-sync', connected: true, authStatus: 'approved', openToTenant: true, defaultEnabled: true, tenantConfigurable: true, health: 'normal' },
])

const connections = reactive([
  { id: 'conn-wecom-001', tenantName: '杭州鹿鸣科技', platformId: 'wecom', appId: 'wecom-suite-main', authSubject: '鹿鸣科技企业微信', authStatus: 'valid', status: 'connected', finalCapabilityCount: 6, callsToday: 12680, lastSync: '2026-05-14 09:55', authScope: ['通讯录读取', '部门读取', '成员读取', '应用消息'], visibleScope: '全公司，3 个一级部门，426 名成员', credentialSummary: 'auth_corpid=wwb8***19，permanent_code 已加密', finalCapabilities: finalCaps(['部门同步','成员同步','组织架构同步','消息推送'], true) },
  { id: 'conn-wecom-002', tenantName: '上海星河贸易', platformId: 'wecom', appId: 'wecom-suite-main', authSubject: '星河贸易企业微信', authStatus: 'valid', status: 'connected', finalCapabilityCount: 5, callsToday: 8920, lastSync: '2026-05-14 09:47', authScope: ['通讯录读取', '部门读取', '成员读取'], visibleScope: '销售中心、客服中心，178 名成员', credentialSummary: 'auth_corpid=ww02***8c，permanent_code 已加密', finalCapabilities: finalCaps(['部门同步','成员同步','组织架构同步'], false) },
  { id: 'conn-wecom-003', tenantName: '深圳云仓供应链', platformId: 'wecom', appId: 'wecom-suite-main', authSubject: '云仓供应链企业微信', authStatus: 'expiring', status: 'warning', finalCapabilityCount: 4, callsToday: 3250, lastSync: '2026-05-14 08:13', authScope: ['通讯录读取'], visibleScope: '仅授权运营部，63 名成员', credentialSummary: 'auth_corpid=ww91***4a，permanent_code 已加密', finalCapabilities: finalCaps(['部门同步','成员同步'], false) },
  { id: 'conn-wecom-test-001', tenantName: '测试租户 Alpha', platformId: 'wecom', appId: 'wecom-suite-test', authSubject: 'Alpha 测试企业', authStatus: 'valid', status: 'connected', finalCapabilityCount: 5, callsToday: 820, lastSync: '2026-05-14 09:21', authScope: ['通讯录读取', '部门读取'], visibleScope: '测试部门，22 名成员', credentialSummary: 'auth_corpid=wwtest***01，permanent_code 已加密', finalCapabilities: finalCaps(['部门同步','成员同步'], false) },
  { id: 'conn-jd-001', tenantName: '杭州鹿鸣科技', platformId: 'jd', appId: 'jd-main', authSubject: '鹿鸣京东旗舰店', authStatus: 'valid', status: 'connected', finalCapabilityCount: 8, callsToday: 20450, lastSync: '2026-05-14 09:59', authScope: ['商品', '订单', '售后'], visibleScope: '店铺全量数据', credentialSummary: 'access_token / refresh_token 已加密', finalCapabilities: finalCaps(['订单同步','商品同步','售后同步'], true) },
  { id: 'conn-jd-002', tenantName: '宁波青禾家居', platformId: 'jd', appId: 'jd-backup', authSubject: '青禾家居京东店', authStatus: 'expired', status: 'error', finalCapabilityCount: 3, callsToday: 120, lastSync: '2026-05-13 22:10', authScope: ['商品', '订单'], visibleScope: '店铺全量数据', credentialSummary: 'access_token 已过期', finalCapabilities: finalCaps(['订单同步','商品同步'], false) },
  { id: 'conn-douyin-001', tenantName: '成都麦田食品', platformId: 'douyin', appId: 'douyin-main', authSubject: '麦田食品抖店', authStatus: 'valid', status: 'error', finalCapabilityCount: 6, callsToday: 16420, lastSync: '2026-05-14 08:45', authScope: ['店铺', '订单', '商品'], visibleScope: '店铺全量数据', credentialSummary: 'shop_id=883***20，access_token 已加密', finalCapabilities: finalCaps(['订单同步'], true) },
  { id: 'conn-taobao-001', tenantName: '广州拾光服饰', platformId: 'taobao', appId: 'taobao-main', authSubject: '拾光天猫旗舰店', authStatus: 'valid', status: 'connected', finalCapabilityCount: 7, callsToday: 14100, lastSync: '2026-05-14 09:52', authScope: ['交易', '商品', '退款'], visibleScope: '店铺全量数据', credentialSummary: 'session_key 已加密', finalCapabilities: finalCaps(['交易同步'], true) },
])

function finalCaps(names, includeMessage) {
  const base = [
    { name: '授权状态检测', code: 'auth.check', enabled: true, appAllowed: true, scopeAllowed: true, tenantSelected: true, reason: '基础运行能力，不允许租户关闭。' },
    { name: 'Token 自动刷新', code: 'token.refresh', enabled: true, appAllowed: true, scopeAllowed: true, tenantSelected: true, reason: '用于保持连接实例可用。' },
  ]
  names.forEach(name => base.push({ name, code: name.toLowerCase().replaceAll(' ', '.'), enabled: true, appAllowed: true, scopeAllowed: true, tenantSelected: true, reason: '应用已接通、租户授权范围允许，默认启用。' }))
  base.push({ name: '消息推送', code: 'message.push', enabled: includeMessage, appAllowed: true, scopeAllowed: includeMessage, tenantSelected: includeMessage, reason: includeMessage ? '租户授权并选择启用。' : '当前未开放或租户未授权消息范围。' })
  return base
}

const syncJobs = reactive([
  { id: 'sync-001', connectionId: 'conn-wecom-001', platformId: 'wecom', job: '鹿鸣企业微信成员增量同步', tenant: '杭州鹿鸣科技', capability: '成员同步', mode: 'incremental', cron: '每 10 分钟', status: 'success', successRate: 99, lastRun: '09:55', nextRun: '10:05' },
  { id: 'sync-002', connectionId: 'conn-wecom-002', platformId: 'wecom', job: '星河部门全量同步', tenant: '上海星河贸易', capability: '部门同步', mode: 'full', cron: '每日 02:00', status: 'success', successRate: 98, lastRun: '02:00', nextRun: '明日 02:00' },
  { id: 'sync-003', connectionId: 'conn-jd-002', platformId: 'jd', job: '青禾京东订单同步', tenant: '宁波青禾家居', capability: '订单同步', mode: 'incremental', cron: '每 5 分钟', status: 'failed', successRate: 42, lastRun: '22:10', nextRun: '暂停' },
  { id: 'sync-004', connectionId: 'conn-douyin-001', platformId: 'douyin', job: '麦田抖店订单同步', tenant: '成都麦田食品', capability: '订单同步', mode: 'incremental', cron: '每 5 分钟', status: 'running', successRate: 76, lastRun: '09:58', nextRun: '执行中' },
  { id: 'sync-005', connectionId: 'conn-taobao-001', platformId: 'taobao', job: '拾光天猫交易同步', tenant: '广州拾光服饰', capability: '交易同步', mode: 'incremental', cron: '每 10 分钟', status: 'success', successRate: 99, lastRun: '09:52', nextRun: '10:02' },
])

const policies = reactive([
  { id: 'policy-platform-wecom', platformId: 'wecom', name: '企业微信平台默认策略', scope: 'platform', scopeLabel: '平台级', dailyLimit: 2000000, monthlyLimit: 60000000, qpsLimit: 500, concurrentLimit: 200, exceedStrategy: '熔断平台', status: 'enabled', priority: 100, isOverride: false },
  { id: 'policy-app-wecom-main', platformId: 'wecom', name: '企微标准应用默认策略', scope: 'provider_app', scopeLabel: '应用级', dailyLimit: 1000000, monthlyLimit: 30000000, qpsLimit: 100, concurrentLimit: 50, exceedStrategy: '排队', status: 'enabled', priority: 200, isOverride: false },
  { id: 'policy-tenant-standard', platformId: 'wecom', name: '企业微信租户通用策略', scope: 'tenant', scopeLabel: '租户级', dailyLimit: 100000, monthlyLimit: 3000000, qpsLimit: 20, concurrentLimit: 5, exceedStrategy: '降频', status: 'enabled', priority: 300, isOverride: false },
  { id: 'policy-tenant-vip-luming', platformId: 'wecom', name: '鹿鸣科技企微专属覆盖', scope: 'tenant', scopeLabel: '租户级', dailyLimit: 300000, monthlyLimit: 9000000, qpsLimit: 50, concurrentLimit: 10, exceedStrategy: '排队', status: 'enabled', priority: 500, isOverride: true },
  { id: 'policy-cap-wecom-user', platformId: 'wecom', name: '成员同步能力策略', scope: 'capability', scopeLabel: '能力级', dailyLimit: 20000, monthlyLimit: 600000, qpsLimit: 10, concurrentLimit: 2, exceedStrategy: '暂停能力', status: 'enabled', priority: 350, isOverride: false },
  { id: 'policy-jd-tenant', platformId: 'jd', name: '京东租户通用策略', scope: 'tenant', scopeLabel: '租户级', dailyLimit: 150000, monthlyLimit: 4500000, qpsLimit: 30, concurrentLimit: 6, exceedStrategy: '排队', status: 'enabled', priority: 300, isOverride: false },
])

const quotaUsages = reactive([
  { id: 'usage-001', policyId: 'policy-tenant-vip-luming', connectionId: 'conn-wecom-001', name: '杭州鹿鸣科技', scope: '租户级', objectPath: '企业微信 / 标准应用 / 鹿鸣科技企业微信', used: 126800, limit: 300000, rate: 42, override: true },
  { id: 'usage-002', policyId: 'policy-tenant-standard', connectionId: 'conn-wecom-002', name: '上海星河贸易', scope: '租户级', objectPath: '企业微信 / 标准应用 / 星河贸易企业微信', used: 92000, limit: 100000, rate: 92, override: false },
  { id: 'usage-003', policyId: 'policy-cap-wecom-user', connectionId: 'conn-wecom-002', name: '星河贸易成员同步', scope: '能力级', objectPath: '连接实例 / 成员同步', used: 19200, limit: 20000, rate: 96, override: false },
  { id: 'usage-004', policyId: 'policy-jd-tenant', connectionId: 'conn-jd-001', name: '鹿鸣京东旗舰店', scope: '连接实例级', objectPath: '京东 / 正式应用 A / 京东旗舰店', used: 121000, limit: 150000, rate: 81, override: false },
  { id: 'usage-005', policyId: 'policy-platform-wecom', connectionId: null, name: '企业微信平台总量', scope: '平台级', objectPath: '企业微信 / 全部应用 / 全部租户', used: 326800, limit: 2000000, rate: 16, override: false },
])

const alerts = reactive([
  { id: 'alert-001', level: 'critical', status: 'open', title: '京东连接授权已过期', message: '宁波青禾家居京东店 access_token 已过期，订单同步失败。', path: '京东电商 / 京东备用应用 B / 青禾京东店', lastAt: '09:38', count: 12 },
  { id: 'alert-002', level: 'warning', status: 'open', title: '星河贸易成员同步接近配额', message: '成员同步能力今日使用率达到 96%，建议降频或配置专属覆盖。', path: '企业微信 / 标准应用 / 星河贸易 / 成员同步', lastAt: '09:45', count: 4 },
  { id: 'alert-003', level: 'critical', status: 'processing', title: '抖音订单同步接口异常', message: '第三方接口连续返回 5xx，已自动进入排队重试。', path: '抖音电商 / 正式应用 / 麦田食品抖店', lastAt: '09:51', count: 23 },
  { id: 'alert-004', level: 'warning', status: 'open', title: '企业微信 suite_ticket 推送延迟', message: '测试应用最近一次 suite_ticket 推送超过预期时间。', path: '企业微信 / 沙箱测试应用', lastAt: '09:31', count: 2 },
  { id: 'alert-005', level: 'info', status: 'resolved', title: '淘宝交易同步恢复', message: '拾光天猫交易同步失败率已恢复到正常范围。', path: '淘宝天猫 / 正式应用 / 拾光天猫旗舰店', lastAt: '08:58', count: 6 },
])

const logs = reactive([
  { id: 'log-001', requestId: 'req_202605140955_wecom_001', type: 'api', typeLabel: '第三方 API', platformId: 'wecom', tenant: '杭州鹿鸣科技', method: 'GET', path: '/cgi-bin/user/list', endpoint: 'api.weixin.qq.com', success: true, status: 200, cost: 382, calledAt: '2026-05-14 09:55:20' },
  { id: 'log-002', requestId: 'req_202605140952_jd_113', type: 'api', typeLabel: '第三方 API', platformId: 'jd', tenant: '宁波青禾家居', method: 'POST', path: '/routerjson?method=360buy.order.search', endpoint: 'api.jd.com', success: false, status: 401, cost: 211, calledAt: '2026-05-14 09:52:11' },
  { id: 'log-003', requestId: 'req_202605140942_wecom_ticket', type: 'callback', typeLabel: '回调接收', platformId: 'wecom', tenant: '-', method: 'POST', path: '/callback/wecom/command', endpoint: 'api.ec-aios.com', success: true, status: 200, cost: 62, calledAt: '2026-05-14 09:42:03' },
  { id: 'log-004', requestId: 'req_202605140945_token_002', type: 'token', typeLabel: 'Token 刷新', platformId: 'wecom', tenant: '上海星河贸易', method: 'POST', path: '/cgi-bin/service/get_corp_token', endpoint: 'api.weixin.qq.com', success: true, status: 200, cost: 301, calledAt: '2026-05-14 09:45:18' },
  { id: 'log-005', requestId: 'req_202605140951_data_883', type: 'data_write', typeLabel: '数据写入', platformId: 'douyin', tenant: '成都麦田食品', method: 'POST', path: '/data-center/write/orders', endpoint: 'data.ec-aios.com', success: false, status: 500, cost: 780, calledAt: '2026-05-14 09:51:41' },
])

const currentNav = computed(() => navItems.find(n => n.key === page.value) || navItems[0])
const overviewMetrics = computed(() => [
  { label: '接入平台', value: platforms.length, trend: 1, trendText: '+1 本月' },
  { label: '服务商应用', value: apps.length, trend: 1, trendText: '+2 本周' },
  { label: '平台能力', value: capabilities.length, trend: 1, trendText: '+5 本周' },
  { label: '租户连接', value: connections.length, trend: 1, trendText: '+18 本月' },
  { label: '今日调用', value: formatNumber(platforms.reduce((s, p) => s + p.callsToday, 0)), trend: 1, trendText: '+12.6%' },
  { label: '待处理异常', value: alerts.filter(a => a.status !== 'resolved').length, trend: -1, trendText: '-3 较昨日' },
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
  },
)

onMounted(() => {
  loadBackendSnapshots()
})

function switchPage(key) { page.value = key }
function primaryAction() {
  if (page.value === 'platforms') return openPlatformModal()
  if (page.value === 'workspace') return openAppModal()
  if (page.value === 'tenantConnections') return notify('租户连接列表已刷新')
  if (page.value === 'quota') return openPolicyModal()
  if (page.value === 'logs') return notify('调用日志已生成导出任务')
  if (page.value === 'alerts') return notify('已批量标记待处理异常为处理中')
  notify('已刷新当前页面数据')
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
function notify(message, type = 'success') { toast.message = message; toast.type = type; toast.show = true; setTimeout(() => { toast.show = false }, 2200) }
function closeDrawer() { drawer.open = false; drawer.type = ''; drawer.data = null }
function openDrawer(type, title, subtitle, desc, data, tab = 'basic') { drawer.type = type; drawer.title = title; drawer.subtitle = subtitle; drawer.desc = desc; drawer.data = data; drawer.open = true; drawerTab.value = tab }
function openPlatformDrawer(platform) { openDrawer('platform', platform.name, '接入平台详情', '平台是聚合根，应用、能力、连接实例都围绕平台展开。', platform) }
function openAppDrawer(app, tab = 'basic') { openDrawer('app', app.name, '服务商应用详情', '应用承载密钥、回调、应用能力连接与授权租户。', app, tab) }
function openConnectionDrawer(conn) { openDrawer('connection', `${conn.tenantName} · ${conn.authSubject}`, '租户连接实例详情', '租户通过应用授权后生成的运行实例。', conn) }
function openCapabilityDrawer(row) { openDrawer('capability', row.name, '能力详情', '能力定义、应用接通情况与租户开放策略。', row) }
function openPolicyDrawer(policy) { openDrawer('policy', policy.name, '配额策略详情', '查看策略范围、优先级、超限动作与绑定情况。', policy) }
function openBindingDrawer(policy) { openDrawer('policy', policy.name, '策略绑定对象', '策略可绑定平台、应用、租户、连接实例或能力。', { policy, bindings: selectedPolicyUsage.value }) }
function openUsageDrawer(row) { openDrawer('usage', row.name, '配额用量明细', '展示用量、来源、覆盖关系与超限记录。', row) }
function openAlertDrawer(row) { openDrawer('alert', row.title, '异常详情', '异常聚合、影响对象、处理动作与恢复状态。', row) }
function openLogDrawer(row) { openDrawer('log', row.requestId, '调用日志详情', '请求、响应、耗时、错误码与调用链路。', row) }
function openSyncDrawer(row) { openDrawer('sync', row.job, '同步任务日志', '查看任务执行记录、失败原因与重试链路。', row) }
async function loadBackendSnapshots() {
  backendState.loading = true
  backendState.error = ''
  try {
    const [platformRes, appRes, connectionRes, syncRes, quotaRes, alertRes, logRes] = await Promise.allSettled([
      fetchIntegrationPlatforms(),
      fetchIntegrationWorkspace(),
      fetchIntegrationTenantConnections(),
      fetchIntegrationSyncMonitor(),
      fetchIntegrationQuota(),
      fetchIntegrationAlerts(),
      fetchIntegrationLogs(),
    ])
    applyBackendSection(platformRes, (items) => replaceRows(platforms, items.map(mapBackendPlatform)))
    applyBackendSection(appRes, (items) => replaceRows(apps, items.map(mapBackendApp)))
    applyBackendSection(connectionRes, (items) => replaceRows(connections, items.map(mapBackendConnection)))
    applyBackendSection(syncRes, (items) => replaceRows(syncJobs, items.map(mapBackendSyncJob)))
    applyBackendSection(quotaRes, (items) => replaceRows(policies, items.map(mapBackendPolicy)))
    applyBackendSection(alertRes, (items) => replaceRows(alerts, items.map(mapBackendAlert)))
    applyBackendSection(logRes, (items) => replaceRows(logs, items.map(mapBackendLog)))
    backendState.loaded = true
  } catch (error) {
    backendState.error = error instanceof Error ? error.message : '后端数据加载失败'
  } finally {
    backendState.loading = false
  }
}

function applyBackendSection(result, apply) {
  if (result.status !== 'fulfilled') return
  const items = Array.isArray(result.value?.items) ? result.value.items : []
  if (items.length) apply(items)
}

function replaceRows(target, rows) {
  target.splice(0, target.length, ...rows)
}

function mapBackendPlatform(row) {
  const code = row.code || row.platform_code || row.PlatformCode || String(row.ID || row.id || '')
  return {
    id: code,
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
    successRate: Number(row.success_rate || row.SuccessRate || (Number(row.open_alert_count || row.OpenAlertCount || 0) > 0 ? 88 : 99)),
    description: row.description || row.Description || '由集成中心数据库返回的平台档案。',
  }
}

function mapBackendApp(row) {
  const id = row.app_code || row.AppCode || String(row.id || row.ID || '')
  const platformId = findPlatformId(row.platform_id || row.PlatformID, row.platform_name || row.PlatformName)
  return {
    id,
    platformId,
    name: row.app_name || row.AppName || id,
    type: '服务商应用',
    env: row.environment || row.Environment || 'prod',
    status: mapStatus(row.status || row.Status),
    capabilityCount: Number(row.capability_count || row.CapabilityCount || 0),
    connectionCount: Number(row.connection_count || row.ConnectionCount || 0),
    alertCount: Number(row.alert_count || row.AlertCount || 0),
    callsToday: Number(row.calls_today || row.CallsToday || 0),
    endpoint: row.endpoint || row.Endpoint || '',
    authMode: row.auth_mode || row.AuthMode || '-',
  }
}

function mapBackendConnection(row) {
  const platformId = findPlatformId(null, row.platform_name || row.PlatformName)
  const appId = findAppId(row.provider_app_name || row.ProviderAppName, platformId)
  return {
    id: String(row.id || row.ID || row.auth_subject_id || row.AuthSubjectID),
    tenantName: row.tenant_name || row.TenantName || `租户 ${row.tenant_id || row.TenantID || '-'}`,
    platformId,
    appId,
    authSubject: row.auth_subject_name || row.AuthSubjectName || '-',
    authStatus: mapAuthStatus(row.auth_status || row.AuthStatus),
    status: mapConnectionStatus(row.connection_status || row.ConnectionStatus),
    finalCapabilityCount: Number(row.final_capability_count || row.FinalCapabilityCount || 0),
    callsToday: Number(row.calls_today || row.CallsToday || 0),
    lastSync: formatBackendTime(row.last_sync_at || row.LastSyncAt),
    authScope: [],
    visibleScope: row.auth_subject_type || row.AuthSubjectType || '-',
    credentialSummary: '凭证由后端密钥引用托管',
    finalCapabilities: finalCaps([], false),
  }
}

function mapBackendSyncJob(row) {
  const total = Number(row.total_count || row.TotalCount || 0)
  const success = Number(row.success_count || row.SuccessCount || 0)
  return {
    id: String(row.id || row.ID),
    connectionId: String(row.tenant_connection_id || row.TenantConnectionID || ''),
    platformId: 'all',
    job: row.job_type || row.JobType || '同步任务',
    tenant: `租户 ${row.tenant_id || row.TenantID || '-'}`,
    capability: row.capability_code || row.CapabilityCode || '-',
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
    platformId: 'all',
    name: row.policy_name || row.PolicyName || '-',
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

function mapBackendAlert(row) {
  return {
    id: String(row.id || row.ID),
    level: row.severity || row.Severity || 'warning',
    title: row.title || row.Title || '-',
    message: row.message || row.Message || '',
    path: row.alert_type || row.AlertType || '-',
    status: mapAlertStatus(row.status || row.Status),
    count: 1,
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
    platformId: 'all',
    tenant: row.tenant_id || row.TenantID ? `租户 ${row.tenant_id || row.TenantID}` : '-',
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
  const byIndex = platforms.find(p => String(p.sortWeight) === String(platformID) || String(p.id) === String(platformID))
  return byIndex?.id || 'all'
}

function findAppId(appName, platformId) {
  const app = apps.find(a => a.name === appName && (platformId === 'all' || a.platformId === platformId))
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
function openPlatformModal(platform) {
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
  const name = String(form.name || '').trim()
  const code = String(form.code || '').trim()
  if (!name || !code) {
    notify('请填写平台名称与平台编码', 'error')
    return
  }
  const existing = form.id && platforms.find(p => p.id === form.id)
  const shortName = String(form.shortName || '').trim() || name
  const slug = code.replace(/[^a-z0-9_-]/gi, '_').toLowerCase()
  const icon = String(form.icon || '').trim() || shortName.slice(0, 1) || '?'
  const patch = {
    name,
    shortName,
    code,
    type: form.type,
    accessType: form.accessType,
    icon,
    status: form.status,
    tenantVisible: form.tenantVisible !== false,
    owner: String(form.owner || '').trim() || '未指定',
    officialUrl: String(form.officialUrl || '').trim(),
    sortWeight: Number(form.sortWeight) || 100,
    description: String(form.description || ''),
  }
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
  if (backendState.loaded || !backendState.error) {
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
      closeModal()
      notify(existing ? '平台档案已更新' : '平台档案已创建')
      if (goWorkspace) jumpWorkspace(savedId, 'apps')
      return
    } catch (error) {
      backendState.error = error instanceof Error ? error.message : '接入平台保存失败'
      notify('后端保存失败，已保留本地编辑结果', 'error')
    }
  }
  if (existing) {
    Object.assign(existing, patch)
    savedId = existing.id
  } else {
    let id = slug
    if (platforms.some(p => p.id === id)) id = `pl-${Date.now()}`
    savedId = id
    platforms.push({
      id: savedId,
      ...patch,
      appCount: 0,
      capabilityCount: 0,
      connectionCount: 0,
      alertCount: 0,
      callsToday: 0,
      successRate: 0,
    })
  }
  closeModal()
  notify(existing ? '平台档案已更新' : '平台档案已创建')
  if (goWorkspace) jumpWorkspace(savedId, 'apps')
}
function openAppModal(app) { openModal('app', app ? '编辑服务商应用' : '新增服务商应用', '服务商应用', app || { name: '', platformId: workspacePlatformId.value === 'all' ? 'wecom' : workspacePlatformId.value, type: '正式应用', env: '正式', endpoint: '' }) }
function openCapabilityModal(cap) { openModal('capability', cap ? '编辑平台能力' : '新增平台能力', '平台能力', cap || {}) }
function openPolicyModal(policy) { openModal('policy', policy ? '编辑配额策略' : '新增配额策略', '配额与限流', policy || { name: '', scope: 'tenant', dailyLimit: 100000, monthlyLimit: 3000000, qpsLimit: 20, concurrentLimit: 5, exceedStrategy: '告警', isOverride: false }) }
function openOverrideModal(row) { openModal('override', '配置专属覆盖', '特殊企业配额覆盖', { name: row.name, limit: row.limit || 300000, qps: 50, concurrent: 10, period: '长期有效', remark: '大客户专属提额' }) }
function openModal(type, title, subtitle, data) { Object.keys(form).forEach(k => delete form[k]); Object.assign(form, JSON.parse(JSON.stringify(data || {}))); modal.type = type; modal.title = title; modal.subtitle = subtitle; modal.open = true }
function closeModal() { modal.open = false }
function saveModal() { notify(`${modal.title}已保存`); closeModal() }
function toggleAppCapability(row, key) { const target = appCapabilities.find(ac => ac.id === row.id); if (target) target[key] = !target[key]; notify(`${row.name}：${key} 已更新`) }
function runHealthCheck(target) { notify(`${target}连通性检测已完成：存在 1 条预警`) }
function refreshConnection(row) { row.authStatus = 'valid'; row.status = 'connected'; notify(`${row.tenantName} 授权状态已刷新`) }
function pauseConnection(row) { row.status = row.status === 'paused' ? 'connected' : 'paused'; notify(`${row.tenantName} 连接已${row.status === 'paused' ? '暂停' : '恢复'}`) }
function retryConnection(row) { notify(`${row.tenantName} 同步任务已加入重试队列`) }
function retrySync(row) { row.status = 'running'; notify(`${row.job} 已开始重试`) }
function toggleSync(row) { row.status = row.status === 'paused' ? 'running' : 'paused'; notify(`${row.job} 已${row.status === 'paused' ? '暂停' : '恢复'}`) }
function copyPolicy(policy) { const copy = { ...policy, id: `${policy.id}-copy-${Date.now()}`, name: `${policy.name} - 专属覆盖`, isOverride: true, priority: policy.priority + 200, dailyLimit: Math.round(policy.dailyLimit * 2) }; policies.push(copy); selectedPolicyId.value = copy.id; notify('已复制为专属覆盖策略') }
function toggleStatus(obj) { obj.status = obj.status === 'enabled' ? 'disabled' : 'enabled'; notify(`${obj.name || obj.title} 已${obj.status === 'enabled' ? '启用' : '停用'}`) }
function processAlert(alert) { alert.status = 'processing'; notify('异常已标记处理中') }
function resolveAlert(alert) { alert.status = 'resolved'; notify('异常已标记恢复') }
function ignoreAlert(alert) { alert.status = 'ignored'; notify('异常已忽略') }
function copyText(text) { navigator?.clipboard?.writeText(text); notify('已复制') }
function appSecretItems(app) { return [{ label: 'suite_id / app_key', value: app.suiteId || 'app_key_******' }, { label: 'suite_secret / app_secret', value: app.suiteSecret || '****** 加密存储' }, { label: 'Token', value: app.token || '******' }, { label: 'EncodingAESKey', value: app.encodingAesKey || '******' }, { label: 'Endpoint', value: app.endpoint || '-' }] }

const StatusBadge = defineComponent({
  props: { status: { type: [String, Boolean], default: 'normal' } },
  setup(props) {
    const labels = { enabled: '启用', disabled: '停用', draft: '草稿', maintenance: '维护中', connected: '正常', normal: '正常', success: '成功', valid: '有效', approved: '已授权', warning: '预警', testing: '测试中', expiring: '即将过期', running: '执行中', open: '待处理', processing: '处理中', error: '异常', failed: '失败', expired: '已过期', exceeded: '超限', not_connected: '未接通', not_applied: '未申请', paused: '暂停', ignored: '已忽略', resolved: '已恢复' }
    return () => h('span', { class: ['status-badge', props.status] }, labels[props.status] || props.status)
  }
})

const SwitchToggle = defineComponent({
  props: { modelValue: Boolean },
  emits: ['update:modelValue'],
  setup(props, { emit }) { return () => h('button', { class: ['switch', props.modelValue ? 'on' : ''], onClick: () => emit('update:modelValue', !props.modelValue) }, [h('i')]) }
})

const DataTable = defineComponent({
  props: { columns: Array, rows: Array, minWidth: { type: String, default: '960px' } },
  setup(props, { slots }) {
    return () => h('div', { class: 'data-table-wrap' }, [h('table', { style: { minWidth: props.minWidth } }, [
      h('thead', [h('tr', props.columns.map(c => h('th', c.label)))]),
      h('tbody', props.rows.map(row => h('tr', { key: row.id }, props.columns.map(c => h('td', slots[`cell-${c.key}`] ? slots[`cell-${c.key}`]({ row }) : row[c.key])))))
    ])])
  }
})

const RankingList = defineComponent({
  props: { items: Array },
  setup(props) { return () => h('div', { class: 'ranking-list' }, props.items.map((item, index) => h('div', { class: 'ranking-row', key: item.name }, [h('span', { class: 'rank-index' }, index + 1), h('div', [h('strong', item.name), h('p', item.desc)]), h('b', item.value)]))) }
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
  props: { rows: Array },
  emits: ['open', 'refresh', 'pause', 'retry'],
  setup(props, { emit }) {
    const cols = [
      { key: 'tenant', label: '租户 / 授权主体' }, { key: 'platform', label: '平台' }, { key: 'app', label: '服务商应用' }, { key: 'auth', label: '授权状态' }, { key: 'status', label: '连接状态' }, { key: 'caps', label: '最终能力' }, { key: 'calls', label: '今日调用' }, { key: 'last', label: '最近同步' }, { key: 'actions', label: '操作' }
    ]
    return () => h(DataTable, { columns: cols, rows: props.rows, minWidth: '1260px' }, {
      'cell-tenant': ({ row }) => h('div', [h('strong', row.tenantName), h('p', { class: 'muted' }, row.authSubject)]),
      'cell-platform': ({ row }) => getPlatformName(row.platformId),
      'cell-app': ({ row }) => getAppName(row.appId),
      'cell-auth': ({ row }) => h(StatusBadge, { status: row.authStatus }),
      'cell-status': ({ row }) => h(StatusBadge, { status: row.status }),
      'cell-caps': ({ row }) => `${row.finalCapabilityCount} 项`,
      'cell-calls': ({ row }) => formatNumber(row.callsToday),
      'cell-last': ({ row }) => row.lastSync,
      'cell-actions': ({ row }) => h('div', { class: 'actions wide' }, [h('button', { onClick: () => emit('open', row) }, '详情'), h('button', { onClick: () => emit('refresh', row) }, '刷新授权'), h('button', { onClick: () => emit('pause', row) }, row.status === 'paused' ? '恢复' : '暂停'), h('button', { onClick: () => emit('retry', row) }, '重试')])
    })
  }
})
</script>

<style scoped src="./integrationCenter.css"></style>
