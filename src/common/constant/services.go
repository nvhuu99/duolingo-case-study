package constant

const (
	SV_INP_MESG     = "message_input_api"
	SV_NOTI_BUILDER = "notification_builder"
	SV_PUSH_SENDER  = "push_notification_sender"
)

var ServiceTypes = map[string]string{
	SV_INP_MESG:     "api",
	SV_NOTI_BUILDER: "worker",
	SV_PUSH_SENDER:  "worker",
}

const (
	INP_MESG_REQUEST = "message_input_request"
	RELAY_INP_MESG   = "relay_input_message"
	BUILD_NOTI_MSG   = "build_push_notification_message"
	SEND_PUSH_NOTI   = "send_push_notification"
)
