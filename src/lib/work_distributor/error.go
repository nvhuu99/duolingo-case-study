package work_distributor

const (
	ERR_UNKNOWN                     = 501
	ERR_CONNECTION_FAILURE          = 502
	ERR_LOCK                        = 503
	ERR_WORKLOAD_EMPTY              = 504
	ERR_WORKLOAD_NOT_FOUND          = 505
	ERR_WORKLOAD_NOT_SET            = 506
	ERR_WORKLOAD_ATTRIBUTES_INVALID = 507
	ERR_WORKLOAD_CREATION           = 508
	ERR_WORKLOAD_DUPLICATION        = 509
	ERR_WORKLOAD_SWITCH             = 510
	ERR_WORKLOAD_ASSIGNMENT         = 511
	ERR_PURGE_FAILURE               = 512
)

var ErrMessages = map[int]string{
	ERR_UNKNOWN:                     "501 - unknown error",
	ERR_CONNECTION_FAILURE:          "502 - connection failure",
	ERR_LOCK:                        "503 - can not acquire lock",
	ERR_WORKLOAD_EMPTY:              "504 - workload empty",
	ERR_WORKLOAD_NOT_FOUND:          "505 - workload does not exist",
	ERR_WORKLOAD_NOT_SET:            "506 - workload not set",
	ERR_WORKLOAD_ATTRIBUTES_INVALID: "507 - workload attributes invalid",
	ERR_WORKLOAD_CREATION:           "508 - fail to register new workload",
	ERR_WORKLOAD_DUPLICATION:        "509 - workload duplication",
	ERR_WORKLOAD_SWITCH:             "510 - workload duplication",
	ERR_WORKLOAD_ASSIGNMENT:         "511 - workload assignment error",
	ERR_PURGE_FAILURE:               "512 - fail to purge workloads",
}
