package handlers

import (
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	appuser "saas_baseon_go/internal/application/user"
	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
	"saas_baseon_go/internal/interfaces/http/response"
)

func (h *IdentityHandler) Users(c *gin.Context) {
	user, ok := h.currentUser(c)
	if !ok {
		response.Error(c, 401, response.CodeUnauthorized, "请先登录")
		return
	}
	tenantID := h.requestTenantID(c)
	skip, limit := paginationParams(c)
	query := h.tenantScope().Active(tenantID)
	query = h.dataScopeService().ApplyToUserQuery(query, user)
	if raw := strings.TrimSpace(c.Query("status")); raw != "" {
		status, err := strconv.Atoi(raw)
		if err != nil {
			response.Error(c, 400, response.CodeBadRequest, "请求参数错误")
			return
		}
		query = query.Where("status = ?", status)
	}
	if id := parseOptionalUintQuery(c, "company_id"); id != nil {
		query = query.Where("company_id = ?", *id)
	}
	if id := parseOptionalUintQuery(c, "department_id"); id != nil {
		var userIDs []uint64
		_ = h.db.Model(&models.AppUserDepartment{}).Where("department_id = ?", *id).Distinct().Pluck("user_id", &userIDs).Error
		if len(userIDs) > 0 {
			query = query.Where("(department_id = ? OR id IN ?)", *id, userIDs)
		} else {
			query = query.Where("department_id = ?", *id)
		}
	}
	keyword := strings.TrimSpace(c.Query("keyword"))
	if keyword == "" {
		keyword = strings.TrimSpace(c.Query("kw"))
	}
	if keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("(name LIKE ? OR employee_no LIKE ? OR phone LIKE ?)", like, like, like)
	}
	var total int64
	_ = query.Model(&models.AppUser{}).Count(&total).Error
	var rows []models.AppUser
	_ = query.Order("id desc").Offset(skip).Limit(limit).Find(&rows).Error
	items := make([]gin.H, 0, len(rows))
	for _, row := range rows {
		items = append(items, h.userToJSON(row))
	}
	response.OK(c, paginatedWithTotal(items, total, skip, limit))
}

