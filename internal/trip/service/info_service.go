package service

import (
	"context"

	"job4j.ru/share_trip/internal/storage"
	"job4j.ru/share_trip/internal/trip/usecase"
)

type CommonService interface {
	GetDBInfo(ctx context.Context) (string, error)
}
type InfoService struct {
	infoCase usecase.BaseInfo
	repo     storage.InfoRepository
}

// NewInfoService - Конструктор
func NewInfoService(
	useCase usecase.BaseInfo,
	r storage.InfoRepository,
) *InfoService {
	return &InfoService{
		infoCase: useCase,
		repo:     r,
	}
}
