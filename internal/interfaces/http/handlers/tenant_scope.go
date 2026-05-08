package handlers

import "gorm.io/gorm"

type TenantScopeService struct {
	db *gorm.DB
}

func (h *IdentityHandler) tenantScope() TenantScopeService {
	return TenantScopeService{db: h.db}
}

func (s TenantScopeService) Query(tenantID uint64) *gorm.DB {
	return s.db.Where("tenant_id = ?", tenantID)
}

func (s TenantScopeService) Active(tenantID uint64) *gorm.DB {
	return s.Query(tenantID).Where("deleted_at IS NULL")
}

func (s TenantScopeService) ActiveByID(tenantID uint64, id interface{}) *gorm.DB {
	return s.Active(tenantID).Where("id = ?", id)
}

func (s TenantScopeService) ActiveByCode(tenantID uint64, code string) *gorm.DB {
	return s.Active(tenantID).Where("code = ?", code)
}
