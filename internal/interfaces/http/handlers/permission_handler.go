package handlers

import (
	"github.com/gin-gonic/gin"

	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
	"saas_baseon_go/internal/interfaces/http/response"
)

func (h *IdentityHandler) Permissions(c *gin.Context) {
	user, ok := h.currentUser(c)
	if !ok {
		response.Error(c, 401, response.CodeUnauthorized, "请先登录")
		return
	}
	skip, limit := paginationParams(c)
	var rows []models.Permission
	query := h.tenantScope().Active(user.TenantID)
	var total int64
	_ = query.Model(&models.Permission{}).Count(&total).Error
	_ = query.Order("id asc").Offset(skip).Limit(limit).Find(&rows).Error
	items := make([]gin.H, 0, len(rows))
	for _, row := range rows {
		items = append(items, h.permissionToJSON(row))
	}
	response.OK(c, paginatedWithTotal(items, total, skip, limit))
}

func (h *IdentityHandler) PermissionTree(c *gin.Context) {
	user, ok := h.currentUser(c)
	if !ok {
		response.Error(c, 401, response.CodeUnauthorized, "请先登录")
		return
	}
	var rows []models.Permission
	_ = h.tenantScope().Active(user.TenantID).Order("sort_order asc, id asc").Find(&rows).Error
	byParent := map[uint64][]models.Permission{}
	roots := make([]models.Permission, 0)
	for _, row := range rows {
		if row.ParentID == nil {
			roots = append(roots, row)
			continue
		}
		byParent[*row.ParentID] = append(byParent[*row.ParentID], row)
	}
	var walk func([]models.Permission) []gin.H
	walk = func(nodes []models.Permission) []gin.H {
		items := make([]gin.H, 0, len(nodes))
		for _, node := range nodes {
			items = append(items, gin.H{
				"id":        node.ID,
				"parent_id": node.ParentID,
				"name":      node.Name,
				"path":      node.Path,
				"perm_type": node.PermType,
				"children":  walk(byParent[node.ID]),
			})
		}
		return items
	}
	response.OK(c, walk(roots))
}

func (h *IdentityHandler) Permission(c *gin.Context) {
	user, ok := h.currentUser(c)
	if !ok {
		response.Error(c, 401, response.CodeUnauthorized, "请先登录")
		return
	}
	var row models.Permission
	if err := h.tenantScope().ActiveByID(user.TenantID, c.Param("id")).First(&row).Error; err != nil {
		response.Error(c, 404, response.CodeNotFound, "权限不存在")
		return
	}
	response.OK(c, h.permissionToJSON(row))
}
