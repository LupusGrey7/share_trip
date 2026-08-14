package model

type CheckResult struct {
	Allowed bool
	Reason  string
}

type CheckServiceRequest struct {
	CompanyID   string `json:"company_id"`
	ServiceCode string `json:"service_code"`
}

type CheckServiceResponse struct {
	Allowed bool   `json:"allowed"`
	Reason  string `json:"reason"`
}

func (c *CheckResult) IsAllowed() bool {
	return c.Allowed
}
