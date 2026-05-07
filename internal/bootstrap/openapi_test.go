package bootstrap

import (
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestOpenAPISpecCoversRegisteredAPIRoutes(t *testing.T) {
	spec := openAPISpec()
	paths := spec["paths"].(gin.H)

	router := NewRouter(Config{AppEnv: "test", AuthSecret: "test-secret", TokenTTLHours: 1}, nil, nil)
	for _, route := range router.Routes() {
		if !strings.HasPrefix(route.Path, "/api/") {
			continue
		}
		path := ginPathToOpenAPIPath(route.Path)
		operations, ok := paths[path]
		require.Truef(t, ok, "missing OpenAPI path for %s %s", route.Method, route.Path)
		_, ok = operations.(gin.H)[strings.ToLower(route.Method)]
		require.Truef(t, ok, "missing OpenAPI operation for %s %s", route.Method, route.Path)
	}
}

func TestOpenAPISpecIncludesSchemas(t *testing.T) {
	spec := openAPISpec()
	components := spec["components"].(gin.H)
	schemas := components["schemas"].(gin.H)
	require.Contains(t, schemas, "User")
	require.Contains(t, schemas, "RoleCreateRequest")
	require.Contains(t, schemas, "TenantPage")

	paths := spec["paths"].(gin.H)
	login := paths["/api/auth/login"].(gin.H)["post"].(gin.H)
	require.Contains(t, login, "requestBody")
	loginResponses := login["responses"].(gin.H)
	require.Equal(t, "OK", loginResponses["200"].(gin.H)["description"])

	createRole := paths["/api/roles"].(gin.H)["post"].(gin.H)
	require.Contains(t, createRole, "requestBody")
	roleResponse := createRole["responses"].(gin.H)["200"].(gin.H)
	content := roleResponse["content"].(gin.H)["application/json"].(gin.H)
	require.Contains(t, content["schema"].(gin.H), "properties")
}

func ginPathToOpenAPIPath(path string) string {
	parts := strings.Split(path, "/")
	for i, part := range parts {
		if strings.HasPrefix(part, ":") {
			parts[i] = "{" + strings.TrimPrefix(part, ":") + "}"
		}
	}
	return strings.Join(parts, "/")
}
