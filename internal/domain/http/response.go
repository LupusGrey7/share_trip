package http

// Response model info
// Response is a common response structure
// @Description Common API response format
type Response struct {
	Success bool        `json:"success" example:"false"`
	Message string      `json:"error" example:"Invalid request format"`
	Data    interface{} `json:"data"`
}
