package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	aiccservices "saas_baseon_go/internal/apps/ai_capability_center/services"
	"saas_baseon_go/internal/apps/ai_geo/dto"
	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
)

const (
	draftGenerationScenarioCode        = "ai_geo_draft_generation"
	channelContentStandardScenarioCode = "channel_content_standard_generate"
	channelContentEditorScenarioCode   = "ai_geo_channel_content_editor"
	auditSuggestionScenarioCode        = "ai_geo_audit_suggestion"
)

type aiGatewayInvoker interface {
	Invoke(ctx context.Context, req aiccservices.InvokeRequest) (aiccservices.InvokeResponse, error)
}

type GatewayInvokeRequest = aiccservices.InvokeRequest
type GatewayStreamEvent = aiccservices.InvokeStreamEvent

type GatewayStreamInvoker interface {
	InvokeStream(ctx context.Context, req aiccservices.InvokeRequest, emit func(aiccservices.InvokeStreamEvent) error) error
}

type gatewayDraftGenerator struct {
	gateway aiGatewayInvoker
}

type gatewayChannelContentGenerator struct {
	gateway aiGatewayInvoker
}

type gatewayAuditAdvisor struct {
	gateway aiGatewayInvoker
}

type localDraftGenerator struct{}
type localChannelContentGenerator struct{}
type localAuditAdvisor struct{}

func NewGatewayDraftGenerator(gateway aiGatewayInvoker) DraftGenerator {
	if gateway == nil {
		return localDraftGenerator{}
	}
	return gatewayDraftGenerator{gateway: gateway}
}

func NewGatewayChannelContentGenerator(gateway aiGatewayInvoker) ChannelContentGenerator {
	if gateway == nil {
		return localChannelContentGenerator{}
	}
	return gatewayChannelContentGenerator{gateway: gateway}
}

func NewGatewayAuditAdvisor(gateway aiGatewayInvoker) AuditAdvisor {
	if gateway == nil {
		return localAuditAdvisor{}
	}
	return gatewayAuditAdvisor{gateway: gateway}
}

func (s *Service) StreamGatewayScenario(ctx context.Context, viewer dto.Viewer, req GatewayInvokeRequest, allowedScenario string, emit func(GatewayStreamEvent) error) error {
	if s.streamGateway == nil {
		return errors.New("AI Gateway 流式服务未配置")
	}
	if viewer.TenantID == 0 || strings.TrimSpace(allowedScenario) == "" {
		return ErrInvalidInput
	}
	if strings.TrimSpace(req.AIScenarioCode) != allowedScenario {
		return ErrInvalidInput
	}
	req.TenantID = fmt.Sprintf("%d", viewer.TenantID)
	req.UserID = fmt.Sprintf("%d", viewer.UserID)
	req.AppCode = "ai-geo"
	req.AppName = "AI GEO"
	if strings.TrimSpace(req.RequestID) == "" {
		req.RequestID = fmt.Sprintf("ai_geo_stream_%d_%d", viewer.TenantID, time.Now().UnixNano())
	}
	return s.streamGateway.InvokeStream(ctx, req, emit)
}

func (localDraftGenerator) GenerateDraft(_ context.Context, req DraftGenerationRequest) (DraftGenerationResult, error) {
	prompt := strings.TrimSpace(req.Payload.Prompt)
	title := prompt
	if len([]rune(title)) > 42 {
		title = string([]rune(title)[:42])
	}
	styleRule := ""
	if req.StyleTemplate != nil {
		styleRule = fmt.Sprintf("\n\n参考写作风格：学习“%s”的结构、语气和表达手法，但不得复制原文句子或引用外部文章事实。", req.StyleTemplate.TemplateName)
	}
	body := fmt.Sprintf("围绕“%s”生成一篇可直接生成渠道内容的母稿。\n\n创作要求：结合品牌资料、商品卖点、渠道语境和热点素材，输出结构化内容，后续可生成小红书、知乎、独立站等渠道版本。%s", prompt, styleRule)
	return DraftGenerationResult{
		Title:    title,
		Summary:  "AI 工作台生成母稿，可继续精修或直接生成渠道内容。",
		Body:     body,
		Keywords: []string{"AI生成", "母稿"},
		Source:   "ai_workbench",
	}, nil
}

func (localChannelContentGenerator) GenerateChannelContent(_ context.Context, req ChannelContentGenerationRequest) (ChannelContentGenerationResult, error) {
	packageBody := channelPackageJSON(req, map[string]interface{}{
		"title":   defaultString(req.Payload.Title, channelDefaultTitle(req)),
		"summary": stringValueFromPtr(req.Draft.Summary),
		"body":    defaultString(req.Payload.Body, req.Draft.Body),
		"tags":    defaultChannelTags(req.Channel.ChannelName),
	})
	return ChannelContentGenerationResult{
		Title: defaultString(req.Payload.Title, channelDefaultTitle(req)),
		Body:  packageBody,
	}, nil
}

