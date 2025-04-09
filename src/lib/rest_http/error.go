package rest_http

const (
	ERR_URL_DECODE_FAILURE    = 501
	ERR_ARGUMENT_NOT_ENCLOSED = 502
	ERR_ROUTE_NOT_FOUND       = 503
	ERR_SERVER_PANIC          = 504
)

var ErrMessages = map[int]string{
	ERR_URL_DECODE_FAILURE:    "501 - request url contains invalid url-encode character",
	ERR_ARGUMENT_NOT_ENCLOSED: "502 - route argument is not enclosed with \"{}\"",
	ERR_ROUTE_NOT_FOUND:       "503 - the requested route does not exist",
	ERR_SERVER_PANIC:          "504 - server panic",
}
