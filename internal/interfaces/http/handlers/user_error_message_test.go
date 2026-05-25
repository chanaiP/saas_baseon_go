package handlers

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSafeUserErrorMessageKeepsBusinessReason(t *testing.T) {
	require.Equal(t, "工号已存在", safeUserErrorMessage(errors.New("工号已存在")))
	require.Equal(t, "组织节点不存在或不属于当前主体", safeUserErrorMessage(errors.New("组织节点不存在或不属于当前主体")))
}

func TestSafeUserErrorMessageKeepsQuotaReason(t *testing.T) {
	require.Equal(
		t,
		"用户数已达到平台配额上限（当前已用 1 / 上限 1），请调整主体套餐配额或停用不需要的用户后再新增",
		safeUserErrorMessage(&quotaExceededError{QuotaName: "用户数", Used: 1, Limit: 1}),
	)
	require.Equal(
		t,
		"用户数已达到平台配额上限，请调整主体套餐配额或停用不需要的用户后再新增",
		safeUserErrorMessage(errors.New("用户数已超出套餐配额")),
	)
}

func TestSafeUserErrorMessageMapsUniqueConstraint(t *testing.T) {
	require.Equal(t, "手机号已存在", safeUserErrorMessage(errors.New(`ERROR: duplicate key value violates unique constraint "idx_app_user_tenant_phone"`)))
	require.Equal(t, "工号已存在", safeUserErrorMessage(errors.New(`ERROR: duplicate key value violates unique constraint "idx_app_user_tenant_employee"`)))
}
