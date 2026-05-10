package permission

import "context"

type RoleListQuery struct {
	TenantID uint64
	Skip     int
	Limit    int
	Keyword  string
}

type RoleRepository interface {
	List(ctx context.Context, query RoleListQuery) ([]Role, int64, error)
	FindByID(ctx context.Context, tenantID uint64, id uint64) (Role, error)
	Create(ctx context.Context, role Role) (Role, error)
	Update(ctx context.Context, role Role, updatePermissions bool) (Role, error)
	Delete(ctx context.Context, tenantID uint64, id uint64) error
}
