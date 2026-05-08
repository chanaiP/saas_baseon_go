package handlers

import (
	"bytes"
	"encoding/csv"
	"strings"

	"github.com/gin-gonic/gin"

	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
	"saas_baseon_go/internal/interfaces/http/response"
)

func csvHeaderIndex(headers []string) map[string]int {
	index := map[string]int{}
	for i, header := range headers {
		index[strings.TrimSpace(header)] = i
	}
	return index
}

func csvCell(record []string, index map[string]int, key string) string {
	i, ok := index[key]
	if !ok || i < 0 || i >= len(record) {
		return ""
	}
	return strings.TrimSpace(record[i])
}

func csvSafe(value string) string {
	if strings.HasPrefix(value, "=") || strings.HasPrefix(value, "+") || strings.HasPrefix(value, "-") || strings.HasPrefix(value, "@") {
		return "'" + value
	}
	return value
}

func decodeCSVContent(content []byte) []byte {
	return bytes.TrimPrefix(content, []byte{0xEF, 0xBB, 0xBF})
}

func (h *IdentityHandler) sendCSV(c *gin.Context, filename string, rows [][]string) {
	var buf bytes.Buffer
	buf.Write([]byte{0xEF, 0xBB, 0xBF})
	writer := csv.NewWriter(&buf)
	for _, row := range rows {
		safe := make([]string, len(row))
		for i, cell := range row {
			safe[i] = csvSafe(cell)
		}
		_ = writer.Write(safe)
	}
	writer.Flush()
	c.Header("Content-Disposition", `attachment; filename="`+filename+`"`)
	c.Data(200, "text/csv; charset=utf-8", buf.Bytes())
}

func (h *IdentityHandler) orgNameMaps(tenantID uint64) (map[uint64]string, map[uint64]string) {
	var rows []models.OrgNode
	_ = h.db.Where("tenant_id = ? AND deleted_at IS NULL", tenantID).Find(&rows).Error
	companies := map[uint64]string{}
	depts := map[uint64]string{}
	for _, row := range rows {
		switch row.NodeType {
		case "company":
			companies[row.ID] = row.Name
		case "department":
			depts[row.ID] = row.Name
		}
	}
	return companies, depts
}

func (h *IdentityHandler) resolveCompanyIDByName(tenantID uint64, name string) *uint64 {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil
	}
	var row models.OrgNode
	if err := h.db.Where("tenant_id = ? AND node_type = ? AND name = ? AND deleted_at IS NULL", tenantID, "company", name).First(&row).Error; err != nil {
		return nil
	}
	return &row.ID
}

func (h *IdentityHandler) resolveDepartmentIDByName(tenantID uint64, companyID *uint64, name string) *uint64 {
	name = strings.TrimSpace(name)
	if companyID == nil || name == "" {
		return nil
	}
	var row models.OrgNode
	if err := h.db.Where("tenant_id = ? AND node_type = ? AND company_id = ? AND name = ? AND deleted_at IS NULL", tenantID, "department", *companyID, name).First(&row).Error; err != nil {
		return nil
	}
	return &row.ID
}

func firstStrings(values []string, limit int) []string {
	if limit <= 0 || len(values) <= limit {
		return values
	}
	return values[:limit]
}

func truncateString(value string, limit int) string {
	if limit <= 0 || len(value) <= limit {
		return value
	}
	return value[:limit]
}

func (h *IdentityHandler) exportOrgCSV(c *gin.Context, filename, nodeType string, headers []string) {
	user, ok := h.currentUser(c)
	if !ok {
		response.Error(c, 401, response.CodeUnauthorized, "请先登录")
		return
	}
	if err := h.requireFeatureAccess(user.TenantID, "export_data"); err != nil {
		response.Error(c, 403, response.CodeForbidden, err.Error())
		return
	}
	if err := h.consumeQuota(user.TenantID, "daily_export_times", 1); err != nil {
		response.Error(c, 429, response.CodeBadRequest, err.Error())
		return
	}
	var rows []models.OrgNode
	_ = h.db.Where("tenant_id = ? AND node_type = ? AND deleted_at IS NULL", user.TenantID, nodeType).Order("id asc").Find(&rows).Error
	var all []models.OrgNode
	_ = h.db.Where("tenant_id = ? AND deleted_at IS NULL", user.TenantID).Find(&all).Error
	names := map[uint64]models.OrgNode{}
	for _, row := range all {
		names[row.ID] = row
	}
	out := [][]string{headers}
	for _, row := range rows {
		parentName := ""
		if row.ParentID != nil {
			parentName = names[*row.ParentID].Name
		}
		status := "启用"
		if row.Status != 1 {
			status = "停用"
		}
		if nodeType == "company" {
			out = append(out, []string{row.Name, derefString(row.Code), derefString(row.CompanyType), parentName, status})
		} else {
			companyName := ""
			if row.CompanyID != nil {
				companyName = names[*row.CompanyID].Name
			}
			if parentName != "" && names[valueOrZero(row.ParentID)].NodeType != "department" {
				parentName = ""
			}
			out = append(out, []string{row.Name, derefString(row.Code), companyName, parentName, status})
		}
	}
	h.sendCSV(c, filename, out)
}