func (localAuditAdvisor) Advise(_ context.Context, req AuditAdviceRequest) (AuditAdviceResult, error) {
	title := ""
	body := ""
	if req.Draft != nil {
		title = req.Draft.Title
		body = req.Draft.Body
	}
	if req.Content != nil {
		title = req.Content.Title
		body = req.Content.Body
	}
	passed := strings.TrimSpace(title) != "" && strings.TrimSpace(body) != ""
	risk := "low"
	summary := "内容基础字段完整，建议人工复核渠道口径和禁用词。"
	if !passed {
		risk = "high"
		summary = "标题或正文缺失，建议补齐后再继续流转。"
	}
	return AuditAdviceResult{
		RiskLevel: risk,
		Passed:    passed,
		Summary:   summary,
		Suggestions: []map[string]interface{}{
			{"type": "content_quality", "content": summary},
		},
		ModelCode: "local-ai-geo-audit",
	}, nil
}

func (g gatewayDraftGenerator) GenerateDraft(ctx context.Context, req DraftGenerationRequest) (DraftGenerationResult, error) {
	if g.gateway == nil {
		return DraftGenerationResult{}, errors.New("AI Gateway 未配置")
	}
	resp, err := g.gateway.Invoke(ctx, aiccservices.InvokeRequest{
		TenantID:       fmt.Sprintf("%d", req.Viewer.TenantID),
		AppCode:        "ai-geo",
		AppName:        "AI GEO",
		AIScenarioCode: draftGenerationScenarioCode,
		UserID:         fmt.Sprintf("%d", req.Viewer.UserID),
		RequestID:      fmt.Sprintf("ai_geo_draft_%d_%d", req.Viewer.TenantID, time.Now().UnixNano()),
		Params: map[string]interface{}{
			"usage_amount": 1,
			"usage_unit":   "calls",
			"temperature":  0.7,
			"max_tokens":   1800,
		},
		Input: draftGenerationMessages(req),
	})
	if err != nil {
		return DraftGenerationResult{}, err
	}
	if resp.Status != "success" {
		return DraftGenerationResult{}, fmt.Errorf("AI Gateway 调用失败: %s", resp.Status)
	}
	return draftGenerationFromGatewayData(resp.Data, localDraftGeneratorResult(req)), nil
}

func (g gatewayChannelContentGenerator) GenerateChannelContent(ctx context.Context, req ChannelContentGenerationRequest) (ChannelContentGenerationResult, error) {
	if g.gateway == nil {
		return ChannelContentGenerationResult{}, errors.New("AI Gateway 未配置")
	}
	resp, err := g.gateway.Invoke(ctx, aiccservices.InvokeRequest{
		TenantID:       fmt.Sprintf("%d", req.Viewer.TenantID),
		AppCode:        "ai-geo",
		AppName:        "AI GEO",
		AIScenarioCode: channelContentStandardScenarioCode,
		UserID:         fmt.Sprintf("%d", req.Viewer.UserID),
		RequestID:      fmt.Sprintf("ai_geo_channel_%d_%d", req.Viewer.TenantID, time.Now().UnixNano()),
		Params: map[string]interface{}{
			"usage_amount": 1,
			"usage_unit":   "calls",
			"temperature":  0.6,
			"max_tokens":   2600,
		},
		Input: channelRewriteMessages(req),
	})
	if err != nil {
		return ChannelContentGenerationResult{}, err
	}
	if resp.Status != "success" {
		return ChannelContentGenerationResult{}, fmt.Errorf("AI Gateway 调用失败: %s", resp.Status)
	}
	return channelContentFromGatewayData(resp.Data, localChannelContentGeneratorResult(req)), nil
}

func (g gatewayAuditAdvisor) Advise(ctx context.Context, req AuditAdviceRequest) (AuditAdviceResult, error) {
	if g.gateway == nil {
		return AuditAdviceResult{}, errors.New("AI Gateway 未配置")
	}
	resp, err := g.gateway.Invoke(ctx, aiccservices.InvokeRequest{
		TenantID:       fmt.Sprintf("%d", req.Viewer.TenantID),
		AppCode:        "ai-geo",
		AppName:        "AI GEO",
		AIScenarioCode: auditSuggestionScenarioCode,
		UserID:         fmt.Sprintf("%d", req.Viewer.UserID),
		RequestID:      fmt.Sprintf("ai_geo_audit_%d_%d", req.Viewer.TenantID, time.Now().UnixNano()),
		Params: map[string]interface{}{
			"usage_amount": 1,
			"usage_unit":   "calls",
			"temperature":  0.2,
			"max_tokens":   1000,
		},
		Input: auditSuggestionMessages(req),
	})
	if err != nil {
		return AuditAdviceResult{}, err
	}
	if resp.Status != "success" {
		return AuditAdviceResult{}, fmt.Errorf("AI Gateway 调用失败: %s", resp.Status)
	}
	return auditAdviceFromGatewayData(resp.Data, localAuditAdvisorResult(req)), nil
}

