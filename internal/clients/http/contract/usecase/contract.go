package usecase

import (
	"context"

	contracts "job4j.ru/share_trip/internal/clients/http/contract"
)

type BaseContractUsecase interface {
	CheckAvailableService(ctx context.Context, companyID string, serviceCode string) (contracts.CheckResult, error)
}

type ContractUsecase struct {
	contractClient contracts.BaseContractClient
}

func NewContractUsecase(contractClient contracts.BaseContractClient) *ContractUsecase {
	return &ContractUsecase{
		contractClient: contractClient,
	}
}
