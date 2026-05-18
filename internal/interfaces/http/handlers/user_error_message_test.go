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

func TestSafeUserErrorMessageMapsUniqueConstraint(t *testing.T) {
	require.Equal(t, "手机号已存在", safeUserErrorMessage(errors.New(`ERROR: duplicate key value violates unique constraint "idx_app_user_tenant_phone"`)))
	require.Equal(t, "工号已存在", safeUserErrorMessage(errors.New(`ERROR: duplicate key value violates unique constraint "idx_app_user_tenant_employee"`)))
}
