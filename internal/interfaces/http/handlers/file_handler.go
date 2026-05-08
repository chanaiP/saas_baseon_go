package handlers

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"time"
	"unicode/utf8"

	"github.com/gin-gonic/gin"

	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
	"saas_baseon_go/internal/interfaces/http/response"
)

var safeFileIDPattern = regexp.MustCompile(`^[0-9a-f]{32}$`)
var errInvalidFileID = errors.New("invalid file_id")
var allowedUploadExtensions = map[string]struct{}{
	".csv": {}, ".doc": {}, ".docx": {}, ".gif": {}, ".jpeg": {}, ".jpg": {}, ".pdf": {}, ".png": {},
	".ppt": {}, ".pptx": {}, ".txt": {}, ".webp": {}, ".xls": {}, ".xlsx": {}, ".zip": {},
}

func (h *IdentityHandler) UploadFile(c *gin.Context) {
	user, ok := h.currentUser(c)
	if !ok {
		response.Error(c, 401, response.CodeUnauthorized, "请先登录")
		return
	}
	if err := h.requireFeatureAccess(user.TenantID, "file_manage"); err != nil {
		response.Error(c, 403, response.CodeForbidden, err.Error())
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 50*1024*1024+1024)
	file, err := c.FormFile("file")
	if err != nil {
		response.Error(c, 400, response.CodeBadRequest, "请选择文件")
		return
	}
	maxSize := int64(50 * 1024 * 1024)
	if quota, ok := h.currentQuotaLimitByCode(user.TenantID, "max_file_size_mb"); ok {
		if quota == 0 {
			response.Error(c, 403, response.CodeForbidden, "当前套餐不支持文件上传")
			return
		}
		if quota > 0 {
			quotaSize := int64(quota) * 1024 * 1024
			if quotaSize < maxSize {
				maxSize = quotaSize
			}
		}
	}
	if file.Size > maxSize {
		response.Error(c, 413, response.CodeBadRequest, fmt.Sprintf("文件过大（最大 %d MB）", maxSize/(1024*1024)))
		return
	}
	if quota, ok := h.currentQuotaLimitByCode(user.TenantID, "max_storage_gb"); ok {
		if quota == 0 {
			response.Error(c, 403, response.CodeForbidden, "当前套餐不支持文件存储")
			return
		}
		if quota > 0 {
			used, err := tenantStorageBytes(user.TenantID)
			if err != nil {
				response.Error(c, 500, response.CodeInternal, "读取存储用量失败")
				return
			}
			if used+file.Size > int64(quota)*1024*1024*1024 {
				response.Error(c, 429, response.CodeBadRequest, "租户存储空间不足")
				return
			}
		}
	}
	fileID := randomHex(16)
	originalName := safeOriginalName(file.Filename)
	ext, err := validateUploadFile(originalName, file)
	if err != nil {
		response.Error(c, 400, response.CodeBadRequest, err.Error())
		return
	}
	dir := uploadDir(user.TenantID, time.Now())
	if err := os.MkdirAll(dir, 0o750); err != nil {
		response.Error(c, 500, response.CodeInternal, "创建上传目录失败")
		return
	}
	dst := filepath.Join(dir, fileID+ext)
	if err := saveUploadedFile(file, dst); err != nil {
		_ = os.Remove(dst)
		response.Error(c, 500, response.CodeInternal, "保存文件失败")
		return
	}
	now := time.Now()
	storedName := fileID + ext
	meta := models.FileObject{
		TenantID:     user.TenantID,
		FileID:       fileID,
		CreatedBy:    user.ID,
		OriginalName: originalName,
		StoredName:   storedName,
		StoragePath:  dst,
		MimeType:     normalizeContentType(file.Header.Get("Content-Type")),
		FileSize:     file.Size,
		Status:       1,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := h.db.Create(&meta).Error; err != nil {
		_ = os.Remove(dst)
		response.Error(c, 500, response.CodeInternal, "保存文件元数据失败")
		return
	}
	h.audit(c, user.TenantID, user.ID, "file", "upload", "上传文件 "+originalName, gin.H{"file_id": fileID, "file_name": originalName, "size": file.Size})
	response.OK(c, gin.H{"file_id": fileID, "file_name": originalName, "size": file.Size, "url": "/api/files/download/" + fileID})
}

func (h *IdentityHandler) DownloadFile(c *gin.Context) {
	user, ok := h.currentUser(c)
	if !ok {
		response.Error(c, 401, response.CodeUnauthorized, "请先登录")
		return
	}
	fileID := c.Param("file_id")
	fileObject, err := h.findTenantFileObject(user, fileID, "file:download")
	if err != nil {
		if errors.Is(err, errInvalidFileID) {
			response.Error(c, 400, response.CodeBadRequest, "file_id 非法")
			return
		}
		response.Error(c, 404, response.CodeNotFound, "文件不存在")
		return
	}
	h.audit(c, user.TenantID, user.ID, "file", "download", "下载文件 "+fileID, gin.H{"file_id": fileID, "file_name": fileObject.OriginalName, "owner_id": fileObject.CreatedBy})
	c.FileAttachment(fileObject.StoragePath, fileObject.OriginalName)
}

func (h *IdentityHandler) DeleteFile(c *gin.Context) {
	user, ok := h.currentUser(c)
	if !ok {
		response.Error(c, 401, response.CodeUnauthorized, "请先登录")
		return
	}
	fileID := c.Param("file_id")
	fileObject, err := h.findTenantFileObject(user, fileID, "file:delete")
	if err != nil {
		if errors.Is(err, errInvalidFileID) {
			response.Error(c, 400, response.CodeBadRequest, "file_id 非法")
			return
		}
		response.Error(c, 404, response.CodeNotFound, "文件不存在")
		return
	}
	trash := filepath.Join(uploadRoot(), strconv.FormatUint(user.TenantID, 10), ".trash", time.Now().Format("2006/01/02"))
	_ = os.MkdirAll(trash, 0o750)
	trashPath := filepath.Join(trash, time.Now().Format("150405")+"_"+filepath.Base(fileObject.StoragePath))
	if err := os.Rename(fileObject.StoragePath, trashPath); err != nil {
		response.Error(c, 500, response.CodeInternal, "删除文件失败")
		return
	}
	now := time.Now()
	_ = h.db.Model(&fileObject).Updates(map[string]interface{}{"status": 0, "deleted_at": now, "updated_at": now, "storage_path": trashPath}).Error
	h.audit(c, user.TenantID, user.ID, "file", "delete", "删除文件 "+fileID, gin.H{"file_id": fileID, "trash_path": trashPath})
	response.OK(c, gin.H{"message": "已删除"})
}

func (h *IdentityHandler) ExportUsersCSV(c *gin.Context) {
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
	var users []models.AppUser
	query := h.db.Where("tenant_id = ? AND deleted_at IS NULL", user.TenantID)
	query = h.dataScopeService().ApplyToUserQuery(query, user)
	_ = query.Order("id asc").Find(&users).Error
	companyNames, deptNames := h.orgNameMaps(user.TenantID)
	rows := make([][]string, 0, len(users)+1)
	rows = append(rows, []string{"employee_no", "name", "phone", "email", "company_name", "department_name", "status"})
	for _, row := range users {
		status := "启用"
		if row.Status != 1 {
			status = "停用"
		}
		rows = append(rows, []string{row.EmployeeNo, row.Name, derefString(row.Phone), derefString(row.Email), companyNames[valueOrZero(row.CompanyID)], deptNames[valueOrZero(row.DepartmentID)], status})
	}
	h.sendCSV(c, "users_export.csv", rows)
}

func (h *IdentityHandler) ImportUsersCSV(c *gin.Context) {
	user, ok := h.currentUser(c)
	if !ok {
		response.Error(c, 401, response.CodeUnauthorized, "请先登录")
		return
	}
	if err := h.requireFeatureAccess(user.TenantID, "import_data"); err != nil {
		response.Error(c, 403, response.CodeForbidden, err.Error())
		return
	}
	if err := h.consumeQuota(user.TenantID, "daily_import_times", 1); err != nil {
		response.Error(c, 429, response.CodeBadRequest, err.Error())
		return
	}
	file, err := c.FormFile("file")
	if err != nil {
		response.Error(c, 400, response.CodeBadRequest, "请选择 CSV 文件")
		return
	}
	if file.Size > 5*1024*1024 {
		response.Error(c, 413, response.CodeBadRequest, "文件过大（最大 5 MB）")
		return
	}
	opened, err := file.Open()
	if err != nil {
		response.Error(c, 400, response.CodeBadRequest, "读取文件失败")
		return
	}
	defer opened.Close()
	content, _ := io.ReadAll(opened)
	content = decodeCSVContent(content)
	if !utf8.Valid(content) {
		response.Error(c, 400, response.CodeBadRequest, "CSV 格式错误")
		return
	}
	reader := csv.NewReader(bytes.NewReader(content))
	records, err := reader.ReadAll()
	if err != nil || len(records) == 0 {
		response.Error(c, 400, response.CodeBadRequest, "CSV 格式错误")
		return
	}
	index := csvHeaderIndex(records[0])
	created, skipped := 0, 0
	errors := []string{}
	for line, record := range records[1:] {
		employeeNo := csvCell(record, index, "employee_no")
		name := csvCell(record, index, "name")
		if employeeNo == "" || name == "" {
			errors = append(errors, fmt.Sprintf("第 %d 行：工号和姓名不能为空", line+2))
			continue
		}
		var count int64
		h.db.Model(&models.AppUser{}).Where("tenant_id = ? AND employee_no = ? AND deleted_at IS NULL", user.TenantID, employeeNo).Count(&count)
		if count > 0 {
			skipped++
			continue
		}
		status := 1
		if csvCell(record, index, "status") == "停用" || csvCell(record, index, "status") == "0" {
			status = 0
		}
		if status == 1 {
			if err := h.requireQuotaAvailable(user.TenantID, "max_users", 1); err != nil {
				response.Error(c, 429, response.CodeBadRequest, err.Error())
				return
			}
		}
		companyID := h.resolveCompanyIDByName(user.TenantID, csvCell(record, index, "company_name"))
		departmentID := h.resolveDepartmentIDByName(user.TenantID, companyID, csvCell(record, index, "department_name"))
		phone, msg := normalizeOptionalPhone(nullableFromString(csvCell(record, index, "phone")))
		if msg != "" {
			errors = append(errors, fmt.Sprintf("第 %d 行：%s", line+2, msg))
			continue
		}
		password := mustHashPassword(employeeNo)
		newUser := models.AppUser{TenantID: user.TenantID, EmployeeNo: employeeNo, Account: employeeNo, PasswordHash: password, Name: name, Phone: phone, Email: nullableFromString(csvCell(record, index, "email")), CompanyID: companyID, DepartmentID: departmentID, Status: status}
		if err := h.db.Create(&newUser).Error; err != nil {
			errors = append(errors, fmt.Sprintf("第 %d 行：%s", line+2, err.Error()))
			continue
		}
		created++
	}
	h.audit(c, user.TenantID, user.ID, "user", "batch_import", fmt.Sprintf("批量导入用户：创建 %d，跳过 %d", created, skipped), gin.H{"created": created, "skipped": skipped, "errors": firstStrings(errors, 10)})
	response.OK(c, gin.H{"created": created, "skipped": skipped, "errors": firstStrings(errors, 20)})
}

func (h *IdentityHandler) ExportCompaniesCSV(c *gin.Context) {
	h.exportOrgCSV(c, "companies_export.csv", "company", []string{"name", "code", "company_type", "parent_name", "status"})
}

func (h *IdentityHandler) ExportDepartmentsCSV(c *gin.Context) {
	h.exportOrgCSV(c, "departments_export.csv", "department", []string{"name", "code", "company_name", "parent_name", "status"})
}
