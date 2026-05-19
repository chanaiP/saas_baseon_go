package businessunit

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"saas_baseon_go/internal/application/dictionary"
	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
	"saas_baseon_go/internal/infrastructure/persistence/postgres/repositories"
)

const DictCode = "business_unit"

type Service struct {
	repo *repositories.BusinessUnitRepository
	dict *dictionary.Service
}

func NewService(repo *repositories.BusinessUnitRepository, dict *dictionary.Service) *Service {
	return &Service{repo: repo, dict: dict}
}

func (s *Service) DictionaryTree(ctx context.Context, viewer dictionary.Viewer) (string, []DictNode, error) {
	code, items, err := s.dict.ItemsByCode(ctx, viewer, DictCode)
	if err != nil {
		return code, nil, err
	}
	return code, dictItemsToTree(items), nil
}

func (s *Service) Summary(ctx context.Context, viewer dictionary.Viewer, tenantID uint64) ([]SummaryType, error) {
	rows, err := s.repo.ListSummary(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	_, tree, _ := s.DictionaryTree(ctx, viewer)
	names := dictNameMap(tree)
	byCode := map[string]*SummaryType{}
	out := []SummaryType{}
	for _, r := range rows {
		typeName := coalesceName(names[r.UnitTypeCode], r.UnitTypeName)
		groupName := coalesceName(names[r.UnitGroupCode], r.UnitGroupName)
		item := byCode[r.UnitTypeCode]
		if item == nil {
			out = append(out, SummaryType{Code: r.UnitTypeCode, Name: typeName})
			item = &out[len(out)-1]
			byCode[r.UnitTypeCode] = item
		}
		item.UnitCount += r.UnitCount
		item.Groups = append(item.Groups, SummaryGroup{Code: r.UnitGroupCode, Name: groupName, UnitCount: r.UnitCount})
	}
	return out, nil
}

func (s *Service) List(ctx context.Context, tenantID uint64, q ListQuery) (Page[UnitDTO], error) {
	repoQuery := repositories.BusinessUnitListQuery{
		TenantID:      tenantID,
		Skip:          q.Skip,
		Limit:         q.Limit,
		UnitTypeCode:  strings.TrimSpace(q.UnitTypeCode),
		UnitGroupCode: strings.TrimSpace(q.UnitGroupCode),
		Status:        strings.TrimSpace(q.Status),
		ScopedIDs:     q.ScopedIDs,
		UseScope:      q.UseScope,
	}
	if strings.TrimSpace(q.Keyword) != "" {
		repoQuery.KeywordLike = "%" + strings.TrimSpace(q.Keyword) + "%"
	}
	rows, total, err := s.repo.List(ctx, repoQuery)
	if err != nil {
		return Page[UnitDTO]{}, err
	}
	items := make([]UnitDTO, 0, len(rows))
	for _, r := range rows {
		items = append(items, unitDTO(r))
	}
	return Page[UnitDTO]{Items: items, Total: total, Skip: q.Skip, Limit: q.Limit}, nil
}

func (s *Service) Create(ctx context.Context, viewer dictionary.Viewer, tenantID uint64, body UnitPayload) (UnitDTO, error) {
	row, err := s.unitFromPayload(ctx, viewer, tenantID, body, nil)
	if err != nil {
		return UnitDTO{}, err
	}
	if err := s.repo.Create(ctx, &row); err != nil {
		return UnitDTO{}, err
	}
	return unitDTO(row), nil
}

func (s *Service) Update(ctx context.Context, viewer dictionary.Viewer, tenantID uint64, id uint64, body UnitPayload) (UnitDTO, error) {
	existing, err := s.repo.Get(ctx, tenantID, id)
	if err != nil {
		return UnitDTO{}, err
	}
	next, err := s.unitFromPayload(ctx, viewer, tenantID, body, &existing)
	if err != nil {
		return UnitDTO{}, err
	}
	if err := s.repo.Update(ctx, &existing, map[string]interface{}{
		"name": next.Name, "code": next.Code, "unit_type_code": next.UnitTypeCode, "unit_type_name": next.UnitTypeName,
		"unit_group_code": next.UnitGroupCode, "unit_group_name": next.UnitGroupName, "bu_type": next.BUType, "unit_form": next.UnitForm,
		"parent_id": next.ParentID, "attr_template_id": next.AttrTemplateID, "attrs": next.Attrs, "status": next.Status,
		"billing_enabled": next.BillingEnabled, "statistic_enabled": next.StatisticEnabled, "operation_enabled": next.OperationEnabled,
		"settlement_enabled": next.SettlementEnabled, "remark": next.Remark,
	}); err != nil {
		return UnitDTO{}, err
	}
	return unitDTO(existing), nil
}

func (s *Service) Archive(ctx context.Context, tenantID uint64, id uint64, tombstoneCode string) error {
	row, err := s.repo.Get(ctx, tenantID, id)
	if err != nil {
		return err
	}
	if err := s.ensureArchivable(ctx, tenantID, id); err != nil {
		return err
	}
	return s.repo.Archive(ctx, &row, tombstoneCode)
}

func (s *Service) Actors(ctx context.Context, tenantID uint64, unitID uint64) ([]ActorDTO, error) {
	if _, err := s.unitByID(ctx, tenantID, unitID); err != nil {
		return nil, err
	}
	rows, err := s.repo.ListActors(ctx, tenantID, unitID)
	if err != nil {
		return nil, err
	}
	items := make([]ActorDTO, 0, len(rows))
	for _, r := range rows {
		items = append(items, actorDTO(r))
	}
	return items, nil
}

func (s *Service) SaveActors(ctx context.Context, tenantID uint64, unitID uint64, actors []ActorPayload) ([]ActorDTO, error) {
	if _, err := s.unitByID(ctx, tenantID, unitID); err != nil {
		return nil, err
	}
	rows := make([]models.BusinessUnitActor, 0, len(actors))
	for _, input := range actors {
		row, err := s.actorFromPayload(ctx, tenantID, unitID, input)
		if err != nil {
			return nil, err
		}
		rows = append(rows, row)
	}
	if err := s.repo.ReplaceActors(ctx, tenantID, unitID, rows); err != nil {
		return nil, err
	}
	return s.Actors(ctx, tenantID, unitID)
}

func (s *Service) Relations(ctx context.Context, tenantID uint64, unitID uint64) ([]RelationDTO, error) {
	if _, err := s.unitByID(ctx, tenantID, unitID); err != nil {
		return nil, err
	}
	rows, err := s.repo.ListRelations(ctx, tenantID, unitID)
	if err != nil {
		return nil, err
	}
	items := make([]RelationDTO, 0, len(rows))
	for _, r := range rows {
		items = append(items, relationDTO(r))
	}
	return items, nil
}

func (s *Service) SaveRelations(ctx context.Context, tenantID uint64, unitID uint64, relations []RelationPayload) ([]RelationDTO, error) {
	if _, err := s.unitByID(ctx, tenantID, unitID); err != nil {
		return nil, err
	}
	rows := make([]models.BusinessUnitRelation, 0, len(relations))
	for _, input := range relations {
		row, err := s.relationFromPayload(ctx, tenantID, unitID, input)
		if err != nil {
			return nil, err
		}
		rows = append(rows, row)
	}
	if err := s.repo.ReplaceRelations(ctx, tenantID, unitID, rows); err != nil {
		return nil, err
	}
	return s.Relations(ctx, tenantID, unitID)
}

func (s *Service) MatchTemplate(ctx context.Context, tenantID uint64, unitTypeCode string, unitGroupCode string) (*TemplateDTO, []FieldDTO, error) {
	tpl, fields, err := s.repo.MatchTemplate(ctx, tenantID, strings.TrimSpace(unitTypeCode), strings.TrimSpace(unitGroupCode))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, []FieldDTO{}, nil
		}
		return nil, nil, err
	}
	t := templateDTO(tpl)
	return &t, fieldDTOs(fields), nil
}

