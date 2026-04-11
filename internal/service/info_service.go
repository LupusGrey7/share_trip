package service

import (
	"context"
	"job4j.ru/share_trip/internal/repository"
	"job4j.ru/share_trip/internal/service/use_case"
)

type CommonService interface {
	GetDBInfo(ctx context.Context) (string, error)
}
type InfoService struct {
	infoCase use_case.BaseInfo
	repo     repository.InfoRepository
}

// NewInfoService - Конструктор
func NewInfoService(
	useCase use_case.BaseInfo,
	r repository.InfoRepository,
) *InfoService {
	return &InfoService{
		infoCase: useCase,
		repo:     r,
	}
}
