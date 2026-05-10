package handlers

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
	"saas_baseon_go/internal/interfaces/http/response"
)

func (h *IdentityHandler) Tenants(c *gin.Context) {
	skip, limit := paginationParams(c)
	query := h.db.Where("deleted_at IS NULL")
	var total int64
	_ = query.Model(&models.Tenant{}).Count(&total).Error
	var rows []models.Tenant
	_ = query.Order("id desc").Offset(skip).Limit(limit).Find(&rows).Error
	items := make([]gin.H, 0, len(rows))
	for _, row := range rows {
		items = append(items, tenantToJSON(h.db, row))
	}
	response.OK(c, paginatedWithTotal(items, total, skip, limit))
}

func (h *IdentityHandler) Tenant(c *gin.Context) {
	var row models.Tenant
	if err := h.db.Where("deleted_at IS NULL").First(&row, c.Param("id")).Error; err != nil {
		response.Error(c, 404, response.CodeNotFound, "主体不存在")
		return
	}
	response.OK(c, tenantToJSON(h.db, row))
}

func (h *IdentityHandler) CreateTenant(c *gin.Context) {
	var body struct {
		Code            string  `json:"code"`
		Name            string  `json:"name"`
		Status          int     `json:"status"`
		AdminName       string  `json:"admin_name"`
		AdminEmployeeNo string  `json:"admin_employee_no"`
		AdminPhone      *string `json:"admin_phone"`
		AdminPassword   string  `json:"admin_password"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, 400, response.CodeBadRequest, "请求参数错误")
		return
	}
	tenant, err := h.createTenantWithAdmin(body.Code, body.Name, body.Status, body.AdminName, body.AdminEmployeeNo, body.AdminPhone, body.AdminPassword)
	if err != nil {
		respondBadRequest(c, err)
		return
	}
	h.auditCurrentUser(c, "tenant", "create", "创建主体 "+tenant.Name, gin.H{"tenant_id": tenant.ID, "code": tenant.Code, "name": tenant.Name})
	response.OK(c, tenantToJSON(h.db, tenant))
}

func (h *IdentityHandler) CreateTenantWithPackage(c *gin.Context) {
	var body struct {
		Tenant struct {
			Code            string  `json:"code"`
			Name            string  `json:"name"`
			Status          int     `json:"status"`
			AdminName       string  `json:"admin_name"`
			AdminEmployeeNo string  `json:"admin_employee_no"`
			AdminPhone      *string `json:"admin_phone"`
			AdminPassword   string  `json:"admin_password"`
		} `json:"tenant"`
		Package tenantPackagePayload `json:"package"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, 400, response.CodeBadRequest, "请求参数错误")
		return
	}
	var tenant models.Tenant
	err := h.db.Transaction(func(tx *gorm.DB) error {
		created, err := h.createTenantWithAdminOnDB(tx, body.Tenant.Code, body.Tenant.Name, body.Tenant.Status, body.Tenant.AdminName, body.Tenant.AdminEmployeeNo, body.Tenant.AdminPhone, body.Tenant.AdminPassword)
		if err != nil {
			return err
		}
		tenant = created
		return h.saveTenantPackageOnDB(tx, tenant.ID, body.Package)
	})
	if err != nil {
		respondBadRequest(c, err)
		return
	}
	h.auditCurrentUser(c, "tenant", "create_with_package", "创建主体并配置套餐 "+tenant.Name, gin.H{"tenant_id": tenant.ID, "code": tenant.Code, "name": tenant.Name})
	response.OK(c, tenantToJSON(h.db, tenant))
}

