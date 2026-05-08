package handlers

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPlanUpdatePayloadWhitelistsFields(t *testing.T) {
	name := " Enterprise "
	status := 1
	payload := planUpdatePayload{PlanName: &name, Status: &status}

	updates := payload.Updates()

	require.Equal(t, "Enterprise", updates["plan_name"])
	require.Equal(t, 1, updates["status"])
	require.NotContains(t, updates, "id")
	require.NotContains(t, updates, "created_at")
}

func TestFeatureAndQuotaUpdatePayloadTrimWhitelistedFields(t *testing.T) {
	featureName := " Users "
	quotaCode := " max_users "

	require.Equal(t, "Users", featureUpdatePayload{FeatureName: &featureName}.Updates()["feature_name"])
	require.Equal(t, "max_users", quotaUpdatePayload{QuotaCode: &quotaCode}.Updates()["quota_code"])
}
