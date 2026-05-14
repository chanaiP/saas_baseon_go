package services

import (
	"context"
	"fmt"
	"time"

	"saas_baseon_go/internal/apps/integration_center/repositories"
)

type Service struct {
	repo *repositories.Repository
}

func NewService(repo *repositories.Repository) *Service {
	return &Service{repo: repo}
}

type Metric struct {
	Label string `json:"label"`
	Value string `json:"value"`
	Trend string `json:"trend"`
	Tone  string `json:"tone"`
}

type IntegrationConnector struct {
	Code        string   `json:"code"`
	Name        string   `json:"name"`
	Vendor      string   `json:"vendor"`
	Category    string   `json:"category"`
	Status      string   `json:"status"`
	AuthMode    string   `json:"auth_mode"`
	DeployMode  string   `json:"deploy_mode"`
	Channels    []string `json:"channels"`
	LastSyncAt  string   `json:"last_sync_at"`
	HealthScore int      `json:"health_score"`
	Description string   `json:"description"`
}

type IntegrationEvent struct {
	ID        string `json:"id"`
	Source    string `json:"source"`
	EventType string `json:"event_type"`
	Status    string `json:"status"`
	LatencyMS int    `json:"latency_ms"`
	Occurred  string `json:"occurred"`
}

type IntegrationChecklistItem struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
}

type Overview struct {
	Metrics    []Metric                   `json:"metrics"`
	Connectors []IntegrationConnector     `json:"connectors"`
	Events     []IntegrationEvent         `json:"events"`
	Checklist  []IntegrationChecklistItem `json:"checklist"`
	Channels   []map[string]string        `json:"channels"`
}

type SectionSummary struct {
	AppCode string      `json:"app_code"`
	Section string      `json:"section"`
	Items   interface{} `json:"items"`
}

