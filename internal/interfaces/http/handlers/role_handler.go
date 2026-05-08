package handlers

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	apppermission "saas_baseon_go/internal/application/permission"
	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
	"saas_baseon_go/internal/infrastructure/persistence/postgres/repositories"
	"saas_baseon_go/internal/interfaces/http/response"
)

func (h *IdentityHandler) AssignableRoles(c *gin.Context) {
	roles, _, err := h.roleService().List(c.Request.Context(), apppermission.RoleListQuery{TenantID: h.requestTenantID(c), Limit: 200})
	if err != nil {
		respondBadRequest(c, err)
		return
	}
	items := make([]gin.H, 0, len(roles))
	for _, role := range roles {
		items = append(items, gin.H{"id": role.ID, "code": role.Code, "name": role.Name})
	}
	response.OK(c, paginated(items))
}

func (h *IdentityHandler) Roles(c *gin.Context) {
	user, ok := h.currentUser(c)
	if !ok {
		response.Error(c, 401, response.CodeUnauthorized, "请先登录")
		return
	}
	skip, limit := paginationParams(c)
	query := h.tenantScope().Active(user.TenantID)
	if kw := strings.TrimSpace(c.Query("kw")); kw != "" {
		like := "%" + strings.ToLower(kw) + "%"
		query = query.Where("lower(code) LIKE ? OR lower(name) LIKE ?", like, like)
	}
	var total int64
	_ = query.Model(&models.Role{}).Count(&total).Error
	var roles []models.Role
	_ = query.Order("id asc").Offset(skip).Limit(limit).Find(&roles).Error
	filterForSubscription := !h.viewerHasPlatformScope(user)
	items := make([]gin.H, 0, len(roles))
	for _, role := range roles {
		items = append(items, h.roleToJSON(role, filterForSubscription))
	}
	response.OK(c, paginatedWithTotal(items, total, skip, limit))
}

func (h *IdentityHandler) Role(c *gin.Context) {
	user, ok := h.currentUser(c)
	if !ok {
		response.Error(c, 401, response.CodeUnauthorized, "请先登录")
		return
	}
	var role models.Role
	if err := h.tenantScope().ActiveByID(user.TenantID, parseUintParam(c, "id")).First(&role).Error; err != nil {
		response.Error(c, 404, response.CodeNotFound, "角色不存在")
		return
	}
	response.OK(c, h.roleToJSON(role, false))
}

func (h *IdentityHandler) CreateRole(c *gin.Context) {
	user, ok := h.currentUser(c)
	if !ok {
		response.Error(c, 401, response.CodeUnauthorized, "请先登录")
		return
	}
	var body rolePayload
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, 400, response.CodeBadRequest, "请求参数错误")
		return
	}
	if err := h.requireFeatureAccess(user.TenantID, "role_manage"); err != nil {
		respondForbidden(c, err)
		return
	}
	if err := h.requireQuotaAvailable(user.TenantID, "max_roles", 1); err != nil {
		respondBadRequest(c, err)
		return
	}
	code := strings.TrimSpace(body.Code)
	name := strings.TrimSpace(derefString(body.Name))
	if code == "" || name == "" {
		response.Error(c, 400, response.CodeBadRequest, "角色编码和名称不能为空")
		return
	}
	role := models.Role{TenantID: user.TenantID, Code: code, Name: name, Description: nullableTrimmed(body.Description), Status: 1}
	if err := h.db.Create(&role).Error; err != nil {
		respondBadRequest(c, err)
		return
	}
	h.audit(c, user.TenantID, user.ID, "role", "create", "创建角色 "+role.Name, gin.H{"id": role.ID, "code": role.Code})
	response.OK(c, h.roleToJSON(role, false))
}

func (h *IdentityHandler) UpdateRole(c *gin.Context) {
	user, ok := h.currentUser(c)
	if !ok {
		response.Error(c, 401, response.CodeUnauthorized, "请先登录")
		return
	}
	var body rolePayload
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, 400, response.CodeBadRequest, "请求参数错误")
		return
	}
	var role models.Role
	if err := h.tenantScope().ActiveByID(user.TenantID, parseUintParam(c, "id")).First(&role).Error; err != nil {
		response.Error(c, 404, response.CodeNotFound, "角色不存在")
		return
	}
	if body.PermissionIDs != nil {
		if err := h.validateRolePermissionIDs(user, body.PermissionIDs); err != nil {
			respondBadRequest(c, err)
			return
		}
		if err := h.validateRoleDataOverrides(user.TenantID, body.DataOverrides); err != nil {
			respondBadRequest(c, err)
			return
		}
	}
	if err := h.db.Transaction(func(tx *gorm.DB) error {
		updates := map[string]interface{}{}
		if body.Name != nil {
			updates["name"] = strings.TrimSpace(*body.Name)
		}
		if body.Description != nil {
			updates["description"] = nullableTrimmed(body.Description)
		}
		if len(updates) > 0 {
			if err := tx.Model(&role).Updates(updates).Error; err != nil {
				return err
			}
		}
		if body.PermissionIDs != nil {
			if err := replaceRolePermissionsWithOverrides(tx, role.ID, body.PermissionIDs, body.DataOverrides); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		respondBadRequest(c, err)
		return
	}
	_ = h.db.First(&role, role.ID).Error
	h.invalidateRoleAuthorizationCache(role.ID)
	h.audit(c, user.TenantID, user.ID, "role", "update", "编辑角色 "+role.Name, gin.H{"id": role.ID, "code": role.Code, "permission_ids": body.PermissionIDs, "data_overrides": body.DataOverrides})
	response.OK(c, h.roleToJSON(role, false))
}

func (h *IdentityHandler) DeleteRole(c *gin.Context) {
	user, ok := h.currentUser(c)
	if !ok {
		response.Error(c, 401, response.CodeUnauthorized, "请先登录")
		return
	}
	var role models.Role
	if err := h.tenantScope().ActiveByID(user.TenantID, parseUintParam(c, "id")).First(&role).Error; err != nil {
		response.Error(c, 404, response.CodeNotFound, "角色不存在")
		return
	}
	now := time.Now()
	if err := h.db.Model(&role).Updates(map[string]interface{}{"deleted_at": now, "code": tombstoneUniqueValue(role.Code, role.ID, 64)}).Error; err != nil {
		respondBadRequest(c, err)
		return
	}
	h.invalidateRoleAuthorizationCache(role.ID)
	h.audit(c, user.TenantID, user.ID, "role", "delete", "删除角色 "+role.Name, gin.H{"id": role.ID})
	response.OK(c, gin.H{"deleted": parseUintParam(c, "id")})
}

func (h *IdentityHandler) roleService() *apppermission.RoleService {
	return apppermission.NewRoleService(repositories.NewRoleRepository(h.db))
}
