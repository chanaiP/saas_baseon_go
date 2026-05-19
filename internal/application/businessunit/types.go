package businessunit

import "encoding/json"

type DictNode struct {
	ID        uint64     `json:"id"`
	Code      string     `json:"code"`
	Name      string     `json:"name"`
	SortOrder int        `json:"sort_order"`
	Children  []DictNode `json:"children"`
}

type SummaryType struct {
	Code      string         `json:"code"`
	Name      string         `json:"name"`
	UnitCount int64          `json:"unit_count"`
	Groups    []SummaryGroup `json:"groups"`
}

type SummaryGroup struct {
	Code      string `json:"code"`
	Name      string `json:"name"`
	UnitCount int64  `json:"unit_count"`
}

type Page[T any] struct {
	Items []T   `json:"items"`
	Total int64 `json:"total"`
	Skip  int   `json:"skip"`
	Limit int   `json:"limit"`
}

type ListQuery struct {
	Skip          int
	Limit         int
	Keyword       string
	UnitTypeCode  string
	UnitGroupCode string
	Status        string
	ScopedIDs     []uint64
	UseScope      bool
}

type UnitPayload struct {
	UnitTypeCode   string          `json:"unit_type_code"`
	UnitGroupCode  string          `json:"unit_group_code"`
	Name           string          `json:"name"`
	Code           string          `json:"code"`
	ParentID       *uint64         `json:"parent_id"`
	AttrTemplateID *uint64         `json:"attr_template_id"`
	Attrs          json.RawMessage `json:"attrs"`
	Status         int             `json:"status"`
	Remark         *string         `json:"remark"`
}

type UnitDTO struct {
	ID             uint64  `json:"id"`
	TenantID       uint64  `json:"tenant_id"`
	Name           string  `json:"name"`
	Code           string  `json:"code"`
	UnitTypeCode   *string `json:"unit_type_code"`
	UnitTypeName   *string `json:"unit_type_name"`
	UnitGroupCode  *string `json:"unit_group_code"`
	UnitGroupName  *string `json:"unit_group_name"`
	ParentID       *uint64 `json:"parent_id"`
	AttrTemplateID *uint64 `json:"attr_template_id"`
	Attrs          *string `json:"attrs"`
	Status         int     `json:"status"`
	Remark         *string `json:"remark"`
}

type ActorPayload struct {
	ActorType       string `json:"actor_type"`
	ActorID         uint64 `json:"actor_id"`
	RoleType        string `json:"role_type"`
	IncludeChildren *bool  `json:"include_children"`
	Status          string `json:"status"`
}

type ActorDTO struct {
	ID              uint64 `json:"id"`
	BusinessUnitID  uint64 `json:"business_unit_id"`
	ActorType       string `json:"actor_type"`
	ActorID         uint64 `json:"actor_id"`
	RoleType        string `json:"role_type"`
	IncludeChildren bool   `json:"include_children"`
	Status          string `json:"status"`
}

type RelationPayload struct {
	TargetUnitID     uint64  `json:"target_unit_id"`
	RelationTypeCode string  `json:"relation_type_code"`
	RelationTypeName string  `json:"relation_type_name"`
	Status           string  `json:"status"`
	Remark           *string `json:"remark"`
}

type RelationDTO struct {
	ID               uint64  `json:"id"`
	SourceUnitID     uint64  `json:"source_unit_id"`
	TargetUnitID     uint64  `json:"target_unit_id"`
	RelationTypeCode string  `json:"relation_type_code"`
	RelationTypeName string  `json:"relation_type_name"`
	Status           string  `json:"status"`
	Remark           *string `json:"remark"`
}

type TemplatePayload struct {
	TemplateName  string         `json:"template_name"`
	UnitTypeCode  string         `json:"unit_type_code"`
	UnitGroupCode *string        `json:"unit_group_code"`
	SortOrder     int            `json:"sort_order"`
	Status        string         `json:"status"`
	Remark        *string        `json:"remark"`
	Fields        []FieldPayload `json:"fields"`
}

type FieldPayload struct {
	FieldKey     string          `json:"field_key"`
	FieldLabel   string          `json:"field_label"`
	FieldType    string          `json:"field_type"`
	Required     *bool           `json:"required"`
	DefaultValue *string         `json:"default_value"`
	Placeholder  *string         `json:"placeholder"`
	OptionsJSON  json.RawMessage `json:"options_json"`
	SortOrder    int             `json:"sort_order"`
	Status       string          `json:"status"`
}

type TemplateDTO struct {
	ID            uint64  `json:"id"`
	TemplateName  string  `json:"template_name"`
	UnitTypeCode  string  `json:"unit_type_code"`
	UnitTypeName  string  `json:"unit_type_name"`
	UnitGroupCode *string `json:"unit_group_code"`
	UnitGroupName *string `json:"unit_group_name"`
	SortOrder     int     `json:"sort_order"`
	Status        string  `json:"status"`
	Remark        *string `json:"remark"`
}

type FieldDTO struct {
	ID           uint64  `json:"id"`
	TemplateID   uint64  `json:"template_id"`
	FieldKey     string  `json:"field_key"`
	FieldLabel   string  `json:"field_label"`
	FieldType    string  `json:"field_type"`
	Required     bool    `json:"required"`
	DefaultValue *string `json:"default_value"`
	Placeholder  *string `json:"placeholder"`
	OptionsJSON  *string `json:"options_json"`
	SortOrder    int     `json:"sort_order"`
	Status       string  `json:"status"`
}