func (s *Service) Overview(ctx context.Context) (Overview, error) {
	if s.repo != nil {
		counts, err := s.repo.Counts(ctx)
		if err != nil {
			return Overview{}, err
		}
		connectors, err := s.Connectors(ctx)
		if err != nil {
			return Overview{}, err
		}
		alerts, err := s.repo.ListAlerts(ctx, 3)
		if err != nil {
			return Overview{}, err
		}
		events, err := s.eventsFromLogs(ctx)
		if err != nil {
			return Overview{}, err
		}
		checklist := []IntegrationChecklistItem{
			{Title: "应用 Manifest", Description: "菜单、权限、API、套餐功能点与配额均由 integration-center Manifest 声明。", Status: "done"},
			{Title: "业务表初始化", Description: "平台、能力、服务商应用、连接实例、同步任务、配额与日志已具备持久化结构。", Status: "done"},
			{Title: "待处理异常", Description: fmt.Sprintf("当前还有 %d 个未恢复异常需要平台侧处理。", counts.OpenAlerts), Status: statusFromCount(counts.OpenAlerts)},
		}
		if len(alerts) == 0 {
			checklist = append(checklist, IntegrationChecklistItem{Title: "异常监控", Description: "当前没有待处理集成异常。", Status: "done"})
		}
		return Overview{
			Metrics: []Metric{
				{Label: "接入平台", Value: fmt.Sprintf("%d", counts.Platforms), Trend: fmt.Sprintf("%d 个服务商应用", counts.ProviderApps), Tone: "primary"},
				{Label: "租户连接", Value: fmt.Sprintf("%d", counts.Connections), Trend: "按授权主体独立计数", Tone: "success"},
				{Label: "今日调用", Value: fmt.Sprintf("%d", counts.TodayAPICalls), Trend: "含 API / Token / Webhook", Tone: "info"},
				{Label: "待处理异常", Value: fmt.Sprintf("%d", counts.OpenAlerts), Trend: fmt.Sprintf("%d 个同步任务运行中", counts.RunningSyncJobs), Tone: alertTone(counts.OpenAlerts)},
			},
			Connectors: connectors,
			Events:     events,
			Checklist:  checklist,
			Channels: []map[string]string{
				{"name": "平台 API", "description": "外部系统调用底座开放 API，使用应用凭证、scope 和租户上下文。"},
				{"name": "Webhook", "description": "第三方事件进入底座，强制签名、时间戳和幂等键。"},
				{"name": "数据同步", "description": "批量或增量同步组织、用户、订单等外部数据，记录游标和失败补偿。"},
				{"name": "网关代理", "description": "流量经底座鉴权、限流、审计后转发到目标系统。"},
			},
		}, nil
	}
	connectors, err := s.Connectors(ctx)
	if err != nil {
		return Overview{}, err
	}
	now := time.Now()
	return Overview{
		Metrics: []Metric{
			{Label: "连接器", Value: "4", Trend: "2 个可上线", Tone: "primary"},
			{Label: "开放 API", Value: "18", Trend: "统一鉴权与审计", Tone: "success"},
			{Label: "Webhook", Value: "9", Trend: "平均延迟 132ms", Tone: "warning"},
			{Label: "同步任务", Value: "6", Trend: "5 个健康运行", Tone: "info"},
		},
		Connectors: connectors,
		Events: []IntegrationEvent{
			{ID: "evt_10086", Source: "企业微信", EventType: "contact.changed", Status: "success", LatencyMS: 96, Occurred: now.Add(-18 * time.Minute).Format(time.RFC3339)},
			{ID: "evt_10087", Source: "钉钉", EventType: "approval.finished", Status: "success", LatencyMS: 128, Occurred: now.Add(-41 * time.Minute).Format(time.RFC3339)},
			{ID: "evt_10088", Source: "飞书", EventType: "message.card.action", Status: "retrying", LatencyMS: 310, Occurred: now.Add(-74 * time.Minute).Format(time.RFC3339)},
		},
		Checklist: []IntegrationChecklistItem{
			{Title: "应用主档", Description: "合并部署应用已在应用中心登记，后续通过 Manifest 托管菜单、权限与 API。", Status: "done"},
			{Title: "连接器凭证", Description: "仅保存凭证引用和轮换状态，禁止在 Manifest 中写入真实密钥。", Status: "active"},
			{Title: "租户授权", Description: "按租户开通连接器能力，运行时校验套餐、角色权限与配额。", Status: "planned"},
			{Title: "事件审计", Description: "Webhook、同步任务、OAuth 授权变更统一写入操作日志。", Status: "planned"},
		},
		Channels: []map[string]string{
			{"name": "平台 API", "description": "外部系统调用底座开放 API，使用应用凭证、scope 和租户上下文。"},
			{"name": "Webhook", "description": "第三方事件进入底座，强制签名、时间戳和幂等键。"},
			{"name": "数据同步", "description": "批量或增量同步组织、用户、订单等外部数据，记录游标和失败补偿。"},
			{"name": "网关代理", "description": "流量经底座鉴权、限流、审计后转发到目标系统。"},
		},
	}, nil
}