func (h *IdentityHandler) UpdateTenant(c *gin.Context) {
	var tenant models.Tenant
	if err := h.db.Where("deleted_at IS NULL").First(&tenant, c.Param("id")).Error; err != nil {
		response.Error(c, 404, response.CodeNotFound, "主体不存在")
		return
	}
	var body struct {
		Name             *string `json:"name"`
		Status           *int    `json:"status"`
		ContactName      *string `json:"contact_name"`
		ContactPhone     *string `json:"contact_phone"`
		BrandDisplayName *string `json:"brand_display_name"`
		BrandLogoData    *string `json:"brand_logo_data"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, 400, response.CodeBadRequest, "请求参数错误")
		return
	}
	updates := map[string]interface{}{}
	if body.Name != nil {
		updates["name"] = strings.TrimSpace(*body.Name)
	}
	if body.Status != nil {
		updates["status"] = *body.Status
	}
	if body.ContactName != nil {
		updates["contact_name"] = nullableTrimmed(body.ContactName)
	}
	if body.ContactPhone != nil {
		phone, msg := normalizeOptionalPhone(body.ContactPhone)
		if msg != "" {
			response.Error(c, 400, response.CodeBadRequest, strings.Replace(msg, "手机号", "联系人手机号", 1))
			return
		}
		updates["contact_phone"] = phone
	}
	if body.BrandDisplayName != nil {
		updates["brand_display_name"] = nullableTrimmed(body.BrandDisplayName)
	}
	if body.BrandLogoData != nil {
		logo := nullableTrimmed(body.BrandLogoData)
		if logo != nil && len(*logo) > 800000 {
			response.Error(c, 400, response.CodeBadRequest, "Logo 数据过大")
			return
		}
		updates["brand_logo_data"] = logo
	}
	if len(updates) > 0 {
		if err := h.db.Model(&tenant).Updates(updates).First(&tenant, tenant.ID).Error; err != nil {
			respondBadRequest(c, err)
			return
		}
	}
	if body.Status != nil && *body.Status != 1 {
		h.invalidateSessionsForTenant(tenant.ID)
	}
	h.auditCurrentUser(c, "tenant", "update", "更新主体 "+tenant.Name, gin.H{"tenant_id": tenant.ID, "changes": updates})
	response.OK(c, tenantToJSON(h.db, tenant))
}

func (h *IdentityHandler) UpdateTenantStatus(c *gin.Context) {
	var body struct {
		Status int `json:"status"`
	}
	_ = c.ShouldBindJSON(&body)
	if err := h.db.Model(&models.Tenant{}).Where("id = ? AND deleted_at IS NULL", c.Param("id")).Update("status", body.Status).Error; err != nil {
		respondBadRequest(c, err)
		return
	}
	if body.Status == 0 {
		h.invalidateSessionsForTenant(parseUintParam(c, "id"))
	}
	var tenant models.Tenant
	_ = h.db.First(&tenant, c.Param("id")).Error
	h.auditCurrentUser(c, "tenant", "status_update", "更新主体状态 "+tenant.Name, gin.H{"tenant_id": tenant.ID, "status": body.Status})
	response.OK(c, tenantToJSON(h.db, tenant))
}

func (h *IdentityHandler) DeleteTenant(c *gin.Context) {
	id := parseUintParam(c, "id")
	var tenant models.Tenant
	if err := h.db.Where("id = ? AND deleted_at IS NULL", id).First(&tenant).Error; err != nil {
		response.Error(c, 400, response.CodeBadRequest, "主体不存在")
		return
	}
	var platformAdminCount int64
	_ = h.db.Model(&models.AppUser{}).Where("tenant_id = ? AND is_platform_admin = ? AND deleted_at IS NULL", id, true).Count(&platformAdminCount).Error
	if platformAdminCount > 0 {
		response.Error(c, 400, response.CodeBadRequest, "该主体下存在系统管理员账号，请先取消其平台权限或迁移后再删除")
		return
	}
	h.invalidateSessionsForTenant(id)
	now := time.Now()
	if err := h.db.Model(&tenant).Updates(map[string]interface{}{"deleted_at": now, "status": 0, "code": tombstoneUniqueValue(tenant.Code, tenant.ID, 64)}).Error; err != nil {
		respondBadRequest(c, err)
		return
	}
	h.auditCurrentUser(c, "tenant", "delete", "删除主体「"+tenant.Name+"」", gin.H{"tenant_id": id})
	response.OK(c, gin.H{"deleted": id})
}
