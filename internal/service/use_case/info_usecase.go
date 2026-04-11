package use_case

import (
	"context"
	"job4j.ru/share_trip/internal/repository"
)

type BaseInfo interface {
	GetConnectInfo(ctx context.Context, repo repository.InfoRepository) (string, error)
}

type InfoUseCase struct {
}

func NewInfoUseCase() *InfoUseCase {
	return &InfoUseCase{}
}
