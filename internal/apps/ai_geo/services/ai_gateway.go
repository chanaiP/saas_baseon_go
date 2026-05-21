package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	aiccservices "saas_baseon_go/internal/apps/ai_capability_center/services"
	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
)

const (
	draftGenerationScenarioCode = "ai_geo_draft_generation"
	channelRewriteScenarioCode  = "ai_geo_channel_rewrite"
	auditSuggestionScenarioCode = "ai_geo_audit_suggestion"
)

type aiGatewayInvoker interface {
	Invoke(ctx context.Context, req aiccservices.InvokeRequest) (aiccservices.InvokeResponse, error)
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

func (localDraftGenerator) GenerateDraft(_ context.Context, req DraftGenerationRequest) (DraftGenerationResult, error) {
	prompt := strings.TrimSpace(req.Payload.Prompt)
	title := prompt
	if len([]rune(title)) > 42 {
		title = string([]rune(title)[:42])
	}
	body := fmt.Sprintf("围绕“%s”生成一篇可进入审核的母稿。\n\n创作要求：结合品牌资料、商品卖点、渠道语境和热点素材，输出结构化内容，后续可生成小红书、知乎、独立站等渠道版本。", prompt)
	return DraftGenerationResult{
		Title:    title,
		Summary:  "AI 工作台生成母稿，待人工审核确认。",
		Body:     body,
		Keywords: []string{"AI生成", "母稿"},
		Source:   "ai_workbench",
	}, nil
}

func (localChannelContentGenerator) GenerateChannelContent(_ context.Context, req ChannelContentGenerationRequest) (ChannelContentGenerationResult, error) {
	return ChannelContentGenerationResult{
		Title: defaultString(req.Payload.Title, req.Draft.Title),
		Body:  defaultString(req.Payload.Body, req.Draft.Body),
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
		summary = "标题或正文缺失，建议补齐后再提交审核。"
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
		AIScenarioCode: channelRewriteScenarioCode,
		UserID:         fmt.Sprintf("%d", req.Viewer.UserID),
		RequestID:      fmt.Sprintf("ai_geo_channel_%d_%d", req.Viewer.TenantID, time.Now().UnixNano()),
		Params: map[string]interface{}{
			"usage_amount": 1,
			"usage_unit":   "calls",
			"temperature":  0.6,
			"max_tokens":   1600,
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
		"prompt":  req.Payload.Prompt,
		"skill":   req.Payload.Skill,
		"brand":   brandContext(req.Brand),
		"product": productContext(req.Product),
		"skus":    skuContext(req.SKUs),
		"hotspot": hotspotContext(req.Hotspot),
	}
	raw, _ := json.Marshal(context)
	return map[string]interface{}{
		"messages": []map[string]string{
			{
				"role": "system",
				"content": strings.Join([]string{
					"你是一个熟悉 GEO 内容增长、品牌种草和渠道改写的内容策略专家。",
					"你必须基于输入的品牌、商品、技能、用户提示和对话收敛结果，生成可审核的中文母稿。",
					"母稿是可发布正文资产，不是聊天回复；严禁输出“好的、明白、我建议、请确认、现在可以生成、母稿使用说明”等沟通过程话术。",
					"body 只能包含文章正文：开头、观点/方案、场景建议、选择理由、结论；不要包含生成说明、替换占位符、导师分析或确认问题。",
					"只返回合法 JSON，不要 Markdown，不要解释性前后缀。",
				}, "\n"),
			},
			{
				"role": "user",
				"content": fmt.Sprintf(`请生成一篇 AI GEO 母稿，并严格返回 JSON：
{
  "title": "不超过 42 个中文字符的标题",
  "summary": "一句可直接展示给用户的文章摘要，不要写生成过程",
  "body": "完整可发布文章正文，只写文章内容，不写聊天确认、不写创作说明、不写使用说明",
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
			"draft_code": req.Draft.DraftCode,
			"title":      req.Draft.Title,
			"summary":    stringValueFromPtr(req.Draft.Summary),
			"body":       req.Draft.Body,
			"keywords":   req.Draft.Keywords,
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
	}
	raw, _ := json.Marshal(context)
	return map[string]interface{}{
		"messages": []map[string]string{
			{
				"role": "system",
				"content": strings.Join([]string{
					"你是 GEO 多渠道内容改写专家。",
					"你必须基于母稿和渠道资料输出适合该渠道的内容版本。",
					"只返回合法 JSON，不要 Markdown，不要解释性前后缀。",
				}, "\n"),
			},
			{
				"role": "user",
				"content": fmt.Sprintf(`请将母稿改写成渠道内容，并严格返回 JSON：
{
  "title": "渠道标题",
  "body": "渠道正文"
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
					"你是品牌内容审核专家，负责检查 GEO 内容的事实一致性、平台合规、口径风险和内容质量。",
					"只返回合法 JSON，不要 Markdown，不要解释性前后缀。",
				}, "\n"),
			},
			{
				"role": "user",
				"content": fmt.Sprintf(`请审核以下内容，并严格返回 JSON：
{
  "risk_level": "low|medium|high|critical",
  "passed": true,
  "summary": "审核结论摘要",
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

func normalizeGatewayJSONContent(content string) string {
	value := strings.TrimSpace(content)
	value = strings.TrimPrefix(value, "```json")
	value = strings.TrimPrefix(value, "```JSON")
	value = strings.TrimPrefix(value, "```")
	value = strings.TrimSuffix(value, "```")
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
