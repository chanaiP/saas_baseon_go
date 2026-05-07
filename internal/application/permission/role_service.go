package permission

import (
	"context"
	"strings"

	domain "saas_baseon_go/internal/domain/permission"
)

type RoleService struct {
	repo domain.RoleRepository
}

func NewRoleService(repo domain.RoleRepository) *RoleService {
	return &RoleService{repo: repo}
}

type RoleListQuery struct {
	TenantID uint64
	Skip     int
	Limit    int
	Keyword  string
}

type RoleCreateCommand struct {
	TenantID      uint64
	Code          string
	Name          string
	Description   *string
	PermissionIDs []uint64
}

type RoleUpdateCommand struct {
	ID            uint64
	TenantID      uint64
	Name          *string
	Description   *string
	PermissionIDs []uint64
	UpdatePerms   bool
}

func (s *RoleService) List(ctx context.Context, query RoleListQuery) ([]domain.Role, int64, error) {
	if query.TenantID == 0 {
		query.TenantID = 1
	}
	if query.Limit <= 0 || query.Limit > 200 {
		query.Limit = 50
	}
	if query.Skip < 0 {
		query.Skip = 0
	}
	return s.repo.List(ctx, domain.RoleListQuery{
		TenantID: query.TenantID,
		Skip:     query.Skip,
		Limit:    query.Limit,
		Keyword:  strings.TrimSpace(query.Keyword),
	})
}

func (s *RoleService) Get(ctx context.Context, id uint64) (domain.Role, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *RoleService) Create(ctx context.Context, cmd RoleCreateCommand) (domain.Role, error) {
	role, err := domain.NewRole(cmd.TenantID, cmd.Code, cmd.Name, cmd.Description, cmd.PermissionIDs)
	if err != nil {
		return domain.Role{}, err
	}
	return s.repo.Create(ctx, role)
}

func (s *RoleService) Update(ctx context.Context, cmd RoleUpdateCommand) (domain.Role, error) {
	role, err := s.repo.FindByID(ctx, cmd.ID)
	if err != nil {
		return domain.Role{}, err
	}
	if cmd.TenantID > 0 && role.TenantID != cmd.TenantID {
		return domain.Role{}, domain.ErrRoleNotFound
	}
	if cmd.Name != nil {
		role.Name = strings.TrimSpace(*cmd.Name)
		if role.Name == "" {
			return domain.Role{}, domain.ErrRoleInvalid
		}
	}
	if cmd.Description != nil {
		role.Description = cmd.Description
	}
	if cmd.UpdatePerms {
		role.PermissionIDs = domain.NormalizePermissionIDs(cmd.PermissionIDs)
	}
	return s.repo.Update(ctx, role, cmd.UpdatePerms)
}

func (s *RoleService) Delete(ctx context.Context, id uint64) error {
	return s.repo.Delete(ctx, id)
}