func draftGenerationMessages(req DraftGenerationRequest) map[string]interface{} {
	context := map[string]interface{}{
		"prompt":          req.Payload.Prompt,
		"skill":           req.Payload.Skill,
		"brand":           brandContext(req.Brand),
		"product":         productContext(req.Product),
		"skus":            skuContext(req.SKUs),
		"hotspot":         hotspotContext(req.Hotspot),
		"style_template":  styleTemplateSnapshot(req.StyleTemplate),
		"recent_titles":   req.RecentTitles,
		"source_snapshot": req.Payload.SourceSnapshot,
	}
	raw, _ := json.Marshal(context)
	return map[string]interface{}{
		"messages": []map[string]string{
			{
				"role": "system",
				"content": strings.Join([]string{
					"你是 GEO 母稿共创专家，不是普通文章生成器。",
					"你必须基于输入的品牌、商品、技能、用户提示和对话收敛结果，生成可拆分、可被 AI 搜索和问答引擎引用的中文内容母版。",
					"母稿是多渠道内容源文件，不是聊天回复；严禁输出“好的、明白、我建议、请确认、现在可以生成、母稿使用说明”等沟通过程话术。",
					"如果 recent_titles 非空，title 必须避开这些历史标题，换一个具体内容角度，不得复用同一句标题或只做标点改写。",
					"body 必须按内容母版写作，包含核心问题、可被 AI 引用的核心答案、目标用户、品牌/产品定位、使用场景、用户痛点、选择理由、对比逻辑、证据与论点、FAQ、GEO 关键词结构、渠道改写建议。",
					"FAQ 至少 6 个问题，渠道改写建议至少包含小红书、知乎、抖音、公众号；不要包含生成说明、替换占位符、导师分析或确认问题。",
					"如果 style_template 存在，必须学习其结构路径、语气人设、句式手法和 prompt_fragment；只迁移写法，不复制原文句子，不把外部来源事实当作当前品牌或商品事实。",
					"只返回合法 JSON，不要 Markdown，不要解释性前后缀。",
				}, "\n"),
			},
			{
				"role": "user",
				"content": fmt.Sprintf(`请生成一篇 AI GEO 母稿，并严格返回 JSON：
{
  "title": "不超过 42 个中文字符的标题",
  "summary": "一句可直接展示给用户的内容母版摘要，不要写生成过程",
  "body": "完整可发布 GEO 内容母版，只写内容本身，不写聊天确认、不写创作说明、不写使用说明",
  "keywords": ["关键词1", "关键词2"]
}

输入上下文：
%s`, string(raw)),
			},
		},
	}
}

func channelRewriteMessages(req ChannelContentGenerationRequest) map[string]interface{} {
	context := map[string]interface{}{
		"draft": map[string]interface{}{
			"draft_code":      req.Draft.DraftCode,
			"title":           req.Draft.Title,
			"summary":         stringValueFromPtr(req.Draft.Summary),
			"body":            req.Draft.Body,
			"keywords":        req.Draft.Keywords,
			"source_snapshot": req.Draft.SourceSnapshot,
		},
		"channel": map[string]interface{}{
			"channel_code":         req.Channel.ChannelCode,
			"channel_name":         req.Channel.ChannelName,
			"channel_type":         req.Channel.ChannelType,
			"content_forms":        req.Channel.ContentForms,
			"support_modes":        req.Channel.SupportModes,
			"default_publish_mode": req.Channel.DefaultPublishMode,
		},
		"override": map[string]interface{}{
			"title": req.Payload.Title,
			"body":  req.Payload.Body,
		},
		"skill": "channel_content_standard_generate",
	}
	raw, _ := json.Marshal(context)
	return map[string]interface{}{
		"messages": []map[string]string{
			{
				"role": "system",
				"content": strings.Join([]string{
					"你是渠道内容标准生成 Skill。",
					"你只负责“母稿 -> 当前平台渠道内容包”。",
					"必须基于母稿、品牌/商品基础资料、渠道规范和素材规则生成，不得新增未确认事实。",
					"不同平台必须输出不同结构：小红书图文笔记、知乎问答回答、微信公众号图文文章、抖音视频脚本、微博短帖、百家号图文文章、独立站 SEO/FAQ 内容。",
					"生成完成不等于可发布，nextAction 必须体现待编辑、待补充素材或可加入发布计划。",
					"只返回合法 JSON，不要 Markdown，不要解释性前后缀。",
				}, "\n"),
			},
			{
				"role": "user",
				"content": fmt.Sprintf(`请为当前平台生成结构化渠道内容包，并严格返回 JSON：
{
  "channel": "平台名称",
  "status": "generated",
  "contentType": "图文笔记|问答回答|图文文章|视频脚本|短帖|SEO文章",
  "contentPayload": {
    "title": "",
    "summary": "",
    "body": "",
    "tags": [],
    "extraFields": {}
  },
  "assetPayload": {
    "requiredAssets": [],
    "matchedAssets": [],
    "missingAssets": [],
    "aiGenerateSuggestions": []
  },
  "geoPayload": {
    "brandEntityIncluded": true,
    "productEntityIncluded": true,
    "keywords": [],
    "geoSuggestions": []
  },
  "riskNotes": [],
  "nextAction": "待编辑|待补充素材|可加入发布计划"
}

输入上下文：
%s`, string(raw)),
			},
		},
	}
}

