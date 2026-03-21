package domain

import (
	"context"
	"job4j.ru/share_trip/internal/storage"
)

type CommonService interface {
	GetDBInfo(ctx context.Context) (string, error)
}
type Service struct {
	repo *storage.RepoPg
}

// NewCommonService - Конструктор
func NewCommonService(r *storage.RepoPg) *Service {
	return &Service{repo: r}
}

func (s *Service) GetDBInfo(ctx context.Context) (string, error) {
	v, err := s.repo.GetDbConnectInfo(ctx)
	if err != nil {
		return "", err
	}
	return v, nil
}
