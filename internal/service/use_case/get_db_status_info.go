package use_case

import (
	"context"
	"job4j.ru/share_trip/internal/repository"
)

func (s *InfoUseCase) GetConnectInfo(ctx context.Context, repo repository.InfoRepository) (string, error) {
	v, err := repo.GetDbConnectInfo(ctx)
	if err != nil {
		return "", err
	}
	return v, nil
}
