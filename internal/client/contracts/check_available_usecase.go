package contracts

import (
	"context"
)

func (c *ContractUsecase) CheckAvailableService(ctx context.Context, companyID string, serviceCode string) (CheckResult, error) {
	return c.contractClient.CheckService(ctx, companyID, serviceCode)
}
