package bootstrap

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBootstrapVerifyResultPassed(t *testing.T) {
	require.True(t, BootstrapVerifyResult{Checks: []BootstrapCheck{{Passed: true}}}.Passed())
	require.False(t, BootstrapVerifyResult{Checks: []BootstrapCheck{{Passed: true}, {Passed: false}}}.Passed())
}

func TestBootstrapVerifyResultSummary(t *testing.T) {
	summary := BootstrapVerifyResult{Checks: []BootstrapCheck{
		{Name: "platform_tenant", Passed: true, Actual: 1, Expected: "1"},
	}}.Summary()

	require.True(t, strings.HasPrefix(summary, "bootstrap_verify_result=PASS"))
	require.Contains(t, summary, "platform_tenant actual=1 expected=1 result=PASS")
}
