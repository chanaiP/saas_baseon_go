package dictionary

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDictTypeMutationRequiresPlatformAdmin(t *testing.T) {
	svc := NewService(nil)
	viewer := Viewer{TenantID: 2, IsPlatformAdmin: false}

	_, err := svc.CreateType(context.Background(), viewer, CreateTypeCommand{Code: "demo", Name: "演示"})
	require.True(t, errors.Is(err, ErrPlatformOnly))

	_, err = svc.UpdateType(context.Background(), viewer, 1, UpdateTypeCommand{})
	require.True(t, errors.Is(err, ErrPlatformOnly))

	_, err = svc.DeleteType(context.Background(), viewer, 1)
	require.True(t, errors.Is(err, ErrPlatformOnly))
}