func (s *Service) Connectors(ctx context.Context) ([]IntegrationConnector, error) {
	if s.repo != nil {
		rows, err := s.repo.ListPlatforms(ctx, 20)
		if err != nil {
			return nil, err
		}
		connectors := make([]IntegrationConnector, 0, len(rows))
		for _, row := range rows {
			name := row.Name
			if row.ShortName != nil && *row.ShortName != "" {
				name = *row.ShortName
			}
			connectors = append(connectors, IntegrationConnector{
				Code:        row.Code,
				Name:        name,
				Vendor:      row.Name,
				Category:    row.PlatformType,
				Status:      row.Status,
				AuthMode:    row.AccessMode,
				DeployMode:  "MERGED",
				Channels:    channelsForAccessMode(row.AccessMode),
				LastSyncAt:  "",
				HealthScore: healthScoreForStatus(row.Status, row.OpenAlertCount),
				Description: fmt.Sprintf("%d 个服务商应用，%d 项平台能力，%d 条租户连接。", row.AppCount, row.CapabilityCount, row.ConnectionCount),
			})
		}
		return connectors, nil
	}
	return []IntegrationConnector{
		{
			Code:        "wechat-work",
			Name:        "企业微信",
			Vendor:      "Tencent",
			Category:    "协同办公",
			Status:      "online",
			AuthMode:    "OAuth + 通讯录密钥",
			DeployMode:  "MERGED",
			Channels:    []string{"Webhook", "数据同步"},
			LastSyncAt:  time.Now().Add(-22 * time.Minute).Format(time.RFC3339),
			HealthScore: 96,
			Description: "同步组织、用户与消息卡片事件，支持租户级授权和回调审计。",
		},
		{
			Code:        "dingtalk",
			Name:        "钉钉",
			Vendor:      "Alibaba",
			Category:    "协同办公",
			Status:      "beta",
			AuthMode:    "OAuth",
			DeployMode:  "MERGED",
			Channels:    []string{"Webhook", "平台 API"},
			LastSyncAt:  time.Now().Add(-55 * time.Minute).Format(time.RFC3339),
			HealthScore: 91,
			Description: "接入审批、组织和待办事件，统一纳入租户开通与权限治理。",
		},
		{
			Code:        "feishu",
			Name:        "飞书",
			Vendor:      "ByteDance",
			Category:    "协同办公",
			Status:      "developing",
			AuthMode:    "OAuth + 事件订阅",
			DeployMode:  "MERGED",
			Channels:    []string{"Webhook"},
			LastSyncAt:  time.Now().Add(-3 * time.Hour).Format(time.RFC3339),
			HealthScore: 82,
			Description: "用于消息卡片、通讯录和审批事件接入，当前处于联调阶段。",
		},
		{
			Code:        "generic-openapi",
			Name:        "通用 OpenAPI",
			Vendor:      "Platform",
			Category:    "开放接口",
			Status:      "planned",
			AuthMode:    "API Key + HMAC",
			DeployMode:  "MERGED",
			Channels:    []string{"平台 API", "网关代理"},
			LastSyncAt:  "",
			HealthScore: 0,
			Description: "面向未标准化第三方系统的通用接入模板，声明 scope、签名和限流策略。",
		},
	}, nil
}

func (s *Service) Platforms(ctx context.Context) (SectionSummary, error) {
	if s.repo != nil {
		rows, err := s.repo.ListPlatforms(ctx, 50)
		if err != nil {
			return SectionSummary{}, err
		}
		return SectionSummary{AppCode: "integration-center", Section: "platforms", Items: rows}, nil
	}
	connectors, err := s.Connectors(ctx)
	if err != nil {
		return SectionSummary{}, err
	}
	return SectionSummary{AppCode: "integration-center", Section: "platforms", Items: connectors}, nil
}

func (s *Service) Workspace(ctx context.Context) (SectionSummary, error) {
	if s.repo != nil {
		rows, err := s.repo.ListProviderApps(ctx, 50)
		if err != nil {
			return SectionSummary{}, err
		}
		return SectionSummary{AppCode: "integration-center", Section: "workspace", Items: rows}, nil
	}
	connectors, err := s.Connectors(ctx)
	if err != nil {
		return SectionSummary{}, err
	}
	return SectionSummary{AppCode: "integration-center", Section: "workspace", Items: connectors}, nil
}

func (s *Service) TenantConnections(ctx context.Context) (SectionSummary, error) {
	if s.repo != nil {
		rows, err := s.repo.ListTenantConnections(ctx, 50)
		if err != nil {
			return SectionSummary{}, err
		}
		return SectionSummary{AppCode: "integration-center", Section: "tenant-connections", Items: rows}, nil
	}
	now := time.Now()
	return SectionSummary{AppCode: "integration-center", Section: "tenant-connections", Items: []map[string]interface{}{
		{"tenant_name": "杭州鹿鸣科技", "platform": "企业微信", "provider_app": "企业微信第三方标准应用", "auth_subject_type": "corp", "auth_subject_name": "鹿鸣科技企业微信", "status": "connected", "last_sync_at": now.Add(-5 * time.Minute).Format(time.RFC3339)},
		{"tenant_name": "上海星河贸易", "platform": "企业微信", "provider_app": "企业微信第三方标准应用", "auth_subject_type": "corp", "auth_subject_name": "星河贸易企业微信", "status": "warning", "last_sync_at": now.Add(-13 * time.Minute).Format(time.RFC3339)},
	}}, nil
}

