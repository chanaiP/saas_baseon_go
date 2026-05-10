package handlers

import (
	"encoding/json"
	"errors"
	"strings"

	"gorm.io/gorm"

	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
)

type orgNodePayload struct {
	NodeType           string  `json:"node_type"`
	Name               string  `json:"name"`
	Code               *string `json:"code"`
	CompanyType        *string `json:"company_type"`
	CompanyID          *uint64 `json:"company_id"`
	ParentID           *uint64 `json:"parent_id"`
	ParentIDSet        bool    `json:"-"`
	ParentCompanyID    *uint64 `json:"parent_company_id"`
	ParentDepartmentID *uint64 `json:"parent_department_id"`
	ParentStoreID      *uint64 `json:"parent_store_id"`
	StoreID            *uint64 `json:"store_id"`
	Status             int     `json:"status"`
	StatusSet          bool    `json:"-"`
}

func (p *orgNodePayload) UnmarshalJSON(data []byte) error {
	type alias orgNodePayload
	var raw alias
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}
	_, raw.ParentIDSet = fields["parent_id"]
	_, raw.StatusSet = fields["status"]
	*p = orgNodePayload(raw)
	return nil
}

func (h *IdentityHandler) resolveOrgNodeCreatePlacement(tenantID uint64, body orgNodePayload) (*uint64, *uint64, error) {
	nodeType := strings.TrimSpace(body.NodeType)
	switch nodeType {
	case "company":
		if body.ParentID != nil && body.ParentDepartmentID != nil {
			return nil, nil, errors.New("不能同时指定上级公司与上级部门")
		}
		parentID := body.ParentID
		if body.ParentDepartmentID != nil {
			if err := h.assertOrgNodeType(tenantID, *body.ParentDepartmentID, "department", "上级部门不存在"); err != nil {
				return nil, nil, err
			}
			parentID = body.ParentDepartmentID
		} else if body.ParentID != nil {
			if err := h.assertOrgNodeType(tenantID, *body.ParentID, "company", "上级公司不存在"); err != nil {
				return nil, nil, err
			}
		}
		return parentID, nil, nil
	case "department":
		if body.CompanyID == nil && body.ParentID != nil {
			body.CompanyID = h.resolveCompanyIDFromParent(tenantID, body.ParentID)
		}
		if body.CompanyID == nil {
			return nil, nil, errors.New("所属公司不能为空")
		}
		if err := h.assertOrgNodeType(tenantID, *body.CompanyID, "company", "公司不存在"); err != nil {
			return nil, nil, err
		}
		parentID := body.CompanyID
		if body.ParentID != nil {
			parent, err := h.orgNodeByID(tenantID, *body.ParentID)
			if err != nil {
				return nil, nil, errors.New("父级节点不存在")
			}
			switch parent.NodeType {
			case "company":
				if parent.ID != *body.CompanyID {
					return nil, nil, errors.New("父级公司与所属公司不一致")
				}
				parentID = body.ParentID
			case "store":
				if parent.CompanyID == nil || *parent.CompanyID != *body.CompanyID {
					return nil, nil, errors.New("门店与所属公司不一致")
				}
				body.StoreID = body.ParentID
				parentID = body.ParentID
			case "department":
				if parent.CompanyID == nil || *parent.CompanyID != *body.CompanyID {
					return nil, nil, errors.New("上级部门与所属公司不一致")
				}
				if !sameOptionalUint64(h.storeAncestorID(tenantID, parent.ID), body.StoreID) {
					return nil, nil, errors.New("子部门与上级的门店归属不一致")
				}
				parentID = body.ParentID
			default:
				return nil, nil, errors.New("部门只能挂在公司、门店或部门下")
			}
			return parentID, body.CompanyID, nil
		}
		if body.StoreID != nil {
			store, err := h.orgNodeByType(tenantID, *body.StoreID, "store")
			if err != nil {
				return nil, nil, errors.New("门店不存在")
			}
			if store.CompanyID == nil || *store.CompanyID != *body.CompanyID {
				return nil, nil, errors.New("门店与所属公司不一致")
			}
			parentID = body.StoreID
		}
		if body.ParentID != nil {
			parent, err := h.orgNodeByType(tenantID, *body.ParentID, "department")
			if err != nil {
				return nil, nil, errors.New("上级部门不存在")
			}
			if parent.CompanyID == nil || *parent.CompanyID != *body.CompanyID {
				return nil, nil, errors.New("上级部门与所属公司不一致")
			}
			if !sameOptionalUint64(h.storeAncestorID(tenantID, parent.ID), body.StoreID) {
				return nil, nil, errors.New("子部门与上级的门店归属不一致")
			}
			parentID = body.ParentID
		}
		return parentID, body.CompanyID, nil
	case "store":
		var parentID *uint64
		var companyID *uint64
		switch {
		case body.ParentCompanyID != nil:
			if err := h.assertOrgNodeType(tenantID, *body.ParentCompanyID, "company", "上级公司不存在"); err != nil {
				return nil, nil, err
			}
			parentID = body.ParentCompanyID
			companyID = body.ParentCompanyID
		case body.ParentDepartmentID != nil:
			dep, err := h.orgNodeByType(tenantID, *body.ParentDepartmentID, "department")
			if err != nil {
				return nil, nil, errors.New("上级部门不存在")
			}
			parentID = body.ParentDepartmentID
			companyID = dep.CompanyID
		case body.ParentStoreID != nil:
			store, err := h.orgNodeByType(tenantID, *body.ParentStoreID, "store")
			if err != nil {
				return nil, nil, errors.New("上级门店不存在")
			}
			parentID = body.ParentStoreID
			companyID = store.CompanyID
		case body.ParentID != nil:
			parent, err := h.orgNodeByID(tenantID, *body.ParentID)
			if err != nil {
				return nil, nil, errors.New("父级节点不存在")
			}
			parentID = body.ParentID
			if parent.NodeType == "company" {
				companyID = &parent.ID
			} else {
				companyID = parent.CompanyID
			}
		default:
			return nil, nil, errors.New("上级节点不能为空")
		}
		return parentID, companyID, nil
	default:
		parentID := body.ParentID
		var companyID *uint64
		if parentID != nil {
			if _, err := h.orgNodeByID(tenantID, *parentID); err != nil {
				return nil, nil, errors.New("父级节点不存在")
			}
			companyID = h.resolveCompanyIDFromParent(tenantID, parentID)
		}
		if nodeType == "company" {
			companyID = nil
		}
		return parentID, companyID, nil
	}
}