func (s *Service) Templates(ctx context.Context, tenantID uint64, unitTypeCode string, unitGroupCode string, status string) ([]TemplateDTO, error) {
	rows, err := s.repo.ListTemplates(ctx, tenantID, unitTypeCode, unitGroupCode, status)
	if err != nil {
		return nil, err
	}
	items := make([]TemplateDTO, 0, len(rows))
	for _, row := range rows {
		items = append(items, templateDTO(row))
	}
	return items, nil
}

func (s *Service) Template(ctx context.Context, tenantID uint64, id uint64) (TemplateDTO, []FieldDTO, error) {
	tpl, fields, err := s.repo.GetTemplate(ctx, tenantID, id)
	if err != nil {
		return TemplateDTO{}, nil, err
	}
	return templateDTO(tpl), fieldDTOs(fields), nil
}

func (s *Service) CreateTemplate(ctx context.Context, viewer dictionary.Viewer, tenantID uint64, body TemplatePayload) (TemplateDTO, []FieldDTO, error) {
	tpl, fields, err := s.templateFromPayload(ctx, viewer, tenantID, body)
	if err != nil {
		return TemplateDTO{}, nil, err
	}
	if err := s.ensureTemplateUnique(ctx, tenantID, tpl, 0); err != nil {
		return TemplateDTO{}, nil, err
	}
	if err := s.repo.CreateTemplate(ctx, &tpl, fields); err != nil {
		return TemplateDTO{}, nil, err
	}
	return s.Template(ctx, tenantID, tpl.ID)
}

