package system

import (
	"context"

	domain "saas_baseon_go/internal/domain/system"
)

type ParamService struct {
	repo domain.ParamRepository
}

func NewParamService(repo domain.ParamRepository) *ParamService {
	return &ParamService{repo: repo}
}

type CreateParamCommand struct {
	Key    string `json:"key"`
	Value  string `json:"value"`
	Remark string `json:"remark"`
}

func (s *ParamService) List(ctx context.Context) ([]domain.Param, error) {
	return s.repo.List(ctx)
}

func (s *ParamService) GetByKey(ctx context.Context, key string) (domain.Param, error) {
	return s.repo.FindByKey(ctx, key)
}

func (s *ParamService) Create(ctx context.Context, cmd CreateParamCommand) (domain.Param, error) {
	param, err := domain.NewParam(cmd.Key, cmd.Value, cmd.Remark)
	if err != nil {
		return domain.Param{}, err
	}
	return s.repo.Create(ctx, param)
}
