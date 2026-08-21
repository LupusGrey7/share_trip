package usecase

import (
	"context"

	"job4j.ru/share_trip/internal/storage"
)

func (s *InfoUseCase) GetConnectInfo(ctx context.Context, repo storage.InfoRepository) (string, error) {
	v, err := repo.GetDbConnectInfo(ctx)
	if err != nil {
		return "", err
	}
	return v, nil
}