func (s *Service) UpdateTemplate(ctx context.Context, viewer dictionary.Viewer, tenantID uint64, id uint64, body TemplatePayload) (TemplateDTO, []FieldDTO, error) {
	existing, _, err := s.repo.GetTemplate(ctx, tenantID, id)
	if err != nil {
		return TemplateDTO{}, nil, err
	}
	if existing.TenantID == nil || *existing.TenantID != tenantID {
		return TemplateDTO{}, nil, errors.New("只能维护本主体属性模板")
	}
	next, fields, err := s.templateFromPayload(ctx, viewer, tenantID, body)
	if err != nil {
		return TemplateDTO{}, nil, err
	}
	if err := s.ensureTemplateUnique(ctx, tenantID, next, existing.ID); err != nil {
		return TemplateDTO{}, nil, err
	}
	if err := s.repo.UpdateTemplate(ctx, &existing, map[string]interface{}{
		"template_name": next.TemplateName, "unit_type_code": next.UnitTypeCode, "unit_type_name": next.UnitTypeName,
		"unit_group_code": next.UnitGroupCode, "unit_group_name": next.UnitGroupName, "sort_order": next.SortOrder,
		"status": next.Status, "remark": next.Remark,
	}, fields); err != nil {
		return TemplateDTO{}, nil, err
	}
	return s.Template(ctx, tenantID, existing.ID)
}

func (s *Service) ensureTemplateUnique(ctx context.Context, tenantID uint64, tpl models.BusinessUnitAttrTemplate, excludeID uint64) error {
	groupCode := ""
	if tpl.UnitGroupCode != nil {
		groupCode = *tpl.UnitGroupCode
	}
	exists, err := s.repo.TemplateScopeExists(ctx, tenantID, tpl.TemplateName, tpl.UnitTypeCode, groupCode, excludeID)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("属性模板已存在：同一适用范围下模板名称不能重复")
	}
	return nil
}

func (s *Service) ArchiveTemplate(ctx context.Context, tenantID uint64, id uint64) error {
	return s.repo.ArchiveTemplate(ctx, tenantID, id)
}

