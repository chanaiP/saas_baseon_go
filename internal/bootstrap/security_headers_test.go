package bootstrap

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestSecurityHeadersMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(securityHeadersMiddleware("production"))
	router.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)

	router.ServeHTTP(rec, req)

	require.Equal(t, "nosniff", rec.Header().Get("X-Content-Type-Options"))
	require.Equal(t, "DENY", rec.Header().Get("X-Frame-Options"))
	require.Equal(t, "strict-origin-when-cross-origin", rec.Header().Get("Referrer-Policy"))
	require.Contains(t, rec.Header().Get("Content-Security-Policy"), "default-src 'self'")
	require.Contains(t, rec.Header().Get("Strict-Transport-Security"), "max-age=31536000")
}
