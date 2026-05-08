package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"saas_baseon_go/internal/interfaces/http/response"
)

type deletionBlockedError struct {
	Message string
}

func (e *deletionBlockedError) Error() string { return e.Message }

func (h *IdentityHandler) blockDeleteIfReferenced(c *gin.Context, resource string, refs ...deletionReference) bool {
	if err := checkDeletionReferences(h.db, resource, refs...); err != nil {
		h.respondDeletionError(c, err)
		return true
	}
	return false
}

func checkDeletionReferences(db *gorm.DB, resource string, refs ...deletionReference) error {
	for _, item := range refs {
		var count int64
		if err := db.Model(item.Model).Where(item.Query, item.Args...).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return &deletionBlockedError{Message: fmt.Sprintf("%s已被%s引用，不能删除", resource, item.Name)}
		}
	}
	return nil
}

func (h *IdentityHandler) respondDeletionError(c *gin.Context, err error) {
	var blocked *deletionBlockedError
	if errors.As(err, &blocked) {
		response.Error(c, 400, response.CodeBadRequest, blocked.Message)
		return
	}
	response.Error(c, 400, response.CodeBadRequest, safeDBErrorMessage(err))
}

func requiredDeletionGuards() []string {
	return []string{
		"permission:role_permission",
		"tenant:app_user",
		"plan:tenant_subscription",
		"org_node:app_user",
		"org_node:business_unit_org_map",
		"position_type:position",
		"position:app_user_position",
		"business_unit:business_unit_org_map",
		"business_unit:business_unit_scope",
		"dict_type:dict_item",
		"dict_item:tenant_dict_item_override",
		"sys_param:tenant_param_value",
	}
}

func (h *IdentityHandler) deleteByID(c *gin.Context, model interface{}) bool {
	id := parseUintParam(c, "id")
	if err := h.db.Delete(model, id).Error; err != nil {
		response.Error(c, 400, response.CodeBadRequest, err.Error())
		return false
	}
	response.OK(c, gin.H{"deleted": 1, "id": id})
	return true
}

func (h *IdentityHandler) deleteTenantScopedByID(c *gin.Context, model interface{}) bool {
	id := parseUintParam(c, "id")
	result := h.db.Where("tenant_id = ?", h.requestTenantID(c)).Delete(model, id)
	if result.Error != nil {
		response.Error(c, 400, response.CodeBadRequest, result.Error.Error())
		return false
	}
	if result.RowsAffected == 0 {
		response.Error(c, 404, response.CodeNotFound, "数据不存在")
		return false
	}
	response.OK(c, gin.H{"deleted": 1, "id": id})
	return true
}

func nullableTrimmed(value *string) *string {
	if value == nil {
		return nil
	}
	return nullableFromString(*value)
}

func trimmedStringPtr(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	return &trimmed
}

func nullableFromString(value string) *string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func normalizeOptionalPhone(value *string) (*string, string) {
	if value == nil {
		return nil, ""
	}
	raw := strings.TrimSpace(*value)
	if raw == "" {
		return nil, ""
	}
	var digits strings.Builder
	for _, ch := range raw {
		switch {
		case ch >= '0' && ch <= '9':
			digits.WriteRune(ch)
		case ch == ' ' || ch == '-' || ch == '(' || ch == ')':
		default:
			return nil, "手机号只能包含数字、空格、短横线或括号"
		}
	}
	normalized := digits.String()
	if len(normalized) < 10 || len(normalized) > 15 {
		return nil, "手机号需为 10-15 位数字"
	}
	return &normalized, ""
}

func uniqueUint64s(values []uint64) []uint64 {
	if values == nil {
		return nil
	}
	seen := map[uint64]struct{}{}
	out := make([]uint64, 0, len(values))
	for _, value := range values {
		if value == 0 {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}

func normalizedUserDepartmentIDs(departmentID *uint64, departmentIDs []uint64) []uint64 {
	ids := uniqueUint64s(departmentIDs)
	if departmentID == nil || *departmentID == 0 {
		return ids
	}
	for _, id := range ids {
		if id == *departmentID {
			return ids
		}
	}
	return append([]uint64{*departmentID}, ids...)
}

func tombstoneUniqueValue(value string, id uint64, maxLen int) string {
	suffix := fmt.Sprintf("__deleted_%d_%d", id, time.Now().Unix())
	base := value
	if len(base)+len(suffix) > maxLen {
		keep := maxLen - len(suffix)
		if keep < 0 {
			keep = 0
		}
		if len(base) > keep {
			base = base[:keep]
		}
	}
	return base + suffix
}

func derefString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func valueOrZero(value *uint64) uint64 {
	if value == nil {
		return 0
	}
	return *value
}

func randomHex(size int) string {
	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		return strconv.FormatInt(time.Now().UnixNano(), 16)
	}
	return hex.EncodeToString(buf)
}

func randomCode(size int) string {
	const chars = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		return "ABCD"
	}
	out := make([]byte, size)
	for i, b := range buf {
		out[i] = chars[int(b)%len(chars)]
	}
	return string(out)
}