func (s *Service) unitFromPayload(ctx context.Context, viewer dictionary.Viewer, tenantID uint64, body UnitPayload, existing *models.BusinessUnit) (models.BusinessUnit, error) {
	name, code := strings.TrimSpace(body.Name), strings.TrimSpace(body.Code)
	if name == "" || code == "" {
		return models.BusinessUnit{}, errors.New("请填写业务单元名称和编码")
	}
	typeName, groupName, err := s.validateDictPair(ctx, viewer, body.UnitTypeCode, body.UnitGroupCode, false)
	if err != nil {
		return models.BusinessUnit{}, err
	}
	if err := s.validateParent(ctx, tenantID, existingID(existing), body.ParentID); err != nil {
		return models.BusinessUnit{}, err
	}
	attrs, err := normalizeJSONObject(body.Attrs)
	if err != nil {
		return models.BusinessUnit{}, err
	}
	if err := s.validateAttrs(ctx, tenantID, body.AttrTemplateID, attrs); err != nil {
		return models.BusinessUnit{}, err
	}
	status := body.Status
	if status != 0 && status != 1 {
		return models.BusinessUnit{}, errors.New("业务单元状态不正确")
	}
	unitTypeCode, unitGroupCode := strings.TrimSpace(body.UnitTypeCode), strings.TrimSpace(body.UnitGroupCode)
	return models.BusinessUnit{
		TenantID: tenantID, Name: name, Code: code, UnitTypeCode: &unitTypeCode, UnitTypeName: &typeName, UnitGroupCode: &unitGroupCode, UnitGroupName: &groupName,
		BUType: &unitTypeCode, UnitForm: &unitTypeCode, ParentID: body.ParentID, AttrTemplateID: body.AttrTemplateID, Attrs: attrs,
		Status: status, BillingEnabled: status == 1, StatisticEnabled: true, OperationEnabled: true, SettlementEnabled: status == 1, Remark: trimPtr(body.Remark),
	}, nil
}

func (s *Service) templateFromPayload(ctx context.Context, viewer dictionary.Viewer, tenantID uint64, body TemplatePayload) (models.BusinessUnitAttrTemplate, []models.BusinessUnitAttrTemplateField, error) {
	name := strings.TrimSpace(body.TemplateName)
	if name == "" {
		return models.BusinessUnitAttrTemplate{}, nil, errors.New("请填写模板名称")
	}
	groupCode := ""
	if body.UnitGroupCode != nil {
		groupCode = strings.TrimSpace(*body.UnitGroupCode)
	}
	typeName, groupName, err := s.validateDictPair(ctx, viewer, body.UnitTypeCode, groupCode, true)
	if err != nil {
		return models.BusinessUnitAttrTemplate{}, nil, err
	}
	status := strings.TrimSpace(body.Status)
	if status == "" {
		status = "active"
	}
	if status != "active" && status != "disabled" {
		return models.BusinessUnitAttrTemplate{}, nil, errors.New("模板状态不正确")
	}
	unitTypeCode := strings.TrimSpace(body.UnitTypeCode)
	var unitGroupCodePtr *string
	var unitGroupNamePtr *string
	if groupCode != "" {
		unitGroupCodePtr = &groupCode
		unitGroupNamePtr = &groupName
	}
	fields := make([]models.BusinessUnitAttrTemplateField, 0, len(body.Fields))
	for _, field := range body.Fields {
		row, err := templateFieldFromPayload(tenantID, field)
		if err != nil {
			return models.BusinessUnitAttrTemplate{}, nil, err
		}
		fields = append(fields, row)
	}
	return models.BusinessUnitAttrTemplate{
		TenantID: &tenantID, TemplateName: name, UnitTypeCode: unitTypeCode, UnitTypeName: typeName,
		UnitGroupCode: unitGroupCodePtr, UnitGroupName: unitGroupNamePtr, SortOrder: body.SortOrder,
		Status: status, Remark: trimPtr(body.Remark),
	}, fields, nil
}

func (s *Service) validateDictPair(ctx context.Context, viewer dictionary.Viewer, unitTypeCode string, unitGroupCode string, allowEmptyGroup bool) (string, string, error) {
	unitTypeCode, unitGroupCode = strings.TrimSpace(unitTypeCode), strings.TrimSpace(unitGroupCode)
	if unitTypeCode == "" {
		return "", "", errors.New("请选择业务单元类型")
	}
	if unitGroupCode == "" && !allowEmptyGroup {
		return "", "", errors.New("请选择业务单元分组")
	}
	_, roots, err := s.DictionaryTree(ctx, viewer)
	if err != nil {
		return "", "", err
	}
	for _, root := range roots {
		if root.Code != unitTypeCode {
			continue
		}
		if unitGroupCode == "" && allowEmptyGroup {
			return root.Name, "", nil
		}
		for _, child := range root.Children {
			if child.Code == unitGroupCode {
				return root.Name, child.Name, nil
			}
		}
		return "", "", errors.New("业务单元分组不属于所选业务单元类型")
	}
	return "", "", errors.New("业务单元类型不在字典范围内")
}

