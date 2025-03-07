package model

type PushNotiMessage struct {
	Id			string	`json:"id"`
	DeviceToken	string  `json:"device_token"`
	Content		string	`json:"content"`
}
