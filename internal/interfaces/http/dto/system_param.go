package dto

type SystemParamResponse struct {
	ID             uint64 `json:"id"`
	Key            string `json:"param_key"`
	DefaultValue   string `json:"default_value"`
	ParamValue     string `json:"param_value"`
	Remark         string `json:"remark"`
	ValueType      string `json:"value_type"`
	TenantEditable bool   `json:"tenant_editable"`
	IsPlatformOnly bool   `json:"is_platform_only"`
	IsOverride     bool   `json:"is_override"`
}

type SystemParamBatchResponse struct {
	Values map[string]*string `json:"values"`
}
