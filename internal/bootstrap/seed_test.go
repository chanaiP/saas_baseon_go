package bootstrap

import (
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
)

func TestSeedInitialAdminPasswordAllowsDevelopmentDefault(t *testing.T) {
	t.Setenv("APP_ENV", "development")
	t.Setenv("BOOTSTRAP_ADMIN_PASSWORD", "")

	password, err := seedInitialAdminPassword()

	require.NoError(t, err)
	require.Equal(t, "112233", password)
}

func TestSeedInitialAdminPasswordRejectsProductionDefault(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("BOOTSTRAP_ADMIN_PASSWORD", "112233")

	_, err := seedInitialAdminPassword()

	require.Error(t, err)
	require.Contains(t, err.Error(), "BOOTSTRAP_ADMIN_PASSWORD")
}

func TestSeedInitialAdminPasswordAllowsExplicitProductionPassword(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("BOOTSTRAP_ADMIN_PASSWORD", "StrongBootstrap123")

	password, err := seedInitialAdminPassword()

	require.NoError(t, err)
	require.Equal(t, "StrongBootstrap123", password)
}

func TestAIProviderSeedPreservesOperationalOverrides(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.Exec(`CREATE TABLE ai_providers (
		id TEXT PRIMARY KEY,
		name TEXT,
		code TEXT,
		type TEXT,
		base_url TEXT,
		auth_type TEXT,
		status TEXT,
		priority INTEGER,
		region TEXT,
		qps_limit INTEGER,
		monthly_budget REAL,
		owner TEXT,
		remark TEXT,
		created_at DATETIME,
		updated_at DATETIME,
		deleted_at DATETIME
	)`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE ai_provider_accounts (
		id TEXT PRIMARY KEY,
		provider_id TEXT,
		account_name TEXT,
		endpoint TEXT,
		key_alias TEXT,
		login_method TEXT,
		login_account TEXT,
		maintainer TEXT,
		maintainer_contact TEXT,
		encrypted_api_key TEXT,
		encrypted_secret TEXT,
		quota_limit REAL,
		used_quota REAL,
		status TEXT,
		created_at DATETIME,
		updated_at DATETIME,
		deleted_at DATETIME
	)`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE ai_provider_apis (
		id TEXT PRIMARY KEY,
		provider_id TEXT,
		account_id TEXT,
		api_name TEXT,
		api_path TEXT,
		api_type TEXT,
		capabilities TEXT,
		auth_type TEXT,
		qps_limit INTEGER,
		timeout_ms INTEGER,
		status TEXT,
		last_called_at DATETIME,
		health_status TEXT,
		health_message TEXT,
		health_checked_at DATETIME,
		created_at DATETIME,
		updated_at DATETIME,
		deleted_at DATETIME
	)`).Error)

	now := time.Now()
	provider := models.AIProvider{
		ID:           "provider-1",
		Name:         "OpenAI",
		Code:         "openai",
		Type:         "openai-compatible",
		BaseURL:      "https://old.example",
		AuthType:     "bearer",
		Status:       "inactive",
		AITimeFields: models.AITimeFields{CreatedAt: now, UpdatedAt: now},
	}
	require.NoError(t, db.Create(&provider).Error)
	account := models.AIProviderAccount{
		ID:                "account-1",
		ProviderID:        provider.ID,
		AccountName:       "prod-main",
		Endpoint:          "https://old.example",
		KeyAlias:          "CUSTOM_OPENAI_KEY",
		LoginMethod:       "email",
		LoginAccount:      "ops@example.com",
		Maintainer:        "ops",
		MaintainerContact: "ops@example.com",
		Status:            "inactive",
		AITimeFields:      models.AITimeFields{CreatedAt: now, UpdatedAt: now},
	}
	require.NoError(t, db.Create(&account).Error)
	api := models.AIProviderAPI{
		ID:           "api-1",
		ProviderID:   provider.ID,
		AccountID:    account.ID,
		APIName:      "chat.completions",
		APIPath:      "/old",
		APIType:      "chat",
		Capabilities: []string{"chat_completion"},
		AuthType:     "bearer",
		Status:       "inactive",
		AITimeFields: models.AITimeFields{CreatedAt: now, UpdatedAt: now},
	}
	require.NoError(t, db.Create(&api).Error)

	seed := aiProviderSeed{
		name: "OpenAI", code: "openai", providerType: "openai-compatible", baseURL: "https://new.example", authType: "bearer",
	}
	providerID, err := upsertAIProviderSeed(db, seed, 0)
	require.NoError(t, err)
	require.NoError(t, upsertAIProviderAccountAndAPIs(db, providerID, seed))

	var gotProvider models.AIProvider
	require.NoError(t, db.Where("id = ?", provider.ID).First(&gotProvider).Error)
	require.Equal(t, "inactive", gotProvider.Status)
	require.Equal(t, "https://new.example", gotProvider.BaseURL)

	var gotAccount models.AIProviderAccount
	require.NoError(t, db.Where("id = ?", account.ID).First(&gotAccount).Error)
	require.Equal(t, "inactive", gotAccount.Status)
	require.Equal(t, "CUSTOM_OPENAI_KEY", gotAccount.KeyAlias)
	require.Equal(t, "https://new.example", gotAccount.Endpoint)

	var gotAPI models.AIProviderAPI
	require.NoError(t, db.Where("id = ?", api.ID).First(&gotAPI).Error)
	require.Equal(t, "inactive", gotAPI.Status)
	require.Equal(t, "/old", gotAPI.APIPath)
}
