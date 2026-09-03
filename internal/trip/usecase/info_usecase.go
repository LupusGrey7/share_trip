package usecase

import (
	"context"

	"job4j.ru/share_trip/internal/storage"
)

type BaseInfo interface {
	GetConnectInfo(ctx context.Context, repo storage.InfoRepository) (string, error)
}

type InfoUseCase struct {
}

func NewInfoUseCase() *InfoUseCase {
	return &InfoUseCase{}
}
