package usecase

import (
	"context"

	"job4j.ru/share_trip/internal/client/contracts"
)

func (c *ContractUsecase) CheckAvailableService(ctx context.Context, companyID string, serviceCode string) (contracts.CheckResult, error) {
	return c.contractClient.CheckAvailableService(ctx, companyID, serviceCode)
}