func (s *Service) unitByID(ctx context.Context, tenantID uint64, id uint64) (models.BusinessUnit, error) {
	return s.repo.Get(ctx, tenantID, id)
}

func (s *Service) validateParent(ctx context.Context, tenantID uint64, currentID uint64, parentID *uint64) error {
	if parentID == nil {
		return nil
	}
	if currentID > 0 && *parentID == currentID {
		return errors.New("父级业务单元不能选择自身")
	}
	parent, err := s.unitByID(ctx, tenantID, *parentID)
	if err != nil {
		return errors.New("父级业务单元不存在")
	}
	seen := map[uint64]struct{}{currentID: {}}
	for parent.ParentID != nil {
		if _, ok := seen[*parent.ParentID]; ok {
			return errors.New("父级业务单元不能形成循环")
		}
		seen[*parent.ParentID] = struct{}{}
		next, err := s.unitByID(ctx, tenantID, *parent.ParentID)
		if err != nil {
			return errors.New("父级业务单元不存在")
		}
		parent = next
	}
	return nil
}

func (s *Service) ensureArchivable(ctx context.Context, tenantID uint64, id uint64) error {
	childCount, err := s.repo.CountActiveChildren(ctx, tenantID, id)
	if err != nil {
		return err
	}
	if childCount > 0 {
		return errors.New("存在有效下级业务单元，不能归档")
	}
	relationCount, err := s.repo.CountActiveRelations(ctx, tenantID, id)
	if err != nil {
		return err
	}
	if relationCount > 0 {
		return errors.New("存在有效业务单元关联，不能归档")
	}
	return nil
}

func (s *Service) validateAttrs(ctx context.Context, tenantID uint64, templateID *uint64, attrs *string) error {
	if templateID == nil {
		return nil
	}
	fields, err := s.repo.ListTemplateFields(ctx, *templateID)
	if err != nil {
		return err
	}
	values := map[string]interface{}{}
	if attrs != nil && strings.TrimSpace(*attrs) != "" {
		if err := json.Unmarshal([]byte(*attrs), &values); err != nil {
			return errors.New("自定义属性必须是合法 JSON 对象")
		}
	}
	for _, f := range fields {
		if f.Required {
			if v, ok := values[f.FieldKey]; !ok || v == nil || strings.TrimSpace(fmt.Sprint(v)) == "" {
				return fmt.Errorf("请填写自定义属性：%s", f.FieldLabel)
			}
		}
	}
	return nil
}

func (s *Service) actorFromPayload(ctx context.Context, tenantID uint64, unitID uint64, input ActorPayload) (models.BusinessUnitActor, error) {
	actorType := strings.TrimSpace(input.ActorType)
	if actorType != "org" && actorType != "user" {
		return models.BusinessUnitActor{}, errors.New("责任方类型不正确")
	}
	if actorType == "org" {
		ok, err := s.repo.OrgExists(ctx, tenantID, input.ActorID)
		if err != nil || !ok {
			return models.BusinessUnitActor{}, errors.New("负责组织不存在")
		}
	} else {
		ok, err := s.repo.UserExists(ctx, tenantID, input.ActorID)
		if err != nil || !ok {
			return models.BusinessUnitActor{}, errors.New("负责人不存在")
		}
	}
	roleType, status := defaultString(input.RoleType, "owner"), defaultString(input.Status, "active")
	return models.BusinessUnitActor{TenantID: tenantID, BusinessUnitID: unitID, ActorType: actorType, ActorID: input.ActorID, RoleType: roleType, IncludeChildren: boolValue(input.IncludeChildren, actorType == "org"), Status: status}, nil
}

