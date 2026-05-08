package dto

type ListResponse[T any] struct {
	Items []T `json:"items"`
	Total int `json:"total"`
}

type PaginatedResponse[T any] struct {
	Items []T   `json:"items"`
	Total int64 `json:"total"`
	Skip  int   `json:"skip"`
	Limit int   `json:"limit"`
}

type IDResponse struct {
	ID uint64 `json:"id"`
}

type DeletedResponse struct {
	Deleted uint64 `json:"deleted"`
}
