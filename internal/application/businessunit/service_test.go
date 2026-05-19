package businessunit

import (
	"testing"

	"github.com/stretchr/testify/require"

	"saas_baseon_go/internal/application/dictionary"
)

func TestDictItemsToTreeDeduplicatesBusinessUnitCodes(t *testing.T) {
	storeA := uint64(1)
	storeB := uint64(2)
	brandA := uint64(7)
	brandB := uint64(8)
	items := []dictionary.DictItemResponse{
		{ID: storeA, Label: "门店", Value: "store", SortOrder: 10, Enabled: true},
		{ID: storeB, Label: "店铺门店", Value: "store", SortOrder: 900, Enabled: true},
		{ID: brandA, Label: "品牌", Value: "brand", SortOrder: 20, Enabled: true},
		{ID: brandB, Label: "brand", Value: "brand", SortOrder: 900, Enabled: true},
		{ID: 3, ParentID: &storeA, Label: "抖音", Value: "douyin", SortOrder: 11, Enabled: true},
		{ID: 4, ParentID: &storeB, Label: "抖音", Value: "douyin", SortOrder: 900, Enabled: true},
		{ID: 5, ParentID: &storeB, Label: "京东", Value: "jd", SortOrder: 12, Enabled: true},
		{ID: 9, ParentID: &brandB, Label: "运营", Value: "operation", SortOrder: 51, Enabled: true},
	}

	tree := dictItemsToTree(items)

	require.Len(t, tree, 2)
	require.Equal(t, "store", tree[0].Code)
	require.Equal(t, "门店", tree[0].Name)
	require.Len(t, tree[0].Children, 2)
	require.Equal(t, "douyin", tree[0].Children[0].Code)
	require.Equal(t, "jd", tree[0].Children[1].Code)
	require.Equal(t, "brand", tree[1].Code)
	require.Equal(t, "品牌", tree[1].Name)
	require.Len(t, tree[1].Children, 1)
	require.Equal(t, "operation", tree[1].Children[0].Code)
}
