package handlers

import (
	"errors"
	"testing"

	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"

	"github.com/stretchr/testify/require"
)

func TestSysParamBatchKeysTrimEmptyAndDeduplicate(t *testing.T) {
	require.Equal(t, []string{"a", "b", "c"}, splitCSVParam(" a, ,b,, c,a "))
	require.Empty(t, splitCSVParam(" , , "))
}

func TestDictItemJSONAppliesTenantOverride(t *testing.T) {
	label := "标准客户"
	sortOrder := 1
	enabled := false
	item := dictItemJSON(models.DictItem{
		ID:         7,
		DictTypeID: 3,
		Label:      "普通客户",
		Value:      "C",
		SortOrder:  3,
		Enabled:    true,
	}, &models.TenantDictItemOverride{
		CustomLabel: &label,
		SortOrder:   &sortOrder,
		Enabled:     &enabled,
	})

	require.Equal(t, "标准客户", item["label"])
	require.Equal(t, "C", item["value"])
	require.Equal(t, 1, item["sort_order"])
	require.Equal(t, false, item["enabled"])
	require.Equal(t, true, item["is_override"])
}

func TestSafeDBErrorMessageHidesConstraintDetails(t *testing.T) {
	msg := safeDBErrorMessage(errors.New("duplicate key value violates unique constraint uq_sys_param_key"))

	require.NotContains(t, msg, "duplicate key")
	require.NotContains(t, msg, "unique constraint")
	require.Equal(t, "数据已存在，请检查唯一字段", msg)
}
