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

const draftGenerationScenarioCode = "ai_geo_draft_generation"

type aiGatewayInvoker interface {
	Invoke(ctx context.Context, req aiccservices.InvokeRequest) (aiccservices.InvokeResponse, error)
}

type gatewayDraftGenerator struct {
	gateway aiGatewayInvoker
}

type localDraftGenerator struct{}

func NewGatewayDraftGenerator(gateway aiGatewayInvoker) DraftGenerator {
	if gateway == nil {
		return localDraftGenerator{}
	}
	return gatewayDraftGenerator{gateway: gateway}
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

func draftGenerationMessages(req DraftGenerationRequest) map[string]interface{} {
	context := map[string]interface{}{
		"prompt":  req.Payload.Prompt,
		"skill":   req.Payload.Skill,
		"brand":   brandContext(req.Brand),
		"product": productContext(req.Product),
	}
	raw, _ := json.Marshal(context)
	return map[string]interface{}{
		"messages": []map[string]string{
			{
				"role": "system",
				"content": strings.Join([]string{
					"你是一个熟悉 GEO 内容增长、品牌种草和渠道改写的内容策略专家。",
					"你必须基于输入的品牌、商品、技能和用户提示生成可审核的中文母稿。",
					"只返回合法 JSON，不要 Markdown，不要解释性前后缀。",
				}, "\n"),
			},
			{
				"role": "user",
				"content": fmt.Sprintf(`请生成一篇 AI GEO 母稿，并严格返回 JSON：
{
  "title": "不超过 42 个中文字符的标题",
  "summary": "一句话说明内容方向",
  "body": "完整母稿正文，包含开头、核心卖点、场景表达和收束",
  "keywords": ["关键词1", "关键词2"]
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

func draftGenerationFromGatewayData(data map[string]interface{}, fallback DraftGenerationResult) DraftGenerationResult {
	if parsed, ok := draftGenerationFromMap(data, fallback); ok {
		return parsed
	}
	content := gatewayTextContent(data)
	if content == "" {
		return fallback
	}
	var decoded map[string]interface{}
	if err := json.Unmarshal([]byte(normalizeGatewayJSONContent(content)), &decoded); err == nil {
		if parsed, ok := draftGenerationFromMap(decoded, fallback); ok {
			return parsed
		}
	}
	fallback.Body = content
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