func auditSuggestionMessages(req AuditAdviceRequest) map[string]interface{} {
	context := map[string]interface{}{
		"object_type": req.ObjectType,
		"draft":       map[string]interface{}{},
		"content":     map[string]interface{}{},
		"channel":     map[string]interface{}{},
	}
	if req.Draft != nil {
		context["draft"] = map[string]interface{}{
			"draft_code": req.Draft.DraftCode,
			"title":      req.Draft.Title,
			"summary":    stringValueFromPtr(req.Draft.Summary),
			"body":       req.Draft.Body,
			"keywords":   req.Draft.Keywords,
		}
	}
	if req.Content != nil {
		context["content"] = map[string]interface{}{
			"id":    req.Content.ID,
			"title": req.Content.Title,
			"body":  req.Content.Body,
		}
	}
	if req.Channel != nil {
		context["channel"] = map[string]interface{}{
			"channel_code":  req.Channel.ChannelCode,
			"channel_name":  req.Channel.ChannelName,
			"content_forms": req.Channel.ContentForms,
		}
	}
	raw, _ := json.Marshal(context)
	return map[string]interface{}{
		"messages": []map[string]string{
			{
				"role": "system",
				"content": strings.Join([]string{
					"你是品牌内容质量检查专家，负责检查 GEO 内容的事实一致性、平台合规、口径风险和内容质量。",
					"只返回合法 JSON，不要 Markdown，不要解释性前后缀。",
				}, "\n"),
			},
			{
				"role": "user",
				"content": fmt.Sprintf(`请检查以下内容，并严格返回 JSON：
{
  "risk_level": "low|medium|high|critical",
  "passed": true,
  "summary": "检查结论摘要",
  "suggestions": [
    {"type": "risk|quality|compliance", "content": "具体建议"}
  ],
  "model_code": "模型或规则名称"
}

输入上下文：
%s`, string(raw)),
			},
		},
	}
}

func localDraftGeneratorResult(req DraftGenerationRequest) DraftGenerationResult {
	result, _ := localDraftGenerator{}.GenerateDraft(context.Background(), req)
	return result
}

func localChannelContentGeneratorResult(req ChannelContentGenerationRequest) ChannelContentGenerationResult {
	result, _ := localChannelContentGenerator{}.GenerateChannelContent(context.Background(), req)
	return result
}

func channelPackageJSON(req ChannelContentGenerationRequest, contentPayload map[string]interface{}) string {
	contentType := channelContentType(req.Channel.ChannelName)
	packageBody := map[string]interface{}{
		"channel":        req.Channel.ChannelName,
		"status":         "generated",
		"contentType":    contentType,
		"contentPayload": contentPayload,
		"assetPayload": map[string]interface{}{
			"requiredAssets":        requiredAssetsForChannel(req.Channel.ChannelName),
			"matchedAssets":         []map[string]interface{}{},
			"missingAssets":         missingAssetsForChannel(req.Channel.ChannelName),
			"aiGenerateSuggestions": aiAssetSuggestionsForChannel(req.Channel.ChannelName),
		},
		"geoPayload": map[string]interface{}{
			"brandEntityIncluded":   true,
			"productEntityIncluded": req.Draft.ProductID != nil,
			"keywords":              parseStringSliceJSON(req.Draft.Keywords),
			"geoSuggestions":        []string{"保留品牌实体、商品实体、人群词、场景词和问题词。"},
		},
		"riskNotes":  []string{"渠道内容生成完成后仍需补齐素材、账号和发布时间。"},
		"nextAction": "待编辑",
	}
	raw, _ := json.MarshalIndent(packageBody, "", "  ")
	return string(raw)
}

func channelDefaultTitle(req ChannelContentGenerationRequest) string {
	if req.Channel.ChannelName == "" {
		return req.Draft.Title
	}
	return fmt.Sprintf("%s｜%s", req.Draft.Title, req.Channel.ChannelName)
}

func channelContentType(channelName string) string {
	switch channelName {
	case "小红书":
		return "图文笔记"
	case "知乎":
		return "问答回答"
	case "微信公众号", "百家号":
		return "图文文章"
	case "抖音":
		return "视频脚本"
	case "微博":
		return "短帖"
	case "独立站":
		return "SEO文章"
	default:
		return "渠道内容"
	}
}

func defaultChannelTags(channelName string) []string {
	switch channelName {
	case "小红书":
		return []string{"GEO", "种草", "穿搭建议"}
	case "微博":
		return []string{"GEO", "品牌内容"}
	default:
		return []string{"GEO"}
	}
}

