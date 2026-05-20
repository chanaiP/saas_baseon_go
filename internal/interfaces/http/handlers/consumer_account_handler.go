package handlers

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
	"saas_baseon_go/internal/interfaces/http/response"
)

func (h *IdentityHandler) ConsumerRegister(c *gin.Context) {
	var body struct {
		Phone    *string `json:"phone"`
		Account  string  `json:"account"`
		Password string  `json:"password"`
		Name     string  `json:"name"`
		Email    *string `json:"email"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, 400, response.CodeBadRequest, "请求参数错误")
		return
	}
	phone, msg := normalizeOptionalPhone(body.Phone)
	if msg != "" {
		response.Error(c, 400, response.CodeBadRequest, msg)
		return
	}
	loginAccount := strings.TrimSpace(body.Account)
	if loginAccount == "" && phone != nil {
		loginAccount = *phone
	}
	if loginAccount == "" {
		response.Error(c, 400, response.CodeBadRequest, "登录账号或手机号不能为空")
		return
	}
	if strings.TrimSpace(body.Password) == "" {
		response.Error(c, 400, response.CodeBadRequest, "密码不能为空")
		return
	}
	displayName := strings.TrimSpace(body.Name)
	if displayName == "" {
		displayName = loginAccount
	}
	hashed, err := hashPassword(body.Password)
	if err != nil {
		response.Error(c, 500, response.CodeInternal, "密码加密失败")
		return
	}
	var user models.AppUser
	err = h.db.Transaction(func(tx *gorm.DB) error {
		now := time.Now()
		account := models.Account{
			LoginAccount: loginAccount,
			Phone:        phone,
			Email:        nullableTrimmed(body.Email),
			PasswordHash: hashed,
			Status:       1,
			CreatedAt:    now,
			UpdatedAt:    now,
		}
		if err := tx.Create(&account).Error; err != nil {
			return err
		}
		identity := models.UserIdentity{
			AccountID:   account.ID,
			DisplayName: displayName,
			Phone:       phone,
			Email:       nullableTrimmed(body.Email),
			Status:      1,
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		if err := tx.Create(&identity).Error; err != nil {
			return err
		}
		tenant := models.Tenant{
			Code:       "personal_" + randomHex(8),
			Name:       displayName + "的个人空间",
			Status:     1,
			TenantType: TenantTypePersonal,
			CreatedAt:  now,
			UpdatedAt:  now,
		}
		if err := tx.Create(&tenant).Error; err != nil {
			return err
		}
		companyCode := "personal"
		company := models.OrgNode{TenantID: tenant.ID, NodeType: "company", Name: "个人空间", Code: &companyCode, Status: 1, CreatedAt: now, UpdatedAt: now}
		if err := tx.Create(&company).Error; err != nil {
			return err
		}
		user = models.AppUser{
			TenantID:       tenant.ID,
			AccountID:      &account.ID,
			IdentityUserID: &identity.ID,
			MemberType:     "personal_owner",
			CompanyID:      &company.ID,
			EmployeeNo:     loginAccount,
			Account:        loginAccount,
			PasswordHash:   hashed,
			Name:           displayName,
			Phone:          phone,
			Email:          nullableTrimmed(body.Email),
			Status:         1,
			IsTenantAdmin:  true,
			CreatedAt:      now,
			UpdatedAt:      now,
		}
		if err := tx.Create(&user).Error; err != nil {
			return err
		}
		if err := tx.Model(&tenant).Update("owner_user_id", user.ID).Error; err != nil {
			return err
		}
		role := models.Role{TenantID: tenant.ID, Code: "personal_owner", Name: "个人空间所有者", Status: 1, CreatedAt: now, UpdatedAt: now}
		if err := tx.Create(&role).Error; err != nil {
			return err
		}
		if err := tx.Create(&models.UserRole{UserID: user.ID, RoleID: role.ID, CreatedAt: now}).Error; err != nil {
			return err
		}
		var permissions []models.Permission
		if err := tx.Where("enabled = ? AND deleted_at IS NULL AND tenant_scope IN ?", true, []string{TenantScopeAll, TenantScopeEnterprisePersonal, TenantScopePersonalOnly, TenantScopePlatformPersonal}).
			Where("perm_type IN ?", []int{2, 3}).
			Find(&permissions).Error; err != nil {
			return err
		}
		for _, permission := range permissions {
			if !tenantScopeAllows(normalizeTenantScope(permission.TenantScope, permission.IsPlatformOnly), TenantTypePersonal) {
				continue
			}
			if err := tx.Create(&models.RolePermission{RoleID: role.ID, PermissionID: permission.ID, CreatedAt: now}).Error; err != nil {
				return err
			}
		}
		plan, err := ensurePersonalPlan(tx, now)
		if err != nil {
			return err
		}
		if err := ensurePersonalBasicPlan(tx, now); err != nil {
			return err
		}
		sub := models.TenantSubscription{TenantID: tenant.ID, PlanID: plan.ID, SubscriptionStatus: "ACTIVE", StartTime: now, AutoRenew: true, Source: "SYSTEM", SourceRef: nullableFromString("consumer_register"), CreatedAt: now, UpdatedAt: now}
		return tx.Create(&sub).Error
	})
	if err != nil {
		respondBadRequest(c, err)
		return
	}
	token, err := h.issueLoginToken(user)
	if err != nil {
		response.Error(c, 500, response.CodeInternal, "令牌生成失败")
		return
	}
	h.recordLogin(c, user.Account, &user.ID, &user.TenantID, true, "C端注册登录")
	response.OK(c, gin.H{"token": token, "token_type": "bearer", "tenant_id": user.TenantID, "user_id": user.ID})
}

func ensurePersonalPlan(tx *gorm.DB, now time.Time) (models.SaasPlan, error) {
	var plan models.SaasPlan
	err := tx.Where("plan_code = ? AND deleted_at IS NULL", "PERSONAL_FREE").First(&plan).Error
	if err == nil {
		return plan, ensurePersonalPlanQuotas(tx, plan.ID, now)
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return plan, err
	}
	plan = models.SaasPlan{PlanCode: "PERSONAL_FREE", PlanName: "个人免费版", PlanType: "PERSONAL", BillingCycle: "MONTH", Status: 1, IsDefault: false, SortOrder: 5, Description: nullableFromString("C端个人空间默认套餐"), CreatedAt: now, UpdatedAt: now}
	if err := tx.Create(&plan).Error; err != nil {
		return plan, err
	}
	return plan, ensurePersonalPlanQuotas(tx, plan.ID, now)
}

func ensurePersonalBasicPlan(tx *gorm.DB, now time.Time) error {
	var plan models.SaasPlan
	err := tx.Where("plan_code = ? AND deleted_at IS NULL", "PERSONAL_BASIC").First(&plan).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		plan = models.SaasPlan{PlanCode: "PERSONAL_BASIC", PlanName: "个人基础版", PlanType: "PERSONAL", BillingCycle: "MONTH", Status: 1, IsDefault: false, SortOrder: 6, Description: nullableFromString("C端个人空间基础套餐"), CreatedAt: now, UpdatedAt: now}
		if err := tx.Create(&plan).Error; err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	return ensurePersonalBasicPlanQuotas(tx, plan.ID, now)
}

func ensurePersonalPlanQuotas(tx *gorm.DB, planID uint64, now time.Time) error {
	quotas := []struct {
		code       string
		name       string
		quotaType  string
		periodType *string
		unit       *string
		value      int
	}{
		{code: "max_users", name: "最大用户数", quotaType: "STATIC", periodType: nullableFromString("NONE"), unit: nullableFromString("COUNT"), value: 1},
		{code: "daily_import_times", name: "每日导入次数", quotaType: "DYNAMIC", periodType: nullableFromString("DAY"), unit: nullableFromString("TIMES"), value: 3},
		{code: "daily_ai_calls", name: "每日 AI 调用次数", quotaType: "DYNAMIC", periodType: nullableFromString("DAY"), unit: nullableFromString("TIMES"), value: 20},
		{code: "max_storage_gb", name: "存储空间", quotaType: "STATIC", periodType: nullableFromString("NONE"), unit: nullableFromString("GB"), value: 1},
		{code: "max_file_size_mb", name: "单文件大小", quotaType: "STATIC", periodType: nullableFromString("NONE"), unit: nullableFromString("MB"), value: 10},
		{code: "max_tasks", name: "最大任务数量", quotaType: "STATIC", periodType: nullableFromString("NONE"), unit: nullableFromString("COUNT"), value: 50},
	}
	for _, item := range quotas {
		quota := models.SaasQuota{QuotaCode: item.code, QuotaName: item.name, QuotaType: item.quotaType, PeriodType: item.periodType, Unit: item.unit, Status: 1, CreatedAt: now, UpdatedAt: now}
		if err := tx.Where("quota_code = ?", item.code).FirstOrCreate(&quota).Error; err != nil {
			return err
		}
		link := models.SaasPlanQuota{PlanID: planID, QuotaID: quota.ID, QuotaValue: item.value, CreatedAt: now, UpdatedAt: now}
		if err := tx.Where("plan_id = ? AND quota_id = ?", planID, quota.ID).FirstOrCreate(&link).Error; err != nil {
			return err
		}
		if err := tx.Model(&link).Updates(map[string]interface{}{"quota_value": item.value, "updated_at": now}).Error; err != nil {
			return err
		}
	}
	return nil
}

func ensurePersonalBasicPlanQuotas(tx *gorm.DB, planID uint64, now time.Time) error {
	values := map[string]int{"max_users": 1, "daily_import_times": 20, "daily_ai_calls": 200, "max_storage_gb": 10, "max_file_size_mb": 100, "max_tasks": 500}
	for code, value := range values {
		var quota models.SaasQuota
		if err := tx.Where("quota_code = ? AND status = ?", code, 1).First(&quota).Error; err != nil {
			return err
		}
		link := models.SaasPlanQuota{PlanID: planID, QuotaID: quota.ID, QuotaValue: value, CreatedAt: now, UpdatedAt: now}
		if err := tx.Where("plan_id = ? AND quota_id = ?", planID, quota.ID).FirstOrCreate(&link).Error; err != nil {
			return err
		}
		if err := tx.Model(&link).Updates(map[string]interface{}{"quota_value": value, "updated_at": now}).Error; err != nil {
			return err
		}
	}
	return nil
}

func (h *IdentityHandler) loginWithAccount(c *gin.Context, accountRaw string, password string, tenantID *uint64, tenantCode string) bool {
	var account models.Account
	accountValue := strings.TrimSpace(accountRaw)
	if accountValue == "" {
		return false
	}
	err := h.db.Where("deleted_at IS NULL").Where("login_account = ? OR phone = ? OR email = ?", accountValue, accountValue, accountValue).Order("id asc").First(&account).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false
	}
	if missingAccountTable(err) {
		return false
	}
	if err != nil {
		response.Error(c, 500, response.CodeInternal, "登录失败")
		return true
	}
	if account.Status != 1 {
		h.recordLoginFailure(c, accountValue, nil, nil, "账号已停用")
		response.Error(c, 400, response.CodeBadRequest, "账号已停用")
		return true
	}
	if !verifyPassword(password, account.PasswordHash) {
		h.recordLoginFailure(c, accountValue, nil, nil, "账号或密码错误")
		response.Error(c, 400, response.CodeBadRequest, "账号或密码错误")
		return true
	}
	query := h.db.Where("app_user.account_id = ? AND app_user.deleted_at IS NULL", account.ID)
	if tenantID != nil {
		query = query.Where("app_user.tenant_id = ?", *tenantID)
	}
	if strings.TrimSpace(tenantCode) != "" {
		query = query.Joins("JOIN tenant t ON t.id = app_user.tenant_id AND t.code = ?", strings.TrimSpace(tenantCode))
	}
	var candidates []models.AppUser
	if err := query.Order("app_user.is_platform_admin desc, app_user.id asc").Find(&candidates).Error; err != nil {
		response.Error(c, 500, response.CodeInternal, "登录失败")
		return true
	}
	if len(candidates) == 0 {
		h.recordLoginFailure(c, accountValue, nil, nil, "账号或密码错误")
		response.Error(c, 400, response.CodeBadRequest, "账号或密码错误")
		return true
	}
	if len(candidates) > 1 && tenantID == nil && strings.TrimSpace(tenantCode) == "" {
		c.JSON(200, response.Body{Code: 2, Message: "请选择空间", Data: gin.H{"token": nil, "token_type": "bearer", "captcha_required": false, "tenants": h.activeLoginTenantOptions(candidates)}})
		return true
	}
	if h.finishLogin(c, accountValue, candidates[0]) {
		return true
	}
	return true
}

func (h *IdentityHandler) finishLogin(c *gin.Context, account string, user models.AppUser) bool {
	if user.Status != 1 {
		h.recordLoginFailure(c, account, &user.ID, &user.TenantID, "账号已停用")
		response.Error(c, 400, response.CodeBadRequest, "账号已停用")
		return false
	}
	var tenant models.Tenant
	if err := h.db.First(&tenant, user.TenantID).Error; err == nil && tenant.Status != 1 {
		h.recordLoginFailure(c, account, &user.ID, &user.TenantID, "所属空间已停用")
		response.Error(c, 400, response.CodeBadRequest, "所属空间已停用，请联系管理员")
		return false
	}
	if !h.subscriptionAllowsLogin(user.TenantID) {
		h.recordLoginFailure(c, account, &user.ID, &user.TenantID, "账户已到期")
		response.Error(c, 400, response.CodeBadRequest, "账户已到期，请联系管理员续费")
		return false
	}
	token, err := h.issueLoginToken(user)
	if err != nil {
		response.Error(c, 500, response.CodeInternal, "令牌生成失败")
		return false
	}
	h.resetLoginFail(c, account)
	h.resetIPLoginFail(c)
	h.recordLogin(c, account, &user.ID, &user.TenantID, true, "登录成功")
	response.OK(c, gin.H{"token": token, "token_type": "bearer", "captcha_required": false})
	return true
}

func accountBackfillKey(phone *string, account string) string {
	if phone != nil && strings.TrimSpace(*phone) != "" {
		return strings.TrimSpace(*phone)
	}
	return strings.TrimSpace(account)
}

func (h *IdentityHandler) backfillAccountForLegacyUser(user *models.AppUser, password string) {
	if user == nil || user.ID == 0 || user.AccountID != nil {
		return
	}
	loginAccount := accountBackfillKey(user.Phone, user.Account)
	if loginAccount == "" || password == "" {
		return
	}
	_ = h.db.Transaction(func(tx *gorm.DB) error {
		now := time.Now()
		var account models.Account
		err := tx.Where("deleted_at IS NULL").Where("login_account = ? OR phone = ?", loginAccount, loginAccount).First(&account).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			account = models.Account{LoginAccount: loginAccount, Phone: user.Phone, Email: user.Email, PasswordHash: user.PasswordHash, Status: 1, CreatedAt: now, UpdatedAt: now}
			if err := tx.Create(&account).Error; err != nil {
				return err
			}
		} else if err != nil {
			return err
		}
		var identity models.UserIdentity
		err = tx.Where("account_id = ? AND deleted_at IS NULL", account.ID).Order("id asc").First(&identity).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			identity = models.UserIdentity{AccountID: account.ID, DisplayName: user.Name, AvatarURL: user.AvatarURL, Phone: user.Phone, Email: user.Email, Status: 1, CreatedAt: now, UpdatedAt: now}
			if err := tx.Create(&identity).Error; err != nil {
				return err
			}
		} else if err != nil {
			return err
		}
		user.AccountID = &account.ID
		user.IdentityUserID = &identity.ID
		return tx.Model(user).Updates(map[string]interface{}{"account_id": account.ID, "identity_user_id": identity.ID}).Error
	})
}

func (h *IdentityHandler) accountIdentityForEnterpriseMember(phone *string, email *string) (*uint64, *uint64) {
	if h == nil || h.db == nil {
		return nil, nil
	}
	phoneValue := ""
	if phone != nil {
		phoneValue = strings.TrimSpace(*phone)
	}
	emailValue := ""
	if email != nil {
		emailValue = strings.TrimSpace(*email)
	}
	if phoneValue == "" && emailValue == "" {
		return nil, nil
	}
	var account models.Account
	query := h.db.Where("deleted_at IS NULL")
	if phoneValue != "" && emailValue != "" {
		query = query.Where("phone = ? OR email = ?", phoneValue, emailValue)
	} else if phoneValue != "" {
		query = query.Where("phone = ?", phoneValue)
	} else {
		query = query.Where("email = ?", emailValue)
	}
	if err := query.Order("id asc").First(&account).Error; err != nil {
		return nil, nil
	}
	var identity models.UserIdentity
	if err := h.db.Where("account_id = ? AND deleted_at IS NULL", account.ID).Order("id asc").First(&identity).Error; err != nil {
		return &account.ID, nil
	}
	return &account.ID, &identity.ID
}

func consumerRegisterAuditDetail(user models.AppUser) string {
	return fmt.Sprintf("consumer_register:user=%d tenant=%d", user.ID, user.TenantID)
}

func missingAccountTable(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "no such table: account") || strings.Contains(message, "relation \"account\" does not exist")
}
