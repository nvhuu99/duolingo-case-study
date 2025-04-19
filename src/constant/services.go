package constant

const (
	SV_INP_MESG     = "input_message_api"
	SV_NOTI_BUILDER = "noti_builder"
	SV_PUSH_SENDER  = "push_noti_sender"
)

var ServiceTypes = map[string]string{
	SV_INP_MESG:     "api",
	SV_NOTI_BUILDER: "worker",
	SV_PUSH_SENDER:  "worker",
}

const (
	INP_MESG_REQUEST     = "input_message_request"
	RELAY_INP_MESG       = "relay_input_message"
	BUILD_PUSH_NOTI_MESG = "build_push_notification_message"
	SEND_PUSH_NOTI       = "send_push_notification"
)
