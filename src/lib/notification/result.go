package notification

type Result struct {
	Success       bool     `json:"success"`
	Error         error    `json:"error"`
	SuccessCount  int      `json:"success_count"`
	FailureCount  int      `json:"failure_count"`
	FailureTokens []string `json:"failure_tokens"`
}
