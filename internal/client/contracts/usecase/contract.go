package usecase

import (
	"context"

	"job4j.ru/share_trip/internal/client/contracts"
	"job4j.ru/share_trip/internal/client/contracts/model"
)

type BaseContractUsecase interface {
	CheckAvailableService(ctx context.Context, companyID string, serviceCode string) (model.CheckResult, error)
}

type ContractUsecase struct {
	contractClient contracts.BaseContractClient
}

func NewContractUsecase(contractClient contracts.BaseContractClient) *ContractUsecase {
	return &ContractUsecase{
		contractClient: contractClient,
	}
}
