package model

type PushNotiMessage struct {
	Id				string		`json:"id"`
	Content			string		`json:"content"`
	DeviceTokens	[]string  	`json:"device_tokens"`
}
