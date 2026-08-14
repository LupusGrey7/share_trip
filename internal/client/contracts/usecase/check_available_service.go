package usecase

import (
	"context"

	"job4j.ru/share_trip/internal/client/contracts/model"
)

func (c *ContractUsecase) CheckAvailableService(ctx context.Context, companyID string, serviceCode string) (model.CheckResult, error) {
	return c.contractClient.CheckService(ctx, companyID, serviceCode)
}