func (s *Service) relationFromPayload(ctx context.Context, tenantID uint64, unitID uint64, input RelationPayload) (models.BusinessUnitRelation, error) {
	if input.TargetUnitID == 0 || input.TargetUnitID == unitID {
		return models.BusinessUnitRelation{}, errors.New("关联业务单元不能选择自身")
	}
	target, err := s.unitByID(ctx, tenantID, input.TargetUnitID)
	if err != nil || target.Status != 1 {
		return models.BusinessUnitRelation{}, errors.New("关联业务单元不存在或未启用")
	}
	relationTypeCode := strings.TrimSpace(input.RelationTypeCode)
	if relationTypeCode == "" {
		return models.BusinessUnitRelation{}, errors.New("请选择关联关系类型")
	}
	return models.BusinessUnitRelation{TenantID: tenantID, SourceUnitID: unitID, TargetUnitID: input.TargetUnitID, RelationTypeCode: relationTypeCode, RelationTypeName: defaultString(input.RelationTypeName, relationTypeCode), Status: defaultString(input.Status, "active"), Remark: trimPtr(input.Remark)}, nil
}

func templateFieldFromPayload(tenantID uint64, input FieldPayload) (models.BusinessUnitAttrTemplateField, error) {
	key, label, fieldType := strings.TrimSpace(input.FieldKey), strings.TrimSpace(input.FieldLabel), strings.TrimSpace(input.FieldType)
	if key == "" || label == "" {
		return models.BusinessUnitAttrTemplateField{}, errors.New("模板字段 key 和名称不能为空")
	}
	if fieldType == "" {
		fieldType = "text"
	}
	switch fieldType {
	case "text", "number", "date", "select", "textarea", "switch":
	default:
		return models.BusinessUnitAttrTemplateField{}, errors.New("模板字段类型不正确")
	}
	status := defaultString(input.Status, "active")
	if status != "active" && status != "disabled" {
		return models.BusinessUnitAttrTemplateField{}, errors.New("模板字段状态不正确")
	}
	var options *string
	if len(input.OptionsJSON) > 0 && strings.TrimSpace(string(input.OptionsJSON)) != "" {
		var value interface{}
		if err := json.Unmarshal(input.OptionsJSON, &value); err != nil {
			return models.BusinessUnitAttrTemplateField{}, errors.New("字段选项必须是合法 JSON")
		}
		normalized, _ := json.Marshal(value)
		out := string(normalized)
		options = &out
	}
	return models.BusinessUnitAttrTemplateField{
		TenantID: &tenantID, FieldKey: key, FieldLabel: label, FieldType: fieldType, Required: boolValue(input.Required, false),
		DefaultValue: trimPtr(input.DefaultValue), Placeholder: trimPtr(input.Placeholder), OptionsJSON: options, SortOrder: input.SortOrder, Status: status,
	}, nil
}

func dictItemsToTree(items []dictionary.DictItemResponse) []DictNode {
	roots := []DictNode{}
	rootByCode := map[string]int{}
	canonicalID := map[uint64]uint64{}
	children := map[uint64][]DictNode{}
	childByParentCode := map[uint64]map[string]int{}
	for _, item := range items {
		if !item.Enabled {
			continue
		}
		node := DictNode{ID: item.ID, Code: item.Value, Name: item.Label, SortOrder: item.SortOrder, Children: []DictNode{}}
		if item.ParentID == nil {
			if index, ok := rootByCode[node.Code]; ok {
				canonicalID[item.ID] = roots[index].ID
				continue
			}
			rootByCode[node.Code] = len(roots)
			canonicalID[item.ID] = item.ID
			roots = append(roots, node)
			continue
		}
		parentID := *item.ParentID
		if resolved, ok := canonicalID[parentID]; ok {
			parentID = resolved
		}
		if childByParentCode[parentID] == nil {
			childByParentCode[parentID] = map[string]int{}
		}
		if _, ok := childByParentCode[parentID][node.Code]; ok {
			canonicalID[item.ID] = children[parentID][childByParentCode[parentID][node.Code]].ID
			continue
		}
		childByParentCode[parentID][node.Code] = len(children[parentID])
		canonicalID[item.ID] = item.ID
		children[parentID] = append(children[parentID], node)
	}
	for i := range roots {
		roots[i].Children = children[roots[i].ID]
	}
	return roots
}

