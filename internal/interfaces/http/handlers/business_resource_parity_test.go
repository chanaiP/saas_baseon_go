package handlers

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
)

func setupBusinessResourceTestHandler(t *testing.T) (*IdentityHandler, *gin.Context) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&models.Tenant{},
		&models.DictType{},
		&models.DictItem{},
		&models.TenantDictItemOverride{},
		&models.BusinessUnit{},
		&models.BusinessResource{},
		&models.BusinessResourceActor{},
		&models.BusinessResourceFieldConfig{},
		&models.BusinessUnitResource{},
		&models.BusinessResourceRelation{},
		&models.AppUser{},
		&models.OrgNode{},
	))
	require.NoError(t, db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS uq_bu_resource_primary_owner_test ON business_unit_resource (tenant_id, resource_id) WHERE is_primary = 1 AND relation_type = 'owner' AND deleted_at IS NULL`).Error)
	now := time.Now()
	require.NoError(t, db.Create(&models.Tenant{ID: 1, Code: "platform", Name: "平台主体", Status: 1, IsPlatform: true, CreatedAt: now, UpdatedAt: now}).Error)
	require.NoError(t, db.Create(&models.Tenant{ID: 4, Code: "tenant4", Name: "租户4", Status: 1, CreatedAt: now, UpdatedAt: now}).Error)
	require.NoError(t, db.Create(&models.Tenant{ID: 5, Code: "tenant5", Name: "租户5", Status: 1, CreatedAt: now, UpdatedAt: now}).Error)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/", nil)
	return &IdentityHandler{db: db}, c
}

func seedDict(t *testing.T, db *gorm.DB, code string, items ...models.DictItem) {
	t.Helper()
	now := time.Now()
	dictType := models.DictType{TenantID: 1, Code: code, Name: code, Scope: "platform", TenantEditable: true, CreatedAt: now, UpdatedAt: now}
	require.NoError(t, db.Create(&dictType).Error)
	for i := range items {
		items[i].TenantID = 1
		items[i].DictTypeID = dictType.ID
		items[i].CreatedAt = now
		items[i].UpdatedAt = now
		require.NoError(t, db.Create(&items[i]).Error)
	}
}

func seedBusinessResourceDicts(t *testing.T, db *gorm.DB) (uint64, uint64) {
	t.Helper()
	seedDict(t, db, "business_resource.resource_category",
		models.DictItem{Label: "门店", Value: "store", SortOrder: 10, Enabled: true},
		models.DictItem{Label: "渠道", Value: "channel", SortOrder: 20, Enabled: true},
	)
	now := time.Now()
	resourceType := models.DictType{TenantID: 1, Code: "business_resource.resource_type", Name: "资源类型", Scope: "platform", TenantEditable: true, CreatedAt: now, UpdatedAt: now}
	require.NoError(t, db.Create(&resourceType).Error)
	storeGroup := models.DictItem{TenantID: 1, DictTypeID: resourceType.ID, Label: "门店", Value: "store", SortOrder: 10, Enabled: true, CreatedAt: now, UpdatedAt: now}
	channelGroup := models.DictItem{TenantID: 1, DictTypeID: resourceType.ID, Label: "渠道", Value: "channel", SortOrder: 20, Enabled: true, CreatedAt: now, UpdatedAt: now}
	require.NoError(t, db.Create(&storeGroup).Error)
	require.NoError(t, db.Create(&channelGroup).Error)
	storeGroupID := storeGroup.ID
	channelGroupID := channelGroup.ID
	require.NoError(t, db.Create(&models.DictItem{TenantID: 1, DictTypeID: resourceType.ID, ParentID: &storeGroupID, Label: "抖音门店", Value: "douyin_shop", SortOrder: 11, Enabled: true, CreatedAt: now, UpdatedAt: now}).Error)
	require.NoError(t, db.Create(&models.DictItem{TenantID: 1, DictTypeID: resourceType.ID, ParentID: &channelGroupID, Label: "直播间", Value: "live_room", SortOrder: 21, Enabled: true, CreatedAt: now, UpdatedAt: now}).Error)
	seedDict(t, db, "business_resource.source_mode",
		models.DictItem{Label: "原生", Value: "native", SortOrder: 10, Enabled: true},
		models.DictItem{Label: "引用", Value: "reference", SortOrder: 20, Enabled: true},
	)
	seedDict(t, db, "business_resource.resource_status",
		models.DictItem{Label: "启用", Value: "active", SortOrder: 10, Enabled: true},
		models.DictItem{Label: "归档", Value: "archived", SortOrder: 20, Enabled: true},
	)
	seedDict(t, db, "business_resource.unit_resource_relation_type",
		models.DictItem{Label: "归属", Value: "owner", SortOrder: 10, Enabled: true},
	)
	seedDict(t, db, "business_resource.resource_relation_type",
		models.DictItem{Label: "父子", Value: "parent_child", SortOrder: 10, Enabled: true},
	)
	seedDict(t, db, "base.status",
		models.DictItem{Label: "启用", Value: "active", SortOrder: 10, Enabled: true},
	)
	return storeGroupID, channelGroupID
}

func seedBusinessUnitDict(t *testing.T, db *gorm.DB) {
	t.Helper()
	now := time.Now()
	dictType := models.DictType{TenantID: 1, Code: "business_unit", Name: "业务单元", Scope: "platform", TenantEditable: true, CreatedAt: now, UpdatedAt: now}
	require.NoError(t, db.Create(&dictType).Error)
	store := models.DictItem{TenantID: 1, DictTypeID: dictType.ID, Label: "店铺门店", Value: "store", SortOrder: 10, Enabled: true, CreatedAt: now, UpdatedAt: now}
	ad := models.DictItem{TenantID: 1, DictTypeID: dictType.ID, Label: "投放广告", Value: "ad", SortOrder: 20, Enabled: true, CreatedAt: now, UpdatedAt: now}
	require.NoError(t, db.Create(&store).Error)
	require.NoError(t, db.Create(&ad).Error)
	storeID := store.ID
	adID := ad.ID
	require.NoError(t, db.Create(&models.DictItem{TenantID: 1, DictTypeID: dictType.ID, ParentID: &storeID, Label: "抖音电商", Value: "douyin_ecommerce", SortOrder: 101, Enabled: true, CreatedAt: now, UpdatedAt: now}).Error)
	require.NoError(t, db.Create(&models.DictItem{TenantID: 1, DictTypeID: dictType.ID, ParentID: &storeID, Label: "天猫电商", Value: "tmall_ecommerce", SortOrder: 102, Enabled: true, CreatedAt: now, UpdatedAt: now}).Error)
	require.NoError(t, db.Create(&models.DictItem{TenantID: 1, DictTypeID: dictType.ID, ParentID: &adID, Label: "巨量千川", Value: "qianchuan", SortOrder: 201, Enabled: true, CreatedAt: now, UpdatedAt: now}).Error)
}

func TestFetchDictItemsByCodeKeepsParentIDForResourceTypeTree(t *testing.T) {
	h, _ := setupBusinessResourceTestHandler(t)
	storeGroupID, _ := seedBusinessResourceDicts(t, h.db)

	_, items, err := h.dictionaryService().ItemsByCode(context.Background(), h.dictionaryViewer(models.AppUser{TenantID: 4}), "business_resource.resource_type")

	require.NoError(t, err)
	var foundChild bool
	for _, item := range items {
		if item.Value == "douyin_shop" {
			require.NotNil(t, item.ParentID)
			require.Equal(t, storeGroupID, *item.ParentID)
			foundChild = true
		}
	}
	require.True(t, foundChild, "resource type child item should be returned")
}

func TestBusinessResourceTypeValidationUsesDictParentID(t *testing.T) {
	h, c := setupBusinessResourceTestHandler(t)
	seedBusinessResourceDicts(t, h.db)
	user := models.AppUser{TenantID: 4}

	require.NoError(t, h.validateBusinessResourceType(c, user, "store", "douyin_shop"))
	require.EqualError(t, h.validateBusinessResourceType(c, user, "channel", "douyin_shop"), "资源类型不属于所选资源大类")
	require.EqualError(t, h.validateBusinessResourceType(c, user, "store", "store"), "资源类型不能选择资源大类分组")
}

func TestBusinessResourceV2UsesBusinessUnitDictionaryWithoutChangingDictionarySchema(t *testing.T) {
	h, c := setupBusinessResourceTestHandler(t)
	seedBusinessResourceDicts(t, h.db)
	seedBusinessUnitDict(t, h.db)
	user := models.AppUser{TenantID: 4}

	row, err := h.businessResourceFromBody(c, user, businessResourceBody{
		UnitTypeCode:     "store",
		BusinessUnitCode: "douyin_ecommerce",
		ResourceName:     "抖音一店",
		ResourceCode:     "DY001",
		SourceMode:       "native",
		Status:           "active",
	}, nil)

	require.NoError(t, err)
	require.Equal(t, "store", row.UnitTypeCode)
	require.Equal(t, "店铺门店", row.UnitTypeName)
	require.Equal(t, "douyin_ecommerce", row.BusinessUnitCode)
	require.Equal(t, "抖音电商", row.BusinessUnitName)
	require.Equal(t, row.UnitTypeCode, row.ResourceCategory)
	require.Equal(t, row.BusinessUnitCode, row.ResourceType)

	_, err = h.businessResourceFromBody(c, user, businessResourceBody{
		UnitTypeCode:     "store",
		BusinessUnitCode: "qianchuan",
		ResourceName:     "错配资源",
		ResourceCode:     "BADTREE",
		SourceMode:       "native",
		Status:           "active",
	}, nil)
	require.EqualError(t, err, "业务单元不属于所选业务单元类型")
}

func TestBusinessResourceFieldConfigItemsPreferTenantOverrideAndImportVisibility(t *testing.T) {
	h, _ := setupBusinessResourceTestHandler(t)
	now := time.Now()
	tenantID := uint64(4)
	require.NoError(t, h.db.Create(&models.BusinessResourceFieldConfig{UnitTypeCode: "store", BusinessUnitCode: "douyin_ecommerce", FieldKey: "shop_level", FieldLabel: "平台等级", FieldType: "text", ShowInDetail: true, ShowInImport: true, SortOrder: 10, Status: "active", CreatedAt: now, UpdatedAt: now}).Error)
	require.NoError(t, h.db.Create(&models.BusinessResourceFieldConfig{TenantID: &tenantID, UnitTypeCode: "store", BusinessUnitCode: "douyin_ecommerce", FieldKey: "shop_level", FieldLabel: "店铺等级", FieldType: "text", ShowInDetail: true, ShowInImport: true, SortOrder: 10, Status: "active", CreatedAt: now, UpdatedAt: now}).Error)
	require.NoError(t, h.db.Create(&models.BusinessResourceFieldConfig{TenantID: &tenantID, UnitTypeCode: "store", BusinessUnitCode: "douyin_ecommerce", FieldKey: "internal_note", FieldLabel: "内部备注", FieldType: "text", ShowInDetail: true, ShowInImport: false, SortOrder: 20, Status: "active", CreatedAt: now, UpdatedAt: now}).Error)

	detailItems := h.businessResourceFieldConfigItems(tenantID, "store", "douyin_ecommerce", true)
	require.Len(t, detailItems, 2)
	require.Equal(t, "店铺等级", detailItems[0]["field_label"])

	importItems := h.businessResourceFieldConfigItems(tenantID, "store", "douyin_ecommerce", false)
	require.Len(t, importItems, 1)
	require.Equal(t, "shop_level", importItems[0]["field_key"])
}

func TestBusinessResourceSummaryOnlyUsesActiveResourcesAndDictionaryNames(t *testing.T) {
	h, c := setupBusinessResourceTestHandler(t)
	seedBusinessUnitDict(t, h.db)
	user := models.AppUser{TenantID: 4}
	now := time.Now()
	deletedAt := now
	rows := []models.BusinessResource{
		{TenantID: 4, UnitTypeCode: "store", UnitTypeName: "store", BusinessUnitCode: "douyin_ecommerce", BusinessUnitName: "douyin_ecommerce", ResourceName: "抖音一店", ResourceCode: "DY001", ResourceCategory: "store", ResourceType: "douyin_ecommerce", SourceMode: "native", Status: "active", CreatedAt: now, UpdatedAt: now},
		{TenantID: 4, UnitTypeCode: "store", UnitTypeName: "store", BusinessUnitCode: "douyin_ecommerce", BusinessUnitName: "douyin_ecommerce", ResourceName: "抖音二店", ResourceCode: "DY002", ResourceCategory: "store", ResourceType: "douyin_ecommerce", SourceMode: "native", Status: "active", CreatedAt: now, UpdatedAt: now},
		{TenantID: 4, UnitTypeCode: "store", UnitTypeName: "store", BusinessUnitCode: "tmall_ecommerce", BusinessUnitName: "tmall_ecommerce", ResourceName: "天猫停用店", ResourceCode: "TM001", ResourceCategory: "store", ResourceType: "tmall_ecommerce", SourceMode: "native", Status: "inactive", CreatedAt: now, UpdatedAt: now},
		{TenantID: 4, UnitTypeCode: "ad", UnitTypeName: "ad", BusinessUnitCode: "qianchuan", BusinessUnitName: "qianchuan", ResourceName: "已归档投放", ResourceCode: "AD001", ResourceCategory: "ad", ResourceType: "qianchuan", SourceMode: "native", Status: "active", CreatedAt: now, UpdatedAt: now, DeletedAt: &deletedAt},
		{TenantID: 5, UnitTypeCode: "ad", UnitTypeName: "ad", BusinessUnitCode: "qianchuan", BusinessUnitName: "qianchuan", ResourceName: "其他租户投放", ResourceCode: "AD002", ResourceCategory: "ad", ResourceType: "qianchuan", SourceMode: "native", Status: "active", CreatedAt: now, UpdatedAt: now},
	}
	require.NoError(t, h.db.Create(&rows).Error)

	items, err := h.businessResourceSummaryItems(c, user)

	require.NoError(t, err)
	require.Len(t, items, 1)
	require.Equal(t, "store", items[0]["code"])
	require.Equal(t, "店铺门店", items[0]["name"])
	require.Equal(t, int64(2), items[0]["resource_count"])
	units := items[0]["business_units"].([]gin.H)
	require.Len(t, units, 1)
	require.Equal(t, "douyin_ecommerce", units[0]["code"])
	require.Equal(t, "抖音电商", units[0]["name"])
	require.Equal(t, int64(2), units[0]["resource_count"])
}

func TestBusinessResourceActorRejectsCrossTenantRefs(t *testing.T) {
	h, c := setupBusinessResourceTestHandler(t)
	seedDict(t, h.db, "business_resource.actor_type",
		models.DictItem{Label: "负责组织", Value: "org_node", SortOrder: 10, Enabled: true},
		models.DictItem{Label: "负责人员", Value: "user", SortOrder: 20, Enabled: true},
	)
	seedDict(t, h.db, "business_resource.actor_role",
		models.DictItem{Label: "负责人", Value: "owner", SortOrder: 10, Enabled: true},
	)
	now := time.Now()
	user := models.AppUser{TenantID: 4}
	org := models.OrgNode{TenantID: 4, NodeType: "department", Name: "本租户组织", Status: 1, CreatedAt: now, UpdatedAt: now}
	otherOrg := models.OrgNode{TenantID: 5, NodeType: "department", Name: "其他租户组织", Status: 1, CreatedAt: now, UpdatedAt: now}
	owner := models.AppUser{TenantID: 4, EmployeeNo: "A1", Account: "a1", PasswordHash: "x", Name: "本租户用户", Status: 1, CreatedAt: now, UpdatedAt: now}
	otherOwner := models.AppUser{TenantID: 5, EmployeeNo: "B1", Account: "b1", PasswordHash: "x", Name: "其他租户用户", Status: 1, CreatedAt: now, UpdatedAt: now}
	require.NoError(t, h.db.Create(&org).Error)
	require.NoError(t, h.db.Create(&otherOrg).Error)
	require.NoError(t, h.db.Create(&owner).Error)
	require.NoError(t, h.db.Create(&otherOwner).Error)

	_, err := h.businessResourceActorFromInput(c, user, 10, businessResourceActorInput{ActorType: "org_node", ActorID: org.ID, RoleType: "owner", Status: "active"})
	require.NoError(t, err)

	_, err = h.businessResourceActorFromInput(c, user, 10, businessResourceActorInput{ActorType: "org_node", ActorID: otherOrg.ID, RoleType: "owner", Status: "active"})
	require.EqualError(t, err, "负责组织不存在")

	_, err = h.businessResourceActorFromInput(c, user, 10, businessResourceActorInput{ActorType: "user", ActorID: owner.ID, RoleType: "owner", Status: "active"})
	require.NoError(t, err)

	_, err = h.businessResourceActorFromInput(c, user, 10, businessResourceActorInput{ActorType: "user", ActorID: otherOwner.ID, RoleType: "owner", Status: "active"})
	require.EqualError(t, err, "负责人员不存在")
}

func TestBusinessResourceByIDEnforcesTenantIsolation(t *testing.T) {
	h, _ := setupBusinessResourceTestHandler(t)
	now := time.Now()
	row := models.BusinessResource{TenantID: 5, ResourceName: "其他租户门店", ResourceCode: "other-shop", ResourceCategory: "store", ResourceType: "douyin_shop", SourceMode: "native", Status: "active", CreatedAt: now, UpdatedAt: now}
	require.NoError(t, h.db.Create(&row).Error)

	_, err := h.businessResourceByID(4, row.ID)

	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestBusinessUnitRefsValidateTenantScopeAndNewFieldsJSON(t *testing.T) {
	h, _ := setupBusinessResourceTestHandler(t)
	now := time.Now()
	parent := models.BusinessUnit{TenantID: 4, Name: "总部BU", Code: "HQ", Status: 1, CreatedAt: now, UpdatedAt: now}
	user := models.AppUser{TenantID: 4, EmployeeNo: "E4", Account: "e4", PasswordHash: "x", Name: "负责人", Status: 1, CreatedAt: now, UpdatedAt: now}
	org := models.OrgNode{TenantID: 4, NodeType: "department", Name: "运营部", Status: 1, CreatedAt: now, UpdatedAt: now}
	otherTenantParent := models.BusinessUnit{TenantID: 5, Name: "其他BU", Code: "OTHER", Status: 1, CreatedAt: now, UpdatedAt: now}
	require.NoError(t, h.db.Create(&parent).Error)
	require.NoError(t, h.db.Create(&user).Error)
	require.NoError(t, h.db.Create(&org).Error)
	require.NoError(t, h.db.Create(&otherTenantParent).Error)

	require.NoError(t, h.validateBusinessUnitRefs(4, 0, &parent.ID, &user.ID, &org.ID))
	require.EqualError(t, h.validateBusinessUnitRefs(4, 0, &otherTenantParent.ID, nil, nil), "上级业务单元不存在")

	scenario := "retail_store"
	form := "profit_center"
	row := models.BusinessUnit{TenantID: 4, Name: "门店BU", Code: "STORE", UnitScenario: &scenario, UnitForm: &form, ParentID: &parent.ID, OwnerUserID: &user.ID, OwnerOrgID: &org.ID, OperationEnabled: true, SettlementEnabled: true, DataScopeEnabled: true}
	payload := businessUnitToJSON(row)
	require.Equal(t, scenario, *(payload["unit_scenario"].(*string)))
	require.Equal(t, form, *(payload["unit_form"].(*string)))
	require.Equal(t, true, payload["operation_enabled"])
	require.Equal(t, true, payload["settlement_enabled"])
	require.Equal(t, true, payload["data_scope_enabled"])
}

func TestBusinessUnitResourceTenantIsolationAndPrimaryOwnerUniqueness(t *testing.T) {
	h, _ := setupBusinessResourceTestHandler(t)
	now := time.Now()
	bu := models.BusinessUnit{TenantID: 4, Name: "业务单元", Code: "BU", Status: 1, CreatedAt: now, UpdatedAt: now}
	resource := models.BusinessResource{TenantID: 4, ResourceName: "门店", ResourceCode: "SHOP", ResourceCategory: "store", ResourceType: "douyin_shop", SourceMode: "native", Status: "active", CreatedAt: now, UpdatedAt: now}
	otherTenantBU := models.BusinessUnit{TenantID: 5, Name: "其他业务单元", Code: "OTHERBU", Status: 1, CreatedAt: now, UpdatedAt: now}
	require.NoError(t, h.db.Create(&bu).Error)
	require.NoError(t, h.db.Create(&resource).Error)
	require.NoError(t, h.db.Create(&otherTenantBU).Error)

	require.NoError(t, h.assertBusinessUnit(4, bu.ID))
	require.ErrorIs(t, h.assertBusinessUnit(4, otherTenantBU.ID), gorm.ErrRecordNotFound)
	require.NoError(t, h.db.Create(&models.BusinessUnitResource{TenantID: 4, BusinessUnitID: bu.ID, ResourceID: resource.ID, ResourceCategory: "store", ResourceType: "douyin_shop", RelationType: "owner", IsPrimary: true, CreatedAt: now, UpdatedAt: now}).Error)
	err := h.db.Create(&models.BusinessUnitResource{TenantID: 4, BusinessUnitID: bu.ID + 100, ResourceID: resource.ID, ResourceCategory: "store", ResourceType: "douyin_shop", RelationType: "owner", IsPrimary: true, CreatedAt: now, UpdatedAt: now}).Error
	require.Error(t, err)
}

func TestBusinessResourceRelationRejectsCrossTenantResources(t *testing.T) {
	h, _ := setupBusinessResourceTestHandler(t)
	now := time.Now()
	parent := models.BusinessResource{TenantID: 4, ResourceName: "本租户门店", ResourceCode: "SHOP", ResourceCategory: "store", ResourceType: "douyin_shop", SourceMode: "native", Status: "active", CreatedAt: now, UpdatedAt: now}
	child := models.BusinessResource{TenantID: 5, ResourceName: "跨租户门店", ResourceCode: "OTHER", ResourceCategory: "store", ResourceType: "douyin_shop", SourceMode: "native", Status: "active", CreatedAt: now, UpdatedAt: now}
	require.NoError(t, h.db.Create(&parent).Error)
	require.NoError(t, h.db.Create(&child).Error)

	_, err := h.businessResourceByID(4, child.ID)

	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestBusinessResourceValidationFailureBranches(t *testing.T) {
	h, c := setupBusinessResourceTestHandler(t)
	seedBusinessResourceDicts(t, h.db)
	user := models.AppUser{TenantID: 4}

	invalidStatus := "disabled"
	err := h.validateBusinessUnitDictValue(context.Background(), user, "business_resource.resource_status", &invalidStatus, "资源状态")
	require.EqualError(t, err, "资源状态不在字典范围内")

	_, err = normalizeJSON(json.RawMessage(`{"bad"`))
	require.EqualError(t, err, "资源扩展属性必须是合法JSON")

	_, err = h.businessResourceFromBody(c, user, businessResourceBody{
		ResourceName:     "引用资源",
		ResourceCode:     "REF",
		ResourceCategory: "store",
		ResourceType:     "douyin_shop",
		SourceMode:       "reference",
		Status:           "active",
	}, nil)
	require.EqualError(t, err, "引用资源必须填写来源应用、来源表和来源ID")

	_, err = h.businessResourceFromBody(c, user, businessResourceBody{
		ResourceName:     "无效分类",
		ResourceCode:     "BAD",
		ResourceCategory: "store",
		ResourceType:     "live_room",
		SourceMode:       "native",
		Status:           "active",
	}, nil)
	require.EqualError(t, err, "资源类型不属于所选资源大类")
}