func requiredAssetsForChannel(channelName string) []map[string]string {
	switch channelName {
	case "小红书":
		return []map[string]string{{"slot": "封面图", "requirement": "商品图或模特图，3-6 张优先", "type": "image"}}
	case "微信公众号", "百家号":
		return []map[string]string{{"slot": "封面图", "requirement": "适合图文文章封面的品牌/商品图", "type": "image"}}
	case "抖音":
		return []map[string]string{{"slot": "商品视频", "requirement": "商品或模特视频素材，缺失时输出拍摄建议", "type": "video"}}
	case "独立站":
		return []map[string]string{{"slot": "商品图", "requirement": "可用于 SEO 页面或商品详情的真实商品图", "type": "image"}}
	default:
		return []map[string]string{}
	}
}

func missingAssetsForChannel(channelName string) []map[string]string {
	required := requiredAssetsForChannel(channelName)
	items := make([]map[string]string, 0, len(required))
	for _, item := range required {
		items = append(items, map[string]string{
			"slot":        item["slot"],
			"requirement": item["requirement"],
			"suggestion":  "从资料中心选择已授权素材；若为氛围图或分镜参考，可进入渠道图片生成。",
		})
	}
	return items
}

func aiAssetSuggestionsForChannel(channelName string) []string {
	if channelName == "抖音" {
		return []string{"可生成分镜参考图，但不能替代真实商品视频。"}
	}
	if len(requiredAssetsForChannel(channelName)) > 0 {
		return []string{"可生成封面背景或氛围图；商品主图、SKU 图必须来自资料中心。"}
	}
	return []string{}
}

func localAuditAdvisorResult(req AuditAdviceRequest) AuditAdviceResult {
	result, _ := localAuditAdvisor{}.Advise(context.Background(), req)
	return result
}

func draftGenerationFromGatewayData(data map[string]interface{}, fallback DraftGenerationResult) DraftGenerationResult {
	if parsed, ok := draftGenerationFromMap(data, fallback); ok {
		return sanitizeDraftGenerationResult(parsed, fallback)
	}
	content := gatewayTextContent(data)
	if content == "" {
		return fallback
	}
	var decoded map[string]interface{}
	if err := json.Unmarshal([]byte(normalizeGatewayJSONContent(content)), &decoded); err == nil {
		if parsed, ok := draftGenerationFromMap(decoded, fallback); ok {
			return sanitizeDraftGenerationResult(parsed, fallback)
		}
	}
	if decoded := looseDraftJSONMap(content); len(decoded) > 0 {
		if parsed, ok := draftGenerationFromMap(decoded, fallback); ok {
			return sanitizeDraftGenerationResult(parsed, fallback)
		}
	}
	return draftGenerationFromArticleText(content, fallback)
}

func channelContentFromGatewayData(data map[string]interface{}, fallback ChannelContentGenerationResult) ChannelContentGenerationResult {
	if parsed, ok := channelContentFromMap(data, fallback); ok {
		return parsed
	}
	content := gatewayTextContent(data)
	if content == "" {
		return fallback
	}
	var decoded map[string]interface{}
	if err := json.Unmarshal([]byte(normalizeGatewayJSONContent(content)), &decoded); err == nil {
		if parsed, ok := channelContentFromMap(decoded, fallback); ok {
			return parsed
		}
	}
	fallback.Body = content
	return fallback
}

func auditAdviceFromGatewayData(data map[string]interface{}, fallback AuditAdviceResult) AuditAdviceResult {
	if parsed, ok := auditAdviceFromMap(data, fallback); ok {
		return parsed
	}
	content := gatewayTextContent(data)
	if content == "" {
		return fallback
	}
	var decoded map[string]interface{}
	if err := json.Unmarshal([]byte(normalizeGatewayJSONContent(content)), &decoded); err == nil {
		if parsed, ok := auditAdviceFromMap(decoded, fallback); ok {
			return parsed
		}
	}
	fallback.Summary = content
	return fallback
}

func draftGenerationFromMap(data map[string]interface{}, fallback DraftGenerationResult) (DraftGenerationResult, bool) {
	if nested, ok := mapValue(data, "draft"); ok {
		data = nested
	}
	result := fallback
	matched := false
	if title := stringValue(data, "title", ""); title != "" {
		result.Title = trimRunes(title, 42)
		matched = true
	}
	if summary := stringValue(data, "summary", ""); summary != "" {
		result.Summary = summary
		matched = true
	}
	if body := stringValue(data, "body", ""); body != "" {
		result.Body = body
		matched = true
	}
	if keywords, ok := stringSliceValue(data, "keywords"); ok {
		result.Keywords = keywords
		matched = true
	}
	result.Source = "ai_workbench"
	return result, matched
}

func draftGenerationFromArticleText(content string, fallback DraftGenerationResult) DraftGenerationResult {
	body := cleanDraftBody(content)
	if body == "" {
		body = cleanDraftBody(fallback.Body)
	}
	result := fallback
	result.Body = defaultString(body, fallback.Body)
	result.Title = cleanDraftTitle(result.Title, result.Body, fallback.Title)
	result.Summary = cleanDraftSummary(result.Summary, result.Body)
	result.Source = "ai_workbench"
	return result
}