func (h *IdentityHandler) assertOrgNodeType(tenantID, id uint64, nodeType string, message string) error {
	if _, err := h.orgNodeByType(tenantID, id, nodeType); err != nil {
		return errors.New(message)
	}
	return nil
}

func (h *IdentityHandler) orgNodeByType(tenantID, id uint64, nodeType string) (models.OrgNode, error) {
	var row models.OrgNode
	err := h.db.Where("id = ? AND tenant_id = ? AND node_type = ? AND deleted_at IS NULL", id, tenantID, nodeType).First(&row).Error
	return row, err
}

func (h *IdentityHandler) orgNodeByID(tenantID, id uint64) (models.OrgNode, error) {
	var row models.OrgNode
	err := h.db.Where("id = ? AND tenant_id = ? AND deleted_at IS NULL", id, tenantID).First(&row).Error
	return row, err
}

func (h *IdentityHandler) storeAncestorID(tenantID, nodeID uint64) *uint64 {
	rows := []models.OrgNode{}
	_ = h.db.Where("tenant_id = ? AND deleted_at IS NULL", tenantID).Find(&rows).Error
	byID := map[uint64]models.OrgNode{}
	for _, row := range rows {
		byID[row.ID] = row
	}
	node, ok := byID[nodeID]
	for ok && node.ParentID != nil {
		parent, exists := byID[*node.ParentID]
		if !exists {
			return nil
		}
		if parent.NodeType == "store" {
			return &parent.ID
		}
		node = parent
	}
	return nil
}

func sameOptionalUint64(a, b *uint64) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return *a == *b
}

func (h *IdentityHandler) normalizedCompanyType(nodeType string, value *string) *string {
	if nodeType != "group" && nodeType != "company" && nodeType != "store" {
		return nil
	}
	return nullableTrimmed(value)
}

func (h *IdentityHandler) validateOrgParentChange(tenantID, nodeID uint64, parentID *uint64) error {
	if parentID == nil {
		return nil
	}
	if *parentID == nodeID {
		return errors.New("父级不能是当前节点本身")
	}
	if _, err := h.orgNodeByID(tenantID, *parentID); err != nil {
		return errors.New("父级节点不存在")
	}
	for _, childID := range h.descendantOrgNodeIDs(tenantID, nodeID) {
		if childID == *parentID && childID != nodeID {
			return errors.New("父级不能是当前节点或其子节点")
		}
	}
	return nil
}

func (h *IdentityHandler) resolveCompanyIDFromParent(tenantID uint64, parentID *uint64) *uint64 {
	if parentID == nil {
		return nil
	}
	parent, err := h.orgNodeByID(tenantID, *parentID)
	if err != nil {
		return nil
	}
	if parent.NodeType == "company" {
		return &parent.ID
	}
	return parent.CompanyID
}

func (h *IdentityHandler) cascadeRefreshCompanyIDs(tx *gorm.DB, tenantID, rootID uint64) error {
	var rows []models.OrgNode
	if err := tx.Where("tenant_id = ? AND deleted_at IS NULL", tenantID).Find(&rows).Error; err != nil {
		return err
	}
	byID := map[uint64]models.OrgNode{}
	childrenByParent := map[uint64][]models.OrgNode{}
	for _, row := range rows {
		byID[row.ID] = row
		if row.ParentID != nil {
			childrenByParent[*row.ParentID] = append(childrenByParent[*row.ParentID], row)
		}
	}
	companyIDFor := func(row models.OrgNode) *uint64 {
		if row.NodeType == "company" || row.ParentID == nil {
			return nil
		}
		parent, ok := byID[*row.ParentID]
		if !ok {
			return nil
		}
		if parent.NodeType == "company" {
			return &parent.ID
		}
		return parent.CompanyID
	}
	stack := []models.OrgNode{}
	if root, ok := byID[rootID]; ok {
		stack = append(stack, root)
	}
	for len(stack) > 0 {
		node := stack[0]
		stack = stack[1:]
		nextCompanyID := companyIDFor(node)
		if err := tx.Model(&models.OrgNode{}).Where("id = ?", node.ID).Update("company_id", nextCompanyID).Error; err != nil {
			return err
		}
		node.CompanyID = nextCompanyID
		byID[node.ID] = node
		stack = append(stack, childrenByParent[node.ID]...)
	}
	return nil
}
