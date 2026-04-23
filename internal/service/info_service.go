package service

import (
	"context"
	"job4j.ru/share_trip/internal/repository"
	"job4j.ru/share_trip/internal/service/usecase"
)

type CommonService interface {
	GetDBInfo(ctx context.Context) (string, error)
}
type InfoService struct {
	infoCase usecase.BaseInfo
	repo     repository.InfoRepository
}

// NewInfoService - Конструктор
func NewInfoService(
	useCase usecase.BaseInfo,
	r repository.InfoRepository,
) *InfoService {
	return &InfoService{
		infoCase: useCase,
		repo:     r,
	}
}
