package data

type HttpRequest struct {
	RequestId        string `json:"request_id"`
	Timestamp        string `json:"timestamp"`
	Method           string `json:"method"`
	URL              string `json:"url"`
	StatusCode       int    `json:"status_code"`
	ClientAddr       string `json:"client_address"`
	UserAgent        string `json:"user_agent"`
	Referer          string `json:"referer"`
	Query            any    `json:"query"`
	Inputs           any    `json:"inputs"`
	Headers          any    `json:"headers"`
	ResponseTimeMs   int    `json:"response_time_ms"`
	ResponseBodySize int    `json:"response_body_size"`
	ResponseBodyData any    `json:"response_body_data"`
}