func dictNameMap(nodes []DictNode) map[string]string {
	names := map[string]string{}
	var walk func([]DictNode)
	walk = func(list []DictNode) {
		for _, n := range list {
			names[n.Code] = n.Name
			walk(n.Children)
		}
	}
	walk(nodes)
	return names
}

func unitDTO(r models.BusinessUnit) UnitDTO {
	return UnitDTO{ID: r.ID, TenantID: r.TenantID, Name: r.Name, Code: r.Code, UnitTypeCode: r.UnitTypeCode, UnitTypeName: r.UnitTypeName, UnitGroupCode: r.UnitGroupCode, UnitGroupName: r.UnitGroupName, ParentID: r.ParentID, AttrTemplateID: r.AttrTemplateID, Attrs: r.Attrs, Status: r.Status, Remark: r.Remark}
}

func actorDTO(r models.BusinessUnitActor) ActorDTO {
	return ActorDTO{ID: r.ID, BusinessUnitID: r.BusinessUnitID, ActorType: r.ActorType, ActorID: r.ActorID, RoleType: r.RoleType, IncludeChildren: r.IncludeChildren, Status: r.Status}
}

func relationDTO(r models.BusinessUnitRelation) RelationDTO {
	return RelationDTO{ID: r.ID, SourceUnitID: r.SourceUnitID, TargetUnitID: r.TargetUnitID, RelationTypeCode: r.RelationTypeCode, RelationTypeName: r.RelationTypeName, Status: r.Status, Remark: r.Remark}
}

func templateDTO(r models.BusinessUnitAttrTemplate) TemplateDTO {
	return TemplateDTO{ID: r.ID, TemplateName: r.TemplateName, UnitTypeCode: r.UnitTypeCode, UnitTypeName: r.UnitTypeName, UnitGroupCode: r.UnitGroupCode, UnitGroupName: r.UnitGroupName, SortOrder: r.SortOrder, Status: r.Status, Remark: r.Remark}
}

func fieldDTOs(rows []models.BusinessUnitAttrTemplateField) []FieldDTO {
	items := make([]FieldDTO, 0, len(rows))
	for _, r := range rows {
		items = append(items, FieldDTO{ID: r.ID, TemplateID: r.TemplateID, FieldKey: r.FieldKey, FieldLabel: r.FieldLabel, FieldType: r.FieldType, Required: r.Required, DefaultValue: r.DefaultValue, Placeholder: r.Placeholder, OptionsJSON: r.OptionsJSON, SortOrder: r.SortOrder, Status: r.Status})
	}
	return items
}

func normalizeJSONObject(raw json.RawMessage) (*string, error) {
	if len(raw) == 0 || strings.TrimSpace(string(raw)) == "" {
		out := "{}"
		return &out, nil
	}
	var obj map[string]interface{}
	if err := json.Unmarshal(raw, &obj); err != nil {
		return nil, errors.New("自定义属性必须是合法 JSON 对象")
	}
	normalized, _ := json.Marshal(obj)
	out := string(normalized)
	return &out, nil
}

func existingID(row *models.BusinessUnit) uint64 {
	if row == nil {
		return 0
	}
	return row.ID
}

func trimPtr(v *string) *string {
	if v == nil || strings.TrimSpace(*v) == "" {
		return nil
	}
	out := strings.TrimSpace(*v)
	return &out
}

func boolValue(v *bool, fallback bool) bool {
	if v == nil {
		return fallback
	}
	return *v
}

func defaultString(v string, fallback string) string {
	if strings.TrimSpace(v) == "" {
		return fallback
	}
	return strings.TrimSpace(v)
}

func coalesceName(primary string, fallback string) string {
	if strings.TrimSpace(primary) != "" {
		return primary
	}
	return fallback
}
