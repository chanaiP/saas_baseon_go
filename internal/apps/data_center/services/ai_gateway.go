package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	aiccservices "saas_baseon_go/internal/apps/ai_capability_center/services"
	"saas_baseon_go/internal/apps/data_center/domain"
	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
)

const anomalyAnalysisScenarioCode = "anomaly_analysis"

type aiGatewayInvoker interface {
	Invoke(ctx context.Context, req aiccservices.InvokeRequest) (aiccservices.InvokeResponse, error)
}

type gatewayAIAnalyzer struct {
	gateway aiGatewayInvoker
}

func NewGatewayAIAnalyzer(gateway aiGatewayInvoker) AIAnalyzer {
	if gateway == nil {
		return localAIAnalyzer{}
	}
	return gatewayAIAnalyzer{gateway: gateway}
}

func (g gatewayAIAnalyzer) AnalyzeAnomaly(ctx context.Context, anomaly models.DataCenterAnomalyRecord, rule models.DataCenterAnomalyRule) (localDiagnosis, error) {
	if g.gateway == nil {
		return localDiagnosis{}, errors.New("AI Gateway 未配置")
	}
	req := aiccservices.InvokeRequest{
		TenantID:       fmt.Sprintf("%d", anomaly.TenantID),
		AppCode:        domain.AppCode,
		AppName:        "Ai经营决策中心",
		AIScenarioCode: anomalyAnalysisScenarioCode,
		RequestID:      fmt.Sprintf("dc_anomaly_%d_%d", anomaly.ID, time.Now().UnixNano()),
		Params: map[string]interface{}{
			"usage_amount": 1,
			"usage_unit":   "calls",
			"temperature":  0.2,
			"max_tokens":   1200,
		},
		Input: anomalyAnalysisMessages(anomaly, rule),
	}
	resp, err := g.gateway.Invoke(ctx, req)
	if err != nil {
		return localDiagnosis{}, err
	}
	if resp.Status != "success" {
		return localDiagnosis{}, fmt.Errorf("AI Gateway 调用失败: %s", resp.Status)
	}
	return diagnosisFromGatewayData(resp.Data, buildLocalDiagnosis(anomaly, rule)), nil
}

func anomalyAnalysisMessages(anomaly models.DataCenterAnomalyRecord, rule models.DataCenterAnomalyRule) map[string]interface{} {
	context := map[string]interface{}{
		"anomaly": map[string]interface{}{
			"anomaly_code":    anomaly.AnomalyCode,
			"title":           anomaly.Title,
			"business_domain": anomaly.BusinessDomain,
			"object_type":     anomaly.ObjectType,
			"object_code":     anomaly.ObjectCode,
			"object_name":     anomaly.ObjectName,
			"stat_date":       anomaly.StatDate.Format("2006-01-02"),
			"level":           anomaly.AnomalyLevel,
			"confidence":      anomaly.ConfidenceScore,
			"impact_amount":   anomaly.ImpactAmount,
			"evidence":        evidenceToMetrics(anomaly.EvidenceJSON),
		},
		"rule": map[string]interface{}{
			"rule_code":          rule.RuleCode,
			"rule_name":          rule.RuleName,
			"target_object_type": rule.TargetObjectType,
			"metric_conditions":  rule.MetricConditionsJSON,
			"level_config":       rule.LevelConfigJSON,
		},
	}
	raw, _ := json.Marshal(context)
	return map[string]interface{}{
		"messages": []map[string]string{
			{
				"role": "system",
				"content": strings.Join([]string{
					"你是服装零售集团的经营分析专家。",
					"你只基于输入的异常、规则和证据做原因分析、影响评估和整改建议。",
					"异常是否成立由规则引擎决定，你不得推翻异常结论。",
					"只返回一段合法 JSON，不要 Markdown，不要解释性前后缀。",
				}, "\n"),
			},
			{
				"role": "user",
				"content": fmt.Sprintf(`请分析以下经营异常，并严格返回 JSON：
{
  "problem_summary": "一句话说明异常问题",
  "impact_summary": "说明业务影响、范围和金额",
  "reasons": [
    {"type": "root_cause", "label": "原因标题", "content": "结合证据说明原因"}
  ],
  "suggestions": [
    {"type": "action", "content": "可执行的整改动作"}
  ],
  "task_suggestion": {
    "title": "整改任务标题",
    "owner_role": "建议责任角色",
    "deadline_days": 3,
    "priority": "low|medium|high|critical",
    "target_desc": "整改目标"
  },
  "confidence_explanation": "解释置信度来自规则证据、数据完整度和偏离程度"
}

输入上下文：
%s`, string(raw)),
			},
		},
	}
}

func diagnosisFromGatewayData(data map[string]interface{}, fallback localDiagnosis) localDiagnosis {
	if parsed, ok := diagnosisFromMap(data, fallback); ok {
		return parsed
	}
	content := gatewayTextContent(data)
	if content == "" {
		return fallback
	}
	var decoded map[string]interface{}
	if err := json.Unmarshal([]byte(normalizeGatewayJSONContent(content)), &decoded); err == nil {
		if parsed, ok := diagnosisFromMap(decoded, fallback); ok {
			return parsed
		}
	}
	fallback.ProblemSummary = content
	return fallback
}

func diagnosisFromMap(data map[string]interface{}, fallback localDiagnosis) (localDiagnosis, bool) {
	if nested, ok := mapValue(data, "diagnosis"); ok {
		data = nested
	}
	result := fallback
	matched := false
	result.ProblemSummary = stringValue(data, "problem_summary", result.ProblemSummary)
	if strings.TrimSpace(fmt.Sprint(data["problem_summary"])) != "" && strings.TrimSpace(fmt.Sprint(data["problem_summary"])) != "<nil>" {
		matched = true
	}
	result.ImpactSummary = stringValue(data, "impact_summary", result.ImpactSummary)
	if strings.TrimSpace(fmt.Sprint(data["impact_summary"])) != "" && strings.TrimSpace(fmt.Sprint(data["impact_summary"])) != "<nil>" {
		matched = true
	}
	result.ConfidenceExplanation = stringValue(data, "confidence_explanation", result.ConfidenceExplanation)
	if strings.TrimSpace(fmt.Sprint(data["confidence_explanation"])) != "" && strings.TrimSpace(fmt.Sprint(data["confidence_explanation"])) != "<nil>" {
		matched = true
	}
	if reasons, ok := mapSliceValue(data, "reasons"); ok {
		result.Reasons = reasons
		matched = true
	}
	if suggestions, ok := mapSliceValue(data, "suggestions"); ok {
		result.Suggestions = suggestions
		matched = true
	}
	if task, ok := mapValue(data, "task_suggestion"); ok {
		result.TaskSuggestion = task
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
