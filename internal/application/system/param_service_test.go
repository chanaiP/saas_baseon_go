package system

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	domain "saas_baseon_go/internal/domain/system"
)

type memoryParamRepo struct {
	items []domain.Param
}

func (r *memoryParamRepo) List(context.Context) ([]domain.Param, error) {
	return r.items, nil
}

func (r *memoryParamRepo) FindByKey(_ context.Context, key string) (domain.Param, error) {
	for _, item := range r.items {
		if item.Key == key {
			return item, nil
		}
	}
	return domain.Param{}, domain.ErrParamNotFound
}

func (r *memoryParamRepo) Create(_ context.Context, param domain.Param) (domain.Param, error) {
	param.ID = uint64(len(r.items) + 1)
	r.items = append(r.items, param)
	return param, nil
}

func TestParamServiceCreateValidatesKey(t *testing.T) {
	service := NewParamService(&memoryParamRepo{})

	_, err := service.Create(context.Background(), CreateParamCommand{Key: " ", Value: "enabled"})

	require.ErrorIs(t, err, domain.ErrParamKeyRequired)
}

func TestParamServiceCreateValidatesValue(t *testing.T) {
	service := NewParamService(&memoryParamRepo{})

	_, err := service.Create(context.Background(), CreateParamCommand{Key: "site.mode", Value: " "})

	require.ErrorIs(t, err, domain.ErrParamValueRequired)
}

func TestParamServiceCreatePersistsParam(t *testing.T) {
	repo := &memoryParamRepo{}
	service := NewParamService(repo)

	created, err := service.Create(context.Background(), CreateParamCommand{
		Key:    "site.mode",
		Value:  "production",
		Remark: "runtime mode",
	})

	require.NoError(t, err)
	require.Equal(t, uint64(1), created.ID)
	require.Equal(t, "site.mode", created.Key)
	require.Equal(t, "production", created.Value)
}
