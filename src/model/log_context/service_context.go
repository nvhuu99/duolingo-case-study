package log_context

type ServiceContext struct {
	Type            string `json:"type"`
	Name            string `json:"name"`
	Operation       string `json:"operation"`
	InstanceId      string `json:"instance_id"`
	InstanceAddress string `json:"instance_address"`
}
