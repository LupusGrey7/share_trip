package usecase

import (
	"context"

	contracts "job4j.ru/share_trip/internal/clients/http/contract"
)

func (c *ContractUsecase) CheckAvailableService(ctx context.Context, companyID string, serviceCode string) (contracts.CheckResult, error) {
	return c.contractClient.CheckAvailableService(ctx, companyID, serviceCode)
}
