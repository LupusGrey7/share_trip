package contracts

import (
	"context"
)

type BaseContractUsecase interface {
	CheckAvailableService(ctx context.Context, companyID string, serviceCode string) (CheckResult, error)
}

type ContractUsecase struct {
	contractClient BaseContractClient
}

func NewContractUsecase(contractClient BaseContractClient) *ContractUsecase {
	return &ContractUsecase{
		contractClient: contractClient,
	}
}
