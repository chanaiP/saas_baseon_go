package dto

import "encoding/json"

type BusinessUnitMutationRequest struct {
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

type BusinessUnitActorsSaveRequest struct {
	Actors []BusinessUnitActorRequest `json:"actors"`
}

type BusinessUnitActorRequest struct {
	ActorType       string `json:"actor_type"`
	ActorID         uint64 `json:"actor_id"`
	RoleType        string `json:"role_type"`
	IncludeChildren *bool  `json:"include_children"`
	Status          string `json:"status"`
}

type BusinessUnitRelationsSaveRequest struct {
	Relations []BusinessUnitRelationRequest `json:"relations"`
}

type BusinessUnitRelationRequest struct {
	TargetUnitID     uint64  `json:"target_unit_id"`
	RelationTypeCode string  `json:"relation_type_code"`
	RelationTypeName string  `json:"relation_type_name"`
	Status           string  `json:"status"`
	Remark           *string `json:"remark"`
}
