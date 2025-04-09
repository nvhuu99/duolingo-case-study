package rest_http

const (
	STATUS_OK         = 200
	STATUS_CREATED    = 201
	STATUS_INVALID    = 400
	STATUS_NOT_FOUND  = 404
	STATUS_SERVER_ERR = 500
)

var (
	defaultMessages = map[int]string{
		STATUS_OK:         "Ok",
		STATUS_CREATED:    "Created",
		STATUS_INVALID:    "Invalid Request",
		STATUS_NOT_FOUND:  "Not Found",
		STATUS_SERVER_ERR: "Internal Server Error",
	}
)
