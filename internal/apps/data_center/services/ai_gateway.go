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
		},
		Input: map[string]interface{}{
			"anomaly": map[string]interface{}{
				"anomaly_code":    anomaly.AnomalyCode,
				"title":           anomaly.Title,
				"business_domain": anomaly.BusinessDomain,
				"object_type":     anomaly.ObjectType,
				"object_code":     anomaly.ObjectCode,
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
			"expected_json_schema": map[string]interface{}{
				"problem_summary":        "string",
				"impact_summary":         "string",
				"reasons":                "array",
				"suggestions":            "array",
				"task_suggestion":        "object",
				"confidence_explanation": "string",
			},
		},
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

func diagnosisFromGatewayData(data map[string]interface{}, fallback localDiagnosis) localDiagnosis {
	if parsed, ok := diagnosisFromMap(data, fallback); ok {
		return parsed
	}
	content := gatewayTextContent(data)
	if content == "" {
		return fallback
	}
	var decoded map[string]interface{}
	if err := json.Unmarshal([]byte(content), &decoded); err == nil {
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
	result.ProblemSummary = stringValue(data, "problem_summary", result.ProblemSummary)
	result.ImpactSummary = stringValue(data, "impact_summary", result.ImpactSummary)
	result.ConfidenceExplanation = stringValue(data, "confidence_explanation", result.ConfidenceExplanation)
	if reasons, ok := mapSliceValue(data, "reasons"); ok {
		result.Reasons = reasons
	}
	if suggestions, ok := mapSliceValue(data, "suggestions"); ok {
		result.Suggestions = suggestions
	}
	if task, ok := mapValue(data, "task_suggestion"); ok {
		result.TaskSuggestion = task
	}
	return result, result.ProblemSummary != "" || len(result.Reasons) > 0 || len(result.Suggestions) > 0
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
