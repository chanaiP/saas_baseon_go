package handlers

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
	"saas_baseon_go/internal/interfaces/http/response"
)

func (h *IdentityHandler) ResetUserPassword(c *gin.Context) {
	targetUserID := parseUintParam(c, "id")
	if message, blocked := h.rateLimitExceeded(c, fmt.Sprintf("rate:reset_password:ip:%s", c.ClientIP()), 10, 10*time.Minute); blocked {
		c.JSON(429, response.Body{Code: 42900, Message: message})
		return
	}
	if message, blocked := h.rateLimitExceeded(c, fmt.Sprintf("rate:reset_password:user:%d", targetUserID), 5, 10*time.Minute); blocked {
		c.JSON(429, response.Body{Code: 42900, Message: message})
		return
	}
	newPassword := generateRandomPassword(14)
	result := h.db.Model(&models.AppUser{}).Where("id = ? AND tenant_id = ? AND deleted_at IS NULL", targetUserID, h.requestTenantID(c)).Updates(map[string]interface{}{
		"password_hash":       devPasswordHash(newPassword),
		"session_version":     gorm.Expr("session_version + 1"),
		"password_changed_at": time.Now(),
	})
	if result.Error != nil {
		respondBadRequest(c, result.Error)
		return
	}
	if result.RowsAffected == 0 {
		response.Error(c, 404, response.CodeNotFound, "用户不存在")
		return
	}
	h.invalidateSessionsForUser(targetUserID)
	h.auditCurrentUser(c, "user", "password_reset", "重置用户密码", gin.H{"user_id": targetUserID})
	response.OK(c, gin.H{"new_password": newPassword})
}

func (h *IdentityHandler) DeleteUser(c *gin.Context) {
	id := parseUintParam(c, "id")
	viewer, ok := h.currentUser(c)
	if ok && viewer.ID == id {
		response.Error(c, 400, response.CodeBadRequest, "不能删除当前登录用户")
		return
	}
	tenantID := h.requestTenantID(c)
	var user models.AppUser
	if err := h.tenantScope().ActiveByID(tenantID, id).First(&user).Error; err != nil {
		response.Error(c, 404, response.CodeNotFound, "用户不存在")
		return
	}
	if h.userIsInitialSuperAdmin(user) {
		response.Error(c, 400, response.CodeBadRequest, "初始超级管理员不能删除")
		return
	}
	now := time.Now()
	updates := map[string]interface{}{"deleted_at": now, "status": 0, "employee_no": tombstoneUniqueValue(user.EmployeeNo, user.ID, 64), "account": tombstoneUniqueValue(user.Account, user.ID, 64), "session_version": gorm.Expr("session_version + 1"), "password_changed_at": now}
	if user.Phone != nil {
		updates["phone"] = tombstoneUniqueValue(*user.Phone, user.ID, 32)
	}
	if err := h.userService().Delete(c.Request.Context(), &user, updates); err != nil {
		response.Error(c, 400, response.CodeBadRequest, safeDBErrorMessage(err))
		return
	}
	h.invalidateSessionsForUser(user.ID)
	h.auditCurrentUser(c, "user", "delete", "删除用户 "+user.Name, gin.H{"user_id": id})
	response.OK(c, gin.H{"deleted": id})
}
