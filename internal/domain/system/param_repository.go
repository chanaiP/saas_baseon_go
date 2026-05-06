package system

import "context"

type ParamRepository interface {
	List(ctx context.Context) ([]Param, error)
	FindByKey(ctx context.Context, key string) (Param, error)
	Create(ctx context.Context, param Param) (Param, error)
}
