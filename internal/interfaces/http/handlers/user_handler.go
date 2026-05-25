package handlers

import (
	"errors"
	"fmt"
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
	accountID, identityID := h.accountIdentityForEnterpriseMember(phone, body.Email)
	user := models.AppUser{TenantID: tenantID, EmployeeNo: body.EmployeeNo, Account: body.EmployeeNo, PasswordHash: devPasswordHash(initialPassword), Name: body.Name, Phone: phone, Email: body.Email, CompanyID: companyID, DepartmentID: departmentID, Status: body.Status, AccountID: accountID, IdentityUserID: identityID, MemberType: "enterprise_member"}
	var err error
	user, err = h.userService().CreateWithRelations(c.Request.Context(), user, appuser.Relations{RoleIDs: uniqueUint64s(body.RoleIDs), PositionIDs: uniqueUint64s(body.PositionIDs), DepartmentIDs: departmentIDs})
	if err != nil {
		response.Error(c, 400, response.CodeBadRequest, safeUserErrorMessage(err))
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
	if h.userIsInitialSuperAdmin(user) {
		response.Error(c, 400, response.CodeBadRequest, "初始超级管理员不能编辑")
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
		response.Error(c, 400, response.CodeBadRequest, safeUserErrorMessage(err))
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

func safeUserErrorMessage(err error) string {
	if err == nil {
		return ""
	}
	var quotaErr *quotaExceededError
	if errors.As(err, &quotaErr) {
		return userQuotaExceededMessage(quotaErr.QuotaName, quotaErr.Used, quotaErr.Limit)
	}
	raw := strings.TrimSpace(err.Error())
	lower := strings.ToLower(raw)
	if strings.Contains(lower, "duplicate key") || strings.Contains(lower, "unique constraint") {
		switch {
		case strings.Contains(lower, "employee_no") || strings.Contains(lower, "idx_app_user_tenant_employee"):
			return "工号已存在"
		case strings.Contains(lower, "phone") || strings.Contains(lower, "idx_app_user_tenant_phone"):
			return "手机号已存在"
		default:
			return "数据已存在，请检查唯一字段"
		}
	}
	for _, message := range []string{
		"工号和姓名不能为空",
		"工号已存在",
		"手机号已存在",
		"手机号格式不正确",
		"组织节点不存在或不属于当前主体",
		"岗位不存在或不属于当前主体",
		"角色不存在或不属于当前主体",
		"配额不足",
	} {
		if raw == message || strings.Contains(raw, message) {
			return raw
		}
	}
	if strings.Contains(raw, "已超出套餐配额") {
		return userQuotaExceededMessage(strings.TrimSuffix(raw, "已超出套餐配额"), 0, 0)
	}
	return safeDBErrorMessage(err)
}

func userQuotaExceededMessage(quotaName string, used int, limit int) string {
	name := strings.TrimSpace(quotaName)
	if name == "" {
		name = "用户数"
	}
	if used > 0 || limit > 0 {
		return fmt.Sprintf("%s已达到平台配额上限（当前已用 %d / 上限 %d），请调整主体套餐配额或停用不需要的用户后再新增", name, used, limit)
	}
	return name + "已达到平台配额上限，请调整主体套餐配额或停用不需要的用户后再新增"
}
