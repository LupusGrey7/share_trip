package service

import (
	"context"
	"job4j.ru/share_trip/internal/repository"
)

type CommonService interface {
	GetDBInfo(ctx context.Context) (string, error)
}
type InfoService struct {
	repo *repository.RepoPg
}

// NewInfoService - Конструктор
func NewInfoService(r *repository.RepoPg) *InfoService {
	return &InfoService{repo: r}
}

func (s *InfoService) GetDBInfo(ctx context.Context) (string, error) {
	v, err := s.repo.GetDbConnectInfo(ctx)
	if err != nil {
		return "", err
	}
	return v, nil
}
