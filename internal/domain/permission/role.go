package permission

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrRoleNotFound = errors.New("role not found")
	ErrRoleInvalid  = errors.New("role is invalid")
)

type Role struct {
	ID            uint64
	TenantID      uint64
	Code          string
	Name          string
	Description   *string
	Status        int
	PermissionIDs []uint64
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func NewRole(tenantID uint64, code, name string, description *string, permissionIDs []uint64) (Role, error) {
	code = strings.TrimSpace(code)
	name = strings.TrimSpace(name)
	if tenantID == 0 || code == "" || name == "" {
		return Role{}, ErrRoleInvalid
	}
	return Role{
		TenantID:      tenantID,
		Code:          code,
		Name:          name,
		Description:   description,
		Status:        1,
		PermissionIDs: NormalizePermissionIDs(permissionIDs),
	}, nil
}

func NormalizePermissionIDs(ids []uint64) []uint64 {
	if ids == nil {
		return nil
	}
	seen := map[uint64]struct{}{}
	out := make([]uint64, 0, len(ids))
	for _, id := range ids {
		if id == 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}
