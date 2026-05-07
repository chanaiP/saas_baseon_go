package handlers

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRequiredAuditEvents(t *testing.T) {
	events := map[string]bool{}
	for _, event := range requiredAuditEvents() {
		events[event.Module+":"+event.Action] = true
	}

	for _, key := range []string{
		"tenant:create",
		"tenant:status_update",
		"user:create",
		"user:password_reset",
		"role:update",
		"organization:delete",
		"business_unit:org_mapping_create",
	} {
		require.True(t, events[key], "missing audit event %s", key)
	}
}