func (s *Service) SyncMonitor(ctx context.Context) (SectionSummary, error) {
	if s.repo != nil {
		rows, err := s.repo.ListSyncJobs(ctx, 50)
		if err != nil {
			return SectionSummary{}, err
		}
		return SectionSummary{AppCode: "integration-center", Section: "sync-monitor", Items: rows}, nil
	}
	return SectionSummary{AppCode: "integration-center", Section: "sync-monitor", Items: []map[string]interface{}{
		{"job": "企微成员增量同步", "tenant": "杭州鹿鸣科技", "capability": "成员同步", "status": "success", "success_rate": 99},
		{"job": "京东订单同步", "tenant": "宁波青禾家居", "capability": "订单同步", "status": "failed", "success_rate": 42},
	}}, nil
}

func (s *Service) Quota(ctx context.Context) (SectionSummary, error) {
	if s.repo != nil {
		rows, err := s.repo.ListQuotaPolicies(ctx, 50)
		if err != nil {
			return SectionSummary{}, err
		}
		return SectionSummary{AppCode: "integration-center", Section: "quota", Items: rows}, nil
	}
	return SectionSummary{AppCode: "integration-center", Section: "quota", Items: []map[string]interface{}{
		{"policy": "企业微信租户通用策略", "scope": "tenant", "daily_limit": 100000, "qps": 20, "status": "enabled"},
		{"policy": "鹿鸣科技企微专属覆盖", "scope": "tenant", "daily_limit": 300000, "qps": 50, "status": "enabled"},
	}}, nil
}

func (s *Service) Alerts(ctx context.Context) (SectionSummary, error) {
	if s.repo != nil {
		rows, err := s.repo.ListAlerts(ctx, 50)
		if err != nil {
			return SectionSummary{}, err
		}
		return SectionSummary{AppCode: "integration-center", Section: "alerts", Items: rows}, nil
	}
	return SectionSummary{AppCode: "integration-center", Section: "alerts", Items: []map[string]interface{}{
		{"title": "京东连接授权已过期", "level": "critical", "status": "open"},
		{"title": "成员同步接近配额", "level": "warning", "status": "open"},
	}}, nil
}

func (s *Service) Logs(ctx context.Context) (SectionSummary, error) {
	if s.repo != nil {
		rows, err := s.repo.ListAPICallLogs(ctx, 100)
		if err != nil {
			return SectionSummary{}, err
		}
		return SectionSummary{AppCode: "integration-center", Section: "logs", Items: rows}, nil
	}
	return SectionSummary{AppCode: "integration-center", Section: "logs", Items: []map[string]interface{}{
		{"request_id": "req_wecom_8e92", "type": "api", "path": "/cgi-bin/department/list", "status": 200, "success": true},
		{"request_id": "req_jd_7a21", "type": "api", "path": "/api/order/search", "status": 200, "success": true},
	}}, nil
}

func (s *Service) eventsFromLogs(ctx context.Context) ([]IntegrationEvent, error) {
	rows, err := s.repo.ListAPICallLogs(ctx, 5)
	if err != nil {
		return nil, err
	}
	events := make([]IntegrationEvent, 0, len(rows))
	for _, row := range rows {
		status := "success"
		if row.Status != "success" {
			status = row.Status
		}
		events = append(events, IntegrationEvent{
			ID:        row.RequestID,
			Source:    row.CallType,
			EventType: endpointValue(row.Endpoint),
			Status:    status,
			LatencyMS: row.DurationMS,
			Occurred:  row.CalledAt.Format(time.RFC3339),
		})
	}
	return events, nil
}

func endpointValue(value *string) string {
	if value == nil || *value == "" {
		return "integration.event"
	}
	return *value
}

func channelsForAccessMode(accessMode string) []string {
	switch accessMode {
	case "OAuth2":
		return []string{"平台 API", "数据同步"}
	case "第三方服务商":
		return []string{"Webhook", "数据同步"}
	case "Webhook":
		return []string{"Webhook"}
	case "API Key":
		return []string{"平台 API", "网关代理"}
	default:
		return []string{"平台 API"}
	}
}

func healthScoreForStatus(status string, openAlerts int64) int {
	if openAlerts > 0 {
		return 72
	}
	switch status {
	case "online", "enabled":
		return 96
	case "beta":
		return 88
	case "maintenance":
		return 65
	default:
		return 0
	}
}

func alertTone(openAlerts int64) string {
	if openAlerts > 0 {
		return "warning"
	}
	return "success"
}

func statusFromCount(count int64) string {
	if count > 0 {
		return "active"
	}
	return "done"
}