func sanitizeDraftGenerationResult(result DraftGenerationResult, fallback DraftGenerationResult) DraftGenerationResult {
	result.Body = defaultString(cleanDraftBody(result.Body), cleanDraftBody(fallback.Body))
	result.Title = cleanDraftTitle(defaultString(result.Title, fallback.Title), result.Body, fallback.Title)
	result.Summary = cleanDraftSummary(defaultString(result.Summary, fallback.Summary), result.Body)
	if len(result.Keywords) == 0 {
		result.Keywords = fallback.Keywords
	}
	result.Source = defaultString(result.Source, "ai_workbench")
	return result
}

func cleanDraftTitle(title string, body string, fallback string) string {
	value := strings.TrimSpace(stripMarkdownTokens(title))
	for _, marker := range []string{"\n", ">", "：>", "正文", "摘要", "好的", "明白", "收到", "我会", "请确认"} {
		if idx := strings.Index(value, marker); idx > 0 {
			value = strings.TrimSpace(value[:idx])
		}
	}
	if value == "" || isMentorTalk(value) {
		if fallbackValue := strings.TrimSpace(stripMarkdownTokens(fallback)); fallbackValue != "" && !isMentorTalk(fallbackValue) {
			value = fallbackValue
		}
	}
	if value == "" || isMentorTalk(value) {
		value = firstArticleHeading(body)
	}
	if value == "" {
		value = "AI GEO 母稿"
	}
	return trimRunes(value, 42)
}

func cleanDraftSummary(summary string, body string) string {
	value := strings.TrimSpace(stripMarkdownTokens(summary))
	if value == "" || isMentorTalk(value) {
		value = firstSentence(body)
	}
	return trimRunes(value, 120)
}

func cleanDraftBody(content string) string {
	value := stripMarkdownFence(strings.TrimSpace(content))
	if value == "" {
		return ""
	}
	if extracted := extractMarkedArticleBody(value); extracted != "" {
		value = extracted
	}
	lines := strings.Split(value, "\n")
	cleaned := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(stripMarkdownTokens(line))
		line = strings.Trim(line, "- ")
		if line == "" {
			if len(cleaned) > 0 && cleaned[len(cleaned)-1] != "" {
				cleaned = append(cleaned, "")
			}
			continue
		}
		if isMentorTalk(line) || isDraftInstructionLine(line) {
			continue
		}
		cleaned = append(cleaned, line)
	}
	value = strings.TrimSpace(strings.Join(cleaned, "\n"))
	value = strings.ReplaceAll(value, "\n\n\n", "\n\n")
	return value
}

func extractMarkedArticleBody(content string) string {
	markers := []string{"正文（", "正文:", "正文：", "正文\n", "完整母稿正文", "文章正文"}
	for _, marker := range markers {
		if idx := strings.Index(content, marker); idx >= 0 {
			rest := content[idx+len(marker):]
			if marker == "正文（" {
				if end := strings.Index(rest, "）："); end >= 0 {
					rest = rest[end+len("）："):]
				} else if end := strings.Index(rest, "):"); end >= 0 {
					rest = rest[end+len("):"):]
				}
			}
			return strings.TrimSpace(rest)
		}
	}
	return ""
}

func isMentorTalk(line string) bool {
	prefixes := []string{"好的", "明白", "收到", "我会先", "我先", "建议先", "现在可以", "现在，我可以", "你可以", "请确认", "在生成前", "如果你", "这决定了", "我不会", "我建议"}
	for _, prefix := range prefixes {
		if strings.HasPrefix(line, prefix) {
			return true
		}
	}
	return strings.Contains(line, "补充这篇文章要回答") || strings.Contains(line, "继续帮你放大和收敛")
}

func isDraftInstructionLine(line string) bool {
	markers := []string{"母稿结构", "母稿使用说明", "替换占位符", "补充细节", "调整语气", "生成方向", "输出要求", "输入上下文", "收敛 brief", "用户原始想法", "本轮对话"}
	for _, marker := range markers {
		if strings.Contains(line, marker) {
			return true
		}
	}
	return false
}

func firstArticleHeading(body string) string {
	for _, line := range strings.Split(body, "\n") {
		line = strings.TrimSpace(stripMarkdownTokens(line))
		line = strings.Trim(line, "#- ")
		if line == "" || isMentorTalk(line) || isDraftInstructionLine(line) {
			continue
		}
		if strings.HasPrefix(line, "标题") {
			parts := strings.SplitN(line, "：", 2)
			if len(parts) == 2 {
				return strings.TrimSpace(parts[1])
			}
		}
		return line
	}
	return ""
}

func firstSentence(body string) string {
	body = strings.ReplaceAll(strings.TrimSpace(body), "\n", " ")
	for _, sep := range []string{"。", "！", "？"} {
		if idx := strings.Index(body, sep); idx > 0 {
			return strings.TrimSpace(body[:idx+len(sep)])
		}
	}
	return body
}

