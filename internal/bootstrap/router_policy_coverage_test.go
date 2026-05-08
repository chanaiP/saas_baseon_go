package bootstrap_test

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"saas_baseon_go/internal/bootstrap"
	"saas_baseon_go/internal/interfaces/http/handlers"
)

func TestProtectedAPIRoutesHavePermissionPolicy(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := bootstrap.NewRouter(bootstrap.Config{AppEnv: "test", AuthSecret: "test-secret", TokenTTLHours: 1}, nil, nil)

	missing := handlers.UnclassifiedAPIRoutes(router.Routes())

	require.Empty(t, missing)
}
