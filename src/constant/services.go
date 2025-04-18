package constant

import "fmt"

const (
	SV_INP_MESG     = "api:input_message_api"
	SV_NOTI_BUILDER = "worker:noti_builder"
	SV_PUSH_SENDER  = "worker:push_noti_sender"
)

var ServiceTypes = map[string]string{
	SV_INP_MESG:     "api",
	SV_NOTI_BUILDER: "worker",
	SV_PUSH_SENDER:  "worker",
}

var (
	INP_MESG_REQUEST     = fmt.Sprintf("%v:%v", SV_INP_MESG, "input_message_request")
	RELAY_INP_MESG       = fmt.Sprintf("%v:%v", SV_NOTI_BUILDER, "relay_input_message")
	BUILD_PUSH_NOTI_MESG = fmt.Sprintf("%v:%v", SV_NOTI_BUILDER, "build_push_notification_message")
	SEND_PUSH_NOTI       = fmt.Sprintf("%v:%v", SV_PUSH_SENDER, "send_push_notification")
)
