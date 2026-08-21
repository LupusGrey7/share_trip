package service

import (
	"context"
)

func (s *InfoService) GetDBInfo(ctx context.Context) (string, error) {

	v, err := s.infoCase.GetConnectInfo(ctx, s.repo)
	if err != nil {
		return "", err
	}
	return v, nil
}