func stripMarkdownFence(content string) string {
	value := strings.TrimSpace(content)
	value = strings.TrimPrefix(value, "```json")
	value = strings.TrimPrefix(value, "```JSON")
	value = strings.TrimPrefix(value, "```")
	value = strings.TrimSuffix(value, "```")
	return strings.TrimSpace(value)
}

func stripMarkdownTokens(content string) string {
	value := strings.ReplaceAll(content, "**", "")
	value = strings.ReplaceAll(value, "###", "")
	value = strings.ReplaceAll(value, "##", "")
	value = strings.ReplaceAll(value, "#", "")
	value = strings.ReplaceAll(value, "`", "")
	return strings.TrimSpace(value)
}

func channelContentFromMap(data map[string]interface{}, fallback ChannelContentGenerationResult) (ChannelContentGenerationResult, bool) {
	if nested, ok := mapValue(data, "channel_content"); ok {
		data = nested
	}
	if _, hasPackage := data["contentPayload"]; hasPackage {
		result := fallback
		if contentPayload, ok := mapValue(data, "contentPayload"); ok {
			if title := firstStringValue(contentPayload, "title", "question_title", "answer_title", "video_title", "post_text", "seo_title", "page_title"); title != "" {
				result.Title = trimRunes(title, 80)
			}
		}
		if result.Title == "" {
			result.Title = fallback.Title
		}
		raw, _ := json.MarshalIndent(data, "", "  ")
		result.Body = string(raw)
		return result, true
	}
	result := fallback
	matched := false
	if title := stringValue(data, "title", ""); title != "" {
		result.Title = trimRunes(title, 80)
		matched = true
	}
	if body := stringValue(data, "body", ""); body != "" {
		result.Body = body
		matched = true
	}
	return result, matched
}

func firstStringValue(data map[string]interface{}, keys ...string) string {
	for _, key := range keys {
		if value := stringValue(data, key, ""); value != "" {
			return value
		}
	}
	return ""
}

func auditAdviceFromMap(data map[string]interface{}, fallback AuditAdviceResult) (AuditAdviceResult, bool) {
	if nested, ok := mapValue(data, "audit"); ok {
		data = nested
	}
	result := fallback
	matched := false
	if risk := stringValue(data, "risk_level", ""); risk != "" {
		result.RiskLevel = risk
		matched = true
	}
	if passed, ok := data["passed"].(bool); ok {
		result.Passed = passed
		matched = true
	}
	if summary := stringValue(data, "summary", ""); summary != "" {
		result.Summary = summary
		matched = true
	}
	if modelCode := stringValue(data, "model_code", ""); modelCode != "" {
		result.ModelCode = modelCode
		matched = true
	}
	if suggestions, ok := mapSliceValue(data, "suggestions"); ok {
		result.Suggestions = suggestions
		matched = true
	}
	return result, matched
}

func gatewayTextContent(data map[string]interface{}) string {
	for _, key := range []string{"content", "text", "output_text"} {
		if value, ok := data[key].(string); ok && strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	choices, ok := data["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return ""
	}
	choice, ok := choices[0].(map[string]interface{})
	if !ok {
		return ""
	}
	if message, ok := choice["message"].(map[string]interface{}); ok {
		if content, ok := message["content"].(string); ok {
			return strings.TrimSpace(content)
		}
	}
	if text, ok := choice["text"].(string); ok {
		return strings.TrimSpace(text)
	}
	return ""
}

func mapValue(data map[string]interface{}, key string) (map[string]interface{}, bool) {
	value, ok := data[key].(map[string]interface{})
	return value, ok
}

func parseJSONMap(raw string) map[string]interface{} {
	var value map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &value); err != nil || value == nil {
		return map[string]interface{}{}
	}
	return value
}

func normalizeGatewayJSONContent(content string) string {
	value := strings.TrimSpace(content)
	value = strings.TrimPrefix(value, "```json")
	value = strings.TrimPrefix(value, "```JSON")
	value = strings.TrimPrefix(value, "```")
	value = strings.TrimSuffix(value, "```")
	return strings.TrimSpace(value)
}

