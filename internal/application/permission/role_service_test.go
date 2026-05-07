package permission

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	domain "saas_baseon_go/internal/domain/permission"
)

func TestRoleServiceCreateNormalizesPermissions(t *testing.T) {
	repo := newFakeRoleRepo()
	service := NewRoleService(repo)

	role, err := service.Create(context.Background(), RoleCreateCommand{
		TenantID:      1,
		Code:          "admin",
		Name:          "管理员",
		PermissionIDs: []uint64{3, 3, 0, 8},
	})

	require.NoError(t, err)
	require.Equal(t, []uint64{3, 8}, role.PermissionIDs)
	require.Equal(t, "admin", repo.roles[role.ID].Code)
}

func TestRoleServiceRejectsEmptyRoleName(t *testing.T) {
	repo := newFakeRoleRepo()
	service := NewRoleService(repo)

	_, err := service.Create(context.Background(), RoleCreateCommand{TenantID: 1, Code: "empty"})

	require.ErrorIs(t, err, domain.ErrRoleInvalid)
}

func TestRoleServiceListDefaultsPagination(t *testing.T) {
	repo := newFakeRoleRepo()
	service := NewRoleService(repo)

	_, _, err := service.List(context.Background(), RoleListQuery{TenantID: 1, Limit: 1000, Skip: -1})

	require.NoError(t, err)
	require.Equal(t, domain.RoleListQuery{TenantID: 1, Skip: 0, Limit: 50}, repo.lastQuery)
}

func TestRoleServiceUpdateCanReplacePermissions(t *testing.T) {
	repo := newFakeRoleRepo()
	service := NewRoleService(repo)
	existing, err := repo.Create(context.Background(), domain.Role{TenantID: 1, Code: "ops", Name: "运营", PermissionIDs: []uint64{1}})
	require.NoError(t, err)
	name := "高级运营"

	updated, err := service.Update(context.Background(), RoleUpdateCommand{
		ID:            existing.ID,
		Name:          &name,
		PermissionIDs: []uint64{2, 2, 5},
		UpdatePerms:   true,
	})

	require.NoError(t, err)
	require.Equal(t, "高级运营", updated.Name)
	require.Equal(t, []uint64{2, 5}, updated.PermissionIDs)
}

type fakeRoleRepo struct {
	nextID    uint64
	roles     map[uint64]domain.Role
	lastQuery domain.RoleListQuery
}

func newFakeRoleRepo() *fakeRoleRepo {
	return &fakeRoleRepo{nextID: 1, roles: map[uint64]domain.Role{}}
}

func (r *fakeRoleRepo) List(_ context.Context, query domain.RoleListQuery) ([]domain.Role, int64, error) {
	r.lastQuery = query
	items := make([]domain.Role, 0, len(r.roles))
	for _, role := range r.roles {
		items = append(items, role)
	}
	return items, int64(len(items)), nil
}

func (r *fakeRoleRepo) FindByID(_ context.Context, id uint64) (domain.Role, error) {
	role, ok := r.roles[id]
	if !ok {
		return domain.Role{}, domain.ErrRoleNotFound
	}
	return role, nil
}

func (r *fakeRoleRepo) Create(_ context.Context, role domain.Role) (domain.Role, error) {
	role.ID = r.nextID
	r.nextID++
	r.roles[role.ID] = role
	return role, nil
}

func (r *fakeRoleRepo) Update(_ context.Context, role domain.Role, _ bool) (domain.Role, error) {
	r.roles[role.ID] = role
	return role, nil
}

func (r *fakeRoleRepo) Delete(_ context.Context, id uint64) error {
	delete(r.roles, id)
	return nil
}