func (h *IdentityHandler) CreateUser(c *gin.Context) {
	var body struct {
		EmployeeNo    string   `json:"employee_no"`
		Password      string   `json:"password"`
		Name          string   `json:"name"`
		Phone         *string  `json:"phone"`
		Email         *string  `json:"email"`
		CompanyID     *uint64  `json:"company_id"`
		DepartmentID  *uint64  `json:"department_id"`
		DepartmentIDs []uint64 `json:"department_ids"`
		PositionIDs   []uint64 `json:"position_ids"`
		RoleIDs       []uint64 `json:"role_ids"`
		Status        int      `json:"status"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, 400, response.CodeBadRequest, "请求参数错误")
		return
	}
	if body.Status == 0 {
		body.Status = 1
	}
	tenantID := h.requestTenantID(c)
	phone, msg := normalizeOptionalPhone(body.Phone)
	if msg != "" {
		response.Error(c, 400, response.CodeBadRequest, msg)
		return
	}
	departmentIDs := normalizedUserDepartmentIDs(body.DepartmentID, body.DepartmentIDs)
	companyID := body.CompanyID
	departmentID := (*uint64)(nil)
	if len(departmentIDs) > 0 {
		departmentID = &departmentIDs[0]
		if companyID == nil {
			if found := h.companyIDForDepartment(tenantID, departmentIDs[0]); found != nil {
				companyID = found
			}
		}
	} else if companyID == nil {
		companyID = h.rootCompanyIDForTenant(tenantID)
	}
	initialPassword := strings.TrimSpace(body.Password)
	if initialPassword == "" {
		initialPassword = generateRandomPassword(14)
	}
	user := models.AppUser{TenantID: tenantID, EmployeeNo: body.EmployeeNo, Account: body.EmployeeNo, PasswordHash: devPasswordHash(initialPassword), Name: body.Name, Phone: phone, Email: body.Email, CompanyID: companyID, DepartmentID: departmentID, Status: body.Status}
	var err error
	user, err = h.userService().CreateWithRelations(c.Request.Context(), user, appuser.Relations{RoleIDs: uniqueUint64s(body.RoleIDs), PositionIDs: uniqueUint64s(body.PositionIDs), DepartmentIDs: departmentIDs})
	if err != nil {
		response.Error(c, 400, response.CodeBadRequest, safeDBErrorMessage(err))
		return
	}
	h.auditCurrentUser(c, "user", "create", "创建用户 "+user.Name, gin.H{"user_id": user.ID, "tenant_id": user.TenantID, "employee_no": user.EmployeeNo})
	item := h.userToJSON(user)
	item["initial_password"] = initialPassword
	response.OK(c, item)
}

func (h *IdentityHandler) UpdateUser(c *gin.Context) {
	tenantID := h.requestTenantID(c)
	var user models.AppUser
	if err := h.tenantScope().ActiveByID(tenantID, c.Param("id")).First(&user).Error; err != nil {
		response.Error(c, 404, response.CodeNotFound, "用户不存在")
		return
	}
	var body struct {
		Name            *string  `json:"name"`
		Phone           *string  `json:"phone"`
		Email           *string  `json:"email"`
		CompanyID       *uint64  `json:"company_id"`
		DepartmentID    *uint64  `json:"department_id"`
		DepartmentIDs   []uint64 `json:"department_ids"`
		PositionIDs     []uint64 `json:"position_ids"`
		RoleIDs         []uint64 `json:"role_ids"`
		Status          *int     `json:"status"`
		IsPlatformAdmin *bool    `json:"is_platform_admin"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, 400, response.CodeBadRequest, "请求参数错误")
		return
	}
	updates := map[string]interface{}{}
	if body.Name != nil {
		updates["name"] = *body.Name
	}
	if body.Phone != nil {
		phone, msg := normalizeOptionalPhone(body.Phone)
		if msg != "" {
			response.Error(c, 400, response.CodeBadRequest, msg)
			return
		}
		updates["phone"] = phone
	}
	if body.Email != nil {
		updates["email"] = *body.Email
	}
	if body.CompanyID != nil {
		updates["company_id"] = *body.CompanyID
	}
	if body.DepartmentID != nil {
		updates["department_id"] = *body.DepartmentID
	}
	if body.Status != nil {
		updates["status"] = *body.Status
		if *body.Status != 1 {
			updates["session_version"] = gorm.Expr("session_version + 1")
			updates["password_changed_at"] = time.Now()
		}
	}
	if body.IsPlatformAdmin != nil {
		viewer, ok := h.currentUser(c)
		if !ok || !viewer.IsPlatformAdmin {
			response.Error(c, 403, response.CodeForbidden, "无权限设置系统管理员")
			return
		}
		if user.IsPlatformAdmin && !*body.IsPlatformAdmin && h.platformAdminCountExcept(user.ID) < 1 {
			response.Error(c, 400, response.CodeBadRequest, "至少保留一名系统管理员")
			return
		}
		updates["is_platform_admin"] = *body.IsPlatformAdmin
	}
	departmentIDs := body.DepartmentIDs
	if departmentIDs != nil {
		departmentIDs = uniqueUint64s(departmentIDs)
		if len(departmentIDs) > 0 {
			updates["department_id"] = departmentIDs[0]
			if body.CompanyID == nil {
				if companyID := h.companyIDForDepartment(tenantID, departmentIDs[0]); companyID != nil {
					updates["company_id"] = *companyID
				}
			}
		} else {
			updates["department_id"] = nil
		}
	} else if body.DepartmentID != nil {
		departmentIDs = []uint64{*body.DepartmentID}
	}
	positionIDs := body.PositionIDs
	if positionIDs != nil {
		positionIDs = uniqueUint64s(positionIDs)
	}
	roleIDs := body.RoleIDs
	if roleIDs != nil {
		roleIDs = uniqueUint64s(roleIDs)
	}
	if err := h.userService().UpdateWithRelations(c.Request.Context(), &user, updates, appuser.Relations{RoleIDs: roleIDs, PositionIDs: positionIDs, DepartmentIDs: departmentIDs}); err != nil {
		response.Error(c, 400, response.CodeBadRequest, safeDBErrorMessage(err))
		return
	}
	if body.Status != nil && *body.Status != 1 {
		h.invalidateSessionsForUser(user.ID)
	}
	h.auditCurrentUser(c, "user", "update", "更新用户 "+user.Name, gin.H{"user_id": user.ID, "tenant_id": user.TenantID, "changes": updates})
	response.OK(c, h.userToJSON(user))
}

func (h *IdentityHandler) userService() *appuser.Service {
	return appuser.NewService(h.db, h)
}