func looseDraftJSONMap(content string) map[string]interface{} {
	value := normalizeGatewayJSONContent(content)
	if !strings.Contains(value, `"title"`) && !strings.Contains(value, `"body"`) {
		return nil
	}
	result := map[string]interface{}{}
	if title := looseJSONStringField(value, "title", []string{"summary", "body", "keywords"}); title != "" {
		result["title"] = title
	}
	if summary := looseJSONStringField(value, "summary", []string{"body", "keywords"}); summary != "" {
		result["summary"] = summary
	}
	if body := looseJSONStringField(value, "body", []string{"keywords"}); body != "" {
		result["body"] = body
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

func looseJSONStringField(raw string, key string, nextKeys []string) string {
	marker := `"` + key + `"`
	idx := strings.Index(raw, marker)
	if idx < 0 {
		return ""
	}
	rest := raw[idx+len(marker):]
	colon := strings.Index(rest, ":")
	if colon < 0 {
		return ""
	}
	rest = strings.TrimSpace(rest[colon+1:])
	if !strings.HasPrefix(rest, `"`) {
		return ""
	}
	rest = rest[1:]
	end := len(rest)
	for _, nextKey := range nextKeys {
		for _, nextMarker := range []string{`",` + "\n" + `  "` + nextKey + `"`, `",` + "\n" + `    "` + nextKey + `"`, `",` + "\r\n" + `  "` + nextKey + `"`, `", "` + nextKey + `"`} {
			if nextIdx := strings.Index(rest, nextMarker); nextIdx >= 0 && nextIdx < end {
				end = nextIdx
			}
		}
	}
	value := strings.TrimSpace(rest[:end])
	value = strings.TrimSuffix(value, `"`)
	value = strings.TrimSuffix(value, ",")
	value = strings.TrimSuffix(value, `"`)
	value = strings.ReplaceAll(value, `\n`, "\n")
	value = strings.ReplaceAll(value, `\"`, `"`)
	value = strings.ReplaceAll(value, `\\`, `\`)
	return strings.TrimSpace(value)
}

func stringValue(data map[string]interface{}, key string, fallback string) string {
	if value, ok := data[key].(string); ok && strings.TrimSpace(value) != "" {
		return strings.TrimSpace(value)
	}
	return fallback
}

func stringSliceValue(data map[string]interface{}, key string) ([]string, bool) {
	raw, ok := data[key].([]interface{})
	if !ok {
		if typed, ok := data[key].([]string); ok {
			return typed, true
		}
		return nil, false
	}
	items := make([]string, 0, len(raw))
	for _, item := range raw {
		value := strings.TrimSpace(fmt.Sprint(item))
		if value != "" && value != "<nil>" {
			items = append(items, value)
		}
	}
	return items, true
}

func parseStringSliceJSON(raw string) []string {
	var items []string
	if err := json.Unmarshal([]byte(raw), &items); err == nil {
		return items
	}
	var loose []interface{}
	if err := json.Unmarshal([]byte(raw), &loose); err != nil {
		return []string{}
	}
	result := make([]string, 0, len(loose))
	for _, item := range loose {
		value := strings.TrimSpace(fmt.Sprint(item))
		if value != "" && value != "<nil>" {
			result = append(result, value)
		}
	}
	return result
}

func mapSliceValue(data map[string]interface{}, key string) ([]map[string]interface{}, bool) {
	raw, ok := data[key].([]interface{})
	if !ok {
		if typed, ok := data[key].([]map[string]interface{}); ok {
			return typed, true
		}
		return nil, false
	}
	items := make([]map[string]interface{}, 0, len(raw))
	for _, item := range raw {
		if typed, ok := item.(map[string]interface{}); ok {
			items = append(items, typed)
		}
	}
	return items, true
}

func brandContext(brand *models.AiGeoBrandCard) map[string]interface{} {
	if brand == nil {
		return map[string]interface{}{}
	}
	return map[string]interface{}{
		"brand_code":      brand.BrandCode,
		"brand_name":      brand.BrandName,
		"positioning":     stringValueFromPtr(brand.Positioning),
		"target_audience": stringValueFromPtr(brand.TargetAudience),
		"price_band":      stringValueFromPtr(brand.PriceBand),
		"tone":            stringValueFromPtr(brand.Tone),
		"keywords":        brand.Keywords,
	}
}

func productContext(product *models.AiGeoProductCard) map[string]interface{} {
	if product == nil {
		return map[string]interface{}{}
	}
	return map[string]interface{}{
		"product_code":   product.ProductCode,
		"product_name":   product.ProductName,
		"category_name":  stringValueFromPtr(product.CategoryName),
		"selling_points": product.SellingPoints,
		"faq":            product.FAQ,
		"content_angles": product.ContentAngles,
	}
}

func skuContext(skus []models.AiGeoSKU) []map[string]interface{} {
	items := make([]map[string]interface{}, 0, len(skus))
	for _, sku := range skus {
		items = append(items, map[string]interface{}{
			"sku_code":     sku.SKUCode,
			"sku_name":     sku.SKUName,
			"attributes":   sku.Attributes,
			"price":        sku.Price,
			"image_url":    stringValueFromPtr(sku.ImageURL),
			"stock_status": sku.StockStatus,
		})
	}
	return items
}

func hotspotContext(hotspot *models.AiGeoHotspot) map[string]interface{} {
	if hotspot == nil {
		return map[string]interface{}{}
	}
	return map[string]interface{}{
		"platform":    hotspot.Platform,
		"title":       hotspot.Title,
		"heat_score":  hotspot.HeatScore,
		"source_url":  stringValueFromPtr(hotspot.SourceURL),
		"captured_at": hotspot.CapturedAt.Format(time.RFC3339),
		"metadata":    parseJSONMap(hotspot.Metadata),
	}
}

func stringValueFromPtr(value *string) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}

func trimRunes(value string, max int) string {
	runes := []rune(strings.TrimSpace(value))
	if len(runes) <= max {
		return string(runes)
	}
	return string(runes[:max])
}
