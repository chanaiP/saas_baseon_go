package handlers

import (
	"gorm.io/gorm"

	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
)

type DataScopeService struct {
	handler *IdentityHandler
}

func (h *IdentityHandler) dataScopeService() DataScopeService {
	return DataScopeService{handler: h}
}

func (s DataScopeService) ApplyToUserQuery(query *gorm.DB, user models.AppUser) *gorm.DB {
	return s.handler.applyUserDataScopeFilter(query, user)
}
