package contracts

type CheckResult struct {
	Allowed bool
	Reason  string
}

type CheckServiceRequest struct {
	CompanyID   string
	ServiceCode string
}

type CheckServiceResponse struct {
	Allowed bool   `json:"allowed"`
	Reason  string `json:"reason"`
}

func (c *CheckServiceResponse) ToCheckResult() CheckResult {
	return CheckResult{
		Allowed: c.Allowed,
		Reason:  c.Reason,
	}
}

func (c *CheckResult) IsAllowed() bool {
	return c.Allowed
}
