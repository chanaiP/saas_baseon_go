package repositories

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTombstoneRoleCodeKeepsUniqueCodeLength(t *testing.T) {
	code := tombstoneRoleCode(strings.Repeat("a", 80), 42)

	require.LessOrEqual(t, len(code), 64)
	require.Contains(t, code, "__deleted_42_")
}
