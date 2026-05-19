package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"
)

type ruleConditionGroup struct {
	Logic      string              `json:"logic"`
	Conditions []ruleConditionNode `json:"conditions"`
}

type ruleConditionNode struct {
	Logic      string              `json:"logic"`
	MetricCode string              `json:"metric_code"`
	Operator   string              `json:"operator"`
	Value      *float64            `json:"value"`
	Min        *float64            `json:"min"`
	Max        *float64            `json:"max"`
	Conditions []ruleConditionNode `json:"conditions"`
}

func evaluateRuleConditions(raw string, metrics map[string]float64) (bool, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "{}" || raw == "[]" {
		return false, nil
	}
	var group ruleConditionGroup
	if err := json.Unmarshal([]byte(raw), &group); err != nil {
		return false, err
	}
	if group.Logic == "" {
		group.Logic = "AND"
	}
	return evaluateConditionNodes(group.Logic, group.Conditions, metrics)
}

func evaluateConditionNodes(logic string, nodes []ruleConditionNode, metrics map[string]float64) (bool, error) {
	if len(nodes) == 0 {
		return false, nil
	}
	logic = strings.ToUpper(strings.TrimSpace(logic))
	if logic == "" {
		logic = "AND"
	}
	if logic != "AND" && logic != "OR" {
		return false, fmt.Errorf("unsupported condition logic: %s", logic)
	}
	matched := logic == "AND"
	for _, node := range nodes {
		ok, err := evaluateConditionNode(node, metrics)
		if err != nil {
			return false, err
		}
		if logic == "AND" && !ok {
			return false, nil
		}
		if logic == "OR" && ok {
			return true, nil
		}
		matched = ok
	}
	return matched, nil
}

func evaluateConditionNode(node ruleConditionNode, metrics map[string]float64) (bool, error) {
	if len(node.Conditions) > 0 {
		logic := node.Logic
		if logic == "" {
			logic = "AND"
		}
		return evaluateConditionNodes(logic, node.Conditions, metrics)
	}
	metricCode := strings.TrimSpace(node.MetricCode)
	if metricCode == "" {
		return false, errors.New("metric_code is required")
	}
	current, ok := metrics[metricCode]
	if !ok {
		return false, nil
	}
	operator := strings.ToLower(strings.TrimSpace(node.Operator))
	switch operator {
	case "gt", ">":
		return node.Value != nil && current > *node.Value, nil
	case "gte", ">=":
		return node.Value != nil && current >= *node.Value, nil
	case "lt", "<":
		return node.Value != nil && current < *node.Value, nil
	case "lte", "<=":
		return node.Value != nil && current <= *node.Value, nil
	case "eq", "=":
		return node.Value != nil && math.Abs(current-*node.Value) < 0.0001, nil
	case "neq", "!=":
		return node.Value != nil && math.Abs(current-*node.Value) >= 0.0001, nil
	case "between":
		return node.Min != nil && node.Max != nil && current >= *node.Min && current <= *node.Max, nil
	default:
		return false, fmt.Errorf("unsupported condition operator: %s", node.Operator)
	}
}
