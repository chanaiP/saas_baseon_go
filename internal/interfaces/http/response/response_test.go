package response

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestSafeMessageRedactsInternalErrors(t *testing.T) {
	require.Equal(t, "请求处理失败", SafeMessage(400, "SQLSTATE 23505 duplicate key violates unique constraint"))
	require.Equal(t, "请求处理失败", SafeMessage(400, "Authorization: Bearer secret-token"))
	require.Equal(t, "服务暂时不可用", SafeMessage(500, "unexpected panic"))
}

func TestSafeMessageKeepsBusinessMessage(t *testing.T) {
	require.Equal(t, "用户不存在", SafeMessage(404, "用户不存在"))
	require.Equal(t, "手机号格式错误", SafeMessage(400, "手机号格式错误"))
}

func TestErrorMapsUniqueConstraintToConflict(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)

	Error(c, http.StatusBadRequest, CodeBadRequest, "ERROR: duplicate key value violates unique constraint")

	require.Equal(t, http.StatusConflict, rec.Code)
	require.Contains(t, rec.Body.String(), `"code":40900`)
	require.Contains(t, rec.Body.String(), "数据已存在")
}
